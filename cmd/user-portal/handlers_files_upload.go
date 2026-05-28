package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
	"Recontext.online/pkg/database"
	"Recontext.online/pkg/storage"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

// ─── POST /api/v1/files/init ──────────────────────────────────────────────────

// InitFileUpload godoc
// @Summary Инициализировать S3 multipart загрузку
// @Description Создаёт запись в БД, инициирует S3 multipart upload и возвращает presigned PUT URL
// @Description для каждой части. Клиент загружает части параллельно (PUT на upload_url),
// @Description собирает ETags и передаёт их в POST /api/v1/files/{id}/confirm.
// @Tags Files
// @Accept json
// @Produce json
// @Param request body models.InitUploadRequest true "Параметры файла (chunk_size опционален, по умолчанию 10 МБ)"
// @Success 201 {object} models.InitUploadResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 503 {object} models.ErrorResponse "MinIO недоступен"
// @Security BearerAuth
// @Router /api/v1/files/init [post]
func (up *UserPortal) initFileUploadHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	var req models.InitUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	if req.FileName == "" {
		up.respondWithError(w, http.StatusBadRequest, "file_name is required", "")
		return
	}
	if req.FileSize <= 0 {
		up.respondWithError(w, http.StatusBadRequest, "file_size must be positive", "")
		return
	}

	if up.minioClient == nil {
		up.respondWithError(w, http.StatusServiceUnavailable, "Storage service unavailable", "MinIO not configured")
		return
	}

	bucket := up.minioClient.GetBucket()
	fileID := uuid.New()
	objectPath := fmt.Sprintf("uploads/%s/%s/%s", claims.UserID, fileID, req.FileName)

	title := req.Title
	if title == "" {
		title = req.FileName
	}
	language := req.Language
	if language == "" {
		language = "auto"
	}
	mimeType := req.MimeType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	dbFile := &database.UploadedFile{
		ID:           fileID,
		Title:        title,
		Filename:     fmt.Sprintf("%d-%s", time.Now().Unix(), req.FileName),
		OriginalName: req.FileName,
		FileSize:     req.FileSize,
		MimeType:     mimeType,
		StoragePath:  objectPath,
		Bucket:       bucket,
		UserID:       claims.UserID,
		Language:     language,
		Status:       "pending",
		UploadedAt:   time.Now(),
	}

	if err := up.db.CreateUploadedFileV2(dbFile); err != nil {
		up.logger.Errorf("[Files/Init] Failed to create file record: %v", err)
		up.respondWithError(w, http.StatusInternalServerError, "Failed to create file record", err.Error())
		return
	}

	expiry := 30 * time.Minute
	uploadID, storageParts, err := up.minioClient.InitiateMultipartUpload(
		r.Context(), objectPath, mimeType, req.FileSize, req.ChunkSize, expiry,
	)
	if err != nil {
		up.logger.Errorf("[Files/Init] Failed to initiate multipart upload: %v", err)
		_ = up.db.DeleteUploadedFile(fileID.String())
		up.respondWithError(w, http.StatusInternalServerError, "Failed to initiate multipart upload", err.Error())
		return
	}

	parts := make([]models.MultipartPart, len(storageParts))
	for i, p := range storageParts {
		parts[i] = models.MultipartPart{
			PartNumber: p.PartNumber,
			UploadURL:  p.UploadURL,
			Offset:     p.Offset,
			Size:       p.Size,
		}
	}

	resp := models.InitUploadResponse{
		FileID:       fileID,
		UploadID:     uploadID,
		UploadMethod: "MULTIPART",
		Parts:        parts,
		StoragePath:  objectPath,
		ExpiresAt:    time.Now().Add(expiry),
	}

	up.logger.Infof("[Files/Init] File %s created for user %s, multipart uploadId=%s, %d parts",
		fileID, claims.UserID, uploadID, len(parts))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// ─── POST /api/v1/files/{id}/confirm ─────────────────────────────────────────

// ConfirmFileUpload godoc
// @Summary Завершить S3 multipart загрузку
// @Description Принимает upload_id и ETags всех частей, вызывает CompleteMultipartUpload и ставит задачу в очередь.
// @Description При success=false — AbortMultipartUpload и удаление записи из БД.
// @Tags Files
// @Accept json
// @Produce json
// @Param id path string true "ID файла"
// @Param request body models.ConfirmUploadRequest true "upload_id + parts с ETags (или success=false для отмены)"
// @Success 200 {object} models.ConfirmUploadResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/files/{id}/confirm [post]
func (up *UserPortal) confirmFileUploadHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	fileIDStr := extractPathSegment(r.URL.Path, "/api/v1/files/", "/confirm")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid file ID", err.Error())
		return
	}

	dbFile, err := up.db.GetUploadedFileV2(fileID)
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "File not found", err.Error())
		return
	}
	if dbFile.UserID != claims.UserID && claims.Role != models.RoleAdmin {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "")
		return
	}
	if dbFile.Status != "pending" {
		up.respondWithError(w, http.StatusConflict, "File already confirmed or processed", dbFile.Status)
		return
	}

	var req models.ConfirmUploadRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	// success=false — клиент сигнализирует об отмене или ошибке загрузки
	if req.Success != nil && !*req.Success {
		if up.minioClient != nil && req.UploadID != "" {
			if abortErr := up.minioClient.AbortMultipartUpload(context.Background(), dbFile.StoragePath, req.UploadID); abortErr != nil {
				up.logger.Errorf("[Files/Confirm] AbortMultipartUpload failed for %s: %v", fileID, abortErr)
			}
		}
		_ = up.db.DeleteUploadedFile(fileID.String())
		up.logger.Infof("[Files/Confirm] Upload cancelled for file %s by user %s", fileID, claims.UserID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.ConfirmUploadResponse{
			FileID:  fileID,
			Status:  "cancelled",
			Message: "Upload cancelled, multipart aborted, file record deleted",
		})
		return
	}

	// Валидация обязательных полей для успешного confirm
	if req.UploadID == "" {
		up.respondWithError(w, http.StatusBadRequest, "upload_id is required", "")
		return
	}
	if len(req.Parts) == 0 {
		up.respondWithError(w, http.StatusBadRequest, "parts is required and must not be empty", "")
		return
	}

	// Завершаем multipart upload — MinIO собирает объект из загруженных частей
	if up.minioClient != nil {
		completeParts := make([]storage.CompletePart, len(req.Parts))
		for i, p := range req.Parts {
			completeParts[i] = storage.CompletePart{PartNumber: p.PartNumber, ETag: p.ETag}
		}
		if completeErr := up.minioClient.CompleteMultipartUpload(r.Context(), dbFile.StoragePath, req.UploadID, completeParts); completeErr != nil {
			up.logger.Errorf("[Files/Confirm] CompleteMultipartUpload failed for %s: %v", fileID, completeErr)
			up.respondWithError(w, http.StatusUnprocessableEntity, "Failed to complete multipart upload", completeErr.Error())
			return
		}
	}

	if err := up.db.ConfirmUpload(fileID, ""); err != nil {
		up.logger.Errorf("[Files/Confirm] Failed to confirm upload in DB: %v", err)
		up.respondWithError(w, http.StatusInternalServerError, "Failed to confirm upload", err.Error())
		return
	}

	// Queue transcription task
	if up.redisPublisher != nil {
		bucket := dbFile.Bucket
		if bucket == "" {
			bucket = "recontext"
		}
		if err := up.redisPublisher.PublishTranscriptionTask(
			fileID,
			claims.UserID,
			bucket,
			dbFile.StoragePath,
			"upload",
			dbFile.Language,
		); err != nil {
			up.logger.Errorf("[Files/Confirm] Failed to queue transcription: %v", err)
			_ = up.db.UpdateFileStatus(fileID.String(), models.StatusFailed)
			up.respondWithError(w, http.StatusInternalServerError, "Failed to queue transcription task", err.Error())
			return
		}
		up.logger.Infof("[Files/Confirm] Transcription task queued for file %s", fileID)
	} else {
		up.logger.Info("[Files/Confirm] Redis not available — transcription task not queued")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.ConfirmUploadResponse{
		FileID:  fileID,
		Status:  "queued",
		Message: "File confirmed, transcription task queued",
	})
}

// ─── GET /api/v1/files/{id}/status ───────────────────────────────────────────

// GetFileStatus godoc
// @Summary Статус обработки файла
// @Description Возвращает текущий статус транскрибации файла для polling
// @Tags Files
// @Produce json
// @Param id path string true "ID файла"
// @Success 200 {object} models.FileStatusResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/files/{id}/status [get]
func (up *UserPortal) getFileStatusHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	fileIDStr := extractPathSegment(r.URL.Path, "/api/v1/files/", "/status")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid file ID", err.Error())
		return
	}

	dbFile, err := up.db.GetUploadedFileV2(fileID)
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "File not found", err.Error())
		return
	}
	if dbFile.UserID != claims.UserID && claims.Role != models.RoleAdmin {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "")
		return
	}

	errMsg := ""
	if dbFile.ErrorMessage != nil {
		errMsg = *dbFile.ErrorMessage
	}
	resp := models.FileStatusResponse{
		FileID:    fileID,
		Status:    dbFile.Status,
		Progress:  dbFile.Progress,
		Stage:     dbFile.Stage,
		Error:     errMsg,
		UpdatedAt: dbFile.UploadedAt,
	}
	if dbFile.ProcessedAt != nil {
		resp.UpdatedAt = *dbFile.ProcessedAt
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ─── GET /api/v1/files ────────────────────────────────────────────────────────

// ListFilesV2 godoc
// @Summary Список файлов пользователя
// @Description Постраничный список загруженных файлов с фильтрами и поиском
// @Tags Files
// @Produce json
// @Param page      query int    false "Номер страницы (с 1)"                    default(1)
// @Param page_size query int    false "Размер страницы (макс 100)"              default(20)
// @Param status    query string false "Фильтр по статусу: pending,queued,processing,completed,failed"
// @Param search    query string false "Поиск по названию или имени файла (ILIKE)"
// @Param mime_type query string false "Фильтр по MIME-типу (например video/mp4)"
// @Param language  query string false "Фильтр по языку (ru, en, auto…)"
// @Param date_from query string false "Загружен не ранее (RFC3339, например 2024-01-01T00:00:00Z)"
// @Param date_to   query string false "Загружен не позднее (RFC3339)"
// @Success 200 {object} models.FileListResponse
// @Failure 401 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/files [get]
func (up *UserPortal) listFilesV2Handler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	page := queryInt(r, "page", 1)
	pageSize := queryInt(r, "page_size", 20)
	if pageSize > 100 {
		pageSize = 100
	}

	q := r.URL.Query()
	status := q.Get("status")
	search := q.Get("search")
	mimeType := q.Get("mime_type")
	language := q.Get("language")

	var dateFrom, dateTo *time.Time
	if v := q.Get("date_from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			dateFrom = &t
		} else {
			up.respondWithError(w, http.StatusBadRequest, "Invalid date_from format, use RFC3339", err.Error())
			return
		}
	}
	if v := q.Get("date_to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			dateTo = &t
		} else {
			up.respondWithError(w, http.StatusBadRequest, "Invalid date_to format, use RFC3339", err.Error())
			return
		}
	}

	files, total, err := up.db.ListUploadedFilesV2(
		claims.UserID, page, pageSize,
		status, search, mimeType, language,
		dateFrom, dateTo,
	)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to list files", err.Error())
		return
	}

	items := make([]models.FileListItem, 0, len(files))
	for _, f := range files {
		items = append(items, models.FileListItem{
			ID:          f.ID,
			Title:       f.Title,
			FileName:    f.OriginalName,
			FileSize:    f.FileSize,
			MimeType:    f.MimeType,
			Duration:    f.Duration,
			Status:      f.Status,
			Progress:    f.Progress,
			Language:    f.Language,
			UploadedAt:  f.UploadedAt,
			CompletedAt: f.ProcessedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.FileListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// ─── GET /api/v1/files/{id} ───────────────────────────────────────────────────

// GetFileDetail godoc
// @Summary Детали файла
// @Description Возвращает полные метаданные файла, включая сводку по транскрипции
// @Tags Files
// @Produce json
// @Param id path string true "ID файла"
// @Success 200 {object} models.FileDetailResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/files/{id} [get]
func (up *UserPortal) getFileDetailHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	fileID, err := uuid.Parse(lastPathSegment(r.URL.Path))
	if err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid file ID", err.Error())
		return
	}

	dbFile, err := up.db.GetUploadedFileV2(fileID)
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "File not found", err.Error())
		return
	}
	if dbFile.UserID != claims.UserID && claims.Role != models.RoleAdmin {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "")
		return
	}

	errMsg := ""
	if dbFile.ErrorMessage != nil {
		errMsg = *dbFile.ErrorMessage
	}

	resp := models.FileDetailResponse{
		ID:          fileID,
		Title:       dbFile.Title,
		FileName:    dbFile.OriginalName,
		FileSize:    dbFile.FileSize,
		MimeType:    dbFile.MimeType,
		Duration:    dbFile.Duration,
		Status:      dbFile.Status,
		Progress:    dbFile.Progress,
		Stage:       dbFile.Stage,
		Language:    dbFile.Language,
		ErrorMsg:    errMsg,
		UploadedAt:  dbFile.UploadedAt,
		CompletedAt: dbFile.ProcessedAt,
		VideoURL:    fmt.Sprintf("/api/v1/files/%s/video", fileID),
	}

	// Attach transcript summary when processing is complete
	if dbFile.Status == "completed" {
		var phraseCount int64
		up.db.DB.Model(&database.FileTranscriptionPhrase{}).Where("file_id = ?", fileID).Count(&phraseCount)
		speakers, _ := up.db.GetFileSpeakers(fileID)
		summary, _ := up.db.GetFileSummary(fileID)

		resp.Transcript = &struct {
			PhraseCount  int64    `json:"phrase_count"`
			SpeakerCount int      `json:"speaker_count"`
			HasSummary   bool     `json:"has_summary"`
			Speakers     []string `json:"speakers,omitempty"`
		}{
			PhraseCount:  phraseCount,
			SpeakerCount: len(speakers),
			HasSummary:   summary != nil && summary.Status == "completed",
			Speakers:     speakers,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ─── DELETE /api/v1/files/{id} ────────────────────────────────────────────────

// DeleteFile godoc
// @Summary Удалить файл
// @Description Удаляет файл из БД и из MinIO
// @Tags Files
// @Produce json
// @Param id path string true "ID файла"
// @Success 204 "Файл удалён"
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/files/{id} [delete]
func (up *UserPortal) deleteFileHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	fileID, err := uuid.Parse(lastPathSegment(r.URL.Path))
	if err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid file ID", err.Error())
		return
	}

	dbFile, err := up.db.GetUploadedFileV2(fileID)
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "File not found", err.Error())
		return
	}
	if dbFile.UserID != claims.UserID && claims.Role != models.RoleAdmin {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "")
		return
	}

	// Remove from MinIO (best-effort)
	if up.minioClient != nil {
		if err := up.minioClient.DeleteFile(context.Background(), dbFile.StoragePath); err != nil {
			up.logger.Infof("[Files/Delete] MinIO delete failed for %s: %v", dbFile.StoragePath, err)
		}
	}

	if err := up.db.DeleteUploadedFile(fileID.String()); err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to delete file", err.Error())
		return
	}

	up.logger.Infof("[Files/Delete] File %s deleted by user %s", fileID, claims.UserID)
	w.WriteHeader(http.StatusNoContent)
}

// ─── GET /api/v1/files/{id}/transcript ───────────────────────────────────────

// GetFileTranscript godoc
// @Summary Транскрипция файла
// @Description Возвращает постраничный список фраз транскрипции с временными метками
// @Tags Files
// @Produce json
// @Param id path string true "ID файла"
// @Param page query int false "Страница" default(1)
// @Param page_size query int false "Размер страницы" default(100)
// @Param speaker query string false "Фильтр по спикеру (SPEAKER_00 и т.д.)"
// @Success 200 {object} models.FileTranscriptResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/files/{id}/transcript [get]
func (up *UserPortal) getFileTranscriptHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	fileIDStr := extractPathSegment(r.URL.Path, "/api/v1/files/", "/transcript")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid file ID", err.Error())
		return
	}

	dbFile, err := up.db.GetUploadedFileV2(fileID)
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "File not found", err.Error())
		return
	}
	if dbFile.UserID != claims.UserID && claims.Role != models.RoleAdmin {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "")
		return
	}

	page := queryInt(r, "page", 1)
	pageSize := queryInt(r, "page_size", 100)
	if pageSize > 500 {
		pageSize = 500
	}
	speaker := r.URL.Query().Get("speaker")

	phrases, total, err := up.db.GetFilePhrasesPage(fileID, page, pageSize, speaker)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get transcript", err.Error())
		return
	}

	speakers, _ := up.db.GetFileSpeakers(fileID)

	items := make([]models.FilePhraseItem, 0, len(phrases))
	for _, p := range phrases {
		items = append(items, models.FilePhraseItem{
			PhraseIndex: p.PhraseIndex,
			StartTime:   p.StartTime,
			EndTime:     p.EndTime,
			Text:        p.Text,
			Speaker:     p.Speaker,
			Confidence:  p.Confidence,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.FileTranscriptResponse{
		FileID:   fileID,
		Language: dbFile.Language,
		Duration: dbFile.Duration,
		Speakers: speakers,
		Phrases:  items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// ─── GET /api/v1/files/{id}/summary ──────────────────────────────────────────

// GetFileSummary godoc
// @Summary Резюме транскрипции
// @Description Возвращает краткое содержание, ключевые темы и задачи по транскрипции файла
// @Tags Files
// @Produce json
// @Param id path string true "ID файла"
// @Success 200 {object} models.FileSummaryResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/files/{id}/summary [get]
func (up *UserPortal) getFileSummaryHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	fileIDStr := extractPathSegment(r.URL.Path, "/api/v1/files/", "/summary")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid file ID", err.Error())
		return
	}

	dbFile, err := up.db.GetUploadedFileV2(fileID)
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "File not found", err.Error())
		return
	}
	if dbFile.UserID != claims.UserID && claims.Role != models.RoleAdmin {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "")
		return
	}

	summary, err := up.db.GetFileSummary(fileID)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get summary", err.Error())
		return
	}
	if summary == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.FileSummaryResponse{
			FileID: fileID,
			Status: "pending",
		})
		return
	}

	resp := models.FileSummaryResponse{
		FileID:      fileID,
		Summary:     summary.Summary,
		SummaryRu:   summary.SummaryRu,
		KeyTopics:   []string(summary.KeyTopics),
		ActionItems: []string(summary.ActionItems),
		Status:      summary.Status,
		GeneratedAt: summary.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ─── GET /api/v1/files/{id}/video ────────────────────────────────────────────

// GetFileVideo godoc
// @Summary Получить видео файл
// @Description Возвращает presigned GET URL для стриминга видео из MinIO
// @Tags Files
// @Produce json
// @Param id path string true "ID файла"
// @Success 200 {object} map[string]string
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/files/{id}/video [get]
func (up *UserPortal) getFileVideoHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	fileIDStr := extractPathSegment(r.URL.Path, "/api/v1/files/", "/video")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid file ID", err.Error())
		return
	}

	dbFile, err := up.db.GetUploadedFileV2(fileID)
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "File not found", err.Error())
		return
	}
	if dbFile.UserID != claims.UserID && claims.Role != models.RoleAdmin {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "")
		return
	}

	if up.minioClient == nil {
		up.respondWithError(w, http.StatusServiceUnavailable, "Storage service unavailable", "")
		return
	}

	// Stream directly if Range header is present (for seekable video player)
	if r.Header.Get("Range") != "" {
		up.streamFileFromMinIO(w, r, dbFile.StoragePath, dbFile.MimeType)
		return
	}

	presignedURL, err := up.minioClient.PresignedGetObject(r.Context(), dbFile.StoragePath, 2*time.Hour)
	if err != nil {
		up.logger.Errorf("[Files/Video] Failed to generate presigned GET URL: %v", err)
		up.respondWithError(w, http.StatusInternalServerError, "Failed to generate video URL", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": presignedURL})
}

// streamFileFromMinIO proxies a MinIO object with Range support.
func (up *UserPortal) streamFileFromMinIO(w http.ResponseWriter, r *http.Request, objectPath, mimeType string) {
	if up.minioClient == nil {
		http.Error(w, "storage unavailable", http.StatusServiceUnavailable)
		return
	}

	obj, err := up.minioClient.GetClient().GetObject(
		r.Context(),
		up.minioClient.GetBucket(),
		objectPath,
		buildMinIOGetOptions(r),
	)
	if err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	defer obj.Close()

	stat, err := obj.Stat()
	if err != nil {
		http.Error(w, "object not found", http.StatusNotFound)
		return
	}

	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size, 10))
	w.Header().Set("Cache-Control", "private, max-age=3600")

	io.Copy(w, obj)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// extractPathSegment returns the segment between prefix and suffix in a URL path.
// e.g. "/api/v1/files/abc-123/confirm", "/api/v1/files/", "/confirm" → "abc-123"
func extractPathSegment(path, prefix, suffix string) string {
	s := strings.TrimPrefix(path, prefix)
	s = strings.TrimSuffix(s, suffix)
	return s
}

// lastPathSegment returns the final path segment (after the last /).
func lastPathSegment(path string) string {
	parts := strings.Split(strings.TrimRight(path, "/"), "/")
	return parts[len(parts)-1]
}

// queryInt reads an integer query parameter with a default fallback.
func queryInt(r *http.Request, key string, def int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

// buildMinIOGetOptions converts the HTTP Range header into MinIO GetObjectOptions.
func buildMinIOGetOptions(r *http.Request) minio.GetObjectOptions {
	opts := minio.GetObjectOptions{}
	if rangeHeader := r.Header.Get("Range"); rangeHeader != "" {
		// format: "bytes=start-end"
		rangeHeader = strings.TrimPrefix(rangeHeader, "bytes=")
		parts := strings.SplitN(rangeHeader, "-", 2)
		if len(parts) == 2 {
			start, errS := strconv.ParseInt(parts[0], 10, 64)
			end, errE := strconv.ParseInt(parts[1], 10, 64)
			if errS == nil && errE == nil {
				_ = opts.SetRange(start, end)
			} else if errS == nil && parts[1] == "" {
				_ = opts.SetRange(start, 0)
			}
		}
	}
	return opts
}

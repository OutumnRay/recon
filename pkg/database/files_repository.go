package database

import (
	"encoding/json"
	"fmt"
	"time"

	"Recontext.online/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─── Presigned-upload flow ────────────────────────────────────────────────────

// CreateUploadedFileV2 creates a file record for the presigned-URL upload flow.
// Status starts as "pending"; the client will PUT the file directly to MinIO and
// then call ConfirmUpload to transition to "queued".
func (db *DB) CreateUploadedFileV2(f *UploadedFile) error {
	return db.DB.Create(f).Error
}

// GetUploadedFileV2 fetches a file record by ID without soft-delete filter.
func (db *DB) GetUploadedFileV2(id uuid.UUID) (*UploadedFile, error) {
	var f UploadedFile
	if err := db.DB.Where("id = ?", id).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

// ConfirmUpload transitions a file from "pending" to "queued" and records the ETag.
func (db *DB) ConfirmUpload(id uuid.UUID, etag string) error {
	return db.DB.Model(&UploadedFile{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]interface{}{
			"status":     "queued",
			"etag":       etag,
			"updated_at": time.Now(),
		}).Error
}

// UpdateFileProgress updates status, progress percentage, and current stage.
func (db *DB) UpdateFileProgress(id uuid.UUID, status, stage string, progress int) error {
	return db.DB.Model(&UploadedFile{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]interface{}{
			"status":     status,
			"stage":      stage,
			"progress":   progress,
			"updated_at": time.Now(),
		}).Error
}

// SetFileCompleted marks the file as completed, records duration.
func (db *DB) SetFileCompleted(id uuid.UUID, duration float64) error {
	now := time.Now()
	return db.DB.Model(&UploadedFile{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]interface{}{
			"status":       "completed",
			"stage":        "",
			"progress":     100,
			"duration":     duration,
			"processed_at": &now,
			"updated_at":   now,
		}).Error
}

// SetFileFailed marks the file as failed with an error message.
func (db *DB) SetFileFailed(id uuid.UUID, errMsg string) error {
	return db.DB.Model(&UploadedFile{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": errMsg,
			"updated_at":    time.Now(),
		}).Error
}

// ListUploadedFilesV2 returns a paginated list of files for a user with optional
// status and title-search filters.
func (db *DB) ListUploadedFilesV2(userID uuid.UUID, page, pageSize int, status, search string) ([]UploadedFile, int64, error) {
	var total int64
	q := db.DB.Model(&UploadedFile{}).Where("user_id = ? AND deleted_at IS NULL", userID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if search != "" {
		q = q.Where("title ILIKE ? OR original_name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var files []UploadedFile
	offset := (page - 1) * pageSize
	if err := q.Order("uploaded_at DESC").Limit(pageSize).Offset(offset).Find(&files).Error; err != nil {
		return nil, 0, err
	}
	return files, total, nil
}

// ─── Transcription phrases ────────────────────────────────────────────────────

// BulkInsertPhrases inserts a batch of phrases in one statement.
func (db *DB) BulkInsertPhrases(phrases []FileTranscriptionPhrase) error {
	if len(phrases) == 0 {
		return nil
	}
	return db.DB.CreateInBatches(phrases, 500).Error
}

// GetFilePhrasesPage returns a paginated slice of phrases for a file.
// speaker filter is optional (empty string = all speakers).
func (db *DB) GetFilePhrasesPage(fileID uuid.UUID, page, pageSize int, speaker string) ([]FileTranscriptionPhrase, int64, error) {
	var total int64
	q := db.DB.Model(&FileTranscriptionPhrase{}).Where("file_id = ?", fileID)
	if speaker != "" {
		q = q.Where("speaker = ?", speaker)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var phrases []FileTranscriptionPhrase
	offset := (page - 1) * pageSize
	if err := q.Order("phrase_index ASC").Limit(pageSize).Offset(offset).Find(&phrases).Error; err != nil {
		return nil, 0, err
	}
	return phrases, total, nil
}

// GetFileSpeakers returns the distinct speaker identifiers for a file.
func (db *DB) GetFileSpeakers(fileID uuid.UUID) ([]string, error) {
	var speakers []string
	err := db.DB.Model(&FileTranscriptionPhrase{}).
		Where("file_id = ? AND speaker <> ''", fileID).
		Distinct("speaker").
		Order("speaker").
		Pluck("speaker", &speakers).Error
	return speakers, err
}

// DeleteFilePhrases removes all phrases for a file (used before re-transcription).
func (db *DB) DeleteFilePhrases(fileID uuid.UUID) error {
	return db.DB.Where("file_id = ?", fileID).Delete(&FileTranscriptionPhrase{}).Error
}

// ─── File summaries ───────────────────────────────────────────────────────────

// UpsertFileSummary creates or replaces the summary for a file.
func (db *DB) UpsertFileSummary(s *FileSummary) error {
	return db.DB.
		Where(FileSummary{FileID: s.FileID}).
		Assign(FileSummary{
			Summary:      s.Summary,
			SummaryRu:    s.SummaryRu,
			KeyTopics:    s.KeyTopics,
			ActionItems:  s.ActionItems,
			Status:       s.Status,
			ErrorMessage: s.ErrorMessage,
			UpdatedAt:    time.Now(),
		}).
		FirstOrCreate(s).Error
}

// UpdateFileSummaryStatus updates only the status (and optional error) of a summary.
func (db *DB) UpdateFileSummaryStatus(fileID uuid.UUID, status string, errMsg *string) error {
	updates := map[string]interface{}{"status": status, "updated_at": time.Now()}
	if errMsg != nil {
		updates["error_message"] = *errMsg
	}
	return db.DB.Model(&FileSummary{}).Where("file_id = ?", fileID).Updates(updates).Error
}

// GetFileSummary returns the summary for a file (nil, nil if not yet generated).
func (db *DB) GetFileSummary(fileID uuid.UUID) (*FileSummary, error) {
	var s FileSummary
	err := db.DB.Where("file_id = ?", fileID).First(&s).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &s, err
}

// CreateUploadedFile creates a new file upload record
func (db *DB) CreateUploadedFile(file *models.UploadedFile) error {
	metadata, err := json.Marshal(file.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	dbFile := &UploadedFile{
		ID:           file.ID,
		Filename:     file.Filename,
		OriginalName: file.OriginalName,
		FileSize:     file.FileSize,
		MimeType:     file.MimeType,
		StoragePath:  file.StoragePath,
		UserID:       file.UserID,
		GroupID:      file.GroupID,
		Status:       string(file.Status),
		Metadata:     string(metadata),
		UploadedAt:   file.UploadedAt,
	}

	if file.TranscriptionID != nil {
		dbFile.TranscriptionID = file.TranscriptionID
	}

	if err := db.DB.Create(dbFile).Error; err != nil {
		return fmt.Errorf("failed to create uploaded file: %w", err)
	}

	return nil
}

// GetUploadedFileByID retrieves a file by its ID
func (db *DB) GetUploadedFileByID(id string) (*models.UploadedFile, error) {
	uuidID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	var dbFile UploadedFile
	err = db.DB.Where("id = ?", uuidID).First(&dbFile).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("failed to get uploaded file: %w", err)
	}

	file := &models.UploadedFile{
		ID:           dbFile.ID,
		Filename:     dbFile.Filename,
		OriginalName: dbFile.OriginalName,
		FileSize:     dbFile.FileSize,
		MimeType:     dbFile.MimeType,
		StoragePath:  dbFile.StoragePath,
		UserID:       dbFile.UserID,
		GroupID:      dbFile.GroupID,
		Status:       models.TranscriptionStatus(dbFile.Status),
		UploadedAt:   dbFile.UploadedAt,
		ProcessedAt:  dbFile.ProcessedAt,
	}

	if dbFile.TranscriptionID != nil {
		file.TranscriptionID = dbFile.TranscriptionID
	}

	if dbFile.Metadata != "" {
		if err := json.Unmarshal([]byte(dbFile.Metadata), &file.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return file, nil
}

// ListUploadedFilesByUser retrieves all files uploaded by a specific user
func (db *DB) ListUploadedFilesByUser(userID uuid.UUID, page, pageSize int) ([]models.UploadedFile, int, error) {
	var total int64
	var dbFiles []UploadedFile

	// Get total count
	if err := db.DB.Model(&UploadedFile{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count files: %w", err)
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := db.DB.Where("user_id = ?", userID).
		Order("uploaded_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&dbFiles).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list files: %w", err)
	}

	files := make([]models.UploadedFile, 0, len(dbFiles))
	for _, dbFile := range dbFiles {
		file := models.UploadedFile{
			ID:           dbFile.ID,
			Filename:     dbFile.Filename,
			OriginalName: dbFile.OriginalName,
			FileSize:     dbFile.FileSize,
			MimeType:     dbFile.MimeType,
			StoragePath:  dbFile.StoragePath,
			UserID:       dbFile.UserID,
			GroupID:      dbFile.GroupID,
			Status:       models.TranscriptionStatus(dbFile.Status),
			UploadedAt:   dbFile.UploadedAt,
			ProcessedAt:  dbFile.ProcessedAt,
		}

		if dbFile.TranscriptionID != nil {
			file.TranscriptionID = dbFile.TranscriptionID
		}

		if dbFile.Metadata != "" {
			json.Unmarshal([]byte(dbFile.Metadata), &file.Metadata)
		}

		files = append(files, file)
	}

	return files, int(total), nil
}

// UpdateFileStatus updates the status of a file
func (db *DB) UpdateFileStatus(fileID string, status models.TranscriptionStatus) error {
	uuidID, err := uuid.Parse(fileID)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       string(status),
		"processed_at": &now,
	}

	result := db.DB.Model(&UploadedFile{}).Where("id = ?", uuidID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update file status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("file not found")
	}

	return nil
}

// DeleteUploadedFile deletes a file record
func (db *DB) DeleteUploadedFile(fileID string) error {
	uuidID, err := uuid.Parse(fileID)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	result := db.DB.Where("id = ?", uuidID).Delete(&UploadedFile{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete file: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("file not found")
	}

	return nil
}

// CheckUserHasFilePermission checks if a user has permission to access files
func (db *DB) CheckUserHasFilePermission(userID uuid.UUID, action string) (bool, error) {
	_, groupID, err := db.getUserGroupWithPermission(userID, action)
	if err != nil {
		return false, err
	}
	return groupID != uuid.Nil, nil
}

// GetUserGroupIDWithPermission returns the group ID of the first group the user belongs to
// that has the specified file action permission.
func (db *DB) GetUserGroupIDWithPermission(userID uuid.UUID, action string) (uuid.UUID, error) {
	_, groupID, err := db.getUserGroupWithPermission(userID, action)
	return groupID, err
}

func (db *DB) getUserGroupWithPermission(userID uuid.UUID, action string) ([]Group, uuid.UUID, error) {
	var groups []Group
	err := db.DB.Table("groups g").
		Select("g.id, g.permissions").
		Joins("INNER JOIN group_memberships gm ON g.id = gm.group_id").
		Where("gm.user_id = ?", userID).
		Find(&groups).Error

	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("failed to check permission: %w", err)
	}

	for _, group := range groups {
		var permissions map[string]interface{}
		if err := json.Unmarshal([]byte(group.Permissions), &permissions); err != nil {
			continue
		}

		if filesPerms, ok := permissions["files"].(map[string]interface{}); ok {
			if actions, ok := filesPerms["actions"].([]interface{}); ok {
				for _, a := range actions {
					if a == action {
						return groups, group.ID, nil
					}
				}
			}
		}
	}

	return groups, uuid.Nil, nil
}

package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
	"Recontext.online/pkg/database"
	"github.com/google/uuid"
)

// ─── POST /api/v1/files/{id}/share ───────────────────────────────────────────

// ShareFile godoc
// @Summary Поделиться файлом
// @Description Предоставляет доступ к файлу конкретному пользователю или всей организации.
// @Description Только владелец файла может выполнить эту операцию.
// @Description Если shared_with_id не указан — доступ получает вся организация владельца.
// @Tags Files
// @Accept json
// @Produce json
// @Param id path string true "ID файла"
// @Param request body models.FileShareRequest true "Параметры доступа"
// @Success 201 {object} models.FileShareInfo
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse "Только владелец может поделиться файлом"
// @Failure 404 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse "Такой доступ уже существует"
// @Security BearerAuth
// @Router /api/v1/files/{id}/share [post]
func (up *UserPortal) shareFileHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	fileIDStr := extractPathSegment(r.URL.Path, "/api/v1/files/", "/share")
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
	// Only the file owner (or admin) can share it
	if dbFile.UserID != claims.UserID && claims.Role != models.RoleAdmin {
		up.respondWithError(w, http.StatusForbidden, "Only the file owner can share this file", "")
		return
	}

	var req models.FileShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	permission := req.Permission
	if permission == "" {
		permission = "view"
	}
	if permission != "view" && permission != "edit" {
		up.respondWithError(w, http.StatusBadRequest, "Permission must be 'view' or 'edit'", "")
		return
	}

	// Validate target user exists and belongs to the same organization
	if req.SharedWithID != nil {
		if *req.SharedWithID == claims.UserID {
			up.respondWithError(w, http.StatusBadRequest, "Cannot share a file with yourself", "")
			return
		}

		var targetUser database.User
		if err := up.db.DB.Where("id = ? AND deleted_at IS NULL", *req.SharedWithID).First(&targetUser).Error; err != nil {
			up.respondWithError(w, http.StatusNotFound, "Target user not found", "")
			return
		}

		// Enforce organization isolation: target must be in the same org as the file owner
		if dbFile.OrganizationID != nil {
			if targetUser.OrganizationID == nil || *targetUser.OrganizationID != *dbFile.OrganizationID {
				up.respondWithError(w, http.StatusForbidden, "Cannot share file with a user from a different organization", "")
				return
			}
		}
	}

	// Check for duplicate share
	exists, err := up.db.FileShareExists(fileID, req.SharedWithID)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to check existing shares", err.Error())
		return
	}
	if exists {
		up.respondWithError(w, http.StatusConflict, "This share already exists", "")
		return
	}

	share := &database.FileShare{
		FileID:       fileID,
		SharedByID:   claims.UserID,
		SharedWithID: req.SharedWithID,
		Permission:   permission,
		CreatedAt:    time.Now(),
	}
	if err := up.db.CreateFileShare(share); err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to create share", err.Error())
		return
	}

	resp := models.FileShareInfo{
		ID:           share.ID,
		FileID:       share.FileID,
		SharedByID:   share.SharedByID,
		SharedWithID: share.SharedWithID,
		Permission:   share.Permission,
		CreatedAt:    share.CreatedAt,
	}

	up.logger.Infof("[Files/Share] File %s shared by %s with %v (permission=%s)",
		fileID, claims.UserID, req.SharedWithID, permission)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// ─── GET /api/v1/files/{id}/shares ───────────────────────────────────────────

// ListFileShares godoc
// @Summary Список доступов к файлу
// @Description Возвращает всех пользователей и группы, которым предоставлен доступ к файлу.
// @Description Только владелец файла или администратор может просматривать этот список.
// @Tags Files
// @Produce json
// @Param id path string true "ID файла"
// @Success 200 {object} models.FileShareListResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/files/{id}/shares [get]
func (up *UserPortal) listFileSharesHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	fileIDStr := extractPathSegment(r.URL.Path, "/api/v1/files/", "/shares")
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
		up.respondWithError(w, http.StatusForbidden, "Only the file owner can view shares", "")
		return
	}

	shares, err := up.db.ListFileShares(fileID)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to list shares", err.Error())
		return
	}

	// Enrich with user info for explicit per-user shares
	userIDs := make([]uuid.UUID, 0)
	for _, s := range shares {
		if s.SharedWithID != nil {
			userIDs = append(userIDs, *s.SharedWithID)
		}
	}

	userMap := make(map[uuid.UUID]database.User)
	if len(userIDs) > 0 {
		var users []database.User
		if err := up.db.DB.Where("id IN ?", userIDs).Find(&users).Error; err == nil {
			for _, u := range users {
				userMap[u.ID] = u
			}
		}
	}

	infos := make([]models.FileShareInfo, 0, len(shares))
	for _, s := range shares {
		info := models.FileShareInfo{
			ID:           s.ID,
			FileID:       s.FileID,
			SharedByID:   s.SharedByID,
			SharedWithID: s.SharedWithID,
			Permission:   s.Permission,
			CreatedAt:    s.CreatedAt,
		}
		if s.SharedWithID != nil {
			if u, ok := userMap[*s.SharedWithID]; ok {
				info.SharedWithUsername = &u.Username
				info.SharedWithEmail = &u.Email
			}
		}
		infos = append(infos, info)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.FileShareListResponse{
		FileID: fileID,
		Shares: infos,
	})
}

// ─── DELETE /api/v1/files/{id}/shares/{shareId} ──────────────────────────────

// DeleteFileShare godoc
// @Summary Отозвать доступ к файлу
// @Description Удаляет запись о доступе. Только владелец файла или администратор может выполнить эту операцию.
// @Tags Files
// @Produce json
// @Param id      path string true "ID файла"
// @Param shareId path string true "ID записи о доступе"
// @Success 204 "Доступ отозван"
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/files/{id}/shares/{shareId} [delete]
func (up *UserPortal) deleteFileShareHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Path: /api/v1/files/{id}/shares/{shareId}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/files/"), "/")
	if len(parts) < 3 {
		up.respondWithError(w, http.StatusBadRequest, "Invalid path", "")
		return
	}
	fileID, err := uuid.Parse(parts[0])
	if err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid file ID", err.Error())
		return
	}
	shareID, err := uuid.Parse(parts[2])
	if err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid share ID", err.Error())
		return
	}

	dbFile, err := up.db.GetUploadedFileV2(fileID)
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "File not found", err.Error())
		return
	}
	if dbFile.UserID != claims.UserID && claims.Role != models.RoleAdmin {
		up.respondWithError(w, http.StatusForbidden, "Only the file owner can revoke shares", "")
		return
	}

	share, err := up.db.GetFileShare(shareID)
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "Share not found", err.Error())
		return
	}
	if share.FileID != fileID {
		up.respondWithError(w, http.StatusNotFound, "Share not found for this file", "")
		return
	}

	if err := up.db.DeleteFileShare(shareID); err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to delete share", err.Error())
		return
	}

	up.logger.Infof("[Files/Share] Share %s revoked by %s (file %s)", shareID, claims.UserID, fileID)
	w.WriteHeader(http.StatusNoContent)
}

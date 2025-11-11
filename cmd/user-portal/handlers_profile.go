package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
)

const (
	maxAvatarSize = 5 * 1024 * 1024 // 5MB
	avatarsDir    = "uploads/avatars"
)

// UploadAvatar godoc
// @Summary Upload user avatar
// @Description Upload a user avatar image
// @Tags Profile
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "Avatar image file"
// @Success 200 {object} models.UploadAvatarResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/users/{id}/avatar [post]
func (up *UserPortal) uploadAvatarHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Extract user ID from path
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/users/"), "/")
	if len(pathParts) < 2 {
		up.respondWithError(w, http.StatusBadRequest, "Invalid URL", "")
		return
	}
	userID := pathParts[0]

	// Users can only update their own avatar (unless admin)
	if userID != claims.UserID && claims.Role != models.RoleAdmin {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "You can only update your own avatar")
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(maxAvatarSize); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Failed to parse form", err.Error())
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Avatar file is required", err.Error())
		return
	}
	defer file.Close()

	// Check file size
	if header.Size > maxAvatarSize {
		up.respondWithError(w, http.StatusBadRequest, "File too large", fmt.Sprintf("Maximum file size is %d MB", maxAvatarSize/(1024*1024)))
		return
	}

	// Check file type
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		up.respondWithError(w, http.StatusBadRequest, "Invalid file type", "Only image files are allowed")
		return
	}

	// Create avatars directory if it doesn't exist
	if err := os.MkdirAll(avatarsDir, 0755); err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to create directory", err.Error())
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		// Try to determine extension from content type
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		default:
			ext = ".jpg"
		}
	}

	// Create unique filename using timestamp and hash
	hash := md5.New()
	hash.Write([]byte(fmt.Sprintf("%s-%d", userID, time.Now().UnixNano())))
	filename := fmt.Sprintf("%s%s", hex.EncodeToString(hash.Sum(nil)), ext)
	filePath := filepath.Join(avatarsDir, filename)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to save file", err.Error())
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to save file", err.Error())
		return
	}

	// Generate avatar URL
	avatarURL := fmt.Sprintf("/uploads/avatars/%s", filename)

	// Update user avatar in database
	if err := up.userRepo.UpdateAvatar(userID, avatarURL); err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to update user avatar", err.Error())
		return
	}

	up.logger.Infof("Avatar uploaded for user %s: %s", userID, avatarURL)

	response := models.UploadAvatarResponse{
		AvatarURL: avatarURL,
		Message:   "Avatar uploaded successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update user profile information (self-service)
// @Tags Profile
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body models.UpdateProfileRequest true "Profile data"
// @Success 200 {object} models.UserInfo
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/users/{id} [put]
func (up *UserPortal) updateProfileHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Extract user ID from path
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/users/"), "/")
	if len(pathParts) == 0 {
		up.respondWithError(w, http.StatusBadRequest, "Invalid URL", "")
		return
	}
	userID := pathParts[0]

	// Users can only update their own profile (unless admin)
	if userID != claims.UserID && claims.Role != models.RoleAdmin {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "You can only update your own profile")
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get current user
	user, err := up.userRepo.GetByID(userID)
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "User not found", err.Error())
		return
	}

	// Update fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Language != "" {
		user.Language = req.Language
	}

	user.UpdatedAt = time.Now()

	// Update in database
	if err := up.userRepo.Update(user); err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to update profile", err.Error())
		return
	}

	up.logger.Infof("Profile updated for user %s", userID)

	// Return updated user info
	userInfo := models.UserInfo{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Role:         user.Role,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Phone:        user.Phone,
		Bio:          user.Bio,
		Avatar:       user.Avatar,
		DepartmentID: user.DepartmentID,
		Permissions:  user.Permissions,
		Language:     user.Language,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get user profile information
// @Tags Profile
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} models.UserInfo
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/users/{id} [get]
func (up *UserPortal) getProfileHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Extract user ID from path
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/users/"), "/")
	if len(pathParts) == 0 {
		up.respondWithError(w, http.StatusBadRequest, "Invalid URL", "")
		return
	}
	userID := pathParts[0]

	// Users can view their own profile or admins can view any
	if userID != claims.UserID && claims.Role != models.RoleAdmin {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "You can only view your own profile")
		return
	}

	// Get user
	user, err := up.userRepo.GetByID(userID)
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "User not found", err.Error())
		return
	}

	// Return user info
	userInfo := models.UserInfo{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Role:         user.Role,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Phone:        user.Phone,
		Bio:          user.Bio,
		Avatar:       user.Avatar,
		DepartmentID: user.DepartmentID,
		Permissions:  user.Permissions,
		Language:     user.Language,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

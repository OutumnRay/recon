package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
	"github.com/google/uuid"
)

// UpdateProfile godoc
// @Summary Обновить профиль пользователя
// @Description Обновить информацию профиля пользователя (самообслуживание)
// @Tags Profile
// @Accept json
// @Produce json
// @Param id path string true "Идентификатор пользователя"
// @Param request body models.UpdateProfileRequest true "Данные профиля"
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
	if userID != claims.UserID.String() && claims.Role != models.RoleAdmin {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "You can only update your own profile")
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get current user
	user, err := up.userRepo.GetByID(uuid.Must(uuid.Parse(userID)))
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "User not found", err.Error())
		return
	}

	// Update fields (allow empty values to clear fields)
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Phone = req.Phone
	user.Bio = req.Bio

	// Only update avatar if provided (don't clear it)
	if req.AvatarURL != "" {
		user.Avatar = req.AvatarURL
	}

	// Only update language if provided (don't clear it)
	if req.Language != "" {
		user.Language = req.Language
	}

	// Validate and update notification preferences
	if req.NotificationPreferences != "" {
		if req.NotificationPreferences == "tracks" || req.NotificationPreferences == "rooms" || req.NotificationPreferences == "both" {
			user.NotificationPreferences = req.NotificationPreferences
		}
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
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Bio:       user.Bio,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

// GetProfile godoc
// @Summary Получить профиль пользователя
// @Description Получить информацию о профиле пользователя
// @Tags Profile
// @Produce json
// @Param id path string true "Идентификатор пользователя"
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
	if userID != claims.UserID.String() && claims.Role != models.RoleAdmin {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "You can only view your own profile")
		return
	}

	// Get user
	user, err := up.userRepo.GetByID(uuid.Must(uuid.Parse(userID)))
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "User not found", err.Error())
		return
	}

	// Return user info
	userInfo := models.UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Bio:       user.Bio,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

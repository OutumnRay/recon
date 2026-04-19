package main

// MinimalUserInfoHandler for /profile and /update-profile

import (
	"encoding/json"
	"net/http"
	"time"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
)

// GetMyProfile godoc
// @Summary Получить текущий профиль
// @Description Получить данные текущего пользователя (минимальные)
// @Tags Profile
// @Produce json
// @Success 200 {object} models.MinimalUserInfo
// @Failure 401 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/profile [get]
func (up *UserPortal) getMyProfileHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	user, err := up.userRepo.GetByID(claims.UserID)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to fetch profile", err.Error())
		return
	}

	userInfo := models.MinimalUserInfo{
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

// UpdateMyProfile godoc
// @Summary Обновить текущий профиль
// @Description Обновить bio, имя и другие поля текущего пользователя
// @Tags Profile
// @Accept json
// @Produce json
// @Param request body models.UpdateProfileRequest true "Обновленные данные профиля"
// @Success 200 {object} models.MinimalUserInfo
// @Failure 400 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/update-profile [put]
func (up *UserPortal) updateMyProfileHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	user, err := up.userRepo.GetByID(claims.UserID)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to fetch profile", err.Error())
		return
	}

	// Update allowed fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	user.Bio = req.Bio // allow empty to clear

	user.UpdatedAt = time.Now()
	if err := up.userRepo.Update(user); err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to update profile", err.Error())
		return
	}

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

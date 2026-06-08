package main

// MinimalUserInfoHandler for /profile and /update-profile

import (
	"encoding/json"
	"net/http"
	"time"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
	"github.com/google/uuid"
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
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		Role:           user.Role,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Bio:            user.Bio,
		Groups:         user.Groups,
		OrganizationID: user.OrganizationID,
		DepartmentID:   user.DepartmentID,
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
	if req.AvatarURL != "" {
		user.Avatar = req.AvatarURL
	}

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

// DepartmentMemberItem - публичный профиль коллеги для выбора участников встречи
type DepartmentMemberItem struct {
	ID           uuid.UUID       `json:"id"`
	Username     string          `json:"username"`
	Email        string          `json:"email"`
	FirstName    string          `json:"first_name"`
	LastName     string          `json:"last_name"`
	Avatar       string          `json:"avatar,omitempty"`
	Role         models.UserRole `json:"role"`
	DepartmentID *uuid.UUID      `json:"department_id,omitempty"`
}

// GetDepartmentMembers godoc
// @Summary Список коллег по отделу
// @Description Возвращает активных пользователей того же отдела и организации, что и текущий пользователь
// @Tags Profile
// @Produce json
// @Success 200 {array} DepartmentMemberItem
// @Failure 401 {object} models.ErrorResponse
// @Failure 400 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/users/colleagues [get]
func (up *UserPortal) getDepartmentMembersHandler(w http.ResponseWriter, r *http.Request) {
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

	if user.DepartmentID == nil || user.OrganizationID == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]DepartmentMemberItem{})
		return
	}

	members, err := up.userRepo.ListDepartmentMembers(*user.DepartmentID, *user.OrganizationID)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to fetch department members", err.Error())
		return
	}

	result := make([]DepartmentMemberItem, 0, len(members))
	for _, m := range members {
		result = append(result, DepartmentMemberItem{
			ID:           m.ID,
			Username:     m.Username,
			Email:        m.Email,
			FirstName:    m.FirstName,
			LastName:     m.LastName,
			Avatar:       m.Avatar,
			Role:         m.Role,
			DepartmentID: m.DepartmentID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

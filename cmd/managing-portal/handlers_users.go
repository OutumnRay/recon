package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
)

// ListUsers godoc
// @Summary Список всех пользователей
// @Description Получить постраничный список пользователей с дополнительными фильтрами
// @Tags Users
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param role query string false "Filter by role"
// @Param is_active query bool false "Filter by active status"
// @Success 200 {object} models.ListUsersResponse
// @Security BearerAuth
// @Router /api/v1/users [get]
func (mp *ManagingPortal) listUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Get optional filters from query params
	role := r.URL.Query().Get("role")
	isActiveStr := r.URL.Query().Get("is_active")

	var isActive *bool
	if isActiveStr != "" {
		val := isActiveStr == "true"
		isActive = &val
	}

	users, err := mp.userRepo.List(role, isActive, nil) // nil organization = all organizations
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to list users", err.Error())
		return
	}

	var usersList []models.UserInfo
	for _, user := range users {
		usersList = append(usersList, models.UserInfo{
			ID:             user.ID,
			Username:       user.Username,
			Email:          user.Email,
			Role:           user.Role,
			OrganizationID: user.OrganizationID,
			DepartmentID:   user.DepartmentID,
			Permissions:    user.Permissions,
			Language:     user.Language,
		})
	}

	response := models.ListUsersResponse{
		Items:    usersList,
		Total:    len(usersList),
		Offset:   0,
		PageSize: len(usersList),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUser godoc
// @Summary Получить пользователя по ID
// @Description Получить детальную информацию о конкретном пользователе
// @Tags Users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} models.UserInfo
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/users/{id} [get]
func (mp *ManagingPortal) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")

	user, err := mp.userRepo.GetByID(uuid.Must(uuid.Parse(userID)))
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "User not found", err.Error())
		return
	}

	userInfo := models.UserInfo{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		Role:           user.Role,
		OrganizationID: user.OrganizationID,
		DepartmentID:   user.DepartmentID,
		Permissions:    user.Permissions,
		Language:       user.Language,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

// UpdateUser godoc
// @Summary Обновить пользователя
// @Description Обновить информацию о пользователе (только администратор)
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body models.UpdateUserRequest true "Update data"
// @Success 200 {object} models.UserInfo
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/users/{id} [put]
func (mp *ManagingPortal) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Find user
	foundUser, err := mp.userRepo.GetByID(uuid.Must(uuid.Parse(userID)))
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "User not found", err.Error())
		return
	}

	// Update fields
	if req.Email != "" {
		foundUser.Email = req.Email
	}
	if req.Password != "" {
		// Hash the new password
		mp.logger.Infof("=== PASSWORD UPDATE DEBUG ===")
		mp.logger.Infof("User: %s (%s)", foundUser.Username, foundUser.ID)
		mp.logger.Infof("Old password hash (first 20 chars): %s...", foundUser.Password[:20])
		mp.logger.Infof("New password (plain): %s", req.Password)
		hashedPassword := auth.HashPassword(req.Password)
		mp.logger.Infof("New password hash (first 20 chars): %s...", hashedPassword[:20])
		foundUser.Password = hashedPassword
		mp.logger.Infof("Password field set in user object")
		mp.logger.Infof("============================")
	}
	if req.Role != "" {
		foundUser.Role = req.Role
	}
	if req.OrganizationID != nil {
		foundUser.OrganizationID = req.OrganizationID
	}
	if req.DepartmentID != nil {
		foundUser.DepartmentID = req.DepartmentID
	}
	if req.Groups != nil {
		foundUser.Groups = req.Groups
	}
	if req.Permissions != nil {
		foundUser.Permissions = *req.Permissions
	}
	if req.Language != "" {
		foundUser.Language = req.Language
	}
	if req.IsActive != nil {
		foundUser.IsActive = *req.IsActive
	}

	mp.logger.Infof("Calling userRepo.Update() for user: %s", foundUser.ID)
	if err := mp.userRepo.Update(foundUser); err != nil {
		mp.logger.Errorf("Failed to update user in database: %v", err)
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to update user", err.Error())
		return
	}

	mp.logger.Infof("✓ User updated successfully in database: %s (%s)", foundUser.Username, foundUser.ID)
	if req.Password != "" {
		mp.logger.Infof("✓ Password was changed and saved to database")
	}

	userInfo := models.UserInfo{
		ID:             foundUser.ID,
		Username:       foundUser.Username,
		Email:          foundUser.Email,
		Role:           foundUser.Role,
		OrganizationID: foundUser.OrganizationID,
		DepartmentID:   foundUser.DepartmentID,
		Permissions:    foundUser.Permissions,
		Language:       foundUser.Language,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

// DeleteUser godoc
// @Summary Удалить пользователя
// @Description Удалить пользователя из системы (только администратор)
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/users/{id} [delete]
func (mp *ManagingPortal) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")

	// Get user first to log the username
	parsedUserID := uuid.Must(uuid.Parse(userID))
	user, err := mp.userRepo.GetByID(parsedUserID)
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "User not found", err.Error())
		return
	}

	// Delete user
	if err := mp.userRepo.Delete(parsedUserID); err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to delete user", err.Error())
		return
	}

	mp.logger.Infof("User deleted: %s (%s)", user.Username, user.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User deleted successfully",
		"user_id": userID,
	})
}

// ChangePassword godoc
// @Summary Изменить пароль пользователя
// @Description Изменить пароль для аутентифицированного пользователя
// @Tags Users
// @Accept json
// @Produce json
// @Param request body models.ChangePasswordRequest true "Password change request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/users/password [put]
func (mp *ManagingPortal) changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		mp.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Find user
	user, err := mp.userRepo.GetByUsername(claims.Username)
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "User not found", err.Error())
		return
	}

	// Verify old password
	if !auth.VerifyPassword(req.OldPassword, user.Password) {
		mp.respondWithError(w, http.StatusUnauthorized, "Invalid old password", "")
		return
	}

	// Update password
	hashedPassword := auth.HashPassword(req.NewPassword)
	if err := mp.userRepo.UpdatePassword(user.ID, hashedPassword); err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to update password", err.Error())
		return
	}

	mp.logger.Infof("Password changed for user: %s", claims.Username)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password changed successfully",
	})
}

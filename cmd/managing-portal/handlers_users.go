package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
)

// ListUsers godoc
// @Summary List all users
// @Description Get a paginated list of users with optional filters
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

	users, err := mp.userRepo.List(role, isActive)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to list users", err.Error())
		return
	}

	var usersList []models.UserInfo
	for _, user := range users {
		usersList = append(usersList, models.UserInfo{
			ID:           user.ID,
			Username:     user.Username,
			Email:        user.Email,
			Role:         user.Role,
			DepartmentID: user.DepartmentID,
			Permissions:  user.Permissions,
		})
	}

	response := models.ListUsersResponse{
		Users:    usersList,
		Total:    len(usersList),
		Page:     1,
		PageSize: 20,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get detailed information about a specific user
// @Tags Users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} models.UserInfo
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/users/{id} [get]
func (mp *ManagingPortal) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")

	user, err := mp.userRepo.GetByID(userID)
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "User not found", err.Error())
		return
	}

	userInfo := models.UserInfo{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Role:         user.Role,
		DepartmentID: user.DepartmentID,
		Permissions:  user.Permissions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

// UpdateUser godoc
// @Summary Update user
// @Description Update user information (admin only)
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
	foundUser, err := mp.userRepo.GetByID(userID)
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "User not found", err.Error())
		return
	}

	// Update fields
	if req.Email != "" {
		foundUser.Email = req.Email
	}
	if req.Role != "" {
		foundUser.Role = req.Role
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
	if req.IsActive != nil {
		foundUser.IsActive = *req.IsActive
	}

	if err := mp.userRepo.Update(foundUser); err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to update user", err.Error())
		return
	}

	mp.logger.Infof("User updated: %s (%s)", foundUser.Username, foundUser.ID)

	userInfo := models.UserInfo{
		ID:           foundUser.ID,
		Username:     foundUser.Username,
		Email:        foundUser.Email,
		Role:         foundUser.Role,
		DepartmentID: foundUser.DepartmentID,
		Permissions:  foundUser.Permissions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete a user from the system (admin only)
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
	user, err := mp.userRepo.GetByID(userID)
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "User not found", err.Error())
		return
	}

	// Delete user
	if err := mp.userRepo.Delete(userID); err != nil {
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
// @Summary Change user password
// @Description Change password for the authenticated user
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

package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

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
	var usersList []models.UserInfo
	for _, user := range mp.users {
		if user.IsActive {
			usersList = append(usersList, models.UserInfo{
				ID:       user.ID,
				Username: user.Username,
				Email:    user.Email,
				Role:     user.Role,
			})
		}
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

	for _, user := range mp.users {
		if user.ID == userID {
			userInfo := models.UserInfo{
				ID:       user.ID,
				Username: user.Username,
				Email:    user.Email,
				Role:     user.Role,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(userInfo)
			return
		}
	}

	mp.respondWithError(w, http.StatusNotFound, "User not found", "")
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
	var foundUser *models.User
	for _, user := range mp.users {
		if user.ID == userID {
			foundUser = user
			break
		}
	}

	if foundUser == nil {
		mp.respondWithError(w, http.StatusNotFound, "User not found", "")
		return
	}

	// Update fields
	if req.Email != "" {
		foundUser.Email = req.Email
	}
	if req.Role != "" {
		foundUser.Role = req.Role
	}
	if req.Groups != nil {
		foundUser.Groups = req.Groups
	}
	if req.IsActive != nil {
		foundUser.IsActive = *req.IsActive
	}
	foundUser.UpdatedAt = time.Now()

	mp.users[foundUser.Username] = foundUser

	userInfo := models.UserInfo{
		ID:       foundUser.ID,
		Username: foundUser.Username,
		Email:    foundUser.Email,
		Role:     foundUser.Role,
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

	// Find and delete user
	for username, user := range mp.users {
		if user.ID == userID {
			delete(mp.users, username)
			mp.logger.Infof("User deleted: %s (%s)", user.Username, user.ID)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "User deleted successfully",
				"user_id": userID,
			})
			return
		}
	}

	mp.respondWithError(w, http.StatusNotFound, "User not found", "")
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
	user, exists := mp.users[claims.Username]
	if !exists {
		mp.respondWithError(w, http.StatusNotFound, "User not found", "")
		return
	}

	// Verify old password
	if !auth.VerifyPassword(req.OldPassword, user.Password) {
		mp.respondWithError(w, http.StatusUnauthorized, "Invalid old password", "")
		return
	}

	// Update password
	user.Password = auth.HashPassword(req.NewPassword)
	user.UpdatedAt = time.Now()
	mp.users[claims.Username] = user

	mp.logger.Infof("Password changed for user: %s", claims.Username)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password changed successfully",
	})
}

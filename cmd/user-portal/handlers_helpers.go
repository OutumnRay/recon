package main

import (
	"encoding/json"
	"net/http"

	"Recontext.online/pkg/auth"
)

// listUsersHandler godoc
// @Summary Список пользователей для участников встречи
// @Description Получить список пользователей для выбора участников встречи (только из той же организации)
// @Tags Helpers
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/users [get]
func (up *UserPortal) listUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Get current user to determine organization
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	user, err := up.userRepo.GetByID(claims.UserID)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get user", err.Error())
		return
	}

	// Get all active users from the same organization
	isActive := true
	users, err := up.userRepo.List("", &isActive, user.OrganizationID) // Get all active users from same organization
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to fetch users", err.Error())
		return
	}

	// Return standardized response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items":     users,
		"total":     len(users),
		"offset":    0,
		"page_size": len(users),
	})
}

// listDepartmentsHandler godoc
// @Summary Список отделов для приглашений на встречи
// @Description Получить список отделов для массового выбора участников (только из той же организации)
// @Tags Helpers
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/departments [get]
func (up *UserPortal) listDepartmentsHandler(w http.ResponseWriter, r *http.Request) {
	// Get current user to determine organization
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	user, err := up.userRepo.GetByID(claims.UserID)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get user", err.Error())
		return
	}

	// Get all active departments from the same organization
	departments, err := up.departmentRepo.List(nil, false, user.OrganizationID) // Get all departments from same organization, active only
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to fetch departments", err.Error())
		return
	}

	// Return standardized response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items":     departments,
		"total":     len(departments),
		"offset":    0,
		"page_size": len(departments),
	})
}

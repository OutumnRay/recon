package main

import (
	"encoding/json"
	"net/http"
)

// listUsersHandler godoc
// @Summary Список пользователей для участников встречи
// @Description Получить список пользователей для выбора участников встречи
// @Tags Helpers
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/users [get]
func (up *UserPortal) listUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Get all active users from database
	isActive := true
	users, err := up.userRepo.List("", &isActive) // Get all active users, any role
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
// @Description Получить список отделов для массового выбора участников
// @Tags Helpers
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/departments [get]
func (up *UserPortal) listDepartmentsHandler(w http.ResponseWriter, r *http.Request) {
	// Get all active departments from database
	departments, err := up.departmentRepo.List(nil, false) // Get all departments, active only (false = don't include inactive)
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

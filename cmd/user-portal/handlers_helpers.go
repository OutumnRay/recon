package main

import (
	"encoding/json"
	"net/http"
)

// listUsersHandler godoc
// @Summary List users for meeting participants
// @Description Get list of users for participant selection in meetings
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

	// Return simplified user list
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": users,
	})
}

// listDepartmentsHandler godoc
// @Summary List departments for meeting invitations
// @Description Get list of departments for bulk participant selection
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

	// Return departments list
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": departments,
	})
}

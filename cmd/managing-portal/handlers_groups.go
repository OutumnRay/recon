package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"Recontext.online/internal/models"
)

// ListGroups godoc
// @Summary List all groups
// @Description Get a list of all user groups
// @Tags Groups
// @Produce json
// @Success 200 {object} models.ListGroupsResponse
// @Security BearerAuth
// @Router /api/v1/groups [get]
func (mp *ManagingPortal) listGroupsHandler(w http.ResponseWriter, r *http.Request) {
	groups, err := mp.groupRepo.List()
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to list groups", err.Error())
		return
	}

	var groupsList []models.UserGroup
	for _, group := range groups {
		groupsList = append(groupsList, *group)
	}

	response := models.ListGroupsResponse{
		Items:    groupsList,
		Total:    len(groupsList),
		Offset:   0,
		PageSize: len(groupsList),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetGroup godoc
// @Summary Get group by ID
// @Description Get detailed information about a specific group
// @Tags Groups
// @Produce json
// @Param id path string true "Group ID"
// @Success 200 {object} models.UserGroup
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/groups/{id} [get]
func (mp *ManagingPortal) getGroupHandler(w http.ResponseWriter, r *http.Request) {
	groupID := strings.TrimPrefix(r.URL.Path, "/api/v1/groups/")

	group, err := mp.groupRepo.GetByID(groupID)
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "Group not found", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

// CreateGroup godoc
// @Summary Create a new group
// @Description Create a new user group with permissions (admin only)
// @Tags Groups
// @Accept json
// @Produce json
// @Param request body models.CreateGroupRequest true "Group data"
// @Success 201 {object} models.UserGroup
// @Failure 400 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/groups [post]
func (mp *ManagingPortal) createGroupHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Check if group name already exists
	if exists, _ := mp.groupRepo.NameExists(req.Name); exists {
		mp.respondWithError(w, http.StatusConflict, "Group name already exists", "")
		return
	}

	groupID := fmt.Sprintf("group-%d", time.Now().Unix())
	group := &models.UserGroup{
		ID:          groupID,
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := mp.groupRepo.Create(group); err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to create group", err.Error())
		return
	}

	mp.logger.Infof("Group created: %s (%s)", group.Name, group.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(group)
}

// UpdateGroup godoc
// @Summary Update group
// @Description Update group information and permissions (admin only)
// @Tags Groups
// @Accept json
// @Produce json
// @Param id path string true "Group ID"
// @Param request body models.UpdateGroupRequest true "Update data"
// @Success 200 {object} models.UserGroup
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/groups/{id} [put]
func (mp *ManagingPortal) updateGroupHandler(w http.ResponseWriter, r *http.Request) {
	groupID := strings.TrimPrefix(r.URL.Path, "/api/v1/groups/")

	var req models.UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	group, err := mp.groupRepo.GetByID(groupID)
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "Group not found", err.Error())
		return
	}

	// Update fields
	if req.Name != "" {
		group.Name = req.Name
	}
	if req.Description != "" {
		group.Description = req.Description
	}
	if req.Permissions != nil {
		group.Permissions = req.Permissions
	}

	if err := mp.groupRepo.Update(group); err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to update group", err.Error())
		return
	}

	mp.logger.Infof("Group updated: %s (%s)", group.Name, group.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

// DeleteGroup godoc
// @Summary Delete group
// @Description Delete a user group (admin only)
// @Tags Groups
// @Produce json
// @Param id path string true "Group ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/groups/{id} [delete]
func (mp *ManagingPortal) deleteGroupHandler(w http.ResponseWriter, r *http.Request) {
	groupID := strings.TrimPrefix(r.URL.Path, "/api/v1/groups/")

	// Get group first to log the name
	group, err := mp.groupRepo.GetByID(groupID)
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "Group not found", err.Error())
		return
	}

	if err := mp.groupRepo.Delete(groupID); err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to delete group", err.Error())
		return
	}

	mp.logger.Infof("Group deleted: %s (%s)", group.Name, group.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Group deleted successfully",
		"group_id": groupID,
	})
}

// AddUserToGroup godoc
// @Summary Add user to group
// @Description Add a user to a specific group (admin only)
// @Tags Groups
// @Accept json
// @Produce json
// @Param request body models.AddUserToGroupRequest true "User and group IDs"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/groups/add-user [post]
func (mp *ManagingPortal) addUserToGroupHandler(w http.ResponseWriter, r *http.Request) {
	var req models.AddUserToGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Check if group exists
	if _, err := mp.groupRepo.GetByID(req.GroupID); err != nil {
		mp.respondWithError(w, http.StatusNotFound, "Group not found", err.Error())
		return
	}

	// Check if user exists
	user, err := mp.userRepo.GetByID(req.UserID)
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "User not found", err.Error())
		return
	}

	// Check if already in group
	for _, gid := range user.Groups {
		if gid == req.GroupID {
			mp.respondWithError(w, http.StatusConflict, "User already in group", "")
			return
		}
	}

	// Add user to group
	if err := mp.groupRepo.AddUserToGroup(req.UserID, req.GroupID, "admin"); err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to add user to group", err.Error())
		return
	}

	mp.logger.Infof("User %s added to group %s", user.Username, req.GroupID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "User added to group successfully",
		"user_id":  req.UserID,
		"group_id": req.GroupID,
	})
}

// RemoveUserFromGroup godoc
// @Summary Remove user from group
// @Description Remove a user from a specific group (admin only)
// @Tags Groups
// @Accept json
// @Produce json
// @Param request body models.AddUserToGroupRequest true "User and group IDs"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/groups/remove-user [post]
func (mp *ManagingPortal) removeUserFromGroupHandler(w http.ResponseWriter, r *http.Request) {
	var req models.AddUserToGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Check if group exists
	if _, err := mp.groupRepo.GetByID(req.GroupID); err != nil {
		mp.respondWithError(w, http.StatusNotFound, "Group not found", err.Error())
		return
	}

	// Check if user exists
	user, err := mp.userRepo.GetByID(req.UserID)
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "User not found", err.Error())
		return
	}

	// Remove user from group
	if err := mp.groupRepo.RemoveUserFromGroup(req.UserID, req.GroupID); err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to remove user from group", err.Error())
		return
	}

	mp.logger.Infof("User %s removed from group %s", user.Username, req.GroupID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "User removed from group successfully",
		"user_id":  req.UserID,
		"group_id": req.GroupID,
	})
}

// CheckPermission godoc
// @Summary Check user permission
// @Description Check if a user has permission to perform an action on a resource
// @Tags Groups
// @Accept json
// @Produce json
// @Param request body models.PermissionCheckRequest true "Permission check data"
// @Success 200 {object} models.PermissionCheckResponse
// @Failure 400 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/groups/check-permission [post]
func (mp *ManagingPortal) checkPermissionHandler(w http.ResponseWriter, r *http.Request) {
	var req models.PermissionCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Find user
	foundUser, err := mp.userRepo.GetByID(req.UserID)
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "User not found", err.Error())
		return
	}

	// Check admin role (admins have all permissions)
	if foundUser.Role == models.RoleAdmin {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.PermissionCheckResponse{
			Allowed: true,
			Reason:  "User has admin role",
		})
		return
	}

	// Check group permissions
	allowed := false
	reason := "No permission found"

	for _, groupID := range foundUser.Groups {
		group, err := mp.groupRepo.GetByID(groupID)
		if err != nil {
			continue
		}

		// Check if group permissions contain the resource
		if resourcePerms, ok := group.Permissions[req.Resource]; ok {
			// Check if the action is allowed
			if permsMap, ok := resourcePerms.(map[string]interface{}); ok {
				if actions, ok := permsMap["actions"].([]interface{}); ok {
					for _, action := range actions {
						if actionStr, ok := action.(string); ok && actionStr == req.Action {
							allowed = true
							reason = fmt.Sprintf("Permission granted via group %s", group.Name)
							break
						}
					}
				}
			}
		}

		if allowed {
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.PermissionCheckResponse{
		Allowed: allowed,
		Reason:  reason,
	})
}

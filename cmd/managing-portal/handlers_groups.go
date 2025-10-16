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
	var groupsList []models.UserGroup
	for _, group := range mp.groups {
		groupsList = append(groupsList, *group)
	}

	response := models.ListGroupsResponse{
		Groups: groupsList,
		Total:  len(groupsList),
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

	group, exists := mp.groups[groupID]
	if !exists {
		mp.respondWithError(w, http.StatusNotFound, "Group not found", "")
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

	groupID := fmt.Sprintf("group-%d", time.Now().Unix())
	group := &models.UserGroup{
		ID:          groupID,
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mp.groups[groupID] = group
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

	group, exists := mp.groups[groupID]
	if !exists {
		mp.respondWithError(w, http.StatusNotFound, "Group not found", "")
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
	group.UpdatedAt = time.Now()

	mp.groups[groupID] = group

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

	group, exists := mp.groups[groupID]
	if !exists {
		mp.respondWithError(w, http.StatusNotFound, "Group not found", "")
		return
	}

	delete(mp.groups, groupID)
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
	if _, exists := mp.groups[req.GroupID]; !exists {
		mp.respondWithError(w, http.StatusNotFound, "Group not found", "")
		return
	}

	// Find and update user
	var foundUser *models.User
	for _, user := range mp.users {
		if user.ID == req.UserID {
			foundUser = user
			break
		}
	}

	if foundUser == nil {
		mp.respondWithError(w, http.StatusNotFound, "User not found", "")
		return
	}

	// Add group to user's groups
	if foundUser.Groups == nil {
		foundUser.Groups = []string{}
	}
	// Check if already in group
	for _, gid := range foundUser.Groups {
		if gid == req.GroupID {
			mp.respondWithError(w, http.StatusConflict, "User already in group", "")
			return
		}
	}
	foundUser.Groups = append(foundUser.Groups, req.GroupID)
	foundUser.UpdatedAt = time.Now()

	mp.users[foundUser.Username] = foundUser
	mp.logger.Infof("User %s added to group %s", foundUser.Username, req.GroupID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User added to group successfully",
		"user_id": req.UserID,
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
	var foundUser *models.User
	for _, user := range mp.users {
		if user.ID == req.UserID {
			foundUser = user
			break
		}
	}

	if foundUser == nil {
		mp.respondWithError(w, http.StatusNotFound, "User not found", "")
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
		group, exists := mp.groups[groupID]
		if !exists {
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

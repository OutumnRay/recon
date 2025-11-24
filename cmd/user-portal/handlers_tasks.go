package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
	"github.com/google/uuid"
)

// respondWithJSON sends a JSON response
func (up *UserPortal) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to marshal JSON", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// parseJSONBody is a helper function to parse JSON request body
func parseJSONBody(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(v); err != nil {
		return err
	}
	return nil
}

// @Summary Get my tasks
// @Description Get all tasks assigned to the current user
// @Tags Tasks
// @Accept json
// @Produce json
// @Param status query string false "Filter by status" Enums(pending, in_progress, completed, cancelled)
// @Success 200 {array} models.TaskWithDetails
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/my-tasks [get]
func (up *UserPortal) getMyTasksHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Get status filter if provided
	var status *string
	if statusParam := r.URL.Query().Get("status"); statusParam != "" {
		status = &statusParam
	}

	// Get user's tasks with details
	tasks, err := up.taskRepo.GetMyTasksWithDetails(claims.UserID, status)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to fetch tasks", err.Error())
		return
	}

	up.respondWithJSON(w, http.StatusOK, tasks)
}

// @Summary Update task status
// @Description Update the status of a task (for assigned user)
// @Tags Tasks
// @Accept json
// @Produce json
// @Param task_id path string true "Task ID (UUID)"
// @Param status body models.UpdateTaskStatusRequest true "New status"
// @Success 200 {object} models.TaskWithDetails
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/tasks/{task_id}/status [put]
func (up *UserPortal) updateTaskStatusHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Get task_id from URL
	taskIDStr := r.URL.Query().Get("task_id")
	if taskIDStr == "" {
		up.respondWithError(w, http.StatusBadRequest, "task_id is required", "")
		return
	}

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid task_id format", err.Error())
		return
	}

	// Parse request body
	var req models.UpdateTaskStatusRequest
	if err := parseJSONBody(r, &req); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate status
	var newStatus models.TaskStatus
	switch req.Status {
	case string(models.TaskStatusPending):
		newStatus = models.TaskStatusPending
	case string(models.TaskStatusInProgress):
		newStatus = models.TaskStatusInProgress
	case string(models.TaskStatusCompleted):
		newStatus = models.TaskStatusCompleted
	case string(models.TaskStatusCancelled):
		newStatus = models.TaskStatusCancelled
	default:
		up.respondWithError(w, http.StatusBadRequest, "Invalid status", "")
		return
	}

	// Check access - user must be assigned to the task or be admin
	hasAccess, err := up.taskRepo.CheckTaskAccess(taskID, claims.UserID, string(claims.Role))
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to check access", err.Error())
		return
	}
	if !hasAccess {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "You are not assigned to this task")
		return
	}

	// Update status
	if err := up.taskRepo.UpdateTaskStatus(taskID, newStatus); err != nil {
		if err.Error() == "record not found" {
			up.respondWithError(w, http.StatusNotFound, "Task not found", "")
			return
		}
		up.respondWithError(w, http.StatusInternalServerError, "Failed to update task status", err.Error())
		return
	}

	// Return updated task
	task, err := up.taskRepo.GetTaskWithDetails(taskID)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to fetch updated task", err.Error())
		return
	}

	up.respondWithJSON(w, http.StatusOK, task)
}

// @Summary Get tasks for a meeting
// @Description Get all tasks for a specific meeting
// @Tags Tasks
// @Accept json
// @Produce json
// @Param meeting_id path string true "Meeting ID (UUID)"
// @Success 200 {array} models.TaskWithDetails
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/meetings/{meeting_id}/tasks [get]
func (up *UserPortal) getMeetingTasksHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Get meeting_id from path: /api/v1/meetings/{meeting_id}/tasks
	pathPrefix := "/api/v1/meetings/"
	pathSuffix := "/tasks"
	path := r.URL.Path

	if !strings.HasPrefix(path, pathPrefix) || !strings.HasSuffix(path, pathSuffix) {
		up.respondWithError(w, http.StatusBadRequest, "Invalid URL format", "")
		return
	}

	meetingIDStr := strings.TrimSuffix(strings.TrimPrefix(path, pathPrefix), pathSuffix)
	if meetingIDStr == "" {
		up.respondWithError(w, http.StatusBadRequest, "meeting_id is required", "")
		return
	}

	meetingID, err := uuid.Parse(meetingIDStr)
	if err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid meeting_id format", err.Error())
		return
	}

	// Check if user has access to the meeting
	// (either participant, creator, or admin)
	meeting, err := up.meetingRepo.GetMeetingByID(meetingID)
	if err != nil {
		if err.Error() == "record not found" {
			up.respondWithError(w, http.StatusNotFound, "Meeting not found", "")
			return
		}
		up.respondWithError(w, http.StatusInternalServerError, "Failed to fetch meeting", err.Error())
		return
	}

	// Check access
	hasAccess := false
	if claims.Role == "admin" || meeting.CreatedBy == claims.UserID {
		hasAccess = true
	} else {
		// Check if user is participant
		participants, err := up.meetingRepo.GetMeetingParticipants(meetingID)
		if err == nil {
			for _, p := range participants {
				if p.UserID == claims.UserID {
					hasAccess = true
					break
				}
			}
		}
	}

	if !hasAccess {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "You don't have access to this meeting")
		return
	}

	// Get tasks
	req := &models.ListTasksRequest{
		MeetingID: &meetingID,
	}

	tasks, _, err := up.taskRepo.ListTasksWithDetails(req)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to fetch tasks", err.Error())
		return
	}

	up.respondWithJSON(w, http.StatusOK, tasks)
}
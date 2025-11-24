package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
	"github.com/google/uuid"
)

// respondWithJSON sends a JSON response
func (mp *ManagingPortal) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to marshal JSON", err.Error())
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
		return fmt.Errorf("failed to decode JSON: %w", err)
	}
	return nil
}

// @Summary List tasks for a session
// @Description Get all tasks for a specific session with optional filters
// @Tags Tasks
// @Accept json
// @Produce json
// @Param session_id path string true "Session ID (UUID)"
// @Param status query string false "Filter by status" Enums(pending, in_progress, completed, cancelled)
// @Param priority query string false "Filter by priority" Enums(low, medium, high, urgent)
// @Param assigned_to query string false "Filter by assigned user ID (UUID)"
// @Param page_size query int false "Page size" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} models.ListTasksResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/sessions/{session_id}/tasks [get]
func (mp *ManagingPortal) listSessionTasksHandler(w http.ResponseWriter, r *http.Request) {
	// Get session_id from URL
	sessionIDStr := r.URL.Query().Get("session_id")
	if sessionIDStr == "" {
		mp.respondWithError(w, http.StatusBadRequest, "session_id is required", "")
		return
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid session_id format", err.Error())
		return
	}

	// Parse query parameters
	req := &models.ListTasksRequest{
		SessionID: &sessionID,
		PageSize:  20,
		Offset:    0,
	}

	if status := r.URL.Query().Get("status"); status != "" {
		req.Status = &status
	}
	if priority := r.URL.Query().Get("priority"); priority != "" {
		req.Priority = &priority
	}
	if assignedTo := r.URL.Query().Get("assigned_to"); assignedTo != "" {
		assignedToUUID, err := uuid.Parse(assignedTo)
		if err == nil {
			req.AssignedTo = &assignedToUUID
		}
	}

	// Parse pagination
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		var pageSize int
		if _, err := fmt.Sscanf(pageSizeStr, "%d", &pageSize); err == nil && pageSize > 0 {
			req.PageSize = pageSize
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		var offset int
		if _, err := fmt.Sscanf(offsetStr, "%d", &offset); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	// Get tasks with details
	tasks, total, err := mp.taskRepo.ListTasksWithDetails(req)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to fetch tasks", err.Error())
		return
	}

	response := models.ListTasksResponse{
		Items:    tasks,
		Total:    total,
		PageSize: req.PageSize,
		Offset:   req.Offset,
	}

	mp.respondWithJSON(w, http.StatusOK, response)
}

// @Summary Create a task
// @Description Create a new task for a session
// @Tags Tasks
// @Accept json
// @Produce json
// @Param session_id path string true "Session ID (UUID)"
// @Param task body models.CreateTaskRequest true "Task details"
// @Success 201 {object} models.TaskWithDetails
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/sessions/{session_id}/tasks [post]
func (mp *ManagingPortal) createTaskHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		mp.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Get session_id from URL
	sessionIDStr := r.URL.Query().Get("session_id")
	if sessionIDStr == "" {
		mp.respondWithError(w, http.StatusBadRequest, "session_id is required", "")
		return
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid session_id format", err.Error())
		return
	}

	// Parse request body
	var req models.CreateTaskRequest
	if err := parseJSONBody(r, &req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate priority
	priority := models.TaskPriorityMedium
	if req.Priority != "" {
		switch req.Priority {
		case string(models.TaskPriorityLow), string(models.TaskPriorityMedium),
			string(models.TaskPriorityHigh), string(models.TaskPriorityUrgent):
			priority = models.TaskPriority(req.Priority)
		default:
			mp.respondWithError(w, http.StatusBadRequest, "Invalid priority", "")
			return
		}
	}

	// Create task
	task := &models.Task{
		SessionID:     sessionID,
		Title:         req.Title,
		Description:   req.Description,
		Hint:          req.Hint,
		AssignedTo:    req.AssignedTo,
		AssignedBy:    &claims.UserID,
		Priority:      priority,
		Status:        models.TaskStatusPending,
		DueDate:       req.DueDate,
		ExtractedByAI: false,
	}

	if err := mp.taskRepo.CreateTask(task); err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to create task", err.Error())
		return
	}

	// Return task with details
	taskWithDetails, err := mp.taskRepo.GetTaskWithDetails(task.ID)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to fetch task details", err.Error())
		return
	}

	mp.respondWithJSON(w, http.StatusCreated, taskWithDetails)
}

// @Summary Get task details
// @Description Get details of a specific task
// @Tags Tasks
// @Accept json
// @Produce json
// @Param session_id path string true "Session ID (UUID)"
// @Param task_id path string true "Task ID (UUID)"
// @Success 200 {object} models.TaskWithDetails
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/sessions/{session_id}/tasks/{task_id} [get]
func (mp *ManagingPortal) getTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Get task_id from URL
	taskIDStr := r.URL.Query().Get("task_id")
	if taskIDStr == "" {
		mp.respondWithError(w, http.StatusBadRequest, "task_id is required", "")
		return
	}

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid task_id format", err.Error())
		return
	}

	// Get task with details
	task, err := mp.taskRepo.GetTaskWithDetails(taskID)
	if err != nil {
		if err.Error() == "record not found" {
			mp.respondWithError(w, http.StatusNotFound, "Task not found", "")
			return
		}
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to fetch task", err.Error())
		return
	}

	mp.respondWithJSON(w, http.StatusOK, task)
}

// @Summary Update a task
// @Description Update an existing task
// @Tags Tasks
// @Accept json
// @Produce json
// @Param session_id path string true "Session ID (UUID)"
// @Param task_id path string true "Task ID (UUID)"
// @Param task body models.UpdateTaskRequest true "Updated task details"
// @Success 200 {object} models.TaskWithDetails
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/sessions/{session_id}/tasks/{task_id} [put]
func (mp *ManagingPortal) updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Get task_id from URL
	taskIDStr := r.URL.Query().Get("task_id")
	if taskIDStr == "" {
		mp.respondWithError(w, http.StatusBadRequest, "task_id is required", "")
		return
	}

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid task_id format", err.Error())
		return
	}

	// Parse request body
	var req models.UpdateTaskRequest
	if err := parseJSONBody(r, &req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Hint != nil {
		updates["hint"] = *req.Hint
	}
	if req.AssignedTo != nil {
		updates["assigned_to"] = *req.AssignedTo
	}
	if req.Priority != nil {
		// Validate priority
		switch *req.Priority {
		case string(models.TaskPriorityLow), string(models.TaskPriorityMedium),
			string(models.TaskPriorityHigh), string(models.TaskPriorityUrgent):
			updates["priority"] = *req.Priority
		default:
			mp.respondWithError(w, http.StatusBadRequest, "Invalid priority", "")
			return
		}
	}
	if req.Status != nil {
		// Validate status
		switch *req.Status {
		case string(models.TaskStatusPending), string(models.TaskStatusInProgress),
			string(models.TaskStatusCompleted), string(models.TaskStatusCancelled):
			updates["status"] = *req.Status
		default:
			mp.respondWithError(w, http.StatusBadRequest, "Invalid status", "")
			return
		}
	}
	if req.DueDate != nil {
		updates["due_date"] = *req.DueDate
	}

	// Update task
	if err := mp.taskRepo.UpdateTask(taskID, updates); err != nil {
		if err.Error() == "record not found" {
			mp.respondWithError(w, http.StatusNotFound, "Task not found", "")
			return
		}
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to update task", err.Error())
		return
	}

	// Return updated task with details
	task, err := mp.taskRepo.GetTaskWithDetails(taskID)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to fetch updated task", err.Error())
		return
	}

	mp.respondWithJSON(w, http.StatusOK, task)
}

// @Summary Delete a task
// @Description Delete a task
// @Tags Tasks
// @Accept json
// @Produce json
// @Param session_id path string true "Session ID (UUID)"
// @Param task_id path string true "Task ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/sessions/{session_id}/tasks/{task_id} [delete]
func (mp *ManagingPortal) deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Get task_id from URL
	taskIDStr := r.URL.Query().Get("task_id")
	if taskIDStr == "" {
		mp.respondWithError(w, http.StatusBadRequest, "task_id is required", "")
		return
	}

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid task_id format", err.Error())
		return
	}

	// Delete task
	if err := mp.taskRepo.DeleteTask(taskID); err != nil {
		if err.Error() == "record not found" {
			mp.respondWithError(w, http.StatusNotFound, "Task not found", "")
			return
		}
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to delete task", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

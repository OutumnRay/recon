package database

import (
	"fmt"
	"time"

	"Recontext.online/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskRepository handles task-related database operations
type TaskRepository struct {
	db *gorm.DB
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// CreateTask creates a new task
func (r *TaskRepository) CreateTask(task *models.Task) error {
	return r.db.Create(task).Error
}

// GetTaskByID retrieves a task by ID
func (r *TaskRepository) GetTaskByID(taskID uuid.UUID) (*models.Task, error) {
	var task models.Task
	err := r.db.Where("id = ?", taskID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetTaskWithDetails retrieves a task with user details
func (r *TaskRepository) GetTaskWithDetails(taskID uuid.UUID) (*models.TaskWithDetails, error) {
	task, err := r.GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}

	taskWithDetails := &models.TaskWithDetails{
		Task: *task,
	}

	// Load assigned_to user
	if task.AssignedTo != nil {
		var assignedToUser User
		if err := r.db.Where("id = ?", task.AssignedTo).First(&assignedToUser).Error; err == nil {
			firstName := ""
			if assignedToUser.FirstName != nil {
				firstName = *assignedToUser.FirstName
			}
			lastName := ""
			if assignedToUser.LastName != nil {
				lastName = *assignedToUser.LastName
			}
			taskWithDetails.AssignedToUser = &models.UserInfo{
				ID:        assignedToUser.ID,
				Username:  assignedToUser.Username,
				Email:     assignedToUser.Email,
				FirstName: firstName,
				LastName:  lastName,
				Role:      models.UserRole(assignedToUser.Role),
			}
		}
	}

	// Load assigned_by user
	if task.AssignedBy != nil {
		var assignedByUser User
		if err := r.db.Where("id = ?", task.AssignedBy).First(&assignedByUser).Error; err == nil {
			firstName := ""
			if assignedByUser.FirstName != nil {
				firstName = *assignedByUser.FirstName
			}
			lastName := ""
			if assignedByUser.LastName != nil {
				lastName = *assignedByUser.LastName
			}
			taskWithDetails.AssignedByUser = &models.UserInfo{
				ID:        assignedByUser.ID,
				Username:  assignedByUser.Username,
				Email:     assignedByUser.Email,
				FirstName: firstName,
				LastName:  lastName,
				Role:      models.UserRole(assignedByUser.Role),
			}
		}
	}

	return taskWithDetails, nil
}

// ListTasks retrieves tasks with filters and pagination
func (r *TaskRepository) ListTasks(req *models.ListTasksRequest) ([]models.Task, int64, error) {
	var tasks []models.Task
	var total int64

	query := r.db.Model(&models.Task{})

	// Apply filters
	if req.SessionID != nil {
		query = query.Where("session_id = ?", *req.SessionID)
	}
	if req.MeetingID != nil {
		query = query.Where("meeting_id = ?", *req.MeetingID)
	}
	if req.AssignedTo != nil {
		query = query.Where("assigned_to = ?", *req.AssignedTo)
	}
	if req.Status != nil && *req.Status != "" {
		query = query.Where("status = ?", *req.Status)
	}
	if req.Priority != nil && *req.Priority != "" {
		query = query.Where("priority = ?", *req.Priority)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	query = query.Order("created_at DESC")
	if req.PageSize > 0 {
		query = query.Limit(req.PageSize)
	}
	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	// Execute query
	if err := query.Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// ListTasksWithDetails retrieves tasks with user details
func (r *TaskRepository) ListTasksWithDetails(req *models.ListTasksRequest) ([]models.TaskWithDetails, int64, error) {
	tasks, total, err := r.ListTasks(req)
	if err != nil {
		return nil, 0, err
	}

	// Collect all user IDs
	userIDs := make(map[uuid.UUID]bool)
	for _, task := range tasks {
		if task.AssignedTo != nil {
			userIDs[*task.AssignedTo] = true
		}
		if task.AssignedBy != nil {
			userIDs[*task.AssignedBy] = true
		}
	}

	// Batch load users
	userIDList := make([]uuid.UUID, 0, len(userIDs))
	for id := range userIDs {
		userIDList = append(userIDList, id)
	}

	var users []User
	userMap := make(map[uuid.UUID]*models.UserInfo)
	if len(userIDList) > 0 {
		if err := r.db.Where("id IN ?", userIDList).Find(&users).Error; err == nil {
			for _, user := range users {
				firstName := ""
				if user.FirstName != nil {
					firstName = *user.FirstName
				}
				lastName := ""
				if user.LastName != nil {
					lastName = *user.LastName
				}
				userMap[user.ID] = &models.UserInfo{
					ID:        user.ID,
					Username:  user.Username,
					Email:     user.Email,
					FirstName: firstName,
					LastName:  lastName,
					Role:      models.UserRole(user.Role),
				}
			}
		}
	}

	// Build result with details
	result := make([]models.TaskWithDetails, len(tasks))
	for i, task := range tasks {
		result[i] = models.TaskWithDetails{
			Task: task,
		}
		if task.AssignedTo != nil {
			result[i].AssignedToUser = userMap[*task.AssignedTo]
		}
		if task.AssignedBy != nil {
			result[i].AssignedByUser = userMap[*task.AssignedBy]
		}
	}

	return result, total, nil
}

// UpdateTask updates a task
func (r *TaskRepository) UpdateTask(taskID uuid.UUID, updates map[string]interface{}) error {
	// Add updated_at timestamp
	updates["updated_at"] = time.Now()

	// If status is being changed to completed, set completed_at
	if status, ok := updates["status"].(string); ok && status == string(models.TaskStatusCompleted) {
		if _, hasCompletedAt := updates["completed_at"]; !hasCompletedAt {
			updates["completed_at"] = time.Now()
		}
	}

	// If status is being changed from completed to something else, clear completed_at
	if status, ok := updates["status"].(string); ok && status != string(models.TaskStatusCompleted) {
		updates["completed_at"] = nil
	}

	return r.db.Model(&models.Task{}).Where("id = ?", taskID).Updates(updates).Error
}

// UpdateTaskStatus updates only the status of a task
func (r *TaskRepository) UpdateTaskStatus(taskID uuid.UUID, status models.TaskStatus) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == models.TaskStatusCompleted {
		updates["completed_at"] = time.Now()
	} else {
		updates["completed_at"] = nil
	}

	result := r.db.Model(&models.Task{}).Where("id = ?", taskID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// DeleteTask deletes a task (hard delete)
func (r *TaskRepository) DeleteTask(taskID uuid.UUID) error {
	result := r.db.Delete(&models.Task{}, "id = ?", taskID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetTasksBySession retrieves all tasks for a session
func (r *TaskRepository) GetTasksBySession(sessionID uuid.UUID) ([]models.Task, error) {
	var tasks []models.Task
	err := r.db.Where("session_id = ?", sessionID).Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

// GetTasksByMeeting retrieves all tasks for a meeting
func (r *TaskRepository) GetTasksByMeeting(meetingID uuid.UUID) ([]models.Task, error) {
	var tasks []models.Task
	err := r.db.Where("meeting_id = ?", meetingID).Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

// GetMyTasks retrieves all tasks assigned to a user
func (r *TaskRepository) GetMyTasks(userID uuid.UUID, status *string) ([]models.Task, error) {
	query := r.db.Where("assigned_to = ?", userID)

	if status != nil && *status != "" {
		query = query.Where("status = ?", *status)
	}

	var tasks []models.Task
	err := query.Order("CASE WHEN due_date IS NULL THEN 1 ELSE 0 END, due_date ASC, created_at DESC").Find(&tasks).Error
	return tasks, err
}

// GetMyTasksWithDetails retrieves all tasks assigned to a user with details
func (r *TaskRepository) GetMyTasksWithDetails(userID uuid.UUID, status *string) ([]models.TaskWithDetails, error) {
	query := r.db.Where("assigned_to = ?", userID)

	if status != nil && *status != "" {
		query = query.Where("status = ?", *status)
	}

	var tasks []models.Task
	err := query.Order("CASE WHEN due_date IS NULL THEN 1 ELSE 0 END, due_date ASC, created_at DESC").Find(&tasks).Error
	if err != nil {
		return nil, err
	}

	// Collect all user IDs for batch loading
	userIDs := make(map[uuid.UUID]bool)
	for _, task := range tasks {
		if task.AssignedBy != nil {
			userIDs[*task.AssignedBy] = true
		}
	}

	// Batch load users
	userIDList := make([]uuid.UUID, 0, len(userIDs))
	for id := range userIDs {
		userIDList = append(userIDList, id)
	}

	var users []User
	userMap := make(map[uuid.UUID]*models.UserInfo)
	if len(userIDList) > 0 {
		if err := r.db.Where("id IN ?", userIDList).Find(&users).Error; err == nil {
			for _, user := range users {
				firstName := ""
				if user.FirstName != nil {
					firstName = *user.FirstName
				}
				lastName := ""
				if user.LastName != nil {
					lastName = *user.LastName
				}
				userMap[user.ID] = &models.UserInfo{
					ID:        user.ID,
					Username:  user.Username,
					Email:     user.Email,
					FirstName: firstName,
					LastName:  lastName,
					Role:      models.UserRole(user.Role),
				}
			}
		}
	}

	// Also load current user info for assigned_to
	var currentUser User
	var currentUserInfo *models.UserInfo
	if err := r.db.Where("id = ?", userID).First(&currentUser).Error; err == nil {
		firstName := ""
		if currentUser.FirstName != nil {
			firstName = *currentUser.FirstName
		}
		lastName := ""
		if currentUser.LastName != nil {
			lastName = *currentUser.LastName
		}
		currentUserInfo = &models.UserInfo{
			ID:        currentUser.ID,
			Username:  currentUser.Username,
			Email:     currentUser.Email,
			FirstName: firstName,
			LastName:  lastName,
			Role:      models.UserRole(currentUser.Role),
		}
	}

	// Build result with details
	result := make([]models.TaskWithDetails, len(tasks))
	for i, task := range tasks {
		result[i] = models.TaskWithDetails{
			Task:           task,
			AssignedToUser: currentUserInfo,
		}
		if task.AssignedBy != nil {
			result[i].AssignedByUser = userMap[*task.AssignedBy]
		}
	}

	return result, nil
}

// GetTaskStats returns task statistics for a session
func (r *TaskRepository) GetTaskStats(sessionID uuid.UUID) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count by status
	var statusCounts []struct {
		Status string
		Count  int64
	}

	err := r.db.Model(&models.Task{}).
		Where("session_id = ?", sessionID).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts).Error

	if err != nil {
		return nil, err
	}

	for _, sc := range statusCounts {
		stats[sc.Status] = sc.Count
	}

	// Total count
	var total int64
	err = r.db.Model(&models.Task{}).Where("session_id = ?", sessionID).Count(&total).Error
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	return stats, nil
}

// CheckTaskAccess verifies if a user has access to a task
// Returns true if user is assigned to the task, created it, or is admin
func (r *TaskRepository) CheckTaskAccess(taskID uuid.UUID, userID uuid.UUID, role string) (bool, error) {
	// Admins have access to all tasks
	if role == "admin" {
		return true, nil
	}

	var task models.Task
	err := r.db.Where("id = ?", taskID).First(&task).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	// Check if user is assigned to the task
	if task.AssignedTo != nil && *task.AssignedTo == userID {
		return true, nil
	}

	// Check if user created the task
	if task.AssignedBy != nil && *task.AssignedBy == userID {
		return true, nil
	}

	// Check if user is participant of the meeting/session
	if task.MeetingID != nil {
		var count int64
		err := r.db.Model(&MeetingParticipant{}).
			Where("meeting_id = ? AND user_id = ?", *task.MeetingID, userID).
			Count(&count).Error
		if err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}

		// Check if user is meeting creator
		var meeting Meeting
		err = r.db.Where("id = ? AND created_by = ?", *task.MeetingID, userID).First(&meeting).Error
		if err == nil {
			return true, nil
		}
	}

	return false, nil
}

// GetUserByUsername finds a user by username (for AI task assignment)
func (r *TaskRepository) GetUserByUsername(username string) (*uuid.UUID, error) {
	var user User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found: %s", username)
		}
		return nil, err
	}
	return &user.ID, nil
}

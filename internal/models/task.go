package models

import (
	"time"

	"github.com/google/uuid"
)

// TaskStatus представляет статус задачи
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"     // Ожидает выполнения
	TaskStatusInProgress TaskStatus = "in_progress" // В работе
	TaskStatusCompleted  TaskStatus = "completed"   // Завершена
	TaskStatusCancelled  TaskStatus = "cancelled"   // Отменена
)

// TaskPriority представляет приоритет задачи
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"    // Низкий
	TaskPriorityMedium TaskPriority = "medium" // Средний
	TaskPriorityHigh   TaskPriority = "high"   // Высокий
	TaskPriorityUrgent TaskPriority = "urgent" // Срочно
)

// Task представляет задачу из встречи/сессии
type Task struct {
	// Уникальный идентификатор задачи
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id" db:"id"`

	// Связь с сессией LiveKit (обязательно)
	SessionID uuid.UUID `gorm:"type:uuid;not null;index" json:"session_id" db:"session_id"`

	// Связь со встречей (опционально)
	MeetingID *uuid.UUID `gorm:"type:uuid;index" json:"meeting_id,omitempty" db:"meeting_id"`

	// Информация о задаче
	Title       string  `gorm:"type:varchar(500);not null" json:"title" db:"title"`                // Краткое описание
	Description *string `gorm:"type:text" json:"description,omitempty" db:"description"`           // Подробное описание
	Hint        *string `gorm:"type:text" json:"hint,omitempty" db:"hint"`                         // Подсказка как решить

	// Назначение задачи
	AssignedTo *uuid.UUID `gorm:"type:uuid;index" json:"assigned_to,omitempty" db:"assigned_to"` // Кому назначена (NULL = никому)
	AssignedBy *uuid.UUID `gorm:"type:uuid" json:"assigned_by,omitempty" db:"assigned_by"`        // Кто назначил

	// Статус и приоритет
	Status   TaskStatus   `gorm:"type:varchar(50);not null;default:'pending';index" json:"status" db:"status"`      // Статус задачи
	Priority TaskPriority `gorm:"type:varchar(50);default:'medium'" json:"priority" db:"priority"`                  // Приоритет

	// Дедлайн
	DueDate *time.Time `gorm:"index" json:"due_date,omitempty" db:"due_date"` // Срок выполнения

	// Метаданные AI-извлечения
	ExtractedByAI bool     `gorm:"default:false" json:"extracted_by_ai" db:"extracted_by_ai"`                    // Извлечена AI или создана вручную
	AIConfidence  *float64 `gorm:"type:double precision" json:"ai_confidence,omitempty" db:"ai_confidence"`      // Уверенность AI (0-1)
	SourceSegment *string  `gorm:"type:text" json:"source_segment,omitempty" db:"source_segment"`                // Фрагмент транскрипции

	// Временные метки
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at" db:"updated_at"`
}

// TaskWithDetails содержит задачу с дополнительной информацией о пользователях
type TaskWithDetails struct {
	Task
	AssignedToUser *UserInfo `json:"assigned_to_user,omitempty"`
	AssignedByUser *UserInfo `json:"assigned_by_user,omitempty"`
}

// CreateTaskRequest представляет запрос на создание задачи
type CreateTaskRequest struct {
	Title       string     `json:"title" binding:"required,max=500"`
	Description *string    `json:"description,omitempty"`
	Hint        *string    `json:"hint,omitempty"`
	AssignedTo  *uuid.UUID `json:"assigned_to,omitempty"`
	Priority    string     `json:"priority,omitempty"` // low, medium, high, urgent
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// UpdateTaskRequest представляет запрос на обновление задачи
type UpdateTaskRequest struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	Hint        *string    `json:"hint,omitempty"`
	AssignedTo  *uuid.UUID `json:"assigned_to,omitempty"`
	Status      *string    `json:"status,omitempty"`    // pending, in_progress, completed, cancelled
	Priority    *string    `json:"priority,omitempty"`  // low, medium, high, urgent
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// UpdateTaskStatusRequest представляет запрос на обновление только статуса задачи
type UpdateTaskStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending in_progress completed cancelled"`
}

// ListTasksRequest представляет параметры запроса списка задач
type ListTasksRequest struct {
	SessionID  *uuid.UUID `form:"session_id"`
	MeetingID  *uuid.UUID `form:"meeting_id"`
	AssignedTo *uuid.UUID `form:"assigned_to"`
	Status     *string    `form:"status"`     // pending, in_progress, completed, cancelled
	Priority   *string    `form:"priority"`   // low, medium, high, urgent
	PageSize   int        `form:"page_size"`
	Offset     int        `form:"offset"`
}

// ListTasksResponse представляет ответ со списком задач
type ListTasksResponse struct {
	Items    []TaskWithDetails `json:"items"`
	Total    int64             `json:"total"`
	PageSize int               `json:"page_size"`
	Offset   int               `json:"offset"`
}

// ExtractTasksRequest представляет запрос на AI-извлечение задач
type ExtractTasksRequest struct {
	LLMProvider   string  `json:"llm_provider" binding:"required,oneof=openai anthropic ollama"` // openai, anthropic, ollama
	Model         string  `json:"model" binding:"required"`                                       // gpt-4, claude-3-opus, llama2
	AutoAssign    bool    `json:"auto_assign"`                                                    // Автоматически назначать по username
	MinConfidence float64 `json:"min_confidence"`                                                 // Минимальная уверенность (0-1)
}

// ExtractedTask представляет задачу, извлеченную AI из транскрипции
type ExtractedTask struct {
	Title              string  `json:"title"`
	Description        string  `json:"description"`
	Hint               *string `json:"hint,omitempty"`
	AssignedToUsername *string `json:"assigned_to_username,omitempty"`
	Priority           string  `json:"priority"`
	Confidence         float64 `json:"confidence"`
	SourceSegment      string  `json:"source_segment"`
}

// ExtractTasksResponse представляет ответ на запрос AI-извлечения задач
type ExtractTasksResponse struct {
	ExtractedCount int               `json:"extracted_count"` // Сколько извлечено
	SavedCount     int               `json:"saved_count"`     // Сколько сохранено
	SkippedCount   int               `json:"skipped_count"`   // Сколько пропущено (низкая confidence)
	Tasks          []TaskWithDetails `json:"tasks"`           // Сохраненные задачи
}

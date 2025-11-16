package models

import (
	"time"

	"github.com/google/uuid"
)

// UserGroup представляет группу пользователей со специфическими разрешениями
type UserGroup struct {
	// Уникальный идентификатор группы
	ID          uuid.UUID              `json:"id" db:"id"`
	// Название группы
	Name        string                 `json:"name" db:"name"`
	// Описание группы
	Description string                 `json:"description" db:"description"`
	// Динамические разрешения на основе JSON
	Permissions map[string]interface{} `json:"permissions" db:"permissions"`
	// ID организации, к которой относится группа
	OrganizationID *uuid.UUID          `json:"organization_id,omitempty" db:"organization_id"`
	// Время создания группы
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	// Время последнего обновления группы
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// Permission представляет динамическое правило разрешения
type Permission struct {
	// Тип ресурса (recordings, transcripts, users и т.д.)
	Resource string                 `json:"resource" example:"recordings"`
	// Разрешенные действия (read, write, delete и т.д.)
	Actions  []string               `json:"actions" example:"read,write"`
	// Область действия (all, own, group)
	Scope    string                 `json:"scope" example:"own"`
	// Дополнительные фильтры
	Filters  map[string]interface{} `json:"filters,omitempty"`
}

// PermissionSet представляет полный набор разрешений для группы
type PermissionSet struct {
	// Идентификатор группы
	GroupID     uuid.UUID              `json:"group_id"`
	// Список разрешений
	Permissions []Permission           `json:"permissions"`
	// Пользовательские JSON правила для расширенных разрешений
	CustomRules map[string]interface{} `json:"custom_rules,omitempty"`
}

// CreateGroupRequest представляет запрос на создание новой группы
type CreateGroupRequest struct {
	// Название группы
	Name        string                 `json:"name" binding:"required" example:"Editors"`
	// Описание группы
	Description string                 `json:"description" example:"Users who can edit recordings"`
	// Разрешения группы
	Permissions map[string]interface{} `json:"permissions" binding:"required"`
	// ID организации, к которой относится группа
	OrganizationID *uuid.UUID          `json:"organization_id"`
}

// UpdateGroupRequest представляет запрос на обновление группы
type UpdateGroupRequest struct {
	// Название группы
	Name        string                 `json:"name,omitempty" example:"Senior Editors"`
	// Описание группы
	Description string                 `json:"description,omitempty"`
	// Разрешения группы
	Permissions map[string]interface{} `json:"permissions,omitempty"`
	// ID организации, к которой относится группа
	OrganizationID *uuid.UUID          `json:"organization_id"`
}

// GroupMembership представляет членство пользователя в группе
type GroupMembership struct {
	// Идентификатор пользователя
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	// Идентификатор группы
	GroupID   uuid.UUID `json:"group_id" db:"group_id"`
	// Время добавления в группу
	AddedAt   time.Time `json:"added_at" db:"added_at"`
	// Кем был добавлен
	AddedBy   uuid.UUID `json:"added_by" db:"added_by"`
}

// AddUserToGroupRequest представляет запрос на добавление пользователя в группу
type AddUserToGroupRequest struct {
	// Идентификатор пользователя
	UserID  uuid.UUID `json:"user_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Идентификатор группы
	GroupID uuid.UUID `json:"group_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// ListGroupsResponse представляет список групп
type ListGroupsResponse struct {
	// Список групп
	Items    []UserGroup `json:"items"`
	// Общее количество групп
	Total    int         `json:"total"`
	// Смещение от начала
	Offset   int         `json:"offset"`
	// Размер страницы
	PageSize int         `json:"page_size"`
}

// PermissionCheckRequest представляет запрос на проверку разрешений
type PermissionCheckRequest struct {
	// Идентификатор пользователя
	UserID   uuid.UUID `json:"user_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Ресурс для проверки
	Resource string    `json:"resource" binding:"required" example:"recordings"`
	// Действие для проверки
	Action   string    `json:"action" binding:"required" example:"write"`
}

// PermissionCheckResponse представляет результат проверки разрешений
type PermissionCheckResponse struct {
	// Разрешено ли действие
	Allowed bool   `json:"allowed" example:"true"`
	// Причина решения
	Reason  string `json:"reason,omitempty" example:"User has write access to recordings"`
}

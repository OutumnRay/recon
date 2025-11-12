package models

import (
	"time"

	"github.com/google/uuid"
)

// UserGroup представляет группу пользователей со специфическими разрешениями
type UserGroup struct {
	// ID - уникальный идентификатор группы
	ID          uuid.UUID              `json:"id" db:"id"`
	// Name - название группы
	Name        string                 `json:"name" db:"name"`
	// Description - описание группы
	Description string                 `json:"description" db:"description"`
	// Permissions - динамические разрешения на основе JSON
	Permissions map[string]interface{} `json:"permissions" db:"permissions"`
	// CreatedAt - время создания группы
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	// UpdatedAt - время последнего обновления группы
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// Permission представляет динамическое правило разрешения
type Permission struct {
	// Resource - тип ресурса (recordings, transcripts, users и т.д.)
	Resource string                 `json:"resource" example:"recordings"`
	// Actions - разрешенные действия (read, write, delete и т.д.)
	Actions  []string               `json:"actions" example:"read,write"`
	// Scope - область действия (all, own, group)
	Scope    string                 `json:"scope" example:"own"`
	// Filters - дополнительные фильтры
	Filters  map[string]interface{} `json:"filters,omitempty"`
}

// PermissionSet представляет полный набор разрешений для группы
type PermissionSet struct {
	// GroupID - идентификатор группы
	GroupID     uuid.UUID              `json:"group_id"`
	// Permissions - список разрешений
	Permissions []Permission           `json:"permissions"`
	// CustomRules - пользовательские JSON правила для расширенных разрешений
	CustomRules map[string]interface{} `json:"custom_rules,omitempty"`
}

// CreateGroupRequest представляет запрос на создание новой группы
type CreateGroupRequest struct {
	// Name - название группы
	Name        string                 `json:"name" binding:"required" example:"Editors"`
	// Description - описание группы
	Description string                 `json:"description" example:"Users who can edit recordings"`
	// Permissions - разрешения группы
	Permissions map[string]interface{} `json:"permissions" binding:"required"`
}

// UpdateGroupRequest представляет запрос на обновление группы
type UpdateGroupRequest struct {
	// Name - название группы
	Name        string                 `json:"name,omitempty" example:"Senior Editors"`
	// Description - описание группы
	Description string                 `json:"description,omitempty"`
	// Permissions - разрешения группы
	Permissions map[string]interface{} `json:"permissions,omitempty"`
}

// GroupMembership представляет членство пользователя в группе
type GroupMembership struct {
	// UserID - идентификатор пользователя
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	// GroupID - идентификатор группы
	GroupID   uuid.UUID `json:"group_id" db:"group_id"`
	// AddedAt - время добавления в группу
	AddedAt   time.Time `json:"added_at" db:"added_at"`
	// AddedBy - кем был добавлен
	AddedBy   uuid.UUID `json:"added_by" db:"added_by"`
}

// AddUserToGroupRequest представляет запрос на добавление пользователя в группу
type AddUserToGroupRequest struct {
	// UserID - идентификатор пользователя
	UserID  uuid.UUID `json:"user_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// GroupID - идентификатор группы
	GroupID uuid.UUID `json:"group_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// ListGroupsResponse представляет список групп
type ListGroupsResponse struct {
	// Items - список групп
	Items    []UserGroup `json:"items"`
	// Total - общее количество групп
	Total    int         `json:"total"`
	// Offset - смещение от начала
	Offset   int         `json:"offset"`
	// PageSize - размер страницы
	PageSize int         `json:"page_size"`
}

// PermissionCheckRequest представляет запрос на проверку разрешений
type PermissionCheckRequest struct {
	// UserID - идентификатор пользователя
	UserID   uuid.UUID `json:"user_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Resource - ресурс для проверки
	Resource string    `json:"resource" binding:"required" example:"recordings"`
	// Action - действие для проверки
	Action   string    `json:"action" binding:"required" example:"write"`
}

// PermissionCheckResponse представляет результат проверки разрешений
type PermissionCheckResponse struct {
	// Allowed - разрешено ли действие
	Allowed bool   `json:"allowed" example:"true"`
	// Reason - причина решения
	Reason  string `json:"reason,omitempty" example:"User has write access to recordings"`
}

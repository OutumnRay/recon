package models

import "time"

// UserGroup represents a group of users with specific permissions
type UserGroup struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Permissions map[string]interface{} `json:"permissions" db:"permissions"` // JSON-based dynamic permissions
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// Permission represents a dynamic permission rule
type Permission struct {
	Resource string                 `json:"resource" example:"recordings"`  // Resource type (recordings, transcripts, users, etc.)
	Actions  []string               `json:"actions" example:"read,write"`   // Allowed actions (read, write, delete, etc.)
	Scope    string                 `json:"scope" example:"own"`            // Scope (all, own, group)
	Filters  map[string]interface{} `json:"filters,omitempty"`              // Additional filters
}

// PermissionSet represents a complete set of permissions for a group
type PermissionSet struct {
	GroupID     string                 `json:"group_id"`
	Permissions []Permission           `json:"permissions"`
	CustomRules map[string]interface{} `json:"custom_rules,omitempty"` // Custom JSON rules for advanced permissions
}

// CreateGroupRequest represents a request to create a new group
type CreateGroupRequest struct {
	Name        string                 `json:"name" binding:"required" example:"Editors"`
	Description string                 `json:"description" example:"Users who can edit recordings"`
	Permissions map[string]interface{} `json:"permissions" binding:"required"`
}

// UpdateGroupRequest represents a request to update a group
type UpdateGroupRequest struct {
	Name        string                 `json:"name,omitempty" example:"Senior Editors"`
	Description string                 `json:"description,omitempty"`
	Permissions map[string]interface{} `json:"permissions,omitempty"`
}

// GroupMembership represents a user's membership in a group
type GroupMembership struct {
	UserID    string    `json:"user_id" db:"user_id"`
	GroupID   string    `json:"group_id" db:"group_id"`
	AddedAt   time.Time `json:"added_at" db:"added_at"`
	AddedBy   string    `json:"added_by" db:"added_by"`
}

// AddUserToGroupRequest represents a request to add a user to a group
type AddUserToGroupRequest struct {
	UserID  string `json:"user_id" binding:"required" example:"user-001"`
	GroupID string `json:"group_id" binding:"required" example:"group-001"`
}

// ListGroupsResponse represents a list of groups
type ListGroupsResponse struct {
	Groups []UserGroup `json:"groups"`
	Total  int         `json:"total"`
}

// PermissionCheckRequest represents a request to check permissions
type PermissionCheckRequest struct {
	UserID   string `json:"user_id" binding:"required" example:"user-001"`
	Resource string `json:"resource" binding:"required" example:"recordings"`
	Action   string `json:"action" binding:"required" example:"write"`
}

// PermissionCheckResponse represents the result of a permission check
type PermissionCheckResponse struct {
	Allowed bool   `json:"allowed" example:"true"`
	Reason  string `json:"reason,omitempty" example:"User has write access to recordings"`
}

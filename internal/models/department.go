package models

import (
	"time"

	"github.com/google/uuid"
)

// Department represents an organizational department/unit
// Departments can be organized hierarchically with parent-child relationships
type Department struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"` // NULL for root departments
	Level       int        `json:"level" db:"level"`                    // Depth in hierarchy (0 for root)
	Path        string     `json:"path" db:"path"`                      // Full path like "root/child/grandchild"
	IsActive    bool       `json:"is_active" db:"is_active"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// DepartmentTreeNode represents a department with its children for tree display
type DepartmentTreeNode struct {
	Department
	Children []*DepartmentTreeNode `json:"children,omitempty"`
}

// CreateDepartmentRequest represents a request to create a new department
type CreateDepartmentRequest struct {
	Name        string     `json:"name" binding:"required" example:"IT Department"`
	Description string     `json:"description" example:"Information Technology Department"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// UpdateDepartmentRequest represents a request to update a department
type UpdateDepartmentRequest struct {
	Name        string     `json:"name" example:"IT Department"`
	Description string     `json:"description" example:"Updated description"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	IsActive    *bool      `json:"is_active" example:"true"`
}

// ListDepartmentsRequest represents parameters for listing departments
type ListDepartmentsRequest struct {
	ParentID   *uuid.UUID `json:"parent_id" form:"parent_id"`
	IncludeAll bool       `json:"include_all" form:"include_all"` // Include inactive departments
	Page       int        `json:"page" form:"page" example:"1"`
	PageSize   int        `json:"page_size" form:"page_size" example:"20"`
}

// ListDepartmentsResponse represents a paginated list of departments
type ListDepartmentsResponse struct {
	Items    []Department `json:"items"`
	Total    int          `json:"total"`
	Offset   int          `json:"offset"`
	PageSize int          `json:"page_size"`
}

// DepartmentWithStats represents a department with statistics
type DepartmentWithStats struct {
	Department
	UserCount       int `json:"user_count" db:"user_count"`
	ChildCount      int `json:"child_count" db:"child_count"`
	TotalUsersCount int `json:"total_users_count"` // Including all sub-departments
}

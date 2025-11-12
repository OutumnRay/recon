package models

import (
	"time"

	"github.com/google/uuid"
)

// Department представляет организационный отдел/подразделение
// Отделы могут быть организованы иерархически с отношениями родитель-потомок
type Department struct {
	// ID - уникальный идентификатор отдела
	ID          uuid.UUID  `json:"id" db:"id"`
	// Name - название отдела
	Name        string     `json:"name" db:"name"`
	// Description - описание отдела
	Description string     `json:"description" db:"description"`
	// ParentID - идентификатор родительского отдела (NULL для корневых отделов)
	ParentID    *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	// Level - глубина в иерархии (0 для корневого)
	Level       int        `json:"level" db:"level"`
	// Path - полный путь вида "root/child/grandchild"
	Path        string     `json:"path" db:"path"`
	// IsActive - активен ли отдел
	IsActive    bool       `json:"is_active" db:"is_active"`
	// CreatedAt - время создания отдела
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	// UpdatedAt - время последнего обновления отдела
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// DepartmentTreeNode представляет отдел с его дочерними элементами для древовидного отображения
type DepartmentTreeNode struct {
	Department
	// Children - дочерние отделы
	Children []*DepartmentTreeNode `json:"children,omitempty"`
}

// CreateDepartmentRequest представляет запрос на создание нового отдела
type CreateDepartmentRequest struct {
	// Name - название отдела
	Name        string     `json:"name" binding:"required" example:"IT Department"`
	// Description - описание отдела
	Description string     `json:"description" example:"Information Technology Department"`
	// ParentID - идентификатор родительского отдела (опционально)
	ParentID    *uuid.UUID `json:"parent_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// UpdateDepartmentRequest представляет запрос на обновление отдела
type UpdateDepartmentRequest struct {
	// Name - название отдела
	Name        string     `json:"name" example:"IT Department"`
	// Description - описание отдела
	Description string     `json:"description" example:"Updated description"`
	// ParentID - идентификатор родительского отдела
	ParentID    *uuid.UUID `json:"parent_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	// IsActive - активен ли отдел
	IsActive    *bool      `json:"is_active" example:"true"`
}

// ListDepartmentsRequest представляет параметры для получения списка отделов
type ListDepartmentsRequest struct {
	// ParentID - фильтр по родительскому отделу
	ParentID   *uuid.UUID `json:"parent_id" form:"parent_id"`
	// IncludeAll - включать неактивные отделы
	IncludeAll bool       `json:"include_all" form:"include_all"`
	// Page - номер страницы
	Page       int        `json:"page" form:"page" example:"1"`
	// PageSize - размер страницы
	PageSize   int        `json:"page_size" form:"page_size" example:"20"`
}

// ListDepartmentsResponse представляет постраничный список отделов
type ListDepartmentsResponse struct {
	// Items - список отделов
	Items    []Department `json:"items"`
	// Total - общее количество отделов
	Total    int          `json:"total"`
	// Offset - смещение от начала
	Offset   int          `json:"offset"`
	// PageSize - размер страницы
	PageSize int          `json:"page_size"`
}

// DepartmentWithStats представляет отдел с статистикой
type DepartmentWithStats struct {
	Department
	// UserCount - количество пользователей в отделе
	UserCount       int `json:"user_count" db:"user_count"`
	// ChildCount - количество дочерних отделов
	ChildCount      int `json:"child_count" db:"child_count"`
	// TotalUsersCount - общее количество пользователей включая все подотделы
	TotalUsersCount int `json:"total_users_count"`
}

package models

import (
	"time"

	"github.com/google/uuid"
)

// Department представляет организационный отдел/подразделение
// Отделы могут быть организованы иерархически с отношениями родитель-потомок
type Department struct {
	// Уникальный идентификатор отдела
	ID          uuid.UUID  `json:"id" db:"id"`
	// Название отдела
	Name        string     `json:"name" db:"name"`
	// Описание отдела
	Description string     `json:"description" db:"description"`
	// Идентификатор родительского отдела (NULL для корневых отделов)
	ParentID    *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	// ID организации, к которой относится отдел
	OrganizationID *uuid.UUID `json:"organization_id,omitempty" db:"organization_id"`
	// Глубина в иерархии (0 для корневого)
	Level       int        `json:"level" db:"level"`
	// Полный путь вида "root/child/grandchild"
	Path        string     `json:"path" db:"path"`
	// Активен ли отдел
	IsActive    bool       `json:"is_active" db:"is_active"`
	// Время создания отдела
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	// Время последнего обновления отдела
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// DepartmentTreeNode представляет отдел с его дочерними элементами для древовидного отображения
type DepartmentTreeNode struct {
	Department
	// Дочерние отделы
	Children []*DepartmentTreeNode `json:"children,omitempty"`
}

// CreateDepartmentRequest представляет запрос на создание нового отдела
type CreateDepartmentRequest struct {
	// Название отдела
	Name        string     `json:"name" binding:"required" example:"IT Department"`
	// Описание отдела
	Description string     `json:"description" example:"Information Technology Department"`
	// Идентификатор родительского отдела (опционально)
	ParentID    *uuid.UUID `json:"parent_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	// ID организации, к которой относится отдел
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
}

// UpdateDepartmentRequest представляет запрос на обновление отдела
type UpdateDepartmentRequest struct {
	// Название отдела
	Name        string     `json:"name" example:"IT Department"`
	// Описание отдела
	Description string     `json:"description" example:"Updated description"`
	// Идентификатор родительского отдела
	ParentID    *uuid.UUID `json:"parent_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	// ID организации, к которой относится отдел
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	// Активен ли отдел
	IsActive    *bool      `json:"is_active" example:"true"`
}

// ListDepartmentsRequest представляет параметры для получения списка отделов
type ListDepartmentsRequest struct {
	// Фильтр по родительскому отделу
	ParentID   *uuid.UUID `json:"parent_id" form:"parent_id"`
	// Включать неактивные отделы
	IncludeAll bool       `json:"include_all" form:"include_all"`
	// Номер страницы
	Page       int        `json:"page" form:"page" example:"1"`
	// Размер страницы
	PageSize   int        `json:"page_size" form:"page_size" example:"20"`
}

// ListDepartmentsResponse представляет постраничный список отделов
type ListDepartmentsResponse struct {
	// Список отделов
	Items    []Department `json:"items"`
	// Общее количество отделов
	Total    int          `json:"total"`
	// Смещение от начала
	Offset   int          `json:"offset"`
	// Размер страницы
	PageSize int          `json:"page_size"`
}

// DepartmentWithStats представляет отдел с статистикой
type DepartmentWithStats struct {
	Department
	// Количество пользователей в отделе
	UserCount       int `json:"user_count" db:"user_count"`
	// Количество дочерних отделов
	ChildCount      int `json:"child_count" db:"child_count"`
	// Общее количество пользователей включая все подотделы
	TotalUsersCount int `json:"total_users_count"`
}

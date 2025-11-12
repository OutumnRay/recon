package models

import "github.com/google/uuid"

// UpdateUserRequest представляет запрос на обновление информации о пользователе
type UpdateUserRequest struct {
	// Email - адрес электронной почты
	Email        string           `json:"email,omitempty" example:"newemail@example.com"`
	// Password - пароль
	Password     string           `json:"password,omitempty" example:"newpassword123"`
	// Role - роль пользователя
	Role         UserRole         `json:"role,omitempty" example:"operator"`
	// FirstName - имя
	FirstName    string           `json:"first_name,omitempty" example:"John"`
	// LastName - фамилия
	LastName     string           `json:"last_name,omitempty" example:"Doe"`
	// Phone - номер телефона
	Phone        string           `json:"phone,omitempty" example:"+1234567890"`
	// Bio - биография
	Bio          string           `json:"bio,omitempty" example:"Software developer"`
	// Avatar - URL аватара
	Avatar       string           `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
	// DepartmentID - идентификатор отдела
	DepartmentID *uuid.UUID       `json:"department_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Groups - список групп
	Groups       []uuid.UUID      `json:"groups,omitempty"`
	// Permissions - разрешения пользователя
	Permissions  *UserPermissions `json:"permissions,omitempty"`
	// Language - предпочитаемый язык
	Language     string           `json:"language,omitempty" example:"en"`
	// IsActive - активен ли аккаунт
	IsActive     *bool            `json:"is_active,omitempty" example:"true"`
}

// UpdateProfileRequest представляет запрос на обновление профиля пользователя (ограниченные поля для самостоятельного обновления)
type UpdateProfileRequest struct {
	// FirstName - имя
	FirstName string `json:"first_name,omitempty" example:"John"`
	// LastName - фамилия
	LastName  string `json:"last_name,omitempty" example:"Doe"`
	// Phone - номер телефона
	Phone     string `json:"phone,omitempty" example:"+1234567890"`
	// Bio - биография
	Bio       string `json:"bio,omitempty" example:"Software developer"`
	// Avatar - URL аватара
	Avatar    string `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
	// Language - предпочитаемый язык
	Language  string `json:"language,omitempty" example:"en"`
}

// UploadAvatarResponse представляет ответ после загрузки аватара
type UploadAvatarResponse struct {
	// AvatarURL - URL загруженного аватара
	AvatarURL string `json:"avatar_url" example:"https://example.com/avatars/user-123.jpg"`
	// Message - статусное сообщение
	Message   string `json:"message" example:"Avatar uploaded successfully"`
}

// ChangePasswordRequest представляет запрос на изменение пароля
type ChangePasswordRequest struct {
	// OldPassword - старый пароль
	OldPassword string `json:"old_password" binding:"required" example:"oldpass123"`
	// NewPassword - новый пароль
	NewPassword string `json:"new_password" binding:"required,min=8" example:"newpass123"`
}

// ListUsersRequest представляет параметры для получения списка пользователей
type ListUsersRequest struct {
	// Page - номер страницы
	Page         int        `json:"page" form:"page" example:"1"`
	// PageSize - размер страницы
	PageSize     int        `json:"page_size" form:"page_size" example:"20"`
	// Role - фильтр по роли
	Role         string     `json:"role" form:"role" example:"user"`
	// DepartmentID - фильтр по отделу
	DepartmentID *uuid.UUID `json:"department_id" form:"department_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// GroupID - фильтр по группе
	GroupID      *uuid.UUID `json:"group_id" form:"group_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// IsActive - фильтр по активности
	IsActive     *bool      `json:"is_active" form:"is_active" example:"true"`
}

// ListUsersResponse представляет постраничный список пользователей
type ListUsersResponse struct {
	// Items - список пользователей
	Items    []UserInfo `json:"items"`
	// Total - общее количество пользователей
	Total    int        `json:"total"`
	// Offset - смещение от начала
	Offset   int        `json:"offset"`
	// PageSize - размер страницы
	PageSize int        `json:"page_size"`
}

// DeleteUserRequest представляет запрос на удаление пользователя
type DeleteUserRequest struct {
	// UserID - идентификатор пользователя для удаления
	UserID uuid.UUID `json:"user_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Reason - причина удаления
	Reason string    `json:"reason,omitempty" example:"Account requested deletion"`
}

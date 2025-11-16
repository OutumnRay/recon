package models

import "github.com/google/uuid"

// UpdateUserRequest представляет запрос на обновление информации о пользователе
type UpdateUserRequest struct {
	// Адрес электронной почты
	Email        string           `json:"email,omitempty" example:"newemail@example.com"`
	// Пароль
	Password     string           `json:"password,omitempty" example:"newpassword123"`
	// Роль пользователя
	Role         UserRole         `json:"role,omitempty" example:"operator"`
	// Имя
	FirstName    string           `json:"first_name,omitempty" example:"John"`
	// Фамилия
	LastName     string           `json:"last_name,omitempty" example:"Doe"`
	// Номер телефона
	Phone        string           `json:"phone,omitempty" example:"+1234567890"`
	// Биография
	Bio          string           `json:"bio,omitempty" example:"Software developer"`
	// URL аватара
	Avatar       string           `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
	// Идентификатор организации
	OrganizationID *uuid.UUID     `json:"organization_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Идентификатор отдела
	DepartmentID *uuid.UUID       `json:"department_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Список групп
	Groups       []uuid.UUID      `json:"groups,omitempty"`
	// Разрешения пользователя
	Permissions  *UserPermissions `json:"permissions,omitempty"`
	// Предпочитаемый язык
	Language     string           `json:"language,omitempty" example:"en"`
	// Активен ли аккаунт
	IsActive     *bool            `json:"is_active,omitempty" example:"true"`
}

// UpdateProfileRequest представляет запрос на обновление профиля пользователя (ограниченные поля для самостоятельного обновления)
type UpdateProfileRequest struct {
	// Имя
	FirstName string `json:"first_name,omitempty" example:"John"`
	// Фамилия
	LastName  string `json:"last_name,omitempty" example:"Doe"`
	// Номер телефона
	Phone     string `json:"phone,omitempty" example:"+1234567890"`
	// Биография
	Bio       string `json:"bio,omitempty" example:"Software developer"`
	// URL аватара
	Avatar    string `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
	// Предпочитаемый язык
	Language  string `json:"language,omitempty" example:"en"`
	// Настройки уведомлений пользователя (tracks, rooms, both)
	NotificationPreferences string `json:"notification_preferences,omitempty" example:"rooms"`
}

// UploadAvatarResponse представляет ответ после загрузки аватара
type UploadAvatarResponse struct {
	// URL загруженного аватара
	AvatarURL string `json:"avatar_url" example:"https://example.com/avatars/user-123.jpg"`
	// Статусное сообщение
	Message   string `json:"message" example:"Avatar uploaded successfully"`
}

// ChangePasswordRequest представляет запрос на изменение пароля
type ChangePasswordRequest struct {
	// Старый пароль
	OldPassword string `json:"old_password" binding:"required" example:"oldpass123"`
	// Новый пароль
	NewPassword string `json:"new_password" binding:"required,min=8" example:"newpass123"`
}

// ListUsersRequest представляет параметры для получения списка пользователей
type ListUsersRequest struct {
	// Номер страницы
	Page         int        `json:"page" form:"page" example:"1"`
	// Размер страницы
	PageSize     int        `json:"page_size" form:"page_size" example:"20"`
	// Фильтр по роли
	Role         string     `json:"role" form:"role" example:"user"`
	// Фильтр по отделу
	DepartmentID *uuid.UUID `json:"department_id" form:"department_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Фильтр по группе
	GroupID      *uuid.UUID `json:"group_id" form:"group_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Фильтр по активности
	IsActive     *bool      `json:"is_active" form:"is_active" example:"true"`
}

// ListUsersResponse представляет постраничный список пользователей
type ListUsersResponse struct {
	// Список пользователей
	Items    []UserInfo `json:"items"`
	// Общее количество пользователей
	Total    int        `json:"total"`
	// Смещение от начала
	Offset   int        `json:"offset"`
	// Размер страницы
	PageSize int        `json:"page_size"`
}

// DeleteUserRequest представляет запрос на удаление пользователя
type DeleteUserRequest struct {
	// Идентификатор пользователя для удаления
	UserID uuid.UUID `json:"user_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Причина удаления
	Reason string    `json:"reason,omitempty" example:"Account requested deletion"`
}

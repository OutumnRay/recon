package models

import (
	"time"

	"github.com/google/uuid"
)

// User представляет пользователя в системе
type User struct {
	// Уникальный идентификатор пользователя
	ID uuid.UUID `json:"id" db:"id"`
	// Имя пользователя для входа
	Username string `json:"username" db:"username"`
	// Адрес электронной почты пользователя
	Email string `json:"email" db:"email"`
	// Хешированный пароль (никогда не передается в JSON)
	Password string `json:"-" db:"password"`
	// Роль пользователя в системе
	Role UserRole `json:"role" db:"role"`
	// Имя пользователя
	FirstName string `json:"first_name,omitempty" db:"first_name"`
	// Фамилия пользователя
	LastName string `json:"last_name,omitempty" db:"last_name"`
	// Номер телефона пользователя
	Phone string `json:"phone,omitempty" db:"phone"`
	// Биография пользователя
	Bio string `json:"bio,omitempty" db:"bio"`
	// URL аватара или base64 изображение
	Avatar string `json:"avatar,omitempty" db:"avatar"`
	// Идентификатор организации, к которой принадлежит пользователь
	OrganizationID *uuid.UUID `json:"organization_id,omitempty" db:"organization_id"`
	// Идентификатор отдела, к которому принадлежит пользователь
	DepartmentID *uuid.UUID `json:"department_id,omitempty" db:"department_id"`
	// Идентификаторы групп, к которым принадлежит пользователь
	Groups []uuid.UUID `json:"groups,omitempty" db:"groups"`
	// Специфические разрешения пользователя
	Permissions UserPermissions `json:"permissions" db:"permissions"`
	// Предпочитаемый язык пользователя (ru, en и т.д.)
	Language string `json:"language" db:"language"`
	// Настройки уведомлений пользователя (tracks, rooms, both)
	NotificationPreferences string `json:"notification_preferences" db:"notification_preferences"`
	// Активен ли аккаунт пользователя
	IsActive bool `json:"is_active" db:"is_active"`
	// Время последнего входа пользователя
	LastLogin *time.Time `json:"last_login,omitempty" db:"last_login"`
	// Время создания аккаунта пользователя
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	// Время последнего обновления аккаунта пользователя
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserPermissions представляет специфические разрешения пользователя
type UserPermissions struct {
	// Разрешение на планирование видеовстреч
	CanScheduleMeetings bool `json:"can_schedule_meetings"`
	// Разрешение на управление отделом
	CanManageDepartment bool `json:"can_manage_department"`
	// Разрешение на утверждение записей
	CanApproveRecordings bool `json:"can_approve_recordings"`
}

// UserRole представляет роли пользователей
type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleUser     UserRole = "user"
	RoleOperator UserRole = "operator"
	RoleService  UserRole = "service" // Роль для сервисной аутентификации (межсервисные вызовы)
)

// LoginRequest представляет запрос на вход в систему
type LoginRequest struct {
	// Имя пользователя или email для входа
	Username string `json:"username" binding:"required" example:"admin@recontext.online"`
	// Пароль пользователя
	Password string `json:"password" binding:"required" example:"password123"`
}

// LoginResponse представляет ответ на запрос входа
type LoginResponse struct {
	// JWT токен аутентификации
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	// Время истечения срока действия токена
	ExpiresAt time.Time `json:"expires_at"`
	// Информация об аутентифицированном пользователе
	User UserInfo `json:"user"`
}

// UserInfo представляет публичную информацию о пользователе
type UserInfo struct {
	// Уникальный идентификатор пользователя
	ID uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Имя пользователя для входа
	Username string `json:"username" example:"admin"`
	// Адрес электронной почты пользователя
	Email string `json:"email" example:"admin@recontext.online"`
	// Роль пользователя в системе
	Role UserRole `json:"role" example:"admin"`
	// Идентификатор организации, к которой принадлежит пользователь
	OrganizationID *uuid.UUID `json:"organization_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Имя пользователя
	FirstName string `json:"first_name" example:"John"`
	// Фамилия пользователя
	LastName string `json:"last_name" example:"Doe"`
	// Номер телефона пользователя
	Phone string `json:"phone,omitempty" example:"+1234567890"`
	// Биография пользователя
	Bio string `json:"bio,omitempty" example:"Software developer"`
	// URL аватара или base64 изображение
	Avatar string `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
	// Идентификатор отдела, к которому принадлежит пользователь
	DepartmentID *uuid.UUID `json:"department_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Специфические разрешения пользователя
	Permissions UserPermissions `json:"permissions"`
	// Предпочитаемый язык пользователя
	Language string `json:"language" example:"en"`
	// Настройки уведомлений пользователя (tracks, rooms, both)
	NotificationPreferences string `json:"notification_preferences" example:"rooms"`
}

// RegisterRequest представляет запрос на регистрацию
type RegisterRequest struct {
	// Желаемое имя пользователя для нового аккаунта
	Username string `json:"username" binding:"required" example:"newuser"`
	// Адрес электронной почты для нового аккаунта
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
	// Желаемый пароль (минимум 8 символов)
	Password string `json:"password" binding:"required,min=8" example:"password123"`
	// Подтверждение пароля (должно совпадать с password)
	ConfirmPassword string `json:"confirm_password" binding:"required" example:"password123"`
}

// TokenClaims представляет claims JWT токена
type TokenClaims struct {
	// Идентификатор пользователя, которому принадлежит токен
	UserID uuid.UUID `json:"user_id"`
	// Имя пользователя
	Username string `json:"username"`
	// Роль пользователя
	Role UserRole `json:"role"`
	// Идентификатор организации пользователя
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	// Unix timestamp когда токен был выпущен
	IssuedAt int64 `json:"iat"`
	// Unix timestamp когда токен истекает
	ExpiresAt int64 `json:"exp"`
}

// RefreshTokenRequest представляет запрос на обновление токена
type RefreshTokenRequest struct {
	// Существующий токен для обновления
	Token string `json:"token" binding:"required"`
}

// PasswordResetToken представляет токен сброса пароля
type PasswordResetToken struct {
	// Уникальный идентификатор токена сброса
	ID uuid.UUID `json:"id" db:"id"`
	// Идентификатор пользователя, запросившего сброс
	UserID uuid.UUID `json:"user_id" db:"user_id"`
	// Адрес электронной почты, связанный со сбросом
	Email string `json:"email" db:"email"`
	// 6-значный код подтверждения (никогда не передается в JSON)
	Code string `json:"-" db:"code"`
	// Время истечения срока действия токена сброса
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	// Использован ли токен
	Used bool `json:"used" db:"used"`
	// Время создания токена сброса
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// RequestPasswordResetRequest представляет запрос на сброс пароля
type RequestPasswordResetRequest struct {
	// Адрес электронной почты для отправки кода сброса
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

// RequestPasswordResetResponse представляет ответ после запроса на сброс пароля
type RequestPasswordResetResponse struct {
	// Статусное сообщение
	Message string `json:"message" example:"Password reset code sent to your email"`
	// Идентификатор сгенерированного токена сброса
	TokenID uuid.UUID `json:"token_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// VerifyResetCodeRequest представляет запрос на проверку кода сброса
type VerifyResetCodeRequest struct {
	// Идентификатор токена сброса для проверки
	TokenID uuid.UUID `json:"token_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// 6-значный код подтверждения
	Code string `json:"code" binding:"required,len=6" example:"123456"`
}

// VerifyResetCodeResponse представляет ответ после проверки кода сброса
type VerifyResetCodeResponse struct {
	// Валиден ли код
	Valid bool `json:"valid" example:"true"`
	// Дополнительная информация о проверке
	Message string `json:"message" example:"Code verified successfully"`
}

// ResetPasswordRequest представляет запрос на сброс пароля с кодом
type ResetPasswordRequest struct {
	// Идентификатор токена сброса
	TokenID uuid.UUID `json:"token_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// 6-значный код подтверждения
	Code string `json:"code" binding:"required,len=6" example:"123456"`
	// Новый пароль (минимум 8 символов)
	NewPassword string `json:"new_password" binding:"required,min=8" example:"newpassword123"`
}

// ResetPasswordResponse представляет ответ после сброса пароля
type ResetPasswordResponse struct {
	// Статусное сообщение
	Message string `json:"message" example:"Password reset successfully"`
}

// MinimalUserInfo - минимальная информация о пользователе для LoginResponse (по API.md)
type MinimalUserInfo struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username  string    `json:"username" example:"admin"`
	Email     string    `json:"email" example:"admin@recontext.online"`
	Role      UserRole  `json:"role" example:"admin"`
	FirstName string    `json:"first_name" example:"John"`
	LastName  string    `json:"last_name" example:"Doe"`
	Bio       string    `json:"bio,omitempty" example:"Software developer"`
}

// MinimalLoginResponse - минимальный ответ логина без sensitive полей
type MinimalLoginResponse struct {
	Token     string          `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt time.Time       `json:"expires_at"`
	User      MinimalUserInfo `json:"user"`
}

package models

import (
	"time"

	"github.com/google/uuid"
)

// User представляет пользователя в системе
type User struct {
	// ID - уникальный идентификатор пользователя
	ID           uuid.UUID       `json:"id" db:"id"`
	// Username - имя пользователя для входа
	Username     string          `json:"username" db:"username"`
	// Email - адрес электронной почты пользователя
	Email        string          `json:"email" db:"email"`
	// Password - хешированный пароль (никогда не передается в JSON)
	Password     string          `json:"-" db:"password"`
	// Role - роль пользователя в системе
	Role         UserRole        `json:"role" db:"role"`
	// FirstName - имя пользователя
	FirstName    string          `json:"first_name,omitempty" db:"first_name"`
	// LastName - фамилия пользователя
	LastName     string          `json:"last_name,omitempty" db:"last_name"`
	// Phone - номер телефона пользователя
	Phone        string          `json:"phone,omitempty" db:"phone"`
	// Bio - биография пользователя
	Bio          string          `json:"bio,omitempty" db:"bio"`
	// Avatar - URL аватара или base64 изображение
	Avatar       string          `json:"avatar,omitempty" db:"avatar"`
	// DepartmentID - идентификатор отдела, к которому принадлежит пользователь
	DepartmentID *uuid.UUID      `json:"department_id,omitempty" db:"department_id"`
	// Groups - идентификаторы групп, к которым принадлежит пользователь
	Groups       []uuid.UUID     `json:"groups,omitempty" db:"groups"`
	// Permissions - специфические разрешения пользователя
	Permissions  UserPermissions `json:"permissions" db:"permissions"`
	// Language - предпочитаемый язык пользователя (ru, en и т.д.)
	Language     string          `json:"language" db:"language"`
	// IsActive - активен ли аккаунт пользователя
	IsActive     bool            `json:"is_active" db:"is_active"`
	// LastLogin - время последнего входа пользователя
	LastLogin    *time.Time      `json:"last_login,omitempty" db:"last_login"`
	// CreatedAt - время создания аккаунта пользователя
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	// UpdatedAt - время последнего обновления аккаунта пользователя
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

// UserPermissions представляет специфические разрешения пользователя
type UserPermissions struct {
	// CanScheduleMeetings - разрешение на планирование видеовстреч
	CanScheduleMeetings bool `json:"can_schedule_meetings"`
	// CanManageDepartment - разрешение на управление отделом
	CanManageDepartment bool `json:"can_manage_department"`
	// CanApproveRecordings - разрешение на утверждение записей
	CanApproveRecordings bool `json:"can_approve_recordings"`
}

// UserRole представляет роли пользователей
type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleUser     UserRole = "user"
	RoleOperator UserRole = "operator"
	RoleService  UserRole = "service" // For service-to-service authentication
)

// LoginRequest представляет запрос на вход в систему
type LoginRequest struct {
	// Username - имя пользователя или email для входа
	Username string `json:"username" binding:"required" example:"admin@recontext.online"`
	// Password - пароль пользователя
	Password string `json:"password" binding:"required" example:"password123"`
}

// LoginResponse представляет ответ на запрос входа
type LoginResponse struct {
	// Token - JWT токен аутентификации
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	// ExpiresAt - время истечения срока действия токена
	ExpiresAt time.Time `json:"expires_at"`
	// User - информация об аутентифицированном пользователе
	User      UserInfo  `json:"user"`
}

// UserInfo представляет публичную информацию о пользователе
type UserInfo struct {
	// ID - уникальный идентификатор пользователя
	ID           uuid.UUID       `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Username - имя пользователя для входа
	Username     string          `json:"username" example:"admin"`
	// Email - адрес электронной почты пользователя
	Email        string          `json:"email" example:"admin@recontext.online"`
	// Role - роль пользователя в системе
	Role         UserRole        `json:"role" example:"admin"`
	// FirstName - имя пользователя
	FirstName    string          `json:"first_name,omitempty" example:"John"`
	// LastName - фамилия пользователя
	LastName     string          `json:"last_name,omitempty" example:"Doe"`
	// Phone - номер телефона пользователя
	Phone        string          `json:"phone,omitempty" example:"+1234567890"`
	// Bio - биография пользователя
	Bio          string          `json:"bio,omitempty" example:"Software developer"`
	// Avatar - URL аватара или base64 изображение
	Avatar       string          `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
	// DepartmentID - идентификатор отдела, к которому принадлежит пользователь
	DepartmentID *uuid.UUID      `json:"department_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Permissions - специфические разрешения пользователя
	Permissions  UserPermissions `json:"permissions"`
	// Language - предпочитаемый язык пользователя
	Language     string          `json:"language" example:"en"`
}

// RegisterRequest представляет запрос на регистрацию
type RegisterRequest struct {
	// Username - желаемое имя пользователя для нового аккаунта
	Username string `json:"username" binding:"required" example:"newuser"`
	// Email - адрес электронной почты для нового аккаунта
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	// Password - желаемый пароль (минимум 8 символов)
	Password string `json:"password" binding:"required,min=8" example:"password123"`
	// Language - предпочитаемый язык (опционально)
	Language string `json:"language,omitempty" example:"en"`
}

// TokenClaims представляет claims JWT токена
type TokenClaims struct {
	// UserID - идентификатор пользователя, которому принадлежит токен
	UserID   uuid.UUID `json:"user_id"`
	// Username - имя пользователя
	Username string    `json:"username"`
	// Role - роль пользователя
	Role     UserRole  `json:"role"`
	// IssuedAt - Unix timestamp когда токен был выпущен
	IssuedAt int64     `json:"iat"`
	// ExpiresAt - Unix timestamp когда токен истекает
	ExpiresAt int64    `json:"exp"`
}

// RefreshTokenRequest представляет запрос на обновление токена
type RefreshTokenRequest struct {
	// Token - существующий токен для обновления
	Token string `json:"token" binding:"required"`
}

// PasswordResetToken представляет токен сброса пароля
type PasswordResetToken struct {
	// ID - уникальный идентификатор токена сброса
	ID        uuid.UUID `json:"id" db:"id"`
	// UserID - идентификатор пользователя, запросившего сброс
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	// Email - адрес электронной почты, связанный со сбросом
	Email     string    `json:"email" db:"email"`
	// Code - 6-значный код подтверждения (никогда не передается в JSON)
	Code      string    `json:"-" db:"code"`
	// ExpiresAt - время истечения срока действия токена сброса
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	// Used - использован ли токен
	Used      bool      `json:"used" db:"used"`
	// CreatedAt - время создания токена сброса
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// RequestPasswordResetRequest представляет запрос на сброс пароля
type RequestPasswordResetRequest struct {
	// Email - адрес электронной почты для отправки кода сброса
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

// RequestPasswordResetResponse представляет ответ после запроса на сброс пароля
type RequestPasswordResetResponse struct {
	// Message - статусное сообщение
	Message string    `json:"message" example:"Password reset code sent to your email"`
	// TokenID - идентификатор сгенерированного токена сброса
	TokenID uuid.UUID `json:"token_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// VerifyResetCodeRequest представляет запрос на проверку кода сброса
type VerifyResetCodeRequest struct {
	// TokenID - идентификатор токена сброса для проверки
	TokenID uuid.UUID `json:"token_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Code - 6-значный код подтверждения
	Code    string    `json:"code" binding:"required,len=6" example:"123456"`
}

// VerifyResetCodeResponse представляет ответ после проверки кода сброса
type VerifyResetCodeResponse struct {
	// Valid - валиден ли код
	Valid   bool   `json:"valid" example:"true"`
	// Message - дополнительная информация о проверке
	Message string `json:"message" example:"Code verified successfully"`
}

// ResetPasswordRequest представляет запрос на сброс пароля с кодом
type ResetPasswordRequest struct {
	// TokenID - идентификатор токена сброса
	TokenID     uuid.UUID `json:"token_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Code - 6-значный код подтверждения
	Code        string    `json:"code" binding:"required,len=6" example:"123456"`
	// NewPassword - новый пароль (минимум 8 символов)
	NewPassword string    `json:"new_password" binding:"required,min=8" example:"newpassword123"`
}

// ResetPasswordResponse представляет ответ после сброса пароля
type ResetPasswordResponse struct {
	// Message - статусное сообщение
	Message string `json:"message" example:"Password reset successfully"`
}

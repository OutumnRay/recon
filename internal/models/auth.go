package models

import "time"

// User represents a user in the system
type User struct {
	ID           string          `json:"id" db:"id"`
	Username     string          `json:"username" db:"username"`
	Email        string          `json:"email" db:"email"`
	Password     string          `json:"-" db:"password"` // Never expose password in JSON
	Role         UserRole        `json:"role" db:"role"`
	FirstName    string          `json:"first_name,omitempty" db:"first_name"` // User's first name
	LastName     string          `json:"last_name,omitempty" db:"last_name"`   // User's last name
	Phone        string          `json:"phone,omitempty" db:"phone"`           // User's phone number
	Bio          string          `json:"bio,omitempty" db:"bio"`               // User's biography
	Avatar       string          `json:"avatar,omitempty" db:"avatar"`         // Avatar URL or base64
	DepartmentID *string         `json:"department_id,omitempty" db:"department_id"` // Department user belongs to
	Groups       []string        `json:"groups,omitempty" db:"groups"`                // Group IDs user belongs to
	Permissions  UserPermissions `json:"permissions" db:"permissions"`                // User-specific permissions
	Language     string          `json:"language" db:"language"`                      // User's preferred language (ru, en, etc.)
	IsActive     bool            `json:"is_active" db:"is_active"`
	LastLogin    *time.Time      `json:"last_login,omitempty" db:"last_login"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

// UserPermissions represents user-specific permissions
type UserPermissions struct {
	CanScheduleMeetings bool `json:"can_schedule_meetings"` // Permission to schedule video meetings
	CanManageDepartment bool `json:"can_manage_department"` // Permission to manage department
	CanApproveRecordings bool `json:"can_approve_recordings"` // Permission to approve recordings
}

// UserRole represents user roles
type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleUser     UserRole = "user"
	RoleOperator UserRole = "operator"
	RoleService  UserRole = "service" // For service-to-service authentication
)

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin@recontext.online"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

// UserInfo represents public user information
type UserInfo struct {
	ID           string          `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username     string          `json:"username" example:"admin"`
	Email        string          `json:"email" example:"admin@recontext.online"`
	Role         UserRole        `json:"role" example:"admin"`
	FirstName    string          `json:"first_name,omitempty" example:"John"`
	LastName     string          `json:"last_name,omitempty" example:"Doe"`
	Phone        string          `json:"phone,omitempty" example:"+1234567890"`
	Bio          string          `json:"bio,omitempty" example:"Software developer"`
	Avatar       string          `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
	DepartmentID *string         `json:"department_id,omitempty" example:"dept-001"`
	Permissions  UserPermissions `json:"permissions"`
	Language     string          `json:"language" example:"en"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required" example:"newuser"`
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"password123"`
	Language string `json:"language,omitempty" example:"en"` // Optional language preference
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Role     UserRole `json:"role"`
	IssuedAt int64    `json:"iat"`
	ExpiresAt int64   `json:"exp"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Email     string    `json:"email" db:"email"`
	Code      string    `json:"-" db:"code"` // 6-digit code, never expose in JSON
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	Used      bool      `json:"used" db:"used"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// RequestPasswordResetRequest represents a password reset request
type RequestPasswordResetRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

// RequestPasswordResetResponse represents the response after requesting password reset
type RequestPasswordResetResponse struct {
	Message string `json:"message" example:"Password reset code sent to your email"`
	TokenID string `json:"token_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// VerifyResetCodeRequest represents a request to verify reset code
type VerifyResetCodeRequest struct {
	TokenID string `json:"token_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Code    string `json:"code" binding:"required,len=6" example:"123456"`
}

// VerifyResetCodeResponse represents the response after verifying reset code
type VerifyResetCodeResponse struct {
	Valid   bool   `json:"valid" example:"true"`
	Message string `json:"message" example:"Code verified successfully"`
}

// ResetPasswordRequest represents a request to reset password with code
type ResetPasswordRequest struct {
	TokenID     string `json:"token_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Code        string `json:"code" binding:"required,len=6" example:"123456"`
	NewPassword string `json:"new_password" binding:"required,min=8" example:"newpassword123"`
}

// ResetPasswordResponse represents the response after resetting password
type ResetPasswordResponse struct {
	Message string `json:"message" example:"Password reset successfully"`
}

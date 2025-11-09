package models

import "time"

// User represents a user in the system
type User struct {
	ID           string         `json:"id" db:"id"`
	Username     string         `json:"username" db:"username"`
	Email        string         `json:"email" db:"email"`
	Password     string         `json:"-" db:"password"` // Never expose password in JSON
	Role         UserRole       `json:"role" db:"role"`
	DepartmentID *string        `json:"department_id,omitempty" db:"department_id"` // Department user belongs to
	Groups       []string       `json:"groups,omitempty" db:"groups"`                // Group IDs user belongs to
	Permissions  UserPermissions `json:"permissions" db:"permissions"`                // User-specific permissions
	IsActive     bool           `json:"is_active" db:"is_active"`
	LastLogin    *time.Time     `json:"last_login,omitempty" db:"last_login"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
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
	DepartmentID *string         `json:"department_id,omitempty" example:"dept-001"`
	Permissions  UserPermissions `json:"permissions"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required" example:"newuser"`
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"password123"`
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

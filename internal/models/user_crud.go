package models

// UpdateUserRequest represents a request to update user information
type UpdateUserRequest struct {
	Email        string           `json:"email,omitempty" example:"newemail@example.com"`
	Password     string           `json:"password,omitempty" example:"newpassword123"`
	Role         UserRole         `json:"role,omitempty" example:"operator"`
	DepartmentID *string          `json:"department_id,omitempty" example:"dept-001"`
	Groups       []string         `json:"groups,omitempty"`
	Permissions  *UserPermissions `json:"permissions,omitempty"`
	Language     string           `json:"language,omitempty" example:"en"`
	IsActive     *bool            `json:"is_active,omitempty" example:"true"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"oldpass123"`
	NewPassword string `json:"new_password" binding:"required,min=8" example:"newpass123"`
}

// ListUsersRequest represents parameters for listing users
type ListUsersRequest struct {
	Page         int    `json:"page" form:"page" example:"1"`
	PageSize     int    `json:"page_size" form:"page_size" example:"20"`
	Role         string `json:"role" form:"role" example:"user"`
	DepartmentID string `json:"department_id" form:"department_id" example:"dept-001"`
	GroupID      string `json:"group_id" form:"group_id" example:"group-001"`
	IsActive     *bool  `json:"is_active" form:"is_active" example:"true"`
}

// ListUsersResponse represents a paginated list of users
type ListUsersResponse struct {
	Items    []UserInfo `json:"items"`
	Total    int        `json:"total"`
	Offset   int        `json:"offset"`
	PageSize int        `json:"page_size"`
}

// DeleteUserRequest represents a user deletion request
type DeleteUserRequest struct {
	UserID string `json:"user_id" binding:"required" example:"user-123"`
	Reason string `json:"reason,omitempty" example:"Account requested deletion"`
}

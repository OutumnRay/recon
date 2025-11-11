package database

import (
	"database/sql"
	"fmt"
	"time"

	"Recontext.online/internal/models"
	"github.com/lib/pq"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (id, username, email, password, role, groups, language, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(
		query,
		user.ID,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		pq.Array(user.Groups),
		user.Language,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id string) (*models.User, error) {
	query := `
		SELECT id, username, email, password, role, first_name, last_name, phone, bio, avatar_url,
		       groups, language, is_active, last_login, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &models.User{}
	var lastLogin sql.NullTime
	var firstName, lastName, phone, bio, avatarURL sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&firstName,
		&lastName,
		&phone,
		&bio,
		&avatarURL,
		pq.Array(&user.Groups),
		&user.Language,
		&user.IsActive,
		&lastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if firstName.Valid {
		user.FirstName = firstName.String
	}
	if lastName.Valid {
		user.LastName = lastName.String
	}
	if phone.Valid {
		user.Phone = phone.String
	}
	if bio.Valid {
		user.Bio = bio.String
	}
	if avatarURL.Valid {
		user.Avatar = avatarURL.String
	}
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password, role, groups, language, is_active, last_login, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	user := &models.User{}
	var lastLogin sql.NullTime

	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		pq.Array(&user.Groups),
		&user.Language,
		&user.IsActive,
		&lastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return user, nil
}

// GetByEmail retrieves a user by email using GORM
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var dbUser User
	err := r.db.Where("email = ?", email).First(&dbUser).Error

	if err != nil {
		if err.Error() == "record not found" {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Convert database model to API model
	user := &models.User{
		ID:           dbUser.ID,
		Username:     dbUser.Username,
		Email:        dbUser.Email,
		Password:     dbUser.Password,
		Role:         models.UserRole(dbUser.Role),
		Groups:       dbUser.Groups,
		Language:     dbUser.Language,
		IsActive:     dbUser.IsActive,
		LastLogin:    dbUser.LastLogin,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		DepartmentID: dbUser.DepartmentID,
	}

	// Convert pointer fields to string fields
	if dbUser.FirstName != nil {
		user.FirstName = *dbUser.FirstName
	}
	if dbUser.LastName != nil {
		user.LastName = *dbUser.LastName
	}
	if dbUser.Phone != nil {
		user.Phone = *dbUser.Phone
	}
	if dbUser.Bio != nil {
		user.Bio = *dbUser.Bio
	}
	if dbUser.AvatarURL != nil {
		user.Avatar = *dbUser.AvatarURL
	}

	return user, nil
}

// List retrieves all users with optional filters
func (r *UserRepository) List(role string, isActive *bool) ([]*models.User, error) {
	query := `
		SELECT id, username, email, password, role, groups, language, is_active, last_login, created_at, updated_at
		FROM users
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if role != "" {
		query += fmt.Sprintf(" AND role = $%d", argIdx)
		args = append(args, role)
		argIdx++
	}

	if isActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argIdx)
		args = append(args, *isActive)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := []*models.User{}
	for rows.Next() {
		user := &models.User{}
		var lastLogin sql.NullTime

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.Role,
			pq.Array(&user.Groups),
			&user.Language,
			&user.IsActive,
			&lastLogin,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		if lastLogin.Valid {
			user.LastLogin = &lastLogin.Time
		}

		users = append(users, user)
	}

	return users, nil
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users
		SET email = $2, role = $3, first_name = $4, last_name = $5, phone = $6, bio = $7, avatar_url = $8,
		    groups = $9, language = $10, is_active = $11, updated_at = $12
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	result, err := r.db.Exec(
		query,
		user.ID,
		user.Email,
		user.Role,
		user.FirstName,
		user.LastName,
		user.Phone,
		user.Bio,
		user.Avatar,
		pq.Array(user.Groups),
		user.Language,
		user.IsActive,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdateAvatar updates a user's avatar
func (r *UserRepository) UpdateAvatar(userID, avatarURL string) error {
	query := `
		UPDATE users
		SET avatar_url = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.Exec(query, userID, avatarURL, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update avatar: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(userID, hashedPassword string) error {
	query := `
		UPDATE users
		SET password = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.Exec(query, userID, hashedPassword, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdateLastLogin updates the last login timestamp
func (r *UserRepository) UpdateLastLogin(userID string) error {
	query := `
		UPDATE users
		SET last_login = $2
		WHERE id = $1
	`

	_, err := r.db.Exec(query, userID, time.Now())
	return err
}

// Delete deletes a user
func (r *UserRepository) Delete(userID string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UsernameExists checks if a username already exists
func (r *UserRepository) UsernameExists(username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	var exists bool
	err := r.db.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return exists, nil
}

// EmailExists checks if an email already exists
func (r *UserRepository) EmailExists(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

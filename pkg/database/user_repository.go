package database

import (
	"fmt"
	"time"

	"Recontext.online/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	// Convert API model to DB model
	dbUser := &User{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		Password:       user.Password,
		Role:           string(user.Role),
		OrganizationID: user.OrganizationID,
		Language:       user.Language,
		IsActive:       user.IsActive,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}

	// Convert UUID slice to string slice for groups
	if len(user.Groups) > 0 {
		dbUser.Groups = make([]string, len(user.Groups))
		for i, g := range user.Groups {
			dbUser.Groups[i] = g.String()
		}
	}

	// Handle optional fields
	if user.FirstName != "" {
		dbUser.FirstName = &user.FirstName
	}
	if user.LastName != "" {
		dbUser.LastName = &user.LastName
	}
	if user.Phone != "" {
		dbUser.Phone = &user.Phone
	}
	if user.Bio != "" {
		dbUser.Bio = &user.Bio
	}
	if user.Avatar != "" {
		dbUser.AvatarURL = &user.Avatar
	}
	if user.DepartmentID != nil {
		dbUser.DepartmentID = user.DepartmentID
	}

	if err := r.db.DB.Create(dbUser).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var dbUser User
	err := r.db.DB.Where("id = ?", id).First(&dbUser).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Convert database model to API model
	user := &models.User{
		ID:             dbUser.ID,
		Username:       dbUser.Username,
		Email:          dbUser.Email,
		Password:       dbUser.Password,
		Role:           models.UserRole(dbUser.Role),
		OrganizationID: dbUser.OrganizationID,
		Language:       dbUser.Language,
		IsActive:       dbUser.IsActive,
		LastLogin:      dbUser.LastLogin,
		CreatedAt:      dbUser.CreatedAt,
		UpdatedAt:      dbUser.UpdatedAt,
	}

	// Convert string slice to UUID slice for groups
	if len(dbUser.Groups) > 0 {
		user.Groups = make([]uuid.UUID, len(dbUser.Groups))
		for i, g := range dbUser.Groups {
			user.Groups[i], _ = uuid.Parse(g)
		}
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
	if dbUser.DepartmentID != nil {
		user.DepartmentID = dbUser.DepartmentID
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var dbUser User
	err := r.db.DB.Where("username = ?", username).First(&dbUser).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Convert database model to API model
	user := &models.User{
		ID:             dbUser.ID,
		Username:       dbUser.Username,
		Email:          dbUser.Email,
		Password:       dbUser.Password,
		Role:           models.UserRole(dbUser.Role),
		OrganizationID: dbUser.OrganizationID,
		Language:       dbUser.Language,
		IsActive:       dbUser.IsActive,
		LastLogin:      dbUser.LastLogin,
		CreatedAt:      dbUser.CreatedAt,
		UpdatedAt:      dbUser.UpdatedAt,
	}

	// Convert string slice to UUID slice for groups
	if len(dbUser.Groups) > 0 {
		user.Groups = make([]uuid.UUID, len(dbUser.Groups))
		for i, g := range dbUser.Groups {
			user.Groups[i], _ = uuid.Parse(g)
		}
	}

	if dbUser.DepartmentID != nil {
		user.DepartmentID = dbUser.DepartmentID
	}

	return user, nil
}

// GetByEmail retrieves a user by email using GORM
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var dbUser User
	err := r.db.DB.Where("email = ?", email).First(&dbUser).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Convert database model to API model
	user := &models.User{
		ID:             dbUser.ID,
		Username:       dbUser.Username,
		Email:          dbUser.Email,
		Password:       dbUser.Password,
		Role:           models.UserRole(dbUser.Role),
		OrganizationID: dbUser.OrganizationID,
		Language:       dbUser.Language,
		IsActive:       dbUser.IsActive,
		LastLogin:      dbUser.LastLogin,
		CreatedAt:      dbUser.CreatedAt,
		UpdatedAt:      dbUser.UpdatedAt,
	}

	// Convert string slice to UUID slice for groups
	if len(dbUser.Groups) > 0 {
		user.Groups = make([]uuid.UUID, len(dbUser.Groups))
		for i, g := range dbUser.Groups {
			user.Groups[i], _ = uuid.Parse(g)
		}
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
	if dbUser.DepartmentID != nil {
		user.DepartmentID = dbUser.DepartmentID
	}

	return user, nil
}

// List retrieves all users with optional filters
func (r *UserRepository) List(role string, isActive *bool) ([]*models.User, error) {
	var dbUsers []User
	query := r.db.DB.Model(&User{})

	if role != "" {
		query = query.Where("role = ?", role)
	}

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Order("created_at DESC").Find(&dbUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	users := make([]*models.User, 0, len(dbUsers))
	for _, dbUser := range dbUsers {
		user := &models.User{
			ID:        dbUser.ID,
			Username:  dbUser.Username,
			Email:     dbUser.Email,
			Password:  dbUser.Password,
			Role:      models.UserRole(dbUser.Role),
			Language:  dbUser.Language,
			IsActive:  dbUser.IsActive,
			LastLogin: dbUser.LastLogin,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
		}

		// Convert string slice to UUID slice for groups
		if len(dbUser.Groups) > 0 {
			user.Groups = make([]uuid.UUID, len(dbUser.Groups))
			for i, g := range dbUser.Groups {
				user.Groups[i], _ = uuid.Parse(g)
			}
		}

		if dbUser.DepartmentID != nil {
			user.DepartmentID = dbUser.DepartmentID
		}

		users = append(users, user)
	}

	return users, nil
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	user.UpdatedAt = time.Now()

	// Convert API model to DB model updates
	updates := map[string]interface{}{
		"email":           user.Email,
		"password":        user.Password,
		"role":            string(user.Role),
		"organization_id": user.OrganizationID,
		"language":        user.Language,
		"is_active":       user.IsActive,
		"updated_at":      user.UpdatedAt,
	}

	// Convert UUID slice to string slice for groups
	if user.Groups != nil {
		groupStrings := make([]string, len(user.Groups))
		for i, g := range user.Groups {
			groupStrings[i] = g.String()
		}
		updates["groups"] = pq.StringArray(groupStrings)
	}

	// Handle optional fields
	if user.FirstName != "" {
		updates["first_name"] = user.FirstName
	}
	if user.LastName != "" {
		updates["last_name"] = user.LastName
	}
	if user.Phone != "" {
		updates["phone"] = user.Phone
	}
	if user.Bio != "" {
		updates["bio"] = user.Bio
	}
	if user.Avatar != "" {
		updates["avatar_url"] = user.Avatar
	}
	if user.DepartmentID != nil {
		updates["department_id"] = user.DepartmentID
	}

	result := r.db.DB.Model(&User{}).Where("id = ?", user.ID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdateAvatar updates a user's avatar
func (r *UserRepository) UpdateAvatar(userID uuid.UUID, avatarURL string) error {
	result := r.db.DB.Model(&User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"avatar_url": avatarURL,
		"updated_at": time.Now(),
	})

	if result.Error != nil {
		return fmt.Errorf("failed to update avatar: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(userID uuid.UUID, hashedPassword string) error {
	result := r.db.DB.Model(&User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"password":   hashedPassword,
		"updated_at": time.Now(),
	})

	if result.Error != nil {
		return fmt.Errorf("failed to update password: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdateLastLogin updates the last login timestamp
func (r *UserRepository) UpdateLastLogin(userID uuid.UUID) error {
	return r.db.DB.Model(&User{}).Where("id = ?", userID).Update("last_login", time.Now()).Error
}

// Delete deletes a user
func (r *UserRepository) Delete(userID uuid.UUID) error {
	result := r.db.DB.Where("id = ?", userID).Delete(&User{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UsernameExists checks if a username already exists
func (r *UserRepository) UsernameExists(username string) (bool, error) {
	var count int64
	err := r.db.DB.Model(&User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return count > 0, nil
}

// EmailExists checks if an email already exists
func (r *UserRepository) EmailExists(email string) (bool, error) {
	var count int64
	err := r.db.DB.Model(&User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}

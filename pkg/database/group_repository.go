package database

import (
	"encoding/json"
	"fmt"
	"time"

	"Recontext.online/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GroupRepository struct {
	db *DB
}

func NewGroupRepository(db *DB) *GroupRepository {
	return &GroupRepository{db: db}
}

// Create creates a new group
func (r *GroupRepository) Create(group *models.UserGroup) error {
	permissionsJSON, err := json.Marshal(group.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	dbGroup := &Group{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		Permissions: string(permissionsJSON),
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}

	if err := r.db.DB.Create(dbGroup).Error; err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	return nil
}

// GetByID retrieves a group by ID
func (r *GroupRepository) GetByID(id uuid.UUID) (*models.UserGroup, error) {
	var dbGroup Group
	err := r.db.DB.Where("id = ?", id).First(&dbGroup).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("group not found")
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	group := &models.UserGroup{
		ID:          dbGroup.ID,
		Name:        dbGroup.Name,
		Description: dbGroup.Description,
		CreatedAt:   dbGroup.CreatedAt,
		UpdatedAt:   dbGroup.UpdatedAt,
	}

	if err := json.Unmarshal([]byte(dbGroup.Permissions), &group.Permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	return group, nil
}

// GetByName retrieves a group by name
func (r *GroupRepository) GetByName(name string) (*models.UserGroup, error) {
	var dbGroup Group
	err := r.db.DB.Where("name = ?", name).First(&dbGroup).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("group not found")
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	group := &models.UserGroup{
		ID:          dbGroup.ID,
		Name:        dbGroup.Name,
		Description: dbGroup.Description,
		CreatedAt:   dbGroup.CreatedAt,
		UpdatedAt:   dbGroup.UpdatedAt,
	}

	if err := json.Unmarshal([]byte(dbGroup.Permissions), &group.Permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	return group, nil
}

// List retrieves all groups
func (r *GroupRepository) List() ([]*models.UserGroup, error) {
	var dbGroups []Group
	if err := r.db.DB.Order("created_at DESC").Find(&dbGroups).Error; err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	groups := make([]*models.UserGroup, 0, len(dbGroups))
	for _, dbGroup := range dbGroups {
		group := &models.UserGroup{
			ID:          dbGroup.ID,
			Name:        dbGroup.Name,
			Description: dbGroup.Description,
			CreatedAt:   dbGroup.CreatedAt,
			UpdatedAt:   dbGroup.UpdatedAt,
		}

		if err := json.Unmarshal([]byte(dbGroup.Permissions), &group.Permissions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// Update updates a group
func (r *GroupRepository) Update(group *models.UserGroup) error {
	permissionsJSON, err := json.Marshal(group.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	group.UpdatedAt = time.Now()

	updates := map[string]interface{}{
		"name":        group.Name,
		"description": group.Description,
		"permissions": string(permissionsJSON),
		"updated_at":  group.UpdatedAt,
	}

	result := r.db.DB.Model(&Group{}).Where("id = ?", group.ID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update group: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("group not found")
	}

	return nil
}

// Delete deletes a group
func (r *GroupRepository) Delete(groupID uuid.UUID) error {
	result := r.db.DB.Where("id = ?", groupID).Delete(&Group{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete group: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("group not found")
	}

	return nil
}

// AddUserToGroup adds a user to a group
func (r *GroupRepository) AddUserToGroup(userID, groupID uuid.UUID, addedBy *uuid.UUID) error {
	membership := &GroupMembership{
		UserID:  userID,
		GroupID: groupID,
		AddedAt: time.Now(),
	}
	if addedBy != nil {
		membership.AddedBy = addedBy
	}

	// Use FirstOrCreate to handle ON CONFLICT
	if err := r.db.DB.Where(GroupMembership{UserID: userID, GroupID: groupID}).FirstOrCreate(membership).Error; err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	// Also update user's groups array using GORM
	groupIDStr := groupID.String()
	result := r.db.DB.Model(&User{}).
		Where("id = ? AND NOT (? = ANY(groups))", userID, groupIDStr).
		Updates(map[string]interface{}{
			"groups":     gorm.Expr("array_append(groups, ?)", groupIDStr),
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update user groups: %w", result.Error)
	}

	return nil
}

// RemoveUserFromGroup removes a user from a group
func (r *GroupRepository) RemoveUserFromGroup(userID, groupID uuid.UUID) error {
	if err := r.db.DB.Where("user_id = ? AND group_id = ?", userID, groupID).Delete(&GroupMembership{}).Error; err != nil {
		return fmt.Errorf("failed to remove user from group: %w", err)
	}

	// Also update user's groups array using GORM
	groupIDStr := groupID.String()
	result := r.db.DB.Model(&User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"groups":     gorm.Expr("array_remove(groups, ?)", groupIDStr),
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update user groups: %w", result.Error)
	}

	return nil
}

// GetGroupMembers retrieves all users in a group
func (r *GroupRepository) GetGroupMembers(groupID uuid.UUID) ([]string, error) {
	var memberships []GroupMembership
	if err := r.db.DB.Where("group_id = ?", groupID).Find(&memberships).Error; err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}

	userIDs := make([]string, 0, len(memberships))
	for _, membership := range memberships {
		userIDs = append(userIDs, membership.UserID.String())
	}

	return userIDs, nil
}

// NameExists checks if a group name already exists
func (r *GroupRepository) NameExists(name string) (bool, error) {
	var count int64
	err := r.db.DB.Model(&Group{}).Where("name = ?", name).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check group name existence: %w", err)
	}

	return count > 0, nil
}

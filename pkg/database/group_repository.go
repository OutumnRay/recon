package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"Recontext.online/internal/models"
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

	query := `
		INSERT INTO groups (id, name, description, permissions, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = r.db.Exec(
		query,
		group.ID,
		group.Name,
		group.Description,
		permissionsJSON,
		group.CreatedAt,
		group.UpdatedAt,
	)

	if err != nil{
		return fmt.Errorf("failed to create group: %w", err)
	}

	return nil
}

// GetByID retrieves a group by ID
func (r *GroupRepository) GetByID(id string) (*models.UserGroup, error) {
	query := `
		SELECT id, name, description, permissions, created_at, updated_at
		FROM groups
		WHERE id = $1
	`

	group := &models.UserGroup{}
	var permissionsJSON []byte

	err := r.db.QueryRow(query, id).Scan(
		&group.ID,
		&group.Name,
		&group.Description,
		&permissionsJSON,
		&group.CreatedAt,
		&group.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("group not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	if err := json.Unmarshal(permissionsJSON, &group.Permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	return group, nil
}

// GetByName retrieves a group by name
func (r *GroupRepository) GetByName(name string) (*models.UserGroup, error) {
	query := `
		SELECT id, name, description, permissions, created_at, updated_at
		FROM groups
		WHERE name = $1
	`

	group := &models.UserGroup{}
	var permissionsJSON []byte

	err := r.db.QueryRow(query, name).Scan(
		&group.ID,
		&group.Name,
		&group.Description,
		&permissionsJSON,
		&group.CreatedAt,
		&group.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("group not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	if err := json.Unmarshal(permissionsJSON, &group.Permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	return group, nil
}

// List retrieves all groups
func (r *GroupRepository) List() ([]*models.UserGroup, error) {
	query := `
		SELECT id, name, description, permissions, created_at, updated_at
		FROM groups
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}
	defer rows.Close()

	groups := []*models.UserGroup{}
	for rows.Next() {
		group := &models.UserGroup{}
		var permissionsJSON []byte

		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.Description,
			&permissionsJSON,
			&group.CreatedAt,
			&group.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}

		if err := json.Unmarshal(permissionsJSON, &group.Permissions); err != nil {
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

	query := `
		UPDATE groups
		SET name = $2, description = $3, permissions = $4, updated_at = $5
		WHERE id = $1
	`

	group.UpdatedAt = time.Now()

	result, err := r.db.Exec(
		query,
		group.ID,
		group.Name,
		group.Description,
		permissionsJSON,
		group.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("group not found")
	}

	return nil
}

// Delete deletes a group
func (r *GroupRepository) Delete(groupID string) error {
	query := `DELETE FROM groups WHERE id = $1`

	result, err := r.db.Exec(query, groupID)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("group not found")
	}

	return nil
}

// AddUserToGroup adds a user to a group
func (r *GroupRepository) AddUserToGroup(userID, groupID, addedBy string) error {
	query := `
		INSERT INTO group_memberships (user_id, group_id, added_at, added_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, group_id) DO NOTHING
	`

	_, err := r.db.Exec(query, userID, groupID, time.Now(), addedBy)
	if err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	// Also update user's groups array
	updateUserQuery := `
		UPDATE users
		SET groups = array_append(groups, $2), updated_at = $3
		WHERE id = $1 AND NOT ($2 = ANY(groups))
	`

	_, err = r.db.Exec(updateUserQuery, userID, groupID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update user groups: %w", err)
	}

	return nil
}

// RemoveUserFromGroup removes a user from a group
func (r *GroupRepository) RemoveUserFromGroup(userID, groupID string) error {
	query := `
		DELETE FROM group_memberships
		WHERE user_id = $1 AND group_id = $2
	`

	_, err := r.db.Exec(query, userID, groupID)
	if err != nil {
		return fmt.Errorf("failed to remove user from group: %w", err)
	}

	// Also update user's groups array
	updateUserQuery := `
		UPDATE users
		SET groups = array_remove(groups, $2), updated_at = $3
		WHERE id = $1
	`

	_, err = r.db.Exec(updateUserQuery, userID, groupID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update user groups: %w", err)
	}

	return nil
}

// GetGroupMembers retrieves all users in a group
func (r *GroupRepository) GetGroupMembers(groupID string) ([]string, error) {
	query := `
		SELECT user_id
		FROM group_memberships
		WHERE group_id = $1
	`

	rows, err := r.db.Query(query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}
	defer rows.Close()

	userIDs := []string{}
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// NameExists checks if a group name already exists
func (r *GroupRepository) NameExists(name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM groups WHERE name = $1)`

	var exists bool
	err := r.db.QueryRow(query, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check group name existence: %w", err)
	}

	return exists, nil
}

package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"Recontext.online/internal/models"
)

// DepartmentRepository handles database operations for departments
type DepartmentRepository struct {
	db *DB
}

// NewDepartmentRepository creates a new DepartmentRepository
func NewDepartmentRepository(db *DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

// Create creates a new department
func (r *DepartmentRepository) Create(dept *models.Department) error {
	// If parent_id is provided, get parent's level and path
	if dept.ParentID != nil {
		parent, err := r.GetByID(*dept.ParentID)
		if err != nil {
			return fmt.Errorf("parent department not found: %w", err)
		}
		dept.Level = parent.Level + 1
		dept.Path = parent.Path + "/" + dept.Name
	} else {
		dept.Level = 0
		dept.Path = dept.Name
	}

	query := `
		INSERT INTO departments (id, name, description, parent_id, level, path, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(
		query,
		dept.ID,
		dept.Name,
		dept.Description,
		dept.ParentID,
		dept.Level,
		dept.Path,
		dept.IsActive,
		dept.CreatedAt,
		dept.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create department: %w", err)
	}

	return nil
}

// GetByID retrieves a department by ID
func (r *DepartmentRepository) GetByID(id string) (*models.Department, error) {
	query := `
		SELECT id, name, description, parent_id, level, path, is_active, created_at, updated_at
		FROM departments
		WHERE id = $1
	`

	dept := &models.Department{}
	var parentID sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&dept.ID,
		&dept.Name,
		&dept.Description,
		&parentID,
		&dept.Level,
		&dept.Path,
		&dept.IsActive,
		&dept.CreatedAt,
		&dept.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("department not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get department: %w", err)
	}

	if parentID.Valid {
		dept.ParentID = &parentID.String
	}

	return dept, nil
}

// List retrieves departments with optional filters
func (r *DepartmentRepository) List(parentID *string, includeAll bool) ([]*models.Department, error) {
	query := `
		SELECT id, name, description, parent_id, level, path, is_active, created_at, updated_at
		FROM departments
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if parentID != nil {
		query += fmt.Sprintf(" AND parent_id = $%d", argIdx)
		args = append(args, *parentID)
		argIdx++
	}

	if !includeAll {
		query += " AND is_active = true"
	}

	query += " ORDER BY path ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list departments: %w", err)
	}
	defer rows.Close()

	departments := []*models.Department{}
	for rows.Next() {
		dept := &models.Department{}
		var parentID sql.NullString

		err := rows.Scan(
			&dept.ID,
			&dept.Name,
			&dept.Description,
			&parentID,
			&dept.Level,
			&dept.Path,
			&dept.IsActive,
			&dept.CreatedAt,
			&dept.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan department: %w", err)
		}

		if parentID.Valid {
			dept.ParentID = &parentID.String
		}

		departments = append(departments, dept)
	}

	return departments, nil
}

// GetTree builds a hierarchical tree of departments
func (r *DepartmentRepository) GetTree(rootID *string) (*models.DepartmentTreeNode, error) {
	// Get all departments
	allDepts, err := r.List(nil, false)
	if err != nil {
		return nil, err
	}

	// Create a map for quick lookups
	deptMap := make(map[string]*models.DepartmentTreeNode)
	for _, dept := range allDepts {
		node := &models.DepartmentTreeNode{
			Department: *dept,
			Children:   []*models.DepartmentTreeNode{},
		}
		deptMap[dept.ID] = node
	}

	// Build tree structure
	var root *models.DepartmentTreeNode
	for _, node := range deptMap {
		if node.ParentID == nil || (rootID != nil && node.ID == *rootID) {
			root = node
		} else if parent, ok := deptMap[*node.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		}
	}

	if root == nil && rootID != nil {
		return nil, fmt.Errorf("root department not found")
	}

	return root, nil
}

// Update updates a department
func (r *DepartmentRepository) Update(dept *models.Department) error {
	// If parent changed, recalculate level and path
	if dept.ParentID != nil {
		parent, err := r.GetByID(*dept.ParentID)
		if err != nil {
			return fmt.Errorf("parent department not found: %w", err)
		}

		// Prevent circular reference
		if strings.HasPrefix(parent.Path, dept.Path+"/") {
			return fmt.Errorf("cannot set child department as parent (circular reference)")
		}

		dept.Level = parent.Level + 1
		dept.Path = parent.Path + "/" + dept.Name
	} else {
		dept.Level = 0
		dept.Path = dept.Name
	}

	dept.UpdatedAt = time.Now()

	query := `
		UPDATE departments
		SET name = $2, description = $3, parent_id = $4, level = $5, path = $6, is_active = $7, updated_at = $8
		WHERE id = $1
	`

	result, err := r.db.Exec(
		query,
		dept.ID,
		dept.Name,
		dept.Description,
		dept.ParentID,
		dept.Level,
		dept.Path,
		dept.IsActive,
		dept.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update department: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("department not found")
	}

	// Update paths for all children
	return r.updateChildrenPaths(dept.ID, dept.Path)
}

// updateChildrenPaths recursively updates paths for all child departments
func (r *DepartmentRepository) updateChildrenPaths(parentID, parentPath string) error {
	query := `
		UPDATE departments
		SET path = $2 || '/' || name,
		    level = (
		        SELECT level + 1 FROM departments WHERE id = $1
		    ),
		    updated_at = NOW()
		WHERE parent_id = $1
		RETURNING id, path
	`

	rows, err := r.db.Query(query, parentID, parentPath)
	if err != nil {
		return fmt.Errorf("failed to update children paths: %w", err)
	}
	defer rows.Close()

	// Recursively update grandchildren
	for rows.Next() {
		var childID, childPath string
		if err := rows.Scan(&childID, &childPath); err != nil {
			return fmt.Errorf("failed to scan child: %w", err)
		}
		if err := r.updateChildrenPaths(childID, childPath); err != nil {
			return err
		}
	}

	return nil
}

// Delete soft deletes a department (marks as inactive)
func (r *DepartmentRepository) Delete(id string) error {
	query := `
		UPDATE departments
		SET is_active = false, updated_at = $2
		WHERE id = $1
	`

	result, err := r.db.Exec(query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete department: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("department not found")
	}

	return nil
}

// GetWithStats retrieves a department with statistics
func (r *DepartmentRepository) GetWithStats(id string) (*models.DepartmentWithStats, error) {
	query := `
		SELECT
			d.id, d.name, d.description, d.parent_id, d.level, d.path,
			d.is_active, d.created_at, d.updated_at,
			COUNT(DISTINCT u.id) as user_count,
			COUNT(DISTINCT c.id) as child_count
		FROM departments d
		LEFT JOIN users u ON u.department_id = d.id AND u.is_active = true
		LEFT JOIN departments c ON c.parent_id = d.id AND c.is_active = true
		WHERE d.id = $1
		GROUP BY d.id, d.name, d.description, d.parent_id, d.level, d.path, d.is_active, d.created_at, d.updated_at
	`

	stats := &models.DepartmentWithStats{}
	var parentID sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&stats.ID,
		&stats.Name,
		&stats.Description,
		&parentID,
		&stats.Level,
		&stats.Path,
		&stats.IsActive,
		&stats.CreatedAt,
		&stats.UpdatedAt,
		&stats.UserCount,
		&stats.ChildCount,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("department not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get department stats: %w", err)
	}

	if parentID.Valid {
		stats.ParentID = &parentID.String
	}

	// Get total users count (including sub-departments)
	totalUsersQuery := `
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		JOIN departments d ON u.department_id = d.id
		WHERE d.path LIKE $1 AND u.is_active = true
	`

	pathPattern := stats.Path + "%"
	err = r.db.QueryRow(totalUsersQuery, pathPattern).Scan(&stats.TotalUsersCount)
	if err != nil {
		stats.TotalUsersCount = stats.UserCount // Fallback to direct count
	}

	return stats, nil
}

// GetChildren retrieves all child departments of a given department
func (r *DepartmentRepository) GetChildren(parentID string) ([]*models.Department, error) {
	query := `
		SELECT id, name, description, parent_id, level, path, is_active, created_at, updated_at
		FROM departments
		WHERE parent_id = $1 AND is_active = true
		ORDER BY name ASC
	`

	rows, err := r.db.Query(query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}
	defer rows.Close()

	children := []*models.Department{}
	for rows.Next() {
		dept := &models.Department{}
		var parentID sql.NullString

		err := rows.Scan(
			&dept.ID,
			&dept.Name,
			&dept.Description,
			&parentID,
			&dept.Level,
			&dept.Path,
			&dept.IsActive,
			&dept.CreatedAt,
			&dept.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan department: %w", err)
		}

		if parentID.Valid {
			dept.ParentID = &parentID.String
		}

		children = append(children, dept)
	}

	return children, nil
}

// NameExists checks if a department name already exists at the same level
func (r *DepartmentRepository) NameExists(name string, parentID *string, excludeID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM departments
			WHERE name = $1
			AND ($2::VARCHAR IS NULL AND parent_id IS NULL OR parent_id = $2)
			AND id != $3
			AND is_active = true
		)
	`

	var exists bool
	err := r.db.QueryRow(query, name, parentID, excludeID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check name existence: %w", err)
	}

	return exists, nil
}

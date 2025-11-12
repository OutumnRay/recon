package database

import (
	"fmt"
	"strings"
	"time"

	"Recontext.online/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
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

	dbDept := &Department{
		ID:          dept.ID,
		Name:        dept.Name,
		Description: dept.Description,
		Level:       dept.Level,
		Path:        dept.Path,
		IsActive:    dept.IsActive,
		CreatedAt:   dept.CreatedAt,
		UpdatedAt:   dept.UpdatedAt,
	}

	if dept.ParentID != nil {
		dbDept.ParentID = dept.ParentID
	}

	if err := r.db.DB.Create(dbDept).Error; err != nil {
		return fmt.Errorf("failed to create department: %w", err)
	}

	return nil
}

// GetByID retrieves a department by ID
func (r *DepartmentRepository) GetByID(id uuid.UUID) (*models.Department, error) {
	var dbDept Department
	err := r.db.DB.Where("id = ?", id).First(&dbDept).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("department not found")
		}
		return nil, fmt.Errorf("failed to get department: %w", err)
	}

	dept := &models.Department{
		ID:          dbDept.ID,
		Name:        dbDept.Name,
		Description: dbDept.Description,
		Level:       dbDept.Level,
		Path:        dbDept.Path,
		IsActive:    dbDept.IsActive,
		CreatedAt:   dbDept.CreatedAt,
		UpdatedAt:   dbDept.UpdatedAt,
	}

	if dbDept.ParentID != nil {
		dept.ParentID = dbDept.ParentID
	}

	return dept, nil
}

// List retrieves departments with optional filters
func (r *DepartmentRepository) List(parentID *string, includeAll bool) ([]*models.Department, error) {
	var dbDepts []Department
	query := r.db.DB.Model(&Department{})

	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	}

	if !includeAll {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Order("path ASC").Find(&dbDepts).Error; err != nil {
		return nil, fmt.Errorf("failed to list departments: %w", err)
	}

	departments := make([]*models.Department, 0, len(dbDepts))
	for _, dbDept := range dbDepts {
		dept := &models.Department{
			ID:          dbDept.ID,
			Name:        dbDept.Name,
			Description: dbDept.Description,
			Level:       dbDept.Level,
			Path:        dbDept.Path,
			IsActive:    dbDept.IsActive,
			CreatedAt:   dbDept.CreatedAt,
			UpdatedAt:   dbDept.UpdatedAt,
		}
		if dbDept.ParentID != nil {
			dept.ParentID = dbDept.ParentID
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
	deptMap := make(map[uuid.UUID]*models.DepartmentTreeNode)
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
		if node.ParentID == nil || (rootID != nil && node.ID.String() == *rootID) {
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

	updates := map[string]interface{}{
		"name":        dept.Name,
		"description": dept.Description,
		"level":       dept.Level,
		"path":        dept.Path,
		"is_active":   dept.IsActive,
		"updated_at":  dept.UpdatedAt,
	}

	if dept.ParentID != nil {
		updates["parent_id"] = dept.ParentID
	} else {
		updates["parent_id"] = nil
	}

	result := r.db.DB.Model(&Department{}).Where("id = ?", dept.ID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update department: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("department not found")
	}

	// Update paths for all children
	return r.updateChildrenPaths(dept.ID, dept.Path)
}

// updateChildrenPaths recursively updates paths for all child departments
func (r *DepartmentRepository) updateChildrenPaths(parentID uuid.UUID, parentPath string) error {
	// Use raw SQL for this complex operation as GORM doesn't handle it well
	// First, get parent level
	var parentDept Department
	if err := r.db.DB.Select("level").Where("id = ?", parentID).First(&parentDept).Error; err != nil {
		return fmt.Errorf("failed to get parent level: %w", err)
	}

	// Get all children
	var children []Department
	if err := r.db.DB.Where("parent_id = ?", parentID).Find(&children).Error; err != nil {
		return fmt.Errorf("failed to get children: %w", err)
	}

	// Update each child
	type ChildResult struct {
		ID   uuid.UUID
		Path string
	}
	var childResults []ChildResult

	for _, child := range children {
		newPath := parentPath + "/" + child.Name
		newLevel := parentDept.Level + 1

		if err := r.db.DB.Model(&child).Updates(map[string]interface{}{
			"path":       newPath,
			"level":      newLevel,
			"updated_at": time.Now(),
		}).Error; err != nil {
			return fmt.Errorf("failed to update child path: %w", err)
		}

		childResults = append(childResults, ChildResult{
			ID:   child.ID,
			Path: newPath,
		})
	}

	// Recursively update grandchildren
	for _, child := range childResults {
		if err := r.updateChildrenPaths(child.ID, child.Path); err != nil {
			return err
		}
	}

	return nil
}

// Delete soft deletes a department (marks as inactive)
func (r *DepartmentRepository) Delete(id uuid.UUID) error {
	updates := map[string]interface{}{
		"is_active":  false,
		"updated_at": time.Now(),
	}

	result := r.db.DB.Model(&Department{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to delete department: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("department not found")
	}

	return nil
}

// GetWithStats retrieves a department with statistics
func (r *DepartmentRepository) GetWithStats(id uuid.UUID) (*models.DepartmentWithStats, error) {
	stats := &models.DepartmentWithStats{}

	err := r.db.DB.Table("departments d").
		Select(`
			d.id, d.name, d.description, d.parent_id, d.level, d.path,
			d.is_active, d.created_at, d.updated_at,
			COUNT(DISTINCT u.id) as user_count,
			COUNT(DISTINCT c.id) as child_count
		`).
		Joins("LEFT JOIN users u ON u.department_id = d.id AND u.is_active = true").
		Joins("LEFT JOIN departments c ON c.parent_id = d.id AND c.is_active = true").
		Where("d.id = ?", id).
		Group("d.id, d.name, d.description, d.parent_id, d.level, d.path, d.is_active, d.created_at, d.updated_at").
		Scan(stats).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("department not found")
		}
		return nil, fmt.Errorf("failed to get department stats: %w", err)
	}

	// Get total users count (including sub-departments)
	pathPattern := stats.Path + "%"
	err = r.db.DB.Table("users u").
		Select("COUNT(DISTINCT u.id)").
		Joins("JOIN departments d ON u.department_id = d.id").
		Where("d.path LIKE ? AND u.is_active = true", pathPattern).
		Scan(&stats.TotalUsersCount).Error

	if err != nil {
		stats.TotalUsersCount = stats.UserCount // Fallback to direct count
	}

	return stats, nil
}

// GetChildren retrieves all child departments of a given department
func (r *DepartmentRepository) GetChildren(parentID uuid.UUID) ([]*models.Department, error) {
	var dbDepts []Department
	if err := r.db.DB.Where("parent_id = ? AND is_active = ?", parentID, true).Order("name ASC").Find(&dbDepts).Error; err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}

	children := make([]*models.Department, 0, len(dbDepts))
	for _, dbDept := range dbDepts {
		dept := &models.Department{
			ID:          dbDept.ID,
			Name:        dbDept.Name,
			Description: dbDept.Description,
			Level:       dbDept.Level,
			Path:        dbDept.Path,
			IsActive:    dbDept.IsActive,
			CreatedAt:   dbDept.CreatedAt,
			UpdatedAt:   dbDept.UpdatedAt,
		}
		if dbDept.ParentID != nil {
			dept.ParentID = dbDept.ParentID
		}
		children = append(children, dept)
	}

	return children, nil
}

// NameExists checks if a department name already exists at the same level
func (r *DepartmentRepository) NameExists(name string, parentID *string, excludeID string) (bool, error) {
	var count int64
	query := r.db.DB.Model(&Department{}).Where("name = ? AND id != ? AND is_active = ?", name, excludeID, true)

	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check name existence: %w", err)
	}

	return count > 0, nil
}

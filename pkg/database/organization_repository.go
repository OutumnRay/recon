package database

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrganizationRepository handles organization-related database operations
type OrganizationRepository struct {
	db *DB
}

// NewOrganizationRepository creates a new organization repository
func NewOrganizationRepository(db *DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// GetAllOrganizations retrieves all organizations
func (r *OrganizationRepository) GetAllOrganizations() ([]Organization, error) {
	var organizations []Organization
	if err := r.db.Find(&organizations).Error; err != nil {
		return nil, fmt.Errorf("failed to get organizations: %w", err)
	}
	return organizations, nil
}

// GetOrganizationByID retrieves an organization by ID
func (r *OrganizationRepository) GetOrganizationByID(id uuid.UUID) (*Organization, error) {
	var organization Organization
	if err := r.db.First(&organization, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	return &organization, nil
}

// GetOrganizationByName retrieves an organization by name
func (r *OrganizationRepository) GetOrganizationByName(name string) (*Organization, error) {
	var organization Organization
	if err := r.db.First(&organization, "name = ?", name).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	return &organization, nil
}

// CreateOrganization creates a new organization
func (r *OrganizationRepository) CreateOrganization(org *Organization) error {
	// Check if organization with same name already exists
	var existingOrg Organization
	if err := r.db.Where("name = ?", org.Name).First(&existingOrg).Error; err == nil {
		return fmt.Errorf("organization with name '%s' already exists", org.Name)
	}

	// Set defaults
	org.CreatedAt = time.Now()
	org.UpdatedAt = time.Now()
	if org.Settings == "" {
		org.Settings = "{}"
	}

	if err := r.db.Create(org).Error; err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}
	return nil
}

// UpdateOrganization updates an existing organization
func (r *OrganizationRepository) UpdateOrganization(org *Organization) error {
	// Check if organization exists
	var existingOrg Organization
	if err := r.db.First(&existingOrg, "id = ?", org.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("organization not found")
		}
		return fmt.Errorf("failed to get organization: %w", err)
	}

	// Check if name is being changed to an existing name
	if org.Name != existingOrg.Name {
		var duplicateOrg Organization
		if err := r.db.Where("name = ? AND id != ?", org.Name, org.ID).First(&duplicateOrg).Error; err == nil {
			return fmt.Errorf("organization with name '%s' already exists", org.Name)
		}
	}

	org.UpdatedAt = time.Now()
	if err := r.db.Save(org).Error; err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}
	return nil
}

// DeleteOrganization soft deletes an organization
func (r *OrganizationRepository) DeleteOrganization(id uuid.UUID) error {
	// Check if organization exists
	var org Organization
	if err := r.db.First(&org, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("organization not found")
		}
		return fmt.Errorf("failed to get organization: %w", err)
	}

	// Check if organization has associated users
	var userCount int64
	if err := r.db.Model(&User{}).Where("organization_id = ?", id).Count(&userCount).Error; err != nil {
		return fmt.Errorf("failed to check organization users: %w", err)
	}
	if userCount > 0 {
		return fmt.Errorf("cannot delete organization with %d associated users", userCount)
	}

	// Check if organization has associated departments
	var deptCount int64
	if err := r.db.Model(&Department{}).Where("organization_id = ?", id).Count(&deptCount).Error; err != nil {
		return fmt.Errorf("failed to check organization departments: %w", err)
	}
	if deptCount > 0 {
		return fmt.Errorf("cannot delete organization with %d associated departments", deptCount)
	}

	if err := r.db.Delete(&org).Error; err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}
	return nil
}

// GetOrganizationStats retrieves statistics for an organization
func (r *OrganizationRepository) GetOrganizationStats(id uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count users
	var userCount int64
	if err := r.db.Model(&User{}).Where("organization_id = ?", id).Count(&userCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}
	stats["users"] = userCount

	// Count departments
	var deptCount int64
	if err := r.db.Model(&Department{}).Where("organization_id = ?", id).Count(&deptCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count departments: %w", err)
	}
	stats["departments"] = deptCount

	// Count meetings
	var meetingCount int64
	if err := r.db.Model(&Meeting{}).
		Joins("JOIN users ON users.id = meetings.created_by").
		Where("users.organization_id = ?", id).
		Count(&meetingCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count meetings: %w", err)
	}
	stats["meetings"] = meetingCount

	return stats, nil
}

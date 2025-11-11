package database

import (
	"fmt"
	"time"
)

// EgressRepository handles LiveKit egress database operations
type EgressRepository struct {
	db *DB
}

// NewEgressRepository creates a new egress repository
func NewEgressRepository(db *DB) *EgressRepository {
	return &EgressRepository{db: db}
}

// Create creates a new egress record
func (r *EgressRepository) Create(egress *LiveKitEgress) error {
	if err := r.db.Create(egress).Error; err != nil {
		return fmt.Errorf("failed to create egress: %w", err)
	}
	return nil
}

// GetByID retrieves an egress by ID
func (r *EgressRepository) GetByID(egressID string) (*LiveKitEgress, error) {
	var egress LiveKitEgress
	if err := r.db.Where("id = ?", egressID).First(&egress).Error; err != nil {
		return nil, fmt.Errorf("failed to get egress: %w", err)
	}
	return &egress, nil
}

// GetByRoomSID retrieves all egress sessions for a room
func (r *EgressRepository) GetByRoomSID(roomSID string) ([]*LiveKitEgress, error) {
	var egresses []*LiveKitEgress
	if err := r.db.Where("room_sid = ?", roomSID).Order("created_at DESC").Find(&egresses).Error; err != nil {
		return nil, fmt.Errorf("failed to get egresses by room: %w", err)
	}
	return egresses, nil
}

// GetActiveByRoomSID retrieves active egress sessions for a room
func (r *EgressRepository) GetActiveByRoomSID(roomSID string) ([]*LiveKitEgress, error) {
	var egresses []*LiveKitEgress
	if err := r.db.Where("room_sid = ? AND status IN ?", roomSID, []string{"pending", "active"}).
		Order("created_at DESC").
		Find(&egresses).Error; err != nil {
		return nil, fmt.Errorf("failed to get active egresses: %w", err)
	}
	return egresses, nil
}

// UpdateStatus updates the status of an egress
func (r *EgressRepository) UpdateStatus(egressID string, status string) error {
	if err := r.db.Model(&LiveKitEgress{}).
		Where("id = ?", egressID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("failed to update egress status: %w", err)
	}
	return nil
}

// MarkStarted marks an egress as started
func (r *EgressRepository) MarkStarted(egressID string) error {
	now := time.Now()
	if err := r.db.Model(&LiveKitEgress{}).
		Where("id = ?", egressID).
		Updates(map[string]interface{}{
			"status":     "active",
			"started_at": &now,
			"updated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("failed to mark egress as started: %w", err)
	}
	return nil
}

// MarkCompleted marks an egress as completed with file info
func (r *EgressRepository) MarkCompleted(egressID string, filePath *string, fileSize *int64, duration *int64) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":     "complete",
		"ended_at":   &now,
		"updated_at": now,
	}

	if filePath != nil {
		updates["file_path"] = filePath
	}
	if fileSize != nil {
		updates["file_size"] = fileSize
	}
	if duration != nil {
		updates["duration"] = duration
	}

	if err := r.db.Model(&LiveKitEgress{}).
		Where("id = ?", egressID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to mark egress as completed: %w", err)
	}
	return nil
}

// MarkFailed marks an egress as failed with error message
func (r *EgressRepository) MarkFailed(egressID string, errorMsg string) error {
	now := time.Now()
	if err := r.db.Model(&LiveKitEgress{}).
		Where("id = ?", egressID).
		Updates(map[string]interface{}{
			"status":     "failed",
			"error":      &errorMsg,
			"ended_at":   &now,
			"updated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("failed to mark egress as failed: %w", err)
	}
	return nil
}

// List retrieves egresses with optional filters
func (r *EgressRepository) List(roomName *string, status *string, limit int, offset int) ([]*LiveKitEgress, int64, error) {
	query := r.db.Model(&LiveKitEgress{})

	if roomName != nil {
		query = query.Where("room_name = ?", *roomName)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count egresses: %w", err)
	}

	var egresses []*LiveKitEgress
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&egresses).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list egresses: %w", err)
	}

	return egresses, total, nil
}

// Delete deletes an egress record (soft delete not needed, hard delete)
func (r *EgressRepository) Delete(egressID string) error {
	if err := r.db.Where("id = ?", egressID).Delete(&LiveKitEgress{}).Error; err != nil {
		return fmt.Errorf("failed to delete egress: %w", err)
	}
	return nil
}

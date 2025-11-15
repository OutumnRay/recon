package database

import (
	"fmt"
	"time"

	"Recontext.online/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FCMDeviceRepository handles database operations for FCM devices
type FCMDeviceRepository struct {
	db *gorm.DB
}

// NewFCMDeviceRepository creates a new FCMDeviceRepository
func NewFCMDeviceRepository(db *gorm.DB) *FCMDeviceRepository {
	return &FCMDeviceRepository{db: db}
}

// RegisterDevice registers a new device or updates an existing one
func (r *FCMDeviceRepository) RegisterDevice(userID uuid.UUID, req *models.RegisterFCMDeviceRequest) (*models.FCMDevice, bool, error) {
	// First, try to find existing device by FCM token
	var existingDevice models.FCMDevice
	err := r.db.Where("fcm_token = ?", req.FCMToken).First(&existingDevice).Error

	if err == nil {
		// Device exists - update it
		existingDevice.UserID = userID // Update user ID in case device was transferred
		existingDevice.Platform = req.Platform
		existingDevice.DeviceModel = req.DeviceModel
		existingDevice.AppVersion = req.AppVersion
		existingDevice.OSVersion = req.OSVersion
		existingDevice.IsActive = true
		existingDevice.LastActiveAt = time.Now()

		if err := r.db.Save(&existingDevice).Error; err != nil {
			return nil, false, fmt.Errorf("failed to update device: %w", err)
		}

		return &existingDevice, false, nil
	}

	// Device doesn't exist - create new one
	newDevice := &models.FCMDevice{
		ID:           uuid.New(),
		UserID:       userID,
		FCMToken:     req.FCMToken,
		Platform:     req.Platform,
		DeviceModel:  req.DeviceModel,
		AppVersion:   req.AppVersion,
		OSVersion:    req.OSVersion,
		IsActive:     true,
		LastActiveAt: time.Now(),
	}

	if err := r.db.Create(newDevice).Error; err != nil {
		return nil, false, fmt.Errorf("failed to create device: %w", err)
	}

	return newDevice, true, nil
}

// UnregisterDevice unregisters a device by FCM token
func (r *FCMDeviceRepository) UnregisterDevice(userID uuid.UUID, fcmToken string) error {
	// Mark device as inactive instead of deleting
	result := r.db.Model(&models.FCMDevice{}).
		Where("fcm_token = ? AND user_id = ?", fcmToken, userID).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to unregister device: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("device not found or already inactive")
	}

	return nil
}

// GetUserDevices retrieves all active devices for a user
func (r *FCMDeviceRepository) GetUserDevices(userID uuid.UUID) ([]models.FCMDevice, error) {
	var devices []models.FCMDevice
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).
		Order("last_active_at DESC").
		Find(&devices).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user devices: %w", err)
	}

	return devices, nil
}

// GetUserFCMTokens retrieves all active FCM tokens for a user
func (r *FCMDeviceRepository) GetUserFCMTokens(userID uuid.UUID) ([]string, error) {
	var tokens []string
	err := r.db.Model(&models.FCMDevice{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Pluck("fcm_token", &tokens).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user FCM tokens: %w", err)
	}

	return tokens, nil
}

// DeleteDevice permanently deletes a device record
func (r *FCMDeviceRepository) DeleteDevice(id uuid.UUID, userID uuid.UUID) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.FCMDevice{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete device: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("device not found")
	}

	return nil
}

// CleanupInactiveDevices removes devices that haven't been active for a specified duration
func (r *FCMDeviceRepository) CleanupInactiveDevices(inactiveDuration time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-inactiveDuration)

	result := r.db.Where("last_active_at < ? AND is_active = ?", cutoffTime, false).
		Delete(&models.FCMDevice{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup inactive devices: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// UpdateDeviceActivity updates the last active time for a device
func (r *FCMDeviceRepository) UpdateDeviceActivity(fcmToken string) error {
	result := r.db.Model(&models.FCMDevice{}).
		Where("fcm_token = ?", fcmToken).
		Update("last_active_at", time.Now())

	if result.Error != nil {
		return fmt.Errorf("failed to update device activity: %w", result.Error)
	}

	return nil
}

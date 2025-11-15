package models

import (
	"time"

	"github.com/google/uuid"
)

// FCMDevice represents a device registered for push notifications
type FCMDevice struct {
	// Unique identifier for the device record
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	// User ID who owns this device
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	// Firebase Cloud Messaging token for this device
	FCMToken string `json:"fcm_token" gorm:"type:text;not null;uniqueIndex"`
	// Device platform (ios, android, web)
	Platform string `json:"platform" gorm:"type:varchar(20);not null"`
	// Device model/name (e.g., "iPhone 14 Pro", "Samsung Galaxy S23")
	DeviceModel string `json:"device_model,omitempty" gorm:"type:varchar(255)"`
	// App version running on this device
	AppVersion string `json:"app_version,omitempty" gorm:"type:varchar(50)"`
	// OS version running on this device
	OSVersion string `json:"os_version,omitempty" gorm:"type:varchar(50)"`
	// Whether this device is currently active
	IsActive bool `json:"is_active" gorm:"default:true"`
	// Last time this device token was verified/updated
	LastActiveAt time.Time `json:"last_active_at" gorm:"autoUpdateTime"`
	// Time when device was registered
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	// Time when device info was last updated
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for FCMDevice model
func (FCMDevice) TableName() string {
	return "fcm_devices"
}

// RegisterFCMDeviceRequest represents a request to register a device for push notifications
type RegisterFCMDeviceRequest struct {
	// Firebase Cloud Messaging token
	FCMToken string `json:"fcm_token" binding:"required"`
	// Device platform (ios, android, web)
	Platform string `json:"platform" binding:"required,oneof=ios android web"`
	// Device model/name (optional)
	DeviceModel string `json:"device_model,omitempty"`
	// App version (optional)
	AppVersion string `json:"app_version,omitempty"`
	// OS version (optional)
	OSVersion string `json:"os_version,omitempty"`
}

// RegisterFCMDeviceResponse represents the response after registering a device
type RegisterFCMDeviceResponse struct {
	// Unique identifier for the device record
	ID uuid.UUID `json:"id"`
	// Status message
	Message string `json:"message"`
	// Whether this is a new registration or an update
	IsNew bool `json:"is_new"`
}

// UnregisterFCMDeviceRequest represents a request to unregister a device
type UnregisterFCMDeviceRequest struct {
	// Firebase Cloud Messaging token to unregister
	FCMToken string `json:"fcm_token" binding:"required"`
}

// UnregisterFCMDeviceResponse represents the response after unregistering a device
type UnregisterFCMDeviceResponse struct {
	// Status message
	Message string `json:"message"`
}

package database

import (
	"encoding/json"
	"fmt"
	"time"

	"Recontext.online/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateUploadedFile creates a new file upload record
func (db *DB) CreateUploadedFile(file *models.UploadedFile) error {
	metadata, err := json.Marshal(file.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	dbFile := &UploadedFile{
		ID:           file.ID,
		Filename:     file.Filename,
		OriginalName: file.OriginalName,
		FileSize:     file.FileSize,
		MimeType:     file.MimeType,
		StoragePath:  file.StoragePath,
		UserID:       file.UserID,
		GroupID:      file.GroupID,
		Status:       string(file.Status),
		Metadata:     string(metadata),
		UploadedAt:   file.UploadedAt,
	}

	if file.TranscriptionID != nil {
		dbFile.TranscriptionID = file.TranscriptionID
	}

	if err := db.DB.Create(dbFile).Error; err != nil {
		return fmt.Errorf("failed to create uploaded file: %w", err)
	}

	return nil
}

// GetUploadedFileByID retrieves a file by its ID
func (db *DB) GetUploadedFileByID(id string) (*models.UploadedFile, error) {
	uuidID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	var dbFile UploadedFile
	err = db.DB.Where("id = ?", uuidID).First(&dbFile).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("failed to get uploaded file: %w", err)
	}

	file := &models.UploadedFile{
		ID:           dbFile.ID,
		Filename:     dbFile.Filename,
		OriginalName: dbFile.OriginalName,
		FileSize:     dbFile.FileSize,
		MimeType:     dbFile.MimeType,
		StoragePath:  dbFile.StoragePath,
		UserID:       dbFile.UserID,
		GroupID:      dbFile.GroupID,
		Status:       models.TranscriptionStatus(dbFile.Status),
		UploadedAt:   dbFile.UploadedAt,
		ProcessedAt:  dbFile.ProcessedAt,
	}

	if dbFile.TranscriptionID != nil {
		file.TranscriptionID = dbFile.TranscriptionID
	}

	if dbFile.Metadata != "" {
		if err := json.Unmarshal([]byte(dbFile.Metadata), &file.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return file, nil
}

// ListUploadedFilesByUser retrieves all files uploaded by a specific user
func (db *DB) ListUploadedFilesByUser(userID uuid.UUID, page, pageSize int) ([]models.UploadedFile, int, error) {
	var total int64
	var dbFiles []UploadedFile

	// Get total count
	if err := db.DB.Model(&UploadedFile{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count files: %w", err)
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := db.DB.Where("user_id = ?", userID).
		Order("uploaded_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&dbFiles).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list files: %w", err)
	}

	files := make([]models.UploadedFile, 0, len(dbFiles))
	for _, dbFile := range dbFiles {
		file := models.UploadedFile{
			ID:           dbFile.ID,
			Filename:     dbFile.Filename,
			OriginalName: dbFile.OriginalName,
			FileSize:     dbFile.FileSize,
			MimeType:     dbFile.MimeType,
			StoragePath:  dbFile.StoragePath,
			UserID:       dbFile.UserID,
			GroupID:      dbFile.GroupID,
			Status:       models.TranscriptionStatus(dbFile.Status),
			UploadedAt:   dbFile.UploadedAt,
			ProcessedAt:  dbFile.ProcessedAt,
		}

		if dbFile.TranscriptionID != nil {
			file.TranscriptionID = dbFile.TranscriptionID
		}

		if dbFile.Metadata != "" {
			json.Unmarshal([]byte(dbFile.Metadata), &file.Metadata)
		}

		files = append(files, file)
	}

	return files, int(total), nil
}

// UpdateFileStatus updates the status of a file
func (db *DB) UpdateFileStatus(fileID string, status models.TranscriptionStatus) error {
	uuidID, err := uuid.Parse(fileID)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       string(status),
		"processed_at": &now,
	}

	result := db.DB.Model(&UploadedFile{}).Where("id = ?", uuidID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update file status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("file not found")
	}

	return nil
}

// DeleteUploadedFile deletes a file record
func (db *DB) DeleteUploadedFile(fileID string) error {
	uuidID, err := uuid.Parse(fileID)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	result := db.DB.Where("id = ?", uuidID).Delete(&UploadedFile{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete file: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("file not found")
	}

	return nil
}

// CheckUserHasFilePermission checks if a user has permission to access files
func (db *DB) CheckUserHasFilePermission(userID uuid.UUID, action string) (bool, error) {
	var groups []Group
	err := db.DB.Table("groups g").
		Select("g.permissions").
		Joins("INNER JOIN group_memberships gm ON g.id = gm.group_id").
		Where("gm.user_id = ?", userID).
		Find(&groups).Error

	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	for _, group := range groups {
		var permissions map[string]interface{}
		if err := json.Unmarshal([]byte(group.Permissions), &permissions); err != nil {
			continue
		}

		// Check if user has "files" permission with the required action
		if filesPerms, ok := permissions["files"].(map[string]interface{}); ok {
			if actions, ok := filesPerms["actions"].([]interface{}); ok {
				for _, a := range actions {
					if a == action {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}

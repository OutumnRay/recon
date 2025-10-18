package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"Recontext.online/internal/models"
)

// CreateUploadedFile creates a new file upload record
func (db *DB) CreateUploadedFile(file *models.UploadedFile) error {
	metadata, err := json.Marshal(file.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO uploaded_files (
			id, filename, original_name, file_size, mime_type, storage_path,
			user_id, group_id, status, metadata, uploaded_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = db.Exec(
		query,
		file.ID, file.Filename, file.OriginalName, file.FileSize, file.MimeType,
		file.StoragePath, file.UserID, file.GroupID, file.Status, metadata, file.UploadedAt,
	)

	return err
}

// GetUploadedFileByID retrieves a file by its ID
func (db *DB) GetUploadedFileByID(id string) (*models.UploadedFile, error) {
	var file models.UploadedFile
	var metadataJSON []byte

	query := `
		SELECT id, filename, original_name, file_size, mime_type, storage_path,
			   user_id, group_id, status, transcription_id, metadata, uploaded_at, processed_at
		FROM uploaded_files
		WHERE id = $1
	`

	err := db.QueryRow(query, id).Scan(
		&file.ID, &file.Filename, &file.OriginalName, &file.FileSize, &file.MimeType,
		&file.StoragePath, &file.UserID, &file.GroupID, &file.Status, &file.TranscriptionID,
		&metadataJSON, &file.UploadedAt, &file.ProcessedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("file not found")
		}
		return nil, err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &file.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &file, nil
}

// ListUploadedFilesByUser retrieves all files uploaded by a specific user
func (db *DB) ListUploadedFilesByUser(userID string, page, pageSize int) ([]models.UploadedFile, int, error) {
	var files []models.UploadedFile
	var total int

	// Get total count
	countQuery := `SELECT COUNT(*) FROM uploaded_files WHERE user_id = $1`
	if err := db.QueryRow(countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	query := `
		SELECT id, filename, original_name, file_size, mime_type, storage_path,
			   user_id, group_id, status, transcription_id, metadata, uploaded_at, processed_at
		FROM uploaded_files
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var file models.UploadedFile
		var metadataJSON []byte

		if err := rows.Scan(
			&file.ID, &file.Filename, &file.OriginalName, &file.FileSize, &file.MimeType,
			&file.StoragePath, &file.UserID, &file.GroupID, &file.Status, &file.TranscriptionID,
			&metadataJSON, &file.UploadedAt, &file.ProcessedAt,
		); err != nil {
			return nil, 0, err
		}

		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &file.Metadata)
		}

		files = append(files, file)
	}

	return files, total, nil
}

// UpdateFileStatus updates the status of a file
func (db *DB) UpdateFileStatus(fileID string, status models.TranscriptionStatus) error {
	query := `
		UPDATE uploaded_files
		SET status = $1, processed_at = $2
		WHERE id = $3
	`

	_, err := db.Exec(query, status, time.Now(), fileID)
	return err
}

// DeleteUploadedFile deletes a file record
func (db *DB) DeleteUploadedFile(fileID string) error {
	query := `DELETE FROM uploaded_files WHERE id = $1`
	_, err := db.Exec(query, fileID)
	return err
}

// CheckUserHasFilePermission checks if a user has permission to access files
func (db *DB) CheckUserHasFilePermission(userID string, action string) (bool, error) {
	query := `
		SELECT g.permissions
		FROM groups g
		INNER JOIN group_memberships gm ON g.id = gm.group_id
		WHERE gm.user_id = $1
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var permissionsJSON []byte
		if err := rows.Scan(&permissionsJSON); err != nil {
			continue
		}

		var permissions map[string]interface{}
		if err := json.Unmarshal(permissionsJSON, &permissions); err != nil {
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

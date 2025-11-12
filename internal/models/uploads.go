package models

import (
	"time"

	"github.com/google/uuid"
)

// UploadedFile represents a file uploaded for transcription
type UploadedFile struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	Filename        string                 `json:"filename" db:"filename"`
	OriginalName    string                 `json:"original_name" db:"original_name"`
	FileSize        int64                  `json:"file_size" db:"file_size"`
	MimeType        string                 `json:"mime_type" db:"mime_type"`
	StoragePath     string                 `json:"storage_path" db:"storage_path"`
	UserID          uuid.UUID              `json:"user_id" db:"user_id"`
	GroupID         uuid.UUID              `json:"group_id" db:"group_id"`
	Status          TranscriptionStatus    `json:"status" db:"status"`
	TranscriptionID *uuid.UUID             `json:"transcription_id,omitempty" db:"transcription_id"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	UploadedAt      time.Time              `json:"uploaded_at" db:"uploaded_at"`
	ProcessedAt     *time.Time             `json:"processed_at,omitempty" db:"processed_at"`
}

// TranscriptionStatus represents the status of a transcription
type TranscriptionStatus string

const (
	StatusPending    TranscriptionStatus = "pending"
	StatusProcessing TranscriptionStatus = "processing"
	StatusCompleted  TranscriptionStatus = "completed"
	StatusFailed     TranscriptionStatus = "failed"
)

// FileUploadRequest represents a request to upload a file
type FileUploadRequest struct {
	File        interface{} `form:"file" binding:"required"`
	Description string      `form:"description"`
}

// FileUploadResponse represents the response after uploading a file
type FileUploadResponse struct {
	ID           uuid.UUID           `json:"id"`
	Filename     string              `json:"filename"`
	OriginalName string              `json:"original_name"`
	FileSize     int64               `json:"file_size"`
	Status       TranscriptionStatus `json:"status"`
	UploadedAt   time.Time           `json:"uploaded_at"`
}

// ListFilesRequest represents parameters for listing uploaded files
type ListFilesRequest struct {
	Page     int        `json:"page" form:"page" example:"1"`
	PageSize int        `json:"page_size" form:"page_size" example:"20"`
	Status   string     `json:"status" form:"status" example:"completed"`
	GroupID  *uuid.UUID `json:"group_id" form:"group_id"`
}

// ListFilesResponse represents a paginated list of uploaded files
type ListFilesResponse struct {
	Files    []UploadedFile `json:"files"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
}

// FileTranscription represents a transcription result
type FileTranscription struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	FileID          uuid.UUID              `json:"file_id" db:"file_id"`
	Text            string                 `json:"text" db:"text"`
	Language        string                 `json:"language" db:"language"`
	Confidence      float64                `json:"confidence" db:"confidence"`
	Duration        float64                `json:"duration" db:"duration"` // in seconds
	Segments        map[string]interface{} `json:"segments,omitempty" db:"segments"`
	TranscribedAt   time.Time              `json:"transcribed_at" db:"transcribed_at"`
	TranscribedBy   uuid.UUID              `json:"transcribed_by" db:"transcribed_by"` // service name or user ID
}

// DownloadFileRequest represents a request to download a file
type DownloadFileRequest struct {
	FileID uuid.UUID `json:"file_id" binding:"required"`
}

// DeleteFileRequest represents a request to delete a file
type DeleteFileRequest struct {
	FileID uuid.UUID `json:"file_id" binding:"required"`
	Reason string    `json:"reason,omitempty"`
}

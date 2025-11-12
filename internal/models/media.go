package models

import (
	"time"

	"github.com/google/uuid"
)

// Recording represents a recorded audio/video file
type Recording struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	UserID        uuid.UUID       `json:"user_id" db:"user_id"`
	Title         string          `json:"title" db:"title"`
	FileName      string          `json:"file_name" db:"file_name"`
	FileSize      int64           `json:"file_size" db:"file_size"`
	Duration      float64         `json:"duration" db:"duration"` // in seconds
	MimeType      string          `json:"mime_type" db:"mime_type"`
	StoragePath   string          `json:"storage_path" db:"storage_path"`
	Status        RecordingStatus `json:"status" db:"status"`
	TranscriptID  *uuid.UUID      `json:"transcript_id,omitempty" db:"transcript_id"`
	UploadedAt    time.Time       `json:"uploaded_at" db:"uploaded_at"`
	ProcessedAt   *time.Time      `json:"processed_at,omitempty" db:"processed_at"`
}

// RecordingStatus represents the processing status of a recording
type RecordingStatus string

const (
	RecordingStatusUploading    RecordingStatus = "uploading"
	RecordingStatusQueued       RecordingStatus = "queued"
	RecordingStatusTranscribing RecordingStatus = "transcribing"
	RecordingStatusCompleted    RecordingStatus = "completed"
	RecordingStatusFailed       RecordingStatus = "failed"
)

// Transcript represents a transcription result
type Transcript struct {
	ID           uuid.UUID           `json:"id" db:"id"`
	RecordingID  uuid.UUID           `json:"recording_id" db:"recording_id"`
	Text         string              `json:"text" db:"text"`
	Language     string              `json:"language" db:"language"`
	Segments     []TranscriptSegment `json:"segments"`
	Summary      string              `json:"summary,omitempty" db:"summary"`
	Status       string              `json:"status" db:"status"`
	ProcessedAt  time.Time           `json:"processed_at" db:"processed_at"`
	DurationSecs float64             `json:"duration_secs" db:"duration_secs"`
	CreatedAt    time.Time           `json:"created_at" db:"created_at"`
}

// TranscriptSegment represents a segment of transcription with timing
type TranscriptSegment struct {
	StartTime float64 `json:"start_time"` // seconds
	EndTime   float64 `json:"end_time"`   // seconds
	Text      string  `json:"text"`
	Speaker   string  `json:"speaker,omitempty"` // Speaker ID from diarization
	Confidence float64 `json:"confidence"`
}

// UploadRequest represents a file upload request
type UploadRequest struct {
	Title    string `json:"title" form:"title" binding:"required" example:"Meeting with Client"`
	FileName string `json:"file_name" form:"file_name" binding:"required" example:"meeting.mp4"`
}

// UploadResponse represents a file upload response
type UploadResponse struct {
	RecordingID uuid.UUID `json:"recording_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UploadURL   string    `json:"upload_url" example:"https://minio:9000/recontext/uploads/550e8400-e29b-41d4-a716-446655440000.mp4"`
	Message     string    `json:"message" example:"File uploaded successfully"`
}

// ListRecordingsRequest represents parameters for listing recordings
type ListRecordingsRequest struct {
	Page     int    `json:"page" form:"page" example:"1"`
	PageSize int    `json:"page_size" form:"page_size" example:"20"`
	Status   string `json:"status" form:"status" example:"completed"`
}

// ListRecordingsResponse represents a paginated list of recordings
type ListRecordingsResponse struct {
	Recordings []Recording `json:"recordings"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
}

// SearchRequest represents a transcript search request
type SearchRequest struct {
	Query    string `json:"query" form:"query" binding:"required" example:"budget discussion"`
	Page     int    `json:"page" form:"page" example:"1"`
	PageSize int    `json:"page_size" form:"page_size" example:"10"`
}

// SearchResult represents a search result
type SearchResult struct {
	RecordingID  uuid.UUID `json:"recording_id"`
	Title        string    `json:"title"`
	Snippet      string    `json:"snippet"`
	Timestamp    float64   `json:"timestamp"`
	Relevance    float64   `json:"relevance"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

// SearchResponse represents search results
type SearchResponse struct {
	Results  []SearchResult `json:"results"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// TranscriptionTask represents a task for the transcription worker
type TranscriptionTask struct {
	ID          uuid.UUID  `json:"id"`
	RecordingID uuid.UUID  `json:"recording_id"`
	AudioURL    string     `json:"audio_url"`
	Language    string     `json:"language,omitempty"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Error       string     `json:"error,omitempty"`
}

// SummarizationTask represents a task for the summarization worker
type SummarizationTask struct {
	ID           uuid.UUID  `json:"id"`
	TranscriptID uuid.UUID  `json:"transcript_id"`
	RecordingID  uuid.UUID  `json:"recording_id"`
	Text         string     `json:"text"`
	Status       string     `json:"status"`
	Summary      string     `json:"summary,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	Error        string     `json:"error,omitempty"`
}

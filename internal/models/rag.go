package models

import (
	"time"

	"github.com/google/uuid"
)

// DocumentChunk represents a chunk of text from a transcription with its embedding
type DocumentChunk struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	FileID          uuid.UUID              `json:"file_id" db:"file_id"`
	TranscriptionID *uuid.UUID             `json:"transcription_id,omitempty" db:"transcription_id"`
	ChunkText       string                 `json:"chunk_text" db:"chunk_text"`
	ChunkIndex      int                    `json:"chunk_index" db:"chunk_index"`
	Embedding       []float32              `json:"-" db:"embedding"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// RAGSearchRequest represents a search query
type RAGSearchRequest struct {
	Query     string `json:"query" binding:"required"`
	TopK      int    `json:"top_k,omitempty"` // Number of results to return (default: 5)
	Threshold float64 `json:"threshold,omitempty"` // Similarity threshold (default: 0.7)
}

// RAGSearchResult represents a single search result
type RAGSearchResult struct {
	ChunkID       uuid.UUID `json:"chunk_id"`
	FileID        uuid.UUID `json:"file_id"`
	FileName      string    `json:"file_name"`
	ChunkText     string    `json:"chunk_text"`
	ChunkIndex    int       `json:"chunk_index"`
	Similarity    float64   `json:"similarity"`
	UploadedAt    time.Time `json:"uploaded_at"`
}

// RAGSearchResponse represents the response from a search query
type RAGSearchResponse struct {
	Query   string            `json:"query"`
	Results []RAGSearchResult `json:"results"`
	Count   int               `json:"count"`
}

// RAGStatusResponse represents the status of RAG system
type RAGStatusResponse struct {
	TotalChunks     int  `json:"total_chunks"`
	IndexedFiles    int  `json:"indexed_files"`
	IsReady         bool `json:"is_ready"`
}

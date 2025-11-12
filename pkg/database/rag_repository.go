package database

import (
	"encoding/json"
	"fmt"

	"Recontext.online/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// CreateDocumentChunk stores a document chunk with its embedding
func (db *DB) CreateDocumentChunk(chunk *models.DocumentChunk) error {
	metadata, err := json.Marshal(chunk.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Convert float32 slice to PostgreSQL array format for storage
	// Note: GORM will handle the conversion, but we need to store as string for the vector type
	embeddingJSON, err := json.Marshal(chunk.Embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	dbChunk := &DocumentChunk{
		ID:         chunk.ID,
		FileID:     chunk.FileID,
		ChunkText:  chunk.ChunkText,
		ChunkIndex: chunk.ChunkIndex,
		Embedding:  string(embeddingJSON),
		Metadata:   string(metadata),
		CreatedAt:  chunk.CreatedAt,
	}

	if chunk.TranscriptionID != nil {
		dbChunk.TranscriptionID = chunk.TranscriptionID
	}

	if err := db.DB.Create(dbChunk).Error; err != nil {
		return fmt.Errorf("failed to create document chunk: %w", err)
	}

	return nil
}

// SearchSimilarChunks performs semantic search using vector similarity
func (db *DB) SearchSimilarChunks(queryEmbedding []float32, topK int, threshold float64) ([]models.RAGSearchResult, error) {
	if topK <= 0 {
		topK = 5
	}
	if threshold <= 0 {
		threshold = 0.7
	}

	embedding := pq.Array(queryEmbedding)

	var results []models.RAGSearchResult
	err := db.DB.Raw(`
		SELECT
			dc.id as chunk_id,
			dc.file_id,
			uf.original_name as file_name,
			dc.chunk_text,
			dc.chunk_index,
			1 - (dc.embedding <=> ?::vector) as similarity,
			uf.uploaded_at
		FROM document_chunks dc
		INNER JOIN uploaded_files uf ON dc.file_id = uf.id
		WHERE 1 - (dc.embedding <=> ?::vector) >= ?
		ORDER BY dc.embedding <=> ?::vector
		LIMIT ?
	`, embedding, embedding, threshold, embedding, topK).Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to search similar chunks: %w", err)
	}

	return results, nil
}

// GetRAGStatus returns statistics about the RAG system
func (db *DB) GetRAGStatus() (*models.RAGStatusResponse, error) {
	var status models.RAGStatusResponse

	// Count total chunks
	var totalChunks int64
	err := db.DB.Model(&DocumentChunk{}).Count(&totalChunks).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count chunks: %w", err)
	}
	status.TotalChunks = int(totalChunks)

	// Count indexed files
	var indexedFiles int64
	err = db.DB.Model(&DocumentChunk{}).Distinct("file_id").Count(&indexedFiles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count indexed files: %w", err)
	}
	status.IndexedFiles = int(indexedFiles)

	status.IsReady = status.TotalChunks > 0

	return &status, nil
}

// DeleteChunksByFileID deletes all chunks for a specific file
func (db *DB) DeleteChunksByFileID(fileID string) error {
	uuidID, err := uuid.Parse(fileID)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	if err := db.DB.Where("file_id = ?", uuidID).Delete(&DocumentChunk{}).Error; err != nil {
		return fmt.Errorf("failed to delete chunks: %w", err)
	}
	return nil
}

// CheckUserHasRAGPermission checks if a user has permission to use RAG
func (db *DB) CheckUserHasRAGPermission(userID uuid.UUID) (bool, error) {
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

		// Check if user has "rag" permission with "search" action
		if ragPerms, ok := permissions["rag"].(map[string]interface{}); ok {
			if actions, ok := ragPerms["actions"].([]interface{}); ok {
				for _, action := range actions {
					if action == "search" {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}

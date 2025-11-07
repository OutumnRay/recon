package database

import (
	"encoding/json"
	"fmt"

	"Recontext.online/internal/models"
	"github.com/lib/pq"
)

// CreateDocumentChunk stores a document chunk with its embedding
func (db *DB) CreateDocumentChunk(chunk *models.DocumentChunk) error {
	metadata, err := json.Marshal(chunk.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Convert float32 slice to PostgreSQL array format
	embedding := pq.Array(chunk.Embedding)

	query := `
		INSERT INTO document_chunks (
			id, file_id, transcription_id, chunk_text, chunk_index,
			embedding, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = db.Exec(
		query,
		chunk.ID, chunk.FileID, chunk.TranscriptionID, chunk.ChunkText, chunk.ChunkIndex,
		embedding, metadata, chunk.CreatedAt,
	)

	return err
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

	query := `
		SELECT
			dc.id as chunk_id,
			dc.file_id,
			uf.original_name as file_name,
			dc.chunk_text,
			dc.chunk_index,
			1 - (dc.embedding <=> $1::vector) as similarity,
			uf.uploaded_at
		FROM document_chunks dc
		INNER JOIN uploaded_files uf ON dc.file_id = uf.id
		WHERE 1 - (dc.embedding <=> $1::vector) >= $2
		ORDER BY dc.embedding <=> $1::vector
		LIMIT $3
	`

	rows, err := db.Query(query, embedding, threshold, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar chunks: %w", err)
	}
	defer rows.Close()

	var results []models.RAGSearchResult
	for rows.Next() {
		var result models.RAGSearchResult
		err := rows.Scan(
			&result.ChunkID,
			&result.FileID,
			&result.FileName,
			&result.ChunkText,
			&result.ChunkIndex,
			&result.Similarity,
			&result.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetRAGStatus returns statistics about the RAG system
func (db *DB) GetRAGStatus() (*models.RAGStatusResponse, error) {
	var status models.RAGStatusResponse

	// Count total chunks
	err := db.QueryRow("SELECT COUNT(*) FROM document_chunks").Scan(&status.TotalChunks)
	if err != nil {
		return nil, err
	}

	// Count indexed files
	err = db.QueryRow(`
		SELECT COUNT(DISTINCT file_id)
		FROM document_chunks
	`).Scan(&status.IndexedFiles)
	if err != nil {
		return nil, err
	}

	status.IsReady = status.TotalChunks > 0

	return &status, nil
}

// DeleteChunksByFileID deletes all chunks for a specific file
func (db *DB) DeleteChunksByFileID(fileID string) error {
	query := `DELETE FROM document_chunks WHERE file_id = $1`
	_, err := db.Exec(query, fileID)
	return err
}

// CheckUserHasRAGPermission checks if a user has permission to use RAG
func (db *DB) CheckUserHasRAGPermission(userID string) (bool, error) {
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

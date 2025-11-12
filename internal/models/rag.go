package models

import (
	"time"

	"github.com/google/uuid"
)

// DocumentChunk представляет фрагмент текста из транскрипции с его векторным представлением
type DocumentChunk struct {
	// ID - уникальный идентификатор фрагмента
	ID              uuid.UUID              `json:"id" db:"id"`
	// FileID - идентификатор файла
	FileID          uuid.UUID              `json:"file_id" db:"file_id"`
	// TranscriptionID - идентификатор транскрипции (если есть)
	TranscriptionID *uuid.UUID             `json:"transcription_id,omitempty" db:"transcription_id"`
	// ChunkText - текст фрагмента
	ChunkText       string                 `json:"chunk_text" db:"chunk_text"`
	// ChunkIndex - индекс фрагмента в документе
	ChunkIndex      int                    `json:"chunk_index" db:"chunk_index"`
	// Embedding - векторное представление фрагмента (не передается в JSON)
	Embedding       []float32              `json:"-" db:"embedding"`
	// Metadata - дополнительные метаданные фрагмента
	Metadata        map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	// CreatedAt - время создания
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// RAGSearchRequest представляет поисковый запрос
type RAGSearchRequest struct {
	// Query - текст запроса
	Query     string `json:"query" binding:"required"`
	// TopK - количество возвращаемых результатов (по умолчанию: 5)
	TopK      int    `json:"top_k,omitempty"`
	// Threshold - порог схожести (по умолчанию: 0.7)
	Threshold float64 `json:"threshold,omitempty"`
}

// RAGSearchResult представляет один результат поиска
type RAGSearchResult struct {
	// ChunkID - идентификатор фрагмента
	ChunkID       uuid.UUID `json:"chunk_id"`
	// FileID - идентификатор файла
	FileID        uuid.UUID `json:"file_id"`
	// FileName - имя файла
	FileName      string    `json:"file_name"`
	// ChunkText - текст фрагмента
	ChunkText     string    `json:"chunk_text"`
	// ChunkIndex - индекс фрагмента
	ChunkIndex    int       `json:"chunk_index"`
	// Similarity - уровень схожести с запросом
	Similarity    float64   `json:"similarity"`
	// UploadedAt - время загрузки файла
	UploadedAt    time.Time `json:"uploaded_at"`
}

// RAGSearchResponse представляет ответ на поисковый запрос
type RAGSearchResponse struct {
	// Query - исходный запрос
	Query   string            `json:"query"`
	// Results - список результатов
	Results []RAGSearchResult `json:"results"`
	// Count - количество найденных результатов
	Count   int               `json:"count"`
}

// RAGStatusResponse представляет статус системы RAG
type RAGStatusResponse struct {
	// TotalChunks - общее количество фрагментов
	TotalChunks     int  `json:"total_chunks"`
	// IndexedFiles - количество проиндексированных файлов
	IndexedFiles    int  `json:"indexed_files"`
	// IsReady - готова ли система к работе
	IsReady         bool `json:"is_ready"`
}

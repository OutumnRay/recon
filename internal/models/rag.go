package models

import (
	"time"

	"github.com/google/uuid"
)

// DocumentChunk представляет фрагмент текста из транскрипции с его векторным представлением
type DocumentChunk struct {
	// Уникальный идентификатор фрагмента
	ID              uuid.UUID              `json:"id" db:"id"`
	// Идентификатор файла
	FileID          uuid.UUID              `json:"file_id" db:"file_id"`
	// Идентификатор транскрипции (если есть)
	TranscriptionID *uuid.UUID             `json:"transcription_id,omitempty" db:"transcription_id"`
	// Текст фрагмента
	ChunkText       string                 `json:"chunk_text" db:"chunk_text"`
	// Индекс фрагмента в документе
	ChunkIndex      int                    `json:"chunk_index" db:"chunk_index"`
	// Векторное представление фрагмента (не передается в JSON)
	Embedding       []float32              `json:"-" db:"embedding"`
	// Дополнительные метаданные фрагмента
	Metadata        map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	// Время создания
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// RAGSearchRequest представляет поисковый запрос
type RAGSearchRequest struct {
	// Текст запроса
	Query     string `json:"query" binding:"required"`
	// Количество возвращаемых результатов (по умолчанию: 5)
	TopK      int    `json:"top_k,omitempty"`
	// Порог схожести (по умолчанию: 0.7)
	Threshold float64 `json:"threshold,omitempty"`
}

// RAGSearchResult представляет один результат поиска
type RAGSearchResult struct {
	// Идентификатор фрагмента
	ChunkID       uuid.UUID `json:"chunk_id"`
	// Идентификатор файла
	FileID        uuid.UUID `json:"file_id"`
	// Имя файла
	FileName      string    `json:"file_name"`
	// Текст фрагмента
	ChunkText     string    `json:"chunk_text"`
	// Индекс фрагмента
	ChunkIndex    int       `json:"chunk_index"`
	// Уровень схожести с запросом
	Similarity    float64   `json:"similarity"`
	// Время загрузки файла
	UploadedAt    time.Time `json:"uploaded_at"`
}

// RAGSearchResponse представляет ответ на поисковый запрос
type RAGSearchResponse struct {
	// Исходный запрос
	Query   string            `json:"query"`
	// Список результатов
	Results []RAGSearchResult `json:"results"`
	// Количество найденных результатов
	Count   int               `json:"count"`
}

// RAGStatusResponse представляет статус системы RAG
type RAGStatusResponse struct {
	// Общее количество фрагментов
	TotalChunks     int  `json:"total_chunks"`
	// Количество проиндексированных файлов
	IndexedFiles    int  `json:"indexed_files"`
	// Готова ли система к работе
	IsReady         bool `json:"is_ready"`
}

package models

import (
	"time"

	"github.com/google/uuid"
)

// Recording представляет записанный аудио/видео файл
type Recording struct {
	// ID - уникальный идентификатор записи
	ID            uuid.UUID       `json:"id" db:"id"`
	// UserID - идентификатор пользователя
	UserID        uuid.UUID       `json:"user_id" db:"user_id"`
	// Title - название записи
	Title         string          `json:"title" db:"title"`
	// FileName - имя файла
	FileName      string          `json:"file_name" db:"file_name"`
	// FileSize - размер файла в байтах
	FileSize      int64           `json:"file_size" db:"file_size"`
	// Duration - длительность в секундах
	Duration      float64         `json:"duration" db:"duration"`
	// MimeType - MIME тип файла
	MimeType      string          `json:"mime_type" db:"mime_type"`
	// StoragePath - путь к файлу в хранилище
	StoragePath   string          `json:"storage_path" db:"storage_path"`
	// Status - статус обработки записи
	Status        RecordingStatus `json:"status" db:"status"`
	// TranscriptID - идентификатор транскрипции (если есть)
	TranscriptID  *uuid.UUID      `json:"transcript_id,omitempty" db:"transcript_id"`
	// UploadedAt - время загрузки
	UploadedAt    time.Time       `json:"uploaded_at" db:"uploaded_at"`
	// ProcessedAt - время завершения обработки
	ProcessedAt   *time.Time      `json:"processed_at,omitempty" db:"processed_at"`
}

// RecordingStatus представляет статус обработки записи
type RecordingStatus string

const (
	RecordingStatusUploading    RecordingStatus = "uploading"    // Загружается
	RecordingStatusQueued       RecordingStatus = "queued"       // В очереди
	RecordingStatusTranscribing RecordingStatus = "transcribing" // Транскрибируется
	RecordingStatusCompleted    RecordingStatus = "completed"    // Завершено
	RecordingStatusFailed       RecordingStatus = "failed"       // Ошибка
)

// Transcript представляет результат транскрипции
type Transcript struct {
	// ID - уникальный идентификатор транскрипции
	ID           uuid.UUID           `json:"id" db:"id"`
	// RecordingID - идентификатор записи
	RecordingID  uuid.UUID           `json:"recording_id" db:"recording_id"`
	// Text - полный текст транскрипции
	Text         string              `json:"text" db:"text"`
	// Language - язык транскрипции
	Language     string              `json:"language" db:"language"`
	// Segments - сегменты транскрипции с временными метками
	Segments     []TranscriptSegment `json:"segments"`
	// Summary - краткое содержание (если есть)
	Summary      string              `json:"summary,omitempty" db:"summary"`
	// Status - статус обработки
	Status       string              `json:"status" db:"status"`
	// ProcessedAt - время обработки
	ProcessedAt  time.Time           `json:"processed_at" db:"processed_at"`
	// DurationSecs - длительность в секундах
	DurationSecs float64             `json:"duration_secs" db:"duration_secs"`
	// CreatedAt - время создания
	CreatedAt    time.Time           `json:"created_at" db:"created_at"`
}

// TranscriptSegment представляет сегмент транскрипции с временными метками
type TranscriptSegment struct {
	// StartTime - время начала в секундах
	StartTime float64 `json:"start_time"`
	// EndTime - время окончания в секундах
	EndTime   float64 `json:"end_time"`
	// Text - текст сегмента
	Text      string  `json:"text"`
	// Speaker - идентификатор говорящего из диаризации (опционально)
	Speaker   string  `json:"speaker,omitempty"`
	// Confidence - уровень уверенности распознавания
	Confidence float64 `json:"confidence"`
}

// UploadRequest представляет запрос на загрузку файла
type UploadRequest struct {
	// Title - название записи
	Title    string `json:"title" form:"title" binding:"required" example:"Meeting with Client"`
	// FileName - имя файла
	FileName string `json:"file_name" form:"file_name" binding:"required" example:"meeting.mp4"`
}

// UploadResponse представляет ответ после загрузки файла
type UploadResponse struct {
	// RecordingID - идентификатор созданной записи
	RecordingID uuid.UUID `json:"recording_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// UploadURL - URL для загрузки файла
	UploadURL   string    `json:"upload_url" example:"https://minio:9000/recontext/uploads/550e8400-e29b-41d4-a716-446655440000.mp4"`
	// Message - статусное сообщение
	Message     string    `json:"message" example:"File uploaded successfully"`
}

// ListRecordingsRequest представляет параметры для получения списка записей
type ListRecordingsRequest struct {
	// Page - номер страницы
	Page     int    `json:"page" form:"page" example:"1"`
	// PageSize - размер страницы
	PageSize int    `json:"page_size" form:"page_size" example:"20"`
	// Status - фильтр по статусу
	Status   string `json:"status" form:"status" example:"completed"`
}

// ListRecordingsResponse представляет постраничный список записей
type ListRecordingsResponse struct {
	// Recordings - список записей
	Recordings []Recording `json:"recordings"`
	// Total - общее количество записей
	Total      int         `json:"total"`
	// Page - номер текущей страницы
	Page       int         `json:"page"`
	// PageSize - размер страницы
	PageSize   int         `json:"page_size"`
}

// SearchRequest представляет запрос на поиск по транскриптам
type SearchRequest struct {
	// Query - поисковый запрос
	Query    string `json:"query" form:"query" binding:"required" example:"budget discussion"`
	// Page - номер страницы
	Page     int    `json:"page" form:"page" example:"1"`
	// PageSize - размер страницы
	PageSize int    `json:"page_size" form:"page_size" example:"10"`
}

// SearchResult представляет результат поиска
type SearchResult struct {
	// RecordingID - идентификатор записи
	RecordingID  uuid.UUID `json:"recording_id"`
	// Title - название записи
	Title        string    `json:"title"`
	// Snippet - фрагмент текста с совпадением
	Snippet      string    `json:"snippet"`
	// Timestamp - временная метка в секундах
	Timestamp    float64   `json:"timestamp"`
	// Relevance - релевантность результата
	Relevance    float64   `json:"relevance"`
	// UploadedAt - время загрузки
	UploadedAt   time.Time `json:"uploaded_at"`
}

// SearchResponse представляет результаты поиска
type SearchResponse struct {
	// Results - список результатов
	Results  []SearchResult `json:"results"`
	// Total - общее количество результатов
	Total    int            `json:"total"`
	// Page - номер текущей страницы
	Page     int            `json:"page"`
	// PageSize - размер страницы
	PageSize int            `json:"page_size"`
}

// TranscriptionTask представляет задачу для воркера транскрипции
type TranscriptionTask struct {
	// ID - уникальный идентификатор задачи
	ID          uuid.UUID  `json:"id"`
	// RecordingID - идентификатор записи
	RecordingID uuid.UUID  `json:"recording_id"`
	// AudioURL - URL аудиофайла
	AudioURL    string     `json:"audio_url"`
	// Language - язык для транскрипции (опционально)
	Language    string     `json:"language,omitempty"`
	// Status - статус задачи
	Status      string     `json:"status"`
	// CreatedAt - время создания задачи
	CreatedAt   time.Time  `json:"created_at"`
	// StartedAt - время начала обработки
	StartedAt   *time.Time `json:"started_at,omitempty"`
	// CompletedAt - время завершения обработки
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	// Error - описание ошибки (если есть)
	Error       string     `json:"error,omitempty"`
}

// SummarizationTask представляет задачу для воркера суммаризации
type SummarizationTask struct {
	// ID - уникальный идентификатор задачи
	ID           uuid.UUID  `json:"id"`
	// TranscriptID - идентификатор транскрипции
	TranscriptID uuid.UUID  `json:"transcript_id"`
	// RecordingID - идентификатор записи
	RecordingID  uuid.UUID  `json:"recording_id"`
	// Text - текст для суммаризации
	Text         string     `json:"text"`
	// Status - статус задачи
	Status       string     `json:"status"`
	// Summary - результат суммаризации
	Summary      string     `json:"summary,omitempty"`
	// CreatedAt - время создания задачи
	CreatedAt    time.Time  `json:"created_at"`
	// StartedAt - время начала обработки
	StartedAt    *time.Time `json:"started_at,omitempty"`
	// CompletedAt - время завершения обработки
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	// Error - описание ошибки (если есть)
	Error        string     `json:"error,omitempty"`
}

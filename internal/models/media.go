package models

import (
	"time"

	"github.com/google/uuid"
)

// Recording представляет записанный аудио/видео файл
type Recording struct {
	// Уникальный идентификатор записи
	ID            uuid.UUID       `json:"id" db:"id"`
	// Идентификатор пользователя
	UserID        uuid.UUID       `json:"user_id" db:"user_id"`
	// Название записи
	Title         string          `json:"title" db:"title"`
	// Имя файла
	FileName      string          `json:"file_name" db:"file_name"`
	// Размер файла в байтах
	FileSize      int64           `json:"file_size" db:"file_size"`
	// Длительность в секундах
	Duration      float64         `json:"duration" db:"duration"`
	// MIME тип файла
	MimeType      string          `json:"mime_type" db:"mime_type"`
	// Путь к файлу в хранилище
	StoragePath   string          `json:"storage_path" db:"storage_path"`
	// Статус обработки записи
	Status        RecordingStatus `json:"status" db:"status"`
	// Идентификатор транскрипции (если есть)
	TranscriptID  *uuid.UUID      `json:"transcript_id,omitempty" db:"transcript_id"`
	// Время загрузки
	UploadedAt    time.Time       `json:"uploaded_at" db:"uploaded_at"`
	// Время завершения обработки
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
	// Уникальный идентификатор транскрипции
	ID           uuid.UUID           `json:"id" db:"id"`
	// Идентификатор записи
	RecordingID  uuid.UUID           `json:"recording_id" db:"recording_id"`
	// Полный текст транскрипции
	Text         string              `json:"text" db:"text"`
	// Язык транскрипции
	Language     string              `json:"language" db:"language"`
	// Сегменты транскрипции с временными метками
	Segments     []TranscriptSegment `json:"segments"`
	// Краткое содержание (если есть)
	Summary      string              `json:"summary,omitempty" db:"summary"`
	// Статус обработки
	Status       string              `json:"status" db:"status"`
	// Время обработки
	ProcessedAt  time.Time           `json:"processed_at" db:"processed_at"`
	// Длительность в секундах
	DurationSecs float64             `json:"duration_secs" db:"duration_secs"`
	// Время создания
	CreatedAt    time.Time           `json:"created_at" db:"created_at"`
}

// TranscriptSegment представляет сегмент транскрипции с временными метками
type TranscriptSegment struct {
	// Время начала в секундах
	StartTime float64 `json:"start_time"`
	// Время окончания в секундах
	EndTime   float64 `json:"end_time"`
	// Текст сегмента
	Text      string  `json:"text"`
	// Идентификатор говорящего из диаризации (опционально)
	Speaker   string  `json:"speaker,omitempty"`
	// Уровень уверенности распознавания
	Confidence float64 `json:"confidence"`
}

// UploadRequest представляет запрос на загрузку файла
type UploadRequest struct {
	// Название записи
	Title    string `json:"title" form:"title" binding:"required" example:"Meeting with Client"`
	// Имя файла
	FileName string `json:"file_name" form:"file_name" binding:"required" example:"meeting.mp4"`
}

// UploadResponse представляет ответ после загрузки файла
type UploadResponse struct {
	// Идентификатор созданной записи
	RecordingID uuid.UUID `json:"recording_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// URL для загрузки файла
	UploadURL   string    `json:"upload_url" example:"https://minio:9000/recontext/uploads/550e8400-e29b-41d4-a716-446655440000.mp4"`
	// Статусное сообщение
	Message     string    `json:"message" example:"File uploaded successfully"`
}

// ListRecordingsRequest представляет параметры для получения списка записей
type ListRecordingsRequest struct {
	// Номер страницы
	Page     int    `json:"page" form:"page" example:"1"`
	// Размер страницы
	PageSize int    `json:"page_size" form:"page_size" example:"20"`
	// Фильтр по статусу
	Status   string `json:"status" form:"status" example:"completed"`
}

// ListRecordingsResponse представляет постраничный список записей
type ListRecordingsResponse struct {
	// Список записей
	Recordings []Recording `json:"recordings"`
	// Общее количество записей
	Total      int         `json:"total"`
	// Номер текущей страницы
	Page       int         `json:"page"`
	// Размер страницы
	PageSize   int         `json:"page_size"`
}

// SearchRequest представляет запрос на поиск по транскриптам
type SearchRequest struct {
	// Поисковый запрос
	Query    string `json:"query" form:"query" binding:"required" example:"budget discussion"`
	// Номер страницы
	Page     int    `json:"page" form:"page" example:"1"`
	// Размер страницы
	PageSize int    `json:"page_size" form:"page_size" example:"10"`
}

// SearchResult представляет результат поиска
type SearchResult struct {
	// Идентификатор записи
	RecordingID  uuid.UUID `json:"recording_id"`
	// Название записи
	Title        string    `json:"title"`
	// Фрагмент текста с совпадением
	Snippet      string    `json:"snippet"`
	// Временная метка в секундах
	Timestamp    float64   `json:"timestamp"`
	// Релевантность результата
	Relevance    float64   `json:"relevance"`
	// Время загрузки
	UploadedAt   time.Time `json:"uploaded_at"`
}

// SearchResponse представляет результаты поиска
type SearchResponse struct {
	// Список результатов
	Results  []SearchResult `json:"results"`
	// Общее количество результатов
	Total    int            `json:"total"`
	// Номер текущей страницы
	Page     int            `json:"page"`
	// Размер страницы
	PageSize int            `json:"page_size"`
}

// TranscriptionTask представляет задачу для воркера транскрипции
type TranscriptionTask struct {
	// Уникальный идентификатор задачи
	ID          uuid.UUID  `json:"id"`
	// Идентификатор записи
	RecordingID uuid.UUID  `json:"recording_id"`
	// URL аудиофайла
	AudioURL    string     `json:"audio_url"`
	// Язык для транскрипции (опционально)
	Language    string     `json:"language,omitempty"`
	// Статус задачи
	Status      string     `json:"status"`
	// Время создания задачи
	CreatedAt   time.Time  `json:"created_at"`
	// Время начала обработки
	StartedAt   *time.Time `json:"started_at,omitempty"`
	// Время завершения обработки
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	// Описание ошибки (если есть)
	Error       string     `json:"error,omitempty"`
}

// SummarizationTask представляет задачу для воркера суммаризации
type SummarizationTask struct {
	// Уникальный идентификатор задачи
	ID           uuid.UUID  `json:"id"`
	// Идентификатор транскрипции
	TranscriptID uuid.UUID  `json:"transcript_id"`
	// Идентификатор записи
	RecordingID  uuid.UUID  `json:"recording_id"`
	// Текст для суммаризации
	Text         string     `json:"text"`
	// Статус задачи
	Status       string     `json:"status"`
	// Результат суммаризации
	Summary      string     `json:"summary,omitempty"`
	// Время создания задачи
	CreatedAt    time.Time  `json:"created_at"`
	// Время начала обработки
	StartedAt    *time.Time `json:"started_at,omitempty"`
	// Время завершения обработки
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	// Описание ошибки (если есть)
	Error        string     `json:"error,omitempty"`
}

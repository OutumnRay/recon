package models

import (
	"time"

	"github.com/google/uuid"
)

// UploadedFile представляет файл, загруженный для транскрипции
type UploadedFile struct {
	// Уникальный идентификатор файла
	ID              uuid.UUID              `json:"id" db:"id"`
	// Имя файла в хранилище
	Filename        string                 `json:"filename" db:"filename"`
	// Оригинальное имя файла
	OriginalName    string                 `json:"original_name" db:"original_name"`
	// Размер файла в байтах
	FileSize        int64                  `json:"file_size" db:"file_size"`
	// MIME тип файла
	MimeType        string                 `json:"mime_type" db:"mime_type"`
	// Путь к файлу в хранилище
	StoragePath     string                 `json:"storage_path" db:"storage_path"`
	// Идентификатор пользователя, загрузившего файл
	UserID          uuid.UUID              `json:"user_id" db:"user_id"`
	// Идентификатор группы
	GroupID         uuid.UUID              `json:"group_id" db:"group_id"`
	// Статус обработки файла
	Status          TranscriptionStatus    `json:"status" db:"status"`
	// Идентификатор транскрипции (если есть)
	TranscriptionID *uuid.UUID             `json:"transcription_id,omitempty" db:"transcription_id"`
	// Дополнительные метаданные файла
	Metadata        map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	// Время загрузки файла
	UploadedAt      time.Time              `json:"uploaded_at" db:"uploaded_at"`
	// Время завершения обработки файла
	ProcessedAt     *time.Time             `json:"processed_at,omitempty" db:"processed_at"`
}

// TranscriptionStatus представляет статус транскрипции
type TranscriptionStatus string

const (
	StatusPending    TranscriptionStatus = "pending"    // Ожидает обработки
	StatusProcessing TranscriptionStatus = "processing" // В процессе обработки
	StatusCompleted  TranscriptionStatus = "completed"  // Завершено
	StatusFailed     TranscriptionStatus = "failed"     // Ошибка
)

// FileUploadRequest представляет запрос на загрузку файла
type FileUploadRequest struct {
	// Файл для загрузки
	File        interface{} `form:"file" binding:"required"`
	// Описание файла
	Description string      `form:"description"`
}

// FileUploadResponse представляет ответ после загрузки файла
type FileUploadResponse struct {
	// Уникальный идентификатор загруженного файла
	ID           uuid.UUID           `json:"id"`
	// Имя файла
	Filename     string              `json:"filename"`
	// Оригинальное имя файла
	OriginalName string              `json:"original_name"`
	// Размер файла в байтах
	FileSize     int64               `json:"file_size"`
	// Статус обработки
	Status       TranscriptionStatus `json:"status"`
	// Время загрузки
	UploadedAt   time.Time           `json:"uploaded_at"`
}

// ListFilesRequest представляет параметры для получения списка загруженных файлов
type ListFilesRequest struct {
	// Номер страницы
	Page     int        `json:"page" form:"page" example:"1"`
	// Размер страницы
	PageSize int        `json:"page_size" form:"page_size" example:"20"`
	// Фильтр по статусу
	Status   string     `json:"status" form:"status" example:"completed"`
	// Фильтр по группе
	GroupID  *uuid.UUID `json:"group_id" form:"group_id"`
}

// ListFilesResponse представляет постраничный список загруженных файлов
type ListFilesResponse struct {
	// Список файлов
	Files    []UploadedFile `json:"files"`
	// Общее количество файлов
	Total    int            `json:"total"`
	// Номер текущей страницы
	Page     int            `json:"page"`
	// Размер страницы
	PageSize int            `json:"pageSize"`
}

// FileTranscription представляет результат транскрипции файла
type FileTranscription struct {
	// Уникальный идентификатор транскрипции
	ID              uuid.UUID              `json:"id" db:"id"`
	// Идентификатор файла
	FileID          uuid.UUID              `json:"file_id" db:"file_id"`
	// Текст транскрипции
	Text            string                 `json:"text" db:"text"`
	// Язык транскрипции
	Language        string                 `json:"language" db:"language"`
	// Уровень уверенности распознавания
	Confidence      float64                `json:"confidence" db:"confidence"`
	// Длительность аудио в секундах
	Duration        float64                `json:"duration" db:"duration"`
	// Сегменты транскрипции с временными метками
	Segments        map[string]interface{} `json:"segments,omitempty" db:"segments"`
	// Время создания транскрипции
	TranscribedAt   time.Time              `json:"transcribed_at" db:"transcribed_at"`
	// Идентификатор сервиса или пользователя, создавшего транскрипцию
	TranscribedBy   uuid.UUID              `json:"transcribed_by" db:"transcribed_by"`
}

// DownloadFileRequest представляет запрос на скачивание файла
type DownloadFileRequest struct {
	// Идентификатор файла для скачивания
	FileID uuid.UUID `json:"file_id" binding:"required"`
}

// DeleteFileRequest представляет запрос на удаление файла
type DeleteFileRequest struct {
	// Идентификатор файла для удаления
	FileID uuid.UUID `json:"file_id" binding:"required"`
	// Причина удаления
	Reason string    `json:"reason,omitempty"`
}

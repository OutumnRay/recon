package models

import (
	"time"

	"github.com/google/uuid"
)

// UploadedFile представляет файл, загруженный для транскрипции
type UploadedFile struct {
	// ID - уникальный идентификатор файла
	ID              uuid.UUID              `json:"id" db:"id"`
	// Filename - имя файла в хранилище
	Filename        string                 `json:"filename" db:"filename"`
	// OriginalName - оригинальное имя файла
	OriginalName    string                 `json:"original_name" db:"original_name"`
	// FileSize - размер файла в байтах
	FileSize        int64                  `json:"file_size" db:"file_size"`
	// MimeType - MIME тип файла
	MimeType        string                 `json:"mime_type" db:"mime_type"`
	// StoragePath - путь к файлу в хранилище
	StoragePath     string                 `json:"storage_path" db:"storage_path"`
	// UserID - идентификатор пользователя, загрузившего файл
	UserID          uuid.UUID              `json:"user_id" db:"user_id"`
	// GroupID - идентификатор группы
	GroupID         uuid.UUID              `json:"group_id" db:"group_id"`
	// Status - статус обработки файла
	Status          TranscriptionStatus    `json:"status" db:"status"`
	// TranscriptionID - идентификатор транскрипции (если есть)
	TranscriptionID *uuid.UUID             `json:"transcription_id,omitempty" db:"transcription_id"`
	// Metadata - дополнительные метаданные файла
	Metadata        map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	// UploadedAt - время загрузки файла
	UploadedAt      time.Time              `json:"uploaded_at" db:"uploaded_at"`
	// ProcessedAt - время завершения обработки файла
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
	// File - файл для загрузки
	File        interface{} `form:"file" binding:"required"`
	// Description - описание файла
	Description string      `form:"description"`
}

// FileUploadResponse представляет ответ после загрузки файла
type FileUploadResponse struct {
	// ID - уникальный идентификатор загруженного файла
	ID           uuid.UUID           `json:"id"`
	// Filename - имя файла
	Filename     string              `json:"filename"`
	// OriginalName - оригинальное имя файла
	OriginalName string              `json:"original_name"`
	// FileSize - размер файла в байтах
	FileSize     int64               `json:"file_size"`
	// Status - статус обработки
	Status       TranscriptionStatus `json:"status"`
	// UploadedAt - время загрузки
	UploadedAt   time.Time           `json:"uploaded_at"`
}

// ListFilesRequest представляет параметры для получения списка загруженных файлов
type ListFilesRequest struct {
	// Page - номер страницы
	Page     int        `json:"page" form:"page" example:"1"`
	// PageSize - размер страницы
	PageSize int        `json:"page_size" form:"page_size" example:"20"`
	// Status - фильтр по статусу
	Status   string     `json:"status" form:"status" example:"completed"`
	// GroupID - фильтр по группе
	GroupID  *uuid.UUID `json:"group_id" form:"group_id"`
}

// ListFilesResponse представляет постраничный список загруженных файлов
type ListFilesResponse struct {
	// Files - список файлов
	Files    []UploadedFile `json:"files"`
	// Total - общее количество файлов
	Total    int            `json:"total"`
	// Page - номер текущей страницы
	Page     int            `json:"page"`
	// PageSize - размер страницы
	PageSize int            `json:"pageSize"`
}

// FileTranscription представляет результат транскрипции файла
type FileTranscription struct {
	// ID - уникальный идентификатор транскрипции
	ID              uuid.UUID              `json:"id" db:"id"`
	// FileID - идентификатор файла
	FileID          uuid.UUID              `json:"file_id" db:"file_id"`
	// Text - текст транскрипции
	Text            string                 `json:"text" db:"text"`
	// Language - язык транскрипции
	Language        string                 `json:"language" db:"language"`
	// Confidence - уровень уверенности распознавания
	Confidence      float64                `json:"confidence" db:"confidence"`
	// Duration - длительность аудио в секундах
	Duration        float64                `json:"duration" db:"duration"`
	// Segments - сегменты транскрипции с временными метками
	Segments        map[string]interface{} `json:"segments,omitempty" db:"segments"`
	// TranscribedAt - время создания транскрипции
	TranscribedAt   time.Time              `json:"transcribed_at" db:"transcribed_at"`
	// TranscribedBy - идентификатор сервиса или пользователя, создавшего транскрипцию
	TranscribedBy   uuid.UUID              `json:"transcribed_by" db:"transcribed_by"`
}

// DownloadFileRequest представляет запрос на скачивание файла
type DownloadFileRequest struct {
	// FileID - идентификатор файла для скачивания
	FileID uuid.UUID `json:"file_id" binding:"required"`
}

// DeleteFileRequest представляет запрос на удаление файла
type DeleteFileRequest struct {
	// FileID - идентификатор файла для удаления
	FileID uuid.UUID `json:"file_id" binding:"required"`
	// Reason - причина удаления
	Reason string    `json:"reason,omitempty"`
}

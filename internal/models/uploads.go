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
	FileID uuid.UUID `json:"file_id" binding:"required"`
}

// DeleteFileRequest представляет запрос на удаление файла
type DeleteFileRequest struct {
	FileID uuid.UUID `json:"file_id" binding:"required"`
	Reason string    `json:"reason,omitempty"`
}

// ─── S3 Multipart Upload flow ─────────────────────────────────────────────────
//
// Полный цикл загрузки:
//
//  1. POST /api/v1/files/init
//     Бэкенд инициирует S3 multipart upload и возвращает presigned PUT URL для
//     каждой части. Фронтенд загружает части параллельно и собирает ETags.
//
//  2. PUT <part.upload_url>   (прямо в MinIO, без Authorization)
//     Тело — ровно part.size байт начиная с part.offset.
//     MinIO возвращает ETag в заголовке ответа.
//
//  3. POST /api/v1/files/{id}/confirm  { upload_id, parts: [{part_number, etag}...] }
//     Бэкенд вызывает CompleteMultipartUpload и ставит задачу в очередь.
//     При success=false — AbortMultipartUpload + удаление записи.

// InitUploadRequest — тело POST /api/v1/files/init
type InitUploadRequest struct {
	Title     string `json:"title"`
	FileName  string `json:"file_name"`
	FileSize  int64  `json:"file_size"`
	MimeType  string `json:"mime_type"`
	Language  string `json:"language"`  // "ru", "en", "auto"
	ChunkSize int64  `json:"chunk_size"` // байт на часть; 0 → 10 МБ; мин 5 МБ
}

// MultipartPart — одна часть для загрузки (presigned PUT URL + диапазон байт)
type MultipartPart struct {
	PartNumber int    `json:"part_number"`
	UploadURL  string `json:"upload_url"`
	Offset     int64  `json:"offset"`
	Size       int64  `json:"size"`
}

// InitUploadResponse — ответ на POST /api/v1/files/init
type InitUploadResponse struct {
	FileID       uuid.UUID       `json:"file_id"`
	UploadID     string          `json:"upload_id"`     // S3 multipart uploadId
	UploadMethod string          `json:"upload_method"` // всегда "MULTIPART"
	Parts        []MultipartPart `json:"parts"`
	StoragePath  string          `json:"storage_path"`
	ExpiresAt    time.Time       `json:"expires_at"`
}

// UploadedPart — ETag загруженной части, отправляется клиентом в confirm
type UploadedPart struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}

// ConfirmUploadRequest — тело POST /api/v1/files/{id}/confirm
//
// success=false: загрузка не завершена (ошибка, отмена пользователем).
// Бэкенд вызовет AbortMultipartUpload и удалит запись из БД.
type ConfirmUploadRequest struct {
	UploadID string         `json:"upload_id"`
	Parts    []UploadedPart `json:"parts"`
	// Success=false — отмена; nil/true — нормальное подтверждение
	Success  *bool          `json:"success,omitempty"`
}

// ConfirmUploadResponse — ответ на POST /api/v1/files/{id}/confirm
type ConfirmUploadResponse struct {
	FileID  uuid.UUID `json:"file_id"`
	Status  string    `json:"status"`
	Message string    `json:"message"`
}

// FileStatusResponse — ответ на GET /api/v1/files/{id}/status
type FileStatusResponse struct {
	FileID    uuid.UUID  `json:"file_id"`
	Status    string     `json:"status"`
	Progress  int        `json:"progress"`
	Stage     string     `json:"stage,omitempty"`
	Error     string     `json:"error,omitempty"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// FileListItem — элемент в списке файлов
type FileListItem struct {
	ID           uuid.UUID  `json:"id"`
	Title        string     `json:"title"`
	FileName     string     `json:"file_name"`
	FileSize     int64      `json:"file_size"`
	MimeType     string     `json:"mime_type"`
	Duration     *float64   `json:"duration,omitempty"`
	Status       string     `json:"status"`
	Progress     int        `json:"progress"`
	Language     string     `json:"language"`
	UploadedAt   time.Time  `json:"uploaded_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// FileListResponse — ответ на GET /api/v1/files
type FileListResponse struct {
	Items    []FileListItem `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// FileDetailResponse — ответ на GET /api/v1/files/{id}
type FileDetailResponse struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	FileName    string     `json:"file_name"`
	FileSize    int64      `json:"file_size"`
	MimeType    string     `json:"mime_type"`
	Duration    *float64   `json:"duration,omitempty"`
	Status      string     `json:"status"`
	Progress    int        `json:"progress"`
	Stage       string     `json:"stage,omitempty"`
	Language    string     `json:"language"`
	ErrorMsg    string     `json:"error,omitempty"`
	UploadedAt  time.Time  `json:"uploaded_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Transcript  *struct {
		PhraseCount  int64    `json:"phrase_count"`
		SpeakerCount int      `json:"speaker_count"`
		HasSummary   bool     `json:"has_summary"`
		Speakers     []string `json:"speakers,omitempty"`
	} `json:"transcript_summary,omitempty"`
	VideoURL string `json:"video_url,omitempty"`
}

// FilePhraseItem — одна фраза в ответе транскрипции
type FilePhraseItem struct {
	PhraseIndex int      `json:"phrase_index"`
	StartTime   float64  `json:"start_time"`
	EndTime     float64  `json:"end_time"`
	Text        string   `json:"text"`
	Speaker     string   `json:"speaker,omitempty"`
	Confidence  *float64 `json:"confidence,omitempty"`
}

// FileTranscriptResponse — ответ на GET /api/v1/files/{id}/transcript
type FileTranscriptResponse struct {
	FileID   uuid.UUID        `json:"file_id"`
	Language string           `json:"language"`
	Duration *float64         `json:"duration,omitempty"`
	Speakers []string         `json:"speakers"`
	Phrases  []FilePhraseItem `json:"phrases"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// FileSummaryResponse — ответ на GET /api/v1/files/{id}/summary
type FileSummaryResponse struct {
	FileID      uuid.UUID `json:"file_id"`
	Summary     string    `json:"summary,omitempty"`
	SummaryRu   string    `json:"summary_ru,omitempty"`
	KeyTopics   []string  `json:"key_topics,omitempty"`
	ActionItems []string  `json:"action_items,omitempty"`
	Status      string    `json:"status"`
	GeneratedAt time.Time `json:"generated_at"`
}

// FileResultCallback — тело POST /internal/files/{id}/result (от Python-воркера)
type FileResultCallback struct {
	TaskID          string  `json:"task_id"`
	SessionID       string  `json:"session_id"`
	Status          string  `json:"status"`
	Duration        float64 `json:"duration"`
	ParagraphsCount int     `json:"paragraphs_count"`
	ChunksCount     int     `json:"chunks_count"`
	TranscriptPath  string  `json:"transcript_path"`
	SRTPath         string  `json:"srt_path"`
	Error           string  `json:"error"`
}

// FileResultParagraph — один абзац из paragraphs.json воркера
type FileResultParagraph struct {
	Start   float64 `json:"start"`
	End     float64 `json:"end"`
	Text    string  `json:"text"`
	Speaker string  `json:"speaker"`
}

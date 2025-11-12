package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// LiveKitWebhookEvent представляет событие webhook верхнего уровня от LiveKit
type LiveKitWebhookEvent struct {
	// Event - тип события
	Event     string          `json:"event"`
	// Room - данные комнаты (необработанный JSON)
	Room      json.RawMessage `json:"room,omitempty"`
	// Participant - данные участника (необработанный JSON)
	Participant json.RawMessage `json:"participant,omitempty"`
	// Track - данные дорожки (необработанный JSON)
	Track     json.RawMessage `json:"track,omitempty"`
	// ID - идентификатор события
	ID        uuid.UUID       `json:"id"`
	// CreatedAt - время создания события
	CreatedAt string          `json:"createdAt"`
}

// Room представляет комнату LiveKit
type Room struct {
	// ID - внутренний уникальный идентификатор
	ID                uuid.UUID      `json:"id" db:"id"`
	// SID - идентификатор сессии LiveKit
	SID               string         `json:"sid" db:"sid"`
	// Name - название комнаты
	Name              string         `json:"name" db:"name"`
	// EmptyTimeout - таймаут пустой комнаты в секундах
	EmptyTimeout      int            `json:"emptyTimeout" db:"empty_timeout"`
	// DepartureTimeout - таймаут после ухода участника в секундах
	DepartureTimeout  int            `json:"departureTimeout" db:"departure_timeout"`
	// CreationTime - время создания (строка)
	CreationTime      string         `json:"creationTime" db:"creation_time"`
	// CreationTimeMs - время создания в миллисекундах (строка)
	CreationTimeMs    string         `json:"creationTimeMs" db:"creation_time_ms"`
	// TurnPassword - пароль TURN сервера
	TurnPassword      string         `json:"turnPassword,omitempty" db:"turn_password"`
	// EnabledCodecs - включенные кодеки
	EnabledCodecs     []EnabledCodec `json:"enabledCodecs" db:"-"`
	// EnabledCodecsJSON - включенные кодеки в формате JSON (для БД)
	EnabledCodecsJSON string         `json:"-" db:"enabled_codecs"`
	// Status - статус комнаты (active, finished)
	Status            string         `json:"status" db:"status"`
	// StartedAt - время начала сессии
	StartedAt         time.Time      `json:"startedAt" db:"started_at"`
	// FinishedAt - время завершения сессии
	FinishedAt        *time.Time     `json:"finishedAt,omitempty" db:"finished_at"`
	// CreatedAtDB - время создания записи в БД
	CreatedAtDB       time.Time      `json:"-" db:"created_at"`
	// UpdatedAt - время последнего обновления
	UpdatedAt         time.Time      `json:"-" db:"updated_at"`
}

// EnabledCodec представляет кодек, включенный в комнате
type EnabledCodec struct {
	// Mime - MIME тип кодека
	Mime string `json:"mime"`
}

// Participant представляет участника комнаты LiveKit
type Participant struct {
	// ID - внутренний уникальный идентификатор
	ID              uuid.UUID       `json:"id" db:"id"`
	// SID - идентификатор сессии участника LiveKit
	SID             string          `json:"sid" db:"sid"`
	// RoomSID - идентификатор сессии комнаты
	RoomSID         string          `json:"-" db:"room_sid"`
	// Identity - идентификатор пользователя
	Identity        string          `json:"identity" db:"identity"`
	// Name - отображаемое имя участника
	Name            string          `json:"name" db:"name"`
	// State - состояние участника (ACTIVE, DISCONNECTED)
	State           string          `json:"state" db:"state"`
	// JoinedAt - время присоединения (строка)
	JoinedAt        string          `json:"joinedAt" db:"joined_at"`
	// JoinedAtMs - время присоединения в миллисекундах (строка)
	JoinedAtMs      string          `json:"joinedAtMs" db:"joined_at_ms"`
	// Version - версия протокола
	Version         int             `json:"version" db:"version"`
	// Permission - разрешения участника (необработанный JSON)
	Permission      json.RawMessage `json:"permission" db:"permission"`
	// IsPublisher - является ли участник издателем
	IsPublisher     bool            `json:"isPublisher,omitempty" db:"is_publisher"`
	// DisconnectReason - причина отключения
	DisconnectReason string         `json:"disconnectReason,omitempty" db:"disconnect_reason"`
	// LeftAt - время выхода из комнаты
	LeftAt          *time.Time      `json:"leftAt,omitempty" db:"left_at"`
	// CreatedAtDB - время создания записи в БД
	CreatedAtDB     time.Time       `json:"-" db:"created_at"`
	// UpdatedAt - время последнего обновления
	UpdatedAt       time.Time       `json:"-" db:"updated_at"`
}

// Track представляет аудио или видео дорожку
type Track struct {
	// ID - внутренний уникальный идентификатор
	ID               uuid.UUID       `json:"id" db:"id"`
	// SID - идентификатор сессии дорожки LiveKit
	SID              string          `json:"sid" db:"sid"`
	// ParticipantSID - идентификатор сессии участника
	ParticipantSID   string          `json:"-" db:"participant_sid"`
	// RoomSID - идентификатор сессии комнаты
	RoomSID          string          `json:"-" db:"room_sid"`
	// Type - тип дорожки (VIDEO, AUDIO)
	Type             string          `json:"type,omitempty" db:"type"`
	// Source - источник дорожки (MICROPHONE, CAMERA)
	Source           string          `json:"source" db:"source"`
	// MimeType - MIME тип дорожки
	MimeType         string          `json:"mimeType" db:"mime_type"`
	// Mid - идентификатор медиа
	Mid              string          `json:"mid" db:"mid"`
	// Width - ширина видео (для видео дорожек)
	Width            int             `json:"width,omitempty" db:"width"`
	// Height - высота видео (для видео дорожек)
	Height           int             `json:"height,omitempty" db:"height"`
	// Simulcast - используется ли simulcast
	Simulcast        bool            `json:"simulcast,omitempty" db:"simulcast"`
	// Layers - слои simulcast (необработанный JSON)
	Layers           json.RawMessage `json:"layers,omitempty" db:"layers"`
	// Codecs - используемые кодеки (необработанный JSON)
	Codecs           json.RawMessage `json:"codecs,omitempty" db:"codecs"`
	// Stream - идентификатор потока
	Stream           string          `json:"stream,omitempty" db:"stream"`
	// Version - версия (необработанный JSON)
	Version          json.RawMessage `json:"version,omitempty" db:"version"`
	// AudioFeatures - аудио функции
	AudioFeatures    []string        `json:"audioFeatures,omitempty" db:"-"`
	// AudioFeaturesJSON - аудио функции в формате JSON (для БД)
	AudioFeaturesJSON string         `json:"-" db:"audio_features"`
	// BackupCodecPolicy - политика резервного кодека
	BackupCodecPolicy string         `json:"backupCodecPolicy,omitempty" db:"backup_codec_policy"`
	// Status - статус дорожки (published, unpublished)
	Status           string          `json:"status" db:"status"`
	// PublishedAt - время публикации
	PublishedAt      time.Time       `json:"publishedAt" db:"published_at"`
	// UnpublishedAt - время отмены публикации
	UnpublishedAt    *time.Time      `json:"unpublishedAt,omitempty" db:"unpublished_at"`
	// CreatedAtDB - время создания записи в БД
	CreatedAtDB      time.Time       `json:"-" db:"created_at"`
	// UpdatedAt - время последнего обновления
	UpdatedAt        time.Time       `json:"-" db:"updated_at"`
}

// WebhookEventLog представляет лог всех полученных событий webhook
type WebhookEventLog struct {
	// ID - уникальный идентификатор записи лога
	ID          uuid.UUID       `json:"id" db:"id"`
	// EventType - тип события
	EventType   string          `json:"event_type" db:"event_type"`
	// EventID - идентификатор события от LiveKit
	EventID     string          `json:"event_id" db:"event_id"`
	// RoomSID - идентификатор комнаты (если применимо)
	RoomSID     string          `json:"room_sid,omitempty" db:"room_sid"`
	// ParticipantSID - идентификатор участника (если применимо)
	ParticipantSID string       `json:"participant_sid,omitempty" db:"participant_sid"`
	// TrackSID - идентификатор дорожки (если применимо)
	TrackSID    string          `json:"track_sid,omitempty" db:"track_sid"`
	// Payload - полезная нагрузка события (необработанный JSON)
	Payload     json.RawMessage `json:"payload" db:"payload"`
	// CreatedAt - время создания записи
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// WebhookRequest представляет входящую полезную нагрузку webhook
type WebhookRequest struct {
	// Event - тип события
	Event       string                 `json:"event" binding:"required"`
	// Room - данные комнаты
	Room        map[string]interface{} `json:"room,omitempty"`
	// Participant - данные участника
	Participant map[string]interface{} `json:"participant,omitempty"`
	// Track - данные дорожки
	Track       map[string]interface{} `json:"track,omitempty"`
	// ID - идентификатор события
	ID          string                 `json:"id"`
	// CreatedAt - время создания события
	CreatedAt   string                 `json:"createdAt"`
}

// WebhookResponse представляет ответ на webhook
type WebhookResponse struct {
	// Status - статус обработки
	Status  string `json:"status"`
	// Message - сообщение о результате
	Message string `json:"message"`
}

// TableName переопределяет имя таблицы, используемое для Room
func (Room) TableName() string {
	return "livekit_rooms"
}

// TableName переопределяет имя таблицы, используемое для Participant
func (Participant) TableName() string {
	return "livekit_participants"
}

// TableName переопределяет имя таблицы, используемое для Track
func (Track) TableName() string {
	return "livekit_tracks"
}

// TableName переопределяет имя таблицы, используемое для WebhookEventLog
func (WebhookEventLog) TableName() string {
	return "livekit_webhook_events"
}

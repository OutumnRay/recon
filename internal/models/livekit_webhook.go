package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// LiveKitWebhookEvent представляет событие webhook верхнего уровня от LiveKit
type LiveKitWebhookEvent struct {
	// Тип события
	Event     string          `json:"event"`
	// Данные комнаты (необработанный JSON)
	Room      json.RawMessage `json:"room,omitempty"`
	// Данные участника (необработанный JSON)
	Participant json.RawMessage `json:"participant,omitempty"`
	// Данные дорожки (необработанный JSON)
	Track     json.RawMessage `json:"track,omitempty"`
	// Идентификатор события
	ID        uuid.UUID       `json:"id"`
	// Время создания события
	CreatedAt string          `json:"createdAt"`
}

// Room представляет комнату LiveKit
type Room struct {
	// Внутренний уникальный идентификатор
	ID                uuid.UUID      `json:"id" db:"id"`
	// Идентификатор сессии LiveKit
	SID               string         `json:"sid" db:"sid"`
	// Название комнаты
	Name              string         `json:"name" db:"name"`
	// Таймаут пустой комнаты в секундах
	EmptyTimeout      int            `json:"emptyTimeout" db:"empty_timeout"`
	// Таймаут после ухода участника в секундах
	DepartureTimeout  int            `json:"departureTimeout" db:"departure_timeout"`
	// Время создания (строка)
	CreationTime      string         `json:"creationTime" db:"creation_time"`
	// Время создания в миллисекундах (строка)
	CreationTimeMs    string         `json:"creationTimeMs" db:"creation_time_ms"`
	// Пароль TURN сервера
	TurnPassword      string         `json:"turnPassword,omitempty" db:"turn_password"`
	// Включенные кодеки
	EnabledCodecs     []EnabledCodec `json:"enabledCodecs" db:"-"`
	// Включенные кодеки в формате JSON (для БД)
	EnabledCodecsJSON string         `json:"-" db:"enabled_codecs"`
	// Статус комнаты (active, finished)
	Status            string         `json:"status" db:"status"`
	// Время начала сессии
	StartedAt         time.Time      `json:"startedAt" db:"started_at"`
	// Время завершения сессии
	FinishedAt        *time.Time     `json:"finishedAt,omitempty" db:"finished_at"`
	// Время создания записи в БД
	CreatedAtDB       time.Time      `json:"-" db:"created_at"`
	// Время последнего обновления
	UpdatedAt         time.Time      `json:"-" db:"updated_at"`
}

// EnabledCodec представляет кодек, включенный в комнате
type EnabledCodec struct {
	// MIME тип кодека
	Mime string `json:"mime"`
}

// Participant представляет участника комнаты LiveKit
type Participant struct {
	// Внутренний уникальный идентификатор
	ID              uuid.UUID       `json:"id" db:"id"`
	// Идентификатор сессии участника LiveKit
	SID             string          `json:"sid" db:"sid"`
	// Идентификатор сессии комнаты
	RoomSID         string          `json:"-" db:"room_sid"`
	// Идентификатор пользователя
	Identity        string          `json:"identity" db:"identity"`
	// Отображаемое имя участника
	Name            string          `json:"name" db:"name"`
	// Состояние участника (ACTIVE, DISCONNECTED)
	State           string          `json:"state" db:"state"`
	// Время присоединения (строка)
	JoinedAt        string          `json:"joinedAt" db:"joined_at"`
	// Время присоединения в миллисекундах (строка)
	JoinedAtMs      string          `json:"joinedAtMs" db:"joined_at_ms"`
	// Версия протокола
	Version         int             `json:"version" db:"version"`
	// Разрешения участника (необработанный JSON)
	Permission      json.RawMessage `json:"permission" db:"permission"`
	// Является ли участник издателем
	IsPublisher     bool            `json:"isPublisher,omitempty" db:"is_publisher"`
	// Причина отключения
	DisconnectReason string         `json:"disconnectReason,omitempty" db:"disconnect_reason"`
	// Время выхода из комнаты
	LeftAt          *time.Time      `json:"leftAt,omitempty" db:"left_at"`
	// Время создания записи в БД
	CreatedAtDB     time.Time       `json:"-" db:"created_at"`
	// Время последнего обновления
	UpdatedAt       time.Time       `json:"-" db:"updated_at"`
}

// Track представляет аудио или видео дорожку
type Track struct {
	// Внутренний уникальный идентификатор
	ID               uuid.UUID       `json:"id" db:"id"`
	// Идентификатор сессии дорожки LiveKit
	SID              string          `json:"sid" db:"sid"`
	// Идентификатор сессии участника
	ParticipantSID   string          `json:"-" db:"participant_sid"`
	// Идентификатор сессии комнаты
	RoomSID          string          `json:"-" db:"room_sid"`
	// Тип дорожки (VIDEO, AUDIO)
	Type             string          `json:"type,omitempty" db:"type"`
	// Источник дорожки (MICROPHONE, CAMERA)
	Source           string          `json:"source" db:"source"`
	// MIME тип дорожки
	MimeType         string          `json:"mimeType" db:"mime_type"`
	// Идентификатор медиа
	Mid              string          `json:"mid" db:"mid"`
	// Ширина видео (для видео дорожек)
	Width            int             `json:"width,omitempty" db:"width"`
	// Высота видео (для видео дорожек)
	Height           int             `json:"height,omitempty" db:"height"`
	// Используется ли simulcast
	Simulcast        bool            `json:"simulcast,omitempty" db:"simulcast"`
	// Слои simulcast (необработанный JSON)
	Layers           json.RawMessage `json:"layers,omitempty" db:"layers"`
	// Используемые кодеки (необработанный JSON)
	Codecs           json.RawMessage `json:"codecs,omitempty" db:"codecs"`
	// Идентификатор потока
	Stream           string          `json:"stream,omitempty" db:"stream"`
	// Версия (необработанный JSON)
	Version          json.RawMessage `json:"version,omitempty" db:"version"`
	// Аудио функции
	AudioFeatures    []string        `json:"audioFeatures,omitempty" db:"-"`
	// Аудио функции в формате JSON (для БД)
	AudioFeaturesJSON string         `json:"-" db:"audio_features"`
	// Политика резервного кодека
	BackupCodecPolicy string         `json:"backupCodecPolicy,omitempty" db:"backup_codec_policy"`
	// Статус дорожки (published, unpublished)
	Status           string          `json:"status" db:"status"`
	// Время публикации
	PublishedAt      time.Time       `json:"publishedAt" db:"published_at"`
	// Время отмены публикации
	UnpublishedAt    *time.Time      `json:"unpublishedAt,omitempty" db:"unpublished_at"`
	// Время создания записи в БД
	CreatedAtDB      time.Time       `json:"-" db:"created_at"`
	// Время последнего обновления
	UpdatedAt        time.Time       `json:"-" db:"updated_at"`
}

// WebhookEventLog представляет лог всех полученных событий webhook
type WebhookEventLog struct {
	// Уникальный идентификатор записи лога
	ID          uuid.UUID       `json:"id" db:"id"`
	// Тип события
	EventType   string          `json:"event_type" db:"event_type"`
	// Идентификатор события от LiveKit
	EventID     string          `json:"event_id" db:"event_id"`
	// Идентификатор комнаты (если применимо)
	RoomSID     string          `json:"room_sid,omitempty" db:"room_sid"`
	// Идентификатор участника (если применимо)
	ParticipantSID string       `json:"participant_sid,omitempty" db:"participant_sid"`
	// Идентификатор дорожки (если применимо)
	TrackSID    string          `json:"track_sid,omitempty" db:"track_sid"`
	// Полезная нагрузка события (необработанный JSON)
	Payload     json.RawMessage `json:"payload" db:"payload"`
	// Время создания записи
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// WebhookRequest представляет входящую полезную нагрузку webhook
type WebhookRequest struct {
	// Тип события
	Event       string                 `json:"event" binding:"required"`
	// Данные комнаты
	Room        map[string]interface{} `json:"room,omitempty"`
	// Данные участника
	Participant map[string]interface{} `json:"participant,omitempty"`
	// Данные дорожки
	Track       map[string]interface{} `json:"track,omitempty"`
	// Идентификатор события
	ID          string                 `json:"id"`
	// Время создания события
	CreatedAt   string                 `json:"createdAt"`
}

// WebhookResponse представляет ответ на webhook
type WebhookResponse struct {
	// Статус обработки
	Status  string `json:"status"`
	// Сообщение о результате
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

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
	ID                uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid()" json:"id" db:"id"`
	// Идентификатор сессии LiveKit
	SID               string         `gorm:"column:sid;uniqueIndex;type:varchar(255);not null" json:"sid" db:"sid"`
	// Название комнаты
	Name              string         `gorm:"type:varchar(255);not null" json:"name" db:"name"`
	// Таймаут пустой комнаты в секундах
	EmptyTimeout      int            `gorm:"default:300" json:"emptyTimeout" db:"empty_timeout"`
	// Таймаут после ухода участника в секундах
	DepartureTimeout  int            `gorm:"default:20" json:"departureTimeout" db:"departure_timeout"`
	// Время создания (строка)
	CreationTime      string         `gorm:"type:varchar(50)" json:"creationTime" db:"creation_time"`
	// Время создания в миллисекундах (строка)
	CreationTimeMs    string         `gorm:"type:varchar(50)" json:"creationTimeMs" db:"creation_time_ms"`
	// Пароль TURN сервера
	TurnPassword      string         `gorm:"type:text" json:"turnPassword,omitempty" db:"turn_password"`
	// Включенные кодеки
	EnabledCodecs     []EnabledCodec `json:"enabledCodecs" gorm:"-" db:"-"`
	// Включенные кодеки в формате JSON (для БД)
	EnabledCodecsJSON string         `gorm:"column:enabled_codecs;type:jsonb;default:'[]'" json:"-" db:"enabled_codecs"`
	// Статус комнаты (active, finished)
	Status            string         `gorm:"type:varchar(50);not null;default:'active'" json:"status" db:"status"`
	// Время начала сессии
	StartedAt         time.Time      `gorm:"column:started_at" json:"startedAt" db:"started_at"`
	// Время завершения сессии
	FinishedAt        *time.Time     `gorm:"column:finished_at" json:"finishedAt,omitempty" db:"finished_at"`
	// Время создания записи в БД
	CreatedAtDB       time.Time      `gorm:"column:created_at;autoCreateTime" json:"-" db:"created_at"`
	// Время последнего обновления
	UpdatedAt         time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"-" db:"updated_at"`
}

// EnabledCodec представляет кодек, включенный в комнате
type EnabledCodec struct {
	// MIME тип кодека
	Mime string `json:"mime"`
}

// Participant представляет участника комнаты LiveKit
type Participant struct {
	// Внутренний уникальный идентификатор
	ID              uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid()" json:"id" db:"id"`
	// Идентификатор сессии участника LiveKit
	SID             string          `gorm:"column:sid;uniqueIndex;type:varchar(255);not null" json:"sid" db:"sid"`
	// Идентификатор сессии комнаты
	RoomSID         string          `gorm:"column:room_sid;type:varchar(255);not null" json:"-" db:"room_sid"`
	// Идентификатор пользователя
	Identity        string          `gorm:"column:identity;type:varchar(255);not null" json:"identity" db:"identity"`
	// Отображаемое имя участника
	Name            string          `gorm:"column:name;type:varchar(255)" json:"name" db:"name"`
	// Состояние участника (ACTIVE, DISCONNECTED)
	State           string          `gorm:"column:state;type:varchar(50);not null" json:"state" db:"state"`
	// Время присоединения (строка)
	JoinedAt        string          `gorm:"column:joined_at;type:varchar(50)" json:"joinedAt" db:"joined_at"`
	// Время присоединения в миллисекундах (строка)
	JoinedAtMs      string          `gorm:"column:joined_at_ms;type:varchar(50)" json:"joinedAtMs" db:"joined_at_ms"`
	// Версия протокола
	Version         int             `gorm:"column:version;default:0" json:"version" db:"version"`
	// Разрешения участника (необработанный JSON)
	Permission      json.RawMessage `gorm:"column:permission;type:jsonb" json:"permission" db:"permission"`
	// Является ли участник издателем
	IsPublisher     bool            `gorm:"column:is_publisher;default:false" json:"isPublisher,omitempty" db:"is_publisher"`
	// Причина отключения
	DisconnectReason string         `gorm:"column:disconnect_reason;type:text" json:"disconnectReason,omitempty" db:"disconnect_reason"`
	// Время выхода из комнаты
	LeftAt          *time.Time      `gorm:"column:left_at" json:"leftAt,omitempty" db:"left_at"`
	// Время создания записи в БД
	CreatedAtDB     time.Time       `gorm:"column:created_at_db;autoCreateTime" json:"-" db:"created_at"`
	// Время последнего обновления
	UpdatedAt       time.Time       `gorm:"column:updated_at;autoUpdateTime" json:"-" db:"updated_at"`
}

// Track представляет аудио или видео дорожку
type Track struct {
	// Внутренний уникальный идентификатор
	ID               uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid()" json:"id" db:"id"`
	// Идентификатор сессии дорожки LiveKit
	SID              string          `gorm:"column:sid;uniqueIndex;type:varchar(255);not null" json:"sid" db:"sid"`
	// Идентификатор сессии участника
	ParticipantSID   string          `gorm:"column:participant_sid;type:varchar(255);not null" json:"-" db:"participant_sid"`
	// Идентификатор сессии комнаты
	RoomSID          string          `gorm:"column:room_sid;type:varchar(255);not null" json:"-" db:"room_sid"`
	// Тип дорожки (VIDEO, AUDIO)
	Type             string          `gorm:"column:type;type:varchar(50)" json:"type,omitempty" db:"type"`
	// Источник дорожки (MICROPHONE, CAMERA)
	Source           string          `gorm:"column:source;type:varchar(50);not null" json:"source" db:"source"`
	// MIME тип дорожки
	MimeType         string          `gorm:"column:mime_type;type:varchar(100)" json:"mimeType" db:"mime_type"`
	// Идентификатор медиа
	Mid              string          `gorm:"column:mid;type:varchar(255)" json:"mid" db:"mid"`
	// Ширина видео (для видео дорожек)
	Width            int             `gorm:"column:width;default:0" json:"width,omitempty" db:"width"`
	// Высота видео (для видео дорожек)
	Height           int             `gorm:"column:height;default:0" json:"height,omitempty" db:"height"`
	// Используется ли simulcast
	Simulcast        bool            `gorm:"column:simulcast;default:false" json:"simulcast,omitempty" db:"simulcast"`
	// Слои simulcast (необработанный JSON)
	Layers           json.RawMessage `gorm:"column:layers;type:jsonb" json:"layers,omitempty" db:"layers"`
	// Используемые кодеки (необработанный JSON)
	Codecs           json.RawMessage `gorm:"column:codecs;type:jsonb" json:"codecs,omitempty" db:"codecs"`
	// Идентификатор потока
	Stream           string          `gorm:"column:stream;type:varchar(255)" json:"stream,omitempty" db:"stream"`
	// Версия (необработанный JSON)
	Version          json.RawMessage `gorm:"column:version;type:jsonb" json:"version,omitempty" db:"version"`
	// Аудио функции
	AudioFeatures    []string        `gorm:"-" json:"audioFeatures,omitempty" db:"-"`
	// Аудио функции в формате JSON (для БД)
	AudioFeaturesJSON string         `gorm:"column:audio_features;type:jsonb" json:"-" db:"audio_features"`
	// Политика резервного кодека
	BackupCodecPolicy string         `gorm:"column:backup_codec_policy;type:varchar(100)" json:"backupCodecPolicy,omitempty" db:"backup_codec_policy"`
	// Статус дорожки (published, unpublished)
	Status           string          `gorm:"column:status;type:varchar(50);not null" json:"status" db:"status"`
	// Время публикации
	PublishedAt      time.Time       `gorm:"column:published_at" json:"publishedAt" db:"published_at"`
	// Время отмены публикации
	UnpublishedAt    *time.Time      `gorm:"column:unpublished_at" json:"unpublishedAt,omitempty" db:"unpublished_at"`
	// Время создания записи в БД
	CreatedAtDB      time.Time       `gorm:"column:created_at_db;autoCreateTime" json:"-" db:"created_at"`
	// Время последнего обновления
	UpdatedAt        time.Time       `gorm:"column:updated_at;autoUpdateTime" json:"-" db:"updated_at"`
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

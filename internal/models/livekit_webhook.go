package models

import (
	"encoding/json"
	"time"
)

// LiveKitWebhookEvent represents the top-level webhook event from LiveKit
type LiveKitWebhookEvent struct {
	Event     string          `json:"event"`
	Room      json.RawMessage `json:"room,omitempty"`
	Participant json.RawMessage `json:"participant,omitempty"`
	Track     json.RawMessage `json:"track,omitempty"`
	ID        string          `json:"id"`
	CreatedAt string          `json:"createdAt"`
}

// Room represents a LiveKit room
type Room struct {
	ID                string         `json:"id" db:"id"`
	SID               string         `json:"sid" db:"sid"`
	Name              string         `json:"name" db:"name"`
	EmptyTimeout      int            `json:"emptyTimeout" db:"empty_timeout"`
	DepartureTimeout  int            `json:"departureTimeout" db:"departure_timeout"`
	CreationTime      string         `json:"creationTime" db:"creation_time"`
	CreationTimeMs    string         `json:"creationTimeMs" db:"creation_time_ms"`
	TurnPassword      string         `json:"turnPassword,omitempty" db:"turn_password"`
	EnabledCodecs     []EnabledCodec `json:"enabledCodecs" db:"-"`
	EnabledCodecsJSON string         `json:"-" db:"enabled_codecs"`
	Status            string         `json:"status" db:"status"` // active, finished
	StartedAt         time.Time      `json:"startedAt" db:"started_at"`
	FinishedAt        *time.Time     `json:"finishedAt,omitempty" db:"finished_at"`
	CreatedAtDB       time.Time      `json:"-" db:"created_at"`
	UpdatedAt         time.Time      `json:"-" db:"updated_at"`
}

// EnabledCodec represents a codec enabled in the room
type EnabledCodec struct {
	Mime string `json:"mime"`
}

// Participant represents a participant in a LiveKit room
type Participant struct {
	ID              string          `json:"id" db:"id"`
	SID             string          `json:"sid" db:"sid"`
	RoomSID         string          `json:"-" db:"room_sid"`
	Identity        string          `json:"identity" db:"identity"`
	Name            string          `json:"name" db:"name"`
	State           string          `json:"state" db:"state"` // ACTIVE, DISCONNECTED
	JoinedAt        string          `json:"joinedAt" db:"joined_at"`
	JoinedAtMs      string          `json:"joinedAtMs" db:"joined_at_ms"`
	Version         int             `json:"version" db:"version"`
	Permission      json.RawMessage `json:"permission" db:"permission"`
	IsPublisher     bool            `json:"isPublisher,omitempty" db:"is_publisher"`
	DisconnectReason string         `json:"disconnectReason,omitempty" db:"disconnect_reason"`
	LeftAt          *time.Time      `json:"leftAt,omitempty" db:"left_at"`
	CreatedAtDB     time.Time       `json:"-" db:"created_at"`
	UpdatedAt       time.Time       `json:"-" db:"updated_at"`
}

// Track represents an audio or video track
type Track struct {
	ID               string          `json:"id" db:"id"`
	SID              string          `json:"sid" db:"sid"`
	ParticipantSID   string          `json:"-" db:"participant_sid"`
	RoomSID          string          `json:"-" db:"room_sid"`
	Type             string          `json:"type,omitempty" db:"type"` // VIDEO, AUDIO
	Source           string          `json:"source" db:"source"` // MICROPHONE, CAMERA
	MimeType         string          `json:"mimeType" db:"mime_type"`
	Mid              string          `json:"mid" db:"mid"`
	Width            int             `json:"width,omitempty" db:"width"`
	Height           int             `json:"height,omitempty" db:"height"`
	Simulcast        bool            `json:"simulcast,omitempty" db:"simulcast"`
	Layers           json.RawMessage `json:"layers,omitempty" db:"layers"`
	Codecs           json.RawMessage `json:"codecs,omitempty" db:"codecs"`
	Stream           string          `json:"stream,omitempty" db:"stream"`
	Version          json.RawMessage `json:"version,omitempty" db:"version"`
	AudioFeatures    []string        `json:"audioFeatures,omitempty" db:"-"`
	AudioFeaturesJSON string         `json:"-" db:"audio_features"`
	BackupCodecPolicy string         `json:"backupCodecPolicy,omitempty" db:"backup_codec_policy"`
	Status           string          `json:"status" db:"status"` // published, unpublished
	PublishedAt      time.Time       `json:"publishedAt" db:"published_at"`
	UnpublishedAt    *time.Time      `json:"unpublishedAt,omitempty" db:"unpublished_at"`
	CreatedAtDB      time.Time       `json:"-" db:"created_at"`
	UpdatedAt        time.Time       `json:"-" db:"updated_at"`
}

// WebhookEventLog represents a log of all webhook events received
type WebhookEventLog struct {
	ID          string          `json:"id" db:"id"`
	EventType   string          `json:"event_type" db:"event_type"`
	EventID     string          `json:"event_id" db:"event_id"`
	RoomSID     string          `json:"room_sid,omitempty" db:"room_sid"`
	ParticipantSID string       `json:"participant_sid,omitempty" db:"participant_sid"`
	TrackSID    string          `json:"track_sid,omitempty" db:"track_sid"`
	Payload     json.RawMessage `json:"payload" db:"payload"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// WebhookRequest represents the incoming webhook payload
type WebhookRequest struct {
	Event       string                 `json:"event" binding:"required"`
	Room        map[string]interface{} `json:"room,omitempty"`
	Participant map[string]interface{} `json:"participant,omitempty"`
	Track       map[string]interface{} `json:"track,omitempty"`
	ID          string                 `json:"id"`
	CreatedAt   string                 `json:"createdAt"`
}

// WebhookResponse represents the response to a webhook
type WebhookResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

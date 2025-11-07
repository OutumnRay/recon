package models

import "time"

// JitsiSession represents a Jitsi Meet conference session
type JitsiSession struct {
	ID            string    `json:"id" db:"id"`
	RoomName      string    `json:"room_name" db:"room_name"`
	RoomURL       string    `json:"room_url" db:"room_url"`
	Status        JitsiSessionStatus `json:"status" db:"status"`
	RecordingID   string    `json:"recording_id,omitempty" db:"recording_id"`
	StartedAt     time.Time `json:"started_at" db:"started_at"`
	EndedAt       *time.Time `json:"ended_at,omitempty" db:"ended_at"`
	Participants  int       `json:"participants" db:"participants"`
}

// JitsiSessionStatus represents the status of a Jitsi session
type JitsiSessionStatus string

const (
	JitsiSessionStatusWaiting   JitsiSessionStatus = "waiting"
	JitsiSessionStatusRecording JitsiSessionStatus = "recording"
	JitsiSessionStatusCompleted JitsiSessionStatus = "completed"
	JitsiSessionStatusFailed    JitsiSessionStatus = "failed"
)

// JitsiAgentConfig represents configuration for the Jitsi agent
type JitsiAgentConfig struct {
	AgentID          string `json:"agent_id"`
	JitsiServerURL   string `json:"jitsi_server_url"`
	MaxConcurrentSessions int `json:"max_concurrent_sessions"`
	RecordingFormat  string `json:"recording_format"` // mp4, webm, etc.
	AudioOnly        bool   `json:"audio_only"`
}

// StartRecordingRequest represents a request to start recording a Jitsi session
type StartRecordingRequest struct {
	RoomName string `json:"room_name" binding:"required" example:"meeting-room-123"`
	RoomURL  string `json:"room_url" binding:"required" example:"https://meet.jit.si/meeting-room-123"`
	UserID   string `json:"user_id" binding:"required" example:"user-001"`
}

// StartRecordingResponse represents the response to starting a recording
type StartRecordingResponse struct {
	SessionID   string    `json:"session_id" example:"session-123"`
	Status      string    `json:"status" example:"recording"`
	StartedAt   time.Time `json:"started_at"`
	Message     string    `json:"message" example:"Recording started successfully"`
}

// StopRecordingRequest represents a request to stop recording
type StopRecordingRequest struct {
	SessionID string `json:"session_id" binding:"required" example:"session-123"`
}

// StopRecordingResponse represents the response to stopping a recording
type StopRecordingResponse struct {
	SessionID   string    `json:"session_id" example:"session-123"`
	RecordingID string    `json:"recording_id" example:"rec-456"`
	EndedAt     time.Time `json:"ended_at"`
	Message     string    `json:"message" example:"Recording stopped and saved"`
}

// WebRTCConnection represents a WebRTC connection state
type WebRTCConnection struct {
	SessionID    string    `json:"session_id"`
	PeerID       string    `json:"peer_id"`
	IsConnected  bool      `json:"is_connected"`
	StreamActive bool      `json:"stream_active"`
	ConnectedAt  time.Time `json:"connected_at"`
}

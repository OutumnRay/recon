package models

import (
	"time"

	"github.com/google/uuid"
)

// TranscriptionPhrase represents a single transcribed phrase from audio
type TranscriptionPhrase struct {
	ID                    uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TrackID               uuid.UUID `gorm:"type:uuid;not null;index:idx_transcription_phrases_track_id" json:"track_id"`
	UserID                uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	PhraseIndex           int       `gorm:"not null" json:"phrase_index"`
	StartTime             float64   `gorm:"type:numeric(10,3);not null" json:"start_time"`              // seconds from track start / секунды от начала трека
	EndTime               float64   `gorm:"type:numeric(10,3);not null" json:"end_time"`                // seconds from track start / секунды от начала трека
	AbsoluteStartTime     float64   `gorm:"type:numeric(10,3);not null" json:"absolute_start_time"`     // seconds from meeting start / секунды от начала встречи
	AbsoluteEndTime       float64   `gorm:"type:numeric(10,3);not null" json:"absolute_end_time"`       // seconds from meeting start / секунды от начала встречи
	Text                  string    `gorm:"type:text;not null" json:"text"`
	Confidence            *float64  `gorm:"type:numeric(5,4)" json:"confidence,omitempty"`
	Language              *string   `gorm:"type:varchar(10)" json:"language,omitempty"`
	CreatedAt             time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt             time.Time `gorm:"default:now()" json:"updated_at"`
}

// TableName specifies the table name for TranscriptionPhrase
func (TranscriptionPhrase) TableName() string {
	return "transcription_phrases"
}

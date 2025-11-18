package database

import (
	"encoding/json"
	"fmt"
	"time"

	"Recontext.online/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LiveKitRepository struct {
	db *DB
}

func NewLiveKitRepository(db *DB) *LiveKitRepository {
	return &LiveKitRepository{db: db}
}

// ============================================================================
// Room Operations
// ============================================================================

// CreateRoom creates a new room
func (r *LiveKitRepository) CreateRoom(room *models.Room) error {
	// Convert enabled codecs to JSON
	codecsJSON, err := json.Marshal(room.EnabledCodecs)
	if err != nil {
		return fmt.Errorf("failed to marshal enabled codecs: %w", err)
	}
	room.EnabledCodecsJSON = string(codecsJSON)

	// Use UPSERT to handle race condition where placeholder room may already exist
	// Update all fields except ID and SID when conflict occurs
	err = r.db.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "sid"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name",
			"status",
			"empty_timeout",
			"departure_timeout",
			"creation_time",
			"creation_time_ms",
			"turn_password",
			"enabled_codecs",
			"meeting_id",
			"egress_id",
			"updated_at",
		}),
	}).Create(room).Error

	if err != nil {
		return fmt.Errorf("failed to create room: %w", err)
	}

	return nil
}

// GetRoomBySID retrieves a room by SID
func (r *LiveKitRepository) GetRoomBySID(sid string) (*models.Room, error) {
	room := &models.Room{}

	err := r.db.DB.Where("sid = ?", sid).First(room).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("room not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	// Unmarshal enabled codecs from JSON string
	if len(room.EnabledCodecsJSON) > 0 {
		if err := json.Unmarshal([]byte(room.EnabledCodecsJSON), &room.EnabledCodecs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal enabled codecs: %w", err)
		}
	}

	return room, nil
}

// GetRoomByName retrieves a room by name
func (r *LiveKitRepository) GetRoomByName(name string) (*models.Room, error) {
	room := &models.Room{}

	err := r.db.DB.Where("name = ?", name).Order("created_at DESC").First(room).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("room not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	// Unmarshal enabled codecs from JSON string
	if len(room.EnabledCodecsJSON) > 0 {
		if err := json.Unmarshal([]byte(room.EnabledCodecsJSON), &room.EnabledCodecs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal enabled codecs: %w", err)
		}
	}

	return room, nil
}

// FinishRoom marks a room as finished
func (r *LiveKitRepository) FinishRoom(sid string) error {
	now := time.Now()
	result := r.db.DB.Model(&models.Room{}).Where("sid = ?", sid).Updates(map[string]interface{}{
		"status":      "finished",
		"finished_at": now,
		"updated_at":  now,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to finish room: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("room not found")
	}

	return nil
}

// ListRooms retrieves all rooms with optional status filter
func (r *LiveKitRepository) ListRooms(status string, limit, offset int) ([]*models.Room, error) {
	var rooms []*models.Room

	query := r.db.DB.Model(&models.Room{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query = query.Order("started_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&rooms).Error; err != nil {
		return nil, fmt.Errorf("failed to list rooms: %w", err)
	}

	// Unmarshal enabled codecs for each room
	for _, room := range rooms {
		if len(room.EnabledCodecsJSON) > 0 {
			if err := json.Unmarshal([]byte(room.EnabledCodecsJSON), &room.EnabledCodecs); err != nil {
				return nil, fmt.Errorf("failed to unmarshal enabled codecs: %w", err)
			}
		}
	}

	return rooms, nil
}

// ============================================================================
// Participant Operations
// ============================================================================

// CreateParticipant creates a new participant
func (r *LiveKitRepository) CreateParticipant(participant *models.Participant) error {
	err := r.db.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "sid"}},
		DoUpdates: clause.AssignmentColumns([]string{"state", "version", "is_publisher", "updated_at"}),
	}).Create(participant).Error

	if err != nil {
		return fmt.Errorf("failed to create participant: %w", err)
	}

	return nil
}

// UpdateParticipantLeft marks a participant as left
func (r *LiveKitRepository) UpdateParticipantLeft(sid, disconnectReason string) error {
	now := time.Now()
	result := r.db.DB.Model(&models.Participant{}).Where("sid = ?", sid).Updates(map[string]interface{}{
		"state":             "DISCONNECTED",
		"disconnect_reason": disconnectReason,
		"left_at":           now,
		"updated_at":        now,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to update participant: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("participant not found")
	}

	return nil
}

// GetParticipantBySID retrieves a participant by SID
func (r *LiveKitRepository) GetParticipantBySID(sid string) (*models.Participant, error) {
	participant := &models.Participant{}

	err := r.db.DB.Where("sid = ?", sid).First(participant).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("participant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}

	return participant, nil
}

// ListParticipantsByRoom retrieves all participants in a room
func (r *LiveKitRepository) ListParticipantsByRoom(roomSID string) ([]*models.Participant, error) {
	var participants []*models.Participant

	err := r.db.DB.Where("room_sid = ?", roomSID).Order("joined_at ASC").Find(&participants).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list participants: %w", err)
	}

	return participants, nil
}

// ============================================================================
// Track Operations
// ============================================================================

// CreateTrack creates a new track
func (r *LiveKitRepository) CreateTrack(track *models.Track) error {
	// Note: For AudioFeatures, we need to handle it manually as JSON
	// First, let's save the track using GORM
	err := r.db.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "sid"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "updated_at"}),
	}).Create(track).Error

	if err != nil {
		return fmt.Errorf("failed to create track: %w", err)
	}

	// If AudioFeatures is not empty, marshal to JSON and update
	if len(track.AudioFeatures) > 0 {
		audioFeaturesJSON, err := json.Marshal(track.AudioFeatures)
		if err != nil {
			return fmt.Errorf("failed to marshal audio features: %w", err)
		}
		fmt.Printf("[LiveKit Repository] Updating audio_features for track %s: %s -> JSON: %s\n", track.SID, track.AudioFeatures, string(audioFeaturesJSON))
		err = r.db.DB.Model(track).Update("audio_features", audioFeaturesJSON).Error
		if err != nil {
			fmt.Printf("[LiveKit Repository] ERROR updating audio_features: %v\n", err)
			return fmt.Errorf("failed to update audio features: %w", err)
		}
		fmt.Printf("[LiveKit Repository] Successfully updated audio_features for track %s\n", track.SID)
	}

	return nil
}

// UnpublishTrack marks a track as unpublished
func (r *LiveKitRepository) UnpublishTrack(sid string) error {
	now := time.Now()
	result := r.db.DB.Model(&models.Track{}).Where("sid = ?", sid).Updates(map[string]interface{}{
		"status":         "unpublished",
		"unpublished_at": now,
		"updated_at":     now,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to unpublish track: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("track not found")
	}

	return nil
}

// GetTrackBySID retrieves a track by SID
func (r *LiveKitRepository) GetTrackBySID(sid string) (*models.Track, error) {
	track := &models.Track{}

	err := r.db.DB.Where("sid = ?", sid).First(track).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("track not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get track: %w", err)
	}

	// Unmarshal AudioFeatures from JSON
	if track.AudioFeaturesJSON != nil && len(track.AudioFeaturesJSON) > 0 {
		err = json.Unmarshal(track.AudioFeaturesJSON, &track.AudioFeatures)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal audio features: %w", err)
		}
	}

	return track, nil
}

// ListTracksByParticipant retrieves all tracks for a participant
func (r *LiveKitRepository) ListTracksByParticipant(participantSID string) ([]*models.Track, error) {
	var tracks []*models.Track

	err := r.db.DB.Where("participant_sid = ?", participantSID).Order("published_at ASC").Find(&tracks).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list tracks: %w", err)
	}

	// Unmarshal AudioFeatures for each track
	for _, track := range tracks {
		if track.AudioFeaturesJSON != nil && len(track.AudioFeaturesJSON) > 0 {
			err = json.Unmarshal(track.AudioFeaturesJSON, &track.AudioFeatures)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal audio features: %w", err)
			}
		}
	}

	return tracks, nil
}

// ListTracksByRoom retrieves all tracks in a room
func (r *LiveKitRepository) ListTracksByRoom(roomSID string) ([]*models.Track, error) {
	var tracks []*models.Track

	err := r.db.DB.Where("room_sid = ?", roomSID).Order("published_at ASC").Find(&tracks).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list tracks: %w", err)
	}

	// Unmarshal AudioFeatures for each track
	for _, track := range tracks {
		if track.AudioFeaturesJSON != nil && len(track.AudioFeaturesJSON) > 0 {
			err = json.Unmarshal(track.AudioFeaturesJSON, &track.AudioFeatures)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal audio features: %w", err)
			}
		}
	}

	return tracks, nil
}

// ============================================================================
// Webhook Event Log Operations
// ============================================================================

// LogWebhookEvent logs a webhook event
func (r *LiveKitRepository) LogWebhookEvent(event *models.WebhookEventLog) error {
	err := r.db.DB.Create(event).Error
	if err != nil {
		return fmt.Errorf("failed to log webhook event: %w", err)
	}

	return nil
}

// GetWebhookEvents retrieves webhook events with filters
func (r *LiveKitRepository) GetWebhookEvents(eventType, roomSID string, limit, offset int) ([]*models.WebhookEventLog, error) {
	var events []*models.WebhookEventLog

	query := r.db.DB.Model(&models.WebhookEventLog{})

	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}

	if roomSID != "" {
		query = query.Where("room_sid = ?", roomSID)
	}

	query = query.Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to get webhook events: %w", err)
	}

	return events, nil
}

// ============================================================================
// Egress Operations
// ============================================================================

// UpdateEgressStatus updates the status of an egress recording
func (r *LiveKitRepository) UpdateEgressStatus(egressID string, status string) error {
	// Try to find egress in livekit_egress table
	var egress LiveKitEgress
	err := r.db.DB.Where("egress_id = ?", egressID).First(&egress).Error

	if err == nil {
		// Found in livekit_egress table, update it
		egress.Status = status
		if err := r.db.DB.Save(&egress).Error; err != nil {
			return fmt.Errorf("failed to update egress status in livekit_egress: %w", err)
		}
		return nil
	}

	// Not found in livekit_egress, just skip (not critical)
	return nil
}

// GetRoomsByName retrieves all rooms by name (meeting ID)
func (r *LiveKitRepository) GetRoomsByName(roomName string) ([]*models.Room, error) {
	var rooms []*models.Room

	err := r.db.DB.Where("name = ?", roomName).Order("created_at DESC").Find(&rooms).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get rooms by name: %w", err)
	}

	// Unmarshal enabled codecs for each room
	for _, room := range rooms {
		if len(room.EnabledCodecsJSON) > 0 {
			if err := json.Unmarshal([]byte(room.EnabledCodecsJSON), &room.EnabledCodecs); err != nil {
				return nil, fmt.Errorf("failed to unmarshal enabled codecs: %w", err)
			}
		}
	}

	return rooms, nil
}

// GetTracksByRoomSID retrieves all tracks for a room
func (r *LiveKitRepository) GetTracksByRoomSID(roomSID string) ([]*models.Track, error) {
	var tracks []*models.Track

	err := r.db.DB.Where("room_sid = ?", roomSID).Order("published_at ASC").Find(&tracks).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get tracks by room SID: %w", err)
	}

	// Unmarshal AudioFeatures for each track
	for _, track := range tracks {
		if track.AudioFeaturesJSON != nil && len(track.AudioFeaturesJSON) > 0 {
			err = json.Unmarshal(track.AudioFeaturesJSON, &track.AudioFeatures)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal audio features: %w", err)
			}
		}
	}

	return tracks, nil
}

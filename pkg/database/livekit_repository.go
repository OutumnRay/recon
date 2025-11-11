package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"Recontext.online/internal/models"
	"github.com/lib/pq"
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

	query := `
		INSERT INTO livekit_rooms (
			id, sid, name, empty_timeout, departure_timeout,
			creation_time, creation_time_ms, turn_password, enabled_codecs,
			status, started_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (sid) DO UPDATE SET
			name = EXCLUDED.name,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`

	_, err = r.db.Exec(
		query,
		room.ID,
		room.SID,
		room.Name,
		room.EmptyTimeout,
		room.DepartureTimeout,
		room.CreationTime,
		room.CreationTimeMs,
		room.TurnPassword,
		codecsJSON,
		room.Status,
		room.StartedAt,
		room.CreatedAtDB,
		room.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create room: %w", err)
	}

	return nil
}

// GetRoomBySID retrieves a room by SID
func (r *LiveKitRepository) GetRoomBySID(sid string) (*models.Room, error) {
	query := `
		SELECT id, sid, name, empty_timeout, departure_timeout,
			creation_time, creation_time_ms, turn_password, enabled_codecs,
			status, started_at, finished_at, created_at, updated_at
		FROM livekit_rooms
		WHERE sid = $1
	`

	room := &models.Room{}
	var finishedAt sql.NullTime
	var codecsJSON []byte

	err := r.db.QueryRow(query, sid).Scan(
		&room.ID,
		&room.SID,
		&room.Name,
		&room.EmptyTimeout,
		&room.DepartureTimeout,
		&room.CreationTime,
		&room.CreationTimeMs,
		&room.TurnPassword,
		&codecsJSON,
		&room.Status,
		&room.StartedAt,
		&finishedAt,
		&room.CreatedAtDB,
		&room.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("room not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	if finishedAt.Valid {
		room.FinishedAt = &finishedAt.Time
	}

	// Unmarshal enabled codecs
	if len(codecsJSON) > 0 {
		if err := json.Unmarshal(codecsJSON, &room.EnabledCodecs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal enabled codecs: %w", err)
		}
	}

	return room, nil
}

// GetRoomByName retrieves a room by name
func (r *LiveKitRepository) GetRoomByName(name string) (*models.Room, error) {
	query := `
		SELECT id, sid, name, empty_timeout, departure_timeout,
			creation_time, creation_time_ms, turn_password, enabled_codecs,
			status, started_at, finished_at, created_at, updated_at
		FROM livekit_rooms
		WHERE name = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	room := &models.Room{}
	var finishedAt sql.NullTime
	var codecsJSON []byte

	err := r.db.QueryRow(query, name).Scan(
		&room.ID,
		&room.SID,
		&room.Name,
		&room.EmptyTimeout,
		&room.DepartureTimeout,
		&room.CreationTime,
		&room.CreationTimeMs,
		&room.TurnPassword,
		&codecsJSON,
		&room.Status,
		&room.StartedAt,
		&finishedAt,
		&room.CreatedAtDB,
		&room.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("room not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	if finishedAt.Valid {
		room.FinishedAt = &finishedAt.Time
	}

	// Unmarshal enabled codecs
	if len(codecsJSON) > 0 {
		if err := json.Unmarshal(codecsJSON, &room.EnabledCodecs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal enabled codecs: %w", err)
		}
	}

	return room, nil
}

// FinishRoom marks a room as finished
func (r *LiveKitRepository) FinishRoom(sid string) error {
	query := `
		UPDATE livekit_rooms
		SET status = 'finished', finished_at = $2, updated_at = $3
		WHERE sid = $1
	`

	now := time.Now()
	result, err := r.db.Exec(query, sid, now, now)
	if err != nil {
		return fmt.Errorf("failed to finish room: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("room not found")
	}

	return nil
}

// ListRooms retrieves all rooms with optional status filter
func (r *LiveKitRepository) ListRooms(status string, limit, offset int) ([]*models.Room, error) {
	query := `
		SELECT id, sid, name, empty_timeout, departure_timeout,
			creation_time, creation_time_ms, turn_password, enabled_codecs,
			status, started_at, finished_at, created_at, updated_at
		FROM livekit_rooms
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	query += " ORDER BY started_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, limit)
		argIdx++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list rooms: %w", err)
	}
	defer rows.Close()

	rooms := []*models.Room{}
	for rows.Next() {
		room := &models.Room{}
		var finishedAt sql.NullTime
		var codecsJSON []byte

		err := rows.Scan(
			&room.ID,
			&room.SID,
			&room.Name,
			&room.EmptyTimeout,
			&room.DepartureTimeout,
			&room.CreationTime,
			&room.CreationTimeMs,
			&room.TurnPassword,
			&codecsJSON,
			&room.Status,
			&room.StartedAt,
			&finishedAt,
			&room.CreatedAtDB,
			&room.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}

		if finishedAt.Valid {
			room.FinishedAt = &finishedAt.Time
		}

		// Unmarshal enabled codecs
		if len(codecsJSON) > 0 {
			if err := json.Unmarshal(codecsJSON, &room.EnabledCodecs); err != nil {
				return nil, fmt.Errorf("failed to unmarshal enabled codecs: %w", err)
			}
		}

		rooms = append(rooms, room)
	}

	return rooms, nil
}

// ============================================================================
// Participant Operations
// ============================================================================

// CreateParticipant creates a new participant
func (r *LiveKitRepository) CreateParticipant(participant *models.Participant) error {
	query := `
		INSERT INTO livekit_participants (
			id, sid, room_sid, identity, name, state,
			joined_at, joined_at_ms, version, permission,
			is_publisher, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (sid) DO UPDATE SET
			state = EXCLUDED.state,
			version = EXCLUDED.version,
			is_publisher = EXCLUDED.is_publisher,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(
		query,
		participant.ID,
		participant.SID,
		participant.RoomSID,
		participant.Identity,
		participant.Name,
		participant.State,
		participant.JoinedAt,
		participant.JoinedAtMs,
		participant.Version,
		participant.Permission,
		participant.IsPublisher,
		participant.CreatedAtDB,
		participant.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create participant: %w", err)
	}

	return nil
}

// UpdateParticipantLeft marks a participant as left
func (r *LiveKitRepository) UpdateParticipantLeft(sid, disconnectReason string) error {
	query := `
		UPDATE livekit_participants
		SET state = 'DISCONNECTED',
		    disconnect_reason = $2,
		    left_at = $3,
		    updated_at = $4
		WHERE sid = $1
	`

	now := time.Now()
	result, err := r.db.Exec(query, sid, disconnectReason, now, now)
	if err != nil {
		return fmt.Errorf("failed to update participant: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("participant not found")
	}

	return nil
}

// GetParticipantBySID retrieves a participant by SID
func (r *LiveKitRepository) GetParticipantBySID(sid string) (*models.Participant, error) {
	query := `
		SELECT id, sid, room_sid, identity, name, state,
			joined_at, joined_at_ms, version, permission,
			is_publisher, disconnect_reason, left_at, created_at, updated_at
		FROM livekit_participants
		WHERE sid = $1
	`

	participant := &models.Participant{}
	var leftAt sql.NullTime

	err := r.db.QueryRow(query, sid).Scan(
		&participant.ID,
		&participant.SID,
		&participant.RoomSID,
		&participant.Identity,
		&participant.Name,
		&participant.State,
		&participant.JoinedAt,
		&participant.JoinedAtMs,
		&participant.Version,
		&participant.Permission,
		&participant.IsPublisher,
		&participant.DisconnectReason,
		&leftAt,
		&participant.CreatedAtDB,
		&participant.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("participant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}

	if leftAt.Valid {
		participant.LeftAt = &leftAt.Time
	}

	return participant, nil
}

// ListParticipantsByRoom retrieves all participants in a room
func (r *LiveKitRepository) ListParticipantsByRoom(roomSID string) ([]*models.Participant, error) {
	query := `
		SELECT id, sid, room_sid, identity, name, state,
			joined_at, joined_at_ms, version, permission,
			is_publisher, disconnect_reason, left_at, created_at, updated_at
		FROM livekit_participants
		WHERE room_sid = $1
		ORDER BY joined_at ASC
	`

	rows, err := r.db.Query(query, roomSID)
	if err != nil {
		return nil, fmt.Errorf("failed to list participants: %w", err)
	}
	defer rows.Close()

	participants := []*models.Participant{}
	for rows.Next() {
		participant := &models.Participant{}
		var leftAt sql.NullTime

		err := rows.Scan(
			&participant.ID,
			&participant.SID,
			&participant.RoomSID,
			&participant.Identity,
			&participant.Name,
			&participant.State,
			&participant.JoinedAt,
			&participant.JoinedAtMs,
			&participant.Version,
			&participant.Permission,
			&participant.IsPublisher,
			&participant.DisconnectReason,
			&leftAt,
			&participant.CreatedAtDB,
			&participant.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}

		if leftAt.Valid {
			participant.LeftAt = &leftAt.Time
		}

		participants = append(participants, participant)
	}

	return participants, nil
}

// ============================================================================
// Track Operations
// ============================================================================

// CreateTrack creates a new track
func (r *LiveKitRepository) CreateTrack(track *models.Track) error {
	query := `
		INSERT INTO livekit_tracks (
			id, sid, participant_sid, room_sid, type, source,
			mime_type, mid, width, height, simulcast,
			layers, codecs, stream, version, audio_features,
			backup_codec_policy, status, published_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		ON CONFLICT (sid) DO UPDATE SET
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(
		query,
		track.ID,
		track.SID,
		track.ParticipantSID,
		track.RoomSID,
		track.Type,
		track.Source,
		track.MimeType,
		track.Mid,
		track.Width,
		track.Height,
		track.Simulcast,
		track.Layers,
		track.Codecs,
		track.Stream,
		track.Version,
		pq.Array(track.AudioFeatures),
		track.BackupCodecPolicy,
		track.Status,
		track.PublishedAt,
		track.CreatedAtDB,
		track.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create track: %w", err)
	}

	return nil
}

// UnpublishTrack marks a track as unpublished
func (r *LiveKitRepository) UnpublishTrack(sid string) error {
	query := `
		UPDATE livekit_tracks
		SET status = 'unpublished', unpublished_at = $2, updated_at = $3
		WHERE sid = $1
	`

	now := time.Now()
	result, err := r.db.Exec(query, sid, now, now)
	if err != nil {
		return fmt.Errorf("failed to unpublish track: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("track not found")
	}

	return nil
}

// GetTrackBySID retrieves a track by SID
func (r *LiveKitRepository) GetTrackBySID(sid string) (*models.Track, error) {
	query := `
		SELECT id, sid, participant_sid, room_sid, type, source,
			mime_type, mid, width, height, simulcast,
			layers, codecs, stream, version, audio_features,
			backup_codec_policy, status, published_at, unpublished_at,
			created_at, updated_at
		FROM livekit_tracks
		WHERE sid = $1
	`

	track := &models.Track{}
	var unpublishedAt sql.NullTime

	err := r.db.QueryRow(query, sid).Scan(
		&track.ID,
		&track.SID,
		&track.ParticipantSID,
		&track.RoomSID,
		&track.Type,
		&track.Source,
		&track.MimeType,
		&track.Mid,
		&track.Width,
		&track.Height,
		&track.Simulcast,
		&track.Layers,
		&track.Codecs,
		&track.Stream,
		&track.Version,
		pq.Array(&track.AudioFeatures),
		&track.BackupCodecPolicy,
		&track.Status,
		&track.PublishedAt,
		&unpublishedAt,
		&track.CreatedAtDB,
		&track.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("track not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get track: %w", err)
	}

	if unpublishedAt.Valid {
		track.UnpublishedAt = &unpublishedAt.Time
	}

	return track, nil
}

// ListTracksByParticipant retrieves all tracks for a participant
func (r *LiveKitRepository) ListTracksByParticipant(participantSID string) ([]*models.Track, error) {
	query := `
		SELECT id, sid, participant_sid, room_sid, type, source,
			mime_type, mid, width, height, simulcast,
			layers, codecs, stream, version, audio_features,
			backup_codec_policy, status, published_at, unpublished_at,
			created_at, updated_at
		FROM livekit_tracks
		WHERE participant_sid = $1
		ORDER BY published_at ASC
	`

	rows, err := r.db.Query(query, participantSID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tracks: %w", err)
	}
	defer rows.Close()

	tracks := []*models.Track{}
	for rows.Next() {
		track := &models.Track{}
		var unpublishedAt sql.NullTime

		err := rows.Scan(
			&track.ID,
			&track.SID,
			&track.ParticipantSID,
			&track.RoomSID,
			&track.Type,
			&track.Source,
			&track.MimeType,
			&track.Mid,
			&track.Width,
			&track.Height,
			&track.Simulcast,
			&track.Layers,
			&track.Codecs,
			&track.Stream,
			&track.Version,
			pq.Array(&track.AudioFeatures),
			&track.BackupCodecPolicy,
			&track.Status,
			&track.PublishedAt,
			&unpublishedAt,
			&track.CreatedAtDB,
			&track.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan track: %w", err)
		}

		if unpublishedAt.Valid {
			track.UnpublishedAt = &unpublishedAt.Time
		}

		tracks = append(tracks, track)
	}

	return tracks, nil
}

// ListTracksByRoom retrieves all tracks in a room
func (r *LiveKitRepository) ListTracksByRoom(roomSID string) ([]*models.Track, error) {
	query := `
		SELECT id, sid, participant_sid, room_sid, type, source,
			mime_type, mid, width, height, simulcast,
			layers, codecs, stream, version, audio_features,
			backup_codec_policy, status, published_at, unpublished_at,
			created_at, updated_at
		FROM livekit_tracks
		WHERE room_sid = $1
		ORDER BY published_at ASC
	`

	rows, err := r.db.Query(query, roomSID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tracks: %w", err)
	}
	defer rows.Close()

	tracks := []*models.Track{}
	for rows.Next() {
		track := &models.Track{}
		var unpublishedAt sql.NullTime

		err := rows.Scan(
			&track.ID,
			&track.SID,
			&track.ParticipantSID,
			&track.RoomSID,
			&track.Type,
			&track.Source,
			&track.MimeType,
			&track.Mid,
			&track.Width,
			&track.Height,
			&track.Simulcast,
			&track.Layers,
			&track.Codecs,
			&track.Stream,
			&track.Version,
			pq.Array(&track.AudioFeatures),
			&track.BackupCodecPolicy,
			&track.Status,
			&track.PublishedAt,
			&unpublishedAt,
			&track.CreatedAtDB,
			&track.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan track: %w", err)
		}

		if unpublishedAt.Valid {
			track.UnpublishedAt = &unpublishedAt.Time
		}

		tracks = append(tracks, track)
	}

	return tracks, nil
}

// ============================================================================
// Webhook Event Log Operations
// ============================================================================

// LogWebhookEvent logs a webhook event
func (r *LiveKitRepository) LogWebhookEvent(event *models.WebhookEventLog) error {
	query := `
		INSERT INTO livekit_webhook_events (
			id, event_type, event_id, room_sid, participant_sid, track_sid, payload, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(
		query,
		event.ID,
		event.EventType,
		event.EventID,
		event.RoomSID,
		event.ParticipantSID,
		event.TrackSID,
		event.Payload,
		event.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to log webhook event: %w", err)
	}

	return nil
}

// GetWebhookEvents retrieves webhook events with filters
func (r *LiveKitRepository) GetWebhookEvents(eventType, roomSID string, limit, offset int) ([]*models.WebhookEventLog, error) {
	query := `
		SELECT id, event_type, event_id, room_sid, participant_sid, track_sid, payload, created_at
		FROM livekit_webhook_events
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if eventType != "" {
		query += fmt.Sprintf(" AND event_type = $%d", argIdx)
		args = append(args, eventType)
		argIdx++
	}

	if roomSID != "" {
		query += fmt.Sprintf(" AND room_sid = $%d", argIdx)
		args = append(args, roomSID)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, limit)
		argIdx++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook events: %w", err)
	}
	defer rows.Close()

	events := []*models.WebhookEventLog{}
	for rows.Next() {
		event := &models.WebhookEventLog{}
		var participantSID, trackSID sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.EventID,
			&event.RoomSID,
			&participantSID,
			&trackSID,
			&event.Payload,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook event: %w", err)
		}

		if participantSID.Valid {
			event.ParticipantSID = participantSID.String
		}
		if trackSID.Valid {
			event.TrackSID = trackSID.String
		}

		events = append(events, event)
	}

	return events, nil
}

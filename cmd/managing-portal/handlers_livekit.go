package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"Recontext.online/internal/models"
	"github.com/google/uuid"
)

// LiveKitWebhook godoc
// @Summary LiveKit webhook endpoint
// @Description Receives webhook events from LiveKit server (room_started, participant_joined, track_published, track_unpublished, participant_left, room_finished)
// @Tags LiveKit
// @Accept json
// @Produce json
// @Param webhook body models.WebhookRequest true "Webhook payload"
// @Success 200 {object} models.WebhookResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /webhook/meet [post]
func (mp *ManagingPortal) liveKitWebhookHandler(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		mp.logger.Errorf("Failed to read webhook body: %s", err.Error())
		mp.respondWithError(w, http.StatusBadRequest, "Failed to read request body", err.Error())
		return
	}
	defer r.Body.Close()

	// Parse webhook event
	var webhookReq models.WebhookRequest
	if err := json.Unmarshal(body, &webhookReq); err != nil {
		mp.logger.Errorf("Failed to parse webhook JSON: %s", err.Error())
		mp.respondWithError(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}

	mp.logger.Infof("Received LiveKit webhook event: %s (ID: %s)", webhookReq.Event, webhookReq.ID)

	// Log the raw event to database
	eventLog := &models.WebhookEventLog{
		ID:        uuid.New().String(),
		EventType: webhookReq.Event,
		EventID:   webhookReq.ID,
		Payload:   body,
		CreatedAt: time.Now(),
	}

	// Extract IDs from webhook
	if webhookReq.Room != nil {
		if sid, ok := webhookReq.Room["sid"].(string); ok {
			eventLog.RoomSID = sid
		}
	}
	if webhookReq.Participant != nil {
		if sid, ok := webhookReq.Participant["sid"].(string); ok {
			eventLog.ParticipantSID = sid
		}
	}
	if webhookReq.Track != nil {
		if sid, ok := webhookReq.Track["sid"].(string); ok {
			eventLog.TrackSID = sid
		}
	}

	// Save event log
	if err := mp.liveKitRepo.LogWebhookEvent(eventLog); err != nil {
		mp.logger.Errorf("Failed to log webhook event: %s", err.Error())
		// Continue processing even if logging fails
	}

	// Process event based on type
	switch webhookReq.Event {
	case "room_started":
		err = mp.handleRoomStarted(webhookReq)
	case "participant_joined":
		err = mp.handleParticipantJoined(webhookReq)
	case "track_published":
		err = mp.handleTrackPublished(webhookReq)
	case "track_unpublished":
		err = mp.handleTrackUnpublished(webhookReq)
	case "participant_left":
		err = mp.handleParticipantLeft(webhookReq)
	case "room_finished":
		err = mp.handleRoomFinished(webhookReq)
	default:
		mp.logger.Errorf("Unknown webhook event type: %s", webhookReq.Event)
		mp.respondWithError(w, http.StatusBadRequest, "Unknown event type", "")
		return
	}

	if err != nil {
		mp.logger.Errorf("Failed to process %s event: %s", webhookReq.Event, err.Error())
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to process event", err.Error())
		return
	}

	// Return success response
	response := models.WebhookResponse{
		Status:  "success",
		Message: fmt.Sprintf("Event %s processed successfully", webhookReq.Event),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (mp *ManagingPortal) handleRoomStarted(req models.WebhookRequest) error {
	if req.Room == nil {
		return fmt.Errorf("room data is missing")
	}

	room := &models.Room{
		ID:     uuid.New().String(),
		Status: "active",
		StartedAt: time.Now(),
		CreatedAtDB: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Extract room data
	if sid, ok := req.Room["sid"].(string); ok {
		room.SID = sid
	}
	if name, ok := req.Room["name"].(string); ok {
		room.Name = name
	}
	if emptyTimeout, ok := req.Room["emptyTimeout"].(float64); ok {
		room.EmptyTimeout = int(emptyTimeout)
	}
	if departureTimeout, ok := req.Room["departureTimeout"].(float64); ok {
		room.DepartureTimeout = int(departureTimeout)
	}
	if creationTime, ok := req.Room["creationTime"].(string); ok {
		room.CreationTime = creationTime
	}
	if creationTimeMs, ok := req.Room["creationTimeMs"].(string); ok {
		room.CreationTimeMs = creationTimeMs
	}
	if turnPassword, ok := req.Room["turnPassword"].(string); ok {
		room.TurnPassword = turnPassword
	}

	// Parse enabled codecs
	if enabledCodecs, ok := req.Room["enabledCodecs"].([]interface{}); ok {
		for _, codec := range enabledCodecs {
			if codecMap, ok := codec.(map[string]interface{}); ok {
				if mime, ok := codecMap["mime"].(string); ok {
					room.EnabledCodecs = append(room.EnabledCodecs, models.EnabledCodec{Mime: mime})
				}
			}
		}
	}

	return mp.liveKitRepo.CreateRoom(room)
}

func (mp *ManagingPortal) handleParticipantJoined(req models.WebhookRequest) error {
	if req.Participant == nil {
		return fmt.Errorf("participant data is missing")
	}

	participant := &models.Participant{
		ID:    uuid.New().String(),
		State: "ACTIVE",
		CreatedAtDB: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Extract participant data
	if sid, ok := req.Participant["sid"].(string); ok {
		participant.SID = sid
	}
	if identity, ok := req.Participant["identity"].(string); ok {
		participant.Identity = identity
	}
	if name, ok := req.Participant["name"].(string); ok {
		participant.Name = name
	}
	if state, ok := req.Participant["state"].(string); ok {
		participant.State = state
	}
	if joinedAt, ok := req.Participant["joinedAt"].(string); ok {
		participant.JoinedAt = joinedAt
	}
	if joinedAtMs, ok := req.Participant["joinedAtMs"].(string); ok {
		participant.JoinedAtMs = joinedAtMs
	}
	if version, ok := req.Participant["version"].(float64); ok {
		participant.Version = int(version)
	}
	if isPublisher, ok := req.Participant["isPublisher"].(bool); ok {
		participant.IsPublisher = isPublisher
	}

	// Marshal permission as JSON
	if permission, ok := req.Participant["permission"]; ok {
		permJSON, _ := json.Marshal(permission)
		participant.Permission = permJSON
	}

	// Extract room SID
	if req.Room != nil {
		if roomSID, ok := req.Room["sid"].(string); ok {
			participant.RoomSID = roomSID
		}
	}

	return mp.liveKitRepo.CreateParticipant(participant)
}

func (mp *ManagingPortal) handleTrackPublished(req models.WebhookRequest) error {
	if req.Track == nil {
		return fmt.Errorf("track data is missing")
	}

	track := &models.Track{
		ID:     uuid.New().String(),
		Status: "published",
		PublishedAt: time.Now(),
		CreatedAtDB: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Extract track data
	if sid, ok := req.Track["sid"].(string); ok {
		track.SID = sid
	}
	if trackType, ok := req.Track["type"].(string); ok {
		track.Type = trackType
	}
	if source, ok := req.Track["source"].(string); ok {
		track.Source = source
	}
	if mimeType, ok := req.Track["mimeType"].(string); ok {
		track.MimeType = mimeType
	}
	if mid, ok := req.Track["mid"].(string); ok {
		track.Mid = mid
	}
	if width, ok := req.Track["width"].(float64); ok {
		track.Width = int(width)
	}
	if height, ok := req.Track["height"].(float64); ok {
		track.Height = int(height)
	}
	if simulcast, ok := req.Track["simulcast"].(bool); ok {
		track.Simulcast = simulcast
	}
	if stream, ok := req.Track["stream"].(string); ok {
		track.Stream = stream
	}
	if backupCodecPolicy, ok := req.Track["backupCodecPolicy"].(string); ok {
		track.BackupCodecPolicy = backupCodecPolicy
	}

	// Marshal layers as JSON
	if layers, ok := req.Track["layers"]; ok {
		layersJSON, _ := json.Marshal(layers)
		track.Layers = layersJSON
	}

	// Marshal codecs as JSON
	if codecs, ok := req.Track["codecs"]; ok {
		codecsJSON, _ := json.Marshal(codecs)
		track.Codecs = codecsJSON
	}

	// Marshal version as JSON
	if version, ok := req.Track["version"]; ok {
		versionJSON, _ := json.Marshal(version)
		track.Version = versionJSON
	}

	// Extract audio features
	if audioFeatures, ok := req.Track["audioFeatures"].([]interface{}); ok {
		for _, feature := range audioFeatures {
			if featureStr, ok := feature.(string); ok {
				track.AudioFeatures = append(track.AudioFeatures, featureStr)
			}
		}
	}

	// Extract participant SID
	if req.Participant != nil {
		if participantSID, ok := req.Participant["sid"].(string); ok {
			track.ParticipantSID = participantSID
		}
	}

	// Extract room SID
	if req.Room != nil {
		if roomSID, ok := req.Room["sid"].(string); ok {
			track.RoomSID = roomSID
		}
	}

	return mp.liveKitRepo.CreateTrack(track)
}

func (mp *ManagingPortal) handleTrackUnpublished(req models.WebhookRequest) error {
	if req.Track == nil {
		return fmt.Errorf("track data is missing")
	}

	var trackSID string
	if sid, ok := req.Track["sid"].(string); ok {
		trackSID = sid
	} else {
		return fmt.Errorf("track SID is missing")
	}

	return mp.liveKitRepo.UnpublishTrack(trackSID)
}

func (mp *ManagingPortal) handleParticipantLeft(req models.WebhookRequest) error {
	if req.Participant == nil {
		return fmt.Errorf("participant data is missing")
	}

	var participantSID string
	if sid, ok := req.Participant["sid"].(string); ok {
		participantSID = sid
	} else {
		return fmt.Errorf("participant SID is missing")
	}

	var disconnectReason string
	if reason, ok := req.Participant["disconnectReason"].(string); ok {
		disconnectReason = reason
	}

	return mp.liveKitRepo.UpdateParticipantLeft(participantSID, disconnectReason)
}

func (mp *ManagingPortal) handleRoomFinished(req models.WebhookRequest) error {
	if req.Room == nil {
		return fmt.Errorf("room data is missing")
	}

	var roomSID string
	if sid, ok := req.Room["sid"].(string); ok {
		roomSID = sid
	} else {
		return fmt.Errorf("room SID is missing")
	}

	return mp.liveKitRepo.FinishRoom(roomSID)
}

// GetLiveKitRooms godoc
// @Summary Get LiveKit rooms
// @Description Get a list of LiveKit rooms with optional status filter
// @Tags LiveKit
// @Produce json
// @Param status query string false "Filter by status (active, finished)"
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} models.Room
// @Security BearerAuth
// @Router /api/v1/livekit/rooms [get]
func (mp *ManagingPortal) listLiveKitRoomsHandler(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	rooms, err := mp.liveKitRepo.ListRooms(status, limit, offset)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to list rooms", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

// GetLiveKitRoom godoc
// @Summary Get LiveKit room details
// @Description Get detailed information about a specific room
// @Tags LiveKit
// @Produce json
// @Param sid path string true "Room SID"
// @Success 200 {object} models.Room
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/livekit/rooms/{sid} [get]
func (mp *ManagingPortal) getLiveKitRoomHandler(w http.ResponseWriter, r *http.Request) {
	sid := r.URL.Query().Get("sid")
	if sid == "" {
		mp.respondWithError(w, http.StatusBadRequest, "Room SID is required", "")
		return
	}

	room, err := mp.liveKitRepo.GetRoomBySID(sid)
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "Room not found", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(room)
}

// GetLiveKitParticipants godoc
// @Summary Get room participants
// @Description Get all participants for a specific room
// @Tags LiveKit
// @Produce json
// @Param room_sid query string true "Room SID"
// @Success 200 {array} models.Participant
// @Security BearerAuth
// @Router /api/v1/livekit/participants [get]
func (mp *ManagingPortal) listLiveKitParticipantsHandler(w http.ResponseWriter, r *http.Request) {
	roomSID := r.URL.Query().Get("room_sid")
	if roomSID == "" {
		mp.respondWithError(w, http.StatusBadRequest, "Room SID is required", "")
		return
	}

	participants, err := mp.liveKitRepo.ListParticipantsByRoom(roomSID)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to list participants", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(participants)
}

// GetLiveKitTracks godoc
// @Summary Get room tracks
// @Description Get all tracks for a specific room
// @Tags LiveKit
// @Produce json
// @Param room_sid query string true "Room SID"
// @Success 200 {array} models.Track
// @Security BearerAuth
// @Router /api/v1/livekit/tracks [get]
func (mp *ManagingPortal) listLiveKitTracksHandler(w http.ResponseWriter, r *http.Request) {
	roomSID := r.URL.Query().Get("room_sid")
	if roomSID == "" {
		mp.respondWithError(w, http.StatusBadRequest, "Room SID is required", "")
		return
	}

	tracks, err := mp.liveKitRepo.ListTracksByRoom(roomSID)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to list tracks", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tracks)
}

// GetWebhookEvents godoc
// @Summary Get webhook event logs
// @Description Get a list of webhook events with optional filters
// @Tags LiveKit
// @Produce json
// @Param event_type query string false "Filter by event type"
// @Param room_sid query string false "Filter by room SID"
// @Param limit query int false "Limit results" default(100)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} models.WebhookEventLog
// @Security BearerAuth
// @Router /api/v1/livekit/webhook-events [get]
func (mp *ManagingPortal) listWebhookEventsHandler(w http.ResponseWriter, r *http.Request) {
	eventType := r.URL.Query().Get("event_type")
	roomSID := r.URL.Query().Get("room_sid")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 100
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	events, err := mp.liveKitRepo.GetWebhookEvents(eventType, roomSID, limit, offset)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to list webhook events", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

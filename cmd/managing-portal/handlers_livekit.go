package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/database"
	"github.com/google/uuid"
)

// LiveKitWebhook godoc
// @Summary Конечная точка webhook LiveKit
// @Description Получает события webhook от сервера LiveKit (room_started, participant_joined, track_published, track_unpublished, participant_left, room_finished)
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
		ID:        uuid.New(),
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
		ID:     uuid.New(),
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

	// Save room to database
	if err := mp.liveKitRepo.CreateRoom(room); err != nil {
		return err
	}

	// Auto-start room composite recording if egressClient is configured
	if mp.egressClient != nil && room.Name != "" {
		mp.logger.Infof("Auto-starting room composite recording for room: %s", room.Name)
		go mp.autoStartRoomRecording(room.SID, room.Name)
	}

	return nil
}

func (mp *ManagingPortal) handleParticipantJoined(req models.WebhookRequest) error {
	if req.Participant == nil {
		return fmt.Errorf("participant data is missing")
	}

	participant := &models.Participant{
		ID:    uuid.New(),
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
		ID:     uuid.New(),
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

	// Stop all active egress sessions for this room
	if mp.egressClient != nil {
		mp.logger.Infof("Auto-stopping active recordings for room: %s", roomSID)
		go mp.autoStopRoomRecordings(roomSID)
	}

	return mp.liveKitRepo.FinishRoom(roomSID)
}

// GetLiveKitRooms godoc
// @Summary Получить комнаты LiveKit
// @Description Получить список комнат LiveKit с дополнительным фильтром по статусу
// @Tags LiveKit
// @Produce json
// @Param status query string false "Filter by status (active, finished)"
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} map[string]interface{}
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

	// Return standardized response format
	response := map[string]interface{}{
		"items":     rooms,
		"total":     len(rooms),
		"offset":    offset,
		"page_size": limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetLiveKitRoom godoc
// @Summary Получить детали комнаты LiveKit
// @Description Получить детальную информацию о конкретной комнате
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
// @Summary Получить участников комнаты
// @Description Получить всех участников для конкретной комнаты
// @Tags LiveKit
// @Produce json
// @Param room_sid query string true "Room SID"
// @Success 200 {object} map[string]interface{}
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

	// Return standardized response format
	response := map[string]interface{}{
		"items":     participants,
		"total":     len(participants),
		"offset":    0,
		"page_size": len(participants),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetLiveKitTracks godoc
// @Summary Получить треки комнаты
// @Description Получить все треки для конкретной комнаты
// @Tags LiveKit
// @Produce json
// @Param room_sid query string true "Room SID"
// @Success 200 {object} map[string]interface{}
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

	// Return standardized response format
	response := map[string]interface{}{
		"items":     tracks,
		"total":     len(tracks),
		"offset":    0,
		"page_size": len(tracks),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetWebhookEvents godoc
// @Summary Получить логи событий webhook
// @Description Получить список событий webhook с дополнительными фильтрами
// @Tags LiveKit
// @Produce json
// @Param event_type query string false "Filter by event type"
// @Param room_sid query string false "Filter by room SID"
// @Param limit query int false "Limit results" default(100)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} map[string]interface{}
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

	// Return standardized response format
	response := map[string]interface{}{
		"items":     events,
		"total":     len(events),
		"offset":    offset,
		"page_size": limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ============================================================================
// Auto-recording helper functions
// ============================================================================

// autoStartRoomRecording automatically starts recording when a room starts
func (mp *ManagingPortal) autoStartRoomRecording(roomSID string, roomName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mp.logger.Infof("Starting automatic room recording for: %s (%s)", roomName, roomSID)

	// Start room composite egress (audio + video)
	egressInfo, err := mp.egressClient.StartRoomCompositeEgress(ctx, roomName, false)
	if err != nil {
		mp.logger.Errorf("Failed to auto-start room recording: %v", err)
		return
	}

	mp.logger.Infof("Room recording started successfully. Egress ID: %s", egressInfo.EgressId)

	// Store egress info in database
	egressID, err := uuid.Parse(egressInfo.EgressId)
	if err != nil {
		mp.logger.Errorf("Invalid egress ID format: %v", err)
		return
	}

	egress := &database.LiveKitEgress{
		ID:        egressID,
		RoomSID:   roomSID,
		RoomName:  roomName,
		Type:      "room_composite",
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := mp.egressRepo.Create(egress); err != nil {
		mp.logger.Errorf("Failed to store egress in database: %v", err)
	}
}

// autoStopRoomRecordings automatically stops all active recordings when room ends
func (mp *ManagingPortal) autoStopRoomRecordings(roomSID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mp.logger.Infof("Stopping active recordings for room: %s", roomSID)

	// Get all active egress sessions for this room
	egresses, err := mp.egressRepo.GetActiveByRoomSID(roomSID)
	if err != nil {
		mp.logger.Errorf("Failed to get active egress sessions: %v", err)
		return
	}

	if len(egresses) == 0 {
		mp.logger.Infof("No active recordings found for room: %s", roomSID)
		return
	}

	// Stop each egress session
	for _, egress := range egresses {
		mp.logger.Infof("Stopping egress: %s", egress.ID)

		_, err := mp.egressClient.StopEgress(ctx, egress.ID.String())
		if err != nil {
			mp.logger.Errorf("Failed to stop egress %s: %v", egress.ID, err)
			continue
		}

		// Update status in database
		if err := mp.egressRepo.UpdateStatus(egress.ID, "finishing"); err != nil {
			mp.logger.Errorf("Failed to update egress status: %v", err)
		}

		mp.logger.Infof("Egress %s stopped successfully", egress.ID)
	}
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"Recontext.online/pkg/database"
)

// startRoomRecordingHandler godoc
// @Summary Start recording entire room
// @Description Starts recording the entire room with composite view
// @Tags LiveKit Egress
// @Accept json
// @Produce json
// @Param request body StartRoomRecordingRequest true "Room recording request"
// @Success 200 {object} map[string]interface{} "Egress started"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/livekit/egress/room/start [post]
func (mp *ManagingPortal) startRoomRecordingHandler(w http.ResponseWriter, r *http.Request) {
	var req StartRoomRecordingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if req.RoomName == "" {
		mp.respondWithError(w, http.StatusBadRequest, "room_name is required", "")
		return
	}

	// Get room by name to get room_sid
	room, err := mp.liveKitRepo.GetRoomByName(req.RoomName)
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "Room not found", err.Error())
		return
	}

	// Check if room is active
	if room.Status != "active" {
		mp.respondWithError(w, http.StatusBadRequest, "Room is not active", "")
		return
	}

	// Start room composite egress
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	audioOnly := req.AudioOnly != nil && *req.AudioOnly
	egressInfo, err := mp.egressClient.StartRoomCompositeEgress(ctx, req.RoomName, audioOnly)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to start room recording", err.Error())
		return
	}

	// Store egress info in database
	egressID, err := uuid.Parse(egressInfo.EgressId)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Invalid egress ID format", err.Error())
		return
	}

	egress := &database.LiveKitEgress{
		ID:        egressID,
		RoomSID:   room.SID,
		RoomName:  req.RoomName,
		Type:      "room_composite",
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := mp.egressRepo.Create(egress); err != nil {
		log.Printf("Warning: Failed to store egress in database: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"egress_id": egressInfo.EgressId,
		"room_sid":  room.SID,
		"room_name": req.RoomName,
		"status":    egressInfo.Status.String(),
		"message":   "Room recording started successfully",
	})
}

// stopRoomRecordingHandler godoc
// @Summary Stop room recording
// @Description Stops an ongoing room recording
// @Tags LiveKit Egress
// @Accept json
// @Produce json
// @Param request body StopEgressRequest true "Stop egress request"
// @Success 200 {object} map[string]interface{} "Egress stopped"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/livekit/egress/stop [post]
func (mp *ManagingPortal) stopRoomRecordingHandler(w http.ResponseWriter, r *http.Request) {
	var req StopEgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.EgressID == "" {
		mp.respondWithError(w, http.StatusBadRequest, "egress_id is required", "")
		return
	}

	// Stop egress
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	egressInfo, err := mp.egressClient.StopEgress(ctx, req.EgressID)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to stop recording", err.Error())
		return
	}

	// Update status in database
	if err := mp.egressRepo.UpdateStatus(uuid.Must(uuid.Parse(req.EgressID)), "finishing"); err != nil {
		log.Printf("Warning: Failed to update egress status in database: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"egress_id": egressInfo.EgressId,
		"status":    egressInfo.Status.String(),
		"message":   "Recording stopped successfully",
	})
}

// startTrackRecordingHandler godoc
// @Summary Start recording specific track
// @Description Starts recording a specific audio or video track
// @Tags LiveKit Egress
// @Accept json
// @Produce json
// @Param request body StartTrackRecordingRequest true "Track recording request"
// @Success 200 {object} map[string]interface{} "Track recording started"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/livekit/egress/track/start [post]
func (mp *ManagingPortal) startTrackRecordingHandler(w http.ResponseWriter, r *http.Request) {
	var req StartTrackRecordingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if req.RoomName == "" || (req.AudioTrackID == "" && req.VideoTrackID == "") {
		mp.respondWithError(w, http.StatusBadRequest, "room_name and at least one track_id are required", "")
		return
	}

	// Get room by name
	room, err := mp.liveKitRepo.GetRoomByName(req.RoomName)
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "Room not found", err.Error())
		return
	}

	// Start track composite egress
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	audioTrackID := req.AudioTrackID
	if audioTrackID == "" {
		audioTrackID = ""
	}

	videoTrackID := req.VideoTrackID
	if videoTrackID == "" {
		videoTrackID = ""
	}

	egressInfo, err := mp.egressClient.StartTrackCompositeEgress(ctx, req.RoomName, audioTrackID, videoTrackID)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to start track recording", err.Error())
		return
	}

	// Store egress info in database
	egressID, err := uuid.Parse(egressInfo.EgressId)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Invalid egress ID format", err.Error())
		return
	}

	egress := &database.LiveKitEgress{
		ID:       egressID,
		RoomSID:  room.SID,
		RoomName: req.RoomName,
		Type:     "track_composite",
		Status:   "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if audioTrackID != "" {
		egress.AudioTrackID = &audioTrackID
	}
	if videoTrackID != "" {
		egress.VideoTrackID = &videoTrackID
	}

	if err := mp.egressRepo.Create(egress); err != nil {
		log.Printf("Warning: Failed to store egress in database: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"egress_id": egressInfo.EgressId,
		"room_sid":  room.SID,
		"room_name": req.RoomName,
		"status":    egressInfo.Status.String(),
		"message":   "Track recording started successfully",
	})
}

// listEgressHandler godoc
// @Summary List egress sessions
// @Description List all egress sessions with optional filters
// @Tags LiveKit Egress
// @Produce json
// @Param room_name query string false "Filter by room name"
// @Param status query string false "Filter by status (pending, active, finishing, complete, failed)"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{} "List of egress sessions"
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/livekit/egress [get]
func (mp *ManagingPortal) listEgressHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	var roomName *string
	if rn := queryParams.Get("room_name"); rn != "" {
		roomName = &rn
	}

	var status *string
	if s := queryParams.Get("status"); s != "" {
		status = &s
	}

	limit := 50
	if l := queryParams.Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	offset := 0
	if o := queryParams.Get("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	egresses, total, err := mp.egressRepo.List(roomName, status, limit, offset)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to list egress sessions", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items":  egresses,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// Request/Response models

type StartRoomRecordingRequest struct {
	RoomName  string `json:"room_name"`
	AudioOnly *bool  `json:"audio_only,omitempty"`
}

type StartTrackRecordingRequest struct {
	RoomName     string `json:"room_name"`
	AudioTrackID string `json:"audio_track_id,omitempty"`
	VideoTrackID string `json:"video_track_id,omitempty"`
}

type StopEgressRequest struct {
	EgressID string `json:"egress_id"`
}

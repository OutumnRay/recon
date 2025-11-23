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


// startTrackRecordingHandler запускает запись конкретного трека (аудио или видео)
// Создаёт egress для указанного трека, сохраняет информацию в базе данных
// @Summary Начать запись конкретного трека
// @Description Начинает запись конкретного аудио или видео трека
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

	egressInfo, err := mp.egressClient.StartTrackCompositeEgress(ctx, req.RoomName, room.SID, audioTrackID, videoTrackID)
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

// startRoomRecordingHandler запускает композитную запись всей комнаты
// Объединяет все треки комнаты в одну запись, поддерживает режим только аудио
// @Summary Начать запись всей комнаты
// @Description Начинает запись всей комнаты (composite)
// @Tags LiveKit Egress
// @Accept json
// @Produce json
// @Param request body StartRoomRecordingRequest true "Room recording request"
// @Success 200 {object} map[string]interface{} "Room recording started"
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

	// Get room by name
	room, err := mp.liveKitRepo.GetRoomByName(req.RoomName)
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "Room not found", err.Error())
		return
	}

	// Determine if audio only
	audioOnly := false
	if req.AudioOnly != nil {
		audioOnly = *req.AudioOnly
	}

	// Start room composite egress
	egressID, err := mp.startRoomCompositeEgress(req.RoomName, room.SID, audioOnly)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to start room recording", err.Error())
		return
	}

	// Store egress info in database
	egressUUID, err := uuid.Parse(egressID)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Invalid egress ID format", err.Error())
		return
	}

	egress := &database.LiveKitEgress{
		ID:       egressUUID,
		RoomSID:  room.SID,
		RoomName: req.RoomName,
		Type:     "room_composite",
		Status:   "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := mp.egressRepo.Create(egress); err != nil {
		log.Printf("Warning: Failed to store egress in database: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"egress_id": egressID,
		"room_sid":  room.SID,
		"room_name": req.RoomName,
		"audio_only": audioOnly,
		"message":   "Room recording started successfully",
	})
}

// stopRoomRecordingHandler останавливает активную сессию записи egress
// Завершает запись, обновляет статус в базе данных
// @Summary Остановить запись (egress)
// @Description Останавливает активную сессию egress
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

	// Validate request
	if req.EgressID == "" {
		mp.respondWithError(w, http.StatusBadRequest, "egress_id is required", "")
		return
	}

	// Stop the egress
	if err := mp.stopEgress(req.EgressID); err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to stop egress", err.Error())
		return
	}

	// Update egress status in database
	egressUUID, err := uuid.Parse(req.EgressID)
	if err != nil {
		log.Printf("Warning: Invalid egress ID format: %v", err)
	} else {
		if err := mp.egressRepo.UpdateStatus(egressUUID, "stopped"); err != nil {
			log.Printf("Warning: Failed to update egress status in database: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"egress_id": req.EgressID,
		"message":   "Egress stopped successfully",
	})
}

// listEgressHandler возвращает список всех сессий egress с фильтрацией
// Поддерживает фильтрацию по комнате, статусу, пагинацию
// @Summary Список сессий egress
// @Description Получить список всех сессий egress с дополнительными фильтрами
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

// StartRoomRecordingRequest represents a request to start recording an entire room
type StartRoomRecordingRequest struct {
	RoomName  string `json:"room_name"`
	AudioOnly *bool  `json:"audio_only,omitempty"`
}

// StartTrackRecordingRequest represents a request to start recording specific tracks
type StartTrackRecordingRequest struct {
	RoomName     string `json:"room_name"`
	AudioTrackID string `json:"audio_track_id,omitempty"`
	VideoTrackID string `json:"video_track_id,omitempty"`
}

// StopEgressRequest represents a request to stop an egress session
type StopEgressRequest struct {
	EgressID string `json:"egress_id"`
}

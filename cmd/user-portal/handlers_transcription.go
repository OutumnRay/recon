package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
)

// ForceTranscribeTrack godoc
// @Summary Force transcription for a track
// @Description Manually triggers transcription for a specific track recording
// @Tags Tracks
// @Accept json
// @Produce json
// @Param id path string true "Track SID (e.g., TR_xxxxx)"
// @Success 200 {object} map[string]interface{} "Transcription started"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/tracks/{id}/transcribe [post]
func (up *UserPortal) forceTranscribeTrackHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 || pathParts[0] != "api" || pathParts[1] != "tracks" {
		up.respondWithError(w, http.StatusBadRequest, "Invalid URL format", "Expected /api/tracks/{sid}/transcribe")
		return
	}

	trackSID := pathParts[2]
	if trackSID == "" {
		up.respondWithError(w, http.StatusBadRequest, "Track SID is required", "")
		return
	}

	var track models.Track
	err := up.db.DB.Where("sid = ?", trackSID).First(&track).Error
	if err != nil {
		up.logger.Errorf("[Transcription] Track not found: %s, error: %v", trackSID, err)
		up.respondWithError(w, http.StatusNotFound, "Track not found", err.Error())
		return
	}

	var room models.Room
	err = up.db.DB.Where("sid = ?", track.RoomSID).First(&room).Error
	if err != nil {
		up.logger.Errorf("[Transcription] Room not found for track %s: %v", trackSID, err)
		up.respondWithError(w, http.StatusNotFound, "Room not found", err.Error())
		return
	}

	if room.MeetingID == nil {
		up.respondWithError(w, http.StatusBadRequest, "Track not associated with a meeting", "")
		return
	}

	meeting, err := up.meetingRepo.GetMeetingByID(*room.MeetingID)
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "Meeting not found", err.Error())
		return
	}

	hasAccess := false
	if meeting.CreatedBy == claims.UserID {
		hasAccess = true
	} else {
		participants, err := up.meetingRepo.GetMeetingParticipants(*room.MeetingID)
		if err == nil {
			for _, p := range participants {
				if p.UserID == claims.UserID {
					hasAccess = true
					break
				}
			}
		}
	}

	if !hasAccess {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "You don't have permission to transcribe this track")
		return
	}

	if track.TranscriptionStatus == "processing" {
		up.respondWithError(w, http.StatusConflict, "Track is already being transcribed", "")
		return
	}

	if track.EgressID == "" {
		up.respondWithError(w, http.StatusBadRequest, "Track recording not available", "Track must be recorded before transcription")
		return
	}

	bucket := "recontext"
	audioURL := fmt.Sprintf("https://api.storage.recontext.online/%s/%s_%s/tracks/%s.m3u8",
		bucket, meeting.ID.String(), room.SID, track.SID)
	minioObjectPath := fmt.Sprintf("%s_%s/tracks/%s.m3u8", meeting.ID.String(), room.SID, track.SID)
	up.logger.Infof("[Transcription] Track %s audio URL: %s", trackSID, audioURL)

	track.TranscriptionStatus = "processing"
	if err := up.db.DB.Save(&track).Error; err != nil {
		up.logger.Errorf("[Transcription] Failed to update track status: %v", err)
		up.respondWithError(w, http.StatusInternalServerError, "Failed to update track status", err.Error())
		return
	}

	if up.rabbitMQPublisher != nil {
		err := up.rabbitMQPublisher.PublishTranscriptionTask(
			track.ID,
			claims.UserID,
			audioURL,
			"", // Auto-detect language
			"", // No auth token needed for MinIO (internal access)
		)
		if err != nil {
			up.logger.Errorf("[Transcription] Failed to publish transcription task: %v", err)
			// Revert track status
			track.TranscriptionStatus = "pending"
			up.db.DB.Save(&track)
			up.respondWithError(w, http.StatusInternalServerError, "Failed to queue transcription task", err.Error())
			return
		}
		up.logger.Infof("[Transcription] Task queued for track %s", trackSID)
	} else {
		up.logger.Error("[Transcription] RabbitMQ publisher not available")
		up.respondWithError(w, http.StatusServiceUnavailable, "Transcription service unavailable", "")
		return
	}

	if up.redisPublisher != nil {
		if err := up.redisPublisher.PublishTranscriptionTask(
			track.ID,
			claims.UserID,
			bucket,
			minioObjectPath,
			"conference", // LiveKit meetings are multi-participant
			"ru",
		); err != nil {
			up.logger.Errorf("[Transcription] Failed to publish task to Redis: %v", err)
		} else {
			up.logger.Infof("[Transcription] Task also queued in Redis for Python worker: %s", trackSID)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"track_sid": trackSID,
		"track_id":  track.ID.String(),
		"status":    "processing",
		"message":   "Transcription task started",
	})
}

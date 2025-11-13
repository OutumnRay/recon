package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
	"github.com/google/uuid"
)

// RecordingInfo представляет информацию о записи
type RecordingInfo struct {
	ID            string   `json:"id"`
	Type          string   `json:"type"` // "room" or "track"
	Status        string   `json:"status"`
	StartedAt     string   `json:"started_at"`
	EndedAt       *string  `json:"ended_at,omitempty"`
	PlaylistURL   string   `json:"playlist_url"`
	ParticipantID *string  `json:"participant_id,omitempty"`
	TrackID       *string  `json:"track_id,omitempty"`
}

// GetMeetingRecordings godoc
// @Summary Get recordings for a meeting
// @Description Returns all recordings (room composite and tracks) for a meeting
// @Tags Recordings
// @Produce json
// @Param id path string true "Meeting ID"
// @Success 200 {array} RecordingInfo
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/meetings/{id}/recordings [get]
func (up *UserPortal) getMeetingRecordingsHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Extract meeting ID from path
	meetingID := strings.TrimPrefix(r.URL.Path, "/api/v1/meetings/")
	meetingID = strings.TrimSuffix(meetingID, "/recordings")

	// Get meeting
	meeting, err := up.meetingRepo.GetMeetingByID(uuid.Must(uuid.Parse(meetingID)))
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "Meeting not found", err.Error())
		return
	}

	// Check permissions - only participants can view recordings
	participants, err := up.meetingRepo.GetMeetingParticipants(uuid.Must(uuid.Parse(meetingID)))
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to check permissions", err.Error())
		return
	}

	isParticipant := meeting.CreatedBy == claims.UserID
	for _, p := range participants {
		if p.UserID == claims.UserID {
			isParticipant = true
			break
		}
	}

	if !isParticipant && claims.Role != models.RoleAdmin {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "")
		return
	}

	// Get all egress recordings for this meeting's rooms
	// Query LiveKit rooms by meeting ID (room name = meeting ID)
	rooms, err := up.liveKitRepo.GetRoomsByName(meetingID)
	if err != nil {
		up.logger.Infof("No rooms found for meeting %s: %v", meetingID, err)
	}

	recordings := []RecordingInfo{}

	// Collect room composite recordings
	for _, room := range rooms {
		if room.EgressID != "" {
			rec := RecordingInfo{
				ID:          room.EgressID,
				Type:        "room",
				Status:      "completed",
				StartedAt:   room.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
				PlaylistURL: "/api/v1/recordings/" + room.EgressID + "/playlist",
			}
			if room.FinishedAt != nil {
				endedAt := room.FinishedAt.Format("2006-01-02T15:04:05Z07:00")
				rec.EndedAt = &endedAt
			}
			recordings = append(recordings, rec)
		}

		// Get tracks for this room
		tracks, err := up.liveKitRepo.GetTracksByRoomSID(room.SID)
		if err == nil {
			for _, track := range tracks {
				if track.EgressID != "" && track.Source == "MICROPHONE" {
					rec := RecordingInfo{
						ID:        track.EgressID,
						Type:      "track",
						Status:    "completed",
						StartedAt: track.PublishedAt.Format("2006-01-02T15:04:05Z07:00"),
						PlaylistURL: "/api/v1/recordings/" + track.EgressID + "/playlist",
						TrackID:   &track.SID,
						ParticipantID: &track.ParticipantSID,
					}
					if track.UnpublishedAt != nil {
						endedAt := track.UnpublishedAt.Format("2006-01-02T15:04:05Z07:00")
						rec.EndedAt = &endedAt
					}
					recordings = append(recordings, rec)
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recordings)
}

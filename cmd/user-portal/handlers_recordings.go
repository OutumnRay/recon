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

// TrackRecordingInfo представляет информацию о записи трека участника
type TrackRecordingInfo struct {
	ID            string           `json:"id"`
	Status        string           `json:"status"`
	StartedAt     string           `json:"started_at"`
	EndedAt       *string          `json:"ended_at,omitempty"`
	PlaylistURL   string           `json:"playlist_url"`
	ParticipantID string           `json:"participant_id"`
	TrackID       string           `json:"track_id"`
	Participant   *models.UserInfo `json:"participant,omitempty"`
}

// RoomRecordingInfo представляет информацию о записи комнаты с треками
type RoomRecordingInfo struct {
	ID          string               `json:"id"`
	RoomSID     string               `json:"room_sid"`
	Status      string               `json:"status"`
	StartedAt   string               `json:"started_at"`
	EndedAt     *string              `json:"ended_at,omitempty"`
	PlaylistURL string               `json:"playlist_url,omitempty"` // Only if room has egress recording
	Tracks      []TrackRecordingInfo `json:"tracks"`
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
	up.logger.Infof("📹 [RECORDINGS] Request received: %s", r.URL.Path)

	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.logger.Errorf("📹 [RECORDINGS] Unauthorized access attempt")
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Extract meeting ID from path
	meetingID := strings.TrimPrefix(r.URL.Path, "/api/v1/meetings/")
	meetingID = strings.TrimSuffix(meetingID, "/recordings")
	up.logger.Infof("📹 [RECORDINGS] Meeting ID: %s, User ID: %s", meetingID, claims.UserID)

	// Get meeting
	meeting, err := up.meetingRepo.GetMeetingByID(uuid.Must(uuid.Parse(meetingID)))
	if err != nil {
		up.logger.Errorf("📹 [RECORDINGS] Meeting not found: %s, error: %v", meetingID, err)
		up.respondWithError(w, http.StatusNotFound, "Meeting not found", err.Error())
		return
	}
	up.logger.Infof("📹 [RECORDINGS] Meeting found: %s, CreatedBy: %s", meeting.ID, meeting.CreatedBy)

	// Check permissions - only participants can view recordings
	participants, err := up.meetingRepo.GetMeetingParticipants(uuid.Must(uuid.Parse(meetingID)))
	if err != nil {
		up.logger.Errorf("📹 [RECORDINGS] Failed to get participants: %v", err)
		up.respondWithError(w, http.StatusInternalServerError, "Failed to check permissions", err.Error())
		return
	}
	up.logger.Infof("📹 [RECORDINGS] Participants count: %d", len(participants))

	isParticipant := meeting.CreatedBy == claims.UserID
	for _, p := range participants {
		if p.UserID == claims.UserID {
			isParticipant = true
			break
		}
	}

	if !isParticipant && claims.Role != models.RoleAdmin {
		up.logger.Errorf("📹 [RECORDINGS] Access denied for user %s (not participant, role: %s)", claims.UserID, claims.Role)
		up.respondWithError(w, http.StatusForbidden, "Access denied", "")
		return
	}
	up.logger.Infof("📹 [RECORDINGS] Access granted (isParticipant: %v, role: %s)", isParticipant, claims.Role)

	// Get all egress recordings for this meeting's rooms
	// Query LiveKit rooms by meeting ID (room name = meeting ID)
	rooms, err := up.liveKitRepo.GetRoomsByName(meetingID)
	if err != nil {
		up.logger.Errorf("📹 [RECORDINGS] Error getting rooms for meeting %s: %v", meetingID, err)
	} else {
		up.logger.Infof("📹 [RECORDINGS] Found %d rooms for meeting %s", len(rooms), meetingID)
	}

	roomRecordings := []RoomRecordingInfo{}

	// Collect room recordings with their tracks
	for i, room := range rooms {
		up.logger.Infof("📹 [RECORDINGS] Room %d: SID=%s, Name=%s, EgressID=%s, Status=%s",
			i, room.SID, room.Name, room.EgressID, room.Status)

		roomRec := RoomRecordingInfo{
			ID:        room.SID,
			RoomSID:   room.SID,
			Status:    room.Status,
			StartedAt: room.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
			Tracks:    []TrackRecordingInfo{},
		}

		if room.FinishedAt != nil {
			endedAt := room.FinishedAt.Format("2006-01-02T15:04:05Z07:00")
			roomRec.EndedAt = &endedAt
		}

		// Add room composite recording URL if exists
		if room.EgressID != "" {
			roomRec.PlaylistURL = "/api/v1/recordings/" + room.EgressID + "/playlist"
			up.logger.Infof("📹 [RECORDINGS] Room has composite recording: %s", room.EgressID)
		}

		// Get tracks for this room
		tracks, err := up.liveKitRepo.GetTracksByRoomSID(room.SID)
		if err == nil {
			up.logger.Infof("📹 [RECORDINGS] Found %d tracks for room %s", len(tracks), room.SID)
			for j, track := range tracks {
				up.logger.Infof("📹 [RECORDINGS]   Track %d: SID=%s, Source=%s, EgressID=%s, Type=%s",
					j, track.SID, track.Source, track.EgressID, track.Type)

				if track.EgressID != "" && track.Source == "MICROPHONE" {
					trackRec := TrackRecordingInfo{
						ID:            track.SID,
						Status:        "completed",
						StartedAt:     track.PublishedAt.Format("2006-01-02T15:04:05Z07:00"),
						PlaylistURL:   "/api/v1/recordings/track/" + track.SID + "/playlist",
						TrackID:       track.SID,
						ParticipantID: track.ParticipantSID,
					}
					if track.UnpublishedAt != nil {
						endedAt := track.UnpublishedAt.Format("2006-01-02T15:04:05Z07:00")
						trackRec.EndedAt = &endedAt
					}

					// Get participant info to retrieve user details
					participant, err := up.liveKitRepo.GetParticipantBySID(track.ParticipantSID)
					if err == nil && participant != nil {
						// Parse user ID from participant identity
						userID, err := uuid.Parse(participant.Identity)
						if err == nil {
							// Get user info
							user, err := up.userRepo.GetByID(userID)
							if err == nil && user != nil {
								// Create UserInfo from User
								userInfo := &models.UserInfo{
									ID:           user.ID,
									Username:     user.Username,
									Email:        user.Email,
									Role:         user.Role,
									FirstName:    user.FirstName,
									LastName:     user.LastName,
									Phone:        user.Phone,
									Bio:          user.Bio,
									Avatar:       user.Avatar,
									DepartmentID: user.DepartmentID,
									Permissions:  user.Permissions,
									Language:     user.Language,
								}
								trackRec.Participant = userInfo
								up.logger.Infof("📹 [RECORDINGS] Added participant info: %s %s", user.FirstName, user.LastName)
							}
						}
					}

					roomRec.Tracks = append(roomRec.Tracks, trackRec)
					up.logger.Infof("📹 [RECORDINGS] Added track recording: %s", track.EgressID)
				}
			}
		} else {
			up.logger.Errorf("📹 [RECORDINGS] Error getting tracks for room %s: %v", room.SID, err)
		}

		roomRecordings = append(roomRecordings, roomRec)
	}

	up.logger.Infof("📹 [RECORDINGS] Returning %d rooms with recordings for meeting %s", len(roomRecordings), meetingID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roomRecordings)
}

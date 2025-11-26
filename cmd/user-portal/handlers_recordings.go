package main

import (
	"encoding/json"
	"fmt"
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

// TranscriptionPhrase представляет одну фразу в транскрипции
type TranscriptionPhrase struct {
	Start   float64 `json:"start"`             // Время начала в секундах
	End     float64 `json:"end"`               // Время окончания в секундах
	Text    string  `json:"text"`              // Текст фразы
	Speaker *string `json:"speaker,omitempty"` // Идентификатор говорящего (если есть диаризация)
}

// TrackRecordingInfo представляет информацию о записи трека участника
type TrackRecordingInfo struct {
	ID                   string                 `json:"id"`
	Status               string                 `json:"status"`
	StartedAt            string                 `json:"started_at"`
	EndedAt              *string                `json:"ended_at,omitempty"`
	PlaylistURL          string                 `json:"playlist_url"`
	ParticipantID        string                 `json:"participant_id"`
	TrackID              string                 `json:"track_id"`
	Participant          *models.UserInfo       `json:"participant,omitempty"`
	TranscriptionStatus  *string                `json:"transcription_status,omitempty"`
	TranscriptionPhrases []TranscriptionPhrase  `json:"transcription_phrases,omitempty"`
}

// RoomRecordingInfo представляет информацию о записи комнаты с треками
type RoomRecordingInfo struct {
	ID                string               `json:"id"`
	RoomSID           string               `json:"room_sid"`
	Status            string               `json:"status"`
	StartedAt         string               `json:"started_at"`
	EndedAt           *string              `json:"ended_at,omitempty"`
	PlaylistURL       string               `json:"playlist_url,omitempty"`        // Only if room has egress recording
	HasCompositeVideo bool                 `json:"has_composite_video"`           // Флаг наличия композитного видео (composite.m3u8 в корне)
	Tracks            []TrackRecordingInfo `json:"tracks"`
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

	// Pre-collect all participant SIDs for batch loading
	// Собираем участников для всех треков с записями (аудио и видео)
	// Collect participants for all tracks with recordings (audio and video)
	allParticipantSIDsSet := make(map[string]bool)
	for _, room := range rooms {
		tracks, err := up.liveKitRepo.GetTracksByRoomSID(room.SID)
		if err == nil {
			for _, track := range tracks {
				// Включаем всех участников, у которых есть записанные треки (audio or video)
				// Include all participants who have recorded tracks (audio or video)
				if track.EgressID != "" {
					allParticipantSIDsSet[track.ParticipantSID] = true
				}
			}
		}
	}

	// Batch load all participants once
	participantsMap := make(map[string]*models.UserInfo)
	for participantSID := range allParticipantSIDsSet {
		participant, err := up.liveKitRepo.GetParticipantBySID(participantSID)
		if err == nil && participant != nil {
			// Try to get user from users table first
			userID, err := uuid.Parse(participant.Identity)
			if err == nil {
				user, err := up.userRepo.GetByID(userID)
				if err == nil && user != nil {
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
					participantsMap[participantSID] = userInfo
				} else {
					// User not found in users table, this might be an anonymous user
					// Use the participant's name from LiveKit (set from display name)
					if participant.Name != "" {
						// Create a minimal UserInfo with the display name
						displayName := participant.Name
						userInfo := &models.UserInfo{
							ID:        userID,
							Username:  displayName,
							FirstName: displayName,
							Role:      "guest", // Mark as guest/anonymous
						}
						participantsMap[participantSID] = userInfo
					}
				}
			}
		}
	}
	if len(participantsMap) > 0 {
		up.logger.Infof("📹 [RECORDINGS] Batch loaded %d unique participants", len(participantsMap))
	}

	// Collect room recordings with their tracks
	for i, room := range rooms {
		up.logger.Infof("📹 [RECORDINGS] Room %d: SID=%s, Name=%s, EgressID=%s, Status=%s",
			i, room.SID, room.Name, room.EgressID, room.Status)

		roomRec := RoomRecordingInfo{
			ID:                room.SID,
			RoomSID:           room.SID,
			Status:            room.Status,
			StartedAt:         room.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
			HasCompositeVideo: room.HasCompositeVideo,
			Tracks:            []TrackRecordingInfo{},
		}

		if room.FinishedAt != nil {
			endedAt := room.FinishedAt.Format("2006-01-02T15:04:05Z07:00")
			roomRec.EndedAt = &endedAt
		}

		// Add composite video URL if exists (created by VideoPostProcessor)
		// VideoPostProcessor assembles composite video from individual tracks after transcription
		// Если есть композитное видео, оно лежит в корне: {meetingID}_{roomSID}/composite.m3u8
		if room.HasCompositeVideo {
			// Return API proxy URL instead of MinIO path (for authenticated access)
			// API proxy will serve: {meetingID}_{roomSID}/composite.m3u8 from MinIO
			compositePlaylistURL := fmt.Sprintf("/api/v1/recordings/%s/playlist", room.SID)
			roomRec.PlaylistURL = compositePlaylistURL
			up.logger.Infof("📹 [RECORDINGS] Room has composite video: %s", compositePlaylistURL)
		}

		// Get tracks for this room
		tracks, err := up.liveKitRepo.GetTracksByRoomSID(room.SID)
		if err == nil {
			up.logger.Infof("📹 [RECORDINGS] Found %d tracks for room %s", len(tracks), room.SID)

			// Process tracks (transcriptions will be fetched separately via /api/v1/rooms/{roomSid}/transcripts)
			for j, track := range tracks {
				up.logger.Infof("📹 [RECORDINGS]   Track %d: SID=%s, Source=%s, EgressID=%s, Type=%s",
					j, track.SID, track.Source, track.EgressID, track.Type)

				// Включаем треки с записью (egress_id не пустой)
				// Это могут быть: аудио (MICROPHONE), видео (CAMERA), шаринг экрана (SCREEN_SHARE, SCREEN_SHARE_AUDIO)
				// Include tracks that have recordings (egress_id is not empty)
				// This can be: audio (MICROPHONE), video (CAMERA), screen sharing (SCREEN_SHARE, SCREEN_SHARE_AUDIO)
				if track.EgressID != "" {
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

					// Add transcription status if available
					if track.TranscriptionStatus != "" {
						trackRec.TranscriptionStatus = &track.TranscriptionStatus
					}

					// Note: Transcription phrases are fetched separately via /api/v1/rooms/{roomSid}/transcripts

					// Get participant info from pre-loaded map
					if userInfo, exists := participantsMap[track.ParticipantSID]; exists {
						trackRec.Participant = userInfo
					}

					roomRec.Tracks = append(roomRec.Tracks, trackRec)
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

// RoomTranscriptsResponse представляет ответ с транскрипциями для комнаты
type RoomTranscriptsResponse struct {
	RoomSID      string                        `json:"room_sid"`
	Tracks       []TrackTranscriptInfo         `json:"tracks"`
	Memo         string                        `json:"memo,omitempty"`
	MemoRu       string                        `json:"memoRu,omitempty"`
}

// TrackTranscriptInfo представляет транскрипцию трека
type TrackTranscriptInfo struct {
	TrackID              string                `json:"track_id"`
	ParticipantID        string                `json:"participant_id"`
	Participant          *models.UserInfo      `json:"participant,omitempty"`
	StartedAt            string                `json:"started_at"`
	TranscriptionPhrases []TranscriptionPhrase `json:"transcription_phrases,omitempty"`
}

// getRoomTranscriptsHandler получает транскрипции для конкретной комнаты
// @Summary Get room transcripts
// @Description Get transcriptions for all tracks in a specific room
// @Tags recordings
// @Accept json
// @Produce json
// @Param roomSid path string true "Room SID"
// @Success 200 {object} RoomTranscriptsResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/rooms/{roomSid}/transcripts [get]
func (up *UserPortal) getRoomTranscriptsHandler(w http.ResponseWriter, r *http.Request) {
	up.logger.Infof("📝 [TRANSCRIPTS] Request received: %s", r.URL.Path)

	// Extract room SID from path
	roomSID := strings.TrimPrefix(r.URL.Path, "/api/v1/rooms/")
	roomSID = strings.TrimSuffix(roomSID, "/transcripts")
	up.logger.Infof("📝 [TRANSCRIPTS] Room SID: %s", roomSID)

	// Get room to verify it exists
	room, err := up.liveKitRepo.GetRoomBySID(roomSID)
	if err != nil || room == nil {
		up.logger.Errorf("📝 [TRANSCRIPTS] Room not found: %s, error: %v", roomSID, err)
		http.Error(w, `{"error":"Room not found"}`, http.StatusNotFound)
		return
	}

	// Get tracks for this room
	tracks, err := up.liveKitRepo.GetTracksByRoomSID(roomSID)
	if err != nil {
		up.logger.Errorf("📝 [TRANSCRIPTS] Failed to get tracks for room %s: %v", roomSID, err)
		http.Error(w, `{"error":"Failed to get tracks"}`, http.StatusInternalServerError)
		return
	}

	// Collect track IDs for transcription loading
	// Транскрипции делаются только для аудио-треков (MICROPHONE, SCREEN_SHARE_AUDIO)
	// Transcriptions are only done for audio tracks (MICROPHONE, SCREEN_SHARE_AUDIO)
	var trackIDs []uuid.UUID
	trackMap := make(map[uuid.UUID]*models.Track)
	for i := range tracks {
		// Для транскрипций: egress_id не пустой, это аудио-трек, и транскрипция завершена
		// For transcriptions: egress_id not empty, it's an audio track, and transcription is completed
		isAudioTrack := tracks[i].Source == "MICROPHONE" || tracks[i].Source == "SCREEN_SHARE_AUDIO"
		if tracks[i].EgressID != "" && isAudioTrack && tracks[i].TranscriptionStatus == "completed" {
			trackIDs = append(trackIDs, tracks[i].ID)
			trackMap[tracks[i].ID] = tracks[i]
		}
	}

	// Load transcriptions for all tracks in this room
	transcriptionsMap := make(map[uuid.UUID][]models.TranscriptionPhrase)
	if len(trackIDs) > 0 {
		var roomPhrases []models.TranscriptionPhrase
		err := up.db.DB.Where("track_id IN ?", trackIDs).
			Order("track_id, phrase_index ASC").
			Find(&roomPhrases).Error

		if err == nil && len(roomPhrases) > 0 {
			for _, phrase := range roomPhrases {
				transcriptionsMap[phrase.TrackID] = append(transcriptionsMap[phrase.TrackID], phrase)
			}
			up.logger.Infof("📝 [TRANSCRIPTS] Loaded %d phrases for %d tracks in room %s",
				len(roomPhrases), len(trackIDs), roomSID)
		} else if err != nil {
			up.logger.Errorf("📝 [TRANSCRIPTS] Failed to load transcriptions: %v", err)
		}
	}

	// Pre-load participants
	// Загружаем участников для всех треков с записью (аудио и видео)
	// Load participants for all tracks with recordings (audio and video)
	participantSIDsSet := make(map[string]bool)
	for _, track := range tracks {
		if track.EgressID != "" {
			participantSIDsSet[track.ParticipantSID] = true
		}
	}

	participantsMap := make(map[string]*models.UserInfo)
	for participantSID := range participantSIDsSet {
		participant, err := up.liveKitRepo.GetParticipantBySID(participantSID)
		if err == nil && participant != nil {
			// Try to get user from users table first
			userID, err := uuid.Parse(participant.Identity)
			if err == nil {
				user, err := up.userRepo.GetByID(userID)
				if err == nil && user != nil {
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
					participantsMap[participantSID] = userInfo
				} else {
					// User not found in users table, this might be an anonymous user
					// Use the participant's name from LiveKit (set from display name)
					if participant.Name != "" {
						// Create a minimal UserInfo with the display name
						displayName := participant.Name
						userInfo := &models.UserInfo{
							ID:        userID,
							Username:  displayName,
							FirstName: displayName,
							Role:      "guest", // Mark as guest/anonymous
						}
						participantsMap[participantSID] = userInfo
					}
				}
			}
		}
	}

	// Build response
	response := RoomTranscriptsResponse{
		RoomSID: roomSID,
		Tracks:  []TrackTranscriptInfo{},
		Memo:    room.Memo,
		MemoRu:  room.MemoRu,
	}

	for trackID, phrases := range transcriptionsMap {
		if track, exists := trackMap[trackID]; exists {
			trackInfo := TrackTranscriptInfo{
				TrackID:       track.SID,
				ParticipantID: track.ParticipantSID,
				StartedAt:     track.PublishedAt.Format("2006-01-02T15:04:05Z07:00"),
			}

			// Add participant info
			if userInfo, exists := participantsMap[track.ParticipantSID]; exists {
				trackInfo.Participant = userInfo
			}

			// Add transcription phrases
			trackInfo.TranscriptionPhrases = make([]TranscriptionPhrase, len(phrases))
			for i, p := range phrases {
				trackInfo.TranscriptionPhrases[i] = TranscriptionPhrase{
					Start:   p.StartTime,
					End:     p.EndTime,
					Text:    p.Text,
					Speaker: nil,
				}
			}

			response.Tracks = append(response.Tracks, trackInfo)
		}
	}

	up.logger.Infof("📝 [TRANSCRIPTS] Returning %d tracks with transcriptions for room %s", len(response.Tracks), roomSID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

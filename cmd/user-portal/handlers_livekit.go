package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
	"github.com/google/uuid"
	lkauth "github.com/livekit/protocol/auth"
)

// GetMeetingTokenResponse описывает ответ с токеном LiveKit для подключения к встрече
type GetMeetingTokenResponse struct {
	// Токен доступа LiveKit
	Token           string `json:"token"`
	// URL сервера LiveKit
	URL             string `json:"url"`
	// Название комнаты LiveKit
	RoomName        string `json:"roomName"`
	// Имя участника, отображаемое в комнате
	ParticipantName string `json:"participantName"`
	// Идентификатор встречи
	MeetingID       string `json:"meetingId"`
	// Запланированное время начала встречи (строка RFC3339)
	ScheduledAt     string `json:"scheduledAt"`
	// Длительность встречи в минутах
	Duration        int    `json:"duration"`
	// Время принудительного завершения встречи (если задано)
	ForceEndAt      string `json:"forceEndAt,omitempty"`
}

// GetMeetingToken godoc
// @Summary Получить токен LiveKit для присоединения к встрече
// @Description Получить токен доступа LiveKit для присоединения к комнате встречи (можно присоединиться за 10 минут до начала)
// @Tags Meetings
// @Accept json
// @Produce json
// @Param meetingId path string true "Идентификатор встречи"
// @Success 200 {object} GetMeetingTokenResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/meetings/{meetingId}/token [get]
func (up *UserPortal) getMeetingTokenHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		fmt.Printf("ERROR: Unauthorized token request\n")
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Get meeting ID from URL path
	meetingID := r.URL.Path[len("/api/v1/meetings/"):]
	if idx := len(meetingID) - len("/token"); idx > 0 && meetingID[idx:] == "/token" {
		meetingID = meetingID[:idx]
	}

	if meetingID == "" {
		fmt.Printf("ERROR: Meeting ID is empty in token request\n")
		up.respondWithError(w, http.StatusBadRequest, "Meeting ID is required", "")
		return
	}

	fmt.Printf("INFO: User %s requesting token for meeting %s\n", claims.UserID, meetingID)

	// Get meeting with details
	meeting, err := up.meetingRepo.GetMeetingByID(func() uuid.UUID { id, _ := uuid.Parse(meetingID); return id }())
	if err != nil {
		fmt.Printf("ERROR: Meeting %s not found: %v\n", meetingID, err)
		up.respondWithError(w, http.StatusNotFound, "Meeting not found", err.Error())
		return
	}

	fmt.Printf("INFO: Meeting %s found, status: %s, title: %s\n", meetingID, meeting.Status, meeting.Title)

	// Check if meeting is cancelled
	if meeting.Status == models.MeetingStatusCancelled {
		fmt.Printf("ERROR: Meeting %s is cancelled\n", meetingID)
		up.respondWithError(w, http.StatusForbidden, "Meeting is cancelled", "")
		return
	}

	// Check if meeting is already completed
	if meeting.Status == models.MeetingStatusCompleted {
		fmt.Printf("ERROR: Meeting %s is already completed\n", meetingID)
		up.respondWithError(w, http.StatusForbidden, "Meeting has ended", "")
		return
	}

	// Check if user is a participant
	participants, err := up.meetingRepo.GetMeetingParticipants(uuid.Must(uuid.Parse(meetingID)))
	if err != nil {
		fmt.Printf("ERROR: Failed to get participants for meeting %s: %v\n", meetingID, err)
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get participants", err.Error())
		return
	}

	fmt.Printf("INFO: Meeting %s has %d participants\n", meetingID, len(participants))

	isParticipant := false
	participantRole := "participant"

	// Check if user is the meeting creator
	if meeting.CreatedBy == claims.UserID {
		isParticipant = true
		participantRole = "organizer"
		fmt.Printf("INFO: User %s is the creator of meeting %s\n", claims.UserID, meetingID)
	} else {
		// Check if user is in the participants list
		for _, p := range participants {
			if p.UserID == claims.UserID {
				isParticipant = true
				participantRole = p.Role
				break
			}
		}
	}

	if !isParticipant {
		fmt.Printf("ERROR: User %s is not invited to meeting %s\n", claims.UserID, meetingID)
		up.respondWithError(w, http.StatusForbidden, "You are not invited to this meeting", "")
		return
	}

	fmt.Printf("INFO: User %s is a %s in meeting %s\n", claims.UserID, participantRole, meetingID)

	// Get current time
	now := time.Now()
	fmt.Printf("DEBUG: Current time: %s\n", now.Format(time.RFC3339))

	// TEMPORARY: Allow joining at any time (removed time restrictions)
	// TODO: Uncomment these checks when time-based restrictions are needed again

	// // Check if meeting can be joined (10 minutes before start)
	// tenMinutesBeforeStart := meeting.ScheduledAt.Add(-10 * time.Minute)
	// if now.Before(tenMinutesBeforeStart) {
	// 	minutesUntil := int(meeting.ScheduledAt.Sub(now).Minutes())
	// 	up.respondWithError(w, http.StatusForbidden,
	// 		fmt.Sprintf("Meeting starts in %d minutes. You can join 10 minutes before start.", minutesUntil),
	// 		"")
	// 	return
	// }

	// Calculate when meeting ends
	meetingEnd := meeting.ScheduledAt.Add(time.Duration(meeting.Duration) * time.Minute)
	fmt.Printf("DEBUG: Meeting scheduled: %s, duration: %d min, ends: %s\n",
		meeting.ScheduledAt.Format(time.RFC3339), meeting.Duration, meetingEnd.Format(time.RFC3339))

	// // If force_end_at_duration is true, meeting must end exactly at scheduled time
	// // Otherwise, allow joining if within scheduled time or up to end time
	// if meeting.ForceEndAtDuration && now.After(meetingEnd) {
	// 	up.respondWithError(w, http.StatusForbidden, "Meeting has ended", "")
	// 	return
	// }

	// Get LiveKit configuration
	fmt.Printf("DEBUG: Getting LiveKit configuration...\n")
	apiKey := os.Getenv("LIVEKIT_API_KEY")
	apiSecret := os.Getenv("LIVEKIT_API_SECRET")
	livekitURL := os.Getenv("LIVEKIT_URL")

	fmt.Printf("DEBUG: LIVEKIT_API_KEY present: %v, LIVEKIT_API_SECRET present: %v, LIVEKIT_URL: %s\n",
		apiKey != "", apiSecret != "", livekitURL)

	if apiKey == "" || apiSecret == "" {
		fmt.Printf("ERROR: LiveKit credentials not configured!\n")
		up.respondWithError(w, http.StatusInternalServerError, "LiveKit not configured", "")
		return
	}

	if livekitURL == "" {
		livekitURL = "ws://localhost:7880"
		fmt.Printf("DEBUG: Using default LiveKit URL: %s\n", livekitURL)
	}

	// Get room ID (should be set during meeting creation)
	fmt.Printf("DEBUG: Checking LiveKit room ID...\n")
	var roomID string
	if meeting.LiveKitRoomID != nil {
		roomID = meeting.LiveKitRoomID.String()
		fmt.Printf("DEBUG: Found LiveKit room ID: %s\n", roomID)
	} else {
		fmt.Printf("ERROR: Meeting %s has no LiveKit room ID assigned\n", meetingID)
		up.respondWithError(w, http.StatusInternalServerError, "Meeting room not initialized", "")
		return
	}

	fmt.Printf("INFO: Using LiveKit room: %s\n", roomID)

	// Get user info for participant name
	fmt.Printf("DEBUG: Getting user info for user ID: %s\n", claims.UserID.String())
	user, err := up.userRepo.GetByID(claims.UserID)
	if err != nil {
		fmt.Printf("ERROR: Failed to get user info for %s: %v\n", claims.UserID.String(), err)
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get user info", err.Error())
		return
	}

	participantName := user.Username
	fmt.Printf("DEBUG: Participant name: %s\n", participantName)

	// Calculate token validity duration
	fmt.Printf("DEBUG: Calculating token validity...\n")
	var tokenValidity time.Duration

	if meeting.ForceEndAtDuration {
		// Token valid until meeting end + 10 minutes
		tokenValidity = meetingEnd.Add(10 * time.Minute).Sub(now)
		fmt.Printf("DEBUG: ForceEndAtDuration=true, token valid until: %s\n", now.Add(tokenValidity).Format(time.RFC3339))
	} else {
		// Token valid for meeting duration from now + 10 minutes
		remainingDuration := meetingEnd.Sub(now)
		if remainingDuration < 0 {
			remainingDuration = time.Duration(meeting.Duration) * time.Minute
		}
		tokenValidity = remainingDuration + 10*time.Minute
		fmt.Printf("DEBUG: ForceEndAtDuration=false, remaining: %v, token validity: %v\n", remainingDuration, tokenValidity)
	}

	// Ensure minimum validity of 10 minutes
	if tokenValidity < 10*time.Minute {
		tokenValidity = 10 * time.Minute
		fmt.Printf("DEBUG: Token validity adjusted to minimum: 10 minutes\n")
	}
	fmt.Printf("DEBUG: Final token validity: %v\n", tokenValidity)

	// Generate LiveKit token
	fmt.Printf("DEBUG: Generating LiveKit token...\n")
	canPublish := true
	canSubscribe := true

	fmt.Printf("DEBUG: Creating access token with API key (length: %d)\n", len(apiKey))
	at := lkauth.NewAccessToken(apiKey, apiSecret)
	grant := &lkauth.VideoGrant{
		RoomJoin:     true,
		Room:         roomID,
		CanPublish:   &canPublish,
		CanSubscribe: &canSubscribe,
	}
	fmt.Printf("DEBUG: Video grant created for room: %s\n", roomID)

	metadata := fmt.Sprintf(`{"meetingId":"%s","role":"%s"}`, meetingID, participantRole)
	fmt.Printf("DEBUG: Setting token metadata: %s\n", metadata)

	at.SetVideoGrant(grant).
		SetIdentity(claims.UserID.String()).
		SetName(participantName).
		SetValidFor(tokenValidity).
		SetMetadata(metadata)

	fmt.Printf("DEBUG: Calling ToJWT()...\n")

	token, err := at.ToJWT()
	if err != nil {
		fmt.Printf("ERROR: Failed to generate JWT token for user %s in meeting %s: %v\n", claims.UserID, meetingID, err)
		up.respondWithError(w, http.StatusInternalServerError, "Failed to generate token", err.Error())
		return
	}

	fmt.Printf("INFO: Successfully generated token for user %s (%s) in room %s\n", claims.UserID, participantName, roomID)

	// Prepare response
	response := GetMeetingTokenResponse{
		Token:           token,
		URL:             livekitURL,
		RoomName:        roomID,
		ParticipantName: participantName,
		MeetingID:       meetingID,
		ScheduledAt:     meeting.ScheduledAt.Format(time.RFC3339),
		Duration:        meeting.Duration,
	}

	if meeting.ForceEndAtDuration {
		response.ForceEndAt = meetingEnd.Format(time.RFC3339)
		fmt.Printf("DEBUG: Force end at: %s\n", response.ForceEndAt)
	}

	fmt.Printf("DEBUG: Preparing to send response...\n")
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("ERROR: Failed to encode JSON response: %v\n", err)
		up.respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err.Error())
		return
	}
	fmt.Printf("INFO: Successfully sent token response to user %s\n", claims.UserID)
}

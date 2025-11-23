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

	// Get meeting with details
	meeting, err := up.meetingRepo.GetMeetingByID(func() uuid.UUID { id, _ := uuid.Parse(meetingID); return id }())
	if err != nil {
		fmt.Printf("ERROR: Meeting %s not found: %v\n", meetingID, err)
		up.respondWithError(w, http.StatusNotFound, "Meeting not found", err.Error())
		return
	}

	fmt.Printf("INFO: Meeting %s found, status: %s, title: %s, allow_anonymous: %v\n", meetingID, meeting.Status, meeting.Title, meeting.AllowAnonymous)

	// Check if meeting is cancelled
	if meeting.Status == models.MeetingStatusCancelled {
		fmt.Printf("ERROR: Meeting %s is cancelled\n", meetingID)
		up.respondWithError(w, http.StatusForbidden, "Meeting is cancelled", "")
		return
	}

	// Check if meeting is already completed (unless it's a permanent meeting)
	isPermanent := meeting.IsPermanent || meeting.Recurrence == models.MeetingRecurrencePermanent
	if meeting.Status == models.MeetingStatusCompleted && !isPermanent {
		fmt.Printf("ERROR: Meeting %s is already completed (not permanent)\n", meetingID)
		up.respondWithError(w, http.StatusForbidden, "Meeting has ended", "")
		return
	}

	if isPermanent {
		fmt.Printf("INFO: Meeting %s is permanent, allowing access regardless of status\n", meetingID)
	}

	// Check authentication - required unless meeting allows anonymous
	claims, ok := auth.GetUserFromContext(r.Context())
	var userID uuid.UUID
	participantRole := "participant"
	isAnonymous := false

	if !ok {
		// No authentication - check if anonymous is allowed
		if !meeting.AllowAnonymous {
			fmt.Printf("ERROR: Unauthorized token request for non-anonymous meeting %s\n", meetingID)
			up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
			return
		}
		// Anonymous user
		isAnonymous = true
		fmt.Printf("INFO: Anonymous user requesting token for meeting %s\n", meetingID)
	} else {
		// Authenticated user
		userID = claims.UserID
		fmt.Printf("INFO: User %s requesting token for meeting %s\n", userID, meetingID)
	}

	// For authenticated users, check if they are a participant
	if !isAnonymous {
		participants, err := up.meetingRepo.GetMeetingParticipants(uuid.Must(uuid.Parse(meetingID)))
		if err != nil {
			fmt.Printf("ERROR: Failed to get participants for meeting %s: %v\n", meetingID, err)
			up.respondWithError(w, http.StatusInternalServerError, "Failed to get participants", err.Error())
			return
		}

		fmt.Printf("INFO: Meeting %s has %d participants\n", meetingID, len(participants))

		isParticipant := false

		// Check if user is the meeting creator
		if meeting.CreatedBy == userID {
			isParticipant = true
			participantRole = "organizer"
			fmt.Printf("INFO: User %s is the creator of meeting %s\n", userID, meetingID)
		} else {
			// Check if user is in the participants list
			for _, p := range participants {
				if p.UserID == userID {
					isParticipant = true
					participantRole = p.Role
					break
				}
			}
		}

		if !isParticipant && !meeting.AllowAnonymous {
			fmt.Printf("ERROR: User %s is not invited to meeting %s\n", userID, meetingID)
			up.respondWithError(w, http.StatusForbidden, "You are not invited to this meeting", "")
			return
		}

		if isParticipant {
			fmt.Printf("INFO: User %s is a %s in meeting %s\n", userID, participantRole, meetingID)
		} else {
			fmt.Printf("INFO: User %s joining anonymous meeting %s\n", userID, meetingID)
		}
	}

	// Получаем текущее время для проверки доступа к встрече
	now := time.Now()
	fmt.Printf("DEBUG: Current time: %s\n", now.Format(time.RFC3339))

	// Вычисляем время окончания встречи для проверки доступа
	meetingEnd := meeting.ScheduledAt.Add(time.Duration(meeting.Duration) * time.Minute)
	fmt.Printf("DEBUG: Meeting scheduled: %s, duration: %d min, ends: %s\n",
		meeting.ScheduledAt.Format(time.RFC3339), meeting.Duration, meetingEnd.Format(time.RFC3339))

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

	// Use meeting ID as room name for LiveKit
	// This allows managing-portal to find the meeting when processing webhooks
	fmt.Printf("DEBUG: Using meeting ID as LiveKit room name...\n")
	roomName := meetingID
	fmt.Printf("INFO: Using LiveKit room name: %s (meeting ID)\n", roomName)

	// Get participant name
	var participantName string
	if isAnonymous {
		// For anonymous users, use "Guest" as default name
		// The actual name will be set by the AnonymousJoin page through location state
		participantName = "Guest"
		fmt.Printf("DEBUG: Anonymous participant, using default name: %s\n", participantName)
	} else {
		// Get user info for participant name
		fmt.Printf("DEBUG: Getting user info for user ID: %s\n", userID.String())
		user, err := up.userRepo.GetByID(userID)
		if err != nil {
			fmt.Printf("ERROR: Failed to get user info for %s: %v\n", userID.String(), err)
			up.respondWithError(w, http.StatusInternalServerError, "Failed to get user info", err.Error())
			return
		}

		participantName = user.Username
		fmt.Printf("DEBUG: Participant name: %s\n", participantName)
	}

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
		Room:         roomName,
		CanPublish:   &canPublish,
		CanSubscribe: &canSubscribe,
	}
	fmt.Printf("DEBUG: Video grant created for room: %s\n", roomName)

	metadata := fmt.Sprintf(`{"meetingId":"%s","role":"%s"}`, meetingID, participantRole)
	fmt.Printf("DEBUG: Setting token metadata: %s\n", metadata)

	// Use userID for identity, or generate a random one for anonymous users
	var identity string
	if isAnonymous {
		identity = uuid.New().String()
		fmt.Printf("DEBUG: Generated identity for anonymous user: %s\n", identity)
	} else {
		identity = userID.String()
	}

	at.SetVideoGrant(grant).
		SetIdentity(identity).
		SetName(participantName).
		SetValidFor(tokenValidity).
		SetMetadata(metadata)

	fmt.Printf("DEBUG: Calling ToJWT()...\n")

	token, err := at.ToJWT()
	if err != nil {
		fmt.Printf("ERROR: Failed to generate JWT token for identity %s in meeting %s: %v\n", identity, meetingID, err)
		up.respondWithError(w, http.StatusInternalServerError, "Failed to generate token", err.Error())
		return
	}

	fmt.Printf("INFO: Successfully generated token for identity %s (%s) in room %s\n", identity, participantName, roomName)

	// Prepare response
	response := GetMeetingTokenResponse{
		Token:           token,
		URL:             livekitURL,
		RoomName:        roomName,
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

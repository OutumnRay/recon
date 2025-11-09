package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
	lkauth "github.com/livekit/protocol/auth"
)

// GetMeetingTokenResponse represents response with LiveKit token
type GetMeetingTokenResponse struct {
	Token            string `json:"token"`
	URL              string `json:"url"`
	RoomName         string `json:"roomName"`
	ParticipantName  string `json:"participantName"`
	MeetingID        string `json:"meetingId"`
	ScheduledAt      string `json:"scheduledAt"`
	Duration         int    `json:"duration"`
	ForceEndAt       string `json:"forceEndAt,omitempty"`
}

// GetMeetingToken godoc
// @Summary Get LiveKit token for joining a meeting
// @Description Get LiveKit access token for joining a meeting room (can join 10 minutes before start)
// @Tags Meetings
// @Accept json
// @Produce json
// @Param meetingId path string true "Meeting ID"
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
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Get meeting ID from URL path
	meetingID := r.URL.Path[len("/api/v1/meetings/"):]
	if idx := len(meetingID) - len("/token"); idx > 0 && meetingID[idx:] == "/token" {
		meetingID = meetingID[:idx]
	}

	if meetingID == "" {
		up.respondWithError(w, http.StatusBadRequest, "Meeting ID is required", "")
		return
	}

	// Get meeting with details
	meeting, err := up.meetingRepo.GetMeetingByID(meetingID)
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "Meeting not found", err.Error())
		return
	}

	// Check if meeting is cancelled
	if meeting.Status == models.MeetingStatusCancelled {
		up.respondWithError(w, http.StatusForbidden, "Meeting is cancelled", "")
		return
	}

	// Check if meeting is already completed
	if meeting.Status == models.MeetingStatusCompleted {
		up.respondWithError(w, http.StatusForbidden, "Meeting has ended", "")
		return
	}

	// Check if user is a participant
	participants, err := up.meetingRepo.GetMeetingParticipants(meetingID)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get participants", err.Error())
		return
	}

	isParticipant := false
	participantRole := "participant"
	for _, p := range participants {
		if p.UserID == claims.UserID {
			isParticipant = true
			participantRole = p.Role
			break
		}
	}

	if !isParticipant {
		up.respondWithError(w, http.StatusForbidden, "You are not invited to this meeting", "")
		return
	}

	// Get current time
	now := time.Now()

	// Check if meeting can be joined (10 minutes before start)
	tenMinutesBeforeStart := meeting.ScheduledAt.Add(-10 * time.Minute)
	if now.Before(tenMinutesBeforeStart) {
		minutesUntil := int(meeting.ScheduledAt.Sub(now).Minutes())
		up.respondWithError(w, http.StatusForbidden,
			fmt.Sprintf("Meeting starts in %d minutes. You can join 10 minutes before start.", minutesUntil),
			"")
		return
	}

	// Calculate when meeting ends
	meetingEnd := meeting.ScheduledAt.Add(time.Duration(meeting.Duration) * time.Minute)

	// If force_end_at_duration is true, meeting must end exactly at scheduled time
	// Otherwise, allow joining if within scheduled time or up to end time
	if meeting.ForceEndAtDuration && now.After(meetingEnd) {
		up.respondWithError(w, http.StatusForbidden, "Meeting has ended", "")
		return
	}

	// Get LiveKit configuration
	apiKey := os.Getenv("LIVEKIT_API_KEY")
	apiSecret := os.Getenv("LIVEKIT_API_SECRET")
	livekitURL := os.Getenv("LIVEKIT_URL")

	if apiKey == "" || apiSecret == "" {
		up.respondWithError(w, http.StatusInternalServerError, "LiveKit not configured", "")
		return
	}

	if livekitURL == "" {
		livekitURL = "ws://localhost:7880"
	}

	// Get room ID (use LiveKitRoomID from meeting)
	roomID := ""
	if meeting.LiveKitRoomID != nil {
		roomID = *meeting.LiveKitRoomID
	} else {
		up.respondWithError(w, http.StatusInternalServerError, "Meeting room not initialized", "")
		return
	}

	// Get user info for participant name
	user, err := up.userRepo.GetByID(claims.UserID)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get user info", err.Error())
		return
	}

	participantName := user.Username

	// Calculate token validity duration
	var tokenValidity time.Duration

	if meeting.ForceEndAtDuration {
		// Token valid until meeting end + 10 minutes
		tokenValidity = meetingEnd.Add(10 * time.Minute).Sub(now)
	} else {
		// Token valid for meeting duration from now + 10 minutes
		remainingDuration := meetingEnd.Sub(now)
		if remainingDuration < 0 {
			remainingDuration = time.Duration(meeting.Duration) * time.Minute
		}
		tokenValidity = remainingDuration + 10*time.Minute
	}

	// Ensure minimum validity of 10 minutes
	if tokenValidity < 10*time.Minute {
		tokenValidity = 10 * time.Minute
	}

	// Generate LiveKit token
	canPublish := true
	canSubscribe := true

	at := lkauth.NewAccessToken(apiKey, apiSecret)
	grant := &lkauth.VideoGrant{
		RoomJoin:     true,
		Room:         roomID,
		CanPublish:   &canPublish,
		CanSubscribe: &canSubscribe,
	}

	at.SetVideoGrant(grant).
		SetIdentity(claims.UserID).
		SetName(participantName).
		SetValidFor(tokenValidity).
		SetMetadata(fmt.Sprintf(`{"meetingId":"%s","role":"%s"}`, meetingID, participantRole))

	token, err := at.ToJWT()
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to generate token", err.Error())
		return
	}

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
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

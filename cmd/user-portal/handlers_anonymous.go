package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"Recontext.online/pkg/database"
	"github.com/google/uuid"
	lkauth "github.com/livekit/protocol/auth"
)

// AnonymousJoinRequest describes the request body for anonymous meeting join
type AnonymousJoinRequest struct {
	DisplayName string `json:"displayName"` // Display name entered by anonymous user
}

// AnonymousJoinResponse describes the response for anonymous meeting join
type AnonymousJoinResponse struct {
	// LiveKit access token
	Token           string `json:"token"`
	// LiveKit server URL
	URL             string `json:"url"`
	// LiveKit room name
	RoomName        string `json:"roomName"`
	// Participant display name
	ParticipantName string `json:"participantName"`
	// Meeting ID
	MeetingID       string `json:"meetingId"`
	// Meeting title
	MeetingTitle    string `json:"meetingTitle"`
	// Scheduled meeting time (RFC3339 string)
	ScheduledAt     string `json:"scheduledAt"`
	// Meeting duration in minutes
	Duration        int    `json:"duration"`
	// Force end time if configured
	ForceEndAt      string `json:"forceEndAt,omitempty"`
}

// generateSessionID generates a random session ID for temporary user
func generateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// anonymousJoinHandler godoc
// @Summary Anonymous meeting join
// @Description Join a meeting anonymously by providing a display name (requires meeting to have allow_anonymous enabled). If user is authenticated, redirects to meeting room directly.
// @Tags Meetings
// @Accept json
// @Produce json
// @Param meetingId path string true "Meeting ID"
// @Param request body AnonymousJoinRequest true "Anonymous user display name"
// @Success 200 {object} AnonymousJoinResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/meetings/{meetingId}/join-anonymous [post]
func (up *UserPortal) anonymousJoinHandler(w http.ResponseWriter, r *http.Request) {
	// Extract meeting ID from URL path
	// Path format: /api/v1/meetings/{meetingId}/join-anonymous
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		up.respondWithError(w, http.StatusBadRequest, "Invalid URL path", "")
		return
	}
	meetingIDStr := pathParts[3]

	meetingID, err := uuid.Parse(meetingIDStr)
	if err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid meeting ID", err.Error())
		return
	}

	// Check if user is authenticated
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		// User is authenticated, redirect to regular meeting join
		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := up.jwtManager.VerifyToken(token)
		if err == nil && claims != nil {
			userID := claims.UserID
			// User is valid, use regular join handler
			fmt.Printf("INFO: Authenticated user %s attempting to join meeting %s via anonymous link, redirecting to regular join\n", userID, meetingID)

			// Get meeting to verify access
			meeting, err := up.meetingRepo.GetMeetingByID(meetingID)
			if err != nil {
				up.respondWithError(w, http.StatusNotFound, "Meeting not found", err.Error())
				return
			}

			// Check if user has access (is participant or creator)
			hasAccess := false
			if meeting.CreatedBy == userID {
				hasAccess = true
			} else {
				participants, _ := up.meetingRepo.GetMeetingParticipants(meetingID)
				for _, p := range participants {
					if p.UserID == userID {
						hasAccess = true
						break
					}
				}
			}

			if !hasAccess {
				up.respondWithError(w, http.StatusForbidden, "You don't have access to this meeting", "")
				return
			}

			// User has access, return redirect flag
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"redirect": true,
				"meetingId": meetingID.String(),
			})
			return
		}
	}

	// Parse request body
	var req AnonymousJoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate display name
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.DisplayName == "" {
		up.respondWithError(w, http.StatusBadRequest, "Display name is required", "")
		return
	}

	if len(req.DisplayName) > 255 {
		up.respondWithError(w, http.StatusBadRequest, "Display name too long (max 255 characters)", "")
		return
	}

	fmt.Printf("INFO: Anonymous join request for meeting %s with name: %s\n", meetingID, req.DisplayName)

	// Get meeting
	meeting, err := up.meetingRepo.GetMeetingByID(meetingID)
	if err != nil {
		fmt.Printf("ERROR: Meeting %s not found: %v\n", meetingID, err)
		up.respondWithError(w, http.StatusNotFound, "Meeting not found", err.Error())
		return
	}

	// Check if anonymous access is allowed
	if !meeting.AllowAnonymous {
		fmt.Printf("ERROR: Meeting %s does not allow anonymous access\n", meetingID)
		up.respondWithError(w, http.StatusForbidden, "This meeting does not allow anonymous access", "")
		return
	}

	// Check if meeting is cancelled
	if meeting.Status == "cancelled" {
		fmt.Printf("ERROR: Meeting %s is cancelled\n", meetingID)
		up.respondWithError(w, http.StatusForbidden, "Meeting is cancelled", "")
		return
	}

	// Check if meeting is already completed (unless it's permanent)
	isPermanent := meeting.IsPermanent || meeting.Recurrence == "permanent"
	if meeting.Status == "completed" && !isPermanent {
		fmt.Printf("ERROR: Meeting %s is already completed (not permanent)\n", meetingID)
		up.respondWithError(w, http.StatusForbidden, "Meeting has ended", "")
		return
	}

	fmt.Printf("INFO: Meeting %s allows anonymous access, status: %s, permanent: %v\n",
		meetingID, meeting.Status, isPermanent)

	// Generate session ID
	sessionID, err := generateSessionID()
	if err != nil {
		fmt.Printf("ERROR: Failed to generate session ID: %v\n", err)
		up.respondWithError(w, http.StatusInternalServerError, "Failed to generate session ID", err.Error())
		return
	}

	// Create temporary user
	tempUser := &database.TemporaryUser{
		DisplayName:  req.DisplayName,
		MeetingID:    meetingID,
		SessionID:    sessionID,
		CreatedAt:    time.Now(),
		LastActiveAt: &[]time.Time{time.Now()}[0],
	}

	// Save temporary user to database
	if err := up.db.Create(tempUser).Error; err != nil {
		fmt.Printf("ERROR: Failed to create temporary user: %v\n", err)
		up.respondWithError(w, http.StatusInternalServerError, "Failed to create temporary user", err.Error())
		return
	}

	fmt.Printf("INFO: Created temporary user %s for meeting %s (session: %s)\n",
		tempUser.ID, meetingID, sessionID)

	// Get LiveKit configuration
	apiKey := os.Getenv("LIVEKIT_API_KEY")
	apiSecret := os.Getenv("LIVEKIT_API_SECRET")
	livekitURL := os.Getenv("LIVEKIT_URL")

	if apiKey == "" || apiSecret == "" {
		fmt.Printf("ERROR: LiveKit credentials not configured!\n")
		up.respondWithError(w, http.StatusInternalServerError, "LiveKit not configured", "")
		return
	}

	if livekitURL == "" {
		livekitURL = "ws://localhost:7880"
	}

	// Use meeting ID as room name
	roomName := meetingID.String()

	// Calculate meeting end time
	meetingEnd := meeting.ScheduledAt.Add(time.Duration(meeting.Duration) * time.Minute)
	now := time.Now()

	// Calculate token validity
	remainingDuration := meetingEnd.Sub(now)
	if remainingDuration < 0 {
		remainingDuration = time.Duration(meeting.Duration) * time.Minute
	}
	tokenValidity := remainingDuration + 10*time.Minute

	// Ensure minimum validity of 10 minutes
	if tokenValidity < 10*time.Minute {
		tokenValidity = 10 * time.Minute
	}

	fmt.Printf("DEBUG: Token validity for anonymous user: %v\n", tokenValidity)

	// Generate LiveKit token
	canPublish := true
	canSubscribe := true

	at := lkauth.NewAccessToken(apiKey, apiSecret)
	grant := &lkauth.VideoGrant{
		RoomJoin:     true,
		Room:         roomName,
		CanPublish:   &canPublish,
		CanSubscribe: &canSubscribe,
	}

	// Use temporary user ID as identity, include session ID in metadata
	metadata := fmt.Sprintf(`{"meetingId":"%s","role":"participant","temporary":true,"sessionId":"%s"}`,
		meetingID, sessionID)

	at.SetVideoGrant(grant).
		SetIdentity(tempUser.ID.String()).
		SetName(req.DisplayName).
		SetValidFor(tokenValidity).
		SetMetadata(metadata)

	token, err := at.ToJWT()
	if err != nil {
		fmt.Printf("ERROR: Failed to generate JWT token for anonymous user: %v\n", err)
		up.respondWithError(w, http.StatusInternalServerError, "Failed to generate token", err.Error())
		return
	}

	fmt.Printf("INFO: Successfully generated token for anonymous user %s (%s) in room %s\n",
		tempUser.ID, req.DisplayName, roomName)

	// Prepare response
	response := AnonymousJoinResponse{
		Token:           token,
		URL:             livekitURL,
		RoomName:        roomName,
		ParticipantName: req.DisplayName,
		MeetingID:       meetingID.String(),
		MeetingTitle:    meeting.Title,
		ScheduledAt:     meeting.ScheduledAt.Format(time.RFC3339),
		Duration:        meeting.Duration,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("ERROR: Failed to encode JSON response: %v\n", err)
		up.respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err.Error())
		return
	}

	fmt.Printf("INFO: Successfully sent anonymous join response to user %s\n", req.DisplayName)
}

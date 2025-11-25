package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"Recontext.online/pkg/auth"
	"Recontext.online/pkg/summary"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// @Summary Generate meeting summary
// @Description Manually trigger summary generation for a meeting
// @Tags meetings
// @Accept json
// @Produce json
// @Param id path string true "Meeting ID"
// @Success 202 {object} map[string]interface{} "Summary generation started successfully"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/meetings/{id}/generate-summary [post]
func (up *UserPortal) generateMeetingSummaryHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Extract meeting ID from URL path: /api/v1/meetings/{id}/generate-summary
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/meetings/")
	meetingIDStr := strings.TrimSuffix(path, "/generate-summary")
	meetingID, err := uuid.Parse(meetingIDStr)
	if err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid meeting ID", err.Error())
		return
	}

	// Get meeting and verify ownership
	meeting, err := up.meetingRepo.GetMeetingWithDetails(meetingID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			up.respondWithError(w, http.StatusNotFound, "Meeting not found", "")
			return
		}
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get meeting", err.Error())
		return
	}

	// Verify user has access to this meeting
	if meeting.CreatedBy != claims.UserID {
		// Check if user is a participant
		hasAccess := false
		for _, p := range meeting.Participants {
			if p.UserID == claims.UserID {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			up.respondWithError(w, http.StatusForbidden, "Access denied", "")
			return
		}
	}

	// Get room SID for this meeting
	var roomSID string
	err = up.db.DB.Raw(`
		SELECT sid
		FROM livekit_rooms
		WHERE meeting_id = ?
		LIMIT 1
	`, meetingID).Scan(&roomSID).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			up.respondWithError(w, http.StatusNotFound, "No room found for this meeting", "")
			return
		}
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get room", err.Error())
		return
	}

	// Start summary generation in background
	go func() {
		if err := up.generateSummaryForRoom(roomSID, meetingID, meeting.Title); err != nil {
			up.logger.Errorf("Failed to generate summary for meeting %s: %v", meetingID, err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Summary generation started",
	})
}

// generateSummaryForRoom generates summary for a specific room
func (up *UserPortal) generateSummaryForRoom(roomSID string, meetingID uuid.UUID, meetingTitle string) error {
	up.logger.Infof("📝 Generating summary for room: %s (meeting: %s)", roomSID, meetingID)

	ctx := context.Background()

	// Get transcriptions for all audio tracks
	type TranscriptionData struct {
		ParticipantName string
		Segments        string // JSON
	}

	var transcriptions []TranscriptionData
	err := up.db.DB.Raw(`
		SELECT
			COALESCE(p.name, 'Unknown') as participant_name,
			t.transcription_segments as segments
		FROM livekit_tracks t
		LEFT JOIN livekit_participants p ON t.participant_sid = p.sid
		WHERE t.room_sid = ?
			AND t.type = 'audio'
			AND t.transcription_segments IS NOT NULL
		ORDER BY t.published_at ASC
	`, roomSID).Scan(&transcriptions).Error

	if err != nil {
		return fmt.Errorf("failed to get transcriptions: %w", err)
	}

	if len(transcriptions) == 0 {
		return fmt.Errorf("no transcriptions found for room %s", roomSID)
	}

	up.logger.Infof("   Found %d transcription(s) to process", len(transcriptions))

	// Collect all transcript segments
	var allSegments []summary.TranscriptSegment

	for _, trans := range transcriptions {
		var segments []struct {
			Start float64 `json:"start"`
			End   float64 `json:"end"`
			Text  string  `json:"text"`
		}

		if err := json.Unmarshal([]byte(trans.Segments), &segments); err != nil {
			up.logger.Infof("Failed to parse segments for %s: %v", trans.ParticipantName, err)
			continue
		}

		for _, seg := range segments {
			allSegments = append(allSegments, summary.TranscriptSegment{
				ParticipantName: trans.ParticipantName,
				StartTime:       seg.Start,
				EndTime:         seg.End,
				Text:            seg.Text,
			})
		}
	}

	if len(allSegments) == 0 {
		return fmt.Errorf("no valid transcript segments found")
	}

	up.logger.Infof("   Total segments: %d", len(allSegments))

	// Generate summaries
	generator, err := summary.NewSummaryGenerator()
	if err != nil {
		return fmt.Errorf("failed to create summary generator: %w", err)
	}

	summaries, err := generator.GenerateSummaries(ctx, meetingTitle, allSegments)
	if err != nil {
		return fmt.Errorf("failed to generate summaries: %w", err)
	}

	// Save summaries to database
	summariesJSON, err := json.Marshal(summaries)
	if err != nil {
		return fmt.Errorf("failed to marshal summaries: %w", err)
	}

	err = up.db.DB.Exec(`
		UPDATE livekit_rooms
		SET summaries = ?
		WHERE sid = ?
	`, summariesJSON, roomSID).Error

	if err != nil {
		return fmt.Errorf("failed to save summaries: %w", err)
	}

	up.logger.Infof("✅ Summary generation completed for meeting %s", meetingID)
	return nil
}

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

	// Get transcriptions from transcription_phrases table
	type TranscriptionPhrase struct {
		ParticipantName string
		StartTime       float64
		EndTime         float64
		Text            string
	}

	var phrases []TranscriptionPhrase
	err := up.db.DB.Raw(`
		SELECT
			COALESCE(p.name, 'Unknown') as participant_name,
			tp.absolute_start_time as start_time,
			tp.absolute_end_time as end_time,
			tp.text
		FROM transcription_phrases tp
		INNER JOIN livekit_tracks t ON tp.track_id = t.id
		LEFT JOIN livekit_participants p ON t.participant_sid = p.sid
		WHERE t.room_sid = ?
			AND t.type = 'audio'
		ORDER BY tp.absolute_start_time ASC
	`, roomSID).Scan(&phrases).Error

	if err != nil {
		return fmt.Errorf("failed to get transcriptions: %w", err)
	}

	if len(phrases) == 0 {
		return fmt.Errorf("no transcriptions found for room %s", roomSID)
	}

	up.logger.Infof("   Found %d transcription phrase(s) to process", len(phrases))

	// Collect all transcript segments
	var allSegments []summary.TranscriptSegment

	for _, phrase := range phrases {
		allSegments = append(allSegments, summary.TranscriptSegment{
			ParticipantName: phrase.ParticipantName,
			StartTime:       phrase.StartTime,
			EndTime:         phrase.EndTime,
			Text:            phrase.Text,
		})
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

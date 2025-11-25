package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"Recontext.online/internal/models"
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

	// Get room SID for this meeting using GORM
	var room models.Room
	err = up.db.DB.Where("meeting_id = ?", meetingID).First(&room).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			up.respondWithError(w, http.StatusNotFound, "No room found for this meeting", "")
			return
		}
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get room", err.Error())
		return
	}

	roomSID := room.SID

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

	// Get all audio tracks for this room using GORM
	var tracks []models.Track
	if err := up.db.DB.Where("room_sid = ? AND type = ?", roomSID, "audio").
		Find(&tracks).Error; err != nil {
		return fmt.Errorf("failed to get audio tracks: %w", err)
	}

	if len(tracks) == 0 {
		return fmt.Errorf("no audio tracks found for room %s", roomSID)
	}

	// Get track IDs and participant SIDs
	trackIDs := make([]uuid.UUID, len(tracks))
	participantSIDs := make([]string, 0, len(tracks))
	participantSIDSet := make(map[string]bool)
	trackToParticipantSID := make(map[uuid.UUID]string)

	for i, track := range tracks {
		trackIDs[i] = track.ID
		trackToParticipantSID[track.ID] = track.ParticipantSID
		if !participantSIDSet[track.ParticipantSID] {
			participantSIDs = append(participantSIDs, track.ParticipantSID)
			participantSIDSet[track.ParticipantSID] = true
		}
	}

	// Get participants using GORM
	var participants []models.Participant
	if err := up.db.DB.Where("sid IN ?", participantSIDs).
		Find(&participants).Error; err != nil {
		return fmt.Errorf("failed to get participants: %w", err)
	}

	// Build participant name map
	participantNameMap := make(map[string]string)
	for _, p := range participants {
		if p.Name != "" {
			participantNameMap[p.SID] = p.Name
		} else {
			participantNameMap[p.SID] = "Unknown"
		}
	}

	// Build track to participant name map
	trackParticipantMap := make(map[uuid.UUID]string)
	for trackID, participantSID := range trackToParticipantSID {
		if name, ok := participantNameMap[participantSID]; ok {
			trackParticipantMap[trackID] = name
		} else {
			trackParticipantMap[trackID] = "Unknown"
		}
	}

	// Get transcription phrases for these tracks using GORM
	var phrases []models.TranscriptionPhrase
	if err := up.db.DB.Where("track_id IN ?", trackIDs).
		Order("absolute_start_time ASC").
		Find(&phrases).Error; err != nil {
		return fmt.Errorf("failed to get transcriptions: %w", err)
	}

	if len(phrases) == 0 {
		return fmt.Errorf("no transcriptions found for room %s", roomSID)
	}

	up.logger.Infof("   Found %d transcription phrase(s) to process", len(phrases))

	// Collect all transcript segments
	var allSegments []summary.TranscriptSegment

	for _, phrase := range phrases {
		participantName := trackParticipantMap[phrase.TrackID]
		allSegments = append(allSegments, summary.TranscriptSegment{
			ParticipantName: participantName,
			StartTime:       phrase.AbsoluteStartTime,
			EndTime:         phrase.AbsoluteEndTime,
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

	// Save summaries to database using GORM
	summariesJSON, err := json.Marshal(summaries)
	if err != nil {
		return fmt.Errorf("failed to marshal summaries: %w", err)
	}

	if err := up.db.DB.Model(&models.Room{}).
		Where("sid = ?", roomSID).
		Update("summaries", summariesJSON).Error; err != nil {
		return fmt.Errorf("failed to save summaries: %w", err)
	}

	up.logger.Infof("✅ Summary generation completed for meeting %s", meetingID)
	return nil
}

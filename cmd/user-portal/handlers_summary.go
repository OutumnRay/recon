package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
	"Recontext.online/pkg/notifications"
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

	// Get room_sid from query parameter (optional - if not provided, use first room)
	roomSIDParam := r.URL.Query().Get("room_sid")
	up.logger.Infof("📝 [SUMMARY] Generate summary request: meetingID=%s, room_sid param=%s", meetingID, roomSIDParam)

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

	// Get room - either by room_sid or first room for meeting
	var room models.Room
	if roomSIDParam != "" {
		// Use specific room_sid provided by client
		err = up.db.DB.Where("sid = ? AND meeting_id = ?", roomSIDParam, meetingID).First(&room).Error
		up.logger.Infof("📝 [SUMMARY] Looking for specific room: sid=%s, meetingID=%s", roomSIDParam, meetingID)
	} else {
		// Fallback to first room (for backwards compatibility)
		err = up.db.DB.Where("meeting_id = ?", meetingID).First(&room).Error
		up.logger.Infof("📝 [SUMMARY] No room_sid provided, using first room for meetingID=%s", meetingID)
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			up.respondWithError(w, http.StatusNotFound, "No room found for this meeting", "")
			return
		}
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get room", err.Error())
		return
	}

	roomSID := room.SID
	up.logger.Infof("📝 [SUMMARY] Using room: %s for summary generation", roomSID)

	// Set summary status to processing
	if err := up.db.DB.Model(&models.Room{}).
		Where("sid = ?", roomSID).
		Updates(map[string]interface{}{
			"summary_status": "processing",
			"summary_error":  "",
		}).Error; err != nil {
		up.logger.Errorf("Failed to update summary status: %v", err)
	}

	// Send real-time notification: summary started
	up.notificationService.Notify(notifications.NewSummaryStatusNotification(
		meetingID,
		"processing",
		notifications.EventSummaryStarted,
		"",
	))

	// Start summary generation in background
	go func() {
		if err := up.generateSummaryForRoom(roomSID, meetingID, meeting.Title); err != nil {
			up.logger.Errorf("Failed to generate summary for meeting %s: %v", meetingID, err)
			// Set status to failed with error message
			up.db.DB.Model(&models.Room{}).
				Where("sid = ?", roomSID).
				Updates(map[string]interface{}{
					"summary_status": "failed",
					"summary_error":  err.Error(),
				})

			// Send real-time notification: summary failed
			up.notificationService.Notify(notifications.NewSummaryStatusNotification(
				meetingID,
				"failed",
				notifications.EventSummaryFailed,
				err.Error(),
			))
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Summary generation started",
		"status":  "processing",
	})
}

// generateSummaryForRoom generates summary for a specific room
func (up *UserPortal) generateSummaryForRoom(roomSID string, meetingID uuid.UUID, meetingTitle string) error {
	up.logger.Infof("📝 Generating summary for room: %s (meeting: %s)", roomSID, meetingID)

	ctx := context.Background()

	// Get all transcription phrases for this room
	// First, get all tracks for this room (not just audio - transcription can exist for any track)
	var tracks []models.Track
	if err := up.db.DB.Where("room_sid = ?", roomSID).
		Find(&tracks).Error; err != nil {
		return fmt.Errorf("failed to get tracks: %w", err)
	}

	if len(tracks) == 0 {
		return fmt.Errorf("no tracks found for room %s", roomSID)
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

	// Get participants with user information using JOIN
	type ParticipantWithUser struct {
		ParticipantSID  string
		ParticipantName string
		Identity        string
		UserFirstName   *string
		UserLastName    *string
	}

	var participantsWithUsers []ParticipantWithUser
	err := up.db.DB.Table("livekit_participants").
		Select("livekit_participants.sid as participant_sid, livekit_participants.name as participant_name, livekit_participants.identity, users.first_name as user_first_name, users.last_name as user_last_name").
		Joins("LEFT JOIN users ON livekit_participants.identity = users.id::text").
		Where("livekit_participants.sid IN ?", participantSIDs).
		Scan(&participantsWithUsers).Error

	if err != nil {
		return fmt.Errorf("failed to get participants with users: %w", err)
	}

	// Build participant name map with real names (first_name + last_name)
	participantNameMap := make(map[string]string)
	for _, p := range participantsWithUsers {
		// Priority: real name (first_name + last_name) > participant name > "Unknown"
		if p.UserFirstName != nil && p.UserLastName != nil && *p.UserFirstName != "" && *p.UserLastName != "" {
			participantNameMap[p.ParticipantSID] = *p.UserFirstName + " " + *p.UserLastName
		} else if p.ParticipantName != "" {
			participantNameMap[p.ParticipantSID] = p.ParticipantName
		} else {
			participantNameMap[p.ParticipantSID] = "Unknown"
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

	// Extract full summary text for memo fields
	var memoEN, memoRU string
	if summaries.English != nil {
		memoEN = summaries.English.FullSummary
	}
	if summaries.Russian != nil {
		memoRU = summaries.Russian.FullSummary
	}

	up.logger.Infof("📝 Saving summaries to room %s:", roomSID)
	up.logger.Infof("   memo (EN): %d chars", len(memoEN))
	up.logger.Infof("   memo_ru (RU): %d chars", len(memoRU))
	up.logger.Infof("   summaries JSON: %d bytes", len(summariesJSON))

	// Save summaries and update status to completed
	result := up.db.DB.Model(&models.Room{}).
		Where("sid = ?", roomSID).
		Updates(map[string]interface{}{
			"summaries":      summariesJSON,
			"memo":           memoEN,
			"memo_ru":        memoRU,
			"summary_status": "completed",
			"summary_error":  "",
		})

	if result.Error != nil {
		up.logger.Errorf("❌ Failed to save summaries to room %s: %v", roomSID, result.Error)
		return fmt.Errorf("failed to save summaries: %w", result.Error)
	}

	up.logger.Infof("✅ Summaries saved to livekit_rooms - rows affected: %d", result.RowsAffected)

	// Verify the save was successful
	var room models.Room
	if err := up.db.DB.Where("sid = ?", roomSID).First(&room).Error; err != nil {
		up.logger.Infof("⚠️ Could not verify save: %v", err)
	} else {
		up.logger.Infof("📋 Verification - memo: %d chars, memo_ru: %d chars, status: %s",
			len(room.Memo), len(room.MemoRu), room.SummaryStatus)
	}

	// Send real-time notification: summary completed
	up.notificationService.Notify(notifications.NewSummaryStatusNotification(
		meetingID,
		"completed",
		notifications.EventSummaryCompleted,
		"",
	))

	up.logger.Infof("✅ Summary generation completed for meeting %s (room %s)", meetingID, roomSID)
	return nil
}

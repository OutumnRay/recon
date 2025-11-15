package main

import (
	"fmt"
	"time"

	"Recontext.online/internal/models"
	"github.com/google/uuid"
)

// TranscriptionScheduler manages automatic transcription of pending tracks
type TranscriptionScheduler struct {
	up       *UserPortal
	ticker   *time.Ticker
	stopChan chan bool
}

// NewTranscriptionScheduler creates a new transcription scheduler
func NewTranscriptionScheduler(up *UserPortal) *TranscriptionScheduler {
	return &TranscriptionScheduler{
		up:       up,
		stopChan: make(chan bool),
	}
}

// Start begins the transcription scheduler
func (ts *TranscriptionScheduler) Start() {
	ts.up.logger.Info("🎙️ [TRANSCRIPTION SCHEDULER] Starting...")

	// Process pending tracks immediately on startup
	ts.processPendingTracks()

	// Start periodic check every 10 minutes
	ts.ticker = time.NewTicker(10 * time.Minute)
	go func() {
		for {
			select {
			case <-ts.ticker.C:
				ts.processPendingTracks()
			case <-ts.stopChan:
				ts.ticker.Stop()
				ts.up.logger.Info("🎙️ [TRANSCRIPTION SCHEDULER] Stopped")
				return
			}
		}
	}()

	ts.up.logger.Info("🎙️ [TRANSCRIPTION SCHEDULER] Started (checking every 10 minutes)")
}

// Stop stops the transcription scheduler
func (ts *TranscriptionScheduler) Stop() {
	ts.stopChan <- true
}

// processPendingTracks finds and queues all pending tracks for transcription
func (ts *TranscriptionScheduler) processPendingTracks() {
	ts.up.logger.Info("🎙️ [TRANSCRIPTION SCHEDULER] Checking for pending tracks...")

	// Find all tracks that:
	// 1. Have egress_id (recording is complete)
	// 2. Transcription status is 'pending', 'processing' (interrupted), or empty
	// 3. Are microphone tracks (audio only)
	var pendingTracks []models.Track
	err := ts.up.db.DB.Where(
		"egress_id != '' AND egress_id IS NOT NULL AND source = ? AND (transcription_status = ? OR transcription_status = ? OR transcription_status IS NULL OR transcription_status = '')",
		"MICROPHONE",
		"pending",
		"processing",
	).Find(&pendingTracks).Error

	if err != nil {
		ts.up.logger.Errorf("🎙️ [TRANSCRIPTION SCHEDULER] Error querying pending tracks: %v", err)
		return
	}

	if len(pendingTracks) == 0 {
		ts.up.logger.Info("🎙️ [TRANSCRIPTION SCHEDULER] No pending tracks found")
		return
	}

	ts.up.logger.Infof("🎙️ [TRANSCRIPTION SCHEDULER] Found %d pending tracks to transcribe", len(pendingTracks))

	successCount := 0
	failCount := 0

	for _, track := range pendingTracks {
		if ts.queueTrackForTranscription(track) {
			successCount++
		} else {
			failCount++
		}
	}

	ts.up.logger.Infof("🎙️ [TRANSCRIPTION SCHEDULER] Completed: %d queued, %d failed", successCount, failCount)
}

// queueTrackForTranscription queues a single track for transcription
func (ts *TranscriptionScheduler) queueTrackForTranscription(track models.Track) bool {
	// Get room to find meeting ID
	var room models.Room
	err := ts.up.db.DB.Where("sid = ?", track.RoomSID).First(&room).Error
	if err != nil {
		ts.up.logger.Errorf("🎙️ [TRANSCRIPTION SCHEDULER] Room not found for track %s: %v", track.SID, err)
		return false
	}

	// Check if track is associated with a meeting
	if room.MeetingID == nil {
		ts.up.logger.Errorf("🎙️ [TRANSCRIPTION SCHEDULER] Track %s not associated with a meeting", track.SID)
		return false
	}

	// Get meeting
	meeting, err := ts.up.meetingRepo.GetMeetingByID(*room.MeetingID)
	if err != nil {
		ts.up.logger.Errorf("🎙️ [TRANSCRIPTION SCHEDULER] Meeting not found for track %s: %v", track.SID, err)
		return false
	}

	// Construct audio URL for MinIO
	storageURL := "https://api.storage.recontext.online"
	bucket := "recontext"
	audioURL := fmt.Sprintf("%s/%s/%s_%s/tracks/%s.m3u8",
		storageURL, bucket, meeting.ID.String(), room.SID, track.SID)

	// Update track status to processing
	track.TranscriptionStatus = "processing"
	if err := ts.up.db.DB.Save(&track).Error; err != nil {
		ts.up.logger.Errorf("🎙️ [TRANSCRIPTION SCHEDULER] Failed to update track %s status: %v", track.SID, err)
		return false
	}

	// Send task to RabbitMQ
	if ts.up.rabbitMQPublisher != nil {
		// Use system user ID (00000000-0000-0000-0000-000000000000) for automatic transcriptions
		systemUserID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

		err := ts.up.rabbitMQPublisher.PublishTranscriptionTask(
			track.ID,
			systemUserID,
			audioURL,
			"", // Auto-detect language
			"", // No auth token needed for MinIO (internal access)
		)
		if err != nil {
			ts.up.logger.Errorf("🎙️ [TRANSCRIPTION SCHEDULER] Failed to publish task for track %s: %v", track.SID, err)
			// Revert track status
			track.TranscriptionStatus = "pending"
			ts.up.db.DB.Save(&track)
			return false
		}

		ts.up.logger.Infof("🎙️ [TRANSCRIPTION SCHEDULER] Queued track %s (Meeting: %s, Room: %s)",
			track.SID, meeting.ID, room.SID)
		return true
	}

	ts.up.logger.Error("🎙️ [TRANSCRIPTION SCHEDULER] RabbitMQ publisher not available")
	// Revert track status
	track.TranscriptionStatus = "pending"
	ts.up.db.DB.Save(&track)
	return false
}

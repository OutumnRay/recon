package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/database"
	"Recontext.online/pkg/fcm"
	"Recontext.online/pkg/llm"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// TranscriptionNotifier handles notifications when transcriptions are completed
type TranscriptionNotifier struct {
	up           *UserPortal
	fcmService   *fcm.FCMService
	llmClient    *llm.Client
	conn         *amqp.Connection
	channel      *amqp.Channel
	stopChan     chan bool
	rabbitmqURL  string
	reconnecting bool
}

// TranscriptionCompletedMessage represents the message format from transcription service
type TranscriptionCompletedMessage struct {
	TrackID string `json:"track_id"`
	Event   string `json:"event"`
}

// NewTranscriptionNotifier creates a new transcription notifier
func NewTranscriptionNotifier(up *UserPortal, fcmService *fcm.FCMService, llmClient *llm.Client) *TranscriptionNotifier {
	return &TranscriptionNotifier{
		up:         up,
		fcmService: fcmService,
		llmClient:  llmClient,
		stopChan:   make(chan bool),
	}
}

// Start begins listening for transcription completed events
func (tn *TranscriptionNotifier) Start(rabbitmqURL string) error {
	tn.rabbitmqURL = rabbitmqURL
	tn.up.logger.Info("📢 [TRANSCRIPTION NOTIFIER] Starting...")

	// Start connection in goroutine with automatic reconnection
	go tn.maintainConnection()

	return nil
}

// maintainConnection maintains the RabbitMQ connection with auto-reconnect
func (tn *TranscriptionNotifier) maintainConnection() {
	for {
		select {
		case <-tn.stopChan:
			tn.up.logger.Info("📢 [TRANSCRIPTION NOTIFIER] Stopped")
			return
		default:
			err := tn.connect()
			if err != nil {
				tn.up.logger.Errorf("📢 [TRANSCRIPTION NOTIFIER] Connection failed: %v. Retrying in 5 seconds...", err)
				time.Sleep(5 * time.Second)
				continue
			}

			// Connection established, now consume messages
			tn.consumeMessages()

			// If we get here, connection was lost
			tn.up.logger.Info("📢 [TRANSCRIPTION NOTIFIER] Connection lost. Reconnecting in 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}
}

// connect establishes connection to RabbitMQ
func (tn *TranscriptionNotifier) connect() error {
	// Clean up old connection if exists
	if tn.channel != nil {
		tn.channel.Close()
		tn.channel = nil
	}
	if tn.conn != nil {
		tn.conn.Close()
		tn.conn = nil
	}

	// Connect to RabbitMQ
	conn, err := amqp.Dial(tn.rabbitmqURL)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	tn.conn = conn

	// Create channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}
	tn.channel = ch

	// Declare queue
	queue, err := ch.QueueDeclare(
		"transcription_completed", // queue name
		true,                      // durable
		false,                     // delete when unused
		false,                     // exclusive
		false,                     // no-wait
		nil,                       // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Set QoS
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	tn.up.logger.Infof("📢 [TRANSCRIPTION NOTIFIER] Connected to RabbitMQ (queue: %s)", queue.Name)
	return nil
}

// consumeMessages consumes messages from the queue
func (tn *TranscriptionNotifier) consumeMessages() {
	// Start consuming messages
	msgs, err := tn.channel.Consume(
		"transcription_completed", // queue
		"",                        // consumer
		false,                     // auto-ack
		false,                     // exclusive
		false,                     // no-local
		false,                     // no-wait
		nil,                       // args
	)
	if err != nil {
		tn.up.logger.Errorf("📢 [TRANSCRIPTION NOTIFIER] Failed to start consuming: %v", err)
		return
	}

	tn.up.logger.Info("📢 [TRANSCRIPTION NOTIFIER] Listening for completion events...")

	// Listen for connection closure
	closeChan := make(chan *amqp.Error)
	tn.channel.NotifyClose(closeChan)

	emptyMessageCount := 0
	for {
		select {
		case err := <-closeChan:
			if err != nil {
				tn.up.logger.Infof("📢 [TRANSCRIPTION NOTIFIER] Channel closed: %v", err)
			}
			return
		case msg, ok := <-msgs:
			if !ok {
				tn.up.logger.Info("📢 [TRANSCRIPTION NOTIFIER] Message channel closed")
				return
			}

			// Check if message body is empty
			if len(msg.Body) == 0 {
				emptyMessageCount++
				// Only log the first empty message to avoid spam
				if emptyMessageCount == 1 {
					tn.up.logger.Debug("📢 [TRANSCRIPTION NOTIFIER] Received empty message (connection may be unstable)")
				}
				msg.Ack(false)

				// If we get multiple empty messages, connection is likely broken
				if emptyMessageCount > 10 {
					tn.up.logger.Info("📢 [TRANSCRIPTION NOTIFIER] Too many empty messages, reconnecting...")
					return
				}
				continue
			}

			// Reset empty message counter on valid message
			emptyMessageCount = 0
			tn.processMessage(msg)
		case <-tn.stopChan:
			return
		}
	}
}

// Stop stops the transcription notifier
func (tn *TranscriptionNotifier) Stop() {
	if tn.channel != nil {
		tn.channel.Close()
	}
	if tn.conn != nil {
		tn.conn.Close()
	}
	tn.stopChan <- true
}

// processMessage processes a transcription completed message
func (tn *TranscriptionNotifier) processMessage(msg amqp.Delivery) {
	// Parse message
	var completedMsg TranscriptionCompletedMessage
	err := json.Unmarshal(msg.Body, &completedMsg)
	if err != nil {
		tn.up.logger.Errorf("📢 [TRANSCRIPTION NOTIFIER] Failed to parse message: %v. Raw message body (first 200 chars): %s", err, string(msg.Body[:min(len(msg.Body), 200)]))
		msg.Ack(false) // Acknowledge to remove malformed message from queue instead of requeuing
		return
	}

	tn.up.logger.Infof("📢 [TRANSCRIPTION NOTIFIER] Received completion event for track: %s", completedMsg.TrackID)

	// Parse track ID
	trackID, err := uuid.Parse(completedMsg.TrackID)
	if err != nil {
		tn.up.logger.Errorf("📢 [TRANSCRIPTION NOTIFIER] Invalid track ID: %v", err)
		msg.Nack(false, false)
		return
	}

	// Send push notifications
	err = tn.sendPushNotifications(trackID)
	if err != nil {
		tn.up.logger.Errorf("📢 [TRANSCRIPTION NOTIFIER] Failed to send push notifications: %v", err)
		msg.Nack(false, true) // Requeue on error
		return
	}

	// Acknowledge message
	msg.Ack(false)
	tn.up.logger.Infof("📢 [TRANSCRIPTION NOTIFIER] Successfully processed track: %s", completedMsg.TrackID)
}

// sendPushNotifications sends push notifications to all meeting participants based on their preferences
func (tn *TranscriptionNotifier) sendPushNotifications(trackID uuid.UUID) error {
	// Get track from database
	var track models.Track
	err := tn.up.db.DB.Where("id = ?", trackID).First(&track).Error
	if err != nil {
		return fmt.Errorf("track not found: %w", err)
	}

	// Get room
	var room models.Room
	err = tn.up.db.DB.Where("sid = ?", track.RoomSID).First(&room).Error
	if err != nil {
		return fmt.Errorf("room not found: %w", err)
	}

	// Check if room is associated with a meeting
	if room.MeetingID == nil {
		tn.up.logger.Infof("📢 [TRANSCRIPTION NOTIFIER] Track %s not associated with a meeting, skipping notification", trackID)
		return nil
	}

	// Get meeting
	meeting, err := tn.up.meetingRepo.GetMeetingByID(*room.MeetingID)
	if err != nil {
		return fmt.Errorf("meeting not found: %w", err)
	}

	// Get meeting participants
	participants, err := tn.up.meetingRepo.GetMeetingParticipants(*room.MeetingID)
	if err != nil {
		return fmt.Errorf("failed to get participants: %w", err)
	}

	// Collect all user IDs (participants + creator)
	userIDs := make(map[uuid.UUID]bool)
	userIDs[meeting.CreatedBy] = true
	for _, p := range participants {
		userIDs[p.UserID] = true
	}

	tn.up.logger.Infof("📢 [TRANSCRIPTION NOTIFIER] Found %d users for meeting %s", len(userIDs), meeting.ID)

	// Check if all tracks in the room are transcribed
	var allTracks []models.Track
	err = tn.up.db.DB.Where("room_sid = ? AND source = ? AND egress_id != ? AND egress_id IS NOT NULL",
		track.RoomSID, "MICROPHONE", "").Find(&allTracks).Error
	if err != nil {
		tn.up.logger.Errorf("📢 [TRANSCRIPTION NOTIFIER] Failed to get room tracks: %v", err)
	}

	allRoomTracksCompleted := true
	for _, t := range allTracks {
		if t.TranscriptionStatus != "completed" {
			allRoomTracksCompleted = false
			break
		}
	}

	tn.up.logger.Infof("📢 [TRANSCRIPTION NOTIFIER] Room %s transcription status: %d/%d tracks completed, all=%v",
		track.RoomSID, countCompleted(allTracks), len(allTracks), allRoomTracksCompleted)

	// Collect FCM tokens based on user preferences
	trackTokens := []string{}   // Users who want track notifications
	roomTokens := []string{}    // Users who want room notifications (when all complete)

	for userID := range userIDs {
		// Get user to check notification preferences
		user, err := tn.up.userRepo.GetByID(userID)
		if err != nil {
			tn.up.logger.Errorf("📢 [TRANSCRIPTION NOTIFIER] Failed to get user %s: %v", userID, err)
			continue
		}

		// Get FCM tokens for this user
		tokens, err := tn.up.fcmDeviceRepo.GetUserFCMTokens(userID)
		if err != nil {
			tn.up.logger.Errorf("📢 [TRANSCRIPTION NOTIFIER] Failed to get FCM tokens for user %s: %v", userID, err)
			continue
		}

		// Sort tokens based on user preference
		pref := user.NotificationPreferences
		if pref == "" {
			pref = "both" // Default to both
		}

		switch pref {
		case "tracks":
			trackTokens = append(trackTokens, tokens...)
		case "rooms":
			roomTokens = append(roomTokens, tokens...)
		case "both":
			trackTokens = append(trackTokens, tokens...)
			roomTokens = append(roomTokens, tokens...)
		}
	}

	tn.up.logger.Infof("📢 [TRANSCRIPTION NOTIFIER] Token distribution: %d for tracks, %d for rooms",
		len(trackTokens), len(roomTokens))

	// Send track completion notification
	if len(trackTokens) > 0 && tn.fcmService != nil {
		ctx := context.Background()
		response, err := tn.fcmService.SendTranscriptionCompletedNotification(
			ctx,
			trackTokens,
			fmt.Sprintf("Track transcription completed: %s", meeting.Title),
			meeting.ID.String(),
		)

		if err != nil {
			tn.up.logger.Errorf("📢 [TRANSCRIPTION NOTIFIER] Failed to send track notification: %v", err)
		} else {
			tn.up.logger.Infof("📢 [TRANSCRIPTION NOTIFIER] Track notification sent: %d success, %d failure",
				response.SuccessCount, response.FailureCount)
		}
	}

	// Send room completion notification if all tracks are done
	if allRoomTracksCompleted && len(roomTokens) > 0 && tn.fcmService != nil {
		ctx := context.Background()
		response, err := tn.fcmService.SendTranscriptionCompletedNotification(
			ctx,
			roomTokens,
			fmt.Sprintf("Room transcription completed: %s", meeting.Title),
			meeting.ID.String(),
		)

		if err != nil {
			tn.up.logger.Errorf("📢 [TRANSCRIPTION NOTIFIER] Failed to send room notification: %v", err)
		} else {
			tn.up.logger.Infof("📢 [TRANSCRIPTION NOTIFIER] Room notification sent: %d success, %d failure",
				response.SuccessCount, response.FailureCount)
		}
	}

	if tn.fcmService == nil {
		tn.up.logger.Info("📢 [TRANSCRIPTION NOTIFIER] FCM service not available, skipping push notification")
	}

	// Generate memo if all tracks are completed and LLM is configured
	if allRoomTracksCompleted && tn.llmClient != nil && tn.llmClient.IsConfigured() {
		tn.up.logger.Infof("📝 [MEMO GENERATION] Starting memo generation for room %s", track.RoomSID)

		// Generate memo in background to avoid blocking
		go func() {
			if err := tn.generateAndSaveMemo(track.RoomSID, meeting.ID, meeting.Title, allTracks, userIDs); err != nil {
				tn.up.logger.Errorf("📝 [MEMO GENERATION] Failed: %v", err)
			}
		}()
	}

	return nil
}

// countCompleted counts how many tracks have completed transcription
func countCompleted(tracks []models.Track) int {
	count := 0
	for _, t := range tracks {
		if t.TranscriptionStatus == "completed" {
			count++
		}
	}
	return count
}

// generateAndSaveMemo generates a memo using LLM and saves it to the room
func (tn *TranscriptionNotifier) generateAndSaveMemo(roomSID string, meetingID uuid.UUID, meetingTitle string, tracks []models.Track, userIDs map[uuid.UUID]bool) error {
	// Load all transcription phrases for the room
	var trackIDs []uuid.UUID
	trackParticipants := make(map[uuid.UUID]string)

	for _, track := range tracks {
		trackIDs = append(trackIDs, track.ID)

		// Get participant name
		participantName := "Unknown"
		if track.ParticipantSID != "" {
			var participant database.LiveKitParticipant
			if err := tn.up.db.DB.Where("sid = ?", track.ParticipantSID).First(&participant).Error; err == nil {
				// First try to get from registered users
				if participant.UserID != nil {
					if user, err := tn.up.userRepo.GetByID(*participant.UserID); err == nil {
						if user.FirstName != "" {
							participantName = user.FirstName
							if user.LastName != "" {
								participantName += " " + user.LastName
							}
						} else {
							participantName = user.Username
						}
					}
				} else if participant.Name != "" {
					// Not a registered user, use the LiveKit participant name
					// (which contains the display name for anonymous users)
					participantName = participant.Name
				}
			}
		}
		trackParticipants[track.ID] = participantName
	}

	tn.up.logger.Infof("📝 [MEMO GENERATION] Loading transcripts for %d tracks", len(trackIDs))

	var allPhrases []models.TranscriptionPhrase
	err := tn.up.db.DB.Where("track_id IN ?", trackIDs).
		Order("track_id, phrase_index ASC").
		Find(&allPhrases).Error
	if err != nil {
		return fmt.Errorf("failed to load transcription phrases: %w", err)
	}

	tn.up.logger.Infof("📝 [MEMO GENERATION] Loaded %d phrases", len(allPhrases))

	// Build sequential dialogue with timestamps
	type TimedPhrase struct {
		Timestamp  float64
		Speaker    string
		Text       string
		TrackStart float64
	}

	var timedPhrases []TimedPhrase
	trackStartTimes := make(map[uuid.UUID]float64)

	for _, track := range tracks {
		trackStartTimes[track.ID] = float64(track.PublishedAt.Unix())
	}

	for _, phrase := range allPhrases {
		trackStart := trackStartTimes[phrase.TrackID]
		absoluteTimestamp := trackStart + phrase.StartTime

		timedPhrases = append(timedPhrases, TimedPhrase{
			Timestamp:  absoluteTimestamp,
			Speaker:    trackParticipants[phrase.TrackID],
			Text:       phrase.Text,
			TrackStart: trackStart,
		})
	}

	// Sort by absolute timestamp
	sort.Slice(timedPhrases, func(i, j int) bool {
		return timedPhrases[i].Timestamp < timedPhrases[j].Timestamp
	})

	// Build dialogue text
	var dialogueBuilder strings.Builder
	dialogueBuilder.WriteString(fmt.Sprintf("Meeting: %s\n\n", meetingTitle))

	for _, phrase := range timedPhrases {
		dialogueBuilder.WriteString(fmt.Sprintf("%s: %s\n", phrase.Speaker, phrase.Text))
	}

	dialogue := dialogueBuilder.String()
	tn.up.logger.Infof("📝 [MEMO GENERATION] Built dialogue with %d phrases, length: %d chars", len(timedPhrases), len(dialogue))

	// Generate memo using LLM (English)
	messages := []llm.Message{
		{
			Role:    "system",
			Content: "You are a professional meeting assistant. Summarize the whole dialogue naturally and concisely. Highlight the key points discussed and any action items or tasks if necessary. Write in a clear, readable format without forcing a specific structure. Write in English.",
		},
		{
			Role:    "user",
			Content: dialogue,
		},
	}

	tn.up.logger.Info("📝 [MEMO GENERATION] Calling LLM API for English memo...")
	memo, err := tn.llmClient.GenerateChatCompletion(messages)
	if err != nil {
		return fmt.Errorf("failed to generate memo: %w", err)
	}

	tn.up.logger.Infof("📝 [MEMO GENERATION] Generated English memo, length: %d chars", len(memo))

	// Generate Russian translation
	messagesRu := []llm.Message{
		{
			Role:    "system",
			Content: "Ты профессиональный помощник по ведению совещаний. Создай естественное и краткое резюме всего диалога. Выдели ключевые моменты обсуждения и любые задачи или действия, если они есть. Пиши в ясном, читаемом формате без навязывания конкретной структуры. Пиши на русском языке.",
		},
		{
			Role:    "user",
			Content: dialogue,
		},
	}

	tn.up.logger.Info("📝 [MEMO GENERATION] Calling LLM API for Russian memo...")
	memoRu, err := tn.llmClient.GenerateChatCompletion(messagesRu)
	if err != nil {
		tn.up.logger.Infof("⚠️ [MEMO GENERATION] Failed to generate Russian memo: %v", err)
		memoRu = "" // Fallback to empty if Russian generation fails
	} else {
		tn.up.logger.Infof("📝 [MEMO GENERATION] Generated Russian memo, length: %d chars", len(memoRu))
	}

	// Save both memos to room
	err = tn.up.db.DB.Model(&database.LiveKitRoom{}).
		Where("sid = ?", roomSID).
		Updates(map[string]interface{}{
			"memo":    memo,
			"memo_ru": memoRu,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to save memos: %w", err)
	}

	tn.up.logger.Infof("📝 [MEMO GENERATION] Saved memos to room %s", roomSID)

	// Send push notification about memo completion
	if tn.fcmService != nil {
		var tokens []string
		for userID := range userIDs {
			userTokens, err := tn.up.fcmDeviceRepo.GetUserFCMTokens(userID)
			if err == nil {
				tokens = append(tokens, userTokens...)
			}
		}

		if len(tokens) > 0 {
			ctx := context.Background()
			response, err := tn.fcmService.SendTranscriptionCompletedNotification(
				ctx,
				tokens,
				fmt.Sprintf("Meeting memo generated: %s", meetingTitle),
				meetingID.String(),
			)

			if err != nil {
				tn.up.logger.Errorf("📝 [MEMO GENERATION] Failed to send notification: %v", err)
			} else {
				tn.up.logger.Infof("📝 [MEMO GENERATION] Notification sent: %d success, %d failure",
					response.SuccessCount, response.FailureCount)
			}
		}
	}

	return nil
}

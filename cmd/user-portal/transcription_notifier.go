package main

import (
	"context"
	"encoding/json"
	"fmt"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/fcm"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// TranscriptionNotifier handles notifications when transcriptions are completed
type TranscriptionNotifier struct {
	up          *UserPortal
	fcmService  *fcm.FCMService
	conn        *amqp.Connection
	channel     *amqp.Channel
	stopChan    chan bool
}

// TranscriptionCompletedMessage represents the message format from transcription service
type TranscriptionCompletedMessage struct {
	TrackID string `json:"track_id"`
	Event   string `json:"event"`
}

// NewTranscriptionNotifier creates a new transcription notifier
func NewTranscriptionNotifier(up *UserPortal, fcmService *fcm.FCMService) *TranscriptionNotifier {
	return &TranscriptionNotifier{
		up:         up,
		fcmService: fcmService,
		stopChan:   make(chan bool),
	}
}

// Start begins listening for transcription completed events
func (tn *TranscriptionNotifier) Start(rabbitmqURL string) error {
	tn.up.logger.Info("📢 [TRANSCRIPTION NOTIFIER] Starting...")

	// Connect to RabbitMQ
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	tn.conn = conn

	// Create channel
	ch, err := conn.Channel()
	if err != nil {
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
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Set QoS
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Start consuming messages
	msgs, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	tn.up.logger.Info("📢 [TRANSCRIPTION NOTIFIER] Started (listening for completion events)")

	// Process messages in goroutine
	go func() {
		for {
			select {
			case msg := <-msgs:
				tn.processMessage(msg)
			case <-tn.stopChan:
				tn.up.logger.Info("📢 [TRANSCRIPTION NOTIFIER] Stopped")
				return
			}
		}
	}()

	return nil
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
		tn.up.logger.Errorf("📢 [TRANSCRIPTION NOTIFIER] Failed to parse message: %v", err)
		msg.Nack(false, false)
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

// sendPushNotifications sends push notifications to all meeting participants
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

	// Collect FCM tokens for all users
	allTokens := []string{}
	for userID := range userIDs {
		tokens, err := tn.up.fcmDeviceRepo.GetUserFCMTokens(userID)
		if err != nil {
			tn.up.logger.Errorf("📢 [TRANSCRIPTION NOTIFIER] Failed to get FCM tokens for user %s: %v", userID, err)
			continue
		}

		allTokens = append(allTokens, tokens...)
	}

	if len(allTokens) == 0 {
		tn.up.logger.Infof("📢 [TRANSCRIPTION NOTIFIER] No FCM tokens found for meeting %s participants", meeting.ID)
		return nil
	}

	tn.up.logger.Infof("📢 [TRANSCRIPTION NOTIFIER] Sending push notification to %d devices", len(allTokens))

	// Send push notification
	if tn.fcmService != nil {
		ctx := context.Background()
		response, err := tn.fcmService.SendTranscriptionCompletedNotification(
			ctx,
			allTokens,
			meeting.Title,
			meeting.ID.String(),
		)

		if err != nil {
			return fmt.Errorf("failed to send FCM notification: %w", err)
		}

		tn.up.logger.Infof("📢 [TRANSCRIPTION NOTIFIER] FCM notification sent: %d success, %d failure",
			response.SuccessCount, response.FailureCount)

		// Log failed tokens
		if response.FailureCount > 0 {
			for i, sendResponse := range response.Responses {
				if !sendResponse.Success {
					tn.up.logger.Errorf("📢 [TRANSCRIPTION NOTIFIER] Failed to send to token %d: %v",
						i, sendResponse.Error)
				}
			}
		}
	} else {
		tn.up.logger.Info("📢 [TRANSCRIPTION NOTIFIER] FCM service not available, skipping push notification")
	}

	return nil
}

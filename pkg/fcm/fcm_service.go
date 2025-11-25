package fcm

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// FCMService provides Firebase Cloud Messaging functionality
type FCMService struct {
	client *messaging.Client
}

// NewFCMService creates a new FCM service
func NewFCMService(credentialsPath string) (*FCMService, error) {
	ctx := context.Background()

	// Initialize Firebase app
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	// Get messaging client
	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get messaging client: %w", err)
	}

	return &FCMService{
		client: client,
	}, nil
}

// SendNotification sends a push notification to specific FCM tokens
func (s *FCMService) SendNotification(ctx context.Context, tokens []string, title, body string, data map[string]string) (*messaging.BatchResponse, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("no tokens provided")
	}

	// Create the message
	message := &messaging.MulticastMessage{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data:   data,
		Tokens: tokens,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				ChannelID: "transcription_notifications",
				Priority:  messaging.PriorityHigh,
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: title,
						Body:  body,
					},
					Sound: "default",
				},
			},
		},
	}

	// Send the message
	response, err := s.client.SendEachForMulticast(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send FCM message: %w", err)
	}

	return response, nil
}

// SendTranscriptionCompletedNotification sends a notification about completed transcription
func (s *FCMService) SendTranscriptionCompletedNotification(ctx context.Context, tokens []string, meetingTitle string, meetingID string) (*messaging.BatchResponse, error) {
	title := "Расшифровка готова"
	body := fmt.Sprintf("Расшифровка встречи \"%s\" завершена", meetingTitle)

	data := map[string]string{
		"type":       "transcription_completed",
		"meeting_id": meetingID,
	}

	return s.SendNotification(ctx, tokens, title, body, data)
}

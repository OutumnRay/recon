package fcm

import (
	"context"
	"encoding/base64"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// FCMService provides Firebase Cloud Messaging functionality
type FCMService struct {
	client *messaging.Client
}

// NewFCMService creates a new FCM service from a credentials file path
func NewFCMService(credentialsPath string) (*FCMService, error) {
	ctx := context.Background()

	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get messaging client: %w", err)
	}

	return &FCMService{client: client}, nil
}

// NewFCMServiceFromJSON creates a new FCM service from a base64-encoded JSON string.
// Use this when deploying to environments without persistent file storage (e.g. AppPlatform).
// Set the FCM_CREDENTIALS_JSON environment variable to the base64-encoded service account JSON.
func NewFCMServiceFromJSON(credentialsBase64 string) (*FCMService, error) {
	jsonBytes, err := base64.StdEncoding.DecodeString(credentialsBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode FCM credentials (expected base64): %w", err)
	}

	ctx := context.Background()
	opt := option.WithCredentialsJSON(jsonBytes)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get messaging client: %w", err)
	}

	return &FCMService{client: client}, nil
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

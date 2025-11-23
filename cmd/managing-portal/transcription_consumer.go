package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// TranscriptionResult represents the result message from transcription service
type TranscriptionResult struct {
	Event        string                 `json:"event"`
	TrackID      string                 `json:"track_id"`
	UserID       string                 `json:"user_id"`
	AudioURL     string                 `json:"audio_url"`
	JSONURL      string                 `json:"json_url"`
	Transcription TranscriptionData     `json:"transcription"`
	Timestamp    string                 `json:"timestamp"`
	Status       string                 `json:"status"`
}

// TranscriptionData contains the transcription details
type TranscriptionData struct {
	Phrases      []TranscriptionPhrase `json:"phrases"`
	PhraseCount  int                   `json:"phrase_count"`
	TotalDuration float64              `json:"total_duration"`
}

// TranscriptionPhrase represents a single transcribed phrase
type TranscriptionPhrase struct {
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Language   string  `json:"language"`
}

// StartTranscriptionConsumer starts consuming transcription results from RabbitMQ
func StartTranscriptionConsumer() {
	// Get RabbitMQ connection details from environment
	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://recontext:je9rO4k6CQ3M@5.129.227.21:5672/"
	}

	resultQueue := os.Getenv("RABBITMQ_RESULT_QUEUE")
	if resultQueue == "" {
		resultQueue = "transcription_results"
	}

	log.Printf("Starting transcription result consumer...")
	log.Printf("RabbitMQ URL: %s", rabbitmqURL)
	log.Printf("Result Queue: %s", resultQueue)

	// Connect to RabbitMQ with retry
	var conn *amqp.Connection
	var err error

	for retries := 0; retries < 5; retries++ {
		conn, err = amqp.Dial(rabbitmqURL)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ (attempt %d/5): %v", retries+1, err)
		time.Sleep(time.Duration(retries+1) * 2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ after 5 attempts: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	// Declare result queue
	q, err := ch.QueueDeclare(
		resultQueue, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// Set QoS to process one message at a time
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	// Start consuming messages
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack (manual ack to ensure processing)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	log.Printf("✅ Transcription consumer started, waiting for results...")

	// Process messages
	go func() {
		for msg := range msgs {
			processTranscriptionResult(msg.Body)
			msg.Ack(false) // Acknowledge message
		}
	}()

	// Block forever
	select {}
}

// processTranscriptionResult processes a transcription result message
func processTranscriptionResult(body []byte) {
	var result TranscriptionResult

	err := json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("Error parsing transcription result: %v", err)
		return
	}

	log.Printf("\n" + strings.Repeat("=", 60))
	log.Printf("📥 Received transcription result:")
	log.Printf("  Track ID: %s", result.TrackID)
	log.Printf("  User ID: %s", result.UserID)
	log.Printf("  Status: %s", result.Status)
	log.Printf("  JSON URL: %s", result.JSONURL)
	log.Printf("  Phrases: %d", result.Transcription.PhraseCount)
	log.Printf("  Duration: %.2f seconds", result.Transcription.TotalDuration)
	log.Printf(strings.Repeat("=", 60) + "\n")

	// TODO: Update database with transcription result
	// Example:
	// - Update track status to "transcribed"
	// - Store JSON URL in database
	// - Update track metadata
	// - Send notification to user

	// For now, just log the result
	updateTrackTranscriptionStatus(result.TrackID, result.JSONURL, result.Transcription.PhraseCount, result.Transcription.TotalDuration)
}

// updateTrackTranscriptionStatus updates the track with transcription information
func updateTrackTranscriptionStatus(trackID string, jsonURL string, phraseCount int, duration float64) {
	// TODO: Implement database update
	// This is a placeholder - you should update your database schema to include:
	// - transcription_json_url (TEXT)
	// - transcription_status (VARCHAR) - e.g., "pending", "processing", "completed", "failed"
	// - transcription_phrase_count (INTEGER)
	// - transcription_duration (FLOAT)

	log.Printf("📝 Updating track %s with transcription data", trackID)
	log.Printf("   JSON URL: %s", jsonURL)
	log.Printf("   Phrases: %d", phraseCount)
	log.Printf("   Duration: %.2f", duration)

	// Example update query (uncomment and adapt to your schema):
	/*
	db := getDB() // Your database connection
	_, err := db.Exec(`
		UPDATE livekit_tracks
		SET
			transcription_status = 'completed',
			transcription_json_url = $1,
			transcription_phrase_count = $2,
			transcription_duration = $3,
			updated_at = NOW()
		WHERE id = $4
	`, jsonURL, phraseCount, duration, trackID)

	if err != nil {
		log.Printf("❌ Failed to update track %s: %v", trackID, err)
	} else {
		log.Printf("✅ Track %s updated successfully", trackID)
	}
	*/
}

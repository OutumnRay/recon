package rabbitmq

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher handles RabbitMQ publishing operations
type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

// TranscriptionMessage represents a message for transcription queue
type TranscriptionMessage struct {
	TrackID  string `json:"track_id"`
	UserID   string `json:"user_id"`
	AudioURL string `json:"audio_url"`
	Language string `json:"language,omitempty"`
	Token    string `json:"token,omitempty"`
}

// NewPublisher creates a new RabbitMQ publisher
func NewPublisher(host string, port int, user string, password string, queue string) (*Publisher, error) {
	connStr := fmt.Sprintf("amqp://%s:%s@%s:%d/", user, password, host, port)

	conn, err := amqp.Dial(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare queue
	_, err = channel.QueueDeclare(
		queue, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &Publisher{
		conn:    conn,
		channel: channel,
		queue:   queue,
	}, nil
}

// PublishTranscriptionTask publishes a transcription task to the queue
func (p *Publisher) PublishTranscriptionTask(trackID uuid.UUID, userID uuid.UUID, audioURL string, language string, token string) error {
	message := TranscriptionMessage{
		TrackID:  trackID.String(),
		UserID:   userID.String(),
		AudioURL: audioURL,
		Language: language,
		Token:    token,
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = p.channel.Publish(
		"",      // exchange
		p.queue, // routing key (queue name)
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Close closes the RabbitMQ connection
func (p *Publisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

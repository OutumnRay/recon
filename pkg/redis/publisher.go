// Package redis provides a simple Redis client for publishing transcription tasks.
// Tasks are pushed with LPUSH; the Python worker consumes them with BRPOP.
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const defaultQueue = "recontext:transcription:queue"

// TranscriptionTask is the task payload consumed by the Python worker.
type TranscriptionTask struct {
	TaskID     string `json:"task_id"`
	SessionID  string `json:"session_id"`   // usually the LiveKit track ID
	UserID     string `json:"user_id"`
	VideoType  string `json:"video_type"`   // "conference" | "lecture"
	Bucket     string `json:"bucket"`
	ObjectPath string `json:"object_path"`  // MinIO object key of the source file
	Language   string `json:"language,omitempty"`
	CreatedAt  string `json:"created_at"`
}

// Publisher publishes transcription tasks to a Redis list.
type Publisher struct {
	rdb   *redis.Client
	queue string
}

// NewPublisher creates a connected Redis publisher.
func NewPublisher(host string, port int, password string, db int, queue string) (*Publisher, error) {
	if queue == "" {
		queue = defaultQueue
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Password:     password,
		DB:           db,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &Publisher{rdb: rdb, queue: queue}, nil
}

// PublishTranscriptionTask enqueues a task for the Python transcription worker.
func (p *Publisher) PublishTranscriptionTask(
	sessionID uuid.UUID,
	userID uuid.UUID,
	bucket string,
	objectPath string,
	videoType string,
	language string,
) error {
	task := TranscriptionTask{
		TaskID:     uuid.New().String(),
		SessionID:  sessionID.String(),
		UserID:     userID.String(),
		VideoType:  videoType,
		Bucket:     bucket,
		ObjectPath: objectPath,
		Language:   language,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := p.rdb.LPush(ctx, p.queue, payload).Err(); err != nil {
		return fmt.Errorf("redis LPUSH failed: %w", err)
	}

	return nil
}

// Close closes the underlying Redis connection.
func (p *Publisher) Close() error {
	return p.rdb.Close()
}

// Ping checks connectivity.
func (p *Publisher) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return p.rdb.Ping(ctx).Err()
}

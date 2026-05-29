package main

// FileResultConsumer consumes transcription results that the Python worker pushes
// to the Redis result queue (REDIS_RESULT_QUEUE, default "recontext:transcription:results").
//
// For every completed task it:
//  1. Downloads paragraphs.json from MinIO
//  2. Bulk-inserts phrases into file_transcription_phrases
//  3. Updates the uploaded_files row (status=completed, duration)
//
// For failed tasks it marks the file as failed with the error message.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/database"
	"Recontext.online/pkg/storage"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
)

const defaultResultQueue = "recontext:transcription:results"

// FileResultConsumer listens on the Redis result queue and persists transcription
// results produced by the Python worker into PostgreSQL.
type FileResultConsumer struct {
	rdb    *redis.Client
	queue  string
	db     *database.DB
	minio  *storage.MinIOClient
	up     *UserPortal
	stopCh chan struct{}
}

// newFileResultConsumer creates the consumer.
// Returns nil, nil when REDIS_HOST is unset (consumer disabled gracefully).
func newFileResultConsumer(up *UserPortal) (*FileResultConsumer, error) {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		up.logger.Info("[FileConsumer] REDIS_HOST not set — result consumer disabled")
		return nil, nil
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	password := os.Getenv("REDIS_PASSWORD")
	queue := os.Getenv("REDIS_RESULT_QUEUE")
	if queue == "" {
		queue = defaultResultQueue
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", host, port),
		Password:     password,
		DB:           0,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  35 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		up.logger.Infof("[FileConsumer] Redis ping failed (%v) — result consumer disabled", err)
		return nil, nil
	}

	up.logger.Infof("[FileConsumer] Connected, watching queue=%s", queue)
	return &FileResultConsumer{
		rdb:    rdb,
		queue:  queue,
		db:     up.db,
		minio:  up.minioClient,
		up:     up,
		stopCh: make(chan struct{}),
	}, nil
}

// Start runs the consumer loop in a background goroutine.
func (c *FileResultConsumer) Start() {
	go c.loop()
}

// Stop signals the consumer to stop.
func (c *FileResultConsumer) Stop() {
	close(c.stopCh)
}

func (c *FileResultConsumer) loop() {
	c.up.logger.Infof("[FileConsumer] Started — listening on %s", c.queue)
	for {
		select {
		case <-c.stopCh:
			c.up.logger.Infof("[FileConsumer] Stopped")
			return
		default:
		}

		ctx, cancel := context.WithTimeout(context.Background(), 32*time.Second)
		res, err := c.rdb.BRPop(ctx, 30*time.Second, c.queue).Result()
		cancel()

		if err == redis.Nil {
			continue
		}
		if err != nil {
			c.up.logger.Errorf("[FileConsumer] BRPOP error: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if len(res) < 2 {
			continue
		}

		c.handleResult(res[1])
	}
}

// isTransientDBError reports whether err is a recoverable PostgreSQL/network error
// that warrants a retry (e.g. server in recovery mode, connection reset).
func isTransientDBError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "recovery mode") ||
		strings.Contains(msg, "57P03") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "EOF")
}

// retryDB retries fn up to maxAttempts times when the error is transient.
// Delays: 2 s, 4 s, 6 s … (linear back-off).
func retryDB(maxAttempts int, fn func() error) error {
	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		if lastErr = fn(); lastErr == nil {
			return nil
		}
		if !isTransientDBError(lastErr) {
			return lastErr
		}
		time.Sleep(time.Duration(i+1) * 2 * time.Second)
	}
	return lastErr
}

func (c *FileResultConsumer) handleResult(payload string) {
	var cb models.FileResultCallback
	if err := json.Unmarshal([]byte(payload), &cb); err != nil {
		c.up.logger.Errorf("[FileConsumer] Failed to parse result: %v", err)
		return
	}

	c.up.logger.Infof("[FileConsumer] Result: session=%s status=%s", cb.SessionID, cb.Status)

	fileID, err := uuid.Parse(cb.SessionID)
	if err != nil {
		c.up.logger.Errorf("[FileConsumer] Invalid session_id UUID %q: %v", cb.SessionID, err)
		return
	}

	if cb.Status == "failed" {
		_ = c.db.SetFileFailed(fileID, cb.Error)
		c.up.logger.Infof("[FileConsumer] File %s failed: %s", fileID, cb.Error)
		return
	}

	_ = retryDB(5, func() error {
		return c.db.UpdateFileProgress(fileID, "transcribing", "saving", 90)
	})

	if cb.TranscriptPath != "" {
		if err := retryDB(5, func() error {
			return c.savePhrases(fileID, cb.TranscriptPath)
		}); err != nil {
			c.up.logger.Errorf("[FileConsumer] Failed to save phrases for %s: %v", fileID, err)
			_ = retryDB(3, func() error {
				return c.db.SetFileFailed(fileID, fmt.Sprintf("failed to save phrases: %v", err))
			})
			return
		}
	}

	if err := retryDB(5, func() error {
		return c.db.SetFileCompleted(fileID, cb.Duration)
	}); err != nil {
		c.up.logger.Errorf("[FileConsumer] Failed to mark file %s completed: %v", fileID, err)
		return
	}

	c.up.logger.Infof("[FileConsumer] File %s completed — paragraphs=%d duration=%.1fs",
		fileID, cb.ParagraphsCount, cb.Duration)
}

// savePhrases downloads paragraphs.json from MinIO and bulk-inserts phrases.
func (c *FileResultConsumer) savePhrases(fileID uuid.UUID, transcriptPath string) error {
	if c.minio == nil {
		return fmt.Errorf("MinIO client not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	obj, err := c.minio.GetClient().GetObject(
		ctx,
		c.minio.GetBucket(),
		transcriptPath,
		minio.GetObjectOptions{},
	)
	if err != nil {
		return fmt.Errorf("MinIO GetObject: %w", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return fmt.Errorf("read paragraphs.json: %w", err)
	}

	// Format from Python worker:
	// {"session_id": "...", "paragraphs": [{"start":0,"end":4,"text":"...","speaker":"SPEAKER_00"}]}
	var doc struct {
		Paragraphs []models.FileResultParagraph `json:"paragraphs"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("parse paragraphs.json: %w", err)
	}

	if len(doc.Paragraphs) == 0 {
		c.up.logger.Infof("[FileConsumer] paragraphs.json empty for file %s", fileID)
		return nil
	}

	phrases := make([]database.FileTranscriptionPhrase, 0, len(doc.Paragraphs))
	for i, p := range doc.Paragraphs {
		phrases = append(phrases, database.FileTranscriptionPhrase{
			FileID:      fileID,
			PhraseIndex: i,
			StartTime:   p.Start,
			EndTime:     p.End,
			Text:        p.Text,
			Speaker:     p.Speaker,
		})
	}

	_ = c.db.DeleteFilePhrases(fileID) // idempotent
	return c.db.BulkInsertPhrases(phrases)
}

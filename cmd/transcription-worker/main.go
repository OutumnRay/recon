package main

import (
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"Recontext.online/internal/config"
	"Recontext.online/internal/models"
	"Recontext.online/pkg/logger"
	"github.com/google/uuid"
)

// @title Recontext Transcription Worker API
// @version 1.0
// @description Transcription worker service for audio/video processing using Whisper
// @host localhost:8082
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token

type TranscriptionWorker struct {
	config   *config.Config
	logger   *logger.Logger
	tasks    map[string]*models.TranscriptionTask
	tasksMux sync.RWMutex
	isActive bool
}

func main() {
	log := logger.New()
	log.Info("Starting Transcription Worker...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	worker := &TranscriptionWorker{
		config:   cfg,
		logger:   log,
		tasks:    make(map[string]*models.TranscriptionTask),
		isActive: true,
	}

	// Start HTTP server for health checks
	http.HandleFunc("/health", worker.healthHandler)
	http.HandleFunc("/status", worker.statusHandler)

	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8082"
		}
		log.Infof("HTTP server listening on port %s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Errorf("HTTP server error: %v", err)
		}
	}()

	// Start worker loop
	go worker.processLoop()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down gracefully...")
	worker.isActive = false
	time.Sleep(2 * time.Second)
	log.Info("Transcription Worker stopped")
}

func (w *TranscriptionWorker) processLoop() {
	w.logger.Info("Transcription worker loop started")
	w.logger.Info("Waiting for transcription tasks from RabbitMQ...")

	// TODO: Implement RabbitMQ connection and message consumption
	// For now, this is a placeholder that logs periodically
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if w.isActive {
				w.logger.Debug("Transcription worker is running (waiting for RabbitMQ integration)")
			} else {
				return
			}
		}
	}
}

// simulateTranscription - Placeholder for actual Whisper transcription
func (w *TranscriptionWorker) simulateTranscription(taskID, audioURL string) (*models.Transcript, error) {
	w.logger.Infof("Processing transcription task %s for %s", taskID, audioURL)

	// TODO: Implement actual Whisper integration
	// 1. Download audio file from MinIO
	// 2. Run Whisper transcription
	// 3. Perform speaker diarization
	// 4. Generate transcript segments
	// 5. Upload results to MinIO
	// 6. Save metadata to PostgreSQL
	// 7. Send results to RabbitMQ

	// Simulate processing time
	time.Sleep(2 * time.Second)

	transcript := &models.Transcript{
		ID:           uuid.New(),
		RecordingID:  uuid.MustParse(taskID),
		Language:     "en",
		Status:       "completed",
		Segments:     []models.TranscriptSegment{},
		ProcessedAt:  time.Now(),
		DurationSecs: 0,
	}

	return transcript, nil
}

// healthHandler godoc
// @Summary Health check
// @Description Check if the transcription worker is healthy
// @Tags Health
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Router /health [get]
func (w *TranscriptionWorker) healthHandler(rw http.ResponseWriter, r *http.Request) {
	response := models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(response)
}

// statusHandler godoc
// @Summary Get worker status
// @Description Get current status and statistics of the transcription worker
// @Tags Status
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /status [get]
func (w *TranscriptionWorker) statusHandler(rw http.ResponseWriter, r *http.Request) {
	w.tasksMux.RLock()
	defer w.tasksMux.RUnlock()

	status := map[string]interface{}{
		"service":         "transcription-worker",
		"status":          "active",
		"active_tasks":    len(w.tasks),
		"uptime":          time.Now().Format(time.RFC3339),
		"rabbitmq_status": "not_connected", // TODO: Implement actual status
		"whisper_status":  "not_initialized", // TODO: Implement actual status
	}

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(status)
}

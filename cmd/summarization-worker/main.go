package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"Recontext.online/internal/config"
	"Recontext.online/internal/models"
	"Recontext.online/pkg/logger"
)

// @title Recontext Summarization Worker API
// @version 1.0
// @description Summarization worker service for generating summaries from transcripts using LLMs
// @host localhost:8083
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token

type SummarizationWorker struct {
	config   *config.Config
	logger   *logger.Logger
	tasks    map[string]*models.SummarizationTask
	tasksMux sync.RWMutex
	isActive bool
}

func main() {
	log := logger.New()
	log.Info("Starting Summarization Worker...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	worker := &SummarizationWorker{
		config:   cfg,
		logger:   log,
		tasks:    make(map[string]*models.SummarizationTask),
		isActive: true,
	}

	// Start HTTP server for health checks
	http.HandleFunc("/health", worker.healthHandler)
	http.HandleFunc("/status", worker.statusHandler)

	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8083"
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
	log.Info("Summarization Worker stopped")
}

func (w *SummarizationWorker) processLoop() {
	w.logger.Info("Summarization worker loop started")
	w.logger.Info("Waiting for summarization tasks from RabbitMQ...")

	// TODO: Implement RabbitMQ connection and message consumption
	// For now, this is a placeholder that logs periodically
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if w.isActive {
			w.logger.Debug("Summarization worker is running (waiting for RabbitMQ integration)")
		} else {
			return
		}
	}
}

// simulateSummarization - Placeholder for actual LLM summarization
func (w *SummarizationWorker) simulateSummarization(taskID, transcriptText string) (string, error) {
	w.logger.Infof("Processing summarization task %s", taskID)

	// TODO: Implement actual LLM integration
	// 1. Fetch transcript from PostgreSQL
	// 2. Chunk text if needed (for long transcripts)
	// 3. Call LLM API (OpenAI, Anthropic, local model)
	// 4. Generate summary with key points
	// 5. Extract action items and topics
	// 6. Save summary to PostgreSQL
	// 7. Send completion message to RabbitMQ

	// Simulate processing time
	time.Sleep(2 * time.Second)

	summary := fmt.Sprintf("Summary generated at %s for transcript %s. This is a placeholder summary that will be replaced with actual LLM-generated content.",
		time.Now().Format(time.RFC3339), taskID)

	return summary, nil
}

// healthHandler godoc
// @Summary Health check
// @Description Check if the summarization worker is healthy
// @Tags Health
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Router /health [get]
func (w *SummarizationWorker) healthHandler(rw http.ResponseWriter, r *http.Request) {
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
// @Description Get current status and statistics of the summarization worker
// @Tags Status
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /status [get]
func (w *SummarizationWorker) statusHandler(rw http.ResponseWriter, r *http.Request) {
	w.tasksMux.RLock()
	defer w.tasksMux.RUnlock()

	status := map[string]interface{}{
		"service":         "summarization-worker",
		"status":          "active",
		"active_tasks":    len(w.tasks),
		"uptime":          time.Now().Format(time.RFC3339),
		"rabbitmq_status": "not_connected", // TODO: Implement actual status
		"llm_status":      "not_initialized", // TODO: Implement actual status
	}

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(status)
}

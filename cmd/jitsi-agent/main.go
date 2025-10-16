package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"Recontext.online/internal/config"
	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
	"Recontext.online/pkg/logger"
)

const version = "0.1.0"

// @title Recontext.online Jitsi Agent API
// @version 0.1.0
// @description API for Jitsi Meet recording agent - connects to conferences via WebRTC and records sessions
// @termsOfService http://recontext.online/terms/

// @contact.name API Support
// @contact.url http://recontext.online/support
// @contact.email support@recontext.online

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8084
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

type JitsiAgent struct {
	config     *config.Config
	logger     *logger.Logger
	jwtManager *auth.JWTManager
	sessions   map[string]*models.JitsiSession
	connections map[string]*models.WebRTCConnection
	mu         sync.RWMutex
	maxSessions int
}

func NewJitsiAgent(cfg *config.Config, log *logger.Logger) *JitsiAgent {
	jwtManager := auth.NewJWTManager("your-secret-key-change-in-production", 24*time.Hour)

	return &JitsiAgent{
		config:      cfg,
		logger:      log,
		jwtManager:  jwtManager,
		sessions:    make(map[string]*models.JitsiSession),
		connections: make(map[string]*models.WebRTCConnection),
		maxSessions: 10, // Max concurrent recording sessions
	}
}

// Health godoc
// @Summary Health check
// @Description Check if the agent is healthy and ready to record
// @Tags Monitoring
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Router /health [get]
func (ja *JitsiAgent) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := models.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   version,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetStatus godoc
// @Summary Get agent status
// @Description Get current status of the recording agent including active sessions
// @Tags Monitoring
// @Produce json
// @Success 200 {object} object{active_sessions=int,max_sessions=int,sessions=[]models.JitsiSession}
// @Security BearerAuth
// @Router /api/v1/status [get]
func (ja *JitsiAgent) statusHandler(w http.ResponseWriter, r *http.Request) {
	ja.mu.RLock()
	defer ja.mu.RUnlock()

	activeSessions := make([]models.JitsiSession, 0)
	for _, session := range ja.sessions {
		if session.Status == models.JitsiSessionStatusRecording {
			activeSessions = append(activeSessions, *session)
		}
	}

	response := map[string]interface{}{
		"active_sessions": len(activeSessions),
		"max_sessions":    ja.maxSessions,
		"sessions":        activeSessions,
		"agent_id":        "jitsi-agent-001",
		"timestamp":       time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StartRecording godoc
// @Summary Start recording a Jitsi session
// @Description Start recording a Jitsi Meet conference by joining the room via WebRTC
// @Tags Recording
// @Accept json
// @Produce json
// @Param request body models.StartRecordingRequest true "Recording start request"
// @Success 200 {object} models.StartRecordingResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 503 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/recording/start [post]
func (ja *JitsiAgent) startRecordingHandler(w http.ResponseWriter, r *http.Request) {
	var req models.StartRecordingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ja.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	ja.mu.Lock()
	defer ja.mu.Unlock()

	// Check capacity
	activeCount := 0
	for _, session := range ja.sessions {
		if session.Status == models.JitsiSessionStatusRecording {
			activeCount++
		}
	}

	if activeCount >= ja.maxSessions {
		ja.respondWithError(w, http.StatusServiceUnavailable, "Maximum concurrent sessions reached", "")
		return
	}

	// Create session
	sessionID := fmt.Sprintf("session-%d", time.Now().Unix())
	session := &models.JitsiSession{
		ID:        sessionID,
		RoomName:  req.RoomName,
		RoomURL:   req.RoomURL,
		Status:    models.JitsiSessionStatusRecording,
		StartedAt: time.Now(),
		Participants: 0,
	}

	ja.sessions[sessionID] = session

	// Start recording in a goroutine (thread)
	go ja.recordSession(sessionID, req)

	ja.logger.Infof("Started recording session %s for room %s", sessionID, req.RoomName)

	response := models.StartRecordingResponse{
		SessionID: sessionID,
		Status:    string(models.JitsiSessionStatusRecording),
		StartedAt: session.StartedAt,
		Message:   "Recording started successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StopRecording godoc
// @Summary Stop recording a Jitsi session
// @Description Stop an active recording session
// @Tags Recording
// @Accept json
// @Produce json
// @Param request body models.StopRecordingRequest true "Recording stop request"
// @Success 200 {object} models.StopRecordingResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/recording/stop [post]
func (ja *JitsiAgent) stopRecordingHandler(w http.ResponseWriter, r *http.Request) {
	var req models.StopRecordingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ja.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	ja.mu.Lock()
	defer ja.mu.Unlock()

	session, exists := ja.sessions[req.SessionID]
	if !exists {
		ja.respondWithError(w, http.StatusNotFound, "Session not found", "")
		return
	}

	// Update session status
	now := time.Now()
	session.Status = models.JitsiSessionStatusCompleted
	session.EndedAt = &now

	// Generate recording ID
	recordingID := fmt.Sprintf("rec-%d", time.Now().Unix())
	session.RecordingID = recordingID

	ja.logger.Infof("Stopped recording session %s, recording saved as %s", req.SessionID, recordingID)

	response := models.StopRecordingResponse{
		SessionID:   req.SessionID,
		RecordingID: recordingID,
		EndedAt:     now,
		Message:     "Recording stopped and saved",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListSessions godoc
// @Summary List recording sessions
// @Description Get list of all recording sessions
// @Tags Recording
// @Produce json
// @Success 200 {object} map[string]models.JitsiSession
// @Security BearerAuth
// @Router /api/v1/sessions [get]
func (ja *JitsiAgent) listSessionsHandler(w http.ResponseWriter, r *http.Request) {
	ja.mu.RLock()
	defer ja.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ja.sessions)
}

// recordSession simulates the WebRTC recording process
// In production, this would use pion/webrtc or similar library
func (ja *JitsiAgent) recordSession(sessionID string, req models.StartRecordingRequest) {
	ja.logger.Infof("Recording thread started for session %s", sessionID)

	// Simulate WebRTC connection
	connection := &models.WebRTCConnection{
		SessionID:    sessionID,
		PeerID:       fmt.Sprintf("peer-%d", time.Now().Unix()),
		IsConnected:  true,
		StreamActive: true,
		ConnectedAt:  time.Now(),
	}

	ja.mu.Lock()
	ja.connections[sessionID] = connection
	ja.mu.Unlock()

	// TODO: Implement actual WebRTC connection using pion/webrtc
	// 1. Connect to Jitsi Meet server
	// 2. Join conference room
	// 3. Receive audio/video streams
	// 4. Write streams to file (MP4/WebM)
	// 5. Upload to MinIO when done
	// 6. Send to RabbitMQ for processing

	ja.logger.Infof("WebRTC connection established for session %s", sessionID)

	// Simulate recording (in production, this would be actual recording)
	// Keep connection alive and record until stopped
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ja.mu.RLock()
			session, exists := ja.sessions[sessionID]
			ja.mu.RUnlock()

			if !exists || session.Status != models.JitsiSessionStatusRecording {
				ja.logger.Infof("Recording session %s ended", sessionID)
				return
			}

			ja.logger.Debugf("Recording session %s still active", sessionID)
		}
	}
}

func (ja *JitsiAgent) respondWithError(w http.ResponseWriter, code int, message string, detail string) {
	response := models.ErrorResponse{
		Error:     message,
		Message:   detail,
		Code:      code,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

func chainMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func (ja *JitsiAgent) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Public endpoints
	mux.HandleFunc("/health", ja.healthHandler)

	// Protected endpoints
	authMiddleware := auth.AuthMiddleware(ja.jwtManager)

	mux.Handle("/api/v1/status", chainMiddleware(
		http.HandlerFunc(ja.statusHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/recording/start", chainMiddleware(
		http.HandlerFunc(ja.startRecordingHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/recording/stop", chainMiddleware(
		http.HandlerFunc(ja.stopRecordingHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/sessions", chainMiddleware(
		http.HandlerFunc(ja.listSessionsHandler),
		authMiddleware,
	))

	return mux
}

func (ja *JitsiAgent) Start() error {
	mux := ja.setupRoutes()

	addr := fmt.Sprintf("%s:%d", ja.config.Server.Host, 8084)
	ja.logger.Infof("Jitsi Agent starting on %s", addr)
	ja.logger.Infof("Version: %s", version)
	ja.logger.Infof("Max concurrent sessions: %d", ja.maxSessions)
	ja.logger.Infof("Swagger docs: http://%s/swagger/index.html", addr)

	// Register with managing portal
	go ja.registerWithManagingPortal()

	return http.ListenAndServe(addr, mux)
}

func (ja *JitsiAgent) registerWithManagingPortal() {
	// TODO: Implement registration with managing portal
	ja.logger.Info("Registering with managing portal...")
}

func main() {
	log := logger.New()
	log.Info("Starting Jitsi Agent...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	agent := NewJitsiAgent(cfg, log)

	if err := agent.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

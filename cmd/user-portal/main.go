package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"Recontext.online/internal/config"
	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
	"Recontext.online/pkg/logger"
)

//go:embed dist/*
var staticFiles embed.FS

const version = "0.1.0"

// @title Recontext.online User Portal API
// @version 0.1.0
// @description API for users to interact with Recontext.online platform - upload recordings, view transcripts, search content
// @termsOfService http://recontext.online/terms/

// @contact.name API Support
// @contact.url http://recontext.online/support
// @contact.email support@recontext.online

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

type UserPortal struct {
	config     *config.Config
	logger     *logger.Logger
	jwtManager *auth.JWTManager
	users      map[string]*models.User      // In-memory user store
	recordings map[string]*models.Recording // In-memory recordings store
}

func NewUserPortal(cfg *config.Config, log *logger.Logger) *UserPortal {
	jwtManager := auth.NewJWTManager("your-secret-key-change-in-production", 24*time.Hour)

	// Create default users
	users := make(map[string]*models.User)
	users["user"] = &models.User{
		ID:        "user-001",
		Username:  "user",
		Email:     "user@recontext.online",
		Password:  auth.HashPassword("user123"),
		Role:      models.RoleUser,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return &UserPortal{
		config:     cfg,
		logger:     log,
		jwtManager: jwtManager,
		users:      users,
		recordings: make(map[string]*models.Recording),
	}
}

// Login godoc
// @Summary User login
// @Description Authenticate user and receive JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/auth/login [post]
func (up *UserPortal) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	user, exists := up.users[req.Username]
	if !exists || !auth.VerifyPassword(req.Password, user.Password) {
		up.respondWithError(w, http.StatusUnauthorized, "Invalid credentials", "username or password incorrect")
		return
	}

	token, expiresAt, err := up.jwtManager.GenerateToken(user)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to generate token", err.Error())
		return
	}

	response := models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: models.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Health godoc
// @Summary Health check
// @Description Check if the service is healthy
// @Tags Monitoring
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Router /health [get]
func (up *UserPortal) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := models.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   version,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UploadRecording godoc
// @Summary Upload a recording
// @Description Upload an audio or video file for transcription
// @Tags Recordings
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "Recording title"
// @Param file formData file true "Audio/video file"
// @Success 200 {object} models.UploadResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/recordings/upload [post]
func (up *UserPortal) uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Parse multipart form (max 500MB)
	if err := r.ParseMultipartForm(500 << 20); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Failed to parse form", err.Error())
		return
	}

	title := r.FormValue("title")
	if title == "" {
		up.respondWithError(w, http.StatusBadRequest, "Title is required", "")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		up.respondWithError(w, http.StatusBadRequest, "File is required", err.Error())
		return
	}
	defer file.Close()

	// Create recording
	recordingID := fmt.Sprintf("rec-%d", time.Now().Unix())
	recording := &models.Recording{
		ID:          recordingID,
		UserID:      claims.UserID,
		Title:       title,
		FileName:    header.Filename,
		FileSize:    header.Size,
		MimeType:    header.Header.Get("Content-Type"),
		StoragePath: fmt.Sprintf("uploads/%s/%s", claims.UserID, header.Filename),
		Status:      models.RecordingStatusQueued,
		UploadedAt:  time.Now(),
	}

	up.recordings[recordingID] = recording
	up.logger.Infof("Recording uploaded: %s by user %s", recordingID, claims.Username)

	// TODO: Upload file to MinIO
	// TODO: Send message to RabbitMQ for processing

	response := models.UploadResponse{
		RecordingID: recordingID,
		UploadURL:   fmt.Sprintf("https://storage.recontext.online/%s", recording.StoragePath),
		Message:     "File uploaded successfully and queued for processing",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListRecordings godoc
// @Summary List user recordings
// @Description Get a paginated list of user's recordings
// @Tags Recordings
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param status query string false "Filter by status"
// @Success 200 {object} models.ListRecordingsResponse
// @Failure 401 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/recordings [get]
func (up *UserPortal) listRecordingsHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Filter recordings by user
	var userRecordings []models.Recording
	for _, rec := range up.recordings {
		if rec.UserID == claims.UserID {
			userRecordings = append(userRecordings, *rec)
		}
	}

	response := models.ListRecordingsResponse{
		Recordings: userRecordings,
		Total:      len(userRecordings),
		Page:       1,
		PageSize:   20,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetRecording godoc
// @Summary Get recording details
// @Description Get detailed information about a specific recording
// @Tags Recordings
// @Produce json
// @Param id path string true "Recording ID"
// @Success 200 {object} models.Recording
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/recordings/{id} [get]
func (up *UserPortal) getRecordingHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Extract recording ID from path (simplified - in production use a router)
	recordingID := r.URL.Path[len("/api/v1/recordings/"):]

	recording, exists := up.recordings[recordingID]
	if !exists {
		up.respondWithError(w, http.StatusNotFound, "Recording not found", "")
		return
	}

	// Check ownership
	if recording.UserID != claims.UserID && claims.Role != models.RoleAdmin {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recording)
}

// SearchTranscripts godoc
// @Summary Search transcripts
// @Description Search through transcripts using semantic or keyword search
// @Tags Search
// @Produce json
// @Param query query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} models.SearchResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/search [get]
func (up *UserPortal) searchHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	query := r.URL.Query().Get("query")
	if query == "" {
		up.respondWithError(w, http.StatusBadRequest, "Query parameter is required", "")
		return
	}

	// TODO: Implement actual search using Qdrant

	response := models.SearchResponse{
		Results:  []models.SearchResult{},
		Total:    0,
		Page:     1,
		PageSize: 10,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (up *UserPortal) respondWithError(w http.ResponseWriter, code int, message string, detail string) {
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

// serveStaticFiles serves the React frontend
func serveStaticFiles() http.Handler {
	// Get the dist subdirectory from embedded files
	distFS, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		panic(fmt.Sprintf("Failed to get dist directory: %v", err))
	}

	fileServer := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if file exists
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Try to open the file
		file, err := distFS.Open(path)
		if err != nil {
			// File doesn't exist, serve index.html for SPA routing
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		file.Close()

		// File exists, serve it
		fileServer.ServeHTTP(w, r)
	})
}

func (up *UserPortal) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Public endpoints
	mux.HandleFunc("/health", up.healthHandler)
	mux.HandleFunc("/api/v1/auth/login", up.loginHandler)

	// Protected endpoints
	authMiddleware := auth.AuthMiddleware(up.jwtManager)

	mux.Handle("/api/v1/recordings/upload", chainMiddleware(
		http.HandlerFunc(up.uploadHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/recordings", chainMiddleware(
		http.HandlerFunc(up.listRecordingsHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/recordings/", chainMiddleware(
		http.HandlerFunc(up.getRecordingHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/search", chainMiddleware(
		http.HandlerFunc(up.searchHandler),
		authMiddleware,
	))

	// Serve React frontend for all other routes
	mux.Handle("/", serveStaticFiles())

	return mux
}

func (up *UserPortal) Start() error {
	mux := up.setupRoutes()

	addr := fmt.Sprintf("%s:%d", up.config.Server.Host, up.config.Server.Port)
	up.logger.Infof("User Portal starting on %s", addr)
	up.logger.Infof("Version: %s", version)
	up.logger.Infof("Swagger docs: http://%s/swagger/index.html", addr)
	up.logger.Infof("Default user credentials: username=user, password=user123")

	return http.ListenAndServe(addr, mux)
}

func main() {
	log := logger.New()
	log.Info("Starting User Portal...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override port for user portal
	cfg.Server.Port = 8081

	portal := NewUserPortal(cfg, log)

	if err := portal.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

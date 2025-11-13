package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"

	"Recontext.online/internal/config"
	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
	"Recontext.online/pkg/database"
	"Recontext.online/pkg/email"
	"Recontext.online/pkg/embeddings"
	"Recontext.online/pkg/logger"
	"Recontext.online/pkg/metrics"

	_ "Recontext.online/cmd/user-portal/docs" // Import generated docs
)

//go:embed dist/*
var staticFiles embed.FS

const version = "0.1.0"

// @title Recontext.online User Portal API
// @version 0.1.0
// @description API пользовательского портала Recontext.online: загрузка записей, транскрибация, поиск и управление материалами
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
	config            *config.Config
	logger            *logger.Logger
	jwtManager        *auth.JWTManager
	db                *database.DB                   // Database connection
	userRepo          *database.UserRepository       // User repository
	departmentRepo    *database.DepartmentRepository // Department repository
	meetingRepo       *database.MeetingRepository    // Meeting repository
	recordings        map[string]*models.Recording   // In-memory recordings store
	prometheusMetrics *metrics.ServiceMetrics        // Prometheus metrics
	embeddingsClient  *embeddings.EmbeddingsClient   // Embeddings client for RAG
	emailService      EmailServiceInterface          // Email service for sending emails
}

// EmailServiceInterface defines the interface for email services
type EmailServiceInterface interface {
	SendPasswordResetEmail(toEmail string, data email.PasswordResetEmailData) error
}

func NewUserPortal(cfg *config.Config, log *logger.Logger) (*UserPortal, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production"
	}
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)

	// Initialize database connection
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "recontext"),
		Password: getEnv("DB_PASSWORD", "recontext"),
		DBName:   getEnv("DB_NAME", "recontext"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}

	db, err := database.NewDB(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info("Database connected and migrations completed")

	// Initialize repositories
	userRepo := database.NewUserRepository(db)
	departmentRepo := database.NewDepartmentRepository(db)
	meetingRepo := database.NewMeetingRepository(db)

	// Initialize embeddings client for RAG
	embeddingsClient := embeddings.NewEmbeddingsClient()

	// Initialize email service using unified mailer
	emailConfig := email.LoadConfigFromEnv()
	mailer := email.NewMailer(emailConfig)

	return &UserPortal{
		config:            cfg,
		logger:            log,
		jwtManager:        jwtManager,
		db:                db,
		userRepo:          userRepo,
		departmentRepo:    departmentRepo,
		meetingRepo:       meetingRepo,
		recordings:        make(map[string]*models.Recording),
		prometheusMetrics: metrics.NewServiceMetrics("user_portal"),
		embeddingsClient:  embeddingsClient,
		emailService:      mailer,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// Login godoc
// @Summary Вход в систему
// @Description Аутентификация пользователя и получение JWT токена
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Учетные данные для входа"
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

	// Find user in database (username field contains email address)
	user, err := up.userRepo.GetByEmail(req.Username)
	if err != nil || !auth.VerifyPassword(req.Password, user.Password) {
		up.respondWithError(w, http.StatusUnauthorized, "Invalid credentials", "email or password incorrect")
		return
	}

	// Check if user is active
	if !user.IsActive {
		up.respondWithError(w, http.StatusUnauthorized, "Account is inactive", "")
		return
	}

	// Update last login
	up.userRepo.UpdateLastLogin(user.ID)

	token, expiresAt, err := up.jwtManager.GenerateToken(user)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to generate token", err.Error())
		return
	}

	response := models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: models.UserInfo{
			ID:           user.ID,
			Username:     user.Username,
			Email:        user.Email,
			Role:         user.Role,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			Phone:        user.Phone,
			Bio:          user.Bio,
			Avatar:       user.Avatar,
			DepartmentID: user.DepartmentID,
			Permissions:  user.Permissions,
			Language:     user.Language,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Health godoc
// @Summary Проверка здоровья
// @Description Проверка, работает ли сервис
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
// @Summary Загрузить запись
// @Description Загрузить аудио или видео файл для транскрибации
// @Tags Recordings
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "Название записи"
// @Param file formData file true "Аудио- или видеофайл"
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
	recording := &models.Recording{
		ID:          uuid.New(),
		UserID:      claims.UserID,
		Title:       title,
		FileName:    header.Filename,
		FileSize:    header.Size,
		MimeType:    header.Header.Get("Content-Type"),
		StoragePath: fmt.Sprintf("uploads/%s/%s", claims.UserID, header.Filename),
		Status:      models.RecordingStatusQueued,
		UploadedAt:  time.Now(),
	}

	up.recordings[recording.ID.String()] = recording
	up.logger.Infof("Recording uploaded: %s by user %s", recording.ID, claims.Username)

	// TODO: Upload file to MinIO
	// TODO: Send message to RabbitMQ for processing

	response := models.UploadResponse{
		RecordingID: recording.ID,
		UploadURL:   fmt.Sprintf("https://storage.recontext.online/%s", recording.StoragePath),
		Message:     "File uploaded successfully and queued for processing",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListRecordings godoc
// @Summary Список записей пользователя
// @Description Получить постраничный список записей пользователя
// @Tags Recordings
// @Produce json
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(20)
// @Param status query string false "Фильтр по статусу"
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
// @Summary Получить детали записи
// @Description Получить детальную информацию о конкретной записи
// @Tags Recordings
// @Produce json
// @Param id path string true "Идентификатор записи"
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
// @Summary Поиск по транскриптам
// @Description Поиск по транскриптам с использованием семантического или ключевого поиска
// @Tags Search
// @Produce json
// @Param query query string true "Поисковый запрос"
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(10)
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

// UploadFile godoc
// @Summary Загрузить файл для транскрибации
// @Description Загрузить аудио или видео файл для транскрибации (требуется разрешение на загрузку файлов)
// @Tags Files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Аудио- или видеофайл"
// @Param description formData string false "Описание файла"
// @Success 200 {object} models.FileUploadResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/files/upload [post]
func (up *UserPortal) uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Check if user has file upload permission
	hasPermission, err := up.db.CheckUserHasFilePermission(claims.UserID, "write")
	if err != nil || !hasPermission {
		up.respondWithError(w, http.StatusForbidden, "You don't have permission to upload files", "Contact administrator to grant file upload access")
		return
	}

	// Parse multipart form (max 500MB)
	if err := r.ParseMultipartForm(500 << 20); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Failed to parse form", err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		up.respondWithError(w, http.StatusBadRequest, "File is required", err.Error())
		return
	}
	defer file.Close()

	// Create file record
	fileID := uuid.New()
	groupFileUploadersID := uuid.Must(uuid.Parse("00000000-0000-0000-0000-000000000001")) // Use a fixed UUID for the file uploaders group
	uploadedFile := &models.UploadedFile{
		ID:           fileID,
		Filename:     fmt.Sprintf("%d-%s", time.Now().Unix(), header.Filename),
		OriginalName: header.Filename,
		FileSize:     header.Size,
		MimeType:     header.Header.Get("Content-Type"),
		StoragePath:  fmt.Sprintf("files/%s/%s", claims.UserID, fileID),
		UserID:       claims.UserID,
		GroupID:      groupFileUploadersID,
		Status:       models.StatusPending,
		UploadedAt:   time.Now(),
	}

	// Save to database
	if err := up.db.CreateUploadedFile(uploadedFile); err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to save file record", err.Error())
		return
	}

	up.logger.Infof("File uploaded: %s by user %s", fileID, claims.Username)

	response := models.FileUploadResponse{
		ID:           fileID,
		Filename:     uploadedFile.Filename,
		OriginalName: uploadedFile.OriginalName,
		FileSize:     uploadedFile.FileSize,
		Status:       uploadedFile.Status,
		UploadedAt:   uploadedFile.UploadedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListFiles godoc
// @Summary Список загруженных файлов
// @Description Получить постраничный список загруженных файлов пользователя
// @Tags Files
// @Produce json
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(20)
// @Success 200 {object} models.ListFilesResponse
// @Failure 401 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/files [get]
func (up *UserPortal) listFilesHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Parse query parameters
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			page = val
		}
	}

	pageSize := 20
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil {
			pageSize = val
		}
	}

	// Get files from database
	files, total, err := up.db.ListUploadedFilesByUser(claims.UserID, page, pageSize)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve files", err.Error())
		return
	}

	response := models.ListFilesResponse{
		Files:    files,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CheckPermission godoc
// @Summary Проверить разрешение на загрузку файлов
// @Description Проверить, имеет ли текущий пользователь разрешение на загрузку файлов
// @Tags Files
// @Produce json
// @Success 200 {object} map[string]bool
// @Failure 401 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/files/permission [get]
func (up *UserPortal) checkFilePermissionHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	hasPermission, err := up.db.CheckUserHasFilePermission(claims.UserID, "write")
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to check permission", err.Error())
		return
	}

	response := map[string]bool{
		"hasPermission": hasPermission,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RAGSearch godoc
// @Summary Семантический поиск по транскрипциям
// @Description Выполнить семантический поиск с использованием RAG по транскрипциям
// @Tags RAG
// @Accept json
// @Produce json
// @Param request body models.RAGSearchRequest true "Параметры поискового запроса"
// @Success 200 {object} models.RAGSearchResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/rag/search [post]
func (up *UserPortal) ragSearchHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Check RAG permission
	hasPermission, err := up.db.CheckUserHasRAGPermission(claims.UserID)
	if err != nil || !hasPermission {
		up.respondWithError(w, http.StatusForbidden, "You don't have permission to use RAG search", "Contact administrator to grant RAG access")
		return
	}

	var req models.RAGSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.Query == "" {
		up.respondWithError(w, http.StatusBadRequest, "Query is required", "")
		return
	}

	// Set defaults
	if req.TopK <= 0 {
		req.TopK = 5
	}
	if req.Threshold <= 0 {
		req.Threshold = 0.7
	}

	// Generate embedding for query
	queryEmbedding, err := up.embeddingsClient.GetEmbedding(req.Query)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to generate query embedding", err.Error())
		return
	}

	// Search for similar chunks
	results, err := up.db.SearchSimilarChunks(queryEmbedding, req.TopK, req.Threshold)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to search", err.Error())
		return
	}

	response := models.RAGSearchResponse{
		Query:   req.Query,
		Results: results,
		Count:   len(results),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CheckRAGPermission godoc
// @Summary Проверить разрешение на RAG
// @Description Проверить, имеет ли текущий пользователь разрешение на использование RAG поиска
// @Tags RAG
// @Produce json
// @Success 200 {object} map[string]bool
// @Failure 401 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/rag/permission [get]
func (up *UserPortal) checkRAGPermissionHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	hasPermission, err := up.db.CheckUserHasRAGPermission(claims.UserID)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to check permission", err.Error())
		return
	}

	response := map[string]bool{
		"hasPermission": hasPermission,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RAGStatus godoc
// @Summary Получить статус системы RAG
// @Description Получить статистику о системе RAG
// @Tags RAG
// @Produce json
// @Success 200 {object} models.RAGStatusResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/rag/status [get]
func (up *UserPortal) ragStatusHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Check RAG permission
	hasPermission, err := up.db.CheckUserHasRAGPermission(claims.UserID)
	if err != nil || !hasPermission {
		up.respondWithError(w, http.StatusForbidden, "You don't have permission to access RAG status", "Contact administrator to grant RAG access")
		return
	}

	status, err := up.db.GetRAGStatus()
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get RAG status", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
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

	// Prometheus metrics endpoint (no auth required for scraping)
	mux.Handle("/metrics", promhttp.Handler())

	// Swagger documentation endpoint
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// Public endpoints
	mux.HandleFunc("/health", up.healthHandler)
	mux.HandleFunc("/api/v1/auth/login", up.loginHandler)

	// Password reset endpoints (public)
	mux.HandleFunc("/api/v1/auth/password-reset/request", up.requestPasswordResetHandler)
	mux.HandleFunc("/api/v1/auth/password-reset/verify", up.verifyResetCodeHandler)
	mux.HandleFunc("/api/v1/auth/password-reset/reset", up.resetPasswordHandler)

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

	// File upload endpoints
	mux.Handle("/api/v1/files/upload", chainMiddleware(
		http.HandlerFunc(up.uploadFileHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/files/permission", chainMiddleware(
		http.HandlerFunc(up.checkFilePermissionHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/files", chainMiddleware(
		http.HandlerFunc(up.listFilesHandler),
		authMiddleware,
	))

	// Meeting endpoints
	mux.Handle("/api/v1/meetings", chainMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				up.listMyMeetingsHandler(w, r)
			} else if r.Method == http.MethodPost {
				up.createMeetingHandler(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}),
		authMiddleware,
	))

	mux.Handle("/api/v1/meetings/", chainMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if this is a token request
			if strings.HasSuffix(r.URL.Path, "/token") && r.Method == http.MethodGet {
				up.getMeetingTokenHandler(w, r)
				return
			}

			// Check if this is a recording control request
			if strings.HasSuffix(r.URL.Path, "/recording/start") && r.Method == http.MethodPost {
				up.startRecordingHandler(w, r)
				return
			}
			if strings.HasSuffix(r.URL.Path, "/recording/stop") && r.Method == http.MethodPost {
				up.stopRecordingHandler(w, r)
				return
			}

			// Check if this is a transcription control request
			if strings.HasSuffix(r.URL.Path, "/transcription/start") && r.Method == http.MethodPost {
				up.startTranscriptionHandler(w, r)
				return
			}
			if strings.HasSuffix(r.URL.Path, "/transcription/stop") && r.Method == http.MethodPost {
				up.stopTranscriptionHandler(w, r)
				return
			}

			switch r.Method {
			case http.MethodGet:
				up.getMeetingHandler(w, r)
			case http.MethodPut:
				up.updateMeetingHandler(w, r)
			case http.MethodDelete:
				up.deleteMeetingHandler(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}),
		authMiddleware,
	))

	// Meeting subjects endpoint
	mux.Handle("/api/v1/meeting-subjects", chainMiddleware(
		http.HandlerFunc(up.listMeetingSubjectsHandler),
		authMiddleware,
	))

	// RAG endpoints
	mux.Handle("/api/v1/rag/search", chainMiddleware(
		http.HandlerFunc(up.ragSearchHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/rag/permission", chainMiddleware(
		http.HandlerFunc(up.checkRAGPermissionHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/rag/status", chainMiddleware(
		http.HandlerFunc(up.ragStatusHandler),
		authMiddleware,
	))

	// Profile endpoints - MUST come before /api/v1/users to avoid conflicts
	mux.Handle("/api/v1/users/", chainMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract path after /api/v1/users/
			pathAfterUsers := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")

			// Check if this is the list endpoint (empty path after users/)
			if pathAfterUsers == "" {
				if r.Method == http.MethodGet {
					up.listUsersHandler(w, r)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
				return
			}

			// Check if this is an avatar upload request
			if strings.HasSuffix(r.URL.Path, "/avatar") && r.Method == http.MethodPost {
				up.uploadAvatarHandler(w, r)
				return
			}

			// Handle profile GET/PUT for specific user ID
			switch r.Method {
			case http.MethodGet:
				up.getProfileHandler(w, r)
			case http.MethodPut:
				up.updateProfileHandler(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}),
		authMiddleware,
	))

	// Departments endpoint
	mux.Handle("/api/v1/departments", chainMiddleware(
		http.HandlerFunc(up.listDepartmentsHandler),
		authMiddleware,
	))

	// Serve uploaded avatars
	avatarsFS := http.FileServer(http.Dir("uploads"))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", avatarsFS))

	// Serve React frontend for all other routes
	mux.Handle("/", serveStaticFiles())

	return mux
}

func (up *UserPortal) Start() error {
	mux := up.setupRoutes()

	// Wrap with recovery and metrics middleware
	handler := recoveryMiddleware(mux)
	metricsMiddleware := metrics.HTTPMetricsMiddleware(up.prometheusMetrics)
	handler = metricsMiddleware(handler)

	addr := fmt.Sprintf("%s:%d", up.config.Server.Host, up.config.Server.Port)
	up.logger.Infof("User Portal starting on %s", addr)
	up.logger.Infof("Version: %s", version)
	up.logger.Infof("Swagger docs: http://%s/swagger/index.html", addr)
	up.logger.Infof("Prometheus metrics: http://%s/metrics", addr)
	up.logger.Infof("Default user credentials: username=user, password=user123")

	return http.ListenAndServe(addr, handler)
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

	portal, err := NewUserPortal(cfg, log)
	if err != nil {
		log.Fatalf("Failed to initialize portal: %v", err)
	}

	if err := portal.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

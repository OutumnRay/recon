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

// @title Recontext.online Managing Portal API
// @version 0.1.0
// @description API for managing and monitoring Recontext.online platform services
// @termsOfService http://recontext.online/terms/

// @contact.name API Support
// @contact.url http://recontext.online/support
// @contact.email support@recontext.online

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

type ManagingPortal struct {
	config     *config.Config
	logger     *logger.Logger
	services   map[string]models.ServiceInfo
	jwtManager *auth.JWTManager
	users      map[string]*models.User         // In-memory user store (replace with DB)
	groups     map[string]*models.UserGroup    // In-memory group store
	metrics    []models.Metric                 // In-memory metrics store
	logs       []models.LogEntry               // In-memory logs store
}

func NewManagingPortal(cfg *config.Config, log *logger.Logger) *ManagingPortal {
	jwtManager := auth.NewJWTManager("your-secret-key-change-in-production", 24*time.Hour)

	// Create default admin user
	users := make(map[string]*models.User)
	adminUser := &models.User{
		ID:        "admin-001",
		Username:  "admin",
		Email:     "admin@recontext.online",
		Password:  auth.HashPassword("admin123"),
		Role:      models.RoleAdmin,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	users["admin"] = adminUser

	// Create default groups
	groups := make(map[string]*models.UserGroup)

	// Editors group - can read and write recordings
	editorsGroup := &models.UserGroup{
		ID:          "group-editors",
		Name:        "Editors",
		Description: "Users who can view and edit recordings",
		Permissions: map[string]interface{}{
			"recordings": map[string]interface{}{
				"actions": []string{"read", "write"},
				"scope":   "all",
			},
			"transcripts": map[string]interface{}{
				"actions": []string{"read"},
				"scope":   "all",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	groups["group-editors"] = editorsGroup

	// Viewers group - read-only access
	viewersGroup := &models.UserGroup{
		ID:          "group-viewers",
		Name:        "Viewers",
		Description: "Users who can only view recordings",
		Permissions: map[string]interface{}{
			"recordings": map[string]interface{}{
				"actions": []string{"read"},
				"scope":   "all",
			},
			"transcripts": map[string]interface{}{
				"actions": []string{"read"},
				"scope":   "all",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	groups["group-viewers"] = viewersGroup

	return &ManagingPortal{
		config:     cfg,
		logger:     log,
		services:   make(map[string]models.ServiceInfo),
		jwtManager: jwtManager,
		users:      users,
		groups:     groups,
		metrics:    make([]models.Metric, 0),
		logs:       make([]models.LogEntry, 0),
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
func (mp *ManagingPortal) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Find user
	user, exists := mp.users[req.Username]
	if !exists || !auth.VerifyPassword(req.Password, user.Password) {
		mp.respondWithError(w, http.StatusUnauthorized, "Invalid credentials", "username or password incorrect")
		return
	}

	// Generate token
	token, expiresAt, err := mp.jwtManager.GenerateToken(user)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to generate token", err.Error())
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

// Register godoc
// @Summary Register new user
// @Description Register a new user account (admin only)
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Registration details"
// @Success 201 {object} models.UserInfo
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/auth/register [post]
func (mp *ManagingPortal) registerHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Check if user already exists
	if _, exists := mp.users[req.Username]; exists {
		mp.respondWithError(w, http.StatusConflict, "User already exists", "username is taken")
		return
	}

	// Create new user
	newUser := &models.User{
		ID:        fmt.Sprintf("user-%d", time.Now().Unix()),
		Username:  req.Username,
		Email:     req.Email,
		Password:  auth.HashPassword(req.Password),
		Role:      models.RoleUser,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mp.users[req.Username] = newUser

	userInfo := models.UserInfo{
		ID:       newUser.ID,
		Username: newUser.Username,
		Email:    newUser.Email,
		Role:     newUser.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(userInfo)
}

// Health godoc
// @Summary Health check
// @Description Check if the service is healthy
// @Tags Monitoring
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Router /health [get]
func (mp *ManagingPortal) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := models.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   version,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetStatus godoc
// @Summary Get system status
// @Description Get overall system status and service health
// @Tags Monitoring
// @Produce json
// @Success 200 {object} models.SystemStatus
// @Security BearerAuth
// @Router /api/v1/status [get]
func (mp *ManagingPortal) statusHandler(w http.ResponseWriter, r *http.Request) {
	status := models.SystemStatus{
		Status:   "operational",
		Services: mp.services,
		Infrastructure: map[string]interface{}{
			"database": "connected",
			"rabbitmq": "connected",
			"minio":    "connected",
		},
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// ListServices godoc
// @Summary List all services
// @Description Get list of all registered services
// @Tags Services
// @Produce json
// @Success 200 {object} map[string]models.ServiceInfo
// @Security BearerAuth
// @Router /api/v1/services [get]
func (mp *ManagingPortal) servicesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mp.services)
}

// RegisterService godoc
// @Summary Register a service
// @Description Register a new service with the managing portal
// @Tags Services
// @Accept json
// @Produce json
// @Param service body models.ServiceInfo true "Service information"
// @Success 201 {object} models.ServiceInfo
// @Failure 400 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/services/register [post]
func (mp *ManagingPortal) registerServiceHandler(w http.ResponseWriter, r *http.Request) {
	var serviceInfo models.ServiceInfo
	if err := json.NewDecoder(r.Body).Decode(&serviceInfo); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	serviceInfo.LastHeartbeat = time.Now()
	if serviceInfo.StartedAt.IsZero() {
		serviceInfo.StartedAt = time.Now()
	}

	mp.services[serviceInfo.ID] = serviceInfo
	mp.logger.Infof("Service registered: %s (%s)", serviceInfo.Name, serviceInfo.Type)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(serviceInfo)
}

// Heartbeat godoc
// @Summary Service heartbeat
// @Description Update service heartbeat to indicate it's still alive
// @Tags Services
// @Accept json
// @Produce json
// @Param heartbeat body object{service_id=string,status=string} true "Heartbeat data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/services/heartbeat [post]
func (mp *ManagingPortal) heartbeatHandler(w http.ResponseWriter, r *http.Request) {
	var heartbeat struct {
		ServiceID string `json:"service_id"`
		Status    string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&heartbeat); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if service, exists := mp.services[heartbeat.ServiceID]; exists {
		service.LastHeartbeat = time.Now()
		if heartbeat.Status != "" {
			service.Status = models.ServiceStatus(heartbeat.Status)
		}
		mp.services[heartbeat.ServiceID] = service

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	} else {
		mp.respondWithError(w, http.StatusNotFound, "Service not found", "")
	}
}

func (mp *ManagingPortal) respondWithError(w http.ResponseWriter, code int, message string, detail string) {
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

// Middleware to chain handlers
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

func (mp *ManagingPortal) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Public endpoints
	mux.HandleFunc("/health", mp.healthHandler)
	mux.HandleFunc("/api/v1/auth/login", mp.loginHandler)

	// Protected endpoints - require authentication
	authMiddleware := auth.AuthMiddleware(mp.jwtManager)

	// Admin only endpoints
	adminMiddleware := auth.RequireRole(models.RoleAdmin, models.RoleService)

	mux.Handle("/api/v1/auth/register", chainMiddleware(
		http.HandlerFunc(mp.registerHandler),
		authMiddleware,
		adminMiddleware,
	))

	// Authenticated endpoints (any role)
	mux.Handle("/api/v1/status", chainMiddleware(
		http.HandlerFunc(mp.statusHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/services", chainMiddleware(
		http.HandlerFunc(mp.servicesHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/services/register", chainMiddleware(
		http.HandlerFunc(mp.registerServiceHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/services/heartbeat", chainMiddleware(
		http.HandlerFunc(mp.heartbeatHandler),
		authMiddleware,
	))

	// User management endpoints (admin only)
	mux.Handle("/api/v1/users", chainMiddleware(
		http.HandlerFunc(mp.listUsersHandler),
		authMiddleware,
		adminMiddleware,
	))

	mux.Handle("/api/v1/users/", chainMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				mp.getUserHandler(w, r)
			case http.MethodPut:
				mp.updateUserHandler(w, r)
			case http.MethodDelete:
				mp.deleteUserHandler(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}),
		authMiddleware,
		adminMiddleware,
	))

	mux.Handle("/api/v1/users/password", chainMiddleware(
		http.HandlerFunc(mp.changePasswordHandler),
		authMiddleware,
	))

	// Group management endpoints (admin only)
	mux.Handle("/api/v1/groups", chainMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				mp.listGroupsHandler(w, r)
			} else if r.Method == http.MethodPost {
				mp.createGroupHandler(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}),
		authMiddleware,
		adminMiddleware,
	))

	mux.Handle("/api/v1/groups/", chainMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				mp.getGroupHandler(w, r)
			case http.MethodPut:
				mp.updateGroupHandler(w, r)
			case http.MethodDelete:
				mp.deleteGroupHandler(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}),
		authMiddleware,
		adminMiddleware,
	))

	mux.Handle("/api/v1/groups/add-user", chainMiddleware(
		http.HandlerFunc(mp.addUserToGroupHandler),
		authMiddleware,
		adminMiddleware,
	))

	mux.Handle("/api/v1/groups/check-permission", chainMiddleware(
		http.HandlerFunc(mp.checkPermissionHandler),
		authMiddleware,
	))

	// Metrics and telemetry endpoints
	mux.Handle("/api/v1/metrics", chainMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				mp.sendMetricsHandler(w, r)
			} else if r.Method == http.MethodGet {
				mp.queryMetricsHandler(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}),
		authMiddleware,
	))

	mux.Handle("/api/v1/metrics/system", chainMiddleware(
		http.HandlerFunc(mp.getSystemMetricsHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/logs", chainMiddleware(
		http.HandlerFunc(mp.sendLogsHandler),
		authMiddleware,
	))

	// Serve React frontend for all other routes
	mux.Handle("/", serveStaticFiles())

	return mux
}

func (mp *ManagingPortal) Start() error {
	mux := mp.setupRoutes()

	addr := fmt.Sprintf("%s:%d", mp.config.Server.Host, mp.config.Server.Port)
	mp.logger.Infof("Managing Portal starting on %s", addr)
	mp.logger.Infof("Version: %s", version)
	mp.logger.Infof("Swagger docs: http://%s/swagger/index.html", addr)
	mp.logger.Infof("Default admin credentials: username=admin, password=admin123")

	// Start heartbeat checker in background
	go mp.checkServiceHeartbeats()

	return http.ListenAndServe(addr, mux)
}

func (mp *ManagingPortal) checkServiceHeartbeats() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		for id, service := range mp.services {
			if now.Sub(service.LastHeartbeat) > 2*time.Minute {
				mp.logger.Errorf("Service %s (%s) heartbeat timeout", service.Name, service.Type)
				service.Status = models.ServiceStatusError
				mp.services[id] = service
			}
		}
	}
}

func main() {
	log := logger.New()
	log.Info("Starting Managing Portal...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	portal := NewManagingPortal(cfg, log)

	if err := portal.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

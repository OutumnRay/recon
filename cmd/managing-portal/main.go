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
	"Recontext.online/pkg/livekit"
	"Recontext.online/pkg/logger"
	"Recontext.online/pkg/metrics"
	"Recontext.online/pkg/prometheus"
	"Recontext.online/pkg/rabbitmq"

	_ "Recontext.online/cmd/managing-portal/docs" // Import generated docs
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
	config            *config.Config
	logger            *logger.Logger
	services          map[string]models.ServiceInfo
	jwtManager        *auth.JWTManager
	db                *database.DB                   // Database connection
	userRepo          *database.UserRepository       // User repository
	groupRepo         *database.GroupRepository      // Group repository
	departmentRepo    *database.DepartmentRepository // Department repository
	meetingRepo       *database.MeetingRepository    // Meeting repository
	liveKitRepo       *database.LiveKitRepository    // LiveKit repository
	egressRepo        *database.EgressRepository     // LiveKit Egress repository
	taskRepo          *database.TaskRepository       // Task repository
	mailer            *email.Mailer                  // Email service
	egressClient      *livekit.EgressClient          // LiveKit Egress client
	metricsData       []models.Metric                // In-memory metrics store
	logs              []models.LogEntry              // In-memory logs store
	prometheusMetrics *metrics.ServiceMetrics        // Prometheus metrics
	prometheusClient  *prometheus.Client             // Prometheus query client
	rabbitMQPublisher *rabbitmq.Publisher            // RabbitMQ publisher for transcription tasks
}

func NewManagingPortal(cfg *config.Config, log *logger.Logger) (*ManagingPortal, error) {
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
	groupRepo := database.NewGroupRepository(db)
	departmentRepo := database.NewDepartmentRepository(db)
	meetingRepo := database.NewMeetingRepository(db)
	liveKitRepo := database.NewLiveKitRepository(db)
	egressRepo := database.NewEgressRepository(db)
	taskRepo := database.NewTaskRepository(db.DB)

	// Initialize email mailer
	emailConfig := email.LoadConfigFromEnv()
	mailer := email.NewMailer(emailConfig)

	// Initialize LiveKit Egress client from environment variables
	egressClient := livekit.NewEgressClientFromEnv()

	// Initialize Prometheus client (uses internal Docker networ
	prometheusClient := prometheus.NewClient("http://prometheus:9090")

	// Initialize RabbitMQ publisher for transcription tasks
	rabbitMQHost := getEnv("RABBITMQ_HOST", "localhost")
	rabbitMQPort := getEnvInt("RABBITMQ_PORT", 5672)
	rabbitMQUser := getEnv("RABBITMQ_USER", "guest")
	rabbitMQPassword := getEnv("RABBITMQ_PASSWORD", "guest")
	rabbitMQQueue := getEnv("RABBITMQ_QUEUE", "transcription_queue")

	rabbitMQPublisher, err := rabbitmq.NewPublisher(
		rabbitMQHost,
		rabbitMQPort,
		rabbitMQUser,
		rabbitMQPassword,
		rabbitMQQueue,
	)
	if err != nil {
		log.Errorf("Failed to initialize RabbitMQ publisher: %v (transcription tasks will not be sent)", err)
		// Don't fail startup if RabbitMQ is not available
		rabbitMQPublisher = nil
	} else {
		log.Info("RabbitMQ publisher initialized successfully")
	}

	return &ManagingPortal{
		config:            cfg,
		logger:            log,
		services:          make(map[string]models.ServiceInfo),
		jwtManager:        jwtManager,
		db:                db,
		userRepo:          userRepo,
		groupRepo:         groupRepo,
		departmentRepo:    departmentRepo,
		meetingRepo:       meetingRepo,
		liveKitRepo:       liveKitRepo,
		egressRepo:        egressRepo,
		taskRepo:          taskRepo,
		mailer:            mailer,
		egressClient:      egressClient,
		metricsData:       make([]models.Metric, 0),
		logs:              make([]models.LogEntry, 0),
		prometheusMetrics: metrics.NewServiceMetrics("managing_portal"),
		prometheusClient:  prometheusClient,
		rabbitMQPublisher: rabbitMQPublisher,
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

	// Find user in database (username field contains email address)
	user, err := mp.userRepo.GetByEmail(req.Username)
	if err != nil || !auth.VerifyPassword(req.Password, user.Password) {
		mp.respondWithError(w, http.StatusUnauthorized, "Invalid credentials", "email or password incorrect")
		return
	}

	// Check if user is active
	if !user.IsActive {
		mp.respondWithError(w, http.StatusUnauthorized, "Account is inactive", "")
		return
	}

	// Update last login
	mp.userRepo.UpdateLastLogin(user.ID)

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
			ID:             user.ID,
			Username:       user.Username,
			Email:          user.Email,
			Role:           user.Role,
			OrganizationID: user.OrganizationID,
			DepartmentID:   user.DepartmentID,
			Permissions:    user.Permissions,
			Language:       user.Language,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Register godoc
// @Summary Регистрация нового пользователя
// @Description Зарегистрировать новую учетную запись пользователя (только администратор)
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

	// Check if username already exists
	if exists, _ := mp.userRepo.UsernameExists(req.Username); exists {
		mp.respondWithError(w, http.StatusConflict, "User already exists", "username is taken")
		return
	}

	// Check if email already exists
	if exists, _ := mp.userRepo.EmailExists(req.Email); exists {
		mp.respondWithError(w, http.StatusConflict, "Email already in use", "email is taken")
		return
	}

	// Create new user with default permissions
	language := req.Language
	if language == "" {
		language = "en" // Default language
	}

	newUser := &models.User{
		ID:       uuid.New(),
		Username: req.Username,
		Email:    req.Email,
		Password: auth.HashPassword(req.Password),
		Role:     models.RoleUser,
		Permissions: models.UserPermissions{
			CanScheduleMeetings:  false,
			CanManageDepartment:  false,
			CanApproveRecordings: false,
		},
		Language:  language,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := mp.userRepo.Create(newUser); err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to create user", err.Error())
		return
	}

	mp.logger.Infof("User created: %s (%s)", newUser.Username, newUser.ID)

	// Send welcome email asynchronously
	go func() {
		loginURL := getEnv("LOGIN_URL", "http://localhost:20080")
		emailData := email.WelcomeEmailData{
			Username: newUser.Username,
			Email:    newUser.Email,
			Password: req.Password, // Send original password before hashing
			Language: newUser.Language,
			LoginURL: loginURL,
		}

		if err := mp.mailer.SendWelcomeEmail(newUser.Email, emailData); err != nil {
			mp.logger.Errorf("Failed to send welcome email to %s: %v", newUser.Email, err)
		} else {
			mp.logger.Infof("Welcome email sent to %s", newUser.Email)
		}
	}()

	userInfo := models.UserInfo{
		ID:             newUser.ID,
		Username:       newUser.Username,
		Email:          newUser.Email,
		Role:           newUser.Role,
		OrganizationID: newUser.OrganizationID,
		DepartmentID:   newUser.DepartmentID,
		Permissions:    newUser.Permissions,
		Language:       newUser.Language,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(userInfo)
}

// Health godoc
// @Summary Проверка здоровья
// @Description Проверка, работает ли сервис
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
// @Summary Получить статус системы
// @Description Получить общий статус системы и здоровье сервисов
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
// @Summary Список всех сервисов
// @Description Получить список всех зарегистрированных сервисов
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
// @Summary Зарегистрировать сервис
// @Description Зарегистрировать новый сервис в управляющем портале
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
// @Summary Heartbeat сервиса
// @Description Обновить heartbeat сервиса, чтобы указать, что он все еще активен
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

	// Prometheus metrics endpoint (no auth required for scraping)
	mux.Handle("/metrics", promhttp.Handler())

	// Swagger documentation endpoint
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

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

	mux.Handle("/api/v1/groups/remove-user", chainMiddleware(
		http.HandlerFunc(mp.removeUserFromGroupHandler),
		authMiddleware,
		adminMiddleware,
	))

	mux.Handle("/api/v1/groups/check-permission", chainMiddleware(
		http.HandlerFunc(mp.checkPermissionHandler),
		authMiddleware,
	))

	// Meeting subject management endpoints (admin only)
	mux.Handle("/api/v1/meeting-subjects", chainMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				mp.listMeetingSubjectsHandler(w, r)
			} else if r.Method == http.MethodPost {
				mp.createMeetingSubjectHandler(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}),
		authMiddleware,
		adminMiddleware,
	))

	mux.Handle("/api/v1/meeting-subjects/", chainMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				mp.getMeetingSubjectHandler(w, r)
			case http.MethodPut:
				mp.updateMeetingSubjectHandler(w, r)
			case http.MethodDelete:
				mp.deleteMeetingSubjectHandler(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}),
		authMiddleware,
		adminMiddleware,
	))

	// Department management endpoints (admin only)
	mux.Handle("/api/v1/departments", chainMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				mp.listDepartmentsHandler(w, r)
			} else if r.Method == http.MethodPost {
				mp.createDepartmentHandler(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}),
		authMiddleware,
		adminMiddleware,
	))

	mux.Handle("/api/v1/departments/", chainMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path ends with /children
			if strings.HasSuffix(r.URL.Path, "/children") {
				if r.Method == http.MethodGet {
					mp.getDepartmentChildrenHandler(w, r)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
				return
			}

			// Handle department CRUD operations
			switch r.Method {
			case http.MethodGet:
				mp.getDepartmentHandler(w, r)
			case http.MethodPut:
				mp.updateDepartmentHandler(w, r)
			case http.MethodDelete:
				mp.deleteDepartmentHandler(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}),
		authMiddleware,
		adminMiddleware,
	))

	// Organization management endpoints (admin only)
	mux.Handle("/api/v1/organizations", chainMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				mp.GetOrganizationsHandler(w, r)
			} else if r.Method == http.MethodPost {
				mp.CreateOrganizationHandler(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}),
		authMiddleware,
		adminMiddleware,
	))

	mux.Handle("/api/v1/organizations/", chainMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if this is a stats request
			if strings.HasSuffix(r.URL.Path, "/stats") && r.Method == http.MethodGet {
				mp.GetOrganizationStatsHandler(w, r)
				return
			}

			// Handle single organization operations
			switch r.Method {
			case http.MethodGet:
				mp.GetOrganizationHandler(w, r)
			case http.MethodPut:
				mp.UpdateOrganizationHandler(w, r)
			case http.MethodDelete:
				mp.DeleteOrganizationHandler(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}),
		authMiddleware,
		adminMiddleware,
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

	// Dashboard endpoints
	mux.Handle("/api/v1/dashboard/stats", chainMiddleware(
		http.HandlerFunc(mp.dashboardStatsHandler),
		authMiddleware,
	))

	// LiveKit webhook endpoint (no auth required for LiveKit server)
	mux.HandleFunc("/webhook/meet", mp.liveKitWebhookHandler)

	// LiveKit management endpoints (authenticated)
	mux.Handle("/api/v1/livekit/rooms", chainMiddleware(
		http.HandlerFunc(mp.listLiveKitRoomsHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/livekit/rooms/", chainMiddleware(
		http.HandlerFunc(mp.getLiveKitRoomHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/livekit/participants", chainMiddleware(
		http.HandlerFunc(mp.listLiveKitParticipantsHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/livekit/tracks", chainMiddleware(
		http.HandlerFunc(mp.listLiveKitTracksHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/livekit/webhook-events", chainMiddleware(
		http.HandlerFunc(mp.listWebhookEventsHandler),
		authMiddleware,
	))

	// LiveKit Egress endpoints (authenticated)
	mux.Handle("/api/v1/livekit/egress/room/start", chainMiddleware(
		http.HandlerFunc(mp.startRoomRecordingHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/livekit/egress/track/start", chainMiddleware(
		http.HandlerFunc(mp.startTrackRecordingHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/livekit/egress/stop", chainMiddleware(
		http.HandlerFunc(mp.stopRoomRecordingHandler),
		authMiddleware,
	))

	mux.Handle("/api/v1/livekit/egress", chainMiddleware(
		http.HandlerFunc(mp.listEgressHandler),
		authMiddleware,
	))

	// Task management endpoints (authenticated)
	mux.Handle("/api/v1/sessions/tasks", chainMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				mp.listSessionTasksHandler(w, r)
			} else if r.Method == http.MethodPost {
				mp.createTaskHandler(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}),
		authMiddleware,
	))

	mux.Handle("/api/v1/sessions/tasks/", chainMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				mp.getTaskHandler(w, r)
			case http.MethodPut:
				mp.updateTaskHandler(w, r)
			case http.MethodDelete:
				mp.deleteTaskHandler(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}),
		authMiddleware,
	))

	// AI task extraction endpoint
	mux.Handle("/api/v1/sessions/extract-tasks", chainMiddleware(
		http.HandlerFunc(mp.extractTasksHandler),
		authMiddleware,
	))

	// Serve React frontend for all other routes
	mux.Handle("/", serveStaticFiles())

	return mux
}

func (mp *ManagingPortal) Start() error {
	mux := mp.setupRoutes()

	// Wrap with metrics middleware
	metricsMiddleware := metrics.HTTPMetricsMiddleware(mp.prometheusMetrics)
	handler := metricsMiddleware(mux)

	addr := fmt.Sprintf("%s:%d", mp.config.Server.Host, mp.config.Server.Port)
	mp.logger.Infof("Managing Portal starting on %s", addr)
	mp.logger.Infof("Version: %s", version)
	mp.logger.Infof("Swagger docs: http://%s/swagger/index.html", addr)
	mp.logger.Infof("Prometheus metrics: http://%s/metrics", addr)
	mp.logger.Infof("Default admin credentials: username=admin, password=admin123")

	// Start heartbeat checker in background
	go mp.checkServiceHeartbeats()

	return http.ListenAndServe(addr, handler)
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

	portal, err := NewManagingPortal(cfg, log)
	if err != nil {
		log.Fatalf("Failed to initialize portal: %v", err)
	}

	// Start transcription result consumer in background
	go func() {
		log.Info("Starting transcription result consumer...")
		StartTranscriptionConsumer()
	}()

	if err := portal.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

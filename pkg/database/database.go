package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type DB struct {
	*gorm.DB
}

// Exec executes a query without returning any rows (wrapper for compatibility)
func (db *DB) Exec(query string, args ...interface{}) (Result, error) {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return nil, err
	}
	return sqlDB.Exec(query, args...)
}

// Query executes a query that returns rows (wrapper for compatibility)
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return nil, err
	}
	return sqlDB.Query(query, args...)
}

// QueryRow executes a query that returns at most one row (wrapper for compatibility)
func (db *DB) QueryRow(query string, args ...interface{}) *Row {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return &Row{err: err}
	}
	return &Row{row: sqlDB.QueryRow(query, args...)}
}

// Result is an alias for sql.Result
type Result interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

// Rows is a wrapper for sql.Rows
type Rows struct {
	*sql.Rows
}

// Row is a wrapper for sql.Row
type Row struct {
	row *sql.Row
	err error
}

// Scan implements sql.Row.Scan
func (r *Row) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	return r.row.Scan(dest...)
}

// NewDB creates a new database connection using GORM
func NewDB(cfg Config) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	// Configure GORM logger
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB for connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

// RunMigrations executes all database migrations using GORM AutoMigrate
func (db *DB) RunMigrations() error {
	log.Println("🔄 Starting database migrations with GORM AutoMigrate...")

	// Enable pgvector extension
	log.Println("→ Enabling pgvector extension...")
	if err := db.DB.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
		return fmt.Errorf("failed to create vector extension: %w", err)
	}
	log.Println("✓ pgvector extension enabled")

	// Smart migration for LiveKit columns that need type changes
	log.Println("→ Checking LiveKit columns for type migrations...")

	// Migration 1: audio_features from text[] to jsonb
	var audioFeaturesType string
	err := db.DB.Raw(`
		SELECT data_type
		FROM information_schema.columns
		WHERE table_name = 'livekit_tracks'
		AND column_name = 'audio_features'
	`).Scan(&audioFeaturesType).Error

	if err == nil && audioFeaturesType == "ARRAY" {
		log.Println("  → Migrating audio_features from ARRAY to jsonb...")

		// Drop the column and let AutoMigrate recreate it
		if err := db.DB.Exec(`ALTER TABLE livekit_tracks DROP COLUMN audio_features`).Error; err != nil {
			log.Printf("  → Warning: Could not drop audio_features column: %v", err)
		} else {
			log.Println("  → audio_features column dropped, will be recreated")
		}
	} else if err == nil {
		log.Printf("  → audio_features is %s, no migration needed", audioFeaturesType)
	}

	// Migration 2: payload from jsonb to bytea (or vice versa)
	var payloadType string
	err = db.DB.Raw(`
		SELECT data_type
		FROM information_schema.columns
		WHERE table_name = 'livekit_webhook_events'
		AND column_name = 'payload'
	`).Scan(&payloadType).Error

	if err == nil && payloadType != "jsonb" {
		log.Printf("  → Migrating payload from %s to jsonb...", payloadType)

		// Drop the column and let AutoMigrate recreate it
		if err := db.DB.Exec(`ALTER TABLE livekit_webhook_events DROP COLUMN payload`).Error; err != nil {
			log.Printf("  → Warning: Could not drop payload column: %v", err)
		} else {
			log.Println("  → payload column dropped, will be recreated")
		}
	} else if err == nil {
		log.Println("  → payload is already jsonb, no migration needed")
	}

	log.Println("✓ Column migrations completed")

	// Run AutoMigrate for all models
	log.Println("→ Running AutoMigrate for all models...")

	dbModels := []interface{}{
		&Department{},
		&User{},
		&Group{},
		&GroupMembership{},
		&UploadedFile{},
		&FileTranscription{},
		&DocumentChunk{},
		&models.Room{},
		&models.Participant{},
		&models.Track{},
		&models.WebhookEventLog{},
		&models.EgressRecording{},
		&LiveKitEgress{},
		&MeetingSubject{},
		&Meeting{},
		&MeetingParticipant{},
		&MeetingDepartment{},
		&TemporaryUser{},
		&PasswordResetToken{},
	}

	if err := db.AutoMigrate(dbModels...); err != nil {
		return fmt.Errorf("failed to run auto migrations: %w", err)
	}
	log.Println("✓ AutoMigrate completed successfully")

	// Create indexes that GORM doesn't create automatically
	log.Println("→ Creating additional indexes...")
	if err := db.createAdditionalIndexes(); err != nil {
		return fmt.Errorf("failed to create additional indexes: %w", err)
	}
	log.Println("✓ Additional indexes created")

	// Insert default data
	log.Println("→ Inserting default data...")
	if err := db.insertDefaultData(); err != nil {
		return fmt.Errorf("failed to insert default data: %w", err)
	}
	log.Println("✓ Default data inserted")

	log.Println("✅ All migrations completed successfully!")
	return nil
}

// createAdditionalIndexes creates indexes that aren't automatically created by GORM
func (db *DB) createAdditionalIndexes() error {
	indexes := []string{
		// Users indexes
		"CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)",
		"CREATE INDEX IF NOT EXISTS idx_users_department_id ON users(department_id)",

		// Departments indexes
		"CREATE INDEX IF NOT EXISTS idx_departments_parent_id ON departments(parent_id)",
		"CREATE INDEX IF NOT EXISTS idx_departments_path ON departments(path)",
		"CREATE INDEX IF NOT EXISTS idx_departments_is_active ON departments(is_active)",
		"CREATE INDEX IF NOT EXISTS idx_departments_name ON departments(name)",

		// Uploaded files indexes
		"CREATE INDEX IF NOT EXISTS idx_uploaded_files_user ON uploaded_files(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_uploaded_files_group ON uploaded_files(group_id)",
		"CREATE INDEX IF NOT EXISTS idx_uploaded_files_status ON uploaded_files(status)",
		"CREATE INDEX IF NOT EXISTS idx_uploaded_files_uploaded_at ON uploaded_files(uploaded_at DESC)",

		// File transcriptions indexes
		"CREATE INDEX IF NOT EXISTS idx_file_transcriptions_file ON file_transcriptions(file_id)",
		"CREATE INDEX IF NOT EXISTS idx_file_transcriptions_language ON file_transcriptions(language)",

		// Document chunks indexes
		"CREATE INDEX IF NOT EXISTS idx_document_chunks_file ON document_chunks(file_id)",
		"CREATE INDEX IF NOT EXISTS idx_document_chunks_transcription ON document_chunks(transcription_id)",
		"CREATE INDEX IF NOT EXISTS idx_document_chunks_embedding ON document_chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)",

		// LiveKit rooms indexes
		"CREATE INDEX IF NOT EXISTS idx_livekit_rooms_name ON livekit_rooms(name)",
		"CREATE INDEX IF NOT EXISTS idx_livekit_rooms_status ON livekit_rooms(status)",
		"CREATE INDEX IF NOT EXISTS idx_livekit_rooms_started_at ON livekit_rooms(started_at DESC)",

		// LiveKit participants indexes
		"CREATE INDEX IF NOT EXISTS idx_livekit_participants_room_sid ON livekit_participants(room_sid)",
		"CREATE INDEX IF NOT EXISTS idx_livekit_participants_identity ON livekit_participants(identity)",
		"CREATE INDEX IF NOT EXISTS idx_livekit_participants_state ON livekit_participants(state)",

		// LiveKit tracks indexes
		"CREATE INDEX IF NOT EXISTS idx_livekit_tracks_participant_sid ON livekit_tracks(participant_sid)",
		"CREATE INDEX IF NOT EXISTS idx_livekit_tracks_room_sid ON livekit_tracks(room_sid)",
		"CREATE INDEX IF NOT EXISTS idx_livekit_tracks_type ON livekit_tracks(type)",
		"CREATE INDEX IF NOT EXISTS idx_livekit_tracks_source ON livekit_tracks(source)",
		"CREATE INDEX IF NOT EXISTS idx_livekit_tracks_status ON livekit_tracks(status)",

		// LiveKit webhook events indexes
		"CREATE INDEX IF NOT EXISTS idx_livekit_webhook_events_event_type ON livekit_webhook_events(event_type)",
		"CREATE INDEX IF NOT EXISTS idx_livekit_webhook_events_event_id ON livekit_webhook_events(event_id)",
		"CREATE INDEX IF NOT EXISTS idx_livekit_webhook_events_room_sid ON livekit_webhook_events(room_sid)",
		"CREATE INDEX IF NOT EXISTS idx_livekit_webhook_events_participant_sid ON livekit_webhook_events(participant_sid)",
		"CREATE INDEX IF NOT EXISTS idx_livekit_webhook_events_created_at ON livekit_webhook_events(created_at DESC)",

		// Meeting subjects indexes
		"CREATE INDEX IF NOT EXISTS idx_meeting_subjects_name ON meeting_subjects(name)",
		"CREATE INDEX IF NOT EXISTS idx_meeting_subjects_is_active ON meeting_subjects(is_active)",
		"CREATE INDEX IF NOT EXISTS idx_meeting_subjects_department_ids ON meeting_subjects USING GIN(department_ids)",

		// Meetings indexes
		"CREATE INDEX IF NOT EXISTS idx_meetings_scheduled_at ON meetings(scheduled_at)",
		"CREATE INDEX IF NOT EXISTS idx_meetings_status ON meetings(status)",
		"CREATE INDEX IF NOT EXISTS idx_meetings_type ON meetings(type)",
		"CREATE INDEX IF NOT EXISTS idx_meetings_subject_id ON meetings(subject_id)",
		"CREATE INDEX IF NOT EXISTS idx_meetings_created_by ON meetings(created_by)",
		"CREATE INDEX IF NOT EXISTS idx_meetings_livekit_room_id ON meetings(livekit_room_id)",

		// Meeting participants indexes
		"CREATE INDEX IF NOT EXISTS idx_meeting_participants_meeting_id ON meeting_participants(meeting_id)",
		"CREATE INDEX IF NOT EXISTS idx_meeting_participants_user_id ON meeting_participants(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_meeting_participants_role ON meeting_participants(role)",
		"CREATE INDEX IF NOT EXISTS idx_meeting_participants_status ON meeting_participants(status)",

		// Meeting departments indexes
		"CREATE INDEX IF NOT EXISTS idx_meeting_departments_meeting_id ON meeting_departments(meeting_id)",
		"CREATE INDEX IF NOT EXISTS idx_meeting_departments_department_id ON meeting_departments(department_id)",

		// Password reset tokens indexes
		"CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_code ON password_reset_tokens(code)",
		"CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_email ON password_reset_tokens(email)",
		"CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at)",
		"CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens(user_id)",
	}

	for _, indexSQL := range indexes {
		if err := db.DB.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
			// Continue with other indexes even if one fails
		}
	}

	return nil
}

// insertDefaultData inserts default users, groups, and departments
func (db *DB) insertDefaultData() error {
	// Get admin credentials from environment
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@recontext.online"
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123"
	}

	userPassword := "user123"

	// Hash passwords
	hashedAdminPassword := auth.HashPassword(adminPassword)
	hashedUserPassword := auth.HashPassword(userPassword)

	// Insert default root department (UUID will be auto-generated)
	dept := Department{
		Name:        "Organization",
		Description: "Root department",
		Level:       0,
		Path:        "Organization",
		IsActive:    true,
	}
	db.Where(Department{Name: dept.Name}).FirstOrCreate(&dept)

	// Insert default admin user (UUID will be auto-generated)
	adminUser := User{
		Username:    "admin",
		Email:       adminEmail,
		Password:    hashedAdminPassword,
		Role:        "admin",
		IsActive:    true,
		Language:    "en",
		Permissions: `{"can_schedule_meetings": true, "can_manage_department": true, "can_approve_recordings": true}`,
	}
	db.Where(User{Username: adminUser.Username}).Assign(&adminUser).FirstOrCreate(&adminUser)

	// Insert default regular user (UUID will be auto-generated)
	regularUser := User{
		Username:    "user",
		Email:       "user@recontext.online",
		Password:    hashedUserPassword,
		Role:        "user",
		IsActive:    true,
		Language:    "en",
		Permissions: `{"can_schedule_meetings": false, "can_manage_department": false, "can_approve_recordings": false}`,
	}
	db.Where(User{Username: regularUser.Username}).FirstOrCreate(&regularUser)

	// Insert default groups (UUID will be auto-generated)
	groups := []Group{
		{
			Name:        "Editors",
			Description: "Users who can view and edit recordings",
			Permissions: `{"recordings": {"actions": ["read", "write"], "scope": "all"}, "transcripts": {"actions": ["read"], "scope": "all"}}`,
		},
		{
			Name:        "Viewers",
			Description: "Users who can only view recordings",
			Permissions: `{"recordings": {"actions": ["read"], "scope": "all"}, "transcripts": {"actions": ["read"], "scope": "all"}}`,
		},
		{
			Name:        "File Uploaders",
			Description: "Users who can upload and transcribe files",
			Permissions: `{"files": {"actions": ["read", "write", "delete"], "scope": "own"}, "transcriptions": {"actions": ["read"], "scope": "own"}}`,
		},
		{
			Name:        "RAG Users",
			Description: "Users who can use RAG semantic search on transcriptions",
			Permissions: `{"rag": {"actions": ["read", "search"], "scope": "all"}, "transcriptions": {"actions": ["read"], "scope": "all"}}`,
		},
	}

	for _, group := range groups {
		db.Where(Group{Name: group.Name}).FirstOrCreate(&group)
	}

	log.Println("✓ Default data created/verified")
	return nil
}

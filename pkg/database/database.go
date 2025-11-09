package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"Recontext.online/pkg/auth"
	_ "github.com/lib/pq"
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
	*sql.DB
}

// NewDB creates a new database connection
func NewDB(cfg Config) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping to verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

// RunMigrations executes all database migrations
func (db *DB) RunMigrations() error {
	// Enable pgvector extension
	if _, err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector"); err != nil {
		return fmt.Errorf("failed to create vector extension: %w", err)
	}

	migrations := []string{
		createDepartmentsTable,
		createUsersTable,
		createGroupsTable,
		createGroupMembershipsTable,
		createUploadedFilesTable,
		createFileTranscriptionsTable,
		createDocumentChunksTable,
		createLiveKitRoomsTable,
		createLiveKitParticipantsTable,
		createLiveKitTracksTable,
		createLiveKitWebhookEventsTable,
		addDepartmentToUsers,
		addPermissionsToUsers,
		createMeetingSubjectsTable,
		createMeetingsTable,
		createMeetingParticipantsTable,
		createMeetingDepartmentsTable,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	// Insert default data with environment-based admin credentials
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@recontext.online"
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123"
	}

	// Hash the admin password
	hashedAdminPassword := auth.HashPassword(adminPassword)

	// Insert default data with dynamic admin credentials
	insertDefaultDataSQL := fmt.Sprintf(`
-- Insert default admin user
INSERT INTO users (id, username, email, password, role, is_active, created_at, updated_at)
VALUES (
	'admin-001',
	'admin',
	'%s',
	'%s',
	'admin',
	true,
	NOW(),
	NOW()
) ON CONFLICT (username) DO UPDATE SET
	email = EXCLUDED.email,
	password = EXCLUDED.password,
	updated_at = NOW();

-- Insert default user (password: user123)
INSERT INTO users (id, username, email, password, role, is_active, created_at, updated_at)
VALUES (
	'user-001',
	'user',
	'user@recontext.online',
	'$2a$10$ZK5z.qH.BvR5dqT3BqKqZ.KZ.1HqJ5J5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Zu',
	'user',
	true,
	NOW(),
	NOW()
) ON CONFLICT (username) DO NOTHING;

-- Insert default groups
INSERT INTO groups (id, name, description, permissions, created_at, updated_at)
VALUES (
	'group-editors',
	'Editors',
	'Users who can view and edit recordings',
	'{"recordings": {"actions": ["read", "write"], "scope": "all"}, "transcripts": {"actions": ["read"], "scope": "all"}}',
	NOW(),
	NOW()
) ON CONFLICT (name) DO NOTHING;

INSERT INTO groups (id, name, description, permissions, created_at, updated_at)
VALUES (
	'group-viewers',
	'Viewers',
	'Users who can only view recordings',
	'{"recordings": {"actions": ["read"], "scope": "all"}, "transcripts": {"actions": ["read"], "scope": "all"}}',
	NOW(),
	NOW()
) ON CONFLICT (name) DO NOTHING;

INSERT INTO groups (id, name, description, permissions, created_at, updated_at)
VALUES (
	'group-file-uploaders',
	'File Uploaders',
	'Users who can upload and transcribe files',
	'{"files": {"actions": ["read", "write", "delete"], "scope": "own"}, "transcriptions": {"actions": ["read"], "scope": "own"}}',
	NOW(),
	NOW()
) ON CONFLICT (name) DO NOTHING;

INSERT INTO groups (id, name, description, permissions, created_at, updated_at)
VALUES (
	'group-rag-users',
	'RAG Users',
	'Users who can use RAG semantic search on transcriptions',
	'{"rag": {"actions": ["read", "search"], "scope": "all"}, "transcriptions": {"actions": ["read"], "scope": "all"}}',
	NOW(),
	NOW()
) ON CONFLICT (name) DO NOTHING;
`, adminEmail, hashedAdminPassword)

	if _, err := db.Exec(insertDefaultDataSQL); err != nil {
		return fmt.Errorf("failed to insert default data: %w", err)
	}

	return nil
}

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
	id VARCHAR(255) PRIMARY KEY,
	username VARCHAR(255) UNIQUE NOT NULL,
	email VARCHAR(255) UNIQUE NOT NULL,
	password VARCHAR(255) NOT NULL,
	role VARCHAR(50) NOT NULL DEFAULT 'user',
	groups TEXT[] DEFAULT '{}',
	is_active BOOLEAN DEFAULT true,
	last_login TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
`

const createGroupsTable = `
CREATE TABLE IF NOT EXISTS groups (
	id VARCHAR(255) PRIMARY KEY,
	name VARCHAR(255) UNIQUE NOT NULL,
	description TEXT,
	permissions JSONB NOT NULL DEFAULT '{}',
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_groups_name ON groups(name);
`

const createGroupMembershipsTable = `
CREATE TABLE IF NOT EXISTS group_memberships (
	user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	group_id VARCHAR(255) NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
	added_at TIMESTAMP NOT NULL DEFAULT NOW(),
	added_by VARCHAR(255),
	PRIMARY KEY (user_id, group_id)
);

CREATE INDEX IF NOT EXISTS idx_group_memberships_user ON group_memberships(user_id);
CREATE INDEX IF NOT EXISTS idx_group_memberships_group ON group_memberships(group_id);
`

const createUploadedFilesTable = `
CREATE TABLE IF NOT EXISTS uploaded_files (
	id VARCHAR(255) PRIMARY KEY,
	filename VARCHAR(500) NOT NULL,
	original_name VARCHAR(500) NOT NULL,
	file_size BIGINT NOT NULL,
	mime_type VARCHAR(255) NOT NULL,
	storage_path TEXT NOT NULL,
	user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	group_id VARCHAR(255) NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
	status VARCHAR(50) NOT NULL DEFAULT 'pending',
	transcription_id VARCHAR(255),
	metadata JSONB DEFAULT '{}',
	uploaded_at TIMESTAMP NOT NULL DEFAULT NOW(),
	processed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_uploaded_files_user ON uploaded_files(user_id);
CREATE INDEX IF NOT EXISTS idx_uploaded_files_group ON uploaded_files(group_id);
CREATE INDEX IF NOT EXISTS idx_uploaded_files_status ON uploaded_files(status);
CREATE INDEX IF NOT EXISTS idx_uploaded_files_uploaded_at ON uploaded_files(uploaded_at DESC);
`

const createFileTranscriptionsTable = `
CREATE TABLE IF NOT EXISTS file_transcriptions (
	id VARCHAR(255) PRIMARY KEY,
	file_id VARCHAR(255) NOT NULL REFERENCES uploaded_files(id) ON DELETE CASCADE,
	text TEXT NOT NULL,
	language VARCHAR(10) NOT NULL,
	confidence DECIMAL(5,4),
	duration DECIMAL(10,2),
	segments JSONB DEFAULT '{}',
	transcribed_at TIMESTAMP NOT NULL DEFAULT NOW(),
	transcribed_by VARCHAR(255) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_file_transcriptions_file ON file_transcriptions(file_id);
CREATE INDEX IF NOT EXISTS idx_file_transcriptions_language ON file_transcriptions(language);
`

const createDocumentChunksTable = `
CREATE TABLE IF NOT EXISTS document_chunks (
	id VARCHAR(255) PRIMARY KEY,
	file_id VARCHAR(255) NOT NULL REFERENCES uploaded_files(id) ON DELETE CASCADE,
	transcription_id VARCHAR(255) REFERENCES file_transcriptions(id) ON DELETE CASCADE,
	chunk_text TEXT NOT NULL,
	chunk_index INTEGER NOT NULL,
	embedding vector(1536),
	metadata JSONB DEFAULT '{}',
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_document_chunks_file ON document_chunks(file_id);
CREATE INDEX IF NOT EXISTS idx_document_chunks_transcription ON document_chunks(transcription_id);
CREATE INDEX IF NOT EXISTS idx_document_chunks_embedding ON document_chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
`

const createLiveKitRoomsTable = `
CREATE TABLE IF NOT EXISTS livekit_rooms (
	id VARCHAR(255) PRIMARY KEY,
	sid VARCHAR(255) UNIQUE NOT NULL,
	name VARCHAR(255) NOT NULL,
	empty_timeout INTEGER DEFAULT 300,
	departure_timeout INTEGER DEFAULT 20,
	creation_time VARCHAR(50),
	creation_time_ms VARCHAR(50),
	turn_password TEXT,
	enabled_codecs JSONB DEFAULT '[]',
	status VARCHAR(50) NOT NULL DEFAULT 'active',
	started_at TIMESTAMP NOT NULL,
	finished_at TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_livekit_rooms_sid ON livekit_rooms(sid);
CREATE INDEX IF NOT EXISTS idx_livekit_rooms_name ON livekit_rooms(name);
CREATE INDEX IF NOT EXISTS idx_livekit_rooms_status ON livekit_rooms(status);
CREATE INDEX IF NOT EXISTS idx_livekit_rooms_started_at ON livekit_rooms(started_at DESC);
`

const createLiveKitParticipantsTable = `
CREATE TABLE IF NOT EXISTS livekit_participants (
	id VARCHAR(255) PRIMARY KEY,
	sid VARCHAR(255) UNIQUE NOT NULL,
	room_sid VARCHAR(255) NOT NULL REFERENCES livekit_rooms(sid) ON DELETE CASCADE,
	identity VARCHAR(255) NOT NULL,
	name VARCHAR(255) NOT NULL,
	state VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
	joined_at VARCHAR(50) NOT NULL,
	joined_at_ms VARCHAR(50) NOT NULL,
	version INTEGER DEFAULT 0,
	permission JSONB DEFAULT '{}',
	is_publisher BOOLEAN DEFAULT false,
	disconnect_reason VARCHAR(255),
	left_at TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_livekit_participants_sid ON livekit_participants(sid);
CREATE INDEX IF NOT EXISTS idx_livekit_participants_room_sid ON livekit_participants(room_sid);
CREATE INDEX IF NOT EXISTS idx_livekit_participants_identity ON livekit_participants(identity);
CREATE INDEX IF NOT EXISTS idx_livekit_participants_state ON livekit_participants(state);
`

const createLiveKitTracksTable = `
CREATE TABLE IF NOT EXISTS livekit_tracks (
	id VARCHAR(255) PRIMARY KEY,
	sid VARCHAR(255) UNIQUE NOT NULL,
	participant_sid VARCHAR(255) NOT NULL REFERENCES livekit_participants(sid) ON DELETE CASCADE,
	room_sid VARCHAR(255) NOT NULL REFERENCES livekit_rooms(sid) ON DELETE CASCADE,
	type VARCHAR(50),
	source VARCHAR(50) NOT NULL,
	mime_type VARCHAR(100) NOT NULL,
	mid VARCHAR(50),
	width INTEGER,
	height INTEGER,
	simulcast BOOLEAN DEFAULT false,
	layers JSONB DEFAULT '[]',
	codecs JSONB DEFAULT '[]',
	stream VARCHAR(255),
	version JSONB DEFAULT '{}',
	audio_features TEXT[] DEFAULT '{}',
	backup_codec_policy VARCHAR(50),
	status VARCHAR(50) NOT NULL DEFAULT 'published',
	published_at TIMESTAMP NOT NULL,
	unpublished_at TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_livekit_tracks_sid ON livekit_tracks(sid);
CREATE INDEX IF NOT EXISTS idx_livekit_tracks_participant_sid ON livekit_tracks(participant_sid);
CREATE INDEX IF NOT EXISTS idx_livekit_tracks_room_sid ON livekit_tracks(room_sid);
CREATE INDEX IF NOT EXISTS idx_livekit_tracks_type ON livekit_tracks(type);
CREATE INDEX IF NOT EXISTS idx_livekit_tracks_source ON livekit_tracks(source);
CREATE INDEX IF NOT EXISTS idx_livekit_tracks_status ON livekit_tracks(status);
`

const createLiveKitWebhookEventsTable = `
CREATE TABLE IF NOT EXISTS livekit_webhook_events (
	id VARCHAR(255) PRIMARY KEY,
	event_type VARCHAR(100) NOT NULL,
	event_id VARCHAR(255) NOT NULL,
	room_sid VARCHAR(255),
	participant_sid VARCHAR(255),
	track_sid VARCHAR(255),
	payload JSONB NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_livekit_webhook_events_event_type ON livekit_webhook_events(event_type);
CREATE INDEX IF NOT EXISTS idx_livekit_webhook_events_event_id ON livekit_webhook_events(event_id);
CREATE INDEX IF NOT EXISTS idx_livekit_webhook_events_room_sid ON livekit_webhook_events(room_sid);
CREATE INDEX IF NOT EXISTS idx_livekit_webhook_events_participant_sid ON livekit_webhook_events(participant_sid);
CREATE INDEX IF NOT EXISTS idx_livekit_webhook_events_created_at ON livekit_webhook_events(created_at DESC);
`

// Department-related migrations
const createDepartmentsTable = `
CREATE TABLE IF NOT EXISTS departments (
	id VARCHAR(255) PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	description TEXT,
	parent_id VARCHAR(255) REFERENCES departments(id) ON DELETE SET NULL,
	level INTEGER NOT NULL DEFAULT 0,
	path TEXT NOT NULL,
	is_active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_departments_parent_id ON departments(parent_id);
CREATE INDEX IF NOT EXISTS idx_departments_path ON departments(path);
CREATE INDEX IF NOT EXISTS idx_departments_is_active ON departments(is_active);
CREATE INDEX IF NOT EXISTS idx_departments_name ON departments(name);

-- Insert default root department
INSERT INTO departments (id, name, description, parent_id, level, path, is_active, created_at, updated_at)
VALUES (
	'dept-root',
	'Organization',
	'Root department',
	NULL,
	0,
	'Organization',
	true,
	NOW(),
	NOW()
) ON CONFLICT (id) DO NOTHING;
`

const addDepartmentToUsers = `
-- Add department_id column to users table
DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_name = 'users' AND column_name = 'department_id'
	) THEN
		ALTER TABLE users ADD COLUMN department_id VARCHAR(255) REFERENCES departments(id) ON DELETE SET NULL;
		CREATE INDEX IF NOT EXISTS idx_users_department_id ON users(department_id);
	END IF;
END $$;
`

const addPermissionsToUsers = `
-- Add permissions column to users table
DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_name = 'users' AND column_name = 'permissions'
	) THEN
		ALTER TABLE users ADD COLUMN permissions JSONB NOT NULL DEFAULT '{"can_schedule_meetings": false, "can_manage_department": false, "can_approve_recordings": false}';

		-- Update admin users to have all permissions
		UPDATE users SET permissions = '{"can_schedule_meetings": true, "can_manage_department": true, "can_approve_recordings": true}' WHERE role = 'admin';
	END IF;
END $$;
`

// Meeting-related migrations
const createMeetingSubjectsTable = `
CREATE TABLE IF NOT EXISTS meeting_subjects (
	id VARCHAR(255) PRIMARY KEY,
	name VARCHAR(255) NOT NULL UNIQUE,
	description TEXT,
	department_ids TEXT[] DEFAULT '{}',
	is_active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_meeting_subjects_name ON meeting_subjects(name);
CREATE INDEX IF NOT EXISTS idx_meeting_subjects_is_active ON meeting_subjects(is_active);
CREATE INDEX IF NOT EXISTS idx_meeting_subjects_department_ids ON meeting_subjects USING GIN(department_ids);
`

const createMeetingsTable = `
CREATE TABLE IF NOT EXISTS meetings (
	id VARCHAR(255) PRIMARY KEY,
	title VARCHAR(500) NOT NULL,
	scheduled_at TIMESTAMP NOT NULL,
	duration INTEGER NOT NULL,
	recurrence VARCHAR(50) NOT NULL DEFAULT 'none',
	type VARCHAR(50) NOT NULL,
	subject_id VARCHAR(255) NOT NULL REFERENCES meeting_subjects(id) ON DELETE RESTRICT,
	status VARCHAR(50) NOT NULL DEFAULT 'scheduled',
	needs_video_record BOOLEAN NOT NULL DEFAULT false,
	needs_audio_record BOOLEAN NOT NULL DEFAULT false,
	additional_notes TEXT,
	livekit_room_id VARCHAR(255) REFERENCES livekit_rooms(sid) ON DELETE SET NULL,
	created_by VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_meetings_scheduled_at ON meetings(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_meetings_status ON meetings(status);
CREATE INDEX IF NOT EXISTS idx_meetings_type ON meetings(type);
CREATE INDEX IF NOT EXISTS idx_meetings_subject_id ON meetings(subject_id);
CREATE INDEX IF NOT EXISTS idx_meetings_created_by ON meetings(created_by);
CREATE INDEX IF NOT EXISTS idx_meetings_livekit_room_id ON meetings(livekit_room_id);
`

const createMeetingParticipantsTable = `
CREATE TABLE IF NOT EXISTS meeting_participants (
	id VARCHAR(255) PRIMARY KEY,
	meeting_id VARCHAR(255) NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
	user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	role VARCHAR(50) NOT NULL,
	status VARCHAR(50) NOT NULL DEFAULT 'invited',
	joined_at TIMESTAMP,
	left_at TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	UNIQUE(meeting_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_meeting_participants_meeting_id ON meeting_participants(meeting_id);
CREATE INDEX IF NOT EXISTS idx_meeting_participants_user_id ON meeting_participants(user_id);
CREATE INDEX IF NOT EXISTS idx_meeting_participants_role ON meeting_participants(role);
CREATE INDEX IF NOT EXISTS idx_meeting_participants_status ON meeting_participants(status);
`

const createMeetingDepartmentsTable = `
CREATE TABLE IF NOT EXISTS meeting_departments (
	id VARCHAR(255) PRIMARY KEY,
	meeting_id VARCHAR(255) NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
	department_id VARCHAR(255) NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	UNIQUE(meeting_id, department_id)
);

CREATE INDEX IF NOT EXISTS idx_meeting_departments_meeting_id ON meeting_departments(meeting_id);
CREATE INDEX IF NOT EXISTS idx_meeting_departments_department_id ON meeting_departments(department_id);
`

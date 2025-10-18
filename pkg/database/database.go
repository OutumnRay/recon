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
		createUsersTable,
		createGroupsTable,
		createGroupMembershipsTable,
		createUploadedFilesTable,
		createFileTranscriptionsTable,
		createDocumentChunksTable,
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

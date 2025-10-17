package database

import (
	"database/sql"
	"fmt"
	"time"

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
	migrations := []string{
		createUsersTable,
		createGroupsTable,
		createGroupMembershipsTable,
		insertDefaultData,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
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

const insertDefaultData = `
-- Insert default admin user (password: admin123)
INSERT INTO users (id, username, email, password, role, is_active, created_at, updated_at)
VALUES (
	'admin-001',
	'admin',
	'admin@recontext.online',
	'$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
	'admin',
	true,
	NOW(),
	NOW()
) ON CONFLICT (username) DO NOTHING;

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
`

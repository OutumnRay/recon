package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	RabbitMQ RabbitMQConfig
	MinIO    MinIOConfig
	Jitsi    JitsiConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host string
	Port int
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// RabbitMQConfig holds RabbitMQ connection configuration
type RabbitMQConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

// MinIOConfig holds MinIO/S3 configuration
type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

// JitsiConfig holds Jitsi Meet server configuration
type JitsiConfig struct {
	ServerURL string
	APIKey    string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnvAsInt("SERVER_PORT", 8080),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "postgres"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "recontext"),
			Password: getEnv("DB_PASSWORD", "recontext"),
			Database: getEnv("DB_NAME", "recontext"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		RabbitMQ: RabbitMQConfig{
			Host:     getEnv("RABBITMQ_HOST", "rabbitmq"),
			Port:     getEnvAsInt("RABBITMQ_PORT", 5672),
			User:     getEnv("RABBITMQ_USER", "guest"),
			Password: getEnv("RABBITMQ_PASSWORD", "guest"),
		},
		MinIO: MinIOConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "minio:9000"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			UseSSL:          getEnvAsBool("MINIO_USE_SSL", false),
			BucketName:      getEnv("MINIO_BUCKET", "recontext"),
		},
		Jitsi: JitsiConfig{
			ServerURL: getEnv("JITSI_SERVER_URL", "https://meet.jit.si"),
			APIKey:    getEnv("JITSI_API_KEY", ""),
		},
	}

	return cfg, nil
}

// GetDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// GetRabbitMQURL returns the RabbitMQ connection URL
func (c *RabbitMQConfig) GetURL() string {
	return fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		c.User, c.Password, c.Host, c.Port,
	)
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

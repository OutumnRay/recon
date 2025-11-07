package models

import "time"

// ServiceStatus represents the status of a service
type ServiceStatus string

const (
	ServiceStatusRunning ServiceStatus = "running"
	ServiceStatusStopped ServiceStatus = "stopped"
	ServiceStatusError   ServiceStatus = "error"
	ServiceStatusUnknown ServiceStatus = "unknown"
)

// ServiceType represents the type of service
type ServiceType string

const (
	ServiceTypeManagingPortal      ServiceType = "managing-portal"
	ServiceTypeUserPortal          ServiceType = "user-portal"
	ServiceTypeTranscriptionWorker ServiceType = "transcription-worker"
	ServiceTypeSummarizationWorker ServiceType = "summarization-worker"
	ServiceTypeJitsiAgent          ServiceType = "jitsi-agent"
)

// ServiceInfo represents information about a service
type ServiceInfo struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Type        ServiceType   `json:"type"`
	Status      ServiceStatus `json:"status"`
	Version     string        `json:"version"`
	StartedAt   time.Time     `json:"started_at"`
	LastHeartbeat time.Time   `json:"last_heartbeat"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SystemStatus represents the overall system status
type SystemStatus struct {
	Status       string                  `json:"status"`
	Services     map[string]ServiceInfo  `json:"services"`
	Infrastructure map[string]interface{} `json:"infrastructure"`
	Timestamp    time.Time               `json:"timestamp"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string    `json:"error"`
	Message string    `json:"message"`
	Code    int       `json:"code"`
	Timestamp time.Time `json:"timestamp"`
}

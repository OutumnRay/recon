package models

import "time"

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// Metric represents a telemetry metric
type Metric struct {
	ID         string                 `json:"id" db:"id"`
	ServiceID  string                 `json:"service_id" db:"service_id"`
	Name       string                 `json:"name" db:"name"`
	Type       MetricType             `json:"type" db:"type"`
	Value      float64                `json:"value" db:"value"`
	Unit       string                 `json:"unit,omitempty" db:"unit"`
	Labels     map[string]string      `json:"labels,omitempty" db:"labels"`
	Metadata   map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	Timestamp  time.Time              `json:"timestamp" db:"timestamp"`
}

// MetricsBatch represents a batch of metrics sent by a service
type MetricsBatch struct {
	ServiceID string    `json:"service_id" binding:"required" example:"service-001"`
	Metrics   []Metric  `json:"metrics" binding:"required"`
	Timestamp time.Time `json:"timestamp"`
}

// SendMetricsRequest represents a request to send metrics
type SendMetricsRequest struct {
	ServiceID string   `json:"service_id" binding:"required" example:"transcription-worker-1"`
	Metrics   []Metric `json:"metrics" binding:"required"`
}

// MetricsQueryRequest represents a request to query metrics
type MetricsQueryRequest struct {
	ServiceID string    `json:"service_id,omitempty" form:"service_id" example:"transcription-worker-1"`
	Name      string    `json:"name,omitempty" form:"name" example:"transcription_duration"`
	From      time.Time `json:"from,omitempty" form:"from"`
	To        time.Time `json:"to,omitempty" form:"to"`
	Limit     int       `json:"limit,omitempty" form:"limit" example:"100"`
}

// MetricsQueryResponse represents metrics query results
type MetricsQueryResponse struct {
	Metrics []Metric `json:"metrics"`
	Total   int      `json:"total"`
}

// ServiceMetricsSummary represents aggregated metrics for a service
type ServiceMetricsSummary struct {
	ServiceID        string            `json:"service_id"`
	TotalRequests    int64             `json:"total_requests"`
	FailedRequests   int64             `json:"failed_requests"`
	AverageLatency   float64           `json:"average_latency_ms"`
	LastReportedAt   time.Time         `json:"last_reported_at"`
	CustomMetrics    map[string]float64 `json:"custom_metrics,omitempty"`
}

// SystemMetrics represents overall system metrics
type SystemMetrics struct {
	ActiveServices    int                              `json:"active_services"`
	TotalRequests     int64                            `json:"total_requests"`
	TotalErrors       int64                            `json:"total_errors"`
	AverageLatency    float64                          `json:"average_latency_ms"`
	ServicesSummaries map[string]ServiceMetricsSummary `json:"services_summaries"`
	Timestamp         time.Time                        `json:"timestamp"`
}

// LogEntry represents a log entry sent by services
type LogEntry struct {
	ID        string                 `json:"id" db:"id"`
	ServiceID string                 `json:"service_id" db:"service_id"`
	Level     string                 `json:"level" db:"level"` // debug, info, warn, error
	Message   string                 `json:"message" db:"message"`
	Context   map[string]interface{} `json:"context,omitempty" db:"context"`
	Timestamp time.Time              `json:"timestamp" db:"timestamp"`
}

// SendLogsRequest represents a request to send logs
type SendLogsRequest struct {
	ServiceID string     `json:"service_id" binding:"required" example:"user-portal"`
	Logs      []LogEntry `json:"logs" binding:"required"`
}

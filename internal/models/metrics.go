package models

import (
	"time"

	"github.com/google/uuid"
)

// MetricType представляет тип метрики
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"   // Счетчик
	MetricTypeGauge     MetricType = "gauge"     // Измеритель
	MetricTypeHistogram MetricType = "histogram" // Гистограмма
	MetricTypeSummary   MetricType = "summary"   // Сводка
)

// Metric представляет телеметрическую метрику
type Metric struct {
	// ID - уникальный идентификатор метрики
	ID         uuid.UUID              `json:"id" db:"id"`
	// ServiceID - идентификатор сервиса
	ServiceID  uuid.UUID              `json:"service_id" db:"service_id"`
	// Name - название метрики
	Name       string                 `json:"name" db:"name"`
	// Type - тип метрики
	Type       MetricType             `json:"type" db:"type"`
	// Value - значение метрики
	Value      float64                `json:"value" db:"value"`
	// Unit - единица измерения
	Unit       string                 `json:"unit,omitempty" db:"unit"`
	// Labels - метки метрики
	Labels     map[string]string      `json:"labels,omitempty" db:"labels"`
	// Metadata - дополнительные метаданные
	Metadata   map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	// Timestamp - временная метка
	Timestamp  time.Time              `json:"timestamp" db:"timestamp"`
}

// MetricsBatch представляет пакет метрик, отправленных сервисом
type MetricsBatch struct {
	// ServiceID - идентификатор сервиса
	ServiceID uuid.UUID `json:"service_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Metrics - список метрик
	Metrics   []Metric  `json:"metrics" binding:"required"`
	// Timestamp - временная метка пакета
	Timestamp time.Time `json:"timestamp"`
}

// SendMetricsRequest представляет запрос на отправку метрик
type SendMetricsRequest struct {
	// ServiceID - идентификатор сервиса
	ServiceID string   `json:"service_id" binding:"required" example:"transcription-worker-1"`
	// Metrics - список метрик
	Metrics   []Metric `json:"metrics" binding:"required"`
}

// MetricsQueryRequest представляет запрос на получение метрик
type MetricsQueryRequest struct {
	// ServiceID - фильтр по сервису
	ServiceID string    `json:"service_id,omitempty" form:"service_id" example:"transcription-worker-1"`
	// Name - фильтр по названию метрики
	Name      string    `json:"name,omitempty" form:"name" example:"transcription_duration"`
	// From - начало временного периода
	From      time.Time `json:"from,omitempty" form:"from"`
	// To - конец временного периода
	To        time.Time `json:"to,omitempty" form:"to"`
	// Limit - максимальное количество результатов
	Limit     int       `json:"limit,omitempty" form:"limit" example:"100"`
}

// MetricsQueryResponse представляет результаты запроса метрик
type MetricsQueryResponse struct {
	// Metrics - список метрик
	Metrics []Metric `json:"metrics"`
	// Total - общее количество метрик
	Total   int      `json:"total"`
}

// ServiceMetricsSummary представляет агрегированные метрики сервиса
type ServiceMetricsSummary struct {
	// ServiceID - идентификатор сервиса
	ServiceID        uuid.UUID         `json:"service_id"`
	// TotalRequests - общее количество запросов
	TotalRequests    int64             `json:"total_requests"`
	// FailedRequests - количество неудачных запросов
	FailedRequests   int64             `json:"failed_requests"`
	// AverageLatency - средняя задержка в миллисекундах
	AverageLatency   float64           `json:"average_latency_ms"`
	// LastReportedAt - время последнего отчета
	LastReportedAt   time.Time         `json:"last_reported_at"`
	// CustomMetrics - пользовательские метрики
	CustomMetrics    map[string]float64 `json:"custom_metrics,omitempty"`
}

// SystemMetrics представляет общие метрики системы
type SystemMetrics struct {
	// ActiveServices - количество активных сервисов
	ActiveServices    int                              `json:"active_services"`
	// TotalRequests - общее количество запросов
	TotalRequests     int64                            `json:"total_requests"`
	// TotalErrors - общее количество ошибок
	TotalErrors       int64                            `json:"total_errors"`
	// AverageLatency - средняя задержка в миллисекундах
	AverageLatency    float64                          `json:"average_latency_ms"`
	// ServicesSummaries - сводки по сервисам
	ServicesSummaries map[string]ServiceMetricsSummary `json:"services_summaries"`
	// Timestamp - временная метка
	Timestamp         time.Time                        `json:"timestamp"`
}

// LogEntry представляет запись лога, отправленную сервисом
type LogEntry struct {
	// ID - уникальный идентификатор записи
	ID        uuid.UUID              `json:"id" db:"id"`
	// ServiceID - идентификатор сервиса
	ServiceID uuid.UUID              `json:"service_id" db:"service_id"`
	// Level - уровень логирования (debug, info, warn, error)
	Level     string                 `json:"level" db:"level"`
	// Message - сообщение лога
	Message   string                 `json:"message" db:"message"`
	// Context - контекст лога
	Context   map[string]interface{} `json:"context,omitempty" db:"context"`
	// Timestamp - временная метка
	Timestamp time.Time              `json:"timestamp" db:"timestamp"`
}

// SendLogsRequest представляет запрос на отправку логов
type SendLogsRequest struct {
	// ServiceID - идентификатор сервиса
	ServiceID string     `json:"service_id" binding:"required" example:"user-portal"`
	// Logs - список записей логов
	Logs      []LogEntry `json:"logs" binding:"required"`
}

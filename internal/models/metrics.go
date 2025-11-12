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
	// Уникальный идентификатор метрики
	ID         uuid.UUID              `json:"id" db:"id"`
	// Идентификатор сервиса
	ServiceID  uuid.UUID              `json:"service_id" db:"service_id"`
	// Название метрики
	Name       string                 `json:"name" db:"name"`
	// Тип метрики
	Type       MetricType             `json:"type" db:"type"`
	// Значение метрики
	Value      float64                `json:"value" db:"value"`
	// Единица измерения
	Unit       string                 `json:"unit,omitempty" db:"unit"`
	// Метки метрики
	Labels     map[string]string      `json:"labels,omitempty" db:"labels"`
	// Дополнительные метаданные
	Metadata   map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	// Временная метка
	Timestamp  time.Time              `json:"timestamp" db:"timestamp"`
}

// MetricsBatch представляет пакет метрик, отправленных сервисом
type MetricsBatch struct {
	// Идентификатор сервиса
	ServiceID uuid.UUID `json:"service_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Список метрик
	Metrics   []Metric  `json:"metrics" binding:"required"`
	// Временная метка пакета
	Timestamp time.Time `json:"timestamp"`
}

// SendMetricsRequest представляет запрос на отправку метрик
type SendMetricsRequest struct {
	// Идентификатор сервиса
	ServiceID string   `json:"service_id" binding:"required" example:"transcription-worker-1"`
	// Список метрик
	Metrics   []Metric `json:"metrics" binding:"required"`
}

// MetricsQueryRequest представляет запрос на получение метрик
type MetricsQueryRequest struct {
	// Фильтр по сервису
	ServiceID string    `json:"service_id,omitempty" form:"service_id" example:"transcription-worker-1"`
	// Фильтр по названию метрики
	Name      string    `json:"name,omitempty" form:"name" example:"transcription_duration"`
	// Начало временного периода
	From      time.Time `json:"from,omitempty" form:"from"`
	// Конец временного периода
	To        time.Time `json:"to,omitempty" form:"to"`
	// Максимальное количество результатов
	Limit     int       `json:"limit,omitempty" form:"limit" example:"100"`
}

// MetricsQueryResponse представляет результаты запроса метрик
type MetricsQueryResponse struct {
	// Список метрик
	Metrics []Metric `json:"metrics"`
	// Общее количество метрик
	Total   int      `json:"total"`
}

// ServiceMetricsSummary представляет агрегированные метрики сервиса
type ServiceMetricsSummary struct {
	// Идентификатор сервиса
	ServiceID        uuid.UUID         `json:"service_id"`
	// Общее количество запросов
	TotalRequests    int64             `json:"total_requests"`
	// Количество неудачных запросов
	FailedRequests   int64             `json:"failed_requests"`
	// Средняя задержка в миллисекундах
	AverageLatency   float64           `json:"average_latency_ms"`
	// Время последнего отчета
	LastReportedAt   time.Time         `json:"last_reported_at"`
	// Пользовательские метрики
	CustomMetrics    map[string]float64 `json:"custom_metrics,omitempty"`
}

// SystemMetrics представляет общие метрики системы
type SystemMetrics struct {
	// Количество активных сервисов
	ActiveServices    int                              `json:"active_services"`
	// Общее количество запросов
	TotalRequests     int64                            `json:"total_requests"`
	// Общее количество ошибок
	TotalErrors       int64                            `json:"total_errors"`
	// Средняя задержка в миллисекундах
	AverageLatency    float64                          `json:"average_latency_ms"`
	// Сводки по сервисам
	ServicesSummaries map[string]ServiceMetricsSummary `json:"services_summaries"`
	// Временная метка
	Timestamp         time.Time                        `json:"timestamp"`
}

// LogEntry представляет запись лога, отправленную сервисом
type LogEntry struct {
	// Уникальный идентификатор записи
	ID        uuid.UUID              `json:"id" db:"id"`
	// Идентификатор сервиса
	ServiceID uuid.UUID              `json:"service_id" db:"service_id"`
	// Уровень логирования (debug, info, warn, error)
	Level     string                 `json:"level" db:"level"`
	// Сообщение лога
	Message   string                 `json:"message" db:"message"`
	// Контекст лога
	Context   map[string]interface{} `json:"context,omitempty" db:"context"`
	// Временная метка
	Timestamp time.Time              `json:"timestamp" db:"timestamp"`
}

// SendLogsRequest представляет запрос на отправку логов
type SendLogsRequest struct {
	// Идентификатор сервиса
	ServiceID string     `json:"service_id" binding:"required" example:"user-portal"`
	// Список записей логов
	Logs      []LogEntry `json:"logs" binding:"required"`
}

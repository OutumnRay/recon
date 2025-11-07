package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ServiceMetrics contains all Prometheus metrics for Recontext services
type ServiceMetrics struct {
	// Job/Task metrics
	JobsInQueue     prometheus.Gauge
	JobsInProgress  prometheus.Gauge
	JobDuration     prometheus.Histogram
	JobsTotal       *prometheus.CounterVec
	ErrorsTotal     *prometheus.CounterVec

	// Worker-specific metrics
	WorkerTasksProcessed *prometheus.CounterVec
	WorkerErrors         *prometheus.CounterVec

	// Conference/Meeting metrics (for Jitsi agent)
	ActiveConferences prometheus.Gauge
	RecordingBytes    *prometheus.CounterVec

	// HTTP metrics
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPResponseSize     *prometheus.HistogramVec
}

// NewServiceMetrics creates and registers Prometheus metrics for a service
func NewServiceMetrics(serviceName string) *ServiceMetrics {
	return &ServiceMetrics{
		JobsInQueue: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "recontext",
			Subsystem: serviceName,
			Name:      "jobs_in_queue",
			Help:      "Number of jobs currently in queue",
		}),

		JobsInProgress: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "recontext",
			Subsystem: serviceName,
			Name:      "jobs_in_progress",
			Help:      "Number of jobs currently being processed",
		}),

		JobDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "recontext",
			Subsystem: serviceName,
			Name:      "job_duration_seconds",
			Help:      "Duration of job processing in seconds",
			Buckets:   prometheus.ExponentialBuckets(0.1, 2, 10), // 0.1s to ~100s
		}),

		JobsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "recontext",
			Subsystem: serviceName,
			Name:      "jobs_total",
			Help:      "Total number of jobs processed by status",
		}, []string{"status", "type"}),

		ErrorsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "recontext",
			Subsystem: serviceName,
			Name:      "errors_total",
			Help:      "Total number of errors by type",
		}, []string{"type"}),

		WorkerTasksProcessed: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "recontext",
			Subsystem: serviceName,
			Name:      "worker_tasks_processed_total",
			Help:      "Total number of tasks processed by worker",
		}, []string{"worker_id", "task_type"}),

		WorkerErrors: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "recontext",
			Subsystem: serviceName,
			Name:      "worker_errors_total",
			Help:      "Total number of worker errors by type",
		}, []string{"worker_id", "error_type"}),

		ActiveConferences: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "recontext",
			Subsystem: serviceName,
			Name:      "active_conferences",
			Help:      "Number of currently active conferences",
		}),

		RecordingBytes: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "recontext",
			Subsystem: serviceName,
			Name:      "recording_bytes_total",
			Help:      "Total bytes recorded",
		}, []string{"format"}),

		HTTPRequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "recontext",
			Subsystem: serviceName,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		}, []string{"method", "path", "status"}),

		HTTPRequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "recontext",
			Subsystem: serviceName,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "path"}),

		HTTPResponseSize: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "recontext",
			Subsystem: serviceName,
			Name:      "http_response_size_bytes",
			Help:      "HTTP response size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8), // 100B to 100MB
		}, []string{"method", "path"}),
	}
}

// RecordJobStart records the start of a job
func (m *ServiceMetrics) RecordJobStart() {
	m.JobsInProgress.Inc()
	m.JobsInQueue.Dec()
}

// RecordJobComplete records the completion of a job
func (m *ServiceMetrics) RecordJobComplete(duration float64, jobType, status string) {
	m.JobsInProgress.Dec()
	m.JobDuration.Observe(duration)
	m.JobsTotal.WithLabelValues(status, jobType).Inc()
}

// RecordError records an error
func (m *ServiceMetrics) RecordError(errorType string) {
	m.ErrorsTotal.WithLabelValues(errorType).Inc()
}

// RecordHTTPRequest records an HTTP request
func (m *ServiceMetrics) RecordHTTPRequest(method, path, status string, duration float64, responseSize float64) {
	m.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
	m.HTTPResponseSize.WithLabelValues(method, path).Observe(responseSize)
}

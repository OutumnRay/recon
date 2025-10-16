package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"Recontext.online/internal/models"
)

// SendMetrics godoc
// @Summary Send metrics
// @Description Services send telemetry metrics to the managing portal
// @Tags Metrics
// @Accept json
// @Produce json
// @Param request body models.SendMetricsRequest true "Metrics data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/metrics [post]
func (mp *ManagingPortal) sendMetricsHandler(w http.ResponseWriter, r *http.Request) {
	var req models.SendMetricsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Store metrics
	for _, metric := range req.Metrics {
		metric.ServiceID = req.ServiceID
		metric.Timestamp = time.Now()

		mp.metrics = append(mp.metrics, metric)
	}

	mp.logger.Debugf("Received %d metrics from service %s", len(req.Metrics), req.ServiceID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Metrics received successfully",
		"count":   fmt.Sprintf("%d", len(req.Metrics)),
	})
}

// QueryMetrics godoc
// @Summary Query metrics
// @Description Query metrics with optional filters
// @Tags Metrics
// @Produce json
// @Param service_id query string false "Filter by service ID"
// @Param name query string false "Filter by metric name"
// @Param limit query int false "Limit results" default(100)
// @Success 200 {object} models.MetricsQueryResponse
// @Security BearerAuth
// @Router /api/v1/metrics [get]
func (mp *ManagingPortal) queryMetricsHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("service_id")
	name := r.URL.Query().Get("name")

	var filteredMetrics []models.Metric
	for _, metric := range mp.metrics {
		if serviceID != "" && metric.ServiceID != serviceID {
			continue
		}
		if name != "" && metric.Name != name {
			continue
		}
		filteredMetrics = append(filteredMetrics, metric)
	}

	// Limit to last 100
	if len(filteredMetrics) > 100 {
		filteredMetrics = filteredMetrics[len(filteredMetrics)-100:]
	}

	response := models.MetricsQueryResponse{
		Metrics: filteredMetrics,
		Total:   len(filteredMetrics),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSystemMetrics godoc
// @Summary Get system metrics
// @Description Get aggregated system-wide metrics
// @Tags Metrics
// @Produce json
// @Success 200 {object} models.SystemMetrics
// @Security BearerAuth
// @Router /api/v1/metrics/system [get]
func (mp *ManagingPortal) getSystemMetricsHandler(w http.ResponseWriter, r *http.Request) {
	// Aggregate metrics by service
	serviceSummaries := make(map[string]models.ServiceMetricsSummary)

	for _, metric := range mp.metrics {
		summary, exists := serviceSummaries[metric.ServiceID]
		if !exists {
			summary = models.ServiceMetricsSummary{
				ServiceID:      metric.ServiceID,
				CustomMetrics:  make(map[string]float64),
				LastReportedAt: metric.Timestamp,
			}
		}

		// Update summary based on metric type
		switch metric.Name {
		case "requests_total":
			summary.TotalRequests += int64(metric.Value)
		case "requests_failed":
			summary.FailedRequests += int64(metric.Value)
		case "request_latency_ms":
			if summary.AverageLatency == 0 {
				summary.AverageLatency = metric.Value
			} else {
				summary.AverageLatency = (summary.AverageLatency + metric.Value) / 2
			}
		default:
			summary.CustomMetrics[metric.Name] = metric.Value
		}

		if metric.Timestamp.After(summary.LastReportedAt) {
			summary.LastReportedAt = metric.Timestamp
		}

		serviceSummaries[metric.ServiceID] = summary
	}

	// Calculate system-wide totals
	var totalRequests, totalErrors int64
	var avgLatency float64
	latencyCount := 0

	for _, summary := range serviceSummaries {
		totalRequests += summary.TotalRequests
		totalErrors += summary.FailedRequests
		if summary.AverageLatency > 0 {
			avgLatency += summary.AverageLatency
			latencyCount++
		}
	}

	if latencyCount > 0 {
		avgLatency /= float64(latencyCount)
	}

	systemMetrics := models.SystemMetrics{
		ActiveServices:    len(serviceSummaries),
		TotalRequests:     totalRequests,
		TotalErrors:       totalErrors,
		AverageLatency:    avgLatency,
		ServicesSummaries: serviceSummaries,
		Timestamp:         time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(systemMetrics)
}

// SendLogs godoc
// @Summary Send logs
// @Description Services send log entries to the managing portal
// @Tags Metrics
// @Accept json
// @Produce json
// @Param request body models.SendLogsRequest true "Log entries"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/logs [post]
func (mp *ManagingPortal) sendLogsHandler(w http.ResponseWriter, r *http.Request) {
	var req models.SendLogsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Store logs
	for _, logEntry := range req.Logs {
		logEntry.ServiceID = req.ServiceID
		logEntry.Timestamp = time.Now()

		mp.logs = append(mp.logs, logEntry)

		// Also log to our own logger based on level
		switch logEntry.Level {
		case "error":
			mp.logger.Errorf("[%s] %s", req.ServiceID, logEntry.Message)
		case "warn":
			mp.logger.Infof("[%s] WARN: %s", req.ServiceID, logEntry.Message)
		case "debug":
			mp.logger.Debugf("[%s] %s", req.ServiceID, logEntry.Message)
		default:
			mp.logger.Infof("[%s] %s", req.ServiceID, logEntry.Message)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logs received successfully",
		"count":   fmt.Sprintf("%d", len(req.Logs)),
	})
}

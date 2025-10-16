package main

import (
	"encoding/json"
	"net/http"
	"time"

	"Recontext.online/internal/models"
)

type DashboardStats struct {
	Users struct {
		Total  int `json:"total"`
		Active int `json:"active"`
	} `json:"users"`
	Groups struct {
		Total int `json:"total"`
	} `json:"groups"`
	Workers struct {
		Transcription struct {
			Total  int `json:"total"`
			Active int `json:"active"`
		} `json:"transcription"`
		Summarization struct {
			Total  int `json:"total"`
			Active int `json:"active"`
		} `json:"summarization"`
	} `json:"workers"`
	Storage struct {
		Used  int `json:"used"`
		Total int `json:"total"`
	} `json:"storage"`
	Recordings struct {
		Total      int `json:"total"`
		Processing int `json:"processing"`
	} `json:"recordings"`
}

// GetDashboardStats godoc
// @Summary Get dashboard statistics
// @Description Get comprehensive dashboard statistics including users, workers, storage, and recordings
// @Tags Dashboard
// @Produce json
// @Success 200 {object} DashboardStats
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/dashboard/stats [get]
func (mp *ManagingPortal) dashboardStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats := DashboardStats{}

	// Count users
	stats.Users.Total = len(mp.users)
	stats.Users.Active = 0
	for _, user := range mp.users {
		if user.IsActive {
			stats.Users.Active++
		}
	}

	// Count groups
	stats.Groups.Total = len(mp.groups)

	// Count workers by type
	now := time.Now()
	heartbeatTimeout := 2 * time.Minute

	for _, service := range mp.services {
		isActive := now.Sub(service.LastHeartbeat) < heartbeatTimeout && service.Status == models.ServiceStatusRunning

		switch service.Type {
		case "transcription-worker":
			stats.Workers.Transcription.Total++
			if isActive {
				stats.Workers.Transcription.Active++
			}
		case "summarization-worker":
			stats.Workers.Summarization.Total++
			if isActive {
				stats.Workers.Summarization.Active++
			}
		}
	}

	// Storage stats (placeholder - replace with actual MinIO queries)
	stats.Storage.Total = 100 // GB
	stats.Storage.Used = 15   // GB

	// Recording stats (placeholder - replace with actual database queries)
	stats.Recordings.Total = 42
	stats.Recordings.Processing = 3

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

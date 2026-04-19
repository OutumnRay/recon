package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
	"github.com/google/uuid"
)

type VideoInfo struct {
	ID            string  `json:"id"`
	Type          string  `json:"type"` // "upload" or "meeting"
	Title         string  `json:"title"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
	Url           string  `json:"url"`
	HasTranscript bool    `json:"has_transcript"`
	RoomSID       *string `json:"room_sid,omitempty"`
	MeetingID     *string `json:"meeting_id,omitempty"`
	FileID        *string `json:"file_id,omitempty"`
}

type ListVideosResponse struct {
	Videos   []VideoInfo `json:"videos"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// ListVideos godoc
// @Summary Список загруженных видео
// @Description Получить список загруженных файлов и записей встреч пользователя
// @Tags Videos
// @Produce json
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(20)
// @Success 200 {object} ListVideosResponse
// @Security BearerAuth
// @Router /api/v1/videos [get]
func (up *UserPortal) listVideosHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Parse pagination
	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	pageSizeStr := r.URL.Query().Get("page_size")
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	videos := []VideoInfo{}

	// 1. Uploaded files (/api/v1/files)
	files, totalFiles, err := up.db.ListUploadedFilesByUser(claims.UserID, page, pageSize)
	if err == nil {
		for _, file := range files {
			videos = append(videos, VideoInfo{
				ID:            file.ID.String(),
				Type:          "upload",
				Title:         file.OriginalName,
				Status:        string(file.Status),
				CreatedAt:     file.UploadedAt.Format(time.RFC3339),
				Url:           fmt.Sprintf("https://storage.recontext.online/%s", file.StoragePath),
				HasTranscript: file.Status == models.StatusCompleted,
				FileID:        &file.ID.String(),
			})
		}
	}

	// 2. User meetings recordings (simple list without full join for demo)
	meetings, _ := up.meetingRepo.ListMeetings(claims.UserID, 1, 10) // Adjust to user filter if needed
	for _, meeting := range meetings {
		rooms, _ := up.liveKitRepo.GetRoomsByName(meeting.ID.String())
		for _, room := range rooms {
			if room.Status == "finished" || room.Status == "completed" {
				video := VideoInfo{
					ID:            room.SID,
					Type:          "meeting",
					Title:         meeting.Title,
					Status:        room.Status,
					CreatedAt:     room.StartedAt.Format(time.RFC3339),
					Url:           fmt.Sprintf("/api/v1/meetings/%s/recordings", meeting.ID),
					HasTranscript: room.SummaryStatus == "completed",
					RoomSID:       &room.SID,
					MeetingID:     &meeting.ID.String(),
				}
				videos = append(videos, video)
			}
		}
	}

	response := ListVideosResponse{
		Videos:   videos,
		Total:    len(videos),
		Page:     page,
		PageSize: pageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetVideoTranscript godoc
// @Summary Получить транскрипцию видео
// @Description Получить транскрипцию и memo для видео (roomSid or fileID)
// @Tags Videos
// @Produce json
// @Param videoId path string true "Video ID (roomSid or fileID)"
// @Success 200 {object} models.RoomTranscriptsResponse
// @Security BearerAuth
// @Router /api/v1/videos/{videoId} [get]
func (up *UserPortal) getVideoTranscriptHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	videoId := strings.TrimPrefix(r.URL.Path, "/api/v1/videos/")
	videoId = strings.TrimSuffix(videoId, "/")

	// Try as room SID (meeting transcript)
	room, err := up.liveKitRepo.GetRoomBySID(videoId)
	if err == nil && room != nil {
		// Call room transcripts handler logic
		roomTranscripts := up.getRoomTranscriptsResponse(room.SID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(roomTranscripts)
		return
	}

	// Try as file ID (uploaded file)
	fileID, err := uuid.Parse(videoId)
	if err == nil {
		file, err := up.db.GetUploadedFile(fileID)
		if err == nil && file != nil && file.UserID == claims.UserID {
			phrases := up.getFileTranscripts(fileID)
			response := map[string]interface{}{
				"file_id": videoId,
				"phrases": phrases,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	up.respondWithError(w, http.StatusNotFound, "Video not found or no transcript", "")
}

// Helper to get room transcripts (duplicate from getRoomTranscriptsHandler but return struct)
func (up *UserPortal) getRoomTranscriptsResponse(roomSID string) models.RoomTranscriptsResponse {
	// Simplified - duplicate logic or extract to method
	room, _ := up.liveKitRepo.GetRoomBySID(roomSID)
	if room == nil {
		return models.RoomTranscriptsResponse{}
	}

	// Get tracks, phrases, merge etc. (as in getRoomTranscriptsHandler)
	// For now stub
	return models.RoomTranscriptsResponse{
		RoomSID: roomSID,
		Memo:    room.Memo,
	}
}

// Helper to get file transcripts
func (up *UserPortal) getFileTranscripts(fileID uuid.UUID) []TranscriptionPhrase {
	var phrases []models.TranscriptionPhrase
	up.db.DB.Where("file_id = ?", fileID).Order("phrase_index ASC").Find(&phrases)

	rawPhrases := make([]TranscriptionPhrase, len(phrases))
	for i, p := range phrases {
		rawPhrases[i] = TranscriptionPhrase{
			Start: p.StartTime,
			End:   p.EndTime,
			Text:  p.Text,
		}
	}

	return mergePhrases(rawPhrases, 100, 3.0)
}

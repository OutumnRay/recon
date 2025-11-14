package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
	"github.com/google/uuid"
)

// StartRecordingRequest represents a request to start recording
type StartRecordingRequest struct {
	AudioOnly bool `json:"audio_only"`
}

// StartRecording godoc
// @Summary Start recording for a meeting
// @Description Starts room composite recording (audio/video) for an active meeting
// @Tags Recording
// @Accept json
// @Produce json
// @Param id path string true "Meeting ID"
// @Param request body StartRecordingRequest true "Recording options"
// @Success 200 {object} models.MeetingWithDetails
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/meetings/{id}/recording/start [post]
func (up *UserPortal) startRecordingHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Extract meeting ID from path
	meetingID := strings.TrimPrefix(r.URL.Path, "/api/v1/meetings/")
	meetingID = strings.TrimSuffix(meetingID, "/recording/start")

	var req StartRecordingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get meeting
	meeting, err := up.meetingRepo.GetMeetingByID(uuid.Must(uuid.Parse(meetingID)))
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "Meeting not found", err.Error())
		return
	}

	// Check permissions
	if claims.Role != models.RoleAdmin && meeting.CreatedBy != claims.UserID {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "Only meeting creator or admin can control recording")
		return
	}

	// Update meeting status
	meeting.IsRecording = true
	if err := up.meetingRepo.UpdateMeeting(meeting); err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to update meeting", err.Error())
		return
	}

	up.logger.Infof("Recording started for meeting %s by user %s", meetingID, claims.UserID)

	// Get full meeting details
	details, err := up.meetingRepo.GetMeetingWithDetails(uuid.Must(uuid.Parse(meetingID)))
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get meeting details", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}

// StopRecording godoc
// @Summary Stop recording for a meeting
// @Description Stops room composite recording for an active meeting
// @Tags Recording
// @Accept json
// @Produce json
// @Param id path string true "Meeting ID"
// @Success 200 {object} models.MeetingWithDetails
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/meetings/{id}/recording/stop [post]
func (up *UserPortal) stopRecordingHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Extract meeting ID from path
	meetingID := strings.TrimPrefix(r.URL.Path, "/api/v1/meetings/")
	meetingID = strings.TrimSuffix(meetingID, "/recording/stop")

	// Get meeting
	meeting, err := up.meetingRepo.GetMeetingByID(uuid.Must(uuid.Parse(meetingID)))
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "Meeting not found", err.Error())
		return
	}

	// Check permissions
	if claims.Role != models.RoleAdmin && meeting.CreatedBy != claims.UserID {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "Only meeting creator or admin can control recording")
		return
	}

	// Update meeting status
	meeting.IsRecording = false
	if err := up.meetingRepo.UpdateMeeting(meeting); err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to update meeting", err.Error())
		return
	}

	up.logger.Infof("Recording stopped for meeting %s by user %s", meetingID, claims.UserID)

	// Get full meeting details
	details, err := up.meetingRepo.GetMeetingWithDetails(uuid.Must(uuid.Parse(meetingID)))
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get meeting details", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}

// StartTranscription godoc
// @Summary Start transcription for a meeting
// @Description Starts track composite recording (individual audio tracks) for transcription
// @Tags Recording
// @Accept json
// @Produce json
// @Param id path string true "Meeting ID"
// @Success 200 {object} models.MeetingWithDetails
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/meetings/{id}/transcription/start [post]
func (up *UserPortal) startTranscriptionHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Extract meeting ID from path
	meetingID := strings.TrimPrefix(r.URL.Path, "/api/v1/meetings/")
	meetingID = strings.TrimSuffix(meetingID, "/transcription/start")

	// Get meeting
	meeting, err := up.meetingRepo.GetMeetingByID(uuid.Must(uuid.Parse(meetingID)))
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "Meeting not found", err.Error())
		return
	}

	// Check permissions
	if claims.Role != models.RoleAdmin && meeting.CreatedBy != claims.UserID {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "Only meeting creator or admin can control transcription")
		return
	}

	// Update meeting status
	meeting.IsTranscribing = true
	if err := up.meetingRepo.UpdateMeeting(meeting); err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to update meeting", err.Error())
		return
	}

	up.logger.Infof("Transcription started for meeting %s by user %s", meetingID, claims.UserID)

	// Get full meeting details
	details, err := up.meetingRepo.GetMeetingWithDetails(uuid.Must(uuid.Parse(meetingID)))
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get meeting details", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}

// StopTranscription godoc
// @Summary Stop transcription for a meeting
// @Description Stops track composite recording for transcription
// @Tags Recording
// @Accept json
// @Produce json
// @Param id path string true "Meeting ID"
// @Success 200 {object} models.MeetingWithDetails
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/meetings/{id}/transcription/stop [post]
func (up *UserPortal) stopTranscriptionHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Extract meeting ID from path
	meetingID := strings.TrimPrefix(r.URL.Path, "/api/v1/meetings/")
	meetingID = strings.TrimSuffix(meetingID, "/transcription/stop")

	// Get meeting
	meeting, err := up.meetingRepo.GetMeetingByID(uuid.Must(uuid.Parse(meetingID)))
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "Meeting not found", err.Error())
		return
	}

	// Check permissions
	if claims.Role != models.RoleAdmin && meeting.CreatedBy != claims.UserID {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "Only meeting creator or admin can control transcription")
		return
	}

	// Update meeting status
	meeting.IsTranscribing = false
	if err := up.meetingRepo.UpdateMeeting(meeting); err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to update meeting", err.Error())
		return
	}

	up.logger.Infof("Transcription stopped for meeting %s by user %s", meetingID, claims.UserID)

	// Get full meeting details
	details, err := up.meetingRepo.GetMeetingWithDetails(uuid.Must(uuid.Parse(meetingID)))
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get meeting details", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}

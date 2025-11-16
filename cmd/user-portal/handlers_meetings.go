package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
	"github.com/google/uuid"
)

// CreateMeeting godoc
// @Summary Создать новую встречу
// @Description Создать новую встречу (требуется разрешение can_schedule_meetings или роль admin/operator)
// @Tags Meetings
// @Accept json
// @Produce json
// @Param request body models.CreateMeetingRequest true "Данные встречи"
// @Success 201 {object} models.MeetingWithDetails
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/meetings [post]
func (up *UserPortal) createMeetingHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Get user from database to check permissions and organization
	user, err := up.userRepo.GetByID(claims.UserID)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get user", err.Error())
		return
	}

	// Check if user has permission to schedule meetings
	// Admin and operators can always create meetings
	if claims.Role != models.RoleAdmin && claims.Role != models.RoleOperator {
		if !user.Permissions.CanScheduleMeetings {
			up.respondWithError(w, http.StatusForbidden, "You don't have permission to schedule meetings", "Contact administrator to grant can_schedule_meetings permission")
			return
		}
	}

	var req models.CreateMeetingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate required fields
	if req.Title == "" {
		up.respondWithError(w, http.StatusBadRequest, "Title is required", "")
		return
	}
	if req.Duration <= 0 {
		up.respondWithError(w, http.StatusBadRequest, "Duration must be positive", "")
		return
	}
	if req.Type == "" {
		up.respondWithError(w, http.StatusBadRequest, "Type is required", "")
		return
	}

	// Verify subject exists and belongs to the same organization if provided
	if req.SubjectID != nil {
		subject, err := up.meetingRepo.GetSubjectByID(*req.SubjectID)
		if err != nil {
			up.respondWithError(w, http.StatusBadRequest, "Invalid subject ID", err.Error())
			return
		}
		// Validate subject belongs to user's organization
		if user.OrganizationID != nil && subject.OrganizationID != nil {
			if *user.OrganizationID != *subject.OrganizationID {
				up.respondWithError(w, http.StatusForbidden, "Subject must belong to your organization", "")
				return
			}
		}
	}

	// Validate that all participants belong to the same organization
	allParticipantIDs := req.ParticipantIDs
	if req.SpeakerID != nil && *req.SpeakerID != uuid.Nil {
		allParticipantIDs = append(allParticipantIDs, *req.SpeakerID)
	}
	for _, participantID := range allParticipantIDs {
		participant, err := up.userRepo.GetByID(participantID)
		if err != nil {
			up.respondWithError(w, http.StatusBadRequest, "Invalid participant ID: "+participantID.String(), err.Error())
			return
		}
		// Validate participant belongs to user's organization
		if user.OrganizationID != nil && participant.OrganizationID != nil {
			if *user.OrganizationID != *participant.OrganizationID {
				up.respondWithError(w, http.StatusForbidden, "All participants must belong to your organization", "")
				return
			}
		}
	}

	// Validate that all departments belong to the same organization
	for _, deptID := range req.DepartmentIDs {
		dept, err := up.departmentRepo.GetByID(deptID)
		if err != nil {
			up.respondWithError(w, http.StatusBadRequest, "Invalid department ID: "+deptID.String(), err.Error())
			return
		}
		// Validate department belongs to user's organization
		if user.OrganizationID != nil && dept.OrganizationID != nil {
			if *user.OrganizationID != *dept.OrganizationID {
				up.respondWithError(w, http.StatusForbidden, "All departments must belong to your organization", "")
				return
			}
		}
	}

	// Create meeting
	meetingID := uuid.New()

	// Don't assign LiveKit room ID yet - it will be created when meeting starts
	meeting := &models.Meeting{
		ID:                 meetingID,
		Title:              req.Title,
		ScheduledAt:        req.ScheduledAt,
		Duration:           req.Duration,
		Recurrence:         req.Recurrence,
		Type:               req.Type,
		SubjectID:          req.SubjectID,
		Status:             models.MeetingStatusScheduled,
		NeedsVideoRecord:   req.NeedsVideoRecord,
		NeedsAudioRecord:   req.NeedsAudioRecord,
		NeedsTranscription: req.NeedsTranscription,
		AdditionalNotes:    req.AdditionalNotes,
		ForceEndAtDuration: req.ForceEndAtDuration,
		IsPermanent:        req.IsPermanent,
		AllowAnonymous:     req.AllowAnonymous,
		LiveKitRoomID:      nil, // Will be set below
		CreatedBy:          claims.UserID,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Create LiveKit room ID immediately
	livekitRoomID := uuid.New()
	meeting.LiveKitRoomID = &livekitRoomID

	if err := up.meetingRepo.CreateMeeting(meeting); err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to create meeting", err.Error())
		return
	}

	// Add the creator as an organizer participant
	creatorParticipant := &models.MeetingParticipant{
		ID:        uuid.New(),
		MeetingID: meetingID,
		UserID:    claims.UserID,
		Role:      "organizer",
		Status:    "invited",
		CreatedAt: time.Now(),
	}
	if err := up.meetingRepo.AddParticipant(creatorParticipant); err != nil {
		up.logger.Infof("Failed to add creator as participant to meeting: %v", err)
	}

	// Add speaker if provided (for presentations)
	if req.SpeakerID != nil && *req.SpeakerID != uuid.Nil {
		participant := &models.MeetingParticipant{
			ID:        uuid.New(),
			MeetingID: meetingID,
			UserID:    *req.SpeakerID,
			Role:      "speaker",
			Status:    "invited",
			CreatedAt: time.Now(),
		}
		if err := up.meetingRepo.AddParticipant(participant); err != nil {
			up.logger.Infof("Failed to add speaker to meeting: %v", err)
		}
	}

	// Add participants
	for _, userID := range req.ParticipantIDs {
		participant := &models.MeetingParticipant{
			ID:        uuid.New(),
			MeetingID: meetingID,
			UserID:    userID,
			Role:      "participant",
			Status:    "invited",
			CreatedAt: time.Now(),
		}
		if err := up.meetingRepo.AddParticipant(participant); err != nil {
			up.logger.Infof("Failed to add participant %s to meeting: %v", userID, err)
		}
	}

	// Add departments
	for _, deptID := range req.DepartmentIDs {
		meetingDept := &models.MeetingDepartment{
			ID:           uuid.New(),
			MeetingID:    meetingID,
			DepartmentID: deptID,
			CreatedAt:    time.Now(),
		}
		if err := up.meetingRepo.AddDepartment(meetingDept); err != nil {
			up.logger.Infof("Failed to add department %s to meeting: %v", deptID, err)
		}
	}

	up.logger.Infof("Meeting created: %s by user %s", meetingID, claims.Username)

	// Get full meeting details to return
	details, err := up.meetingRepo.GetMeetingWithDetails(meetingID)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get meeting details", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(details)
}

// ListMyMeetings godoc
// @Summary Список моих встреч
// @Description Получить постраничный список встреч, где пользователь является участником или докладчиком
// @Tags Meetings
// @Produce json
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(20)
// @Param status query string false "Фильтр по статусу" Enums(scheduled, in_progress, completed, cancelled)
// @Param type query string false "Фильтр по типу" Enums(presentation, conference)
// @Param subject_id query string false "Фильтр по идентификатору темы"
// @Param date_from query string false "Фильтр по дате начала (формат RFC3339)"
// @Param date_to query string false "Фильтр по дате окончания (формат RFC3339)"
// @Success 200 {object} models.MeetingsResponse
// @Failure 401 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/meetings [get]
func (up *UserPortal) listMyMeetingsHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Parse pagination parameters
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	pageSize := 20
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 && val <= 100 {
			pageSize = val
		}
	}

	// Build request filters
	req := models.ListMeetingsRequest{
		Page:     page,
		PageSize: pageSize,
		UserID:   &claims.UserID, // Filter by current user
	}

	// Parse optional filters
	if status := r.URL.Query().Get("status"); status != "" {
		meetingStatus := models.MeetingStatus(status)
		req.Status = &meetingStatus
	} else {
		// By default, exclude cancelled meetings
		// This ensures cancelled meetings are hidden unless explicitly requested
		req.ExcludeCancelled = true
	}

	if meetingType := r.URL.Query().Get("type"); meetingType != "" {
		mType := models.MeetingType(meetingType)
		req.Type = &mType
	}

	if subjectID := r.URL.Query().Get("subject_id"); subjectID != "" {
		if subjectUUID, err := uuid.Parse(subjectID); err == nil {
			req.SubjectID = &subjectUUID
		}
	}

	if dateFrom := r.URL.Query().Get("date_from"); dateFrom != "" {
		if t, err := time.Parse(time.RFC3339, dateFrom); err == nil {
			req.DateFrom = &t
		}
	}

	if dateTo := r.URL.Query().Get("date_to"); dateTo != "" {
		if t, err := time.Parse(time.RFC3339, dateTo); err == nil {
			req.DateTo = &t
		}
	}

	// Get meetings from repository
	response, err := up.meetingRepo.ListMeetings(req)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to list meetings", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMeeting godoc
// @Summary Получить детали встречи
// @Description Получить детальную информацию о конкретной встрече (должен быть участником/докладчиком или администратором)
// @Tags Meetings
// @Produce json
// @Param id path string true "Идентификатор встречи"
// @Success 200 {object} models.MeetingWithDetails
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/meetings/{id} [get]
func (up *UserPortal) getMeetingHandler(w http.ResponseWriter, r *http.Request) {
	// Extract meeting ID from path
	meetingID := strings.TrimPrefix(r.URL.Path, "/api/v1/meetings/")

	// Get meeting details
	meeting, err := up.meetingRepo.GetMeetingWithDetails(uuid.Must(uuid.Parse(meetingID)))
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "Meeting not found", err.Error())
		return
	}

	// If meeting allows anonymous access, allow anyone to view it
	if meeting.AllowAnonymous {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(meeting)
		return
	}

	// Otherwise, require authentication
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Check if user is participant, speaker, or admin
	isParticipant := false
	if claims.Role == models.RoleAdmin {
		isParticipant = true
	} else {
		for _, participant := range meeting.Participants {
			if participant.UserID == claims.UserID {
				isParticipant = true
				break
			}
		}
	}

	if !isParticipant {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "You are not a participant of this meeting")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meeting)
}

// UpdateMeeting godoc
// @Summary Обновить встречу
// @Description Обновить информацию о встрече (должен быть создателем или администратором)
// @Tags Meetings
// @Accept json
// @Produce json
// @Param id path string true "Идентификатор встречи"
// @Param request body models.UpdateMeetingRequest true "Данные для обновления встречи"
// @Success 200 {object} models.MeetingWithDetails
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/meetings/{id} [put]
func (up *UserPortal) updateMeetingHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Extract meeting ID from path
	meetingID := strings.TrimPrefix(r.URL.Path, "/api/v1/meetings/")

	var req models.UpdateMeetingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get existing meeting
	meeting, err := up.meetingRepo.GetMeetingByID(func() uuid.UUID { id, _ := uuid.Parse(meetingID); return id }())
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "Meeting not found", err.Error())
		return
	}

	// Check if user is creator or admin
	if claims.Role != models.RoleAdmin && meeting.CreatedBy != claims.UserID {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "Only meeting creator or admin can update meetings")
		return
	}

	// Get user to check organization
	user, err := up.userRepo.GetByID(claims.UserID)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get user", err.Error())
		return
	}

	// Update fields if provided
	if req.Title != nil {
		meeting.Title = *req.Title
	}
	if req.ScheduledAt != nil {
		meeting.ScheduledAt = *req.ScheduledAt
	}
	if req.Duration != nil {
		if *req.Duration <= 0 {
			up.respondWithError(w, http.StatusBadRequest, "Duration must be positive", "")
			return
		}
		meeting.Duration = *req.Duration
	}
	if req.Recurrence != nil {
		meeting.Recurrence = *req.Recurrence
	}
	if req.Type != nil {
		meeting.Type = *req.Type
	}
	if req.SubjectID != nil {
		// Verify subject exists and belongs to the same organization
		subject, err := up.meetingRepo.GetSubjectByID(*req.SubjectID)
		if err != nil {
			up.respondWithError(w, http.StatusBadRequest, "Invalid subject ID", err.Error())
			return
		}
		// Validate subject belongs to user's organization
		if user.OrganizationID != nil && subject.OrganizationID != nil {
			if *user.OrganizationID != *subject.OrganizationID {
				up.respondWithError(w, http.StatusForbidden, "Subject must belong to your organization", "")
				return
			}
		}
		meeting.SubjectID = req.SubjectID
	}
	if req.Status != nil {
		meeting.Status = *req.Status
	}
	if req.NeedsVideoRecord != nil {
		meeting.NeedsVideoRecord = *req.NeedsVideoRecord
	}
	if req.NeedsAudioRecord != nil {
		meeting.NeedsAudioRecord = *req.NeedsAudioRecord
	}
	if req.NeedsTranscription != nil {
		meeting.NeedsTranscription = *req.NeedsTranscription
	}
	if req.ForceEndAtDuration != nil {
		meeting.ForceEndAtDuration = *req.ForceEndAtDuration
	}
	if req.IsPermanent != nil {
		meeting.IsPermanent = *req.IsPermanent
	}
	if req.AllowAnonymous != nil {
		meeting.AllowAnonymous = *req.AllowAnonymous
	}
	if req.AdditionalNotes != nil {
		meeting.AdditionalNotes = *req.AdditionalNotes
	}

	// Update meeting
	if err := up.meetingRepo.UpdateMeeting(meeting); err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to update meeting", err.Error())
		return
	}

	// Handle participant updates if provided
	if req.ParticipantIDs != nil || req.SpeakerID != nil {
		// Validate that all participants belong to the same organization
		allParticipantIDs := []uuid.UUID{}
		if req.ParticipantIDs != nil {
			allParticipantIDs = append(allParticipantIDs, req.ParticipantIDs...)
		}
		if req.SpeakerID != nil && *req.SpeakerID != uuid.Nil {
			allParticipantIDs = append(allParticipantIDs, *req.SpeakerID)
		}
		for _, participantID := range allParticipantIDs {
			participant, err := up.userRepo.GetByID(participantID)
			if err != nil {
				up.respondWithError(w, http.StatusBadRequest, "Invalid participant ID: "+participantID.String(), err.Error())
				return
			}
			// Validate participant belongs to user's organization
			if user.OrganizationID != nil && participant.OrganizationID != nil {
				if *user.OrganizationID != *participant.OrganizationID {
					up.respondWithError(w, http.StatusForbidden, "All participants must belong to your organization", "")
					return
				}
			}
		}

		// Get current participants to remove old ones (except the creator/organizer)
		currentParticipants, _ := up.meetingRepo.GetMeetingParticipants(uuid.Must(uuid.Parse(meetingID)))
		for _, participant := range currentParticipants {
			// Don't remove the meeting creator
			if participant.UserID != meeting.CreatedBy {
				up.meetingRepo.RemoveParticipant(uuid.Must(uuid.Parse(meetingID)), participant.UserID)
			}
		}

		// Add new speaker if provided
		if req.SpeakerID != nil && *req.SpeakerID != uuid.Nil {
			meetingUUID, _ := uuid.Parse(meetingID)
			participant := &models.MeetingParticipant{
				ID:        uuid.New(),
				MeetingID: meetingUUID,
				UserID:    *req.SpeakerID,
				Role:      "speaker",
				Status:    "invited",
				CreatedAt: time.Now(),
			}
			if err := up.meetingRepo.AddParticipant(participant); err != nil {
				up.logger.Infof("Failed to add speaker to meeting: %v", err)
			}
		}

		// Add new participants
		if req.ParticipantIDs != nil {
			meetingUUID, _ := uuid.Parse(meetingID)
			for _, userID := range req.ParticipantIDs {
				participant := &models.MeetingParticipant{
					ID:        uuid.New(),
					MeetingID: meetingUUID,
					UserID:    userID,
					Role:      "participant",
					Status:    "invited",
					CreatedAt: time.Now(),
				}
				if err := up.meetingRepo.AddParticipant(participant); err != nil {
					up.logger.Infof("Failed to add participant %s to meeting: %v", userID, err)
				}
			}
		}
	}

	// Handle department updates if provided
	if req.DepartmentIDs != nil {
		// Validate that all departments belong to the same organization
		for _, deptID := range req.DepartmentIDs {
			dept, err := up.departmentRepo.GetByID(deptID)
			if err != nil {
				up.respondWithError(w, http.StatusBadRequest, "Invalid department ID: "+deptID.String(), err.Error())
				return
			}
			// Validate department belongs to user's organization
			if user.OrganizationID != nil && dept.OrganizationID != nil {
				if *user.OrganizationID != *dept.OrganizationID {
					up.respondWithError(w, http.StatusForbidden, "All departments must belong to your organization", "")
					return
				}
			}
		}

		// Get current departments to remove old ones
		currentDepts, _ := up.meetingRepo.GetMeetingDepartments(uuid.Must(uuid.Parse(meetingID)))
		for _, dept := range currentDepts {
			up.meetingRepo.RemoveDepartment(uuid.Must(uuid.Parse(meetingID)), dept.ID)
		}

		// Add new departments
		meetingUUID, _ := uuid.Parse(meetingID)
		for _, deptID := range req.DepartmentIDs {
			meetingDept := &models.MeetingDepartment{
				ID:           uuid.New(),
				MeetingID:    meetingUUID,
				DepartmentID: deptID,
				CreatedAt:    time.Now(),
			}
			if err := up.meetingRepo.AddDepartment(meetingDept); err != nil {
				up.logger.Infof("Failed to add department %s to meeting: %v", deptID, err)
			}
		}
	}

	up.logger.Infof("Meeting updated: %s by user %s", meetingID, claims.Username)

	// Get full meeting details to return
	details, err := up.meetingRepo.GetMeetingWithDetails(uuid.Must(uuid.Parse(meetingID)))
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get meeting details", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}

// DeleteMeeting godoc
// @Summary Отменить встречу
// @Description Отменить встречу (изменить статус на cancelled) (должен быть создателем или администратором)
// @Tags Meetings
// @Produce json
// @Param id path string true "Идентификатор встречи"
// @Success 200 {object} map[string]string
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/meetings/{id} [delete]
func (up *UserPortal) deleteMeetingHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Extract meeting ID from path
	meetingID := strings.TrimPrefix(r.URL.Path, "/api/v1/meetings/")

	// Get existing meeting
	meeting, err := up.meetingRepo.GetMeetingByID(func() uuid.UUID { id, _ := uuid.Parse(meetingID); return id }())
	if err != nil {
		up.respondWithError(w, http.StatusNotFound, "Meeting not found", err.Error())
		return
	}

	// Check if user is creator or admin
	if claims.Role != models.RoleAdmin && meeting.CreatedBy != claims.UserID {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "Only meeting creator or admin can cancel meetings")
		return
	}

	// Cancel meeting by updating status instead of deleting
	meeting.Status = models.MeetingStatusCancelled
	meeting.UpdatedAt = time.Now()

	if err := up.meetingRepo.UpdateMeeting(meeting); err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to cancel meeting", err.Error())
		return
	}

	up.logger.Infof("Meeting cancelled: %s by user %s", meetingID, claims.Username)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Meeting cancelled successfully",
		"meeting_id": meetingID,
	})
}

// ListMeetingSubjects godoc
// @Summary Список тем встреч
// @Description Получить постраничный список активных тем встреч
// @Tags Meetings
// @Produce json
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(100)
// @Param include_inactive query bool false "Включать неактивные темы" default(false)
// @Success 200 {object} models.MeetingSubjectsResponse
// @Failure 401 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/meeting-subjects [get]
func (up *UserPortal) listMeetingSubjectsHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Get user to determine organization
	user, err := up.userRepo.GetByID(claims.UserID)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get user", err.Error())
		return
	}

	// Parse pagination parameters
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	pageSize := 100
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 && val <= 1000 {
			pageSize = val
		}
	}

	// Parse include_inactive parameter
	includeInactive := false
	if ia := r.URL.Query().Get("include_inactive"); ia == "true" {
		includeInactive = true
	}

	// Get subjects from repository filtered by user's organization
	response, err := up.meetingRepo.ListSubjects(page, pageSize, nil, includeInactive, user.OrganizationID)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to list meeting subjects", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

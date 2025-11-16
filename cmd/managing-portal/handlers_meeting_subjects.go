package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"Recontext.online/internal/models"
)

// CreateMeetingSubject godoc
// @Summary Создать новую тему встречи
// @Description Создать новую тему/категорию встречи (только администратор)
// @Tags Meeting Subjects
// @Accept json
// @Produce json
// @Param request body models.CreateMeetingSubjectRequest true "Subject creation data"
// @Success 201 {object} models.MeetingSubject
// @Failure 400 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/meeting-subjects [post]
func (mp *ManagingPortal) createMeetingSubjectHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateMeetingSubjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate required fields
	if req.Name == "" {
		mp.respondWithError(w, http.StatusBadRequest, "Subject name is required", "")
		return
	}

	subject := &models.MeetingSubject{
		ID:             uuid.New(),
		Name:           req.Name,
		Description:    req.Description,
		DepartmentIDs:  req.DepartmentIDs,
		OrganizationID: req.OrganizationID,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if subject.DepartmentIDs == nil {
		subject.DepartmentIDs = []uuid.UUID{}
	}

	if err := mp.meetingRepo.CreateSubject(subject); err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to create subject", err.Error())
		return
	}

	mp.logger.Infof("Meeting subject created: %s (%s)", subject.Name, subject.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(subject)
}

// ListMeetingSubjects godoc
// @Summary Список тем встреч
// @Description Получить постраничный список тем встреч с дополнительными фильтрами
// @Tags Meeting Subjects
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param department_id query string false "Filter by department ID"
// @Param include_inactive query bool false "Include inactive subjects" default(false)
// @Success 200 {object} models.MeetingSubjectsResponse
// @Security BearerAuth
// @Router /api/v1/meeting-subjects [get]
func (mp *ManagingPortal) listMeetingSubjectsHandler(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	departmentID := r.URL.Query().Get("department_id")
	var deptIDPtr *string
	if departmentID != "" {
		deptIDPtr = &departmentID
	}

	includeInactive := r.URL.Query().Get("include_inactive") == "true"

	response, err := mp.meetingRepo.ListSubjects(page, pageSize, deptIDPtr, includeInactive)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to list subjects", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMeetingSubject godoc
// @Summary Получить тему встречи по ID
// @Description Получить детальную информацию о конкретной теме встречи
// @Tags Meeting Subjects
// @Produce json
// @Param id path string true "Subject ID"
// @Success 200 {object} models.MeetingSubject
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/meeting-subjects/{id} [get]
func (mp *ManagingPortal) getMeetingSubjectHandler(w http.ResponseWriter, r *http.Request) {
	subjectID := strings.TrimPrefix(r.URL.Path, "/api/v1/meeting-subjects/")

	subject, err := mp.meetingRepo.GetSubjectByID(uuid.Must(uuid.Parse(subjectID)))
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "Subject not found", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subject)
}

// UpdateMeetingSubject godoc
// @Summary Обновить тему встречи
// @Description Обновить информацию о теме встречи (только администратор)
// @Tags Meeting Subjects
// @Accept json
// @Produce json
// @Param id path string true "Subject ID"
// @Param request body models.UpdateMeetingSubjectRequest true "Update data"
// @Success 200 {object} models.MeetingSubject
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/meeting-subjects/{id} [put]
func (mp *ManagingPortal) updateMeetingSubjectHandler(w http.ResponseWriter, r *http.Request) {
	subjectID := strings.TrimPrefix(r.URL.Path, "/api/v1/meeting-subjects/")

	var req models.UpdateMeetingSubjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Find subject
	subject, err := mp.meetingRepo.GetSubjectByID(uuid.Must(uuid.Parse(subjectID)))
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "Subject not found", err.Error())
		return
	}

	// Update fields
	if req.Name != "" {
		subject.Name = req.Name
	}
	if req.Description != "" {
		subject.Description = req.Description
	}
	if req.DepartmentIDs != nil {
		subject.DepartmentIDs = req.DepartmentIDs
	}
	if req.IsActive != nil {
		subject.IsActive = *req.IsActive
	}
	if req.OrganizationID != nil {
		subject.OrganizationID = req.OrganizationID
	}

	if err := mp.meetingRepo.UpdateSubject(subject); err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to update subject", err.Error())
		return
	}

	mp.logger.Infof("Meeting subject updated: %s (%s)", subject.Name, subject.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subject)
}

// DeleteMeetingSubject godoc
// @Summary Удалить тему встречи
// @Description Мягкое удаление темы встречи (только администратор)
// @Tags Meeting Subjects
// @Produce json
// @Param id path string true "Subject ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/meeting-subjects/{id} [delete]
func (mp *ManagingPortal) deleteMeetingSubjectHandler(w http.ResponseWriter, r *http.Request) {
	subjectID := strings.TrimPrefix(r.URL.Path, "/api/v1/meeting-subjects/")

	if err := mp.meetingRepo.DeleteSubject(uuid.Must(uuid.Parse(subjectID))); err != nil {
		if strings.Contains(err.Error(), "not found") {
			mp.respondWithError(w, http.StatusNotFound, "Subject not found", err.Error())
		} else {
			mp.respondWithError(w, http.StatusInternalServerError, "Failed to delete subject", err.Error())
		}
		return
	}

	mp.logger.Infof("Meeting subject deleted: %s", subjectID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Subject deleted successfully",
		"subject_id": subjectID,
	})
}

package database

import (
	"fmt"
	"time"

	"Recontext.online/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MeetingRepository handles database operations for meetings
type MeetingRepository struct {
	db *DB
}

// NewMeetingRepository creates a new MeetingRepository
func NewMeetingRepository(db *DB) *MeetingRepository {
	return &MeetingRepository{db: db}
}

// Helper functions for UUID conversion
func uuidSliceToStringSlice(uuids []uuid.UUID) pq.StringArray {
	strs := make(pq.StringArray, len(uuids))
	for i, u := range uuids {
		strs[i] = u.String()
	}
	return strs
}

func stringSliceToUUIDSlice(strs []string) ([]uuid.UUID, error) {
	uuids := make([]uuid.UUID, len(strs))
	for i, s := range strs {
		u, parseErr := uuid.Parse(s)
		if parseErr != nil {
			return nil, parseErr
		}
		uuids[i] = u
	}
	return uuids, nil
}

// ============= Meeting Subjects =============

// CreateSubject creates a new meeting subject
func (r *MeetingRepository) CreateSubject(subject *models.MeetingSubject) error {
	dbSubject := &MeetingSubject{
		ID:             subject.ID,
		Name:           subject.Name,
		Description:    subject.Description,
		DepartmentIDs:  uuidSliceToStringSlice(subject.DepartmentIDs),
		OrganizationID: subject.OrganizationID,
		IsActive:       subject.IsActive,
		CreatedAt:      subject.CreatedAt,
		UpdatedAt:      subject.UpdatedAt,
	}

	if err := r.db.DB.Create(dbSubject).Error; err != nil {
		return fmt.Errorf("failed to create meeting subject: %w", err)
	}

	return nil
}

// GetSubjectByID retrieves a meeting subject by ID
func (r *MeetingRepository) GetSubjectByID(id uuid.UUID) (*models.MeetingSubject, error) {
	var dbSubject MeetingSubject

	if err := r.db.DB.Where("id = ?", id).First(&dbSubject).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("meeting subject not found")
		}
		return nil, fmt.Errorf("failed to get meeting subject: %w", err)
	}

	departmentIDs, err := stringSliceToUUIDSlice(dbSubject.DepartmentIDs)
	if err != nil {
		return nil, err
	}

	subject := &models.MeetingSubject{
		ID:            dbSubject.ID,
		Name:          dbSubject.Name,
		Description:   dbSubject.Description,
		DepartmentIDs: departmentIDs,
		IsActive:      dbSubject.IsActive,
		CreatedAt:     dbSubject.CreatedAt,
		UpdatedAt:     dbSubject.UpdatedAt,
	}

	return subject, nil
}

// ListSubjects retrieves meeting subjects with pagination
func (r *MeetingRepository) ListSubjects(page, pageSize int, departmentID *string, includeInactive bool, organizationID *uuid.UUID) (*models.MeetingSubjectsResponse, error) {
	offset := (page - 1) * pageSize

	// Build query
	query := r.db.DB.Model(&MeetingSubject{})

	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}

	if departmentID != nil && *departmentID != "" {
		query = query.Where("? = ANY(department_ids)", *departmentID)
	}

	// Filter by organization ID if provided
	if organizationID != nil {
		query = query.Where("organization_id = ?", *organizationID)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count subjects: %w", err)
	}

	// Get page data
	var dbSubjects []MeetingSubject
	if err := query.Order("name ASC").Limit(pageSize).Offset(offset).Find(&dbSubjects).Error; err != nil {
		return nil, fmt.Errorf("failed to list subjects: %w", err)
	}

	subjects := make([]models.MeetingSubject, len(dbSubjects))
	for i, dbSubject := range dbSubjects {
		departmentIDs, err := stringSliceToUUIDSlice(dbSubject.DepartmentIDs)
		if err != nil {
			return nil, err
		}

		subjects[i] = models.MeetingSubject{
			ID:            dbSubject.ID,
			Name:          dbSubject.Name,
			Description:   dbSubject.Description,
			DepartmentIDs: departmentIDs,
			IsActive:      dbSubject.IsActive,
			CreatedAt:     dbSubject.CreatedAt,
			UpdatedAt:     dbSubject.UpdatedAt,
		}
	}

	return &models.MeetingSubjectsResponse{
		Items:    subjects,
		Offset:   offset,
		PageSize: pageSize,
		Total:    int(total),
	}, nil
}

// UpdateSubject updates a meeting subject
func (r *MeetingRepository) UpdateSubject(subject *models.MeetingSubject) error {
	subject.UpdatedAt = time.Now()

	result := r.db.DB.Model(&MeetingSubject{}).Where("id = ?", subject.ID).Updates(map[string]interface{}{
		"name":            subject.Name,
		"description":     subject.Description,
		"department_ids":  uuidSliceToStringSlice(subject.DepartmentIDs),
		"organization_id": subject.OrganizationID,
		"is_active":       subject.IsActive,
		"updated_at":      subject.UpdatedAt,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to update subject: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("meeting subject not found")
	}

	return nil
}

// DeleteSubject soft deletes a meeting subject
func (r *MeetingRepository) DeleteSubject(id uuid.UUID) error {
	result := r.db.DB.Model(&MeetingSubject{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_active":  false,
		"updated_at": time.Now(),
	})

	if result.Error != nil {
		return fmt.Errorf("failed to delete subject: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("meeting subject not found")
	}

	return nil
}

// ============= Meetings =============

// CreateMeeting creates a new meeting
func (r *MeetingRepository) CreateMeeting(meeting *models.Meeting) error {
	dbMeeting := &Meeting{
		ID:                 meeting.ID,
		Title:              meeting.Title,
		ScheduledAt:        meeting.ScheduledAt,
		Duration:           meeting.Duration,
		Recurrence:         string(meeting.Recurrence),
		Type:               string(meeting.Type),
		SubjectID:          meeting.SubjectID,
		Status:             string(meeting.Status),
		NeedsRecord:        meeting.NeedsRecord,
		NeedsTranscription: meeting.NeedsTranscription,
		IsRecording:        meeting.IsRecording,
		IsTranscribing:     meeting.IsTranscribing,
		IsPermanent:        meeting.IsPermanent,
		AllowAnonymous:     meeting.AllowAnonymous,
		AdditionalNotes:    meeting.AdditionalNotes,
		LiveKitRoomID:      meeting.LiveKitRoomID,
		CreatedBy:          meeting.CreatedBy,
		CreatedAt:          meeting.CreatedAt,
		UpdatedAt:          meeting.UpdatedAt,
	}

	if err := r.db.DB.Create(dbMeeting).Error; err != nil {
		return fmt.Errorf("failed to create meeting: %w", err)
	}

	return nil
}

// GetMeetingByID retrieves a meeting by ID
func (r *MeetingRepository) GetMeetingByID(id uuid.UUID) (*models.Meeting, error) {
	var dbMeeting Meeting

	if err := r.db.DB.Where("id = ?", id).First(&dbMeeting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("meeting not found")
		}
		return nil, fmt.Errorf("failed to get meeting: %w", err)
	}

	// No parsing needed - dbMeeting fields are already uuid.UUID
	meeting := &models.Meeting{
		ID:                 dbMeeting.ID,
		Title:              dbMeeting.Title,
		ScheduledAt:        dbMeeting.ScheduledAt,
		Duration:           dbMeeting.Duration,
		Recurrence:         models.MeetingRecurrence(dbMeeting.Recurrence),
		Type:               models.MeetingType(dbMeeting.Type),
		SubjectID:          dbMeeting.SubjectID,
		Status:             models.MeetingStatus(dbMeeting.Status),
		NeedsRecord:        dbMeeting.NeedsRecord,
		NeedsTranscription: dbMeeting.NeedsTranscription,
		IsRecording:        dbMeeting.IsRecording,
		IsTranscribing:     dbMeeting.IsTranscribing,
		IsPermanent:        dbMeeting.IsPermanent,
		AllowAnonymous:     dbMeeting.AllowAnonymous,
		AdditionalNotes:    dbMeeting.AdditionalNotes,
		LiveKitRoomID:      dbMeeting.LiveKitRoomID,
		CreatedBy:          dbMeeting.CreatedBy,
		CreatedAt:          dbMeeting.CreatedAt,
		UpdatedAt:          dbMeeting.UpdatedAt,
	}

	return meeting, nil
}

// GetMeetingWithDetails retrieves a meeting with all related information
func (r *MeetingRepository) GetMeetingWithDetails(id uuid.UUID) (*models.MeetingWithDetails, error) {
	meeting, err := r.GetMeetingByID(id)
	if err != nil {
		return nil, err
	}

	details := &models.MeetingWithDetails{
		Meeting: *meeting,
	}

	// Get subject if provided
	if meeting.SubjectID != nil {
		subject, err := r.GetSubjectByID(*meeting.SubjectID)
		if err == nil {
			details.Subject = subject
		}
	}

	// Get participants
	participants, err := r.GetMeetingParticipants(id)
	if err == nil {
		details.Participants = participants
	}

	// Get departments
	departments, err := r.GetMeetingDepartments(id)
	if err == nil {
		details.Departments = departments
	}

	// Get creator info
	userRepo := NewUserRepository(r.db)
	creator, err := userRepo.GetByID(meeting.CreatedBy)
	if err == nil {
		details.CreatedByUser = &models.UserInfo{
			ID:           creator.ID,
			Username:     creator.Username,
			Email:        creator.Email,
			Role:         creator.Role,
			DepartmentID: creator.DepartmentID,
			Permissions:  creator.Permissions,
		}
	}

	// Count active participants in LiveKit room
	if meeting.LiveKitRoomID != nil {
		livekitRepo := NewLiveKitRepository(r.db)
		// Get the room by meeting ID (room name = meeting ID)
		rooms, err := livekitRepo.GetRoomsByName(id.String())
		if err == nil && len(rooms) > 0 {
			// Use the most recent room
			room := rooms[0]

			// Подсчёт активных участников (не ботов/рекордеров)
			// Count participants that are ACTIVE and NOT hidden (not bots/recorders)
			// hidden = true means it's a bot/recorder that should not be counted
			// Используем Unscoped() т.к. у модели Participant нет soft delete
			// Use Unscoped() because Participant model doesn't have soft delete
			var activeCount int64
			err = r.db.DB.Model(&LiveKitParticipant{}).Unscoped().
				Where("room_sid = ? AND state = ? AND (permission->>'hidden' IS NULL OR permission->>'hidden' = 'false')",
					room.SID, "ACTIVE").
				Count(&activeCount).Error
			if err == nil {
				details.ActiveParticipantsCount = int(activeCount)
			}

			// Count anonymous guests - participants in LiveKit who are in temporary_users table
			// and are ACTIVE and not hidden
			if meeting.AllowAnonymous {
				var guestCount int64
				err = r.db.DB.Table("livekit_participants").
					Joins("INNER JOIN temporary_users ON livekit_participants.identity = temporary_users.id::text").
					Where("livekit_participants.room_sid = ? AND livekit_participants.state = ? AND (livekit_participants.permission->>'hidden' IS NULL OR livekit_participants.permission->>'hidden' = 'false')",
						room.SID, "ACTIVE").
					Where("temporary_users.meeting_id = ?", id).
					Count(&guestCount).Error
				if err == nil {
					details.AnonymousGuestsCount = int(guestCount)
				}
			}
		}
	}

	return details, nil
}

// ListMeetings retrieves meetings with pagination and filters
func (r *MeetingRepository) ListMeetings(req models.ListMeetingsRequest) (*models.MeetingsResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	// Build query
	query := r.db.DB.Model(&Meeting{})

	if req.Status != nil {
		// Permanent meetings should ALWAYS be shown regardless of status filter
		query = query.Where("(status = ? OR is_permanent = ?)", *req.Status, true)
	} else if req.ExcludeCancelled {
		// Exclude cancelled meetings unless specifically requested
		// But always include permanent meetings regardless
		query = query.Where("(status != ? OR is_permanent = ?)", "cancelled", true)
	}

	if req.Type != nil {
		query = query.Where("type = ?", *req.Type)
	}

	if req.SubjectID != nil {
		query = query.Where("subject_id = ?", *req.SubjectID)
	}

	if req.DateFrom != nil {
		query = query.Where("scheduled_at >= ?", *req.DateFrom)
	}

	if req.DateTo != nil {
		query = query.Where("scheduled_at <= ?", *req.DateTo)
	}

	// Filter by speaker or participant
	if req.SpeakerID != nil || req.UserID != nil {
		userID := req.SpeakerID
		if req.UserID != nil {
			userID = req.UserID
		}

		query = query.Where("id IN (?)",
			r.db.DB.Table("meeting_participants").Select("meeting_id").Where("user_id = ?", *userID),
		)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count meetings: %w", err)
	}

	// Get page data
	var dbMeetings []Meeting
	if err := query.Order("scheduled_at DESC").Limit(req.PageSize).Offset(offset).Find(&dbMeetings).Error; err != nil {
		return nil, fmt.Errorf("failed to list meetings: %w", err)
	}

	meetingIDs := make([]uuid.UUID, len(dbMeetings))
	meetingsMap := make(map[uuid.UUID]*models.MeetingWithDetails)

	for i, dbMeeting := range dbMeetings {
		// No parsing needed - fields are already uuid.UUID
		meetingID := dbMeeting.ID
		subjectID := dbMeeting.SubjectID
		createdBy := dbMeeting.CreatedBy
		liveKitRoomID := dbMeeting.LiveKitRoomID

		meeting := models.Meeting{
			ID:                 meetingID,
			Title:              dbMeeting.Title,
			ScheduledAt:        dbMeeting.ScheduledAt,
			Duration:           dbMeeting.Duration,
			Recurrence:         models.MeetingRecurrence(dbMeeting.Recurrence),
			Type:               models.MeetingType(dbMeeting.Type),
			SubjectID:          subjectID,
			Status:             models.MeetingStatus(dbMeeting.Status),
			NeedsRecord:        dbMeeting.NeedsRecord,
			NeedsTranscription: dbMeeting.NeedsTranscription,
			IsRecording:        dbMeeting.IsRecording,
			IsTranscribing:     dbMeeting.IsTranscribing,
			IsPermanent:        dbMeeting.IsPermanent,
			AllowAnonymous:     dbMeeting.AllowAnonymous,
			AdditionalNotes:    dbMeeting.AdditionalNotes,
			LiveKitRoomID:      liveKitRoomID,
			CreatedBy:          createdBy,
			CreatedAt:          dbMeeting.CreatedAt,
			UpdatedAt:          dbMeeting.UpdatedAt,
		}

		meetingIDs[i] = dbMeeting.ID
		meetingsMap[dbMeeting.ID] = &models.MeetingWithDetails{
			Meeting:      meeting,
			Participants: []models.MeetingParticipantInfo{},
			Departments:  []models.Department{},
		}
	}

	// Load related data for all meetings
	if len(meetingIDs) > 0 {
		// Load subjects
		r.loadMeetingSubjects(meetingsMap, meetingIDs)

		// Load participants
		r.loadMeetingParticipants(meetingsMap, meetingIDs)

		// Load departments
		r.loadMeetingDepartments(meetingsMap, meetingIDs)

		// Load creators
		r.loadMeetingCreators(meetingsMap, meetingIDs)

		// Load participant counts
		r.loadParticipantCounts(meetingsMap, meetingIDs)

		// Load recordings counts
		r.loadRecordingsCounts(meetingsMap, meetingIDs)
	}

	// Convert map to slice
	meetings := make([]models.MeetingWithDetails, len(meetingIDs))
	for i, id := range meetingIDs {
		meetings[i] = *meetingsMap[id]
	}

	return &models.MeetingsResponse{
		Items:    meetings,
		Offset:   offset,
		PageSize: req.PageSize,
		Total:    int(total),
	}, nil
}

// Helper methods for loading related data
func (r *MeetingRepository) loadMeetingSubjects(meetingsMap map[uuid.UUID]*models.MeetingWithDetails, meetingIDs []uuid.UUID) {
	subjectIDs := []string{}
	for _, meeting := range meetingsMap {
		// Skip meetings without a subject
		if meeting.SubjectID != nil {
			subjectIDs = append(subjectIDs, meeting.SubjectID.String())
		}
	}

	if len(subjectIDs) == 0 {
		return
	}

	var dbSubjects []MeetingSubject
	if err := r.db.DB.Where("id IN ?", subjectIDs).Find(&dbSubjects).Error; err != nil {
		return
	}

	subjectsMap := make(map[uuid.UUID]*models.MeetingSubject)
	for _, dbSubject := range dbSubjects {
		subjectID := dbSubject.ID

		departmentIDs, parseErr := stringSliceToUUIDSlice(dbSubject.DepartmentIDs)
		if parseErr != nil {
			// Log error and skip this subject
			continue
		}

		subject := &models.MeetingSubject{
			ID:            subjectID,
			Name:          dbSubject.Name,
			Description:   dbSubject.Description,
			DepartmentIDs: departmentIDs,
			IsActive:      dbSubject.IsActive,
			CreatedAt:     dbSubject.CreatedAt,
			UpdatedAt:     dbSubject.UpdatedAt,
		}
		subjectsMap[dbSubject.ID] = subject
	}

	for _, meeting := range meetingsMap {
		if meeting.SubjectID != nil {
			if subject, ok := subjectsMap[*meeting.SubjectID]; ok {
				meeting.Subject = subject
			}
		}
	}
}

func (r *MeetingRepository) loadMeetingParticipants(meetingsMap map[uuid.UUID]*models.MeetingWithDetails, meetingIDs []uuid.UUID) {
	type ParticipantWithUser struct {
		MeetingParticipant
		Username     string  `gorm:"column:username"`
		Email        string  `gorm:"column:email"`
		FirstName    *string `gorm:"column:first_name"`
		LastName     *string `gorm:"column:last_name"`
		AvatarURL    *string `gorm:"column:avatar_url"`
		UserRole     string  `gorm:"column:user_role"`
		DepartmentID *string `gorm:"column:department_id"`
	}

	var results []ParticipantWithUser
	if err := r.db.DB.Table("meeting_participants mp").
		Select("mp.id, mp.meeting_id, mp.user_id, mp.role, mp.status, mp.joined_at, mp.left_at, mp.created_at, u.username, u.email, u.first_name, u.last_name, u.avatar_url, u.role as user_role, u.department_id").
		Joins("JOIN users u ON mp.user_id = u.id").
		Where("mp.meeting_id IN ?", meetingIDs).
		Order("mp.role DESC, u.username ASC").
		Find(&results).Error; err != nil {
		return
	}

	for _, result := range results {
		partID := result.ID

		meetingID := result.MeetingID

		userID := result.UserID

		var deptID *uuid.UUID
		if result.DepartmentID != nil {
			did, err := uuid.Parse(*result.DepartmentID)
			if err == nil {
				deptID = &did
			}
		}

		userInfo := &models.UserInfo{
			ID:           userID,
			Username:     result.Username,
			Email:        result.Email,
			Role:         models.UserRole(result.UserRole),
			DepartmentID: deptID,
		}

		// Add optional fields
		if result.FirstName != nil {
			userInfo.FirstName = *result.FirstName
		}
		if result.LastName != nil {
			userInfo.LastName = *result.LastName
		}
		if result.AvatarURL != nil {
			userInfo.Avatar = *result.AvatarURL
		}

		participant := models.MeetingParticipantInfo{
			MeetingParticipant: models.MeetingParticipant{
				ID:        partID,
				MeetingID: meetingID,
				UserID:    userID,
				Role:      result.Role,
				Status:    result.Status,
				JoinedAt:  result.JoinedAt,
				LeftAt:    result.LeftAt,
				CreatedAt: result.CreatedAt,
			},
			User: userInfo,
		}

		if meeting, ok := meetingsMap[result.MeetingID]; ok {
			meeting.Participants = append(meeting.Participants, participant)
		}
	}
}

func (r *MeetingRepository) loadMeetingDepartments(meetingsMap map[uuid.UUID]*models.MeetingWithDetails, meetingIDs []uuid.UUID) {
	type MeetingDeptWithInfo struct {
		MeetingID   uuid.UUID  `gorm:"column:meeting_id"`
		ID          uuid.UUID  `gorm:"column:id"`
		Name        string  `gorm:"column:name"`
		Description string  `gorm:"column:description"`
		ParentID    *string `gorm:"column:parent_id"`
		Level       int     `gorm:"column:level"`
		Path        string  `gorm:"column:path"`
		IsActive    bool    `gorm:"column:is_active"`
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}

	var results []MeetingDeptWithInfo
	if err := r.db.DB.Table("meeting_departments md").
		Select("md.meeting_id, d.id, d.name, d.description, d.parent_id, d.level, d.path, d.is_active, d.created_at, d.updated_at").
		Joins("JOIN departments d ON md.department_id = d.id").
		Where("md.meeting_id IN ?", meetingIDs).
		Order("d.name ASC").
		Find(&results).Error; err != nil {
		return
	}

	for _, result := range results {
		deptID := result.ID

		var parentID *uuid.UUID
		if result.ParentID != nil {
			pid, err := uuid.Parse(*result.ParentID)
			if err == nil {
				parentID = &pid
			}
		}

		dept := models.Department{
			ID:          deptID,
			Name:        result.Name,
			Description: result.Description,
			ParentID:    parentID,
			Level:       result.Level,
			Path:        result.Path,
			IsActive:    result.IsActive,
			CreatedAt:   result.CreatedAt,
			UpdatedAt:   result.UpdatedAt,
		}

		if meeting, ok := meetingsMap[result.MeetingID]; ok {
			meeting.Departments = append(meeting.Departments, dept)
		}
	}
}

func (r *MeetingRepository) loadMeetingCreators(meetingsMap map[uuid.UUID]*models.MeetingWithDetails, meetingIDs []uuid.UUID) {
	creatorIDs := []string{}
	for _, meeting := range meetingsMap {
		creatorIDs = append(creatorIDs, meeting.CreatedBy.String())
	}

	if len(creatorIDs) == 0 {
		return
	}

	var dbUsers []User
	if err := r.db.DB.Select("id, username, email, role, department_id").
		Where("id IN ?", creatorIDs).
		Find(&dbUsers).Error; err != nil {
		return
	}

	creatorsMap := make(map[uuid.UUID]*models.UserInfo)
	for _, dbUser := range dbUsers {
		userID := dbUser.ID

		var deptID *uuid.UUID
		if dbUser.DepartmentID != nil {
			deptID = dbUser.DepartmentID
		}

		user := &models.UserInfo{
			ID:           userID,
			Username:     dbUser.Username,
			Email:        dbUser.Email,
			Role:         models.UserRole(dbUser.Role),
			DepartmentID: deptID,
		}
		creatorsMap[dbUser.ID] = user
	}

	for _, meeting := range meetingsMap {
		if creator, ok := creatorsMap[meeting.CreatedBy]; ok {
			meeting.CreatedByUser = creator
		}
	}
}

func (r *MeetingRepository) loadParticipantCounts(meetingsMap map[uuid.UUID]*models.MeetingWithDetails, meetingIDs []uuid.UUID) {
	livekitRepo := NewLiveKitRepository(r.db)

	for meetingID, meeting := range meetingsMap {
		// Count active participants in LiveKit room
		if meeting.LiveKitRoomID != nil {
			// Get the room by meeting ID (room name = meeting ID)
			rooms, err := livekitRepo.GetRoomsByName(meetingID.String())
			if err == nil && len(rooms) > 0 {
				// Use the most recent room
				room := rooms[0]

				// Count participants that are ACTIVE and NOT hidden (not bots/recorders)
				// hidden = true means it's a bot/recorder that should not be counted
				// Используем Unscoped() т.к. у модели Participant нет soft delete
				// Use Unscoped() because Participant model doesn't have soft delete
				var activeCount int64
				err = r.db.DB.Model(&LiveKitParticipant{}).Unscoped().
					Where("room_sid = ? AND state = ? AND (permission->>'hidden' IS NULL OR permission->>'hidden' = 'false')",
						room.SID, "ACTIVE").
					Count(&activeCount).Error
				if err == nil {
					meeting.ActiveParticipantsCount = int(activeCount)
				}

				// Count anonymous guests - participants in LiveKit who are in temporary_users table
				// and are ACTIVE and not hidden
				if meeting.AllowAnonymous {
					var guestCount int64
					err = r.db.DB.Table("livekit_participants").
						Joins("INNER JOIN temporary_users ON livekit_participants.identity = temporary_users.id::text").
						Where("livekit_participants.room_sid = ? AND livekit_participants.state = ? AND (livekit_participants.permission->>'hidden' IS NULL OR livekit_participants.permission->>'hidden' = 'false')",
							room.SID, "ACTIVE").
						Where("temporary_users.meeting_id = ?", meetingID).
						Count(&guestCount).Error
					if err == nil {
						meeting.AnonymousGuestsCount = int(guestCount)
					}
				}
			}
		}
	}
}

// loadRecordingsCounts loads the number of recording rooms for each meeting
func (r *MeetingRepository) loadRecordingsCounts(meetingsMap map[uuid.UUID]*models.MeetingWithDetails, meetingIDs []uuid.UUID) {
	livekitRepo := NewLiveKitRepository(r.db)

	for meetingID, meeting := range meetingsMap {
		// Count rooms by meeting ID (room name = meeting ID)
		rooms, err := livekitRepo.GetRoomsByName(meetingID.String())
		if err == nil {
			meeting.RecordingsCount = len(rooms)
		}
	}
}

// UpdateMeeting updates a meeting
func (r *MeetingRepository) UpdateMeeting(meeting *models.Meeting) error {
	meeting.UpdatedAt = time.Now()

	var liveKitRoomID *string
	if meeting.LiveKitRoomID != nil {
		roomIDStr := meeting.LiveKitRoomID.String()
		liveKitRoomID = &roomIDStr
	}

	var subjectID *string
	if meeting.SubjectID != nil {
		subjectIDStr := meeting.SubjectID.String()
		subjectID = &subjectIDStr
	}

	result := r.db.DB.Model(&Meeting{}).Where("id = ?", meeting.ID.String()).Updates(map[string]interface{}{
		"title":                 meeting.Title,
		"scheduled_at":          meeting.ScheduledAt,
		"duration":              meeting.Duration,
		"recurrence":            string(meeting.Recurrence),
		"type":                  string(meeting.Type),
		"subject_id":            subjectID,
		"status":                string(meeting.Status),
		"needs_record":          meeting.NeedsRecord,
		"needs_transcription":   meeting.NeedsTranscription,
		"is_recording":          meeting.IsRecording,
		"is_transcribing":       meeting.IsTranscribing,
		"is_permanent":          meeting.IsPermanent,
		"allow_anonymous":       meeting.AllowAnonymous,
		"additional_notes":      meeting.AdditionalNotes,
		"livekit_room_id":       liveKitRoomID,
		"updated_at":            meeting.UpdatedAt,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to update meeting: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("meeting not found")
	}

	return nil
}

// DeleteMeeting deletes a meeting
func (r *MeetingRepository) DeleteMeeting(id uuid.UUID) error {
	result := r.db.DB.Where("id = ?", id).Delete(&Meeting{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete meeting: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("meeting not found")
	}

	return nil
}

// ============= Meeting Participants =============

// AddParticipant adds a participant to a meeting
func (r *MeetingRepository) AddParticipant(participant *models.MeetingParticipant) error {
	dbParticipant := &MeetingParticipant{
		ID:        participant.ID,
		MeetingID: participant.MeetingID,
		UserID:    participant.UserID,
		Role:      participant.Role,
		Status:    participant.Status,
		CreatedAt: participant.CreatedAt,
	}

	// Use GORM Clauses for ON CONFLICT
	if err := r.db.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "meeting_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"role", "status"}),
	}).Create(dbParticipant).Error; err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	return nil
}

// RemoveParticipant removes a participant from a meeting
func (r *MeetingRepository) RemoveParticipant(meetingID, userID uuid.UUID) error {
	result := r.db.DB.Where("meeting_id = ? AND user_id = ?", meetingID, userID).Delete(&MeetingParticipant{})

	if result.Error != nil {
		return fmt.Errorf("failed to remove participant: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("participant not found")
	}

	return nil
}

// GetMeetingParticipants retrieves all participants of a meeting with user info
func (r *MeetingRepository) GetMeetingParticipants(meetingID uuid.UUID) ([]models.MeetingParticipantInfo, error) {
	type ParticipantWithUser struct {
		MeetingParticipant
		Username     string  `gorm:"column:username"`
		Email        string  `gorm:"column:email"`
		FirstName    *string `gorm:"column:first_name"`
		LastName     *string `gorm:"column:last_name"`
		AvatarURL    *string `gorm:"column:avatar_url"`
		UserRole     string  `gorm:"column:user_role"`
		DepartmentID *string `gorm:"column:department_id"`
	}

	var results []ParticipantWithUser
	if err := r.db.DB.Table("meeting_participants mp").
		Select("mp.id, mp.meeting_id, mp.user_id, mp.role, mp.status, mp.joined_at, mp.left_at, mp.created_at, u.username, u.email, u.first_name, u.last_name, u.avatar_url, u.role as user_role, u.department_id").
		Joins("JOIN users u ON mp.user_id = u.id").
		Where("mp.meeting_id = ?", meetingID).
		Order("mp.role DESC, u.username ASC").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}

	participants := make([]models.MeetingParticipantInfo, 0, len(results))
	for _, result := range results {
		partID := result.ID

		meetID := result.MeetingID

		userID := result.UserID

		var deptID *uuid.UUID
		if result.DepartmentID != nil {
			did, err := uuid.Parse(*result.DepartmentID)
			if err == nil {
				deptID = &did
			}
		}

		userInfo := &models.UserInfo{
			ID:           userID,
			Username:     result.Username,
			Email:        result.Email,
			Role:         models.UserRole(result.UserRole),
			DepartmentID: deptID,
		}

		// Add optional fields
		if result.FirstName != nil {
			userInfo.FirstName = *result.FirstName
		}
		if result.LastName != nil {
			userInfo.LastName = *result.LastName
		}
		if result.AvatarURL != nil {
			userInfo.Avatar = *result.AvatarURL
		}

		participants = append(participants, models.MeetingParticipantInfo{
			MeetingParticipant: models.MeetingParticipant{
				ID:        partID,
				MeetingID: meetID,
				UserID:    userID,
				Role:      result.Role,
				Status:    result.Status,
				JoinedAt:  result.JoinedAt,
				LeftAt:    result.LeftAt,
				CreatedAt: result.CreatedAt,
			},
			User: userInfo,
		})
	}

	return participants, nil
}

// ============= Meeting Departments =============

// AddDepartment adds a department to a meeting
func (r *MeetingRepository) AddDepartment(meetingDept *models.MeetingDepartment) error {
	dbMeetingDept := &MeetingDepartment{
		ID:           meetingDept.ID,
		MeetingID:    meetingDept.MeetingID,
		DepartmentID: meetingDept.DepartmentID,
		CreatedAt:    meetingDept.CreatedAt,
	}

	// Use GORM Clauses for ON CONFLICT DO NOTHING
	if err := r.db.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "meeting_id"}, {Name: "department_id"}},
		DoNothing: true,
	}).Create(dbMeetingDept).Error; err != nil {
		return fmt.Errorf("failed to add department: %w", err)
	}

	return nil
}

// RemoveDepartment removes a department from a meeting
func (r *MeetingRepository) RemoveDepartment(meetingID, departmentID uuid.UUID) error {
	result := r.db.DB.Where("meeting_id = ? AND department_id = ?", meetingID, departmentID).Delete(&MeetingDepartment{})

	if result.Error != nil {
		return fmt.Errorf("failed to remove department: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("department not found in meeting")
	}

	return nil
}

// GetMeetingDepartments retrieves all departments invited to a meeting
func (r *MeetingRepository) GetMeetingDepartments(meetingID uuid.UUID) ([]models.Department, error) {
	type DeptInfo struct {
		ID          uuid.UUID  `gorm:"column:id"`
		Name        string     `gorm:"column:name"`
		Description string     `gorm:"column:description"`
		ParentID    *uuid.UUID `gorm:"column:parent_id"`
		Level       int        `gorm:"column:level"`
		Path        string     `gorm:"column:path"`
		IsActive    bool       `gorm:"column:is_active"`
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}

	var results []DeptInfo
	if err := r.db.DB.Table("meeting_departments md").
		Select("d.id, d.name, d.description, d.parent_id, d.level, d.path, d.is_active, d.created_at, d.updated_at").
		Joins("JOIN departments d ON md.department_id = d.id").
		Where("md.meeting_id = ?", meetingID).
		Order("d.name ASC").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get departments: %w", err)
	}

	departments := make([]models.Department, 0, len(results))
	for _, result := range results {
		departments = append(departments, models.Department{
			ID:          result.ID,
			Name:        result.Name,
			Description: result.Description,
			ParentID:    result.ParentID,
			Level:       result.Level,
			Path:        result.Path,
			IsActive:    result.IsActive,
			CreatedAt:   result.CreatedAt,
			UpdatedAt:   result.UpdatedAt,
		})
	}

	return departments, nil
}

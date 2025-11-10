package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"Recontext.online/internal/models"
	"github.com/lib/pq"
)

// MeetingRepository handles database operations for meetings
type MeetingRepository struct {
	db *DB
}

// NewMeetingRepository creates a new MeetingRepository
func NewMeetingRepository(db *DB) *MeetingRepository {
	return &MeetingRepository{db: db}
}

// ============= Meeting Subjects =============

// CreateSubject creates a new meeting subject
func (r *MeetingRepository) CreateSubject(subject *models.MeetingSubject) error {
	query := `
		INSERT INTO meeting_subjects (id, name, description, department_ids, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(
		query,
		subject.ID,
		subject.Name,
		subject.Description,
		pq.Array(subject.DepartmentIDs),
		subject.IsActive,
		subject.CreatedAt,
		subject.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create meeting subject: %w", err)
	}

	return nil
}

// GetSubjectByID retrieves a meeting subject by ID
func (r *MeetingRepository) GetSubjectByID(id string) (*models.MeetingSubject, error) {
	query := `
		SELECT id, name, description, department_ids, is_active, created_at, updated_at
		FROM meeting_subjects
		WHERE id = $1
	`

	subject := &models.MeetingSubject{}
	var departmentIDs pq.StringArray

	err := r.db.QueryRow(query, id).Scan(
		&subject.ID,
		&subject.Name,
		&subject.Description,
		&departmentIDs,
		&subject.IsActive,
		&subject.CreatedAt,
		&subject.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("meeting subject not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get meeting subject: %w", err)
	}

	subject.DepartmentIDs = departmentIDs

	return subject, nil
}

// ListSubjects retrieves meeting subjects with pagination
func (r *MeetingRepository) ListSubjects(page, pageSize int, departmentID *string, includeInactive bool) (*models.MeetingSubjectsResponse, error) {
	offset := (page - 1) * pageSize

	// Build WHERE conditions
	conditions := []string{}
	args := []interface{}{}
	argIdx := 1

	if !includeInactive {
		conditions = append(conditions, "is_active = true")
	}

	if departmentID != nil && *departmentID != "" {
		conditions = append(conditions, fmt.Sprintf("$%d = ANY(department_ids)", argIdx))
		args = append(args, *departmentID)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM meeting_subjects %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count subjects: %w", err)
	}

	// Get page data
	query := fmt.Sprintf(`
		SELECT id, name, description, department_ids, is_active, created_at, updated_at
		FROM meeting_subjects
		%s
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list subjects: %w", err)
	}
	defer rows.Close()

	subjects := []models.MeetingSubject{}
	for rows.Next() {
		subject := models.MeetingSubject{}
		var departmentIDs pq.StringArray

		err := rows.Scan(
			&subject.ID,
			&subject.Name,
			&subject.Description,
			&departmentIDs,
			&subject.IsActive,
			&subject.CreatedAt,
			&subject.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subject: %w", err)
		}

		subject.DepartmentIDs = departmentIDs
		subjects = append(subjects, subject)
	}

	return &models.MeetingSubjectsResponse{
		Items:    subjects,
		Offset:   offset,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

// UpdateSubject updates a meeting subject
func (r *MeetingRepository) UpdateSubject(subject *models.MeetingSubject) error {
	subject.UpdatedAt = time.Now()

	query := `
		UPDATE meeting_subjects
		SET name = $2, description = $3, department_ids = $4, is_active = $5, updated_at = $6
		WHERE id = $1
	`

	result, err := r.db.Exec(
		query,
		subject.ID,
		subject.Name,
		subject.Description,
		pq.Array(subject.DepartmentIDs),
		subject.IsActive,
		subject.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update subject: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("meeting subject not found")
	}

	return nil
}

// DeleteSubject soft deletes a meeting subject
func (r *MeetingRepository) DeleteSubject(id string) error {
	query := `
		UPDATE meeting_subjects
		SET is_active = false, updated_at = $2
		WHERE id = $1
	`

	result, err := r.db.Exec(query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete subject: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("meeting subject not found")
	}

	return nil
}

// ============= Meetings =============

// CreateMeeting creates a new meeting
func (r *MeetingRepository) CreateMeeting(meeting *models.Meeting) error {
	query := `
		INSERT INTO meetings (
			id, title, scheduled_at, duration, recurrence, type, subject_id, status,
			needs_video_record, needs_audio_record, additional_notes, livekit_room_id,
			created_by, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := r.db.Exec(
		query,
		meeting.ID,
		meeting.Title,
		meeting.ScheduledAt,
		meeting.Duration,
		meeting.Recurrence,
		meeting.Type,
		meeting.SubjectID,
		meeting.Status,
		meeting.NeedsVideoRecord,
		meeting.NeedsAudioRecord,
		meeting.AdditionalNotes,
		meeting.LiveKitRoomID,
		meeting.CreatedBy,
		meeting.CreatedAt,
		meeting.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create meeting: %w", err)
	}

	return nil
}

// GetMeetingByID retrieves a meeting by ID
func (r *MeetingRepository) GetMeetingByID(id string) (*models.Meeting, error) {
	query := `
		SELECT id, title, scheduled_at, duration, recurrence, type, subject_id, status,
			   needs_video_record, needs_audio_record, additional_notes, livekit_room_id,
			   created_by, created_at, updated_at
		FROM meetings
		WHERE id = $1
	`

	meeting := &models.Meeting{}
	var liveKitRoomID sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&meeting.ID,
		&meeting.Title,
		&meeting.ScheduledAt,
		&meeting.Duration,
		&meeting.Recurrence,
		&meeting.Type,
		&meeting.SubjectID,
		&meeting.Status,
		&meeting.NeedsVideoRecord,
		&meeting.NeedsAudioRecord,
		&meeting.AdditionalNotes,
		&liveKitRoomID,
		&meeting.CreatedBy,
		&meeting.CreatedAt,
		&meeting.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("meeting not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get meeting: %w", err)
	}

	if liveKitRoomID.Valid {
		meeting.LiveKitRoomID = &liveKitRoomID.String
	}

	return meeting, nil
}

// GetMeetingWithDetails retrieves a meeting with all related information
func (r *MeetingRepository) GetMeetingWithDetails(id string) (*models.MeetingWithDetails, error) {
	meeting, err := r.GetMeetingByID(id)
	if err != nil {
		return nil, err
	}

	details := &models.MeetingWithDetails{
		Meeting: *meeting,
	}

	// Get subject
	subject, err := r.GetSubjectByID(meeting.SubjectID)
	if err == nil {
		details.Subject = subject
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

	// Build WHERE conditions
	conditions := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Status != nil {
		conditions = append(conditions, fmt.Sprintf("m.status = $%d", argIdx))
		args = append(args, *req.Status)
		argIdx++
	}

	if req.Type != nil {
		conditions = append(conditions, fmt.Sprintf("m.type = $%d", argIdx))
		args = append(args, *req.Type)
		argIdx++
	}

	if req.SubjectID != nil {
		conditions = append(conditions, fmt.Sprintf("m.subject_id = $%d", argIdx))
		args = append(args, *req.SubjectID)
		argIdx++
	}

	if req.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("m.scheduled_at >= $%d", argIdx))
		args = append(args, *req.DateFrom)
		argIdx++
	}

	if req.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("m.scheduled_at <= $%d", argIdx))
		args = append(args, *req.DateTo)
		argIdx++
	}

	// Filter by speaker or participant
	if req.SpeakerID != nil || req.UserID != nil {
		userID := req.SpeakerID
		if req.UserID != nil {
			userID = req.UserID
		}

		conditions = append(conditions, fmt.Sprintf(`
			m.id IN (
				SELECT meeting_id FROM meeting_participants WHERE user_id = $%d
			)
		`, argIdx))
		args = append(args, *userID)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM meetings m %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count meetings: %w", err)
	}

	// Get page data
	query := fmt.Sprintf(`
		SELECT m.id, m.title, m.scheduled_at, m.duration, m.recurrence, m.type, m.subject_id,
			   m.status, m.needs_video_record, m.needs_audio_record, m.additional_notes,
			   m.livekit_room_id, m.created_by, m.created_at, m.updated_at
		FROM meetings m
		%s
		ORDER BY m.scheduled_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, req.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list meetings: %w", err)
	}
	defer rows.Close()

	meetingIDs := []string{}
	meetingsMap := make(map[string]*models.MeetingWithDetails)

	for rows.Next() {
		meeting := models.Meeting{}
		var liveKitRoomID sql.NullString

		err := rows.Scan(
			&meeting.ID,
			&meeting.Title,
			&meeting.ScheduledAt,
			&meeting.Duration,
			&meeting.Recurrence,
			&meeting.Type,
			&meeting.SubjectID,
			&meeting.Status,
			&meeting.NeedsVideoRecord,
			&meeting.NeedsAudioRecord,
			&meeting.AdditionalNotes,
			&liveKitRoomID,
			&meeting.CreatedBy,
			&meeting.CreatedAt,
			&meeting.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan meeting: %w", err)
		}

		if liveKitRoomID.Valid {
			meeting.LiveKitRoomID = &liveKitRoomID.String
		}

		meetingIDs = append(meetingIDs, meeting.ID)
		meetingsMap[meeting.ID] = &models.MeetingWithDetails{
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
	}

	// Convert map to slice
	meetings := []models.MeetingWithDetails{}
	for _, id := range meetingIDs {
		meetings = append(meetings, *meetingsMap[id])
	}

	return &models.MeetingsResponse{
		Items:    meetings,
		Offset:   offset,
		PageSize: req.PageSize,
		Total:    total,
	}, nil
}

// Helper methods for loading related data
func (r *MeetingRepository) loadMeetingSubjects(meetingsMap map[string]*models.MeetingWithDetails, meetingIDs []string) {
	subjectIDs := []string{}
	for _, meeting := range meetingsMap {
		subjectIDs = append(subjectIDs, meeting.SubjectID)
	}

	if len(subjectIDs) == 0 {
		return
	}

	query := `
		SELECT id, name, description, department_ids, is_active, created_at, updated_at
		FROM meeting_subjects
		WHERE id = ANY($1)
	`

	rows, err := r.db.Query(query, pq.Array(subjectIDs))
	if err != nil {
		return
	}
	defer rows.Close()

	subjectsMap := make(map[string]*models.MeetingSubject)
	for rows.Next() {
		subject := &models.MeetingSubject{}
		var departmentIDs pq.StringArray

		err := rows.Scan(
			&subject.ID,
			&subject.Name,
			&subject.Description,
			&departmentIDs,
			&subject.IsActive,
			&subject.CreatedAt,
			&subject.UpdatedAt,
		)
		if err == nil {
			subject.DepartmentIDs = departmentIDs
			subjectsMap[subject.ID] = subject
		}
	}

	for _, meeting := range meetingsMap {
		if subject, ok := subjectsMap[meeting.SubjectID]; ok {
			meeting.Subject = subject
		}
	}
}

func (r *MeetingRepository) loadMeetingParticipants(meetingsMap map[string]*models.MeetingWithDetails, meetingIDs []string) {
	query := `
		SELECT mp.id, mp.meeting_id, mp.user_id, mp.role, mp.status, mp.joined_at, mp.left_at, mp.created_at,
			   u.id, u.username, u.email, u.role, u.department_id
		FROM meeting_participants mp
		JOIN users u ON mp.user_id = u.id
		WHERE mp.meeting_id = ANY($1)
		ORDER BY mp.role DESC, u.username ASC
	`

	rows, err := r.db.Query(query, pq.Array(meetingIDs))
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		participant := models.MeetingParticipantInfo{
			User: &models.UserInfo{},
		}
		var joinedAt, leftAt sql.NullTime
		var departmentID sql.NullString

		err := rows.Scan(
			&participant.ID,
			&participant.MeetingID,
			&participant.UserID,
			&participant.Role,
			&participant.Status,
			&joinedAt,
			&leftAt,
			&participant.CreatedAt,
			&participant.User.ID,
			&participant.User.Username,
			&participant.User.Email,
			&participant.User.Role,
			&departmentID,
		)
		if err != nil {
			continue
		}

		if joinedAt.Valid {
			participant.JoinedAt = &joinedAt.Time
		}
		if leftAt.Valid {
			participant.LeftAt = &leftAt.Time
		}
		if departmentID.Valid {
			participant.User.DepartmentID = &departmentID.String
		}

		if meeting, ok := meetingsMap[participant.MeetingID]; ok {
			meeting.Participants = append(meeting.Participants, participant)
		}
	}
}

func (r *MeetingRepository) loadMeetingDepartments(meetingsMap map[string]*models.MeetingWithDetails, meetingIDs []string) {
	query := `
		SELECT md.meeting_id, d.id, d.name, d.description, d.parent_id, d.level, d.path, d.is_active, d.created_at, d.updated_at
		FROM meeting_departments md
		JOIN departments d ON md.department_id = d.id
		WHERE md.meeting_id = ANY($1)
		ORDER BY d.name ASC
	`

	rows, err := r.db.Query(query, pq.Array(meetingIDs))
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var meetingID string
		dept := models.Department{}
		var parentID sql.NullString

		err := rows.Scan(
			&meetingID,
			&dept.ID,
			&dept.Name,
			&dept.Description,
			&parentID,
			&dept.Level,
			&dept.Path,
			&dept.IsActive,
			&dept.CreatedAt,
			&dept.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if parentID.Valid {
			dept.ParentID = &parentID.String
		}

		if meeting, ok := meetingsMap[meetingID]; ok {
			meeting.Departments = append(meeting.Departments, dept)
		}
	}
}

func (r *MeetingRepository) loadMeetingCreators(meetingsMap map[string]*models.MeetingWithDetails, meetingIDs []string) {
	creatorIDs := []string{}
	for _, meeting := range meetingsMap {
		creatorIDs = append(creatorIDs, meeting.CreatedBy)
	}

	if len(creatorIDs) == 0 {
		return
	}

	query := `
		SELECT id, username, email, role, department_id
		FROM users
		WHERE id = ANY($1)
	`

	rows, err := r.db.Query(query, pq.Array(creatorIDs))
	if err != nil {
		return
	}
	defer rows.Close()

	creatorsMap := make(map[string]*models.UserInfo)
	for rows.Next() {
		user := &models.UserInfo{}
		var departmentID sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Role,
			&departmentID,
		)
		if err == nil {
			if departmentID.Valid {
				user.DepartmentID = &departmentID.String
			}
			creatorsMap[user.ID] = user
		}
	}

	for _, meeting := range meetingsMap {
		if creator, ok := creatorsMap[meeting.CreatedBy]; ok {
			meeting.CreatedByUser = creator
		}
	}
}

// UpdateMeeting updates a meeting
func (r *MeetingRepository) UpdateMeeting(meeting *models.Meeting) error {
	meeting.UpdatedAt = time.Now()

	query := `
		UPDATE meetings
		SET title = $2, scheduled_at = $3, duration = $4, recurrence = $5, type = $6,
		    subject_id = $7, status = $8, needs_video_record = $9, needs_audio_record = $10,
		    additional_notes = $11, livekit_room_id = $12, updated_at = $13
		WHERE id = $1
	`

	result, err := r.db.Exec(
		query,
		meeting.ID,
		meeting.Title,
		meeting.ScheduledAt,
		meeting.Duration,
		meeting.Recurrence,
		meeting.Type,
		meeting.SubjectID,
		meeting.Status,
		meeting.NeedsVideoRecord,
		meeting.NeedsAudioRecord,
		meeting.AdditionalNotes,
		meeting.LiveKitRoomID,
		meeting.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update meeting: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("meeting not found")
	}

	return nil
}

// DeleteMeeting deletes a meeting
func (r *MeetingRepository) DeleteMeeting(id string) error {
	query := `DELETE FROM meetings WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete meeting: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("meeting not found")
	}

	return nil
}

// ============= Meeting Participants =============

// AddParticipant adds a participant to a meeting
func (r *MeetingRepository) AddParticipant(participant *models.MeetingParticipant) error {
	query := `
		INSERT INTO meeting_participants (id, meeting_id, user_id, role, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (meeting_id, user_id) DO UPDATE
		SET role = EXCLUDED.role, status = EXCLUDED.status
	`

	_, err := r.db.Exec(
		query,
		participant.ID,
		participant.MeetingID,
		participant.UserID,
		participant.Role,
		participant.Status,
		participant.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	return nil
}

// RemoveParticipant removes a participant from a meeting
func (r *MeetingRepository) RemoveParticipant(meetingID, userID string) error {
	query := `DELETE FROM meeting_participants WHERE meeting_id = $1 AND user_id = $2`

	result, err := r.db.Exec(query, meetingID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("participant not found")
	}

	return nil
}

// GetMeetingParticipants retrieves all participants of a meeting with user info
func (r *MeetingRepository) GetMeetingParticipants(meetingID string) ([]models.MeetingParticipantInfo, error) {
	query := `
		SELECT mp.id, mp.meeting_id, mp.user_id, mp.role, mp.status, mp.joined_at, mp.left_at, mp.created_at,
			   u.id, u.username, u.email, u.role, u.department_id
		FROM meeting_participants mp
		JOIN users u ON mp.user_id = u.id
		WHERE mp.meeting_id = $1
		ORDER BY mp.role DESC, u.username ASC
	`

	rows, err := r.db.Query(query, meetingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}
	defer rows.Close()

	participants := []models.MeetingParticipantInfo{}
	for rows.Next() {
		participant := models.MeetingParticipantInfo{
			User: &models.UserInfo{},
		}
		var joinedAt, leftAt sql.NullTime
		var departmentID sql.NullString

		err := rows.Scan(
			&participant.ID,
			&participant.MeetingID,
			&participant.UserID,
			&participant.Role,
			&participant.Status,
			&joinedAt,
			&leftAt,
			&participant.CreatedAt,
			&participant.User.ID,
			&participant.User.Username,
			&participant.User.Email,
			&participant.User.Role,
			&departmentID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}

		if joinedAt.Valid {
			participant.JoinedAt = &joinedAt.Time
		}
		if leftAt.Valid {
			participant.LeftAt = &leftAt.Time
		}
		if departmentID.Valid {
			participant.User.DepartmentID = &departmentID.String
		}

		participants = append(participants, participant)
	}

	return participants, nil
}

// ============= Meeting Departments =============

// AddDepartment adds a department to a meeting
func (r *MeetingRepository) AddDepartment(meetingDept *models.MeetingDepartment) error {
	query := `
		INSERT INTO meeting_departments (id, meeting_id, department_id, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (meeting_id, department_id) DO NOTHING
	`

	_, err := r.db.Exec(
		query,
		meetingDept.ID,
		meetingDept.MeetingID,
		meetingDept.DepartmentID,
		meetingDept.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to add department: %w", err)
	}

	return nil
}

// RemoveDepartment removes a department from a meeting
func (r *MeetingRepository) RemoveDepartment(meetingID, departmentID string) error {
	query := `DELETE FROM meeting_departments WHERE meeting_id = $1 AND department_id = $2`

	result, err := r.db.Exec(query, meetingID, departmentID)
	if err != nil {
		return fmt.Errorf("failed to remove department: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("department not found in meeting")
	}

	return nil
}

// GetMeetingDepartments retrieves all departments invited to a meeting
func (r *MeetingRepository) GetMeetingDepartments(meetingID string) ([]models.Department, error) {
	query := `
		SELECT d.id, d.name, d.description, d.parent_id, d.level, d.path, d.is_active, d.created_at, d.updated_at
		FROM meeting_departments md
		JOIN departments d ON md.department_id = d.id
		WHERE md.meeting_id = $1
		ORDER BY d.name ASC
	`

	rows, err := r.db.Query(query, meetingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get departments: %w", err)
	}
	defer rows.Close()

	departments := []models.Department{}
	for rows.Next() {
		dept := models.Department{}
		var parentID sql.NullString

		err := rows.Scan(
			&dept.ID,
			&dept.Name,
			&dept.Description,
			&parentID,
			&dept.Level,
			&dept.Path,
			&dept.IsActive,
			&dept.CreatedAt,
			&dept.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan department: %w", err)
		}

		if parentID.Valid {
			dept.ParentID = &parentID.String
		}

		departments = append(departments, dept)
	}

	return departments, nil
}

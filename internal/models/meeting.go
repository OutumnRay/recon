package models

import "time"

// MeetingType represents the type of meeting
type MeetingType string

const (
	MeetingTypePresentation MeetingType = "presentation" // Доклад
	MeetingTypeConference   MeetingType = "conference"   // Совещание
)

// MeetingStatus represents the status of a meeting
type MeetingStatus string

const (
	MeetingStatusScheduled MeetingStatus = "scheduled" // Запланирована
	MeetingStatusInProgress MeetingStatus = "in_progress" // Идет
	MeetingStatusCompleted MeetingStatus = "completed" // Завершена
	MeetingStatusCancelled MeetingStatus = "cancelled" // Отменена
)

// MeetingRecurrence represents how often a meeting repeats
type MeetingRecurrence string

const (
	MeetingRecurrenceNone    MeetingRecurrence = "none"    // Не повторяется
	MeetingRecurrenceDaily   MeetingRecurrence = "daily"   // Ежедневно
	MeetingRecurrenceWeekly  MeetingRecurrence = "weekly"  // Еженедельно
	MeetingRecurrenceMonthly MeetingRecurrence = "monthly" // Ежемесячно
)

// MeetingSubject represents a meeting subject/topic category
type MeetingSubject struct {
	ID            string    `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	Description   string    `json:"description" db:"description"`
	DepartmentIDs []string  `json:"department_ids" db:"department_ids"` // Departments this subject is linked to
	IsActive      bool      `json:"is_active" db:"is_active"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// Meeting represents a video meeting
type Meeting struct {
	ID                   string            `json:"id" db:"id"`
	Title                string            `json:"title" db:"title"`
	ScheduledAt          time.Time         `json:"scheduled_at" db:"scheduled_at"`
	Duration             int               `json:"duration" db:"duration"` // Duration in minutes
	Recurrence           MeetingRecurrence `json:"recurrence" db:"recurrence"`
	Type                 MeetingType       `json:"type" db:"type"`
	SubjectID            string            `json:"subject_id" db:"subject_id"`
	Status               MeetingStatus     `json:"status" db:"status"`
	NeedsVideoRecord     bool              `json:"needs_video_record" db:"needs_video_record"`
	NeedsAudioRecord     bool              `json:"needs_audio_record" db:"needs_audio_record"`
	AdditionalNotes      string            `json:"additional_notes" db:"additional_notes"`
	ForceEndAtDuration   bool              `json:"force_end_at_duration" db:"force_end_at_duration"` // Force end meeting after duration
	LiveKitRoomID        *string           `json:"livekit_room_id,omitempty" db:"livekit_room_id"`   // Link to LiveKit room if started
	CreatedBy            string            `json:"created_by" db:"created_by"`                        // User ID who created the meeting
	CreatedAt            time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at" db:"updated_at"`
}

// MeetingParticipant represents a participant in a meeting
type MeetingParticipant struct {
	ID          string    `json:"id" db:"id"`
	MeetingID   string    `json:"meeting_id" db:"meeting_id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Role        string    `json:"role" db:"role"` // "speaker", "participant"
	Status      string    `json:"status" db:"status"` // "invited", "accepted", "declined", "attended"
	JoinedAt    *time.Time `json:"joined_at,omitempty" db:"joined_at"`
	LeftAt      *time.Time `json:"left_at,omitempty" db:"left_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// MeetingDepartment represents a department invited to a meeting
type MeetingDepartment struct {
	ID           string    `json:"id" db:"id"`
	MeetingID    string    `json:"meeting_id" db:"meeting_id"`
	DepartmentID string    `json:"department_id" db:"department_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// MeetingWithDetails represents a meeting with all related information
type MeetingWithDetails struct {
	Meeting
	Subject         *MeetingSubject        `json:"subject,omitempty"`
	Participants    []MeetingParticipantInfo `json:"participants"`
	Departments     []Department           `json:"departments"`
	CreatedByUser   *UserInfo              `json:"created_by_user,omitempty"`
}

// MeetingParticipantInfo represents participant information with user details
type MeetingParticipantInfo struct {
	MeetingParticipant
	User *UserInfo `json:"user,omitempty"`
}

// CreateMeetingSubjectRequest represents a request to create a meeting subject
type CreateMeetingSubjectRequest struct {
	Name          string   `json:"name" binding:"required" example:"Техническое обсуждение"`
	Description   string   `json:"description" example:"Обсуждение технических вопросов разработки"`
	DepartmentIDs []string `json:"department_ids" example:"dept-001,dept-002"`
}

// UpdateMeetingSubjectRequest represents a request to update a meeting subject
type UpdateMeetingSubjectRequest struct {
	Name          string   `json:"name" example:"Техническое обсуждение"`
	Description   string   `json:"description" example:"Обновленное описание"`
	DepartmentIDs []string `json:"department_ids" example:"dept-001,dept-002"`
	IsActive      *bool    `json:"is_active" example:"true"`
}

// CreateMeetingRequest represents a request to create a meeting
type CreateMeetingRequest struct {
	Title              string            `json:"title" binding:"required" example:"Еженедельное совещание"`
	ScheduledAt        time.Time         `json:"scheduled_at" binding:"required" example:"2025-01-15T10:00:00Z"`
	Duration           int               `json:"duration" binding:"required" example:"60"`
	Recurrence         MeetingRecurrence `json:"recurrence" example:"weekly"`
	Type               MeetingType       `json:"type" binding:"required" example:"conference"`
	SubjectID          string            `json:"subject_id" binding:"required" example:"subj-001"`
	NeedsVideoRecord   bool              `json:"needs_video_record" example:"true"`
	NeedsAudioRecord   bool              `json:"needs_audio_record" example:"true"`
	AdditionalNotes    string            `json:"additional_notes" example:"Подготовить отчеты"`
	ForceEndAtDuration bool              `json:"force_end_at_duration" example:"true"` // Force end meeting after duration
	SpeakerID          *string           `json:"speaker_id,omitempty" example:"user-001"` // For presentations
	ParticipantIDs     []string          `json:"participant_ids" example:"user-002,user-003"`
	DepartmentIDs      []string          `json:"department_ids" example:"dept-001,dept-002"`
}

// UpdateMeetingRequest represents a request to update a meeting
type UpdateMeetingRequest struct {
	Title              *string            `json:"title,omitempty" example:"Обновленное название"`
	ScheduledAt        *time.Time         `json:"scheduled_at,omitempty" example:"2025-01-15T10:00:00Z"`
	Duration           *int               `json:"duration,omitempty" example:"90"`
	Recurrence         *MeetingRecurrence `json:"recurrence,omitempty" example:"weekly"`
	Type               *MeetingType       `json:"type,omitempty" example:"conference"`
	SubjectID          *string            `json:"subject_id,omitempty" example:"subj-001"`
	Status             *MeetingStatus     `json:"status,omitempty" example:"in_progress"`
	NeedsVideoRecord   *bool              `json:"needs_video_record,omitempty" example:"true"`
	NeedsAudioRecord   *bool              `json:"needs_audio_record,omitempty" example:"true"`
	AdditionalNotes    *string            `json:"additional_notes,omitempty" example:"Обновленные комментарии"`
	ForceEndAtDuration *bool              `json:"force_end_at_duration,omitempty" example:"true"`
	SpeakerID          *string            `json:"speaker_id,omitempty" example:"user-001"`
	ParticipantIDs     []string           `json:"participant_ids,omitempty" example:"user-002,user-003"`
	DepartmentIDs      []string           `json:"department_ids,omitempty" example:"dept-001"`
}

// ListMeetingsRequest represents parameters for listing meetings
type ListMeetingsRequest struct {
	Page         int            `json:"page" form:"page" example:"1"`
	PageSize     int            `json:"page_size" form:"page_size" example:"20"`
	Status       *MeetingStatus `json:"status" form:"status" example:"scheduled"`
	Type         *MeetingType   `json:"type" form:"type" example:"conference"`
	SubjectID    *string        `json:"subject_id" form:"subject_id" example:"subj-001"`
	SpeakerID    *string        `json:"speaker_id" form:"speaker_id" example:"user-001"`
	DateFrom     *time.Time     `json:"date_from" form:"date_from" example:"2025-01-01T00:00:00Z"`
	DateTo       *time.Time     `json:"date_to" form:"date_to" example:"2025-12-31T23:59:59Z"`
	UserID       *string        `json:"user_id" form:"user_id"` // Filter by participant or speaker
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Items    interface{} `json:"items"`
	Offset   int         `json:"offset"`
	PageSize int         `json:"page_size"`
	Total    int         `json:"total"`
}

// MeetingsResponse represents a paginated list of meetings
type MeetingsResponse struct {
	Items    []MeetingWithDetails `json:"items"`
	Offset   int                  `json:"offset"`
	PageSize int                  `json:"page_size"`
	Total    int                  `json:"total"`
}

// MeetingSubjectsResponse represents a paginated list of meeting subjects
type MeetingSubjectsResponse struct {
	Items    []MeetingSubject `json:"items"`
	Offset   int              `json:"offset"`
	PageSize int              `json:"page_size"`
	Total    int              `json:"total"`
}

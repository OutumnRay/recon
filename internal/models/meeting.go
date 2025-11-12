package models

import (
	"time"

	"github.com/google/uuid"
)

// MeetingType представляет тип встречи
type MeetingType string

const (
	MeetingTypePresentation MeetingType = "presentation" // Доклад
	MeetingTypeConference   MeetingType = "conference"   // Совещание
)

// MeetingStatus представляет статус встречи
type MeetingStatus string

const (
	MeetingStatusScheduled MeetingStatus = "scheduled" // Запланирована
	MeetingStatusInProgress MeetingStatus = "in_progress" // Идет
	MeetingStatusCompleted MeetingStatus = "completed" // Завершена
	MeetingStatusCancelled MeetingStatus = "cancelled" // Отменена
)

// MeetingRecurrence представляет частоту повторения встречи
type MeetingRecurrence string

const (
	MeetingRecurrenceNone    MeetingRecurrence = "none"    // Не повторяется
	MeetingRecurrenceDaily   MeetingRecurrence = "daily"   // Ежедневно
	MeetingRecurrenceWeekly  MeetingRecurrence = "weekly"  // Еженедельно
	MeetingRecurrenceMonthly MeetingRecurrence = "monthly" // Ежемесячно
)

// MeetingSubject представляет тему/категорию встречи
type MeetingSubject struct {
	// ID - уникальный идентификатор темы
	ID            uuid.UUID    `json:"id" db:"id"`
	// Name - название темы
	Name          string       `json:"name" db:"name"`
	// Description - описание темы
	Description   string       `json:"description" db:"description"`
	// DepartmentIDs - отделы, с которыми связана эта тема
	DepartmentIDs []uuid.UUID  `json:"department_ids" db:"department_ids"`
	// IsActive - активна ли тема
	IsActive      bool         `json:"is_active" db:"is_active"`
	// CreatedAt - время создания
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
	// UpdatedAt - время последнего обновления
	UpdatedAt     time.Time    `json:"updated_at" db:"updated_at"`
}

// Meeting представляет видеовстречу
type Meeting struct {
	// ID - уникальный идентификатор встречи
	ID                   uuid.UUID         `json:"id" db:"id"`
	// Title - название встречи
	Title                string            `json:"title" db:"title"`
	// ScheduledAt - запланированное время начала
	ScheduledAt          time.Time         `json:"scheduled_at" db:"scheduled_at"`
	// Duration - длительность в минутах
	Duration             int               `json:"duration" db:"duration"`
	// Recurrence - частота повторения
	Recurrence           MeetingRecurrence `json:"recurrence" db:"recurrence"`
	// Type - тип встречи
	Type                 MeetingType       `json:"type" db:"type"`
	// SubjectID - идентификатор темы встречи
	SubjectID            uuid.UUID         `json:"subject_id" db:"subject_id"`
	// Status - текущий статус встречи
	Status               MeetingStatus     `json:"status" db:"status"`
	// NeedsVideoRecord - требуется ли видеозапись
	NeedsVideoRecord     bool              `json:"needs_video_record" db:"needs_video_record"`
	// NeedsAudioRecord - требуется ли аудиозапись
	NeedsAudioRecord     bool              `json:"needs_audio_record" db:"needs_audio_record"`
	// AdditionalNotes - дополнительные заметки
	AdditionalNotes      string            `json:"additional_notes" db:"additional_notes"`
	// ForceEndAtDuration - принудительно завершить встречу после истечения времени
	ForceEndAtDuration   bool              `json:"force_end_at_duration" db:"force_end_at_duration"`
	// LiveKitRoomID - ссылка на комнату LiveKit если встреча начата
	LiveKitRoomID        *uuid.UUID        `json:"livekit_room_id,omitempty" db:"livekit_room_id"`
	// CreatedBy - идентификатор пользователя, создавшего встречу
	CreatedBy            uuid.UUID         `json:"created_by" db:"created_by"`
	// CreatedAt - время создания
	CreatedAt            time.Time         `json:"created_at" db:"created_at"`
	// UpdatedAt - время последнего обновления
	UpdatedAt            time.Time         `json:"updated_at" db:"updated_at"`
}

// MeetingParticipant представляет участника встречи
type MeetingParticipant struct {
	// ID - уникальный идентификатор участника
	ID          uuid.UUID  `json:"id" db:"id"`
	// MeetingID - идентификатор встречи
	MeetingID   uuid.UUID  `json:"meeting_id" db:"meeting_id"`
	// UserID - идентификатор пользователя
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	// Role - роль участника ("speaker", "participant")
	Role        string     `json:"role" db:"role"`
	// Status - статус участника ("invited", "accepted", "declined", "attended")
	Status      string     `json:"status" db:"status"`
	// JoinedAt - время присоединения к встрече
	JoinedAt    *time.Time `json:"joined_at,omitempty" db:"joined_at"`
	// LeftAt - время выхода из встречи
	LeftAt      *time.Time `json:"left_at,omitempty" db:"left_at"`
	// CreatedAt - время создания записи
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// MeetingDepartment представляет отдел, приглашенный на встречу
type MeetingDepartment struct {
	// ID - уникальный идентификатор записи
	ID           uuid.UUID `json:"id" db:"id"`
	// MeetingID - идентификатор встречи
	MeetingID    uuid.UUID `json:"meeting_id" db:"meeting_id"`
	// DepartmentID - идентификатор отдела
	DepartmentID uuid.UUID `json:"department_id" db:"department_id"`
	// CreatedAt - время создания записи
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// MeetingWithDetails представляет встречу со всей связанной информацией
type MeetingWithDetails struct {
	Meeting
	// Subject - тема встречи
	Subject         *MeetingSubject        `json:"subject,omitempty"`
	// Participants - список участников
	Participants    []MeetingParticipantInfo `json:"participants"`
	// Departments - список отделов
	Departments     []Department           `json:"departments"`
	// CreatedByUser - пользователь, создавший встречу
	CreatedByUser   *UserInfo              `json:"created_by_user,omitempty"`
}

// MeetingParticipantInfo представляет информацию об участнике с деталями пользователя
type MeetingParticipantInfo struct {
	MeetingParticipant
	// User - информация о пользователе
	User *UserInfo `json:"user,omitempty"`
}

// CreateMeetingSubjectRequest представляет запрос на создание темы встречи
type CreateMeetingSubjectRequest struct {
	// Name - название темы
	Name          string       `json:"name" binding:"required" example:"Техническое обсуждение"`
	// Description - описание темы
	Description   string       `json:"description" example:"Обсуждение технических вопросов разработки"`
	// DepartmentIDs - идентификаторы отделов
	DepartmentIDs []uuid.UUID  `json:"department_ids" example:"[\"550e8400-e29b-41d4-a716-446655440000\"]"`
}

// UpdateMeetingSubjectRequest представляет запрос на обновление темы встречи
type UpdateMeetingSubjectRequest struct {
	// Name - название темы
	Name          string       `json:"name" example:"Техническое обсуждение"`
	// Description - описание темы
	Description   string       `json:"description" example:"Обновленное описание"`
	// DepartmentIDs - идентификаторы отделов
	DepartmentIDs []uuid.UUID  `json:"department_ids" example:"[\"550e8400-e29b-41d4-a716-446655440000\"]"`
	// IsActive - активна ли тема
	IsActive      *bool        `json:"is_active" example:"true"`
}

// CreateMeetingRequest представляет запрос на создание встречи
type CreateMeetingRequest struct {
	// Title - название встречи
	Title              string            `json:"title" binding:"required" example:"Еженедельное совещание"`
	// ScheduledAt - запланированное время начала
	ScheduledAt        time.Time         `json:"scheduled_at" binding:"required" example:"2025-01-15T10:00:00Z"`
	// Duration - длительность в минутах
	Duration           int               `json:"duration" binding:"required" example:"60"`
	// Recurrence - частота повторения
	Recurrence         MeetingRecurrence `json:"recurrence" example:"weekly"`
	// Type - тип встречи
	Type               MeetingType       `json:"type" binding:"required" example:"conference"`
	// SubjectID - идентификатор темы
	SubjectID          uuid.UUID         `json:"subject_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// NeedsVideoRecord - требуется ли видеозапись
	NeedsVideoRecord   bool              `json:"needs_video_record" example:"true"`
	// NeedsAudioRecord - требуется ли аудиозапись
	NeedsAudioRecord   bool              `json:"needs_audio_record" example:"true"`
	// AdditionalNotes - дополнительные заметки
	AdditionalNotes    string            `json:"additional_notes" example:"Подготовить отчеты"`
	// ForceEndAtDuration - принудительно завершить встречу после истечения времени
	ForceEndAtDuration bool              `json:"force_end_at_duration" example:"true"`
	// SpeakerID - идентификатор докладчика (для презентаций)
	SpeakerID          *uuid.UUID        `json:"speaker_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	// ParticipantIDs - идентификаторы участников
	ParticipantIDs     []uuid.UUID       `json:"participant_ids" example:"[\"550e8400-e29b-41d4-a716-446655440000\"]"`
	// DepartmentIDs - идентификаторы отделов
	DepartmentIDs      []uuid.UUID       `json:"department_ids" example:"[\"550e8400-e29b-41d4-a716-446655440000\"]"`
}

// UpdateMeetingRequest представляет запрос на обновление встречи
type UpdateMeetingRequest struct {
	// Title - название встречи
	Title              *string            `json:"title,omitempty" example:"Обновленное название"`
	// ScheduledAt - запланированное время начала
	ScheduledAt        *time.Time         `json:"scheduled_at,omitempty" example:"2025-01-15T10:00:00Z"`
	// Duration - длительность в минутах
	Duration           *int               `json:"duration,omitempty" example:"90"`
	// Recurrence - частота повторения
	Recurrence         *MeetingRecurrence `json:"recurrence,omitempty" example:"weekly"`
	// Type - тип встречи
	Type               *MeetingType       `json:"type,omitempty" example:"conference"`
	// SubjectID - идентификатор темы
	SubjectID          *uuid.UUID         `json:"subject_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Status - статус встречи
	Status             *MeetingStatus     `json:"status,omitempty" example:"in_progress"`
	// NeedsVideoRecord - требуется ли видеозапись
	NeedsVideoRecord   *bool              `json:"needs_video_record,omitempty" example:"true"`
	// NeedsAudioRecord - требуется ли аудиозапись
	NeedsAudioRecord   *bool              `json:"needs_audio_record,omitempty" example:"true"`
	// AdditionalNotes - дополнительные заметки
	AdditionalNotes    *string            `json:"additional_notes,omitempty" example:"Обновленные комментарии"`
	// ForceEndAtDuration - принудительно завершить встречу после истечения времени
	ForceEndAtDuration *bool              `json:"force_end_at_duration,omitempty" example:"true"`
	// SpeakerID - идентификатор докладчика
	SpeakerID          *uuid.UUID         `json:"speaker_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	// ParticipantIDs - идентификаторы участников
	ParticipantIDs     []uuid.UUID        `json:"participant_ids,omitempty" example:"[\"550e8400-e29b-41d4-a716-446655440000\"]"`
	// DepartmentIDs - идентификаторы отделов
	DepartmentIDs      []uuid.UUID        `json:"department_ids,omitempty" example:"[\"550e8400-e29b-41d4-a716-446655440000\"]"`
}

// ListMeetingsRequest представляет параметры для получения списка встреч
type ListMeetingsRequest struct {
	// Page - номер страницы
	Page         int            `json:"page" form:"page" example:"1"`
	// PageSize - размер страницы
	PageSize     int            `json:"page_size" form:"page_size" example:"20"`
	// Status - фильтр по статусу
	Status       *MeetingStatus `json:"status" form:"status" example:"scheduled"`
	// Type - фильтр по типу
	Type         *MeetingType   `json:"type" form:"type" example:"conference"`
	// SubjectID - фильтр по теме
	SubjectID    *uuid.UUID     `json:"subject_id" form:"subject_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// SpeakerID - фильтр по докладчику
	SpeakerID    *uuid.UUID     `json:"speaker_id" form:"speaker_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// DateFrom - начальная дата периода
	DateFrom     *time.Time     `json:"date_from" form:"date_from" example:"2025-01-01T00:00:00Z"`
	// DateTo - конечная дата периода
	DateTo       *time.Time     `json:"date_to" form:"date_to" example:"2025-12-31T23:59:59Z"`
	// UserID - фильтр по участнику или докладчику
	UserID       *uuid.UUID     `json:"user_id" form:"user_id"`
}

// PaginatedResponse представляет постраничный ответ
type PaginatedResponse struct {
	// Items - список элементов
	Items    interface{} `json:"items"`
	// Offset - смещение от начала
	Offset   int         `json:"offset"`
	// PageSize - размер страницы
	PageSize int         `json:"page_size"`
	// Total - общее количество элементов
	Total    int         `json:"total"`
}

// MeetingsResponse представляет постраничный список встреч
type MeetingsResponse struct {
	// Items - список встреч
	Items    []MeetingWithDetails `json:"items"`
	// Offset - смещение от начала
	Offset   int                  `json:"offset"`
	// PageSize - размер страницы
	PageSize int                  `json:"page_size"`
	// Total - общее количество встреч
	Total    int                  `json:"total"`
}

// MeetingSubjectsResponse представляет постраничный список тем встреч
type MeetingSubjectsResponse struct {
	// Items - список тем
	Items    []MeetingSubject `json:"items"`
	// Offset - смещение от начала
	Offset   int              `json:"offset"`
	// PageSize - размер страницы
	PageSize int              `json:"page_size"`
	// Total - общее количество тем
	Total    int              `json:"total"`
}

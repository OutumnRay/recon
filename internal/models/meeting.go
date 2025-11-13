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
	// Уникальный идентификатор темы
	ID            uuid.UUID    `json:"id" db:"id"`
	// Название темы
	Name          string       `json:"name" db:"name"`
	// Описание темы
	Description   string       `json:"description" db:"description"`
	// Отделы, с которыми связана эта тема
	DepartmentIDs []uuid.UUID  `json:"department_ids" db:"department_ids"`
	// Активна ли тема
	IsActive      bool         `json:"is_active" db:"is_active"`
	// Время создания
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
	// Время последнего обновления
	UpdatedAt     time.Time    `json:"updated_at" db:"updated_at"`
}

// Meeting представляет видеовстречу
type Meeting struct {
	// Уникальный идентификатор встречи
	ID                   uuid.UUID         `json:"id" db:"id"`
	// Название встречи
	Title                string            `json:"title" db:"title"`
	// Запланированное время начала
	ScheduledAt          time.Time         `json:"scheduled_at" db:"scheduled_at"`
	// Длительность в минутах
	Duration             int               `json:"duration" db:"duration"`
	// Частота повторения
	Recurrence           MeetingRecurrence `json:"recurrence" db:"recurrence"`
	// Тип встречи
	Type                 MeetingType       `json:"type" db:"type"`
	// Идентификатор темы встречи
	SubjectID            uuid.UUID         `json:"subject_id" db:"subject_id"`
	// Текущий статус встречи
	Status               MeetingStatus     `json:"status" db:"status"`
	// Требуется ли видеозапись
	NeedsVideoRecord     bool              `json:"needs_video_record" db:"needs_video_record"`
	// Требуется ли аудиозапись
	NeedsAudioRecord     bool              `json:"needs_audio_record" db:"needs_audio_record"`
	// Требуется ли транскрибация (запись отдельных аудио треков участников)
	NeedsTranscription   bool              `json:"needs_transcription" db:"needs_transcription"`
	// Идет ли сейчас запись
	IsRecording          bool              `json:"is_recording" db:"is_recording"`
	// Идет ли сейчас транскрибация
	IsTranscribing       bool              `json:"is_transcribing" db:"is_transcribing"`
	// Дополнительные заметки
	AdditionalNotes      string            `json:"additional_notes" db:"additional_notes"`
	// Принудительно завершить встречу после истечения времени
	ForceEndAtDuration   bool              `json:"force_end_at_duration" db:"force_end_at_duration"`
	// Ссылка на комнату LiveKit если встреча начата
	LiveKitRoomID        *uuid.UUID        `json:"livekit_room_id,omitempty" db:"livekit_room_id"`
	// Идентификатор пользователя, создавшего встречу
	CreatedBy            uuid.UUID         `json:"created_by" db:"created_by"`
	// Время создания
	CreatedAt            time.Time         `json:"created_at" db:"created_at"`
	// Время последнего обновления
	UpdatedAt            time.Time         `json:"updated_at" db:"updated_at"`
}

// MeetingParticipant представляет участника встречи
type MeetingParticipant struct {
	// Уникальный идентификатор участника
	ID          uuid.UUID  `json:"id" db:"id"`
	// Идентификатор встречи
	MeetingID   uuid.UUID  `json:"meeting_id" db:"meeting_id"`
	// Идентификатор пользователя
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	// Роль участника ("speaker", "participant")
	Role        string     `json:"role" db:"role"`
	// Статус участника ("invited", "accepted", "declined", "attended")
	Status      string     `json:"status" db:"status"`
	// Время присоединения к встрече
	JoinedAt    *time.Time `json:"joined_at,omitempty" db:"joined_at"`
	// Время выхода из встречи
	LeftAt      *time.Time `json:"left_at,omitempty" db:"left_at"`
	// Время создания записи
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// MeetingDepartment представляет отдел, приглашенный на встречу
type MeetingDepartment struct {
	// Уникальный идентификатор записи
	ID           uuid.UUID `json:"id" db:"id"`
	// Идентификатор встречи
	MeetingID    uuid.UUID `json:"meeting_id" db:"meeting_id"`
	// Идентификатор отдела
	DepartmentID uuid.UUID `json:"department_id" db:"department_id"`
	// Время создания записи
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// MeetingWithDetails представляет встречу со всей связанной информацией
type MeetingWithDetails struct {
	Meeting
	// Тема встречи
	Subject         *MeetingSubject        `json:"subject,omitempty"`
	// Список участников
	Participants    []MeetingParticipantInfo `json:"participants"`
	// Список отделов
	Departments     []Department           `json:"departments"`
	// Пользователь, создавший встречу
	CreatedByUser   *UserInfo              `json:"created_by_user,omitempty"`
}

// MeetingParticipantInfo представляет информацию об участнике с деталями пользователя
type MeetingParticipantInfo struct {
	MeetingParticipant
	// Информация о пользователе
	User *UserInfo `json:"user,omitempty"`
}

// CreateMeetingSubjectRequest представляет запрос на создание темы встречи
type CreateMeetingSubjectRequest struct {
	// Название темы
	Name          string       `json:"name" binding:"required" example:"Техническое обсуждение"`
	// Описание темы
	Description   string       `json:"description" example:"Обсуждение технических вопросов разработки"`
	// Идентификаторы отделов
	DepartmentIDs []uuid.UUID  `json:"department_ids" example:"[\"550e8400-e29b-41d4-a716-446655440000\"]"`
}

// UpdateMeetingSubjectRequest представляет запрос на обновление темы встречи
type UpdateMeetingSubjectRequest struct {
	// Название темы
	Name          string       `json:"name" example:"Техническое обсуждение"`
	// Описание темы
	Description   string       `json:"description" example:"Обновленное описание"`
	// Идентификаторы отделов
	DepartmentIDs []uuid.UUID  `json:"department_ids" example:"[\"550e8400-e29b-41d4-a716-446655440000\"]"`
	// Активна ли тема
	IsActive      *bool        `json:"is_active" example:"true"`
}

// CreateMeetingRequest представляет запрос на создание встречи
type CreateMeetingRequest struct {
	// Название встречи
	Title              string            `json:"title" binding:"required" example:"Еженедельное совещание"`
	// Запланированное время начала
	ScheduledAt        time.Time         `json:"scheduled_at" binding:"required" example:"2025-01-15T10:00:00Z"`
	// Длительность в минутах
	Duration           int               `json:"duration" binding:"required" example:"60"`
	// Частота повторения
	Recurrence         MeetingRecurrence `json:"recurrence" example:"weekly"`
	// Тип встречи
	Type               MeetingType       `json:"type" binding:"required" example:"conference"`
	// Идентификатор темы
	SubjectID          uuid.UUID         `json:"subject_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Требуется ли видеозапись
	NeedsVideoRecord   bool              `json:"needs_video_record" example:"true"`
	// Требуется ли аудиозапись
	NeedsAudioRecord   bool              `json:"needs_audio_record" example:"true"`
	// Требуется ли транскрибация
	NeedsTranscription bool              `json:"needs_transcription" example:"true"`
	// Дополнительные заметки
	AdditionalNotes    string            `json:"additional_notes" example:"Подготовить отчеты"`
	// Принудительно завершить встречу после истечения времени
	ForceEndAtDuration bool              `json:"force_end_at_duration" example:"true"`
	// Идентификатор докладчика (для презентаций)
	SpeakerID          *uuid.UUID        `json:"speaker_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Идентификаторы участников
	ParticipantIDs     []uuid.UUID       `json:"participant_ids" example:"[\"550e8400-e29b-41d4-a716-446655440000\"]"`
	// Идентификаторы отделов
	DepartmentIDs      []uuid.UUID       `json:"department_ids" example:"[\"550e8400-e29b-41d4-a716-446655440000\"]"`
}

// UpdateMeetingRequest представляет запрос на обновление встречи
type UpdateMeetingRequest struct {
	// Название встречи
	Title              *string            `json:"title,omitempty" example:"Обновленное название"`
	// Запланированное время начала
	ScheduledAt        *time.Time         `json:"scheduled_at,omitempty" example:"2025-01-15T10:00:00Z"`
	// Длительность в минутах
	Duration           *int               `json:"duration,omitempty" example:"90"`
	// Частота повторения
	Recurrence         *MeetingRecurrence `json:"recurrence,omitempty" example:"weekly"`
	// Тип встречи
	Type               *MeetingType       `json:"type,omitempty" example:"conference"`
	// Идентификатор темы
	SubjectID          *uuid.UUID         `json:"subject_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Статус встречи
	Status             *MeetingStatus     `json:"status,omitempty" example:"in_progress"`
	// Требуется ли видеозапись
	NeedsVideoRecord   *bool              `json:"needs_video_record,omitempty" example:"true"`
	// Требуется ли аудиозапись
	NeedsAudioRecord   *bool              `json:"needs_audio_record,omitempty" example:"true"`
	// Требуется ли транскрибация
	NeedsTranscription *bool              `json:"needs_transcription,omitempty" example:"true"`
	// Дополнительные заметки
	AdditionalNotes    *string            `json:"additional_notes,omitempty" example:"Обновленные комментарии"`
	// Принудительно завершить встречу после истечения времени
	ForceEndAtDuration *bool              `json:"force_end_at_duration,omitempty" example:"true"`
	// Идентификатор докладчика
	SpeakerID          *uuid.UUID         `json:"speaker_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Идентификаторы участников
	ParticipantIDs     []uuid.UUID        `json:"participant_ids,omitempty" example:"[\"550e8400-e29b-41d4-a716-446655440000\"]"`
	// Идентификаторы отделов
	DepartmentIDs      []uuid.UUID        `json:"department_ids,omitempty" example:"[\"550e8400-e29b-41d4-a716-446655440000\"]"`
}

// ListMeetingsRequest представляет параметры для получения списка встреч
type ListMeetingsRequest struct {
	// Номер страницы
	Page         int            `json:"page" form:"page" example:"1"`
	// Размер страницы
	PageSize     int            `json:"page_size" form:"page_size" example:"20"`
	// Фильтр по статусу
	Status       *MeetingStatus `json:"status" form:"status" example:"scheduled"`
	// Фильтр по типу
	Type         *MeetingType   `json:"type" form:"type" example:"conference"`
	// Фильтр по теме
	SubjectID    *uuid.UUID     `json:"subject_id" form:"subject_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Фильтр по докладчику
	SpeakerID    *uuid.UUID     `json:"speaker_id" form:"speaker_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Начальная дата периода
	DateFrom     *time.Time     `json:"date_from" form:"date_from" example:"2025-01-01T00:00:00Z"`
	// Конечная дата периода
	DateTo       *time.Time     `json:"date_to" form:"date_to" example:"2025-12-31T23:59:59Z"`
	// Фильтр по участнику или докладчику
	UserID       *uuid.UUID     `json:"user_id" form:"user_id"`
}

// PaginatedResponse представляет постраничный ответ
type PaginatedResponse struct {
	// Список элементов
	Items    interface{} `json:"items"`
	// Смещение от начала
	Offset   int         `json:"offset"`
	// Размер страницы
	PageSize int         `json:"page_size"`
	// Общее количество элементов
	Total    int         `json:"total"`
}

// MeetingsResponse представляет постраничный список встреч
type MeetingsResponse struct {
	// Список встреч
	Items    []MeetingWithDetails `json:"items"`
	// Смещение от начала
	Offset   int                  `json:"offset"`
	// Размер страницы
	PageSize int                  `json:"page_size"`
	// Общее количество встреч
	Total    int                  `json:"total"`
}

// MeetingSubjectsResponse представляет постраничный список тем встреч
type MeetingSubjectsResponse struct {
	// Список тем
	Items    []MeetingSubject `json:"items"`
	// Смещение от начала
	Offset   int              `json:"offset"`
	// Размер страницы
	PageSize int              `json:"page_size"`
	// Общее количество тем
	Total    int              `json:"total"`
}

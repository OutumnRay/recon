package database

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Department represents a department in the organization
type Department struct {
	ID          string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	ParentID    *string        `gorm:"type:varchar(255)" json:"parent_id"`
	Level       int            `gorm:"not null;default:0" json:"level"`
	Path        string         `gorm:"type:text;not null" json:"path"`
	IsActive    bool           `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Parent   *Department  `gorm:"foreignKey:ParentID" json:"-"`
	Children []Department `gorm:"foreignKey:ParentID" json:"-"`
}

// User represents a user in the system
type User struct {
	ID           string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	Username     string         `gorm:"uniqueIndex;type:varchar(255);not null" json:"username"`
	Email        string         `gorm:"uniqueIndex;type:varchar(255);not null" json:"email"`
	Password     string         `gorm:"type:varchar(255);not null" json:"-"`
	Role         string         `gorm:"type:varchar(50);not null;default:'user'" json:"role"`
	Groups       pq.StringArray `gorm:"type:text[];default:'{}'" json:"groups"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	LastLogin    *time.Time     `json:"last_login"`
	Language     string         `gorm:"type:varchar(10);not null;default:'en'" json:"language"`
	DepartmentID *string        `gorm:"type:varchar(255)" json:"department_id"`
	Permissions  string         `gorm:"type:jsonb;not null;default:'{\"can_schedule_meetings\": false, \"can_manage_department\": false, \"can_approve_recordings\": false}'" json:"permissions"`
	FirstName    *string        `gorm:"column:first_name;type:varchar(255)" json:"first_name"`
	LastName     *string        `gorm:"column:last_name;type:varchar(255)" json:"last_name"`
	Phone        *string        `gorm:"column:phone;type:varchar(50)" json:"phone"`
	Bio          *string        `gorm:"column:bio;type:text" json:"bio"`
	AvatarURL    *string        `gorm:"column:avatar_url;type:text" json:"avatar_url"`
	CreatedAt    time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Department *Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
}

// Group represents a user group with permissions
type Group struct {
	ID          string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	Name        string         `gorm:"uniqueIndex;type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Permissions string         `gorm:"type:jsonb;not null;default:'{}'" json:"permissions"`
	CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// GroupMembership represents the many-to-many relationship between users and groups
type GroupMembership struct {
	UserID  string    `gorm:"primaryKey;type:varchar(255);not null" json:"user_id"`
	GroupID string    `gorm:"primaryKey;type:varchar(255);not null" json:"group_id"`
	AddedAt time.Time `gorm:"not null;default:now()" json:"added_at"`
	AddedBy *string   `gorm:"type:varchar(255)" json:"added_by"`

	// Relations
	User  User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Group Group `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE" json:"-"`
}

// UploadedFile represents an uploaded audio/video file
type UploadedFile struct {
	ID              string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	Filename        string         `gorm:"type:varchar(500);not null" json:"filename"`
	OriginalName    string         `gorm:"type:varchar(500);not null" json:"original_name"`
	FileSize        int64          `gorm:"not null" json:"file_size"`
	MimeType        string         `gorm:"type:varchar(255);not null" json:"mime_type"`
	StoragePath     string         `gorm:"type:text;not null" json:"storage_path"`
	UserID          string         `gorm:"type:varchar(255);not null" json:"user_id"`
	GroupID         string         `gorm:"type:varchar(255);not null" json:"group_id"`
	Status          string         `gorm:"type:varchar(50);not null;default:'pending'" json:"status"`
	TranscriptionID *string        `gorm:"type:varchar(255)" json:"transcription_id"`
	Metadata        string         `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	UploadedAt      time.Time      `gorm:"not null;default:now()" json:"uploaded_at"`
	ProcessedAt     *time.Time     `json:"processed_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User  User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Group Group `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE" json:"-"`
}

// FileTranscription represents a transcription of an uploaded file
type FileTranscription struct {
	ID            string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	FileID        string         `gorm:"type:varchar(255);not null" json:"file_id"`
	Text          string         `gorm:"type:text;not null" json:"text"`
	Language      string         `gorm:"type:varchar(10);not null" json:"language"`
	Confidence    *float64       `gorm:"type:decimal(5,4)" json:"confidence"`
	Duration      *float64       `gorm:"type:decimal(10,2)" json:"duration"`
	Segments      string         `gorm:"type:jsonb;default:'{}'" json:"segments"`
	TranscribedAt time.Time      `gorm:"not null;default:now()" json:"transcribed_at"`
	TranscribedBy string         `gorm:"type:varchar(255);not null" json:"transcribed_by"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	File UploadedFile `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"-"`
}

// DocumentChunk represents a chunk of a document with embeddings
type DocumentChunk struct {
	ID              string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	FileID          string         `gorm:"type:varchar(255);not null" json:"file_id"`
	TranscriptionID *string        `gorm:"type:varchar(255)" json:"transcription_id"`
	ChunkText       string         `gorm:"type:text;not null" json:"chunk_text"`
	ChunkIndex      int            `gorm:"not null" json:"chunk_index"`
	Embedding       string         `gorm:"type:vector(1536)" json:"-"` // pgvector type
	Metadata        string         `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt       time.Time      `gorm:"not null;default:now()" json:"created_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	File          UploadedFile       `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"-"`
	Transcription *FileTranscription `gorm:"foreignKey:TranscriptionID;constraint:OnDelete:CASCADE" json:"-"`
}

// LiveKitRoom represents a LiveKit room
type LiveKitRoom struct {
	ID               string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	Sid              string         `gorm:"uniqueIndex;type:varchar(255);not null" json:"sid"`
	Name             string         `gorm:"type:varchar(255);not null" json:"name"`
	EmptyTimeout     int            `gorm:"default:300" json:"empty_timeout"`
	DepartureTimeout int            `gorm:"default:20" json:"departure_timeout"`
	CreationTime     string         `gorm:"type:varchar(50)" json:"creation_time"`
	CreationTimeMs   string         `gorm:"type:varchar(50)" json:"creation_time_ms"`
	TurnPassword     string         `gorm:"type:text" json:"turn_password"`
	EnabledCodecs    string         `gorm:"type:jsonb;default:'[]'" json:"enabled_codecs"`
	Status           string         `gorm:"type:varchar(50);not null;default:'active'" json:"status"`
	StartedAt        time.Time      `gorm:"not null" json:"started_at"`
	FinishedAt       *time.Time     `json:"finished_at"`
	CreatedAt        time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// LiveKitParticipant represents a participant in a LiveKit room
type LiveKitParticipant struct {
	ID               string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	Sid              string         `gorm:"uniqueIndex;type:varchar(255);not null" json:"sid"`
	RoomSid          string         `gorm:"type:varchar(255);not null" json:"room_sid"`
	Identity         string         `gorm:"type:varchar(255);not null" json:"identity"`
	Name             string         `gorm:"type:varchar(255);not null" json:"name"`
	State            string         `gorm:"type:varchar(50);not null;default:'ACTIVE'" json:"state"`
	JoinedAt         string         `gorm:"type:varchar(50);not null" json:"joined_at"`
	JoinedAtMs       string         `gorm:"type:varchar(50);not null" json:"joined_at_ms"`
	Version          int            `gorm:"default:0" json:"version"`
	Permission       string         `gorm:"type:jsonb;default:'{}'" json:"permission"`
	IsPublisher      bool           `gorm:"default:false" json:"is_publisher"`
	DisconnectReason string         `gorm:"type:varchar(255)" json:"disconnect_reason"`
	LeftAt           *time.Time     `json:"left_at"`
	CreatedAt        time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Room LiveKitRoom `gorm:"foreignKey:RoomSid;references:Sid;constraint:OnDelete:CASCADE" json:"-"`
}

// LiveKitTrack represents a media track in a LiveKit room
type LiveKitTrack struct {
	ID                string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	Sid               string         `gorm:"uniqueIndex;type:varchar(255);not null" json:"sid"`
	ParticipantSid    string         `gorm:"type:varchar(255);not null" json:"participant_sid"`
	RoomSid           string         `gorm:"type:varchar(255);not null" json:"room_sid"`
	Type              string         `gorm:"type:varchar(50)" json:"type"`
	Source            string         `gorm:"type:varchar(50);not null" json:"source"`
	MimeType          string         `gorm:"type:varchar(100);not null" json:"mime_type"`
	Mid               string         `gorm:"type:varchar(50)" json:"mid"`
	Width             *int           `json:"width"`
	Height            *int           `json:"height"`
	Simulcast         bool           `gorm:"default:false" json:"simulcast"`
	Layers            string         `gorm:"type:jsonb;default:'[]'" json:"layers"`
	Codecs            string         `gorm:"type:jsonb;default:'[]'" json:"codecs"`
	Stream            string         `gorm:"type:varchar(255)" json:"stream"`
	Version           string         `gorm:"type:jsonb;default:'{}'" json:"version"`
	AudioFeatures     pq.StringArray `gorm:"type:text[];default:'{}'" json:"audio_features"`
	BackupCodecPolicy string         `gorm:"type:varchar(50)" json:"backup_codec_policy"`
	Status            string         `gorm:"type:varchar(50);not null;default:'published'" json:"status"`
	PublishedAt       time.Time      `gorm:"not null" json:"published_at"`
	UnpublishedAt     *time.Time     `json:"unpublished_at"`
	CreatedAt         time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Participant LiveKitParticipant `gorm:"foreignKey:ParticipantSid;references:Sid;constraint:OnDelete:CASCADE" json:"-"`
	Room        LiveKitRoom        `gorm:"foreignKey:RoomSid;references:Sid;constraint:OnDelete:CASCADE" json:"-"`
}

// LiveKitWebhookEvent represents a webhook event from LiveKit
type LiveKitWebhookEvent struct {
	ID             string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	EventType      string         `gorm:"type:varchar(100);not null" json:"event_type"`
	EventID        string         `gorm:"type:varchar(255);not null" json:"event_id"`
	RoomSid        *string        `gorm:"type:varchar(255)" json:"room_sid"`
	ParticipantSid *string        `gorm:"type:varchar(255)" json:"participant_sid"`
	TrackSid       *string        `gorm:"type:varchar(255)" json:"track_sid"`
	Payload        string         `gorm:"type:jsonb;not null" json:"payload"`
	CreatedAt      time.Time      `gorm:"not null;default:now()" json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// MeetingSubject represents a meeting subject/topic
type MeetingSubject struct {
	ID            string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	Name          string         `gorm:"uniqueIndex;type:varchar(255);not null" json:"name"`
	Description   string         `gorm:"type:text" json:"description"`
	DepartmentIDs pq.StringArray `gorm:"type:text[];default:'{}'" json:"department_ids"`
	IsActive      bool           `gorm:"not null;default:true" json:"is_active"`
	CreatedAt     time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// Meeting represents a scheduled or ongoing meeting
type Meeting struct {
	ID                 string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	Title              string         `gorm:"type:varchar(500);not null" json:"title"`
	ScheduledAt        time.Time      `gorm:"not null" json:"scheduled_at"`
	Duration           int            `gorm:"not null" json:"duration"`
	Recurrence         string         `gorm:"type:varchar(50);not null;default:'none'" json:"recurrence"`
	Type               string         `gorm:"type:varchar(50);not null" json:"type"`
	SubjectID          string         `gorm:"type:varchar(255);not null" json:"subject_id"`
	Status             string         `gorm:"type:varchar(50);not null;default:'scheduled'" json:"status"`
	NeedsVideoRecord   bool           `gorm:"not null;default:false" json:"needs_video_record"`
	NeedsAudioRecord   bool           `gorm:"not null;default:false" json:"needs_audio_record"`
	ForceEndAtDuration bool           `gorm:"column:force_end_at_duration;not null;default:false" json:"force_end_at_duration"`
	AdditionalNotes    string         `gorm:"column:additional_notes;type:text" json:"additional_notes"`
	LiveKitRoomID      *string        `gorm:"column:livekit_room_id;type:varchar(255)" json:"livekit_room_id"`
	CreatedBy          string         `gorm:"type:varchar(255);not null" json:"created_by"`
	CreatedAt          time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Subject MeetingSubject `gorm:"foreignKey:SubjectID;constraint:OnDelete:RESTRICT" json:"subject,omitempty"`
	Creator User           `gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE" json:"creator,omitempty"`
}

// MeetingParticipant represents a participant in a meeting
type MeetingParticipant struct {
	ID        string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	MeetingID string         `gorm:"type:varchar(255);not null;uniqueIndex:idx_meeting_user" json:"meeting_id"`
	UserID    string         `gorm:"type:varchar(255);not null;uniqueIndex:idx_meeting_user" json:"user_id"`
	Role      string         `gorm:"type:varchar(50);not null" json:"role"`
	Status    string         `gorm:"type:varchar(50);not null;default:'invited'" json:"status"`
	JoinedAt  *time.Time     `json:"joined_at"`
	LeftAt    *time.Time     `json:"left_at"`
	CreatedAt time.Time      `gorm:"not null;default:now()" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Meeting Meeting `gorm:"foreignKey:MeetingID;constraint:OnDelete:CASCADE" json:"-"`
	User    User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// MeetingDepartment represents departments invited to a meeting
type MeetingDepartment struct {
	ID           string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	MeetingID    string         `gorm:"type:varchar(255);not null;uniqueIndex:idx_meeting_dept" json:"meeting_id"`
	DepartmentID string         `gorm:"type:varchar(255);not null;uniqueIndex:idx_meeting_dept" json:"department_id"`
	CreatedAt    time.Time      `gorm:"not null;default:now()" json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Meeting    Meeting    `gorm:"foreignKey:MeetingID;constraint:OnDelete:CASCADE" json:"-"`
	Department Department `gorm:"foreignKey:DepartmentID;constraint:OnDelete:CASCADE" json:"-"`
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID    string    `gorm:"type:varchar(255);not null" json:"user_id"`
	Email     string    `gorm:"type:varchar(255);not null" json:"email"`
	Code      string    `gorm:"type:varchar(6);not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Used      bool      `gorm:"not null;default:false" json:"used"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`

	// Relations
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// LiveKitEgress represents a LiveKit egress (recording) session
type LiveKitEgress struct {
	ID             string     `gorm:"primaryKey;type:varchar(255)" json:"id"` // EgressID from LiveKit
	RoomSID        string     `gorm:"type:varchar(255);not null;index" json:"room_sid"`
	RoomName       string     `gorm:"type:varchar(255);not null;index" json:"room_name"`
	Type           string     `gorm:"type:varchar(50);not null" json:"type"` // room_composite, track_composite, track
	Status         string     `gorm:"type:varchar(50);not null;index" json:"status"` // pending, active, finishing, complete, failed
	TrackID        *string    `gorm:"type:varchar(255)" json:"track_id,omitempty"` // For track egress
	AudioTrackID   *string    `gorm:"type:varchar(255)" json:"audio_track_id,omitempty"` // For track composite
	VideoTrackID   *string    `gorm:"type:varchar(255)" json:"video_track_id,omitempty"` // For track composite
	FilePath       *string    `gorm:"type:text" json:"file_path,omitempty"` // S3 path or local path
	FileSize       *int64     `gorm:"type:bigint" json:"file_size,omitempty"` // File size in bytes
	Duration       *int64     `gorm:"type:bigint" json:"duration,omitempty"` // Duration in seconds
	Error          *string    `gorm:"type:text" json:"error,omitempty"` // Error message if failed
	StartedAt      *time.Time `gorm:"type:timestamp" json:"started_at,omitempty"`
	EndedAt        *time.Time `gorm:"type:timestamp" json:"ended_at,omitempty"`
	CreatedAt      time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"not null;default:now()" json:"updated_at"`

	// Relations
	Room LiveKitRoom `gorm:"foreignKey:RoomSID;references:Sid;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName overrides for consistency
func (Department) TableName() string          { return "departments" }
func (User) TableName() string                { return "users" }
func (Group) TableName() string               { return "groups" }
func (GroupMembership) TableName() string     { return "group_memberships" }
func (UploadedFile) TableName() string        { return "uploaded_files" }
func (FileTranscription) TableName() string   { return "file_transcriptions" }
func (DocumentChunk) TableName() string       { return "document_chunks" }
func (LiveKitRoom) TableName() string         { return "livekit_rooms" }
func (LiveKitParticipant) TableName() string  { return "livekit_participants" }
func (LiveKitTrack) TableName() string        { return "livekit_tracks" }
func (LiveKitWebhookEvent) TableName() string { return "livekit_webhook_events" }
func (LiveKitEgress) TableName() string       { return "livekit_egress" }
func (MeetingSubject) TableName() string      { return "meeting_subjects" }
func (Meeting) TableName() string             { return "meetings" }
func (MeetingParticipant) TableName() string  { return "meeting_participants" }
func (MeetingDepartment) TableName() string   { return "meeting_departments" }
func (PasswordResetToken) TableName() string  { return "password_reset_tokens" }

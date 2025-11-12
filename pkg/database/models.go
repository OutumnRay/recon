package database

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Department - отдел/департамент в организации
type Department struct {
	// ID - уникальный идентификатор отдела
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	// Name - название отдела
	Name string `gorm:"type:varchar(255);not null" json:"name"`
	// Description - описание отдела и его функций
	Description string `gorm:"type:text" json:"description"`
	// ParentID - ID родительского отдела (NULL для корневых отделов)
	ParentID *uuid.UUID `gorm:"type:uuid" json:"parent_id"`
	// Level - уровень вложенности в иерархии (0 для корневых отделов)
	Level int `gorm:"not null;default:0" json:"level"`
	// Path - путь в иерархии отделов (например, "parent/child/grandchild")
	Path string `gorm:"type:text;not null" json:"path"`
	// IsActive - активен ли отдел
	IsActive bool `gorm:"not null;default:true" json:"is_active"`
	// CreatedAt - время создания отдела
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	// UpdatedAt - время последнего обновления отдела
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
	// DeletedAt - время мягкого удаления отдела
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	// Parent - родительский отдел
	Parent *Department `gorm:"foreignKey:ParentID" json:"-"`
	// Children - дочерние отделы
	Children []Department `gorm:"foreignKey:ParentID" json:"-"`
}

// User - пользователь системы
type User struct {
	// ID - уникальный идентификатор пользователя
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	// Username - уникальное имя пользователя для входа в систему
	Username string `gorm:"uniqueIndex;type:varchar(255);not null" json:"username"`
	// Email - адрес электронной почты пользователя
	Email string `gorm:"uniqueIndex;type:varchar(255);not null" json:"email"`
	// Password - хеш пароля пользователя
	Password string `gorm:"type:varchar(255);not null" json:"-"`
	// Role - роль пользователя в системе (admin, user и т.д.)
	Role string `gorm:"type:varchar(50);not null;default:'user'" json:"role"`
	// Groups - массив идентификаторов групп, к которым принадлежит пользователь
	Groups pq.StringArray `gorm:"type:text[];default:'{}'" json:"groups"`
	// IsActive - активен ли пользователь (может входить в систему)
	IsActive bool `gorm:"default:true" json:"is_active"`
	// LastLogin - время последнего входа пользователя в систему
	LastLogin *time.Time `json:"last_login"`
	// Language - предпочитаемый язык интерфейса пользователя
	Language string `gorm:"type:varchar(10);not null;default:'en'" json:"language"`
	// DepartmentID - ID отдела, к которому относится пользователь
	DepartmentID *uuid.UUID `gorm:"type:uuid" json:"department_id"`
	// Permissions - JSON с правами доступа пользователя
	Permissions string `gorm:"type:jsonb;not null;default:'{\"can_schedule_meetings\": false, \"can_manage_department\": false, \"can_approve_recordings\": false}'" json:"permissions"`
	// FirstName - имя пользователя
	FirstName *string `gorm:"column:first_name;type:varchar(255)" json:"first_name"`
	// LastName - фамилия пользователя
	LastName *string `gorm:"column:last_name;type:varchar(255)" json:"last_name"`
	// Phone - номер телефона пользователя
	Phone *string `gorm:"column:phone;type:varchar(50)" json:"phone"`
	// Bio - краткая биография или описание пользователя
	Bio *string `gorm:"column:bio;type:text" json:"bio"`
	// AvatarURL - URL аватара пользователя
	AvatarURL *string `gorm:"column:avatar_url;type:text" json:"avatar_url"`
	// CreatedAt - время создания учетной записи пользователя
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	// UpdatedAt - время последнего обновления данных пользователя
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
	// DeletedAt - время мягкого удаления учетной записи пользователя
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	// Department - отдел, к которому относится пользователь
	Department *Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
}

// Group - группа пользователей с общими правами доступа
type Group struct {
	// ID - уникальный идентификатор группы
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	// Name - уникальное название группы
	Name string `gorm:"uniqueIndex;type:varchar(255);not null" json:"name"`
	// Description - описание группы и ее назначения
	Description string `gorm:"type:text" json:"description"`
	// Permissions - JSON с правами доступа для группы
	Permissions string `gorm:"type:jsonb;not null;default:'{}'" json:"permissions"`
	// CreatedAt - время создания группы
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	// UpdatedAt - время последнего обновления группы
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
	// DeletedAt - время мягкого удаления группы
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// GroupMembership - членство пользователя в группе (связь многие-ко-многим между пользователями и группами)
type GroupMembership struct {
	// UserID - ID пользователя, который является членом группы
	UserID uuid.UUID `gorm:"primaryKey;type:uuid;not null" json:"user_id"`
	// GroupID - ID группы, в которую входит пользователь
	GroupID uuid.UUID `gorm:"primaryKey;type:uuid;not null" json:"group_id"`
	// AddedAt - время добавления пользователя в группу
	AddedAt time.Time `gorm:"not null;default:now()" json:"added_at"`
	// AddedBy - ID пользователя, который добавил данного пользователя в группу
	AddedBy *uuid.UUID `gorm:"type:uuid" json:"added_by"`

	// Relations
	// User - пользователь, который является членом группы
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	// Group - группа, в которую входит пользователь
	Group Group `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE" json:"-"`
}

// UploadedFile - загруженный аудио/видео файл
type UploadedFile struct {
	// ID - уникальный идентификатор файла
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	// Filename - имя файла в системе хранения
	Filename string `gorm:"type:varchar(500);not null" json:"filename"`
	// OriginalName - оригинальное имя файла при загрузке
	OriginalName string `gorm:"type:varchar(500);not null" json:"original_name"`
	// FileSize - размер файла в байтах
	FileSize int64 `gorm:"not null" json:"file_size"`
	// MimeType - MIME-тип файла (audio/mp3, video/mp4 и т.д.)
	MimeType string `gorm:"type:varchar(255);not null" json:"mime_type"`
	// StoragePath - путь к файлу в системе хранения (S3, MinIO и т.д.)
	StoragePath string `gorm:"type:text;not null" json:"storage_path"`
	// UserID - ID пользователя, который загрузил файл
	UserID uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	// GroupID - ID группы, к которой относится файл
	GroupID uuid.UUID `gorm:"type:uuid;not null" json:"group_id"`
	// Status - статус обработки файла (pending, processing, completed, failed)
	Status string `gorm:"type:varchar(50);not null;default:'pending'" json:"status"`
	// TranscriptionID - ID транскрипции файла (если есть)
	TranscriptionID *uuid.UUID `gorm:"type:uuid" json:"transcription_id"`
	// Metadata - JSON с дополнительными метаданными файла
	Metadata string `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	// UploadedAt - время загрузки файла
	UploadedAt time.Time `gorm:"not null;default:now()" json:"uploaded_at"`
	// ProcessedAt - время завершения обработки файла
	ProcessedAt *time.Time `json:"processed_at"`
	// DeletedAt - время мягкого удаления файла
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	// User - пользователь, который загрузил файл
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	// Group - группа, к которой относится файл
	Group Group `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE" json:"-"`
}

// FileTranscription - транскрипция загруженного аудио/видео файла
type FileTranscription struct {
	// ID - уникальный идентификатор транскрипции
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	// FileID - ID файла, для которого выполнена транскрипция
	FileID uuid.UUID `gorm:"type:uuid;not null" json:"file_id"`
	// Text - полный текст транскрипции
	Text string `gorm:"type:text;not null" json:"text"`
	// Language - язык транскрипции (ru, en и т.д.)
	Language string `gorm:"type:varchar(10);not null" json:"language"`
	// Confidence - уровень уверенности распознавания речи (0-1)
	Confidence *float64 `gorm:"type:decimal(5,4)" json:"confidence"`
	// Duration - длительность аудио/видео в секундах
	Duration *float64 `gorm:"type:decimal(10,2)" json:"duration"`
	// Segments - JSON с сегментами транскрипции (временные метки, спикеры и т.д.)
	Segments string `gorm:"type:jsonb;default:'{}'" json:"segments"`
	// TranscribedAt - время выполнения транскрипции
	TranscribedAt time.Time `gorm:"not null;default:now()" json:"transcribed_at"`
	// TranscribedBy - ID пользователя или системы, которая выполнила транскрипцию
	TranscribedBy uuid.UUID `gorm:"type:uuid;not null" json:"transcribed_by"`
	// DeletedAt - время мягкого удаления транскрипции
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	// File - файл, для которого выполнена транскрипция
	File UploadedFile `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"-"`
}

// DocumentChunk - фрагмент документа с векторными эмбеддингами для семантического поиска
type DocumentChunk struct {
	// ID - уникальный идентификатор фрагмента
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	// FileID - ID файла, к которому относится фрагмент
	FileID uuid.UUID `gorm:"type:uuid;not null" json:"file_id"`
	// TranscriptionID - ID транскрипции, к которой относится фрагмент
	TranscriptionID *uuid.UUID `gorm:"type:uuid" json:"transcription_id"`
	// ChunkText - текст фрагмента документа
	ChunkText string `gorm:"type:text;not null" json:"chunk_text"`
	// ChunkIndex - порядковый номер фрагмента в документе
	ChunkIndex int `gorm:"not null" json:"chunk_index"`
	// Embedding - векторное представление фрагмента для семантического поиска (pgvector)
	Embedding string `gorm:"type:vector(1536)" json:"-"`
	// Metadata - JSON с дополнительными метаданными фрагмента
	Metadata string `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	// CreatedAt - время создания фрагмента
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	// DeletedAt - время мягкого удаления фрагмента
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	// File - файл, к которому относится фрагмент
	File UploadedFile `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"-"`
	// Transcription - транскрипция, к которой относится фрагмент
	Transcription *FileTranscription `gorm:"foreignKey:TranscriptionID;constraint:OnDelete:CASCADE" json:"-"`
}

// LiveKitRoom - комната LiveKit для видеоконференций
type LiveKitRoom struct {
	// ID - уникальный идентификатор комнаты
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	// Sid - уникальный идентификатор комнаты в LiveKit
	Sid string `gorm:"uniqueIndex;type:varchar(255);not null" json:"sid"`
	// Name - название комнаты
	Name string `gorm:"type:varchar(255);not null" json:"name"`
	// EmptyTimeout - таймаут закрытия пустой комнаты (в секундах)
	EmptyTimeout int `gorm:"default:300" json:"empty_timeout"`
	// DepartureTimeout - таймаут ожидания переподключения участника (в секундах)
	DepartureTimeout int `gorm:"default:20" json:"departure_timeout"`
	// CreationTime - время создания комнаты (Unix timestamp в строковом формате)
	CreationTime string `gorm:"type:varchar(50)" json:"creation_time"`
	// CreationTimeMs - время создания комнаты в миллисекундах
	CreationTimeMs string `gorm:"type:varchar(50)" json:"creation_time_ms"`
	// TurnPassword - пароль для TURN сервера
	TurnPassword string `gorm:"type:text" json:"turn_password"`
	// EnabledCodecs - JSON с включенными кодеками для комнаты
	EnabledCodecs string `gorm:"type:jsonb;default:'[]'" json:"enabled_codecs"`
	// Status - статус комнаты (active, finished)
	Status string `gorm:"type:varchar(50);not null;default:'active'" json:"status"`
	// StartedAt - время начала конференции в комнате
	StartedAt time.Time `gorm:"not null" json:"started_at"`
	// FinishedAt - время завершения конференции в комнате
	FinishedAt *time.Time `json:"finished_at"`
	// CreatedAt - время создания записи о комнате
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	// UpdatedAt - время последнего обновления записи о комнате
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
	// DeletedAt - время мягкого удаления комнаты
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// LiveKitParticipant - участник комнаты LiveKit
type LiveKitParticipant struct {
	// ID - уникальный идентификатор участника
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	// Sid - уникальный идентификатор участника в LiveKit
	Sid string `gorm:"uniqueIndex;type:varchar(255);not null" json:"sid"`
	// RoomSid - идентификатор комнаты, в которой находится участник
	RoomSid string `gorm:"type:varchar(255);not null" json:"room_sid"`
	// Identity - уникальный идентификатор личности участника
	Identity string `gorm:"type:varchar(255);not null" json:"identity"`
	// Name - отображаемое имя участника
	Name string `gorm:"type:varchar(255);not null" json:"name"`
	// State - состояние участника (ACTIVE, DISCONNECTED и т.д.)
	State string `gorm:"type:varchar(50);not null;default:'ACTIVE'" json:"state"`
	// JoinedAt - время присоединения к комнате (Unix timestamp)
	JoinedAt string `gorm:"type:varchar(50);not null" json:"joined_at"`
	// JoinedAtMs - время присоединения к комнате в миллисекундах
	JoinedAtMs string `gorm:"type:varchar(50);not null" json:"joined_at_ms"`
	// Version - версия протокола участника
	Version int `gorm:"default:0" json:"version"`
	// Permission - JSON с разрешениями участника в комнате
	Permission string `gorm:"type:jsonb;default:'{}'" json:"permission"`
	// IsPublisher - публикует ли участник медиа-потоки
	IsPublisher bool `gorm:"default:false" json:"is_publisher"`
	// DisconnectReason - причина отключения участника
	DisconnectReason string `gorm:"type:varchar(255)" json:"disconnect_reason"`
	// LeftAt - время выхода участника из комнаты
	LeftAt *time.Time `json:"left_at"`
	// CreatedAt - время создания записи об участнике
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	// UpdatedAt - время последнего обновления записи об участнике
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
	// DeletedAt - время мягкого удаления участника
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	// Room - комната, в которой находится участник
	Room LiveKitRoom `gorm:"foreignKey:RoomSid;references:Sid;constraint:OnDelete:CASCADE" json:"-"`
}

// LiveKitTrack - медиа-трек (аудио/видео поток) в комнате LiveKit
type LiveKitTrack struct {
	// ID - уникальный идентификатор трека
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	// Sid - уникальный идентификатор трека в LiveKit
	Sid string `gorm:"uniqueIndex;type:varchar(255);not null" json:"sid"`
	// ParticipantSid - идентификатор участника, которому принадлежит трек
	ParticipantSid string `gorm:"type:varchar(255);not null" json:"participant_sid"`
	// RoomSid - идентификатор комнаты, в которой находится трек
	RoomSid string `gorm:"type:varchar(255);not null" json:"room_sid"`
	// Type - тип трека (audio, video)
	Type string `gorm:"type:varchar(50)" json:"type"`
	// Source - источник трека (camera, microphone, screen_share и т.д.)
	Source string `gorm:"type:varchar(50);not null" json:"source"`
	// MimeType - MIME-тип медиа (video/h264, audio/opus и т.д.)
	MimeType string `gorm:"type:varchar(100);not null" json:"mime_type"`
	// Mid - идентификатор медиа в SDP
	Mid string `gorm:"type:varchar(50)" json:"mid"`
	// Width - ширина видео в пикселях (для видео треков)
	Width *int `json:"width"`
	// Height - высота видео в пикселях (для видео треков)
	Height *int `json:"height"`
	// Simulcast - включен ли симулкаст для трека
	Simulcast bool `gorm:"default:false" json:"simulcast"`
	// Layers - JSON с информацией о слоях симулкаста
	Layers string `gorm:"type:jsonb;default:'[]'" json:"layers"`
	// Codecs - JSON с информацией о кодеках трека
	Codecs string `gorm:"type:jsonb;default:'[]'" json:"codecs"`
	// Stream - идентификатор потока
	Stream string `gorm:"type:varchar(255)" json:"stream"`
	// Version - JSON с информацией о версии трека
	Version string `gorm:"type:jsonb;default:'{}'" json:"version"`
	// AudioFeatures - массив аудио-функций трека (stereo, red и т.д.)
	AudioFeatures pq.StringArray `gorm:"type:text[];default:'{}'" json:"audio_features"`
	// BackupCodecPolicy - политика резервного кодека
	BackupCodecPolicy string `gorm:"type:varchar(50)" json:"backup_codec_policy"`
	// Status - статус трека (published, unpublished)
	Status string `gorm:"type:varchar(50);not null;default:'published'" json:"status"`
	// PublishedAt - время публикации трека
	PublishedAt time.Time `gorm:"not null" json:"published_at"`
	// UnpublishedAt - время отмены публикации трека
	UnpublishedAt *time.Time `json:"unpublished_at"`
	// CreatedAt - время создания записи о треке
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	// UpdatedAt - время последнего обновления записи о треке
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
	// DeletedAt - время мягкого удаления трека
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	// Participant - участник, которому принадлежит трек
	Participant LiveKitParticipant `gorm:"foreignKey:ParticipantSid;references:Sid;constraint:OnDelete:CASCADE" json:"-"`
	// Room - комната, в которой находится трек
	Room LiveKitRoom `gorm:"foreignKey:RoomSid;references:Sid;constraint:OnDelete:CASCADE" json:"-"`
}

// LiveKitWebhookEvent - событие вебхука от LiveKit сервера
type LiveKitWebhookEvent struct {
	// ID - уникальный идентификатор события
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	// EventType - тип события (room_started, participant_joined, track_published и т.д.)
	EventType string `gorm:"type:varchar(100);not null" json:"event_type"`
	// EventID - уникальный идентификатор события от LiveKit
	EventID string `gorm:"type:varchar(255);not null" json:"event_id"`
	// RoomSid - идентификатор комнаты, к которой относится событие
	RoomSid *uuid.UUID `gorm:"type:uuid" json:"room_sid"`
	// ParticipantSid - идентификатор участника, к которому относится событие
	ParticipantSid *uuid.UUID `gorm:"type:uuid" json:"participant_sid"`
	// TrackSid - идентификатор трека, к которому относится событие
	TrackSid *uuid.UUID `gorm:"type:uuid" json:"track_sid"`
	// Payload - JSON с полными данными события от LiveKit
	Payload string `gorm:"type:jsonb;not null" json:"payload"`
	// CreatedAt - время получения события
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	// DeletedAt - время мягкого удаления события
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// MeetingSubject - тема/предмет встречи
type MeetingSubject struct {
	// ID - уникальный идентификатор темы встречи
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	// Name - уникальное название темы встречи
	Name string `gorm:"uniqueIndex;type:varchar(255);not null" json:"name"`
	// Description - описание темы встречи
	Description string `gorm:"type:text" json:"description"`
	// DepartmentIDs - массив ID отделов, связанных с этой темой
	DepartmentIDs pq.StringArray `gorm:"type:text[];default:'{}'" json:"department_ids"`
	// IsActive - активна ли тема встречи (может использоваться для создания новых встреч)
	IsActive bool `gorm:"not null;default:true" json:"is_active"`
	// CreatedAt - время создания темы встречи
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	// UpdatedAt - время последнего обновления темы встречи
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
	// DeletedAt - время мягкого удаления темы встречи
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Meeting - запланированная или проходящая встреча/совещание
type Meeting struct {
	// ID - уникальный идентификатор встречи
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	// Title - название встречи
	Title string `gorm:"type:varchar(500);not null" json:"title"`
	// ScheduledAt - запланированное время начала встречи
	ScheduledAt time.Time `gorm:"not null" json:"scheduled_at"`
	// Duration - длительность встречи в минутах
	Duration int `gorm:"not null" json:"duration"`
	// Recurrence - периодичность встречи (none, daily, weekly, monthly)
	Recurrence string `gorm:"type:varchar(50);not null;default:'none'" json:"recurrence"`
	// Type - тип встречи (online, offline, hybrid)
	Type string `gorm:"type:varchar(50);not null" json:"type"`
	// SubjectID - ID темы встречи
	SubjectID uuid.UUID `gorm:"type:uuid;not null" json:"subject_id"`
	// Status - статус встречи (scheduled, in_progress, completed, cancelled)
	Status string `gorm:"type:varchar(50);not null;default:'scheduled'" json:"status"`
	// NeedsVideoRecord - требуется ли видеозапись встречи
	NeedsVideoRecord bool `gorm:"not null;default:false" json:"needs_video_record"`
	// NeedsAudioRecord - требуется ли аудиозапись встречи
	NeedsAudioRecord bool `gorm:"not null;default:false" json:"needs_audio_record"`
	// ForceEndAtDuration - принудительно завершить встречу по истечении времени
	ForceEndAtDuration bool `gorm:"column:force_end_at_duration;not null;default:false" json:"force_end_at_duration"`
	// AdditionalNotes - дополнительные заметки о встрече
	AdditionalNotes string `gorm:"column:additional_notes;type:text" json:"additional_notes"`
	// LiveKitRoomID - ID комнаты LiveKit для онлайн-встречи
	LiveKitRoomID *uuid.UUID `gorm:"column:livekit_room_id;type:uuid" json:"livekit_room_id"`
	// CreatedBy - ID пользователя, который создал встречу
	CreatedBy uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	// CreatedAt - время создания встречи
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	// UpdatedAt - время последнего обновления встречи
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
	// DeletedAt - время мягкого удаления встречи
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	// Subject - тема встречи
	Subject MeetingSubject `gorm:"foreignKey:SubjectID;constraint:OnDelete:RESTRICT" json:"subject,omitempty"`
	// Creator - пользователь, создавший встречу
	Creator User `gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE" json:"creator,omitempty"`
}

// MeetingParticipant - участник встречи
type MeetingParticipant struct {
	// ID - уникальный идентификатор участника встречи
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	// MeetingID - ID встречи, в которой участвует пользователь
	MeetingID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_meeting_user" json:"meeting_id"`
	// UserID - ID пользователя-участника встречи
	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_meeting_user" json:"user_id"`
	// Role - роль участника на встрече (organizer, presenter, attendee)
	Role string `gorm:"type:varchar(50);not null" json:"role"`
	// Status - статус участника (invited, confirmed, declined, attended)
	Status string `gorm:"type:varchar(50);not null;default:'invited'" json:"status"`
	// JoinedAt - время присоединения к встрече
	JoinedAt *time.Time `json:"joined_at"`
	// LeftAt - время выхода участника из встречи
	LeftAt *time.Time `json:"left_at"`
	// CreatedAt - время добавления участника к встрече
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	// DeletedAt - время мягкого удаления участника из встречи
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	// Meeting - встреча, в которой участвует пользователь
	Meeting Meeting `gorm:"foreignKey:MeetingID;constraint:OnDelete:CASCADE" json:"-"`
	// User - пользователь-участник встречи
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// MeetingDepartment - отдел, приглашенный на встречу
type MeetingDepartment struct {
	// ID - уникальный идентификатор связи встречи и отдела
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	// MeetingID - ID встречи, на которую приглашен отдел
	MeetingID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_meeting_dept" json:"meeting_id"`
	// DepartmentID - ID отдела, приглашенного на встречу
	DepartmentID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_meeting_dept" json:"department_id"`
	// CreatedAt - время приглашения отдела на встречу
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	// DeletedAt - время мягкого удаления связи между встречей и отделом
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	// Meeting - встреча, на которую приглашен отдел
	Meeting Meeting `gorm:"foreignKey:MeetingID;constraint:OnDelete:CASCADE" json:"-"`
	// Department - отдел, приглашенный на встречу
	Department Department `gorm:"foreignKey:DepartmentID;constraint:OnDelete:CASCADE" json:"-"`
}

// PasswordResetToken - токен для сброса пароля пользователя
type PasswordResetToken struct {
	// ID - уникальный идентификатор токена
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	// UserID - ID пользователя, для которого создан токен
	UserID uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	// Email - адрес электронной почты пользователя
	Email string `gorm:"type:varchar(255);not null" json:"email"`
	// Code - 6-значный код для сброса пароля
	Code string `gorm:"type:varchar(6);not null" json:"-"`
	// ExpiresAt - время истечения срока действия токена
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	// Used - был ли использован токен для сброса пароля
	Used bool `gorm:"not null;default:false" json:"used"`
	// CreatedAt - время создания токена
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`

	// Relations
	// User - пользователь, для которого создан токен
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// LiveKitEgress - запись (egress) сессии LiveKit на диск или в облако
type LiveKitEgress struct {
	// ID - уникальный идентификатор записи (EgressID от LiveKit)
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	// RoomSID - идентификатор комнаты LiveKit, которая записывается
	RoomSID string `gorm:"type:varchar(255);not null;index" json:"room_sid"`
	// RoomName - название комнаты LiveKit
	RoomName string `gorm:"type:varchar(255);not null;index" json:"room_name"`
	// Type - тип записи (room_composite, track_composite, track)
	Type string `gorm:"type:varchar(50);not null" json:"type"`
	// Status - статус записи (pending, active, finishing, complete, failed)
	Status string `gorm:"type:varchar(50);not null;index" json:"status"`
	// TrackID - ID трека для записи (для типа track)
	TrackID *string `gorm:"type:varchar(255)" json:"track_id,omitempty"`
	// AudioTrackID - ID аудио-трека для записи (для типа track_composite)
	AudioTrackID *string `gorm:"type:varchar(255)" json:"audio_track_id,omitempty"`
	// VideoTrackID - ID видео-трека для записи (для типа track_composite)
	VideoTrackID *string `gorm:"type:varchar(255)" json:"video_track_id,omitempty"`
	// FilePath - путь к файлу записи (S3 или локальный путь)
	FilePath *string `gorm:"type:text" json:"file_path,omitempty"`
	// FileSize - размер файла записи в байтах
	FileSize *int64 `gorm:"type:bigint" json:"file_size,omitempty"`
	// Duration - длительность записи в секундах
	Duration *int64 `gorm:"type:bigint" json:"duration,omitempty"`
	// Error - сообщение об ошибке, если запись не удалась
	Error *string `gorm:"type:text" json:"error,omitempty"`
	// StartedAt - время начала записи
	StartedAt *time.Time `gorm:"type:timestamp" json:"started_at,omitempty"`
	// EndedAt - время завершения записи
	EndedAt *time.Time `gorm:"type:timestamp" json:"ended_at,omitempty"`
	// CreatedAt - время создания записи о записи
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	// UpdatedAt - время последнего обновления записи о записи
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`

	// Relations
	// Room - комната LiveKit, которая записывается
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

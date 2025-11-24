/**
 * Meeting Types - TypeScript interfaces matching backend Go models
 */

export type MeetingType = 'presentation' | 'conference';
export type MeetingStatus = 'scheduled' | 'in_progress' | 'completed' | 'cancelled';
export type MeetingRecurrence = 'none' | 'daily' | 'weekly' | 'monthly' | 'permanent';
export type ParticipantRole = 'speaker' | 'participant';
export type ParticipantStatus = 'invited' | 'accepted' | 'declined' | 'tentative';

export interface MeetingSubject {
  id: string;
  name: string;
  description: string;
  department_ids: string[];
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Meeting {
  id: string;
  title: string;
  scheduled_at: string;
  duration: number; // in minutes
  recurrence: MeetingRecurrence;
  type: MeetingType;
  subject_id?: string;
  status: MeetingStatus;
  needs_record: boolean;
  needs_transcription: boolean;
  is_recording: boolean;
  is_transcribing: boolean;
  additional_notes: string;
  is_permanent: boolean;
  allow_anonymous: boolean;
  livekit_room_id?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface User {
  id: string;
  username: string;
  email: string;
  first_name?: string;
  last_name?: string;
  avatar?: string;
  role: string;
}

export interface Department {
  id: string;
  name: string;
  description: string;
}

export interface MeetingParticipant {
  meeting_id: string;
  user_id: string;
  role: ParticipantRole;
  status: ParticipantStatus;
  joined_at?: string;
  left_at?: string;
  created_at: string;
}

export interface MeetingParticipantInfo {
  meeting_id: string;
  user_id: string;
  role: ParticipantRole;
  status: ParticipantStatus;
  joined_at?: string;
  left_at?: string;
  created_at: string;
  user: User;
}

export interface MeetingDepartment {
  meeting_id: string;
  department_id: string;
  created_at: string;
}

export interface MeetingWithDetails {
  id: string;
  title: string;
  scheduled_at: string;
  duration: number;
  recurrence: MeetingRecurrence;
  type: MeetingType;
  subject_id?: string;
  status: MeetingStatus;
  needs_record: boolean;
  needs_transcription: boolean;
  is_recording: boolean;
  is_transcribing: boolean;
  additional_notes: string;
  is_permanent: boolean;
  allow_anonymous: boolean;
  livekit_room_id?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
  subject?: MeetingSubject;
  participants: MeetingParticipantInfo[];
  departments: Department[];
  creator?: User;
  active_participants_count: number;
  anonymous_guests_count: number;
  recordings_count: number;
}

// Request types
export interface CreateMeetingRequest {
  title: string;
  scheduled_at: string;
  duration: number;
  recurrence: MeetingRecurrence;
  type: MeetingType;
  subject_id?: string;
  needs_record: boolean;
  needs_transcription: boolean;
  additional_notes?: string;
  is_permanent: boolean;
  allow_anonymous: boolean;
  speaker_id?: string; // For presentations
  participant_ids: string[];
  department_ids: string[];
}

export interface UpdateMeetingRequest {
  title?: string;
  scheduled_at?: string;
  duration?: number;
  recurrence?: MeetingRecurrence;
  type?: MeetingType;
  subject_id?: string;
  status?: MeetingStatus;
  needs_record?: boolean;
  needs_transcription?: boolean;
  additional_notes?: string;
  is_permanent?: boolean;
  allow_anonymous?: boolean;
  speaker_id?: string;
  participant_ids?: string[];
  department_ids?: string[];
}

export interface ListMeetingsRequest {
  page?: number;
  page_size?: number;
  status?: MeetingStatus;
  type?: MeetingType;
  subject_id?: string;
  date_from?: string;
  date_to?: string;
  user_id?: string;
  speaker_id?: string;
}

// Response types
export interface MeetingsResponse {
  items: MeetingWithDetails[];
  offset: number;
  page_size: number;
  total: number;
}

export interface MeetingSubjectsResponse {
  items: MeetingSubject[];
  offset: number;
  page_size: number;
  total: number;
}

export interface MeetingTokenResponse {
  token: string;
  url: string;
  roomName: string;
  participantName: string;
  meetingId: string;
  scheduledAt: string;
  duration: number;
  forceEndAt?: string;
}

export interface CreateMeetingSubjectRequest {
  name: string;
  description?: string;
  department_ids?: string[];
}

export interface UpdateMeetingSubjectRequest {
  name?: string;
  description?: string;
  department_ids?: string[];
  is_active?: boolean;
}

export interface Recording {
  id: string;
  type: 'room' | 'track';
  status: string;
  started_at: string;
  ended_at?: string;
  playlist_url: string;
  participant_id?: string;
  track_id?: string;
  participant?: User;
}

export interface TranscriptionPhrase {
  start: number; // Start time in seconds
  end: number; // End time in seconds
  text: string; // Transcribed text
  speaker?: string; // Speaker identifier (if diarization is available)
}

export interface TrackRecording {
  id: string;
  status: string;
  started_at: string;
  ended_at?: string;
  playlist_url: string;
  participant_id: string;
  track_id: string;
  participant?: User;
  transcription?: string; // URL or identifier for the transcription
  transcription_status?: string; // Transcription status: pending, processing, completed, failed
  transcription_phrases?: TranscriptionPhrase[]; // Parsed transcription phrases
  type?: string; // Track type (audio, video)
}

export interface RoomRecording {
  id: string;
  room_sid: string;
  status: string;
  started_at: string;
  ended_at?: string;
  playlist_url?: string; // Room composite recording (optional)
  audio_only?: boolean; // Whether this is an audio-only recording
  tracks: TrackRecording[];
}

export interface TrackTranscript {
  track_id: string;
  participant_id: string;
  participant?: User;
  started_at: string;
  transcription_phrases?: TranscriptionPhrase[];
}

export interface RoomTranscripts {
  room_sid: string;
  tracks: TrackTranscript[];
  memo?: string;
  memoRu?: string;
}

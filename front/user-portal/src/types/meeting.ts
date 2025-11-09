/**
 * Meeting Types - TypeScript interfaces matching backend Go models
 */

export type MeetingType = 'presentation' | 'conference';
export type MeetingStatus = 'scheduled' | 'in_progress' | 'completed' | 'cancelled';
export type MeetingRecurrence = 'none' | 'daily' | 'weekly' | 'monthly';
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
  subject_id: string;
  status: MeetingStatus;
  needs_video_record: boolean;
  needs_audio_record: boolean;
  additional_notes: string;
  force_end_at_duration: boolean;
  livekit_room_id?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface User {
  id: string;
  username: string;
  email: string;
  full_name: string;
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
  subject_id: string;
  status: MeetingStatus;
  needs_video_record: boolean;
  needs_audio_record: boolean;
  additional_notes: string;
  force_end_at_duration: boolean;
  livekit_room_id?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
  subject?: MeetingSubject;
  participants: MeetingParticipantInfo[];
  departments: Department[];
  creator?: User;
}

// Request types
export interface CreateMeetingRequest {
  title: string;
  scheduled_at: string;
  duration: number;
  recurrence: MeetingRecurrence;
  type: MeetingType;
  subject_id: string;
  needs_video_record: boolean;
  needs_audio_record: boolean;
  additional_notes?: string;
  force_end_at_duration: boolean;
  speaker_id?: string; // For presentations
  participant_user_ids: string[];
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
  needs_video_record?: boolean;
  needs_audio_record?: boolean;
  additional_notes?: string;
  force_end_at_duration?: boolean;
  speaker_id?: string;
  participant_user_ids?: string[];
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

/**
 * Meetings API Service
 * Handles all API calls related to meetings
 */

import type {
  Meeting,
  MeetingWithDetails,
  MeetingsResponse,
  MeetingSubject,
  MeetingSubjectsResponse,
  CreateMeetingRequest,
  UpdateMeetingRequest,
  ListMeetingsRequest,
  MeetingTokenResponse,
} from '../types/meeting';

const API_BASE_URL = '/api/v1';

/**
 * Get authentication token from storage
 */
const getAuthToken = (): string | null => {
  return localStorage.getItem('token') || sessionStorage.getItem('token');
};

/**
 * Build headers with authentication
 */
const getHeaders = (): HeadersInit => {
  const token = getAuthToken();
  return {
    'Content-Type': 'application/json',
    ...(token && { Authorization: `Bearer ${token}` }),
  };
};

/**
 * Handle API response errors
 */
const handleResponse = async <T>(response: Response): Promise<T> => {
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ message: 'Request failed' }));
    throw new Error(errorData.message || `HTTP ${response.status}: ${response.statusText}`);
  }
  return response.json();
};

/**
 * Build query string from object
 */
const buildQueryString = (params: Record<string, any>): string => {
  const filtered = Object.entries(params)
    .filter(([_, value]) => value !== undefined && value !== null && value !== '')
    .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`);

  return filtered.length > 0 ? `?${filtered.join('&')}` : '';
};

// ===========================
// Meeting CRUD Operations
// ===========================

/**
 * List my meetings (where I'm a participant or speaker)
 */
export const listMyMeetings = async (request: ListMeetingsRequest = {}): Promise<MeetingsResponse> => {
  const queryString = buildQueryString({
    page: request.page || 1,
    page_size: request.page_size || 20,
    status: request.status,
    type: request.type,
    subject_id: request.subject_id,
    date_from: request.date_from,
    date_to: request.date_to,
    speaker_id: request.speaker_id,
  });

  const response = await fetch(`${API_BASE_URL}/meetings${queryString}`, {
    method: 'GET',
    headers: getHeaders(),
  });

  return handleResponse<MeetingsResponse>(response);
};

/**
 * Get meeting by ID
 */
export const getMeeting = async (meetingId: string): Promise<MeetingWithDetails> => {
  const response = await fetch(`${API_BASE_URL}/meetings/${meetingId}`, {
    method: 'GET',
    headers: getHeaders(),
  });

  return handleResponse<MeetingWithDetails>(response);
};

/**
 * Create a new meeting
 */
export const createMeeting = async (request: CreateMeetingRequest): Promise<MeetingWithDetails> => {
  const response = await fetch(`${API_BASE_URL}/meetings`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify(request),
  });

  return handleResponse<MeetingWithDetails>(response);
};

/**
 * Update an existing meeting
 */
export const updateMeeting = async (
  meetingId: string,
  request: UpdateMeetingRequest
): Promise<MeetingWithDetails> => {
  const response = await fetch(`${API_BASE_URL}/meetings/${meetingId}`, {
    method: 'PUT',
    headers: getHeaders(),
    body: JSON.stringify(request),
  });

  return handleResponse<MeetingWithDetails>(response);
};

/**
 * Delete a meeting
 */
export const deleteMeeting = async (meetingId: string): Promise<{ message: string; meeting_id: string }> => {
  const response = await fetch(`${API_BASE_URL}/meetings/${meetingId}`, {
    method: 'DELETE',
    headers: getHeaders(),
  });

  return handleResponse<{ message: string; meeting_id: string }>(response);
};

/**
 * Get LiveKit token for joining a meeting
 */
export const getMeetingToken = async (meetingId: string): Promise<MeetingTokenResponse> => {
  const response = await fetch(`${API_BASE_URL}/meetings/${meetingId}/token`, {
    method: 'GET',
    headers: getHeaders(),
  });

  return handleResponse<MeetingTokenResponse>(response);
};

// ===========================
// Recording Control Operations
// ===========================

/**
 * Start recording for a meeting
 */
export const startRecording = async (meetingId: string, audioOnly: boolean = false): Promise<MeetingWithDetails> => {
  const response = await fetch(`${API_BASE_URL}/meetings/${meetingId}/recording/start`, {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify({ audio_only: audioOnly }),
  });

  return handleResponse<MeetingWithDetails>(response);
};

/**
 * Stop recording for a meeting
 */
export const stopRecording = async (meetingId: string): Promise<MeetingWithDetails> => {
  const response = await fetch(`${API_BASE_URL}/meetings/${meetingId}/recording/stop`, {
    method: 'POST',
    headers: getHeaders(),
  });

  return handleResponse<MeetingWithDetails>(response);
};

/**
 * Start transcription for a meeting
 */
export const startTranscription = async (meetingId: string): Promise<MeetingWithDetails> => {
  const response = await fetch(`${API_BASE_URL}/meetings/${meetingId}/transcription/start`, {
    method: 'POST',
    headers: getHeaders(),
  });

  return handleResponse<MeetingWithDetails>(response);
};

/**
 * Stop transcription for a meeting
 */
export const stopTranscription = async (meetingId: string): Promise<MeetingWithDetails> => {
  const response = await fetch(`${API_BASE_URL}/meetings/${meetingId}/transcription/stop`, {
    method: 'POST',
    headers: getHeaders(),
  });

  return handleResponse<MeetingWithDetails>(response);
};

// ===========================
// Meeting Subjects Operations
// ===========================

/**
 * List meeting subjects
 */
export const listMeetingSubjects = async (
  page: number = 1,
  pageSize: number = 100,
  departmentId?: string,
  includeInactive: boolean = false
): Promise<MeetingSubjectsResponse> => {
  const queryString = buildQueryString({
    page,
    page_size: pageSize,
    department_id: departmentId,
    include_inactive: includeInactive,
  });

  const response = await fetch(`${API_BASE_URL}/meeting-subjects${queryString}`, {
    method: 'GET',
    headers: getHeaders(),
  });

  return handleResponse<MeetingSubjectsResponse>(response);
};

/**
 * Get meeting subject by ID
 */
export const getMeetingSubject = async (subjectId: string): Promise<MeetingSubject> => {
  const response = await fetch(`${API_BASE_URL}/meeting-subjects/${subjectId}`, {
    method: 'GET',
    headers: getHeaders(),
  });

  return handleResponse<MeetingSubject>(response);
};

// ===========================
// Utility Functions
// ===========================

/**
 * Format meeting date for display
 * Dates from backend are in UTC ISO format, displayed in user's local timezone with 24-hour format
 */
export const formatMeetingDate = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false, // Use 24-hour format
  });
};

/**
 * Format meeting duration
 */
export const formatDuration = (minutes: number): string => {
  if (minutes < 60) {
    return `${minutes}m`;
  }
  const hours = Math.floor(minutes / 60);
  const mins = minutes % 60;
  return mins > 0 ? `${hours}h ${mins}m` : `${hours}h`;
};

/**
 * Get meeting status display name and color class
 */
export const getMeetingStatusInfo = (status: string): { label: string; className: string } => {
  switch (status) {
    case 'scheduled':
      return { label: 'Scheduled', className: 'status-scheduled' };
    case 'in_progress':
      return { label: 'In Progress', className: 'status-in-progress' };
    case 'completed':
      return { label: 'Completed', className: 'status-completed' };
    case 'cancelled':
      return { label: 'Cancelled', className: 'status-cancelled' };
    default:
      return { label: status, className: 'status-default' };
  }
};

/**
 * Get meeting type display name and icon
 */
export const getMeetingTypeInfo = (type: string): { label: string; icon: string } => {
  switch (type) {
    case 'presentation':
      return { label: 'Presentation', icon: '' };
    case 'conference':
      return { label: 'Conference', icon: '' };
    default:
      return { label: type, icon: '' };
  }
};

/**
 * Check if meeting is upcoming (scheduled and in the future)
 */
export const isMeetingUpcoming = (meeting: Meeting | MeetingWithDetails): boolean => {
  return meeting.status === 'scheduled' && new Date(meeting.scheduled_at) > new Date();
};

/**
 * Check if meeting is happening now (in_progress or scheduled time is within duration window)
 */
export const isMeetingNow = (meeting: Meeting | MeetingWithDetails): boolean => {
  if (meeting.status === 'in_progress') {
    return true;
  }

  const now = new Date();
  const scheduledTime = new Date(meeting.scheduled_at);
  const endTime = new Date(scheduledTime.getTime() + meeting.duration * 60000);

  return now >= scheduledTime && now <= endTime && meeting.status === 'scheduled';
};

/**
 * Check if meeting is past (completed or cancelled, or scheduled time has passed)
 */
export const isMeetingPast = (meeting: Meeting | MeetingWithDetails): boolean => {
  if (meeting.status === 'completed' || meeting.status === 'cancelled') {
    return true;
  }

  const now = new Date();
  const scheduledTime = new Date(meeting.scheduled_at);
  const endTime = new Date(scheduledTime.getTime() + meeting.duration * 60000);

  return now > endTime && meeting.status === 'scheduled';
};

/**
 * Get participant role display
 */
export const getParticipantRoleLabel = (role: string): string => {
  switch (role) {
    case 'speaker':
      return 'Speaker';
    case 'participant':
      return 'Participant';
    default:
      return role;
  }
};

/**
 * Get participant status display
 */
export const getParticipantStatusInfo = (status: string): { label: string; className: string } => {
  switch (status) {
    case 'invited':
      return { label: 'Invited', className: 'participant-invited' };
    case 'accepted':
      return { label: 'Accepted', className: 'participant-accepted' };
    case 'declined':
      return { label: 'Declined', className: 'participant-declined' };
    case 'tentative':
      return { label: 'Tentative', className: 'participant-tentative' };
    default:
      return { label: status, className: 'participant-default' };
  }
};

/**
 * Get recordings for a meeting
 */
export const getMeetingRecordings = async (meetingId: string): Promise<any[]> => {
  const response = await fetch(`${API_BASE_URL}/meetings/${meetingId}/recordings`, {
    method: 'GET',
    headers: getHeaders(),
  });

  return handleResponse<any[]>(response);
};

const API_BASE_URL = import.meta.env.VITE_API_URL || '';

export interface MeetingSubject {
  id: string;
  name: string;
  description: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateMeetingSubjectRequest {
  name: string;
  description?: string;
}

export interface UpdateMeetingSubjectRequest {
  name?: string;
  description?: string;
  is_active?: boolean;
}

export interface MeetingSubjectsResponse {
  items: MeetingSubject[];
  total: number;
  page: number;
  page_size: number;
}

const getAuthHeaders = () => {
  const token = localStorage.getItem('token') || sessionStorage.getItem('token');
  return {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };
};

export const meetingSubjectsApi = {
  async getSubjects(page: number = 1, pageSize: number = 50): Promise<MeetingSubjectsResponse> {
    const response = await fetch(`${API_BASE_URL}/api/v1/meeting-subjects?page=${page}&page_size=${pageSize}`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch meeting subjects');
    }

    return response.json();
  },

  async getSubject(id: string): Promise<MeetingSubject> {
    const response = await fetch(`${API_BASE_URL}/api/v1/meeting-subjects/${id}`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch meeting subject');
    }

    return response.json();
  },

  async createSubject(data: CreateMeetingSubjectRequest): Promise<MeetingSubject> {
    const response = await fetch(`${API_BASE_URL}/api/v1/meeting-subjects`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to create meeting subject');
    }

    return response.json();
  },

  async updateSubject(id: string, data: UpdateMeetingSubjectRequest): Promise<MeetingSubject> {
    const response = await fetch(`${API_BASE_URL}/api/v1/meeting-subjects/${id}`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to update meeting subject');
    }

    return response.json();
  },

  async deleteSubject(id: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/api/v1/meeting-subjects/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to delete meeting subject');
    }
  },
};

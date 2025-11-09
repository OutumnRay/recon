const API_BASE_URL = import.meta.env.VITE_API_URL || '';

export interface Room {
  id: string;
  sid: string;
  name: string;
  status: string;
  startedAt: string;
  finishedAt?: string;
  emptyTimeout: number;
  departureTimeout: number;
  creationTime: string;
}

export interface Participant {
  id: string;
  sid: string;
  identity: string;
  name: string;
  state: string;
  joinedAt: string;
  leftAt?: string;
  isPublisher: boolean;
  disconnectReason?: string;
}

export interface Track {
  id: string;
  sid: string;
  type: string;
  source: string;
  mimeType: string;
  status: string;
  publishedAt: string;
  unpublishedAt?: string;
  width?: number;
  height?: number;
  simulcast?: boolean;
}

const getAuthHeaders = () => {
  const token = localStorage.getItem('token') || sessionStorage.getItem('token');
  return {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };
};

export const liveKitApi = {
  async getRooms(status?: string, limit: number = 50, offset: number = 0): Promise<Room[]> {
    const params = new URLSearchParams();
    if (status) params.append('status', status);
    params.append('limit', limit.toString());
    params.append('offset', offset.toString());

    const response = await fetch(`${API_BASE_URL}/api/v1/livekit/rooms?${params}`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch rooms');
    }

    return response.json();
  },

  async getRoom(sid: string): Promise<Room> {
    const response = await fetch(`${API_BASE_URL}/api/v1/livekit/rooms?sid=${sid}`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch room');
    }

    return response.json();
  },

  async getParticipants(roomSid: string): Promise<Participant[]> {
    const response = await fetch(`${API_BASE_URL}/api/v1/livekit/participants?room_sid=${roomSid}`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch participants');
    }

    return response.json();
  },

  async getTracks(roomSid: string): Promise<Track[]> {
    const response = await fetch(`${API_BASE_URL}/api/v1/livekit/tracks?room_sid=${roomSid}`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch tracks');
    }

    return response.json();
  },

  async getWebhookEvents(eventType?: string, roomSid?: string, limit: number = 100, offset: number = 0) {
    const params = new URLSearchParams();
    if (eventType) params.append('event_type', eventType);
    if (roomSid) params.append('room_sid', roomSid);
    params.append('limit', limit.toString());
    params.append('offset', offset.toString());

    const response = await fetch(`${API_BASE_URL}/api/v1/livekit/webhook-events?${params}`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch webhook events');
    }

    return response.json();
  },
};

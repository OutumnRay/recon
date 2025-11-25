/**
 * Real-time Notification Service using WebSocket
 * Connects to /api/v1/notifications/ws and receives real-time updates
 */

export type EventType =
  | 'meeting.status_changed'
  | 'meeting.updated'
  | 'recording.started'
  | 'recording.completed'
  | 'recording.failed'
  | 'transcription.started'
  | 'transcription.completed'
  | 'transcription.failed'
  | 'summary.started'
  | 'summary.completed'
  | 'summary.failed'
  | 'composite_video.started'
  | 'composite_video.completed'
  | 'composite_video.failed'
  | 'system.connected'
  | 'system.pong';

export interface Notification {
  id: string;
  type: EventType;
  entity_type: string; // meeting, recording, transcription, summary
  entity_id: string;
  meeting_id?: string;
  user_id?: string;
  changed_fields?: Record<string, any>;
  timestamp: string;
  message?: string;
}

export type NotificationHandler = (notification: Notification) => void;

export class NotificationService {
  private ws: WebSocket | null = null;
  private handlers: Set<NotificationHandler> = new Set();
  private reconnectTimeout: NodeJS.Timeout | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private baseReconnectDelay = 1000; // 1 second
  private isIntentionallyClosed = false;

  constructor(private baseUrl: string, private getToken: () => string | null) {}

  /**
   * Connect to the WebSocket endpoint
   */
  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      console.log('[NotificationService] Already connected');
      return;
    }

    const token = this.getToken();
    if (!token) {
      console.error('[NotificationService] No auth token available');
      return;
    }

    this.isIntentionallyClosed = false;

    // Build WebSocket URL
    const protocol = this.baseUrl.startsWith('https') ? 'wss' : 'ws';
    const wsUrl = `${protocol}://${this.baseUrl.replace(/^https?:\/\//, '')}/api/v1/notifications/ws`;

    console.log('[NotificationService] Connecting to:', wsUrl);

    try {
      this.ws = new WebSocket(wsUrl);

      // Send auth token after connection
      this.ws.onopen = () => {
        console.log('[NotificationService] ✅ Connected to notification service');
        this.reconnectAttempts = 0;

        // Send authentication in first message
        this.send({ type: 'auth', token });
      };

      this.ws.onmessage = (event) => {
        try {
          const notification: Notification = JSON.parse(event.data);
          console.log('[NotificationService] 📨 Received notification:', notification);

          // Notify all registered handlers
          this.handlers.forEach((handler) => {
            try {
              handler(notification);
            } catch (error) {
              console.error('[NotificationService] Error in notification handler:', error);
            }
          });
        } catch (error) {
          console.error('[NotificationService] Failed to parse notification:', error);
        }
      };

      this.ws.onerror = (error) => {
        console.error('[NotificationService] ❌ WebSocket error:', error);
      };

      this.ws.onclose = (event) => {
        console.log('[NotificationService] 🔌 Disconnected:', event.code, event.reason);
        this.ws = null;

        // Attempt to reconnect if not intentionally closed
        if (!this.isIntentionallyClosed) {
          this.scheduleReconnect();
        }
      };
    } catch (error) {
      console.error('[NotificationService] Failed to create WebSocket:', error);
      this.scheduleReconnect();
    }
  }

  /**
   * Disconnect from the WebSocket
   */
  disconnect(): void {
    console.log('[NotificationService] Disconnecting...');
    this.isIntentionallyClosed = true;

    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.reconnectAttempts = 0;
  }

  /**
   * Schedule reconnection attempt with exponential backoff
   */
  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('[NotificationService] Max reconnect attempts reached');
      return;
    }

    if (this.reconnectTimeout) {
      return; // Already scheduled
    }

    const delay = Math.min(
      this.baseReconnectDelay * Math.pow(2, this.reconnectAttempts),
      30000 // Max 30 seconds
    );

    console.log(
      `[NotificationService] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts + 1}/${this.maxReconnectAttempts})`
    );

    this.reconnectTimeout = setTimeout(() => {
      this.reconnectTimeout = null;
      this.reconnectAttempts++;
      this.connect();
    }, delay);
  }

  /**
   * Subscribe to notifications
   */
  subscribe(handler: NotificationHandler): () => void {
    this.handlers.add(handler);

    // Return unsubscribe function
    return () => {
      this.handlers.delete(handler);
    };
  }

  /**
   * Send a message to the server
   */
  private send(data: any): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  /**
   * Send ping to keep connection alive
   */
  ping(): void {
    this.send({ type: 'ping' });
  }
}

// Singleton instance
let notificationService: NotificationService | null = null;

/**
 * Get or create the notification service instance
 */
export function getNotificationService(
  baseUrl: string,
  getToken: () => string | null
): NotificationService {
  if (!notificationService) {
    notificationService = new NotificationService(baseUrl, getToken);
  }
  return notificationService;
}

/**
 * React hook for using notifications
 */
export function useNotificationService(
  baseUrl: string,
  getToken: () => string | null,
  handler?: NotificationHandler
): NotificationService {
  const service = getNotificationService(baseUrl, getToken);

  // Subscribe to notifications if handler provided
  if (handler) {
    const unsubscribe = service.subscribe(handler);
    // Note: In a real React hook, this should be wrapped in useEffect with cleanup
    return service;
  }

  return service;
}

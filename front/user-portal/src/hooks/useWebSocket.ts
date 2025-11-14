import { useEffect, useRef, useCallback, useState } from 'react';

export interface WSMessage {
  type: string;
  meeting_id?: string;
  user_id?: string;
  data?: any;
  timestamp: string;
}

interface UseWebSocketOptions {
  meetingId: string;
  enabled?: boolean;
  onMessage?: (message: WSMessage) => void;
  onConnect?: () => void;
  onDisconnect?: () => void;
  onError?: (error: Event) => void;
}

export function useWebSocket({
  meetingId,
  enabled = true,
  onMessage,
  onConnect,
  onDisconnect,
  onError,
}: UseWebSocketOptions) {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<number | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<WSMessage | null>(null);

  const connect = useCallback(() => {
    if (!enabled || !meetingId) return;

    // Get auth token from localStorage
    const token = localStorage.getItem('token');
    if (!token) {
      // Silently skip connection if no token - this is expected before login
      return;
    }

    // Determine WebSocket protocol based on window location
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//${window.location.host}/api/v1/meetings/${meetingId}/ws`;

    console.log('[WebSocket] Connecting to:', wsUrl);

    try {
      const ws = new WebSocket(wsUrl);

      ws.onopen = () => {
        console.log('[WebSocket] Connected');
        setIsConnected(true);
        if (onConnect) onConnect();

        // Clear any pending reconnect timeout
        if (reconnectTimeoutRef.current) {
          clearTimeout(reconnectTimeoutRef.current);
          reconnectTimeoutRef.current = null;
        }
      };

      ws.onmessage = (event) => {
        try {
          const message: WSMessage = JSON.parse(event.data);
          console.log('[WebSocket] Message received:', message);
          setLastMessage(message);
          if (onMessage) onMessage(message);
        } catch (err) {
          console.error('[WebSocket] Failed to parse message:', err);
        }
      };

      ws.onerror = (error) => {
        console.error('[WebSocket] Error:', error);
        if (onError) onError(error);
      };

      ws.onclose = () => {
        console.log('[WebSocket] Disconnected');
        setIsConnected(false);
        if (onDisconnect) onDisconnect();

        // Attempt to reconnect after 3 seconds
        if (enabled) {
          console.log('[WebSocket] Scheduling reconnect in 3 seconds...');
          reconnectTimeoutRef.current = window.setTimeout(() => {
            console.log('[WebSocket] Attempting to reconnect...');
            connect();
          }, 3000);
        }
      };

      wsRef.current = ws;
    } catch (err) {
      console.error('[WebSocket] Failed to create WebSocket:', err);
    }
  }, [meetingId, enabled, onConnect, onMessage, onDisconnect, onError]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (wsRef.current) {
      console.log('[WebSocket] Disconnecting...');
      wsRef.current.close();
      wsRef.current = null;
    }
    setIsConnected(false);
  }, []);

  const sendMessage = useCallback((type: string, data?: any) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      console.warn('[WebSocket] Cannot send message - not connected');
      return false;
    }

    const message: WSMessage = {
      type,
      meeting_id: meetingId,
      data,
      timestamp: new Date().toISOString(),
    };

    try {
      wsRef.current.send(JSON.stringify(message));
      console.log('[WebSocket] Message sent:', message);
      return true;
    } catch (err) {
      console.error('[WebSocket] Failed to send message:', err);
      return false;
    }
  }, [meetingId]);

  useEffect(() => {
    if (enabled) {
      connect();
    }

    return () => {
      disconnect();
    };
  }, [enabled, connect, disconnect]);

  return {
    isConnected,
    lastMessage,
    sendMessage,
    disconnect,
  };
}

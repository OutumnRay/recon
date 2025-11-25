import 'dart:async';
import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:web_socket_channel/status.dart' as status;

/// Event types for real-time notifications
enum NotificationEventType {
  meetingStatusChanged('meeting.status_changed'),
  meetingUpdated('meeting.updated'),
  recordingStarted('recording.started'),
  recordingCompleted('recording.completed'),
  recordingFailed('recording.failed'),
  transcriptionStarted('transcription.started'),
  transcriptionCompleted('transcription.completed'),
  transcriptionFailed('transcription.failed'),
  summaryStarted('summary.started'),
  summaryCompleted('summary.completed'),
  summaryFailed('summary.failed'),
  compositeVideoStarted('composite_video.started'),
  compositeVideoCompleted('composite_video.completed'),
  compositeVideoFailed('composite_video.failed'),
  systemConnected('system.connected'),
  systemPong('system.pong');

  const NotificationEventType(this.value);
  final String value;

  static NotificationEventType? fromString(String value) {
    for (var type in NotificationEventType.values) {
      if (type.value == value) return type;
    }
    return null;
  }
}

/// Real-time notification model
class Notification {
  final String id;
  final NotificationEventType type;
  final String entityType;
  final String entityId;
  final String? meetingId;
  final String? userId;
  final Map<String, dynamic>? changedFields;
  final DateTime timestamp;
  final String? message;

  Notification({
    required this.id,
    required this.type,
    required this.entityType,
    required this.entityId,
    this.meetingId,
    this.userId,
    this.changedFields,
    required this.timestamp,
    this.message,
  });

  factory Notification.fromJson(Map<String, dynamic> json) {
    return Notification(
      id: json['id'] as String,
      type: NotificationEventType.fromString(json['type'] as String) ??
          NotificationEventType.systemConnected,
      entityType: json['entity_type'] as String? ?? '',
      entityId: json['entity_id'] as String? ?? '',
      meetingId: json['meeting_id'] as String?,
      userId: json['user_id'] as String?,
      changedFields: json['changed_fields'] as Map<String, dynamic>?,
      timestamp: json['timestamp'] != null
          ? DateTime.parse(json['timestamp'] as String)
          : DateTime.now(),
      message: json['message'] as String?,
    );
  }
}

/// Callback for notification events
typedef NotificationHandler = void Function(Notification notification);

/// WebSocket-based real-time notification service
class NotificationService {
  final String baseUrl;
  final String Function() getToken;

  WebSocketChannel? _channel;
  StreamSubscription? _subscription;
  Timer? _reconnectTimer;
  Timer? _pingTimer;
  int _reconnectAttempts = 0;
  static const int _maxReconnectAttempts = 10;
  static const int _baseReconnectDelay = 1000; // milliseconds
  bool _isIntentionallyClosed = false;

  final List<NotificationHandler> _handlers = [];

  NotificationService({
    required this.baseUrl,
    required this.getToken,
  });

  /// Connect to the WebSocket endpoint
  Future<void> connect() async {
    if (_channel != null) {
      debugPrint('[NotificationService] Already connected');
      return;
    }

    final token = getToken();
    if (token.isEmpty) {
      debugPrint('[NotificationService] ❌ No auth token available');
      return;
    }

    _isIntentionallyClosed = false;

    // Build WebSocket URL
    final protocol = baseUrl.startsWith('https') ? 'wss' : 'ws';
    final wsUrl = '$protocol://${baseUrl.replaceAll(RegExp(r'^https?://'), '')}/api/v1/notifications/ws';

    debugPrint('[NotificationService] Connecting to: $wsUrl');

    try {
      // Note: Authorization header doesn't work with WebSocket in browsers
      // The server should handle auth through a separate mechanism or query params
      _channel = WebSocketChannel.connect(
        Uri.parse(wsUrl),
      );

      debugPrint('[NotificationService] ✅ Connected to notification service');
      _reconnectAttempts = 0;

      // Send authentication message
      _sendMessage({'type': 'auth', 'token': token});

      // Listen to messages
      _subscription = _channel!.stream.listen(
        _handleMessage,
        onError: (error) {
          debugPrint('[NotificationService] ❌ Error: $error');
          _handleDisconnect();
        },
        onDone: () {
          debugPrint('[NotificationService] 🔌 Disconnected');
          _handleDisconnect();
        },
      );

      // Start ping timer to keep connection alive
      _startPingTimer();
    } catch (error) {
      debugPrint('[NotificationService] ❌ Failed to connect: $error');
      _scheduleReconnect();
    }
  }

  /// Disconnect from the WebSocket
  void disconnect() {
    debugPrint('[NotificationService] Disconnecting...');
    _isIntentionallyClosed = true;

    _stopPingTimer();
    _stopReconnectTimer();

    _subscription?.cancel();
    _subscription = null;

    _channel?.sink.close(status.goingAway);
    _channel = null;

    _reconnectAttempts = 0;
  }

  /// Subscribe to notifications
  void Function() subscribe(NotificationHandler handler) {
    _handlers.add(handler);

    // Return unsubscribe function
    return () {
      _handlers.remove(handler);
    };
  }

  /// Check if connected
  bool get isConnected => _channel != null;

  /// Send a ping message
  void ping() {
    _sendMessage({'type': 'ping'});
  }

  /// Handle incoming messages
  void _handleMessage(dynamic data) {
    try {
      final json = jsonDecode(data as String) as Map<String, dynamic>;
      final notification = Notification.fromJson(json);

      debugPrint('[NotificationService] 📨 Received: ${notification.type.value}');

      // Notify all handlers
      for (final handler in _handlers) {
        try {
          handler(notification);
        } catch (error) {
          debugPrint('[NotificationService] Error in handler: $error');
        }
      }
    } catch (error) {
      debugPrint('[NotificationService] Failed to parse notification: $error');
    }
  }

  /// Handle disconnection
  void _handleDisconnect() {
    _subscription?.cancel();
    _subscription = null;
    _channel = null;
    _stopPingTimer();

    if (!_isIntentionallyClosed) {
      _scheduleReconnect();
    }
  }

  /// Schedule reconnection attempt
  void _scheduleReconnect() {
    if (_reconnectAttempts >= _maxReconnectAttempts) {
      debugPrint('[NotificationService] ❌ Max reconnect attempts reached');
      return;
    }

    if (_reconnectTimer != null) {
      return; // Already scheduled
    }

    final delay = Duration(
      milliseconds: (_baseReconnectDelay *
              (1 << _reconnectAttempts).clamp(1, 30))
          .toInt(),
    );

    debugPrint(
        '[NotificationService] Reconnecting in ${delay.inSeconds}s (attempt ${_reconnectAttempts + 1}/$_maxReconnectAttempts)');

    _reconnectTimer = Timer(delay, () {
      _reconnectTimer = null;
      _reconnectAttempts++;
      connect();
    });
  }

  /// Stop reconnect timer
  void _stopReconnectTimer() {
    _reconnectTimer?.cancel();
    _reconnectTimer = null;
  }

  /// Start ping timer
  void _startPingTimer() {
    _pingTimer = Timer.periodic(const Duration(seconds: 30), (timer) {
      if (_channel != null) {
        ping();
      }
    });
  }

  /// Stop ping timer
  void _stopPingTimer() {
    _pingTimer?.cancel();
    _pingTimer = null;
  }

  /// Send a message to the server
  void _sendMessage(Map<String, dynamic> data) {
    if (_channel != null) {
      _channel!.sink.add(jsonEncode(data));
    }
  }

  /// Dispose the service
  void dispose() {
    disconnect();
  }
}

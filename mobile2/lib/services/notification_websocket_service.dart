import 'dart:async';
import 'dart:convert';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'storage_service.dart';
import 'config_service.dart';
import '../utils/logger.dart';

/// Типы событий уведомлений
class NotificationEventType {
  // Meeting events
  static const String meetingStatusChanged = 'meeting.status_changed';
  static const String meetingUpdated = 'meeting.updated';
  static const String meetingParticipantJoin = 'meeting.participant_join';
  static const String meetingParticipantLeave = 'meeting.participant_leave';

  // Recording events
  static const String recordingStarted = 'recording.started';
  static const String recordingCompleted = 'recording.completed';
  static const String recordingFailed = 'recording.failed';
  static const String recordingProcessing = 'recording.processing';

  // Transcription events
  static const String transcriptionStarted = 'transcription.started';
  static const String transcriptionCompleted = 'transcription.completed';
  static const String transcriptionFailed = 'transcription.failed';
  static const String transcriptionProgress = 'transcription.progress';

  // Summary events
  static const String summaryStarted = 'summary.started';
  static const String summaryCompleted = 'summary.completed';
  static const String summaryFailed = 'summary.failed';
  static const String summaryProgress = 'summary.progress';

  // Composite video events
  static const String compositeVideoStarted = 'composite_video.started';
  static const String compositeVideoCompleted = 'composite_video.completed';
  static const String compositeVideoFailed = 'composite_video.failed';

  // System events
  static const String systemConnected = 'system.connected';
  static const String systemPong = 'system.pong';
}

/// Модель уведомления
class NotificationMessage {
  final String id;
  final String type;
  final String? entityType;
  final String? entityId;
  final String? meetingId;
  final String? userId;
  final Map<String, dynamic>? changedFields;
  final DateTime timestamp;
  final String? message;

  NotificationMessage({
    required this.id,
    required this.type,
    this.entityType,
    this.entityId,
    this.meetingId,
    this.userId,
    this.changedFields,
    required this.timestamp,
    this.message,
  });

  factory NotificationMessage.fromJson(Map<String, dynamic> json) {
    return NotificationMessage(
      id: json['id'] as String? ?? '',
      type: json['type'] as String? ?? '',
      entityType: json['entity_type'] as String?,
      entityId: json['entity_id'] as String?,
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

/// Сервис для WebSocket уведомлений
class NotificationWebSocketService {
  WebSocketChannel? _channel;
  final _notificationController = StreamController<NotificationMessage>.broadcast();
  final _connectionStateController = StreamController<bool>.broadcast();
  Timer? _reconnectTimer;
  Timer? _pingTimer;
  bool _isConnected = false;
  bool _shouldReconnect = true;
  int _reconnectAttempts = 0;
  static const int _maxReconnectAttempts = 5;
  static const Duration _reconnectDelay = Duration(seconds: 5);
  static const Duration _pingInterval = Duration(seconds: 30);

  /// Поток уведомлений
  Stream<NotificationMessage> get notifications => _notificationController.stream;

  /// Поток состояния подключения
  Stream<bool> get connectionState => _connectionStateController.stream;

  /// Текущее состояние подключения
  bool get isConnected => _isConnected;

  /// Подключиться к WebSocket серверу уведомлений
  Future<void> connect() async {
    if (_isConnected) {
      Logger.logInfo('WebSocket already connected');
      return;
    }

    _shouldReconnect = true;

    try {
      final storageService = StorageService();
      final token = await storageService.getToken();

      if (token == null) {
        Logger.logError('No access token for WebSocket connection');
        return;
      }

      final configService = ConfigService();
      final apiUrl = await configService.getApiUrl();

      // Преобразуем HTTP URL в WebSocket URL
      final wsUrl = apiUrl
          .replaceFirst('https://', 'wss://')
          .replaceFirst('http://', 'ws://')
          .replaceFirst(RegExp(r'/api/v\d+$'), '/api/v1/notifications/ws');

      Logger.logInfo('Connecting to WebSocket: $wsUrl');

      _channel = WebSocketChannel.connect(
        Uri.parse(wsUrl),
        protocols: ['Authorization', token],
      );

      // Слушаем входящие сообщения
      _channel!.stream.listen(
        _handleMessage,
        onError: _handleError,
        onDone: _handleDone,
      );

      // Отправляем авторизационный токен в первом сообщении
      _channel!.sink.add(jsonEncode({
        'type': 'auth',
        'token': token,
      }));

      _isConnected = true;
      _reconnectAttempts = 0;
      _connectionStateController.add(true);

      // Запускаем ping для поддержания соединения
      _startPingTimer();

      Logger.logSuccess('WebSocket connected successfully');
    } catch (e) {
      Logger.logError('Failed to connect WebSocket: $e');
      _handleReconnect();
    }
  }

  /// Обработка входящих сообщений
  void _handleMessage(dynamic message) {
    try {
      final data = jsonDecode(message as String) as Map<String, dynamic>;
      final notification = NotificationMessage.fromJson(data);

      Logger.logInfo('WebSocket notification received: ${notification.type}');

      // Обрабатываем системные сообщения
      if (notification.type == NotificationEventType.systemConnected) {
        Logger.logSuccess('WebSocket connection confirmed by server');
        return;
      }

      if (notification.type == NotificationEventType.systemPong) {
        return; // Игнорируем pong
      }

      // Отправляем уведомление подписчикам
      _notificationController.add(notification);
    } catch (e) {
      Logger.logError('Failed to parse WebSocket message: $e');
    }
  }

  /// Обработка ошибок
  void _handleError(dynamic error) {
    Logger.logError('WebSocket error: $error');
    _setDisconnected();
    _handleReconnect();
  }

  /// Обработка закрытия соединения
  void _handleDone() {
    Logger.logInfo('WebSocket connection closed');
    _setDisconnected();
    _handleReconnect();
  }

  /// Установка состояния отключения
  void _setDisconnected() {
    _isConnected = false;
    _connectionStateController.add(false);
    _pingTimer?.cancel();
    _pingTimer = null;
  }

  /// Переподключение
  void _handleReconnect() {
    if (!_shouldReconnect) return;

    _reconnectAttempts++;

    if (_reconnectAttempts > _maxReconnectAttempts) {
      Logger.logError('Max WebSocket reconnect attempts reached');
      return;
    }

    Logger.logInfo('Attempting WebSocket reconnect in ${_reconnectDelay.inSeconds}s (attempt $_reconnectAttempts)');

    _reconnectTimer?.cancel();
    _reconnectTimer = Timer(_reconnectDelay, () {
      connect();
    });
  }

  /// Запуск таймера ping
  void _startPingTimer() {
    _pingTimer?.cancel();
    _pingTimer = Timer.periodic(_pingInterval, (_) {
      if (_isConnected && _channel != null) {
        try {
          _channel!.sink.add(jsonEncode({'type': 'ping'}));
        } catch (e) {
          Logger.logError('Failed to send ping: $e');
        }
      }
    });
  }

  /// Отключиться от WebSocket
  void disconnect() {
    _shouldReconnect = false;
    _reconnectTimer?.cancel();
    _pingTimer?.cancel();

    if (_channel != null) {
      _channel!.sink.close();
      _channel = null;
    }

    _setDisconnected();
    Logger.logInfo('WebSocket disconnected');
  }

  /// Освобождение ресурсов
  void dispose() {
    disconnect();
    _notificationController.close();
    _connectionStateController.close();
  }
}

/// Глобальный экземпляр сервиса уведомлений
final notificationWebSocketService = NotificationWebSocketService();

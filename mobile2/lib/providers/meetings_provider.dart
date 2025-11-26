import 'dart:async';
import 'package:flutter/foundation.dart';
import '../models/meeting.dart';
import '../models/recording.dart';
import '../models/transcript.dart';
import '../services/meetings_service.dart';
import '../services/api_client.dart';
import '../services/config_service.dart';
import '../services/notification_websocket_service.dart';
import '../utils/logger.dart';
import '../main.dart';

/// Глобальный провайдер для управления данными встреч с реактивными обновлениями
/// через WebSocket уведомления.
class MeetingsProvider extends ChangeNotifier {
  final ConfigService _configService = ConfigService();
  MeetingsService? _meetingsService;
  StreamSubscription<NotificationMessage>? _notificationSubscription;

  // Кэш встреч
  final Map<String, MeetingWithDetails> _meetingsCache = {};

  // Кэш записей для каждой встречи
  final Map<String, List<RoomRecording>> _recordingsCache = {};

  // Кэш транскриптов для каждой комнаты
  final Map<String, RoomTranscripts> _transcriptsCache = {};

  // Состояния загрузки
  bool _isLoadingMeetings = false;
  String? _meetingsError;

  // Список встреч (для экрана списка)
  List<MeetingWithDetails> _meetings = [];
  List<MeetingWithDetails> get meetings => _meetings;
  bool get isLoadingMeetings => _isLoadingMeetings;
  String? get meetingsError => _meetingsError;

  /// Инициализация провайдера
  Future<void> initialize() async {
    await _initializeService();
    _subscribeToNotifications();
  }

  /// Инициализация MeetingsService
  Future<void> _initializeService() async {
    try {
      final baseUrl = await _configService.getApiUrl();
      final apiClient = ApiClient(baseUrl: baseUrl, navigatorKey: navigatorKey);
      _meetingsService = MeetingsService(apiClient);
      Logger.logInfo('MeetingsProvider initialized');
    } catch (e) {
      Logger.logError('Failed to initialize MeetingsProvider', error: e);
    }
  }

  /// Подписка на WebSocket уведомления
  void _subscribeToNotifications() {
    _notificationSubscription?.cancel();
    _notificationSubscription = notificationWebSocketService.notifications.listen(
      _handleNotification,
      onError: (e) => Logger.logError('Notification stream error', error: e),
    );
    Logger.logInfo('📡 MeetingsProvider subscribed to notifications');
  }

  /// Обработка уведомлений
  void _handleNotification(NotificationMessage notification) {
    Logger.logInfo('📩 MeetingsProvider received: ${notification.type}', data: {
      'entityId': notification.entityId,
      'meetingId': notification.meetingId,
    });

    final meetingId = notification.meetingId;
    final entityId = notification.entityId;

    switch (notification.type) {
      // Meeting events
      case NotificationEventType.meetingStatusChanged:
      case NotificationEventType.meetingUpdated:
        if (meetingId != null) {
          _refreshMeeting(meetingId);
        }
        break;

      case NotificationEventType.meetingParticipantJoin:
      case NotificationEventType.meetingParticipantLeave:
        if (meetingId != null) {
          _refreshMeeting(meetingId);
        }
        break;

      // Recording events
      case NotificationEventType.recordingStarted:
      case NotificationEventType.recordingCompleted:
      case NotificationEventType.recordingFailed:
      case NotificationEventType.recordingProcessing:
        if (meetingId != null) {
          _refreshRecordings(meetingId);
        }
        break;

      // Transcription events
      case NotificationEventType.transcriptionStarted:
      case NotificationEventType.transcriptionCompleted:
      case NotificationEventType.transcriptionFailed:
      case NotificationEventType.transcriptionProgress:
        // entityId для транскрипции - это track_sid или room_sid
        if (entityId != null) {
          _refreshTranscripts(entityId);
        }
        if (meetingId != null) {
          _refreshRecordings(meetingId);
        }
        break;

      // Summary events
      case NotificationEventType.summaryStarted:
      case NotificationEventType.summaryCompleted:
      case NotificationEventType.summaryFailed:
      case NotificationEventType.summaryProgress:
        // entityId для summary - это room_sid
        if (entityId != null) {
          _refreshTranscripts(entityId);
        }
        // Также обновляем meeting (там может быть summary_status)
        if (meetingId != null) {
          _refreshMeeting(meetingId);
        }
        break;

      // Composite video events
      case NotificationEventType.compositeVideoStarted:
      case NotificationEventType.compositeVideoCompleted:
      case NotificationEventType.compositeVideoFailed:
        if (meetingId != null) {
          _refreshRecordings(meetingId);
        }
        break;
    }
  }

  /// Загрузка списка встреч
  Future<void> loadMeetings({
    String? status,
    String? type,
    bool forceRefresh = false,
  }) async {
    if (_meetingsService == null) {
      await _initializeService();
    }

    if (_meetingsService == null) {
      _meetingsError = 'Service not initialized';
      notifyListeners();
      return;
    }

    _isLoadingMeetings = true;
    _meetingsError = null;
    notifyListeners();

    try {
      _meetings = await _meetingsService!.getMeetings(
        status: status,
        pageSize: 100,
      );

      // Обновляем кэш
      for (final meeting in _meetings) {
        _meetingsCache[meeting.id] = meeting;
      }

      Logger.logSuccess('Loaded ${_meetings.length} meetings');
    } catch (e) {
      _meetingsError = e.toString();
      Logger.logError('Failed to load meetings', error: e);
    } finally {
      _isLoadingMeetings = false;
      notifyListeners();
    }
  }

  /// Получить встречу по ID (из кэша или загрузить)
  MeetingWithDetails? getMeetingById(String meetingId) {
    return _meetingsCache[meetingId];
  }

  /// Загрузить встречу по ID
  Future<MeetingWithDetails?> loadMeeting(String meetingId) async {
    if (_meetingsService == null) {
      await _initializeService();
    }

    if (_meetingsService == null) return null;

    try {
      final meeting = await _meetingsService!.getMeeting(meetingId);
      _meetingsCache[meetingId] = meeting;

      // Обновляем в списке если есть
      final index = _meetings.indexWhere((m) => m.id == meetingId);
      if (index >= 0) {
        _meetings[index] = meeting;
      }

      notifyListeners();
      return meeting;
    } catch (e) {
      Logger.logError('Failed to load meeting $meetingId', error: e);
      return null;
    }
  }

  /// Обновить встречу из сети
  Future<void> _refreshMeeting(String meetingId) async {
    Logger.logInfo('🔄 Refreshing meeting: $meetingId');
    await loadMeeting(meetingId);
  }

  /// Получить записи встречи (из кэша или загрузить)
  List<RoomRecording>? getRecordings(String meetingId) {
    return _recordingsCache[meetingId];
  }

  /// Загрузить записи встречи
  Future<List<RoomRecording>?> loadRecordings(String meetingId) async {
    if (_meetingsService == null) {
      await _initializeService();
    }

    if (_meetingsService == null) return null;

    try {
      final recordings = await _meetingsService!.getMeetingRecordings(meetingId);
      _recordingsCache[meetingId] = recordings;
      notifyListeners();
      return recordings;
    } catch (e) {
      Logger.logError('Failed to load recordings for $meetingId', error: e);
      return null;
    }
  }

  /// Обновить записи из сети
  Future<void> _refreshRecordings(String meetingId) async {
    Logger.logInfo('🔄 Refreshing recordings for meeting: $meetingId');
    await loadRecordings(meetingId);
  }

  /// Получить транскрипты комнаты (из кэша или загрузить)
  RoomTranscripts? getTranscripts(String roomSid) {
    return _transcriptsCache[roomSid];
  }

  /// Загрузить транскрипты комнаты
  Future<RoomTranscripts?> loadTranscripts(String roomSid) async {
    if (_meetingsService == null) {
      await _initializeService();
    }

    if (_meetingsService == null) return null;

    try {
      final transcripts = await _meetingsService!.getRoomTranscripts(roomSid);
      _transcriptsCache[roomSid] = transcripts;
      notifyListeners();
      return transcripts;
    } catch (e) {
      Logger.logError('Failed to load transcripts for $roomSid', error: e);
      return null;
    }
  }

  /// Обновить транскрипты из сети
  Future<void> _refreshTranscripts(String roomSid) async {
    Logger.logInfo('🔄 Refreshing transcripts for room: $roomSid');
    await loadTranscripts(roomSid);
  }

  /// Удалить встречу из кэша (после удаления)
  void removeMeeting(String meetingId) {
    _meetingsCache.remove(meetingId);
    _recordingsCache.remove(meetingId);
    _meetings.removeWhere((m) => m.id == meetingId);
    notifyListeners();
  }

  /// Удалить запись из кэша
  void removeRecording(String meetingId, String roomSid) {
    final recordings = _recordingsCache[meetingId];
    if (recordings != null) {
      recordings.removeWhere((r) => r.roomSid == roomSid);
      _transcriptsCache.remove(roomSid);
      notifyListeners();
    }
  }

  /// Очистить весь кэш
  void clearCache() {
    _meetingsCache.clear();
    _recordingsCache.clear();
    _transcriptsCache.clear();
    _meetings.clear();
    notifyListeners();
  }

  /// Подключиться к WebSocket (вызывается после логина)
  Future<void> connectWebSocket() async {
    await notificationWebSocketService.connect();
  }

  /// Отключиться от WebSocket (вызывается при логауте)
  void disconnectWebSocket() {
    notificationWebSocketService.disconnect();
    clearCache();
  }

  @override
  void dispose() {
    _notificationSubscription?.cancel();
    super.dispose();
  }
}

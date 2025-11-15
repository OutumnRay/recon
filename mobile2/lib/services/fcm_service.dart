import 'dart:io';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import '../utils/logger.dart';
import 'api_client.dart';

/// FCM Service for handling push notifications
class FCMService {
  final ApiClient _apiClient;
  final FirebaseMessaging _firebaseMessaging = FirebaseMessaging.instance;
  final FlutterLocalNotificationsPlugin _localNotifications =
      FlutterLocalNotificationsPlugin();

  String? _currentToken;
  bool _isInitialized = false;

  FCMService(this._apiClient);

  /// Initialize FCM and request permissions
  Future<void> initialize() async {
    if (_isInitialized) {
      Logger.logWarning('FCM Service already initialized');
      return;
    }

    try {
      Logger.logInfo('Initializing FCM Service');

      // Request notification permissions
      final settings = await _firebaseMessaging.requestPermission(
        alert: true,
        announcement: false,
        badge: true,
        carPlay: false,
        criticalAlert: false,
        provisional: false,
        sound: true,
      );

      if (settings.authorizationStatus == AuthorizationStatus.authorized) {
        Logger.logSuccess('User granted notification permissions');
      } else if (settings.authorizationStatus ==
          AuthorizationStatus.provisional) {
        Logger.logInfo('User granted provisional notification permissions');
      } else {
        Logger.logWarning('User declined notification permissions');
        return;
      }

      // Initialize local notifications
      await _initializeLocalNotifications();

      // Get FCM token
      _currentToken = await _firebaseMessaging.getToken();
      if (_currentToken != null) {
        Logger.logSuccess('FCM Token obtained: ${_currentToken!.substring(0, 20)}...');
      } else {
        Logger.logWarning('Failed to obtain FCM token');
      }

      // Listen for token refresh
      _firebaseMessaging.onTokenRefresh.listen((newToken) {
        Logger.logInfo('FCM Token refreshed');
        _currentToken = newToken;
        // Register the new token
        registerDevice();
      });

      // Configure message handlers
      _configureForegroundMessageHandler();
      _configureBackgroundMessageHandler();

      _isInitialized = true;
      Logger.logSuccess('FCM Service initialized successfully');
    } catch (e) {
      Logger.logError('Failed to initialize FCM Service', error: e);
    }
  }

  /// Initialize local notifications plugin
  Future<void> _initializeLocalNotifications() async {
    const androidSettings = AndroidInitializationSettings('@mipmap/ic_launcher');
    const iosSettings = DarwinInitializationSettings(
      requestAlertPermission: true,
      requestBadgePermission: true,
      requestSoundPermission: true,
    );

    const initSettings = InitializationSettings(
      android: androidSettings,
      iOS: iosSettings,
    );

    await _localNotifications.initialize(
      initSettings,
      onDidReceiveNotificationResponse: _onNotificationTapped,
    );
  }

  /// Configure foreground message handler
  void _configureForegroundMessageHandler() {
    FirebaseMessaging.onMessage.listen((RemoteMessage message) {
      Logger.logInfo('Foreground message received', data: {
        'title': message.notification?.title,
        'body': message.notification?.body,
      });

      // Show local notification when app is in foreground
      _showLocalNotification(message);
    });
  }

  /// Configure background message handler
  void _configureBackgroundMessageHandler() {
    FirebaseMessaging.onMessageOpenedApp.listen((RemoteMessage message) {
      Logger.logInfo('Notification opened app', data: {
        'title': message.notification?.title,
        'body': message.notification?.body,
      });

      // Handle notification tap
      _handleNotificationTap(message.data);
    });
  }

  /// Show local notification
  Future<void> _showLocalNotification(RemoteMessage message) async {
    const androidDetails = AndroidNotificationDetails(
      'default_channel',
      'Default Notifications',
      channelDescription: 'Default notification channel for Recontext',
      importance: Importance.high,
      priority: Priority.high,
      showWhen: true,
    );

    const iosDetails = DarwinNotificationDetails(
      presentAlert: true,
      presentBadge: true,
      presentSound: true,
    );

    const notificationDetails = NotificationDetails(
      android: androidDetails,
      iOS: iosDetails,
    );

    await _localNotifications.show(
      message.hashCode,
      message.notification?.title ?? 'Recontext',
      message.notification?.body ?? '',
      notificationDetails,
      payload: message.data.toString(),
    );
  }

  /// Handle notification tap
  void _onNotificationTapped(NotificationResponse response) {
    Logger.logInfo('Notification tapped', data: {'payload': response.payload});
    // TODO: Navigate to appropriate screen based on payload
  }

  /// Handle notification tap from background
  void _handleNotificationTap(Map<String, dynamic> data) {
    Logger.logInfo('Handling notification tap', data: data);
    // TODO: Navigate to appropriate screen based on data
  }

  /// Register device with backend
  Future<bool> registerDevice() async {
    if (_currentToken == null) {
      Logger.logWarning('Cannot register device: FCM token is null');
      return false;
    }

    try {
      final platform = Platform.isIOS ? 'ios' : Platform.isAndroid ? 'android' : 'web';

      final response = await _apiClient.post(
        '/api/v1/fcm/register',
        {
          'fcm_token': _currentToken,
          'platform': platform,
          'device_model': await _getDeviceModel(),
          'app_version': '1.0.0', // TODO: Get from package info
          'os_version': Platform.operatingSystemVersion,
        },
      );

      if (response.statusCode == 200) {
        Logger.logSuccess('Device registered for push notifications');
        return true;
      } else {
        Logger.logError('Failed to register device', error: 'Status: ${response.statusCode}');
        return false;
      }
    } catch (e) {
      Logger.logError('Error registering device', error: e);
      return false;
    }
  }

  /// Unregister device from backend
  Future<bool> unregisterDevice() async {
    if (_currentToken == null) {
      Logger.logWarning('Cannot unregister device: FCM token is null');
      return false;
    }

    try {
      final response = await _apiClient.post(
        '/api/v1/fcm/unregister',
        {
          'fcm_token': _currentToken,
        },
      );

      if (response.statusCode == 200) {
        Logger.logSuccess('Device unregistered from push notifications');
        return true;
      } else {
        Logger.logError('Failed to unregister device', error: 'Status: ${response.statusCode}');
        return false;
      }
    } catch (e) {
      Logger.logError('Error unregistering device', error: e);
      return false;
    }
  }

  /// Get device model information
  Future<String> _getDeviceModel() async {
    // TODO: Use device_info_plus package for more detailed info
    if (Platform.isIOS) {
      return 'iOS Device';
    } else if (Platform.isAndroid) {
      return 'Android Device';
    }
    return 'Unknown';
  }

  /// Get current FCM token
  String? get token => _currentToken;

  /// Check if FCM is initialized
  bool get isInitialized => _isInitialized;
}

/// Background message handler (must be top-level function)
@pragma('vm:entry-point')
Future<void> firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  Logger.logInfo('Background message received', data: {
    'title': message.notification?.title,
    'body': message.notification?.body,
  });
}

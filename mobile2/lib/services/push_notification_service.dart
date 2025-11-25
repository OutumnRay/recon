import 'dart:convert';
import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'package:flutter/material.dart';
import '../utils/logger.dart';

/// Top-level function for handling background messages
@pragma('vm:entry-point')
Future<void> _firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  await Firebase.initializeApp();
  Logger.logInfo('Background message received', data: {
    'messageId': message.messageId,
    'notification': message.notification?.title,
  });
}

/// Push Notification Service using Firebase Cloud Messaging
/// Handles foreground/background notifications and deeplinks
class PushNotificationService {
  static final PushNotificationService _instance = PushNotificationService._internal();
  factory PushNotificationService() => _instance;
  PushNotificationService._internal();

  final FirebaseMessaging _messaging = FirebaseMessaging.instance;
  final FlutterLocalNotificationsPlugin _localNotifications =
      FlutterLocalNotificationsPlugin();

  String? _fcmToken;
  bool _initialized = false;

  // Callback for handling notification taps (deeplinks)
  Function(Map<String, dynamic>)? onNotificationTap;

  // Callback for handling foreground notifications (show in-app)
  Function(RemoteMessage)? onForegroundMessage;

  /// Get FCM token for this device
  String? get fcmToken => _fcmToken;

  /// Initialize push notification service
  Future<void> initialize() async {
    if (_initialized) {
      Logger.logWarning('PushNotificationService already initialized');
      return;
    }

    try {
      // Request permission
      NotificationSettings settings = await _messaging.requestPermission(
        alert: true,
        badge: true,
        sound: true,
        provisional: false,
      );

      if (settings.authorizationStatus == AuthorizationStatus.authorized) {
        Logger.logInfo('Push notifications authorized');
      } else if (settings.authorizationStatus == AuthorizationStatus.provisional) {
        Logger.logInfo('Push notifications provisional');
      } else {
        Logger.logWarning('Push notifications denied');
        return;
      }

      // Initialize local notifications (for foreground)
      await _initializeLocalNotifications();

      // Get FCM token
      _fcmToken = await _messaging.getToken();
      Logger.logInfo('FCM Token obtained', data: {'token': _fcmToken});

      // Listen to token refresh
      _messaging.onTokenRefresh.listen((newToken) {
        _fcmToken = newToken;
        Logger.logInfo('FCM Token refreshed', data: {'token': newToken});
        // TODO: Send updated token to backend
      });

      // Handle foreground messages
      FirebaseMessaging.onMessage.listen(_handleForegroundMessage);

      // Handle notification tap when app is in background
      FirebaseMessaging.onMessageOpenedApp.listen(_handleNotificationTap);

      // Check if app was opened from terminated state via notification
      RemoteMessage? initialMessage = await _messaging.getInitialMessage();
      if (initialMessage != null) {
        _handleNotificationTap(initialMessage);
      }

      // Set background message handler
      FirebaseMessaging.onBackgroundMessage(_firebaseMessagingBackgroundHandler);

      _initialized = true;
      Logger.logInfo('PushNotificationService initialized successfully');
    } catch (e) {
      Logger.logError('Failed to initialize PushNotificationService', error: e);
    }
  }

  /// Initialize flutter_local_notifications for foreground notifications
  Future<void> _initializeLocalNotifications() async {
    const AndroidInitializationSettings androidSettings =
        AndroidInitializationSettings('@mipmap/ic_launcher');

    const DarwinInitializationSettings iosSettings =
        DarwinInitializationSettings(
      requestAlertPermission: true,
      requestBadgePermission: true,
      requestSoundPermission: true,
    );

    const InitializationSettings settings = InitializationSettings(
      android: androidSettings,
      iOS: iosSettings,
    );

    await _localNotifications.initialize(
      settings,
      onDidReceiveNotificationResponse: (NotificationResponse response) {
        // Handle notification tap
        if (response.payload != null) {
          try {
            final Map<String, dynamic> data = jsonDecode(response.payload!);
            onNotificationTap?.call(data);
          } catch (e) {
            Logger.logError('Failed to parse notification payload', error: e);
          }
        }
      },
    );
  }

  /// Handle foreground message (app is open)
  void _handleForegroundMessage(RemoteMessage message) {
    Logger.logInfo('Foreground message received', data: {
      'messageId': message.messageId,
      'notification': message.notification?.title,
      'data': message.data,
    });

    // Call custom handler if provided
    if (onForegroundMessage != null) {
      onForegroundMessage!(message);
      return;
    }

    // Show local notification
    _showLocalNotification(message);
  }

  /// Show local notification for foreground messages
  Future<void> _showLocalNotification(RemoteMessage message) async {
    final notification = message.notification;

    if (notification == null) return;

    // Create notification details
    AndroidNotificationDetails androidDetails = AndroidNotificationDetails(
      'recontext_channel',
      'Recontext Notifications',
      channelDescription: 'Notifications for meetings, transcriptions, and summaries',
      importance: Importance.high,
      priority: Priority.high,
      icon: '@mipmap/ic_launcher',
    );

    DarwinNotificationDetails iosDetails = const DarwinNotificationDetails(
      presentAlert: true,
      presentBadge: true,
      presentSound: true,
    );

    NotificationDetails details = NotificationDetails(
      android: androidDetails,
      iOS: iosDetails,
    );

    // Show notification with payload
    await _localNotifications.show(
      message.hashCode,
      notification.title,
      notification.body,
      details,
      payload: jsonEncode(message.data),
    );
  }

  /// Handle notification tap (deeplink navigation)
  void _handleNotificationTap(RemoteMessage message) {
    Logger.logInfo('Notification tapped', data: {
      'messageId': message.messageId,
      'data': message.data,
    });

    if (onNotificationTap != null) {
      onNotificationTap!(message.data);
    }
  }

  /// Subscribe to a topic
  Future<void> subscribeToTopic(String topic) async {
    try {
      await _messaging.subscribeToTopic(topic);
      Logger.logInfo('Subscribed to topic', data: {'topic': topic});
    } catch (e) {
      Logger.logError('Failed to subscribe to topic', error: e);
    }
  }

  /// Unsubscribe from a topic
  Future<void> unsubscribeFromTopic(String topic) async {
    try {
      await _messaging.unsubscribeFromTopic(topic);
      Logger.logInfo('Unsubscribed from topic', data: {'topic': topic});
    } catch (e) {
      Logger.logError('Failed to unsubscribe from topic', error: e);
    }
  }

  /// Show a custom in-app notification (Flutter SnackBar/Banner)
  static void showInAppNotification(
    BuildContext context, {
    required String title,
    required String message,
    VoidCallback? onTap,
    Duration duration = const Duration(seconds: 4),
  }) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              title,
              style: const TextStyle(
                fontWeight: FontWeight.bold,
                fontSize: 16,
              ),
            ),
            const SizedBox(height: 4),
            Text(message),
          ],
        ),
        duration: duration,
        behavior: SnackBarBehavior.floating,
        action: onTap != null
            ? SnackBarAction(
                label: 'VIEW',
                onPressed: onTap,
              )
            : null,
      ),
    );
  }
}

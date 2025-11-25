import 'package:flutter/material.dart';
import '../utils/logger.dart';
import '../screens/meeting_detail_screen.dart';

/// Handles deeplinks from push notifications
class DeeplinkHandler {
  static final GlobalKey<NavigatorState> navigatorKey = GlobalKey<NavigatorState>();

  /// Handle deeplink navigation from notification data
  static Future<void> handleDeeplink(Map<String, dynamic> data) async {
    try {
      Logger.logInfo('Handling deeplink', data: data);

      final String? type = data['type'];
      final String? meetingId = data['meeting_id'];

      if (type == null) {
        Logger.logWarning('Deeplink type is null');
        return;
      }

      // Get navigator context
      final NavigatorState? navigator = navigatorKey.currentState;
      if (navigator == null) {
        Logger.logWarning('Navigator is not available');
        return;
      }

      // Handle different notification types
      // For all recording/transcription/summary events, navigate to meeting details
      // The meeting detail screen will show the updated data
      switch (type) {
        case 'composite_video.completed':
        case 'recording.completed':
        case 'transcription.completed':
        case 'summary.completed':
        case 'meeting.status_changed':
        case 'meeting.updated':
          // Navigate to meeting details
          if (meetingId != null) {
            _navigateToMeeting(navigator, meetingId);
          }
          break;

        default:
          Logger.logWarning('Unknown notification type: $type');
      }
    } catch (e) {
      Logger.logError('Failed to handle deeplink', error: e);
    }
  }

  /// Navigate to meeting details
  static void _navigateToMeeting(
    NavigatorState navigator,
    String meetingId,
  ) {
    navigator.push(
      MaterialPageRoute(
        builder: (context) => MeetingDetailScreen(
          meetingId: meetingId,
        ),
      ),
    );
  }

  /// Build deeplink from notification data
  static String buildDeeplink({
    required String type,
    required String meetingId,
    String? entityId,
    String? roomSid,
  }) {
    // Format: recontext://notification/{type}?meeting_id={meetingId}&entity_id={entityId}
    final StringBuffer url = StringBuffer('recontext://notification/$type?meeting_id=$meetingId');

    if (entityId != null) {
      url.write('&entity_id=$entityId');
    }

    if (roomSid != null) {
      url.write('&room_sid=$roomSid');
    }

    return url.toString();
  }
}

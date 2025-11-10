import 'package:flutter/foundation.dart';

/// Global logging configuration
/// Set to false to disable all debug logs in production
const bool kEnableDebugLogs = true; // <-- Change to false to disable all logs

/// Logger utility for debugging API requests and responses
class Logger {
  static const bool _enabled = kEnableDebugLogs;

  /// Log API request
  static void logRequest(String method, String url, {Map<String, dynamic>? body}) {
    if (!_enabled) return;

    debugPrint('\n🔵 ===== API REQUEST =====');
    debugPrint('📤 Method: $method');
    debugPrint('🔗 URL: $url');
    if (body != null) {
      debugPrint('📦 Body: $body');
    }
    debugPrint('========================\n');
  }

  /// Log API response
  static void logResponse(String url, int statusCode, String body) {
    if (!_enabled) return;

    final icon = statusCode >= 200 && statusCode < 300 ? '✅' : '❌';
    debugPrint('\n$icon ===== API RESPONSE =====');
    debugPrint('🔗 URL: $url');
    debugPrint('📊 Status: $statusCode');
    debugPrint('📦 Body: $body');
    debugPrint('==========================\n');
  }

  /// Log error
  static void logError(String message, {Object? error, StackTrace? stackTrace}) {
    if (!_enabled) return;

    debugPrint('\n❌ ===== ERROR =====');
    debugPrint('💥 Message: $message');
    if (error != null) {
      debugPrint('🐛 Error: $error');
    }
    if (stackTrace != null) {
      debugPrint('📍 Stack trace:\n$stackTrace');
    }
    debugPrint('===================\n');
  }

  /// Log info message
  static void logInfo(String message, {Map<String, dynamic>? data}) {
    if (!_enabled) return;

    debugPrint('\nℹ️  $message');
    if (data != null) {
      debugPrint('   Data: $data');
    }
  }

  /// Log warning
  static void logWarning(String message) {
    if (!_enabled) return;

    debugPrint('\n⚠️  WARNING: $message\n');
  }

  /// Log success
  static void logSuccess(String message) {
    if (!_enabled) return;

    debugPrint('\n✅ SUCCESS: $message\n');
  }
}

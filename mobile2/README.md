# Recontext Mobile App

Flutter mobile application for Recontext.online platform - video conferences and meeting management.

## Features

- **User Authentication** - Login with username/password
- **Meetings Management** - View, create, and manage meetings
- **LiveKit Integration** - Join video conferences
- **Offline Support** - Token storage for persistent login

## Project Structure

```
lib/
├── models/           # Data models (User, Meeting, etc.)
├── services/         # API services
│   ├── api_client.dart       # HTTP client wrapper
│   ├── auth_service.dart     # Authentication
│   ├── meetings_service.dart # Meetings API
│   └── storage_service.dart  # Local storage
├── screens/          # UI screens (to be implemented)
├── widgets/          # Reusable widgets (to be implemented)
└── main.dart         # App entry point
```

## Setup

### Prerequisites

- Flutter SDK 3.9.2 or higher
- Dart SDK
- iOS: Xcode 14+
- Android: Android Studio with API 21+

### Installation

1. Install dependencies:
```bash
flutter pub get
```

2. Configure API endpoint in `lib/main.dart`:
```dart
final _apiClient = ApiClient(baseUrl: 'https://your-api-url.com');
```

### Run

```bash
# Development
flutter run

# Release build
flutter build apk  # Android
flutter build ios  # iOS
```

## API Services

### AuthService

```dart
final authService = AuthService(apiClient);

// Login
final user = await authService.login('username', 'password');

// Check if logged in
final isLoggedIn = await authService.isLoggedIn();

// Logout
await authService.logout();
```

### MeetingsService

```dart
final meetingsService = MeetingsService(apiClient);

// Get meetings
final meetings = await meetingsService.getMeetings(
  status: 'scheduled',
  limit: 20,
);

// Get single meeting
final meeting = await meetingsService.getMeeting('meeting-id');

// Create meeting
final newMeeting = await meetingsService.createMeeting(
  CreateMeetingRequest(
    title: 'Team Meeting',
    scheduledAt: DateTime.now(),
    duration: 30,
    type: 'conference',
    subjectId: 'subject-id',
    needsVideoRecord: true,
    needsAudioRecord: true,
    participantIds: [],
    departmentIds: [],
  ),
);

// Get LiveKit token
final token = await meetingsService.getLiveKitToken('meeting-id');
```

## Configuration

### Environment Variables

You can use environment variables for different configurations:

```bash
flutter run --dart-define=API_BASE_URL=https://api.recontext.online
```

### Android

Update `android/app/src/main/AndroidManifest.xml` for permissions:

```xml
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.CAMERA" />
<uses-permission android:name="android.permission.RECORD_AUDIO" />
```

### iOS

Update `ios/Runner/Info.plist` for permissions:

```xml
<key>NSCameraUsageDescription</key>
<string>Camera access is required for video calls</string>
<key>NSMicrophoneUsageDescription</key>
<string>Microphone access is required for video calls</string>
```

## Documentation

See [API_DOCUMENTATION.md](API_DOCUMENTATION.md) for complete API documentation and examples.

## Dependencies

- **http** ^1.2.0 - HTTP client
- **shared_preferences** ^2.2.2 - Local storage
- **provider** ^6.1.1 - State management
- **livekit_client** ^2.2.3 - Video conferencing
- **intl** ^0.19.0 - Internationalization

## Development

### Code Style

Follow [Effective Dart](https://dart.dev/guides/language/effective-dart) guidelines.

Run linter:
```bash
flutter analyze
```

Format code:
```bash
flutter format lib/
```

### Testing

```bash
flutter test
```

## Build for Production

### Android

```bash
flutter build apk --release
# or
flutter build appbundle --release
```

Output: `build/app/outputs/flutter-apk/app-release.apk`

### iOS

```bash
flutter build ios --release
```

Then open in Xcode and archive for App Store.

## Default Credentials

For development and testing, the app comes with pre-filled credentials:

- **Username**: `admin@recontext.online`
- **Password**: `admin123`

You can change these in `lib/screens/login_screen.dart` if needed.

## Error Handling & Debugging

The app includes comprehensive error handling with detailed error messages displayed directly in the UI:

### Features:
- **Detailed Error Messages**: All errors include stack traces and debugging information
- **Copy to Clipboard**: Click the copy button on any error to copy full details
- **Selectable Text**: Error messages are selectable for easy copying
- **Retry Functionality**: Most error screens include a retry button
- **Network Diagnostics**: Error messages include helpful troubleshooting tips

### Error Display Widget

The app uses a custom `ErrorDisplay` widget for consistent error presentation:

```dart
// Compact error display
ErrorDisplay(
  error: errorMessage,
  onRetry: retryFunction,
  compact: true,
)

// Full error display with all details
ErrorDisplay(
  error: errorMessage,
  onRetry: retryFunction,
)

// Full-screen error for critical issues
FullScreenError(
  error: errorMessage,
  onRetry: retryFunction,
  title: 'Custom Title',
)
```

All errors are caught and displayed with:
- API errors: Show status codes and response messages
- Connection errors: Include network troubleshooting tips
- Unexpected errors: Display stack traces for debugging

## Troubleshooting

### FlutterDartVMServicePublisher Error

If you see the error "Could not register as server for FlutterDartVMServicePublisher", this is a non-critical warning related to Flutter DevTools network configuration. The app will continue to function normally. To resolve:

1. Check your network settings
2. Restart the application
3. If persistent, try restarting your IDE/editor

### API Connection Issues

1. Check base URL is correct
2. Ensure device can reach the API (use localhost:8081 for local development)
3. For Android emulator, use `10.0.2.2` instead of `localhost`
4. For iOS simulator, use `localhost` or machine's IP

### Token Expiration

The app automatically handles 401 errors and redirects to login when token expires.

### LiveKit Connection

Ensure LiveKit URL is accessible and API credentials are correct in backend configuration.

## License

Copyright © 2025 Recontext. All rights reserved.

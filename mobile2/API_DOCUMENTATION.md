# Recontext Mobile API Documentation

## Overview

This Flutter application provides API integration with Recontext.online User Portal backend. The API is organized into several services:

- **AuthService** - User authentication
- **MeetingsService** - Meeting management and LiveKit integration
- **StorageService** - Local data persistence

## Setup

### 1. Install Dependencies

```bash
flutter pub get
```

### 2. Configure API Base URL

Update the base URL in your app initialization:

```dart
final apiClient = ApiClient(baseUrl: 'https://your-domain.com');
```

## API Services

### AuthService

Handles user authentication and session management.

#### Login

```dart
import 'package:recontext/services/auth_service.dart';
import 'package:recontext/services/api_client.dart';

// Initialize
final apiClient = ApiClient(baseUrl: 'https://your-domain.com');
final authService = AuthService(apiClient);

// Login
try {
  final user = await authService.login('username', 'password');
  print('Logged in as: ${user.username}');
  print('User role: ${user.role}');
} on ApiException catch (e) {
  print('Login failed: ${e.message}');
}
```

#### Check Login Status

```dart
final isLoggedIn = await authService.isLoggedIn();
if (isLoggedIn) {
  final userData = await authService.getCurrentUserData();
  print('Current user: ${userData['username']}');
}
```

#### Logout

```dart
await authService.logout();
```

### MeetingsService

Manages meetings, participants, and LiveKit integration.

#### Get Meetings List

```dart
import 'package:recontext/services/meetings_service.dart';

final meetingsService = MeetingsService(apiClient);

// Get all meetings
final meetings = await meetingsService.getMeetings();

// Get meetings with filters
final upcomingMeetings = await meetingsService.getMeetings(
  status: 'scheduled',
  startDate: DateTime.now(),
  limit: 10,
);

for (var meeting in upcomingMeetings) {
  print('${meeting.title} - ${meeting.scheduledAt}');
  print('Participants: ${meeting.participants.length}');
}
```

#### Get Single Meeting

```dart
final meeting = await meetingsService.getMeeting('meeting-id');
print('Meeting: ${meeting.title}');
print('Status: ${meeting.status}');
print('LiveKit Room: ${meeting.liveKitRoomId}');
```

#### Create Meeting

```dart
import 'package:recontext/models/meeting.dart';

final request = CreateMeetingRequest(
  title: 'Team Standup',
  scheduledAt: DateTime.now().add(Duration(hours: 1)),
  duration: 30, // minutes
  type: 'conference',
  subjectId: 'subject-123',
  needsVideoRecord: true,
  needsAudioRecord: true,
  participantIds: ['user-1', 'user-2'],
  departmentIds: ['dept-1'],
);

try {
  final meeting = await meetingsService.createMeeting(request);
  print('Meeting created: ${meeting.id}');
} on ApiException catch (e) {
  print('Failed to create meeting: ${e.message}');
}
```

#### Update Meeting

```dart
final updatedMeeting = await meetingsService.updateMeeting(
  'meeting-id',
  {
    'title': 'Updated Meeting Title',
    'status': 'in_progress',
  },
);
```

#### Delete Meeting

```dart
await meetingsService.deleteMeeting('meeting-id');
```

#### Get LiveKit Token for Video Conference

```dart
try {
  final liveKitToken = await meetingsService.getLiveKitToken('meeting-id');

  // Use token to connect to LiveKit room
  print('Token: ${liveKitToken.token}');
  print('URL: ${liveKitToken.url}');

  // Connect to LiveKit (requires livekit_client package)
  // See LiveKit Flutter SDK documentation for details
} on ApiException catch (e) {
  if (e.statusCode == 403) {
    print('You are not invited to this meeting');
  }
}
```

#### Get Meeting Subjects

```dart
final subjects = await meetingsService.getMeetingSubjects();
for (var subject in subjects) {
  print('${subject.name}: ${subject.description}');
}
```

## Models

### User

```dart
class User {
  final String id;
  final String username;
  final String email;
  final String role; // 'admin', 'user', 'operator'
  final String? departmentId;
  final UserPermissions permissions;
  final String language; // 'en', 'ru'
}
```

### Meeting

```dart
class Meeting {
  final String id;
  final String title;
  final DateTime scheduledAt;
  final int duration; // in minutes
  final String? recurrence; // 'daily', 'weekly', 'monthly'
  final String type; // 'conference', 'presentation', 'training'
  final String status; // 'scheduled', 'in_progress', 'completed', 'cancelled'
  final bool needsVideoRecord;
  final bool needsAudioRecord;
  final String? liveKitRoomId;
}
```

### MeetingWithDetails

Extends `Meeting` with additional information:

```dart
class MeetingWithDetails extends Meeting {
  final String? subjectName;
  final List<MeetingParticipant> participants;
  final List<String> departments;
}
```

## Error Handling

All API methods can throw `ApiException`:

```dart
try {
  final meetings = await meetingsService.getMeetings();
} on ApiException catch (e) {
  print('API Error: ${e.statusCode} - ${e.message}');

  switch (e.statusCode) {
    case 401:
      // Unauthorized - token expired, need to login again
      await authService.logout();
      // Navigate to login screen
      break;
    case 403:
      // Forbidden - no permission
      print('You don\'t have permission for this action');
      break;
    case 404:
      // Not found
      print('Resource not found');
      break;
    case 500:
      // Server error
      print('Server error, please try again later');
      break;
  }
} catch (e) {
  print('Unexpected error: $e');
}
```

## Complete Example

```dart
import 'package:flutter/material.dart';
import 'package:recontext/services/api_client.dart';
import 'package:recontext/services/auth_service.dart';
import 'package:recontext/services/meetings_service.dart';

void main() {
  runApp(MyApp());
}

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Recontext',
      home: LoginScreen(),
    );
  }
}

class LoginScreen extends StatefulWidget {
  @override
  _LoginScreenState createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _usernameController = TextEditingController();
  final _passwordController = TextEditingController();
  final _apiClient = ApiClient(baseUrl: 'https://recontext.online');
  late final AuthService _authService;
  bool _isLoading = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _authService = AuthService(_apiClient);
    _checkLoginStatus();
  }

  Future<void> _checkLoginStatus() async {
    final isLoggedIn = await _authService.isLoggedIn();
    if (isLoggedIn) {
      _navigateToHome();
    }
  }

  Future<void> _login() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      await _authService.login(
        _usernameController.text,
        _passwordController.text,
      );
      _navigateToHome();
    } on ApiException catch (e) {
      setState(() {
        _error = e.message;
      });
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }

  void _navigateToHome() {
    Navigator.of(context).pushReplacement(
      MaterialPageRoute(builder: (_) => HomeScreen(apiClient: _apiClient)),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Recontext Login')),
      body: Padding(
        padding: EdgeInsets.all(16.0),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            TextField(
              controller: _usernameController,
              decoration: InputDecoration(labelText: 'Username'),
            ),
            SizedBox(height: 16),
            TextField(
              controller: _passwordController,
              decoration: InputDecoration(labelText: 'Password'),
              obscureText: true,
            ),
            SizedBox(height: 24),
            if (_error != null)
              Text(_error!, style: TextStyle(color: Colors.red)),
            SizedBox(height: 16),
            ElevatedButton(
              onPressed: _isLoading ? null : _login,
              child: _isLoading
                  ? CircularProgressIndicator()
                  : Text('Login'),
            ),
          ],
        ),
      ),
    );
  }
}

class HomeScreen extends StatefulWidget {
  final ApiClient apiClient;

  HomeScreen({required this.apiClient});

  @override
  _HomeScreenState createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  late final MeetingsService _meetingsService;
  List<MeetingWithDetails>? _meetings;
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _meetingsService = MeetingsService(widget.apiClient);
    _loadMeetings();
  }

  Future<void> _loadMeetings() async {
    try {
      final meetings = await _meetingsService.getMeetings(
        status: 'scheduled',
        limit: 10,
      );
      setState(() {
        _meetings = meetings;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _isLoading = false;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Failed to load meetings: $e')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('My Meetings')),
      body: _isLoading
          ? Center(child: CircularProgressIndicator())
          : _meetings == null || _meetings!.isEmpty
              ? Center(child: Text('No meetings scheduled'))
              : ListView.builder(
                  itemCount: _meetings!.length,
                  itemBuilder: (context, index) {
                    final meeting = _meetings![index];
                    return ListTile(
                      title: Text(meeting.title),
                      subtitle: Text(
                        '${meeting.scheduledAt.toString()}\n'
                        'Participants: ${meeting.participants.length}',
                      ),
                      trailing: Icon(Icons.chevron_right),
                      onTap: () {
                        // Navigate to meeting details
                      },
                    );
                  },
                ),
      floatingActionButton: FloatingActionButton(
        onPressed: () {
          // Navigate to create meeting screen
        },
        child: Icon(Icons.add),
      ),
    );
  }
}
```

## Environment Configuration

For production use, create a configuration file:

```dart
// lib/config/env_config.dart
class EnvConfig {
  static const String apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://localhost:8081',
  );

  static const String liveKitUrl = String.fromEnvironment(
    'LIVEKIT_URL',
    defaultValue: 'wss://video.recontext.online',
  );
}
```

Run with different environments:

```bash
# Development
flutter run --dart-define=API_BASE_URL=http://localhost:8081

# Production
flutter run --dart-define=API_BASE_URL=https://recontext.online --dart-define=LIVEKIT_URL=wss://video.recontext.online
```

## Testing

Example test for AuthService:

```dart
import 'package:flutter_test/flutter_test.dart';
import 'package:recontext/services/auth_service.dart';
import 'package:recontext/services/api_client.dart';

void main() {
  group('AuthService', () {
    late AuthService authService;

    setUp(() {
      final apiClient = ApiClient(baseUrl: 'http://localhost:8081');
      authService = AuthService(apiClient);
    });

    test('login with valid credentials', () async {
      final user = await authService.login('testuser', 'testpassword');
      expect(user.username, 'testuser');
    });

    test('logout clears stored data', () async {
      await authService.logout();
      final isLoggedIn = await authService.isLoggedIn();
      expect(isLoggedIn, false);
    });
  });
}
```

## Additional Resources

- [User Portal API Documentation](../cmd/user-portal/README.md)
- [LiveKit Flutter SDK](https://docs.livekit.io/client-sdks/flutter/)
- [Flutter HTTP Package](https://pub.dev/packages/http)
- [Shared Preferences](https://pub.dev/packages/shared_preferences)

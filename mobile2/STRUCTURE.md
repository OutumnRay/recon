# Project Structure

## Directory Organization

```
mobile2/
├── lib/                          # Main application code
│   ├── main.dart                 # App entry point
│   ├── models/                   # Data models
│   │   ├── user.dart            # User model
│   │   └── meeting.dart         # Meeting models (Meeting, MeetingWithDetails, etc.)
│   ├── services/                 # Business logic and API services
│   │   ├── api_client.dart      # HTTP client wrapper with error handling
│   │   ├── auth_service.dart    # Authentication service
│   │   ├── meetings_service.dart # Meetings API service
│   │   ├── storage_service.dart  # Local storage (SharedPreferences)
│   │   └── config_service.dart   # App configuration (API URL, etc.)
│   ├── screens/                  # UI screens
│   │   ├── login_screen.dart    # Login screen with API config
│   │   ├── main_screen.dart     # Main screen with bottom navigation
│   │   ├── meetings_screen.dart # Meetings list and management
│   │   ├── documents_screen.dart # Documents and transcripts
│   │   ├── search_screen.dart   # Search functionality
│   │   └── settings_screen.dart # App settings and user profile
│   └── widgets/                  # Reusable widgets
│       └── error_display.dart   # Error display components
├── test/                         # Tests
│   └── widget_test.dart         # Widget tests
├── android/                      # Android platform code
├── ios/                          # iOS platform code
├── pubspec.yaml                 # Dependencies and configuration
└── README.md                    # Project documentation
```

## Key Components

### Entry Point
- **lib/main.dart**: Application entry point, initializes the app and routes to LoginScreen

### Models
Located in `lib/models/`:
- **user.dart**: User data model with authentication info
- **meeting.dart**: Meeting-related models including MeetingWithDetails, CreateMeetingRequest

### Services
Located in `lib/services/`:
- **api_client.dart**: HTTP client wrapper handling requests, responses, and ApiException
- **auth_service.dart**: Handles login, logout, token management
- **meetings_service.dart**: Meeting CRUD operations and LiveKit token retrieval
- **storage_service.dart**: Local data persistence using SharedPreferences
  - Token storage
  - User data (username, email, role)
  - Methods: `getUsername()`, `getEmail()`, `getRole()`, `getUserData()`
- **config_service.dart**: App configuration management (API URL)

### Screens
Located in `lib/screens/`:
- **login_screen.dart**:
  - User authentication
  - API URL configuration
  - Default credentials: admin@recontext.online / admin123
  - Enhanced error display with stack traces
- **main_screen.dart**: Bottom navigation with 4 tabs (Meetings, Documents, Search, Settings)
- **meetings_screen.dart**:
  - Display meetings list with filters (scheduled, in_progress, completed)
  - Meeting details
  - Video conference integration
- **documents_screen.dart**:
  - Documents and transcripts management
  - Filters by document type
- **search_screen.dart**:
  - Search meetings, transcripts, and documents
  - Filter results by type
- **settings_screen.dart**:
  - User profile display
  - API server configuration
  - App settings
  - Logout functionality

### Widgets
Located in `lib/widgets/`:
- **error_display.dart**: Reusable error display components
  - `ErrorDisplay`: Inline error with copy functionality
  - `FullScreenError`: Full-screen error display for critical errors
  - Features: selectable text, copy to clipboard, retry button

## Navigation Flow

```
LoginScreen (initial)
    ↓ (after successful login)
MainScreen (bottom navigation)
    ├── MeetingsScreen (Tab 0)
    ├── DocumentsScreen (Tab 1)
    ├── SearchScreen (Tab 2)
    └── SettingsScreen (Tab 3)
        └── Logout → back to LoginScreen
```

## Error Handling

All screens implement comprehensive error handling:
- API errors with status codes
- Connection errors with troubleshooting tips
- Unexpected errors with stack traces
- Copy-to-clipboard functionality
- Retry buttons for recoverable errors

## State Management

- Uses Flutter's built-in StatefulWidget
- AutomaticKeepAliveClientMixin for persistent screen state
- SharedPreferences for persistent data storage

## API Integration

Base URL configuration:
- Default: `https://portal.recontext.online`
- Configurable via Settings screen or Login screen
- Stored in SharedPreferences

## Dependencies

Key packages (from pubspec.yaml):
- **http**: HTTP client for API requests
- **shared_preferences**: Local storage
- **provider**: State management (planned)
- **livekit_client**: Video conferencing
- **intl**: Date/time formatting
- **flutter_lints**: Code quality

## Recent Changes

### Error Handling Improvements
- Added `ErrorDisplay` and `FullScreenError` widgets
- Enhanced error messages with stack traces
- Added copy-to-clipboard functionality
- Updated all screens to use new error widgets

### Storage Service
- Added individual getter methods: `getUsername()`, `getEmail()`, `getRole()`
- Fixed missing methods in SettingsScreen

### UI Updates
- Replaced deprecated `withOpacity()` with `withValues(alpha:)`
- Fixed BuildContext async gap warnings
- Added default login credentials for development

## Testing

- Widget tests in `test/widget_test.dart`
- Run with: `flutter test`
- Analyze with: `flutter analyze`

## Next Steps / TODO

- Implement real video conference screen with LiveKit
- Add meeting creation/editing forms
- Implement document upload/download
- Add real-time search with API integration
- Implement push notifications
- Add internationalization (i18n)
- Implement actual API endpoints (currently using mock data in some screens)

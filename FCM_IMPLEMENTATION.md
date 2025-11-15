# FCM (Firebase Cloud Messaging) Implementation

This document describes the implementation of push notification support for the Recontext.online mobile application using Firebase Cloud Messaging (FCM).

## Overview

The FCM implementation allows users to receive push notifications on their mobile devices. Each user can have multiple devices registered simultaneously, enabling notifications across all their devices (phones, tablets, etc.).

## Architecture

### Backend Components

#### 1. Database Model (`internal/models/fcm_device.go`)

**FCMDevice Model:**
- Stores device registration information
- Supports multiple devices per user
- Tracks device platform (iOS, Android, Web)
- Maintains device metadata (model, app version, OS version)
- Soft delete support (marks devices as inactive instead of deleting)

**Key Fields:**
- `id` - Unique device identifier
- `user_id` - Owner of the device
- `fcm_token` - Firebase Cloud Messaging token (unique)
- `platform` - Device platform (ios/android/web)
- `device_model` - Device model/name
- `app_version` - Application version
- `os_version` - Operating system version
- `is_active` - Whether device is currently active
- `last_active_at` - Last activity timestamp

#### 2. Database Repository (`pkg/database/fcm_device_repository.go`)

**FCMDeviceRepository** provides methods for:
- `RegisterDevice()` - Register new device or update existing
- `UnregisterDevice()` - Mark device as inactive
- `GetUserDevices()` - Get all active devices for a user
- `GetUserFCMTokens()` - Get all FCM tokens for a user (for multi-device push)
- `DeleteDevice()` - Permanently delete device record
- `CleanupInactiveDevices()` - Remove old inactive devices
- `UpdateDeviceActivity()` - Update last active timestamp

#### 3. API Endpoints (`cmd/user-portal/main.go`)

**POST /api/v1/fcm/register**
- Register or update a device for push notifications
- Requires authentication
- Request body: `RegisterFCMDeviceRequest`
  ```json
  {
    "fcm_token": "string",
    "platform": "ios|android|web",
    "device_model": "string (optional)",
    "app_version": "string (optional)",
    "os_version": "string (optional)"
  }
  ```
- Response: `RegisterFCMDeviceResponse`
  ```json
  {
    "id": "uuid",
    "message": "string",
    "is_new": boolean
  }
  ```

**POST /api/v1/fcm/unregister**
- Unregister a device from push notifications
- Requires authentication
- Request body: `UnregisterFCMDeviceRequest`
  ```json
  {
    "fcm_token": "string"
  }
  ```
- Response: `UnregisterFCMDeviceResponse`
  ```json
  {
    "message": "string"
  }
  ```

#### 4. Database Migration

The FCM devices table is automatically created via GORM AutoMigrate with the following indexes:
- `idx_fcm_devices_user_id` - Fast user lookup
- `idx_fcm_devices_fcm_token` - Unique token constraint
- `idx_fcm_devices_platform` - Platform filtering
- `idx_fcm_devices_is_active` - Active device filtering
- `idx_fcm_devices_last_active_at` - Activity-based queries

### Mobile Components

#### 1. Dependencies (`mobile2/pubspec.yaml`)

Added packages:
- `firebase_core: ^3.8.1` - Firebase core SDK
- `firebase_messaging: ^15.1.5` - Firebase Cloud Messaging
- `flutter_local_notifications: ^18.0.1` - Local notification display

#### 2. FCM Service (`mobile2/lib/services/fcm_service.dart`)

**FCMService** class provides:
- Firebase initialization and permission requests
- FCM token management and refresh handling
- Foreground and background message handling
- Local notification display
- Device registration/unregistration with backend

**Key Methods:**
- `initialize()` - Initialize FCM and request permissions
- `registerDevice()` - Register device with backend API
- `unregisterDevice()` - Unregister device from backend
- `_showLocalNotification()` - Display notification when app is in foreground
- Background message handler for notifications when app is closed

**Features:**
- Automatic token refresh handling
- Foreground notification display
- Background notification handling
- Notification tap handling (placeholder for navigation)
- Platform detection (iOS/Android/Web)
- Device information collection

#### 3. Firebase Initialization (`mobile2/lib/main.dart`)

- Firebase initialized in `main()` before app starts
- Background message handler registered at top level
- Handler function must be top-level (not class method)

#### 4. Login Integration (`mobile2/lib/screens/login_screen.dart`)

**Login Flow with FCM:**
1. User enters credentials and logs in
2. Authentication succeeds
3. User locale is loaded from profile
4. FCM service is initialized:
   - Requests notification permissions
   - Obtains FCM token
   - Registers device with backend
5. User is navigated to main screen

**Error Handling:**
- FCM initialization failures are logged but don't block login
- Ensures app remains functional even if push notifications fail

## Multi-Device Support

The implementation supports multiple devices per user:

1. **Unique Token Constraint:** Each FCM token is unique in the database
2. **User Association:** Each device is linked to a specific user ID
3. **Platform Tracking:** Devices are identified by platform (iOS/Android)
4. **Soft Delete:** Unregistering marks device as inactive rather than deleting
5. **Broadcasting:** `GetUserFCMTokens()` retrieves all active tokens for a user

**Example Use Case:**
A user with an iPhone, iPad, and Android phone:
- All three devices register independently
- Each gets a unique FCM token
- Backend stores all three tokens linked to the user
- When sending a notification, backend can send to all active devices

## Security Considerations

1. **Authentication Required:** All FCM endpoints require valid JWT token
2. **User Isolation:** Users can only manage their own devices
3. **Token Privacy:** FCM tokens are never exposed in logs (truncated)
4. **Inactive Cleanup:** Provides method to cleanup old inactive devices

## Firebase Configuration

**IMPORTANT:** Before the app can work, you need to configure Firebase:

### iOS Configuration
1. Create a Firebase project at https://console.firebase.google.com
2. Add an iOS app to the project
3. Download `GoogleService-Info.plist`
4. Place it in `mobile2/ios/Runner/`
5. Configure APNs (Apple Push Notification service) in Firebase Console

### Android Configuration
1. Add an Android app to the Firebase project
2. Download `google-services.json`
3. Place it in `mobile2/android/app/`
4. The project already has necessary Gradle configurations

### Required Files (Not in Repository)
- `mobile2/ios/Runner/GoogleService-Info.plist`
- `mobile2/android/app/google-services.json`
- `firebase_options.dart` (generated via FlutterFire CLI)

## Testing the Implementation

### Backend Testing

1. **Start the backend:**
   ```bash
   cd cmd/user-portal
   go run main.go
   ```

2. **Check migrations:**
   - Verify `fcm_devices` table is created
   - Verify indexes are created

3. **Test API endpoints:**
   ```bash
   # Register device
   curl -X POST http://localhost:8081/api/v1/fcm/register \
     -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "fcm_token": "test_token_123",
       "platform": "ios",
       "device_model": "iPhone 14 Pro"
     }'

   # Unregister device
   curl -X POST http://localhost:8081/api/v1/fcm/unregister \
     -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "fcm_token": "test_token_123"
     }'
   ```

### Mobile Testing

1. **Install dependencies:**
   ```bash
   cd mobile2
   flutter pub get
   ```

2. **Set up Firebase:**
   - Add Firebase configuration files (see above)
   - Run FlutterFire CLI: `flutterfire configure`

3. **Run the app:**
   ```bash
   flutter run
   ```

4. **Test flow:**
   - Launch app
   - Log in with valid credentials
   - Check logs for FCM initialization
   - Verify device registration in backend database

## Future Enhancements

### Planned Features
1. **Notification Categories:**
   - Meeting reminders
   - Recording completion notifications
   - System alerts
   - Chat messages

2. **Notification Preferences:**
   - Per-user notification settings
   - Quiet hours configuration
   - Notification type filtering

3. **Advanced Features:**
   - Rich notifications with images
   - Action buttons in notifications
   - Deep linking to specific screens
   - Notification history

4. **Analytics:**
   - Track notification delivery rates
   - Monitor user engagement
   - Device platform statistics

### Backend Enhancements
1. **Notification Sending Service:**
   ```go
   func SendNotificationToUser(userID uuid.UUID, notification Notification) error {
       // Get all active FCM tokens for user
       tokens := fcmDeviceRepo.GetUserFCMTokens(userID)

       // Send to all devices
       // Use Firebase Admin SDK
   }
   ```

2. **Batch Notifications:**
   - Send to multiple users efficiently
   - Topic-based notifications
   - User group notifications

3. **Cleanup Job:**
   ```go
   // Cron job to cleanup devices inactive for 90+ days
   fcmDeviceRepo.CleanupInactiveDevices(90 * 24 * time.Hour)
   ```

## Troubleshooting

### Common Issues

1. **FCM Token is null:**
   - Ensure Firebase is properly configured
   - Check that GoogleService-Info.plist/google-services.json are in place
   - Verify app has notification permissions

2. **Permissions denied:**
   - Check device notification settings
   - Re-request permissions after user denies

3. **Backend registration fails:**
   - Verify JWT token is valid
   - Check API URL configuration
   - Review backend logs for errors

4. **Notifications not received:**
   - Verify device is registered and active in database
   - Check Firebase Console for delivery status
   - Ensure APNs/FCM credentials are configured

### Debug Logging

Enable detailed logging in `mobile2/lib/utils/logger.dart`:
```dart
const bool kEnableDebugLogs = true;
```

Check logs for:
- FCM initialization status
- Token generation
- Device registration responses
- Notification delivery

## Files Created/Modified

### Backend
- **Created:**
  - `internal/models/fcm_device.go` - FCM device model
  - `pkg/database/fcm_device_repository.go` - Database operations

- **Modified:**
  - `cmd/user-portal/main.go` - Added FCM endpoints and handlers
  - `pkg/database/database.go` - Added FCM table to migrations

### Mobile
- **Created:**
  - `mobile2/lib/services/fcm_service.dart` - FCM service implementation

- **Modified:**
  - `mobile2/pubspec.yaml` - Added Firebase dependencies
  - `mobile2/lib/main.dart` - Firebase initialization
  - `mobile2/lib/screens/login_screen.dart` - FCM registration on login

## Summary

The FCM implementation provides a complete foundation for push notifications:
- ✅ Multi-device support per user
- ✅ Platform-aware (iOS/Android/Web)
- ✅ Automatic token refresh handling
- ✅ Foreground and background notifications
- ✅ Database tracking and cleanup
- ✅ Secure API endpoints
- ✅ Non-blocking login flow
- ✅ Comprehensive error handling

The system is ready for production use once Firebase is configured with the appropriate credentials.

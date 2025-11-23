# Recording Fields Refactoring Summary

## Overview
This refactoring consolidates the separate `needs_video_record` and `needs_audio_record` fields into a single unified `needs_record` field that handles both audio and video recording together.

## Changes Made

### 1. Backend API Models (`internal/models/meeting.go`)
**Changed:**
- Removed `NeedsVideoRecord` and `NeedsAudioRecord` fields from `Meeting` struct
- Added single `NeedsRecord` field
- Updated `CreateMeetingRequest` to use `needs_record` instead of separate fields
- Updated `UpdateMeetingRequest` to use `needs_record` instead of separate fields

**Files Modified:**
- `internal/models/meeting.go`

### 2. Database Models (`pkg/database/models.go`)
**Changed:**
- Removed `NeedsVideoRecord` and `NeedsAudioRecord` from database `Meeting` struct
- Added `NeedsRecord` field with proper GORM column mapping

**Files Modified:**
- `pkg/database/models.go`

### 3. Database Repository (`pkg/database/meeting_repository.go`)
**Changed:**
- Updated `CreateMeeting()` to use `needs_record` field
- Updated `GetMeetingByID()` to map `needs_record` field
- Updated `ListMeetings()` to use `needs_record` field
- Updated `UpdateMeeting()` to use `needs_record` field

**Files Modified:**
- `pkg/database/meeting_repository.go`

### 4. Webhook Handlers (`cmd/managing-portal/handlers_livekit.go`)
**Changed:**
- `handleRoomStarted()`: Updated to use `meeting.NeedsRecord` for both audio and video
- `handleTrackPublished()`: Updated to use `meeting.NeedsRecord` for determining recording requirements
- Logic simplified: when `needs_record` is true, both audio and video tracks are recorded

**Files Modified:**
- `cmd/managing-portal/handlers_livekit.go` (lines 208-215, 600-605)

**Webhook Behavior:**
- When a meeting starts (`room_started` event), if `needs_record` is true, `is_recording` flag is automatically enabled
- When tracks are published (`track_published` event), the system checks `needs_record` to determine whether to start egress recording
- Recording now starts automatically via webhook when tracks are published (both audio and video)

### 5. Mobile App Models (`mobile2/lib/models/meeting.dart`)
**Changed:**
- `Meeting` class: Removed `needsVideoRecord` and `needsAudioRecord`, added `needsRecord`
- `MeetingWithDetails` class: Updated constructor parameters
- `CreateMeetingRequest` class: Removed separate recording fields, added unified `needsRecord`
- Updated JSON serialization/deserialization for all classes

**Files Modified:**
- `mobile2/lib/models/meeting.dart`

### 6. Mobile App UI (`mobile2/lib/screens/create_meeting_screen.dart`)
**Changed:**
- Removed separate `_needsVideoRecord` and `_needsAudioRecord` state variables
- Added single `_needsRecord` state variable (default: `true`)
- Simplified UI: Single "Record Meeting (Audio & Video)" switch instead of two separate switches
- Updated request creation to use `needsRecord` field
- Updated logging to reflect new field

**Files Modified:**
- `mobile2/lib/screens/create_meeting_screen.dart`

### 7. Localization Files
**Changed English (`mobile2/lib/l10n/app_en.arb`):**
- Removed: `"meetingVideoRecord"` and `"meetingAudioRecord"`
- Added: `"meetingRecord": "Record Meeting (Audio & Video)"`

**Changed Russian (`mobile2/lib/l10n/app_ru.arb`):**
- Removed: `"meetingVideoRecord"` and `"meetingAudioRecord"`
- Added: `"meetingRecord": "Запись встречи (Аудио и Видео)"`

**Files Modified:**
- `mobile2/lib/l10n/app_en.arb`
- `mobile2/lib/l10n/app_ru.arb`

### 8. Database Migration (`migrations/001_merge_recording_fields.sql`)
**Created:**
- New migration file to handle the database schema change
- Adds `needs_record` column
- Migrates existing data: sets `needs_record = true` if either old field was true
- Includes commented-out DROP statements for old columns (to be executed after verification)

**Files Created:**
- `migrations/001_merge_recording_fields.sql`

## Migration Instructions

### 1. Run Database Migration
```sql
-- Run this migration on your database
psql -U your_user -d your_database -f migrations/001_merge_recording_fields.sql
```

### 2. Verify Data Migration
```sql
-- Check that data was migrated correctly
SELECT id, title, needs_record, needs_video_record, needs_audio_record
FROM meetings
LIMIT 10;
```

### 3. Drop Old Columns (After Verification)
```sql
-- Once you've verified everything works correctly:
ALTER TABLE meetings DROP COLUMN IF EXISTS needs_video_record;
ALTER TABLE meetings DROP COLUMN IF EXISTS needs_audio_record;
```

### 4. Rebuild Mobile App
```bash
cd mobile2
flutter pub get
flutter build apk  # or flutter build ios
```

### 5. Restart Backend Services
```bash
# Restart your Go backend services to pick up the new code
systemctl restart managing-portal  # or your service name
```

## Testing Checklist

### Backend Testing
- [ ] Create a new meeting with `needs_record: true`
- [ ] Create a new meeting with `needs_record: false`
- [ ] Update an existing meeting to toggle `needs_record`
- [ ] Verify webhook `room_started` event correctly sets `is_recording`
- [ ] Verify webhook `track_published` event starts egress when `needs_record` is true
- [ ] Verify existing meetings still work correctly

### Mobile App Testing
- [ ] Open create meeting screen
- [ ] Verify only one "Record Meeting" switch is displayed
- [ ] Toggle the switch and create a meeting
- [ ] Verify the meeting is created with correct `needs_record` value
- [ ] View existing meetings and verify they display correctly

### API Testing
```bash
# Test creating a meeting
curl -X POST http://your-api/api/v1/meetings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "Test Meeting",
    "scheduled_at": "2025-01-24T10:00:00Z",
    "duration": 60,
    "type": "conference",
    "needs_record": true,
    "needs_transcription": false
  }'
```

## Breaking Changes

### API Changes
- **POST `/api/v1/meetings`**: Request body changed
  - ❌ Removed: `needs_video_record`, `needs_audio_record`
  - ✅ Added: `needs_record`

- **PUT `/api/v1/meetings/{id}`**: Request body changed
  - ❌ Removed: `needs_video_record`, `needs_audio_record`
  - ✅ Added: `needs_record`

- **GET `/api/v1/meetings`** and **GET `/api/v1/meetings/{id}`**: Response changed
  - ❌ Removed: `needs_video_record`, `needs_audio_record`
  - ✅ Added: `needs_record`

### Client Compatibility
- **Old mobile apps** will need to be updated - they will fail to parse meetings without the old fields
- **Old API clients** sending `needs_video_record` or `needs_audio_record` will receive validation errors

## Rollback Plan

If you need to rollback:

1. Restore old code from git
2. Run this SQL to add back old columns (if dropped):
```sql
ALTER TABLE meetings ADD COLUMN needs_video_record BOOLEAN DEFAULT false;
ALTER TABLE meetings ADD COLUMN needs_audio_record BOOLEAN DEFAULT false;

-- Restore data from needs_record
UPDATE meetings
SET needs_video_record = needs_record,
    needs_audio_record = needs_record
WHERE needs_record = true;
```

## Benefits of This Refactoring

1. **Simplified UX**: Users no longer need to decide between audio and video recording - one switch controls both
2. **Cleaner Code**: Reduced complexity in webhook handlers and business logic
3. **Better Semantics**: Recording a meeting naturally includes both audio and video
4. **Easier Maintenance**: Fewer fields to track and validate
5. **Webhook Simplification**: Automatic recording start is now clearer and more predictable

## Notes

- The `needs_transcription` field remains separate as it serves a different purpose (recording individual participant tracks for transcription)
- The `is_recording` and `is_transcribing` runtime flags remain separate for real-time status tracking
- Webhook events (`egress_started`, `track_published`) now automatically trigger recording based on the `needs_record` flag

# Webhook-Based Recording Flow

## Overview
Recording now starts automatically via webhook events when participants publish their tracks to the meeting room. This eliminates the need for manual recording control from the frontend.

## Recording Flow

### 1. Meeting Creation
```
User creates meeting with needs_record=true
     ↓
Meeting stored in database
     ↓
needs_record=true, is_recording=false, is_transcribing=false
```

### 2. Room Started Event
```
LiveKit sends room_started webhook
     ↓
handlers_livekit.go: handleRoomStarted()
     ↓
Checks if room name is a valid meeting UUID
     ↓
Loads meeting from database
     ↓
If needs_record=true:
  - Sets is_recording=true automatically
  - Updates meeting in database
```

**File**: `cmd/managing-portal/handlers_livekit.go` (lines 206-236)

**Logic**:
```go
if meeting.NeedsRecord && !meeting.IsRecording {
    meeting.IsRecording = true
    updateNeeded = true
    mp.logger.Infof("Auto-enabling is_recording for meeting %s (needs_record=%v)",
        meetingID, meeting.NeedsRecord)
}
```

### 3. Track Published Event
```
Participant publishes audio/video track
     ↓
LiveKit sends track_published webhook
     ↓
handlers_livekit.go: handleTrackPublished()
     ↓
Loads meeting settings from database
     ↓
Checks meeting.NeedsRecord flag
     ↓
If needs_record=true:
  - Determines if track is audio or video
  - Starts egress recording for the track asynchronously
  - Creates EgressRecording database entry
  - Stores egress_id in track record
```

**File**: `cmd/managing-portal/handlers_livekit.go` (lines 593-697)

**Logic**:
```go
// Get meeting settings
meeting, err := mp.meetingRepo.GetMeetingByID(meetingID)
needsAudioRecord = meeting.NeedsRecord  // unified field
needsVideoRecord = meeting.NeedsRecord  // unified field

// Determine track type
isAudioTrack := track.Source == "MICROPHONE" || ...
isVideoTrack := track.Type == "video" || ...

shouldRecordAudio := isAudioTrack && needsAudioRecord
shouldRecordVideo := isVideoTrack && needsVideoRecord

if shouldRecordAudio || shouldRecordVideo {
    // Start track egress asynchronously
    go func() {
        egressID, err := mp.startTrackCompositeEgress(...)
        // Save egress info to database
    }()
}
```

### 4. Egress Started Event
```
LiveKit egress service starts recording
     ↓
LiveKit sends egress_started webhook
     ↓
handlers_livekit.go: handleEgressStarted()
     ↓
Updates EgressRecording status to "active"
     ↓
Recording is now in progress
```

**File**: `cmd/managing-portal/handlers_livekit.go` (lines 1042-1080)

### 5. Track Recording
```
While meeting is active:
  - Each participant's audio/video tracks are recorded separately
  - Egress creates HLS segments (.ts files) and playlists (.m3u8)
  - Files stored in MinIO/S3: {meetingID}_{roomSID}/tracks/{trackSID}/
```

### 6. Track Unpublished Event
```
Participant stops publishing track (leaves or stops camera/mic)
     ↓
LiveKit sends track_unpublished webhook
     ↓
handlers_livekit.go: handleTrackUnpublished()
     ↓
Stops egress recording for that track
     ↓
If track is audio and needs_transcription=true:
  - Creates transcription task
  - Sends to RabbitMQ queue
```

**File**: `cmd/managing-portal/handlers_livekit.go` (lines 713-891)

### 7. Egress Ended Event
```
LiveKit egress finishes recording
     ↓
LiveKit sends egress_ended webhook
     ↓
handlers_livekit.go: handleEgressEnded()
     ↓
Updates EgressRecording with file_path and status
     ↓
If audio track + needs_transcription=true:
  - Creates transcription task
  - Sends to RabbitMQ
```

**File**: `cmd/managing-portal/handlers_livekit.go` (lines 1133-1313)

## Key Differences: Before vs After

### Before (Separate Fields)
```json
{
  "needs_video_record": true,
  "needs_audio_record": true,
  "needs_transcription": false
}
```

**Webhook Logic**:
- Check both `needs_video_record` AND `needs_audio_record`
- Complex conditional logic: "if video, then audio must also be true"
- Two separate switches in UI

### After (Unified Field)
```json
{
  "needs_record": true,
  "needs_transcription": false
}
```

**Webhook Logic**:
- Check single `needs_record` flag
- When true, both audio and video tracks are recorded
- Simple boolean logic
- Single "Record Meeting" switch in UI

## Database Schema

### EgressRecording Table
Tracks all egress recording sessions:

```sql
CREATE TABLE livekit_egress_recordings (
    id UUID PRIMARY KEY,
    egress_id VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL,  -- 'track_composite'
    status VARCHAR(50) NOT NULL, -- 'started', 'active', 'ended', 'failed'
    room_sid VARCHAR(255) NOT NULL,
    room_name VARCHAR(255) NOT NULL,
    meeting_id UUID,
    track_sid VARCHAR(255),
    participant_sid VARCHAR(255),
    file_path TEXT,
    audio_only BOOLEAN DEFAULT false,
    started_at TIMESTAMP NOT NULL,
    ended_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

## Webhook Event Sequence Example

```
User creates meeting with needs_record=true, needs_transcription=true
     ↓
1. room_started
   - Meeting.is_recording = true
     ↓
2. track_published (participant 1, audio)
   - Start audio egress
   - EgressRecording created (status=started)
     ↓
3. egress_started (audio)
   - EgressRecording updated (status=active)
     ↓
4. track_published (participant 1, video)
   - Start video egress
   - EgressRecording created (status=started)
     ↓
5. egress_started (video)
   - EgressRecording updated (status=active)
     ↓
6. track_published (participant 2, audio)
   - Start audio egress for participant 2
   - EgressRecording created (status=started)
     ↓
... recording continues ...
     ↓
7. track_unpublished (participant 1, audio)
   - Stop audio egress
   - Create transcription task → RabbitMQ
     ↓
8. egress_ended (participant 1, audio)
   - EgressRecording updated (status=ended, file_path set)
   - Create transcription task → RabbitMQ (backup)
     ↓
9. room_finished
   - Stop all remaining egress
   - Meeting.is_recording = false
   - Meeting.is_transcribing = false
   - Meeting.status = 'completed' (if not permanent)
```

## Configuration

### Environment Variables
```bash
# Enable egress recording
LIVEKIT_EGRESS_ENABLED=true

# Enable per-track recording (required for our use case)
LIVEKIT_EGRESS_RECORD_TRACKS=true

# S3/MinIO configuration for storing recordings
LIVEKIT_EGRESS_S3_ENDPOINT=minio:9000
LIVEKIT_EGRESS_S3_BUCKET=recontext
LIVEKIT_EGRESS_S3_ACCESS_KEY=minioadmin
LIVEKIT_EGRESS_S3_SECRET=minioadmin
LIVEKIT_EGRESS_S3_REGION=us-east-1
```

### LiveKit Configuration
The webhook URL must be configured in LiveKit server config:
```yaml
webhook:
  api_key: your-api-key
  urls:
    - http://managing-portal:8080/webhook/meet
```

## Monitoring and Debugging

### Check if Recording Started
```sql
-- Check meeting recording status
SELECT id, title, needs_record, is_recording, is_transcribing, status
FROM meetings
WHERE id = 'your-meeting-id';

-- Check active egress recordings
SELECT egress_id, type, status, room_name, track_sid, started_at
FROM livekit_egress_recordings
WHERE meeting_id = 'your-meeting-id'
ORDER BY started_at DESC;
```

### View Webhook Logs
```bash
# Check managing-portal logs
docker-compose logs -f managing-portal | grep "Processing.*event"

# Watch for specific events
docker-compose logs -f managing-portal | grep "track_published\|egress_started"
```

### Common Issues

**Problem**: Recording doesn't start
- Check: `needs_record` is set to `true` in database
- Check: LiveKit webhook URL is configured correctly
- Check: `LIVEKIT_EGRESS_ENABLED=true` in environment
- Check: Egress service is running in LiveKit

**Problem**: Transcription doesn't work
- Check: `needs_transcription` is set to `true`
- Check: RabbitMQ is running and accessible
- Check: Transcription service is consuming from queue
- Check: Audio files are accessible via configured storage URL

## API Endpoints for Manual Control

While recording starts automatically, you can still manually control it:

```bash
# Manually update meeting recording flags
PUT /api/v1/meetings/{id}
{
  "is_recording": true,
  "is_transcribing": true
}

# List egress recordings for a meeting
GET /api/v1/livekit/egress?room_name={meetingId}

# View specific egress status
GET /api/v1/livekit/egress/{egressId}
```

## Summary

The new webhook-based recording flow:
1. ✅ Automatically starts when participants join
2. ✅ Records each participant's tracks separately
3. ✅ Handles transcription task creation
4. ✅ Simplifies frontend - no manual start/stop needed
5. ✅ Unified `needs_record` field controls both audio and video
6. ✅ Tracks all recordings in database for monitoring

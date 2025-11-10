# Meetings API Guide

This guide provides examples for using the Video Meetings API.

## Table of Contents

1. [Meeting Subjects Management](#meeting-subjects-management)
2. [Meetings CRUD Operations](#meetings-crud-operations)
3. [Meeting Filters and Search](#meeting-filters-and-search)
4. [Common Use Cases](#common-use-cases)

## Meeting Subjects Management

### Create a Meeting Subject (Admin Only)

```bash
curl -X POST http://localhost:8080/api/v1/meeting-subjects \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Quarterly Review",
    "description": "Regular quarterly business reviews",
    "department_ids": ["dept-001", "dept-002"]
  }'
```

### List Meeting Subjects

```bash
# List all active subjects
curl http://localhost:8080/api/v1/meeting-subjects \
  -H "Authorization: Bearer $TOKEN"

# Filter by department
curl "http://localhost:8080/api/v1/meeting-subjects?department_id=dept-001" \
  -H "Authorization: Bearer $TOKEN"

# Include inactive subjects
curl "http://localhost:8080/api/v1/meeting-subjects?include_inactive=true" \
  -H "Authorization: Bearer $TOKEN"
```

### Get Subject Details

```bash
curl http://localhost:8080/api/v1/meeting-subjects/subj-abc123 \
  -H "Authorization: Bearer $TOKEN"
```

### Update Subject (Admin Only)

```bash
curl -X PUT http://localhost:8080/api/v1/meeting-subjects/subj-abc123 \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Subject Name",
    "description": "Updated description",
    "department_ids": ["dept-001", "dept-003"],
    "is_active": true
  }'
```

### Delete Subject (Admin Only)

```bash
curl -X DELETE http://localhost:8080/api/v1/meeting-subjects/subj-abc123 \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

## Meetings CRUD Operations

### Create a Meeting

**Conference (all participants equal):**

```bash
curl -X POST http://localhost:8081/api/v1/meetings \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Weekly Team Sync",
    "scheduled_at": "2025-11-15T14:00:00Z",
    "duration": 60,
    "recurrence": "weekly",
    "type": "conference",
    "subject_id": "subj-abc123",
    "needs_video_record": true,
    "needs_audio_record": true,
    "additional_notes": "Agenda:\n1. Project updates\n2. Blockers\n3. Next steps",
    "participant_user_ids": ["user-001", "user-002", "user-003"],
    "department_ids": ["dept-marketing"]
  }'
```

**Presentation (with speaker):**

```bash
curl -X POST http://localhost:8081/api/v1/meetings \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Q4 Results Presentation",
    "scheduled_at": "2025-11-20T15:00:00Z",
    "duration": 90,
    "recurrence": "none",
    "type": "presentation",
    "subject_id": "subj-quarterly",
    "needs_video_record": true,
    "needs_audio_record": false,
    "speaker_id": "user-ceo",
    "participant_user_ids": [],
    "department_ids": ["dept-sales", "dept-marketing", "dept-engineering"]
  }'
```

### List My Meetings

```bash
# List all my meetings (where I'm a participant or speaker)
curl http://localhost:8081/api/v1/meetings \
  -H "Authorization: Bearer $USER_TOKEN"

# With pagination
curl "http://localhost:8081/api/v1/meetings?page=1&page_size=20" \
  -H "Authorization: Bearer $USER_TOKEN"
```

### Get Meeting Details

```bash
curl http://localhost:8081/api/v1/meetings/meet-xyz789 \
  -H "Authorization: Bearer $USER_TOKEN"
```

Response:
```json
{
  "id": "meet-xyz789",
  "title": "Weekly Team Sync",
  "scheduled_at": "2025-11-15T14:00:00Z",
  "duration": 60,
  "recurrence": "weekly",
  "type": "conference",
  "subject_id": "subj-abc123",
  "status": "scheduled",
  "needs_video_record": true,
  "needs_audio_record": true,
  "additional_notes": "Agenda:\n1. Project updates\n2. Blockers\n3. Next steps",
  "livekit_room_id": null,
  "created_by": "user-manager",
  "created_at": "2025-11-09T10:00:00Z",
  "updated_at": "2025-11-09T10:00:00Z",
  "subject": {
    "id": "subj-abc123",
    "name": "Team Sync",
    "description": "Regular team synchronization meetings",
    "department_ids": ["dept-engineering"],
    "is_active": true,
    "created_at": "2025-11-01T00:00:00Z",
    "updated_at": "2025-11-01T00:00:00Z"
  },
  "participants": [
    {
      "meeting_id": "meet-xyz789",
      "user_id": "user-001",
      "role": "participant",
      "status": "accepted",
      "created_at": "2025-11-09T10:00:00Z",
      "user": {
        "id": "user-001",
        "username": "john.doe",
        "email": "john.doe@example.com",
        "full_name": "John Doe",
        "role": "user"
      }
    }
  ],
  "departments": [
    {
      "id": "dept-marketing",
      "name": "Marketing",
      "description": "Marketing department"
    }
  ],
  "creator": {
    "id": "user-manager",
    "username": "manager",
    "email": "manager@example.com",
    "full_name": "Team Manager",
    "role": "user"
  }
}
```

### Update a Meeting

```bash
curl -X PUT http://localhost:8081/api/v1/meetings/meet-xyz789 \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Meeting Title",
    "scheduled_at": "2025-11-16T14:00:00Z",
    "duration": 45,
    "status": "scheduled",
    "participant_user_ids": ["user-001", "user-004"],
    "department_ids": []
  }'
```

### Cancel (Delete) a Meeting

```bash
curl -X DELETE http://localhost:8081/api/v1/meetings/meet-xyz789 \
  -H "Authorization: Bearer $USER_TOKEN"
```

## Meeting Filters and Search

### Filter by Status

```bash
# Get only scheduled meetings
curl "http://localhost:8081/api/v1/meetings?status=scheduled" \
  -H "Authorization: Bearer $USER_TOKEN"

# Get in-progress meetings
curl "http://localhost:8081/api/v1/meetings?status=in_progress" \
  -H "Authorization: Bearer $USER_TOKEN"

# Get completed meetings
curl "http://localhost:8081/api/v1/meetings?status=completed" \
  -H "Authorization: Bearer $USER_TOKEN"
```

### Filter by Type

```bash
# Get only presentations
curl "http://localhost:8081/api/v1/meetings?type=presentation" \
  -H "Authorization: Bearer $USER_TOKEN"

# Get only conferences
curl "http://localhost:8081/api/v1/meetings?type=conference" \
  -H "Authorization: Bearer $USER_TOKEN"
```

### Filter by Subject

```bash
curl "http://localhost:8081/api/v1/meetings?subject_id=subj-abc123" \
  -H "Authorization: Bearer $USER_TOKEN"
```

### Filter by Date Range

```bash
# Meetings from a specific date onwards
curl "http://localhost:8081/api/v1/meetings?date_from=2025-11-15T00:00:00Z" \
  -H "Authorization: Bearer $USER_TOKEN"

# Meetings up to a specific date
curl "http://localhost:8081/api/v1/meetings?date_to=2025-11-30T23:59:59Z" \
  -H "Authorization: Bearer $USER_TOKEN"

# Meetings in a specific date range
curl "http://localhost:8081/api/v1/meetings?date_from=2025-11-15T00:00:00Z&date_to=2025-11-30T23:59:59Z" \
  -H "Authorization: Bearer $USER_TOKEN"
```

### Filter by Speaker (Presentations Only)

```bash
curl "http://localhost:8081/api/v1/meetings?speaker_id=user-ceo" \
  -H "Authorization: Bearer $USER_TOKEN"
```

### Combine Multiple Filters

```bash
# Get all scheduled presentations in November with video recording
curl "http://localhost:8081/api/v1/meetings?status=scheduled&type=presentation&date_from=2025-11-01T00:00:00Z&date_to=2025-11-30T23:59:59Z" \
  -H "Authorization: Bearer $USER_TOKEN"
```

## Common Use Cases

### 1. Schedule a Department-Wide Meeting

```bash
curl -X POST http://localhost:8081/api/v1/meetings \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "All-Hands Meeting",
    "scheduled_at": "2025-11-25T10:00:00Z",
    "duration": 120,
    "recurrence": "monthly",
    "type": "presentation",
    "subject_id": "subj-company-updates",
    "needs_video_record": true,
    "needs_audio_record": true,
    "speaker_id": "user-ceo",
    "participant_user_ids": [],
    "department_ids": ["dept-sales", "dept-marketing", "dept-engineering", "dept-hr"]
  }'
```

### 2. Get Today's Meetings

```bash
# Get meetings scheduled for today
TODAY_START="2025-11-09T00:00:00Z"
TODAY_END="2025-11-09T23:59:59Z"

curl "http://localhost:8081/api/v1/meetings?date_from=${TODAY_START}&date_to=${TODAY_END}" \
  -H "Authorization: Bearer $USER_TOKEN"
```

### 3. Get Upcoming Meetings (Next 7 Days)

```bash
# Get meetings in the next week
NOW="2025-11-09T00:00:00Z"
NEXT_WEEK="2025-11-16T23:59:59Z"

curl "http://localhost:8081/api/v1/meetings?status=scheduled&date_from=${NOW}&date_to=${NEXT_WEEK}" \
  -H "Authorization: Bearer $USER_TOKEN"
```

### 4. Schedule a Recurring Weekly 1:1

```bash
curl -X POST http://localhost:8081/api/v1/meetings \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Weekly 1:1 with Manager",
    "scheduled_at": "2025-11-11T15:00:00Z",
    "duration": 30,
    "recurrence": "weekly",
    "type": "conference",
    "subject_id": "subj-one-on-one",
    "needs_video_record": false,
    "needs_audio_record": false,
    "participant_user_ids": ["user-manager"],
    "department_ids": []
  }'
```

### 5. Update Meeting Time (Reschedule)

```bash
curl -X PUT http://localhost:8081/api/v1/meetings/meet-xyz789 \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "scheduled_at": "2025-11-15T16:00:00Z"
  }'
```

### 6. Add More Participants to Existing Meeting

```bash
# First get the current meeting to see existing participants
MEETING=$(curl http://localhost:8081/api/v1/meetings/meet-xyz789 \
  -H "Authorization: Bearer $USER_TOKEN")

# Then update with the new participant list
curl -X PUT http://localhost:8081/api/v1/meetings/meet-xyz789 \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "participant_user_ids": ["user-001", "user-002", "user-003", "user-new"]
  }'
```

### 7. Get Meeting History (Completed Meetings)

```bash
curl "http://localhost:8081/api/v1/meetings?status=completed&page=1&page_size=50" \
  -H "Authorization: Bearer $USER_TOKEN"
```

## Pagination Response Format

All list endpoints return data in this format:

```json
{
  "items": [
    {
      "id": "meet-001",
      "title": "Meeting 1",
      ...
    },
    {
      "id": "meet-002",
      "title": "Meeting 2",
      ...
    }
  ],
  "offset": 0,
  "page_size": 20,
  "total": 42
}
```

To navigate pages:
- `page=1` (first page)
- `page=2` (second page)
- etc.

## Permission Requirements

### User Portal (Port 8081)

- **Create Meeting**: Requires `can_schedule_meetings` permission OR admin/operator role
- **List Meetings**: Any authenticated user (shows only their meetings)
- **Get Meeting**: Participant, speaker, or admin
- **Update Meeting**: Creator or admin
- **Delete Meeting**: Creator or admin

### Managing Portal (Port 8080)

- **Subject CRUD**: Admin only

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid request body",
  "details": "Title is required"
}
```

### 401 Unauthorized
```json
{
  "error": "Unauthorized",
  "details": "Missing or invalid token"
}
```

### 403 Forbidden
```json
{
  "error": "You don't have permission to schedule meetings",
  "details": "can_schedule_meetings permission required"
}
```

### 404 Not Found
```json
{
  "error": "Meeting not found",
  "details": "Meeting with ID meet-xyz789 does not exist"
}
```

## Notes

1. **Timestamps**: All dates use ISO 8601 format (e.g., `2025-11-15T14:00:00Z`)
2. **Duration**: Specified in minutes
3. **Recurrence**: Values are `none`, `daily`, `weekly`, `monthly`
4. **Meeting Types**: `conference` (all equal), `presentation` (speaker + participants)
5. **Status**: `scheduled`, `in_progress`, `completed`, `cancelled`
6. **Participant Status**: `invited`, `accepted`, `declined`, `tentative`
7. **Roles**: `speaker` (presentations only), `participant`

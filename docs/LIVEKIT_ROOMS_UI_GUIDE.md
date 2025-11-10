# LiveKit Rooms Management UI Guide

## Overview

The Managing Portal now includes a comprehensive UI for monitoring and managing LiveKit video conference rooms. This guide explains how to use the new Rooms feature.

## Accessing the Rooms Section

1. Login to the Managing Portal at `http://localhost:8080` (or your deployed URL)
2. Default credentials: `admin` / `admin123`
3. Click on the **"Rooms"** menu item in the left sidebar (with a video camera icon)

## Rooms List Page

### Features

#### Room Cards
Each room is displayed as a card with the following information:
- **Room Name**: The display name of the meeting room
- **Status Badge**:
  - 🟢 **Active** - Room is currently in session (with pulsing dot animation)
  - ✓ **Finished** - Room session has ended
- **Started Time**: When the room was created
- **Duration**: How long the room has been active/was active
- **Room ID (SID)**: Unique identifier for the room

#### Filtering Options
- **Status Filter**: Filter rooms by status
  - **All** - Show all rooms
  - **Active** - Show only active rooms
  - **Finished** - Show only finished rooms

#### Auto-Refresh
- When viewing "All" or "Active" rooms, the list automatically refreshes every 5 seconds
- A pulsing indicator shows when auto-refresh is enabled
- You can manually refresh at any time using the "Refresh" button

#### Viewing Room Details
- Click on any room card or the "View Details" button to open the room details page

## Room Details Page

### Navigation
- Click the "Back" button to return to the rooms list
- The sidebar "Rooms" menu item remains highlighted

### Statistics Overview
Four statistics cards display:
1. **Participants**: Total number of participants who joined
2. **Tracks**: Total number of audio/video tracks published
3. **Started**: Room creation timestamp
4. **Duration**: Total duration of the room session

### Participants Tab

View all participants who joined the room:

**For Each Participant:**
- **Avatar Icon**: Visual representation
- **Name**: Participant's display name
- **Identity**: Unique participant identifier
- **State**: Connection state
  - **ACTIVE** - Currently connected
  - **DISCONNECTED** - Left the room
- **Joined**: Timestamp when participant joined
- **Left**: Timestamp when participant left (if applicable)
- **Disconnect Reason**: Why the participant left (if available)
- **Publisher Status**: Whether the participant published media

### Tracks Tab

View all audio/video tracks published in the room:

**Track Card Information:**
- **Track Icon**:
  - 🎤 Microphone (audio)
  - 📹 Camera (video)
  - 🖥️ Screen Share
- **Source**: MICROPHONE, CAMERA, or SCREEN_SHARE
- **Type**: AUDIO or VIDEO
- **Status**:
  - **PUBLISHED** - Currently active
  - **UNPUBLISHED** - No longer active
- **Codec**: MIME type (e.g., audio/opus, video/VP8)
- **Resolution**: Width × Height (for video tracks)
- **Simulcast**: Whether simulcast is enabled
- **Published At**: When the track was published
- **Unpublished At**: When the track was removed (if applicable)
- **Track SID**: Unique track identifier

## API Integration

The UI consumes the following backend APIs:

### List Rooms
```bash
GET /api/v1/livekit/rooms?status=active&limit=50&offset=0
Authorization: Bearer <token>
```

### Get Room Participants
```bash
GET /api/v1/livekit/participants?room_sid=RM_xxx
Authorization: Bearer <token>
```

### Get Room Tracks
```bash
GET /api/v1/livekit/tracks?room_sid=RM_xxx
Authorization: Bearer <token>
```

## Use Cases

### 1. Monitor Active Meetings
- Filter by "Active" status
- Watch real-time updates with auto-refresh
- Check participant count and track activity

### 2. Review Past Meetings
- Filter by "Finished" status
- Review participant join/leave history
- Analyze track usage and duration

### 3. Troubleshoot Issues
- Check disconnect reasons for participants
- Verify track publication status
- Review codec and resolution information

### 4. Audit Trail
- View complete meeting history
- Track participant activity
- Monitor track lifecycle events

## Technical Details

### Real-time Updates
- Active rooms auto-refresh every 5 seconds
- Status changes reflected immediately
- No manual refresh needed for active monitoring

### Responsive Design
- Grid layout adapts to screen size
- Cards resize for optimal viewing
- Mobile-friendly interface

### Performance
- Efficient API calls with pagination
- Conditional auto-refresh (only for active rooms)
- Lazy loading of room details

## Configuration

### Environment Variables
Set the API URL in your frontend environment:
```bash
VITE_API_URL=http://localhost:8080
```

### Backend Setup
Ensure the Managing Portal backend is running and the LiveKit webhook is configured:
```bash
# In LiveKit server configuration
webhook_url: http://your-server:8080/webhook/meet
```

## Troubleshooting

### No Rooms Displayed
- Check that LiveKit is sending webhooks to `/webhook/meet`
- Verify backend is running and accessible
- Check browser console for API errors

### Room Details Not Loading
- Verify JWT token is valid
- Check network tab for failed API calls
- Ensure room SID is correct in URL

### Auto-refresh Not Working
- Check that status filter is set to "All" or "Active"
- Verify no JavaScript errors in console
- Ensure page is in focus (some browsers pause timers)

## Future Enhancements

Planned features:
- Export room data to CSV/JSON
- Real-time notifications for new rooms
- Advanced filtering (by date range, participant count)
- Track quality metrics and analytics
- Integration with recording playback

## Support

For issues or feature requests, please contact the development team or create an issue in the project repository.

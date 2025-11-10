# LiveKit Rooms Feature - Managing Portal

## Overview

This document describes the LiveKit Rooms management feature added to the Managing Portal frontend.

## New Files

### Components
```
src/components/
├── Rooms.tsx           # Rooms list page
├── Rooms.css          # Rooms list styles
├── RoomDetails.tsx    # Room details page
└── RoomDetails.css    # Room details styles
```

### Services
```
src/services/
└── livekit.ts         # LiveKit API client
```

## Features Implemented

### 1. Rooms List (`/rooms`)
- Grid display of all meeting rooms
- Status filtering (All/Active/Finished)
- Auto-refresh every 5 seconds for active rooms
- Room cards with:
  - Name and status badge
  - Start time and duration
  - Room SID
  - Click to view details

### 2. Room Details (`/rooms/:sid`)
- Statistics cards (participants, tracks, duration)
- Tabbed interface:
  - **Participants Tab**: List of all participants with join/leave info
  - **Tracks Tab**: Audio/video tracks with codec details
- Back navigation to rooms list

## API Endpoints Used

All endpoints require JWT authentication:

```typescript
// List rooms
GET /api/v1/livekit/rooms?status={status}&limit={limit}&offset={offset}

// Get room by SID
GET /api/v1/livekit/rooms?sid={sid}

// Get participants
GET /api/v1/livekit/participants?room_sid={room_sid}

// Get tracks
GET /api/v1/livekit/tracks?room_sid={room_sid}
```

## TypeScript Interfaces

```typescript
interface Room {
  id: string;
  sid: string;
  name: string;
  status: string;
  startedAt: string;
  finishedAt?: string;
  emptyTimeout: number;
  departureTimeout: number;
  creationTime: string;
}

interface Participant {
  id: string;
  sid: string;
  identity: string;
  name: string;
  state: string;
  joinedAt: string;
  leftAt?: string;
  isPublisher: boolean;
  disconnectReason?: string;
}

interface Track {
  id: string;
  sid: string;
  type: string;
  source: string;
  mimeType: string;
  status: string;
  publishedAt: string;
  unpublishedAt?: string;
  width?: number;
  height?: number;
  simulcast?: boolean;
}
```

## Styling

### CSS Variables Used
```css
--primary-color         /* Blue for buttons and icons */
--success-color         /* Green for active status */
--error-color          /* Red for errors */
--text-primary         /* Main text color */
--text-secondary       /* Secondary text color */
--card-background      /* Card background */
--border-color         /* Border color */
--shadow-sm            /* Small shadow */
--shadow-md            /* Medium shadow */
```

### Key Design Elements
- Card-based layout with hover effects
- Status badges with color coding
- Pulse animation for active status
- Responsive grid layout
- Icon-based navigation

## Integration Points

### App.tsx
Added routes:
```typescript
{currentPath === '/rooms' && <Rooms />}
{roomSid && <RoomDetails roomSid={roomSid} />}
```

### Layout.tsx
Added navigation item:
```typescript
<a href="/rooms" className={`nav-item ${isActive('/rooms')}`}>
  <LuVideo />
  <span>Rooms</span>
</a>
```

## Development

### Running Locally
```bash
cd front/managing-portal
npm install
npm run dev
```

### Building for Production
```bash
npm run build
```

Output will be in `dist/` directory.

### Environment Configuration
Create `.env` file:
```
VITE_API_URL=http://localhost:8080
```

## Testing

### Manual Testing Steps
1. Login to managing portal
2. Click "Rooms" in sidebar
3. Verify rooms list displays
4. Try status filtering
5. Click on a room card
6. Verify room details page:
   - Statistics cards show correct data
   - Switch between Participants and Tracks tabs
   - Check all data displays correctly
7. Click "Back" button
8. Verify return to rooms list

### Test Data
To generate test data, trigger LiveKit webhook events:
```bash
# Create a LiveKit room and join it
# Or use the test script to simulate webhooks:
curl -X POST http://localhost:8080/webhook/meet \
  -H "Content-Type: application/json" \
  -d @test-webhook-data.json
```

## Browser Compatibility

Tested on:
- Chrome 120+
- Firefox 120+
- Safari 17+
- Edge 120+

## Dependencies

New dependencies added:
```json
{
  "react-icons": "^5.0.0"  // For LuVideo, LuUsers, etc.
}
```

Existing dependencies used:
- React
- TypeScript
- Vite

## Performance Considerations

- Auto-refresh uses `setInterval` and cleans up on unmount
- Only refreshes when viewing active rooms
- API calls use pagination (default limit: 50)
- Room details page loads all data in parallel

## Future Improvements

- [ ] Add export functionality (CSV/JSON)
- [ ] Implement search/filter by room name
- [ ] Add date range filtering
- [ ] Show recording status
- [ ] Real-time WebSocket updates
- [ ] Track quality metrics visualization
- [ ] Participant timeline view

## Known Limitations

- Auto-refresh stops when browser tab is in background (browser behavior)
- No real-time updates (polling only)
- No pagination controls yet (shows first 50 rooms)

## Related Documentation

- [Backend API Documentation](../../READY.md)
- [User Guide](../../docs/LIVEKIT_ROOMS_UI_GUIDE.md)
- [LiveKit Integration](../../README.md)

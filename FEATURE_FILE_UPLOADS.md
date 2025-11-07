# File Upload and Transcription Feature Implementation

## Overview
This document describes the implementation of group-based file upload and transcription functionality for Recontext.online.

## Feature Requirements
1. Create a group ("File Uploaders") with permissions to upload files
2. Implement file upload and transcription for users in this group only
3. Manage group permissions in the managing portal
4. Provide file upload interface in the user portal (visible only to authorized users)

## Implementation Status

### ✅ Completed

#### 1. Database Schema
**File**: `pkg/database/database.go`

Added two new tables:
- `uploaded_files`: Stores uploaded file metadata
  - Fields: id, filename, original_name, file_size, mime_type, storage_path, user_id, group_id, status, transcription_id, metadata, uploaded_at, processed_at
  - Indexes: user_id, group_id, status, uploaded_at

- `file_transcriptions`: Stores transcription results
  - Fields: id, file_id, text, language, confidence, duration, segments, transcribed_at, transcribed_by
  - Indexes: file_id, language

Default group created: `group-file-uploaders` with permissions:
```json
{
  "files": {"actions": ["read", "write", "delete"], "scope": "own"},
  "transcriptions": {"actions": ["read"], "scope": "own"}
}
```

#### 2. Data Models
**File**: `internal/models/uploads.go`

Created models for:
- `UploadedFile`: Main file entity
- `FileTranscription`: Transcription result
- `TranscriptionStatus`: Enum (pending, processing, completed, failed)
- Request/Response models for API endpoints

#### 3. Database Repository
**File**: `pkg/database/files_repository.go`

Implemented functions:
- `CreateUploadedFile()`: Save file upload record
- `GetUploadedFileByID()`: Retrieve file by ID
- `ListUploadedFilesByUser()`: Get paginated user files
- `UpdateFileStatus()`: Update transcription status
- `DeleteUploadedFile()`: Remove file record
- `CheckUserHasFilePermission()`: Verify user permissions

### 🚧 To Be Implemented

#### 4. API Endpoints (User Portal)
**Location**: `cmd/user-portal/main.go`

Needed endpoints:
```go
// File upload - multipart/form-data
POST /api/v1/files/upload

// List user's uploaded files
GET /api/v1/files

// Get file details
GET /api/v1/files/{id}

// Download file
GET /api/v1/files/{id}/download

// Delete file
DELETE /api/v1/files/{id}

// Get transcription
GET /api/v1/files/{id}/transcription
```

Implementation notes:
- Check user has "files:write" permission before upload
- Store files in MinIO storage
- Create transcription job in RabbitMQ queue
- Return 403 if user lacks permissions

#### 5. Managing Portal - Group Permissions UI
**Location**: `front/managing-portal/src/components/`

Required components:
- **GroupDetails.tsx**: Edit group permissions
  - Toggle switches for file upload permissions
  - JSON editor for advanced permissions

- **UserGroups.tsx**: Assign users to File Uploaders group
  - Checkbox in user management table
  - Bulk add/remove users

UI Flow:
1. Admin opens Groups page
2. Clicks on "File Uploaders" group
3. Edits permissions (already exists as JSON)
4. In Users page, admin can add users to this group

#### 6. User Portal - File Upload UI
**Location**: `front/user-portal/src/components/`

Required components:
- **FileUpload.tsx**: Upload interface
  - File drag-and-drop zone
  - Progress bar
  - File type validation (audio/video)
  - Max file size check (e.g., 500MB)

- **FilesList.tsx**: Display uploaded files
  - Table with columns: filename, size, status, uploaded date
  - Actions: download, view transcription, delete
  - Status badges (pending/processing/completed/failed)

- **FileTranscription.tsx**: Display transcription result
  - Full text display
  - Confidence score
  - Language detected
  - Download as TXT/JSON

Permission Check:
```typescript
const hasFileUploadPermission = async () => {
  const response = await fetch('/api/v1/permissions/check', {
    method: 'POST',
    body: JSON.stringify({ resource: 'files', action: 'write' })
  });
  return response.json();
};
```

Only show "Documents" tab if user has file upload permission.

## Next Steps

### Priority 1: Backend API Implementation
1. Add file upload handlers to user portal
2. Implement permission checking middleware
3. Test file upload with MinIO storage integration

### Priority 2: Managing Portal UI
1. Add file upload permission toggle in Groups page
2. Update group membership UI in Users page
3. Show which users have file upload access

### Priority 3: User Portal UI
1. Create FileUpload component
2. Create FilesList component
3. Add "Documents" tab (conditionally rendered)
4. Implement file transcription viewer

### Priority 4: Integration
1. Connect file upload to transcription worker
2. Update transcription status after processing
3. Store transcription results in database

## Testing Checklist

- [ ] Create "File Uploaders" group via managing portal
- [ ] Add test user to the group
- [ ] Login as test user in user portal
- [ ] Verify "Documents" tab is visible
- [ ] Upload audio file (MP3, WAV, M4A)
- [ ] Check file appears in files list with "pending" status
- [ ] Verify transcription worker processes the file
- [ ] Check status updates to "completed"
- [ ] View transcription result
- [ ] Download file and transcription
- [ ] Delete file
- [ ] Remove user from group
- [ ] Verify "Documents" tab disappears

## Security Considerations

1. **Permission Verification**: Always check permissions server-side, not just in UI
2. **File Validation**: Validate file type and size on backend
3. **Path Traversal**: Sanitize filenames to prevent directory traversal
4. **Ownership**: Users can only access their own files (scope: "own")
5. **Rate Limiting**: Limit upload requests per user/hour

## Storage Architecture

```
MinIO Bucket: recontext-uploads
  └── {user_id}/
      └── {filename}

Database: uploaded_files table
  - Stores metadata only
  - References MinIO path

Transcription Results: file_transcriptions table
  - Links to uploaded_files via file_id
  - Stores full transcript text
```

## API Example Requests

### Upload File
```bash
curl -X POST http://localhost:10081/api/v1/files/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@recording.mp3" \
  -F "description=Meeting notes"
```

### List Files
```bash
curl http://localhost:10081/api/v1/files?page=1&page_size=20 \
  -H "Authorization: Bearer $TOKEN"
```

### Get Transcription
```bash
curl http://localhost:10081/api/v1/files/{file_id}/transcription \
  -H "Authorization: Bearer $TOKEN"
```

## Database Queries

### Check User Permission
```sql
SELECT g.permissions
FROM groups g
INNER JOIN group_memberships gm ON g.id = gm.group_id
WHERE gm.user_id = 'user-123'
  AND g.permissions @> '{"files": {"actions": ["write"]}}';
```

### Get User's Files
```sql
SELECT * FROM uploaded_files
WHERE user_id = 'user-123'
ORDER BY uploaded_at DESC
LIMIT 20 OFFSET 0;
```

## Environment Variables

Add to docker-compose:
```yaml
UPLOAD_MAX_SIZE: "524288000"  # 500MB in bytes
ALLOWED_MIME_TYPES: "audio/mpeg,audio/wav,audio/m4a,video/mp4"
```

## Future Enhancements

1. **Batch Upload**: Upload multiple files at once
2. **Shared Files**: Allow sharing files between group members
3. **Folders**: Organize files into folders
4. **Search**: Full-text search in transcriptions
5. **Export**: Bulk export transcriptions as ZIP
6. **Analytics**: Show storage usage and transcription stats

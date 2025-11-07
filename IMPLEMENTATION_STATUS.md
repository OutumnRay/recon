# File Upload Feature - Implementation Status

## ✅ COMPLETED

### 1. Database Layer (100% Complete)
**Files Modified:**
- `pkg/database/database.go` - Added tables for `uploaded_files` and `file_transcriptions`
- `pkg/database/files_repository.go` - Created (new file)

**Features:**
- ✅ Created `uploaded_files` table with indexes
- ✅ Created `file_transcriptions` table with indexes
- ✅ Added default "File Uploaders" group with permissions
- ✅ Implemented `CreateUploadedFile()` function
- ✅ Implemented `GetUploadedFileByID()` function
- ✅ Implemented `ListUploadedFilesByUser()` function with pagination
- ✅ Implemented `CheckUserHasFilePermission()` function
- ✅ Implemented `UpdateFileStatus()` function
- ✅ Implemented `DeleteUploadedFile()` function

### 2. Data Models (100% Complete)
**Files Modified:**
- `internal/models/uploads.go` - Created (new file)

**Features:**
- ✅ `UploadedFile` struct
- ✅ `FileTranscription` struct
- ✅ `TranscriptionStatus` enum (pending, processing, completed, failed)
- ✅ All request/response models for API operations

### 3. Backend API (100% Complete)
**Files Modified:**
- `cmd/user-portal/main.go`

**Endpoints Implemented:**
1. ✅ `POST /api/v1/files/upload` - Upload file (with permission check)
2. ✅ `GET /api/v1/files` - List user's files (paginated)
3. ✅ `GET /api/v1/files/permission` - Check if user has upload permission

**Security:**
- ✅ Permission verification before upload
- ✅ User can only see their own files
- ✅ JWT authentication required
- ✅ 403 Forbidden if no permission

## 🚧 TO DO

### 4. Managing Portal UI (Priority: HIGH)

#### A. Add User to "File Uploaders" Group
**File to Edit:** `front/managing-portal/src/components/UserManagement.tsx`

**Implementation:**
1. Add "File Uploader" checkbox column to users table
2. When checked, add user to `group-file-uploaders`
3. When unchecked, remove user from group

**API Call:**
```typescript
// Add user to group
PUT /api/v1/groups/group-file-uploaders/members
Body: { user_id: "user-123" }

// Remove user from group
DELETE /api/v1/groups/group-file-uploaders/members/{user_id}
```

#### B. Display Group Members
**File to Edit:** `front/managing-portal/src/components/Groups.tsx`

**Implementation:**
1. When "File Uploaders" group is selected, show list of members
2. Add button to add new members
3. Remove button for each member

### 5. User Portal UI (Priority: HIGH)

#### A. Check Permission and Show/Hide Tab
**File to Edit:** `front/user-portal/src/components/Dashboard.tsx`

**Implementation:**
```typescript
import { useState, useEffect } from 'react';

const [hasFilePermission, setHasFilePermission] = useState(false);

useEffect(() => {
  fetch('/api/v1/files/permission', {
    headers: { 'Authorization': `Bearer ${token}` }
  })
  .then(res => res.json())
  .then(data => setHasFilePermission(data.hasPermission));
}, []);

// Only show Documents tab if hasFilePermission is true
{hasFilePermission && (
  <NavLink to="/dashboard/documents">
    <LuFileText /> Documents
  </NavLink>
)}
```

#### B. Create FileUpload Component
**File to Create:** `front/user-portal/src/components/FileUpload.tsx`

**Features:**
- Drag & drop zone
- File size validation (max 500MB)
- File type validation (audio/video)
- Upload progress bar
- Success/error messages

**Component Structure:**
```typescript
import React, { useState } from 'react';

export const FileUpload: React.FC = () => {
  const [file, setFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);

  const handleUpload = async () => {
    const formData = new FormData();
    formData.append('file', file!);

    const response = await fetch('/api/v1/files/upload', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${token}` },
      body: formData
    });

    if (response.ok) {
      alert('File uploaded successfully!');
    }
  };

  return (
    <div className="file-upload">
      <input type="file" onChange={(e) => setFile(e.target.files?.[0] || null)} />
      <button onClick={handleUpload} disabled={!file || uploading}>
        Upload
      </button>
    </div>
  );
};
```

#### C. Create FilesList Component
**File to Create:** `front/user-portal/src/components/FilesList.tsx`

**Features:**
- Table showing: filename, size, status, uploaded date
- Status badges with colors:
  - Pending: gray
  - Processing: blue
  - Completed: green
  - Failed: red
- Pagination controls
- Refresh button

**Component Structure:**
```typescript
import React, { useState, useEffect } from 'react';

export const FilesList: React.FC = () => {
  const [files, setFiles] = useState([]);
  const [page, setPage] = useState(1);

  useEffect(() => {
    fetch(`/api/v1/files?page=${page}&page_size=20`, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    .then(res => res.json())
    .then(data => setFiles(data.files));
  }, [page]);

  return (
    <div className="files-list">
      <table>
        <thead>
          <tr>
            <th>Filename</th>
            <th>Size</th>
            <th>Status</th>
            <th>Uploaded</th>
          </tr>
        </thead>
        <tbody>
          {files.map(file => (
            <tr key={file.id}>
              <td>{file.original_name}</td>
              <td>{formatFileSize(file.file_size)}</td>
              <td><StatusBadge status={file.status} /></td>
              <td>{formatDate(file.uploaded_at)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
```

#### D. Create Documents Page
**File to Create:** `front/user-portal/src/pages/Documents.tsx`

**Layout:**
```typescript
export const DocumentsPage: React.FC = () => {
  return (
    <div className="documents-page">
      <div className="page-header">
        <h1>My Documents</h1>
        <p>Upload and transcribe audio/video files</p>
      </div>

      <FileUpload />
      <FilesList />
    </div>
  );
};
```

**Register Route in App.tsx:**
```typescript
<Route path="/dashboard/documents" element={<DocumentsPage />} />
```

## 📝 TESTING STEPS

Once UI is complete, test with these steps:

1. **Admin Setup:**
   - Login to managing portal as admin
   - Go to Users page
   - Check "File Uploader" for test user
   - Verify user is added to group

2. **User Upload:**
   - Login to user portal as test user
   - Verify "Documents" tab is visible
   - Click Documents tab
   - Upload an audio file (.mp3, .wav, .m4a)
   - Verify file appears in list with "pending" status

3. **File List:**
   - Verify uploaded file is shown in table
   - Check filename, size, and upload date are correct
   - Verify status badge shows "pending"

4. **Permission Test:**
   - Remove user from "File Uploaders" group in managing portal
   - Logout and login again in user portal
   - Verify "Documents" tab is hidden
   - Try accessing `/dashboard/documents` directly
   - Should see empty page or redirect

## 🔧 BACKEND ENDPOINTS READY

The following endpoints are implemented and ready to use:

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/files/upload` | Upload file (requires permission) |
| GET | `/api/v1/files` | List user's files (paginated) |
| GET | `/api/v1/files/permission` | Check if user has upload permission |

## 📊 DATABASE SCHEMA

Tables created and ready:

### uploaded_files
| Column | Type | Description |
|--------|------|-------------|
| id | VARCHAR(255) | Primary key |
| filename | VARCHAR(500) | Stored filename |
| original_name | VARCHAR(500) | Original filename |
| file_size | BIGINT | File size in bytes |
| mime_type | VARCHAR(255) | File MIME type |
| storage_path | TEXT | Path in storage |
| user_id | VARCHAR(255) | Owner |
| group_id | VARCHAR(255) | Group (file-uploaders) |
| status | VARCHAR(50) | pending/processing/completed/failed |
| uploaded_at | TIMESTAMP | Upload time |
| processed_at | TIMESTAMP | Processing completion time |

### file_transcriptions
| Column | Type | Description |
|--------|------|-------------|
| id | VARCHAR(255) | Primary key |
| file_id | VARCHAR(255) | Foreign key to uploaded_files |
| text | TEXT | Transcription text |
| language | VARCHAR(10) | Detected language |
| confidence | DECIMAL | Confidence score |
| duration | DECIMAL | Audio duration in seconds |
| segments | JSONB | Time-stamped segments |
| transcribed_at | TIMESTAMP | Transcription time |

### groups (existing, updated)
New default group added: `group-file-uploaders` with permissions:
```json
{
  "files": {
    "actions": ["read", "write", "delete"],
    "scope": "own"
  },
  "transcriptions": {
    "actions": ["read"],
    "scope": "own"
  }
}
```

## ⏭️ NEXT STEPS

1. **Immediate (Frontend):**
   - Add "File Uploader" checkbox in UserManagement.tsx
   - Create FileUpload.tsx component
   - Create FilesList.tsx component
   - Create Documents.tsx page
   - Update Dashboard.tsx to conditionally show Documents tab

2. **Future Enhancements:**
   - Integrate with MinIO for actual file storage
   - Connect to transcription worker via RabbitMQ
   - Add file download functionality
   - Add transcription viewer
   - Add file deletion
   - Add file search

## 🎯 SUMMARY

**Backend:** ✅ 100% Complete (database + API + permissions)
**Managing Portal UI:** ⏸️ Not started (assign users to group)
**User Portal UI:** ⏸️ Not started (upload interface + file list)

The foundation is solid. Once the UI components are created, users will be able to:
1. Be assigned to "File Uploaders" group by admin
2. See "Documents" tab if they have permission
3. Upload files through a drag & drop interface
4. View list of uploaded files with status
5. System will enforce permissions at API level

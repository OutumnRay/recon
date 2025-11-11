# Avatar Upload Implementation Guide

## Overview

This document describes the implementation of avatar upload and user profile management functionality in the Recontext.online platform.

## Features Implemented

### Frontend (React/TypeScript)

1. **Profile Page** (`front/user-portal/src/pages/Profile.tsx`)
   - View and edit user profile information
   - Upload and change avatar (max 5MB)
   - Preview avatar before uploading
   - Fallback to local base64 storage if API fails
   - Responsive design for all screen sizes

2. **User Menu in Header** (`front/user-portal/src/components/Dashboard.tsx`)
   - Avatar display in header
   - Dropdown menu with profile link
   - Quick access to logout

3. **Localization**
   - English and Russian translations
   - Profile-specific messages and error handling

### Backend (Go)

1. **Models** (`internal/models/`)
   - Added profile fields to `User` struct: `FirstName`, `LastName`, `Phone`, `Bio`, `Avatar`
   - Created `UpdateProfileRequest` for self-service profile updates
   - Created `UploadAvatarResponse` for avatar upload responses

2. **Handlers** (`cmd/user-portal/handlers_profile.go`)
   - `uploadAvatarHandler`: Upload avatar images (POST `/api/v1/users/{id}/avatar`)
   - `updateProfileHandler`: Update user profile (PUT `/api/v1/users/{id}`)
   - `getProfileHandler`: Get user profile (GET `/api/v1/users/{id}`)

3. **Repository** (`pkg/database/user_repository.go`)
   - Updated `GetByID` to include new profile fields
   - Updated `Update` to handle new profile fields
   - Added `UpdateAvatar` method for avatar-only updates

4. **Database Migration** (`internal/database/migrations/000014_add_user_profile_fields.*.sql`)
   - Added profile columns to `users` table
   - Created index for full name search

## API Endpoints

### Upload Avatar
```http
POST /api/v1/users/{id}/avatar
Authorization: Bearer {token}
Content-Type: multipart/form-data

{
  "avatar": <file>
}
```

Response:
```json
{
  "avatar_url": "/uploads/avatars/abc123.jpg",
  "message": "Avatar uploaded successfully"
}
```

### Update Profile
```http
PUT /api/v1/users/{id}
Authorization: Bearer {token}
Content-Type: application/json

{
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "bio": "Software developer",
  "avatar": "https://example.com/avatar.jpg",
  "language": "en"
}
```

Response:
```json
{
  "id": "user-123",
  "username": "john",
  "email": "john@example.com",
  "role": "user",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "bio": "Software developer",
  "avatar": "https://example.com/avatar.jpg",
  "language": "en",
  "permissions": {...}
}
```

### Get Profile
```http
GET /api/v1/users/{id}
Authorization: Bearer {token}
```

## Setup and Testing

### 1. Run Database Migration

```bash
cd /Volumes/ExternalData/source/Team21/Recontext.online

# Make sure PostgreSQL is running
# The migration will run automatically on application startup
```

### 2. Create Uploads Directory

```bash
mkdir -p uploads/avatars
chmod 755 uploads/avatars
```

### 3. Build and Run Backend

```bash
cd cmd/user-portal
go build -o user-portal
./user-portal
```

The server will start on `http://localhost:8081`

### 4. Build and Run Frontend

```bash
cd front/user-portal
npm install
npm run dev
```

The frontend will be available on `http://localhost:5173`

### 5. Test the Features

1. **Login** to the application
2. **Click on your avatar** in the header (top right)
3. **Select "Profile"** from the dropdown menu
4. **Click "Edit"** button
5. **Upload an avatar**:
   - Click on the avatar placeholder
   - Select an image file (JPG, PNG, GIF, max 5MB)
   - See the preview
6. **Edit profile information**:
   - First Name, Last Name
   - Phone number
   - Bio/description
7. **Click "Save"**
8. **Check the header** - your avatar should now appear

## File Storage

### Local Storage (Default)
Avatars are stored in: `uploads/avatars/`

Each file is named using MD5 hash of `{userID}-{timestamp}` plus the original extension.

Example: `uploads/avatars/abc123def456.jpg`

### Fallback Mechanism
If the avatar upload API fails, the system will:
1. Use the base64-encoded image from the preview
2. Store it in localStorage/sessionStorage
3. Display it in the UI
4. Show a warning message to the user

## Security Considerations

1. **File Size Limit**: 5MB maximum
2. **File Type Validation**: Only image files allowed
3. **User Authorization**: Users can only update their own profile (except admins)
4. **Path Traversal Protection**: Filenames are generated, not user-provided
5. **Content-Type Validation**: Server checks actual file content type

## Database Schema

```sql
ALTER TABLE users ADD COLUMN first_name VARCHAR(100);
ALTER TABLE users ADD COLUMN last_name VARCHAR(100);
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
ALTER TABLE users ADD COLUMN bio TEXT;
ALTER TABLE users ADD COLUMN avatar TEXT;

CREATE INDEX idx_users_full_name ON users (first_name, last_name);
```

## Troubleshooting

### Avatar Upload Fails
- Check that `uploads/avatars` directory exists and is writable
- Verify file size is under 5MB
- Check file type is an image
- Look at browser console and server logs for error messages

### Profile Update Fails
- Verify JWT token is valid
- Check user has permission to update the profile
- Verify database connection
- Check server logs for SQL errors

### Avatar Not Displaying
- Check the avatar URL in the database
- Verify the file exists in `uploads/avatars/`
- Check `/uploads/` route is correctly configured
- Clear browser cache

## Future Enhancements

1. **MinIO/S3 Integration**: Store avatars in object storage instead of local filesystem
2. **Image Resizing**: Automatically resize uploaded images to standard sizes
3. **Image Cropping**: Allow users to crop images before upload
4. **Multiple Image Formats**: Support WebP for better compression
5. **CDN Integration**: Serve avatars through CDN for better performance
6. **Avatar History**: Keep history of previous avatars
7. **Default Avatars**: Generate default avatars based on initials or identicons

## Testing Checklist

- [ ] Avatar upload with valid image file
- [ ] Avatar upload with invalid file type (should fail)
- [ ] Avatar upload with file > 5MB (should fail)
- [ ] Profile update with all fields
- [ ] Profile update with partial fields
- [ ] Avatar display in header menu
- [ ] Avatar display on profile page
- [ ] Navigation to profile from header menu
- [ ] Profile edit mode toggle
- [ ] Profile cancel button
- [ ] Profile save button
- [ ] Avatar removal
- [ ] Responsive design on mobile
- [ ] Localization (EN/RU)
- [ ] Authentication (only own profile)
- [ ] Database migration
- [ ] API error handling
- [ ] Fallback to base64 when API fails

# Task: Fix and add API handlers for auth, profile, videos, transcription

## Plan Breakdown (Approved)

### Step 1: Create TODO.md ✅ (done)

### Step 2: Fix /auth/login & profile responses - remove forbidden fields ✅
- Files: cmd/user-portal/main.go (login), handlers_profile.go (get/update profile)
- Removed from UserInfo: Avatar, DepartmentID, Phone, Language, NotificationPreferences (Permissions kept minimal)


### Step 3: Add registration handler /api/v1/auth/register POST ✅
- File: cmd/user-portal/main.go
- Added registerHandler + route (public), uses RegisterRequest, userRepo.Create, HashPassword, returns LoginResponse minimal

### Step 4: Add /profile GET (current user data) ✅
### Step 5: Add /update-profile PUT (bio etc.) ✅
- Files: handlers_profile_minimal.go + main.go routes

### Step 6: Add /videos GET (list loaded videos)
- File: cmd/user-portal/main.go
- New listVideosHandler: combine listFilesHandler + meeting recordings? Or use existing /files or /recordings as base. Paginated list of user videos/files.

### Step 7: Add /videos/{videoId} GET (transcription decode)
- File: cmd/user-portal/main.go or handlers_recordings.go
- New getVideoTranscriptHandler: videoId=roomSid/trackSid?, fetch room/track transcripts via DB/liveKitRepo, merge phrases, return with summary/memo.

### Step 8: Add routes for new handlers in setupRoutes()
- File: cmd/user-portal/main.go

### Step 9: Add swaggo docs for new endpoints

### Step 10: Test changes
- Build: go build ./cmd/user-portal
- Run endpoints: curl login/register/profile/update-profile/videos etc.
- Verify no regressions.

### Step 11: attempt_completion

**Current Progress:** Ready for Step 2.

**Next Action:** Confirm TODO.md created, then proceed to read models for UserInfo exact fields if needed, or start edits.


package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOClient manages connection to MinIO/S3
type MinIOClient struct {
	client *minio.Client
	bucket string
}

// NewMinIOClient creates a new MinIO client
func NewMinIOClient() (*MinIOClient, error) {
	endpoint := getEnv("MINIO_ENDPOINT", "localhost:9000")
	accessKey := getEnv("MINIO_ACCESS_KEY", "minioadmin")
	secretKey := getEnv("MINIO_SECRET_KEY", "minioadmin")
	bucket := getEnv("MINIO_BUCKET", "recontext")
	useSSL := getEnv("MINIO_USE_SSL", "false") == "true"

	// Remove protocol prefix if present
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Check if bucket exists, create if not
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		// Create bucket
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket %s: %w", bucket, err)
		}
		fmt.Printf("✅ Created MinIO bucket: %s\n", bucket)
	} else {
		fmt.Printf("✅ MinIO bucket exists: %s\n", bucket)
	}

	return &MinIOClient{
		client: client,
		bucket: bucket,
	}, nil
}

// GetPlaylist godoc
// @Summary Get HLS playlist for recording
// @Description Returns m3u8 playlist with URLs rewritten to proxy through this server
// @Tags Recordings
// @Produce text/plain
// @Param egress_id path string true "Egress ID"
// @Success 200 {string} string "M3U8 playlist"
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/recordings/{egress_id}/playlist [get]
func (up *UserPortal) getPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	up.logger.Infof("📹 [PLAYLIST] Request: %s", r.URL.Path)

	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusOK)
		return
	}

	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.logger.Errorf("📹 [PLAYLIST] Unauthorized - no user in context")
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	up.logger.Infof("📹 [PLAYLIST] User: %s", claims.UserID)

	// Extract egress_id or track id from URL path
	// Supports both /recordings/{egress_id}/playlist and /recordings/track/{track_sid}/playlist
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/recordings/"), "/")
	if len(pathParts) < 2 {
		up.respondWithError(w, http.StatusBadRequest, "Invalid URL", "")
		return
	}

	var egressID string
	var trackSID string
	var meetingID string
	var isTrack bool

	var roomSIDForAccess string

	// Check if this is a track request
	if pathParts[0] == "track" && len(pathParts) >= 3 {
		isTrack = true
		trackSID = pathParts[1]

		// Find track by SID to get room info
		var track models.Track
		err := up.db.DB.Where("sid = ?", trackSID).First(&track).Error
		if err != nil {
			up.logger.Errorf("Track not found: %s, error: %v", trackSID, err)
			up.respondWithError(w, http.StatusNotFound, "Track not found", err.Error())
			return
		}

		// Get room to find meeting ID from room.MeetingID
		var room models.Room
		err = up.db.DB.Where("sid = ?", track.RoomSID).First(&room).Error
		if err != nil {
			up.logger.Errorf("Room not found for track %s: %v", trackSID, err)
			up.respondWithError(w, http.StatusNotFound, "Room not found", err.Error())
			return
		}
		if room.MeetingID == nil {
			up.logger.Errorf("Room %s has no MeetingID", room.SID)
			up.respondWithError(w, http.StatusNotFound, "Meeting not found for room", "")
			return
		}
		meetingID = room.MeetingID.String()
		roomSIDForAccess = room.SID
		egressID = room.EgressID
	} else {
		// For composite video, egressID is actually room SID
		roomSIDForAccess = pathParts[0]
		egressID = pathParts[0]
	}

	// Check access permissions using room SID
	if !up.checkRecordingAccess(roomSIDForAccess, claims.UserID) {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "")
		return
	}

	// Initialize MinIO client
	minioClient, err := NewMinIOClient()
	if err != nil {
		up.logger.Errorf("Failed to create MinIO client: %v", err)
		up.respondWithError(w, http.StatusInternalServerError, "Storage error", err.Error())
		return
	}

	// Determine playlist path based on type
	var playlistPath string
	var roomSID string
	if isTrack {
		// For tracks, we already have meetingID from earlier. Get room SID.
		// First find the track to get room_sid
		var track models.Track
		err := up.db.DB.Where("sid = ?", trackSID).First(&track).Error
		if err != nil {
			up.logger.Errorf("Track not found: %s, error: %v", trackSID, err)
			up.respondWithError(w, http.StatusNotFound, "Track not found", err.Error())
			return
		}

		// Then get the room using the room_sid from track
		var room models.Room
		err = up.db.DB.Where("sid = ?", track.RoomSID).First(&room).Error
		if err != nil {
			up.logger.Errorf("Room not found for track %s: %v", trackSID, err)
			up.respondWithError(w, http.StatusNotFound, "Room not found", err.Error())
			return
		}
		roomSID = room.SID
		playlistPath = fmt.Sprintf("%s_%s/tracks/%s.m3u8", meetingID, roomSID, trackSID)
	} else {
		// For room composites, get room and use meetingID/roomSID/composite.m3u8
		// egressID can be either egress_id or room SID (when using /api/v1/recordings/{room_sid}/playlist)
		var room models.Room

		// Try to find room by SID first (most common case for composite video)
		err := up.db.DB.Where("sid = ?", egressID).First(&room).Error
		if err != nil {
			// If not found by SID, try egress_id (legacy compatibility)
			err = up.db.DB.Where("egress_id = ?", egressID).First(&room).Error
			if err != nil {
				up.logger.Errorf("Room not found for identifier %s: %v", egressID, err)
				up.respondWithError(w, http.StatusNotFound, "Room not found", err.Error())
				return
			}
		}

		if room.MeetingID == nil {
			up.logger.Errorf("Room %s has no MeetingID", room.SID)
			up.respondWithError(w, http.StatusNotFound, "Meeting not found for room", "")
			return
		}
		meetingID = room.MeetingID.String()
		roomSID = room.SID
		egressID = room.EgressID // Update egressID for access check
		playlistPath = fmt.Sprintf("%s_%s/composite.m3u8", meetingID, roomSID)
	}

	// Get the playlist file from MinIO
	up.logger.Infof("📹 [PLAYLIST] Fetching from MinIO: bucket=%s, path=%s", minioClient.bucket, playlistPath)
	object, err := minioClient.client.GetObject(context.Background(), minioClient.bucket, playlistPath, minio.GetObjectOptions{})
	if err != nil {
		up.logger.Errorf("📹 [PLAYLIST] Failed to get from MinIO (path: %s): %v", playlistPath, err)
		up.respondWithError(w, http.StatusInternalServerError, "Failed to read playlist", err.Error())
		return
	}
	defer object.Close()

	// Try to read object info to check if it actually exists
	_, err = object.Stat()
	if err != nil {
		up.logger.Errorf("📹 [PLAYLIST] Playlist file not found in MinIO (path: %s): %v", playlistPath, err)
		up.respondWithError(w, http.StatusNotFound, "Recording not available", "The recording may still be processing or does not exist")
		return
	}

	up.logger.Infof("📹 [PLAYLIST] Successfully fetched from MinIO")

	// Get token from query parameter (if provided) to propagate to segment URLs
	token := r.URL.Query().Get("token")

	// Read and rewrite playlist URLs
	scanner := bufio.NewScanner(object)
	var modifiedPlaylist strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// If line is a segment reference (ends with .ts), rewrite it to proxy URL
		if strings.HasSuffix(line, ".ts") && !strings.HasPrefix(line, "http") && !strings.HasPrefix(line, "#") {
			// Extract just the filename
			filename := line
			if idx := strings.LastIndex(line, "/"); idx != -1 {
				filename = line[idx+1:]
			}
			// Rewrite to proxy URL - use track SID if this is a track request
			if isTrack {
				line = fmt.Sprintf("/api/v1/recordings/track/%s/segment/%s", trackSID, filename)
			} else {
				// For composite video, use room SID (not egress_id)
				line = fmt.Sprintf("/api/v1/recordings/%s/segment/%s", roomSID, filename)
			}
			// Add token to segment URL if it was provided
			if token != "" {
				line = fmt.Sprintf("%s?token=%s", line, token)
			}
		}

		modifiedPlaylist.WriteString(line)
		modifiedPlaylist.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		up.logger.Errorf("Failed to read playlist: %v", err)
		up.respondWithError(w, http.StatusInternalServerError, "Failed to read playlist", err.Error())
		return
	}

	// Return modified playlist
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	w.Header().Set("Access-Control-Max-Age", "3600")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(modifiedPlaylist.String()))
	up.logger.Infof("📹 [PLAYLIST] Sent modified playlist to client (%d bytes)", len(modifiedPlaylist.String()))
}

// GetSegment godoc
// @Summary Get video/audio segment for recording
// @Description Returns TS segment file from storage
// @Tags Recordings
// @Produce video/mp2t
// @Param egress_id path string true "Egress ID"
// @Param filename path string true "Segment filename"
// @Success 200 {file} binary "TS segment"
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/recordings/{egress_id}/segment/{filename} [get]
func (up *UserPortal) getSegmentHandler(w http.ResponseWriter, r *http.Request) {
	up.logger.Infof("📹 [SEGMENT] Request: %s", r.URL.Path)

	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusOK)
		return
	}

	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.logger.Errorf("📹 [SEGMENT] Unauthorized - no user in context")
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Extract egress_id and filename from URL path
	// Supports both /recordings/{egress_id}/segment/{filename} and /recordings/track/{egress_id}/segment/{filename}
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/recordings/"), "/")
	if len(pathParts) < 3 {
		up.respondWithError(w, http.StatusBadRequest, "Invalid URL", "")
		return
	}

	var egressID string
	var trackSID string
	var meetingID string
	var filename string
	var isTrack bool

	// Check if this is a track request
	if pathParts[0] == "track" && len(pathParts) >= 4 {
		isTrack = true
		trackSID = pathParts[1]
		filename = pathParts[3]

		// Find track by SID to get egress_id for access check and meeting_id for MinIO path
		var track models.Track
		err := up.db.DB.Where("sid = ?", trackSID).First(&track).Error
		if err != nil {
			up.logger.Errorf("Track not found: %s, error: %v", trackSID, err)
			up.respondWithError(w, http.StatusNotFound, "Track not found", err.Error())
			return
		}
		egressID = track.EgressID

		// Get room to find meeting ID from room.MeetingID
		var room models.Room
		err = up.db.DB.Where("sid = ?", track.RoomSID).First(&room).Error
		if err != nil {
			up.logger.Errorf("Room not found for track %s: %v", trackSID, err)
			up.respondWithError(w, http.StatusNotFound, "Room not found", err.Error())
			return
		}
		if room.MeetingID == nil {
			up.logger.Errorf("Room %s has no MeetingID", room.SID)
			up.respondWithError(w, http.StatusNotFound, "Meeting not found for room", "")
			return
		}
		meetingID = room.MeetingID.String()
	} else {
		egressID = pathParts[0]
		filename = pathParts[2]
	}

	// Validate filename (prevent directory traversal)
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		up.respondWithError(w, http.StatusBadRequest, "Invalid filename", "")
		return
	}

	// Check access permissions
	if !up.checkRecordingAccess(egressID, claims.UserID) {
		up.respondWithError(w, http.StatusForbidden, "Access denied", "")
		return
	}

	// Initialize MinIO client
	minioClient, err := NewMinIOClient()
	if err != nil {
		up.logger.Errorf("Failed to create MinIO client: %v", err)
		up.respondWithError(w, http.StatusInternalServerError, "Storage error", err.Error())
		return
	}

	// Construct segment path based on type
	var segmentPath string
	if isTrack {
		// For tracks, we already have meetingID from earlier. Get room SID.
		// First find the track to get room_sid
		var track models.Track
		err := up.db.DB.Where("sid = ?", trackSID).First(&track).Error
		if err != nil {
			up.logger.Errorf("Track not found: %s, error: %v", trackSID, err)
			up.respondWithError(w, http.StatusNotFound, "Track not found", err.Error())
			return
		}

		// Then get the room using the room_sid from track
		var room models.Room
		err = up.db.DB.Where("sid = ?", track.RoomSID).First(&room).Error
		if err != nil {
			up.logger.Errorf("Room not found for track %s: %v", trackSID, err)
			up.respondWithError(w, http.StatusNotFound, "Room not found", err.Error())
			return
		}
		segmentPath = fmt.Sprintf("%s_%s/tracks/%s", meetingID, room.SID, filename)
	} else {
		// For room composites, egressID is actually room_sid in the URL
		// Files are stored in: {meetingID}_{roomSID}/composite_XXXXX.ts
		var room models.Room
		err := up.db.DB.Where("sid = ?", egressID).First(&room).Error
		if err != nil {
			up.logger.Errorf("Room not found for SID %s: %v", egressID, err)
			up.respondWithError(w, http.StatusNotFound, "Room not found", err.Error())
			return
		}
		if room.MeetingID == nil {
			up.logger.Errorf("Room %s has no MeetingID", room.SID)
			up.respondWithError(w, http.StatusNotFound, "Meeting not found for room", "")
			return
		}
		meetingID = room.MeetingID.String()
		segmentPath = fmt.Sprintf("%s_%s/%s", meetingID, room.SID, filename)
	}

	up.logger.Infof("📹 [SEGMENT] Fetching from MinIO: bucket=%s, path=%s", minioClient.bucket, segmentPath)
	object, err := minioClient.client.GetObject(context.Background(), minioClient.bucket, segmentPath, minio.GetObjectOptions{})
	if err != nil {
		up.logger.Errorf("📹 [SEGMENT] Failed to get from MinIO (path: %s): %v", segmentPath, err)
		up.respondWithError(w, http.StatusNotFound, "Segment not found", err.Error())
		return
	}
	defer object.Close()

	// Get object info for content length
	objInfo, err := object.Stat()
	if err != nil {
		up.logger.Errorf("📹 [SEGMENT] Failed to get object info: %v", err)
	} else {
		up.logger.Infof("📹 [SEGMENT] Successfully fetched from MinIO (size: %d bytes)", objInfo.Size)
	}

	// Stream the segment to client
	w.Header().Set("Content-Type", "video/mp2t")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	w.Header().Set("Access-Control-Max-Age", "3600")
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache segments for 1 year

	if err == nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", objInfo.Size))
	}

	w.WriteHeader(http.StatusOK)

	// Stream the data
	bytesWritten, err := io.Copy(w, object)
	if err != nil {
		up.logger.Errorf("📹 [SEGMENT] Failed to stream segment: %v", err)
	} else {
		up.logger.Infof("📹 [SEGMENT] Streamed %d bytes to client", bytesWritten)
	}
}

// checkRecordingAccess checks if user has access to a recording by room or egress ID
func (up *UserPortal) checkRecordingAccess(roomSIDOrEgressID string, userID uuid.UUID) bool {
	var room models.Room

	// Try to find room by SID first (for composite video)
	err := up.db.DB.Where("sid = ?", roomSIDOrEgressID).First(&room).Error
	if err != nil {
		// If not found by SID, try by egress_id (for legacy compatibility)
		err = up.db.DB.Where("egress_id = ?", roomSIDOrEgressID).First(&room).Error
		if err != nil {
			// Room not found, access denied
			up.logger.Errorf("Room not found for identifier %s: %v", roomSIDOrEgressID, err)
			return false
		}
	}

	// Check meeting access
	if room.MeetingID == nil {
		// No meeting ID, access denied
		up.logger.Errorf("Room %s has no meeting ID", room.SID)
		return false
	}

	meeting, err := up.meetingRepo.GetMeetingByID(*room.MeetingID)
	if err != nil {
		// Meeting not found, access denied
		up.logger.Errorf("Meeting not found for room %s: %v", room.SID, err)
		return false
	}

	// Check if user is the creator
	if meeting.CreatedBy == userID {
		return true
	}

	// Check if user is a participant
	participants, err := up.meetingRepo.GetMeetingParticipants(*room.MeetingID)
	if err != nil {
		// Failed to get participants, access denied
		up.logger.Errorf("Failed to get participants for meeting %s: %v", room.MeetingID.String(), err)
		return false
	}

	for _, p := range participants {
		if p.UserID == userID {
			return true
		}
	}

	// User is neither creator nor participant, access denied
	up.logger.Errorf("User %s has no access to room %s", userID.String(), room.SID)
	return false
}

// DeleteDirectory удаляет всю директорию (все объекты с указанным префиксом) из MinIO
func (mc *MinIOClient) DeleteDirectory(ctx context.Context, prefix string) (int, error) {
	// Ensure prefix ends with / for directory matching
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	// List all objects with the prefix
	objectsCh := mc.client.ListObjects(ctx, mc.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	deletedCount := 0
	var lastErr error

	// Delete each object
	for object := range objectsCh {
		if object.Err != nil {
			lastErr = object.Err
			continue
		}

		err := mc.client.RemoveObject(ctx, mc.bucket, object.Key, minio.RemoveObjectOptions{})
		if err != nil {
			lastErr = err
			continue
		}

		deletedCount++
	}

	if lastErr != nil && deletedCount == 0 {
		return 0, fmt.Errorf("failed to delete directory: %w", lastErr)
	}

	return deletedCount, nil
}

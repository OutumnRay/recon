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
	endpoint := getEnv("S3_ENDPOINT", "localhost:9000")
	accessKey := getEnv("S3_ACCESS_KEY", "minioadmin")
	secretKey := getEnv("S3_SECRET_KEY", "minioadmin")
	bucket := getEnv("S3_BUCKET", "livekit-recordings")
	useSSL := getEnv("S3_USE_SSL", "false") == "true"

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
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Extract egress_id or track id from URL path
	// Supports both /recordings/{egress_id}/playlist and /recordings/track/{track_sid}/playlist
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/recordings/"), "/")
	if len(pathParts) < 2 {
		up.respondWithError(w, http.StatusBadRequest, "Invalid URL", "")
		return
	}

	var egressID string
	var trackSID string
	var roomSID string
	var isTrack bool

	// Check if this is a track request
	if pathParts[0] == "track" && len(pathParts) >= 3 {
		isTrack = true
		trackSID = pathParts[1]

		// Find track by SID to get egress_id for access check and room_sid for MinIO path
		var track models.Track
		err := up.db.DB.Where("sid = ?", trackSID).First(&track).Error
		if err != nil {
			up.logger.Errorf("Track not found: %s, error: %v", trackSID, err)
			up.respondWithError(w, http.StatusNotFound, "Track not found", err.Error())
			return
		}
		egressID = track.EgressID
		roomSID = track.RoomSID
	} else {
		egressID = pathParts[0]
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

	// Determine playlist path based on type
	var playlistPath string
	if isTrack {
		// For tracks, use room/tracks/track_sid.m3u8 path on MinIO
		playlistPath = fmt.Sprintf("%s/tracks/%s.m3u8", roomSID, trackSID)
	} else {
		// For room composites, get room SID and use room/composite.m3u8
		var room models.Room
		err := up.db.DB.Where("egress_id = ?", egressID).First(&room).Error
		if err != nil {
			up.logger.Errorf("Room not found for egress %s: %v", egressID, err)
			up.respondWithError(w, http.StatusNotFound, "Room not found", err.Error())
			return
		}
		playlistPath = fmt.Sprintf("%s/composite.m3u8", room.SID)
	}

	// Get the playlist file from MinIO
	object, err := minioClient.client.GetObject(context.Background(), minioClient.bucket, playlistPath, minio.GetObjectOptions{})
	if err != nil {
		up.logger.Errorf("Failed to get playlist from MinIO (path: %s): %v", playlistPath, err)
		up.respondWithError(w, http.StatusNotFound, "Playlist not found", err.Error())
		return
	}
	defer object.Close()

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
				line = fmt.Sprintf("/api/v1/recordings/%s/segment/%s", egressID, filename)
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
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(modifiedPlaylist.String()))
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
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
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
	var roomSID string
	var filename string
	var isTrack bool

	// Check if this is a track request
	if pathParts[0] == "track" && len(pathParts) >= 4 {
		isTrack = true
		trackSID = pathParts[1]
		filename = pathParts[3]

		// Find track by SID to get egress_id for access check and room_sid for MinIO path
		var track models.Track
		err := up.db.DB.Where("sid = ?", trackSID).First(&track).Error
		if err != nil {
			up.logger.Errorf("Track not found: %s, error: %v", trackSID, err)
			up.respondWithError(w, http.StatusNotFound, "Track not found", err.Error())
			return
		}
		egressID = track.EgressID
		roomSID = track.RoomSID
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
		// For tracks, use room/tracks/filename path on MinIO (filename already includes track SID prefix)
		segmentPath = fmt.Sprintf("%s/tracks/%s", roomSID, filename)
	} else {
		// For room composites, get room SID and use room/composite_XXXXX.ts
		var room models.Room
		err := up.db.DB.Where("egress_id = ?", egressID).First(&room).Error
		if err != nil {
			up.logger.Errorf("Room not found for egress %s: %v", egressID, err)
			up.respondWithError(w, http.StatusNotFound, "Room not found", err.Error())
			return
		}
		segmentPath = fmt.Sprintf("%s/%s", room.SID, filename)
	}

	object, err := minioClient.client.GetObject(context.Background(), minioClient.bucket, segmentPath, minio.GetObjectOptions{})
	if err != nil {
		up.logger.Errorf("Failed to get segment from MinIO (path: %s): %v", segmentPath, err)
		up.respondWithError(w, http.StatusNotFound, "Segment not found", err.Error())
		return
	}
	defer object.Close()

	// Stream the segment to client
	w.Header().Set("Content-Type", "video/mp2t")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache segments for 1 year

	// Get object info for content length
	objInfo, err := object.Stat()
	if err == nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", objInfo.Size))
	}

	w.WriteHeader(http.StatusOK)

	// Stream the data
	if _, err := io.Copy(w, object); err != nil {
		up.logger.Errorf("Failed to stream segment: %v", err)
	}
}

// checkRecordingAccess checks if user has access to a recording
func (up *UserPortal) checkRecordingAccess(egressID string, userID uuid.UUID) bool {
	// Find the room or track associated with this egress
	// First try to find in rooms
	var room models.Room
	err := up.db.DB.Where("egress_id = ?", egressID).First(&room).Error
	if err == nil {
		// Found room, now check if user is participant in the meeting
		// Get all rooms with this name (meeting ID)
		rooms, err := up.liveKitRepo.GetRoomsByName(room.Name)
		if err == nil && len(rooms) > 0 {
			// room.Name should be the meeting ID
			meetingID, err := uuid.Parse(room.Name)
			if err == nil {
				// Check if user is participant or creator
				meeting, err := up.meetingRepo.GetMeetingByID(meetingID)
				if err == nil {
					if meeting.CreatedBy == userID {
						return true
					}

					participants, err := up.meetingRepo.GetMeetingParticipants(meetingID)
					if err == nil {
						for _, p := range participants {
							if p.UserID == userID {
								return true
							}
						}
					}
				}
			}
		}
		return false
	}

	// Try to find in tracks
	var track models.Track
	err = up.db.DB.Where("egress_id = ?", egressID).First(&track).Error
	if err == nil {
		// Found track, get its room
		err = up.db.DB.Where("sid = ?", track.RoomSID).First(&room).Error
		if err == nil {
			// Check meeting access
			meetingID, err := uuid.Parse(room.Name)
			if err == nil {
				meeting, err := up.meetingRepo.GetMeetingByID(meetingID)
				if err == nil {
					if meeting.CreatedBy == userID {
						return true
					}

					participants, err := up.meetingRepo.GetMeetingParticipants(meetingID)
					if err == nil {
						for _, p := range participants {
							if p.UserID == userID {
								return true
							}
						}
					}
				}
			}
		}
	}

	return false
}

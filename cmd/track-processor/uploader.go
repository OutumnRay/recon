package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"Recontext.online/pkg/database"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TrackUploader struct {
	minioClient *minio.Client
	db          *database.DB
	bucketName  string
	meetingID   string
	roomSID     string
}

func NewTrackUploader(meetingID, roomSID string) (*TrackUploader, error) {
	// MinIO credentials
	endpoint := "api.storage.recontext.online"
	accessKeyID := "minioadmin"
	secretAccessKey := "32a4953d5bff4a1c6aea4d4ccfb757e5"
	useSSL := true

	// Initialize MinIO client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Initialize database connection
	dsn := "host=192.168.5.153 port=5432 user=recontext password=recontext dbname=recontext sslmode=disable"
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db := &database.DB{DB: gormDB}

	return &TrackUploader{
		minioClient: minioClient,
		db:          db,
		bucketName:  "recontext",
		meetingID:   meetingID,
		roomSID:     roomSID,
	}, nil
}

func (tu *TrackUploader) UploadTrack(trackID, localFilePath string) error {
	ctx := context.Background()

	// Construct MinIO path
	objectName := fmt.Sprintf("%s_%s/tracks/%s.mp4", tu.meetingID, tu.roomSID, trackID)

	log.Printf("📤 Uploading %s to MinIO: %s", trackID, objectName)

	// Upload file
	info, err := tu.minioClient.FPutObject(ctx, tu.bucketName, objectName, localFilePath, minio.PutObjectOptions{
		ContentType: "video/mp4",
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	log.Printf("✅ Uploaded %s (size: %d bytes)", trackID, info.Size)

	// Construct playlist URL (for HLS streaming via MinIO/CDN)
	playlistURL := fmt.Sprintf("https://api.storage.recontext.online/%s/%s", tu.bucketName, objectName)

	// Update database - find track by SID
	log.Printf("🔍 Looking for track %s in database", trackID)

	// Проверяем существует ли трек в базе
	var track struct {
		ID             string
		ParticipantSID string
		Type           string
	}

	err = tu.db.DB.Table("livekit_tracks").
		Select("id, participant_sid, type").
		Where("sid = ?", trackID).
		Where("room_sid = ?", tu.roomSID).
		First(&track).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("⚠️ Track %s not found in database, skipping database update", trackID)
			return nil
		}
		return fmt.Errorf("failed to query track: %w", err)
	}

	log.Printf("📝 Found track in database: ID=%s, Type=%s, ParticipantSID=%s", track.ID, track.Type, track.ParticipantSID)

	// Update or create egress_recordings entry using GORM
	// Сначала пробуем найти существующую запись
	var existingRecording struct {
		ID string
	}

	err = tu.db.DB.Table("livekit_egress_recordings").
		Select("id").
		Where("track_sid = ?", trackID).
		Where("room_sid = ?", tu.roomSID).
		First(&existingRecording).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing recording: %w", err)
	}

	if err == gorm.ErrRecordNotFound {
		// Создаем новую запись
		log.Printf("💾 Creating new egress_recordings entry for track %s", trackID)

		meetingUUID, _ := uuid.Parse(tu.meetingID)

		recording := map[string]interface{}{
			"id":              uuid.New().String(),
			"room_sid":        tu.roomSID,
			"meeting_id":      meetingUUID,
			"track_sid":       trackID,
			"participant_sid": track.ParticipantSID,
			"type":            "track_composite",
			"status":          "ended",
			"file_path":       objectName,
			"playlist_url":    playlistURL,
			"audio_only":      track.Type == "audio",
		}

		if err := tu.db.DB.Table("livekit_egress_recordings").Create(recording).Error; err != nil {
			return fmt.Errorf("failed to create recording: %w", err)
		}
	} else {
		// Обновляем существующую запись
		log.Printf("📝 Updating existing egress_recordings entry for track %s", trackID)

		updates := map[string]interface{}{
			"status":       "ended",
			"file_path":    objectName,
			"playlist_url": playlistURL,
		}

		err = tu.db.DB.Table("livekit_egress_recordings").
			Where("track_sid = ?", trackID).
			Where("room_sid = ?", tu.roomSID).
			Updates(updates).Error

		if err != nil {
			return fmt.Errorf("failed to update recording: %w", err)
		}
	}

	log.Printf("✅ Database updated for track %s", trackID)
	return nil
}

func main() {
	meetingID := "082494f4-e0a7-43c0-a19d-6d6e35ef429c"
	roomSID := "RM_4qJSgKmcQmPW"

	uploader, err := NewTrackUploader(meetingID, roomSID)
	if err != nil {
		log.Fatalf("Failed to create uploader: %v", err)
	}

	// Upload video track
	videoTrackID := "TR_VCoyexbMM3nSRm"
	videoFile := filepath.Join("downloads", videoTrackID+".m3u8", videoTrackID+".mp4")
	if err := uploader.UploadTrack(videoTrackID, videoFile); err != nil {
		log.Printf("❌ Failed to upload video track: %v", err)
	}

	// Upload audio track
	audioTrackID := "TR_AMcJtxis6SfUzT"
	audioFile := filepath.Join("downloads", audioTrackID+".m3u8", audioTrackID+".mp4")
	if err := uploader.UploadTrack(audioTrackID, audioFile); err != nil {
		log.Printf("❌ Failed to upload audio track: %v", err)
	}

	log.Printf("\n✅ All tracks uploaded successfully!")
}

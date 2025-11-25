package main

import (
	"log"
	"time"

	"Recontext.online/pkg/database"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=192.168.5.153 port=5432 user=recontext password=recontext dbname=recontext sslmode=disable"
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db := &database.DB{DB: gormDB}

	meetingID := "082494f4-e0a7-43c0-a19d-6d6e35ef429c"
	meetingUUID, _ := uuid.Parse(meetingID)
	roomSID := "RM_4qJSgKmcQmPW"
	roomName := roomSID // или meetingID, зависит от вашей логики

	log.Printf("📝 Creating database records for meeting %s, room %s", meetingID, roomSID)

	// Create room record if doesn't exist
	var roomExists int64
	db.DB.Table("livekit_rooms").Where("sid = ?", roomSID).Count(&roomExists)

	if roomExists == 0 {
		log.Printf("💾 Creating room record")
		room := map[string]interface{}{
			"id":          uuid.New().String(),
			"sid":         roomSID,
			"name":        roomName,
			"meeting_id":  meetingUUID,
			"created_at":  time.Now(),
			"finished_at": time.Now(),
		}
		if err := db.DB.Table("livekit_rooms").Create(room).Error; err != nil {
			log.Printf("❌ Failed to create room: %v", err)
		} else {
			log.Printf("✅ Room record created")
		}
	}

	// Create egress recordings for uploaded tracks
	tracks := []struct {
		EgressID string
		TrackSID string
		Type     string
		FilePath string
		URL      string
	}{
		{
			EgressID: "EG_" + uuid.New().String()[:12],
			TrackSID: "TR_VCoyexbMM3nSRm",
			Type:     "track_composite",
			FilePath: "082494f4-e0a7-43c0-a19d-6d6e35ef429c_RM_4qJSgKmcQmPW/tracks/TR_VCoyexbMM3nSRm.mp4",
			URL:      "https://api.storage.recontext.online/recontext/082494f4-e0a7-43c0-a19d-6d6e35ef429c_RM_4qJSgKmcQmPW/tracks/TR_VCoyexbMM3nSRm.mp4",
		},
		{
			EgressID: "EG_" + uuid.New().String()[:12],
			TrackSID: "TR_AMcJtxis6SfUzT",
			Type:     "track_composite",
			FilePath: "082494f4-e0a7-43c0-a19d-6d6e35ef429c_RM_4qJSgKmcQmPW/tracks/TR_AMcJtxis6SfUzT.mp4",
			URL:      "https://api.storage.recontext.online/recontext/082494f4-e0a7-43c0-a19d-6d6e35ef429c_RM_4qJSgKmcQmPW/tracks/TR_AMcJtxis6SfUzT.mp4",
		},
	}

	for _, track := range tracks {
		log.Printf("\n📝 Processing track: %s (EgressID: %s)", track.TrackSID, track.EgressID)

		// Check if recording already exists
		var exists int64
		db.DB.Table("livekit_egress_recordings").
			Where("track_sid = ? AND room_sid = ?", track.TrackSID, roomSID).
			Count(&exists)

		if exists > 0 {
			log.Printf("✅ Recording already exists, updating...")
			updates := map[string]interface{}{
				"file_path":    track.FilePath,
				"playlist_url": track.URL,
				"status":       "ended",
				"updated_at":   time.Now(),
			}
			err := db.DB.Table("livekit_egress_recordings").
				Where("track_sid = ? AND room_sid = ?", track.TrackSID, roomSID).
				Updates(updates).Error
			if err != nil {
				log.Printf("❌ Failed to update recording: %v", err)
			} else {
				log.Printf("✅ Recording updated")
			}
		} else {
			log.Printf("💾 Creating new recording...")
			now := time.Now()
			recording := map[string]interface{}{
				"id":           uuid.New().String(),
				"egress_id":    track.EgressID,
				"room_sid":     roomSID,
				"room_name":    roomName,
				"meeting_id":   meetingUUID,
				"track_sid":    track.TrackSID,
				"type":         track.Type,
				"status":       "ended",
				"file_path":    track.FilePath,
				"playlist_url": track.URL,
				"audio_only":   false,
				"started_at":   now,
				"ended_at":     now,
				"created_at":   now,
				"updated_at":   now,
			}
			err := db.DB.Table("livekit_egress_recordings").Create(recording).Error
			if err != nil {
				log.Printf("❌ Failed to create recording: %v", err)
			} else {
				log.Printf("✅ Recording created with EgressID: %s", track.EgressID)
			}
		}
	}

	log.Printf("\n✅ All database records created/updated successfully!")
	log.Printf("\n📍 You can now access the tracks:")
	for _, track := range tracks {
		log.Printf("  - %s", track.URL)
	}
}

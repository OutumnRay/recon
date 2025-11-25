package main

import (
	"log"

	"Recontext.online/pkg/database"
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

	// Find room by meeting_id
	var rooms []struct {
		SID       string
		Name      string
		MeetingID string
	}

	err = db.DB.Table("livekit_rooms").
		Select("sid, name, meeting_id").
		Where("meeting_id = ?", meetingID).
		Scan(&rooms).Error

	if err != nil {
		log.Fatalf("Failed to query room: %v", err)
	}

	if len(rooms) == 0 {
		log.Printf("❌ No room found for meeting_id: %s", meetingID)
		return
	}

	room := rooms[0]
	log.Printf("✅ Found room: SID=%s, Name=%s, MeetingID=%s", room.SID, room.Name, room.MeetingID)

	// Check tracks for this room
	var trackCount int64
	db.DB.Table("livekit_tracks").Where("room_sid = ?", room.SID).Count(&trackCount)
	log.Printf("📊 Tracks in this room: %d", trackCount)

	// Check recordings
	var recordingCount int64
	db.DB.Table("livekit_egress_recordings").Where("room_sid = ?", room.SID).Count(&recordingCount)
	log.Printf("📊 Egress recordings in this room: %d", recordingCount)

	// List all tracks
	if trackCount > 0 {
		var tracks []struct {
			SID  string
			Type string
		}
		db.DB.Table("livekit_tracks").
			Select("sid, type").
			Where("room_sid = ?", room.SID).
			Scan(&tracks)

		log.Printf("\n📋 Tracks:")
		for _, t := range tracks {
			log.Printf("  - %s (%s)", t.SID, t.Type)
		}
	}
}

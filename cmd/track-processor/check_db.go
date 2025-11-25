package main

import (
	"log"

	"Recontext.online/pkg/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Initialize database connection
	dsn := "host=192.168.5.153 port=5432 user=recontext password=recontext dbname=recontext sslmode=disable"
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db := &database.DB{DB: gormDB}

	roomSID := "RM_4qJSgKmcQmPW"

	log.Printf("🔍 Checking tracks in database for room: %s\n", roomSID)

	// Get all tracks for this room
	var tracks []struct {
		ID             string
		SID            string
		ParticipantSID string
		Type           string
		PublishedAt    string
	}

	err = db.DB.Table("livekit_tracks").
		Select("id, sid, participant_sid, type, published_at").
		Where("room_sid = ?", roomSID).
		Order("type, published_at").
		Scan(&tracks).Error

	if err != nil {
		log.Fatalf("Failed to query tracks: %v", err)
	}

	log.Printf("\n📊 Found %d tracks:\n", len(tracks))
	for i, track := range tracks {
		log.Printf("%d. SID: %s, Type: %s, ParticipantSID: %s, ID: %s",
			i+1, track.SID, track.Type, track.ParticipantSID, track.ID)
	}

	// Check egress recordings
	log.Printf("\n🔍 Checking egress recordings for room: %s\n", roomSID)

	var recordings []struct {
		ID             string
		TrackSID       *string
		Type           string
		Status         string
		PlaylistURL    string
		ParticipantSID *string
	}

	err = db.DB.Table("livekit_egress_recordings").
		Select("id, track_sid, type, status, playlist_url, participant_sid").
		Where("room_sid = ?", roomSID).
		Order("type").
		Scan(&recordings).Error

	if err != nil {
		log.Fatalf("Failed to query recordings: %v", err)
	}

	log.Printf("\n📊 Found %d egress recordings:\n", len(recordings))
	for i, rec := range recordings {
		trackSID := "NULL"
		if rec.TrackSID != nil {
			trackSID = *rec.TrackSID
		}
		partSID := "NULL"
		if rec.ParticipantSID != nil {
			partSID = *rec.ParticipantSID
		}
		log.Printf("%d. Type: %s, Status: %s, TrackSID: %s, ParticipantSID: %s, PlaylistURL: %s",
			i+1, rec.Type, rec.Status, trackSID, partSID, rec.PlaylistURL)
	}

	// Get participants
	log.Printf("\n🔍 Checking participants for room: %s\n", roomSID)

	var participants []struct {
		ID   string
		SID  string
		Name string
	}

	err = db.DB.Table("livekit_participants").
		Select("id, sid, name").
		Where("room_sid = ?", roomSID).
		Scan(&participants).Error

	if err != nil {
		log.Fatalf("Failed to query participants: %v", err)
	}

	log.Printf("\n📊 Found %d participants:\n", len(participants))
	for i, p := range participants {
		log.Printf("%d. SID: %s, Name: %s, ID: %s", i+1, p.SID, p.Name, p.ID)
	}
}

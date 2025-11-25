package main

import (
	"context"
	"log"
	"time"

	"Recontext.online/pkg/video"
)

func main() {
	// Конфигурация из docker-compose.yml
	config := video.TrackCombinerConfig{
		Endpoint:        "192.168.5.153:9000", // MINIO_ENDPOINT из docker-compose
		AccessKeyID:     "minioadmin",          // MINIO_ACCESS_KEY
		SecretAccessKey: "32a4953d5bff4a1c6aea4d4ccfb757e5", // MINIO_SECRET_KEY
		BucketName:      "recontext",           // MINIO_BUCKET
		UseSSL:          false,                 // Внутренняя сеть без SSL
		WorkDir:         "./temp-tracks",
	}

	// Создаем TrackCombiner
	combiner, err := video.NewTrackCombiner(config)
	if err != nil {
		log.Fatalf("Failed to create track combiner: %v", err)
	}
	defer combiner.CleanupAll()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Параметры встречи
	meetingID := "082494f4-e0a7-43c0-a19d-6d6e35ef429c"
	roomSID := "RM_4qJSgKmcQmPW"

	log.Printf("🎬 Processing tracks for meeting %s, room %s", meetingID, roomSID)

	// Вариант 1: Обработать все треки в комнате
	tracks, err := combiner.CombineTracksByRoom(ctx, meetingID, roomSID)
	if err != nil {
		log.Fatalf("Failed to combine tracks: %v", err)
	}

	log.Printf("\n✅ Successfully processed %d tracks:", len(tracks))
	for i, track := range tracks {
		if track.Error != nil {
			log.Printf("%d. ❌ Track %s: ERROR - %v", i+1, track.TrackID, track.Error)
			continue
		}

		log.Printf("%d. ✅ Track %s:", i+1, track.TrackID)
		log.Printf("   Type: %s", track.Type)
		log.Printf("   Size: %.2f MB", float64(track.Size)/(1024*1024))
		log.Printf("   Duration: %.2f seconds", track.Duration)
		log.Printf("   Local path: %s", track.LocalPath)

		// Загружаем обратно в MinIO
		url, err := combiner.UploadCombinedTrack(ctx, &track, meetingID, roomSID)
		if err != nil {
			log.Printf("   ⚠️ Upload failed: %v", err)
		} else {
			log.Printf("   📤 Uploaded to: %s", url)
		}

		// Опционально: очищаем временные файлы трека
		if err := combiner.Cleanup(track.TrackID); err != nil {
			log.Printf("   ⚠️ Cleanup failed: %v", err)
		}
	}

	log.Printf("\n🎉 All tracks processed!")

	// Вариант 2: Обработать конкретный трек
	/*
		trackID := "TR_VCoyexbMM3nSRm"
		track, err := combiner.CombineSingleTrack(ctx, meetingID, roomSID, trackID)
		if err != nil {
			log.Fatalf("Failed to combine track: %v", err)
		}

		log.Printf("Track %s combined successfully:", track.TrackID)
		log.Printf("  Type: %s", track.Type)
		log.Printf("  Size: %.2f MB", float64(track.Size)/(1024*1024))
		log.Printf("  Duration: %.2f seconds", track.Duration)
		log.Printf("  Path: %s", track.LocalPath)

		// Загружаем в MinIO
		url, err := combiner.UploadCombinedTrack(ctx, track, meetingID, roomSID)
		if err != nil {
			log.Fatalf("Failed to upload: %v", err)
		}
		log.Printf("  Uploaded to: %s", url)
	*/
}

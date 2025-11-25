package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
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
		log.Fatalf("Failed to create MinIO client: %v", err)
	}

	bucketName := "recontext"
	meetingID := "082494f4-e0a7-43c0-a19d-6d6e35ef429c"
	roomSID := "RM_4qJSgKmcQmPW"
	prefix := fmt.Sprintf("%s_%s/tracks/", meetingID, roomSID)

	log.Printf("🔍 Scanning MinIO bucket '%s' with prefix '%s'", bucketName, prefix)

	// List all objects in the tracks directory
	ctx := context.Background()
	objectCh := minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	tracks := make(map[string][]string) // trackID -> list of files

	for object := range objectCh {
		if object.Err != nil {
			log.Printf("Error listing object: %v", object.Err)
			continue
		}

		log.Printf("📁 Found: %s (size: %d bytes)", object.Key, object.Size)

		// Extract track ID from path
		// Expected format: 082494f4-e0a7-43c0-a19d-6d6e35ef429c_RM_4qJSgKmcQmPW/tracks/TR_VCoyexbMM3nSRm/...
		parts := strings.Split(object.Key, "/")
		if len(parts) >= 3 && strings.HasPrefix(parts[2], "TR_") {
			trackID := parts[2]
			tracks[trackID] = append(tracks[trackID], object.Key)
		}
	}

	log.Printf("\n📊 Summary:")
	log.Printf("Found %d tracks:", len(tracks))
	for trackID, files := range tracks {
		log.Printf("  Track %s: %d files", trackID, len(files))
		for _, file := range files {
			log.Printf("    - %s", filepath.Base(file))
		}
	}

	// Download and process tracks
	if len(tracks) == 0 {
		log.Printf("⚠️ No tracks found in MinIO")
		return
	}

	downloadDir := "./downloads"
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		log.Fatalf("Failed to create download directory: %v", err)
	}

	for trackID, files := range tracks {
		log.Printf("\n📥 Processing track: %s", trackID)
		trackDir := filepath.Join(downloadDir, trackID)
		if err := os.MkdirAll(trackDir, 0755); err != nil {
			log.Printf("Failed to create track directory: %v", err)
			continue
		}

		// Find the m3u8 playlist
		var playlistPath string
		for _, file := range files {
			if strings.HasSuffix(file, ".m3u8") {
				playlistPath = file
				log.Printf("  📋 Found playlist: %s", filepath.Base(file))
				break
			}
		}

		if playlistPath == "" {
			log.Printf("  ⚠️ No .m3u8 playlist found for track %s", trackID)
			continue
		}

		// Download the playlist file
		localPlaylistPath := filepath.Join(trackDir, filepath.Base(playlistPath))
		if err := minioClient.FGetObject(ctx, bucketName, playlistPath, localPlaylistPath, minio.GetObjectOptions{}); err != nil {
			log.Printf("  ❌ Failed to download playlist: %v", err)
			continue
		}
		log.Printf("  ✅ Downloaded playlist to: %s", localPlaylistPath)

		// Read playlist to find segment files
		playlistContent, err := os.ReadFile(localPlaylistPath)
		if err != nil {
			log.Printf("  ❌ Failed to read playlist: %v", err)
			continue
		}

		// Download all segment files mentioned in playlist
		lines := strings.Split(string(playlistContent), "\n")
		segmentCount := 0
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// This is a segment file
			segmentPath := filepath.Join(filepath.Dir(playlistPath), line)
			localSegmentPath := filepath.Join(trackDir, line)

			if err := minioClient.FGetObject(ctx, bucketName, segmentPath, localSegmentPath, minio.GetObjectOptions{}); err != nil {
				log.Printf("  ❌ Failed to download segment %s: %v", line, err)
				continue
			}
			segmentCount++
		}
		log.Printf("  ✅ Downloaded %d segments for track %s", segmentCount, trackID)
	}

	log.Printf("\n✅ All tracks processed successfully!")
	log.Printf("📁 Files downloaded to: %s", downloadDir)
}

package main

import (
	"context"
	"log"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	// Конфигурация из docker-compose
	endpoint := "192.168.5.153:9000"
	accessKeyID := "minioadmin"
	secretAccessKey := "32a4953d5bff4a1c6aea4d4ccfb757e5"
	bucketName := "recontext"
	useSSL := false

	// Создаем MinIO клиент
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalf("Failed to create MinIO client: %v", err)
	}

	ctx := context.Background()

	// Ищем папки с композитными видео (формат: meetingID_roomSID)
	log.Printf("📋 Searching for composite videos in bucket '%s':\n", bucketName)

	objectCh := minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Recursive: true,
	})

	compositeFiles := make(map[string][]string)

	for object := range objectCh {
		if object.Err != nil {
			log.Printf("Error: %v", object.Err)
			continue
		}

		// Ищем файлы composite.m3u8 и composite_*.ts
		if object.Key != "" {
			if matched, _ := filepath.Match("*/composite.m3u8", object.Key); matched {
				prefix := filepath.Dir(object.Key)
				compositeFiles[prefix] = append(compositeFiles[prefix], object.Key)
			} else if matched, _ := filepath.Match("*/composite_*.ts", object.Key); matched {
				prefix := filepath.Dir(object.Key)
				compositeFiles[prefix] = append(compositeFiles[prefix], object.Key)
			}
		}
	}

	if len(compositeFiles) == 0 {
		log.Printf("⚠️ No composite videos found")
		return
	}

	log.Printf("\n✅ Found %d meeting(s) with composite video:\n", len(compositeFiles))

	for prefix, files := range compositeFiles {
		log.Printf("\n📁 %s/", prefix)

		var m3u8Count, tsCount int
		for _, file := range files {
			if filepath.Ext(file) == ".m3u8" {
				m3u8Count++
				log.Printf("   ✓ %s", filepath.Base(file))
			} else if filepath.Ext(file) == ".ts" {
				tsCount++
			}
		}

		log.Printf("   📊 Total: %d playlist(s), %d segment(s)", m3u8Count, tsCount)

		// Показываем несколько примеров сегментов
		if tsCount > 0 {
			log.Printf("   📹 Sample segments:")
			count := 0
			for _, file := range files {
				if filepath.Ext(file) == ".ts" && count < 3 {
					log.Printf("      - %s", filepath.Base(file))
					count++
				}
			}
			if tsCount > 3 {
				log.Printf("      ... and %d more", tsCount-3)
			}
		}
	}
}

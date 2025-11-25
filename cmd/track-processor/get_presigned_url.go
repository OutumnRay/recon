package main

import (
	"context"
	"log"
	"time"

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

	// Список файлов для проверки
	files := []string{
		"082494f4-e0a7-43c0-a19d-6d6e35ef429c_RM_4qJSgKmcQmPW/tracks/TR_VCoyexbMM3nSRm.mp4",
		"082494f4-e0a7-43c0-a19d-6d6e35ef429c_RM_4qJSgKmcQmPW/tracks/TR_AMcJtxis6SfUzT.mp4",
	}

	log.Printf("📋 Generating presigned URLs (valid for 7 days):\n")

	for i, objectName := range files {
		// Генерируем presigned URL со сроком действия 7 дней
		presignedURL, err := minioClient.PresignedGetObject(ctx, bucketName, objectName, 7*24*time.Hour, nil)
		if err != nil {
			log.Printf("%d. ❌ %s: Failed - %v", i+1, objectName, err)
			continue
		}

		// Получаем информацию о файле
		stat, err := minioClient.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
		if err != nil {
			log.Printf("%d. ❌ %s: Failed to stat - %v", i+1, objectName, err)
			continue
		}

		log.Printf("\n%d. ✅ %s", i+1, objectName)
		log.Printf("   Size: %.2f MB", float64(stat.Size)/(1024*1024))
		log.Printf("   Type: %s", stat.ContentType)
		log.Printf("   Last Modified: %s", stat.LastModified.Format(time.RFC3339))
		log.Printf("   URL: %s", presignedURL.String())
	}

	// Также проверим оригинальные HLS файлы
	log.Printf("\n\n📋 Checking original HLS files:\n")

	hlsFiles := []string{
		"082494f4-e0a7-43c0-a19d-6d6e35ef429c_RM_4qJSgKmcQmPW/tracks/TR_VCoyexbMM3nSRm.m3u8",
		"082494f4-e0a7-43c0-a19d-6d6e35ef429c_RM_4qJSgKmcQmPW/tracks/TR_AMcJtxis6SfUzT.m3u8",
	}

	for i, objectName := range hlsFiles {
		stat, err := minioClient.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
		if err != nil {
			log.Printf("%d. ❌ %s: Not found - %v", i+1, objectName, err)
			continue
		}

		presignedURL, _ := minioClient.PresignedGetObject(ctx, bucketName, objectName, 7*24*time.Hour, nil)

		log.Printf("\n%d. ✅ %s", i+1, objectName)
		log.Printf("   Size: %d bytes", stat.Size)
		log.Printf("   URL: %s", presignedURL.String())
	}
}

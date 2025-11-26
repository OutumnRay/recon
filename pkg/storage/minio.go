package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOClient универсальный клиент для работы с MinIO
type MinIOClient struct {
	client         *minio.Client
	bucket         string
	publicEndpoint string // Публичный endpoint для формирования URL
	useSSL         bool   // Использовать HTTPS для публичных URL
}

// MinIOConfig конфигурация для MinIO клиента
type MinIOConfig struct {
	// Endpoint для подключения (например, minio:9000 внутри Docker)
	Endpoint string
	// Публичный endpoint для формирования URL (например, 192.168.5.153:9000)
	PublicEndpoint string
	// Credentials
	AccessKey string
	SecretKey string
	// Bucket
	Bucket string
	// SSL
	UseSSL bool
}

// NewMinIOClient создает новый MinIO клиент
func NewMinIOClient(config MinIOConfig) (*MinIOClient, error) {
	// Создаем клиент MinIO
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// Определяем публичный endpoint
	publicEndpoint := config.PublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = config.Endpoint
	}

	return &MinIOClient{
		client:         client,
		bucket:         config.Bucket,
		publicEndpoint: publicEndpoint,
		useSSL:         config.UseSSL,
	}, nil
}

// NewMinIOClientFromEnv создает MinIO клиент из переменных окружения
func NewMinIOClientFromEnv() (*MinIOClient, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		return nil, fmt.Errorf("MINIO_ENDPOINT not set")
	}

	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	if accessKey == "" {
		return nil, fmt.Errorf("MINIO_ACCESS_KEY not set")
	}

	secretKey := os.Getenv("MINIO_SECRET_KEY")
	if secretKey == "" {
		return nil, fmt.Errorf("MINIO_SECRET_KEY not set")
	}

	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		bucket = "recontext"
	}

	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	publicEndpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT")

	config := MinIOConfig{
		Endpoint:       endpoint,
		PublicEndpoint: publicEndpoint,
		AccessKey:      accessKey,
		SecretKey:      secretKey,
		Bucket:         bucket,
		UseSSL:         useSSL,
	}

	return NewMinIOClient(config)
}

// UploadFile загружает файл в MinIO
func (mc *MinIOClient) UploadFile(ctx context.Context, localPath, remotePath string) (string, error) {
	log.Printf("📤 Uploading file to MinIO")
	log.Printf("   Local: %s", localPath)
	log.Printf("   Remote: %s", remotePath)

	// Открываем файл
	file, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Получаем размер файла
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	// Определяем content type
	contentType := getContentType(localPath)

	// Загружаем файл
	_, err = mc.client.PutObject(
		ctx,
		mc.bucket,
		remotePath,
		file,
		fileInfo.Size(),
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Формируем публичный URL
	url := mc.GetPublicURL(remotePath)
	log.Printf("✅ File uploaded: %s", url)

	return url, nil
}

// DownloadFile скачивает файл из MinIO
func (mc *MinIOClient) DownloadFile(ctx context.Context, remotePath, localPath string) error {
	// Создаем директорию если не существует
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Получаем объект из MinIO
	object, err := mc.client.GetObject(ctx, mc.bucket, remotePath, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

	// Создаем локальный файл
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Копируем данные
	if _, err := io.Copy(file, object); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	return nil
}

// UploadDirectory загружает всю директорию в MinIO
func (mc *MinIOClient) UploadDirectory(ctx context.Context, localDir, remotePrefix string) ([]string, error) {
	log.Printf("📤 Uploading directory to MinIO")
	log.Printf("   Local: %s", localDir)
	log.Printf("   Remote prefix: %s", remotePrefix)

	var uploadedURLs []string

	// Проходим по всем файлам в директории
	err := filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем директории
		if info.IsDir() {
			return nil
		}

		// Вычисляем относительный путь
		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Формируем удаленный путь (используем / для MinIO)
		remotePath := filepath.Join(remotePrefix, relPath)
		remotePath = filepath.ToSlash(remotePath)

		// Загружаем файл
		url, err := mc.UploadFile(ctx, path, remotePath)
		if err != nil {
			log.Printf("⚠️ Failed to upload %s: %v", path, err)
			return err
		}

		uploadedURLs = append(uploadedURLs, url)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upload directory: %w", err)
	}

	log.Printf("✅ Directory uploaded: %d files", len(uploadedURLs))
	return uploadedURLs, nil
}

// DeleteFile удаляет файл из MinIO
func (mc *MinIOClient) DeleteFile(ctx context.Context, remotePath string) error {
	err := mc.client.RemoveObject(ctx, mc.bucket, remotePath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// DeleteDirectory удаляет всю директорию (все объекты с указанным префиксом) из MinIO
func (mc *MinIOClient) DeleteDirectory(ctx context.Context, remotePrefix string) (int, error) {
	log.Printf("🗑️  Deleting directory from MinIO: %s", remotePrefix)

	// Ensure prefix ends with / for directory matching
	if !strings.HasSuffix(remotePrefix, "/") {
		remotePrefix += "/"
	}

	// List all objects with the prefix
	objectsCh := mc.client.ListObjects(ctx, mc.bucket, minio.ListObjectsOptions{
		Prefix:    remotePrefix,
		Recursive: true,
	})

	deletedCount := 0
	var lastErr error

	// Delete each object
	for object := range objectsCh {
		if object.Err != nil {
			log.Printf("⚠️ Error listing object: %v", object.Err)
			lastErr = object.Err
			continue
		}

		log.Printf("   Deleting: %s", object.Key)
		err := mc.client.RemoveObject(ctx, mc.bucket, object.Key, minio.RemoveObjectOptions{})
		if err != nil {
			log.Printf("⚠️ Failed to delete %s: %v", object.Key, err)
			lastErr = err
			continue
		}

		deletedCount++
	}

	if lastErr != nil && deletedCount == 0 {
		return 0, fmt.Errorf("failed to delete directory: %w", lastErr)
	}

	log.Printf("✅ Deleted %d files from MinIO", deletedCount)
	return deletedCount, nil
}

// GetPublicURL возвращает публичный URL для объекта
func (mc *MinIOClient) GetPublicURL(objectKey string) string {
	protocol := "http"
	if mc.useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, mc.publicEndpoint, mc.bucket, objectKey)
}

// GetRelativePath извлекает относительный путь из полного URL
func (mc *MinIOClient) GetRelativePath(fullURL string) string {
	// Удаляем протокол и хост
	// Формат: http://host:port/bucket/path -> path
	parts := strings.SplitN(fullURL, "/"+mc.bucket+"/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return fullURL
}

// EnsureBucket проверяет существование бакета и создает его при необходимости
func (mc *MinIOClient) EnsureBucket(ctx context.Context) error {
	exists, err := mc.client.BucketExists(ctx, mc.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		log.Printf("📦 Creating bucket: %s", mc.bucket)
		err = mc.client.MakeBucket(ctx, mc.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("✅ Bucket created: %s", mc.bucket)
	}

	return nil
}

// GetClient возвращает базовый MinIO клиент для прямого доступа
func (mc *MinIOClient) GetClient() *minio.Client {
	return mc.client
}

// GetBucket возвращает имя бакета
func (mc *MinIOClient) GetBucket() string {
	return mc.bucket
}

// getContentType определяет MIME тип файла по расширению
func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".m3u8":
		return "application/vnd.apple.mpegurl"
	case ".ts":
		return "video/MP2T"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".json":
		return "application/json"
	case ".vtt":
		return "text/vtt"
	case ".srt":
		return "application/x-subrip"
	default:
		return "application/octet-stream"
	}
}

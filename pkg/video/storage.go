package video

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

// StorageConfig конфигурация для S3/MinIO хранилища
type StorageConfig struct {
	Endpoint       string
	AccessKey      string
	SecretKey      string
	Bucket         string
	Region         string
	UseSSL         bool
	ForcePathStyle bool // Для MinIO обычно true
}

// StorageUploader загружает файлы в S3/MinIO
type StorageUploader struct {
	client *minio.Client
	config StorageConfig
}

// NewStorageUploader создает новый загрузчик в S3/MinIO
func NewStorageUploader(config StorageConfig) (*StorageUploader, error) {
	// Создаем клиент MinIO (совместим с S3)
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &StorageUploader{
		client: client,
		config: config,
	}, nil
}

// NewStorageUploaderFromEnv создает загрузчик из переменных окружения
func NewStorageUploaderFromEnv() (*StorageUploader, error) {
	// Проверяем MINIO_* переменные в первую очередь, затем S3_*
	endpoint := getEnv("MINIO_ENDPOINT", "")
	if endpoint == "" {
		endpoint = getEnv("S3_ENDPOINT", "localhost:9000")
	}

	accessKey := getEnv("MINIO_ACCESS_KEY", "")
	if accessKey == "" {
		accessKey = getEnv("S3_ACCESS_KEY", "minioadmin")
	}

	secretKey := getEnv("MINIO_SECRET_KEY", "")
	if secretKey == "" {
		secretKey = getEnv("S3_SECRET_KEY", "minioadmin")
	}

	bucket := getEnv("MINIO_BUCKET", "")
	if bucket == "" {
		bucket = getEnv("S3_BUCKET", "recordings")
	}

	useSSL := getEnv("MINIO_USE_SSL", "")
	if useSSL == "" {
		useSSL = getEnv("S3_USE_SSL", "false")
	}

	config := StorageConfig{
		Endpoint:       endpoint,
		AccessKey:      accessKey,
		SecretKey:      secretKey,
		Bucket:         bucket,
		Region:         getEnv("S3_REGION", "us-east-1"),
		UseSSL:         useSSL == "true",
		ForcePathStyle: getEnv("S3_FORCE_PATH_STYLE", "true") == "true",
	}

	return NewStorageUploader(config)
}

// UploadFile загружает один файл в S3/MinIO
func (su *StorageUploader) UploadFile(ctx context.Context, localPath, remotePath string) (string, error) {
	log.Printf("📤 Uploading file to S3")
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
	_, err = su.client.PutObject(
		ctx,
		su.config.Bucket,
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
	url := su.GetPublicURL(remotePath)
	log.Printf("✅ File uploaded: %s", url)

	return url, nil
}

// UploadDirectory загружает всю директорию в S3/MinIO
func (su *StorageUploader) UploadDirectory(ctx context.Context, localDir, remotePrefix string) ([]string, error) {
	log.Printf("📤 Uploading directory to S3")
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

		// Формируем удаленный путь (используем / для S3)
		remotePath := filepath.Join(remotePrefix, relPath)
		remotePath = filepath.ToSlash(remotePath)

		// Загружаем файл
		url, err := su.UploadFile(ctx, path, remotePath)
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

// GetPublicURL возвращает публичный URL для объекта
func (su *StorageUploader) GetPublicURL(objectKey string) string {
	protocol := "http"
	if su.config.UseSSL {
		protocol = "https"
	}

	// Для MinIO с path-style
	if su.config.ForcePathStyle {
		return fmt.Sprintf("%s://%s/%s/%s", protocol, su.config.Endpoint, su.config.Bucket, objectKey)
	}

	// Для S3 с virtual-hosted-style
	return fmt.Sprintf("%s://%s.s3.%s.amazonaws.com/%s", protocol, su.config.Bucket, su.config.Region, objectKey)
}

// EnsureBucket проверяет существование бакета и создает его при необходимости
func (su *StorageUploader) EnsureBucket(ctx context.Context) error {
	exists, err := su.client.BucketExists(ctx, su.config.Bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		log.Printf("📦 Creating bucket: %s", su.config.Bucket)
		err = su.client.MakeBucket(ctx, su.config.Bucket, minio.MakeBucketOptions{
			Region: su.config.Region,
		})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("✅ Bucket created: %s", su.config.Bucket)
	}

	return nil
}

// DeleteFile удаляет файл из S3/MinIO
func (su *StorageUploader) DeleteFile(ctx context.Context, remotePath string) error {
	err := su.client.RemoveObject(ctx, su.config.Bucket, remotePath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// DownloadFile скачивает файл из S3/MinIO
func (su *StorageUploader) DownloadFile(ctx context.Context, remotePath, localPath string) error {
	// Создаем директорию если не существует
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Получаем объект
	object, err := su.client.GetObject(ctx, su.config.Bucket, remotePath, minio.GetObjectOptions{})
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

// GetClient возвращает MinIO клиент для прямого доступа
func (su *StorageUploader) GetClient() *minio.Client {
	return su.client
}

// GetBucket возвращает имя бакета
func (su *StorageUploader) GetBucket() string {
	return su.config.Bucket
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

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

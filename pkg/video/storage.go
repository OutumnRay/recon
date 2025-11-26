package video

import (
	"context"

	"Recontext.online/pkg/storage"
	"github.com/minio/minio-go/v7"
)

// StorageUploader загружает файлы в S3/MinIO
// Это обертка над storage.MinIOClient для обратной совместимости
type StorageUploader struct {
	minioClient *storage.MinIOClient
}

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

// NewStorageUploader создает новый загрузчик в S3/MinIO
func NewStorageUploader(config StorageConfig) (*StorageUploader, error) {
	minioConfig := storage.MinIOConfig{
		Endpoint:       config.Endpoint,
		PublicEndpoint: "", // Используем endpoint как публичный
		AccessKey:      config.AccessKey,
		SecretKey:      config.SecretKey,
		Bucket:         config.Bucket,
		UseSSL:         config.UseSSL,
	}

	client, err := storage.NewMinIOClient(minioConfig)
	if err != nil {
		return nil, err
	}

	return &StorageUploader{
		minioClient: client,
	}, nil
}

// NewStorageUploaderFromEnv создает загрузчик из переменных окружения
func NewStorageUploaderFromEnv() (*StorageUploader, error) {
	client, err := storage.NewMinIOClientFromEnv()
	if err != nil {
		return nil, err
	}

	return &StorageUploader{
		minioClient: client,
	}, nil
}

// UploadFile загружает один файл в S3/MinIO
func (su *StorageUploader) UploadFile(ctx context.Context, localPath, remotePath string) (string, error) {
	return su.minioClient.UploadFile(ctx, localPath, remotePath)
}

// UploadDirectory загружает всю директорию в S3/MinIO
func (su *StorageUploader) UploadDirectory(ctx context.Context, localDir, remotePrefix string) ([]string, error) {
	return su.minioClient.UploadDirectory(ctx, localDir, remotePrefix)
}

// GetPublicURL возвращает публичный URL для объекта
func (su *StorageUploader) GetPublicURL(objectKey string) string {
	return su.minioClient.GetPublicURL(objectKey)
}

// EnsureBucket проверяет существование бакета и создает его при необходимости
func (su *StorageUploader) EnsureBucket(ctx context.Context) error {
	return su.minioClient.EnsureBucket(ctx)
}

// DeleteFile удаляет файл из S3/MinIO
func (su *StorageUploader) DeleteFile(ctx context.Context, remotePath string) error {
	return su.minioClient.DeleteFile(ctx, remotePath)
}

// DownloadFile скачивает файл из S3/MinIO
func (su *StorageUploader) DownloadFile(ctx context.Context, remotePath, localPath string) error {
	return su.minioClient.DownloadFile(ctx, remotePath, localPath)
}

// GetClient возвращает MinIO клиент для прямого доступа
func (su *StorageUploader) GetClient() *minio.Client {
	return su.minioClient.GetClient()
}

// GetBucket возвращает имя бакета
func (su *StorageUploader) GetBucket() string {
	return su.minioClient.GetBucket()
}

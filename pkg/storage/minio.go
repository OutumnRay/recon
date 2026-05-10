package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOClient struct {
	client         *minio.Client // internal client for server-to-server operations
	publicClient   *minio.Client // client signed with public endpoint for presigned URLs
	bucket         string
	publicEndpoint string
	useSSL         bool
}

type MinIOConfig struct {
	Endpoint       string
	PublicEndpoint string
	AccessKey      string
	SecretKey      string
	Bucket         string
	UseSSL         bool
}

func NewMinIOClient(config MinIOConfig) (*MinIOClient, error) {
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	publicEndpoint := config.PublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = config.Endpoint
	}

	// Create a separate client whose endpoint matches the public URL.
	// Presigned URLs embed the signing host, so the client that generates them
	// must use the same host that the browser/Postman will send the request to.
	publicHost, publicSSL := parseEndpointHost(publicEndpoint)
	publicClient := client
	if publicHost != config.Endpoint {
		pc, pcErr := minio.New(publicHost, &minio.Options{
			Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
			Secure: publicSSL,
		})
		if pcErr == nil {
			publicClient = pc
		}
	}

	return &MinIOClient{
		client:         client,
		publicClient:   publicClient,
		bucket:         config.Bucket,
		publicEndpoint: publicEndpoint,
		useSSL:         config.UseSSL,
	}, nil
}

// parseEndpointHost strips the scheme and any path from an endpoint string.
// MinIO SDK requires a bare "host" or "host:port" — no scheme, no path.
//   "http://185.200.240.31:9000"   → ("185.200.240.31:9000", false)
//   "https://24recontext.ru"       → ("24recontext.ru", true)
//   "https://24recontext.ru/minio" → ("24recontext.ru", true)  path stripped
//   "minio:9000"                   → ("minio:9000", false)
func parseEndpointHost(endpoint string) (host string, useSSL bool) {
	rest := endpoint
	switch {
	case strings.HasPrefix(endpoint, "https://"):
		rest = strings.TrimPrefix(endpoint, "https://")
		useSSL = true
	case strings.HasPrefix(endpoint, "http://"):
		rest = strings.TrimPrefix(endpoint, "http://")
	}
	if idx := strings.Index(rest, "/"); idx != -1 {
		rest = rest[:idx]
	}
	return rest, useSSL
}

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

	config := MinIOConfig{
		Endpoint:       endpoint,
		PublicEndpoint: os.Getenv("MINIO_PUBLIC_ENDPOINT"),
		AccessKey:      accessKey,
		SecretKey:      secretKey,
		Bucket:         bucket,
		UseSSL:         os.Getenv("MINIO_USE_SSL") == "true",
	}

	return NewMinIOClient(config)
}

// UploadFile загружает файл из локального пути в MinIO
func (mc *MinIOClient) UploadFile(ctx context.Context, localPath, remotePath string) (string, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	contentType := getContentType(localPath)

	_, err = mc.client.PutObject(ctx, mc.bucket, remotePath, file, fileInfo.Size(),
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	url := mc.GetPublicURL(remotePath)
	log.Printf("✅ File uploaded to MinIO: %s", url)
	return url, nil
}

// UploadReader загружает данные из io.Reader в MinIO (потоковая загрузка)
func (mc *MinIOClient) UploadReader(ctx context.Context, reader io.Reader, size int64, remotePath, contentType string) (string, error) {
	if contentType == "" {
		contentType = getContentType(remotePath)
	}

	_, err := mc.client.PutObject(ctx, mc.bucket, remotePath, reader, size,
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", fmt.Errorf("failed to upload reader to minio: %w", err)
	}

	url := mc.GetPublicURL(remotePath)
	log.Printf("✅ Stream uploaded to MinIO: %s", url)
	return url, nil
}

func (mc *MinIOClient) DownloadFile(ctx context.Context, remotePath, localPath string) error {
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	object, err := mc.client.GetObject(ctx, mc.bucket, remotePath, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, object); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	return nil
}

func (mc *MinIOClient) UploadDirectory(ctx context.Context, localDir, remotePrefix string) ([]string, error) {
	var uploadedURLs []string

	err := filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		remotePath := filepath.ToSlash(filepath.Join(remotePrefix, relPath))
		url, err := mc.UploadFile(ctx, path, remotePath)
		if err != nil {
			return fmt.Errorf("failed to upload %s: %w", path, err)
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

func (mc *MinIOClient) DeleteFile(ctx context.Context, remotePath string) error {
	err := mc.client.RemoveObject(ctx, mc.bucket, remotePath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (mc *MinIOClient) DeleteDirectory(ctx context.Context, remotePrefix string) (int, error) {
	if !strings.HasSuffix(remotePrefix, "/") {
		remotePrefix += "/"
	}

	objectsCh := mc.client.ListObjects(ctx, mc.bucket, minio.ListObjectsOptions{
		Prefix:    remotePrefix,
		Recursive: true,
	})

	deletedCount := 0
	var lastErr error

	for object := range objectsCh {
		if object.Err != nil {
			lastErr = object.Err
			continue
		}

		err := mc.client.RemoveObject(ctx, mc.bucket, object.Key, minio.RemoveObjectOptions{})
		if err != nil {
			lastErr = err
			continue
		}
		deletedCount++
	}

	if lastErr != nil && deletedCount == 0 {
		return 0, fmt.Errorf("failed to delete directory: %w", lastErr)
	}

	return deletedCount, nil
}

func (mc *MinIOClient) GetPublicURL(objectKey string) string {
	protocol := "http"
	if mc.useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, mc.publicEndpoint, mc.bucket, objectKey)
}

func (mc *MinIOClient) GetRelativePath(fullURL string) string {
	parts := strings.SplitN(fullURL, "/"+mc.bucket+"/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return fullURL
}

func (mc *MinIOClient) EnsureBucket(ctx context.Context) error {
	exists, err := mc.client.BucketExists(ctx, mc.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = mc.client.MakeBucket(ctx, mc.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("✅ Bucket created: %s", mc.bucket)
	}

	return nil
}

// PresignedPutObject returns a time-limited URL the client can use to PUT an object
// directly into MinIO without credentials. expiry must not exceed 7 days.
// The URL is signed using publicClient so the embedded host matches the address
// the browser will actually connect to.
func (mc *MinIOClient) PresignedPutObject(ctx context.Context, objectPath string, expiry time.Duration) (string, error) {
	u, err := mc.publicClient.PresignedPutObject(ctx, mc.bucket, objectPath, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned PUT URL: %w", err)
	}
	return u.String(), nil
}

// PresignedGetObject returns a time-limited URL for downloading an object.
func (mc *MinIOClient) PresignedGetObject(ctx context.Context, objectPath string, expiry time.Duration) (string, error) {
	u, err := mc.publicClient.PresignedGetObject(ctx, mc.bucket, objectPath, expiry, make(url.Values))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned GET URL: %w", err)
	}
	return u.String(), nil
}

// StatObject returns metadata for an object (size, ETag, etc.).
// Returns an error if the object does not exist.
func (mc *MinIOClient) StatObject(ctx context.Context, objectPath string) (minio.ObjectInfo, error) {
	return mc.client.StatObject(ctx, mc.bucket, objectPath, minio.StatObjectOptions{})
}

func (mc *MinIOClient) GetClient() *minio.Client {
	return mc.client
}

func (mc *MinIOClient) GetBucket() string {
	return mc.bucket
}

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

package video

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// TrackCombiner объединяет HLS сегменты треков из MinIO/S3 в единые MP4 файлы
type TrackCombiner struct {
	minioClient    *minio.Client
	bucketName     string
	workDir        string
	publicEndpoint string // Публичный endpoint для формирования URL
	useSSL         bool   // Использовать HTTPS для публичных URL
}

// TrackCombinerConfig конфигурация для TrackCombiner
type TrackCombinerConfig struct {
	// MinIO/S3 настройки
	Endpoint        string // Внутренний endpoint для подключения (например, minio:9000)
	PublicEndpoint  string // Публичный endpoint для формирования URL (например, api.storage.recontext.online)
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	UseSSL          bool

	// Рабочая директория для временных файлов
	WorkDir string
}

// CombinedTrack результат объединения трека
type CombinedTrack struct {
	TrackID     string
	LocalPath   string
	Size        int64
	Duration    float64
	Type        string // "audio" or "video"
	Error       error
}

// NewTrackCombiner создает новый TrackCombiner
func NewTrackCombiner(config TrackCombinerConfig) (*TrackCombiner, error) {
	// Создаем MinIO клиент
	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Создаем рабочую директорию
	workDir := config.WorkDir
	if workDir == "" {
		workDir = filepath.Join(os.TempDir(), "track-combiner")
	}
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	// Определяем публичный endpoint
	publicEndpoint := config.PublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = config.Endpoint
	}

	return &TrackCombiner{
		minioClient:    minioClient,
		bucketName:     config.BucketName,
		workDir:        workDir,
		publicEndpoint: publicEndpoint,
		useSSL:         config.UseSSL,
	}, nil
}

// NewTrackCombinerFromEnv создает TrackCombiner из переменных окружения
func NewTrackCombinerFromEnv() (*TrackCombiner, error) {
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

	// Определяем использование SSL
	useSSL := false
	if os.Getenv("MINIO_USE_SSL") == "true" {
		useSSL = true
	}

	// Публичный endpoint для формирования URL (если не указан, используется внутренний)
	publicEndpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT")

	config := TrackCombinerConfig{
		Endpoint:        endpoint,
		PublicEndpoint:  publicEndpoint,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		BucketName:      bucket,
		UseSSL:          useSSL,
		WorkDir:         os.Getenv("TRACK_COMBINER_WORK_DIR"),
	}

	return NewTrackCombiner(config)
}

// CombineTracksByRoom скачивает и объединяет все треки для указанной комнаты
func (tc *TrackCombiner) CombineTracksByRoom(ctx context.Context, meetingID, roomSID string) ([]CombinedTrack, error) {
	// Формируем префикс для поиска файлов
	prefix := fmt.Sprintf("%s_%s/tracks/", meetingID, roomSID)

	// Сканируем все объекты в MinIO
	tracks, err := tc.scanTracks(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to scan tracks: %w", err)
	}

	if len(tracks) == 0 {
		return nil, fmt.Errorf("no tracks found for meeting %s, room %s", meetingID, roomSID)
	}

	// Обрабатываем каждый трек
	var results []CombinedTrack
	for trackID, files := range tracks {
		result := tc.combineTrack(ctx, trackID, files)
		results = append(results, result)
	}

	return results, nil
}

// CombineSingleTrack скачивает и объединяет конкретный трек
func (tc *TrackCombiner) CombineSingleTrack(ctx context.Context, meetingID, roomSID, trackID string) (*CombinedTrack, error) {
	prefix := fmt.Sprintf("%s_%s/tracks/%s", meetingID, roomSID, trackID)

	// Сканируем файлы трека
	tracks, err := tc.scanTracks(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to scan track: %w", err)
	}

	files, ok := tracks[trackID]
	if !ok || len(files) == 0 {
		return nil, fmt.Errorf("track %s not found", trackID)
	}

	result := tc.combineTrack(ctx, trackID, files)
	if result.Error != nil {
		return nil, result.Error
	}

	return &result, nil
}

// scanTracks сканирует MinIO и группирует файлы по трекам
func (tc *TrackCombiner) scanTracks(ctx context.Context, prefix string) (map[string][]string, error) {
	tracks := make(map[string][]string)

	objectCh := tc.minioClient.ListObjects(ctx, tc.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}

		// Извлекаем ID трека из пути
		// Формат: meetingID_roomSID/tracks/TR_xxxxx/file или meetingID_roomSID/tracks/TR_xxxxx.m3u8
		parts := strings.Split(object.Key, "/")
		if len(parts) < 3 {
			continue
		}

		fileName := parts[len(parts)-1]

		// Определяем trackID
		var trackID string
		if len(parts) >= 4 && strings.HasPrefix(parts[2], "TR_") {
			// Файл внутри директории трека
			trackID = parts[2]
		} else if strings.HasPrefix(fileName, "TR_") && strings.Contains(fileName, ".") {
			// Файл трека напрямую
			trackID = strings.Split(fileName, ".")[0]
			// Убираем суффиксы типа -live
			trackID = strings.Split(trackID, "-")[0]
			trackID = strings.Split(trackID, "_00")[0] // Убираем _00000.ts и т.д.
		} else {
			continue
		}

		tracks[trackID] = append(tracks[trackID], object.Key)
	}

	return tracks, nil
}

// combineTrack обрабатывает один трек: скачивает файлы и объединяет их
func (tc *TrackCombiner) combineTrack(ctx context.Context, trackID string, files []string) CombinedTrack {
	result := CombinedTrack{
		TrackID: trackID,
	}

	// Создаем директорию для трека
	trackDir := filepath.Join(tc.workDir, trackID)
	if err := os.MkdirAll(trackDir, 0755); err != nil {
		result.Error = fmt.Errorf("failed to create track directory: %w", err)
		return result
	}

	// Ищем m3u8 плейлист
	var playlistPath string
	for _, file := range files {
		if strings.HasSuffix(file, ".m3u8") && !strings.Contains(file, "-live.m3u8") {
			playlistPath = file
			break
		}
	}

	if playlistPath == "" {
		result.Error = fmt.Errorf("no m3u8 playlist found for track %s", trackID)
		return result
	}

	// Скачиваем плейлист
	localPlaylist := filepath.Join(trackDir, filepath.Base(playlistPath))
	if err := tc.minioClient.FGetObject(ctx, tc.bucketName, playlistPath, localPlaylist, minio.GetObjectOptions{}); err != nil {
		result.Error = fmt.Errorf("failed to download playlist: %w", err)
		return result
	}

	// Читаем плейлист и скачиваем сегменты
	playlistContent, err := os.ReadFile(localPlaylist)
	if err != nil {
		result.Error = fmt.Errorf("failed to read playlist: %w", err)
		return result
	}

	// Скачиваем все сегменты
	lines := strings.Split(string(playlistContent), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Это сегмент
		segmentPath := filepath.Join(filepath.Dir(playlistPath), line)
		localSegment := filepath.Join(trackDir, line)

		if err := tc.minioClient.FGetObject(ctx, tc.bucketName, segmentPath, localSegment, minio.GetObjectOptions{}); err != nil {
			result.Error = fmt.Errorf("failed to download segment %s: %w", line, err)
			return result
		}
	}

	// Определяем тип трека по префиксу и выбираем расширение
	var outputExt string
	var trackType string
	if strings.HasPrefix(trackID, "TR_AM") {
		// Аудио трек - сохраняем как .webm (содержит Opus аудио)
		outputExt = ".webm"
		trackType = "audio"
	} else if strings.HasPrefix(trackID, "TR_VC") {
		// Видео трек - сохраняем как .mp4
		outputExt = ".mp4"
		trackType = "video"
	} else {
		// Неизвестный тип - пытаемся сохранить как .mp4
		outputExt = ".mp4"
		trackType = "unknown"
	}

	// Объединяем сегменты с помощью ffmpeg
	outputPath := filepath.Join(trackDir, trackID+outputExt)
	if err := tc.combineWithFFmpeg(localPlaylist, outputPath); err != nil {
		result.Error = fmt.Errorf("failed to combine segments: %w", err)
		return result
	}

	// Получаем информацию о результирующем файле
	info, err := os.Stat(outputPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to stat output file: %w", err)
		return result
	}

	result.LocalPath = outputPath
	result.Size = info.Size()
	result.Type = trackType // Используем определенный тип трека

	// Определяем продолжительность
	mediaInfo, err := tc.getMediaInfo(outputPath)
	if err == nil {
		result.Duration = mediaInfo.Duration
		// Перезаписываем тип только если ffprobe смог его определить
		if mediaInfo.Type != "" && mediaInfo.Type != "audio" && trackType == "unknown" {
			result.Type = mediaInfo.Type
		}
	}

	return result
}

// combineWithFFmpeg объединяет HLS сегменты (.ts файлы) в файл используя ffmpeg
func (tc *TrackCombiner) combineWithFFmpeg(playlistPath, outputPath string) error {
	// Определяем целевой формат по расширению выходного файла
	isWebM := strings.HasSuffix(outputPath, ".webm")

	var cmd *exec.Cmd
	if isWebM {
		// Для WebM (аудио треки): конвертируем в Opus
		// Входные .ts файлы могут содержать AAC или Opus, ffmpeg сам определит
		cmd = exec.Command("ffmpeg",
			"-i", playlistPath,
			"-c:a", "libopus",    // Конвертируем аудио в Opus
			"-b:a", "128k",       // Битрейт аудио
			"-vn",                // Без видео (только аудио)
			"-y",                 // Перезаписываем выходной файл
			outputPath,
		)
	} else {
		// Для MP4 (видео треки): копируем потоки H.264 + AAC из .ts файлов
		cmd = exec.Command("ffmpeg",
			"-i", playlistPath,
			"-c", "copy",  // Копируем потоки без перекодирования
			"-y",          // Перезаписываем выходной файл
			outputPath,
		)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %w, output: %s", err, string(output))
	}

	return nil
}

// MediaInfo информация о медиафайле
type MediaInfo struct {
	Duration float64 // Продолжительность в секундах
	Type     string  // "audio" или "video"
}

// getMediaInfo получает информацию о медиафайле используя ffprobe
func (tc *TrackCombiner) getMediaInfo(filePath string) (*MediaInfo, error) {
	// Используем ffprobe для получения информации
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-show_entries", "stream=codec_type",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffprobe error: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	info := &MediaInfo{Type: "audio"}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "video" {
			info.Type = "video"
		} else if duration, err := parseFloat(line); err == nil && duration > 0 {
			info.Duration = duration
		}
	}

	return info, nil
}

// Cleanup удаляет временные файлы трека
func (tc *TrackCombiner) Cleanup(trackID string) error {
	trackDir := filepath.Join(tc.workDir, trackID)
	return os.RemoveAll(trackDir)
}

// CleanupAll удаляет всю рабочую директорию
func (tc *TrackCombiner) CleanupAll() error {
	return os.RemoveAll(tc.workDir)
}

// parseFloat безопасно парсит float64 из строки
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// UploadCombinedTrack загружает объединенный трек обратно в MinIO
func (tc *TrackCombiner) UploadCombinedTrack(ctx context.Context, track *CombinedTrack, meetingID, roomSID string) (string, error) {
	if track.Error != nil {
		return "", fmt.Errorf("cannot upload track with error: %w", track.Error)
	}

	// Определяем расширение и content-type на основе типа трека
	var fileExt, contentType string
	if track.Type == "audio" {
		fileExt = ".webm"
		contentType = "audio/webm"
	} else {
		fileExt = ".mp4"
		contentType = "video/mp4"
	}

	// Формируем путь в MinIO
	objectName := fmt.Sprintf("%s_%s/tracks/%s%s", meetingID, roomSID, track.TrackID, fileExt)

	// Загружаем файл
	info, err := tc.minioClient.FPutObject(ctx, tc.bucketName, objectName, track.LocalPath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Формируем публичный URL для доступа к файлу
	protocol := "http"
	if tc.useSSL {
		protocol = "https"
	}
	publicURL := fmt.Sprintf("%s://%s/%s/%s", protocol, tc.publicEndpoint, tc.bucketName, objectName)

	// Логируем размер загруженного файла
	_ = info.Size

	// Примечание: Если нужен presigned URL (для приватных buckets), можно использовать:
	// url, err := tc.minioClient.PresignedGetObject(ctx, tc.bucketName, objectName, 7*24*time.Hour, nil)
	// if err != nil {
	//     return "", fmt.Errorf("failed to generate URL: %w", err)
	// }
	// return url.String(), nil

	return publicURL, nil
}

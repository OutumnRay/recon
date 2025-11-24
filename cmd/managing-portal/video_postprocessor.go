package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"Recontext.online/pkg/database"
	"Recontext.online/pkg/video"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VideoPostProcessor обрабатывает видео после завершения транскрибации
type VideoPostProcessor struct {
	db              *database.DB
	videoProcessor  *video.VideoProcessor
	storageUploader *video.StorageUploader
	workDir         string
}

// NewVideoPostProcessor создает новый пост-процессор видео
func NewVideoPostProcessor(db *database.DB) (*VideoPostProcessor, error) {
	// Создаем процессор видео
	videoProcessor, err := video.NewVideoProcessor(video.DefaultMergeConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create video processor: %w", err)
	}

	// Создаем загрузчик в S3
	storageUploader, err := video.NewStorageUploaderFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to create storage uploader: %w", err)
	}

	// Проверяем/создаем бакет
	ctx := context.Background()
	if err := storageUploader.EnsureBucket(ctx); err != nil {
		log.Printf("⚠️ Failed to ensure bucket exists: %v", err)
	}

	// Рабочая директория для временных файлов
	workDir := os.Getenv("VIDEO_WORK_DIR")
	if workDir == "" {
		workDir = filepath.Join(os.TempDir(), "recontext-video-processing")
	}

	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	return &VideoPostProcessor{
		db:              db,
		videoProcessor:  videoProcessor,
		storageUploader: storageUploader,
		workDir:         workDir,
	}, nil
}

// ProcessMeetingVideo обрабатывает видео встречи после завершения всех транскрибаций
func (vpp *VideoPostProcessor) ProcessMeetingVideo(roomSID string) error {
	log.Print("\n" + "==============================================")
	log.Printf("🎬 Starting video post-processing for room: %s", roomSID)
	log.Print("==============================================")

	ctx := context.Background()

	// 1. Проверяем, что все треки транскрибированы
	allTranscribed, err := vpp.checkAllTracksTranscribed(roomSID)
	if err != nil {
		return fmt.Errorf("failed to check transcription status: %w", err)
	}

	if !allTranscribed {
		log.Printf("⏳ Not all tracks transcribed yet, skipping video processing")
		return nil
	}

	// 2. Получаем информацию о треках
	tracks, err := vpp.getTrackInfoForMerge(roomSID)
	if err != nil {
		return fmt.Errorf("failed to get track info: %w", err)
	}

	if len(tracks) == 0 {
		log.Printf("⚠️ No tracks found for room %s", roomSID)
		return nil
	}

	log.Printf("📋 Found %d tracks to merge", len(tracks))

	// 3. Скачиваем треки из S3 (если они там)
	localTracks, err := vpp.downloadTracks(ctx, tracks)
	if err != nil {
		return fmt.Errorf("failed to download tracks: %w", err)
	}
	defer vpp.cleanupLocalTracks(localTracks)

	// 4. Объединяем видео в режиме picture-in-picture
	mergedVideoPath := filepath.Join(vpp.workDir, fmt.Sprintf("merged_%s.mp4", uuid.New().String()))
	if err := vpp.videoProcessor.MergeTracksPiP(localTracks, mergedVideoPath); err != nil {
		return fmt.Errorf("failed to merge videos: %w", err)
	}
	defer os.Remove(mergedVideoPath)

	log.Printf("✅ Video merged successfully: %s", mergedVideoPath)

	// 5. Конвертируем в HLS формат
	hlsDir := filepath.Join(vpp.workDir, fmt.Sprintf("hls_%s", uuid.New().String()))
	playlistPath, err := vpp.videoProcessor.ConvertToHLS(mergedVideoPath, hlsDir)
	if err != nil {
		return fmt.Errorf("failed to convert to HLS: %w", err)
	}
	defer os.RemoveAll(hlsDir)

	log.Printf("✅ HLS conversion completed: %s", playlistPath)

	// 6. Загружаем HLS файлы в S3
	remotePrefix := fmt.Sprintf("meetings/%s/hls", roomSID)
	uploadedURLs, err := vpp.storageUploader.UploadDirectory(ctx, hlsDir, remotePrefix)
	if err != nil {
		return fmt.Errorf("failed to upload HLS to S3: %w", err)
	}

	log.Printf("✅ Uploaded %d files to S3", len(uploadedURLs))

	// 7. Находим URL плейлиста
	playlistURL := ""
	for _, url := range uploadedURLs {
		if filepath.Base(url) == "playlist.m3u8" {
			playlistURL = url
			break
		}
	}

	if playlistURL == "" {
		return fmt.Errorf("playlist URL not found in uploaded files")
	}

	// 8. Обновляем базу данных с URL плейлиста
	if err := vpp.updateMeetingWithPlaylist(roomSID, playlistURL); err != nil {
		return fmt.Errorf("failed to update database: %w", err)
	}

	log.Print("==============================================")
	log.Printf("✅ Video post-processing completed for room: %s", roomSID)
	log.Printf("   Playlist URL: %s", playlistURL)
	log.Print("==============================================\n")

	return nil
}

// checkAllTracksTranscribed проверяет, что все треки комнаты транскрибированы
func (vpp *VideoPostProcessor) checkAllTracksTranscribed(roomSID string) (bool, error) {
	var count int64

	// Считаем треки без завершенной транскрибации
	err := vpp.db.DB.Table("livekit_tracks").
		Where("room_sid = ?", roomSID).
		Where("type = ?", "audio"). // Проверяем только аудио треки
		Where("(transcription_status IS NULL OR transcription_status != ?)", "completed").
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count == 0, nil
}

// getTrackInfoForMerge получает информацию о треках для объединения
func (vpp *VideoPostProcessor) getTrackInfoForMerge(roomSID string) ([]video.TrackInfo, error) {
	type TrackData struct {
		TrackID         string
		ParticipantSID  string
		ParticipantName string
		Type            string
		PlaylistURL     string
		PublishedAt     time.Time
		Duration        float64
	}

	var trackData []TrackData

	// Получаем информацию о треках с плейлистами из egress_recordings
	err := vpp.db.DB.Raw(`
		SELECT
			t.id as track_id,
			t.participant_sid,
			COALESCE(p.name, 'Unknown') as participant_name,
			t.type,
			er.playlist_url,
			t.published_at,
			COALESCE(t.transcription_duration, 0) as duration
		FROM livekit_tracks t
		LEFT JOIN livekit_participants p ON t.participant_sid = p.sid
		LEFT JOIN egress_recordings er ON er.track_sid = t.sid
		WHERE t.room_sid = ?
			AND t.type IN ('audio', 'video')
			AND er.playlist_url IS NOT NULL
			AND er.status = 'ended'
		ORDER BY t.type DESC, t.published_at ASC
	`, roomSID).Scan(&trackData).Error

	if err != nil {
		return nil, err
	}

	// Преобразуем в TrackInfo
	tracks := make([]video.TrackInfo, len(trackData))
	for i, td := range trackData {
		tracks[i] = video.TrackInfo{
			ParticipantID:   td.ParticipantSID,
			ParticipantName: td.ParticipantName,
			TrackID:         td.TrackID,
			Type:            td.Type,
			FilePath:        td.PlaylistURL, // Изначально это URL
			StartTime:       td.PublishedAt,
			Duration:        td.Duration,
		}
	}

	return tracks, nil
}

// downloadTracks скачивает треки из S3 в локальную файловую систему
func (vpp *VideoPostProcessor) downloadTracks(ctx context.Context, tracks []video.TrackInfo) ([]video.TrackInfo, error) {
	log.Printf("📥 Downloading %d tracks from S3...", len(tracks))

	localTracks := make([]video.TrackInfo, len(tracks))

	for i, track := range tracks {
		// Если это уже локальный путь, пропускаем
		if !isURL(track.FilePath) {
			localTracks[i] = track
			continue
		}

		// Скачиваем файл
		localPath := filepath.Join(vpp.workDir, fmt.Sprintf("track_%s.mp4", track.TrackID))

		// Извлекаем удаленный путь из URL
		remotePath := extractRemotePath(track.FilePath)

		if err := vpp.storageUploader.DownloadFile(ctx, remotePath, localPath); err != nil {
			return nil, fmt.Errorf("failed to download track %s: %w", track.TrackID, err)
		}

		track.FilePath = localPath
		localTracks[i] = track
		log.Printf("   ✅ Downloaded: %s", localPath)
	}

	return localTracks, nil
}

// cleanupLocalTracks удаляет скачанные локальные файлы
func (vpp *VideoPostProcessor) cleanupLocalTracks(tracks []video.TrackInfo) {
	for _, track := range tracks {
		if filepath.IsAbs(track.FilePath) {
			os.Remove(track.FilePath)
		}
	}
}

// updateMeetingWithPlaylist обновляет запись встречи с URL плейлиста
func (vpp *VideoPostProcessor) updateMeetingWithPlaylist(roomSID, playlistURL string) error {
	// Находим meeting_id по room_sid
	var meetingID uuid.UUID
	err := vpp.db.DB.Table("livekit_rooms").
		Select("meeting_id").
		Where("sid = ?", roomSID).
		Where("meeting_id IS NOT NULL").
		First(&meetingID).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("⚠️ No meeting found for room %s", roomSID)
			return nil
		}
		return err
	}

	// Обновляем meeting с URL плейлиста
	result := vpp.db.DB.Exec(`
		UPDATE meetings
		SET
			video_playlist_url = ?,
			updated_at = NOW()
		WHERE id = ?
	`, playlistURL, meetingID)

	if result.Error != nil {
		return result.Error
	}

	// Также обновляем egress_recordings для room_composite
	result = vpp.db.DB.Exec(`
		UPDATE egress_recordings
		SET
			playlist_url = ?,
			updated_at = NOW()
		WHERE room_sid = ?
			AND type = 'room_composite'
	`, playlistURL, roomSID)

	if result.Error != nil {
		return result.Error
	}

	log.Printf("✅ Database updated with playlist URL for meeting %s", meetingID)
	return nil
}

// isURL проверяет, является ли строка URL
func isURL(s string) bool {
	return len(s) > 7 && (s[:7] == "http://" || (len(s) > 8 && s[:8] == "https://"))
}

// extractRemotePath извлекает удаленный путь из URL S3
func extractRemotePath(url string) string {
	// Простое извлечение пути после bucket/
	// Например: http://localhost:9000/recordings/meetings/xyz/track.mp4 -> meetings/xyz/track.mp4

	// Ищем "meetings/" в URL
	idx := strings.Index(url, "/meetings/")
	if idx != -1 {
		return url[idx+1:] // +1 чтобы убрать начальный /
	}

	return filepath.Base(url)
}

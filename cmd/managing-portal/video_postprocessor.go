package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"Recontext.online/pkg/audio"
	"Recontext.online/pkg/database"
	"Recontext.online/pkg/notifications"
	"Recontext.online/pkg/summary"
	"Recontext.online/pkg/video"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VideoPostProcessor обрабатывает видео после завершения транскрибации
type VideoPostProcessor struct {
	db                  *database.DB
	videoProcessor      *video.VideoProcessor
	storageUploader     *video.StorageUploader
	workDir             string
	notificationService *notifications.NotificationService
}

// NewVideoPostProcessor создает новый пост-процессор видео
func NewVideoPostProcessor(db *database.DB, notificationService *notifications.NotificationService) (*VideoPostProcessor, error) {
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
		db:                  db,
		videoProcessor:      videoProcessor,
		storageUploader:     storageUploader,
		workDir:             workDir,
		notificationService: notificationService,
	}, nil
}

// ProcessMeetingVideo обрабатывает видео встречи после завершения всех транскрибаций
func (vpp *VideoPostProcessor) ProcessMeetingVideo(roomSID string) error {
	log.Print("\n" + "==============================================")
	log.Printf("🎬 Starting video post-processing for room: %s", roomSID)
	log.Print("==============================================")

	ctx := context.Background()

	// Get meeting ID from room (room name is meeting ID)
	var room database.LiveKitRoom
	if err := vpp.db.DB.Where("sid = ?", roomSID).First(&room).Error; err != nil {
		return fmt.Errorf("failed to find room: %w", err)
	}

	// Get meeting ID from room.MeetingID (not from room.Name)
	meetingID := room.MeetingID
	if meetingID == nil {
		log.Printf("⚠️ Room %s has no MeetingID, composite video will not be linked to a meeting", roomSID)
	}

	// Send notification: composite video processing started
	if vpp.notificationService != nil && meetingID != nil {
		vpp.notificationService.Notify(notifications.NewCompositeVideoStatusNotification(
			roomSID,
			meetingID,
			"processing",
			notifications.EventCompositeVideoStarted,
		))
	}

	// 1. Проверяем, что все треки транскрибированы
	allTranscribed, err := vpp.checkAllTracksTranscribed(roomSID)
	if err != nil {
		// Send failure notification
		if vpp.notificationService != nil && meetingID != nil {
			vpp.notificationService.Notify(notifications.NewCompositeVideoStatusNotification(
				roomSID,
				meetingID,
				"failed",
				notifications.EventCompositeVideoFailed,
			))
		}
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

	// 4. Анализируем аудио треки для определения активных спикеров
	speakerTimeline, err := vpp.analyzeSpeakerActivity(localTracks)
	if err != nil {
		log.Printf("⚠️ Failed to analyze speaker activity: %v (continuing without timeline)", err)
		speakerTimeline = nil
	}

	// Сохраняем временную линию спикеров в базу данных
	if speakerTimeline != nil {
		if err := vpp.saveSpeakerTimeline(roomSID, speakerTimeline); err != nil {
			log.Printf("⚠️ Failed to save speaker timeline: %v", err)
		}
	}

	// 4.5. Генерируем сводки встречи на русском и английском
	if err := vpp.generateMeetingSummary(roomSID); err != nil {
		log.Printf("⚠️ Failed to generate meeting summary: %v (continuing)", err)
	}

	// 5. Объединяем видео в режиме picture-in-picture с динамическим переключением спикера
	mergedVideoPath := filepath.Join(vpp.workDir, fmt.Sprintf("merged_%s.mp4", uuid.New().String()))

	// Конвертируем audio.SpeakerTimeline в video.SpeakerTimeline
	var videoTimeline *video.SpeakerTimeline
	if speakerTimeline != nil {
		videoTimeline = &video.SpeakerTimeline{
			ActiveSpeaker: speakerTimeline.ActiveSpeaker,
		}
	}

	// Используем новый метод с переключением спикера
	if err := vpp.videoProcessor.MergeTracksPiPWithSpeakerSwitch(localTracks, videoTimeline, mergedVideoPath); err != nil {
		return fmt.Errorf("failed to merge videos: %w", err)
	}
	defer os.Remove(mergedVideoPath)

	log.Printf("✅ Video merged successfully: %s", mergedVideoPath)

	// 6. Конвертируем в HLS формат
	hlsDir := filepath.Join(vpp.workDir, fmt.Sprintf("hls_%s", uuid.New().String()))
	playlistPath, err := vpp.videoProcessor.ConvertToHLS(mergedVideoPath, hlsDir)
	if err != nil {
		return fmt.Errorf("failed to convert to HLS: %w", err)
	}
	defer os.RemoveAll(hlsDir)

	log.Printf("✅ HLS conversion completed: %s", playlistPath)

	// 7. Загружаем HLS файлы в S3
	remotePrefix := fmt.Sprintf("meetings/%s/hls", roomSID)
	uploadedURLs, err := vpp.storageUploader.UploadDirectory(ctx, hlsDir, remotePrefix)
	if err != nil {
		return fmt.Errorf("failed to upload HLS to S3: %w", err)
	}

	log.Printf("✅ Uploaded %d files to S3", len(uploadedURLs))

	// 8. Находим URL плейлиста
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

	// 9. Обновляем базу данных с URL плейлиста
	if err := vpp.updateMeetingWithPlaylist(roomSID, playlistURL); err != nil {
		// Send failure notification
		if vpp.notificationService != nil && meetingID != nil {
			vpp.notificationService.Notify(notifications.NewCompositeVideoStatusNotification(
				roomSID,
				meetingID,
				"failed",
				notifications.EventCompositeVideoFailed,
			))
		}
		return fmt.Errorf("failed to update database: %w", err)
	}

	// Send success notification with playlist URL
	if vpp.notificationService != nil && meetingID != nil {
		notification := notifications.NewCompositeVideoStatusNotification(
			roomSID,
			meetingID,
			"completed",
			notifications.EventCompositeVideoCompleted,
		)
		notification.ChangedFields["video_playlist_url"] = playlistURL
		vpp.notificationService.Notify(notification)
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

	// Используем GORM для получения информации о треках с плейлистами
	err := vpp.db.DB.
		Table("livekit_tracks").
		Select(`
			livekit_tracks.id as track_id,
			livekit_tracks.participant_sid,
			COALESCE(livekit_participants.name, 'Unknown') as participant_name,
			livekit_tracks.type,
			livekit_egress_recordings.playlist_url,
			livekit_tracks.published_at,
			COALESCE(livekit_tracks.transcription_duration, 0) as duration
		`).
		Joins("LEFT JOIN livekit_participants ON livekit_tracks.participant_sid = livekit_participants.sid").
		Joins("LEFT JOIN livekit_egress_recordings ON livekit_egress_recordings.track_sid = livekit_tracks.sid").
		Where("livekit_tracks.room_sid = ?", roomSID).
		Where("livekit_tracks.type IN ?", []string{"audio", "video"}).
		Where("livekit_egress_recordings.playlist_url IS NOT NULL AND livekit_egress_recordings.playlist_url != ''").
		Where("livekit_egress_recordings.status = ?", "ended").
		Order("livekit_tracks.type DESC, livekit_tracks.published_at ASC").
		Scan(&trackData).Error

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
	// Находим meeting_id по room_sid используя GORM
	var room struct {
		MeetingID *uuid.UUID
	}
	err := vpp.db.DB.Table("livekit_rooms").
		Select("meeting_id").
		Where("sid = ?", roomSID).
		Where("meeting_id IS NOT NULL").
		First(&room).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("⚠️ No meeting found for room %s", roomSID)
			return nil
		}
		return err
	}

	if room.MeetingID == nil {
		log.Printf("⚠️ No meeting ID found for room %s", roomSID)
		return nil
	}

	meetingID := *room.MeetingID

	// Обновляем meeting с URL плейлиста используя GORM
	result := vpp.db.DB.Table("meetings").
		Where("id = ?", meetingID).
		Updates(map[string]interface{}{
			"video_playlist_url": playlistURL,
			"updated_at":         time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	// Также обновляем livekit_egress_recordings для room_composite используя GORM
	result = vpp.db.DB.Table("livekit_egress_recordings").
		Where("room_sid = ?", roomSID).
		Where("type = ?", "room_composite").
		Updates(map[string]interface{}{
			"playlist_url": playlistURL,
			"updated_at":   time.Now(),
		})

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

// analyzeSpeakerActivity анализирует аудио треки для определения активных спикеров
func (vpp *VideoPostProcessor) analyzeSpeakerActivity(tracks []video.TrackInfo) (*audio.SpeakerTimeline, error) {
	// Создаем анализатор аудио
	analyzer, err := audio.NewAudioAnalyzer()
	if err != nil {
		return nil, fmt.Errorf("failed to create audio analyzer: %w", err)
	}

	// Конвертируем video.TrackInfo в audio.TrackAudioInfo
	// Извлекаем только аудио треки
	var audioTracks []audio.TrackAudioInfo
	for _, track := range tracks {
		if track.Type == "audio" {
			audioTracks = append(audioTracks, audio.TrackAudioInfo{
				ParticipantID: track.ParticipantID,
				FilePath:      track.FilePath,
				StartTime:     track.StartTime,
				Duration:      track.Duration,
			})
		}
	}

	if len(audioTracks) == 0 {
		log.Printf("⚠️ No audio tracks found for speaker analysis")
		return nil, nil
	}

	// Анализируем аудио треки
	timeline, err := analyzer.AnalyzeAudioTracks(audioTracks)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze audio tracks: %w", err)
	}

	return timeline, nil
}

// saveSpeakerTimeline сохраняет временную линию спикеров в базу данных
func (vpp *VideoPostProcessor) saveSpeakerTimeline(roomSID string, timeline *audio.SpeakerTimeline) error {
	// Сериализуем timeline в JSON
	timelineJSON, err := json.Marshal(timeline)
	if err != nil {
		return fmt.Errorf("failed to marshal timeline: %w", err)
	}

	// Находим meeting_id по room_sid
	var meetingID uuid.UUID
	err = vpp.db.DB.Table("livekit_rooms").
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

	// Обновляем meeting с временной линией спикеров используя GORM
	result := vpp.db.DB.Table("meetings").
		Where("id = ?", meetingID).
		Updates(map[string]interface{}{
			"speaker_timeline": string(timelineJSON),
			"updated_at":       time.Now(),
		})

	if result.Error != nil {
		// Если колонка не существует, логируем предупреждение
		if strings.Contains(result.Error.Error(), "speaker_timeline") {
			log.Printf("⚠️ Column 'speaker_timeline' does not exist yet, skipping timeline save")
			log.Printf("   Please add migration: ALTER TABLE meetings ADD COLUMN speaker_timeline JSONB;")
			return nil
		}
		return result.Error
	}

	log.Printf("✅ Speaker timeline saved for meeting %s (%d segments, %d timeline points)",
		meetingID, len(timeline.Segments), len(timeline.ActiveSpeaker))

	return nil
}

// generateMeetingSummary генерирует сводки встречи на русском и английском
func (vpp *VideoPostProcessor) generateMeetingSummary(roomSID string) error {
	log.Print("\n" + "==============================================")
	log.Printf("📝 Generating meeting summary for room: %s", roomSID)
	log.Print("==============================================")

	ctx := context.Background()

	// Получаем информацию о встрече и транскрипции используя GORM
	type MeetingInfo struct {
		MeetingID    uuid.UUID
		MeetingTitle string
	}

	var meetingInfo MeetingInfo
	err := vpp.db.DB.Table("livekit_rooms r").
		Select("m.id as meeting_id, m.title as meeting_title").
		Joins("JOIN meetings m ON r.meeting_id = m.id").
		Where("r.sid = ?", roomSID).
		Where("r.meeting_id IS NOT NULL").
		Scan(&meetingInfo).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("⚠️ No meeting found for room %s", roomSID)
			return nil
		}
		return fmt.Errorf("failed to get meeting info: %w", err)
	}

	// Получаем транскрипции всех треков используя GORM
	type TranscriptionData struct {
		ParticipantName string
		Segments        string // JSON
	}

	var transcriptions []TranscriptionData
	err = vpp.db.DB.Table("livekit_tracks t").
		Select("COALESCE(p.name, 'Unknown') as participant_name, t.transcription_segments as segments").
		Joins("LEFT JOIN livekit_participants p ON t.participant_sid = p.sid").
		Where("t.room_sid = ?", roomSID).
		Where("t.type = ?", "audio").
		Where("t.transcription_segments IS NOT NULL").
		Order("t.published_at ASC").
		Scan(&transcriptions).Error

	if err != nil {
		return fmt.Errorf("failed to get transcriptions: %w", err)
	}

	if len(transcriptions) == 0 {
		log.Printf("⚠️ No transcriptions found for room %s", roomSID)
		return nil
	}

	log.Printf("   Found %d transcription(s) to process", len(transcriptions))

	// Собираем все сегменты транскрипции
	var allSegments []summary.TranscriptSegment

	for _, trans := range transcriptions {
		var segments []struct {
			Start float64 `json:"start"`
			End   float64 `json:"end"`
			Text  string  `json:"text"`
		}

		if err := json.Unmarshal([]byte(trans.Segments), &segments); err != nil {
			log.Printf("⚠️ Failed to parse segments for %s: %v", trans.ParticipantName, err)
			continue
		}

		for _, seg := range segments {
			allSegments = append(allSegments, summary.TranscriptSegment{
				ParticipantName: trans.ParticipantName,
				StartTime:       seg.Start,
				EndTime:         seg.End,
				Text:            seg.Text,
			})
		}
	}

	if len(allSegments) == 0 {
		log.Printf("⚠️ No valid transcript segments found")
		return nil
	}

	// Создаем генератор сводок
	summaryGen, err := summary.NewSummaryGenerator()
	if err != nil {
		return fmt.Errorf("failed to create summary generator: %w", err)
	}

	// Генерируем сводки
	summaries, err := summaryGen.GenerateSummaries(ctx, meetingInfo.MeetingTitle, allSegments)
	if err != nil {
		return fmt.Errorf("failed to generate summaries: %w", err)
	}

	// Сохраняем сводки в базу данных
	if err := vpp.saveSummariesToDatabase(meetingInfo.MeetingID, summaries); err != nil {
		return fmt.Errorf("failed to save summaries: %w", err)
	}

	log.Print("==============================================")
	log.Printf("✅ Meeting summary generated and saved for meeting: %s", meetingInfo.MeetingID)
	log.Print("==============================================\n")

	return nil
}

// saveSummariesToDatabase сохраняет сводки в базу данных
func (vpp *VideoPostProcessor) saveSummariesToDatabase(meetingID uuid.UUID, summaries *summary.MeetingSummary) error {
	// Сериализуем сводки в JSON
	var summaryEN, summaryRU *string

	if summaries.English != nil {
		jsonEN, err := json.Marshal(summaries.English)
		if err != nil {
			return fmt.Errorf("failed to marshal English summary: %w", err)
		}
		summaryENStr := string(jsonEN)
		summaryEN = &summaryENStr
	}

	if summaries.Russian != nil {
		jsonRU, err := json.Marshal(summaries.Russian)
		if err != nil {
			return fmt.Errorf("failed to marshal Russian summary: %w", err)
		}
		summaryRUStr := string(jsonRU)
		summaryRU = &summaryRUStr
	}

	// Обновляем встречу используя GORM
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if summaryEN != nil {
		updates["summary_en"] = *summaryEN
	}
	if summaryRU != nil {
		updates["summary_ru"] = *summaryRU
	}

	result := vpp.db.DB.Table("meetings").
		Where("id = ?", meetingID).
		Updates(updates)

	if result.Error != nil {
		// Если колонки не существуют, логируем предупреждение
		if strings.Contains(result.Error.Error(), "summary_en") || strings.Contains(result.Error.Error(), "summary_ru") {
			log.Printf("⚠️ Summary columns do not exist yet, skipping summary save")
			log.Printf("   Please add migration: ALTER TABLE meetings ADD COLUMN summary_en JSONB, ADD COLUMN summary_ru JSONB;")
			return nil
		}
		return result.Error
	}

	log.Printf("✅ Summaries saved for meeting %s", meetingID)
	return nil
}

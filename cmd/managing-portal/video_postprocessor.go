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
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

// VideoPostProcessor обрабатывает видео после завершения транскрибации
type VideoPostProcessor struct {
	db                  *database.DB
	videoProcessor      *video.VideoProcessor
	storageUploader     *video.StorageUploader
	trackCombiner       *video.TrackCombiner
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

	// Создаем TrackCombiner для объединения HLS треков
	trackCombiner, err := video.NewTrackCombinerFromEnv()
	if err != nil {
		log.Printf("⚠️ Failed to create track combiner: %v", err)
		trackCombiner = nil // Продолжаем работу без TrackCombiner
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
		trackCombiner:       trackCombiner,
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

	// 0. Объединяем HLS треки в MP4 файлы (если TrackCombiner доступен)
	if vpp.trackCombiner != nil && meetingID != nil {
		if err := vpp.combineHLSTracks(ctx, meetingID.String(), roomSID); err != nil {
			log.Printf("⚠️ Failed to combine HLS tracks: %v (continuing with existing tracks)", err)
		}
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

	// 6. Конвертируем в HLS формат с кастомными именами файлов
	hlsDir := filepath.Join(vpp.workDir, fmt.Sprintf("hls_%s", uuid.New().String()))
	playlistPath, err := vpp.videoProcessor.ConvertToHLSWithCustomNames(
		mergedVideoPath,
		hlsDir,
		"composite.m3u8",
		"composite",
	)
	if err != nil {
		return fmt.Errorf("failed to convert to HLS: %w", err)
	}
	defer os.RemoveAll(hlsDir)

	log.Printf("✅ HLS conversion completed: %s", playlistPath)

	// 7. Загружаем HLS файлы в S3 в формате meetingID_roomSID/
	var remotePrefix string
	if meetingID != nil {
		remotePrefix = fmt.Sprintf("%s_%s", meetingID.String(), roomSID)
	} else {
		remotePrefix = fmt.Sprintf("room_%s", roomSID)
	}

	uploadedURLs, err := vpp.storageUploader.UploadDirectory(ctx, hlsDir, remotePrefix)
	if err != nil {
		return fmt.Errorf("failed to upload HLS to S3: %w", err)
	}

	log.Printf("✅ Uploaded %d files to S3", len(uploadedURLs))

	// 8. Находим URL плейлиста composite.m3u8 и извлекаем относительный путь
	playlistURL := ""
	for _, url := range uploadedURLs {
		if filepath.Base(url) == "composite.m3u8" {
			// Извлекаем только относительный путь без хоста
			// Из: http://192.168.5.153:9000/recontext/meetingID_roomSID/composite.m3u8
			// В: meetingID_roomSID/composite.m3u8
			playlistURL = extractRelativePath(url)
			break
		}
	}

	if playlistURL == "" {
		return fmt.Errorf("composite.m3u8 URL not found in uploaded files")
	}

	log.Printf("📍 Composite playlist relative path: %s", playlistURL)

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

	// 10. Удаляем индивидуальные треки из MinIO после успешного создания композитного видео
	if meetingID != nil {
		if err := vpp.deleteIndividualTracks(ctx, meetingID.String(), roomSID); err != nil {
			log.Printf("⚠️ Failed to delete individual tracks: %v (continuing)", err)
		}
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
	// Упрощенная структура для чтения треков
	type LivekitTrack struct {
		ID                     string
		SID                    string
		ParticipantSID         string `gorm:"column:participant_sid"`
		RoomSID                string `gorm:"column:room_sid"`
		Type                   string
		Source                 string
		PublishedAt            time.Time `gorm:"column:published_at"`
		TranscriptionDuration  float64   `gorm:"column:transcription_duration"`
	}

	// 1. Получаем все треки для комнаты
	var tracks []LivekitTrack
	err := vpp.db.DB.Table("livekit_tracks").
		Where("room_sid = ?", roomSID).
		Where("source IN ?", []string{"CAMERA", "MICROPHONE"}).
		Order("type DESC, published_at ASC").
		Find(&tracks).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get tracks: %w", err)
	}

	log.Printf("📊 Found %d tracks in database for room %s", len(tracks), roomSID)

	if len(tracks) == 0 {
		return []video.TrackInfo{}, nil
	}

	// 2. Получаем информацию об egress recordings для этих треков
	trackSIDs := make([]string, len(tracks))
	for i, t := range tracks {
		trackSIDs[i] = t.SID
	}

	type EgressRecording struct {
		TrackSID    string `gorm:"column:track_sid"`
		PlaylistURL string `gorm:"column:playlist_url"`
		Status      string
	}

	var recordings []EgressRecording
	err = vpp.db.DB.Table("livekit_egress_recordings").
		Where("track_sid IN ?", trackSIDs).
		Where("status = ?", "ended").
		Where("playlist_url IS NOT NULL").
		Where("playlist_url != ?", "").
		Find(&recordings).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get egress recordings: %w", err)
	}

	log.Printf("📊 Found %d egress recordings", len(recordings))

	// Создаем map для быстрого поиска playlist_url по track_sid
	playlistMap := make(map[string]string)
	for _, rec := range recordings {
		playlistMap[rec.TrackSID] = rec.PlaylistURL
	}

	// 3. Получаем имена участников
	participantSIDs := make([]string, 0, len(tracks))
	for _, t := range tracks {
		participantSIDs = append(participantSIDs, t.ParticipantSID)
	}

	type Participant struct {
		SID  string
		Name string
	}

	var participants []Participant
	err = vpp.db.DB.Table("livekit_participants").
		Where("sid IN ?", participantSIDs).
		Find(&participants).Error

	if err != nil {
		log.Printf("⚠️ Failed to get participants: %v", err)
		// Не критично, продолжаем
	}

	// Создаем map для быстрого поиска имени по participant_sid
	nameMap := make(map[string]string)
	for _, p := range participants {
		nameMap[p.SID] = p.Name
	}

	// 4. Собираем результат
	result := make([]video.TrackInfo, 0)
	for _, t := range tracks {
		// Проверяем, есть ли playlist_url для этого трека
		playlistURL, hasPlaylist := playlistMap[t.SID]
		if !hasPlaylist || playlistURL == "" {
			log.Printf("⚠️ Track %s (%s) has no playlist_url, skipping", t.ID, t.Type)
			continue
		}

		participantName := nameMap[t.ParticipantSID]
		if participantName == "" {
			participantName = "Unknown"
		}

		result = append(result, video.TrackInfo{
			ParticipantID:   t.ParticipantSID,
			ParticipantName: participantName,
			TrackID:         t.ID,
			Type:            t.Type,
			FilePath:        playlistURL,
			StartTime:       t.PublishedAt,
			Duration:        t.TranscriptionDuration,
		})
	}

	log.Printf("📊 Returning %d tracks with playlist URLs", len(result))

	return result, nil
}

// downloadTracks скачивает треки из S3 в локальную файловую систему
func (vpp *VideoPostProcessor) downloadTracks(ctx context.Context, tracks []video.TrackInfo) ([]video.TrackInfo, error) {
	log.Printf("📥 Downloading %d tracks from S3...", len(tracks))

	localTracks := make([]video.TrackInfo, len(tracks))

	for i, track := range tracks {
		// Проверяем, является ли путь уже абсолютным локальным путем
		if filepath.IsAbs(track.FilePath) {
			// Это уже скачанный локальный файл
			localTracks[i] = track
			continue
		}

		// Определяем расширение файла по типу трека
		var ext string
		if track.Type == "audio" {
			ext = ".webm"
		} else {
			ext = ".mp4"
		}

		// Скачиваем файл из MinIO
		localPath := filepath.Join(vpp.workDir, fmt.Sprintf("track_%s%s", track.TrackID, ext))

		// FilePath теперь содержит относительный путь в MinIO (например: meetingID_roomSID/tracks/TR_xxx.mp4)
		// Если это URL, извлекаем путь, иначе используем как есть
		remotePath := track.FilePath
		if isURL(track.FilePath) {
			remotePath = extractRemotePath(track.FilePath)
		}

		log.Printf("   📥 Downloading %s from MinIO...", remotePath)

		if err := vpp.storageUploader.DownloadFile(ctx, remotePath, localPath); err != nil {
			return nil, fmt.Errorf("failed to download track %s from %s: %w", track.TrackID, remotePath, err)
		}

		track.FilePath = localPath
		localTracks[i] = track
		log.Printf("   ✅ Downloaded: %s", filepath.Base(localPath))
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

// extractRemotePath извлекает удаленный путь из URL S3/MinIO
func extractRemotePath(url string) string {
	// Удаляем протокол и хост
	// Например: http://localhost:9000/bucket/path/to/file.mp4 -> bucket/path/to/file.mp4
	//          http://192.168.5.153:9000/recontext/meetingID_roomSID/tracks/TR_xxx.mp4 -> recontext/meetingID_roomSID/tracks/TR_xxx.mp4

	// Ищем "meetings/" в URL (старый формат)
	idx := strings.Index(url, "/meetings/")
	if idx != -1 {
		return url[idx+1:] // +1 чтобы убрать начальный /
	}

	// Ищем bucket name в URL (новый формат с meetingID_roomSID)
	// Формат: http://host:port/bucket/path
	parts := strings.SplitN(url, "/", 4)
	if len(parts) >= 4 {
		// parts[0] = "http:" или "https:"
		// parts[1] = ""
		// parts[2] = "host:port"
		// parts[3] = "bucket/path/to/file"
		return parts[3]
	}

	// Fallback: возвращаем имя файла
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

// combineHLSTracks объединяет HLS сегменты треков в единые MP4 файлы
func (vpp *VideoPostProcessor) combineHLSTracks(ctx context.Context, meetingID, roomSID string) error {
	log.Print("\n" + "==============================================")
	log.Printf("🎬 Combining HLS tracks for meeting %s, room %s", meetingID, roomSID)
	log.Print("==============================================")

	// Объединяем все треки комнаты
	tracks, err := vpp.trackCombiner.CombineTracksByRoom(ctx, meetingID, roomSID)
	if err != nil {
		return fmt.Errorf("failed to combine tracks: %w", err)
	}

	if len(tracks) == 0 {
		log.Printf("⚠️ No HLS tracks found to combine")
		return nil
	}

	log.Printf("📊 Processing %d HLS tracks", len(tracks))

	// Обрабатываем каждый трек
	successCount := 0
	for i, track := range tracks {
		if track.Error != nil {
			log.Printf("%d. ❌ Track %s: %v", i+1, track.TrackID, track.Error)
			continue
		}

		log.Printf("%d. ✅ Track %s:", i+1, track.TrackID)
		log.Printf("   Type: %s", track.Type)
		log.Printf("   Size: %.2f MB", float64(track.Size)/(1024*1024))
		log.Printf("   Duration: %.2f seconds", track.Duration)
		log.Printf("   Local path: %s", track.LocalPath)

		// Загружаем объединенный трек обратно в MinIO
		url, err := vpp.trackCombiner.UploadCombinedTrack(ctx, &track, meetingID, roomSID)
		if err != nil {
			log.Printf("   ⚠️ Upload failed: %v", err)
			continue
		}

		log.Printf("   📤 Uploaded to: %s", url)

		// Извлекаем относительный путь для сохранения в базу
		relativePath := extractRelativePath(url)

		// Обновляем запись в базе данных
		if err := vpp.updateTrackWithCombinedURL(track.TrackID, roomSID, relativePath, track.Duration); err != nil {
			log.Printf("   ⚠️ Failed to update database: %v", err)
		} else {
			log.Printf("   💾 Database updated (path: %s)", relativePath)
		}

		// Очищаем временные файлы трека
		if err := vpp.trackCombiner.Cleanup(track.TrackID); err != nil {
			log.Printf("   ⚠️ Cleanup failed: %v", err)
		}

		successCount++
	}

	log.Printf("\n✅ Combined %d/%d tracks successfully", successCount, len(tracks))
	return nil
}

// updateTrackWithCombinedURL обновляет запись трека с URL объединенного файла
func (vpp *VideoPostProcessor) updateTrackWithCombinedURL(trackID, roomSID, url string, duration float64) error {
	// Обновляем или создаем запись в livekit_egress_recordings
	var existingRecording struct {
		ID string
	}

	err := vpp.db.DB.Table("livekit_egress_recordings").
		Select("id").
		Where("track_sid = ?", trackID).
		Where("room_sid = ?", roomSID).
		First(&existingRecording).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing recording: %w", err)
	}

	if err == gorm.ErrRecordNotFound {
		// Создаем новую запись
		log.Printf("   💾 Creating new egress_recordings entry for track %s", trackID)

		// Получаем meeting_id из room
		var room struct {
			MeetingID *uuid.UUID
		}
		if err := vpp.db.DB.Table("livekit_rooms").
			Select("meeting_id").
			Where("sid = ?", roomSID).
			First(&room).Error; err != nil {
			return fmt.Errorf("failed to get room: %w", err)
		}

		now := time.Now()
		recording := map[string]interface{}{
			"id":           uuid.New().String(),
			"egress_id":    "EG_" + uuid.New().String()[:12],
			"room_sid":     roomSID,
			"room_name":    roomSID,
			"meeting_id":   room.MeetingID,
			"track_sid":    trackID,
			"type":         "track_composite",
			"status":       "ended",
			"playlist_url": url,
			"audio_only":   false,
			"started_at":   now,
			"ended_at":     now,
			"created_at":   now,
			"updated_at":   now,
		}

		if err := vpp.db.DB.Table("livekit_egress_recordings").Create(recording).Error; err != nil {
			return fmt.Errorf("failed to create recording: %w", err)
		}
	} else {
		// Обновляем существующую запись
		log.Printf("   📝 Updating existing egress_recordings entry for track %s", trackID)

		updates := map[string]interface{}{
			"playlist_url": url,
			"status":       "ended",
			"updated_at":   time.Now(),
		}

		if err := vpp.db.DB.Table("livekit_egress_recordings").
			Where("track_sid = ?", trackID).
			Where("room_sid = ?", roomSID).
			Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update recording: %w", err)
		}
	}

	// Также обновляем livekit_tracks с продолжительностью
	if duration > 0 {
		vpp.db.DB.Table("livekit_tracks").
			Where("sid = ?", trackID).
			Where("room_sid = ?", roomSID).
			Updates(map[string]interface{}{
				"transcription_duration": duration,
				"updated_at":             time.Now(),
			})
	}

	return nil
}

// deleteIndividualTracks удаляет индивидуальные треки участников из MinIO после создания композитного видео
func (vpp *VideoPostProcessor) deleteIndividualTracks(ctx context.Context, meetingID, roomSID string) error {
	log.Printf("🗑️  Deleting individual tracks for meeting %s, room %s", meetingID, roomSID)

	// Получаем список всех треков из базы данных
	var tracks []struct {
		TrackID string
		Type    string
	}

	err := vpp.db.DB.Table("livekit_tracks").
		Select("sid as track_id, type").
		Where("room_sid = ?", roomSID).
		Where("type IN ?", []string{"audio", "video"}).
		Scan(&tracks).Error

	if err != nil {
		return fmt.Errorf("failed to get tracks from database: %w", err)
	}

	if len(tracks) == 0 {
		log.Printf("ℹ️  No tracks found to delete")
		return nil
	}

	log.Printf("📋 Found %d tracks to delete", len(tracks))

	// Получаем доступ к MinIO клиенту через storageUploader
	minioClient := vpp.storageUploader.GetClient()
	bucketName := vpp.storageUploader.GetBucket()

	// Префикс для всех файлов треков
	tracksPrefix := fmt.Sprintf("%s_%s/tracks/", meetingID, roomSID)

	// Перебираем все объекты в папке tracks/
	objectCh := minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    tracksPrefix,
		Recursive: true,
	})

	deletedCount := 0
	for object := range objectCh {
		if object.Err != nil {
			log.Printf("⚠️  Error listing object: %v", object.Err)
			continue
		}

		// Удаляем все файлы треков (.ts, .m3u8, .mp4, .webm)
		ext := filepath.Ext(object.Key)
		if ext == ".ts" || ext == ".m3u8" || ext == ".mp4" || ext == ".webm" {
			err := vpp.storageUploader.DeleteFile(ctx, object.Key)
			if err != nil {
				log.Printf("⚠️  Failed to delete %s: %v", object.Key, err)
				continue
			}
			deletedCount++
			if deletedCount%10 == 0 {
				log.Printf("   Deleted %d files...", deletedCount)
			}
		}
	}

	log.Printf("✅ Deleted %d track files from MinIO", deletedCount)
	return nil
}

// extractRelativePath извлекает относительный путь из полного URL MinIO
// Из: http://192.168.5.153:9000/recontext/meetingID_roomSID/composite.m3u8
// В: meetingID_roomSID/composite.m3u8
func extractRelativePath(fullURL string) string {
	// Убираем протокол и хост
	// Ищем "/recontext/" и берем все что после него
	bucketPrefix := "/recontext/"
	if idx := strings.Index(fullURL, bucketPrefix); idx != -1 {
		return fullURL[idx+len(bucketPrefix):]
	}

	// Если не нашли /recontext/, пробуем другой формат
	// Возможно URL уже в виде meetingID_roomSID/file.m3u8
	parts := strings.Split(fullURL, "/")
	if len(parts) >= 2 {
		// Берем последние 2 части: meetingID_roomSID/composite.m3u8
		return strings.Join(parts[len(parts)-2:], "/")
	}

	return fullURL
}

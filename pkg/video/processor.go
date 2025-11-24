package video

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TrackInfo представляет информацию о видео/аудио треке участника
type TrackInfo struct {
	ParticipantID   string
	ParticipantName string
	TrackID         string
	Type            string // "audio" or "video"
	FilePath        string // Путь к файлу записи трека
	StartTime       time.Time
	Duration        float64
}

// MergeConfig конфигурация для объединения видео
type MergeConfig struct {
	OutputPath      string
	Layout          string // "pip" (picture-in-picture), "grid", "speaker-switch"
	Width           int
	Height          int
	Framerate       int
	AudioBitrate    string
	VideoBitrate    string
	MainVideoWidth  int // Для PiP: размер основного видео
	MainVideoHeight int
	PipVideoWidth   int // Для PiP: размер маленьких видео
	PipVideoHeight  int
	PipPadding      int // Отступ маленьких видео от краев
}

// DefaultMergeConfig возвращает конфигурацию по умолчанию
func DefaultMergeConfig() *MergeConfig {
	return &MergeConfig{
		Layout:          "pip",
		Width:           1920,
		Height:          1080,
		Framerate:       30,
		AudioBitrate:    "128k",
		VideoBitrate:    "2M",
		MainVideoWidth:  1920,
		MainVideoHeight: 1080,
		PipVideoWidth:   320,
		PipVideoHeight:  180,
		PipPadding:      20,
	}
}

// VideoProcessor обрабатывает видео с помощью FFmpeg
type VideoProcessor struct {
	config      *MergeConfig
	ffmpegPath  string
	ffprobePath string
}

// NewVideoProcessor создает новый процессор видео
func NewVideoProcessor(config *MergeConfig) (*VideoProcessor, error) {
	if config == nil {
		config = DefaultMergeConfig()
	}

	// Проверяем наличие FFmpeg
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, fmt.Errorf("ffmpeg not found in PATH: %w", err)
	}

	ffprobePath, err := exec.LookPath("ffprobe")
	if err != nil {
		log.Printf("⚠️ ffprobe not found, some features may be limited")
		ffprobePath = ""
	}

	return &VideoProcessor{
		config:      config,
		ffmpegPath:  ffmpegPath,
		ffprobePath: ffprobePath,
	}, nil
}

// MergeTracksPiP объединяет треки в режиме picture-in-picture
// Активный спикер показывается большим, остальные участники - маленькими в углах
func (vp *VideoProcessor) MergeTracksPiP(tracks []TrackInfo, outputPath string) error {
	if len(tracks) == 0 {
		return fmt.Errorf("no tracks to merge")
	}

	log.Printf("🎬 Starting picture-in-picture merge of %d tracks", len(tracks))
	log.Printf("   Output: %s", outputPath)

	// Группируем треки по участникам
	participantTracks := vp.groupTracksByParticipant(tracks)

	// Находим видео и аудио треки
	var videoTracks []TrackInfo
	var audioTracks []TrackInfo

	for _, trackList := range participantTracks {
		for _, track := range trackList {
			if track.Type == "video" {
				videoTracks = append(videoTracks, track)
			} else if track.Type == "audio" {
				audioTracks = append(audioTracks, track)
			}
		}
	}

	if len(videoTracks) == 0 {
		return fmt.Errorf("no video tracks found")
	}

	log.Printf("   Video tracks: %d", len(videoTracks))
	log.Printf("   Audio tracks: %d", len(audioTracks))

	// Строим команду FFmpeg для PiP
	args := []string{
		"-y", // Перезаписываем выходной файл
	}

	// Добавляем входные файлы
	for _, track := range videoTracks {
		args = append(args, "-i", track.FilePath)
	}
	for _, track := range audioTracks {
		args = append(args, "-i", track.FilePath)
	}

	// Строим фильтр для PiP
	filterComplex := vp.buildPiPFilter(videoTracks)
	args = append(args, "-filter_complex", filterComplex)

	// Микшируем все аудио треки
	if len(audioTracks) > 0 {
		audioMix := vp.buildAudioMixFilter(len(videoTracks), len(audioTracks))
		args = append(args, "-filter_complex", audioMix)
		args = append(args, "-map", "[aout]")
	}

	// Маппинг видео
	args = append(args, "-map", "[vout]")

	// Параметры кодирования
	args = append(args,
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", "23",
		"-c:a", "aac",
		"-b:a", vp.config.AudioBitrate,
		"-r", fmt.Sprintf("%d", vp.config.Framerate),
		outputPath,
	)

	log.Printf("🎬 Running FFmpeg command...")
	log.Printf("   Command: ffmpeg %s", strings.Join(args, " "))

	cmd := exec.Command(vp.ffmpegPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w", err)
	}

	log.Printf("✅ Video merge completed: %s", outputPath)
	return nil
}

// groupTracksByParticipant группирует треки по участникам
func (vp *VideoProcessor) groupTracksByParticipant(tracks []TrackInfo) map[string][]TrackInfo {
	result := make(map[string][]TrackInfo)
	for _, track := range tracks {
		result[track.ParticipantID] = append(result[track.ParticipantID], track)
	}
	return result
}

// buildPiPFilter строит фильтр FFmpeg для picture-in-picture
func (vp *VideoProcessor) buildPiPFilter(videoTracks []TrackInfo) string {
	if len(videoTracks) == 0 {
		return ""
	}

	if len(videoTracks) == 1 {
		// Только одно видео - просто масштабируем
		return fmt.Sprintf("[0:v]scale=%d:%d[vout]", vp.config.Width, vp.config.Height)
	}

	// Основное видео (первое) масштабируется на весь экран
	filter := fmt.Sprintf("[0:v]scale=%d:%d[main];", vp.config.Width, vp.config.Height)

	// Остальные видео масштабируются и накладываются в углы
	currentBase := "[main]"
	positions := []string{
		fmt.Sprintf("x=%d:y=%d", vp.config.Width-vp.config.PipVideoWidth-vp.config.PipPadding, vp.config.PipPadding),                                                           // Правый верхний
		fmt.Sprintf("x=%d:y=%d", vp.config.Width-vp.config.PipVideoWidth-vp.config.PipPadding, vp.config.Height-vp.config.PipVideoHeight-vp.config.PipPadding),                // Правый нижний
		fmt.Sprintf("x=%d:y=%d", vp.config.PipPadding, vp.config.Height-vp.config.PipVideoHeight-vp.config.PipPadding),                                                         // Левый нижний
		fmt.Sprintf("x=%d:y=%d", vp.config.PipPadding, vp.config.PipPadding),                                                                                                   // Левый верхний
		fmt.Sprintf("x=%d:y=%d", (vp.config.Width-vp.config.PipVideoWidth)/2, vp.config.PipPadding),                                                                            // Центр верхний
		fmt.Sprintf("x=%d:y=%d", (vp.config.Width-vp.config.PipVideoWidth)/2, vp.config.Height-vp.config.PipVideoHeight-vp.config.PipPadding),                                 // Центр нижний
		fmt.Sprintf("x=%d:y=%d", vp.config.PipPadding, (vp.config.Height-vp.config.PipVideoHeight)/2),                                                                          // Левый центр
		fmt.Sprintf("x=%d:y=%d", vp.config.Width-vp.config.PipVideoWidth-vp.config.PipPadding, (vp.config.Height-vp.config.PipVideoHeight)/2),                                 // Правый центр
	}

	for i := 1; i < len(videoTracks) && i < 9; i++ {
		pipLabel := fmt.Sprintf("pip%d", i)
		outLabel := fmt.Sprintf("out%d", i)

		// Масштабируем маленькое видео
		filter += fmt.Sprintf("[%d:v]scale=%d:%d[%s];", i, vp.config.PipVideoWidth, vp.config.PipVideoHeight, pipLabel)

		// Накладываем на текущую базу
		posIndex := i - 1
		if posIndex >= len(positions) {
			posIndex = len(positions) - 1
		}
		filter += fmt.Sprintf("%s[%s]overlay=%s[%s];", currentBase, pipLabel, positions[posIndex], outLabel)
		currentBase = fmt.Sprintf("[%s]", outLabel)
	}

	// Последний выход переименовываем в vout
	filter = strings.TrimSuffix(filter, ";")
	filter = strings.ReplaceAll(filter, currentBase, "[vout]")

	return filter
}

// buildAudioMixFilter строит фильтр для микширования аудио
func (vp *VideoProcessor) buildAudioMixFilter(videoCount, audioCount int) string {
	if audioCount == 0 {
		return ""
	}

	if audioCount == 1 {
		return fmt.Sprintf("[%d:a]anull[aout]", videoCount)
	}

	// Микшируем все аудио треки
	inputs := make([]string, audioCount)
	for i := 0; i < audioCount; i++ {
		inputs[i] = fmt.Sprintf("[%d:a]", videoCount+i)
	}

	return fmt.Sprintf("%samix=inputs=%d:duration=longest[aout]", strings.Join(inputs, ""), audioCount)
}

// ConvertToHLS конвертирует видео в HLS формат (m3u8 + segments)
func (vp *VideoProcessor) ConvertToHLS(inputPath, outputDir string) (string, error) {
	log.Printf("🎬 Converting video to HLS format")
	log.Printf("   Input: %s", inputPath)
	log.Printf("   Output dir: %s", outputDir)

	// Создаем директорию если не существует
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	playlistFile := filepath.Join(outputDir, "playlist.m3u8")

	args := []string{
		"-i", inputPath,
		"-c:v", "libx264",
		"-c:a", "aac",
		"-b:a", "128k",
		"-hls_time", "10", // Длина каждого сегмента: 10 секунд
		"-hls_list_size", "0", // Включить все сегменты в плейлист
		"-hls_segment_filename", filepath.Join(outputDir, "segment_%03d.ts"),
		"-f", "hls",
		playlistFile,
	}

	log.Printf("🎬 Running FFmpeg HLS conversion...")
	cmd := exec.Command(vp.ffmpegPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg HLS conversion failed: %w", err)
	}

	log.Printf("✅ HLS conversion completed: %s", playlistFile)
	return playlistFile, nil
}

// GetVideoDuration получает длительность видео в секундах
func (vp *VideoProcessor) GetVideoDuration(filePath string) (float64, error) {
	if vp.ffprobePath == "" {
		return 0, fmt.Errorf("ffprobe not available")
	}

	args := []string{
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	}

	cmd := exec.Command(vp.ffprobePath, args...)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}

	var duration float64
	if _, err := fmt.Sscanf(string(output), "%f", &duration); err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}

// GenerateTempPath генерирует временный путь для файла
func GenerateTempPath(prefix, ext string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s_%s%s", prefix, uuid.New().String(), ext))
}

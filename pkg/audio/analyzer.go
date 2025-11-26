package audio

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// SpeechSegment представляет сегмент речи участника
type SpeechSegment struct {
	ParticipantID string    `json:"participant_id"`
	StartTime     float64   `json:"start_time"`     // Секунды от начала
	EndTime       float64   `json:"end_time"`       // Секунды от начала
	Energy        float64   `json:"energy"`         // Уровень громкости
	Duration      float64   `json:"duration"`       // Длительность сегмента
	Timestamp     time.Time `json:"timestamp"`      // Абсолютное время
}

// SpeakerTimeline представляет временную линию активных спикеров
type SpeakerTimeline struct {
	Segments      []SpeechSegment    `json:"segments"`
	ActiveSpeaker map[string]string  `json:"active_speaker"` // Время (строка) -> ParticipantID
}

// AudioAnalyzer анализирует аудио треки для определения активных спикеров
type AudioAnalyzer struct {
	ffmpegPath  string
	ffprobePath string
}

// NewAudioAnalyzer создает новый анализатор аудио
func NewAudioAnalyzer() (*AudioAnalyzer, error) {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, fmt.Errorf("ffmpeg not found: %w", err)
	}

	ffprobePath, err := exec.LookPath("ffprobe")
	if err != nil {
		log.Printf("⚠️ ffprobe not found, some features may be limited")
		ffprobePath = ""
	}

	return &AudioAnalyzer{
		ffmpegPath:  ffmpegPath,
		ffprobePath: ffprobePath,
	}, nil
}

// TrackAudioInfo информация об аудио треке для анализа
type TrackAudioInfo struct {
	ParticipantID string
	FilePath      string
	StartTime     time.Time
	Duration      float64
}

// AnalyzeAudioTracks анализирует аудио треки и определяет активных спикеров
// Возвращает временную линию активных спикеров с защитой от мерцания
func (aa *AudioAnalyzer) AnalyzeAudioTracks(tracks []TrackAudioInfo) (*SpeakerTimeline, error) {
	log.Printf("🎙️ Starting audio analysis for %d tracks", len(tracks))

	timeline := &SpeakerTimeline{
		Segments:      []SpeechSegment{},
		ActiveSpeaker: make(map[string]string),
	}

	// Анализируем каждый трек
	for _, track := range tracks {
		segments, err := aa.detectSpeechSegments(track)
		if err != nil {
			log.Printf("⚠️ Failed to analyze track %s: %v", track.ParticipantID, err)
			continue
		}
		timeline.Segments = append(timeline.Segments, segments...)
	}

	// Строим временную линию активных спикеров с защитой от мерцания
	timeline.ActiveSpeaker = aa.buildActiveSpeakerTimeline(timeline.Segments)

	log.Printf("✅ Audio analysis complete: %d speech segments, %d timeline points",
		len(timeline.Segments), len(timeline.ActiveSpeaker))

	return timeline, nil
}

// detectSpeechSegments определяет сегменты речи в аудио треке
func (aa *AudioAnalyzer) detectSpeechSegments(track TrackAudioInfo) ([]SpeechSegment, error) {
	log.Printf("   Analyzing track for participant: %s", track.ParticipantID)

	// Используем FFmpeg silencedetect filter для определения речевых сегментов
	// Параметры: -30dB шумовой порог, 0.5 секунд минимальной длительности тишины
	args := []string{
		"-i", track.FilePath,
		"-af", "silencedetect=noise=-30dB:d=0.5",
		"-f", "null",
		"-",
	}

	cmd := exec.Command(aa.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// FFmpeg возвращает ошибку, но output содержит результаты
		// Проверяем, есть ли полезная информация в output
		if len(output) == 0 {
			return nil, fmt.Errorf("ffmpeg failed: %w", err)
		}
	}

	// Парсим вывод silencedetect
	segments := aa.parseSilenceDetectOutput(string(output), track.ParticipantID)

	// Добавляем информацию о времени
	for i := range segments {
		segments[i].Timestamp = track.StartTime.Add(time.Duration(segments[i].StartTime * float64(time.Second)))
	}

	log.Printf("   Found %d speech segments for %s", len(segments), track.ParticipantID)
	return segments, nil
}

// parseSilenceDetectOutput парсит вывод FFmpeg silencedetect filter
func (aa *AudioAnalyzer) parseSilenceDetectOutput(output, participantID string) []SpeechSegment {
	lines := strings.Split(output, "\n")

	var silenceStart, silenceEnd float64
	var segments []SpeechSegment
	var lastSpeechEnd float64

	for _, line := range lines {
		// Ищем строки типа: [silencedetect @ ...] silence_start: 1.234
		if strings.Contains(line, "silence_start:") {
			parts := strings.Split(line, "silence_start:")
			if len(parts) > 1 {
				silenceStart = parseFloat(strings.TrimSpace(parts[1]))

				// Если есть речь между предыдущей тишиной и текущей
				if silenceStart > lastSpeechEnd {
					segment := SpeechSegment{
						ParticipantID: participantID,
						StartTime:     lastSpeechEnd,
						EndTime:       silenceStart,
						Duration:      silenceStart - lastSpeechEnd,
						Energy:        1.0, // Упрощенно, в реальности нужен анализ энергии
					}
					segments = append(segments, segment)
				}
			}
		}

		// Ищем строки типа: [silencedetect @ ...] silence_end: 2.345 | silence_duration: 1.111
		if strings.Contains(line, "silence_end:") {
			parts := strings.Split(line, "silence_end:")
			if len(parts) > 1 {
				endPart := strings.TrimSpace(parts[1])
				// Убираем часть с silence_duration
				if idx := strings.Index(endPart, "|"); idx > 0 {
					endPart = endPart[:idx]
				}
				silenceEnd = parseFloat(strings.TrimSpace(endPart))
				lastSpeechEnd = silenceEnd
			}
		}
	}

	return segments
}

// buildActiveSpeakerTimeline строит временную линию с защитой от мерцания
// Применяет правило: смена спикера только если новый спикер говорит минимум 2 секунды
func (aa *AudioAnalyzer) buildActiveSpeakerTimeline(segments []SpeechSegment) map[string]string {
	const MIN_SPEAKER_DURATION = 2.0 // Минимальная длительность для смены спикера (секунды)

	timeline := make(map[string]string)

	if len(segments) == 0 {
		return timeline
	}

	// Сортируем сегменты по времени начала
	sortedSegments := make([]SpeechSegment, len(segments))
	copy(sortedSegments, segments)

	// Простая сортировка по StartTime
	for i := 0; i < len(sortedSegments)-1; i++ {
		for j := i + 1; j < len(sortedSegments); j++ {
			if sortedSegments[i].StartTime > sortedSegments[j].StartTime {
				sortedSegments[i], sortedSegments[j] = sortedSegments[j], sortedSegments[i]
			}
		}
	}

	// Инициализируем текущего спикера
	currentSpeaker := ""
	lastChangeTime := 0.0

	// Проходим по всем сегментам и определяем активного спикера
	for _, segment := range sortedSegments {
		// Находим все активные сегменты в текущий момент времени
		activeSegments := aa.getActiveSegmentsAt(sortedSegments, segment.StartTime)

		if len(activeSegments) == 0 {
			continue
		}

		// Находим самый громкий сегмент (с максимальной энергией)
		loudestSegment := activeSegments[0]
		for _, seg := range activeSegments {
			if seg.Energy > loudestSegment.Energy {
				loudestSegment = seg
			}
		}

		candidateSpeaker := loudestSegment.ParticipantID

		// Проверяем защиту от мерцания
		if currentSpeaker == "" {
			// Первый спикер
			currentSpeaker = candidateSpeaker
			timeKey := strconv.FormatFloat(segment.StartTime, 'f', 2, 64)
			timeline[timeKey] = currentSpeaker
			lastChangeTime = segment.StartTime
		} else if candidateSpeaker != currentSpeaker {
			// Потенциальная смена спикера
			timeSinceLastChange := segment.StartTime - lastChangeTime

			// Разрешаем смену только если прошло минимум MIN_SPEAKER_DURATION секунд
			if timeSinceLastChange >= MIN_SPEAKER_DURATION {
				// Проверяем, что новый кандидат будет говорить достаточно долго
				candidateDuration := aa.getTotalDurationForSpeaker(sortedSegments, candidateSpeaker, segment.StartTime, segment.StartTime+MIN_SPEAKER_DURATION)

				if candidateDuration >= MIN_SPEAKER_DURATION * 0.7 { // 70% от минимальной длительности
					currentSpeaker = candidateSpeaker
					timeKey := strconv.FormatFloat(segment.StartTime, 'f', 2, 64)
					timeline[timeKey] = currentSpeaker
					lastChangeTime = segment.StartTime

					log.Printf("   Speaker change at %.2fs: %s", segment.StartTime, currentSpeaker)
				}
			}
		}
	}

	return timeline
}

// getActiveSegmentsAt возвращает все активные сегменты в указанное время
func (aa *AudioAnalyzer) getActiveSegmentsAt(segments []SpeechSegment, time float64) []SpeechSegment {
	var active []SpeechSegment

	for _, seg := range segments {
		if seg.StartTime <= time && seg.EndTime >= time {
			active = append(active, seg)
		}
	}

	return active
}

// getTotalDurationForSpeaker вычисляет общую длительность речи спикера в интервале
func (aa *AudioAnalyzer) getTotalDurationForSpeaker(segments []SpeechSegment, participantID string, startTime, endTime float64) float64 {
	var totalDuration float64

	for _, seg := range segments {
		if seg.ParticipantID != participantID {
			continue
		}

		// Проверяем пересечение с интервалом
		segStart := seg.StartTime
		segEnd := seg.EndTime

		if segEnd < startTime || segStart > endTime {
			continue
		}

		// Вычисляем пересечение
		overlapStart := max(segStart, startTime)
		overlapEnd := min(segEnd, endTime)

		if overlapEnd > overlapStart {
			totalDuration += overlapEnd - overlapStart
		}
	}

	return totalDuration
}

// ExportTimelineToJSON экспортирует временную линию в JSON
func (aa *AudioAnalyzer) ExportTimelineToJSON(timeline *SpeakerTimeline) ([]byte, error) {
	return json.MarshalIndent(timeline, "", "  ")
}

// parseFloat безопасно парсит float из строки
func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}

// max возвращает максимум из двух float64
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// min возвращает минимум из двух float64
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

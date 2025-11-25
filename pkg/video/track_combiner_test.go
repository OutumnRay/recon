package video

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestTrackCombiner_ScanTracks(t *testing.T) {
	// Пример теста для scanTracks
	// В реальной ситуации нужно использовать mock MinIO client

	config := TrackCombinerConfig{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		BucketName:      "test-bucket",
		UseSSL:          false,
		WorkDir:         filepath.Join(os.TempDir(), "track-combiner-test"),
	}

	combiner, err := NewTrackCombiner(config)
	if err != nil {
		t.Skipf("Skipping test: MinIO not available: %v", err)
		return
	}
	defer combiner.CleanupAll()

	ctx := context.Background()

	// Тестируем сканирование треков
	tracks, err := combiner.scanTracks(ctx, "test-prefix/")
	if err != nil {
		t.Errorf("scanTracks failed: %v", err)
	}

	t.Logf("Found %d tracks", len(tracks))
}

func TestTrackCombiner_ParseFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"45.5", 45.5, false},
		{"0.0", 0.0, false},
		{"123.456", 123.456, false},
		{"invalid", 0.0, true},
		{"", 0.0, true},
	}

	for _, tt := range tests {
		result, err := parseFloat(tt.input)
		hasError := err != nil

		if hasError != tt.hasError {
			t.Errorf("parseFloat(%q): expected error=%v, got error=%v", tt.input, tt.hasError, hasError)
		}

		if !hasError && result != tt.expected {
			t.Errorf("parseFloat(%q): expected %f, got %f", tt.input, tt.expected, result)
		}
	}
}

func TestTrackCombiner_GetMediaInfo(t *testing.T) {
	// Этот тест требует реальный медиа файл и установленный ffprobe
	t.Skip("Requires real media file and ffprobe installed")

	config := TrackCombinerConfig{
		WorkDir: filepath.Join(os.TempDir(), "track-combiner-test"),
	}

	combiner := &TrackCombiner{
		workDir: config.WorkDir,
	}

	info, err := combiner.getMediaInfo("/path/to/test/video.mp4")
	if err != nil {
		t.Errorf("getMediaInfo failed: %v", err)
	}

	if info.Duration <= 0 {
		t.Errorf("Expected positive duration, got %f", info.Duration)
	}

	if info.Type != "video" && info.Type != "audio" {
		t.Errorf("Expected type 'video' or 'audio', got %q", info.Type)
	}
}

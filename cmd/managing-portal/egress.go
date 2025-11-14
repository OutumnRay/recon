package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/livekit/protocol/livekit"
)

// EgressConfig содержит настройки для запуска Egress
type EgressConfig struct {
	Enabled        bool
	S3Endpoint     string
	S3Bucket       string
	S3AccessKey    string
	S3Secret       string
	S3Region       string
	RecordTracks   bool
}

// loadEgressConfig загружает конфигурацию Egress из переменных окружения
func loadEgressConfig() EgressConfig {
	enabled, _ := strconv.ParseBool(os.Getenv("LIVEKIT_EGRESS_ENABLED"))
	recordTracks, _ := strconv.ParseBool(os.Getenv("LIVEKIT_EGRESS_RECORD_TRACKS"))

	return EgressConfig{
		Enabled:      enabled,
		S3Endpoint:   os.Getenv("LIVEKIT_EGRESS_S3_ENDPOINT"),
		S3Bucket:     os.Getenv("LIVEKIT_EGRESS_S3_BUCKET"),
		S3AccessKey:  os.Getenv("LIVEKIT_EGRESS_S3_ACCESS_KEY"),
		S3Secret:     os.Getenv("LIVEKIT_EGRESS_S3_SECRET"),
		S3Region:     os.Getenv("LIVEKIT_EGRESS_S3_REGION"),
		RecordTracks: recordTracks,
	}
}

// startRoomCompositeEgress запускает запись всей комнаты
func (mp *ManagingPortal) startRoomCompositeEgress(roomName string, roomSID string, audioOnly bool) (string, error) {
	config := loadEgressConfig()

	if !config.Enabled {
		return "", nil // Egress отключен
	}

	// Создаем Egress клиент
	egressClient := lksdk.NewEgressClient(
		os.Getenv("LIVEKIT_URL"),
		os.Getenv("LIVEKIT_API_KEY"),
		os.Getenv("LIVEKIT_API_SECRET"),
	)

	// Формируем запрос на запись комнаты
	req := &livekit.RoomCompositeEgressRequest{
		RoomName:  roomName,
		Layout:    "speaker",
		AudioOnly: audioOnly,
	}

	// Настройки кодирования
	if audioOnly {
		req.Options = &livekit.RoomCompositeEgressRequest_Preset{
			Preset: livekit.EncodingOptionsPreset_H264_720P_30,
		}
	} else {
		req.Options = &livekit.RoomCompositeEgressRequest_Preset{
			Preset: livekit.EncodingOptionsPreset_H264_720P_30,
		}
	}

	// Настройки сегментированного вывода в S3
	// Используем структуру: {meetingID}_{sessionID} - LiveKit сам добавит /composite
	req.SegmentOutputs = []*livekit.SegmentedFileOutput{
		{
			FilenamePrefix:   fmt.Sprintf("%s_%s", roomName, roomSID),
			PlaylistName:     "composite.m3u8",
			LivePlaylistName: "composite-live.m3u8",
			SegmentDuration:  10,
			Output: &livekit.SegmentedFileOutput_S3{
				S3: &livekit.S3Upload{
					AccessKey:      config.S3AccessKey,
					Secret:         config.S3Secret,
					Endpoint:       config.S3Endpoint,
					Bucket:         config.S3Bucket,
					Region:         config.S3Region,
					ForcePathStyle: true,
				},
			},
		},
	}

	// Запускаем запись
	res, err := egressClient.StartRoomCompositeEgress(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("failed to start room composite egress: %w", err)
	}

	mp.logger.Infof("Started room composite egress for room %s (SID: %s): EgressID=%s", roomName, roomSID, res.EgressId)
	return res.EgressId, nil
}

// startTrackCompositeEgress запускает запись отдельного аудио трека
func (mp *ManagingPortal) startTrackCompositeEgress(roomName, roomSID, trackID string) (string, error) {
	config := loadEgressConfig()

	if !config.Enabled || !config.RecordTracks {
		return "", nil // Запись треков отключена
	}

	// Создаем Egress клиент
	egressClient := lksdk.NewEgressClient(
		os.Getenv("LIVEKIT_URL"),
		os.Getenv("LIVEKIT_API_KEY"),
		os.Getenv("LIVEKIT_API_SECRET"),
	)

	// Формируем запрос на запись трека
	req := &livekit.TrackCompositeEgressRequest{
		RoomName:     roomName,
		AudioTrackId: trackID,
		VideoTrackId: "",
	}

	// Настройки кодирования
	req.Options = &livekit.TrackCompositeEgressRequest_Preset{
		Preset: livekit.EncodingOptionsPreset_H264_720P_30,
	}

	// Настройки сегментированного вывода в S3
	// Используем структуру: {meetingID}_{sessionID} - LiveKit сам добавит /tracks/{trackID}
	req.SegmentOutputs = []*livekit.SegmentedFileOutput{
		{
			FilenamePrefix:   fmt.Sprintf("%s_%s", roomName, roomSID),
			PlaylistName:     fmt.Sprintf("%s.m3u8", trackID),
			LivePlaylistName: fmt.Sprintf("%s-live.m3u8", trackID),
			SegmentDuration:  20,
			Output: &livekit.SegmentedFileOutput_S3{
				S3: &livekit.S3Upload{
					AccessKey:      config.S3AccessKey,
					Secret:         config.S3Secret,
					Endpoint:       config.S3Endpoint,
					Bucket:         config.S3Bucket,
					Region:         config.S3Region,
					ForcePathStyle: true,
				},
			},
		},
	}

	// Запускаем запись
	res, err := egressClient.StartTrackCompositeEgress(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("failed to start track composite egress: %w", err)
	}

	mp.logger.Infof("Started track composite egress for track %s in room %s (SID: %s): EgressID=%s", trackID, roomName, roomSID, res.EgressId)
	return res.EgressId, nil
}

// stopEgress останавливает запись Egress
func (mp *ManagingPortal) stopEgress(egressID string) error {
	if egressID == "" {
		return nil // Нечего останавливать
	}

	config := loadEgressConfig()
	if !config.Enabled {
		return nil
	}

	// Создаем Egress клиент
	egressClient := lksdk.NewEgressClient(
		os.Getenv("LIVEKIT_URL"),
		os.Getenv("LIVEKIT_API_KEY"),
		os.Getenv("LIVEKIT_API_SECRET"),
	)

	req := &livekit.StopEgressRequest{
		EgressId: egressID,
	}

	_, err := egressClient.StopEgress(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to stop egress %s: %w", egressID, err)
	}

	mp.logger.Infof("Stopped egress: %s", egressID)
	return nil
}

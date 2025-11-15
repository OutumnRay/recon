package livekit

import (
	"context"
	"fmt"
	"os"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

// EgressConfig contains configuration for LiveKit Egress
type EgressConfig struct {
	URL       string
	APIKey    string
	APISecret string
	S3Config  S3Config
}

// S3Config contains S3/MinIO configuration for recording storage
type S3Config struct {
	Endpoint       string
	Bucket         string
	Region         string
	AccessKey      string
	SecretKey      string
	ForcePathStyle bool
}

// EgressClient wraps LiveKit Egress client
type EgressClient struct {
	client *lksdk.EgressClient
	config EgressConfig
}

// NewEgressClient creates a new LiveKit Egress client
func NewEgressClient(config EgressConfig) *EgressClient {
	client := lksdk.NewEgressClient(config.URL, config.APIKey, config.APISecret)
	return &EgressClient{
		client: client,
		config: config,
	}
}

// NewEgressClientFromEnv creates a new LiveKit Egress client from environment variables
func NewEgressClientFromEnv() *EgressClient {
	config := EgressConfig{
		URL:       getEnv("LIVEKIT_URL", "https://video.recontext.online"),
		APIKey:    getEnv("LIVEKIT_API_KEY", ""),
		APISecret: getEnv("LIVEKIT_API_SECRET", ""),
		S3Config: S3Config{
			Endpoint:       getEnv("MINIO_ENDPOINT", "https://api.storage.recontext.online"),
			Bucket:         getEnv("MINIO_BUCKET", "recontext"),
			Region:         getEnv("MINIO_REGION", ""),
			AccessKey:      getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey:      getEnv("MINIO_SECRET_KEY", "minioadmin"),
			ForcePathStyle: true,
		},
	}
	return NewEgressClient(config)
}

// StartRoomCompositeEgress starts recording entire room (composite view)
func (ec *EgressClient) StartRoomCompositeEgress(ctx context.Context, roomName string, roomSID string, audioOnly bool) (*livekit.EgressInfo, error) {
	preset := livekit.EncodingOptionsPreset_H264_720P_30
	if audioOnly {
		preset = livekit.EncodingOptionsPreset_H264_720P_30 // LiveKit doesn't have audio-only preset, we use video preset
	}

	req := &livekit.RoomCompositeEgressRequest{
		RoomName:  roomName,
		Layout:    "speaker", // speaker layout - shows active speaker prominently
		AudioOnly: audioOnly,
		Options: &livekit.RoomCompositeEgressRequest_Preset{
			Preset: preset,
		},
		// Структура: {meetingID}_{roomSID}
		SegmentOutputs: []*livekit.SegmentedFileOutput{
			{
				FilenamePrefix:   fmt.Sprintf("%s_%s", roomName, roomSID),
				PlaylistName:     "composite.m3u8",
				LivePlaylistName: "composite-live.m3u8",
				SegmentDuration:  2, // 2 seconds per segment
				Output: &livekit.SegmentedFileOutput_S3{
					S3: &livekit.S3Upload{
						Endpoint:       ec.config.S3Config.Endpoint,
						Bucket:         ec.config.S3Config.Bucket,
						Region:         ec.config.S3Config.Region,
						AccessKey:      ec.config.S3Config.AccessKey,
						Secret:         ec.config.S3Config.SecretKey,
						ForcePathStyle: ec.config.S3Config.ForcePathStyle,
					},
				},
			},
		},
	}

	info, err := ec.client.StartRoomCompositeEgress(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start room composite egress: %w", err)
	}

	return info, nil
}

// StartTrackCompositeEgress starts recording a specific track (audio or video)
func (ec *EgressClient) StartTrackCompositeEgress(ctx context.Context, roomName, roomSID, audioTrackID, videoTrackID string) (*livekit.EgressInfo, error) {
	trackFilename := getTrackFilename(audioTrackID, videoTrackID)

	req := &livekit.TrackCompositeEgressRequest{
		RoomName:     roomName,
		AudioTrackId: audioTrackID,
		VideoTrackId: videoTrackID,
		Options: &livekit.TrackCompositeEgressRequest_Preset{
			Preset: livekit.EncodingOptionsPreset_H264_720P_30,
		},
		// Структура: {meetingID}_{roomSID} - LiveKit сам добавит /tracks/{trackID}
		SegmentOutputs: []*livekit.SegmentedFileOutput{
			{
				FilenamePrefix:   fmt.Sprintf("%s_%s", roomName, roomSID),
				PlaylistName:     fmt.Sprintf("%s.m3u8", trackFilename),
				LivePlaylistName: fmt.Sprintf("%s-live.m3u8", trackFilename),
				SegmentDuration:  20, // 20 seconds per segment for tracks
				Output: &livekit.SegmentedFileOutput_S3{
					S3: &livekit.S3Upload{
						Endpoint:       ec.config.S3Config.Endpoint,
						Bucket:         ec.config.S3Config.Bucket,
						Region:         ec.config.S3Config.Region,
						AccessKey:      ec.config.S3Config.AccessKey,
						Secret:         ec.config.S3Config.SecretKey,
						ForcePathStyle: ec.config.S3Config.ForcePathStyle,
					},
				},
			},
		},
	}

	info, err := ec.client.StartTrackCompositeEgress(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start track composite egress: %w", err)
	}

	return info, nil
}

// StartTrackEgress starts recording a single track to file
func (ec *EgressClient) StartTrackEgress(ctx context.Context, roomName, trackID string) (*livekit.EgressInfo, error) {
	req := &livekit.TrackEgressRequest{
		RoomName: roomName,
		TrackId:  trackID,
		Output: &livekit.TrackEgressRequest_File{
			File: &livekit.DirectFileOutput{
				Filepath: fmt.Sprintf("{room_name}/{track_id}"),
				Output: &livekit.DirectFileOutput_S3{
					S3: &livekit.S3Upload{
						Endpoint:       ec.config.S3Config.Endpoint,
						Bucket:         ec.config.S3Config.Bucket,
						Region:         ec.config.S3Config.Region,
						AccessKey:      ec.config.S3Config.AccessKey,
						Secret:         ec.config.S3Config.SecretKey,
						ForcePathStyle: ec.config.S3Config.ForcePathStyle,
					},
				},
			},
		},
	}

	info, err := ec.client.StartTrackEgress(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start track egress: %w", err)
	}

	return info, nil
}

// StopEgress stops an ongoing egress session
func (ec *EgressClient) StopEgress(ctx context.Context, egressID string) (*livekit.EgressInfo, error) {
	req := &livekit.StopEgressRequest{
		EgressId: egressID,
	}

	info, err := ec.client.StopEgress(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to stop egress: %w", err)
	}

	return info, nil
}

// ListEgress lists all egress sessions for a room
func (ec *EgressClient) ListEgress(ctx context.Context, roomName string) ([]*livekit.EgressInfo, error) {
	req := &livekit.ListEgressRequest{
		RoomName: roomName,
	}

	res, err := ec.client.ListEgress(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list egress: %w", err)
	}

	return res.Items, nil
}

// Helper functions

func getTrackFilename(audioTrackID, videoTrackID string) string {
	if audioTrackID != "" && videoTrackID != "" {
		return fmt.Sprintf("%s_%s", audioTrackID, videoTrackID)
	} else if audioTrackID != "" {
		return audioTrackID
	} else if videoTrackID != "" {
		return videoTrackID
	}
	return "track"
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

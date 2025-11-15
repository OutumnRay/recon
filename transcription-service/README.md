# Transcription Service

Python-based transcription service using Faster Whisper for speech recognition.

## Overview

This service consumes messages from RabbitMQ when track recordings are completed, transcribes the audio using Faster Whisper (medium model), and stores phrase-level transcriptions with timestamps in the database.

## Features

- **Faster Whisper Integration**: Uses the optimized Faster Whisper model for efficient transcription
- **Phrase-level Transcription**: Generates transcriptions with start/end timestamps for each phrase
- **Multi-language Support**: Auto-detects language or accepts language parameter
- **RabbitMQ Consumer**: Listens for track completion messages
- **Database Storage**: Stores transcription phrases linked to track and user
- **Status Tracking**: Maintains transcription status (pending, processing, completed, failed)
- **Error Handling**: Robust error handling with detailed logging

## Requirements

- Python 3.11+
- PostgreSQL
- RabbitMQ
- FFmpeg (for audio processing)

## Installation

### Local Development

1. Install dependencies:
```bash
pip install -r requirements.txt
```

2. Configure environment:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Run the service:
```bash
python main.py
```

### Docker

Build and run with Docker:
```bash
docker build -t transcription-service .
docker run -d --name transcription-service \
  -e RABBITMQ_HOST=rabbitmq \
  -e DB_HOST=postgres \
  transcription-service
```

## Configuration

Environment variables (see `.env.example`):

### RabbitMQ
- `RABBITMQ_HOST`: RabbitMQ server host (default: localhost)
- `RABBITMQ_PORT`: RabbitMQ port (default: 5672)
- `RABBITMQ_USER`: RabbitMQ username (default: guest)
- `RABBITMQ_PASSWORD`: RabbitMQ password (default: guest)
- `RABBITMQ_QUEUE`: Queue name for transcription tasks (default: transcription_queue)

### Database
- `DB_HOST`: PostgreSQL host (default: localhost)
- `DB_PORT`: PostgreSQL port (default: 5432)
- `DB_NAME`: Database name (default: recontext)
- `DB_USER`: Database user (default: postgres)
- `DB_PASSWORD`: Database password (default: postgres)

### Whisper
- `WHISPER_MODEL`: Model size (tiny, base, small, medium, large) (default: medium)
- `WHISPER_DEVICE`: Device to use (cpu or cuda) (default: cpu)
- `WHISPER_COMPUTE_TYPE`: Compute type (int8, float16, float32) (default: int8)

## Message Format

The service expects RabbitMQ messages in the following JSON format:

```json
{
  "track_id": "uuid-string",
  "user_id": "uuid-string",
  "audio_url": "http://minio:9000/recontext/<egress-id>/playlist.m3u8",
  "language": "en",
  "token": "optional-auth-token"
}
```

**Note**: The `audio_url` typically points to an m3u8 playlist file. The service will:
1. Download the playlist
2. Parse it to extract segment URLs
3. Download all segments (.ts files)
4. Combine them using FFmpeg into a single audio file
5. Transcribe the combined file

## Database Schema

### transcription_phrases

Stores individual transcribed phrases with timestamps:

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| track_id | UUID | Reference to track recording |
| user_id | UUID | Reference to user |
| phrase_index | INTEGER | Sequential phrase number |
| start_time | NUMERIC(10,3) | Phrase start time in seconds |
| end_time | NUMERIC(10,3) | Phrase end time in seconds |
| text | TEXT | Transcribed text |
| confidence | NUMERIC(5,4) | Confidence score |
| language | VARCHAR(10) | Detected language code |
| created_at | TIMESTAMP | Creation timestamp |
| updated_at | TIMESTAMP | Last update timestamp |

### transcription_status

Tracks transcription job status:

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| track_id | UUID | Reference to track recording (unique) |
| status | VARCHAR(50) | Status (pending, processing, completed, failed) |
| started_at | TIMESTAMP | Job start time |
| completed_at | TIMESTAMP | Job completion time |
| error_message | TEXT | Error details if failed |
| phrase_count | INTEGER | Number of phrases transcribed |
| total_duration | NUMERIC(10,3) | Total audio duration |

## Architecture

1. **Worker** (`worker.py`): RabbitMQ consumer that processes transcription tasks
2. **Transcriber** (`transcriber.py`): Faster Whisper integration for audio transcription
3. **Database** (`database.py`): PostgreSQL operations for storing transcriptions
4. **Config** (`config.py`): Configuration management from environment variables

## Workflow

1. Backend sends message to RabbitMQ when track recording completes
2. Transcription service receives message from queue
3. Service downloads m3u8 playlist from MinIO
4. Parse m3u8 playlist to get all segment URLs
5. Download all .ts segments sequentially
6. Combine segments into single audio file using FFmpeg
7. Faster Whisper transcribes combined audio into phrases with timestamps
8. Phrases are stored in `transcription_phrases` table
9. Track status is updated to "ready" in `track_recordings` table
10. Transcription status is marked as "completed"
11. Cleanup temporary files (segments and combined audio)

## Performance

- **Model**: Medium model provides good balance of speed and accuracy
- **Device**: Use `cuda` with GPU for faster transcription (5-10x speedup)
- **Compute Type**: `int8` is faster, `float16`/`float32` more accurate (requires GPU)
- **VAD Filter**: Voice Activity Detection filters silence for better performance

## Error Handling

- Failed transcriptions are logged with full error details
- Status is updated to "failed" with error message
- Messages are not requeued to prevent infinite loops
- Database cleanup handles partial transcriptions

## Monitoring

Check logs for:
- RabbitMQ connection status
- Model loading confirmation
- Transcription progress
- Error details

Query database for:
```sql
-- Check transcription status
SELECT * FROM transcription_status WHERE track_id = 'uuid';

-- Get transcription phrases
SELECT * FROM transcription_phrases WHERE track_id = 'uuid' ORDER BY phrase_index;

-- Monitor processing stats
SELECT status, COUNT(*) FROM transcription_status GROUP BY status;
```

## Troubleshooting

### Model Download
First run will download the Whisper model (~1.5GB for medium). Ensure sufficient disk space and internet connection.

### Memory Usage
Medium model requires ~5GB RAM. For limited resources, use `small` or `base` model.

### GPU Support
For CUDA support:
```bash
pip install faster-whisper[cuda]
```
Set `WHISPER_DEVICE=cuda` in environment.

### RabbitMQ Connection
Ensure RabbitMQ is running and accessible:
```bash
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management
```

### M3U8 Playlist Handling

**Issue**: Failed to download or parse m3u8 playlist

**Solution**:
- Verify MinIO URL is accessible from transcription service
- Check that egress recording created the playlist file
- Verify file path construction in managing portal logs
- Test URL manually: `curl http://minio:9000/recontext/<egress-id>/playlist.m3u8`

**Issue**: FFmpeg concatenation fails

**Solution**:
- Ensure FFmpeg is installed in Docker container (already included in Dockerfile)
- Check all segments downloaded successfully
- Verify temp directory has sufficient space
- Check FFmpeg error output in logs

**Issue**: No transcription tasks received

**Solution**:
- Check managing portal logs for "Transcription task sent to RabbitMQ"
- Verify egress is a track egress (not room composite)
- Check RabbitMQ queue has messages: http://localhost:15672 (guest/guest)
- Verify transcription service is connected to RabbitMQ

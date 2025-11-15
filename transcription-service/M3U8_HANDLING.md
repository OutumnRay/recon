# M3U8 Playlist Handling in Transcription Service

## Overview

The transcription service now supports automatic download and processing of m3u8 HLS playlists. This is necessary because LiveKit egress creates track recordings as m3u8 playlists with multiple .ts segments, not single audio files.

## How It Works

### 1. Backend (Go) - URL Construction

When a track recording completes, the managing portal constructs the correct m3u8 URL:

```go
// In cmd/managing-portal/handlers_livekit.go

// Construct URL to m3u8 playlist
// LiveKit stores track recordings in: <bucket>/<egress_id>/playlist.m3u8
var audioURL string
if strings.HasSuffix(filePath, ".m3u8") {
    // File path already points to playlist
    audioURL = fmt.Sprintf("http://%s/%s/%s", storageURL, bucket, filePath)
} else if strings.Contains(filePath, "/") {
    // File path is a directory or file path, append playlist.m3u8
    audioURL = fmt.Sprintf("http://%s/%s/%s/playlist.m3u8", storageURL, bucket, strings.TrimSuffix(filePath, "/"))
} else {
    // File path is just egress ID
    audioURL = fmt.Sprintf("http://%s/%s/%s/playlist.m3u8", storageURL, bucket, filePath)
}
```

**Example URLs**:
- `http://minio:9000/recontext/EG_abc123def456/playlist.m3u8`
- `http://minio:9000/recontext/recontext/EG_abc123def456/playlist.m3u8`

### 2. Python Service - Download and Combine

The transcription service automatically detects m3u8 URLs and processes them:

```python
# In transcription-service/transcriber.py

def download_audio(self, audio_url: str, token: Optional[str] = None) -> str:
    # Check if this is an m3u8 playlist
    if audio_url.endswith('.m3u8') or 'playlist.m3u8' in audio_url:
        return self.download_m3u8_and_combine(audio_url, token)

    # Otherwise download as regular file
    # ...
```

### 3. M3U8 Processing Steps

```
1. Download playlist.m3u8 file
   GET http://minio:9000/recontext/EG_abc123/playlist.m3u8

2. Parse playlist content
   #EXTM3U
   #EXT-X-VERSION:3
   #EXT-X-TARGETDURATION:6
   #EXTINF:6.000000,
   segment_0.ts
   #EXTINF:6.000000,
   segment_1.ts
   ...

3. Extract segment URLs
   http://minio:9000/recontext/EG_abc123/segment_0.ts
   http://minio:9000/recontext/EG_abc123/segment_1.ts
   ...

4. Download each segment
   Saved to: /tmp/tmpXXXXXX/segment_0000.ts
   Saved to: /tmp/tmpXXXXXX/segment_0001.ts
   ...

5. Create FFmpeg concat file
   file '/tmp/tmpXXXXXX/segment_0000.ts'
   file '/tmp/tmpXXXXXX/segment_0001.ts'
   ...

6. Combine with FFmpeg
   ffmpeg -f concat -safe 0 -i concat.txt -c copy output.m4a

7. Return combined file path
   /tmp/tmpYYYYYY.m4a

8. Transcribe combined audio

9. Cleanup temp files
```

## Code Example

### Download M3U8 and Combine

```python
def download_m3u8_and_combine(self, m3u8_url: str, token: Optional[str] = None) -> str:
    print(f"📥 Downloading and combining m3u8 from: {m3u8_url}")

    # Download m3u8 playlist
    response = requests.get(m3u8_url, headers=headers)
    playlist_content = response.text

    # Parse m3u8 to get segment URLs
    base_url = m3u8_url.rsplit('/', 1)[0]
    segments = []

    for line in playlist_content.split('\n'):
        line = line.strip()
        if line and not line.startswith('#'):
            segment_url = f"{base_url}/{line}"
            segments.append(segment_url)

    print(f"📦 Found {len(segments)} segments to download")

    # Download all segments
    for i, segment_url in enumerate(segments):
        print(f"📥 Downloading segment {i+1}/{len(segments)}")
        segment_data = requests.get(segment_url, headers=headers)
        # Save to temp file
        # ...

    # Combine with FFmpeg
    ffmpeg_cmd = [
        'ffmpeg',
        '-f', 'concat',
        '-safe', '0',
        '-i', concat_file_path,
        '-c', 'copy',
        '-y',
        combined_path
    ]

    subprocess.run(ffmpeg_cmd, timeout=300)

    return combined_path
```

## Logging Output

### Successful Processing

```
📥 Downloading and combining m3u8 from: http://minio:9000/recontext/EG_abc123/playlist.m3u8
📄 Playlist content:
#EXTM3U
#EXT-X-VERSION:3
...
📦 Found 25 segments to download
📥 Downloading segment 1/25: http://minio:9000/recontext/EG_abc123/segment_0.ts
✅ Downloaded segment 1/25
📥 Downloading segment 2/25: http://minio:9000/recontext/EG_abc123/segment_1.ts
✅ Downloaded segment 2/25
...
🔧 Combining 25 segments with ffmpeg...
✅ Combined audio saved to: /tmp/tmpXXXXXX.m4a
📊 Combined file size: 15.43 MB

Starting transcription of: /tmp/tmpXXXXXX.m4a
Detected language: en (probability: 0.99)
Transcription completed: 142 phrases
```

## Error Handling

### Common Issues

**1. Playlist not found**
```
❌ HTTP 404: Not Found
Error: Failed to download m3u8 playlist
```
**Solution**: Verify egress completed successfully and created the playlist

**2. Segment download failure**
```
❌ HTTP 403: Forbidden
Error: Failed to download segment_5.ts
```
**Solution**: Check MinIO permissions and network connectivity

**3. FFmpeg concatenation error**
```
❌ FFmpeg stderr: [concat @ 0x...] Impossible to open 'segment_0000.ts'
Error: FFmpeg failed with return code 1
```
**Solution**: Verify all segments downloaded successfully and temp directory has space

## Configuration

No special configuration needed! The service automatically:
- Detects m3u8 URLs by file extension
- Downloads and parses playlists
- Combines segments using FFmpeg
- Cleans up temporary files

FFmpeg is already included in the Docker container via the Dockerfile.

## Testing

### Manual Test

1. Get a track egress ID from the database:
```sql
SELECT egress_id FROM livekit_tracks WHERE egress_id IS NOT NULL LIMIT 1;
```

2. Test m3u8 download:
```bash
curl http://minio:9000/recontext/<egress-id>/playlist.m3u8
```

3. Send test message to RabbitMQ:
```python
import pika
import json

connection = pika.BlockingConnection(
    pika.ConnectionParameters('localhost')
)
channel = connection.channel()

message = {
    "track_id": "your-track-uuid",
    "user_id": "your-user-uuid",
    "audio_url": "http://minio:9000/recontext/EG_abc123/playlist.m3u8",
    "language": "",
    "token": ""
}

channel.basic_publish(
    exchange='',
    routing_key='transcription_queue',
    body=json.dumps(message)
)

connection.close()
```

4. Watch transcription service logs:
```bash
docker logs -f recontext-transcription-service
```

## Performance

### Segment Download
- **Sequential**: Segments downloaded one at a time
- **Speed**: ~1-2 seconds per segment
- **Total Time**: Depends on number of segments (typically 6-second chunks)

### FFmpeg Combining
- **Method**: `-c copy` (stream copy, no re-encoding)
- **Speed**: Very fast (~1-2 seconds for typical recordings)
- **Quality**: Lossless (no quality degradation)

### Example Timeline
For a 5-minute recording (~50 segments):
```
Download playlist:    0.5s
Download 50 segments: 75s
Combine with FFmpeg:  1.5s
Transcribe:          60s (medium model, CPU)
Total:               ~137s (~2.3 minutes)
```

## Benefits

✅ **Automatic Detection**: No configuration needed
✅ **Robust Parsing**: Handles various m3u8 formats
✅ **Fast Combining**: FFmpeg stream copy is very fast
✅ **Memory Efficient**: Streams segments to disk
✅ **Cleanup**: Automatically removes temp files
✅ **Error Recovery**: Detailed logging for debugging

## Future Improvements

- [ ] Parallel segment downloads for faster processing
- [ ] Resume capability for large playlists
- [ ] Validation of segment integrity (checksums)
- [ ] Support for adaptive bitrate playlists (master.m3u8)
- [ ] Progress reporting during download

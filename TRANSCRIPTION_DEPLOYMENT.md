# Transcription Service Deployment Guide

## Recent Updates

### Go Module Dependency Fix

**Issue**: Build failed with missing RabbitMQ package
```
pkg/rabbitmq/publisher.go:9:2: no required module provides package github.com/streadway/amqp
```

**Solution**: Updated to use the official RabbitMQ Go client

**Changes**:
1. Updated `pkg/rabbitmq/publisher.go` import:
   ```go
   // OLD (deprecated):
   import "github.com/streadway/amqp"

   // NEW (official):
   import amqp "github.com/rabbitmq/amqp091-go"
   ```

2. Added dependency to `go.mod`:
   ```go
   require (
       // ... other dependencies
       github.com/rabbitmq/amqp091-go v1.10.0
   )
   ```

### Docker Compose Configuration

The deployment is configured to use external services (RabbitMQ, PostgreSQL, MinIO) running on host `192.168.5.153`:

```yaml
# Managing Portal
- RABBITMQ_HOST=192.168.5.153
- DB_HOST=postgres  # Can be changed to 192.168.5.153 if needed
- MINIO_ENDPOINT=minio:9000

# User Portal
- RABBITMQ_HOST=192.168.5.153
- DB_HOST=192.168.5.153
- MINIO_ENDPOINT=192.168.5.153:9000

# Transcription Service
- RABBITMQ_HOST=192.168.5.153
- DB_HOST=postgres
- STORAGE_BASE_URL=https://api.storage.recontext.online
```

**Important Notes**:
- RabbitMQ, PostgreSQL, and MinIO containers are commented out in docker-compose.yml
- Services point to external instances on `192.168.5.153`
- MinIO uses public endpoint: `https://api.storage.recontext.online`
- MinIO secret updated to: `32a4953d5bff4a1c6aea4d4ccfb757e5`

## Deployment Steps

### 1. Prerequisites

Ensure external services are running on `192.168.5.153`:
- **RabbitMQ**: Port 5672 (guest/guest)
- **PostgreSQL**: Port 5432 (recontext/recontext)
- **MinIO**: Accessible via `https://api.storage.recontext.online`

### 2. Build Services

```bash
cd /Volumes/ExternalData/source/Team21/Recontext.online

# Build managing portal
docker-compose build managing-portal

# Build user portal
docker-compose build user-portal

# Build transcription service
docker-compose build transcription-service
```

### 3. Start Services

```bash
# Start all services
docker-compose up -d

# Or start individually
docker-compose up -d managing-portal
docker-compose up -d user-portal
docker-compose up -d transcription-service
```

### 4. Verify Services

```bash
# Check logs
docker logs -f recontext-managing-portal
docker logs -f recontext-user-portal
docker logs -f recontext-transcription-service

# Check service status
docker ps | grep recontext
```

### 5. Verify RabbitMQ Connection

Check managing portal logs for:
```
✅ RabbitMQ publisher initialized successfully
```

Check transcription service logs for:
```
Connecting to RabbitMQ at 192.168.5.153:5672
Connected to queue: transcription_queue
Waiting for transcription tasks...
```

### 6. Test Transcription Workflow

1. **Create a meeting** in the web interface
2. **Join and record**: Enable track recording for a participant
3. **End recording**: Stop the track recording
4. **Check managing portal logs** for:
   ```
   📝 Sending transcription task for track <track-sid> (egress: <egress-id>)
   📌 Constructed audio URL: http://192.168.5.153:9000/recontext/<egress-id>/playlist.m3u8
   ✅ Transcription task sent to RabbitMQ for track <track-sid>
   ```

5. **Check transcription service logs** for:
   ```
   Processing transcription task:
     Track ID: <uuid>
     User ID: <uuid>
     Audio URL: https://api.storage.recontext.online/recontext/<egress-id>/playlist.m3u8

   📥 Downloading and combining m3u8 from: ...
   📦 Found X segments to download
   🔧 Combining X segments with ffmpeg...
   ✅ Combined audio saved to: /tmp/...

   Starting transcription of: ...
   Detected language: en (probability: 0.99)
   Transcription completed: Y phrases

   ✅ Transcription completed successfully!
   ```

6. **Verify in database**:
   ```sql
   -- Check transcription status
   SELECT * FROM transcription_status ORDER BY started_at DESC LIMIT 5;

   -- Check transcription phrases
   SELECT
       phrase_index,
       start_time,
       end_time,
       text
   FROM transcription_phrases
   WHERE track_id = '<your-track-uuid>'
   ORDER BY phrase_index
   LIMIT 10;
   ```

## Troubleshooting

### Managing Portal Won't Start

**Check**:
- Database connection: `psql -h 192.168.5.153 -U recontext -d recontext`
- RabbitMQ connection: `telnet 192.168.5.153 5672`

**Logs**:
```bash
docker logs recontext-managing-portal 2>&1 | grep -i error
```

### Transcription Service Won't Connect

**Check RabbitMQ**:
```bash
# Access RabbitMQ management UI
# http://192.168.5.153:15672
# Login: guest/guest

# Check if transcription_queue exists
# Check if service is consuming
```

**Check Database**:
```bash
# Connect to PostgreSQL
psql -h 192.168.5.153 -U recontext -d recontext

# Verify tables exist
\dt transcription*
```

### M3U8 Download Fails

**Check MinIO Access**:
```bash
# Test playlist download
curl https://api.storage.recontext.online/recontext/<egress-id>/playlist.m3u8

# If using internal URL, test from container
docker exec -it recontext-transcription-service \
  curl http://192.168.5.153:9000/recontext/<egress-id>/playlist.m3u8
```

**Check Credentials**:
- Verify `MINIO_SECRET_KEY=32a4953d5bff4a1c6aea4d4ccfb757e5`
- Verify bucket name is `recontext`

### FFmpeg Errors

**Check FFmpeg Installation**:
```bash
docker exec -it recontext-transcription-service ffmpeg -version
```

**Check Temp Space**:
```bash
docker exec -it recontext-transcription-service df -h /tmp
```

## Monitoring

### RabbitMQ Management UI

Access: `http://192.168.5.153:15672`

**Check**:
- Queue `transcription_queue` exists
- Messages being consumed
- Consumer count > 0
- No unacked messages piling up

### Database Metrics

```sql
-- Processing statistics
SELECT
    status,
    COUNT(*) as count,
    AVG(phrase_count) as avg_phrases,
    AVG(EXTRACT(EPOCH FROM (completed_at - started_at))) as avg_seconds
FROM transcription_status
GROUP BY status;

-- Recent transcriptions
SELECT
    ts.track_id,
    ts.status,
    ts.phrase_count,
    ts.started_at,
    ts.completed_at,
    EXTRACT(EPOCH FROM (ts.completed_at - ts.started_at)) as duration_seconds
FROM transcription_status ts
ORDER BY ts.started_at DESC
LIMIT 10;

-- Failed transcriptions
SELECT
    track_id,
    error_message,
    started_at
FROM transcription_status
WHERE status = 'failed'
ORDER BY started_at DESC
LIMIT 10;
```

### Docker Logs

```bash
# Follow all logs
docker-compose logs -f

# Follow specific service
docker logs -f recontext-transcription-service

# Last 100 lines
docker logs --tail 100 recontext-transcription-service

# Search for errors
docker logs recontext-transcription-service 2>&1 | grep -i error
```

## Performance Tuning

### Whisper Model Selection

Current: **medium** (default)

For faster processing:
```yaml
- WHISPER_MODEL=small  # 3x faster, slightly less accurate
- WHISPER_MODEL=base   # 8x faster, good for clear audio
```

For better accuracy:
```yaml
- WHISPER_MODEL=large  # Slower but more accurate
```

### GPU Acceleration

If GPU available:
```yaml
transcription-service:
  environment:
    - WHISPER_DEVICE=cuda
    - WHISPER_COMPUTE_TYPE=float16
  deploy:
    resources:
      reservations:
        devices:
          - driver: nvidia
            count: 1
            capabilities: [gpu]
```

**Speedup**: 5-10x faster with GPU

### Scaling

Run multiple transcription workers:
```bash
docker-compose up -d --scale transcription-service=3
```

**Note**: Make sure container names don't conflict by removing `container_name` or using unique names.

## Maintenance

### Clean Up Old Transcriptions

```sql
-- Remove transcriptions older than 90 days
DELETE FROM transcription_phrases
WHERE created_at < NOW() - INTERVAL '90 days';

DELETE FROM transcription_status
WHERE started_at < NOW() - INTERVAL '90 days';
```

### Restart Services

```bash
# Restart all
docker-compose restart

# Restart specific service
docker-compose restart transcription-service

# Rebuild and restart
docker-compose up -d --build transcription-service
```

### Update Code

```bash
# Pull latest changes
git pull

# Rebuild services
docker-compose build

# Restart with new code
docker-compose up -d
```

## Security Considerations

1. **MinIO Credentials**: The secret key `32a4953d5bff4a1c6aea4d4ccfb757e5` is exposed in docker-compose.yml
   - Consider using Docker secrets or environment files
   - Rotate credentials periodically

2. **RabbitMQ**: Using default guest/guest credentials
   - Create dedicated user for production
   - Use strong passwords

3. **Database**: Using simple password "recontext"
   - Use strong password in production
   - Restrict network access

4. **Network**: Services communicate over unencrypted connections
   - Consider TLS for production
   - Use VPN or private network

## Backup and Recovery

### Database Backup

```bash
# Backup transcriptions
pg_dump -h 192.168.5.153 -U recontext -d recontext \
  -t transcription_phrases \
  -t transcription_status \
  > transcriptions_backup_$(date +%Y%m%d).sql

# Restore
psql -h 192.168.5.153 -U recontext -d recontext \
  < transcriptions_backup_20250115.sql
```

### Whisper Models Backup

Models are cached in Docker volume `whisper-models`:

```bash
# Backup volume
docker run --rm -v whisper-models:/data -v $(pwd):/backup \
  alpine tar czf /backup/whisper-models.tar.gz -C /data .

# Restore volume
docker run --rm -v whisper-models:/data -v $(pwd):/backup \
  alpine tar xzf /backup/whisper-models.tar.gz -C /data
```

## Summary

✅ **Fixed**: RabbitMQ dependency using official `rabbitmq/amqp091-go` package
✅ **Configured**: External services on `192.168.5.153`
✅ **Ready**: Complete m3u8 transcription workflow
✅ **Monitored**: Comprehensive logging and metrics
✅ **Scalable**: Can run multiple workers
✅ **Documented**: Full deployment and troubleshooting guide

The transcription service is ready for production deployment!

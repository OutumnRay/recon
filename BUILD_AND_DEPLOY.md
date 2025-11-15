# Build and Deploy Commands

## Quick Start

### 1. Build All Services

```bash
# Navigate to project root
cd /Volumes/ExternalData/source/Team21/Recontext.online

# Build all services
docker-compose build
```

### 2. Build Individual Services

```bash
# Build managing portal (includes RabbitMQ publisher)
docker-compose build managing-portal

# Build user portal
docker-compose build user-portal

# Build transcription service
docker-compose build transcription-service
```

### 3. Start Services

```bash
# Start all services in detached mode
docker-compose up -d

# Or start specific services
docker-compose up -d managing-portal user-portal transcription-service
```

### 4. View Logs

```bash
# Follow all logs
docker-compose logs -f

# Follow specific service
docker logs -f recontext-managing-portal
docker logs -f recontext-user-portal
docker logs -f recontext-transcription-service

# View last 100 lines
docker logs --tail 100 recontext-transcription-service
```

### 5. Restart Services

```bash
# Restart all
docker-compose restart

# Restart specific service
docker-compose restart transcription-service
```

### 6. Stop Services

```bash
# Stop all
docker-compose down

# Stop but keep volumes
docker-compose stop
```

## Verify Deployment

### Check Managing Portal

```bash
# Check logs for RabbitMQ initialization
docker logs recontext-managing-portal 2>&1 | grep -i rabbitmq

# Should see:
# ✅ RabbitMQ publisher initialized successfully
```

### Check Transcription Service

```bash
# Check logs for startup
docker logs recontext-transcription-service 2>&1 | head -50

# Should see:
# Loading Whisper model: medium
# Whisper model loaded successfully
# Initializing database tables...
# Database tables ready
# Connecting to RabbitMQ at 192.168.5.153:5672
# Connected to queue: transcription_queue
# Waiting for transcription tasks...
```

### Test RabbitMQ Connection

```bash
# From host machine
telnet 192.168.5.153 5672

# Should connect successfully
# Press Ctrl+] then type 'quit' to exit
```

### Check Database Tables

```bash
# Connect to PostgreSQL
psql -h 192.168.5.153 -U recontext -d recontext

# Check if transcription tables exist
\dt transcription*

# Should show:
# transcription_phrases
# transcription_status

# Exit
\q
```

## Test Transcription Workflow

### End-to-End Test

1. **Create a meeting** in the web interface at `https://portal.recontext.online`

2. **Join the meeting** and enable microphone

3. **Start track recording** (should be automatic with LiveKit egress)

4. **Speak for a few seconds** to generate audio

5. **End the recording**

6. **Monitor Managing Portal logs**:
   ```bash
   docker logs -f recontext-managing-portal | grep -E "egress_ended|Transcription"
   ```

   Look for:
   ```
   🏁 Processing egress_ended event...
   📌 Egress ID: EG_xxxxxx
   📌 File Path: recontext/EG_xxxxxx/playlist.m3u8
   📝 Sending transcription task for track xxx (egress: EG_xxxxxx)
   📌 Constructed audio URL: https://api.storage.recontext.online/recontext/EG_xxxxxx/playlist.m3u8
   ✅ Transcription task sent to RabbitMQ for track xxx
   ```

7. **Monitor Transcription Service logs**:
   ```bash
   docker logs -f recontext-transcription-service
   ```

   Look for:
   ```
   ============================================================
   Processing transcription task:
     Track ID: xxx
     User ID: xxx
     Audio URL: https://api.storage.recontext.online/recontext/EG_xxx/playlist.m3u8
   ============================================================

   📥 Downloading and combining m3u8 from: ...
   📄 Playlist content: ...
   📦 Found X segments to download
   📥 Downloading segment 1/X: ...
   ✅ Downloaded segment 1/X
   ...
   🔧 Combining X segments with ffmpeg...
   ✅ Combined audio saved to: /tmp/...
   📊 Combined file size: X.XX MB

   Starting transcription of: /tmp/...
   Detected language: en (probability: 0.99)
   Transcription completed: Y phrases

   ============================================================
   Transcription completed successfully!
     Phrases: Y
     Duration: XXX.XX s
   ============================================================
   ```

8. **Check database**:
   ```sql
   -- Check status
   SELECT * FROM transcription_status ORDER BY started_at DESC LIMIT 5;

   -- Check phrases
   SELECT phrase_index, start_time, end_time, text
   FROM transcription_phrases
   WHERE track_id = 'your-track-uuid'
   ORDER BY phrase_index;
   ```

## Troubleshooting

### Build Fails - Missing Dependencies

```bash
# Clean build cache
docker-compose build --no-cache managing-portal

# Or rebuild everything
docker-compose build --no-cache
```

### Go Module Issues

```bash
# Update go.sum
go mod tidy

# Download dependencies
go mod download

# Verify go.mod and go.sum are committed
git status
```

### RabbitMQ Connection Failed

```bash
# Check if RabbitMQ is running on 192.168.5.153
telnet 192.168.5.153 5672

# Check RabbitMQ management UI
# http://192.168.5.153:15672
# Login: guest/guest

# Verify queue exists: transcription_queue
```

### Database Connection Failed

```bash
# Test PostgreSQL connection
psql -h 192.168.5.153 -U recontext -d recontext -c "SELECT version();"

# Check if database is running
pg_isready -h 192.168.5.153 -p 5432 -U recontext
```

### MinIO Access Failed

```bash
# Test MinIO access
curl https://api.storage.recontext.online/recontext/

# Test from inside container
docker exec -it recontext-transcription-service \
  curl https://api.storage.recontext.online/recontext/
```

### Transcription Not Starting

**Check**:
1. Managing portal successfully sent message to RabbitMQ
2. RabbitMQ queue has messages
3. Transcription service is connected to RabbitMQ
4. No errors in transcription service logs

**Verify RabbitMQ Queue**:
- Go to http://192.168.5.153:15672
- Navigate to Queues → transcription_queue
- Check "Messages" count
- Check "Consumers" count (should be > 0)

### FFmpeg Not Found

```bash
# Check if FFmpeg is installed in container
docker exec -it recontext-transcription-service ffmpeg -version

# Should show FFmpeg version
# If not found, rebuild container
docker-compose build --no-cache transcription-service
```

## Update and Redeploy

### After Code Changes

```bash
# Pull latest changes
git pull

# Rebuild affected services
docker-compose build managing-portal
docker-compose build transcription-service

# Restart services
docker-compose up -d managing-portal transcription-service
```

### Update Whisper Model

To change the Whisper model, edit docker-compose.yml:

```yaml
transcription-service:
  environment:
    - WHISPER_MODEL=small  # Change from medium to small
```

Then rebuild:
```bash
docker-compose up -d transcription-service
```

**Note**: First run will download the new model (~500MB for small, ~1.5GB for medium)

### Enable GPU Acceleration

If GPU is available, update docker-compose.yml:

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

## Maintenance

### View Resource Usage

```bash
# Show container stats
docker stats

# Show specific container
docker stats recontext-transcription-service
```

### Clean Up

```bash
# Remove stopped containers
docker-compose down

# Remove volumes (careful - deletes data!)
docker-compose down -v

# Remove unused images
docker image prune -a

# Clean up Whisper model cache (to download fresh)
docker volume rm whisper-models
```

### Backup Whisper Models

```bash
# Backup Whisper models volume
docker run --rm -v whisper-models:/data -v $(pwd):/backup \
  alpine tar czf /backup/whisper-models-backup.tar.gz -C /data .

# Restore
docker run --rm -v whisper-models:/data -v $(pwd):/backup \
  alpine tar xzf /backup/whisper-models-backup.tar.gz -C /data
```

## Production Deployment Checklist

Before deploying to production:

- [ ] Update MinIO secret key (currently exposed)
- [ ] Update RabbitMQ credentials (currently guest/guest)
- [ ] Update database password (currently "recontext")
- [ ] Configure TLS/SSL for external services
- [ ] Set up monitoring and alerting
- [ ] Configure log rotation
- [ ] Set up automated backups
- [ ] Test failover scenarios
- [ ] Document recovery procedures
- [ ] Set up health checks
- [ ] Configure resource limits

## Quick Reference

### Important URLs

- **Managing Portal**: http://localhost:20080
- **User Portal**: http://localhost:20081 or https://portal.recontext.online
- **RabbitMQ Management**: http://192.168.5.153:15672 (guest/guest)
- **MinIO**: https://api.storage.recontext.online

### Important Directories

- **Project Root**: `/Volumes/ExternalData/source/Team21/Recontext.online`
- **Transcription Service**: `./transcription-service/`
- **Go Backend**: `./cmd/managing-portal/`, `./cmd/user-portal/`
- **Mobile App**: `./mobile2/`

### Important Files

- **docker-compose.yml**: Service orchestration
- **go.mod**: Go dependencies
- **transcription-service/requirements.txt**: Python dependencies
- **TRANSCRIPTION_IMPLEMENTATION.md**: Architecture documentation
- **TRANSCRIPTION_DEPLOYMENT.md**: Deployment guide

## Support

For issues or questions:

1. Check logs first: `docker logs -f recontext-transcription-service`
2. Review documentation: `TRANSCRIPTION_IMPLEMENTATION.md`
3. Check RabbitMQ management UI for queue status
4. Verify database tables exist
5. Test external service connectivity

Common log locations:
- Managing portal: `docker logs recontext-managing-portal`
- User portal: `docker logs recontext-user-portal`
- Transcription service: `docker logs recontext-transcription-service`

# Transcription Service 3

AI-powered transcription service using Faster-Whisper for the Recontext platform. This service processes audio/video files from RabbitMQ queues and stores transcription results.

## Features

- **GPU-accelerated transcription** using NVIDIA CUDA
- **Faster-Whisper** for optimized inference
- **Pre-downloaded Whisper large-v3 model** embedded in Docker image (~3GB)
- **RabbitMQ integration** for job queue management
- **MinIO/S3 storage** for audio files and results
- **Docker support** with NVIDIA GPU runtime

## Requirements

### Hardware
- NVIDIA GPU with CUDA support (recommended)
- Minimum 8GB GPU VRAM for medium model
- CPU fallback available but slower

### Software
- Docker 20.10+
- Docker Compose 1.29+
- NVIDIA Docker runtime (for GPU support)

## Quick Start

### Simple Start (Recommended)

```bash
cd transcription-service3

# Start with existing infrastructure (integrated mode)
./docker-start.sh integrated start

# Or start standalone (includes RabbitMQ and MinIO)
./docker-start.sh standalone start

# View logs
./docker-start.sh integrated logs

# Stop service
./docker-start.sh integrated stop
```

**All docker-start.sh commands:**
```bash
./docker-start.sh [mode] [action]

# Modes: integrated (default) | standalone
# Actions: start | stop | restart | logs | status | shell

# Examples:
./docker-start.sh                    # Start integrated mode
./docker-start.sh standalone         # Start standalone mode
./docker-start.sh integrated logs    # Show logs
./docker-start.sh integrated status  # Show status
./docker-start.sh integrated shell   # Open bash in container
```

### Using Docker Compose Directly

#### Standalone Mode (with included RabbitMQ and MinIO)

```bash
# Copy environment file
cp .env.example .env

# Edit .env with your configuration (optional)
nano .env

# Start services
docker-compose up -d

# View logs
docker-compose logs -f transcription-service3

# Stop services
docker-compose down
```

#### Integrated Mode (with existing Recontext infrastructure)

```bash
# Copy environment file
cp .env.example .env

# Edit .env to point to existing services
nano .env

# Start only the transcription service
docker-compose -f docker-compose.integrated.yml up -d

# View logs
docker-compose -f docker-compose.integrated.yml logs -f

# Stop service
docker-compose -f docker-compose.integrated.yml down
```

### Using Make Commands

```bash
# Standalone mode
make up              # Start all services
make logs            # View logs
make down            # Stop services
make restart         # Restart service

# Integrated mode
make up-integrated   # Start transcription service only
make logs-integrated # View logs
make down-integrated # Stop service
```

## Configuration

All configuration is done via environment variables in `.env` file:

### RabbitMQ Settings
```env
RABBITMQ_HOST=rabbitmq          # RabbitMQ host
RABBITMQ_PORT=5672              # RabbitMQ port
RABBITMQ_USER=guest             # RabbitMQ username
RABBITMQ_PASSWORD=guest         # RabbitMQ password
RABBITMQ_QUEUE=transcription_queue
RABBITMQ_RESULT_QUEUE=transcription_results
```

### Whisper Model Settings
```env
WHISPER_MODEL=large-v3          # Model size: tiny, base, small, medium, large-v2, large-v3
WHISPER_DEVICE=cuda             # Device: cuda, cpu
WHISPER_COMPUTE_TYPE=float16    # Compute type: float16, int8, float32
```

**Note:** The Docker image includes pre-downloaded **large-v3** model (~3GB). This means:
- ✅ No download delay on first run
- ✅ Faster container startup
- ✅ Works in air-gapped environments
- ⚠️ Larger Docker image size (~6GB total)

Available models (size vs accuracy):
- `tiny` - Fastest, lowest accuracy (~1GB VRAM)
- `base` - Fast, low accuracy (~1.5GB VRAM)
- `small` - Balanced (~2GB VRAM)
- `medium` - Good accuracy (~5GB VRAM)
- `large-v2` - High accuracy (~10GB VRAM)
- `large-v3` - Highest accuracy (~10GB VRAM) ⭐ **Pre-downloaded in image**

### Storage Settings
```env
MINIO_ENDPOINT=minio:9000       # MinIO endpoint
MINIO_ACCESS_KEY=minioadmin     # MinIO access key
MINIO_SECRET_KEY=minioadmin     # MinIO secret key
MINIO_SECURE=false              # Use HTTPS
MINIO_BUCKET=recontext          # Bucket name
```

## Docker Commands

### Build Image
```bash
# Note: This will download ~3GB Whisper large-v3 model during build
# Build time: ~15-20 minutes depending on internet speed
docker build -t recontext-transcription-service3:latest .
```

### Run Container (standalone)
```bash
docker run -d \
  --name transcription-service3 \
  --gpus all \
  --env-file .env \
  recontext-transcription-service3:latest
```

### View Logs
```bash
docker logs -f transcription-service3
```

### Stop Service
```bash
docker-compose down
```

### Restart Service
```bash
docker-compose restart transcription-service3
```

### Remove Everything (including volumes)
```bash
docker-compose down -v
```

## GPU Support

### Check GPU Availability
```bash
docker run --rm --gpus all nvcr.io/nvidia/pytorch:25.04-py3 nvidia-smi
```

### CPU-Only Mode
If you don't have a GPU, edit `.env`:
```env
WHISPER_DEVICE=cpu
WHISPER_COMPUTE_TYPE=float32
```

Then remove the GPU section from docker-compose.yml:
```yaml
# Comment out or remove this section
# deploy:
#   resources:
#     reservations:
#       devices:
#         - driver: nvidia
#           count: all
#           capabilities: [gpu]
```

## Monitoring

### Health Check
```bash
docker inspect --format='{{json .State.Health}}' transcription-service3 | jq
```

### Resource Usage
```bash
docker stats transcription-service3
```

### GPU Usage (if using GPU)
```bash
docker exec transcription-service3 nvidia-smi
```

## Troubleshooting

### GPU Not Detected
1. Install NVIDIA Docker runtime:
   ```bash
   distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
   curl -s -L https://nvidia.github.io/nvidia-docker/gpgkey | sudo apt-key add -
   curl -s -L https://nvidia.github.io/nvidia-docker/$distribution/nvidia-docker.list | \
     sudo tee /etc/apt/sources.list.d/nvidia-docker.list
   sudo apt-get update && sudo apt-get install -y nvidia-docker2
   sudo systemctl restart docker
   ```

2. Test GPU access:
   ```bash
   docker run --rm --gpus all nvidia/cuda:11.8.0-base-ubuntu22.04 nvidia-smi
   ```

### Connection Issues
- Check RabbitMQ is accessible: `telnet <RABBITMQ_HOST> 5672`
- Check MinIO is accessible: `curl http://<MINIO_ENDPOINT>/minio/health/live`

### Out of Memory
- Use a smaller model (e.g., `small` or `base`)
- Reduce batch size in transcriber.py
- Use int8 compute type: `WHISPER_COMPUTE_TYPE=int8`

## Performance Tuning

### Model Selection Guide
| Model | Size | VRAM | Speed | Quality | Use Case |
|-------|------|------|-------|---------|----------|
| tiny | 39M | ~1GB | Very Fast | Low | Testing/Development |
| base | 74M | ~1.5GB | Fast | Low | Quick drafts |
| small | 244M | ~2GB | Medium | Good | Balanced performance |
| medium | 769M | ~5GB | Medium | Very Good | Production (recommended) |
| large-v2 | 1550M | ~10GB | Slow | Excellent | High accuracy required |
| large-v3 | 1550M | ~10GB | Slow | Best | Maximum accuracy |

### Compute Type Performance
- `float16` - Best for GPUs with Tensor Cores (RTX 20xx+)
- `int8` - Best for memory-constrained GPUs
- `float32` - Required for CPU, slower but most compatible

## Integration with Recontext

This service integrates with the Recontext platform by:

1. **Consuming jobs** from RabbitMQ queue `transcription_queue`
2. **Downloading audio** from MinIO storage
3. **Processing** with Faster-Whisper
4. **Publishing results** to `transcription_results` queue
5. **Uploading transcripts** back to MinIO

## Development

### Local Development
```bash
# Install dependencies
pip install -r requirements.txt

# Copy and edit environment
cp .env.example .env

# Run service
python main.py
```

### Update Dependencies
```bash
pip freeze > requirements.txt
```

## Building and Publishing

### Using Make (Recommended)

```bash
# Interactive build and push (asks for confirmation)
make build-and-push

# Non-interactive build and push
make publish

# Specify custom version
make publish VERSION=v1.0.0

# Legacy command (alias for publish)
make push
```

### Using Build Script Directly

```bash
# Make script executable (first time only)
chmod +x build-and-push.sh

# Build and optionally push
./build-and-push.sh

# Build with specific version
./build-and-push.sh v1.0.0
```

The script will:
1. Build the Docker image (~15-20 min due to Whisper model download)
2. Tag with both version and `latest`
3. Ask if you want to push to Docker Hub
4. Show image size after completion

### CI/CD Integration

#### GitHub Actions

The `.github/workflows/transcription-service.yml` workflow automatically:
- Builds on push to `main` (when files in `transcription-service3/` change)
- Builds on tags matching `transcription-v*`
- Pushes to Docker Hub: `sivanov2018/recontext-transcription-service3`

**Trigger build:**
```bash
# Push changes
git add transcription-service3/
git commit -m "Update transcription service"
git push

# Or create a tag
git tag transcription-v1.0.0
git push origin transcription-v1.0.0
```

#### GitLab CI/CD

The `.gitlab-ci.yml` pipeline includes:
- **Build stage**: Builds image and saves to artifacts
- **Publish stage**: Pushes to Docker Hub (auto on main/tags)
- **Manual publish**: For testing in merge requests

**Required GitLab CI/CD Variables:**
- `DOCKER_HUB_USERNAME`: Your Docker Hub username
- `DOCKER_HUB_PASSWORD`: Your Docker Hub access token

**Trigger build:**
```bash
# Push to main or create tag
git push origin main

# Or create version tag
git tag transcription-v1.0.0
git push origin transcription-v1.0.0
```

### Published Images

Images are available at: `docker pull sivanov2018/recontext-transcription-service3:latest`

Tags:
- `latest` - Latest build from main branch
- `vX.Y.Z` - Semantic version tags
- `transcription-vX.Y.Z` - Service-specific version tags
- `<commit-sha>` - Specific commit builds

## License

Part of the Recontext platform.

## Support

For issues and questions, please refer to the main Recontext documentation.

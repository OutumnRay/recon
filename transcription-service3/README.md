# Transcription Service 3

AI-powered transcription service using Faster-Whisper for the Recontext platform. This service processes audio/video files from RabbitMQ queues and stores transcription results.

## Features

- **GPU-accelerated transcription** using NVIDIA CUDA
- **Faster-Whisper** for optimized inference
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

### 1. Using Standalone Docker Compose (with included RabbitMQ and MinIO)

```bash
# Copy environment file
cp .env.example .env

# Edit .env with your configuration
nano .env

# Build and start services
docker-compose up -d

# View logs
docker-compose logs -f transcription-service3
```

### 2. Using Integrated Mode (with existing Recontext infrastructure)

```bash
# Copy environment file
cp .env.example .env

# Edit .env to point to existing services
nano .env

# Build and start only the transcription service
docker-compose -f docker-compose.integrated.yml up -d

# View logs
docker-compose -f docker-compose.integrated.yml logs -f
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
WHISPER_MODEL=medium            # Model size: tiny, base, small, medium, large-v2, large-v3
WHISPER_DEVICE=cuda             # Device: cuda, cpu
WHISPER_COMPUTE_TYPE=float16    # Compute type: float16, int8, float32
```

Available models (size vs accuracy):
- `tiny` - Fastest, lowest accuracy (~1GB VRAM)
- `base` - Fast, low accuracy (~1.5GB VRAM)
- `small` - Balanced (~2GB VRAM)
- `medium` - Good accuracy (~5GB VRAM) ⭐ Recommended
- `large-v2` - High accuracy (~10GB VRAM)
- `large-v3` - Highest accuracy (~10GB VRAM)

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

## License

Part of the Recontext platform.

## Support

For issues and questions, please refer to the main Recontext documentation.

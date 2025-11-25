# Quick Start Guide - Transcription Service 3

## 🚀 Fastest Way to Start

```bash
cd transcription-service3

# Start the service (integrated mode - uses existing RabbitMQ/MinIO)
./docker-start.sh

# View logs
./docker-start.sh integrated logs

# Check status
./docker-start.sh integrated status

# Stop service
./docker-start.sh integrated stop
```

## 📋 Common Commands

### Using docker-start.sh (Simplest)

```bash
# Start commands
./docker-start.sh                      # Start integrated mode (default)
./docker-start.sh standalone           # Start with local RabbitMQ/MinIO
./docker-start.sh integrated start     # Explicit integrated start

# Management
./docker-start.sh integrated stop      # Stop service
./docker-start.sh integrated restart   # Restart service
./docker-start.sh integrated logs      # View logs (follow mode)
./docker-start.sh integrated status    # Show status and resource usage
./docker-start.sh integrated shell     # Open bash shell in container

# Help
./docker-start.sh help                 # Show all options
```

### Using Make (Alternative)

```bash
# Integrated mode (uses existing infrastructure)
make docker-start         # Start
make docker-stop          # Stop
make docker-logs          # View logs
make docker-status        # Check status

# Standalone mode (includes RabbitMQ and MinIO)
make docker-start-standalone  # Start standalone
make up                      # Alternative: docker-compose up
make down                    # Stop all
```

### Using Docker Compose (Direct)

```bash
# Integrated mode
docker-compose -f docker-compose.integrated.yml up -d
docker-compose -f docker-compose.integrated.yml logs -f
docker-compose -f docker-compose.integrated.yml down

# Standalone mode
docker-compose up -d
docker-compose logs -f transcription-service3
docker-compose down
```

## 🔧 Configuration

### Environment Variables (.env file)

The service will auto-create `.env` from `.env.example` on first start. Key variables:

```bash
# RabbitMQ (required)
RABBITMQ_HOST=192.168.5.153          # Your RabbitMQ host
RABBITMQ_PORT=5672
RABBITMQ_USER=recontext
RABBITMQ_PASSWORD=your_password

# MinIO/S3 (required)
MINIO_ENDPOINT=192.168.5.153:9000    # Your MinIO endpoint
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=your_secret
MINIO_BUCKET=recontext

# Whisper Model (pre-downloaded in image)
WHISPER_MODEL=large-v3               # Pre-installed
WHISPER_DEVICE=cuda                  # or 'cpu' for CPU mode
WHISPER_COMPUTE_TYPE=float16         # or 'float32' for CPU
```

### Edit configuration:

```bash
# Copy and edit
cp .env.example .env
nano .env  # or vim, code, etc.
```

## 🐛 Troubleshooting

### Check if service is running:
```bash
./docker-start.sh integrated status
# or
docker ps | grep transcription
```

### View logs:
```bash
./docker-start.sh integrated logs
# or
docker logs -f recontext-transcription-service3
```

### Check GPU availability:
```bash
docker exec recontext-transcription-service3 nvidia-smi
# or
make gpu-check
```

### Restart service:
```bash
./docker-start.sh integrated restart
# or
docker restart recontext-transcription-service3
```

### Full reset (rebuild):
```bash
# Stop service
./docker-start.sh integrated stop

# Remove containers and images
docker-compose -f docker-compose.integrated.yml down -v
docker rmi sivanov2018/recontext-transcription-service3:latest

# Start fresh
./docker-start.sh integrated start
```

## 📊 Monitoring

### Resource usage:
```bash
./docker-start.sh integrated status
# or
docker stats recontext-transcription-service3
```

### GPU monitoring (if using GPU):
```bash
# Inside container
docker exec recontext-transcription-service3 nvidia-smi

# Continuous monitoring
watch -n 1 'docker exec recontext-transcription-service3 nvidia-smi'
```

### Health check:
```bash
docker inspect --format='{{json .State.Health}}' recontext-transcription-service3 | jq
```

## 🎯 Deployment Modes

### Integrated Mode (Production)
- Connects to existing RabbitMQ and MinIO
- Lightweight (only transcription service)
- Recommended for production with existing infrastructure

```bash
./docker-start.sh integrated start
```

### Standalone Mode (Development/Testing)
- Includes local RabbitMQ and MinIO
- Self-contained for testing
- Requires more resources

```bash
./docker-start.sh standalone start
```

## 📦 Image Information

**Docker Image**: `sivanov2018/recontext-transcription-service3:latest`

**Size**: ~6GB (includes Whisper large-v3 model pre-downloaded)

**Pull latest**:
```bash
docker pull sivanov2018/recontext-transcription-service3:latest
```

## 🔗 Access Points (Standalone Mode)

When using standalone mode:

- **RabbitMQ Management**: http://localhost:15672
  - Username: `guest`
  - Password: `guest`

- **MinIO Console**: http://localhost:9001
  - Username: `minioadmin`
  - Password: `minioadmin`

## 📚 More Information

- Full documentation: [README.md](README.md)
- Build instructions: [README.md#building-and-publishing](README.md#building-and-publishing)
- Configuration details: [README.md#configuration](README.md#configuration)

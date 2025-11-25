#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Recontext Transcription Service 3 - Startup Script ===${NC}"
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}Warning: .env file not found. Creating from .env.example...${NC}"
    if [ -f .env.example ]; then
        cp .env.example .env
        echo -e "${GREEN}Created .env file. Please review and update the configuration.${NC}"
        echo -e "${YELLOW}Press Enter to continue or Ctrl+C to abort...${NC}"
        read
    else
        echo -e "${RED}Error: .env.example not found!${NC}"
        exit 1
    fi
fi

# Check if docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: Docker is not installed!${NC}"
    exit 1
fi

# Check if docker-compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}Error: Docker Compose is not installed!${NC}"
    exit 1
fi

# Ask user which mode to use
echo ""
echo "Select deployment mode:"
echo "1) Standalone (includes RabbitMQ and MinIO)"
echo "2) Integrated (uses existing Recontext infrastructure)"
echo ""
read -p "Enter your choice (1 or 2): " choice

case $choice in
    1)
        COMPOSE_FILE="docker-compose.yml"
        MODE="standalone"
        echo -e "${GREEN}Using standalone mode${NC}"
        ;;
    2)
        COMPOSE_FILE="docker-compose.integrated.yml"
        MODE="integrated"
        echo -e "${GREEN}Using integrated mode${NC}"
        ;;
    *)
        echo -e "${RED}Invalid choice!${NC}"
        exit 1
        ;;
esac

# Check for GPU support
echo ""
echo -e "${YELLOW}Checking for GPU support...${NC}"
if docker run --rm --gpus all nvcr.io/nvidia/pytorch:25.04-py3 nvidia-smi &> /dev/null; then
    echo -e "${GREEN}✓ GPU support detected${NC}"
    USE_GPU=true
else
    echo -e "${YELLOW}⚠ GPU not detected. Service will run in CPU mode.${NC}"
    echo -e "${YELLOW}Note: CPU mode is significantly slower than GPU mode.${NC}"
    USE_GPU=false

    # Warn user about performance
    echo ""
    read -p "Continue without GPU? (y/n): " continue_cpu
    if [ "$continue_cpu" != "y" ] && [ "$continue_cpu" != "Y" ]; then
        echo -e "${RED}Aborted.${NC}"
        exit 1
    fi

    # Update .env to use CPU
    if grep -q "WHISPER_DEVICE=" .env; then
        sed -i.bak 's/WHISPER_DEVICE=.*/WHISPER_DEVICE=cpu/' .env
        sed -i.bak 's/WHISPER_COMPUTE_TYPE=.*/WHISPER_COMPUTE_TYPE=float32/' .env
        echo -e "${GREEN}Updated .env to use CPU mode${NC}"
    fi
fi

# Pull the base image
echo ""
echo -e "${YELLOW}Pulling base image (this may take a while)...${NC}"
docker pull nvcr.io/nvidia/pytorch:25.04-py3

# Build the service
echo ""
echo -e "${YELLOW}Building transcription service...${NC}"
docker-compose -f "$COMPOSE_FILE" build

# Start the services
echo ""
echo -e "${YELLOW}Starting services...${NC}"
docker-compose -f "$COMPOSE_FILE" up -d

# Wait a moment for services to start
sleep 5

# Show status
echo ""
echo -e "${GREEN}=== Service Status ===${NC}"
docker-compose -f "$COMPOSE_FILE" ps

# Show logs
echo ""
echo -e "${GREEN}=== Recent Logs ===${NC}"
docker-compose -f "$COMPOSE_FILE" logs --tail=20 transcription-service3

# Final instructions
echo ""
echo -e "${GREEN}=== Transcription Service Started ===${NC}"
echo ""
echo "View logs:"
echo "  docker-compose -f $COMPOSE_FILE logs -f transcription-service3"
echo ""
echo "Stop service:"
echo "  docker-compose -f $COMPOSE_FILE down"
echo ""
echo "Restart service:"
echo "  docker-compose -f $COMPOSE_FILE restart transcription-service3"
echo ""

if [ "$MODE" = "standalone" ]; then
    echo "Access points:"
    echo "  RabbitMQ Management: http://localhost:15672 (guest/guest)"
    echo "  MinIO Console: http://localhost:9001 (minioadmin/minioadmin)"
    echo ""
fi

if [ "$USE_GPU" = true ]; then
    echo "GPU Status:"
    docker exec recontext-transcription-service3 nvidia-smi || echo "Unable to check GPU status"
    echo ""
fi

echo -e "${GREEN}✓ Setup complete!${NC}"

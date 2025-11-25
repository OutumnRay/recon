#!/bin/bash

# Quick Docker Start Script for Transcription Service 3
# This script provides simple commands to start the transcription service

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Default values
MODE="${1:-integrated}"  # Default to integrated mode
ACTION="${2:-start}"      # Default action is start

usage() {
    echo "Usage: $0 [mode] [action]"
    echo ""
    echo "Modes:"
    echo "  standalone   - Start with RabbitMQ and MinIO (default for testing)"
    echo "  integrated   - Connect to existing infrastructure (default)"
    echo ""
    echo "Actions:"
    echo "  start        - Start services (default)"
    echo "  stop         - Stop services"
    echo "  restart      - Restart services"
    echo "  logs         - Show logs"
    echo "  status       - Show status"
    echo "  shell        - Open shell in container"
    echo ""
    echo "Examples:"
    echo "  $0                          # Start in integrated mode"
    echo "  $0 standalone               # Start in standalone mode"
    echo "  $0 integrated logs          # Show logs"
    echo "  $0 integrated stop          # Stop services"
}

# Parse command
if [ "$MODE" = "help" ] || [ "$MODE" = "-h" ] || [ "$MODE" = "--help" ]; then
    usage
    exit 0
fi

# Determine compose file
if [ "$MODE" = "standalone" ]; then
    COMPOSE_FILE="docker-compose.yml"
    SERVICE_NAME="transcription-service3"
    CONTAINER_NAME="transcription-service3"
elif [ "$MODE" = "integrated" ]; then
    COMPOSE_FILE="docker-compose.integrated.yml"
    SERVICE_NAME="transcription-service3"
    CONTAINER_NAME="recontext-transcription-service3"
else
    echo -e "${RED}Error: Invalid mode '$MODE'${NC}"
    usage
    exit 1
fi

# Execute action
case "$ACTION" in
    start)
        echo -e "${GREEN}Starting transcription service ($MODE mode)...${NC}"

        # Check if .env exists
        if [ ! -f .env ]; then
            echo -e "${YELLOW}Creating .env from .env.example...${NC}"
            cp .env.example .env
            echo -e "${GREEN}Created .env file. Review settings if needed.${NC}"
        fi

        # Start services
        docker-compose -f "$COMPOSE_FILE" up -d

        echo -e "${GREEN}✓ Service started!${NC}"
        echo ""
        echo "View logs: $0 $MODE logs"
        echo "Stop service: $0 $MODE stop"
        ;;

    stop)
        echo -e "${YELLOW}Stopping transcription service...${NC}"
        docker-compose -f "$COMPOSE_FILE" down
        echo -e "${GREEN}✓ Service stopped${NC}"
        ;;

    restart)
        echo -e "${YELLOW}Restarting transcription service...${NC}"
        docker-compose -f "$COMPOSE_FILE" restart "$SERVICE_NAME"
        echo -e "${GREEN}✓ Service restarted${NC}"
        ;;

    logs)
        echo -e "${GREEN}Showing logs (Ctrl+C to exit)...${NC}"
        docker-compose -f "$COMPOSE_FILE" logs -f "$SERVICE_NAME"
        ;;

    status)
        echo -e "${GREEN}Service Status:${NC}"
        docker-compose -f "$COMPOSE_FILE" ps
        echo ""

        # Check if container is running
        if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
            echo -e "${GREEN}✓ Container is running${NC}"
            echo ""
            echo "Resource usage:"
            docker stats "$CONTAINER_NAME" --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}"
        else
            echo -e "${RED}✗ Container is not running${NC}"
        fi
        ;;

    shell)
        echo -e "${GREEN}Opening shell in container...${NC}"
        docker exec -it "$CONTAINER_NAME" /bin/bash
        ;;

    *)
        echo -e "${RED}Error: Unknown action '$ACTION'${NC}"
        usage
        exit 1
        ;;
esac

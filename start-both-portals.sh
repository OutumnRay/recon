#!/bin/bash

# Recontext Both Portals Launcher
# This script automatically detects your OS and uses the appropriate docker-compose file

set -e

# Detect OS
if [[ "$OSTYPE" == "darwin"* ]]; then
    COMPOSE_FILE="docker-compose-both-portals-mac.yml"
    echo "Detected macOS - using $COMPOSE_FILE"
else
    COMPOSE_FILE="docker-compose-both-portals.yml"
    echo "Detected Linux - using $COMPOSE_FILE"
fi

# Check if .env file exists
if [ ! -f .env.both-portals ]; then
    echo "Warning: .env.both-portals not found. Using default values."
    echo "Consider creating .env.both-portals from .env.both-portals.example"
fi

# Parse command
case "${1:-up}" in
    up)
        echo "Starting Recontext Both Portals..."
        docker-compose -f "$COMPOSE_FILE" --env-file .env.both-portals up -d
        echo ""
        echo "Both Portals are starting up..."
        echo ""
        echo "Services:"
        echo "  - Managing Portal: http://localhost:10080"
        echo "  - User Portal:     http://localhost:10081"
        echo "  - MinIO Console:   http://localhost:9001"
        echo "  - PostgreSQL:      localhost:5432"
        echo ""
        echo "Default credentials:"
        echo "  - Managing Portal: admin / admin123"
        echo "  - User Portal:     user / user123"
        echo "  - MinIO:          minioadmin / minioadmin"
        echo ""
        echo "Use './start-both-portals.sh logs' to view logs"
        echo "Use './start-both-portals.sh down' to stop"
        ;;
    down)
        echo "Stopping Recontext Both Portals..."
        docker-compose -f "$COMPOSE_FILE" down
        ;;
    restart)
        echo "Restarting Recontext Both Portals..."
        docker-compose -f "$COMPOSE_FILE" restart
        ;;
    logs)
        docker-compose -f "$COMPOSE_FILE" logs -f "${2:-}"
        ;;
    build)
        echo "Building Both Portals..."
        docker-compose -f "$COMPOSE_FILE" build --no-cache
        ;;
    clean)
        echo "Cleaning up Both Portals (removing volumes)..."
        docker-compose -f "$COMPOSE_FILE" down -v
        echo "All data has been removed."
        ;;
    ps)
        docker-compose -f "$COMPOSE_FILE" ps
        ;;
    *)
        echo "Usage: $0 {up|down|restart|logs|build|clean|ps}"
        echo ""
        echo "Commands:"
        echo "  up      - Start all services (default)"
        echo "  down    - Stop all services"
        echo "  restart - Restart all services"
        echo "  logs    - View logs (use 'logs <service>' for specific service)"
        echo "  build   - Rebuild images"
        echo "  clean   - Stop and remove all data"
        echo "  ps      - Show running services"
        exit 1
        ;;
esac

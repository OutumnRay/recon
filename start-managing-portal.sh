#!/bin/bash

# Recontext Managing Portal Launcher
# This script automatically detects your OS and uses the appropriate docker-compose file

set -e

# Detect OS
if [[ "$OSTYPE" == "darwin"* ]]; then
    COMPOSE_FILE="docker-compose-managing-portal-mac.yml"
    echo "Detected macOS - using $COMPOSE_FILE"
else
    COMPOSE_FILE="docker-compose-managing-portal.yml"
    echo "Detected Linux - using $COMPOSE_FILE"
fi

# Check if .env file exists
if [ ! -f .env.managing-portal ]; then
    echo "Warning: .env.managing-portal not found. Using default values."
    echo "Consider creating .env.managing-portal from .env.managing-portal.example"
fi

# Parse command
case "${1:-up}" in
    up)
        echo "Starting Recontext Managing Portal..."
        docker-compose -f "$COMPOSE_FILE" --env-file .env.managing-portal up -d
        echo ""
        echo "Managing Portal is starting up..."
        echo ""
        echo "Services:"
        echo "  - Managing Portal: http://localhost:10080"
        echo "  - MinIO Console:   http://localhost:9001"
        echo "  - PostgreSQL:      localhost:5432"
        echo ""
        echo "Default credentials:"
        echo "  - Managing Portal: admin / admin123"
        echo "  - MinIO:          minioadmin / minioadmin"
        echo ""
        echo "Use './start-managing-portal.sh logs' to view logs"
        echo "Use './start-managing-portal.sh down' to stop"
        ;;
    down)
        echo "Stopping Recontext Managing Portal..."
        docker-compose -f "$COMPOSE_FILE" down
        ;;
    restart)
        echo "Restarting Recontext Managing Portal..."
        docker-compose -f "$COMPOSE_FILE" restart
        ;;
    logs)
        docker-compose -f "$COMPOSE_FILE" logs -f "${2:-}"
        ;;
    build)
        echo "Building Managing Portal..."
        docker-compose -f "$COMPOSE_FILE" build --no-cache
        ;;
    clean)
        echo "Cleaning up Managing Portal (removing volumes)..."
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

#!/bin/bash

# Detect OS
if [[ "$OSTYPE" == "darwin"* ]]; then
    COMPOSE_FILE="docker-compose-user-portal-mac.yml"
    echo "🍎 Detected macOS - using $COMPOSE_FILE"
else
    COMPOSE_FILE="docker-compose-user-portal.yml"
    echo "🐧 Detected Linux - using $COMPOSE_FILE"
fi

# Parse command line arguments
ACTION="${1:-up}"

case "$ACTION" in
    up|start)
        echo "🚀 Starting User Portal development environment..."
        docker-compose -f "$COMPOSE_FILE" up -d
        echo ""
        echo "✅ Services started!"
        echo ""
        echo "📍 Access points:"
        echo "   User Portal:    http://localhost:10081"
        echo "   MinIO Console:  http://localhost:9001 (minioadmin/minioadmin)"
        echo "   PostgreSQL:     localhost:5432 (recontext/recontext)"
        echo ""
        echo "👤 Default login: user / user123"
        echo ""
        echo "📝 View logs with: ./start-user-portal.sh logs"
        ;;

    down|stop)
        echo "🛑 Stopping User Portal development environment..."
        docker-compose -f "$COMPOSE_FILE" down
        echo "✅ Services stopped!"
        ;;

    restart)
        echo "🔄 Restarting User Portal development environment..."
        docker-compose -f "$COMPOSE_FILE" restart
        echo "✅ Services restarted!"
        ;;

    logs)
        docker-compose -f "$COMPOSE_FILE" logs -f user-portal
        ;;

    build)
        echo "🔨 Building User Portal..."
        docker-compose -f "$COMPOSE_FILE" up --build -d user-portal
        echo "✅ User Portal rebuilt and started!"
        ;;

    clean)
        echo "🧹 Cleaning up (this will delete all data)..."
        read -p "Are you sure? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            docker-compose -f "$COMPOSE_FILE" down -v
            echo "✅ All services and volumes removed!"
        else
            echo "❌ Cancelled"
        fi
        ;;

    ps|status)
        docker-compose -f "$COMPOSE_FILE" ps
        ;;

    *)
        echo "Usage: $0 {up|down|restart|logs|build|clean|ps}"
        echo ""
        echo "Commands:"
        echo "  up       - Start all services"
        echo "  down     - Stop all services"
        echo "  restart  - Restart all services"
        echo "  logs     - View user portal logs"
        echo "  build    - Rebuild and restart user portal"
        echo "  clean    - Stop and remove all data (requires confirmation)"
        echo "  ps       - Show service status"
        exit 1
        ;;
esac

#!/bin/bash

# LiveKit Video Conference - Deployment Script
# This script deploys the application to production server

set -e

echo "🚀 LiveKit Deployment Script"
echo "============================"

# Configuration
PROJECT_DIR="/var/www/html/livekit"
BACKUP_DIR="/backup/livekit"
DATE=$(date +%Y%m%d_%H%M%S)

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running on server
if [ ! -d "$PROJECT_DIR" ]; then
    log_error "Project directory not found: $PROJECT_DIR"
    log_info "Please run this script on the production server"
    exit 1
fi

# Navigate to project directory
cd $PROJECT_DIR || exit 1
log_info "Changed to directory: $PROJECT_DIR"

# Create backup
log_info "Creating backup..."
mkdir -p $BACKUP_DIR
if [ -d "dist" ]; then
    tar -czf $BACKUP_DIR/dist-backup-$DATE.tar.gz dist/ || log_warn "Backup failed, continuing anyway..."
    log_info "Backup created: dist-backup-$DATE.tar.gz"
fi

# Pull latest changes (if using git)
if [ -d ".git" ]; then
    log_info "Pulling latest changes from git..."
    git pull || {
        log_error "Git pull failed"
        exit 1
    }
else
    log_warn "Not a git repository, skipping git pull"
fi

# Install/Update dependencies
log_info "Installing dependencies..."
npm install --production || {
    log_error "npm install failed"
    exit 1
}

# Build project
log_info "Building project..."
npm run build || {
    log_error "Build failed"
    log_info "Restoring from backup..."
    if [ -f "$BACKUP_DIR/dist-backup-$DATE.tar.gz" ]; then
        tar -xzf $BACKUP_DIR/dist-backup-$DATE.tar.gz
    fi
    exit 1
}

# Set proper permissions
log_info "Setting permissions..."
chmod -R 755 dist/
chmod 600 .env 2>/dev/null || log_warn ".env file not found"

# Restart backend with PM2
if command -v pm2 &> /dev/null; then
    log_info "Restarting backend with PM2..."
    pm2 restart livekit-backend || {
        log_warn "PM2 restart failed, trying to start..."
        pm2 start server.js --name livekit-backend
    }
    pm2 save
else
    log_warn "PM2 not found, skipping backend restart"
fi

# Reload Nginx
if command -v nginx &> /dev/null; then
    log_info "Reloading Nginx..."
    sudo systemctl reload nginx || log_warn "Nginx reload failed"
else
    log_warn "Nginx not found, skipping reload"
fi

# Clean old backups (keep last 10)
log_info "Cleaning old backups..."
cd $BACKUP_DIR
ls -t dist-backup-*.tar.gz 2>/dev/null | tail -n +11 | xargs -r rm
cd $PROJECT_DIR

# Show status
echo ""
log_info "Deployment completed successfully!"
echo ""
echo "Status:"
echo "-------"

if command -v pm2 &> /dev/null; then
    pm2 status
fi

echo ""
log_info "Access your application at your configured domain"
log_info "Check logs with: pm2 logs livekit-backend"

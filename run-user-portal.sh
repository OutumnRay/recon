#!/bin/bash

# Run User Portal with default settings

export DB_HOST="${DB_HOST:-localhost}"
export DB_PORT="${DB_PORT:-5432}"
export DB_USER="${DB_USER:-recontext}"
export DB_PASSWORD="${DB_PASSWORD:-recontext}"
export DB_NAME="${DB_NAME:-recontext}"
export DB_SSL_MODE="${DB_SSL_MODE:-disable}"
export JWT_SECRET="${JWT_SECRET:-your-secret-key-change-in-production}"

# Create uploads directory if it doesn't exist
mkdir -p uploads/avatars
chmod 755 uploads/avatars

echo "Starting User Portal on http://localhost:8081"
echo "Uploads directory: $(pwd)/uploads/avatars"
echo ""

./user-portal

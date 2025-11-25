#!/bin/bash

# Build and Push Script for Transcription Service 3
# Usage: ./build-and-push.sh [tag]
# Example: ./build-and-push.sh v1.0.0

set -e

# Configuration
DOCKER_IMAGE="sivanov2018/recontext-transcription-service"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors for output
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

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    log_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Get version tag
if [ -n "$1" ]; then
    VERSION_TAG="$1"
else
    # Try to get from git tag
    VERSION_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    if [ -z "$VERSION_TAG" ]; then
        # Use commit hash as fallback
        VERSION_TAG="$(git rev-parse --short HEAD)"
        log_warn "No version tag provided or found. Using commit hash: $VERSION_TAG"
    else
        log_info "Using git tag: $VERSION_TAG"
    fi
fi

# Build image tags
IMAGE_WITH_VERSION="${DOCKER_IMAGE}:${VERSION_TAG}"
IMAGE_LATEST="${DOCKER_IMAGE}:latest"

log_info "Building Docker image..."
log_info "Image: $IMAGE_WITH_VERSION"
log_info "This will take ~15-20 minutes due to Whisper model download (~3GB)"
echo ""

# Build the image
docker build \
    -t "$IMAGE_WITH_VERSION" \
    -t "$IMAGE_LATEST" \
    -f "$SCRIPT_DIR/Dockerfile" \
    "$SCRIPT_DIR"

if [ $? -eq 0 ]; then
    log_info "✅ Build successful!"
else
    log_error "❌ Build failed!"
    exit 1
fi

# Ask for confirmation before pushing
echo ""
read -p "Do you want to push the image to Docker Hub? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_info "Logging in to Docker Hub..."
    docker login

    if [ $? -eq 0 ]; then
        log_info "Pushing image with tag: $VERSION_TAG"
        docker push "$IMAGE_WITH_VERSION"

        log_info "Pushing image with tag: latest"
        docker push "$IMAGE_LATEST"

        log_info "✅ Push successful!"
        echo ""
        log_info "Published images:"
        log_info "  - $IMAGE_WITH_VERSION"
        log_info "  - $IMAGE_LATEST"
    else
        log_error "❌ Docker login failed!"
        exit 1
    fi
else
    log_info "Skipping push. Image built locally:"
    log_info "  - $IMAGE_WITH_VERSION"
    log_info "  - $IMAGE_LATEST"
fi

echo ""
log_info "Image size:"
docker images "$DOCKER_IMAGE" --format "table {{.Repository}}:{{.Tag}}\t{{.Size}}" | head -n 3

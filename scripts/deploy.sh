#!/bin/bash
# Скрипт деплоя на выделенный сервер.
# Запускается автоматически из GitHub Actions по SSH после успешного билда.
# Можно также запустить вручную: bash /opt/recontext/scripts/deploy.sh

set -euo pipefail

DEPLOY_DIR="${DEPLOY_DIR:-/opt/recontext}"
COMPOSE_FILE="docker-compose.prod.yml"
LOG_FILE="/var/log/recontext-deploy.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

log "=== Начало деплоя ==="

cd "$DEPLOY_DIR"

log "Тянем свежие образы из Docker Hub..."
docker compose -f "$COMPOSE_FILE" pull

log "Перезапускаем сервисы (zero-downtime для stateless-контейнеров)..."
docker compose -f "$COMPOSE_FILE" up -d --remove-orphans

log "Удаляем устаревшие образы..."
docker image prune -f

log "Статус контейнеров:"
docker compose -f "$COMPOSE_FILE" ps

log "=== Деплой завершён успешно ==="

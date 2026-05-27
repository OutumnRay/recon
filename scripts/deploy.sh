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

log "Тянем актуальную конфигурацию из git..."
git pull origin main || git pull new-origin main || log "ПРЕДУПРЕЖДЕНИЕ: git pull не удался, продолжаем с локальной конфигурацией"

log "Тянем свежие образы из Docker Hub..."
docker compose -f "$COMPOSE_FILE" pull

log "Перезапускаем сервисы (zero-downtime для stateless-контейнеров)..."
docker compose -f "$COMPOSE_FILE" up -d --remove-orphans

log "Проверяем наличие модели Ollama..."
OLLAMA_MODEL="${LLM_MODEL:-qwen2.5:3b}"
if ! docker exec recontext-ollama ollama list 2>/dev/null | grep -q "${OLLAMA_MODEL%%:*}"; then
    log "Модель $OLLAMA_MODEL не найдена, скачиваем..."
    docker exec recontext-ollama ollama pull "$OLLAMA_MODEL" || log "ПРЕДУПРЕЖДЕНИЕ: не удалось скачать модель $OLLAMA_MODEL"
else
    log "Модель $OLLAMA_MODEL уже загружена"
fi

log "Удаляем устаревшие образы..."
docker image prune -f

log "Статус контейнеров:"
docker compose -f "$COMPOSE_FILE" ps

log "=== Деплой завершён успешно ==="

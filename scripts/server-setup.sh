#!/bin/bash
# =============================================================================
# Первичная настройка чистого сервера Recontext (Ubuntu 22.04/24.04, Debian 12).
# Запуск под root или через sudo:
#   curl -fsSL https://raw.githubusercontent.com/OutumnRay/recon/main/scripts/server-setup.sh | sudo bash
# или из уже склонированного репозитория:
#   sudo bash /opt/recontext/scripts/server-setup.sh
#
# Что делает (идемпотентно, можно запускать повторно):
#   1. ставит Docker, git, nginx, certbot, ufw
#   2. клонирует репозиторий в /opt/recontext
#   3. создаёт .env из .env.prod.example со случайными секретами (если .env ещё нет)
#   4. раскладывает конфиги host-nginx для домена
#   5. настраивает ufw
#   6. даёт пользователю DEPLOY_USER права на docker и на каталог проекта
# Запуск самих контейнеров — отдельно: bash /opt/recontext/scripts/deploy.sh
# =============================================================================
set -euo pipefail

DEPLOY_DIR="${DEPLOY_DIR:-/opt/recontext}"
REPO_URL="${REPO_URL:-https://github.com/OutumnRay/recon.git}"
BRANCH="${BRANCH:-main}"
DOMAIN="${DOMAIN:-24recontext.ru}"
DEPLOY_USER="${DEPLOY_USER:-${SUDO_USER:-root}}"
LOG_FILE="/var/log/recontext-deploy.log"

log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"; }

if [ "$(id -u)" -ne 0 ]; then
    echo "Запустите через sudo" >&2
    exit 1
fi

log "=== Настройка сервера Recontext (домен: $DOMAIN, пользователь деплоя: $DEPLOY_USER) ==="

# 1. Пакеты
log "Устанавливаем пакеты..."
apt-get update -y
apt-get install -y ca-certificates curl git nginx certbot python3-certbot-nginx ufw openssl

if ! command -v docker &>/dev/null; then
    log "Устанавливаем Docker..."
    curl -fsSL https://get.docker.com | sh
fi
systemctl enable --now docker
docker compose version

# 2. Репозиторий
if [ ! -d "$DEPLOY_DIR/.git" ]; then
    log "Клонируем $REPO_URL в $DEPLOY_DIR..."
    git clone -b "$BRANCH" "$REPO_URL" "$DEPLOY_DIR"
else
    log "Репозиторий уже есть, обновляем..."
    git -C "$DEPLOY_DIR" pull origin "$BRANCH"
fi

# 3. .env со случайными секретами
ENV_FILE="$DEPLOY_DIR/.env"
if [ ! -f "$ENV_FILE" ]; then
    log "Создаём $ENV_FILE со случайными секретами..."
    while IFS= read -r line || [ -n "$line" ]; do
        # hex-секреты безопасны внутри URL (amqp://, postgresql://)
        echo "${line//__GENERATE__/$(openssl rand -hex 24)}"
    done < "$DEPLOY_DIR/.env.prod.example" > "$ENV_FILE"
    sed -i "s/24recontext\.ru/$DOMAIN/g" "$ENV_FILE"
    chmod 600 "$ENV_FILE"
    GENERATED_ENV=1
else
    log "$ENV_FILE уже существует — не трогаем"
    GENERATED_ENV=0
fi

# 4. Host-nginx: копируем (не симлинк) — certbot будет править эти файлы, репозиторий должен остаться чистым
log "Раскладываем конфиги nginx..."
for name in site admin storage livekit grafana prometheus; do
    src="$DEPLOY_DIR/deployments/nginx/$name.conf"
    dst="/etc/nginx/sites-available/recontext-$name.conf"
    if [ ! -f "$dst" ]; then
        sed "s/24recontext\.ru/$DOMAIN/g" "$src" > "$dst"
        ln -sf "$dst" "/etc/nginx/sites-enabled/recontext-$name.conf"
    fi
done
rm -f /etc/nginx/sites-enabled/default
if [ ! -f /etc/nginx/.htpasswd-prometheus ]; then
    PROM_PASS="$(openssl rand -hex 12)"
    printf 'admin:%s\n' "$(openssl passwd -apr1 "$PROM_PASS")" > /etc/nginx/.htpasswd-prometheus
    log "Prometheus basic auth: admin / $PROM_PASS"
fi
nginx -t && systemctl reload nginx

# 5. Файрвол: наружу только SSH, HTTP(S) и медиа LiveKit
log "Настраиваем ufw..."
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 7881/tcp    # LiveKit WebRTC TCP
ufw allow 7882/udp    # LiveKit WebRTC UDP
ufw --force enable

# 6. Права для пользователя деплоя (GitHub Actions заходит под ним по SSH)
touch "$LOG_FILE"
if [ "$DEPLOY_USER" != "root" ]; then
    usermod -aG docker "$DEPLOY_USER"
    chown -R "$DEPLOY_USER":"$DEPLOY_USER" "$DEPLOY_DIR"
    chown "$DEPLOY_USER":"$DEPLOY_USER" "$LOG_FILE"
fi

log "=== Готово ==="
if [ "$GENERATED_ENV" = "1" ]; then
    echo
    echo "Сгенерирован $ENV_FILE. Пароль администратора портала:"
    grep '^ADMIN_EMAIL=\|^ADMIN_PASSWORD=' "$ENV_FILE"
    echo "Допишите при необходимости SMTP_*, FCM_CREDENTIALS_JSON, HF_TOKEN:  nano $ENV_FILE"
fi
cat <<EOF

Дальше:
  1. Запуск стека:   sudo -u $DEPLOY_USER bash $DEPLOY_DIR/scripts/deploy.sh
  2. Когда DNS для $DOMAIN и поддоменов указывает на внешний IP и порты 80/443 проброшены:
     sudo certbot --nginx -d $DOMAIN -d www.$DOMAIN -d admin.$DOMAIN -d storage.$DOMAIN \\
          -d livekit.$DOMAIN -d grafana.$DOMAIN -d prometheus.$DOMAIN
EOF

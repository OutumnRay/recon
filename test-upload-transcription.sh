#!/usr/bin/env bash
# test-upload-transcription.sh
# Полный тест загрузки видео и запуска транскрибации.
#
# Использование:
#   ./test-upload-transcription.sh [путь/к/видео.mp4]
#
# Если видео не указано — создаётся мини-тестовый файл через ffmpeg (если установлен).
# Переменные окружения (можно переопределить):
#   API_URL     — базовый URL user-portal (по умолчанию http://localhost:20081)
#   ADMIN_URL   — базовый URL managing-portal (по умолчанию http://localhost:20080)
#   LOGIN       — логин пользователя (по умолчанию user@example.com)
#   PASSWORD    — пароль пользователя (по умолчанию user123)
#   ADMIN_LOGIN — логин администратора (по умолчанию admin@example.com)
#   ADMIN_PASS  — пароль администратора (по умолчанию admin123)
#   REDIS_HOST  — хост Redis (по умолчанию localhost)
#   REDIS_PORT  — порт Redis (по умолчанию 6380)
#   MINIO_URL   — URL MinIO (по умолчанию http://localhost:9000)

set -euo pipefail

# ─── Настройки ────────────────────────────────────────────────────────────────
API_URL="${API_URL:-http://localhost:20081}"
ADMIN_URL="${ADMIN_URL:-http://localhost:20080}"
LOGIN="${LOGIN:-user@recontext.online}"
PASSWORD="${PASSWORD:-user123}"
ADMIN_LOGIN="${ADMIN_LOGIN:-admin@recontext.online}"
ADMIN_PASS="${ADMIN_PASS:-admin123}"
REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6380}"
MINIO_URL="${MINIO_URL:-http://localhost:9000}"
REDIS_QUEUE="${REDIS_QUEUE:-recontext:transcription:queue}"
REDIS_RESULT="${REDIS_RESULT:-recontext:transcription:results}"

VIDEO_FILE="${1:-}"

# ─── Цвета ────────────────────────────────────────────────────────────────────
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

ok()   { echo -e "${GREEN}[OK]${NC} $*"; }
fail() { echo -e "${RED}[FAIL]${NC} $*"; }
info() { echo -e "${CYAN}[INFO]${NC} $*"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }

step() {
  echo ""
  echo -e "${CYAN}══════════════════════════════════════════════${NC}"
  echo -e "${CYAN}  $*${NC}"
  echo -e "${CYAN}══════════════════════════════════════════════${NC}"
}

# ─── 1. Проверка зависимостей ─────────────────────────────────────────────────
step "Шаг 1: Проверка зависимостей"

for cmd in curl jq; do
  if command -v "$cmd" &>/dev/null; then
    ok "$cmd найден"
  else
    fail "$cmd не установлен (необходимо для тестов)"
    exit 1
  fi
done

if command -v redis-cli &>/dev/null; then
  ok "redis-cli найден"
  HAVE_REDIS_CLI=true
else
  warn "redis-cli не найден — шаги с Redis будут пропущены"
  HAVE_REDIS_CLI=false
fi

if command -v ffmpeg &>/dev/null; then
  ok "ffmpeg найден"
  HAVE_FFMPEG=true
else
  warn "ffmpeg не найден — тестовый видеофайл нужно указать вручную"
  HAVE_FFMPEG=false
fi

# ─── 2. Проверка доступности сервисов ────────────────────────────────────────
step "Шаг 2: Проверка доступности сервисов"

check_http() {
  local name="$1" url="$2"
  local code
  # curl always prints "000" via -w when connection fails — don't append || echo
  code=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 3 "$url" 2>/dev/null)
  # Trim potential CR/whitespace (Git Bash on Windows can add \r)
  code="${code//[$'\r\n ']/}"
  if [[ -n "$code" && "$code" != "000" ]]; then
    ok "$name доступен ($url) — HTTP $code"
    return 0
  else
    fail "$name недоступен ($url)"
    return 1
  fi
}

PORTAL_OK=true
ADMIN_OK=true
MINIO_OK=true

check_http "User Portal"     "$API_URL/health"  || PORTAL_OK=false
check_http "Managing Portal" "$ADMIN_URL/health" || ADMIN_OK=false
check_http "MinIO"           "$MINIO_URL/minio/health/live" || \
  check_http "MinIO"         "$MINIO_URL" || MINIO_OK=false

if [[ "$HAVE_REDIS_CLI" == "true" ]]; then
  if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ping 2>/dev/null | grep -q PONG; then
    ok "Redis доступен ($REDIS_HOST:$REDIS_PORT)"
  else
    warn "Redis недоступен ($REDIS_HOST:$REDIS_PORT) — задача не попадёт в очередь"
  fi
fi

if [[ "$PORTAL_OK" != "true" ]]; then
  fail "User Portal недоступен. Запустите его перед тестом:"
  echo "  docker-compose up -d user-portal"
  echo "  # или локально:"
  echo "  ./run-user-portal.sh"
  exit 1
fi

# ─── 3. Аутентификация ────────────────────────────────────────────────────────
step "Шаг 3: Аутентификация"

info "Вход как пользователь: $LOGIN"
USER_RESP=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$LOGIN\",\"password\":\"$PASSWORD\"}")

USER_TOKEN=$(echo "$USER_RESP" | jq -r '.token // .access_token // empty' 2>/dev/null)

if [[ -z "$USER_TOKEN" ]]; then
  warn "Не удалось войти как '$LOGIN': $(echo "$USER_RESP" | jq -r '.error // .' 2>/dev/null)"
  warn "Попробуем admin-аккаунт для загрузки файлов..."

  info "Вход как администратор: $ADMIN_LOGIN"
  ADMIN_RESP=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$ADMIN_LOGIN\",\"password\":\"$ADMIN_PASS\"}")
  USER_TOKEN=$(echo "$ADMIN_RESP" | jq -r '.token // .access_token // empty' 2>/dev/null)

  if [[ -z "$USER_TOKEN" ]]; then
    fail "Не удалось получить токен. Ответ: $(echo "$ADMIN_RESP" | jq -c .)"
    echo ""
    echo "Подсказка: проверьте логин/пароль в переменных LOGIN/PASSWORD или ADMIN_LOGIN/ADMIN_PASS"
    echo "  Пример: LOGIN=admin@example.com PASSWORD=admin123 ./test-upload-transcription.sh"
    exit 1
  fi
  ok "Вошли как администратор"
else
  ok "Вошли как пользователь"
fi

USER_ID=$(echo "${USER_RESP:-$ADMIN_RESP}" | jq -r '.user.id // empty' 2>/dev/null || echo "")
info "User ID: ${USER_ID:-<неизвестен>}"

# ─── 4. Проверка разрешения на загрузку файлов ────────────────────────────────
step "Шаг 4: Проверка разрешения на загрузку файлов"

PERM_RESP=$(curl -s "$API_URL/api/v1/files/permission" \
  -H "Authorization: Bearer $USER_TOKEN")
HAS_PERM=$(echo "$PERM_RESP" | jq -r '.hasPermission' 2>/dev/null)

if [[ "$HAS_PERM" == "true" ]]; then
  ok "Разрешение на загрузку файлов: ЕСТЬ"
else
  warn "Разрешение на загрузку файлов: НЕТ"
  info "Нужно добавить пользователя в группу с правом files:write."
  echo ""
  echo "  Вариант 1 — через Managing Portal UI (http://localhost:20080):"
  echo "    Groups → выбрать группу → добавить permission: files/write → добавить пользователя"
  echo ""
  echo "  Вариант 2 — через API (нужен admin-токен):"
  echo "    Получить admin-токен и выполнить:"
  echo "    curl -X POST $ADMIN_URL/api/v1/groups/add-user \\"
  echo "      -H 'Authorization: Bearer <admin_token>' \\"
  echo "      -H 'Content-Type: application/json' \\"
  echo "      -d '{\"user_id\":\"<user_uuid>\",\"group_id\":\"<group_uuid>\"}'"
  echo ""
  warn "Продолжаем тест, но загрузка скорее всего вернёт 403..."
fi

# ─── 5. Подготовка тестового видеофайла ──────────────────────────────────────
step "Шаг 5: Подготовка тестового видеофайла"

CLEANUP_VIDEO=false
if [[ -n "$VIDEO_FILE" && -f "$VIDEO_FILE" ]]; then
  ok "Используем указанный файл: $VIDEO_FILE ($(du -h "$VIDEO_FILE" | cut -f1))"
elif [[ "$HAVE_FFMPEG" == "true" ]]; then
  VIDEO_FILE="/tmp/test-recontext-$(date +%s).mp4"
  info "Создаём тестовое видео (5 сек, тишина + синий экран): $VIDEO_FILE"
  ffmpeg -y \
    -f lavfi -i "sine=frequency=440:duration=5" \
    -f lavfi -i "color=c=blue:size=320x240:duration=5" \
    -c:v libx264 -preset ultrafast -crf 40 \
    -c:a aac -b:a 32k \
    -shortest \
    "$VIDEO_FILE" 2>/dev/null
  ok "Тестовое видео создано: $VIDEO_FILE ($(du -h "$VIDEO_FILE" | cut -f1))"
  CLEANUP_VIDEO=true
else
  fail "Не указан путь к видео и ffmpeg не установлен."
  echo "  Укажите файл: ./test-upload-transcription.sh /path/to/video.mp4"
  echo "  Или установите ffmpeg для автогенерации тестового файла"
  exit 1
fi

# ─── 6. Загрузка видео ────────────────────────────────────────────────────────
step "Шаг 6: Загрузка видео через POST /api/v1/files/upload"

info "Загружаем файл: $VIDEO_FILE"
UPLOAD_RESP=$(curl -s -X POST "$API_URL/api/v1/files/upload" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -F "file=@$VIDEO_FILE")

echo "Ответ сервера:"
echo "$UPLOAD_RESP" | jq . 2>/dev/null || echo "$UPLOAD_RESP"

UPLOAD_ERROR=$(echo "$UPLOAD_RESP" | jq -r '.error // empty' 2>/dev/null)
if [[ -n "$UPLOAD_ERROR" ]]; then
  fail "Ошибка загрузки: $UPLOAD_ERROR"
  UPLOAD_DETAIL=$(echo "$UPLOAD_RESP" | jq -r '.details // empty' 2>/dev/null)
  [[ -n "$UPLOAD_DETAIL" ]] && echo "  Детали: $UPLOAD_DETAIL"

  if [[ "$UPLOAD_ERROR" == *"permission"* || "$UPLOAD_ERROR" == *"403"* ]]; then
    echo ""
    warn "Нет прав на загрузку файлов. Смотрите инструкцию в Шаге 4."
  fi
  [[ "$CLEANUP_VIDEO" == "true" ]] && rm -f "$VIDEO_FILE"
  exit 1
fi

FILE_ID=$(echo "$UPLOAD_RESP" | jq -r '.id // empty' 2>/dev/null)
FILE_STATUS=$(echo "$UPLOAD_RESP" | jq -r '.status // empty' 2>/dev/null)

if [[ -z "$FILE_ID" ]]; then
  fail "Не удалось получить ID файла из ответа"
  [[ "$CLEANUP_VIDEO" == "true" ]] && rm -f "$VIDEO_FILE"
  exit 1
fi

ok "Файл загружен! ID: $FILE_ID, статус: $FILE_STATUS"

# ─── 7. Проверка файла в списке ───────────────────────────────────────────────
step "Шаг 7: Проверка файла в списке /api/v1/files"

sleep 1
LIST_RESP=$(curl -s "$API_URL/api/v1/files?page=1&page_size=5" \
  -H "Authorization: Bearer $USER_TOKEN")

FOUND=$(echo "$LIST_RESP" | jq --arg id "$FILE_ID" \
  '.files[]? | select(.id == $id)' 2>/dev/null)

if [[ -n "$FOUND" ]]; then
  ok "Файл найден в списке:"
  echo "$FOUND" | jq '{id,original_name,status,file_size,uploaded_at}'
else
  warn "Файл не найден в списке (возможно задержка БД)"
  info "Последние файлы:"
  echo "$LIST_RESP" | jq '.files[0:3]? | .[] | {id,original_name,status}' 2>/dev/null || echo "$LIST_RESP"
fi

# ─── 8. Проверка задачи в Redis ──────────────────────────────────────────────
step "Шаг 8: Проверка задачи в Redis"

if [[ "$HAVE_REDIS_CLI" == "true" ]]; then
  QUEUE_LEN=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" LLEN "$REDIS_QUEUE" 2>/dev/null || echo "?")
  info "Задач в очереди $REDIS_QUEUE: $QUEUE_LEN"

  if [[ "$QUEUE_LEN" != "0" && "$QUEUE_LEN" != "?" ]]; then
    ok "В очереди есть задачи — Python-воркер подхватит их"
    info "Последняя задача в очереди:"
    redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" LINDEX "$REDIS_QUEUE" 0 2>/dev/null | \
      python3 -c "import sys,json; d=json.load(sys.stdin); print(json.dumps(d,indent=2,ensure_ascii=False))" 2>/dev/null || \
      redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" LINDEX "$REDIS_QUEUE" 0
  else
    warn "Очередь пуста"
    info "Возможные причины:"
    echo "  • Redis недоступен с теми настройками, что использует user-portal"
    echo "  • Задача уже была взята воркером (если он запущен)"
    echo "  • MinIO недоступен — user-portal не отправил задачу"
  fi

  RESULT_LEN=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" LLEN "$REDIS_RESULT" 2>/dev/null || echo "?")
  if [[ "$RESULT_LEN" != "0" && "$RESULT_LEN" != "?" ]]; then
    info "Готовых результатов в $REDIS_RESULT: $RESULT_LEN"
    info "Последний результат:"
    redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" LINDEX "$REDIS_RESULT" 0 2>/dev/null | \
      python3 -c "import sys,json; d=json.load(sys.stdin); print(json.dumps(d,indent=2,ensure_ascii=False))" 2>/dev/null || \
      redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" LINDEX "$REDIS_RESULT" 0
  fi
else
  warn "redis-cli не установлен, проверка очереди пропущена"
  info "Проверьте вручную:"
  echo "  redis-cli -h $REDIS_HOST -p $REDIS_PORT LLEN $REDIS_QUEUE"
fi

# ─── 9. Мониторинг статуса транскрибации ─────────────────────────────────────
step "Шаг 9: Мониторинг статуса файла (30 сек)"

info "Ожидаем изменения статуса файла (проверяем каждые 5 секунд)..."
info "Для пропуска нажмите Ctrl+C"

FINAL_STATUS="$FILE_STATUS"
for i in $(seq 1 6); do
  sleep 5
  STATUS_RESP=$(curl -s "$API_URL/api/v1/files?page=1&page_size=20" \
    -H "Authorization: Bearer $USER_TOKEN")
  CURRENT_STATUS=$(echo "$STATUS_RESP" | jq -r --arg id "$FILE_ID" \
    '.files[]? | select(.id == $id) | .status' 2>/dev/null)

  if [[ -z "$CURRENT_STATUS" ]]; then
    warn "Не удалось получить статус (попытка $i/6)"
    continue
  fi

  echo -n "  [${i}0s] Статус: "
  case "$CURRENT_STATUS" in
    pending)    echo -e "${YELLOW}pending${NC} (ожидает)" ;;
    processing) echo -e "${CYAN}processing${NC} (транскрибация идёт...)" ;;
    completed)  echo -e "${GREEN}completed${NC} ✓"; FINAL_STATUS="completed"; break ;;
    failed)     echo -e "${RED}failed${NC} ✗"; FINAL_STATUS="failed"; break ;;
    *)          echo "$CURRENT_STATUS" ;;
  esac
  FINAL_STATUS="$CURRENT_STATUS"
done

echo ""
case "$FINAL_STATUS" in
  completed)
    ok "Транскрибация успешно завершена!"
    ;;
  processing)
    info "Транскрибация ещё идёт — файл большой или воркер занят"
    info "Продолжайте мониторинг:"
    echo "  watch -n5 'curl -s $API_URL/api/v1/files -H \"Authorization: Bearer \$TOKEN\" | jq .'"
    ;;
  pending)
    warn "Файл всё ещё ожидает — Python-воркер, возможно, не запущен"
    ;;
  failed)
    fail "Транскрибация завершилась с ошибкой"
    info "Проверьте логи воркера:"
    echo "  docker-compose logs transcription-worker"
    ;;
esac

# ─── 10. Итоговая сводка ──────────────────────────────────────────────────────
step "Итоговая сводка"

echo ""
echo -e "  Файл ID:         ${CYAN}$FILE_ID${NC}"
echo -e "  Итоговый статус: ${CYAN}$FINAL_STATUS${NC}"
echo ""
echo "  Полезные команды:"
echo ""
echo "  # Статус файла:"
echo "  curl -s '$API_URL/api/v1/files' -H 'Authorization: Bearer $USER_TOKEN' | jq '.files[] | select(.id == \"$FILE_ID\")'"
echo ""
echo "  # Очередь Redis:"
if [[ "$HAVE_REDIS_CLI" == "true" ]]; then
  echo "  redis-cli -h $REDIS_HOST -p $REDIS_PORT LLEN $REDIS_QUEUE"
  echo "  redis-cli -h $REDIS_HOST -p $REDIS_PORT LRANGE $REDIS_RESULT 0 -1"
fi
echo ""
echo "  # Логи воркера транскрибации:"
echo "  docker-compose logs -f transcription-worker"
echo ""
echo "  # Прямой тест Python-воркера (без user-portal):"
echo "  cd cmd/transcription-worker && python test_push_task.py /path/to/video.mp4"

# ─── Очистка ──────────────────────────────────────────────────────────────────
[[ "$CLEANUP_VIDEO" == "true" ]] && rm -f "$VIDEO_FILE" && info "Временный файл удалён"

echo ""

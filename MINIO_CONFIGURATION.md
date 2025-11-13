# Конфигурация MinIO Storage

Этот документ описывает конфигурацию MinIO для всех сервисов Recontext.online.

## Единый bucket для всех сервисов

**Название bucket:** `recontext`

Все сервисы платформы используют один общий bucket `recontext` для хранения:
- Записей LiveKit Egress (аудио/видео конференций)
- Отдельных аудио-треков участников (для транскрибации)
- Загруженных пользователями файлов
- Документов и медиа-файлов

## Переменные окружения

### Базовые настройки MinIO (для всех сервисов)

```bash
# Endpoint MinIO сервера
MINIO_ENDPOINT=minio:9000

# Credentials для доступа
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin

# Название bucket (ОБЯЗАТЕЛЬНО!)
MINIO_BUCKET=recontext

# Использование SSL (опционально)
MINIO_USE_SSL=false
```

### Настройки LiveKit Egress

LiveKit Egress также использует тот же bucket через S3-совместимый API:

```bash
# Включение записи
LIVEKIT_EGRESS_ENABLED=true

# S3-совместимый endpoint MinIO
LIVEKIT_EGRESS_S3_ENDPOINT=https://api.storage.recontext.online

# Bucket для записей (ДОЛЖЕН СОВПАДАТЬ с MINIO_BUCKET!)
LIVEKIT_EGRESS_S3_BUCKET=recontext

# Credentials (должны совпадать с MinIO)
LIVEKIT_EGRESS_S3_ACCESS_KEY=minioadmin
LIVEKIT_EGRESS_S3_SECRET=minioadmin

# Регион (пусто для MinIO)
LIVEKIT_EGRESS_S3_REGION=

# Записывать отдельные треки участников
LIVEKIT_EGRESS_RECORD_TRACKS=true
```

## Структура файлов в bucket

```
recontext/
├── egress/                     # Записи LiveKit
│   ├── room/                   # Композитные записи комнат
│   │   ├── RM_xxx_audio.m4a   # Аудиозаписи комнат
│   │   └── RM_xxx_video.mp4   # Видеозаписи комнат
│   └── track/                  # Отдельные треки участников
│       ├── TR_xxx_audio.m4a   # Аудио трек 1
│       └── TR_yyy_audio.m4a   # Аудио трек 2
├── uploads/                    # Загруженные пользователями файлы
│   ├── documents/
│   ├── images/
│   └── videos/
└── temp/                       # Временные файлы
```

## Конфигурация в docker-compose файлах

### Managing Portal

```yaml
managing-portal:
  environment:
    # Базовые настройки MinIO
    - MINIO_ENDPOINT=minio:9000
    - MINIO_ACCESS_KEY=minioadmin
    - MINIO_SECRET_KEY=minioadmin
    - MINIO_BUCKET=recontext

    # LiveKit Egress использует тот же bucket
    - LIVEKIT_EGRESS_ENABLED=true
    - LIVEKIT_EGRESS_S3_ENDPOINT=https://api.storage.recontext.online
    - LIVEKIT_EGRESS_S3_BUCKET=recontext  # ← Совпадает с MINIO_BUCKET
    - LIVEKIT_EGRESS_S3_ACCESS_KEY=minioadmin
    - LIVEKIT_EGRESS_S3_SECRET=minioadmin
```

### User Portal

```yaml
user-portal:
  environment:
    # Базовые настройки MinIO (для загрузки файлов)
    - MINIO_ENDPOINT=minio:9000
    - MINIO_ACCESS_KEY=minioadmin
    - MINIO_SECRET_KEY=minioadmin
    - MINIO_BUCKET=recontext
```

## Обновленные файлы

Все docker-compose файлы обновлены с единым bucket:

✅ `docker-compose.yml` - Development
✅ `docker-compose.prod.yml` - Production
✅ `docker-compose-both-portals.yml`
✅ `docker-compose-both-portals-mac.yml`
✅ `docker-compose-managing-portal.yml`
✅ `docker-compose-managing-portal-mac.yml`
✅ `docker-compose-user-portal.yml`
✅ `docker-compose-user-portal-mac.yml`
✅ `docker-compose-mac.yml`

## Создание bucket при первом запуске

MinIO автоматически создаст bucket при первой попытке записи, но рекомендуется создать его вручную:

### Через MinIO Console

1. Откройте MinIO Console: http://localhost:9001
2. Войдите с credentials: `minioadmin` / `minioadmin`
3. Перейдите в раздел "Buckets"
4. Нажмите "Create Bucket"
5. Введите имя: `recontext`
6. Нажмите "Create"

### Через MinIO Client (mc)

```bash
# Настроить alias для MinIO
mc alias set myminio http://localhost:9000 minioadmin minioadmin

# Создать bucket
mc mb myminio/recontext

# Проверить
mc ls myminio/
```

### Через API (curl)

```bash
# Создать bucket
curl -X PUT http://localhost:9000/recontext \
  --user minioadmin:minioadmin

# Проверить список buckets
curl http://localhost:9000 \
  --user minioadmin:minioadmin
```

## Важные замечания

### ⚠️ Единый bucket для всех операций

Все сервисы **ДОЛЖНЫ** использовать один и тот же bucket `recontext`:
- Managing Portal для записей и файлов
- User Portal для загрузки файлов
- LiveKit Egress для записей конференций

### ⚠️ Credentials должны совпадать

```bash
MINIO_ACCESS_KEY = LIVEKIT_EGRESS_S3_ACCESS_KEY
MINIO_SECRET_KEY = LIVEKIT_EGRESS_S3_SECRET
MINIO_BUCKET = LIVEKIT_EGRESS_S3_BUCKET
```

### ⚠️ Endpoint для Egress

LiveKit Egress использует публичный endpoint:
- Internal: `minio:9000` (для сервисов внутри Docker)
- External: `https://api.storage.recontext.online` (для LiveKit Egress)

## Проверка конфигурации

После запуска проверьте логи:

```bash
# Проверить логи managing-portal
docker logs recontext-managing-portal | grep -i minio

# Проверить что bucket используется
docker logs recontext-managing-portal | grep -i bucket

# Проверить egress записи
docker logs recontext-managing-portal | grep -i egress
```

## Troubleshooting

### Bucket не найден

**Ошибка:** `The specified bucket does not exist`

**Решение:** Создайте bucket `recontext` вручную (см. выше)

### Записи не сохраняются

**Проблема:** LiveKit Egress не может записать файлы

**Проверка:**
1. Bucket `recontext` существует
2. `LIVEKIT_EGRESS_S3_BUCKET=recontext` в конфигурации
3. Credentials правильные
4. MinIO доступен по endpoint

### Разные buckets в разных сервисах

**Проблема:** Файлы попадают в разные места

**Решение:** Убедитесь что везде используется `MINIO_BUCKET=recontext`:
```bash
grep "MINIO.*BUCKET" docker-compose.prod.yml
# Все должны быть = recontext
```

## Миграция со старой конфигурации

Если ранее использовался bucket `jitsi-recordings`, нужно:

1. Скопировать данные в новый bucket:
```bash
mc cp --recursive myminio/jitsi-recordings/ myminio/recontext/egress/
```

2. Обновить docker-compose файлы (уже сделано)

3. Перезапустить сервисы:
```bash
docker-compose down
docker-compose up -d
```

## Мониторинг использования

MinIO предоставляет метрики через Prometheus:
- Endpoint: `http://minio:9000/minio/v2/metrics/cluster`
- Размер bucket
- Количество объектов
- Трафик I/O

Пример запроса:
```bash
curl -s http://localhost:9000/minio/v2/metrics/cluster | grep minio_bucket
```

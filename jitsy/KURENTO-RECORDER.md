# 🎥 Jitsi + Kurento Individual Stream Recorder

Полная система для записи индивидуальных аудио/видео потоков участников Jitsi конференций с автоматической загрузкой в MinIO storage.

## 📋 Содержание

- [Обзор](#обзор)
- [Архитектура](#архитектура)
- [Компоненты](#компоненты)
- [Установка](#установка)
- [Использование](#использование)
- [API Документация](#api-документация)
- [Мониторинг](#мониторинг)
- [Troubleshooting](#troubleshooting)

## 🎯 Обзор

Эта система позволяет:

✅ **Записывать индивидуальные потоки** каждого участника Jitsi конференции в отдельный файл
✅ **Управлять записью через REST API** без нагрузки на клиентов
✅ **Автоматически загружать** записи в MinIO (S3-compatible storage)
✅ **Масштабировать** запись множества участников одновременно
✅ **Мониторить** активные записи в реальном времени

## 🏗️ Архитектура

```
┌──────────────────────────────────────────────────────────────┐
│                   Jitsi Meet Infrastructure                  │
│  ┌──────────┐  ┌──────────┐  ┌───────┐  ┌─────────────┐    │
│  │   Web    │  │ Prosody  │  │Jicofo │  │     JVB     │    │
│  │(Nginx)   │  │  (XMPP)  │  │(Focus)│  │(Videobridge)│    │
│  └──────────┘  └──────────┘  └───────┘  └──────┬──────┘    │
└────────────────────────────────────────────────│─────────────┘
                                                  │
                                    RTP/WebRTC forwarding
                                                  │
┌─────────────────────────────────────────────────▼─────────────┐
│                  Recording Infrastructure                     │
│  ┌────────────────────┐         ┌──────────────────────┐     │
│  │ Kurento Media      │         │ Kurento Recorder     │     │
│  │ Server             │◄────────┤ Service (FastAPI)    │     │
│  │ (WebRTC + Record)  │ WS API  │ REST API             │     │
│  └────────┬───────────┘         └──────────────────────┘     │
│           │                                                   │
│           │ Save files                                        │
│           ▼                                                   │
│  ┌────────────────────┐                                      │
│  │ /recordings        │                                      │
│  │ (temporary storage)│                                      │
│  └────────┬───────────┘                                      │
└───────────│──────────────────────────────────────────────────┘
            │
            │ Upload via S3 API
            ▼
┌───────────────────────────────────────────────────────────────┐
│                    MinIO Storage                              │
│         https://api.storage.recontext.online                  │
│                                                               │
│  Bucket: jitsi-recordings/                                   │
│  ├── room1/                                                  │
│  │   ├── user1_20251108_153000.webm                         │
│  │   ├── user2_20251108_153005.webm                         │
│  │   └── ...                                                 │
│  └── room2/                                                  │
│      └── ...                                                  │
└───────────────────────────────────────────────────────────────┘
```

## 📦 Компоненты

### 1. Jitsi Meet Stack
- **jitsi/web** - Web интерфейс (Nginx + Jitsi Meet frontend)
- **jitsi/prosody** - XMPP сервер для сигнализации
- **jitsi/jicofo** - Jitsi Conference Focus (управление конференциями)
- **jitsi/jvb** - Jitsi Videobridge (маршрутизация медиа потоков)

### 2. Recording Stack
- **kurento/kurento-media-server** - WebRTC media server для записи
- **kurento-recorder** - FastAPI сервис с REST API для управления записью
- **minio/minio** - S3-compatible object storage

### 3. Вспомогательные сервисы
- **redis** - Координация между компонентами
- **jitsi-recorder-lite** - XMPP мониторинг участников (wrtc)

## 🚀 Установка

### Требования

- Docker 20.10+
- Docker Compose 2.0+
- 8 GB RAM minimum
- 50 GB disk space

### 1. Клонировать репозиторий

```bash
cd /Volumes/ExternalData/source/Team21/Recontext.online/jitsy
```

### 2. Настроить переменные окружения

Скопировать `.env.example` в `.env` и настроить:

```bash
cp .env.example .env
nano .env
```

Основные переменные:
```env
# Jitsi configuration
PUBLIC_URL=https://meet.recontext.online
ENABLE_RECORDING=1

# MinIO configuration (внешний сервис)
MINIO_ENDPOINT=https://api.storage.recontext.online
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=jitsi-recordings
```

### 3. Запустить все сервисы

```bash
docker-compose up -d
```

### 4. Проверить статус

```bash
docker-compose ps
```

Все сервисы должны быть в статусе `Up (healthy)`.

### 5. Проверить API

```bash
curl http://localhost:9888/health
```

Ответ:
```json
{"status": "healthy", "active_recordings": 0}
```

## 📡 Использование

### Сценарий 1: Запись одного участника

```bash
# 1. Пользователь присоединяется к конференции
# https://meet.recontext.online/testroom

# 2. Начать запись участника john_doe
curl -X POST "http://localhost:9888/record/start?user=john_doe&room=testroom"

# 3. Проверить статус
curl "http://localhost:9888/record/status?user=john_doe"

# 4. После окончания разговора - остановить запись
curl -X POST "http://localhost:9888/record/stop?user=john_doe&room=testroom"

# 5. Файл автоматически загружен в MinIO:
# https://api.storage.recontext.online/jitsi-recordings/testroom/john_doe_*.webm
```

### Сценарий 2: Запись нескольких участников

```bash
# Начать запись для трех участников
curl -X POST "http://localhost:9888/record/start?user=alice&room=meeting123"
curl -X POST "http://localhost:9888/record/start?user=bob&room=meeting123"
curl -X POST "http://localhost:9888/record/start?user=charlie&room=meeting123"

# Проверить все активные записи
curl "http://localhost:9888/record/list"

# Остановить все записи
curl -X DELETE "http://localhost:9888/record/stop-all"
```

### Сценарий 3: Автоматическая запись через webhook

Настроить webhook в Jitsi для автоматического вызова recorder API при join/leave событиях.

## 📚 API Документация

### Base URL
```
http://localhost:9888
```

### Swagger UI (Интерактивная документация)
```
http://localhost:9888/docs
```

### Endpoints

#### `POST /record/start`
Начать запись участника

**Query Parameters:**
- `user` (string, required) - ID участника
- `room` (string, optional) - Название комнаты

**Response:**
```json
{
  "status": "started",
  "user": "john_doe",
  "filepath": "/recordings/john_doe_20251108_153000.webm",
  "started_at": "2025-11-08T15:30:00"
}
```

#### `POST /record/stop`
Остановить запись и загрузить в MinIO

**Query Parameters:**
- `user` (string, required) - ID участника
- `room` (string, optional) - Название комнаты

**Response:**
```json
{
  "status": "stopping",
  "user": "john_doe",
  "stopped_at": "2025-11-08T15:35:00"
}
```

#### `GET /record/status?user=<user>`
Проверить статус записи

#### `GET /record/list`
Список всех активных записей

#### `DELETE /record/stop-all`
Остановить все активные записи

## 📊 Мониторинг

### Логи сервисов

```bash
# Все логи
docker-compose logs -f

# Только Kurento
docker-compose logs -f kurento

# Только Recorder API
docker-compose logs -f kurento-recorder

# JVB (videobridge)
docker-compose logs -f jvb
```

### Метрики

#### Kurento Media Server
```bash
curl http://localhost:8888/kurento/stats
```

#### Recorder Service
```bash
curl http://localhost:9888/record/list
```

#### MinIO
```
https://api.storage.recontext.online
Login: minioadmin / minioadmin
```

### Health Checks

```bash
# Recorder service
curl http://localhost:9888/health

# Kurento (WebSocket) - external port 8889
wscat -c ws://localhost:8889/kurento

# MinIO
curl https://api.storage.recontext.online/minio/health/live
```

**Примечание**: Kurento доступен извне на порту 8889 (порт 8888 занят jicofo).

## 🔧 Troubleshooting

### Проблема: Kurento не может записать

**Симптом:**
```
ERROR: Failed to create recording: Connection refused
```

**Решение:**
```bash
# Проверить что Kurento запущен
docker-compose ps kurento

# Проверить логи
docker-compose logs kurento

# Перезапустить
docker-compose restart kurento
```

### Проблема: Файл не загружается в MinIO

**Симптом:**
```
ERROR: MinIO upload failed: Connection timeout
```

**Решение:**
```bash
# Проверить доступность MinIO
curl https://api.storage.recontext.online

# Проверить credentials
docker-compose exec kurento-recorder env | grep MINIO

# Проверить bucket
docker-compose exec kurento-recorder python3 -c "
import boto3
s3 = boto3.client('s3',
    endpoint_url='https://api.storage.recontext.online',
    aws_access_key_id='minioadmin',
    aws_secret_access_key='minioadmin')
print(s3.list_buckets())
"
```

### Проблема: Пустые файлы записи

**Симптом:**
Файл создается, но размер 0 байт

**Решение:**
- Проверить что RTP порты 40000-40100 открыты
- Проверить что JVB правильно форвардит потоки в Kurento
- Проверить настройки firewall

```bash
# Проверить открытые порты
docker-compose exec kurento netstat -tulpn | grep 40000
```

### Проблема: Out of Memory

**Симптом:**
```
ERROR: Container killed (OOM)
```

**Решение:**
Увеличить memory limits в docker-compose.yml:

```yaml
kurento:
  mem_limit: 4g  # увеличить с 2g

kurento-recorder:
  mem_limit: 1g  # увеличить с 512m
```

## 🔐 Безопасность

### MinIO Access

- Используйте strong passwords в production
- Настройте IAM policies для bucket access
- Используйте HTTPS для всех соединений

### API Security

Добавить аутентификацию в recorder.py:

```python
from fastapi.security import HTTPBearer
security = HTTPBearer()

@app.post("/record/start")
async def start_recording(
    user: str,
    credentials: HTTPAuthorizationCredentials = Security(security)
):
    # Validate token
    if credentials.credentials != "your-secret-token":
        raise HTTPException(401, "Invalid token")
    # ...
```

### Network Security

```bash
# Ограничить доступ к API только с локальной сети
iptables -A INPUT -p tcp --dport 8080 -s 192.168.0.0/24 -j ACCEPT
iptables -A INPUT -p tcp --dport 8080 -j DROP
```

## 📈 Масштабирование

### Horizontal Scaling

Запустить несколько экземпляров Kurento:

```yaml
# docker-compose.yml
kurento-1:
  image: kurento/kurento-media-server:latest
  ports:
    - "8888:8888"
    - "40000-40100:40000-40100/udp"

kurento-2:
  image: kurento/kurento-media-server:latest
  ports:
    - "8889:8888"
    - "40100-40200:40100-40200/udp"
```

### Load Balancing

Использовать Nginx для балансировки между recorder instances:

```nginx
upstream recorder_backend {
    server kurento-recorder-1:8080;
    server kurento-recorder-2:8080;
}

server {
    location /record/ {
        proxy_pass http://recorder_backend;
    }
}
```

## 🧪 Тестирование

### Функциональный тест

```bash
./test-recorder.sh
```

Скрипт `test-recorder.sh`:
```bash
#!/bin/bash
set -e

echo "🧪 Testing Kurento Recorder..."

# 1. Health check
echo "1. Health check"
curl -f http://localhost:9888/health || exit 1

# 2. Start recording
echo "2. Start recording"
curl -X POST "http://localhost:9888/record/start?user=test&room=test" || exit 1

# 3. Check status
echo "3. Check status"
sleep 2
curl "http://localhost:9888/record/status?user=test" || exit 1

# 4. Wait
echo "4. Recording for 30 seconds..."
sleep 30

# 5. Stop recording
echo "5. Stop recording"
curl -X POST "http://localhost:9888/record/stop?user=test&room=test" || exit 1

# 6. Wait for upload
echo "6. Waiting for MinIO upload..."
sleep 5

echo "✅ Test completed successfully!"
```

## 📝 Roadmap

- [ ] Автоматическая транскрипция через Whisper
- [ ] Webhook notifications при завершении записи
- [ ] Real-time streaming в RTMP/HLS
- [ ] Автоматическое разделение по спикерам (diarization)
- [ ] Интеграция с Jitsi events для auto-start/stop
- [ ] Dashboard для визуализации активных записей
- [ ] Metrics export в Prometheus
- [ ] Аутентификация через JWT tokens

## 👥 Авторы

**Святослав Иванов**
AI-инженер и архитектор решений
Специализация: Jitsi / Kurento / WebRTC / FastAPI
2025

## 📄 Лицензия

MIT License

## 🔗 Ссылки

- [Jitsi Meet](https://jitsi.org/)
- [Kurento Documentation](https://doc-kurento.readthedocs.io/)
- [FastAPI](https://fastapi.tiangolo.com/)
- [MinIO](https://min.io/)
- [Docker Compose](https://docs.docker.com/compose/)

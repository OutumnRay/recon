# 🎥 Jitsi Kurento Recorder Service

REST API сервис для записи индивидуальных аудио/видео потоков участников Jitsi конференций через Kurento Media Server.

## 🎯 Возможности

- ✅ REST API для управления записью
- ✅ Запись индивидуальных потоков участников
- ✅ Автоматическая загрузка в MinIO (S3-compatible storage)
- ✅ WebM формат (VP8/VP9 + Opus)
- ✅ Фоновая обработка и загрузка
- ✅ Health checks и мониторинг
- ✅ Список активных записей

## 🏗️ Архитектура

```
┌──────────────┐
│ Jitsi Client │
└──────┬───────┘
       │ WebRTC
┌──────▼───────────┐
│ Jitsi Videobridge│
└──────┬───────────┘
       │ RTP forwarding
┌──────▼───────────┐
│ Kurento Media    │
│ Server           │
└──────┬───────────┘
       │ WebSocket API
┌──────▼───────────┐
│ Recorder Service │
│ (FastAPI)        │
└──────┬───────────┘
       │ S3 API
┌──────▼───────────┐
│ MinIO Storage    │
│ api.storage.recontext.online │
└──────────────────┘
```

## 🚀 Запуск

### Через Docker Compose (рекомендуется)

```bash
cd /path/to/jitsy
docker-compose up -d kurento kurento-recorder
```

### Просмотр логов

```bash
docker-compose logs -f kurento-recorder
```

## 📡 REST API

### Base URL
```
http://localhost:9888
```

### Endpoints

#### 1. Начать запись участника

```bash
POST /record/start?user=<user_id>&room=<room_name>
```

**Параметры:**
- `user` (required): Идентификатор участника
- `room` (optional): Название комнаты

**Пример:**
```bash
curl -X POST "http://localhost:9888/record/start?user=john_doe&room=testmeeting"
```

**Ответ:**
```json
{
  "status": "started",
  "user": "john_doe",
  "message": "Recording started for user john_doe",
  "filepath": "/recordings/john_doe_20251108_153000.webm",
  "started_at": "2025-11-08T15:30:00.000Z"
}
```

---

#### 2. Остановить запись участника

```bash
POST /record/stop?user=<user_id>&room=<room_name>
```

**Параметры:**
- `user` (required): Идентификатор участника
- `room` (optional): Название комнаты

**Пример:**
```bash
curl -X POST "http://localhost:9888/record/stop?user=john_doe&room=testmeeting"
```

**Ответ:**
```json
{
  "status": "stopping",
  "user": "john_doe",
  "message": "Recording stopping for user john_doe",
  "filepath": "/recordings/john_doe_20251108_153000.webm",
  "started_at": "2025-11-08T15:30:00.000Z",
  "stopped_at": "2025-11-08T15:35:00.000Z"
}
```

**Примечание:** Файл автоматически загружается в MinIO в фоновом режиме.

---

#### 3. Проверить статус записи

```bash
GET /record/status?user=<user_id>
```

**Пример:**
```bash
curl "http://localhost:9888/record/status?user=john_doe"
```

**Ответ:**
```json
{
  "status": "recording",
  "user": "john_doe",
  "filepath": "/recordings/john_doe_20251108_153000.webm",
  "started_at": "2025-11-08T15:30:00.000Z",
  "duration": 125.5
}
```

---

#### 4. Список всех активных записей

```bash
GET /record/list
```

**Пример:**
```bash
curl "http://localhost:9888/record/list"
```

**Ответ:**
```json
{
  "active_recordings": 2,
  "recordings": [
    {
      "user": "john_doe",
      "filepath": "/recordings/john_doe_20251108_153000.webm",
      "started_at": "2025-11-08T15:30:00.000Z",
      "duration": 125.5
    },
    {
      "user": "jane_smith",
      "filepath": "/recordings/jane_smith_20251108_153015.webm",
      "started_at": "2025-11-08T15:30:15.000Z",
      "duration": 110.2
    }
  ]
}
```

---

#### 5. Остановить все записи

```bash
DELETE /record/stop-all
```

**Пример:**
```bash
curl -X DELETE "http://localhost:9888/record/stop-all"
```

**Ответ:**
```json
{
  "status": "stopping_all",
  "count": 2,
  "users": ["john_doe", "jane_smith"]
}
```

---

#### 6. Health Check

```bash
GET /health
```

**Пример:**
```bash
curl "http://localhost:9888/health"
```

**Ответ:**
```json
{
  "status": "healthy",
  "active_recordings": 2
}
```

## 🗂️ MinIO Storage

### Конфигурация

Записи автоматически загружаются в MinIO по адресу:
- **Endpoint:** `https://api.storage.recontext.online`
- **Access Key:** `minioadmin`
- **Secret Key:** `minioadmin`
- **Bucket:** `jitsi-recordings`

### Структура объектов в MinIO

```
jitsi-recordings/
├── testmeeting/
│   ├── john_doe_20251108_153000.webm
│   ├── jane_smith_20251108_153015.webm
│   └── ...
└── anotherroom/
    ├── user1_20251108_160000.webm
    └── ...
```

### Доступ к записям через MinIO Console

1. Откройте MinIO Console: https://api.storage.recontext.online
2. Войдите с credentials: minioadmin/minioadmin
3. Перейдите в bucket `jitsi-recordings`
4. Скачайте нужные записи

## ⚙️ Переменные окружения

| Переменная | Значение по умолчанию | Описание |
|------------|----------------------|----------|
| `KURENTO_URI` | `ws://kurento:8888/kurento` | WebSocket URI Kurento Media Server |
| `MINIO_ENDPOINT` | `https://api.storage.recontext.online` | MinIO endpoint URL |
| `MINIO_ACCESS_KEY` | `minioadmin` | MinIO access key |
| `MINIO_SECRET_KEY` | `minioadmin` | MinIO secret key |
| `MINIO_BUCKET` | `jitsi-recordings` | MinIO bucket name |
| `MINIO_SECURE` | `true` | Use HTTPS for MinIO |

## 🔧 Troubleshooting

### Проверка соединения с Kurento

```bash
docker-compose exec kurento-recorder curl ws://kurento:8888/kurento
```

### Проверка MinIO

```bash
docker-compose exec kurento-recorder python3 -c "
import boto3
s3 = boto3.client('s3',
    endpoint_url='https://api.storage.recontext.online',
    aws_access_key_id='minioadmin',
    aws_secret_access_key='minioadmin')
print(s3.list_buckets())
"
```

### Логи Kurento Media Server

```bash
docker-compose logs -f kurento
```

### Логи Recorder Service

```bash
docker-compose logs -f kurento-recorder
```

## 📊 Мониторинг

### Prometheus метрики

FastAPI автоматически экспортирует метрики:
- Количество активных записей
- Время запросов
- Статус health checks

### Health Check endpoint для Kubernetes/Docker

```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:9888/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 20s
```

## 🧪 Тестирование

### Тест полного цикла

```bash
# 1. Начать запись
curl -X POST "http://localhost:9888/record/start?user=test_user&room=test_room"

# 2. Проверить статус
curl "http://localhost:9888/record/status?user=test_user"

# 3. Подождать 30 секунд
sleep 30

# 4. Остановить запись
curl -X POST "http://localhost:9888/record/stop?user=test_user&room=test_room"

# 5. Проверить MinIO
# Файл должен появиться в bucket jitsi-recordings/test_room/test_user_*.webm
```

## 🚨 Ошибки

### 400 Bad Request
```json
{"detail": "Recording already active for user john_doe"}
```
**Решение:** Остановите текущую запись перед началом новой

### 404 Not Found
```json
{"detail": "No active recording found for user john_doe"}
```
**Решение:** Убедитесь что запись была начата для этого пользователя

### 500 Internal Server Error
```json
{"detail": "Failed to create recording: Connection refused"}
```
**Решение:** Проверьте что Kurento Media Server запущен и доступен

## 📝 Примеры интеграции

### Python

```python
import requests

# Начать запись
response = requests.post(
    "http://localhost:9888/record/start",
    params={"user": "john_doe", "room": "meeting123"}
)
print(response.json())

# Остановить через 60 секунд
import time
time.sleep(60)

response = requests.post(
    "http://localhost:9888/record/stop",
    params={"user": "john_doe", "room": "meeting123"}
)
print(response.json())
```

### JavaScript (Node.js)

```javascript
const axios = require('axios');

// Начать запись
async function startRecording(user, room) {
    const response = await axios.post(
        `http://localhost:9888/record/start?user=${user}&room=${room}`
    );
    console.log(response.data);
}

// Остановить запись
async function stopRecording(user, room) {
    const response = await axios.post(
        `http://localhost:9888/record/stop?user=${user}&room=${room}`
    );
    console.log(response.data);
}

startRecording('john_doe', 'meeting123');
```

### Curl

```bash
#!/bin/bash

USER="john_doe"
ROOM="meeting123"
API_URL="http://localhost:9888"

# Start recording
echo "Starting recording for $USER in room $ROOM"
curl -X POST "$API_URL/record/start?user=$USER&room=$ROOM"

# Wait
echo "Recording for 60 seconds..."
sleep 60

# Stop recording
echo "Stopping recording for $USER"
curl -X POST "$API_URL/record/stop?user=$USER&room=$ROOM"

echo "Done! Check MinIO for the recording."
```

## 📚 Дополнительные ресурсы

- [Kurento Documentation](https://doc-kurento.readthedocs.io/)
- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [MinIO Documentation](https://min.io/docs/)
- [Jitsi Developer Guide](https://jitsi.github.io/handbook/)

## 🏆 Автор

**Святослав Иванов**
AI-инженер и архитектор решений
2025

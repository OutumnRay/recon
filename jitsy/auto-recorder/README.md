# 🤖 Jitsi Auto Recorder

Автоматический сервис записи Jitsi конференций через Kurento Media Server.

## 🎯 Функции

- ✅ Автоматическая запись при присоединении участника
- ✅ Автоматическая остановка при выходе участника
- ✅ Интеграция с Kurento Recorder API
- ✅ Webhook от Prosody для событий конференции
- ✅ Автоматическая загрузка на MinIO
- ✅ Health checks и мониторинг

## 🏗️ Архитектура

```
Prosody (XMPP)
    │
    │ webhook events
    ↓
Auto Recorder (Python)
    │
    │ REST API calls
    ↓
Kurento Recorder API
    │
    │ WebSocket JSON-RPC
    ↓
Kurento Media Server
    │
    │ recording files
    ↓
MinIO (S3)
```

## 📡 Webhook События

Auto Recorder слушает следующие события от Prosody:

### 1. participantJoined
```json
{
  "eventType": "participantJoined",
  "roomName": "testmeeting",
  "endpointId": "abc123",
  "participantId": "user@domain/resource",
  "participantName": "user@domain/resource",
  "displayName": "John Doe"
}
```

**Действие**: Вызывает `POST /record/start?user={participantId}&room={roomName}` на Kurento API

---

### 2. participantLeft
```json
{
  "eventType": "participantLeft",
  "roomName": "testmeeting",
  "endpointId": "abc123"
}
```

**Действие**: Вызывает `POST /record/stop?user={participantId}&room={roomName}` на Kurento API

---

### 3. conferenceEnded
```json
{
  "eventType": "conferenceEnded",
  "roomName": "testmeeting"
}
```

**Действие**: Останавливает все активные записи для комнаты

---

## 🚀 Запуск

### Docker Compose (автоматически)

Сервис запускается автоматически вместе с остальным стеком:

```bash
cd /path/to/jitsy
docker-compose up -d auto-recorder
```

### Проверка статуса

```bash
# Проверить здоровье
curl http://localhost:9889/health

# Получить статистику
curl http://localhost:9889/stats
```

## ⚙️ Переменные окружения

| Переменная | Значение по умолчанию | Описание |
|------------|----------------------|----------|
| `KURENTO_API_URL` | `http://kurento-recorder:8080` | URL Kurento Recorder API |
| `REDIS_HOST` | `redis` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `AUTO_RECORD` | `true` | Включить автозапись |

## 📊 API Endpoints

### GET /health

Проверка здоровья сервиса

**Response:**
```json
{
  "status": "healthy",
  "worker_id": "auto-recorder",
  "auto_record": true,
  "active_conferences": 1,
  "active_sessions": 2,
  "kurento_api": "http://kurento-recorder:8080",
  "kurento_status": {
    "status": "healthy",
    "active_recordings": 2
  }
}
```

---

### GET /stats

Статистика активных записей

**Response:**
```json
{
  "active_conferences": [
    {
      "room": "testmeeting",
      "started_at": "2025-11-08T14:30:00",
      "participants": 2
    }
  ],
  "active_sessions": [
    {
      "endpoint_id": "abc123",
      "user": "john@domain/resource",
      "room": "testmeeting",
      "display_name": "John Doe",
      "started_at": "2025-11-08T14:30:05"
    }
  ]
}
```

---

### POST /events

Webhook endpoint для событий от Prosody

См. раздел "Webhook События" выше.

---

## 🔧 Настройка Prosody

В `.env` файле должна быть установлена переменная:

```bash
RECORDER_WEBHOOK_URL=http://auto-recorder:8080/events
```

Prosody автоматически отправляет события на этот URL.

## 📁 Структура записей на MinIO

```
jitsi-recordings/
├── testmeeting/
│   ├── abc123def456_20251108_143000/
│   │   ├── john_doe_20251108_143005.webm
│   │   ├── jane_smith_20251108_143010.webm
│   │   └── metadata.json
│   └── ...
└── anotherroom/
    └── ...
```

## 🧪 Тестирование

### 1. Проверка health check

```bash
curl http://localhost:9889/health
```

Ожидаемый ответ:
```json
{
  "status": "healthy",
  "kurento_status": {
    "status": "healthy",
    "active_recordings": 0
  }
}
```

### 2. Симуляция webhook события

```bash
# Присоединение участника
curl -X POST http://localhost:9889/events \
  -H "Content-Type: application/json" \
  -d '{
    "eventType": "participantJoined",
    "roomName": "test",
    "endpointId": "test123",
    "participantId": "testuser@domain/res",
    "participantName": "testuser@domain/res",
    "displayName": "Test User"
  }'

# Подождать несколько секунд

# Выход участника
curl -X POST http://localhost:9889/events \
  -H "Content-Type: application/json" \
  -d '{
    "eventType": "participantLeft",
    "roomName": "test",
    "endpointId": "test123"
  }'
```

### 3. Проверка записи на MinIO

После теста файл должен появиться в MinIO:
- Bucket: `jitsi-recordings`
- Path: `recordings/test/{conference_id}/testuser_domain_res_{timestamp}.webm`

## 📜 Логи

### Просмотр логов

```bash
# Следить за логами в реальном времени
docker-compose logs -f auto-recorder

# Последние 100 строк
docker-compose logs --tail=100 auto-recorder
```

### Типичные логи

**Присоединение участника:**
```
📥 PARTICIPANT JOINED: John Doe (ID: john@domain/res, endpoint: abc123) in room testmeeting
🎙️  Starting recording: john@domain/res in room testmeeting
✅ Recording started: john@domain/res -> /recordings/testmeeting_john_domain_res_20251108_143000.webm
✅ Auto-recording started for John Doe
```

**Выход участника:**
```
📤 PARTICIPANT LEFT: endpoint abc123 from room testmeeting
⏹️  Stopping recording: john@domain/res in room testmeeting
✅ Recording stopped: john@domain/res -> /recordings/testmeeting_john_domain_res_20251108_143000.webm
✅ Recording stopped for John Doe (duration: 125.5s)
```

**Завершение конференции:**
```
🏁 CONFERENCE ENDED: testmeeting
✅ Conference testmeeting completed:
   Duration: 300.5s
   Participants: 3
📁 Recordings for room 'testmeeting' uploaded to MinIO bucket 'jitsi-recordings'
```

## 🚨 Troubleshooting

### Auto Recorder не стартует

```bash
# Проверить логи
docker-compose logs auto-recorder

# Проверить зависимости
docker-compose ps kurento-recorder redis
```

### Записи не создаются

1. Проверить что Kurento API доступен:
```bash
curl http://localhost:9888/health
```

2. Проверить webhook URL в Prosody:
```bash
grep RECORDER_WEBHOOK_URL .env
# Должно быть: RECORDER_WEBHOOK_URL=http://auto-recorder:8080/events
```

3. Проверить переменную AUTO_RECORD:
```bash
docker-compose exec auto-recorder env | grep AUTO_RECORD
# Должно быть: AUTO_RECORD=true
```

### Записи не загружаются на MinIO

Проверить Kurento Recorder API - он отвечает за загрузку:
```bash
docker-compose logs kurento-recorder | grep -i minio
```

## 🔄 Обновление

```bash
# Пересобрать и перезапустить
cd /path/to/jitsy
docker-compose build auto-recorder
docker-compose up -d auto-recorder

# Проверить что работает
curl http://localhost:9889/health
```

## 📊 Мониторинг

### Prometheus метрики

Auto Recorder экспортирует метрики через `/health` endpoint:
- `active_conferences` - количество активных конференций
- `active_sessions` - количество активных сессий записи

### Health Check для Kubernetes

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 20
  periodSeconds: 30

readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10
```

## 🏆 Автор

**Святослав Иванов**
AI-инженер и архитектор решений
2025

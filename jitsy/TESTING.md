# Тестирование Jitsi Recorder

## Быстрый старт

### 1. Запустить все сервисы

```bash
cd jitsy
docker-compose up -d
```

### 2. Проверить логи рекордера

```bash
docker-compose logs -f recorder
```

Должны увидеть:
```
✅ S3 client initialized - endpoint: https://api.storage.recontext.online, bucket: jitsi-recordings
✅ S3 bucket 'jitsi-recordings' exists and is accessible
✅ Redis connected: redis:6379
🌐 HTTP server started on :8080 (/health, /events)
```

### 3. Проверить MinIO (если локальный)

Если используется локальный MinIO:
```bash
docker-compose ps minio
curl http://localhost:9000/minio/health/live
```

Открыть консоль MinIO: http://localhost:9001
- Логин: `minioadmin`
- Пароль: `minioadmin`

### 4. Проверить Health Check

```bash
curl http://localhost:8080/health
```

Ответ:
```json
{
  "status": "healthy",
  "workerId": "12345678",
  "activeConferences": 0,
  "activeSessions": 0,
  "totalParticipants": 0,
  "s3Configured": true,
  "webhookConfigured": false
}
```

## Тестирование записи

### 1. Открыть Jitsi Meet

http://localhost:8000/testmeet (или https://localhost:8443/testmeet)

### 2. Подключиться к конференции

Зайти с 1-2 участниками (можно в разных вкладках браузера)

### 3. Смотреть логи рекордера

```bash
docker-compose logs -f recorder
```

Должны увидеть:
```
📥 PARTICIPANT JOINED: testmeet@muc.meet.jitsi/a719784e (ID: a719784e-..., endpoint: a719784e)
🎙️  START RECORDING: testmeet_a719784e-5c6e-4988-b6ca-349a419a4e3e_20251108_090100.opus
```

### 4. Отключиться от конференции

Должны увидеть:
```
📤 PARTICIPANT LEFT: endpoint a719784e from room testmeet
⏹️  STOP RECORDING: testmeet_a719784e-..._20251108_090100.opus (duration: 30.5s)
📁 Recording file ready: /tmp/recordings/testmeet_a719784e-..._20251108_090100.opus (245632 bytes)
☁️  Starting upload to s3://jitsi-recordings/recordings/testmeet/{conference_id}/testmeet_a719784e-..._20251108_090100.opus
✅ Upload completed and local file removed
```

### 5. Завершить конференцию

Когда все участники выйдут:
```
🏁 CONFERENCE ENDED: testmeet
✅ Conference testmeet completed:
   Conference ID: cb81d3fd_20251108_090049
   Duration: 30.6s
   Participants: 1
   Total sessions: 1
   Recordings uploaded to S3: 1/1
   S3 Path: s3://jitsi-recordings/recordings/testmeet/cb81d3fd_20251108_090049/
   📁 Recordings:
      ✅ testmeet_a719784e-5c6e-4988-b6ca-349a419a4e3e_20251108_090100.opus (30.5s)
```

## Тестирование переподключений

### Сценарий 1: Нормальное переподключение

1. Зайти в конференцию
2. Остаться в конференции 1 минуту
3. Отключиться от интернета / закрыть вкладку
4. Через 10 секунд снова зайти в ту же конференцию

**Ожидаемый результат:**
- Создано 2 файла для одного участника
- Второй файл помечен `isReconnection: true`
- Оба файла загружены на S3

### Сценарий 2: Долгое отключение

1. Зайти в конференцию
2. Остаться 1 минуту
3. Отключиться
4. Через 2 минуты снова зайти

**Ожидаемый результат:**
- Создано 2 файла
- Второй файл помечен `isReconnection: false` (прошло >30 сек)

### Сценарий 3: Несколько участников

1. Зайти с 3 участниками
2. Каждый остается разное время
3. Один переподключается

**Ожидаемый результат:**
- По 1 файлу для каждого участника
- +1 файл для переподключившегося
- Все файлы в одной папке `recordings/testmeet/{conference_id}/`

## Проверка на S3/MinIO

### Локальный MinIO

1. Открыть http://localhost:9001
2. Зайти: minioadmin / minioadmin
3. Перейти в Buckets → jitsi-recordings
4. Найти папку `recordings/testmeet/{conference_id}/`
5. Проверить файлы:
   - Файлы `.opus`
   - `metadata.json`

### Внешний S3

```bash
# Используя AWS CLI
aws s3 ls s3://jitsi-recordings/recordings/testmeet/ --endpoint-url https://api.storage.recontext.online

# Или MinIO CLI
mc ls minio/jitsi-recordings/recordings/testmeet/
```

## Проверка metadata.json

Скачать metadata.json и проверить структуру:

```json
{
  "conferenceId": "cb81d3fd_20251108_090049",
  "roomName": "testmeet",
  "startTime": "2025-11-08T09:00:49.123Z",
  "endTime": "2025-11-08T09:01:19.456Z",
  "durationSeconds": 30.33,
  "participantsCount": 1,
  "totalSessions": 2,
  "participants": [
    {
      "participantId": "a719784e-5c6e-4988-b6ca-349a419a4e3e@meet.jitsi",
      "displayName": "testmeet@muc.meet.jitsi/a719784e",
      "totalDurationSeconds": 30.1,
      "sessionsCount": 2,
      "sessions": [
        {
          "sessionId": "a719784e_1762591512.328",
          "filename": "testmeet_a719784e-5c6e-4988-b6ca-349a419a4e3e_20251108_090049.opus",
          "s3Key": "recordings/testmeet/cb81d3fd_20251108_090049/testmeet_a719784e-5c6e-4988-b6ca-349a419a4e3e_20251108_090049.opus",
          "joinOffsetSeconds": 0.06,
          "durationSeconds": 15.2,
          "isReconnection": false
        },
        {
          "sessionId": "a719784e_1762591527.456",
          "filename": "testmeet_a719784e-5c6e-4988-b6ca-349a419a4e3e_20251108_090104.opus",
          "s3Key": "recordings/testmeet/cb81d3fd_20251108_090049/testmeet_a719784e-5c6e-4988-b6ca-349a419a4e3e_20251108_090104.opus",
          "joinOffsetSeconds": 15.3,
          "durationSeconds": 14.9,
          "isReconnection": true
        }
      ]
    }
  ]
}
```

## Типичные проблемы

### 1. Файлы не загружаются на S3

**Проверить:**
```bash
docker-compose logs recorder | grep "S3"
```

**Ошибка "S3 API Requests must be made to API port":**
- Неправильный порт в S3_ENDPOINT
- Должен быть `http://minio:9000`, а не `http://minio:9001`

**Ошибка "Access Denied":**
- Неправильные AWS_ACCESS_KEY_ID или AWS_SECRET_ACCESS_KEY
- Проверить credentials в .env

### 2. Нет событий от Prosody

**Проверить:**
```bash
docker-compose logs prosody | grep "recorder"
docker-compose logs recorder | grep "Incoming webhook"
```

**Если нет событий:**
- Проверить что `RECORDER_WEBHOOK_URL=http://recorder:8080/events` в .env
- Проверить что `XMPP_MUC_MODULES=muc_recorder_events` в .env
- Перезапустить Prosody: `docker-compose restart prosody`

### 3. FFmpeg не запускается

**Проверить логи:**
```bash
docker-compose exec recorder ls -la /tmp/recordings/
```

**Если файлов нет:**
- Проверить что stream_url корректный
- Проверить доступность JVB

### 4. Пустые файлы (0 bytes)

**Причина:** FFmpeg не смог подключиться к потоку

**Действия:**
- Проверить что JVB доступен из контейнера recorder
- Проверить настройки STUN/TURN в .env
- Увеличить логирование FFmpeg

## Мониторинг

### Проверка активных конференций

```bash
curl http://localhost:8080/health | jq
```

### Логи в реальном времени

```bash
# Все логи
docker-compose logs -f recorder

# Только важные события
docker-compose logs -f recorder | grep -E "(JOINED|LEFT|RECORDING|UPLOAD|Conference)"

# Только ошибки
docker-compose logs -f recorder | grep -E "(ERROR|❌)"
```

### Проверка использования диска

```bash
docker-compose exec recorder du -sh /tmp/recordings/
```

### Проверка процессов FFmpeg

```bash
docker-compose exec recorder ps aux | grep ffmpeg
```

## Очистка после тестов

```bash
# Остановить все
docker-compose down

# Удалить volumes (ОСТОРОЖНО: удаляет все записи!)
docker-compose down -v

# Очистить только локальные записи
docker-compose exec recorder rm -rf /tmp/recordings/*
```

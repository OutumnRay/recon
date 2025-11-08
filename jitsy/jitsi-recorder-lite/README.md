# Jitsi Recorder Lite - WebRTC Edition

Автоматический рекордер для Jitsi Meet с прямым WebRTC подключением через XMPP.

## Возможности

- ✅ **XMPP + WebRTC подключение** - подключение через XMPP WebSocket и Jingle signaling
- ✅ **Запись индивидуальных audio tracks** - каждый участник записывается в отдельный файл
- ✅ **SFU (Selective Forwarding Unit)** - JVB отправляет индивидуальные потоки каждого участника
- ✅ **Отдельные файлы при переподключениях** - если участник переподключается, создается новый track и файл
- ✅ **Формат Opus** - запись в эффективный Opus формат через PyAV
- ✅ **Автоматическая загрузка на S3** - опциональная загрузка файлов на S3/MinIO
- ✅ **Чистый Python WebRTC** - используется aiortc для WebRTC, без зависимостей от FFmpeg
- ✅ **Без браузера** - не требует Playwright или Selenium

## Архитектура WebRTC Recorder

### Как это работает

```
┌─────────────────┐
│  Jitsi Meet     │
│  (Web UI)       │
└────────┬────────┘
         │
    ┌────▼────────────────────────┐
    │  Prosody XMPP Server        │
    │  (wss://...xmpp-websocket)  │
    └────┬────────────────────┬───┘
         │                    │
    ┌────▼────┐          ┌────▼──────────┐
    │  Focus  │          │  WebRTC Bot   │ ← Наш recorder
    │ (Jicofo)│          │  (Python)     │
    └────┬────┘          └────┬──────────┘
         │                    │
         │  Jingle signaling  │
         │◄───────────────────┤
         │                    │
    ┌────▼────────────────────▼───┐
    │  Jitsi Videobridge (JVB)    │
    │  WebRTC SFU                 │
    │  Sends individual streams   │
    └─────────────────────────────┘
```

1. **XMPP Connection** - подключается к `wss://meet.recontext.online/xmpp-websocket`
2. **SASL Authentication** - проходит ANONYMOUS аутентификацию
3. **MUC Join** - присоединяется к Multi-User Chat комнате
4. **Jingle Signaling** - обменивается Jingle сообщениями с focus.meet.jitsi
5. **WebRTC Setup** - конвертирует Jingle XML в SDP и устанавливает WebRTC соединение с JVB
6. **Audio Tracks** - JVB отправляет индивидуальные audio tracks каждого участника (SFU mode)
7. **Recording** - каждый track записывается в отдельный opus файл используя PyAV
8. **S3 Upload** - опционально загружает файлы на S3/MinIO

### Технологический стек

- **Python 3.11** - основной язык
- **aiortc** - WebRTC implementation для Python
- **PyAV (av)** - кодирование audio в opus формат
- **websockets** - WebSocket клиент для XMPP
- **aiohttp** - HTTP health check сервер
- **boto3** (опционально) - загрузка на S3/MinIO

## Основной компонент - WebRTC Bot

Основной файл: **`jitsi_webrtc_bot.py`** - полноценный WebRTC рекордер.

Этот бот:
- Подключается к XMPP WebSocket (`wss://meet.recontext.online/xmpp-websocket`)
- Проходит SASL ANONYMOUS аутентификацию
- Присоединяется к MUC комнате (`testmeet@muc.meet.jitsi`)
- Запрашивает конференцию у `focus.meet.jitsi`
- Получает Jingle session-initiate с SDP offer
- Конвертирует Jingle XML в SDP для aiortc
- Устанавливает WebRTC PeerConnection с JVB
- Получает индивидуальные audio tracks и записывает их в Opus файлы

**Подробная документация**: см. `WEBRTC_BOT_README.md`

### Быстрый старт

```bash
# Локально
export JITSI_URL=https://meet.recontext.online
export JITSI_ROOM=testmeet
export RECORD_DIR=./recordings

python jitsi_webrtc_bot.py
```

```bash
# В Docker
docker-compose up --build recorder
```

## Вспомогательные компоненты

### jitsi_xmpp_client.py
Минимальный XMPP клиент - базовый компонент для подключения к Jitsi XMPP серверу.
Используется внутри `jitsi_webrtc_bot.py`.

### jingle_sdp.py
Конвертер между Jingle XML и SDP форматами. Преобразует Jingle сообщения от Jicofo
в SDP для aiortc и обратно.

## Как работает запись по потокам

### Одно подключение = один track = один файл

Каждый раз когда участник подключается к конференции, создается **новый файл**:

```
Участник подключился → testmeet_uuid1_20251108_090100.opus (файл #1)
Участник отключился → файл #1 загружается на S3

Участник переподключился → testmeet_uuid1_20251108_090200.opus (файл #2)
Участник отключился → файл #2 загружается на S3

Конференция завершена → metadata.json содержит оба файла
```

### Пример: участник с нестабильным интернетом

```
09:00:00 - Иван подключился
         → testmeet_uuid-ivan_20251108_090000.opus (запись началась)

09:05:30 - У Ивана оборвалась связь
         → testmeet_uuid-ivan_20251108_090000.opus (5.5 минут) ✅ загружено на S3

09:05:45 - Иван переподключился
         → testmeet_uuid-ivan_20251108_090545.opus (новая запись началась, isReconnection=true)

09:10:00 - Конференция завершилась
         → testmeet_uuid-ivan_20251108_090545.opus (4.25 минут) ✅ загружено на S3
         → metadata.json содержит 2 файла Ивана
```

### В итоге на S3:

```
s3://jitsi-recordings/recordings/testmeet/{conference_id}/
  testmeet_uuid-ivan_20251108_090000.opus  (5.5 минут)
  testmeet_uuid-ivan_20251108_090545.opus  (4.25 минут, reconnection)
  metadata.json
```

### Автоматическая загрузка

Файлы загружаются **сразу после остановки записи**, не дожидаясь конца конференции:

- ✅ **Плюс**: файлы доступны на S3 сразу после отключения участника
- ✅ **Плюс**: не теряются данные если рекордер упадет во время конференции
- ✅ **Плюс**: меньше нагрузка на диск рекордера

## Формат имён файлов

```
{room_name}_{participant_id}_{timestamp}.opus
```

**Пример:**
```
testmeet_a719784e-5c6e-4988-b6ca-349a419a4e3e_20251108_084512_270.opus
```

Где:
- `testmeet` - название комнаты
- `a719784e-5c6e-4988-b6ca-349a419a4e3e` - UUID участника (из `participantId` до `@`)
- `20251108_084512_270` - timestamp (год-месяц-день_час-минута-секунда_миллисекунды)

## Структура S3

```
s3://jitsi-recordings/
  recordings/
    {room_name}/
      {conference_id}/
        {participant1_id}_{timestamp}.opus
        {participant2_id}_{timestamp}.opus
        metadata.json
```

**Пример:**
```
s3://jitsi-recordings/recordings/testmeet/a1b2c3d4_20251108_084512/
  testmeet_a719784e-5c6e-4988-b6ca-349a419a4e3e_20251108_084512_270.opus
  testmeet_b829c94f-7d8a-5bc9-c8db-459b520b5f4f_20251108_084515_123.opus
  metadata.json
```

## Конфигурация (.env)

### S3 Storage (обязательно)

**ВАЖНО**: MinIO использует два порта:
- **9000** - API порт (для S3 операций) ← используйте этот!
- **9001** - Console порт (веб-интерфейс)

**Для внутри Docker сети** (рекомендуется):
```bash
S3_ENDPOINT=http://minio:9000
S3_BUCKET=jitsi-recordings
AWS_ACCESS_KEY_ID=minioadmin
AWS_SECRET_ACCESS_KEY=minioadmin
AWS_REGION=us-east-1
```

**Для внешнего доступа**:
```bash
S3_ENDPOINT=https://storage.recontext.online:9000
S3_BUCKET=jitsi-recordings
AWS_ACCESS_KEY_ID=minioadmin
AWS_SECRET_ACCESS_KEY=minioadmin
AWS_REGION=us-east-1
```

### Webhook (опционально)
```bash
WEBHOOK_URL=https://your-domain.com/api/webhook/recording-finished
```

### Аудио настройки
```bash
AUDIO_BITRATE=48k
```

### Prosody интеграция
```bash
RECORDER_WEBHOOK_URL=http://recorder:8080/events
XMPP_MUC_MODULES=muc_recorder_events
```

## Метаданные (metadata.json)

Каждая конференция сохраняет метаданные в `metadata.json`:

```json
{
  "conferenceId": "a1b2c3d4_20251108_084512",
  "roomName": "testmeet",
  "startTime": "2025-11-08T08:45:12.270Z",
  "endTime": "2025-11-08T08:45:41.880Z",
  "durationSeconds": 29.61,
  "participantsCount": 1,
  "totalSessions": 1,
  "participants": [
    {
      "participantId": "a719784e-5c6e-4988-b6ca-349a419a4e3e@meet.jitsi",
      "participantName": "a719784e-5c6e-4988-b6ca-349a419a4e3e@meet.jitsi",
      "displayName": "testmeet@muc.meet.jitsi/a719784e",
      "totalDurationSeconds": 29.54,
      "sessionsCount": 1,
      "sessions": [
        {
          "sessionId": "a719784e_1762591512.328",
          "participantId": "a719784e-5c6e-4988-b6ca-349a419a4e3e@meet.jitsi",
          "filename": "testmeet_a719784e-5c6e-4988-b6ca-349a419a4e3e_20251108_084512_328.opus",
          "s3Key": "recordings/testmeet/a1b2c3d4_20251108_084512/testmeet_a719784e-5c6e-4988-b6ca-349a419a4e3e_20251108_084512_328.opus",
          "joinOffsetSeconds": 0.06,
          "durationSeconds": 29.54,
          "isReconnection": false
        }
      ]
    }
  ]
}
```

## Webhook формат

При завершении конференции отправляется POST запрос на `WEBHOOK_URL`:

```json
{
  "event": "conferenceEnded",
  "conferenceId": "a1b2c3d4_20251108_084512",
  "roomName": "testmeet",
  "startTime": "2025-11-08T08:45:12.270Z",
  "endTime": "2025-11-08T08:45:41.880Z",
  "durationSeconds": 29.61,
  "participantsCount": 1,
  "totalSessions": 1,
  "participants": [
    {
      "participantId": "a719784e-5c6e-4988-b6ca-349a419a4e3e@meet.jitsi",
      "displayName": "testmeet@muc.meet.jitsi/a719784e",
      "totalDurationSeconds": 29.54,
      "sessionsCount": 1,
      "recordings": [
        {
          "filename": "testmeet_a719784e-5c6e-4988-b6ca-349a419a4e3e_20251108_084512_328.opus",
          "s3Key": "recordings/testmeet/a1b2c3d4_20251108_084512/testmeet_a719784e-5c6e-4988-b6ca-349a419a4e3e_20251108_084512_328.opus",
          "s3Url": "s3://jitsi-recordings/recordings/testmeet/a1b2c3d4_20251108_084512/testmeet_a719784e-5c6e-4988-b6ca-349a419a4e3e_20251108_084512_328.opus",
          "joinOffsetSeconds": 0.06,
          "durationSeconds": 29.54,
          "isReconnection": false
        }
      ]
    }
  ],
  "s3Path": "recordings/testmeet/a1b2c3d4_20251108_084512/"
}
```

## Логи

Рекордер выводит подробные логи:

```
2025-11-08 08:45:41 - INFO - ✅ Conference testmeet completed:
2025-11-08 08:45:41 - INFO -    Conference ID: a1b2c3d4_20251108_084512
2025-11-08 08:45:41 - INFO -    Duration: 29.6s
2025-11-08 08:45:41 - INFO -    Participants: 1
2025-11-08 08:45:41 - INFO -    Total sessions: 1
2025-11-08 08:45:41 - INFO -    Recordings uploaded to S3: 1/1
2025-11-08 08:45:41 - INFO -    S3 Path: s3://jitsi-recordings/recordings/testmeet/a1b2c3d4_20251108_084512/
2025-11-08 08:45:41 - INFO -    📁 Recordings:
2025-11-08 08:45:41 - INFO -       ✅ testmeet_a719784e-5c6e-4988-b6ca-349a419a4e3e_20251108_084512_328.opus (29.5s)
```

## Health Check

```bash
curl http://localhost:8080/health
```

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

## Запуск

```bash
docker-compose up -d recorder
```

## Просмотр логов

```bash
docker-compose logs -f recorder
```

## Требования

- **Docker** & **Docker Compose**
- **Redis** - для координации между инстансами
- **S3/MinIO** - для хранения записей
- **FFmpeg** - встроен в Docker образ
- **Python 3.11+** - встроен в Docker образ

## Troubleshooting

### Файлы не загружаются на S3

Проверьте логи при старте:
```bash
docker-compose logs recorder | grep S3
```

Должны быть:
```
✅ S3 client initialized - endpoint: https://storage.recontext.online/, bucket: jitsi-recordings
✅ S3 bucket 'jitsi-recordings' exists and is accessible
```

Если видите ошибки:
1. Проверьте доступность S3_ENDPOINT
2. Проверьте правильность AWS_ACCESS_KEY_ID и AWS_SECRET_ACCESS_KEY
3. Проверьте наличие бакета

### Нет записей участников

Проверьте:
1. Настроена ли Prosody интеграция (`RECORDER_WEBHOOK_URL` в .env)
2. Установлен ли модуль `muc_recorder_events` в Prosody
3. Есть ли webhook'и от Prosody в логах:
```bash
docker-compose logs recorder | grep "Incoming webhook"
```

### ProcessLookupError

Это нормально - означает что FFmpeg процесс уже завершился до вызова terminate(). Рекордер обрабатывает это корректно.

## Разработка

```bash
# Rebuild образа
docker-compose build recorder

# Рестарт
docker-compose restart recorder

# Логи с детализацией
docker-compose logs -f recorder
```

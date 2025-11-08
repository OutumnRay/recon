# 🎬 Jitsi Unified Recorder

**Объединенный сервис** для автоматической записи Jitsi конференций через Kurento Media Server.

Объединяет функционал:
- ✅ Auto Recorder (webhook listener для Prosody)
- ✅ Kurento Recorder API (управление Kurento через WebSocket)
- ✅ MinIO upload (автоматическая загрузка на S3)

## 🎯 Что это делает

1. **Слушает события** от Prosody (participantJoined/Left/conferenceEnded)
2. **Автоматически создает** Kurento recording pipeline для каждого участника
3. **Записывает** аудио/видео через Kurento Media Server
4. **Загружает** файлы на MinIO при завершении

## 🚀 Запуск

### На сервере

```bash
cd ~/jitsi-docker-jitsi-meet-3bc1ebc

# Остановите старые сервисы (если были)
docker-compose stop kurento-recorder auto-recorder

# Удалите старые контейнеры
docker-compose rm -f kurento-recorder auto-recorder

# Соберите unified-recorder
docker-compose build recorder

# Запустите всё
docker-compose up -d kurento recorder

# Перезапустите Prosody для webhook
docker-compose restart prosody

# Проверьте статус
docker-compose ps kurento recorder
```

### Проверка

```bash
# Health check
curl http://localhost:9888/health

# Статистика
curl http://localhost:9888/stats

# Логи
docker-compose logs -f recorder
```

## 📊 API Endpoints

### GET /health

Health check сервиса

**Response:**
```json
{
  "status": "healthy",
  "worker_id": "unified-recorder",
  "auto_record": true,
  "active_conferences": 0,
  "active_sessions": 0,
  "kurento_uri": "ws://kurento:8888/kurento",
  "s3_configured": true
}
```

### GET /stats

Статистика активных записей

**Response:**
```json
{
  "active_conferences": [
    {
      "room": "testmeeting",
      "started_at": "2025-11-08T15:00:00Z",
      "participants": 2
    }
  ],
  "active_sessions": [
    {
      "participant_id": "user@domain/res",
      "display_name": "John Doe",
      "room": "testmeeting",
      "started_at": "2025-11-08T15:00:05Z",
      "duration": 125.5,
      "filename": "testmeeting_user_20251108_150005.webm"
    }
  ]
}
```

### POST /events

Webhook endpoint для Prosody

**Request body:**
```json
{
  "eventType": "participantJoined",
  "roomName": "testmeeting",
  "endpointId": "abc123",
  "participantId": "user@domain/res",
  "displayName": "John Doe"
}
```

## 🏗️ Архитектура

```
┌─────────────┐
│   Prosody   │
└──────┬──────┘
       │ webhook
       ↓
┌──────────────────────┐
│ Unified Recorder     │
│ - Webhook listener   │
│ - Kurento client     │
│ - S3 uploader        │
└──────┬───────────────┘
       │ WebSocket
       ↓
┌──────────────────────┐
│ Kurento Media Server │
│ - MediaPipeline      │
│ - RtpEndpoint        │◄───── НУЖНО: RTP от JVB
│ - RecorderEndpoint   │
└──────┬───────────────┘
       │ write files
       ↓
┌──────────────────────┐
│   /recordings        │
└──────┬───────────────┘
       │ upload
       ↓
┌──────────────────────┐
│   MinIO (S3)         │
└──────────────────────┘
```

## ⚠️ ВАЖНО: Настройка JVB

Для полной работы нужно настроить **JVB для forwarding RTP в Kurento**.

### Что нужно сделать:

1. **Получить SDP offer** от Kurento RtpEndpoint
2. **Отправить в JVB Colibri API** для создания channel
3. **Получить SDP answer** от JVB
4. **Применить к Kurento RtpEndpoint**
5. **Обменяться ICE candidates**

### Пример интеграции с Colibri:

```python
async def setup_jvb_forwarding(session: SessionInfo):
    """Настройка forwarding от JVB к Kurento"""

    # 1. Получить SDP offer от Kurento
    sdp_offer = await kurento.generate_offer(session.rtp_endpoint_id)

    # 2. Создать channel в JVB через Colibri REST API
    async with ClientSession() as http:
        colibri_url = f"http://{JVB_HOST}:{JVB_COLIBRI_PORT}/colibri/conferences"

        # Создать конференцию
        conf_response = await http.post(colibri_url, json={
            "contents": [{
                "name": "audio",
                "channels": [{
                    "endpoint": session.participant_id,
                    "rtp-level-relay-type": "translator"
                }]
            }]
        })
        conf_data = await conf_response.json()
        conference_id = conf_data["id"]

        # Создать channel для participant
        channel_url = f"{colibri_url}/{conference_id}"
        channel_response = await http.patch(channel_url, json={
            "channel-bundles": [{
                "id": "bundle-1",
                "transport": {
                    "xmlns": "urn:xmpp:jingle:transports:ice-udp:1",
                    "rtcp-mux": True,
                    "ice-controlling": True
                }
            }],
            "contents": [{
                "name": "audio",
                "channels": [{
                    "id": session.participant_id,
                    "endpoint": session.participant_id,
                    "initiator": False,
                    "direction": "sendrecv",
                    "channel-bundle-id": "bundle-1",
                    "sources": [],
                    "payload-types": [{
                        "id": 111,
                        "name": "opus",
                        "clockrate": 48000,
                        "channels": 2
                    }]
                }]
            }]
        })

        channel_data = await channel_response.json()

        # 3. Получить SDP answer от JVB
        sdp_answer = channel_data["contents"][0]["channels"][0]["transport"]

        # 4. Применить к Kurento
        await kurento.process_answer(session.rtp_endpoint_id, sdp_answer)

        logger.info(f"✅ JVB forwarding configured for {session.participant_id}")
```

### Альтернативный подход: Jibri-style

Можно использовать подход как у Jibri:
1. Создать "виртуального участника" который присоединяется к конференции
2. Получать медиа через обычный WebRTC как клиент
3. Передавать в Kurento для записи

Это проще, но требует полной WebRTC реализации.

## 🧪 Тестирование (без JVB forwarding)

Текущая реализация создаст Kurento pipeline, но **не получит реальное медиа** от JVB.

Для тестирования можно:

### 1. Проверить что pipeline создается

```bash
# Зайдите в конференцию
open https://meet.recontext.online/test123

# Проверьте логи
docker-compose logs -f recorder

# Должны увидеть:
# 🎙️ Starting recording: ...
# ✅ Recording started: ...
# Pipeline: xxx-yyy-zzz
# RTP Endpoint: aaa-bbb-ccc
# Recorder: ddd-eee-fff
```

### 2. Проверить что файлы создаются (пустые пока нет медиа)

```bash
# Файлы будут созданы, но пустые (0 байт)
ls -lh recordings/

# После выхода из конференции:
# ⚠️ Empty file, skipping upload
```

### 3. Тестовая запись с fake stream

Можно протестировать Kurento напрямую:

```bash
# Войдите в контейнер recorder
docker-compose exec recorder python3

# Создайте тестовую запись
>>> import asyncio
>>> from unified_recorder import kurento, SessionInfo
>>>
>>> async def test():
...     session = SessionInfo("test@test", "testroom", "Test User")
...     await session.start_recording()
...     await asyncio.sleep(5)
...     await session.stop_recording()
>>>
>>> asyncio.run(test())
```

## 📁 Структура на MinIO

Записи сохраняются по пути:
```
jitsi-recordings/
  recordings/
    {room_name}/
      {conference_id}/
        {participant}_{timestamp}.webm
```

Пример:
```
jitsi-recordings/
  recordings/
    testmeeting/
      abc123_20251108/
        john_doe_20251108_150000.webm
        jane_smith_20251108_150005.webm
```

## 🔧 Переменные окружения

| Переменная | По умолчанию | Описание |
|-----------|--------------|----------|
| `KURENTO_URI` | `ws://kurento:8888/kurento` | Kurento WebSocket |
| `MINIO_ENDPOINT` | `https://api.storage.recontext.online` | MinIO endpoint |
| `MINIO_ACCESS_KEY` | `minioadmin` | MinIO access key |
| `MINIO_SECRET_KEY` | `minioadmin` | MinIO secret key |
| `MINIO_BUCKET` | `jitsi-recordings` | Bucket name |
| `AUTO_RECORD` | `true` | Включить автозапись |
| `JITSI_DOMAIN` | `meet.recontext.online` | Jitsi domain |
| `JVB_HOST` | `jvb` | JVB hostname |
| `JVB_COLIBRI_PORT` | `8080` | Colibri REST API port |

## 📝 TODO: Полная интеграция

Для полной работы нужно добавить:

- [ ] SDP offer/answer exchange с JVB Colibri API
- [ ] ICE candidates negotiation
- [ ] DTLS/SRTP setup
- [ ] Participant stream selection (выбор конкретного участника)
- [ ] Retry logic для reconnect
- [ ] Graceful shutdown при сбоях

## 🏆 Авторы

**Святослав Иванов**
AI-инженер и архитектор решений
2025

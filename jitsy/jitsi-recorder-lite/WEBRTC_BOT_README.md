# Jitsi WebRTC Bot - Инструкция по запуску

Полноценный recorder бот который подключается к Jitsi Meet через XMPP WebSocket,
устанавливает WebRTC соединение с JVB (Jitsi Videobridge) и записывает индивидуальные audio tracks.

## Архитектура

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

## Как это работает

1. **XMPP подключение**
   - Бот подключается к `wss://meet.recontext.online/xmpp-websocket`
   - Проходит SASL ANONYMOUS аутентификацию
   - Присоединяется к MUC комнате `testmeet@muc.meet.jitsi`

2. **Jingle Signaling**
   - Отправляет IQ conference к `focus.meet.jitsi`
   - Получает `session-initiate` с описанием медиа потоков
   - Конвертирует Jingle XML в SDP для aiortc
   - Отправляет `session-accept` обратно

3. **WebRTC соединение**
   - Устанавливает WebRTC PeerConnection с JVB
   - Обменивается ICE candidates
   - Получает индивидуальные audio tracks от каждого участника
   - JVB работает как SFU (Selective Forwarding Unit)

4. **Запись**
   - Каждый audio track записывается в отдельный opus файл
   - Формат: `{room}_{track_id}_{timestamp}.opus`
   - Файлы сохраняются в RECORD_DIR

## Установка

### Зависимости

```bash
pip install -r requirements.txt
```

Требуются:
- `websockets` - для XMPP WebSocket
- `aiortc` - для WebRTC
- `av` (PyAV) - для кодирования в opus
- `boto3` - для S3 upload (опционально)

### Системные пакеты

На Mac:
```bash
brew install ffmpeg opus libvpx libsrtp
```

На Ubuntu/Debian:
```bash
apt-get install ffmpeg libopus-dev libvpx-dev libsrtp2-dev
```

## Запуск

### Локальный запуск

```bash
export JITSI_URL=https://meet.recontext.online
export JITSI_ROOM=testmeet
export RECORD_DIR=./recordings
export LOG_LEVEL=DEBUG

python jitsi_webrtc_bot.py
```

### Docker

Рекордер интегрирован в docker-compose.yml основного Jitsi стека:

```bash
# Запуск рекордера
cd /path/to/jitsy
docker-compose up --build recorder

# Просмотр логов
docker-compose logs -f recorder

# Остановка
docker-compose stop recorder
```

Или локально в jitsi-recorder-lite:

```bash
cd jitsi-recorder-lite
docker build -t jitsi-webrtc-bot .
docker run -it --rm \
  -e JITSI_URL=https://meet.recontext.online \
  -e JITSI_ROOM=testmeet \
  -e RECORD_DIR=/tmp/recordings \
  -v /tmp/recordings:/tmp/recordings \
  jitsi-webrtc-bot
```

## Переменные окружения

```bash
# Обязательные
JITSI_URL=https://meet.recontext.online
JITSI_ROOM=testmeet

# Опциональные (есть дефолты)
XMPP_DOMAIN=meet.jitsi
MUC_DOMAIN=muc.meet.jitsi
FOCUS_JID=focus.meet.jitsi
BOT_NICKNAME=recorder
RECORD_DIR=./recordings
LOG_LEVEL=DEBUG

# S3 (опционально)
S3_ENDPOINT=https://api.storage.recontext.online
S3_BUCKET=jitsi-recordings
AWS_ACCESS_KEY_ID=minioadmin
AWS_SECRET_ACCESS_KEY=minioadmin
```

## Тестирование

### 1. Запустите бота

```bash
python jitsi_webrtc_bot.py
```

### 2. Ожидаемые логи

```
🔌 Connecting to xmpp-websocket: wss://meet.recontext.online/xmpp-websocket?room=testmeet
🔐 SASL ANONYMOUS success
📛 Bound JID: 12345678-abcd-efgh@meet.jitsi/resource
🚪 Joined MUC testmeet@muc.meet.jitsi as recorder
✅ XMPP connection ready as 12345678-abcd-efgh@meet.jitsi/resource
🎯 Focus response: type=result
🎬 Jingle session-initiate from focus@auth.meet.jitsi/focus (sid=26p3nmbipb479)
🌐 Setting up WebRTC connection...
📄 Converted SDP:
v=0
o=- 0 0 IN IP4 0.0.0.0
s=-
t=0 0
...
✅ WebRTC setup complete
📤 Sending session-accept...
✅ Session-accept sent
🔗 ICE connection state: checking
🔗 ICE connection state: connected
📥 Received track: kind=audio, id=337153032
🎙️  Started recording track 337153032: testmeet_337153032_20251108_143000.opus
```

### 3. Откройте конференцию

В браузере откройте `https://meet.recontext.online/testmeet`

### 4. Проверьте записи

```bash
ls -lh recordings/
# testmeet_337153032_20251108_143000.opus
# testmeet_1372697452_20251108_143005.opus  <- второй участник
```

## Troubleshooting

### Ошибка: WebSocket connection failed

Проверьте:
- Доступность `wss://meet.recontext.online/xmpp-websocket`
- Правильность JITSI_URL
- TLS сертификат сервера (сейчас проверка отключена)

### Ошибка: SASL authentication failed

Убедитесь что:
- Prosody поддерживает ANONYMOUS auth
- MUC компонент настроен правильно

### Ошибка: No session-initiate received

Проверьте:
- Focus (Jicofo) работает и доступен
- Отправляется правильный IQ conference запрос
- Логи Jicofo на наличие ошибок

### Ошибка: ICE connection failed

Возможные причины:
- Firewall блокирует UDP порты
- STUN/TURN серверы недоступны
- Неправильные ICE candidates

Проверьте:
```bash
# Доступность JVB
ping 185.233.186.144
nc -u 185.233.186.144 10000
```

### No audio tracks received

Проверьте:
- WebRTC connection state = connected
- Логи JVB на наличие ошибок
- SDP offer/answer корректны
- DTLS handshake успешен

## Отладка

### Уровень логирования

Установите `LOG_LEVEL=DEBUG` для детальных логов:

```bash
LOG_LEVEL=DEBUG python jitsi_webrtc_bot.py
```

### Логирование XMPP stanzas

В `jitsi_xmpp_client.py` раскомментируйте:

```python
logger.debug("↩️  %s", ET.tostring(elem, encoding="unicode"))
```

### Логирование WebRTC stats

Добавьте в `jitsi_webrtc_bot.py`:

```python
@self.pc.on("icegatheringstatechange")
async def on_ice_gathering_state_change():
    logger.info(f"🧊 ICE gathering state: {self.pc.iceGatheringState}")
```

## Известные ограничения

1. **TLS verification disabled**
   - Для тестирования отключена проверка TLS сертификатов
   - В production нужно включить и настроить CA certificates

2. **Simplified Jingle-SDP conversion**
   - Конвертация Jingle ↔ SDP упрощенная
   - Поддерживаются основные параметры (audio/video/data, ICE, DTLS)
   - Могут быть проблемы с некоторыми codecs или extensions

3. **Audio only**
   - По умолчанию записывается только audio
   - Video tracks игнорируются (можно добавить)

4. **No reconnection handling**
   - Если WebRTC disconnected - бот не переподключается автоматически
   - Нужно перезапустить

## Следующие шаги

### Интеграция с S3

Добавить автоматическую загрузку записей на S3:

```python
# В AudioTrackRecorder.stop()
if config.S3_ENDPOINT:
    await upload_to_s3(self.filepath, bucket, key)
```

### Интеграция с Prosody webhooks

Для production нужно:
1. Получать события participantJoined/Left от Prosody
2. Связывать WebRTC tracks с participant IDs
3. Генерировать metadata.json с информацией о сессиях

### Поддержка нескольких комнат

Запускать несколько WebRTC ботов параллельно:

```python
async def record_multiple_rooms(rooms):
    tasks = []
    for room in rooms:
        bot = JitsiWebRTCBot(room, output_dir)
        tasks.append(bot.connect())
    await asyncio.gather(*tasks)
```

## Полезные ссылки

- [Jitsi Meet Handbook](https://jitsi.github.io/handbook/)
- [Jingle Protocol (XEP-0166)](https://xmpp.org/extensions/xep-0166.html)
- [aiortc Documentation](https://aiortc.readthedocs.io/)
- [PyAV Documentation](https://pyav.org/)

## Контрибьюция

Если нашли баг или хотите добавить фичу - создайте issue или PR!

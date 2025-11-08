# WebRTC Recorder - Статус реализации

## ✅ Что сделано

### 1. Базовая архитектура
- ✅ Создан `simple_webrtc_recorder.py` - чистый WebRTC recorder на aiortc
- ✅ Интегрирован в `recorder.py` - WebRTC recorder запускается при первом участнике
- ✅ Обновлен Dockerfile - добавлены зависимости для aiortc (opus, libvpx, libsrtp)
- ✅ Обновлен requirements.txt - aiortc, av (PyAV), websockets

### 2. Основной функционал
- ✅ WebSocket подключение к JVB Colibri API
- ✅ WebRTC PeerConnection создание
- ✅ SDP offer/answer обмен
- ✅ ICE candidates handling
- ✅ Audio track запись в opus формат
- ✅ Автоматическое создание файлов для каждого track
- ✅ Интеграция с Prosody webhooks
- ✅ Метаданные и S3 upload (существующая логика)

### 3. Классы и модули
- `SimpleWebRTCRecorder` - основной WebRTC клиент
- `AudioTrackRecorder` - запись одного audio track в файл
- `ConferenceRecording.start_webrtc_recording()` - запуск recorder для конференции
- `ConferenceRecording.stop_webrtc_recording()` - остановка recorder

## ⚠️ Что нужно доработать

### Шаг 1: Тестирование WebSocket подключения к JVB
**Проблема:** Colibri WebSocket URL и протокол могут отличаться в зависимости от версии JVB.

**Что делать:**
1. Проверить URL формат: `ws://{jvb_host}:{jvb_port}/colibri-ws/default-id/{conference_id}`
2. Может потребоваться другой endpoint или формат conference_id
3. Проверить логи JVB на наличие Colibri WebSocket endpoint

**Файл:** `simple_webrtc_recorder.py:93`

### Шаг 2: SDP signaling
**Проблема:** Colibri protocol требует специфичный формат сообщений.

**Что делать:**
1. Изучить формат Colibri messages: https://github.com/jitsi/jitsi-videobridge/blob/master/doc/rest-colibri.md
2. Проверить правильность JSON сообщений (colibriClass, type, sdp)
3. Возможно нужно добавить дополнительные поля (endpoint ID, media types)

**Файл:** `simple_webrtc_recorder.py:117-126`

### Шаг 3: Audio track mapping
**Проблема:** Нужно связать WebRTC tracks с participant IDs от Prosody.

**Что делать:**
1. WebRTC track.id может не совпадать с Prosody endpoint_id
2. Нужно получить mapping из Colibri messages или JVB events
3. Возможно нужен дополнительный lookup через participant SSRC или другие идентификаторы

**Файл:** `simple_webrtc_recorder.py:103` (обработчик on_track)

### Шаг 4: Загрузка файлов на S3
**Проблема:** WebRTC recorder создает файлы, но логика S3 upload была удалена.

**Что делать:**
1. После остановки WebRTC recorder, получить список записанных файлов
2. Загрузить каждый файл на S3 (как в старой FFmpeg версии)
3. Обновить metadata.json с информацией о файлах
4. Связать WebRTC recording info с ParticipantSession metadata

**Файлы:**
- `simple_webrtc_recorder.py:83-86` (метод stop)
- `recorder.py:629` (handle_conference_ended)

### Шаг 5: ICE connectivity
**Проблема:** WebRTC может не установить соединение если ICE candidates неправильные.

**Что делать:**
1. Проверить STUN/TURN настройки в RTCConfiguration
2. Добавить логирование ICE connection state
3. Возможно нужно использовать те же STUN/TURN что и у Jitsi Meet клиентов

**Файл:** `simple_webrtc_recorder.py:89-91`

## 🔍 Отладка

### Как тестировать

1. **Запустить recorder:**
```bash
docker-compose up --build recorder
```

2. **Проверить логи:**
```bash
docker-compose logs -f recorder | grep -E "(WebRTC|WebSocket|track|JVB)"
```

3. **Создать тестовую конференцию:**
- Открыть https://meet.recontext.online/testroom
- Зайти с 1-2 участниками

4. **Ожидаемые логи:**
```
🔌 Starting WebRTC recorder for room: testroom
🔌 Connecting to WebSocket: ws://jvb:8080/colibri-ws/...
✅ Connected to JVB WebSocket
📤 Sent SDP offer to JVB
✅ Set remote description (answer)
📥 Received track: kind=audio, id=...
🎙️  Started recording: testroom_xxx_20251108_120000.opus
```

### Возможные ошибки

**WebSocket connection failed:**
- Проверить JVB_HOST и JVB_PORT в .env
- Убедиться что JVB доступен из контейнера recorder
- Проверить что Colibri WebSocket enabled в JVB config

**No tracks received:**
- Проверить SDP offer/answer в логах
- Проверить ICE connection state
- Убедиться что JVB отправляет audio tracks

**Files not created:**
- Проверить что output_dir существует и доступен для записи
- Проверить логи PyAV/av на ошибки кодирования

## 📚 Полезные ссылки

- [Jitsi Videobridge REST API](https://github.com/jitsi/jitsi-videobridge/blob/master/doc/rest.md)
- [Colibri Protocol](https://github.com/jitsi/jitsi-videobridge/blob/master/doc/rest-colibri.md)
- [aiortc Documentation](https://aiortc.readthedocs.io/)
- [PyAV Documentation](https://pyav.org/)

## 🎯 Приоритет следующих шагов

1. **Высокий:** Шаг 1 - проверить WebSocket подключение к JVB
2. **Высокий:** Шаг 2 - исправить SDP signaling
3. **Средний:** Шаг 3 - mapping tracks к participants
4. **Средний:** Шаг 4 - S3 upload интеграция
5. **Низкий:** Шаг 5 - ICE troubleshooting (если соединение не устанавливается)

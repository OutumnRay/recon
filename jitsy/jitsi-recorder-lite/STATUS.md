# Jitsi Recorder - Текущий статус

## ✅ Что сделано

### 1. Структура проекта
- ✅ Создан `config.py` - централизованная конфигурация с .env support
- ✅ Создан `.env.example` - пример настроек для локального запуска
- ✅ Создан `LOCAL_SETUP.md` - инструкции по локальной разработке

### 2. Recorder варианты

#### Вариант 1: Simple Jitsi Recorder (Playwright)
**Файл:** `simple_jitsi_recorder.py`

- ✅ Использует Playwright + headless Chromium
- ✅ Подключается к Jitsi Meet URL как участник
- ✅ Мониторит WebRTC connections через CDP
- ⏳ TODO: Извлечение и запись audio tracks (не завершено)

**Плюсы:**
- Простая интеграция
- Работает с любым Jitsi сервером
- Не требует XMPP signaling

**Минусы:**
- Требует Chromium (ресурсоемко)
- Сложнее извлечь raw audio data

#### Вариант 2: Jitsi Participant Recorder (aiortc)
**Файл:** `jitsi_participant_recorder.py`

- ✅ Базовая структура создана
- ⏳ TODO: XMPP signaling через aioxmpp
- ⏳ TODO: WebRTC через aiortc
- ⏳ TODO: Запись audio в opus

**Плюсы:**
- Легковесный (чистый Python, без браузера)
- Прямой доступ к audio tracks

**Минусы:**
- Сложная реализация XMPP/Jingle signaling
- Требует больше доработки

#### Вариант 3: Полный Recorder (с Prosody)
**Файл:** `recorder.py`

- ✅ Интеграция с Prosody webhooks
- ✅ Redis координация
- ✅ S3/MinIO upload
- ✅ Metadata tracking
- ⏳ TODO: Реальная запись audio (старая FFmpeg логика удалена)

### 3. Docker
- ✅ Dockerfile обновлен (Debian вместо Alpine)
- ✅ Playwright + Chromium установка
- ✅ Все Python файлы копируются

### 4. Зависимости
- ✅ aiortc - WebRTC для Python
- ✅ av (PyAV) - audio кодирование
- ✅ playwright - браузер автоматизация
- ✅ python-dotenv - .env support
- ✅ boto3, aiohttp, redis, websockets

## ⚠️ Что нужно доработать

### Приоритет 1: Завершить Playwright recorder

**Задача:** Извлечь audio tracks из браузера и записать в файлы.

**Подходы:**

1. **Chrome DevTools Protocol (CDP)** - перехват WebRTC media streams
   - Нужно: подписаться на события `WebRTC.addTrack`
   - Извлечь audio data через CDP Media API
   - Записать в opus файл

2. **Audio Capture через Page Context** - использовать MediaRecorder API
   - Инжектить JS в страницу
   - Создать MediaRecorder для каждого track
   - Скачать записанные blobs

3. **WebAudio API** - извлечь raw PCM data
   - Подключить audio tracks к ScriptProcessorNode
   - Получать PCM samples через JS
   - Передать в Python для кодирования в opus

**Рекомендация:** Попробовать подход #2 (MediaRecorder) - проще всего.

**Файл:** `simple_jitsi_recorder.py:80-120`

### Приоритет 2: Локальное тестирование

**Задача:** Проверить работу на локальной машине.

**Шаги:**
1. Установить зависимости: `pip install -r requirements.txt`
2. Установить Playwright: `playwright install chromium`
3. Создать .env файл
4. Запустить: `python simple_jitsi_recorder.py https://meet.recontext.online/testmeet`
5. Проверить что recorder подключается к конференции
6. Отладить извлечение audio

### Приоритет 3: Интеграция с основным recorder

**Задача:** Интегрировать Playwright recorder в `recorder.py`.

**Что нужно:**
1. Заменить `SimpleWebRTCRecorder` на `SimpleJitsiRecorder` в `recorder.py`
2. Убрать зависимость от JVB WebSocket URL
3. Использовать только Jitsi Meet URL (`https://meet.recontext.online/testmeet`)
4. Сохранить интеграцию с Prosody webhooks для метаданных

**Файл:** `recorder.py:368-410`

### Приоритет 4: XMPP signaling (опционально)

**Задача:** Реализовать чистый WebRTC подход через aioxmpp.

**Что нужно:**
1. XMPP подключение к Prosody (anonymous auth)
2. Join в MUC комнату
3. Jingle signaling для WebRTC (SDP exchange)
4. Обработка audio tracks через aiortc
5. Запись в opus через PyAV

**Файл:** `jitsi_participant_recorder.py`

**Примечание:** Это сложнее и можно отложить на потом, если Playwright подход заработает.

## 🎯 Текущая цель

**Первоочередная задача:** Заставить `simple_jitsi_recorder.py` записывать audio tracks.

**План действий:**

1. **Тестируем подключение:**
   ```bash
   python simple_jitsi_recorder.py https://meet.recontext.online/testmeet
   ```

2. **Добавляем MediaRecorder извлечение:**
   - Инжектим JS в страницу
   - Получаем все audio tracks
   - Запускаем MediaRecorder для каждого
   - Скачиваем blobs в Python
   - Конвертируем в opus если нужно

3. **Тестируем запись:**
   - Открываем конференцию в браузере
   - Проверяем что recorder видит audio tracks
   - Проверяем что создаются файлы

4. **Интегрируем в recorder.py:**
   - Заменяем WebSocket подход на Playwright
   - Проверяем работу с Prosody webhooks

## 📝 Следующие шаги

### Шаг 1: Реализовать MediaRecorder extraction

**Файл:** `simple_jitsi_recorder.py`

Добавить в метод `_monitor_webrtc()`:

```python
# Инжектим MediaRecorder для каждого track
await self.page.evaluate("""
    const tracks = [];
    const mediaElements = document.querySelectorAll('audio, video');

    for (const el of mediaElements) {
        if (el.srcObject) {
            const audioTracks = el.srcObject.getAudioTracks();
            for (const track of audioTracks) {
                const recorder = new MediaRecorder(new MediaStream([track]));
                const chunks = [];

                recorder.ondataavailable = (e) => chunks.push(e.data);
                recorder.onstop = async () => {
                    const blob = new Blob(chunks, {type: 'audio/webm'});
                    // Отправить blob в Python через CDP или page.evaluate
                };

                recorder.start();
            }
        }
    }
""")
```

### Шаг 2: Получить blobs в Python

Использовать CDP Events или expose Python функцию в JS context.

### Шаг 3: Конвертировать и сохранить

Конвертировать WebM в Opus используя FFmpeg или PyAV.

## 🔗 Полезные ресурсы

- [Playwright Python API](https://playwright.dev/python/docs/api/class-page)
- [Chrome DevTools Protocol](https://chromedevtools.github.io/devtools-protocol/)
- [MediaRecorder API](https://developer.mozilla.org/en-US/docs/Web/API/MediaRecorder)
- [PyAV Audio Encoding](https://pyav.org/docs/stable/cookbook/numpy.html)

## ✅ Готово к тестированию

Можно начинать локальное тестирование с `simple_jitsi_recorder.py`!

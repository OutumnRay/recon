# Локальный запуск Jitsi Recorder

Инструкция для тестирования recorder локально на вашем компьютере.

## Требования

- Python 3.11+
- FFmpeg
- Chromium (автоматически устанавливается через Playwright)

## Установка

### 1. Установите зависимости

```bash
cd jitsi-recorder-lite

# Создайте виртуальное окружение
python3 -m venv venv
source venv/bin/activate  # На Windows: venv\Scripts\activate

# Установите Python пакеты
pip install -r requirements.txt

# Установите Playwright браузеры
playwright install chromium
```

### 2. Создайте .env файл

```bash
# Скопируйте example файл
cp .env.example .env

# Отредактируйте .env под ваши нужды
nano .env
```

Минимальные настройки для локального запуска:

```bash
# .env
JITSI_URL=https://meet.recontext.online
RECORD_DIR=./recordings
LOG_LEVEL=DEBUG

# Опционально - для S3 upload
S3_ENDPOINT=https://api.storage.recontext.online
AWS_ACCESS_KEY_ID=your_key
AWS_SECRET_ACCESS_KEY=your_secret
```

## Запуск

### Простой тест (без Prosody)

Подключится к конференции и записывает audio:

```bash
python simple_jitsi_recorder.py https://meet.recontext.online/testmeet
```

Или используя run_local.py:

```bash
python run_local.py testmeet
```

### Полный режим (с Prosody webhooks)

Для этого нужен доступ к Prosody серверу. В .env укажите:

```bash
PROSODY_HOST=your_prosody_host
REDIS_HOST=your_redis_host
```

Затем запустите:

```bash
python recorder.py
```

## Структура проекта

```
jitsi-recorder-lite/
├── config.py                      # Конфигурация (читает из .env)
├── simple_jitsi_recorder.py       # Простой recorder (Playwright)
├── jitsi_participant_recorder.py  # WebRTC recorder (aiortc)
├── recorder.py                    # Полный recorder с Prosody
├── run_local.py                   # Скрипт для локального запуска
├── requirements.txt               # Python зависимости
├── .env.example                   # Пример настроек
└── recordings/                    # Папка с записями
```

## Как это работает

### 1. Simple Jitsi Recorder (Playwright)

- Запускает headless Chromium
- Подключается к Jitsi Meet URL как участник
- Перехватывает WebRTC audio tracks через CDP
- Записывает каждый track в отдельный opus файл

**Плюсы:**
- Простая реализация
- Работает с любым Jitsi сервером
- Не требует настройки XMPP

**Минусы:**
- Требует Chromium (больше ресурсов)
- Сложнее извлечь audio tracks из браузера

### 2. Jitsi Participant Recorder (aiortc)

- Использует чистый Python WebRTC (aiortc)
- Подключается через XMPP signaling
- Записывает audio tracks напрямую

**Плюсы:**
- Легковесный (без браузера)
- Прямой доступ к audio tracks

**Минусы:**
- Требует реализации XMPP/Jingle signaling
- Более сложная настройка

## Тестирование

### 1. Запустите recorder

```bash
python simple_jitsi_recorder.py https://meet.recontext.online/testmeet
```

### 2. Откройте конференцию в браузере

Откройте https://meet.recontext.online/testmeet в обычном браузере и подключитесь.

### 3. Проверьте логи

Вы должны увидеть:

```
🔌 Starting browser and connecting to: https://meet.recontext.online/testmeet
🌐 Navigating to: https://meet.recontext.online/testmeet
👤 Setting display name: Recorder Bot
✅ Connected to conference
🎧 Monitoring WebRTC connections...
🎙️  New audio track: xxx-yyy-zzz (Remote audio)
```

### 4. Проверьте записи

Файлы сохраняются в `./recordings/`:

```bash
ls -lh recordings/
```

## Отладка

### Проблема: Не подключается к конференции

- Проверьте что URL корректный
- Проверьте что Chromium установлен: `playwright install chromium`
- Запустите с `LOG_LEVEL=DEBUG` для детальных логов

### Проблема: Не записывает audio

- Пока не реализовано полное извлечение audio через CDP
- Это требует дополнительной доработки
- Используйте полный recorder с Prosody

### Проблема: Ошибки aiortc

- Убедитесь что FFmpeg установлен: `ffmpeg -version`
- На Mac: `brew install ffmpeg opus libvpx libsrtp`
- На Ubuntu: `apt-get install ffmpeg libopus-dev libvpx-dev libsrtp2-dev`

## Следующие шаги

1. ✅ Базовая структура и конфигурация
2. ✅ Playwright подключение к Jitsi
3. ⏳ Извлечение audio tracks через CDP
4. ⏳ Запись audio в opus файлы
5. ⏳ XMPP signaling для aiortc версии
6. ⏳ Интеграция с S3 upload
7. ⏳ Интеграция с Prosody webhooks

## Полезные ссылки

- [Playwright Python Docs](https://playwright.dev/python/docs/intro)
- [aiortc Documentation](https://aiortc.readthedocs.io/)
- [Jitsi Meet Handbook](https://jitsi.github.io/handbook/)

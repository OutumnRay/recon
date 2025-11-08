# ⚡ Быстрый старт: Автоматическая запись Jitsi

Полностью автоматическая запись всех участников Jitsi конференций с загрузкой на MinIO.

## 🎯 Что это дает?

- **Автоматически** записывает каждого участника конференции в отдельный файл
- **Автоматически** загружает на MinIO после завершения
- **Не требует** ручного управления
- **Работает** для всех конференций на сервере

## 🚀 Запуск за 3 шага

### 1. Запустите сервисы

```bash
cd /path/to/jitsy
docker-compose up -d kurento kurento-recorder auto-recorder
```

### 2. Проверьте что все работает

```bash
# Проверка Kurento
curl http://localhost:9888/health
# Ожидается: {"status":"healthy","active_recordings":0}

# Проверка Auto Recorder
curl http://localhost:9889/health
# Ожидается: {"status":"healthy",...}
```

### 3. Начните конференцию

Просто зайдите на ваш Jitsi: https://meet.recontext.online/любая-комната

**Запись начнется автоматически!**

## 📊 Как это работает

```
┌─────────────────────────────────────────────────────────┐
│  Пользователь заходит в комнату                          │
└────────────────────┬────────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────────┐
│  Prosody отправляет webhook "participantJoined"          │
└────────────────────┬────────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────────┐
│  Auto Recorder вызывает Kurento API для записи           │
└────────────────────┬────────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────────┐
│  Kurento записывает аудио/видео поток                    │
└────────────────────┬────────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────────┐
│  При выходе - файл загружается на MinIO                 │
└─────────────────────────────────────────────────────────┘
```

## 📁 Где найти записи?

### MinIO Web Interface

1. Откройте: https://api.storage.recontext.online
2. Войдите: `minioadmin` / `minioadmin`
3. Bucket: `jitsi-recordings`
4. Структура:
   ```
   recordings/
   ├── название_комнаты/
   │   └── conference_id_timestamp/
   │       ├── participant1_timestamp.webm
   │       ├── participant2_timestamp.webm
   │       └── metadata.json
   ```

### MinIO CLI (опционально)

```bash
# Установите mc (MinIO Client)
brew install minio/stable/mc  # macOS
# или
wget https://dl.min.io/client/mc/release/linux-amd64/mc && chmod +x mc

# Настройте подключение
mc alias set recontext https://api.storage.recontext.online minioadmin minioadmin

# Список записей
mc ls recontext/jitsi-recordings/recordings/

# Скачать запись
mc cp recontext/jitsi-recordings/recordings/testroom/abc_123/user_20251108.webm ./
```

## 🧪 Тестирование

### Простой тест (2 минуты)

```bash
# 1. Откройте конференцию в браузере
open https://meet.recontext.online/autorectest

# 2. Подождите 30 секунд (чтобы была запись)

# 3. Выйдите из конференции

# 4. Через 10 секунд проверьте MinIO
# Должен появиться файл: jitsi-recordings/recordings/autorectest/.../*.webm
```

### Проверка логов

```bash
# Следить за записью в реальном времени
docker-compose logs -f auto-recorder kurento-recorder

# Что искать в логах:
# ✅ "PARTICIPANT JOINED" - участник присоединился
# ✅ "Recording started" - запись началась
# ✅ "PARTICIPANT LEFT" - участник вышел
# ✅ "Recording stopped" - запись остановлена
# ✅ "Upload completed" - файл загружен на MinIO
```

## ⚙️ Настройка

### Включить/Выключить автозапись

В `docker-compose.yml`:

```yaml
auto-recorder:
  environment:
    - AUTO_RECORD=true  # true = включено, false = выключено
```

После изменения:
```bash
docker-compose up -d auto-recorder
```

### Изменить формат записи

В `docker-compose.yml` для `kurento-recorder`:

```yaml
kurento-recorder:
  environment:
    - RECORD_FORMAT=webm  # webm, mp4, mkv
    - VIDEO_CODEC=vp8     # vp8, vp9, h264
    - AUDIO_CODEC=opus    # opus, aac
```

## 📊 Мониторинг

### Статус сервисов

```bash
docker-compose ps kurento kurento-recorder auto-recorder
```

Все должны быть `healthy`.

### Активные записи

```bash
# Проверить сколько сейчас идет записей
curl -s http://localhost:9889/stats | jq '.'

# Пример ответа:
{
  "active_conferences": [
    {
      "room": "testmeeting",
      "started_at": "2025-11-08T14:30:00",
      "participants": 3
    }
  ],
  "active_sessions": [
    {
      "endpoint_id": "abc123",
      "user": "john@domain/res",
      "room": "testmeeting",
      "display_name": "John Doe",
      "started_at": "2025-11-08T14:30:05"
    }
  ]
}
```

### Проверка MinIO

```bash
# Health check
curl https://api.storage.recontext.online/minio/health/live

# Список файлов через API
curl -X GET "https://api.storage.recontext.online" \
  --user "minioadmin:minioadmin"
```

## 🚨 Troubleshooting

### Запись не начинается

1. **Проверьте webhook настройки:**
   ```bash
   grep RECORDER_WEBHOOK .env
   # Должно быть: RECORDER_WEBHOOK_URL=http://auto-recorder:8080/events
   ```

2. **Перезапустите Prosody:**
   ```bash
   docker-compose restart prosody
   ```

3. **Проверьте логи Prosody:**
   ```bash
   docker-compose logs prosody | grep -i webhook
   ```

### Файлы не появляются на MinIO

1. **Проверьте Kurento Recorder:**
   ```bash
   curl http://localhost:9888/health
   docker-compose logs kurento-recorder | grep -i minio
   ```

2. **Проверьте credentials MinIO в docker-compose.yml:**
   ```yaml
   kurento-recorder:
     environment:
       - MINIO_ENDPOINT=https://api.storage.recontext.online
       - MINIO_ACCESS_KEY=minioadmin
       - MINIO_SECRET_KEY=minioadmin
       - MINIO_BUCKET=jitsi-recordings
   ```

### Сервис не стартует

```bash
# Проверить зависимости
docker-compose ps kurento kurento-recorder redis

# Все должны быть running

# Если что-то не работает:
docker-compose up -d kurento kurento-recorder redis
docker-compose up -d auto-recorder
```

## 📈 Производительность

### Ресурсы на 1 участника

- **CPU**: ~5% (1 ядра)
- **RAM**: ~50MB
- **Disk I/O**: ~128 kbit/s (для Opus audio)
- **Network**: ~128 kbit/s upload to MinIO

### Масштабирование

Система может обрабатывать:
- **10 одновременных участников**: легко
- **50 одновременных участников**: требует 2GB RAM для Kurento
- **100+ участников**: рекомендуется scale Kurento horizontally

## 🔐 Безопасность

### Рекомендации для production

1. **Измените MinIO credentials:**
   ```yaml
   kurento-recorder:
     environment:
       - MINIO_ACCESS_KEY=your_secure_key
       - MINIO_SECRET_KEY=your_secure_secret
   ```

2. **Используйте HTTPS для MinIO:**
   ```yaml
   - MINIO_ENDPOINT=https://your-minio.domain.com
   - MINIO_SECURE=true
   ```

3. **Ограничьте доступ к webhook:**
   - Auto-recorder слушает только внутри Docker network
   - Порт 9889 можно не публиковать если webhook идет внутри сети

## 📚 Дополнительная документация

- **Подробная документация**: [auto-recorder/README.md](./auto-recorder/README.md)
- **Kurento Recorder API**: [KURENTO-RECORDER.md](./KURENTO-RECORDER.md)
- **Справка по портам**: [RECORDER-PORTS.md](./RECORDER-PORTS.md)

## 🎉 Готово!

Теперь все конференции записываются автоматически. Просто заходите в комнату и система сама позаботится о записи!

**Вопросы?** Проверьте логи: `docker-compose logs -f auto-recorder`

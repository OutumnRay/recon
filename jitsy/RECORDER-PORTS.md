# 🔌 Порты сервисов Jitsi Recorder

## 📊 Таблица портов

| Сервис | Внешний порт | Внутренний порт | Описание |
|--------|--------------|-----------------|----------|
| **Kurento Media Server** | 8889 | 8888 | WebSocket API для Kurento |
| **Kurento Recorder API** | 9888 | 8080 | REST API для управления записью (ручной) |
| **Auto Recorder** | 9889 | 8080 | Webhook для автоматической записи |
| **Kurento RTP** | 40000-40100 | 40000-40100 | UDP порты для WebRTC медиа |
| **Jicofo REST** | 127.0.0.1:8888 | 8888 | Jicofo REST API (локальный) |

## 🚀 Быстрый доступ

### Auto Recorder (Автоматическая запись)
```bash
# Health check
curl http://localhost:9889/health

# Статистика активных записей
curl http://localhost:9889/stats

# Все записывается АВТОМАТИЧЕСКИ! Просто зайдите в конференцию
open https://meet.recontext.online/любая-комната
```

### Kurento Recorder API (Ручное управление)
```bash
# Health check
curl http://localhost:9888/health

# Swagger UI
open http://localhost:9888/docs

# Начать запись
curl -X POST "http://localhost:9888/record/start?user=john&room=test"

# Проверить статус
curl "http://localhost:9888/record/status?user=john"

# Список записей
curl "http://localhost:9888/record/list"

# Остановить запись
curl -X POST "http://localhost:9888/record/stop?user=john&room=test"
```

### Kurento WebSocket
```bash
# Проверка доступности
wscat -c ws://localhost:8889/kurento
```

### MinIO Storage
```
URL: https://api.storage.recontext.online
Login: minioadmin
Password: minioadmin
Bucket: jitsi-recordings
```

## 📝 Примечания

- **Порт 8888** занят jicofo, поэтому Kurento использует **8889**
- **Порт 8080** занят другим сервисом, поэтому Recorder API использует **9888**
- Внутри Docker сети сервисы общаются на внутренних портах
- Снаружи доступ через внешние порты (проброшенные через docker-compose)

## 📚 Документация

- **Автоматическая запись** (рекомендуется):
  - [QUICKSTART-AUTO-RECORDER.md](./QUICKSTART-AUTO-RECORDER.md) - быстрый старт
  - [auto-recorder/README.md](./auto-recorder/README.md) - полная документация

- **Ручное управление** (для продвинутых):
  - [QUICKSTART-RECORDER.md](./QUICKSTART-RECORDER.md) - быстрый старт
  - [KURENTO-RECORDER.md](./KURENTO-RECORDER.md) - полная документация
  - [recorder/README.md](./recorder/README.md) - API reference

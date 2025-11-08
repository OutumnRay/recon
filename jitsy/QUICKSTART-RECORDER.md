# 🚀 Quick Start - Kurento Recorder

Быстрый старт для записи индивидуальных потоков Jitsi участников.

## ⚡ Запуск за 3 шага

### 1. Запустить сервисы

```bash
cd /Volumes/ExternalData/source/Team21/Recontext.online/jitsy
docker-compose up -d kurento kurento-recorder
```

### 2. Проверить что все работает

```bash
# Проверить статус сервисов
docker-compose ps kurento kurento-recorder

# Проверить health check
curl http://localhost:9888/health
```

Ожидаемый ответ:
```json
{"status": "healthy", "active_recordings": 0}
```

### 3. Запустить тестовую запись

```bash
# Тест полного цикла
./test-recorder.sh
```

Если все прошло успешно, вы увидите:
```
✅ All tests passed successfully!
```

## 📝 Базовое использование

### Начать запись

```bash
curl -X POST "http://localhost:9888/record/start?user=john_doe&room=testroom"
```

### Проверить статус

```bash
curl "http://localhost:9888/record/status?user=john_doe"
```

### Остановить запись

```bash
curl -X POST "http://localhost:9888/record/stop?user=john_doe&room=testroom"
```

### Просмотреть все активные записи

```bash
curl "http://localhost:9888/record/list"
```

## 🗂️ Где найти записи?

### MinIO Storage

Записи автоматически загружаются в MinIO:

- **URL**: https://api.storage.recontext.online
- **Login**: minioadmin
- **Password**: minioadmin
- **Bucket**: jitsi-recordings

### Структура файлов

```
jitsi-recordings/
├── testroom/
│   ├── john_doe_20251108_153000.webm
│   ├── jane_smith_20251108_153005.webm
│   └── ...
└── anotherroom/
    └── ...
```

## 📖 Подробная документация

- **Полная документация**: [KURENTO-RECORDER.md](./KURENTO-RECORDER.md)
- **API документация**: [recorder/README.md](./recorder/README.md)
- **Swagger UI**: http://localhost:9888/docs

## 🔧 Troubleshooting

### Kurento не запускается

```bash
# Посмотреть логи
docker-compose logs kurento

# Проверить порты (8889 для Kurento, 8888 занят jicofo)
docker-compose ps

# Перезапустить
docker-compose restart kurento
```

**Примечание**: Kurento использует внешний порт 8889 (внутренний 8888), так как порт 8888 уже используется jicofo REST API.

### Recorder API не отвечает

```bash
# Посмотреть логи
docker-compose logs kurento-recorder

# Проверить что Kurento запущен
docker-compose ps kurento

# Перезапустить recorder
docker-compose restart kurento-recorder
```

### Файлы не загружаются в MinIO

```bash
# Проверить доступность MinIO
curl https://api.storage.recontext.online

# Проверить credentials
docker-compose exec kurento-recorder env | grep MINIO

# Посмотреть логи recorder
docker-compose logs kurento-recorder | grep -i minio
```

## 📊 Мониторинг

### Логи в реальном времени

```bash
# Все логи recording stack
docker-compose logs -f kurento kurento-recorder

# Только Kurento
docker-compose logs -f kurento

# Только Recorder API
docker-compose logs -f kurento-recorder
```

### Метрики

```bash
# Количество активных записей
curl http://localhost:9888/record/list | jq '.active_recordings'

# Список всех записей
curl http://localhost:9888/record/list | jq '.recordings'
```

## 🛑 Остановить все

### Остановить все записи

```bash
curl -X DELETE "http://localhost:9888/record/stop-all"
```

### Остановить сервисы

```bash
docker-compose stop kurento kurento-recorder
```

### Полностью удалить (включая volumes)

```bash
docker-compose down -v
```

## 🔗 Полезные команды

```bash
# Статус всех сервисов
docker-compose ps

# Рестарт сервисов
docker-compose restart kurento kurento-recorder

# Пересборка recorder после изменений
docker-compose build kurento-recorder
docker-compose up -d kurento-recorder

# Просмотр использования ресурсов
docker stats jitsi-kurento jitsi-kurento-recorder

# Shell в контейнере recorder
docker-compose exec kurento-recorder bash

# Проверка файлов в /recordings
docker-compose exec kurento-recorder ls -lh /recordings
```

## 🎯 Следующие шаги

1. Прочитать [KURENTO-RECORDER.md](./KURENTO-RECORDER.md) для полного понимания системы
2. Изучить [API документацию](./recorder/README.md) для интеграции
3. Настроить автоматическую запись через webhooks
4. Настроить мониторинг через Prometheus/Grafana
5. Добавить аутентификацию для production использования

## 💡 Примеры использования

### Python скрипт

```python
import requests
import time

api_url = "http://localhost:9888"
user = "john_doe"
room = "meeting123"

# Начать запись
response = requests.post(f"{api_url}/record/start", params={"user": user, "room": room})
print(f"Started: {response.json()}")

# Записывать 60 секунд
time.sleep(60)

# Остановить
response = requests.post(f"{api_url}/record/stop", params={"user": user, "room": room})
print(f"Stopped: {response.json()}")
```

### Bash скрипт

```bash
#!/bin/bash
USER="john_doe"
ROOM="meeting123"

# Start
curl -X POST "http://localhost:9888/record/start?user=$USER&room=$ROOM"

# Wait
sleep 60

# Stop
curl -X POST "http://localhost:9888/record/stop?user=$USER&room=$ROOM"
```

---

**Готово!** Теперь вы можете записывать индивидуальные потоки участников Jitsi конференций 🎉

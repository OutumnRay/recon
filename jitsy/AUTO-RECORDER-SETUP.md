# 🎬 Установка Auto Recorder на сервере

Полная инструкция по установке автоматической записи Jitsi конференций.

## 📋 Что было добавлено

### Новые сервисы

1. **Kurento Media Server** (порт 8889)
   - WebRTC media server для захвата потоков

2. **Kurento Recorder API** (порт 9888)
   - REST API для управления записью
   - Автоматическая загрузка на MinIO

3. **Auto Recorder** (порт 9889)
   - Webhook listener для событий Prosody
   - Автоматическое управление записью

### Новые файлы

```
jitsy/
├── kurento.conf.json                  # Конфигурация Kurento
├── recorder/                          # Kurento Recorder API
│   ├── Dockerfile
│   ├── requirements.txt
│   ├── recorder.py
│   ├── README.md
│   ├── example-client.js
│   └── example.html
├── auto-recorder/                     # Auto Recorder сервис
│   ├── Dockerfile
│   ├── requirements.txt
│   ├── auto_recorder.py
│   └── README.md
├── KURENTO-RECORDER.md               # Документация Kurento
├── QUICKSTART-RECORDER.md            # Быстрый старт (ручной)
├── QUICKSTART-AUTO-RECORDER.md       # Быстрый старт (авто)
├── RECORDER-PORTS.md                 # Справка по портам
├── test-recorder.sh                  # Скрипт тестирования
└── AUTO-RECORDER-SETUP.md            # Этот файл
```

## 🚀 Установка на сервере

### 1. Подключитесь к серверу

```bash
ssh root@your-server-ip
cd ~/jitsi-docker-jitsi-meet-3bc1ebc
```

### 2. Убедитесь что все файлы на месте

```bash
# Проверка структуры
ls -la kurento.conf.json
ls -la recorder/
ls -la auto-recorder/
ls -la .env | grep RECORDER_WEBHOOK

# Должно быть:
# kurento.conf.json - ✅
# recorder/ - ✅
# auto-recorder/ - ✅
# RECORDER_WEBHOOK_URL=http://auto-recorder:8080/events - ✅
```

### 3. Проверьте .env файл

```bash
grep "RECORDER_WEBHOOK_URL" .env
```

Должно быть:
```
RECORDER_WEBHOOK_URL=http://auto-recorder:8080/events
```

Если нет - добавьте:
```bash
echo "RECORDER_WEBHOOK_URL=http://auto-recorder:8080/events" >> .env
```

### 4. Соберите и запустите сервисы

```bash
# Соберите образы
docker-compose build kurento-recorder auto-recorder

# Запустите все сервисы записи
docker-compose up -d kurento kurento-recorder auto-recorder

# Перезапустите Prosody чтобы подхватить webhook
docker-compose restart prosody
```

### 5. Проверьте что все запустилось

```bash
# Статус контейнеров
docker-compose ps kurento kurento-recorder auto-recorder

# Все должны быть "running (healthy)"
```

### 6. Проверьте health checks

```bash
# Kurento Media Server (внутренний, через kurento-recorder)
docker-compose exec kurento-recorder curl -s http://kurento:8888/ | head -5

# Kurento Recorder API
curl -s http://localhost:9888/health

# Auto Recorder
curl -s http://localhost:9889/health
```

Ожидаемые ответы:
```json
// Kurento Recorder API
{"status":"healthy","active_recordings":0}

// Auto Recorder
{
  "status":"healthy",
  "worker_id":"auto-recorder",
  "auto_record":true,
  "active_conferences":0,
  "active_sessions":0,
  "kurento_api":"http://kurento-recorder:8080",
  "kurento_status":{"status":"healthy","active_recordings":0}
}
```

## 🧪 Тестирование

### Тест 1: Автоматическая запись через браузер

1. **Откройте конференцию:**
   ```bash
   # На вашем компьютере
   open https://meet.recontext.online/autotest123
   ```

2. **Подождите 30 секунд** (чтобы была хоть какая-то запись)

3. **Выйдите из конференции**

4. **На сервере проверьте логи:**
   ```bash
   docker-compose logs --tail=50 auto-recorder
   ```

   Должны увидеть:
   ```
   📥 PARTICIPANT JOINED: Your Name ...
   🎙️  Starting recording: ...
   ✅ Recording started: ...
   📤 PARTICIPANT LEFT: ...
   ⏹️  Stopping recording: ...
   ✅ Recording stopped: ...
   🏁 CONFERENCE ENDED: autotest123
   ```

5. **Проверьте что файл появился на MinIO:**
   ```bash
   # Если у вас установлен mc (MinIO Client):
   mc ls recontext/jitsi-recordings/recordings/autotest123/

   # Или откройте в браузере:
   # https://api.storage.recontext.online
   # Login: minioadmin / minioadmin
   # Bucket: jitsi-recordings
   ```

### Тест 2: Webhook напрямую

```bash
# На сервере
curl -X POST http://localhost:9889/events \
  -H "Content-Type: application/json" \
  -d '{
    "eventType": "participantJoined",
    "roomName": "webhook-test",
    "endpointId": "test123",
    "participantId": "testuser@domain/res",
    "participantName": "testuser@domain/res",
    "displayName": "Test User"
  }'

# Подождите 10 секунд

curl -X POST http://localhost:9889/events \
  -H "Content-Type: application/json" \
  -d '{
    "eventType": "participantLeft",
    "roomName": "webhook-test",
    "endpointId": "test123"
  }'

# Проверьте MinIO - должен появиться файл
```

### Тест 3: Полный тестовый скрипт

```bash
# На сервере
chmod +x test-recorder.sh

# Запустите тест (30 секунд записи)
./test-recorder.sh

# Скрипт автоматически:
# - Создаст тестовую запись
# - Подождет 30 секунд
# - Остановит запись
# - Загрузит на MinIO
```

## 📊 Мониторинг

### Просмотр логов

```bash
# Все логи записи
docker-compose logs -f kurento kurento-recorder auto-recorder

# Только Auto Recorder
docker-compose logs -f auto-recorder

# Последние 100 строк
docker-compose logs --tail=100 auto-recorder
```

### Статистика

```bash
# Активные записи
curl -s http://localhost:9889/stats | jq '.'

# Статус Kurento
curl -s http://localhost:9888/health | jq '.'

# Список всех активных записей в Kurento
curl -s http://localhost:9888/record/list | jq '.'
```

### Prometheus метрики

Auto Recorder экспортирует метрики через `/health`:
```bash
curl -s http://localhost:9889/health | jq '.active_conferences, .active_sessions'
```

## 🔧 Настройка

### Отключить автозапись

В `docker-compose.yml`:
```yaml
auto-recorder:
  environment:
    - AUTO_RECORD=false  # ← Изменить на false
```

Затем:
```bash
docker-compose up -d auto-recorder
```

### Изменить формат записи

В `recorder/recorder.py` (строка ~90):
```python
async def create_recorder_endpoint(self, pipeline_id: str, uri: str) -> str:
    result = await self._send_request("create", {
        "type": "RecorderEndpoint",
        "constructorParams": {
            "mediaPipeline": pipeline_id,
            "uri": uri,
            "mediaProfile": "WEBM"  # ← WEBM, WEBM_AUDIO_ONLY, WEBM_VIDEO_ONLY, MP4
        },
        ...
    })
```

### Изменить путь на MinIO

В `docker-compose.yml`:
```yaml
kurento-recorder:
  environment:
    - MINIO_BUCKET=my-custom-bucket  # ← Изменить bucket
```

## 🚨 Troubleshooting

### Сервис не стартует

```bash
# Проверить ошибки
docker-compose logs kurento kurento-recorder auto-recorder | grep -i error

# Проверить зависимости
docker-compose ps redis kurento

# Пересоздать
docker-compose up -d --force-recreate kurento kurento-recorder auto-recorder
```

### Записи не создаются

1. **Проверьте webhook в Prosody:**
   ```bash
   docker-compose logs prosody | grep -i "recorder\|webhook"
   ```

2. **Проверьте что Auto Recorder получает события:**
   ```bash
   docker-compose logs auto-recorder | grep "PARTICIPANT"
   ```

3. **Проверьте что Kurento API доступен:**
   ```bash
   curl http://localhost:9888/health
   docker-compose exec auto-recorder curl http://kurento-recorder:8080/health
   ```

### Файлы не загружаются на MinIO

```bash
# Проверьте MinIO credentials
docker-compose logs kurento-recorder | grep -i minio

# Проверьте доступность MinIO
curl -I https://api.storage.recontext.online

# Проверьте что bucket существует
docker-compose exec kurento-recorder python3 -c "
import boto3
s3 = boto3.client('s3',
    endpoint_url='https://api.storage.recontext.online',
    aws_access_key_id='minioadmin',
    aws_secret_access_key='minioadmin')
print(s3.list_buckets())
"
```

## 📈 Производительность

### Текущая конфигурация

- **Kurento**: 2GB RAM limit
- **Kurento Recorder**: 512MB RAM limit
- **Auto Recorder**: 256MB RAM limit

### Рекомендации для масштабирования

**До 10 участников одновременно:**
- Текущая конфигурация OK

**До 50 участников одновременно:**
```yaml
kurento:
  mem_limit: 4g

kurento-recorder:
  mem_limit: 1g
```

**Более 50 участников:**
- Рассмотрите horizontal scaling (несколько Kurento серверов)
- Используйте Redis для координации

## 🔐 Безопасность Production

### 1. Измените MinIO credentials

В `.env` или `docker-compose.yml`:
```yaml
kurento-recorder:
  environment:
    - MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY}
    - MINIO_SECRET_KEY=${MINIO_SECRET_KEY}
```

### 2. Не публикуйте лишние порты

В `docker-compose.yml` для production:
```yaml
auto-recorder:
  # ports:
  #   - "9889:8080"  # ← Закомментировать если не нужен внешний доступ
```

### 3. Используйте HTTPS для MinIO

```yaml
kurento-recorder:
  environment:
    - MINIO_ENDPOINT=https://minio.your-domain.com
    - MINIO_SECURE=true
```

## ✅ Checklist перед production

- [ ] MinIO credentials изменены
- [ ] Auto Recorder работает и видит Kurento API
- [ ] Prosody отправляет webhook события
- [ ] Тестовая запись успешно загружена на MinIO
- [ ] Логи не показывают ошибок
- [ ] Health checks проходят
- [ ] Достаточно места на диске для временных файлов
- [ ] Настроен мониторинг (Prometheus/Grafana)

## 📚 Документация

- [QUICKSTART-AUTO-RECORDER.md](./QUICKSTART-AUTO-RECORDER.md) - Быстрый старт
- [auto-recorder/README.md](./auto-recorder/README.md) - Подробная документация Auto Recorder
- [KURENTO-RECORDER.md](./KURENTO-RECORDER.md) - Документация Kurento
- [RECORDER-PORTS.md](./RECORDER-PORTS.md) - Справка по портам

## 🎉 Готово!

Теперь все конференции на вашем Jitsi сервере записываются автоматически!

**Проверить работу:**
1. Зайдите в любую комнату: https://meet.recontext.online/тест
2. Через 30 секунд выйдите
3. Проверьте MinIO: https://api.storage.recontext.online

**Вопросы?**
- Логи: `docker-compose logs -f auto-recorder`
- Статус: `curl http://localhost:9889/stats`
- Health: `curl http://localhost:9889/health`

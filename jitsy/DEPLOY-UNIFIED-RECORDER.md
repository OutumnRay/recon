# 🚀 Развертывание Unified Recorder на сервере

## 📦 Что было изменено

### Объединено 2 сервиса → 1

**Было:**
- `kurento-recorder` (порт 9888) - REST API для Kurento
- `auto-recorder` (порт 9889) - Webhook listener

**Стало:**
- `recorder` (порт 9888) - **Unified Recorder** (всё в одном)

### Новые файлы

```
jitsy/
├── unified-recorder/
│   ├── Dockerfile
│   ├── requirements.txt
│   ├── unified_recorder.py     ← Объединенный код
│   └── README.md
├── docker-compose.yml          ← Обновлен
└── .env                        ← RECORDER_WEBHOOK_URL обновлен
```

## 🎯 Команды для развертывания

### 1. Подключитесь к серверу

```bash
ssh root@your-server-ip
cd ~/jitsi-docker-jitsi-meet-3bc1ebc
```

### 2. Остановите старые сервисы

```bash
# Остановите и удалите старые контейнеры
docker-compose stop kurento-recorder auto-recorder
docker-compose rm -f kurento-recorder auto-recorder
```

### 3. Проверьте конфигурацию

```bash
# Проверьте webhook URL
grep RECORDER_WEBHOOK .env
# Должно быть: RECORDER_WEBHOOK_URL=http://recorder:8080/events

# Если нет - обновите:
sed -i 's|RECORDER_WEBHOOK_URL=.*|RECORDER_WEBHOOK_URL=http://recorder:8080/events|' .env
```

### 4. Соберите и запустите

```bash
# Соберите unified-recorder
docker-compose build recorder

# Запустите Kurento + Recorder
docker-compose up -d kurento recorder

# Перезапустите Prosody для подключения webhook
docker-compose restart prosody

# Подождите 10 секунд
sleep 10
```

### 5. Проверьте статус

```bash
# Статус контейнеров
docker-compose ps kurento recorder

# Оба должны быть: running (healthy)
```

### 6. Health check

```bash
# Проверка recorder
curl -s http://localhost:9888/health | jq '.'

# Должен вернуть:
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

### 7. Проверьте логи

```bash
# Следите за логами
docker-compose logs -f recorder

# Должны увидеть:
# 🎬 Jitsi Unified Recorder Configuration
# ✅ MinIO bucket 'jitsi-recordings' exists
# 🌐 HTTP Server started on :8080
# ✅ Unified Recorder ready
```

## 🧪 Тестирование

### Тест 1: Webhook работает

```bash
# Зайдите в конференцию через браузер
open https://meet.recontext.online/test123

# Проверьте логи recorder
docker-compose logs --tail=50 recorder

# Должны увидеть:
# 📥 PARTICIPANT JOINED: ...
# 🎙️ Starting recording: ...
# ✅ Recording started: ...
# Pipeline: xxx
# RTP Endpoint: yyy
# Recorder: zzz
```

### Тест 2: Prosody отправляет события

```bash
# Проверьте логи Prosody
docker-compose logs --tail=50 prosody | grep recorder

# Должно быть:
# Webhook URL: http://recorder:8080/events
# ✅ Event sent: participantJoined ...
```

### Тест 3: Kurento pipeline создается

```bash
# Статистика recorder
curl -s http://localhost:9888/stats | jq '.'

# Должен показать активные сессии:
{
  "active_conferences": [...],
  "active_sessions": [
    {
      "participant_id": "...",
      "display_name": "...",
      "room": "test123",
      "duration": 15.5,
      "filename": "test123_..._20251108.webm"
    }
  ]
}
```

## ⚠️ ВАЖНО: Известные ограничения

### Проблема: Нет реального медиа от JVB

**Симптом:**
- ✅ Webhook работает
- ✅ Kurento pipeline создается
- ❌ Файлы пустые (0 байт)
- ❌ Не загружаются на MinIO

**Причина:**
JVB не настроен для forwarding медиа-потоков в Kurento RtpEndpoint.

**Что происходит:**
1. Participant присоединяется → webhook срабатывает ✅
2. Recorder создает Kurento pipeline ✅
3. Kurento создает RtpEndpoint ✅
4. Kurento ждет RTP пакеты... ❌ (не приходят)
5. При выходе - файл пустой ❌

### Что нужно для полной работы

**Вариант 1: Настройка JVB forwarding (сложно)**

Нужно:
1. Получить SDP offer от Kurento RtpEndpoint
2. Создать channel в JVB через Colibri REST API
3. Передать SDP offer в JVB
4. Получить SDP answer от JVB
5. Применить к Kurento
6. Обменяться ICE candidates

**Вариант 2: Jibri-style participant (проще)**

Создать "виртуального участника" который:
1. Присоединяется к конференции как обычный клиент
2. Получает медиа через WebRTC
3. Передает в Kurento для записи

Это требует полноценного WebRTC клиента (как Jibri).

**Вариант 3: Использовать готовый Jibri ⭐ Рекомендую**

Jibri - официальный recorder от Jitsi:
- Стабильный и протестированный
- Записывает композитное видео (все участники)
- Автоматически загружает на S3
- Поддерживается командой Jitsi

## 🔄 Откат на старую версию

Если нужно вернуться:

```bash
# Остановите unified recorder
docker-compose stop recorder
docker-compose rm -f recorder

# Запустите старые сервисы
docker-compose up -d kurento-recorder auto-recorder

# Обновите webhook
sed -i 's|RECORDER_WEBHOOK_URL=http://recorder:8080/events|RECORDER_WEBHOOK_URL=http://auto-recorder:8080/events|' .env

# Перезапустите Prosody
docker-compose restart prosody
```

## 📊 Мониторинг

### Проверка статуса

```bash
# Каждую минуту проверять
watch -n 60 'curl -s http://localhost:9888/health | jq .'
```

### Логи в реальном времени

```bash
# Все компоненты записи
docker-compose logs -f prosody kurento recorder

# Только recorder
docker-compose logs -f recorder

# С фильтрацией
docker-compose logs -f recorder | grep -E "(JOINED|LEFT|ERROR|Recording)"
```

### Disk space

```bash
# Проверить место в /recordings
df -h | grep recordings

# Список файлов
ls -lh recordings/
```

### MinIO проверка

```bash
# Web console
open https://api.storage.recontext.online

# Login: minioadmin / minioadmin
# Bucket: jitsi-recordings
```

## 🚨 Troubleshooting

### Recorder не стартует

```bash
# Проверьте логи
docker-compose logs recorder | grep -i error

# Проверьте зависимости
docker-compose ps kurento redis

# Пересоберите
docker-compose build --no-cache recorder
docker-compose up -d recorder
```

### Webhook не работает

```bash
# Проверьте URL
grep RECORDER_WEBHOOK .env

# Проверьте что Prosody видит recorder
docker-compose exec prosody ping -c 1 recorder

# Проверьте логи Prosody
docker-compose logs prosody | grep webhook
```

### Kurento недоступен

```bash
# Проверьте Kurento
docker-compose logs kurento | grep -i error

# Проверьте WebSocket
docker-compose exec recorder curl -I http://kurento:8888/

# Перезапустите
docker-compose restart kurento
sleep 5
docker-compose restart recorder
```

### Файлы не загружаются на MinIO

```bash
# Проверьте credentials
docker-compose exec recorder python3 -c "
import boto3
s3 = boto3.client('s3',
    endpoint_url='https://api.storage.recontext.online',
    aws_access_key_id='minioadmin',
    aws_secret_access_key='minioadmin')
print(s3.list_buckets())
"

# Проверьте bucket
curl -I https://api.storage.recontext.online/minio/health/live
```

## ✅ Checklist

- [ ] Старые сервисы остановлены и удалены
- [ ] unified-recorder собран
- [ ] RECORDER_WEBHOOK_URL обновлен в .env
- [ ] Prosody перезапущен
- [ ] Health check проходит
- [ ] Webhook события приходят
- [ ] Kurento pipeline создается
- [ ] Логи без ошибок

## 📚 Следующие шаги

1. **Протестируйте текущую версию** - webhook и pipeline создание
2. **Выберите подход** для получения реального медиа:
   - Настроить JVB forwarding (сложно, но гибко)
   - Использовать Jibri (просто, стабильно)
3. **Доработайте интеграцию** согласно выбранному подходу

## 🏆 Результат

Сейчас у вас:
- ✅ Единый сервис вместо двух
- ✅ Webhook от Prosody работает
- ✅ Kurento pipeline создается автоматически
- ⏳ Ждет интеграции с JVB для реального медиа

**Вопросы?**
- Логи: `docker-compose logs -f recorder`
- Stats: `curl http://localhost:9888/stats`
- Health: `curl http://localhost:9888/health`

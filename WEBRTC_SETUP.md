# WebRTC Setup для Recontext

## Проблема: `navigator.mediaDevices.getUserMedia` undefined

Эта ошибка возникает когда портал работает через **HTTP** вместо **HTTPS**.

### Почему WebRTC требует HTTPS?

WebRTC требует **безопасное соединение (HTTPS)** для доступа к камере и микрофону по соображениям безопасности. Это требование браузеров для защиты пользователей.

**Исключение:** `localhost` и `127.0.0.1` работают через HTTP для разработки.

## ✅ Решение для Production

### Вариант 1: Использовать HTTPS (Рекомендуется)

Вам нужно настроить **reverse proxy** с SSL сертификатом.

#### Пример с Nginx:

```nginx
server {
    listen 443 ssl http2;
    server_name portal.recontext.online;

    ssl_certificate /etc/letsencrypt/live/portal.recontext.online/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/portal.recontext.online/privkey.pem;

    # User Portal
    location / {
        proxy_pass http://localhost:20081;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name portal.recontext.online;
    return 301 https://$server_name$request_uri;
}
```

#### Получить бесплатный SSL сертификат (Let's Encrypt):

```bash
# Установить certbot
sudo apt-get update
sudo apt-get install certbot python3-certbot-nginx

# Получить сертификат
sudo certbot --nginx -d portal.recontext.online

# Автоматическое обновление
sudo certbot renew --dry-run
```

### Вариант 2: Использовать Cloudflare (Бесплатно)

1. Добавьте ваш домен в Cloudflare
2. Cloudflare автоматически предоставит SSL
3. Направьте DNS на ваш сервер
4. Cloudflare будет обрабатывать HTTPS

### Вариант 3: Для разработки через туннель

Используйте **ngrok** или **localtunnel** для создания HTTPS туннеля:

```bash
# Установить ngrok
# https://ngrok.com/download

# Создать туннель
ngrok http 20081

# Вы получите HTTPS URL типа: https://abc123.ngrok.io
```

## 🔧 Docker Compose с Nginx и SSL

Добавьте nginx сервис в `docker-compose.yml`:

```yaml
services:
  nginx:
    image: nginx:alpine
    container_name: recontext-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
      - /etc/letsencrypt:/etc/letsencrypt:ro
    depends_on:
      - user-portal
      - managing-portal
    networks:
      - recontext-network
    restart: unless-stopped
```

## 📱 Мобильные устройства

На мобильных устройствах HTTPS **обязателен**. HTTP не будет работать даже на localhost.

### Проверка:

1. Откройте браузер на мобильном
2. Зайдите в DevTools (Chrome Remote Debugging для Android, Safari Web Inspector для iOS)
3. Проверьте консоль на наличие ошибок
4. Проверьте что URL начинается с `https://`

## 🎯 Текущее состояние

### ✅ Что работает:
- Автоматическое включение камеры и микрофона
- Детальная диагностика проблем
- Адаптивный дизайн для мобильных
- Обработка разрешений на iOS
- Информативные сообщения об ошибках

### ⚠️ Требуется для production:
- **HTTPS сертификат** (Let's Encrypt бесплатно)
- **Nginx reverse proxy** с SSL
- **Правильная DNS настройка**

## 🔍 Диагностика

Откройте консоль браузера (F12) и проверьте логи:

```
[Media] Navigator available: object
[Media] Navigator.mediaDevices available: undefined  ← ПРОБЛЕМА!
[Media] Protocol: http:                              ← Нужен https:
```

Если `navigator.mediaDevices` это `undefined`, значит нужен HTTPS.

## 📚 Дополнительная информация

- [WebRTC Security](https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices/getUserMedia#security)
- [Let's Encrypt](https://letsencrypt.org/)
- [Cloudflare SSL](https://www.cloudflare.com/ssl/)
- [LiveKit Deployment](https://docs.livekit.io/realtime/deploy/)

## 🆘 Быстрое решение для тестирования

Если вам нужно **срочно протестировать** без настройки HTTPS:

1. Используйте **Chrome флаг** (только для разработки!):
   - Откройте `chrome://flags/#unsafely-treat-insecure-origin-as-secure`
   - Добавьте ваш HTTP URL (например: `http://192.168.1.100:20081`)
   - Перезапустите Chrome
   - ⚠️ **Не используйте в production!**

2. Или используйте **localhost туннель**:
   ```bash
   ssh -R 80:localhost:20081 serveo.net
   ```

---

**Важно:** Для production всегда используйте правильный HTTPS с валидным сертификатом!

# Quick Installation Guide

Быстрая установка LiveKit Video Conference на сервер `meet.recontext.online`

## 1. Подготовка на локальном компьютере

```bash
cd /Volumes/ExternalData/source/Team21/Recontext.online/livekit/test/vite-prj

# Создайте production build
npm run build

# Создайте архив для деплоя
tar -czf livekit-deploy.tar.gz \
    dist/ \
    server.js \
    package.json \
    nginx.conf \
    deploy.sh \
    .env.example
```

## 2. Копирование на сервер

```bash
# Скопируйте архив на сервер
scp livekit-deploy.tar.gz root@meet.recontext.online:/tmp/

# Подключитесь к серверу
ssh root@meet.recontext.online
```

## 3. Установка на сервере

```bash
# Создайте директорию
mkdir -p /var/www/html/livekit
cd /var/www/html/livekit

# Распакуйте архив
tar -xzf /tmp/livekit-deploy.tar.gz

# Создайте .env файл
cat > .env << 'EOF'
LIVEKIT_API_KEY=APIBj3yrXtyPRNq
LIVEKIT_API_SECRET=2Q66dFk7HWpxTuTneMT4fQlsxeIlmkn47ApjnJiSukiA
LIVEKIT_URL=wss://video.recontext.online
PORT=3000
EOF

# Установите права
chmod 600 .env
chmod 755 dist/
```

## 4. Установите зависимости

```bash
cd /var/www/html/livekit
npm install --production
```

## 5. Настройте PM2 для backend

```bash
# Установите PM2 (если не установлен)
npm install -g pm2

# Запустите backend
pm2 start server.js --name livekit-backend

# Сохраните конфигурацию
pm2 save

# Настройте автозапуск
pm2 startup
# Выполните команду, которую выведет PM2
```

## 6. Настройте Nginx

```bash
# Скопируйте конфигурацию
cp /var/www/html/livekit/nginx.conf /etc/nginx/sites-available/livekit-meet

# Создайте symlink
ln -s /etc/nginx/sites-available/livekit-meet /etc/nginx/sites-enabled/

# Проверьте конфигурацию
nginx -t

# Если OK, перезапустите nginx
systemctl restart nginx
```

## 7. Проверка

```bash
# Проверьте статус backend
pm2 status

# Проверьте логи
pm2 logs livekit-backend

# Проверьте nginx
systemctl status nginx

# Проверьте порты
netstat -tulpn | grep :3000
netstat -tulpn | grep :443

# Протестируйте API
curl https://meet.recontext.online/getToken?room=test&name=user

# Откройте в браузере
# https://meet.recontext.online
```

## 8. Настройка firewall (если нужно)

```bash
# Разрешите HTTPS
ufw allow 443/tcp

# Разрешите HTTP (для редиректа)
ufw allow 80/tcp

# Проверьте статус
ufw status
```

## Обновление приложения

```bash
# На локальном компьютере
npm run build
tar -czf livekit-update.tar.gz dist/
scp livekit-update.tar.gz root@meet.recontext.online:/tmp/

# На сервере
cd /var/www/html/livekit
tar -xzf /tmp/livekit-update.tar.gz
pm2 restart livekit-backend
systemctl reload nginx
```

## Проверка работы

1. Откройте: https://meet.recontext.online
2. Введите имя и комнату
3. Нажмите "Join Room"
4. Разрешите доступ к камере/микрофону
5. Проверьте что видео работает

## Логи

```bash
# Backend логи
pm2 logs livekit-backend

# Nginx access логи
tail -f /var/log/nginx/livekit-access.log

# Nginx error логи
tail -f /var/log/nginx/livekit-error.log
```

## Troubleshooting

### Backend не запускается

```bash
# Проверьте логи PM2
pm2 logs livekit-backend --err

# Проверьте .env файл
cat /var/www/html/livekit/.env

# Проверьте порт 3000
lsof -i :3000

# Перезапустите
pm2 restart livekit-backend
```

### Nginx ошибки

```bash
# Проверьте конфигурацию
nginx -t

# Проверьте логи
tail -100 /var/log/nginx/livekit-error.log

# Проверьте пути
ls -la /var/www/html/livekit/dist/
ls -la /etc/letsencrypt/live/meet.recontext.online/
```

### SSL ошибки

```bash
# Проверьте сертификаты
openssl s_client -connect meet.recontext.online:443 -servername meet.recontext.online

# Обновите сертификат
certbot renew

# Проверьте права
ls -la /etc/letsencrypt/live/meet.recontext.online/
```

## Мониторинг

```bash
# CPU и память
htop

# PM2 мониторинг
pm2 monit

# Дисковое пространство
df -h

# Статус всех сервисов
systemctl status nginx
pm2 status
```

## Контакты

- Домен: https://meet.recontext.online
- Backend API: https://meet.recontext.online/getToken
- LiveKit Server: wss://video.recontext.online

## Готово!

Приложение должно быть доступно по адресу: **https://meet.recontext.online**

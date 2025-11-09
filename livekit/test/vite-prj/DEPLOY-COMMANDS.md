# Команды для деплоя на meet.recontext.online

## На локальном компьютере

```bash
# 1. Перейдите в директорию проекта
cd /Volumes/ExternalData/source/Team21/Recontext.online/livekit/test/vite-prj

# 2. Создайте production build (уже создан)
npm run build

# 3. Создайте deployment archive (уже создан)
# Файл: livekit-deploy.tar.gz (178 KB)

# 4. Скопируйте на сервер
scp livekit-deploy.tar.gz root@meet.recontext.online:/tmp/
```

## На сервере meet.recontext.online

```bash
# 1. Подключитесь к серверу
ssh root@meet.recontext.online

# 2. Создайте директорию
mkdir -p /var/www/html/livekit
cd /var/www/html/livekit

# 3. Распакуйте архив
tar -xzf /tmp/livekit-deploy.tar.gz

# 4. Создайте .env файл
cat > .env << 'EOF'
LIVEKIT_API_KEY=APIBj3yrXtyPRNq
LIVEKIT_API_SECRET=2Q66dFk7HWpxTuTneMT4fQlsxeIlmkn47ApjnJiSukiA
LIVEKIT_URL=wss://video.recontext.online
PORT=3000
EOF

chmod 600 .env

# 5. Установите зависимости
npm install --production

# 6. Запустите backend с PM2
pm2 start server.js --name livekit-backend
pm2 save

# 7. Настройте Nginx
cp nginx.conf /etc/nginx/sites-available/livekit-meet
ln -s /etc/nginx/sites-available/livekit-meet /etc/nginx/sites-enabled/
nginx -t
systemctl restart nginx

# 8. Проверьте
pm2 status
curl https://meet.recontext.online/health
curl https://meet.recontext.online/getToken?room=test&name=user

# 9. Откройте в браузере
# https://meet.recontext.online
```

## Важные пути

- Frontend (статика): `/var/www/html/livekit/dist/`
- Backend (Node.js): `/var/www/html/livekit/server.js`
- Nginx config: `/etc/nginx/sites-available/livekit-meet`
- SSL Certs: `/etc/letsencrypt/live/meet.recontext.online/`
- Логи Nginx: `/var/log/nginx/livekit-*.log`
- Логи PM2: `pm2 logs livekit-backend`

## Быстрое обновление

```bash
# На локальном
npm run build
tar -czf update.tar.gz dist/
scp update.tar.gz root@meet.recontext.online:/tmp/

# На сервере
cd /var/www/html/livekit
tar -xzf /tmp/update.tar.gz
systemctl reload nginx
```

## Проверка работы

1. **Health check**: https://meet.recontext.online/health
2. **Token API**: https://meet.recontext.online/getToken?room=test&name=user
3. **Frontend**: https://meet.recontext.online
4. **Backend status**: `pm2 status`
5. **Nginx status**: `systemctl status nginx`

## Мониторинг

```bash
# Логи backend
pm2 logs livekit-backend

# Логи Nginx
tail -f /var/log/nginx/livekit-access.log
tail -f /var/log/nginx/livekit-error.log

# Статус процессов
pm2 monit

# Системные ресурсы
htop
```

## Troubleshooting

### Backend не работает
```bash
pm2 logs livekit-backend --err
lsof -i :3000
pm2 restart livekit-backend
```

### Nginx ошибки
```bash
nginx -t
tail -100 /var/log/nginx/livekit-error.log
systemctl restart nginx
```

### CORS ошибки
Убедитесь что в nginx.conf:
```nginx
add_header Access-Control-Allow-Origin "https://meet.recontext.online" always;
```

## URLs

- **Production**: https://meet.recontext.online
- **API**: https://meet.recontext.online/getToken
- **LiveKit Server**: wss://video.recontext.online
- **Health**: https://meet.recontext.online/health

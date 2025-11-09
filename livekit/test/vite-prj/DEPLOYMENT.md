# Production Deployment Guide

Полное руководство по деплою LiveKit Video Conference приложения на production сервер.

## Требования к серверу

- Ubuntu 20.04+ / Debian 11+ / CentOS 8+
- Node.js 20.19+ или 22.12+
- Nginx
- PM2 (для управления процессами)
- Минимум 2GB RAM
- SSL сертификат (опционально, но рекомендуется)

## Подготовка сервера

### 1. Обновите систему

```bash
sudo apt update && sudo apt upgrade -y
```

### 2. Установите Node.js

```bash
# Установите Node.js 20.x
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Проверьте версию
node --version
npm --version
```

### 3. Установите Nginx

```bash
sudo apt install nginx -y
sudo systemctl start nginx
sudo systemctl enable nginx
```

### 4. Установите PM2

```bash
sudo npm install -g pm2
```

## Деплой приложения

### Шаг 1: Клонируйте проект на сервер

```bash
# Создайте директорию для проекта
sudo mkdir -p /var/www/html/livekit
sudo chown -R $USER:$USER /var/www/html/livekit

# Перейдите в директорию
cd /var/www/html/livekit
```

### Шаг 2: Скопируйте файлы

Скопируйте следующие файлы/папки на сервер:

```bash
# На вашем локальном компьютере создайте архив
cd /path/to/vite-prj
tar -czf livekit-app.tar.gz dist/ server.js package.json

# Скопируйте на сервер (замените user и server-ip)
scp livekit-app.tar.gz user@server-ip:/var/www/html/livekit/

# На сервере распакуйте
cd /var/www/html/livekit
tar -xzf livekit-app.tar.gz
```

**Или используйте Git:**

```bash
# На сервере
cd /var/www/html/livekit
git clone <your-repo-url> .
npm run build
```

### Шаг 3: Создайте .env файл

```bash
cd /var/www/html/livekit
nano .env
```

Добавьте:

```env
LIVEKIT_API_KEY=APIBj3yrXtyPRNq
LIVEKIT_API_SECRET=2Q66dFk7HWpxTuTneMT4fQlsxeIlmkn47ApjnJiSukiA
LIVEKIT_URL=wss://video.recontext.online
PORT=3000
```

Сохраните (Ctrl+O, Enter, Ctrl+X)

### Шаг 4: Установите зависимости

```bash
cd /var/www/html/livekit

# Установите только production зависимости
npm install --production
```

### Шаг 5: Настройте PM2 для backend

```bash
# Запустите backend сервер
pm2 start server.js --name livekit-backend

# Сохраните конфигурацию PM2
pm2 save

# Настройте автозапуск PM2 при перезагрузке
pm2 startup
# Выполните команду, которую выведет PM2
```

Проверьте статус:

```bash
pm2 status
pm2 logs livekit-backend
```

### Шаг 6: Настройте Nginx

```bash
# Скопируйте конфигурацию
sudo cp nginx.conf /etc/nginx/sites-available/livekit

# Отредактируйте конфигурацию
sudo nano /etc/nginx/sites-available/livekit
```

**Измените следующие параметры:**

1. `server_name` - замените на ваш домен:
   ```nginx
   server_name video.recontext.online;
   ```

2. `root` - путь к статическим файлам:
   ```nginx
   root /var/www/html/livekit/dist;
   ```

Сохраните файл.

```bash
# Создайте symlink
sudo ln -s /etc/nginx/sites-available/livekit /etc/nginx/sites-enabled/

# Удалите default конфигурацию (опционально)
sudo rm /etc/nginx/sites-enabled/default

# Проверьте конфигурацию
sudo nginx -t

# Перезапустите Nginx
sudo systemctl restart nginx
```

### Шаг 7: Настройте Firewall

```bash
# Разрешите HTTP и HTTPS
sudo ufw allow 'Nginx Full'
sudo ufw allow OpenSSH
sudo ufw enable
sudo ufw status
```

## Настройка SSL (HTTPS) - Рекомендуется

### Используйте Let's Encrypt (бесплатный SSL)

```bash
# Установите Certbot
sudo apt install certbot python3-certbot-nginx -y

# Получите SSL сертификат
sudo certbot --nginx -d video.recontext.online

# Certbot автоматически настроит SSL в nginx
```

Certbot автоматически добавит SSL конфигурацию и настроит редирект с HTTP на HTTPS.

### Автообновление сертификата

```bash
# Проверьте автообновление
sudo certbot renew --dry-run

# Сертификат будет обновляться автоматически
```

## Обновление приложения

### Вариант 1: Ручное обновление

```bash
cd /var/www/html/livekit

# Остановите backend
pm2 stop livekit-backend

# Обновите код
git pull  # или скопируйте новые файлы

# Пересоберите проект
npm run build

# Обновите зависимости (если нужно)
npm install --production

# Запустите backend
pm2 restart livekit-backend

# Очистите кеш nginx (опционально)
sudo systemctl reload nginx
```

### Вариант 2: Скрипт автоматического деплоя

Создайте файл `deploy.sh`:

```bash
nano deploy.sh
```

Добавьте:

```bash
#!/bin/bash

# LiveKit App Deployment Script

set -e

echo "🚀 Starting deployment..."

# Navigate to project directory
cd /var/www/html/livekit

# Pull latest changes
echo "📦 Pulling latest changes..."
git pull

# Install dependencies
echo "📚 Installing dependencies..."
npm install --production

# Build project
echo "🔨 Building project..."
npm run build

# Restart backend
echo "🔄 Restarting backend..."
pm2 restart livekit-backend

# Reload Nginx
echo "🌐 Reloading Nginx..."
sudo systemctl reload nginx

echo "✅ Deployment completed successfully!"
pm2 status
```

Сделайте скрипт исполняемым:

```bash
chmod +x deploy.sh
```

Используйте:

```bash
./deploy.sh
```

## Мониторинг и логи

### PM2 мониторинг

```bash
# Статус процессов
pm2 status

# Просмотр логов
pm2 logs livekit-backend

# Просмотр логов в реальном времени
pm2 logs livekit-backend --lines 100

# Мониторинг ресурсов
pm2 monit
```

### Nginx логи

```bash
# Access логи
sudo tail -f /var/log/nginx/livekit-access.log

# Error логи
sudo tail -f /var/log/nginx/livekit-error.log
```

### Системные ресурсы

```bash
# CPU и память
htop

# Дисковое пространство
df -h

# Проверка портов
sudo netstat -tulpn | grep :3000
sudo netstat -tulpn | grep :80
```

## Резервное копирование

### Создайте скрипт бэкапа

```bash
nano backup.sh
```

```bash
#!/bin/bash

BACKUP_DIR="/backup/livekit"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# Бэкап файлов
tar -czf $BACKUP_DIR/livekit-$DATE.tar.gz /var/www/html/livekit

# Удаление старых бэкапов (старше 30 дней)
find $BACKUP_DIR -name "livekit-*.tar.gz" -mtime +30 -delete

echo "Backup created: livekit-$DATE.tar.gz"
```

```bash
chmod +x backup.sh

# Добавьте в cron для ежедневного бэкапа
crontab -e
```

Добавьте строку:

```
0 2 * * * /var/www/html/livekit/backup.sh
```

## Troubleshooting

### Backend не запускается

```bash
# Проверьте логи
pm2 logs livekit-backend --err

# Проверьте порт
sudo lsof -i :3000

# Перезапустите
pm2 restart livekit-backend
```

### Nginx ошибки

```bash
# Проверьте конфигурацию
sudo nginx -t

# Проверьте логи
sudo tail -100 /var/log/nginx/error.log

# Проверьте права доступа
ls -la /var/www/html/livekit/dist
```

### SSL проблемы

```bash
# Проверьте сертификат
sudo certbot certificates

# Обновите сертификат
sudo certbot renew
```

### High CPU/Memory usage

```bash
# Проверьте процессы
pm2 monit

# Перезапустите PM2
pm2 restart all

# Увеличьте ресурсы сервера если нужно
```

## Оптимизация производительности

### 1. Nginx кеширование

Добавьте в `/etc/nginx/nginx.conf`:

```nginx
http {
    # Cache settings
    proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=my_cache:10m max_size=1g inactive=60m;
    proxy_cache_key "$scheme$request_method$host$request_uri";
}
```

### 2. PM2 Cluster Mode

```bash
# Используйте все CPU ядра
pm2 delete livekit-backend
pm2 start server.js -i max --name livekit-backend
pm2 save
```

### 3. Gzip сжатие

Уже настроено в `nginx.conf`

### 4. HTTP/2

Автоматически включается с SSL/HTTPS

## Безопасность

### 1. Защита .env файла

```bash
chmod 600 /var/www/html/livekit/.env
```

### 2. Обновление зависимостей

```bash
npm audit
npm audit fix
```

### 3. Fail2Ban для защиты от DDoS

```bash
sudo apt install fail2ban -y
sudo systemctl enable fail2ban
sudo systemctl start fail2ban
```

### 4. Регулярные обновления

```bash
sudo apt update && sudo apt upgrade -y
```

## Контакты и поддержка

- LiveKit Docs: https://docs.livekit.io/
- Nginx Docs: https://nginx.org/en/docs/
- PM2 Docs: https://pm2.keymetrics.io/docs/

## Чеклист деплоя

- [ ] Сервер подготовлен (Node.js, Nginx, PM2)
- [ ] Проект скопирован на сервер
- [ ] `.env` файл создан с правильными credentials
- [ ] Зависимости установлены
- [ ] Build создан (`npm run build`)
- [ ] Backend запущен через PM2
- [ ] Nginx настроен и перезапущен
- [ ] Firewall настроен
- [ ] SSL сертификат получен и настроен
- [ ] Приложение доступно по домену
- [ ] Мониторинг настроен
- [ ] Бэкапы настроены
- [ ] Документация обновлена

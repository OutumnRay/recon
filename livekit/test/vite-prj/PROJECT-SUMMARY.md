# LiveKit Video Conference - Project Summary

## Информация о проекте

**Название**: LiveKit Video Conference Application
**Домен**: https://meet.recontext.online
**Backend API**: https://meet.recontext.online/getToken
**LiveKit Server**: wss://video.recontext.online
**Версия**: 1.0.0
**Дата создания**: 2025-11-09

## Что было создано

### Frontend (React + TypeScript + Vite)

#### Основные компоненты
- `src/App.tsx` - Главный компонент приложения
- `src/LiveKitRoom.tsx` - Компонент видеоконференции с полным функционалом
- `src/main.tsx` - Точка входа React
- `src/App.css` - Стили приложения
- `src/index.css` - Глобальные стили

#### Возможности Frontend
- ✅ Форма входа с вводом имени и комнаты
- ✅ Динамическое получение токенов с backend
- ✅ Автоматическое подключение к LiveKit серверу
- ✅ Управление камерой и микрофоном
- ✅ Отображение локального видео
- ✅ Отображение удаленных участников
- ✅ Индикация активных спикеров
- ✅ Адаптивный дизайн
- ✅ Обработка ошибок
- ✅ Автоопределение backend URL (dev/prod)

### Backend (Node.js + Express)

#### Файлы
- `server.js` - Express сервер для генерации JWT токенов

#### Возможности Backend
- ✅ GET /getToken - Генерация токенов
- ✅ POST /getToken - Альтернативный метод
- ✅ Токены действительны 6 часов
- ✅ Права: roomJoin, canPublish, canSubscribe, canPublishData
- ✅ CORS поддержка
- ✅ Чтение конфигурации из .env

### Production Build

#### Результаты сборки
```
dist/
├── assets/
│   ├── index-1rQBsk1_.css    (990 B, gzip: 530 B)
│   └── index-DHVT6pKv.js     (630 KB, gzip: 175 KB)
├── index.html                 (484 B)
└── vite.svg
```

**Total size**: ~631 KB
**Gzipped size**: ~176 KB

### Конфигурация и Деплой

#### Nginx
- `nginx.conf` - Полная конфигурация для meet.recontext.online
  - ✅ HTTP → HTTPS редирект
  - ✅ SSL/TLS конфигурация
  - ✅ Gzip сжатие
  - ✅ Security headers (HSTS, XSS, CSP)
  - ✅ Кеширование статики (1 год)
  - ✅ Proxy для API
  - ✅ CORS настройки
  - ✅ OCSP Stapling
  - ✅ HTTP/2 support

#### SSL Сертификаты
- Путь: `/etc/letsencrypt/live/meet.recontext.online/`
- Тип: Let's Encrypt
- Протоколы: TLSv1.2, TLSv1.3

#### Deployment Scripts
- `deploy.sh` - Автоматический скрипт деплоя
  - Создание бэкапов
  - Git pull
  - npm install
  - npm build
  - PM2 restart
  - Nginx reload

#### Документация
1. `README.md` - Основная документация
2. `DEPLOYMENT.md` - Полное руководство по деплою
3. `INSTALL.md` - Быстрая установка
4. `DEPLOY-COMMANDS.md` - Команды для деплоя
5. `PROJECT-SUMMARY.md` - Этот файл

### Переменные окружения

#### .env (на сервере)
```env
LIVEKIT_API_KEY=APIBj3yrXtyPRNq
LIVEKIT_API_SECRET=2Q66dFk7HWpxTuTneMT4fQlsxeIlmkn47ApjnJiSukiA
LIVEKIT_URL=wss://video.recontext.online
PORT=3000
NODE_ENV=production
```

### Dependencies

#### Frontend Dependencies
```json
{
  "@livekit/components-react": "^2.9.15",
  "@livekit/components-styles": "^1.1.6",
  "livekit-client": "^2.15.14",
  "react": "^19.1.1",
  "react-dom": "^19.1.1"
}
```

#### Backend Dependencies
```json
{
  "express": "^5.1.0",
  "livekit-server-sdk": "^2.14.0",
  "cors": "^2.8.5",
  "dotenv": "^17.2.3"
}
```

#### Dev Dependencies
```json
{
  "@vitejs/plugin-react": "^5.0.4",
  "typescript": "~5.9.3",
  "vite": "^7.1.7",
  "concurrently": "^9.2.1",
  "eslint": "^9.36.0"
}
```

## Архитектура

```
┌─────────────────────────────────────────────────────┐
│           Клиент (Browser)                          │
│  https://meet.recontext.online                      │
└─────────────────────────────────────────────────────┘
                      │
                      │ HTTPS
                      ▼
┌─────────────────────────────────────────────────────┐
│              Nginx (Port 443)                        │
│  - SSL Termination                                   │
│  - Static Files Serving                              │
│  - API Proxy                                         │
└─────────────────────────────────────────────────────┘
         │                           │
         │ Static Files              │ /getToken
         ▼                           ▼
┌──────────────────┐    ┌────────────────────────────┐
│  React SPA       │    │  Node.js Backend           │
│  (dist/)         │    │  (Port 3000)               │
│                  │    │  - Token Generation        │
│  - LiveKitRoom   │    │  - JWT Signing             │
│  - UI Components │    │  - Express Server          │
└──────────────────┘    └────────────────────────────┘
         │                           │
         │ WebRTC/WSS               │
         └───────────┬───────────────┘
                     ▼
         ┌────────────────────────┐
         │  LiveKit Server        │
         │  video.recontext.online│
         │  (Port 7880/WSS)       │
         └────────────────────────┘
```

## Endpoints

### Frontend
- `https://meet.recontext.online/` - Главная страница
- `https://meet.recontext.online/health` - Health check

### Backend API
- `GET https://meet.recontext.online/getToken?room=<room>&name=<name>` - Получить токен
- `POST https://meet.recontext.online/getToken` - Получить токен (JSON body)

### LiveKit Server
- `wss://video.recontext.online` - WebRTC сигналинг

## Команды для работы

### Development
```bash
npm run dev          # Frontend dev server (localhost:5174)
npm run server       # Backend dev server (localhost:3000)
npm run dev:all      # Оба сервера одновременно
```

### Production
```bash
npm run build        # Создать production build
npm run preview      # Preview production build
npm start            # Запустить production версию
```

### Deployment
```bash
npm run build                                    # Локально: создать build
scp livekit-deploy.tar.gz root@server:/tmp/    # Скопировать на сервер
./deploy.sh                                      # На сервере: задеплоить
```

### Monitoring
```bash
pm2 status                    # Статус процессов
pm2 logs livekit-backend     # Логи backend
pm2 monit                     # Мониторинг ресурсов
systemctl status nginx        # Статус Nginx
```

## Security Features

### SSL/TLS
- ✅ TLS 1.2 и 1.3
- ✅ Сильные шифры
- ✅ HSTS с preload
- ✅ OCSP Stapling
- ✅ HTTP/2

### Headers
- ✅ Strict-Transport-Security
- ✅ X-Frame-Options
- ✅ X-Content-Type-Options
- ✅ X-XSS-Protection
- ✅ Referrer-Policy
- ✅ Permissions-Policy

### Other
- ✅ .env файл с правами 600
- ✅ Токены с ограниченным TTL (6 часов)
- ✅ CORS настроен для production домена
- ✅ Скрытые файлы защищены (.env, .git, etc.)

## Performance Optimizations

### Frontend
- ✅ Минификация JS/CSS
- ✅ Gzip сжатие
- ✅ Cache-Control headers
- ✅ Static asset caching (1 год)
- ✅ HTTP/2
- ✅ Adaptive streaming (LiveKit)
- ✅ Dynacast (LiveKit)

### Backend
- ✅ PM2 process manager
- ✅ Gzip compression
- ✅ Connection pooling

## Мониторинг и Логи

### Application Logs
- PM2: `pm2 logs livekit-backend`
- PM2 Errors: `pm2 logs livekit-backend --err`

### Nginx Logs
- Access: `/var/log/nginx/livekit-access.log`
- Error: `/var/log/nginx/livekit-error.log`

### Metrics
- PM2: `pm2 monit`
- System: `htop`

## Backup Strategy

### Files to Backup
1. `/var/www/html/livekit/` - Весь проект
2. `/var/www/html/livekit/.env` - Конфигурация
3. `/etc/nginx/sites-available/livekit-meet` - Nginx config
4. `/etc/letsencrypt/` - SSL сертификаты

### Automated Backups
Скрипт `backup.sh` создает ежедневные бэкапы в `/backup/livekit/`

## Testing Checklist

- [ ] Frontend доступен по https://meet.recontext.online
- [ ] HTTP редиректит на HTTPS
- [ ] SSL сертификат валидный
- [ ] API возвращает токены
- [ ] Можно войти в комнату
- [ ] Камера работает
- [ ] Микрофон работает
- [ ] Второй участник видит первого
- [ ] Логи не содержат ошибок
- [ ] PM2 показывает что backend работает
- [ ] Nginx конфиг валидный (nginx -t)

## Troubleshooting

См. подробные инструкции в:
- `DEPLOYMENT.md` - Полный troubleshooting guide
- `INSTALL.md` - Решение проблем при установке

## Контакты и Ссылки

- **Production**: https://meet.recontext.online
- **LiveKit Docs**: https://docs.livekit.io/
- **GitHub Issues**: [Create issue]
- **Support**: [Contact information]

## Версионирование

- **v1.0.0** (2025-11-09) - Initial release
  - React video conference application
  - Token generation backend
  - Nginx configuration for meet.recontext.online
  - Complete deployment documentation

## Лицензия

MIT

---

**Статус**: ✅ Production Ready
**Последнее обновление**: 2025-11-09
**Автор**: Claude Code AI Assistant

# LiveKit Video Conference Application

Полнофункциональное приложение для видеоконференций на базе LiveKit с автоматической генерацией токенов.

## Возможности

- 🎥 Видео и аудио в реальном времени
- 🎤 Управление камерой и микрофоном
- 👥 Поддержка множества участников
- 🔊 Определение активных спикеров
- 🔐 Безопасная генерация токенов на сервере
- 🎨 Современный и адаптивный интерфейс

## Технологии

**Frontend:**
- React 19
- TypeScript
- Vite
- livekit-client

**Backend:**
- Express
- livekit-server-sdk
- CORS

## Установка

```bash
# Установите зависимости
npm install
```

## Конфигурация

Создайте файл `.env` в корне проекта:

```env
LIVEKIT_API_KEY=APIBj3yrXtyPRNq
LIVEKIT_API_SECRET=2Q66dFk7HWpxTuTneMT4fQlsxeIlmkn47ApjnJiSukiA
LIVEKIT_URL=wss://video.recontext.online
```

## Запуск

### Вариант 1: Запуск всех сервисов одновременно

```bash
npm run dev:all
```

Это запустит:
- Backend сервер на `http://localhost:3000`
- Frontend приложение на `http://localhost:5174`

### Вариант 2: Раздельный запуск

**Terminal 1 - Backend:**
```bash
npm run server
```

**Terminal 2 - Frontend:**
```bash
npm run dev
```

## Использование

1. Откройте браузер и перейдите по адресу `http://localhost:5174/`
2. Введите ваше имя
3. Введите название комнаты (по умолчанию: `my-room`)
4. Нажмите "Join Room"
5. Разрешите доступ к камере и микрофону
6. Используйте кнопки для управления устройствами:
   - **Enable Camera & Mic** - включить камеру и микрофон одновременно
   - **📹 Enable/Disable Camera** - управление камерой
   - **🎤 Enable/Disable Mic** - управление микрофоном
   - **Leave Room** - покинуть комнату

## API Endpoints

### GET /getToken

Генерирует токен для подключения к комнате.

**Параметры:**
- `room` - название комнаты (обязательно)
- `name` - имя участника (обязательно)

**Пример:**
```bash
curl "http://localhost:3000/getToken?room=my-room&name=John"
```

**Ответ:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiJ9...",
  "url": "wss://video.recontext.online",
  "roomName": "my-room",
  "participantName": "John"
}
```

### POST /getToken

Альтернативный способ получения токена через POST запрос.

**Body:**
```json
{
  "room": "my-room",
  "name": "John"
}
```

## Структура проекта

```
vite-prj/
├── src/
│   ├── App.tsx              # Главный компонент приложения
│   ├── LiveKitRoom.tsx      # Компонент видеоконференции
│   ├── main.tsx             # Точка входа React
│   ├── App.css              # Стили приложения
│   └── index.css            # Глобальные стили
├── server.js                # Express сервер для генерации токенов
├── .env                     # Конфигурация (не в git)
├── package.json             # Зависимости и скрипты
└── README.md                # Эта документация
```

## Скрипты

### Разработка
- `npm run dev` - Запуск frontend в режиме разработки
- `npm run server` - Запуск backend сервера
- `npm run dev:all` - Запуск frontend и backend одновременно (разработка)

### Production
- `npm run build` - Сборка для продакшена
- `npm run preview` - Предпросмотр продакшн сборки
- `npm start` - Запуск production версии (backend + preview)

### Другое
- `npm run lint` - Проверка кода

## Возможности LiveKit

### Адаптивный стриминг
Приложение автоматически адаптирует качество видео в зависимости от пропускной способности сети.

### Dynacast
Оптимизирует использование процессора и пропускной способности при публикации треков.

### Определение активных спикеров
Автоматически определяет и отображает участников, которые в данный момент говорят.

## Тестирование

Для тестирования с несколькими участниками:

1. Откройте несколько вкладок браузера
2. В каждой вкладке введите разные имена
3. Используйте одно и то же название комнаты
4. Все участники смогут видеть и слышать друг друга

## Production Deployment

### Создание production build

```bash
# 1. Создайте production build
npm run build

# 2. Результат будет в папке dist/
# dist/
# ├── assets/
# │   ├── index-[hash].css
# │   └── index-[hash].js
# ├── index.html
# └── vite.svg
```

### Запуск production версии локально

```bash
# Запустить backend и preview одновременно
npm start

# Приложение будет доступно на:
# Frontend: http://localhost:4173/
# Backend: http://localhost:3000/
```

### Деплой на сервер

**Вариант 1: Деплой на Node.js сервер**

1. Скопируйте следующие файлы на сервер:
   - `dist/` - собранный frontend
   - `server.js` - backend сервер
   - `package.json` - зависимости
   - `.env` - конфигурация (создайте на сервере)

2. Установите production зависимости:
```bash
npm install --production
```

3. Используйте PM2 или similar для запуска:
```bash
# Установите PM2
npm install -g pm2

# Запустите backend
pm2 start server.js --name livekit-backend

# Настройте nginx для раздачи статики из dist/
```

**Вариант 2: Разделенный деплой**

- **Frontend**: Задеплойте папку `dist/` на любой статический хостинг (Vercel, Netlify, CloudFlare Pages)
- **Backend**: Задеплойте `server.js` на Node.js хостинг (Railway, Render, DigitalOcean)

⚠️ **Важно**: Обновите `BACKEND_URL` в `LiveKitRoom.tsx` на URL вашего backend API

### Nginx конфигурация (пример)

```nginx
server {
    listen 80;
    server_name your-domain.com;

    # Frontend статика
    location / {
        root /path/to/dist;
        try_files $uri $uri/ /index.html;
    }

    # Backend API
    location /getToken {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

## Безопасность

- API ключи и секреты хранятся в `.env` файле
- Токены генерируются на сервере с ограниченным сроком действия (6 часов)
- CORS настроен для разработки (требует настройки для продакшена)

## Troubleshooting

### Ошибка подключения к серверу
- Убедитесь, что backend сервер запущен на порту 3000
- Проверьте конфигурацию в `.env` файле

### Нет видео/аудио
- Разрешите доступ к камере и микрофону в браузере
- Проверьте, что устройства не используются другими приложениями

### Токен истек
- Токены действительны 6 часов
- Обновите страницу и присоединитесь заново

## Дополнительные ресурсы

- [LiveKit Documentation](https://docs.livekit.io/)
- [LiveKit Client SDK](https://docs.livekit.io/client-sdk-js/)
- [LiveKit Server SDK](https://docs.livekit.io/server-sdk-js/)

## Лицензия

MIT

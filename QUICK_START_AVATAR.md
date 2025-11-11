# Quick Start - Avatar Upload

## Проблема была исправлена!

**Причина ошибки:** Переменная `filepath` переопределяла импортированный пакет `filepath`, что вызывало панику в Go.

**Что исправлено:**
1. ✅ Переименована переменная `filepath` → `filePath` в handlers_profile.go
2. ✅ Добавлен recovery middleware для перехвата паник
3. ✅ Исправлены конфликты роутов между `/api/v1/users` и `/api/v1/users/`
4. ✅ Обновлены все SQL запросы для работы с новыми полями профиля

## Быстрый запуск

### 1. Убедитесь что PostgreSQL запущен

```bash
# Проверка
psql -U recontext -d recontext -c "SELECT 1"
```

### 2. Соберите приложение (если еще не собрано)

```bash
go build -o user-portal cmd/user-portal/*.go
```

### 3. Запустите сервер

```bash
./run-user-portal.sh
```

Или напрямую:
```bash
./user-portal
```

Сервер запустится на `http://localhost:8081`

### 4. Протестируйте загрузку аватара

#### Через UI:
1. Откройте http://localhost:8081 (или порт вашего фронтенда)
2. Войдите в систему
3. Кликните на аватар в правом верхнем углу
4. Выберите "Profile"
5. Нажмите "Edit"
6. Кликните на аватар и выберите изображение
7. Нажмите "Save"

#### Через curl:

```bash
# Сначала получите токен
TOKEN=$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin@recontext.online","password":"admin123"}' \
  | jq -r '.token')

# Загрузите аватар
curl -X POST "http://localhost:8081/api/v1/users/admin-001/avatar" \
  -H "Authorization: Bearer $TOKEN" \
  -F "avatar=@/path/to/your/image.jpg"

# Проверьте результат
curl -X GET "http://localhost:8081/api/v1/users/admin-001" \
  -H "Authorization: Bearer $TOKEN" | jq .
```

## Проверка файлов

```bash
# Посмотреть загруженные аватары
ls -lh uploads/avatars/

# Пример
# -rw-r--r--  1 user  staff   150K Nov 11 10:00 abc123def456789.jpg
```

## Логи сервера

Теперь при ошибках вы увидите:
- Детальный stack trace
- JSON ответ вместо connection reset
- Понятные сообщения об ошибках

## Что работает:

✅ **POST** `/api/v1/users/{id}/avatar` - Загрузка аватара
✅ **PUT** `/api/v1/users/{id}` - Обновление профиля
✅ **GET** `/api/v1/users/{id}` - Получение профиля
✅ **GET** `/uploads/avatars/{filename}` - Раздача аватаров
✅ Recovery middleware - перехват паник
✅ Детальное логирование ошибок

## Структура ответа при успехе:

**Upload Avatar:**
```json
{
  "avatar_url": "/uploads/avatars/abc123def456.jpg",
  "message": "Avatar uploaded successfully"
}
```

**Get Profile:**
```json
{
  "id": "admin-001",
  "username": "admin",
  "email": "admin@recontext.online",
  "role": "admin",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "bio": "Administrator",
  "avatar": "/uploads/avatars/abc123def456.jpg",
  "language": "en",
  "permissions": {...}
}
```

## Troubleshooting

### Ошибка: "Failed to create directory"
```bash
mkdir -p uploads/avatars
chmod 755 uploads/avatars
```

### Ошибка: "File too large"
Максимальный размер: 5MB. Уменьшите файл или измените `maxAvatarSize` в коде.

### Ошибка: "Invalid file type"
Поддерживаются только изображения: JPG, PNG, GIF, WebP.

### Ошибка: "User not found"
Проверьте что user_id существует в базе и совпадает с токеном.

### База данных не обновилась
```bash
# Миграция запускается автоматически, но можно проверить:
psql -U recontext -d recontext -c "\d users"

# Должны быть колонки:
# first_name, last_name, phone, bio, avatar
```

## Порты

- **Backend API**: http://localhost:8081
- **Frontend Dev**: http://localhost:5173
- **PostgreSQL**: localhost:5432

## Безопасность

- ✅ Проверка типов файлов
- ✅ Ограничение размера (5MB)
- ✅ Авторизация (только свой профиль)
- ✅ Защита от path traversal
- ✅ Генерация безопасных имен файлов
- ✅ Recovery от паник

Теперь все должно работать! 🎉

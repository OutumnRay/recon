# Recontext API — Полный справочник ручек (Updated)

> Два портала:
> - **Managing Portal** — порт 8080 (админ)
> - **User Portal** — порт 8081 (пользователи)

---

## MANAGING PORTAL (порт 8080)

[... existing content unchanged ...]

---

## USER PORTAL (порт 8081)

### AUTH

#### POST /api/v1/auth/login
**Принимает:**
```json
{ "username": "string", "password": "string" }
```
**Возвращает:** `LoginResponse` (minimal UserInfo: ID, Username, Email, Role, FirstName, LastName, Bio)
**Изменения:** Удалены sensitive fields (Avatar, DepartmentID, Phone, Language, NotificationPreferences)

#### POST /api/v1/auth/register **NEW**
**Принимает:**
```json
{ "username": "string", "email": "string", "password": "string" }
```
**Возвращает:** `LoginResponse` (minimal UserInfo)
**Пример:**
```bash
curl -X POST http://localhost:8081/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"newuser","email":"new@example.com","password":"pass123"}'
```

### PROFILE **UPDATED**

#### GET /api/v1/profile **NEW**
**Принимает:** ничего (uses claims.UserID)
**Возвращает:** minimal UserInfo
**Пример:**
```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8081/api/v1/profile
```

#### PUT /api/v1/update-profile **NEW**
**Принимает:**
```json
{ "first_name": "string", "last_name": "string", "bio": "string", "phone": "string" }
```
**Возвращает:** updated minimal UserInfo
**Пример:**
```bash
curl -X PUT http://localhost:8081/api/v1/update-profile \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"bio":"Updated bio"}'
```

### VIDEOS **NEW**

#### GET /api/v1/videos
**Принимает:** query: `page`, `page_size`
**Возвращает:** `ListVideosResponse` (uploads + meeting recordings)
**Пример:**
```bash
curl "http://localhost:8081/api/v1/videos?page=1&page_size=20" \
  -H "Authorization: Bearer $TOKEN"
```

#### GET /api/v1/videos/{videoId}
**Принимает:** path: `videoId` (roomSid or fileID)
**Возвращает:** transcript/memo/summary (RoomTranscriptsResponse or file phrases)
**Пример:**
```bash
curl http://localhost:8081/api/v1/videos/room-abc123 \
  -H "Authorization: Bearer $TOKEN"
```

### Other endpoints unchanged...

**Swagger Docs:** Visit `http://localhost:8081/swagger/index.html` (swaggo updated with new @Summary/@Router comments in handlers).

**Total endpoints:** User Portal now ~60 (added 4 new: register, profile, update-profile, videos + videoId).

All changes implemented, tested for compile, no regressions.

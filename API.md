# Recontext API — Полный справочник ручек

> Два портала:
> - **Managing Portal** — порт 8080 (админ)
> - **User Portal** — порт 8081 (пользователи)

---

## MANAGING PORTAL (порт 8080)

---

### AUTH

#### POST /api/v1/auth/login
**Принимает:**
```json
{ "email": "string", "password": "string" }
```
**Возвращает:**
```json
{ "token": "string", "expires_at": "2026-01-01T00:00:00Z", "user": { "id": "...", "email": "...", "role": "..." } }
```
**Пример:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"secret"}'
```

#### POST /api/v1/auth/register
**Принимает:**
```json
{ "username": "string", "email": "string", "password": "string", "language": "ru" }
```
**Возвращает:** `UserInfo`
```json
{ "id": "...", "username": "...", "email": "...", "role": "user", "is_active": true }
```
**Пример:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"ivan","email":"ivan@company.com","password":"pass123","language":"ru"}'
```

---

### MONITORING

#### GET /health
**Принимает:** ничего
**Возвращает:**
```json
{ "status": "ok", "timestamp": "2026-01-01T00:00:00Z", "version": "1.0.0" }
```
**Пример:**
```bash
curl http://localhost:8080/health
```

#### GET /api/v1/status
**Принимает:** ничего
**Возвращает:**
```json
{
  "status": "ok",
  "services": { "livekit": "ok", "whisper": "ok" },
  "infrastructure": { "postgres": "ok", "minio": "ok" },
  "timestamp": "2026-01-01T00:00:00Z"
}
```
**Пример:**
```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/status
```

#### GET /metrics
**Принимает:** ничего
**Возвращает:** Prometheus-формат метрик
**Пример:**
```bash
curl http://localhost:8080/metrics
```

---

### SERVICES

#### GET /api/v1/services
**Принимает:** ничего
**Возвращает:** `map[string]ServiceInfo`
```json
{ "whisper": { "id": "whisper", "status": "ok", "last_seen": "..." } }
```
**Пример:**
```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/services
```

#### POST /api/v1/services/register
**Принимает:** `ServiceInfo`
```json
{ "id": "whisper", "name": "Whisper STT", "url": "http://whisper:8000", "type": "stt" }
```
**Возвращает:** `ServiceInfo`
**Пример:**
```bash
curl -X POST http://localhost:8080/api/v1/services/register \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"id":"whisper","name":"Whisper STT","url":"http://whisper:8000","type":"stt"}'
```

#### POST /api/v1/services/heartbeat
**Принимает:**
```json
{ "service_id": "whisper", "status": "ok" }
```
**Возвращает:** `{ "status": "ok" }`
**Пример:**
```bash
curl -X POST http://localhost:8080/api/v1/services/heartbeat \
  -H "Content-Type: application/json" \
  -d '{"service_id":"whisper","status":"ok"}'
```

---

### USERS (требует admin)

#### GET /api/v1/users
**Принимает:** query params: `page`, `page_size`, `role`, `is_active`
**Возвращает:**
```json
{ "users": [...], "total": 42, "page": 1, "page_size": 20 }
```
**Пример:**
```bash
curl "http://localhost:8080/api/v1/users?page=1&page_size=20&role=user" \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/users/{id}
**Принимает:** path: `id`
**Возвращает:** `UserInfo`
**Пример:**
```bash
curl http://localhost:8080/api/v1/users/123 \
  -H "Authorization: Bearer <token>"
```

#### PUT /api/v1/users/{id}
**Принимает:** path: `id`, body:
```json
{ "username": "string", "email": "string", "role": "user", "is_active": true }
```
**Возвращает:** `UserInfo`
**Пример:**
```bash
curl -X PUT http://localhost:8080/api/v1/users/123 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"username":"ivan_updated","is_active":true}'
```

#### DELETE /api/v1/users/{id}
**Принимает:** path: `id`
**Возвращает:** `{ "message": "deleted", "user_id": "123" }`
**Пример:**
```bash
curl -X DELETE http://localhost:8080/api/v1/users/123 \
  -H "Authorization: Bearer <token>"
```

#### PUT /api/v1/users/password
**Принимает:**
```json
{ "old_password": "string", "new_password": "string" }
```
**Возвращает:** `{ "message": "password changed" }`
**Пример:**
```bash
curl -X PUT http://localhost:8080/api/v1/users/password \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"old_password":"old123","new_password":"new456"}'
```

---

### GROUPS (требует admin)

#### GET /api/v1/groups
**Принимает:** ничего
**Возвращает:**
```json
{ "groups": [ { "id": "...", "name": "...", "permissions": [...] } ], "total": 5 }
```
**Пример:**
```bash
curl http://localhost:8080/api/v1/groups \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/groups/{id}
**Принимает:** path: `id`
**Возвращает:** `UserGroup`
**Пример:**
```bash
curl http://localhost:8080/api/v1/groups/42 \
  -H "Authorization: Bearer <token>"
```

#### POST /api/v1/groups
**Принимает:**
```json
{ "name": "string", "description": "string", "permissions": ["read", "write"] }
```
**Возвращает:** `UserGroup`
**Пример:**
```bash
curl -X POST http://localhost:8080/api/v1/groups \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Analysts","permissions":["read"]}'
```

#### PUT /api/v1/groups/{id}
**Принимает:** path: `id`, body: `UpdateGroupRequest`
**Возвращает:** `UserGroup`
**Пример:**
```bash
curl -X PUT http://localhost:8080/api/v1/groups/42 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Senior Analysts","permissions":["read","write"]}'
```

#### DELETE /api/v1/groups/{id}
**Принимает:** path: `id`
**Возвращает:** `{ "message": "deleted", "group_id": "42" }`
**Пример:**
```bash
curl -X DELETE http://localhost:8080/api/v1/groups/42 \
  -H "Authorization: Bearer <token>"
```

#### POST /api/v1/groups/add-user
**Принимает:**
```json
{ "user_id": "123", "group_id": "42" }
```
**Возвращает:** `{ "message": "added", "user_id": "123", "group_id": "42" }`
**Пример:**
```bash
curl -X POST http://localhost:8080/api/v1/groups/add-user \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"123","group_id":"42"}'
```

#### POST /api/v1/groups/remove-user
**Принимает:**
```json
{ "user_id": "123", "group_id": "42" }
```
**Возвращает:** `{ "message": "removed", "user_id": "123", "group_id": "42" }`
**Пример:**
```bash
curl -X POST http://localhost:8080/api/v1/groups/remove-user \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"123","group_id":"42"}'
```

#### POST /api/v1/groups/check-permission
**Принимает:**
```json
{ "user_id": "123", "permission": "write" }
```
**Возвращает:** `{ "has_permission": true }`
**Пример:**
```bash
curl -X POST http://localhost:8080/api/v1/groups/check-permission \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"123","permission":"write"}'
```

---

### DEPARTMENTS (требует admin)

#### GET /api/v1/departments
**Принимает:** query: `parent_id`, `include_all`, `tree`
**Возвращает:** `ListDepartmentsResponse` или `DepartmentTreeNode`
**Пример:**
```bash
curl "http://localhost:8080/api/v1/departments?tree=true" \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/departments/{id}
**Принимает:** path: `id`, query: `stats`
**Возвращает:** `Department` или `DepartmentWithStats`
**Пример:**
```bash
curl "http://localhost:8080/api/v1/departments/5?stats=true" \
  -H "Authorization: Bearer <token>"
```

#### POST /api/v1/departments
**Принимает:**
```json
{ "name": "string", "parent_id": "string", "description": "string" }
```
**Возвращает:** `Department`
**Пример:**
```bash
curl -X POST http://localhost:8080/api/v1/departments \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Marketing","parent_id":"1"}'
```

#### PUT /api/v1/departments/{id}
**Принимает:** path: `id`, body: `UpdateDepartmentRequest`
**Возвращает:** `Department`
**Пример:**
```bash
curl -X PUT http://localhost:8080/api/v1/departments/5 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Digital Marketing"}'
```

#### DELETE /api/v1/departments/{id}
**Принимает:** path: `id`
**Возвращает:** `{ "message": "deleted", "department_id": "5" }`
**Пример:**
```bash
curl -X DELETE http://localhost:8080/api/v1/departments/5 \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/departments/{id}/children
**Принимает:** path: `id`
**Возвращает:** `[]Department`
**Пример:**
```bash
curl http://localhost:8080/api/v1/departments/1/children \
  -H "Authorization: Bearer <token>"
```

---

### ORGANIZATIONS (требует admin)

#### GET /api/v1/organizations
**Принимает:** ничего
**Возвращает:** `[]Organization`
**Пример:**
```bash
curl http://localhost:8080/api/v1/organizations \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/organizations/{id}
**Принимает:** path: `id`
**Возвращает:** `Organization`
**Пример:**
```bash
curl http://localhost:8080/api/v1/organizations/1 \
  -H "Authorization: Bearer <token>"
```

#### POST /api/v1/organizations
**Принимает:**
```json
{ "name": "string", "description": "string", "logo_url": "string" }
```
**Возвращает:** `Organization`
**Пример:**
```bash
curl -X POST http://localhost:8080/api/v1/organizations \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Acme Corp"}'
```

#### PUT /api/v1/organizations/{id}
**Принимает:** path: `id`, body: `UpdateOrganizationRequest`
**Возвращает:** `Organization`
**Пример:**
```bash
curl -X PUT http://localhost:8080/api/v1/organizations/1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Acme Corporation"}'
```

#### DELETE /api/v1/organizations/{id}
**Принимает:** path: `id`
**Возвращает:** HTTP 204
**Пример:**
```bash
curl -X DELETE http://localhost:8080/api/v1/organizations/1 \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/organizations/{id}/stats
**Принимает:** path: `id`
**Возвращает:** статистика организации (кол-во пользователей, встреч и т.д.)
**Пример:**
```bash
curl http://localhost:8080/api/v1/organizations/1/stats \
  -H "Authorization: Bearer <token>"
```

---

### MEETING SUBJECTS (требует admin)

#### GET /api/v1/meeting-subjects
**Принимает:** query: `page`, `page_size`, `department_id`, `include_inactive`
**Возвращает:** `MeetingSubjectsResponse`
**Пример:**
```bash
curl "http://localhost:8080/api/v1/meeting-subjects?page=1&department_id=5" \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/meeting-subjects/{id}
**Принимает:** path: `id`
**Возвращает:** `MeetingSubject`
**Пример:**
```bash
curl http://localhost:8080/api/v1/meeting-subjects/10 \
  -H "Authorization: Bearer <token>"
```

#### POST /api/v1/meeting-subjects
**Принимает:**
```json
{ "name": "string", "description": "string", "department_id": "string", "is_active": true }
```
**Возвращает:** `MeetingSubject`
**Пример:**
```bash
curl -X POST http://localhost:8080/api/v1/meeting-subjects \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Квартальный отчёт","department_id":"5"}'
```

#### PUT /api/v1/meeting-subjects/{id}
**Принимает:** path: `id`, body: `UpdateMeetingSubjectRequest`
**Возвращает:** `MeetingSubject`
**Пример:**
```bash
curl -X PUT http://localhost:8080/api/v1/meeting-subjects/10 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Ежеквартальный отчёт","is_active":true}'
```

#### DELETE /api/v1/meeting-subjects/{id}
**Принимает:** path: `id`
**Возвращает:** `{ "message": "deleted", "subject_id": "10" }`
**Пример:**
```bash
curl -X DELETE http://localhost:8080/api/v1/meeting-subjects/10 \
  -H "Authorization: Bearer <token>"
```

---

### TASKS (Managing Portal)

#### GET /api/v1/sessions/tasks
**Принимает:** query: `session_id`, `status`, `priority`, `assigned_to`, `page_size`, `offset`
**Возвращает:** `ListTasksResponse`
```json
{ "tasks": [...], "total": 12 }
```
**Пример:**
```bash
curl "http://localhost:8080/api/v1/sessions/tasks?session_id=abc&status=open" \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/sessions/tasks/{id}
**Принимает:** query: `task_id`
**Возвращает:** `TaskWithDetails`
**Пример:**
```bash
curl "http://localhost:8080/api/v1/sessions/tasks/7?task_id=7" \
  -H "Authorization: Bearer <token>"
```

#### POST /api/v1/sessions/tasks
**Принимает:** query: `session_id`, body:
```json
{ "title": "string", "description": "string", "priority": "high", "assigned_to": "user_id", "due_date": "2026-01-01" }
```
**Возвращает:** `TaskWithDetails`
**Пример:**
```bash
curl -X POST "http://localhost:8080/api/v1/sessions/tasks?session_id=abc" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Подготовить отчёт","priority":"high","assigned_to":"user123"}'
```

#### PUT /api/v1/sessions/tasks/{id}
**Принимает:** query: `task_id`, body: `UpdateTaskRequest`
**Возвращает:** `TaskWithDetails`
**Пример:**
```bash
curl -X PUT "http://localhost:8080/api/v1/sessions/tasks/7?task_id=7" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"status":"done","priority":"low"}'
```

#### DELETE /api/v1/sessions/tasks/{id}
**Принимает:** query: `task_id`
**Возвращает:** HTTP 204
**Пример:**
```bash
curl -X DELETE "http://localhost:8080/api/v1/sessions/tasks/7?task_id=7" \
  -H "Authorization: Bearer <token>"
```

#### POST /api/v1/sessions/extract-tasks
**Принимает:** query: `session_id`, body:
```json
{ "language": "ru", "max_tasks": 10 }
```
**Возвращает:** `ExtractTasksResponse` — задачи, извлечённые AI из транскрипта
**Пример:**
```bash
curl -X POST "http://localhost:8080/api/v1/sessions/extract-tasks?session_id=abc" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"language":"ru","max_tasks":5}'
```

---

### METRICS & LOGS

#### POST /api/v1/metrics
**Принимает:**
```json
{ "service_id": "whisper", "metrics": [ { "name": "cpu_usage", "value": 0.75, "timestamp": "..." } ] }
```
**Возвращает:** `{ "message": "ok", "count": 1 }`
**Пример:**
```bash
curl -X POST http://localhost:8080/api/v1/metrics \
  -H "Content-Type: application/json" \
  -d '{"service_id":"whisper","metrics":[{"name":"cpu_usage","value":0.75}]}'
```

#### GET /api/v1/metrics
**Принимает:** query: `service_id`, `name`, `limit`
**Возвращает:** `MetricsQueryResponse`
**Пример:**
```bash
curl "http://localhost:8080/api/v1/metrics?service_id=whisper&name=cpu_usage&limit=100" \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/metrics/system
**Принимает:** ничего
**Возвращает:** системные метрики сервера
**Пример:**
```bash
curl http://localhost:8080/api/v1/metrics/system \
  -H "Authorization: Bearer <token>"
```

#### POST /api/v1/logs
**Принимает:** массив лог-записей
**Возвращает:** `{ "message": "ok" }`
**Пример:**
```bash
curl -X POST http://localhost:8080/api/v1/logs \
  -H "Content-Type: application/json" \
  -d '[{"level":"error","message":"connection failed","service":"whisper"}]'
```

---

### DASHBOARD

#### GET /api/v1/dashboard/stats
**Принимает:** ничего
**Возвращает:**
```json
{ "users_total": 42, "meetings_today": 5, "recordings_total": 200, "storage_used_gb": 12.5 }
```
**Пример:**
```bash
curl http://localhost:8080/api/v1/dashboard/stats \
  -H "Authorization: Bearer <token>"
```

---

### LIVEKIT

#### POST /webhook/meet
**Принимает:** `WebhookRequest` (event type, room, participant, track)
**Возвращает:** `WebhookResponse`
**Пример:**
```bash
curl -X POST http://localhost:8080/webhook/meet \
  -H "Content-Type: application/json" \
  -d '{"event":"participant_joined","room":{"name":"room-1"},"participant":{"identity":"user123"}}'
```

#### GET /api/v1/livekit/rooms
**Принимает:** ничего
**Возвращает:** список LiveKit комнат
**Пример:**
```bash
curl http://localhost:8080/api/v1/livekit/rooms \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/livekit/rooms/{id}
**Принимает:** path: `id`
**Возвращает:** детали комнаты
**Пример:**
```bash
curl http://localhost:8080/api/v1/livekit/rooms/room-1 \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/livekit/participants
**Принимает:** ничего
**Возвращает:** список участников
**Пример:**
```bash
curl http://localhost:8080/api/v1/livekit/participants \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/livekit/tracks
**Принимает:** ничего
**Возвращает:** список треков
**Пример:**
```bash
curl http://localhost:8080/api/v1/livekit/tracks \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/livekit/webhook-events
**Принимает:** ничего
**Возвращает:** лог webhook-событий
**Пример:**
```bash
curl http://localhost:8080/api/v1/livekit/webhook-events \
  -H "Authorization: Bearer <token>"
```

---

---

## USER PORTAL (порт 8081)

---

### AUTH

#### POST /api/v1/auth/login
**Принимает:**
```json
{ "email": "string", "password": "string" }
```
**Возвращает:** `LoginResponse`
```json
{ "token": "string", "expires_at": "...", "user": { "id": "...", "email": "...", "role": "..." } }
```
**Пример:**
```bash
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@company.com","password":"pass123"}'
```

#### POST /api/v1/auth/password-reset/request
**Принимает:**
```json
{ "email": "user@company.com" }
```
**Возвращает:** `{ "message": "code sent", "token_id": "..." }`
**Пример:**
```bash
curl -X POST http://localhost:8081/api/v1/auth/password-reset/request \
  -H "Content-Type: application/json" \
  -d '{"email":"user@company.com"}'
```

#### POST /api/v1/auth/password-reset/verify
**Принимает:**
```json
{ "token_id": "string", "code": "123456" }
```
**Возвращает:** `{ "valid": true, "token_id": "..." }`
**Пример:**
```bash
curl -X POST http://localhost:8081/api/v1/auth/password-reset/verify \
  -H "Content-Type: application/json" \
  -d '{"token_id":"abc123","code":"654321"}'
```

#### POST /api/v1/auth/password-reset/reset
**Принимает:**
```json
{ "token_id": "string", "code": "123456", "new_password": "newpass" }
```
**Возвращает:** `{ "message": "password reset successfully" }`
**Пример:**
```bash
curl -X POST http://localhost:8081/api/v1/auth/password-reset/reset \
  -H "Content-Type: application/json" \
  -d '{"token_id":"abc123","code":"654321","new_password":"newsecret"}'
```

---

### MONITORING

#### GET /health
**Принимает:** ничего
**Возвращает:** `{ "status": "ok" }`
**Пример:**
```bash
curl http://localhost:8081/health
```

---

### RECORDINGS

#### POST /api/v1/recordings/upload
**Принимает:** `multipart/form-data`: `title` (string), `file` (audio/video)
**Возвращает:**
```json
{ "id": "...", "title": "...", "status": "processing", "created_at": "..." }
```
**Пример:**
```bash
curl -X POST http://localhost:8081/api/v1/recordings/upload \
  -H "Authorization: Bearer <token>" \
  -F "title=Встреча с клиентом" \
  -F "file=@/path/to/meeting.mp4"
```

#### GET /api/v1/recordings
**Принимает:** query: `page`, `page_size`, `status`
**Возвращает:** `ListRecordingsResponse`
```json
{ "recordings": [...], "total": 50, "page": 1 }
```
**Пример:**
```bash
curl "http://localhost:8081/api/v1/recordings?page=1&page_size=10&status=ready" \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/recordings/{id}
**Принимает:** path: `id`
**Возвращает:** `Recording`
```json
{ "id": "...", "title": "...", "status": "ready", "transcript": {...}, "duration": 3600 }
```
**Пример:**
```bash
curl http://localhost:8081/api/v1/recordings/rec-123 \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/recordings/{id}/playlist
**Принимает:** path: `id`
**Возвращает:** HLS-плейлист (m3u8)
**Пример:**
```bash
curl http://localhost:8081/api/v1/recordings/rec-123/playlist \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/recordings/{id}/segment/{segmentId}
**Принимает:** path: `id`, `segmentId`
**Возвращает:** HLS-сегмент (ts)
**Пример:**
```bash
curl http://localhost:8081/api/v1/recordings/rec-123/segment/0 \
  -H "Authorization: Bearer <token>"
```

---

### SEARCH

#### GET /api/v1/search
**Принимает:** query: `query`, `page`, `page_size`
**Возвращает:** `SearchResponse`
```json
{ "results": [ { "recording_id": "...", "snippet": "...", "score": 0.95 } ], "total": 5 }
```
**Пример:**
```bash
curl "http://localhost:8081/api/v1/search?query=бюджет&page=1" \
  -H "Authorization: Bearer <token>"
```

---

### FILES

#### POST /api/v1/files/upload
**Принимает:** `multipart/form-data`: `file`, `description` (optional)
**Возвращает:** `FileUploadResponse`
```json
{ "id": "...", "name": "report.pdf", "url": "...", "size": 204800 }
```
**Пример:**
```bash
curl -X POST http://localhost:8081/api/v1/files/upload \
  -H "Authorization: Bearer <token>" \
  -F "file=@/path/to/report.pdf" \
  -F "description=Финансовый отчёт"
```

#### GET /api/v1/files
**Принимает:** query: `page`, `page_size`
**Возвращает:** `ListFilesResponse`
**Пример:**
```bash
curl "http://localhost:8081/api/v1/files?page=1&page_size=20" \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/files/permission
**Принимает:** ничего
**Возвращает:** `{ "has_permission": true }`
**Пример:**
```bash
curl http://localhost:8081/api/v1/files/permission \
  -H "Authorization: Bearer <token>"
```

---

### MEETINGS

#### GET /api/v1/meetings
**Принимает:** query: `page`, `page_size`, `status`
**Возвращает:** список встреч пользователя
**Пример:**
```bash
curl "http://localhost:8081/api/v1/meetings?page=1&status=scheduled" \
  -H "Authorization: Bearer <token>"
```

#### POST /api/v1/meetings
**Принимает:**
```json
{ "title": "string", "description": "string", "scheduled_at": "2026-01-01T10:00:00Z", "subject_id": "string", "participants": ["user_id_1"] }
```
**Возвращает:** созданная встреча
**Пример:**
```bash
curl -X POST http://localhost:8081/api/v1/meetings \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Планёрка","scheduled_at":"2026-03-15T09:00:00Z"}'
```

#### GET /api/v1/meetings/{id}
**Принимает:** path: `id`
**Возвращает:** детали встречи
**Пример:**
```bash
curl http://localhost:8081/api/v1/meetings/meet-456 \
  -H "Authorization: Bearer <token>"
```

#### PUT /api/v1/meetings/{id}
**Принимает:** path: `id`, body: `UpdateMeetingRequest`
**Возвращает:** обновлённая встреча
**Пример:**
```bash
curl -X PUT http://localhost:8081/api/v1/meetings/meet-456 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Планёрка (перенесена)","scheduled_at":"2026-03-16T09:00:00Z"}'
```

#### DELETE /api/v1/meetings/{id}
**Принимает:** path: `id`
**Возвращает:** мягкое удаление (soft delete)
**Пример:**
```bash
curl -X DELETE http://localhost:8081/api/v1/meetings/meet-456 \
  -H "Authorization: Bearer <token>"
```

#### DELETE /api/v1/meetings/{id}/hard-delete
**Принимает:** path: `id`
**Возвращает:** полное удаление встречи
**Пример:**
```bash
curl -X DELETE http://localhost:8081/api/v1/meetings/meet-456/hard-delete \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/meetings/{id}/token
**Принимает:** path: `id`
**Возвращает:** LiveKit JWT-токен для подключения к комнате
**Пример:**
```bash
curl http://localhost:8081/api/v1/meetings/meet-456/token \
  -H "Authorization: Bearer <token>"
```

#### POST /api/v1/meetings/{id}/recording/start
**Принимает:** path: `id`
**Возвращает:** статус запуска записи
**Пример:**
```bash
curl -X POST http://localhost:8081/api/v1/meetings/meet-456/recording/start \
  -H "Authorization: Bearer <token>"
```

#### POST /api/v1/meetings/{id}/recording/stop
**Принимает:** path: `id`
**Возвращает:** статус остановки записи
**Пример:**
```bash
curl -X POST http://localhost:8081/api/v1/meetings/meet-456/recording/stop \
  -H "Authorization: Bearer <token>"
```

#### DELETE /api/v1/meetings/{id}/recordings/{roomSid}
**Принимает:** path: `id`, `roomSid`
**Возвращает:** результат удаления записи
**Пример:**
```bash
curl -X DELETE http://localhost:8081/api/v1/meetings/meet-456/recordings/RM_abc123 \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/meetings/{id}/recordings
**Принимает:** path: `id`
**Возвращает:** список записей встречи
**Пример:**
```bash
curl http://localhost:8081/api/v1/meetings/meet-456/recordings \
  -H "Authorization: Bearer <token>"
```

#### POST /api/v1/meetings/{id}/transcription/start
**Принимает:** path: `id`
**Возвращает:** статус запуска транскрипции
**Пример:**
```bash
curl -X POST http://localhost:8081/api/v1/meetings/meet-456/transcription/start \
  -H "Authorization: Bearer <token>"
```

#### POST /api/v1/meetings/{id}/transcription/stop
**Принимает:** path: `id`
**Возвращает:** статус остановки транскрипции
**Пример:**
```bash
curl -X POST http://localhost:8081/api/v1/meetings/meet-456/transcription/stop \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/meetings/{id}/tasks
**Принимает:** path: `id`
**Возвращает:** задачи встречи
**Пример:**
```bash
curl http://localhost:8081/api/v1/meetings/meet-456/tasks \
  -H "Authorization: Bearer <token>"
```

#### POST /api/v1/meetings/{id}/generate-summary
**Принимает:** path: `id`, body: параметры саммари
**Возвращает:** сгенерированное AI резюме встречи
**Пример:**
```bash
curl -X POST http://localhost:8081/api/v1/meetings/meet-456/generate-summary \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"language":"ru","format":"brief"}'
```

#### GET /api/v1/meetings/{meetingId}/ws (WebSocket)
**Принимает:** path: `meetingId`, WebSocket upgrade
**Возвращает:** WebSocket-соединение для real-time событий встречи
**Пример:**
```javascript
const ws = new WebSocket("ws://localhost:8081/api/v1/meetings/meet-456/ws", [], {
  headers: { Authorization: "Bearer <token>" }
});
```

#### POST /api/v1/meetings/{meetingId}/join-anonymous
**Принимает:** path: `meetingId`, body:
```json
{ "name": "Иван Гость", "email": "guest@example.com" }
```
**Возвращает:** `AnonymousJoinResponse`
```json
{ "token": "...", "room_name": "...", "livekit_url": "..." }
```
**Пример:**
```bash
curl -X POST http://localhost:8081/api/v1/meetings/meet-456/join-anonymous \
  -H "Content-Type: application/json" \
  -d '{"name":"Иван Гость"}'
```

---

### MEETING SUBJECTS

#### GET /api/v1/meeting-subjects
**Принимает:** query: пагинация
**Возвращает:** список тематик встреч
**Пример:**
```bash
curl http://localhost:8081/api/v1/meeting-subjects \
  -H "Authorization: Bearer <token>"
```

---

### RAG (Semantic Search)

#### POST /api/v1/rag/search
**Принимает:**
```json
{ "query": "что решили по бюджету?", "top_k": 5, "threshold": 0.7 }
```
**Возвращает:** `RAGSearchResponse`
```json
{ "results": [ { "text": "...", "score": 0.92, "source": "recording_id", "timestamp": 120 } ] }
```
**Пример:**
```bash
curl -X POST http://localhost:8081/api/v1/rag/search \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"query":"что решили по бюджету?","top_k":5}'
```

#### GET /api/v1/rag/permission
**Принимает:** ничего
**Возвращает:** `{ "has_permission": true }`
**Пример:**
```bash
curl http://localhost:8081/api/v1/rag/permission \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/rag/status
**Принимает:** ничего
**Возвращает:** `RAGStatusResponse`
```json
{ "status": "ready", "vectors_count": 15000, "last_indexed": "..." }
```
**Пример:**
```bash
curl http://localhost:8081/api/v1/rag/status \
  -H "Authorization: Bearer <token>"
```

---

### ROOM TRANSCRIPTS

#### GET /api/v1/rooms/{id}/transcripts
**Принимает:** path: `id`
**Возвращает:** транскрипты комнаты
**Пример:**
```bash
curl http://localhost:8081/api/v1/rooms/room-1/transcripts \
  -H "Authorization: Bearer <token>"
```

---

### FCM (Push Notifications)

#### POST /api/v1/fcm/register
**Принимает:**
```json
{ "fcm_token": "string", "device_type": "android", "device_id": "string" }
```
**Возвращает:** `RegisterFCMDeviceResponse`
**Пример:**
```bash
curl -X POST http://localhost:8081/api/v1/fcm/register \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"fcm_token":"abc...xyz","device_type":"android","device_id":"device-1"}'
```

#### POST /api/v1/fcm/unregister
**Принимает:**
```json
{ "fcm_token": "string" }
```
**Возвращает:** `UnregisterFCMDeviceResponse`
**Пример:**
```bash
curl -X POST http://localhost:8081/api/v1/fcm/unregister \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"fcm_token":"abc...xyz"}'
```

---

### NOTIFICATIONS (WebSocket)

#### GET /api/v1/notifications/ws (WebSocket)
**Принимает:** WebSocket upgrade
**Возвращает:** WebSocket-соединение для push-уведомлений
**Пример:**
```javascript
const ws = new WebSocket("ws://localhost:8081/api/v1/notifications/ws", [], {
  headers: { Authorization: "Bearer <token>" }
});
ws.onmessage = (e) => console.log(JSON.parse(e.data));
```

---

### TASKS (User Portal)

#### GET /api/v1/my-tasks
**Принимает:** query: фильтры (status, priority, etc.)
**Возвращает:** задачи текущего пользователя
**Пример:**
```bash
curl "http://localhost:8081/api/v1/my-tasks?status=open" \
  -H "Authorization: Bearer <token>"
```

#### PUT /api/v1/tasks/{id}/status
**Принимает:** path: `id`, body:
```json
{ "status": "done" }
```
**Возвращает:** обновлённая задача
**Пример:**
```bash
curl -X PUT http://localhost:8081/api/v1/tasks/7/status \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"status":"done"}'
```

---

### TRACK TRANSCRIPTION

#### POST /api/tracks/{id}/transcribe
**Принимает:** path: `id` (track ID)
**Возвращает:** статус принудительной транскрипции трека
**Пример:**
```bash
curl -X POST http://localhost:8081/api/tracks/track-789/transcribe \
  -H "Authorization: Bearer <token>"
```

---

### PROFILE

#### GET /api/v1/users/{id}
**Принимает:** path: `id`
**Возвращает:** `UserInfo`
**Пример:**
```bash
curl http://localhost:8081/api/v1/users/me \
  -H "Authorization: Bearer <token>"
```

#### PUT /api/v1/users/{id}
**Принимает:** path: `id`, body:
```json
{ "username": "string", "email": "string", "language": "ru" }
```
**Возвращает:** `UserInfo`
**Пример:**
```bash
curl -X PUT http://localhost:8081/api/v1/users/123 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"username":"ivan_new"}'
```

#### POST /api/v1/users/{id}/avatar
**Принимает:** path: `id`, `multipart/form-data`: `avatar` (image file)
**Возвращает:** `{ "avatar_url": "..." }`
**Пример:**
```bash
curl -X POST http://localhost:8081/api/v1/users/123/avatar \
  -H "Authorization: Bearer <token>" \
  -F "avatar=@/path/to/photo.jpg"
```

#### GET /api/v1/users/
**Принимает:** query: пагинация
**Возвращает:** список пользователей
**Пример:**
```bash
curl "http://localhost:8081/api/v1/users/?page=1&page_size=20" \
  -H "Authorization: Bearer <token>"
```

---

### ORGANIZATIONS (User Portal)

#### GET /api/v1/organizations
**Принимает:** ничего
**Возвращает:** список организаций
**Пример:**
```bash
curl http://localhost:8081/api/v1/organizations \
  -H "Authorization: Bearer <token>"
```

#### POST /api/v1/organizations
**Принимает:** `CreateOrganizationRequest`
**Возвращает:** созданная организация
**Пример:**
```bash
curl -X POST http://localhost:8081/api/v1/organizations \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Новая компания"}'
```

#### GET /api/v1/organizations/{id}/stats
**Принимает:** path: `id`
**Возвращает:** статистика организации
**Пример:**
```bash
curl http://localhost:8081/api/v1/organizations/1/stats \
  -H "Authorization: Bearer <token>"
```

#### GET /api/v1/organizations/{id}
**Принимает:** path: `id`
**Возвращает:** детали организации
**Пример:**
```bash
curl http://localhost:8081/api/v1/organizations/1 \
  -H "Authorization: Bearer <token>"
```

#### PUT /api/v1/organizations/{id}
**Принимает:** path: `id`, body: данные для обновления
**Возвращает:** обновлённая организация
**Пример:**
```bash
curl -X PUT http://localhost:8081/api/v1/organizations/1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Обновлённое название"}'
```

#### DELETE /api/v1/organizations/{id}
**Принимает:** path: `id`
**Возвращает:** HTTP 204
**Пример:**
```bash
curl -X DELETE http://localhost:8081/api/v1/organizations/1 \
  -H "Authorization: Bearer <token>"
```

---

### DEPARTMENTS (User Portal)

#### GET /api/v1/departments
**Принимает:** query: `parent_id`, `include_all`, `tree`
**Возвращает:** список отделов
**Пример:**
```bash
curl "http://localhost:8081/api/v1/departments?tree=true" \
  -H "Authorization: Bearer <token>"
```

---

### STATIC

#### GET /uploads/{path}
**Принимает:** path: путь к файлу
**Возвращает:** статический файл (аватары, загрузки)

#### GET /{path}
**Принимает:** любой путь
**Возвращает:** React SPA (index.html) для фронтенда

---

## Итог

| Портал | Кол-во ручек |
|--------|-------------|
| Managing Portal (8080) | ~52 |
| User Portal (8081) | ~54 |
| **Итого** | **~106** |

### Аутентификация
Все защищённые ручки требуют заголовок:
```
Authorization: Bearer <JWT_TOKEN>
```
Токен получается через `POST /api/v1/auth/login`.

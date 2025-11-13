# Логика записи LiveKit конференций

Этот документ описывает обновленную логику записи аудио/видео в системе Recontext.online.

## Типы записи

### 1. Композитная запись комнаты (Room Composite Egress)

**Назначение:** Создание единого файла со всей конференцией.

**Когда запускается:** При старте комнаты (`room_started` event)

**Условия запуска:**
- Если `needs_video_record = true` → записывается аудио + видео
- Если `needs_audio_record = true` (без видео) → записывается только аудио
- Если `needs_video_record = true`, то `needs_audio_record` автоматически устанавливается в `true`

**Результат:** Один файл MP4 (с видео) или один аудиофайл (без видео) со всем контентом комнаты.

### 2. Покадровая запись треков (Track Composite Egress)

**Назначение:** Запись отдельных аудио-треков каждого участника для последующей транскрибации.

**Когда запускается:** При публикации аудио-трека (`track_published` event)

**Условия запуска:**
- `needs_transcription = true` для митинга
- Трек является аудио-треком:
  - `source = "MICROPHONE"` ИЛИ
  - `source = "SCREEN_SHARE_AUDIO"` ИЛИ
  - `mime_type` начинается с `"audio/"`

**Результат:** Отдельный аудиофайл для каждого участника, который можно отправить на транскрибацию.

## Настройки митинга

### Поля в модели Meeting

```go
type Meeting struct {
    // ...
    NeedsVideoRecord   bool  // Требуется ли видеозапись (композитная запись с видео)
    NeedsAudioRecord   bool  // Требуется ли аудиозапись (композитная запись аудио)
    NeedsTranscription bool  // Требуется ли транскрибация (запись отдельных треков)
    // ...
}
```

### Логика на фронтенде

1. **Галочка "Видеозапись"** (`needs_video_record`)
   - Если включена → автоматически включается и становится серой галочка "Аудиозапись"
   - Запускается композитная запись комнаты с аудио + видео

2. **Галочка "Аудиозапись"** (`needs_audio_record`)
   - Если включена (без видео) → запускается композитная запись только аудио
   - Если включено видео → эта галочка автоматически включается и становится неактивной

3. **Галочка "Транскрибация"** (`needs_transcription`)
   - Если включена → для каждого аудио-трека участника запускается отдельный egress
   - Независима от галочек видео/аудио записи
   - Позволяет получать отдельные аудиофайлы для распознавания речи

## Примеры конфигураций

### Пример 1: Только видеозапись всей комнаты
```json
{
  "needs_video_record": true,
  "needs_audio_record": false,
  "needs_transcription": false
}
```
**Результат:** 1 файл MP4 с видео и аудио всей конференции

---

### Пример 2: Только аудиозапись всей комнаты
```json
{
  "needs_video_record": false,
  "needs_audio_record": true,
  "needs_transcription": false
}
```
**Результат:** 1 аудиофайл со всей конференцией

---

### Пример 3: Только транскрибация (отдельные треки)
```json
{
  "needs_video_record": false,
  "needs_audio_record": false,
  "needs_transcription": true
}
```
**Результат:** N аудиофайлов (по одному на каждого участника)

---

### Пример 4: Видеозапись + транскрибация
```json
{
  "needs_video_record": true,
  "needs_audio_record": false,
  "needs_transcription": true
}
```
**Результат:**
- 1 файл MP4 с видео и аудио всей конференции
- N аудиофайлов (по одному на каждого участника)

---

### Пример 5: Всё включено
```json
{
  "needs_video_record": true,
  "needs_audio_record": true,
  "needs_transcription": true
}
```
**Результат:**
- 1 файл MP4 с видео и аудио всей конференции
- N аудиофайлов (по одному на каждого участника)

## Хранение настроек в памяти

Для эффективной работы настройки транскрибации хранятся в памяти на время жизни комнаты:

```go
type RoomSettings struct {
    NeedsTranscription bool
}

// Сохраняется при room_started
mp.setRoomSettings(roomSID, &RoomSettings{
    NeedsTranscription: needsTranscription,
})

// Используется при track_published
roomSettings := mp.getRoomSettings(track.RoomSID)

// Удаляется при room_finished
mp.deleteRoomSettings(roomSID)
```

## Миграция базы данных

Для добавления поля `needs_transcription` выполните:

```bash
psql -U recontext -d recontext -f migrations/add_needs_transcription.sql
```

Или перезапустите приложение - GORM AutoMigrate автоматически добавит новое поле.

## API

### Создание митинга с транскрибацией

```bash
POST /api/v1/meetings
{
  "title": "Совещание с транскрибацией",
  "scheduled_at": "2025-01-15T10:00:00Z",
  "duration": 60,
  "type": "conference",
  "subject_id": "...",
  "needs_video_record": true,
  "needs_audio_record": true,
  "needs_transcription": true,
  "participant_ids": ["..."],
  "department_ids": ["..."]
}
```

### Обновление настроек записи

```bash
PATCH /api/v1/meetings/{id}
{
  "needs_transcription": true
}
```

## Логирование

Система выводит подробные логи для отладки всех событий:

### События комнаты
- 🏠 `room_started` - создание комнаты и запуск композитного egress
- 🏁 `room_finished` - завершение комнаты и остановка всех egress

### События участников
- 👤 `participant_joined` - подключение участника к комнате
- 👋 `participant_left` - выход участника из комнаты

### События треков
- 🎬 `track_published` - публикация трека (аудио/видео)
- 🔴 `track_unpublished` - отмена публикации трека
- 🎤 `Audio track detected` - обнаружен аудио-трек для транскрибации
- 🛑 `Stopping track egress` - остановка egress для трека

### Операции с egress
- 🚀 `Starting track composite egress` - запуск egress для трека
- 🚀 `Starting room composite egress` - запуск egress для комнаты
- ✅ `Track egress started successfully` - egress трека запущен
- ✅ `Room egress started successfully` - egress комнаты запущен
- ✅ `Stopped track egress` - egress трека остановлен
- ✅ `Stopped room egress` - egress комнаты остановлен

### Информационные сообщения
- 📌 Детали событий (SID, имена, типы)
- 💾 Операции с базой данных
- 🔍 Поиск данных в базе
- 🗑️ Очистка данных из памяти
- ℹ️ Причины пропуска операций
- ⚠️ Предупреждения о недостающих данных

## Жизненный цикл трека с транскрибацией

```
track_published (audio) → Запуск Track Egress
           ↓
      [Запись идет]
           ↓
track_unpublished → Остановка Track Egress
```

Важно: при событии `track_unpublished` система:
1. Находит трек в базе данных по SID
2. Проверяет наличие egress ID
3. Останавливает egress через LiveKit API
4. Помечает трек как "unpublished" в БД

## Troubleshooting

### Egress не запускается для аудио-треков

**Проблема:** Логи показывают "Track egress skipped - transcription not enabled"

**Решение:**
1. Убедитесь, что `needs_transcription = true` в настройках митинга
2. Проверьте, что настройки сохранились: смотрите лог "Room settings saved: NeedsTranscription=true"
3. Проверьте, что трек действительно аудио (Source=MICROPHONE или MimeType=audio/*)

### Не создается композитная запись комнаты

**Проблема:** Нет записи всей конференции

**Решение:**
1. Убедитесь, что `needs_video_record = true` ИЛИ `needs_audio_record = true`
2. Проверьте настройки LiveKit Egress Client (переменные окружения)
3. Смотрите логи: "Starting room composite egress..."

### Egress не останавливается при unpublish

**Проблема:** Логи показывают "No egress ID found for track"

**Решение:**
1. Проверьте, что трек был сохранен с egress ID при публикации
2. Убедитесь, что track_published событие обработалось успешно
3. Проверьте логи: "Track saved successfully" и "Track updated with egress ID"

### Треки не записываются после выхода участника

**Проблема:** После `participant_left` треки перестают записываться

**Ожидаемое поведение:**
- Событие `participant_left` только обновляет статус участника в БД
- Треки останавливаются отдельно событием `track_unpublished`
- При корректной работе LiveKit сначала придет `track_unpublished`, потом `participant_left`

# Поток обработки композитного видео

## Обзор

После завершения встречи система автоматически создаёт композитное видео, объединяя треки всех участников.

## Этапы обработки

### 1. Триггер обработки
Обработка запускается когда:
- Все треки встречи завершили транскрибацию
- Вызывается `VideoPostProcessor.ProcessMeetingVideo(roomSID)`

### 2. Объединение HLS треков (Шаг 0)
**Файл:** `cmd/managing-portal/video_postprocessor.go` → `combineHLSTracks()`

**Действия:**
- Скачивает HLS плейлисты (.m3u8) и сегменты (.ts) из MinIO
- Объединяет сегменты каждого трека в MP4 файл с помощью ffmpeg
- Загружает MP4 файлы обратно в MinIO: `{meetingID}_{roomSID}/tracks/TR_xxxxx.mp4`
- Обновляет базу данных:
  - Таблица: `livekit_egress_recordings`
  - Поле: `playlist_url` → URL на MP4 файл
  - Поле: `status` → "ended"

**Результат:**
```
recontext/
  └── {meetingID}_{roomSID}/
      └── tracks/
          ├── TR_VCoyexbMM3nSRm.mp4  (видео трек участника 1)
          └── TR_AMcJtxis6SfUzT.mp4  (аудио трек участника 1)
```

### 3. Скачивание треков (Шаг 2-3)
**Файл:** `cmd/managing-portal/video_postprocessor.go` → `downloadTracks()`

**Действия:**
- Получает информацию о треках из базы (`getTrackInfoForMerge`)
- Скачивает MP4 файлы из MinIO во временную директорию

**SQL запрос для получения треков:**
```sql
SELECT
    livekit_tracks.id as track_id,
    livekit_tracks.participant_sid,
    COALESCE(livekit_participants.name, 'Unknown') as participant_name,
    livekit_tracks.type,
    livekit_egress_recordings.playlist_url,  -- URL на MP4 файл
    livekit_tracks.published_at,
    COALESCE(livekit_tracks.transcription_duration, 0) as duration
FROM livekit_tracks
LEFT JOIN livekit_participants ON livekit_tracks.participant_sid = livekit_participants.sid
LEFT JOIN livekit_egress_recordings ON livekit_egress_recordings.track_sid = livekit_tracks.sid
WHERE livekit_tracks.room_sid = ?
  AND livekit_tracks.type IN ('audio', 'video')
  AND livekit_egress_recordings.playlist_url IS NOT NULL
  AND livekit_egress_recordings.playlist_url != ''
  AND livekit_egress_recordings.status = 'ended'
```

### 4. Анализ спикеров (Шаг 4)
**Файл:** `cmd/managing-portal/video_postprocessor.go` → `analyzeSpeakerActivity()`

**Действия:**
- Анализирует аудио треки для определения активного спикера
- Создаёт временную линию переключения между спикерами
- Сохраняет в таблицу `livekit_speaker_timelines`

### 5. Генерация сводок (Шаг 4.5)
**Файл:** `cmd/managing-portal/video_postprocessor.go` → `generateMeetingSummary()`

**Действия:**
- Генерирует сводку встречи на русском и английском языках
- Сохраняет в таблицу `meeting_summaries`

### 6. Объединение видео (Шаг 5)
**Файл:** `pkg/video/processor.go` → `MergeTracksPiPWithSpeakerSwitch()`

**Действия:**
- Объединяет все треки в один видеофайл
- Использует режим picture-in-picture с динамическим переключением активного спикера
- Создаёт временный MP4 файл: `/tmp/merged_{uuid}.mp4`

### 7. Конвертация в HLS (Шаг 6)
**Файл:** `pkg/video/processor.go` → `ConvertToHLSWithCustomNames()`

**Действия:**
- Конвертирует объединённое MP4 видео в HLS формат
- Создаёт плейлист: `composite.m3u8`
- Разбивает на сегменты: `composite_00001.ts`, `composite_00002.ts`, ...
- Длина сегмента: 10 секунд
- Формат нумерации: 5 цифр с ведущими нулями

**FFmpeg команда:**
```bash
ffmpeg -i merged.mp4 \
  -c:v libx264 \
  -c:a aac \
  -b:a 128k \
  -hls_time 10 \
  -hls_list_size 0 \
  -hls_segment_filename "composite_%05d.ts" \
  -f hls \
  composite.m3u8
```

### 8. Загрузка в MinIO (Шаг 7)
**Файл:** `pkg/video/storage.go` → `UploadDirectory()`

**Действия:**
- Загружает все файлы HLS в MinIO
- Путь: `{meetingID}_{roomSID}/`
- Файлы:
  - `composite.m3u8` - плейлист
  - `composite_00001.ts` - сегмент 1
  - `composite_00002.ts` - сегмент 2
  - ... и т.д.

**Результирующая структура в MinIO:**
```
recontext/
  └── b5090ccd-2b46-4fea-8879-fbed44a3a09e_RM_sJJ72Y37fFpW/
      ├── composite.m3u8           # Плейлист композитного видео
      ├── composite_00001.ts       # Сегмент 1
      ├── composite_00002.ts       # Сегмент 2
      ├── composite_00003.ts       # Сегмент 3
      ├── ...
      └── tracks/                  # Папка с индивидуальными треками
          ├── TR_VCoyexbMM3nSRm.mp4
          └── TR_AMcJtxis6SfUzT.mp4
```

### 9. Обновление базы данных (Шаг 9)
**Файл:** `cmd/managing-portal/video_postprocessor.go` → `updateMeetingWithPlaylist()`

**Действия:**
- Обновляет таблицу `meetings`:
  ```sql
  UPDATE meetings
  SET video_playlist_url = 'http://192.168.5.153:9000/recontext/{meetingID}_{roomSID}/composite.m3u8',
      updated_at = NOW()
  WHERE id = {meetingID}
  ```

- Обновляет таблицу `livekit_egress_recordings`:
  ```sql
  UPDATE livekit_egress_recordings
  SET playlist_url = 'http://192.168.5.153:9000/recontext/{meetingID}_{roomSID}/composite.m3u8',
      updated_at = NOW()
  WHERE room_sid = {roomSID}
    AND type = 'room_composite'
  ```

**Поля в базе данных:**
- `meetings.video_playlist_url` - URL на composite.m3u8
- `livekit_egress_recordings.playlist_url` - URL на composite.m3u8 (для room_composite)
- `livekit_egress_recordings.playlist_url` - URL на MP4 (для track_composite)

## Формат URL

### Публичные URL (для доступа извне Docker)
```
http://192.168.5.153:9000/recontext/{meetingID}_{roomSID}/composite.m3u8
http://192.168.5.153:9000/recontext/{meetingID}_{roomSID}/composite_00001.ts
http://192.168.5.153:9000/recontext/{meetingID}_{roomSID}/tracks/TR_xxxxx.mp4
```

### Внутренние URL (внутри Docker сети)
```
http://minio:9000/recontext/{meetingID}_{roomSID}/composite.m3u8
```

## Переменные окружения

### MinIO настройки
- `MINIO_ENDPOINT` - внутренний endpoint для подключения (например, `minio:9000`)
- `MINIO_PUBLIC_ENDPOINT` - публичный endpoint для формирования URL (например, `192.168.5.153:9000`)
- `MINIO_ACCESS_KEY` - access key
- `MINIO_SECRET_KEY` - secret key
- `MINIO_BUCKET` - имя bucket (по умолчанию `recontext`)
- `MINIO_USE_SSL` - использовать HTTPS (`true`/`false`)

## Таблицы базы данных

### meetings
| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID | ID встречи |
| video_playlist_url | TEXT | URL на composite.m3u8 |
| updated_at | TIMESTAMP | Время последнего обновления |

### livekit_egress_recordings
| Поле | Тип | Описание |
|------|-----|----------|
| id | UUID | ID записи |
| room_sid | VARCHAR | SID комнаты |
| track_sid | VARCHAR | SID трека (для track_composite) |
| type | VARCHAR | Тип: `room_composite` или `track_composite` |
| playlist_url | TEXT | URL на .m3u8 (room_composite) или .mp4 (track_composite) |
| status | VARCHAR | Статус: `ended` |
| updated_at | TIMESTAMP | Время последнего обновления |

### livekit_tracks
| Поле | Тип | Описание |
|------|-----|----------|
| sid | VARCHAR | SID трека |
| room_sid | VARCHAR | SID комнаты |
| participant_sid | VARCHAR | SID участника |
| type | VARCHAR | Тип: `audio` или `video` |
| transcription_duration | FLOAT | Длительность в секундах |
| updated_at | TIMESTAMP | Время последнего обновления |

## Проверка результата

### Проверка файлов в MinIO
```bash
# Запустить скрипт проверки
go run cmd/track-processor/check_composite.go
```

### Проверка базы данных
```sql
-- Проверить URL композитного видео
SELECT
    m.id,
    m.title,
    m.video_playlist_url,
    m.updated_at
FROM meetings m
WHERE m.video_playlist_url LIKE '%composite.m3u8%';

-- Проверить записи в egress_recordings
SELECT
    id,
    room_sid,
    type,
    playlist_url,
    status,
    updated_at
FROM livekit_egress_recordings
WHERE type = 'room_composite'
ORDER BY updated_at DESC
LIMIT 10;
```

## Отладка

### Логи обработки
```bash
docker-compose -f docker-compose-mac.yml logs managing-portal --tail=100 -f
```

### Ключевые логи:
```
🎬 Combining HLS tracks for meeting {meetingID}, room {roomSID}
📊 Processing 2 HLS tracks
✅ Track TR_xxxxx: audio, 1.07 MB, 65.71 seconds
📤 Uploaded to: http://192.168.5.153:9000/recontext/{meetingID}_{roomSID}/tracks/TR_xxxxx.mp4
✅ Combined 2/2 tracks successfully

✅ Video merged successfully: /tmp/merged_{uuid}.mp4
🎬 Converting video to HLS format with custom names
✅ HLS conversion completed: /tmp/hls_{uuid}/composite.m3u8
✅ Uploaded X files to S3
📍 Composite playlist URL: http://192.168.5.153:9000/recontext/{meetingID}_{roomSID}/composite.m3u8
✅ Database updated with playlist URL for meeting {meetingID}
```

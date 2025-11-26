# Video Package

Пакет для работы с видео и аудио треками: скачивание, объединение и загрузка в MinIO/S3.

## TrackCombiner

Класс `TrackCombiner` предназначен для автоматической обработки HLS треков из MinIO/S3:

1. Сканирует треки по ID встречи и комнаты
2. Скачивает HLS плейлисты (.m3u8) и сегменты (.ts)
3. Объединяет сегменты в единые файлы с правильными форматами:
   - `TR_VC*` (видео треки) → `.mp4` файлы (H.264 видео + AAC аудио)
   - `TR_AM*` (аудио треки) → `.webm` файлы (Opus аудио)
4. Загружает обратно в MinIO/S3 с правильным Content-Type
5. Автоматически определяет тип трека (audio/video) и продолжительность

### Использование

#### Базовый пример

```go
import "Recontext.online/pkg/video"

// Конфигурация
config := video.TrackCombinerConfig{
    Endpoint:        "api.storage.recontext.online",
    AccessKeyID:     "your-access-key",
    SecretAccessKey: "your-secret-key",
    BucketName:      "recontext",
    UseSSL:          true,
    WorkDir:         "./temp-tracks", // опционально
}

// Создаем combiner
combiner, err := video.NewTrackCombiner(config)
if err != nil {
    log.Fatal(err)
}
defer combiner.CleanupAll()

ctx := context.Background()

// Обрабатываем все треки комнаты
tracks, err := combiner.CombineTracksByRoom(ctx, meetingID, roomSID)
if err != nil {
    log.Fatal(err)
}

// Обрабатываем результаты
for _, track := range tracks {
    if track.Error != nil {
        log.Printf("Track %s failed: %v", track.TrackID, track.Error)
        continue
    }

    log.Printf("Track: %s, Type: %s, Size: %d, Duration: %.2f sec",
        track.TrackID, track.Type, track.Size, track.Duration)

    // Загружаем в MinIO
    url, err := combiner.UploadCombinedTrack(ctx, &track, meetingID, roomSID)
    if err != nil {
        log.Printf("Upload failed: %v", err)
        continue
    }

    log.Printf("Uploaded to: %s", url)

    // Очищаем временные файлы
    combiner.Cleanup(track.TrackID)
}
```

#### Обработка одного трека

```go
// Обрабатываем конкретный трек
track, err := combiner.CombineSingleTrack(ctx, meetingID, roomSID, trackID)
if err != nil {
    log.Fatal(err)
}

log.Printf("Track combined: %s (%s, %.2f MB, %.2f sec)",
    track.TrackID, track.Type,
    float64(track.Size)/(1024*1024), track.Duration)
```

### Структура данных

#### TrackCombinerConfig

```go
type TrackCombinerConfig struct {
    Endpoint        string // MinIO/S3 endpoint для подключения (например, minio:9000)
    PublicEndpoint  string // Публичный endpoint для формирования URL (например, api.storage.recontext.online)
    AccessKeyID     string // Access key
    SecretAccessKey string // Secret key
    BucketName      string // Bucket name
    UseSSL          bool   // Использовать HTTPS для публичных URL
    WorkDir         string // Рабочая директория (опционально)
}
```

#### CombinedTrack

```go
type CombinedTrack struct {
    TrackID   string  // ID трека (например, TR_VCoyexbMM3nSRm)
    LocalPath string  // Путь к объединенному файлу
    Size      int64   // Размер файла в байтах
    Duration  float64 // Продолжительность в секундах
    Type      string  // "audio" или "video"
    Error     error   // Ошибка, если произошла
}
```

### Методы

#### NewTrackCombiner

```go
func NewTrackCombiner(config TrackCombinerConfig) (*TrackCombiner, error)
```

Создает новый экземпляр TrackCombiner с указанной конфигурацией.

#### CombineTracksByRoom

```go
func (tc *TrackCombiner) CombineTracksByRoom(ctx context.Context, meetingID, roomSID string) ([]CombinedTrack, error)
```

Обрабатывает все треки для указанной встречи и комнаты.

**Параметры:**
- `ctx` - контекст для отмены операции
- `meetingID` - UUID встречи
- `roomSID` - SID комнаты LiveKit (например, RM_4qJSgKmcQmPW)

**Возвращает:**
- Массив `CombinedTrack` с результатами для каждого трека
- Ошибку, если не удалось отсканировать треки

#### CombineSingleTrack

```go
func (tc *TrackCombiner) CombineSingleTrack(ctx context.Context, meetingID, roomSID, trackID string) (*CombinedTrack, error)
```

Обрабатывает конкретный трек.

**Параметры:**
- `ctx` - контекст
- `meetingID` - UUID встречи
- `roomSID` - SID комнаты
- `trackID` - ID трека (например, TR_VCoyexbMM3nSRm)

#### UploadCombinedTrack

```go
func (tc *TrackCombiner) UploadCombinedTrack(ctx context.Context, track *CombinedTrack, meetingID, roomSID string) (string, error)
```

Загружает объединенный трек обратно в MinIO/S3.

**Возвращает:**
- URL для доступа к загруженному файлу
- Ошибку, если загрузка не удалась

#### Cleanup

```go
func (tc *TrackCombiner) Cleanup(trackID string) error
```

Удаляет временные файлы конкретного трека.

#### CleanupAll

```go
func (tc *TrackCombiner) CleanupAll() error
```

Удаляет всю рабочую директорию со всеми временными файлами.

### Требования

1. **ffmpeg** - для объединения HLS сегментов
   ```bash
   # macOS
   brew install ffmpeg

   # Ubuntu/Debian
   apt-get install ffmpeg

   # Alpine (Docker)
   apk add ffmpeg
   ```

2. **ffprobe** - для получения информации о медиафайлах (обычно входит в ffmpeg)

3. **MinIO/S3 доступ** - валидные credentials для доступа к хранилищу

### Формат данных в MinIO

Ожидаемая структура файлов в MinIO:

```
bucket/
  meetingID_roomSID/
    tracks/
      TR_xxxxx.m3u8          # Основной плейлист
      TR_xxxxx-live.m3u8     # Живой плейлист (игнорируется)
      TR_xxxxx_00000.ts      # Сегменты
      TR_xxxxx_00001.ts
      TR_xxxxx_00002.ts
      ...
```

### Примеры использования

#### В веб-сервисе

```go
func handleCombineTracks(w http.ResponseWriter, r *http.Request) {
    meetingID := r.URL.Query().Get("meeting_id")
    roomSID := r.URL.Query().Get("room_sid")

    combiner, err := video.NewTrackCombiner(config)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer combiner.CleanupAll()

    tracks, err := combiner.CombineTracksByRoom(r.Context(), meetingID, roomSID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Загружаем треки в MinIO и возвращаем URLs
    var urls []string
    for _, track := range tracks {
        if track.Error != nil {
            continue
        }
        url, err := combiner.UploadCombinedTrack(r.Context(), &track, meetingID, roomSID)
        if err == nil {
            urls = append(urls, url)
        }
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "tracks": urls,
    })
}
```

#### В background worker

```go
func processCompletedMeeting(meetingID, roomSID string) error {
    config := video.TrackCombinerConfig{
        Endpoint:        os.Getenv("MINIO_ENDPOINT"),
        AccessKeyID:     os.Getenv("MINIO_ACCESS_KEY"),
        SecretAccessKey: os.Getenv("MINIO_SECRET_KEY"),
        BucketName:      os.Getenv("MINIO_BUCKET"),
        UseSSL:          true,
    }

    combiner, err := video.NewTrackCombiner(config)
    if err != nil {
        return err
    }
    defer combiner.CleanupAll()

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
    defer cancel()

    tracks, err := combiner.CombineTracksByRoom(ctx, meetingID, roomSID)
    if err != nil {
        return err
    }

    // Сохраняем результаты в базу данных
    for _, track := range tracks {
        if track.Error != nil {
            log.Printf("Track %s failed: %v", track.TrackID, track.Error)
            continue
        }

        url, err := combiner.UploadCombinedTrack(ctx, &track, meetingID, roomSID)
        if err != nil {
            log.Printf("Upload failed for %s: %v", track.TrackID, err)
            continue
        }

        // Обновляем базу данных
        db.UpdateTrackURL(track.TrackID, url)

        // Очищаем временные файлы
        combiner.Cleanup(track.TrackID)
    }

    return nil
}
```

### Обработка ошибок

Класс возвращает ошибки на уровне трека, что позволяет продолжить обработку остальных треков:

```go
tracks, err := combiner.CombineTracksByRoom(ctx, meetingID, roomSID)
if err != nil {
    // Критическая ошибка (не удалось отсканировать треки)
    log.Fatal(err)
}

// Проверяем ошибки на уровне треков
for _, track := range tracks {
    if track.Error != nil {
        // Трек не удалось обработать, но остальные могут быть успешны
        log.Printf("Track %s error: %v", track.TrackID, track.Error)
        continue
    }
    // Обрабатываем успешный трек
}
```

### Performance

- **Параллельная обработка**: Треки обрабатываются последовательно, но можно легко добавить горутины
- **Память**: Временные файлы хранятся на диске, не в памяти
- **Скорость**: Зависит от размера треков и скорости сети/диска
- **Продолжительность**: ~1-2 минуты на трек (скачивание + ffmpeg)

### TODO

- [ ] Добавить параллельную обработку треков
- [ ] Поддержка resume для прерванных операций
- [ ] Прогресс-бар для длительных операций
- [ ] Кэширование уже обработанных треков
- [ ] Поддержка других форматов (не только HLS)

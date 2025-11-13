# Реализация просмотра записей встреч

## Описание задачи

Необходимо реализовать функциональность просмотра записей встреч:
1. Отображение списка комнат (egress recordings) для встречи
2. Показ статусов, дат начала/завершения
3. Внутри комнаты - треки (общий + отдельные по участникам)
4. Плееры для аудио/видео воспроизведения
5. Прокси к MinIO для получения файлов через playlist (HLS)

## Архитектура

### 1. Хранение файлов

Файлы записей хранятся в MinIO в формате HLS:
- **Room composite**: `{meeting_id}/composite.m3u8` - общая запись всей комнаты
- **Track recordings**: `{meeting_id}/tracks/{track_id}.m3u8` - отдельные треки участников
- **Segments**: `*.ts` файлы - сегменты видео/аудио

### 2. Backend API (user-portal)

#### Endpoints:

```
GET /api/v1/meetings/{id}/recordings
- Возвращает список всех записей встречи
- Требует: участник встречи или админ
- Response: []RecordingInfo

GET /api/v1/recordings/{egress_id}/playlist
- Прокси к m3u8 плейлисту в MinIO
- Возвращает playlist с исправленными URL на сегменты

GET /api/v1/recordings/{egress_id}/segment/{filename}
- Прокси к TS сегменту в MinIO
- Возвращает бинарные данные с правильными заголовками
```

#### Структура данных:

```go
type RecordingInfo struct {
    ID            string   // EgressID
    Type          string   // "room" или "track"
    Status        string   // "active", "completed", "failed"
    StartedAt     string   // ISO 8601 timestamp
    EndedAt       *string  // ISO 8601 timestamp (nullable)
    PlaylistURL   string   // URL к прокси плейлиста
    ParticipantID *string  // ID участника (для треков)
    TrackID       *string  // ID трека
    Participant   *User    // Информация об участнике
}
```

### 3. Database Queries

Нужно добавить методы в `LiveKitRepository`:

```go
// GetRoomsByName - получить все комнаты по имени (meeting ID)
func (r *LiveKitRepository) GetRoomsByName(roomName string) ([]*LiveKitRoom, error)

// GetTracksByRoomSID - получить все треки комнаты
func (r *LiveKitRepository) GetTracksByRoomSID(roomSID string) ([]*LiveKitTrack, error)

// GetEgressByID - получить egress запись по ID
func (r *LiveKitRepository) GetEgressByID(egressID string) (*LiveKitEgress, error)
```

### 4. MinIO Proxy Implementation

Создать новый файл `cmd/user-portal/handlers_minio_proxy.go`:

```go
// Прокси к MinIO с авторизацией
func (up *UserPortal) getPlaylistHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Извлечь egress_id из URL
    // 2. Проверить права доступа (пользователь - участник встречи)
    // 3. Получить файл из MinIO
    // 4. Переписать URL сегментов на наш прокси
    // 5. Вернуть измененный playlist
}

func (up *UserPortal) getSegmentHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Извлечь egress_id и filename из URL
    // 2. Проверить права доступа
    // 3. Получить файл из MinIO
    // 4. Вернуть с заголовками:
    //    Content-Type: video/mp2t
    //    Access-Control-Allow-Origin: *
}
```

### 5. Frontend

#### Страница списка записей

`front/user-portal/src/pages/MeetingRecordings.tsx`:

```typescript
interface Recording {
  id: string;
  type: 'room' | 'track';
  status: string;
  started_at: string;
  ended_at?: string;
  playlist_url: string;
  participant_id?: string;
  track_id?: string;
  participant?: User;
}

export default function MeetingRecordings() {
  const { meetingId } = useParams();
  const [recordings, setRecordings] = useState<Recording[]>([]);
  const [selectedRecording, setSelectedRecording] = useState<Recording | null>(null);
  const [showTracksOnly, setShowTracksOnly] = useState(false);

  // Загрузка записей
  useEffect(() => {
    fetchRecordings(meetingId);
  }, [meetingId]);

  return (
    <div>
      <h1>Recordings for Meeting</h1>

      {/* Переключатель: Room Composite / Individual Tracks */}
      <div className="recording-type-selector">
        <button onClick={() => setShowTracksOnly(false)}>
          Room Composite (All)
        </button>
        <button onClick={() => setShowTracksOnly(true)}>
          Individual Tracks
        </button>
      </div>

      {/* Список записей */}
      <div className="recordings-list">
        {recordings
          .filter(r => !showTracksOnly || r.type === 'track')
          .map(recording => (
            <RecordingCard
              key={recording.id}
              recording={recording}
              onSelect={() => setSelectedRecording(recording)}
            />
          ))}
      </div>

      {/* Плеер */}
      {selectedRecording && (
        <HLSPlayer
          src={selectedRecording.playlist_url}
          autoplay
        />
      )}
    </div>
  );
}
```

#### HLS Player Component

`front/user-portal/src/components/HLSPlayer.tsx`:

Использовать библиотеку `hls.js`:

```bash
npm install hls.js
```

```typescript
import Hls from 'hls.js';
import { useEffect, useRef } from 'react';

interface HLSPlayerProps {
  src: string;
  autoplay?: boolean;
}

export default function HLSPlayer({ src, autoplay }: HLSPlayerProps) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const hlsRef = useRef<Hls | null>(null);

  useEffect(() => {
    if (!videoRef.current) return;

    if (Hls.isSupported()) {
      const hls = new Hls();
      hls.loadSource(src);
      hls.attachMedia(videoRef.current);
      hlsRef.current = hls;

      return () => {
        hls.destroy();
      };
    } else if (videoRef.current.canPlayType('application/vnd.apple.mpegurl')) {
      // Native HLS support (Safari)
      videoRef.current.src = src;
    }
  }, [src]);

  return (
    <video
      ref={videoRef}
      controls
      autoPlay={autoplay}
      style={{ width: '100%', maxWidth: '800px' }}
    />
  );
}
```

## Порядок реализации

### Шаг 1: Repository методы
1. Добавить методы в `pkg/database/livekit_repository.go`
2. Протестировать запросы к базе

### Шаг 2: Backend API
1. Создать `cmd/user-portal/handlers_recordings.go` (уже создан базовый)
2. Добавить методы в `cmd/user-portal/handlers_minio_proxy.go`
3. Зарегистрировать роуты в `cmd/user-portal/main.go`

### Шаг 3: MinIO Integration
1. Настроить MinIO клиент
2. Реализовать прокси для плейлистов
3. Реализовать прокси для сегментов
4. Переписать URLs в плейлистах

### Шаг 4: Frontend
1. Установить `hls.js`
2. Создать HLSPlayer компонент
3. Создать страницу MeetingRecordings
4. Добавить роут и навигацию

## Конфигурация MinIO

В `.env` или переменных окружения:

```bash
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=recontext
MINIO_USE_SSL=false
```

## Пример использования

1. Пользователь открывает страницу встречи
2. Видит кнопку "View Recordings"
3. Переходит на страницу записей
4. Видит список:
   - Room Composite Recording (started: 14:30, ended: 15:45)
   - Track: John Doe (started: 14:30, ended: 15:45)
   - Track: Jane Smith (started: 14:35, ended: 15:40)
5. Переключает между "All" и "Individual Tracks"
6. Кликает на запись → открывается плеер
7. Видео/аудио воспроизводится через HLS

## Безопасность

- Проверка прав: только участники встречи могут просматривать записи
- Прокси обязателен: прямой доступ к MinIO закрыт
- Временные токены (опционально): можно добавить signed URLs с TTL

## Дополнительные улучшения

1. **Скачивание записей**: кнопка Download для сохранения локально
2. **Субтитры**: если есть транскрипция, показывать как субтитры
3. **Thumbnail preview**: генерировать превью для быстрой навигации
4. **Метаданные**: показывать длительность, размер файла, качество
5. **Фильтрация**: по дате, участнику, типу записи

## Статус реализации

- [ ] Repository методы
- [x] Базовый handler для списка записей
- [ ] MinIO прокси handlers
- [ ] Frontend компоненты
- [ ] HLS player интеграция
- [ ] Тестирование

## Следующие шаги

1. Реализовать `GetRoomsByName` и `GetTracksByRoomSID` в repository
2. Создать MinIO proxy handlers
3. Протестировать получение плейлистов
4. Разработать frontend UI

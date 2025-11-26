# Storage Package

Универсальный пакет для работы с MinIO/S3 хранилищем во всём проекте Recontext.

## MinIOClient

Централизованный клиент для работы с MinIO, который используется во всех частях проекта.

### Особенности

- Единая точка конфигурации MinIO
- Автоматическое чтение переменных окружения `MINIO_*`
- Поддержка публичного и внутреннего endpoint
- Методы для загрузки/скачивания файлов и директорий
- Автоматическое определение Content-Type
- Формирование публичных URL

### Использование

#### Создание клиента из переменных окружения

```go
import "Recontext.online/pkg/storage"

// Создаем клиент из переменных окружения
client, err := storage.NewMinIOClientFromEnv()
if err != nil {
    log.Fatal(err)
}
```

#### Создание клиента с явной конфигурацией

```go
config := storage.MinIOConfig{
    Endpoint:       "minio:9000",              // Внутренний endpoint
    PublicEndpoint: "192.168.5.153:9000",      // Публичный endpoint для URL
    AccessKey:      "minioadmin",
    SecretKey:      "secret",
    Bucket:         "recontext",
    UseSSL:         false,
}

client, err := storage.NewMinIOClient(config)
if err != nil {
    log.Fatal(err)
}
```

### Методы

#### UploadFile - Загрузка файла

```go
ctx := context.Background()

// Загружаем файл
url, err := client.UploadFile(ctx, "/local/path/file.mp4", "meeting_123/file.mp4")
if err != nil {
    log.Fatal(err)
}

// url будет: http://192.168.5.153:9000/recontext/meeting_123/file.mp4
fmt.Println("Uploaded:", url)
```

#### DownloadFile - Скачивание файла

```go
ctx := context.Background()

// Скачиваем файл из MinIO
err := client.DownloadFile(ctx, "meeting_123/file.mp4", "/local/path/file.mp4")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Downloaded successfully")
```

#### UploadDirectory - Загрузка директории

```go
ctx := context.Background()

// Загружаем всю директорию
urls, err := client.UploadDirectory(ctx, "/local/dir", "meeting_123")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Uploaded %d files\n", len(urls))
```

#### DeleteFile - Удаление файла

```go
ctx := context.Background()

err := client.DeleteFile(ctx, "meeting_123/file.mp4")
if err != nil {
    log.Fatal(err)
}
```

#### GetPublicURL - Получение публичного URL

```go
// Формируем публичный URL для объекта
url := client.GetPublicURL("meeting_123/file.mp4")
// url: http://192.168.5.153:9000/recontext/meeting_123/file.mp4
```

#### GetRelativePath - Извлечение относительного пути из URL

```go
fullURL := "http://192.168.5.153:9000/recontext/meeting_123/file.mp4"
relativePath := client.GetRelativePath(fullURL)
// relativePath: meeting_123/file.mp4
```

#### EnsureBucket - Создание bucket если не существует

```go
ctx := context.Background()

err := client.EnsureBucket(ctx)
if err != nil {
    log.Fatal(err)
}
```

#### GetClient - Получение базового MinIO клиента

```go
// Для прямого доступа к MinIO API
minioClient := client.GetClient()

// Теперь можно использовать все методы github.com/minio/minio-go/v7
objects := minioClient.ListObjects(ctx, "recontext", minio.ListObjectsOptions{})
```

### Переменные окружения

```bash
# Обязательные
MINIO_ENDPOINT=minio:9000              # Endpoint для подключения (внутри Docker)
MINIO_ACCESS_KEY=minioadmin            # Access key
MINIO_SECRET_KEY=secret                # Secret key

# Опциональные
MINIO_BUCKET=recontext                 # Bucket (по умолчанию: recontext)
MINIO_USE_SSL=false                    # Использовать SSL (по умолчанию: false)
MINIO_PUBLIC_ENDPOINT=192.168.5.153:9000  # Публичный endpoint для формирования URL
```

### Content-Type Detection

MinIOClient автоматически определяет Content-Type по расширению файла:

- `.m3u8` → `application/vnd.apple.mpegurl`
- `.ts` → `video/MP2T`
- `.mp4` → `video/mp4`
- `.webm` → `video/webm`
- `.json` → `application/json`
- `.vtt` → `text/vtt`
- `.srt` → `application/x-subrip`
- Остальные → `application/octet-stream`

### Примеры использования в проекте

#### В video post-processor

```go
// pkg/video/storage.go использует MinIOClient внутри
uploader, err := video.NewStorageUploaderFromEnv()
if err != nil {
    log.Fatal(err)
}

// Скачиваем трек
err = uploader.DownloadFile(ctx, "meeting_123/tracks/TR_xxx.mp4", "/tmp/track.mp4")
```

#### Прямое использование

```go
import "Recontext.online/pkg/storage"

client, err := storage.NewMinIOClientFromEnv()
if err != nil {
    log.Fatal(err)
}

// Загружаем композитное видео
url, err := client.UploadFile(ctx, "/tmp/composite.m3u8", "meeting_123/composite.m3u8")
if err != nil {
    log.Fatal(err)
}

// Сохраняем только относительный путь в БД
relativePath := client.GetRelativePath(url)
db.SavePlaylistURL(meetingID, relativePath)
```

## Преимущества централизованного подхода

1. **Единая конфигурация**: Все части проекта используют одни и те же настройки MinIO
2. **Простота**: Не нужно повторять код создания клиента в каждом месте
3. **Переиспользуемость**: Один helper для всего проекта
4. **Легкость поддержки**: Изменения в логике работы с MinIO делаются в одном месте
5. **Автоматическое определение типов**: Content-Type определяется автоматически
6. **Поддержка публичных URL**: Автоматическое формирование URL для доступа извне Docker

## Миграция с старого кода

Если у вас есть код, который напрямую использует `minio-go`, его легко мигрировать:

### До:

```go
minioClient, err := minio.New(endpoint, &minio.Options{
    Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
    Secure: useSSL,
})

minioClient.FPutObject(ctx, bucket, objectName, filePath, minio.PutObjectOptions{})
```

### После:

```go
client, err := storage.NewMinIOClientFromEnv()

url, err := client.UploadFile(ctx, filePath, objectName)
```

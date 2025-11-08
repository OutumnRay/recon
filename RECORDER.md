# 🎥 Jitsi Individual Stream Recorder with REST API

## 📖 Идея

Цель проекта — **записывать аудио/видео потоки каждого участника Jitsi-конференции в отдельный файл** без нагрузки на клиентов.  
Система полностью серверная и управляется через REST API:

- `POST /record/start?user=user1` — запустить запись участника  
- `POST /record/stop?user=user1` — остановить запись и загрузить в MinIO  

---

## 🧩 Архитектура

```
┌──────────────────┐
│  Пользователи    │
│ (Jitsi Clients)  │
└───────┬──────────┘
        │ WebRTC
┌───────▼──────────┐
│ Jitsi Videobridge │
│ + Prosody + Jicofo│
└───────┬──────────┘
        │ RTP/WebRTC forwarding
┌───────▼──────────┐
│ Kurento Media Server │
│  (получает потоки)   │
└───────┬──────────┘
        │ WebSocket (KMS API)
┌───────▼──────────┐
│ Recorder Service  │
│ (FastAPI + Kurento + MinIO) │
│ Управление /record/start/stop │
└───────┬──────────┘
        │ S3 API
┌───────▼──────────┐
│   MinIO Storage   │
└──────────────────┘
```

---

## 📁 Структура проекта

```
jitsi-recording/
├── docker-compose.yml
├── kurento.conf.json
├── README.md
├── recorder/
│   ├── Dockerfile
│   ├── recorder.py
│   └── requirements.txt
└── recordings/
```

---

## ⚙️ Конфигурация

### **docker-compose.yml**

```yaml
version: '3.8'

services:
  prosody:
    image: jitsi/prosody:stable
    restart: always
    networks:
      - jitsi

  jicofo:
    image: jitsi/jicofo:stable
    restart: always
    environment:
      - XMPP_SERVER=prosody
    depends_on:
      - prosody
    networks:
      - jitsi

  jvb:
    image: jitsi/jvb:stable
    restart: always
    ports:
      - "10000:10000/udp"
    environment:
      - DOCKER_HOST_ADDRESS=jvb
      - XMPP_SERVER=prosody
    depends_on:
      - prosody
    networks:
      - jitsi

  web:
    image: jitsi/web:stable
    restart: always
    ports:
      - "8443:443"
      - "8000:80"
    environment:
      - ENABLE_RECORDING=1
    depends_on:
      - prosody
      - jicofo
      - jvb
    networks:
      - jitsi

  kurento:
    image: kurento/kurento-media-server:latest
    restart: always
    ports:
      - "8888:8888"
      - "40000-45000:40000-45000/udp"
    environment:
      - KMS_MIN_PORT=40000
      - KMS_MAX_PORT=45000
    volumes:
      - ./kurento.conf.json:/etc/kurento/kurento.conf.json
    networks:
      - jitsi

  recorder:
    build: ./recorder
    restart: always
    depends_on:
      - kurento
      - minio
    environment:
      - KURENTO_URI=ws://kurento:8888/kurento
      - MINIO_ENDPOINT=http://minio:9000
      - MINIO_ACCESS_KEY=admin
      - MINIO_SECRET_KEY=admin123
      - MINIO_BUCKET=recordings
    ports:
      - "8080:8080"
    volumes:
      - ./recordings:/recordings
    networks:
      - jitsi

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      - MINIO_ROOT_USER=admin
      - MINIO_ROOT_PASSWORD=admin123
    volumes:
      - ./minio-data:/data
    networks:
      - jitsi

networks:
  jitsi:
    driver: bridge
```

---

### **kurento.conf.json**

```json
{
  "mediaServer": {
    "resources": {
      "uri": "ws://0.0.0.0:8888/kurento"
    }
  }
}
```

---

## 🐍 Recorder-сервис (FastAPI + Kurento + MinIO)

### **recorder/recorder.py**

```python
import os
import time
import threading
import boto3
from fastapi import FastAPI, HTTPException
from kurento_client import KurentoClient

app = FastAPI(title="Jitsi Stream Recorder API")

KURENTO_URI = os.getenv("KURENTO_URI", "ws://kurento:8888/kurento")
MINIO_ENDPOINT = os.getenv("MINIO_ENDPOINT")
MINIO_ACCESS_KEY = os.getenv("MINIO_ACCESS_KEY")
MINIO_SECRET_KEY = os.getenv("MINIO_SECRET_KEY")
MINIO_BUCKET = os.getenv("MINIO_BUCKET")

client = boto3.client(
    "s3",
    endpoint_url=MINIO_ENDPOINT,
    aws_access_key_id=MINIO_ACCESS_KEY,
    aws_secret_access_key=MINIO_SECRET_KEY,
)

active_recordings = {}

def upload_to_minio(filepath):
    filename = os.path.basename(filepath)
    client.upload_file(filepath, MINIO_BUCKET, filename)
    print(f"✅ Uploaded {filename} to MinIO")

def record_user(user_id: str):
    try:
        kms = KurentoClient(KURENTO_URI)
        pipeline = kms.create("MediaPipeline")
        recorder = pipeline.create("RecorderEndpoint", {
            "uri": f"file:///recordings/{user_id}.webm"
        })
        active_recordings[user_id] = (pipeline, recorder)
        recorder.record()
        print(f"🎙️ Recording started for {user_id}")
    except Exception as e:
        print(f"❌ Error starting recording: {e}")

@app.post("/record/start")
def start_record(user: str):
    if user in active_recordings:
        raise HTTPException(status_code=400, detail="Recording already active for this user.")
    thread = threading.Thread(target=record_user, args=(user,))
    thread.start()
    return {"status": "started", "user": user}

@app.post("/record/stop")
def stop_record(user: str):
    if user not in active_recordings:
        raise HTTPException(status_code=404, detail="No active recording for this user.")
    pipeline, recorder = active_recordings[user]
    try:
        recorder.stop()
        pipeline.release()
        del active_recordings[user]
        filepath = f"/recordings/{user}.webm"
        if os.path.exists(filepath):
            upload_to_minio(filepath)
        print(f"🛑 Recording stopped for {user}")
        return {"status": "stopped", "user": user}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
```

---

### **recorder/requirements.txt**

```
boto3
kurento-client
fastapi
uvicorn
```

---

### **recorder/Dockerfile**

```dockerfile
FROM python:3.10-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY recorder.py .

EXPOSE 8080
CMD ["uvicorn", "recorder:app", "--host", "0.0.0.0", "--port", "8080"]
```

---

## 🚀 Запуск

```bash
docker-compose up -d
```

---

## 🌐 REST API

| Метод | URL | Параметры | Описание |
|--------|-----|------------|-----------|
| `POST` | `/record/start` | `?user=user1` | Начать запись участника |
| `POST` | `/record/stop`  | `?user=user1` | Остановить запись и загрузить в MinIO |

### Пример:
```bash
curl -X POST "http://localhost:8080/record/start?user=user1"
curl -X POST "http://localhost:8080/record/stop?user=user1"
```

---

## 🧠 Как это работает

1. Пользователи подключаются к конференции Jitsi.
2. JVB пересылает RTP-потоки в Kurento.
3. Recorder-сервис через REST API создает `MediaPipeline` и `RecorderEndpoint` для каждого участника.
4. После остановки запись сохраняется в `/recordings` и выгружается в MinIO.

---

## 🧾 Расширения

- Поддержка `room_id` и `session_id` для группировки записей.  
- Автоматическая транскрипция через **Whisper (CUDA)**.  
- Асинхронные задачи через **RabbitMQ**.  
- Вебхуки в систему аналитики.

---

## 🧰 Системные требования

| Ресурс | Минимум |
|--------|----------|
| CPU | 4 ядра |
| RAM | 8 ГБ |
| Disk | 20 ГБ |
| OS | Ubuntu 22.04 / Debian 12 |
| ПО | Docker + Docker Compose |

---

## 📦 Компоненты

| Компонент | Назначение |
|------------|-------------|
| **Jitsi Meet Stack** | видеоконференции |
| **Kurento Media Server** | маршрутизация потоков |
| **FastAPI Recorder Service** | REST API + запись |
| **MinIO** | хранилище записей |
| **Docker Compose** | оркестрация |

---

## 📜 Автор

**Святослав Иванов**  
AI-инженер и архитектор решений Jitsi / Kurento / FastAPI / MinIO  
2025

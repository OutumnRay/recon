"""Configuration for the transcription worker (loaded from environment variables)."""
import os

# ─── Redis ────────────────────────────────────────────────────────────────────
REDIS_HOST = os.getenv("REDIS_HOST", "localhost")
REDIS_PORT = int(os.getenv("REDIS_PORT", "6379"))
REDIS_PASSWORD = os.getenv("REDIS_PASSWORD") or None
REDIS_DB = int(os.getenv("REDIS_DB", "0"))
REDIS_TASK_QUEUE = os.getenv("REDIS_TASK_QUEUE", "recontext:transcription:queue")
REDIS_RESULT_QUEUE = os.getenv("REDIS_RESULT_QUEUE", "recontext:transcription:results")
REDIS_TIMEOUT = int(os.getenv("REDIS_TIMEOUT", "30"))  # BRPOP block timeout (seconds)

# ─── MinIO ────────────────────────────────────────────────────────────────────
MINIO_ENDPOINT = os.getenv("MINIO_ENDPOINT", "localhost:9000")
MINIO_ACCESS_KEY = os.getenv("MINIO_ACCESS_KEY", "minioadmin")
MINIO_SECRET_KEY = os.getenv("MINIO_SECRET_KEY", "minioadmin")
MINIO_SECURE = os.getenv("MINIO_SECURE", "false").lower() == "true"
MINIO_BUCKET = os.getenv("MINIO_BUCKET", "recontext")

# ─── Whisper ──────────────────────────────────────────────────────────────────
WHISPER_MODEL = os.getenv("WHISPER_MODEL", "base")
WHISPER_DEVICE = os.getenv("WHISPER_DEVICE", "cpu")
WHISPER_LANGUAGE = os.getenv("WHISPER_LANGUAGE", "ru")

# ─── HuggingFace (for pyannote diarization) ───────────────────────────────────
# Token is read from env; the default below is a fallback for dev only
HF_TOKEN = os.getenv("HF_TOKEN", "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")

# ─── Qdrant (vector DB) ───────────────────────────────────────────────────────
QDRANT_HOST = os.getenv("QDRANT_HOST", "localhost")
QDRANT_PORT = int(os.getenv("QDRANT_PORT", "6333"))
QDRANT_COLLECTION = os.getenv("QDRANT_COLLECTION", "recontext_transcripts")
QDRANT_ENABLED = os.getenv("QDRANT_ENABLED", "false").lower() == "true"
EMBEDDING_MODEL = os.getenv("EMBEDDING_MODEL", "paraphrase-multilingual-MiniLM-L12-v2")

# ─── Processing ───────────────────────────────────────────────────────────────
WORKDIR = os.getenv("WORKDIR", "/tmp/recontext-worker")
# Maximum pause between segments before starting a new paragraph (seconds)
MAX_PAUSE_SECONDS = float(os.getenv("MAX_PAUSE_SECONDS", "1.5"))
# Maximum paragraph length in characters before forcing a split
MAX_PARAGRAPH_CHARS = int(os.getenv("MAX_PARAGRAPH_CHARS", "800"))
ENABLE_DIARIZATION = os.getenv("ENABLE_DIARIZATION", "true").lower() == "true"
ENABLE_AUDIO_ENHANCEMENT = os.getenv("ENABLE_AUDIO_ENHANCEMENT", "true").lower() == "true"

# Health-check HTTP port
HEALTH_PORT = int(os.getenv("HEALTH_PORT", "8085"))

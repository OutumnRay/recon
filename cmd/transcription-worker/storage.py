"""MinIO and Redis client wrappers."""
import io
import json
import logging
from pathlib import Path
from typing import Optional

import redis as redis_lib
from minio import Minio
from minio.error import S3Error

import config as cfg

logger = logging.getLogger(__name__)


# ─── Redis ────────────────────────────────────────────────────────────────────

class RedisClient:
    def __init__(self):
        self._r: Optional[redis_lib.Redis] = None

    def connect(self) -> bool:
        try:
            self._r = redis_lib.Redis(
                host=cfg.REDIS_HOST,
                port=cfg.REDIS_PORT,
                password=cfg.REDIS_PASSWORD,
                db=cfg.REDIS_DB,
                decode_responses=True,
                socket_timeout=10,
                socket_connect_timeout=10,
                retry_on_timeout=True,
            )
            self._r.ping()
            logger.info("Connected to Redis at %s:%d", cfg.REDIS_HOST, cfg.REDIS_PORT)
            return True
        except Exception as exc:
            logger.error("Failed to connect to Redis: %s", exc)
            return False

    def pop_task(self, timeout: int = 30) -> Optional[dict]:
        """Block until a task is available in the queue (BRPOP)."""
        try:
            result = self._r.brpop(cfg.REDIS_TASK_QUEUE, timeout=timeout)
            if result is None:
                return None
            _, raw = result
            return json.loads(raw)
        except Exception as exc:
            logger.error("Error reading task from Redis: %s", exc)
            return None

    def push_result(self, result: dict) -> bool:
        """Push a result dict to the result queue."""
        try:
            self._r.lpush(cfg.REDIS_RESULT_QUEUE, json.dumps(result, ensure_ascii=False))
            return True
        except Exception as exc:
            logger.error("Error pushing result to Redis: %s", exc)
            return False

    def is_connected(self) -> bool:
        try:
            return self._r is not None and bool(self._r.ping())
        except Exception:
            return False


# ─── MinIO ────────────────────────────────────────────────────────────────────

class MinIOClient:
    def __init__(self):
        self._mc: Optional[Minio] = None

    def connect(self) -> bool:
        try:
            self._mc = Minio(
                cfg.MINIO_ENDPOINT,
                access_key=cfg.MINIO_ACCESS_KEY,
                secret_key=cfg.MINIO_SECRET_KEY,
                secure=cfg.MINIO_SECURE,
            )
            if not self._mc.bucket_exists(cfg.MINIO_BUCKET):
                self._mc.make_bucket(cfg.MINIO_BUCKET)
                logger.info("Created MinIO bucket: %s", cfg.MINIO_BUCKET)
            logger.info("Connected to MinIO at %s (bucket=%s)", cfg.MINIO_ENDPOINT, cfg.MINIO_BUCKET)
            return True
        except Exception as exc:
            logger.error("Failed to connect to MinIO: %s", exc)
            return False

    def download(self, object_path: str, local_path: Path) -> bool:
        """Download an object from MinIO to a local file."""
        try:
            local_path.parent.mkdir(parents=True, exist_ok=True)
            self._mc.fget_object(cfg.MINIO_BUCKET, object_path, str(local_path))
            logger.info("Downloaded  minio:%s → %s", object_path, local_path)
            return True
        except S3Error as exc:
            logger.error("MinIO download failed [%s]: %s", object_path, exc)
            return False

    def upload(self, local_path: Path, object_path: str, content_type: str = "application/octet-stream") -> bool:
        """Upload a local file to MinIO."""
        try:
            self._mc.fput_object(cfg.MINIO_BUCKET, object_path, str(local_path), content_type=content_type)
            logger.info("Uploaded    %s → minio:%s", local_path, object_path)
            return True
        except S3Error as exc:
            logger.error("MinIO upload failed [%s]: %s", object_path, exc)
            return False

    def upload_json(self, data: dict, object_path: str) -> bool:
        """Serialize *data* as JSON and upload directly to MinIO."""
        try:
            raw = json.dumps(data, ensure_ascii=False, indent=2).encode("utf-8")
            self._mc.put_object(
                cfg.MINIO_BUCKET, object_path,
                io.BytesIO(raw), len(raw),
                content_type="application/json",
            )
            logger.info("Uploaded    JSON → minio:%s", object_path)
            return True
        except S3Error as exc:
            logger.error("MinIO JSON upload failed [%s]: %s", object_path, exc)
            return False

    def upload_text(self, text: str, object_path: str, content_type: str = "text/plain; charset=utf-8") -> bool:
        """Upload a UTF-8 string directly to MinIO."""
        try:
            raw = text.encode("utf-8")
            self._mc.put_object(
                cfg.MINIO_BUCKET, object_path,
                io.BytesIO(raw), len(raw),
                content_type=content_type,
            )
            logger.info("Uploaded    text → minio:%s", object_path)
            return True
        except S3Error as exc:
            logger.error("MinIO text upload failed [%s]: %s", object_path, exc)
            return False

    def is_connected(self) -> bool:
        try:
            self._mc.bucket_exists(cfg.MINIO_BUCKET)
            return True
        except Exception:
            return False


# ─── Singletons ───────────────────────────────────────────────────────────────
redis_client = RedisClient()
minio_client = MinIOClient()

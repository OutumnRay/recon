#!/usr/bin/env python3
"""
Local testing helper: push a transcription task to Redis.

Usage:
    python test_push_task.py [path/to/local/video.mp4]

If a local file is given it is uploaded to MinIO first, then a task is queued.
If no file is given a dummy task is created so you can test the worker pipeline
without a real video (it will fail at the download step, which is expected).
"""
import json
import os
import sys
import uuid
from pathlib import Path

import redis
from minio import Minio

REDIS_HOST = os.getenv("REDIS_HOST", "localhost")
REDIS_PORT = int(os.getenv("REDIS_PORT", "6379"))
REDIS_TASK_QUEUE = os.getenv("REDIS_TASK_QUEUE", "recontext:transcription:queue")

MINIO_ENDPOINT = os.getenv("MINIO_ENDPOINT", "localhost:9000")
MINIO_ACCESS_KEY = os.getenv("MINIO_ACCESS_KEY", "minioadmin")
MINIO_SECRET_KEY = os.getenv("MINIO_SECRET_KEY", "minioadmin")
MINIO_BUCKET = os.getenv("MINIO_BUCKET", "recontext")


def main():
    local_file = Path(sys.argv[1]) if len(sys.argv) > 1 else None

    # ── Connect to Redis ──────────────────────────────────────────────────────
    r = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)
    r.ping()
    print(f"[OK] Redis connected: {REDIS_HOST}:{REDIS_PORT}")

    # ── Connect to MinIO ──────────────────────────────────────────────────────
    mc = Minio(MINIO_ENDPOINT, access_key=MINIO_ACCESS_KEY, secret_key=MINIO_SECRET_KEY, secure=False)
    if not mc.bucket_exists(MINIO_BUCKET):
        mc.make_bucket(MINIO_BUCKET)
        print(f"[OK] Created bucket: {MINIO_BUCKET}")

    task_id = str(uuid.uuid4())
    session_id = str(uuid.uuid4())
    object_path = f"recordings/test/{session_id}/input.mp4"

    if local_file and local_file.exists():
        # Upload local file to MinIO
        mc.fput_object(MINIO_BUCKET, object_path, str(local_file), content_type="video/mp4")
        print(f"[OK] Uploaded {local_file} → minio:{object_path}")
    else:
        print(f"[WARN] No local file provided – task will fail at download (use for smoke-testing only)")

    task = {
        "task_id": task_id,
        "session_id": session_id,
        "user_id": str(uuid.uuid4()),
        "video_type": "lecture",
        "bucket": MINIO_BUCKET,
        "object_path": object_path,
        "language": "ru",
    }

    r.lpush(REDIS_TASK_QUEUE, json.dumps(task))
    print(f"[OK] Task pushed to '{REDIS_TASK_QUEUE}'")
    print(json.dumps(task, indent=2, ensure_ascii=False))


if __name__ == "__main__":
    main()

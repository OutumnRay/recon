#!/usr/bin/env python3
"""
Recontext Transcription Worker
================================
Consumes transcription tasks from Redis, processes audio/video files stored in
MinIO, performs Whisper transcription + speaker diarization, builds semantic
paragraphs, and stores all results back to MinIO.

Task JSON (pushed by the Go service via LPUSH):
    {
        "task_id":    "<uuid>",
        "session_id": "<uuid>",          # usually the LiveKit track ID
        "user_id":    "<uuid>",
        "video_type": "conference|lecture",
        "bucket":     "recontext",       # optional, defaults to MINIO_BUCKET
        "object_path":"recordings/...",  # MinIO object key of the source file
        "language":   "ru"               # optional, defaults to WHISPER_LANGUAGE
    }

Result JSON (pushed to REDIS_RESULT_QUEUE):
    {
        "task_id":         "<uuid>",
        "session_id":      "<uuid>",
        "status":          "completed|failed",
        "transcript_path": "transcripts/<session_id>/paragraphs.json",
        "srt_path":        "transcripts/<session_id>/transcript.srt",
        "paragraphs_count": int,
        "chunks_count":    int,
        "duration":        float,
        "error":           "..." (only on failure)
    }
"""

import json
import logging
import os
import shutil
import signal
import sys
import threading
import time
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path

import config as cfg
from diarization import assign_speakers, diarize
from equalizer import enhance_audio
from spliter import extract_audio, extract_video_only, get_media_info
from storage import minio_client, redis_client
from transcribation import (
    merge_segments_into_paragraphs,
    paragraphs_to_chunks,
    segments_to_srt,
    transcribe_with_segments,
)
from vector_db import upsert_to_vector_db


# ─── Logging ──────────────────────────────────────────────────────────────────

def _setup_logging():
    level_name = os.getenv("LOG_LEVEL", "INFO").upper()
    level = getattr(logging, level_name, logging.INFO)
    logging.basicConfig(
        level=level,
        format="%(asctime)s [%(levelname)-8s] %(name)s: %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
        handlers=[logging.StreamHandler(sys.stdout)],
    )


_setup_logging()
logger = logging.getLogger("worker")

# ─── Global state ─────────────────────────────────────────────────────────────

_running = True
_stats = {
    "tasks_processed": 0,
    "tasks_failed": 0,
    "start_time": time.time(),
}


# ─── Health-check HTTP server ─────────────────────────────────────────────────

class _HealthHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path in ("/health", "/status"):
            body = json.dumps({
                "status": "healthy" if _running else "stopping",
                "uptime_seconds": int(time.time() - _stats["start_time"]),
                "redis_connected": redis_client.is_connected(),
                "minio_connected": minio_client.is_connected(),
                "tasks_processed": _stats["tasks_processed"],
                "tasks_failed": _stats["tasks_failed"],
            }).encode()
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(body)
        else:
            self.send_response(404)
            self.end_headers()

    # Suppress default per-request access log
    def log_message(self, fmt, *args):  # noqa: D401
        pass


def _start_health_server():
    server = HTTPServer(("0.0.0.0", cfg.HEALTH_PORT), _HealthHandler)
    t = threading.Thread(target=server.serve_forever, daemon=True)
    t.start()
    logger.info("Health endpoint: http://0.0.0.0:%d/health", cfg.HEALTH_PORT)


# ─── Task processing ──────────────────────────────────────────────────────────

def process_task(task: dict) -> dict:
    task_id = task["task_id"]
    session_id = task["session_id"]
    user_id = task.get("user_id", "unknown")
    video_type = task.get("video_type", "lecture")
    object_path = task["object_path"]
    language = task.get("language") or cfg.WHISPER_LANGUAGE

    logger.info("[%s] ── START  session=%s  type=%s", task_id, session_id, video_type)

    workdir = Path(cfg.WORKDIR) / session_id
    workdir.mkdir(parents=True, exist_ok=True)

    try:
        # ── 1. Download source media from MinIO ───────────────────────────────
        ext = Path(object_path).suffix or ".mp4"
        raw_media = workdir / f"input{ext}"

        if not raw_media.exists():
            logger.info("[%s] Downloading  minio:%s", task_id, object_path)
            if not minio_client.download(object_path, raw_media):
                raise RuntimeError(f"MinIO download failed: {object_path}")

        media_info = get_media_info(raw_media)
        logger.info("[%s] Media info: %s", task_id, media_info)

        # ── 2. Split audio and video tracks ───────────────────────────────────
        audio_wav = workdir / "audio.wav"
        video_only = workdir / "video_only.mp4"

        if not audio_wav.exists():
            if not extract_audio(raw_media, audio_wav):
                raise RuntimeError("Audio extraction failed")

        if media_info.get("has_video") and not video_only.exists():
            extract_video_only(raw_media, video_only)  # best-effort

        # ── 3. Enhance audio (noise reduction + normalisation) ─────────────────
        enhanced_wav = workdir / "audio_enhanced.wav"
        if cfg.ENABLE_AUDIO_ENHANCEMENT and not enhanced_wav.exists():
            logger.info("[%s] Enhancing audio...", task_id)
            enhance_audio(audio_wav, enhanced_wav)
        else:
            enhanced_wav = audio_wav

        # ── 4. Whisper transcription ───────────────────────────────────────────
        segments_path = workdir / "segments.json"

        if not segments_path.exists():
            segments = transcribe_with_segments(enhanced_wav, language=language)

            # ── 5. Speaker diarization (conference only) ───────────────────────
            if video_type == "conference" and cfg.ENABLE_DIARIZATION and cfg.HF_TOKEN:
                try:
                    diar = diarize(enhanced_wav)
                    segments = assign_speakers(segments, diar)
                except Exception as exc:
                    logger.warning("[%s] Diarization failed, using single speaker: %s", task_id, exc)
                    for s in segments:
                        s["speaker"] = user_id
            else:
                for s in segments:
                    s["speaker"] = user_id

            segments_path.write_text(
                json.dumps(segments, ensure_ascii=False, indent=2), encoding="utf-8"
            )
        else:
            logger.info("[%s] Reusing cached segments", task_id)
            segments = json.loads(segments_path.read_text(encoding="utf-8"))

        # ── 6. Merge into semantic paragraphs ──────────────────────────────────
        paragraphs = merge_segments_into_paragraphs(
            segments,
            max_pause_sec=cfg.MAX_PAUSE_SECONDS,
            max_chars=cfg.MAX_PARAGRAPH_CHARS,
        )

        # ── 7. Build vector-DB chunks ──────────────────────────────────────────
        chunks = paragraphs_to_chunks(paragraphs)
        for c in chunks:
            c["session_id"] = session_id
            c["user_id"] = user_id

        # ── 8. SRT subtitle file ───────────────────────────────────────────────
        srt_content = segments_to_srt(paragraphs)

        # ── 9. Upload all results to MinIO ─────────────────────────────────────
        base = f"transcripts/{session_id}"

        minio_client.upload_json(
            {"session_id": session_id, "segments": segments},
            f"{base}/segments.json",
        )
        minio_client.upload_json(
            {"session_id": session_id, "paragraphs": paragraphs},
            f"{base}/paragraphs.json",
        )
        minio_client.upload_json(
            {"session_id": session_id, "chunks": chunks},
            f"{base}/chunks.json",
        )
        minio_client.upload_text(srt_content, f"{base}/transcript.srt", "text/plain; charset=utf-8")

        if audio_wav.exists():
            minio_client.upload(audio_wav, f"{base}/audio.wav", "audio/wav")
        if video_only.exists():
            minio_client.upload(video_only, f"{base}/video_only.mp4", "video/mp4")

        # ── 10. Index in Qdrant ────────────────────────────────────────────────
        upsert_to_vector_db(chunks)

        # ── Cleanup temp files ─────────────────────────────────────────────────
        shutil.rmtree(workdir, ignore_errors=True)

        logger.info(
            "[%s] ── DONE   paragraphs=%d  chunks=%d  duration=%.1fs",
            task_id, len(paragraphs), len(chunks), media_info.get("duration", 0),
        )
        return {
            "task_id": task_id,
            "session_id": session_id,
            "status": "completed",
            "paragraphs_count": len(paragraphs),
            "chunks_count": len(chunks),
            "duration": media_info.get("duration", 0),
            "transcript_path": f"{base}/paragraphs.json",
            "srt_path": f"{base}/transcript.srt",
        }

    except Exception as exc:
        logger.exception("[%s] ── FAILED: %s", task_id, exc)
        shutil.rmtree(workdir, ignore_errors=True)
        return {
            "task_id": task_id,
            "session_id": session_id,
            "status": "failed",
            "error": str(exc),
        }


# ─── Connection helpers ────────────────────────────────────────────────────────

def _wait_for(name: str, connect_fn, retries: int = 10, delay: int = 5) -> None:
    for attempt in range(1, retries + 1):
        if connect_fn():
            return
        logger.warning("%s not ready (attempt %d/%d), retrying in %ds...", name, attempt, retries, delay)
        time.sleep(delay)
    logger.error("Could not connect to %s after %d attempts. Exiting.", name, retries)
    sys.exit(1)


# ─── Main ─────────────────────────────────────────────────────────────────────

def main():
    global _running

    logger.info("=" * 60)
    logger.info("Recontext Transcription Worker")
    logger.info("  Whisper model : %s  (%s)", cfg.WHISPER_MODEL, cfg.WHISPER_DEVICE)
    logger.info("  Redis         : %s:%d  queue=%s", cfg.REDIS_HOST, cfg.REDIS_PORT, cfg.REDIS_TASK_QUEUE)
    logger.info("  MinIO         : %s  bucket=%s", cfg.MINIO_ENDPOINT, cfg.MINIO_BUCKET)
    logger.info("  Diarization   : %s", cfg.ENABLE_DIARIZATION)
    logger.info("  Qdrant        : %s", cfg.QDRANT_ENABLED)
    logger.info("=" * 60)

    _start_health_server()

    _wait_for("Redis", redis_client.connect)
    _wait_for("MinIO", minio_client.connect)

    Path(cfg.WORKDIR).mkdir(parents=True, exist_ok=True)

    def _shutdown(sig, _frame):
        global _running
        logger.info("Signal %s received — finishing current task then stopping…", sig)
        _running = False

    signal.signal(signal.SIGTERM, _shutdown)
    signal.signal(signal.SIGINT, _shutdown)

    logger.info("Waiting for tasks on '%s'…", cfg.REDIS_TASK_QUEUE)

    while _running:
        task = redis_client.pop_task(timeout=cfg.REDIS_TIMEOUT)
        if task is None:
            continue  # timeout, loop back to check _running

        result = process_task(task)

        if result["status"] == "completed":
            _stats["tasks_processed"] += 1
        else:
            _stats["tasks_failed"] += 1

        redis_client.push_result(result)
        logger.info("[%s] Result pushed → status=%s", result["task_id"], result["status"])

    logger.info("Worker stopped gracefully.")


if __name__ == "__main__":
    main()

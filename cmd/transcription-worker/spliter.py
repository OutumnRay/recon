"""Media splitting utilities using ffmpeg."""
import json
import logging
import subprocess
from pathlib import Path

logger = logging.getLogger(__name__)


def _run(cmd: list, label: str) -> bool:
    """Run an ffmpeg command, log stderr on failure."""
    result = subprocess.run(cmd, stdout=subprocess.DEVNULL, stderr=subprocess.PIPE)
    if result.returncode != 0:
        logger.error("%s failed (exit %d):\n%s", label, result.returncode, result.stderr.decode(errors="replace"))
        return False
    return True


def extract_audio(video_path: Path, wav_path: Path) -> bool:
    """
    Extract audio from *video_path* as mono 16 kHz 16-bit PCM WAV.
    This format is optimal for Whisper transcription.
    """
    logger.info("Extracting audio: %s → %s", video_path, wav_path)
    return _run(
        [
            "ffmpeg", "-y",
            "-i", str(video_path),
            "-ac", "1",            # mono
            "-ar", "16000",        # 16 kHz
            "-c:a", "pcm_s16le",   # 16-bit PCM
            "-vn",                 # drop video
            str(wav_path),
        ],
        "audio extraction",
    )


def extract_video_only(video_path: Path, output_path: Path) -> bool:
    """
    Copy the video stream from *video_path* without audio.
    Uses stream copy (no re-encoding) for speed.
    """
    logger.info("Extracting video-only: %s → %s", video_path, output_path)
    return _run(
        [
            "ffmpeg", "-y",
            "-i", str(video_path),
            "-an",          # drop audio
            "-c:v", "copy", # stream copy
            str(output_path),
        ],
        "video extraction",
    )


def get_media_info(file_path: Path) -> dict:
    """
    Return basic media metadata via ffprobe:
        {duration: float, has_video: bool, has_audio: bool}
    Returns an empty dict if ffprobe fails.
    """
    result = subprocess.run(
        [
            "ffprobe", "-v", "quiet",
            "-print_format", "json",
            "-show_streams", "-show_format",
            str(file_path),
        ],
        capture_output=True, text=True,
    )
    if result.returncode != 0:
        logger.warning("ffprobe failed for %s: %s", file_path, result.stderr.strip())
        return {}
    try:
        info = json.loads(result.stdout)
        streams = info.get("streams", [])
        return {
            "duration": round(float(info.get("format", {}).get("duration", 0)), 2),
            "has_video": any(s["codec_type"] == "video" for s in streams),
            "has_audio": any(s["codec_type"] == "audio" for s in streams),
        }
    except Exception as exc:
        logger.warning("Failed to parse ffprobe output for %s: %s", file_path, exc)
        return {}

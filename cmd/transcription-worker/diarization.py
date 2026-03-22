"""Speaker diarization using pyannote.audio."""
import logging
from pathlib import Path

import config as cfg

logger = logging.getLogger(__name__)

_pipeline = None


def get_pipeline():
    global _pipeline
    if _pipeline is None:
        if not cfg.HF_TOKEN:
            raise RuntimeError("HF_TOKEN environment variable is required for diarization")
        from pyannote.audio import Pipeline
        logger.info("Loading pyannote/speaker-diarization-3.1 pipeline...")
        _pipeline = Pipeline.from_pretrained(
            "pyannote/speaker-diarization-3.1",
            use_auth_token=cfg.HF_TOKEN,
        )
        logger.info("Diarization pipeline loaded")
    return _pipeline


def diarize(wav_path: Path) -> list:
    """
    Run speaker diarization on *wav_path*.
    Returns a list of {start, end, speaker} dicts.
    """
    logger.info("Running diarization on %s...", wav_path)
    pipeline = get_pipeline()
    diarization = pipeline(str(wav_path))

    segments = [
        {
            "start": round(turn.start, 3),
            "end": round(turn.end, 3),
            "speaker": label,
        }
        for turn, _, label in diarization.itertracks(yield_label=True)
    ]

    n_speakers = len({s["speaker"] for s in segments})
    logger.info("Diarization complete: %d turns, %d speakers", len(segments), n_speakers)
    return segments


def assign_speakers(transcript_segments: list, diarization_segments: list) -> list:
    """
    Label each transcript segment with the dominant speaker from diarization.
    Uses the midpoint of each segment to look up the speaker.
    """
    for seg in transcript_segments:
        mid = (seg["start"] + seg["end"]) / 2
        seg["speaker"] = "unknown"
        for d in diarization_segments:
            if d["start"] <= mid <= d["end"]:
                seg["speaker"] = d["speaker"]
                break
    return transcript_segments

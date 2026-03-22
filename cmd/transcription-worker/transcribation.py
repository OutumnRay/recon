"""Whisper transcription + semantic paragraph merging."""
import logging
import re
from pathlib import Path
from typing import Optional

import whisper

import config as cfg

logger = logging.getLogger(__name__)

_model = None


def get_model():
    global _model
    if _model is None:
        logger.info("Loading Whisper model '%s' on device '%s'...", cfg.WHISPER_MODEL, cfg.WHISPER_DEVICE)
        _model = whisper.load_model(cfg.WHISPER_MODEL, device=cfg.WHISPER_DEVICE)
        logger.info("Whisper model loaded")
    return _model


# ─── Transcription ────────────────────────────────────────────────────────────

def transcribe_with_segments(audio_path: Path, language: Optional[str] = None) -> list:
    """
    Transcribe *audio_path* with Whisper and return a list of segment dicts:
        {start, end, text, words:[{word, start, end, probability}], no_speech_prob}

    word_timestamps=True gives per-word timing used later by the paragraph merger.
    """
    lang = language or cfg.WHISPER_LANGUAGE
    logger.info("Transcribing %s  [language=%s]", audio_path, lang)

    result = get_model().transcribe(
        str(audio_path),
        language=lang,
        fp16=False,
        verbose=False,
        word_timestamps=True,
        condition_on_previous_text=True,
        no_speech_threshold=0.6,
        logprob_threshold=-1.0,
        compression_ratio_threshold=2.4,
    )

    segments = []
    for seg in result["segments"]:
        if seg.get("no_speech_prob", 0.0) > 0.8:
            logger.debug(
                "Skipping silent segment [%.1fs – %.1fs]  no_speech_prob=%.2f",
                seg["start"], seg["end"], seg["no_speech_prob"],
            )
            continue

        words = [
            {
                "word": w["word"].strip(),
                "start": round(w["start"], 3),
                "end": round(w["end"], 3),
                "probability": round(w.get("probability", 1.0), 3),
            }
            for w in seg.get("words", [])
        ]

        text = seg["text"].strip()
        if not text:
            continue

        segments.append({
            "start": round(seg["start"], 3),
            "end": round(seg["end"], 3),
            "text": text,
            "words": words,
            "no_speech_prob": round(seg.get("no_speech_prob", 0.0), 3),
        })

    logger.info("Transcription complete: %d segments", len(segments))
    return segments


# ─── Semantic paragraph merging ───────────────────────────────────────────────

def merge_segments_into_paragraphs(
    segments: list,
    max_pause_sec: float = 1.5,
    max_chars: int = 800,
) -> list:
    """
    Merge raw Whisper segments into *semantic paragraphs*.

    A new paragraph starts when ANY of these conditions is true:
    - Speaker changes (after diarization assignment)
    - Pause between segments exceeds *max_pause_sec*
    - Current paragraph text ends with a sentence boundary (.!?…)
      AND is already longer than 150 characters
    - Combined length would exceed *max_chars*

    This produces human-readable blocks instead of fixed 2-second chunks.
    """
    if not segments:
        return []

    paragraphs = []
    current = None

    for seg in segments:
        text = seg["text"].strip()
        if not text:
            continue

        speaker = seg.get("speaker", "unknown")

        if current is None:
            current = {
                "start": seg["start"],
                "end": seg["end"],
                "speaker": speaker,
                "text": text,
                "words": list(seg.get("words", [])),
            }
            continue

        pause = seg["start"] - current["end"]
        combined_len = len(current["text"]) + 1 + len(text)

        speaker_changed = speaker != current["speaker"]
        long_pause = pause > max_pause_sec
        sentence_end = bool(re.search(r"[.!?…]\s*$", current["text"]))
        too_long = combined_len > max_chars

        should_split = (
            speaker_changed
            or long_pause
            or too_long
            or (sentence_end and len(current["text"]) > 150)
        )

        if should_split:
            paragraphs.append(current)
            current = {
                "start": seg["start"],
                "end": seg["end"],
                "speaker": speaker,
                "text": text,
                "words": list(seg.get("words", [])),
            }
        else:
            current["end"] = seg["end"]
            current["text"] += " " + text
            current["words"].extend(seg.get("words", []))

    if current:
        paragraphs.append(current)

    logger.info(
        "Merged %d segments → %d paragraphs (avg %.0f chars/para)",
        len(segments),
        len(paragraphs),
        sum(len(p["text"]) for p in paragraphs) / max(len(paragraphs), 1),
    )
    return paragraphs


# ─── Chunk splitting for vector DB ────────────────────────────────────────────

def paragraphs_to_chunks(paragraphs: list, max_words: int = 300, overlap_words: int = 20) -> list:
    """
    Convert paragraphs to vector-DB chunks.
    - Short paragraphs (≤ max_words) become one chunk each.
    - Long paragraphs are split at sentence boundaries with *overlap_words* carry-over.
    """
    chunks = []
    for para in paragraphs:
        if len(para["text"].split()) <= max_words:
            chunks.append({
                "start": para["start"],
                "end": para["end"],
                "speaker": para["speaker"],
                "text": para["text"],
            })
        else:
            chunks.extend(_split_long_paragraph(para, max_words, overlap_words))

    logger.debug("Created %d chunks from %d paragraphs", len(chunks), len(paragraphs))
    return chunks


def _split_long_paragraph(para: dict, max_words: int, overlap_words: int) -> list:
    """Split a long paragraph into overlapping chunks at sentence boundaries."""
    # Split on sentence endings; keep the delimiter attached to preceding sentence
    sentences = re.split(r"(?<=[.!?])\s+", para["text"])

    chunks = []
    current_parts: list = []

    for sentence in sentences:
        s_words = sentence.split()

        if current_parts and len(" ".join(current_parts).split()) + len(s_words) > max_words:
            chunks.append({
                "start": para["start"],
                "end": para["end"],
                "speaker": para["speaker"],
                "text": " ".join(current_parts),
            })
            # Carry-over overlap from the end of the previous chunk
            all_words = " ".join(current_parts).split()
            overlap_text = " ".join(all_words[-overlap_words:]) if overlap_words else ""
            current_parts = [overlap_text] if overlap_text else []

        current_parts.append(sentence)

    if current_parts:
        chunks.append({
            "start": para["start"],
            "end": para["end"],
            "speaker": para["speaker"],
            "text": " ".join(current_parts),
        })

    return chunks


# ─── SRT export ───────────────────────────────────────────────────────────────

def segments_to_srt(segments: list) -> str:
    """Convert a list of paragraph/segment dicts to SRT subtitle format."""
    lines = []
    for i, seg in enumerate(segments, 1):
        start = _fmt_srt(seg["start"])
        end = _fmt_srt(seg["end"])
        speaker = seg.get("speaker", "")
        text = f"[{speaker}] {seg['text']}" if speaker and speaker != "unknown" else seg["text"]
        lines.append(f"{i}\n{start} --> {end}\n{text}\n")
    return "\n".join(lines)


def _fmt_srt(seconds: float) -> str:
    h = int(seconds // 3600)
    m = int((seconds % 3600) // 60)
    s = int(seconds % 60)
    ms = int((seconds % 1) * 1000)
    return f"{h:02d}:{m:02d}:{s:02d},{ms:03d}"

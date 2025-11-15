"""Whisper transcription worker."""
import os
import tempfile
import subprocess
from typing import List, Dict, Optional
import requests
from faster_whisper import WhisperModel
from config import Config


class TranscriptionWorker:
    """Worker for transcribing audio using Faster Whisper."""

    def __init__(self):
        """Initialize Whisper model."""
        print(f"Loading Whisper model: {Config.WHISPER_MODEL}")
        self.model = WhisperModel(
            Config.WHISPER_MODEL,
            device=Config.WHISPER_DEVICE,
            compute_type=Config.WHISPER_COMPUTE_TYPE
        )
        print("Whisper model loaded successfully")

    def download_m3u8_and_combine(self, m3u8_url: str, token: Optional[str] = None) -> str:
        """
        Download m3u8 playlist and combine all segments into a single audio file.

        Args:
            m3u8_url: URL to m3u8 playlist
            token: Optional authentication token

        Returns:
            Path to combined audio file
        """
        print(f"📥 Downloading and combining m3u8 from: {m3u8_url}")

        headers = {}
        if token:
            headers['Authorization'] = f'Bearer {token}'

        # Download m3u8 playlist
        response = requests.get(m3u8_url, headers=headers)
        response.raise_for_status()
        playlist_content = response.text

        print(f"📄 Playlist content:\n{playlist_content[:500]}...")

        # Parse m3u8 to get segment URLs
        base_url = m3u8_url.rsplit('/', 1)[0]  # Get base URL without filename
        segments = []

        for line in playlist_content.split('\n'):
            line = line.strip()
            # Skip comments and empty lines
            if line and not line.startswith('#'):
                # Build full URL for segment
                if line.startswith('http'):
                    segment_url = line
                else:
                    segment_url = f"{base_url}/{line}"
                segments.append(segment_url)

        print(f"📦 Found {len(segments)} segments to download")

        if not segments:
            raise ValueError("No segments found in m3u8 playlist")

        # Create temporary directory for segments
        temp_dir = tempfile.mkdtemp()
        segment_files = []

        try:
            # Download all segments
            for i, segment_url in enumerate(segments):
                print(f"📥 Downloading segment {i+1}/{len(segments)}: {segment_url}")

                segment_response = requests.get(segment_url, headers=headers, stream=True)
                segment_response.raise_for_status()

                # Save segment
                segment_path = os.path.join(temp_dir, f"segment_{i:04d}.ts")
                with open(segment_path, 'wb') as f:
                    for chunk in segment_response.iter_content(chunk_size=8192):
                        f.write(chunk)

                segment_files.append(segment_path)
                print(f"✅ Downloaded segment {i+1}/{len(segments)}")

            # Combine segments using ffmpeg
            combined_file = tempfile.NamedTemporaryFile(delete=False, suffix='.m4a')
            combined_path = combined_file.name
            combined_file.close()

            print(f"🔧 Combining {len(segment_files)} segments with ffmpeg...")

            # Create concat file for ffmpeg
            concat_file_path = os.path.join(temp_dir, 'concat.txt')
            with open(concat_file_path, 'w') as f:
                for segment in segment_files:
                    # Escape single quotes in paths
                    escaped_path = segment.replace("'", "'\\''")
                    f.write(f"file '{escaped_path}'\n")

            # Use ffmpeg to concatenate segments
            # -f concat: Use concat demuxer
            # -safe 0: Allow absolute paths
            # -i: Input concat file
            # -c copy: Copy streams without re-encoding (fast)
            # -y: Overwrite output file
            ffmpeg_cmd = [
                'ffmpeg',
                '-f', 'concat',
                '-safe', '0',
                '-i', concat_file_path,
                '-c', 'copy',
                '-y',
                combined_path
            ]

            result = subprocess.run(
                ffmpeg_cmd,
                capture_output=True,
                text=True,
                timeout=300  # 5 minutes timeout
            )

            if result.returncode != 0:
                print(f"❌ FFmpeg stderr:\n{result.stderr}")
                raise RuntimeError(f"FFmpeg failed with return code {result.returncode}")

            print(f"✅ Combined audio saved to: {combined_path}")

            # Get file size
            file_size = os.path.getsize(combined_path) / (1024 * 1024)  # MB
            print(f"📊 Combined file size: {file_size:.2f} MB")

            return combined_path

        finally:
            # Cleanup segment files
            for segment_file in segment_files:
                try:
                    if os.path.exists(segment_file):
                        os.remove(segment_file)
                except Exception as e:
                    print(f"⚠️  Failed to remove segment file {segment_file}: {e}")

            # Cleanup concat file
            try:
                if os.path.exists(concat_file_path):
                    os.remove(concat_file_path)
            except:
                pass

            # Cleanup temp directory
            try:
                os.rmdir(temp_dir)
            except:
                pass

    def download_audio(self, audio_url: str, token: Optional[str] = None) -> str:
        """
        Download audio file from URL.
        Handles both direct audio files and m3u8 playlists.

        Args:
            audio_url: URL to download audio from
            token: Optional authentication token

        Returns:
            Path to downloaded temporary file
        """
        # Check if this is an m3u8 playlist
        if audio_url.endswith('.m3u8') or 'playlist.m3u8' in audio_url:
            return self.download_m3u8_and_combine(audio_url, token)

        headers = {}
        if token:
            headers['Authorization'] = f'Bearer {token}'

        # Download audio file
        response = requests.get(audio_url, headers=headers, stream=True)
        response.raise_for_status()

        # Save to temporary file
        suffix = '.m4a'  # Default suffix
        if 'content-type' in response.headers:
            content_type = response.headers['content-type']
            if 'mp4' in content_type:
                suffix = '.mp4'
            elif 'webm' in content_type:
                suffix = '.webm'
            elif 'wav' in content_type:
                suffix = '.wav'

        temp_file = tempfile.NamedTemporaryFile(delete=False, suffix=suffix)
        for chunk in response.iter_content(chunk_size=8192):
            temp_file.write(chunk)
        temp_file.close()

        return temp_file.name

    def transcribe_audio(
        self,
        audio_path: str,
        language: Optional[str] = None
    ) -> List[Dict]:
        """
        Transcribe audio file using Faster Whisper.

        Args:
            audio_path: Path to audio file
            language: Optional language code (e.g., 'en', 'ru')

        Returns:
            List of phrase dictionaries with start, end, text, confidence
        """
        print(f"Starting transcription of: {audio_path}")

        # Transcribe with Faster Whisper
        segments, info = self.model.transcribe(
            audio_path,
            language=language,
            beam_size=5,
            vad_filter=True,  # Voice activity detection
            vad_parameters=dict(min_silence_duration_ms=500),
        )

        detected_language = info.language
        print(f"Detected language: {detected_language} (probability: {info.language_probability:.2f})")

        # Convert segments to phrases
        phrases = []
        for segment in segments:
            phrases.append({
                'start': segment.start,
                'end': segment.end,
                'text': segment.text.strip(),
                'confidence': segment.avg_logprob,  # Average log probability
                'language': detected_language
            })

        print(f"Transcription completed: {len(phrases)} phrases")
        return phrases

    def transcribe_from_url(
        self,
        audio_url: str,
        language: Optional[str] = None,
        token: Optional[str] = None
    ) -> List[Dict]:
        """
        Download and transcribe audio from URL.

        Args:
            audio_url: URL to audio file
            language: Optional language code
            token: Optional authentication token

        Returns:
            List of phrase dictionaries
        """
        audio_path = None
        try:
            # Download audio
            print(f"Downloading audio from: {audio_url}")
            audio_path = self.download_audio(audio_url, token)

            # Transcribe
            phrases = self.transcribe_audio(audio_path, language)

            return phrases

        finally:
            # Clean up temporary file
            if audio_path and os.path.exists(audio_path):
                os.remove(audio_path)
                print(f"Cleaned up temporary file: {audio_path}")

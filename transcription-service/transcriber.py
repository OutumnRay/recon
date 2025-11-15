"""Whisper transcription worker."""
import os
import tempfile
import subprocess
from typing import List, Dict, Optional
import requests
from minio import Minio
from faster_whisper import WhisperModel
from config import Config


class TranscriptionWorker:
    """Worker for transcribing audio using Faster Whisper."""

    def __init__(self):
        """Initialize Whisper model and MinIO client."""
        print(f"Loading Whisper model: {Config.WHISPER_MODEL}")
        self.model = WhisperModel(
            Config.WHISPER_MODEL,
            device=Config.WHISPER_DEVICE,
            compute_type=Config.WHISPER_COMPUTE_TYPE
        )
        print("Whisper model loaded successfully")

        # Initialize MinIO client
        print(f"Initializing MinIO client: {Config.MINIO_ENDPOINT}")

        # Extract endpoint without protocol (MinIO client doesn't accept URLs with protocol)
        minio_endpoint = Config.MINIO_ENDPOINT
        if minio_endpoint.startswith('https://'):
            minio_endpoint = minio_endpoint.replace('https://', '')
        elif minio_endpoint.startswith('http://'):
            minio_endpoint = minio_endpoint.replace('http://', '')

        # Add default port if not specified
        if ':' not in minio_endpoint:
            minio_endpoint = f"{minio_endpoint}:9000"

        self.minio_client = Minio(
            minio_endpoint,
            access_key=Config.MINIO_ACCESS_KEY,
            secret_key=Config.MINIO_SECRET_KEY,
            secure=Config.MINIO_SECURE
        )
        print(f"MinIO client initialized successfully (endpoint: {minio_endpoint})")

    def parse_minio_url(self, url: str) -> tuple[str, str]:
        """
        Parse MinIO URL to extract bucket and object key.

        Example URL: https://api.storage.recontext.online/recontext/3177b8ef-.../tracks/TR_AMrS....m3u8
        Returns: ('recontext', '3177b8ef-.../tracks/TR_AMrS....m3u8')

        Args:
            url: MinIO object URL

        Returns:
            Tuple of (bucket_name, object_key)
        """
        # Remove protocol and host
        # URL format: https://api.storage.recontext.online/bucket/object/path
        parts = url.split('/')

        # Find the position after the host
        # Format: ['https:', '', 'api.storage.recontext.online', 'bucket', 'object', 'path', ...]
        if len(parts) >= 4:
            bucket = parts[3]  # First part after host is bucket
            object_key = '/'.join(parts[4:])  # Rest is object key
            return bucket, object_key

        raise ValueError(f"Invalid MinIO URL format: {url}")

    def download_from_minio(self, bucket: str, object_key: str) -> str:
        """
        Download file from MinIO using S3 API.

        Args:
            bucket: MinIO bucket name
            object_key: Object key (path) in bucket

        Returns:
            Path to downloaded temporary file
        """
        print(f"📥 Downloading from MinIO: {bucket}/{object_key}")

        # Determine file suffix from object key
        suffix = '.m3u8' if object_key.endswith('.m3u8') else '.ts'

        # Create temporary file
        temp_file = tempfile.NamedTemporaryFile(delete=False, suffix=suffix)
        temp_path = temp_file.name
        temp_file.close()

        try:
            # Download object from MinIO
            self.minio_client.fget_object(bucket, object_key, temp_path)
            print(f"✅ Downloaded to: {temp_path}")
            return temp_path
        except Exception as e:
            # Clean up temp file on error
            if os.path.exists(temp_path):
                os.remove(temp_path)
            raise RuntimeError(f"Failed to download from MinIO: {e}")

    def download_m3u8_and_combine_from_minio(self, m3u8_url: str) -> str:
        """
        Download m3u8 playlist from MinIO and combine all segments into a single audio file.

        Args:
            m3u8_url: MinIO URL to m3u8 playlist

        Returns:
            Path to combined audio file
        """
        print(f"📥 Downloading and combining m3u8 from MinIO: {m3u8_url}")

        # Parse URL to get bucket and object key
        bucket, playlist_key = self.parse_minio_url(m3u8_url)

        # Download m3u8 playlist
        playlist_path = self.download_from_minio(bucket, playlist_key)

        try:
            # Read playlist content
            with open(playlist_path, 'r') as f:
                playlist_content = f.read()

            print(f"📄 Playlist content:\n{playlist_content[:500]}...")

            # Parse m3u8 to get segment paths
            base_path = playlist_key.rsplit('/', 1)[0]  # Get base path without filename
            segments = []

            for line in playlist_content.split('\n'):
                line = line.strip()
                # Skip comments and empty lines
                if line and not line.startswith('#'):
                    # Build full object key for segment
                    if line.startswith('http'):
                        # If absolute URL, parse it
                        _, segment_key = self.parse_minio_url(line)
                        segments.append(segment_key)
                    else:
                        # Relative path - combine with base path
                        segment_key = f"{base_path}/{line}"
                        segments.append(segment_key)

            print(f"📦 Found {len(segments)} segments to download")

            if not segments:
                raise ValueError("No segments found in m3u8 playlist")

            # Create temporary directory for segments
            temp_dir = tempfile.mkdtemp()
            segment_files = []

            try:
                # Download all segments from MinIO
                for i, segment_key in enumerate(segments):
                    print(f"📥 Downloading segment {i+1}/{len(segments)}: {segment_key}")

                    # Download segment from MinIO
                    segment_path = os.path.join(temp_dir, f"segment_{i:04d}.ts")
                    self.minio_client.fget_object(bucket, segment_key, segment_path)

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

                # Use ffmpeg to concatenate MPEG-TS segments and extract audio
                # -c:a aac -b:a 128k: Re-encode audio to AAC (needed for proper TS handling)
                # -vn: No video output
                ffmpeg_cmd = [
                    'ffmpeg',
                    '-f', 'concat',
                    '-safe', '0',
                    '-i', concat_file_path,
                    '-vn',  # No video
                    '-c:a', 'aac',  # Re-encode audio to AAC
                    '-b:a', '128k',  # Audio bitrate
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
                    concat_file_path_check = os.path.join(temp_dir, 'concat.txt')
                    if os.path.exists(concat_file_path_check):
                        os.remove(concat_file_path_check)
                except:
                    pass

                # Cleanup temp directory
                try:
                    os.rmdir(temp_dir)
                except:
                    pass

        finally:
            # Clean up playlist file
            if os.path.exists(playlist_path):
                os.remove(playlist_path)

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

            # Use ffmpeg to concatenate MPEG-TS segments and extract audio
            # -f concat: Use concat demuxer
            # -safe 0: Allow absolute paths
            # -i: Input concat file
            # -vn: No video output
            # -c:a aac -b:a 128k: Re-encode audio to AAC (needed for proper TS handling)
            # -y: Overwrite output file
            ffmpeg_cmd = [
                'ffmpeg',
                '-f', 'concat',
                '-safe', '0',
                '-i', concat_file_path,
                '-vn',  # No video
                '-c:a', 'aac',  # Re-encode audio to AAC
                '-b:a', '128k',  # Audio bitrate
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
        Uses MinIO S3 client for storage URLs, HTTP for external URLs.

        Args:
            audio_url: URL to download audio from
            token: Optional authentication token (unused for MinIO URLs)

        Returns:
            Path to downloaded temporary file
        """
        # Check if this is a MinIO storage URL
        is_minio_url = 'storage.recontext.online' in audio_url or Config.MINIO_ENDPOINT in audio_url

        # Check if this is an m3u8 playlist
        if audio_url.endswith('.m3u8') or 'playlist.m3u8' in audio_url:
            if is_minio_url:
                # Use MinIO client for storage URLs
                return self.download_m3u8_and_combine_from_minio(audio_url)
            else:
                # Use HTTP for external URLs
                return self.download_m3u8_and_combine(audio_url, token)

        # For non-m3u8 files
        if is_minio_url:
            # Download single file from MinIO
            bucket, object_key = self.parse_minio_url(audio_url)
            return self.download_from_minio(bucket, object_key)

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

        # Check file size - if too small, likely empty audio
        file_size = os.path.getsize(audio_path) / (1024 * 1024)  # MB
        if file_size < 0.01:  # Less than 10KB
            print(f"⚠️  Audio file is very small ({file_size:.4f} MB), may be empty or too short")
            print("Returning empty transcription")
            return []

        try:
            # Transcribe with Faster Whisper
            # If language not specified and auto-detection fails on short audio, default to Russian
            if not language:
                language = None  # Auto-detect

            segments, info = self.model.transcribe(
                audio_path,
                language=language,
                beam_size=5,
                vad_filter=True,  # Voice activity detection
                vad_parameters=dict(min_silence_duration_ms=500),
            )

            detected_language = info.language if hasattr(info, 'language') else 'unknown'
            language_prob = info.language_probability if hasattr(info, 'language_probability') else 0.0
            print(f"Detected language: {detected_language} (probability: {language_prob:.2f})")

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

        except ValueError as e:
            if "empty sequence" in str(e):
                print(f"⚠️  Audio appears to be silent or too short for language detection")
                print("Attempting transcription with Russian language specified...")
                # Retry with Russian language specified
                segments, info = self.model.transcribe(
                    audio_path,
                    language='ru',  # Default to Russian
                    beam_size=5,
                    vad_filter=True,
                    vad_parameters=dict(min_silence_duration_ms=500),
                )

                # Convert segments to phrases
                phrases = []
                for segment in segments:
                    phrases.append({
                        'start': segment.start,
                        'end': segment.end,
                        'text': segment.text.strip(),
                        'confidence': segment.avg_logprob,
                        'language': 'ru'
                    })

                print(f"Transcription completed: {len(phrases)} phrases")
                return phrases
            else:
                raise

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

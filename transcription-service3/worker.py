"""RabbitMQ consumer worker for transcription service."""
import json
import traceback
import time
import os
import tempfile
import pika
from minio import Minio
from config import Config
from transcriber import TranscriptionWorker


class TranscriptionConsumer:
    """RabbitMQ consumer for transcription tasks."""

    def __init__(self):
        """Initialize consumer."""
        self.transcriber = TranscriptionWorker()
        self.connection = None
        self.channel = None

        # Initialize MinIO client for uploading JSON files
        print(f"Initializing MinIO client for JSON uploads: {Config.MINIO_ENDPOINT}")
        minio_endpoint = Config.MINIO_ENDPOINT
        if minio_endpoint.startswith('https://'):
            minio_endpoint = minio_endpoint.replace('https://', '')
        elif minio_endpoint.startswith('http://'):
            minio_endpoint = minio_endpoint.replace('http://', '')

        if ':' not in minio_endpoint:
            minio_endpoint = f"{minio_endpoint}:9000"

        self.minio_client = Minio(
            minio_endpoint,
            access_key=Config.MINIO_ACCESS_KEY,
            secret_key=Config.MINIO_SECRET_KEY,
            secure=Config.MINIO_SECURE
        )
        print(f"MinIO client initialized for uploads (endpoint: {minio_endpoint})")

    def connect(self, max_retries=5, retry_delay=5):
        """Connect to RabbitMQ with retry logic."""
        print(f"Connecting to RabbitMQ at {Config.RABBITMQ_HOST}:{Config.RABBITMQ_PORT}")

        for attempt in range(1, max_retries + 1):
            try:
                self.connection = pika.BlockingConnection(Config.get_rabbitmq_connection_params())
                self.channel = self.connection.channel()

                # Declare transcription queue (input)
                self.channel.queue_declare(queue=Config.RABBITMQ_QUEUE, durable=True)

                # Declare result queue (output)
                self.channel.queue_declare(queue=Config.RABBITMQ_RESULT_QUEUE, durable=True)

                # Declare notification queue (legacy)
                self.channel.queue_declare(queue='transcription_completed', durable=True)
                print(f"✅ Connected to RabbitMQ")
                print(f"   Input queue: {Config.RABBITMQ_QUEUE}")
                print(f"   Result queue: {Config.RABBITMQ_RESULT_QUEUE}")
                return
            except pika.exceptions.AMQPConnectionError as e:
                if attempt < max_retries:
                    print(f"⚠️ Failed to connect to RabbitMQ (attempt {attempt}/{max_retries}): {e}")
                    print(f"Retrying in {retry_delay} seconds...")
                    time.sleep(retry_delay)
                else:
                    print(f"❌ Failed to connect to RabbitMQ after {max_retries} attempts")
                    raise

    def send_transcription_completed_notification(self, track_id):
        """Send notification that transcription is completed."""
        message = {
            'track_id': track_id,
            'event': 'transcription_completed'
        }

        self.channel.basic_publish(
            exchange='',
            routing_key='transcription_completed',
            body=json.dumps(message),
            properties=pika.BasicProperties(
                delivery_mode=2,  # Make message persistent
                content_type='application/json'
            )
        )
        print(f"📢 Sent transcription completed notification for track {track_id}")

    def save_json_to_minio(self, track_id, user_id, audio_url, phrases):
        """
        Save transcription JSON to MinIO in the track folder.

        Args:
            track_id: Track UUID
            user_id: User UUID
            audio_url: Original audio URL to extract path
            phrases: List of transcription phrases

        Returns:
            URL to the saved JSON file
        """
        try:
            # Parse the audio URL to extract bucket and base path
            # Example URL: https://api.storage.recontext.online/recontext/user-id/tracks/track-id/playlist.m3u8
            url_parts = audio_url.split('/')
            bucket = url_parts[3] if len(url_parts) > 3 else 'recontext'

            # Find the tracks folder position and build path to track folder
            # Format: user-id/tracks/track-id/
            if 'tracks' in url_parts:
                tracks_index = url_parts.index('tracks')
                # Path up to and including track ID
                base_path = '/'.join(url_parts[4:tracks_index + 2])
                json_object_key = f"{base_path}/transcription.json"
            else:
                # Fallback: create path from user_id and track_id
                json_object_key = f"{user_id}/tracks/{track_id}/transcription.json"

            # Prepare JSON content
            json_content = {
                'track_id': track_id,
                'user_id': user_id,
                'phrases': phrases,
                'phrase_count': len(phrases),
                'total_duration': phrases[-1]['end'] if phrases else 0.0,
                'created_at': time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime())
            }

            # Convert to JSON string
            json_str = json.dumps(json_content, indent=2, ensure_ascii=False)

            # Create temporary file
            temp_file = tempfile.NamedTemporaryFile(mode='w', delete=False, suffix='.json', encoding='utf-8')
            temp_file.write(json_str)
            temp_file.close()

            try:
                # Upload to MinIO
                print(f"📤 Uploading transcription JSON to MinIO: {bucket}/{json_object_key}")
                self.minio_client.fput_object(
                    bucket,
                    json_object_key,
                    temp_file.name,
                    content_type='application/json'
                )

                # Construct public URL
                protocol = 'https' if Config.MINIO_SECURE else 'http'
                endpoint = Config.MINIO_ENDPOINT
                if endpoint.startswith('https://'):
                    endpoint = endpoint.replace('https://', '')
                elif endpoint.startswith('http://'):
                    endpoint = endpoint.replace('http://', '')

                json_url = f"{protocol}://{endpoint}/{bucket}/{json_object_key}"
                print(f"✅ JSON uploaded successfully: {json_url}")
                return json_url

            finally:
                # Clean up temp file
                if os.path.exists(temp_file.name):
                    os.remove(temp_file.name)

        except Exception as e:
            print(f"❌ Failed to save JSON to MinIO: {e}")
            raise

    def send_result_to_queue(self, track_id, user_id, audio_url, phrases, json_url):
        """
        Send transcription result back to RabbitMQ result queue.

        Args:
            track_id: Track UUID
            user_id: User UUID
            audio_url: Audio URL
            phrases: List of transcription phrases
            json_url: URL to the JSON file in MinIO
        """
        try:
            result_payload = {
                'event': 'transcription_completed',
                'track_id': track_id,
                'user_id': user_id,
                'audio_url': audio_url,
                'json_url': json_url,
                'transcription': {
                    'phrases': phrases,
                    'phrase_count': len(phrases),
                    'total_duration': phrases[-1]['end'] if phrases else 0.0
                },
                'timestamp': time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime()),
                'status': 'completed'
            }

            print(f"📤 Sending result to RabbitMQ queue: {Config.RABBITMQ_RESULT_QUEUE}")
            self.channel.basic_publish(
                exchange='',
                routing_key=Config.RABBITMQ_RESULT_QUEUE,
                body=json.dumps(result_payload),
                properties=pika.BasicProperties(
                    delivery_mode=2,  # Make message persistent
                    content_type='application/json'
                )
            )
            print(f"✅ Result sent to queue successfully")

        except Exception as e:
            print(f"⚠️  Failed to send result to queue: {e}")
            # Don't raise - result queue failure shouldn't fail the transcription

    def process_message(self, ch, method, properties, body):
        """
        Process incoming transcription message.

        Expected message format:
        {
            "track_id": "uuid",
            "user_id": "uuid",
            "audio_url": "http://...",
            "language": "en" (optional)
        }
        """
        try:
            # Parse message
            message = json.loads(body)
            track_id = message.get('track_id')
            user_id = message.get('user_id')
            audio_url = message.get('audio_url')
            language = message.get('language')
            token = message.get('token')  # Optional auth token

            print(f"\n{'='*60}")
            print(f"Processing transcription task:")
            print(f"  Track ID: {track_id}")
            print(f"  User ID: {user_id}")
            print(f"  Audio URL: {audio_url}")
            print(f"  Language: {language or 'auto-detect'}")
            print(f"{'='*60}\n")

            # Validate required fields
            if not track_id or not user_id or not audio_url:
                raise ValueError("Missing required fields: track_id, user_id, or audio_url")

            # Transcribe audio
            phrases = self.transcriber.transcribe_from_url(
                audio_url=audio_url,
                language=language,
                token=token
            )

            # Calculate total duration and phrase count
            total_duration = 0
            if phrases:
                total_duration = phrases[-1]['end']
            phrase_count = len(phrases)

            print(f"\n{'='*60}")
            print(f"Transcription completed successfully!")
            print(f"  Phrases: {phrase_count}")
            print(f"  Duration: {total_duration:.2f}s")
            print(f"{'='*60}\n")

            # Save JSON to MinIO
            try:
                json_url = self.save_json_to_minio(track_id, user_id, audio_url, phrases)
            except Exception as e:
                print(f"⚠️  Failed to save JSON to MinIO: {e}")
                json_url = None

            # Send result to RabbitMQ result queue
            try:
                self.send_result_to_queue(track_id, user_id, audio_url, phrases, json_url)
            except Exception as e:
                print(f"⚠️  Failed to send result to queue: {e}")

            # Send legacy notification about transcription completion
            try:
                self.send_transcription_completed_notification(track_id)
            except Exception as e:
                print(f"⚠️  Failed to send legacy completion notification: {e}")

            # Acknowledge message
            ch.basic_ack(delivery_tag=method.delivery_tag)

        except Exception as e:
            error_str = str(e)
            # Compact single-line error for missing files
            if "NoSuchKey" in error_str or "Object does not exist" in error_str:
                print(f"ERROR: Transcription failed - Audio file not found in storage: {error_str.split('object_name:')[-1].strip() if 'object_name:' in error_str else 'unknown'}")
            else:
                error_message = f"Transcription failed: {error_str}\n{traceback.format_exc()}"
                print(f"\n{'='*60}")
                print(f"ERROR: {error_message}")
                print(f"{'='*60}\n")

            # Reject message (don't requeue to avoid infinite loop)
            ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)

    def start(self):
        """Start consuming messages."""
        print("\n" + "="*60)
        print("Transcription Service Starting")
        print("="*60)

        # Connect to RabbitMQ with retry
        try:
            self.connect()
        except pika.exceptions.AMQPConnectionError:
            print("\n" + "="*60)
            print("⚠️ RabbitMQ is not available")
            print("Service will keep running in standby mode")
            print("Waiting for RabbitMQ to become available...")
            print("="*60 + "\n")

            # Keep trying to connect in a loop with longer delays
            while True:
                try:
                    time.sleep(30)  # Wait 30 seconds between connection attempts
                    print("Attempting to reconnect to RabbitMQ...")
                    self.connect(max_retries=1, retry_delay=0)
                    break  # Successfully connected, exit the loop
                except pika.exceptions.AMQPConnectionError:
                    print("Still waiting for RabbitMQ...")
                    continue
                except KeyboardInterrupt:
                    print("\n\nShutting down...")
                    return

        # Set QoS to process one message at a time
        self.channel.basic_qos(prefetch_count=1)

        # Start consuming
        self.channel.basic_consume(
            queue=Config.RABBITMQ_QUEUE,
            on_message_callback=self.process_message
        )

        print("\n" + "="*60)
        print("Waiting for transcription tasks...")
        print("Press CTRL+C to exit")
        print("="*60 + "\n")

        try:
            self.channel.start_consuming()
        except KeyboardInterrupt:
            print("\n\nShutting down gracefully...")
            self.channel.stop_consuming()
            self.connection.close()
            print("Service stopped")


if __name__ == '__main__':
    consumer = TranscriptionConsumer()
    consumer.start()

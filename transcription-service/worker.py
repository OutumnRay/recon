"""RabbitMQ consumer worker for transcription service."""
import json
import traceback
import time
import pika
from config import Config
from database import DatabaseManager
from transcriber import TranscriptionWorker


class TranscriptionConsumer:
    """RabbitMQ consumer for transcription tasks."""

    def __init__(self):
        """Initialize consumer."""
        self.db = DatabaseManager()
        self.transcriber = TranscriptionWorker()
        self.connection = None
        self.channel = None

    def connect(self, max_retries=5, retry_delay=5):
        """Connect to RabbitMQ with retry logic."""
        print(f"Connecting to RabbitMQ at {Config.RABBITMQ_HOST}:{Config.RABBITMQ_PORT}")

        for attempt in range(1, max_retries + 1):
            try:
                self.connection = pika.BlockingConnection(Config.get_rabbitmq_connection_params())
                self.channel = self.connection.channel()

                # Declare transcription queue
                self.channel.queue_declare(queue=Config.RABBITMQ_QUEUE, durable=True)

                # Declare notification queue
                self.channel.queue_declare(queue='transcription_completed', durable=True)
                print(f"✅ Connected to RabbitMQ queue: {Config.RABBITMQ_QUEUE}")
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

            # Update status to processing
            self.db.update_transcription_status(track_id, 'processing')

            # Transcribe audio
            phrases = self.transcriber.transcribe_from_url(
                audio_url=audio_url,
                language=language,
                token=token
            )

            # Calculate total duration
            total_duration = 0
            if phrases:
                total_duration = phrases[-1]['end']

            # Save to database
            phrase_count = self.db.save_transcription_phrases(
                track_id=track_id,
                user_id=user_id,
                phrases=phrases
            )

            # Update status to completed
            self.db.update_transcription_status(
                track_id=track_id,
                status='completed',
                phrase_count=phrase_count,
                total_duration=total_duration
            )

            # Mark track as ready
            self.db.mark_track_ready(track_id)

            print(f"\n{'='*60}")
            print(f"Transcription completed successfully!")
            print(f"  Phrases: {phrase_count}")
            print(f"  Duration: {total_duration:.2f}s")
            print(f"{'='*60}\n")

            # Send notification about transcription completion
            try:
                self.send_transcription_completed_notification(track_id)
            except Exception as e:
                print(f"⚠️  Failed to send completion notification: {e}")

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

            # Update status to failed
            try:
                track_id = json.loads(body).get('track_id')
                if track_id:
                    self.db.update_transcription_status(
                        track_id=track_id,
                        status='failed',
                        error_message=str(e)
                    )
            except:
                pass

            # Reject message (don't requeue to avoid infinite loop)
            ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)

    def start(self):
        """Start consuming messages."""
        print("\n" + "="*60)
        print("Transcription Service Starting")
        print("="*60)

        # Initialize database tables
        print("Initializing database tables...")
        self.db.create_transcription_tables()
        print("Database tables ready")

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
            self.db.close()
            print("Service stopped")


if __name__ == '__main__':
    consumer = TranscriptionConsumer()
    consumer.start()

"""Configuration for transcription service."""
import os
from dotenv import load_dotenv

load_dotenv()


class Config:
    """Application configuration."""

    # RabbitMQ Configuration
    RABBITMQ_HOST = os.getenv('RABBITMQ_HOST', '5.129.227.21')
    RABBITMQ_PORT = int(os.getenv('RABBITMQ_PORT', '5672'))
    RABBITMQ_USER = os.getenv('RABBITMQ_USER', 'recontext')
    RABBITMQ_PASSWORD = os.getenv('RABBITMQ_PASSWORD', 'je9rO4k6CQ3M')
    RABBITMQ_QUEUE = os.getenv('RABBITMQ_QUEUE', 'transcription_queue')
    RABBITMQ_RESULT_QUEUE = os.getenv('RABBITMQ_RESULT_QUEUE', 'transcription_results')

    # Whisper Configuration
    WHISPER_MODEL = os.getenv('WHISPER_MODEL', 'large')
    WHISPER_DEVICE = os.getenv('WHISPER_DEVICE', 'cuda')  # cpu or cuda
    WHISPER_COMPUTE_TYPE = os.getenv('WHISPER_COMPUTE_TYPE', 'float16')  # int8, float16, float32

    # Storage Configuration
    MINIO_ENDPOINT = os.getenv('MINIO_ENDPOINT', 'api.storage.recontext.online')
    MINIO_ACCESS_KEY = os.getenv('MINIO_ACCESS_KEY', 'minioadmin')
    MINIO_SECRET_KEY = os.getenv('MINIO_SECRET_KEY', '32a4953d5bff4a1c6aea4d4ccfb757e5')
    MINIO_SECURE = os.getenv('MINIO_SECURE', 'true').lower() == 'true'  # Use HTTPS by default

    @classmethod
    def get_rabbitmq_connection_params(cls):
        """Get RabbitMQ connection parameters with extended heartbeat for long-running tasks."""
        import pika
        return pika.ConnectionParameters(
            host=cls.RABBITMQ_HOST,
            port=cls.RABBITMQ_PORT,
            credentials=pika.PlainCredentials(cls.RABBITMQ_USER, cls.RABBITMQ_PASSWORD),
            # Heartbeat interval in seconds (0 = disable, default 60)
            # Set to 3600 (1 hour) to support long transcription tasks
            heartbeat=3600,
            # Connection timeout in seconds
            blocked_connection_timeout=300
        )

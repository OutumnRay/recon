"""Configuration for transcription service."""
import os
from dotenv import load_dotenv

load_dotenv()


class Config:
    """Application configuration."""

    # RabbitMQ Configuration
    RABBITMQ_HOST = os.getenv('RABBITMQ_HOST', 'localhost')
    RABBITMQ_PORT = int(os.getenv('RABBITMQ_PORT', '5672'))
    RABBITMQ_USER = os.getenv('RABBITMQ_USER', 'guest')
    RABBITMQ_PASSWORD = os.getenv('RABBITMQ_PASSWORD', 'guest')
    RABBITMQ_QUEUE = os.getenv('RABBITMQ_QUEUE', 'transcription_queue')

    # Database Configuration
    DB_HOST = os.getenv('DB_HOST', 'localhost')
    DB_PORT = int(os.getenv('DB_PORT', '5432'))
    DB_NAME = os.getenv('DB_NAME', 'recontext')
    DB_USER = os.getenv('DB_USER', 'postgres')
    DB_PASSWORD = os.getenv('DB_PASSWORD', 'postgres')

    # Whisper Configuration
    WHISPER_MODEL = os.getenv('WHISPER_MODEL', 'medium')
    WHISPER_DEVICE = os.getenv('WHISPER_DEVICE', 'cpu')  # cpu or cuda
    WHISPER_COMPUTE_TYPE = os.getenv('WHISPER_COMPUTE_TYPE', 'int8')  # int8, float16, float32

    # Storage Configuration
    STORAGE_BASE_URL = os.getenv('STORAGE_BASE_URL', 'http://localhost:9000')
    MINIO_ENDPOINT = os.getenv('MINIO_ENDPOINT', 'api.storage.recontext.online:9000')
    MINIO_ACCESS_KEY = os.getenv('MINIO_ACCESS_KEY', 'minioadmin')
    MINIO_SECRET_KEY = os.getenv('MINIO_SECRET_KEY', 'minioadmin')
    MINIO_SECURE = os.getenv('MINIO_SECURE', 'true').lower() == 'true'  # Use HTTPS by default

    @classmethod
    def get_db_connection_string(cls):
        """Get PostgreSQL connection string."""
        return f"host={cls.DB_HOST} port={cls.DB_PORT} dbname={cls.DB_NAME} user={cls.DB_USER} password={cls.DB_PASSWORD}"

    @classmethod
    def get_rabbitmq_connection_params(cls):
        """Get RabbitMQ connection parameters."""
        import pika
        return pika.ConnectionParameters(
            host=cls.RABBITMQ_HOST,
            port=cls.RABBITMQ_PORT,
            credentials=pika.PlainCredentials(cls.RABBITMQ_USER, cls.RABBITMQ_PASSWORD)
        )

"""
Конфигурация для Jitsi Recorder
"""
import os


class Config:
    """Настройки приложения"""

    # Jitsi Meet
    JITSI_DOMAIN = os.getenv('JITSI_DOMAIN', 'meet.recontext.online')
    JITSI_URL = os.getenv('JITSI_URL', 'https://meet.recontext.online')

    # Prosody XMPP
    PROSODY_HOST = os.getenv('PROSODY_HOST', 'prosody')
    PROSODY_PORT = int(os.getenv('PROSODY_PORT', '5222'))
    XMPP_DOMAIN = os.getenv('XMPP_DOMAIN', 'meet.jitsi')
    MUC_DOMAIN = os.getenv('MUC_DOMAIN', 'muc.meet.jitsi')

    # JVB
    JVB_HOST = os.getenv('JVB_HOST', 'jvb')
    JVB_PORT = os.getenv('JVB_PORT', '8080')

    # S3 Storage
    S3_ENDPOINT = os.getenv('S3_ENDPOINT', 'https://api.storage.recontext.online')
    S3_BUCKET = os.getenv('S3_BUCKET', 'jitsi-recordings')
    AWS_ACCESS_KEY_ID = os.getenv('AWS_ACCESS_KEY_ID', 'minioadmin')
    AWS_SECRET_ACCESS_KEY = os.getenv('AWS_SECRET_ACCESS_KEY', 'minioadmin')
    AWS_REGION = os.getenv('AWS_REGION', 'us-east-1')

    # Webhook
    WEBHOOK_URL = os.getenv('WEBHOOK_URL', '')

    # Redis
    REDIS_HOST = os.getenv('REDIS_HOST', 'redis')
    REDIS_PORT = int(os.getenv('REDIS_PORT', '6379'))
    REDIS_ENABLED = os.getenv('REDIS_ENABLED', 'true').lower() == 'true'

    # Recording
    RECORD_DIR = os.getenv('RECORD_DIR', '/tmp/recordings')
    AUDIO_BITRATE = os.getenv('AUDIO_BITRATE', '48k')
    CHECK_INTERVAL = int(os.getenv('CHECK_INTERVAL', '5'))
    RECONNECT_TIMEOUT = int(os.getenv('RECONNECT_TIMEOUT', '30'))

    # Recorder bot identity
    BOT_NAME = os.getenv('BOT_NAME', 'Recorder Bot')
    BOT_NICKNAME = os.getenv('BOT_NICKNAME', 'recorder')

    # HTTP Server
    HTTP_HOST = os.getenv('HTTP_HOST', '0.0.0.0')
    HTTP_PORT = int(os.getenv('HTTP_PORT', '8080'))

    # Logging
    LOG_LEVEL = os.getenv('LOG_LEVEL', 'DEBUG')

    # Worker ID
    WORKER_ID = os.getenv('HOSTNAME', None)

    @classmethod
    def validate(cls):
        """Проверка обязательных параметров"""
        required = []

        # S3 обязателен для production
        if not cls.AWS_ACCESS_KEY_ID or not cls.AWS_SECRET_ACCESS_KEY:
            print("⚠️  WARNING: S3 credentials not set - uploads will be disabled")

        if not cls.S3_ENDPOINT:
            print("⚠️  WARNING: S3_ENDPOINT not set")

        return True

    @classmethod
    def print_config(cls):
        """Вывод текущей конфигурации"""
        print("🔧 Configuration:")
        print(f"  JITSI_URL: {cls.JITSI_URL}")
        print(f"  JITSI_DOMAIN: {cls.JITSI_DOMAIN}")
        print(f"  PROSODY_HOST: {cls.PROSODY_HOST}:{cls.PROSODY_PORT}")
        print(f"  JVB_HOST: {cls.JVB_HOST}:{cls.JVB_PORT}")
        print(f"  REDIS: {cls.REDIS_HOST}:{cls.REDIS_PORT} (enabled: {cls.REDIS_ENABLED})")
        print(f"  S3_BUCKET: {cls.S3_BUCKET}")
        print(f"  S3_ENDPOINT: {cls.S3_ENDPOINT}")
        print(f"  WEBHOOK_URL: {cls.WEBHOOK_URL if cls.WEBHOOK_URL else 'NOT SET (optional)'}")
        print(f"  RECORD_DIR: {cls.RECORD_DIR}")
        print(f"  LOG_LEVEL: {cls.LOG_LEVEL}")


# Создаем singleton instance
config = Config()

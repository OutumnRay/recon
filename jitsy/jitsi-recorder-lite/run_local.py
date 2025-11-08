#!/usr/bin/env python3
"""
Локальный запуск Jitsi Recorder для тестирования
"""
import asyncio
import logging
import sys
from pathlib import Path

# Загружаем переменные из .env если файл существует
try:
    from dotenv import load_dotenv
    env_file = Path(__file__).parent / '.env'
    if env_file.exists():
        load_dotenv(env_file)
        print(f"✅ Loaded environment from {env_file}")
    else:
        print(f"⚠️  No .env file found at {env_file}")
        print("   Using default configuration")
except ImportError:
    print("⚠️  python-dotenv not installed (pip install python-dotenv)")
    print("   Using system environment variables")

from config import config
from jitsi_participant_recorder import JitsiParticipantRecorder


async def main(room_name: str):
    """Запуск recorder для одной комнаты"""

    # Настройка логирования
    logging.basicConfig(
        level=getattr(logging, config.LOG_LEVEL),
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )

    logger = logging.getLogger(__name__)

    # Вывод конфигурации
    config.print_config()
    print()

    # Создаем output директорию
    Path(config.RECORD_DIR).mkdir(parents=True, exist_ok=True)

    # Создаем recorder
    recorder = JitsiParticipantRecorder(
        room_name=room_name,
        jitsi_url=config.JITSI_URL,
        output_dir=config.RECORD_DIR,
        bot_name=config.BOT_NAME
    )

    try:
        logger.info(f"🎙️  Starting recorder for room: {room_name}")
        await recorder.connect()

        # Записываем пока не прервут
        logger.info(f"📹 Recording... (Press Ctrl+C to stop)")

        while True:
            await asyncio.sleep(1)

    except KeyboardInterrupt:
        logger.info("⚠️  Interrupted by user")
    except Exception as e:
        logger.error(f"❌ Error: {e}", exc_info=True)
    finally:
        await recorder.disconnect()
        logger.info("✅ Recording stopped")


if __name__ == "__main__":
    print("=" * 60)
    print("🎙️  Jitsi Recorder - Local Mode")
    print("=" * 60)
    print()

    if len(sys.argv) < 2:
        print("Usage: python run_local.py <room_name>")
        print()
        print("Example:")
        print("  python run_local.py testmeet")
        print()
        print("This will connect to: https://meet.recontext.online/testmeet")
        print("And record all audio tracks to ./recordings/")
        print()
        sys.exit(1)

    room_name = sys.argv[1]

    try:
        asyncio.run(main(room_name))
    except KeyboardInterrupt:
        print("\n👋 Goodbye!")

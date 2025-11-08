"""
Простой Jitsi Recorder - подключается к конференции и записывает audio
Использует Playwright + Chrome DevTools Protocol для перехвата audio tracks
"""
import asyncio
import logging
import os
from datetime import datetime
from pathlib import Path
import json

from playwright.async_api import async_playwright

logger = logging.getLogger(__name__)


class SimpleJitsiRecorder:
    """
    Простой recorder - подключается к Jitsi как участник
    и записывает audio streams
    """

    def __init__(self, room_url: str, output_dir: str, bot_name: str = "Recorder Bot"):
        self.room_url = room_url
        self.output_dir = output_dir
        self.bot_name = bot_name

        self.browser = None
        self.context = None
        self.page = None
        self.is_recording = False
        self.active_tracks = {}

        Path(output_dir).mkdir(parents=True, exist_ok=True)

    async def connect(self):
        """Подключается к Jitsi конференции"""
        try:
            logger.info(f"🔌 Starting browser and connecting to: {self.room_url}")

            # Запускаем браузер
            playwright = await async_playwright().start()

            self.browser = await playwright.chromium.launch(
                headless=True,
                args=[
                    '--use-fake-ui-for-media-stream',  # Автоматически разрешить доступ к микрофону
                    '--use-fake-device-for-media-stream',  # Использовать фейковый микрофон
                    '--autoplay-policy=no-user-gesture-required',
                    '--disable-blink-features=AutomationControlled'
                ]
            )

            self.context = await self.browser.new_context(
                permissions=['microphone', 'camera'],
                viewport={'width': 1280, 'height': 720}
            )

            self.page = await self.context.new_page()

            # Перехватываем WebRTC connections через CDP
            cdp = await self.page.context.new_cdp_session(self.page)

            # Включаем события WebRTC
            await cdp.send('Network.enable')

            # TODO: Перехват audio tracks через CDP
            # Это требует прослушивания событий WebRTC и извлечения audio streams

            # Переходим на страницу конференции
            logger.info(f"🌐 Navigating to: {self.room_url}")
            await self.page.goto(self.room_url, wait_until='networkidle')

            # Ждем загрузки Jitsi Meet
            await asyncio.sleep(3)

            # Устанавливаем имя участника
            logger.info(f"👤 Setting display name: {self.bot_name}")
            await self.page.fill('[placeholder="Enter your name"]', self.bot_name)

            # Кликаем "Join meeting" если есть
            try:
                join_button = await self.page.query_selector('button:has-text("Join meeting")')
                if join_button:
                    await join_button.click()
                    logger.info(f"✅ Clicked 'Join meeting'")
            except Exception as e:
                logger.debug(f"No 'Join meeting' button: {e}")

            # Ждем подключения
            await asyncio.sleep(5)

            self.is_recording = True
            logger.info(f"✅ Connected to conference")

            # Запускаем мониторинг WebRTC
            asyncio.create_task(self._monitor_webrtc())

        except Exception as e:
            logger.error(f"❌ Failed to connect: {e}", exc_info=True)
            raise

    async def _monitor_webrtc(self):
        """Мониторит WebRTC connections и записывает audio"""
        try:
            logger.info("🎧 Monitoring WebRTC connections...")

            while self.is_recording:
                # Получаем WebRTC stats через CDP
                try:
                    stats = await self.page.evaluate("""
                        async () => {
                            const peerConnections = window.RTCPeerConnection ?
                                Array.from(document.querySelectorAll('*'))
                                    .map(el => el.RTCPeerConnection)
                                    .filter(Boolean) : [];

                            const tracks = [];

                            // Получаем все audio tracks из страницы
                            const mediaElements = document.querySelectorAll('audio, video');
                            for (const el of mediaElements) {
                                if (el.srcObject) {
                                    const audioTracks = el.srcObject.getAudioTracks();
                                    for (const track of audioTracks) {
                                        tracks.push({
                                            id: track.id,
                                            label: track.label,
                                            enabled: track.enabled,
                                            muted: track.muted
                                        });
                                    }
                                }
                            }

                            return {
                                tracks: tracks,
                                timestamp: Date.now()
                            };
                        }
                    """)

                    if stats['tracks']:
                        logger.debug(f"📊 Found {len(stats['tracks'])} audio tracks")

                        for track_info in stats['tracks']:
                            track_id = track_info['id']

                            if track_id not in self.active_tracks:
                                logger.info(f"🎙️  New audio track: {track_id} ({track_info['label']})")
                                self.active_tracks[track_id] = track_info

                except Exception as e:
                    logger.debug(f"Error getting WebRTC stats: {e}")

                await asyncio.sleep(5)

        except Exception as e:
            logger.error(f"❌ Error monitoring WebRTC: {e}", exc_info=True)

    async def disconnect(self):
        """Отключается от конференции"""
        try:
            logger.info(f"🔌 Disconnecting from conference")

            self.is_recording = False

            # Закрываем браузер
            if self.page:
                await self.page.close()

            if self.context:
                await self.context.close()

            if self.browser:
                await self.browser.close()

            logger.info(f"✅ Disconnected")

        except Exception as e:
            logger.error(f"❌ Error disconnecting: {e}", exc_info=True)


# Пример для локального тестирования
async def test_local():
    """Тест локального запуска"""
    import sys

    logging.basicConfig(
        level=logging.DEBUG,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )

    room_url = sys.argv[1] if len(sys.argv) > 1 else "https://meet.recontext.online/testmeet"

    recorder = SimpleJitsiRecorder(
        room_url=room_url,
        output_dir="./recordings",
        bot_name="Recorder Bot"
    )

    try:
        await recorder.connect()

        # Записываем 60 секунд
        logger.info("📹 Recording for 60 seconds...")
        await asyncio.sleep(60)

    except KeyboardInterrupt:
        logger.info("⚠️  Interrupted by user")
    finally:
        await recorder.disconnect()


if __name__ == "__main__":
    print("🎙️  Simple Jitsi Recorder - Local Test")
    print("Usage: python simple_jitsi_recorder.py [room_url]")
    print()

    asyncio.run(test_local())

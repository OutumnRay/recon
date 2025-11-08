"""
Jitsi Participant Recorder - подключается к Jitsi Meet как участник
Использует XMPP для signaling и WebRTC для получения audio streams
"""
import asyncio
import logging
import os
from datetime import datetime
from pathlib import Path
import json

from aiortc import RTCPeerConnection, RTCSessionDescription, RTCConfiguration, RTCIceServer
from av import open as av_open

logger = logging.getLogger(__name__)


class JitsiParticipantRecorder:
    """
    Подключается к Jitsi Meet конференции как обычный участник
    Записывает все audio tracks
    """

    def __init__(self, room_name: str, jitsi_url: str, output_dir: str, bot_name: str = "Recorder Bot"):
        self.room_name = room_name
        self.jitsi_url = jitsi_url
        self.output_dir = output_dir
        self.bot_name = bot_name

        self.full_room_url = f"{jitsi_url}/{room_name}"

        self.pc = None
        self.is_connected = False
        self.active_recorders = {}

        Path(output_dir).mkdir(parents=True, exist_ok=True)

    async def connect(self):
        """Подключается к Jitsi Meet конференции"""
        try:
            logger.info(f"🔌 Connecting to Jitsi Meet: {self.full_room_url}")

            # Создаем WebRTC PeerConnection
            configuration = RTCConfiguration(
                iceServers=[
                    RTCIceServer(urls=["stun:stun.l.google.com:19302"]),
                    RTCIceServer(urls=["stun:stun1.l.google.com:19302"])
                ]
            )
            self.pc = RTCPeerConnection(configuration)

            # Обработчик для входящих audio tracks
            @self.pc.on("track")
            async def on_track(track):
                logger.info(f"📥 Received track: kind={track.kind}, id={track.id}")

                if track.kind == "audio":
                    await self._start_recording_track(track)

            # Обработчик ICE connection state
            @self.pc.on("connectionstatechange")
            async def on_connection_state_change():
                logger.info(f"🔗 ICE connection state: {self.pc.connectionState}")

            # TODO: Реализовать XMPP signaling для подключения к Jitsi
            # Это требует:
            # 1. XMPP подключение к Prosody
            # 2. Join в MUC комнату
            # 3. Jingle signaling через XMPP
            # 4. Обмен SDP offer/answer через Jingle IQ

            logger.warning("⚠️  XMPP signaling not implemented yet")
            logger.warning("    This is a placeholder - full implementation requires:")
            logger.warning("    1. XMPP connection (aioxmpp)")
            logger.warning("    2. MUC join")
            logger.warning("    3. Jingle signaling")

            self.is_connected = True
            logger.info(f"✅ Connected to Jitsi Meet (placeholder)")

        except Exception as e:
            logger.error(f"❌ Failed to connect: {e}", exc_info=True)
            raise

    async def _start_recording_track(self, track):
        """Начинает запись audio track"""
        try:
            participant_id = track.id
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S_%f')[:-3]
            safe_room = self._sanitize_filename(self.room_name)
            filename = f"{safe_room}_{participant_id}_{timestamp}.opus"
            filepath = os.path.join(self.output_dir, filename)

            logger.info(f"🎙️  Starting recording: {filename}")

            # Открываем файл для записи
            container = av_open(filepath, 'w')
            stream = container.add_stream('opus', rate=48000, layout='stereo')

            start_time = datetime.now()

            # Записываем frames
            async def record_frames():
                try:
                    while True:
                        frame = await track.recv()

                        # Кодируем и пишем в файл
                        for packet in stream.encode(frame):
                            container.mux(packet)

                except Exception as e:
                    logger.debug(f"Track ended or error: {e}")
                finally:
                    # Финализируем запись
                    for packet in stream.encode():
                        container.mux(packet)
                    container.close()

                    duration = (datetime.now() - start_time).total_seconds()
                    file_size = os.path.getsize(filepath) if os.path.exists(filepath) else 0

                    logger.info(f"⏹️  Stopped recording: {filename} ({duration:.1f}s, {file_size} bytes)")

                    # Удаляем из active recorders
                    if participant_id in self.active_recorders:
                        del self.active_recorders[participant_id]

            # Запускаем запись в фоне
            task = asyncio.create_task(record_frames())
            self.active_recorders[participant_id] = {
                'task': task,
                'filename': filename,
                'filepath': filepath,
                'start_time': start_time
            }

        except Exception as e:
            logger.error(f"❌ Error starting track recording: {e}", exc_info=True)

    async def disconnect(self):
        """Отключается от конференции"""
        try:
            logger.info(f"🔌 Disconnecting from Jitsi Meet")

            self.is_connected = False

            # Останавливаем все активные записи
            for participant_id, info in list(self.active_recorders.items()):
                task = info['task']
                task.cancel()
                try:
                    await task
                except asyncio.CancelledError:
                    pass

            self.active_recorders.clear()

            # Закрываем WebRTC connection
            if self.pc:
                await self.pc.close()

            logger.info(f"✅ Disconnected")

        except Exception as e:
            logger.error(f"❌ Error disconnecting: {e}", exc_info=True)

    @staticmethod
    def _sanitize_filename(name):
        import re
        return re.sub(r'[^\w\-.]', '_', name)


# Пример использования для локального тестирования
async def test_local():
    """Тест локального запуска"""
    import sys

    logging.basicConfig(
        level=logging.DEBUG,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )

    room_name = sys.argv[1] if len(sys.argv) > 1 else "testmeet"

    recorder = JitsiParticipantRecorder(
        room_name=room_name,
        jitsi_url="https://meet.recontext.online",
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
    print("🎙️  Jitsi Participant Recorder - Local Test")
    print("Usage: python jitsi_participant_recorder.py [room_name]")
    print()

    asyncio.run(test_local())

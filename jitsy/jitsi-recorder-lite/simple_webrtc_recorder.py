"""
Простой WebRTC recorder для Jitsi используя aiortc
Подключается к JVB через WebSocket и записывает audio tracks
"""
import asyncio
import logging
import json
import os
from datetime import datetime
from pathlib import Path

import websockets
from aiortc import RTCPeerConnection, RTCSessionDescription, RTCConfiguration, RTCIceServer
from av import open as av_open

logger = logging.getLogger(__name__)


class AudioTrackRecorder:
    """Записывает один audio track в opus файл"""

    def __init__(self, track, participant_id, room_name, output_dir):
        self.track = track
        self.participant_id = participant_id
        self.room_name = room_name
        self.output_dir = output_dir
        self.is_recording = False
        self.container = None
        self.stream = None
        self.start_time = None
        self.filename = None
        self.filepath = None

    async def start(self):
        """Начинает запись"""
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S_%f')[:-3]
        safe_room = self._sanitize_filename(self.room_name)
        safe_participant = self._sanitize_filename(self.participant_id)

        self.filename = f"{safe_room}_{safe_participant}_{timestamp}.opus"
        self.filepath = os.path.join(self.output_dir, self.filename)

        # Открываем файл для записи
        self.container = av_open(self.filepath, 'w')
        self.stream = self.container.add_stream('opus', rate=48000, layout='stereo')

        self.start_time = datetime.now()
        self.is_recording = True

        logger.info(f"🎙️  Started recording: {self.filename}")

        # Запускаем запись frames
        asyncio.create_task(self._record_frames())

    async def _record_frames(self):
        """Записывает frames из track"""
        try:
            while self.is_recording:
                frame = await self.track.recv()

                # Кодируем и пишем в файл
                for packet in self.stream.encode(frame):
                    self.container.mux(packet)

        except Exception as e:
            logger.debug(f"Track ended or error: {e}")
        finally:
            await self.stop()

    async def stop(self):
        """Останавливает запись"""
        if not self.is_recording:
            return

        self.is_recording = False

        # Финализируем запись
        if self.stream:
            for packet in self.stream.encode():
                self.container.mux(packet)

        if self.container:
            self.container.close()

        duration = (datetime.now() - self.start_time).total_seconds() if self.start_time else 0
        file_size = os.path.getsize(self.filepath) if os.path.exists(self.filepath) else 0

        logger.info(f"⏹️  Stopped recording: {self.filename} ({duration:.1f}s, {file_size} bytes)")

        return {
            'filename': self.filename,
            'filepath': self.filepath,
            'duration': duration,
            'participant_id': self.participant_id
        }

    @staticmethod
    def _sanitize_filename(name):
        import re
        return re.sub(r'[^\w\-.]', '_', name)


class SimpleWebRTCRecorder:
    """Простой WebRTC recorder для Jitsi"""

    def __init__(self, room_name, conference_id, jvb_host, jvb_port, output_dir):
        self.room_name = room_name
        self.conference_id = conference_id
        self.jvb_host = jvb_host
        self.jvb_port = jvb_port
        self.output_dir = output_dir

        self.pc = None
        self.ws = None
        self.is_connected = False
        self.active_recorders = {}

    async def connect(self):
        """Подключается к JVB и начинает получать audio tracks"""
        try:
            logger.info(f"🔌 Connecting to JVB: {self.jvb_host}:{self.jvb_port}")

            # 1. Создаем WebRTC PeerConnection
            configuration = RTCConfiguration(
                iceServers=[RTCIceServer(urls=["stun:stun.l.google.com:19302"])]
            )
            self.pc = RTCPeerConnection(configuration)

            # 2. Обработчик для входящих audio tracks
            @self.pc.on("track")
            async def on_track(track):
                logger.info(f"📥 Received track: kind={track.kind}, id={track.id}")

                if track.kind == "audio":
                    participant_id = track.id
                    recorder = AudioTrackRecorder(track, participant_id, self.room_name, self.output_dir)
                    await recorder.start()
                    self.active_recorders[participant_id] = recorder

            # 3. Подключаемся к JVB через WebSocket
            ws_url = f"ws://{self.jvb_host}:{self.jvb_port}/colibri-ws/default-id/{self.conference_id}"
            logger.info(f"🔌 Connecting to WebSocket: {ws_url}")

            self.ws = await websockets.connect(ws_url)

            # 4. Создаем SDP offer
            offer = await self.pc.createOffer()
            await self.pc.setLocalDescription(offer)

            # 5. Отправляем offer в JVB
            message = {
                "colibriClass": "EndpointMessage",
                "type": "offer",
                "sdp": offer.sdp
            }
            await self.ws.send(json.dumps(message))
            logger.info(f"📤 Sent SDP offer to JVB")

            # 6. Обрабатываем сообщения от JVB
            asyncio.create_task(self._handle_jvb_messages())

            # 7. Обработчик для ICE candidates
            @self.pc.on("icecandidate")
            async def on_ice_candidate(candidate):
                if candidate:
                    message = {
                        "colibriClass": "EndpointMessage",
                        "type": "candidate",
                        "candidate": candidate.candidate,
                        "sdpMid": candidate.sdpMid,
                        "sdpMLineIndex": candidate.sdpMLineIndex
                    }
                    await self.ws.send(json.dumps(message))
                    logger.debug(f"📤 Sent ICE candidate")

            self.is_connected = True
            logger.info(f"✅ Connected to JVB")

        except Exception as e:
            logger.error(f"❌ Failed to connect: {e}", exc_info=True)
            raise

    async def _handle_jvb_messages(self):
        """Обрабатывает сообщения от JVB"""
        try:
            async for message in self.ws:
                data = json.loads(message)
                logger.debug(f"📨 JVB message: {data.get('type', 'unknown')}")

                if data.get('type') == 'answer':
                    # Получили SDP answer
                    answer = RTCSessionDescription(sdp=data['sdp'], type="answer")
                    await self.pc.setRemoteDescription(answer)
                    logger.info(f"✅ Set remote description (answer)")

                elif data.get('type') == 'candidate':
                    # Получили ICE candidate
                    # aiortc автоматически обрабатывает ICE candidates через trickle ICE
                    pass

        except websockets.exceptions.ConnectionClosed:
            logger.info(f"🔌 WebSocket connection closed")
        except Exception as e:
            logger.error(f"❌ Error handling JVB messages: {e}", exc_info=True)

    async def disconnect(self):
        """Отключается от JVB"""
        try:
            logger.info(f"🔌 Disconnecting from JVB")

            self.is_connected = False

            # Останавливаем все записи
            for participant_id, recorder in list(self.active_recorders.items()):
                await recorder.stop()

            self.active_recorders.clear()

            # Закрываем WebRTC connection
            if self.pc:
                await self.pc.close()

            # Закрываем WebSocket
            if self.ws:
                await self.ws.close()

            logger.info(f"✅ Disconnected")

        except Exception as e:
            logger.error(f"❌ Error disconnecting: {e}", exc_info=True)


# Пример использования:
async def example_usage():
    """Пример использования recorder"""
    recorder = SimpleWebRTCRecorder(
        room_name="testroom",
        conference_id="test123",
        jvb_host="jvb",
        jvb_port="8080",
        output_dir="/tmp/recordings"
    )

    try:
        await recorder.connect()

        # Записываем 30 секунд
        await asyncio.sleep(30)

    finally:
        await recorder.disconnect()


if __name__ == "__main__":
    logging.basicConfig(level=logging.DEBUG)
    asyncio.run(example_usage())

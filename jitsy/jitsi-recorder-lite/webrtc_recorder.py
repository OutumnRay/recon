"""
WebRTC Recorder для Jitsi - подключается напрямую к JVB через WebSocket
и записывает индивидуальные audio tracks участников
"""
import asyncio
import logging
import json
import os
from datetime import datetime
from pathlib import Path

import websockets
from aiortc import RTCPeerConnection, RTCSessionDescription
from av import open as av_open
import av

logger = logging.getLogger(__name__)


class WebRTCJitsiRecorder:
    """
    Подключается к Jitsi Videobridge через WebSocket Colibri API
    и записывает audio tracks участников
    """

    def __init__(self, room_name: str, conference_id: str, jvb_host: str, jvb_port: str, output_dir: str):
        self.room_name = room_name
        self.conference_id = conference_id
        self.jvb_host = jvb_host
        self.jvb_port = jvb_port
        self.output_dir = output_dir

        self.pc = None
        self.ws = None
        self.is_connected = False
        self.active_tracks = {}  # endpoint_id -> track recorder

    async def connect(self):
        """Подключается к JVB"""
        try:
            logger.info(f"🔌 Connecting to JVB WebSocket: {self.jvb_host}:{self.jvb_port}")

            # JVB Colibri WebSocket endpoint
            ws_url = f"ws://{self.jvb_host}:{self.jvb_port}/colibri-ws/default-id/{self.conference_id}"

            self.ws = await websockets.connect(ws_url)
            logger.info(f"✅ Connected to JVB WebSocket")

            # Создаем WebRTC peer connection
            self.pc = RTCPeerConnection()

            # Обработчик для входящих audio tracks
            @self.pc.on("track")
            async def on_track(track):
                if track.kind == "audio":
                    logger.info(f"📥 Received audio track: {track.id}")
                    await self._start_recording_track(track)

            # Создаем SDP offer
            await self._create_offer()

            self.is_connected = True

        except Exception as e:
            logger.error(f"❌ Failed to connect to JVB: {e}", exc_info=True)
            raise

    async def _create_offer(self):
        """Создает WebRTC offer и отправляет в JVB"""
        try:
            # Создаем offer
            offer = await self.pc.createOffer()
            await self.pc.setLocalDescription(offer)

            # Отправляем offer в JVB через WebSocket
            message = {
                "colibriClass": "SenderSourceConstraints",
                "sourceName": "recorder-bot",
                "videoConstraints": {},
                "sdp": offer.sdp
            }

            await self.ws.send(json.dumps(message))
            logger.debug(f"📤 Sent SDP offer to JVB")

            # Ждем answer от JVB
            response = await self.ws.recv()
            data = json.loads(response)

            if 'sdp' in data:
                answer = RTCSessionDescription(sdp=data['sdp'], type="answer")
                await self.pc.setRemoteDescription(answer)
                logger.info(f"✅ Received SDP answer from JVB")
            else:
                logger.error(f"❌ No SDP in JVB response: {data}")

        except Exception as e:
            logger.error(f"❌ Error creating offer: {e}", exc_info=True)
            raise

    async def _start_recording_track(self, track):
        """Начинает запись audio track"""
        try:
            endpoint_id = track.id
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S_%f')[:-3]
            filename = f"{self.room_name}_{endpoint_id}_{timestamp}.opus"
            filepath = os.path.join(self.output_dir, filename)

            logger.info(f"🎙️  Starting recording: {filename}")

            # Открываем файл для записи в opus формате
            container = av_open(filepath, 'w')
            stream = container.add_stream('opus', rate=48000)

            # Записываем frames
            async def record_frames():
                try:
                    while True:
                        frame = await track.recv()

                        # Конвертируем в opus и пишем
                        for packet in stream.encode(frame):
                            container.mux(packet)

                except Exception as e:
                    logger.info(f"⏹️  Track ended: {endpoint_id} - {e}")
                finally:
                    # Закрываем файл
                    for packet in stream.encode():
                        container.mux(packet)
                    container.close()

                    file_size = os.path.getsize(filepath) if os.path.exists(filepath) else 0
                    logger.info(f"✅ Recording saved: {filename} ({file_size} bytes)")

            # Запускаем запись в фоне
            self.active_tracks[endpoint_id] = asyncio.create_task(record_frames())

        except Exception as e:
            logger.error(f"❌ Error starting track recording: {e}", exc_info=True)

    async def disconnect(self):
        """Отключается от JVB"""
        try:
            logger.info(f"🔌 Disconnecting from JVB")

            # Закрываем все активные записи
            for task in self.active_tracks.values():
                task.cancel()

            self.active_tracks.clear()

            # Закрываем WebRTC
            if self.pc:
                await self.pc.close()

            # Закрываем WebSocket
            if self.ws:
                await self.ws.close()

            self.is_connected = False
            logger.info(f"✅ Disconnected from JVB")

        except Exception as e:
            logger.error(f"❌ Error disconnecting: {e}", exc_info=True)


# ПРИМЕЧАНИЕ:
# Этот код - упрощенная версия. В реальности JVB Colibri WebSocket API
# имеет более сложный протокол обмена сообщениями.
#
# Для полноценной работы нужно:
# 1. Реализовать полный Colibri protocol
# 2. Обработать ICE candidates
# 3. Настроить DTLS/SRTP
# 4. Обработать различные типы сообщений от JVB
#
# Более простой подход - использовать lib-jitsi-meet через Node.js
# или использовать Jibri (стандартное решение Jitsi для записи)

"""
Jitsi WebRTC Client - подключается к Jitsi Meet конференции через WebRTC
и записывает индивидуальные audio tracks участников
"""
import asyncio
import logging
import os
from datetime import datetime
from pathlib import Path
import uuid
import json

import aioxmpp
import aiortc
from aiortc import RTCPeerConnection, RTCSessionDescription, MediaStreamTrack
from aiortc.contrib.media import MediaRecorder
from av import AudioFrame
import av

logger = logging.getLogger(__name__)


class AudioTrackRecorder:
    """Записывает один audio track в файл"""

    def __init__(self, track: MediaStreamTrack, participant_id: str, room_name: str, output_dir: str):
        self.track = track
        self.participant_id = participant_id
        self.room_name = room_name
        self.output_dir = output_dir
        self.filename = None
        self.filepath = None
        self.recorder = None
        self.is_recording = False
        self.start_time = None
        self.duration = 0

    async def start(self):
        """Начинает запись audio track"""
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S_%f')[:-3]
        safe_room = self._sanitize_filename(self.room_name)
        safe_participant = self._sanitize_filename(self.participant_id)

        self.filename = f"{safe_room}_{safe_participant}_{timestamp}.opus"
        self.filepath = os.path.join(self.output_dir, self.filename)

        # Создаем recorder для opus формата
        self.recorder = MediaRecorder(self.filepath, format='opus')
        self.recorder.addTrack(self.track)

        self.start_time = datetime.now()
        self.is_recording = True

        await self.recorder.start()
        logger.info(f"🎙️  Started recording track: {self.filename}")

    async def stop(self):
        """Останавливает запись"""
        if not self.is_recording:
            return

        self.is_recording = False

        if self.recorder:
            await self.recorder.stop()

        if self.start_time:
            self.duration = (datetime.now() - self.start_time).total_seconds()

        logger.info(f"⏹️  Stopped recording: {self.filename} (duration: {self.duration:.1f}s)")

        return {
            'filename': self.filename,
            'filepath': self.filepath,
            'duration': self.duration,
            'participant_id': self.participant_id
        }

    @staticmethod
    def _sanitize_filename(name):
        """Очистка имени файла"""
        import re
        return re.sub(r'[^\w\-.]', '_', name)


class JitsiWebRTCClient:
    """WebRTC клиент для подключения к Jitsi Meet"""

    def __init__(self, room_name: str, prosody_host: str, jvb_host: str,
                 output_dir: str, on_track_callback=None, on_track_ended_callback=None):
        self.room_name = room_name
        self.prosody_host = prosody_host
        self.jvb_host = jvb_host
        self.output_dir = output_dir
        self.on_track_callback = on_track_callback
        self.on_track_ended_callback = on_track_ended_callback

        self.pc = None  # RTCPeerConnection
        self.xmpp_client = None
        self.muc = None
        self.is_connected = False
        self.active_recorders = {}  # endpoint_id -> AudioTrackRecorder

        # Bot identity
        self.bot_id = f"recorder-{uuid.uuid4().hex[:8]}"
        self.bot_jid = f"{self.bot_id}@meet.jitsi"

    async def connect(self):
        """Подключается к Jitsi конференции"""
        try:
            logger.info(f"🔌 Connecting to Jitsi room: {self.room_name}")

            # 1. Создаем WebRTC peer connection
            self.pc = RTCPeerConnection()

            # 2. Обработчик новых tracks (audio от участников)
            @self.pc.on("track")
            async def on_track(track):
                logger.info(f"📥 Received track: {track.kind} from {track.id}")

                if track.kind == "audio":
                    # Создаем recorder для этого track
                    endpoint_id = track.id  # Используем track.id как endpoint_id
                    await self._handle_new_audio_track(track, endpoint_id)

            # 3. Подключаемся к XMPP (Prosody)
            await self._connect_xmpp()

            # 4. Присоединяемся к MUC комнате
            await self._join_muc()

            # 5. Создаем WebRTC offer и отправляем в JVB через XMPP
            await self._setup_webrtc()

            self.is_connected = True
            logger.info(f"✅ Successfully connected to room: {self.room_name}")

        except Exception as e:
            logger.error(f"❌ Failed to connect to Jitsi: {e}", exc_info=True)
            raise

    async def _connect_xmpp(self):
        """Подключается к XMPP серверу (Prosody)"""
        try:
            # TODO: Нужны credentials для XMPP
            # Jitsi использует anonymous auth или токены
            jid = aioxmpp.JID.fromstr(self.bot_jid)

            # Создаем XMPP клиент
            # Это упрощенная версия - нужно настроить auth
            logger.info(f"🔌 Connecting to XMPP: {self.prosody_host}")

            # В реальности нужно:
            # 1. Подключиться к prosody с anonymous auth
            # 2. Получить session
            # self.xmpp_client = await aioxmpp.Client(...)

            logger.warning("⚠️  XMPP connection not fully implemented yet")

        except Exception as e:
            logger.error(f"❌ XMPP connection failed: {e}", exc_info=True)
            raise

    async def _join_muc(self):
        """Присоединяется к Multi-User Chat комнате"""
        try:
            logger.info(f"🚪 Joining MUC room: {self.room_name}")

            # TODO: Присоединиться к MUC
            # room_jid = f"{self.room_name}@conference.{self.prosody_host}"
            # self.muc = self.xmpp_client.summon(aioxmpp.MUCClient)
            # await self.muc.join(room_jid, self.bot_id)

            logger.warning("⚠️  MUC join not fully implemented yet")

        except Exception as e:
            logger.error(f"❌ MUC join failed: {e}", exc_info=True)
            raise

    async def _setup_webrtc(self):
        """Настраивает WebRTC соединение с JVB"""
        try:
            logger.info(f"🌐 Setting up WebRTC with JVB: {self.jvb_host}")

            # 1. Создаем offer
            offer = await self.pc.createOffer()
            await self.pc.setLocalDescription(offer)

            logger.debug(f"Created SDP offer: {offer.sdp[:200]}...")

            # 2. Отправляем offer в JVB через Colibri/XMPP
            # TODO: Нужно отправить SDP через XMPP signaling
            # В Jitsi это делается через:
            # - Jingle IQ для legacy
            # - или Colibri WebSocket для современных версий

            # 3. Получаем answer от JVB
            # answer_sdp = await self._get_jvb_answer(offer.sdp)
            # await self.pc.setRemoteDescription(RTCSessionDescription(sdp=answer_sdp, type="answer"))

            logger.warning("⚠️  WebRTC signaling not fully implemented yet")

        except Exception as e:
            logger.error(f"❌ WebRTC setup failed: {e}", exc_info=True)
            raise

    async def _handle_new_audio_track(self, track: MediaStreamTrack, endpoint_id: str):
        """Обрабатывает новый audio track от участника"""
        try:
            logger.info(f"🎵 New audio track from endpoint: {endpoint_id}")

            # Создаем recorder для этого track
            recorder = AudioTrackRecorder(
                track=track,
                participant_id=endpoint_id,
                room_name=self.room_name,
                output_dir=self.output_dir
            )

            await recorder.start()
            self.active_recorders[endpoint_id] = recorder

            # Callback для уведомления основного recorder.py
            if self.on_track_callback:
                await self.on_track_callback(endpoint_id, recorder)

            # Ждем окончания track
            @track.on("ended")
            async def on_ended():
                logger.info(f"🔚 Track ended for endpoint: {endpoint_id}")
                await self._handle_track_ended(endpoint_id)

        except Exception as e:
            logger.error(f"❌ Error handling new track: {e}", exc_info=True)

    async def _handle_track_ended(self, endpoint_id: str):
        """Обрабатывает окончание audio track"""
        if endpoint_id not in self.active_recorders:
            return

        recorder = self.active_recorders[endpoint_id]
        recording_info = await recorder.stop()

        del self.active_recorders[endpoint_id]

        # Callback для уведомления основного recorder.py
        if self.on_track_ended_callback:
            await self.on_track_ended_callback(endpoint_id, recording_info)

    async def disconnect(self):
        """Отключается от конференции"""
        try:
            logger.info(f"🔌 Disconnecting from room: {self.room_name}")

            # Останавливаем все активные записи
            for endpoint_id, recorder in list(self.active_recorders.items()):
                await recorder.stop()

            self.active_recorders.clear()

            # Закрываем WebRTC connection
            if self.pc:
                await self.pc.close()

            # Покидаем MUC
            if self.muc:
                # await self.muc.leave()
                pass

            # Отключаемся от XMPP
            if self.xmpp_client:
                # await self.xmpp_client.disconnect()
                pass

            self.is_connected = False
            logger.info(f"✅ Disconnected from room: {self.room_name}")

        except Exception as e:
            logger.error(f"❌ Error during disconnect: {e}", exc_info=True)

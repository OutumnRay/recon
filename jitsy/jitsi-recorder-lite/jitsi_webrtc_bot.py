"""
Jitsi WebRTC Bot - полноценный recorder
Подключается через XMPP WebSocket, обрабатывает Jingle signaling,
устанавливает WebRTC соединение с JVB и записывает audio tracks
"""
import asyncio
import logging
import os
import json
from datetime import datetime
from pathlib import Path
from xml.etree import ElementTree as ET

from aiortc import RTCPeerConnection, RTCSessionDescription, RTCConfiguration, RTCIceServer, RTCIceCandidate
from av import open as av_open
from aiohttp import web

from jitsi_xmpp_client import JitsiXmppClient, NAMESPACES, strip_ns
from jingle_sdp import JingleToSDP, SDPToJingle

logger = logging.getLogger(__name__)


class AudioTrackRecorder:
    """Записывает один audio track в opus файл"""

    def __init__(self, track, track_id, room_name, output_dir):
        self.track = track
        self.track_id = track_id
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
        safe_track_id = self._sanitize_filename(self.track_id)

        self.filename = f"{safe_room}_{safe_track_id}_{timestamp}.opus"
        self.filepath = os.path.join(self.output_dir, self.filename)

        Path(self.output_dir).mkdir(parents=True, exist_ok=True)

        # Открываем файл для записи в opus
        self.container = av_open(self.filepath, 'w')
        self.stream = self.container.add_stream('opus', rate=48000, layout='stereo')

        self.start_time = datetime.now()
        self.is_recording = True

        logger.info(f"🎙️  Started recording track {self.track_id}: {self.filename}")

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
            logger.debug(f"Track {self.track_id} ended or error: {e}")
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

        logger.info(f"⏹️  Stopped recording {self.track_id}: {self.filename} ({duration:.1f}s, {file_size} bytes)")

        return {
            'filename': self.filename,
            'filepath': self.filepath,
            'duration': duration,
            'track_id': self.track_id
        }

    @staticmethod
    def _sanitize_filename(name):
        import re
        return re.sub(r'[^\w\-.]', '_', name)


class JitsiWebRTCBot:
    """
    Полноценный WebRTC bot для Jitsi
    Использует XMPP для signaling и aiortc для WebRTC
    """

    def __init__(self, room_name, output_dir, jitsi_url="https://meet.recontext.online"):
        self.room_name = room_name
        self.output_dir = output_dir
        self.jitsi_url = jitsi_url

        # XMPP клиент
        self.xmpp = JitsiXmppClient(
            room=room_name,
            jitsi_url=jitsi_url,
            nickname="recorder-bot"
        )

        # WebRTC
        self.pc = None
        self.session_id = None
        self.focus_jid = None
        self.initiator = None  # JID инициатора (focus)
        self.active_recorders = {}

        Path(output_dir).mkdir(parents=True, exist_ok=True)

    async def connect(self):
        """Подключается к Jitsi и запускает запись"""
        try:
            # Подключаемся к XMPP
            await self.xmpp.connect()

            # Устанавливаем callback для Jingle
            self.xmpp.on_jingle = self._handle_jingle

            # Запрашиваем конференцию у focus
            await self.xmpp.request_conference()

            logger.info("✅ Bot connected and waiting for Jingle session-initiate...")

        except Exception as e:
            logger.error(f"❌ Failed to connect: {e}", exc_info=True)
            raise

    async def _handle_jingle(self, iq_elem, jingle_elem):
        """Обрабатывает Jingle сообщения"""
        action = jingle_elem.get('action')
        sid = jingle_elem.get('sid')
        initiator = jingle_elem.get('initiator')

        logger.info(f"🎬 Jingle {action} from {initiator} (sid={sid})")

        if action == 'session-initiate':
            # Получили session-initiate от focus - начинаем WebRTC
            self.session_id = sid
            self.focus_jid = iq_elem.get('from')
            self.initiator = initiator

            # Отправляем result на IQ
            await self._send_iq_result(iq_elem)

            # Создаем WebRTC connection
            await self._setup_webrtc(jingle_elem)

            # Отправляем session-accept
            await self._send_session_accept(jingle_elem)

        elif action == 'transport-info':
            # ICE candidate от JVB
            await self._send_iq_result(iq_elem)
            await self._handle_transport_info(jingle_elem)

        elif action == 'content-modify':
            # Изменение медиа контента
            await self._send_iq_result(iq_elem)

        else:
            logger.warning(f"⚠️  Unhandled Jingle action: {action}")
            await self._send_iq_result(iq_elem)

    async def _setup_webrtc(self, jingle_elem):
        """Создает WebRTC PeerConnection и обрабатывает SDP из Jingle"""
        try:
            logger.info("🌐 Setting up WebRTC connection...")

            # Создаем RTCPeerConnection
            configuration = RTCConfiguration(
                iceServers=[
                    RTCIceServer(urls=["stun:stun.l.google.com:19302"]),
                    RTCIceServer(urls=["stun:stun1.l.google.com:19302"])
                ]
            )
            self.pc = RTCPeerConnection(configuration)

            # Обработчик для входящих tracks
            @self.pc.on("track")
            async def on_track(track):
                logger.info(f"📥 Received track: kind={track.kind}, id={track.id}")

                if track.kind == "audio":
                    # Создаем recorder для audio track
                    recorder = AudioTrackRecorder(
                        track=track,
                        track_id=track.id,
                        room_name=self.room_name,
                        output_dir=self.output_dir
                    )
                    await recorder.start()
                    self.active_recorders[track.id] = recorder

            # Обработчик ICE connection state
            @self.pc.on("connectionstatechange")
            async def on_connection_state_change():
                logger.info(f"🔗 ICE connection state: {self.pc.connectionState}")

            # Обработчик для local ICE candidates
            @self.pc.on("icecandidate")
            async def on_ice_candidate(candidate):
                if candidate:
                    # Отправляем transport-info с ICE candidate
                    await self._send_transport_info(candidate)

            # Конвертируем Jingle в SDP используя helper
            sdp = JingleToSDP.convert(jingle_elem)
            logger.debug(f"📄 Converted SDP:\n{sdp}")

            # Устанавливаем remote description
            offer = RTCSessionDescription(sdp=sdp, type="offer")
            await self.pc.setRemoteDescription(offer)

            # Создаем answer
            answer = await self.pc.createAnswer()
            await self.pc.setLocalDescription(answer)

            logger.info("✅ WebRTC setup complete")

        except Exception as e:
            logger.error(f"❌ Failed to setup WebRTC: {e}", exc_info=True)
            raise

    async def _send_session_accept(self, jingle_offer):
        """Отправляет session-accept обратно в focus"""
        try:
            logger.info("📤 Sending session-accept...")

            # Конвертируем local SDP answer в Jingle XML
            local_sdp = self.pc.localDescription.sdp
            initiator = jingle_offer.get('initiator')

            # Используем SDPToJingle конвертер
            jingle_elem = SDPToJingle.convert(local_sdp, self.session_id, initiator)

            # Добавляем responder attribute
            jingle_elem.set('responder', self.xmpp.jid)

            # Конвертируем в XML string
            jingle_xml = ET.tostring(jingle_elem, encoding='unicode')

            await self.xmpp._send_iq(jingle_xml, iq_type='set', to=self.focus_jid)
            logger.info("✅ Session-accept sent")

        except Exception as e:
            logger.error(f"❌ Failed to send session-accept: {e}", exc_info=True)

    async def _send_transport_info(self, candidate):
        """Отправляет transport-info с ICE candidate в JVB"""
        try:
            if not self.session_id or not self.focus_jid:
                return

            # Парсим candidate string из aiortc
            # Формат: candidate:foundation component protocol priority ip port typ type [raddr X rport Y]
            parts = candidate.candidate.split()

            if len(parts) < 8:
                logger.warning(f"Invalid candidate format: {candidate.candidate}")
                return

            cand_attrs = {
                'foundation': parts[0].split(':')[1],
                'component': parts[1],
                'protocol': parts[2],
                'priority': parts[3],
                'ip': parts[4],
                'port': parts[5],
                'type': parts[7]
            }

            # raddr/rport если есть
            if 'raddr' in parts:
                idx = parts.index('raddr')
                cand_attrs['rel-addr'] = parts[idx + 1]
            if 'rport' in parts:
                idx = parts.index('rport')
                cand_attrs['rel-port'] = parts[idx + 1]

            # Создаем Jingle transport-info
            attr_str = ' '.join([f"{k}='{v}'" for k, v in cand_attrs.items()])

            jingle_xml = (
                f"<jingle action='transport-info' "
                f"initiator='{self.initiator}' "
                f"sid='{self.session_id}' "
                f"xmlns='{NAMESPACES['jingle']}'>"
                f"<content name='audio' creator='initiator'>"  # Упрощенно: только audio
                f"<transport xmlns='urn:xmpp:jingle:transports:ice-udp:1'>"
                f"<candidate {attr_str}/>"
                f"</transport>"
                f"</content>"
                f"</jingle>"
            )

            await self.xmpp._send_iq(jingle_xml, iq_type='set', to=self.focus_jid)
            logger.debug(f"📤 Sent ICE candidate: {cand_attrs['ip']}:{cand_attrs['port']}")

        except Exception as e:
            logger.error(f"❌ Error sending transport-info: {e}", exc_info=True)

    async def _handle_transport_info(self, jingle_elem):
        """Обрабатывает transport-info (ICE candidates от JVB)"""
        try:
            contents = jingle_elem.findall(f".//{{{NAMESPACES['jingle']}}}content")

            for content in contents:
                transport = content.find(f".//{{{NAMESPACES['jingle']}}}transport")
                if transport is not None:
                    candidates = transport.findall(f".//{{{NAMESPACES['jingle']}}}candidate")

                    for cand_elem in candidates:
                        # Добавляем ICE candidate в aiortc
                        foundation = cand_elem.get('foundation', '0')
                        component = cand_elem.get('component', '1')
                        protocol = cand_elem.get('protocol', 'udp').upper()
                        priority = cand_elem.get('priority', '0')
                        ip = cand_elem.get('ip')
                        port = cand_elem.get('port')
                        typ = cand_elem.get('type', 'host')

                        if not ip or not port:
                            continue

                        # Формируем candidate string для aiortc
                        cand_str = f"candidate:{foundation} {component} {protocol} {priority} {ip} {port} typ {typ}"

                        # raddr/rport если есть
                        if typ in ['relay', 'srflx']:
                            rel_addr = cand_elem.get('rel-addr')
                            rel_port = cand_elem.get('rel-port')
                            if rel_addr and rel_port:
                                cand_str += f" raddr {rel_addr} rport {rel_port}"

                        # Создаем RTCIceCandidate
                        ice_candidate = RTCIceCandidate(
                            component=int(component),
                            foundation=foundation,
                            ip=ip,
                            port=int(port),
                            priority=int(priority),
                            protocol=protocol.lower(),
                            type=typ,
                            sdpMid=content.get('name'),  # mid из content name
                            sdpMLineIndex=0  # упрощенно
                        )

                        await self.pc.addIceCandidate(ice_candidate)
                        logger.debug(f"📥 Added ICE candidate: {ip}:{port}")

        except Exception as e:
            logger.error(f"❌ Error handling transport-info: {e}", exc_info=True)

    async def _send_iq_result(self, iq_elem):
        """Отправляет IQ result"""
        iq_id = iq_elem.get('id')
        from_jid = iq_elem.get('from')

        result = (
            f"<iq id='{iq_id}' to='{from_jid}' type='result' "
            f"xmlns='{NAMESPACES['client']}'/>"
        )
        await self.xmpp._send_raw(result)

    async def disconnect(self):
        """Отключается от конференции"""
        try:
            logger.info("🔌 Disconnecting...")

            # Останавливаем все записи
            for track_id, recorder in list(self.active_recorders.items()):
                await recorder.stop()

            self.active_recorders.clear()

            # Закрываем WebRTC
            if self.pc:
                await self.pc.close()

            # Закрываем XMPP
            await self.xmpp.disconnect()

            logger.info("✅ Disconnected")

        except Exception as e:
            logger.error(f"❌ Error disconnecting: {e}", exc_info=True)


async def start_health_server():
    """HTTP health check сервер"""
    async def health(request):
        return web.json_response({
            'status': 'healthy',
            'service': 'jitsi-webrtc-bot'
        })

    app = web.Application()
    app.router.add_get('/health', health)

    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, '0.0.0.0', 8080)
    await site.start()
    logger.info("🌐 Health check server started on :8080")


# Пример использования
async def main():
    logging.basicConfig(
        level=logging.DEBUG,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )

    room_name = os.getenv("JITSI_ROOM", "testmeet")
    output_dir = os.getenv("RECORD_DIR", "./recordings")
    jitsi_url = os.getenv("JITSI_URL", "https://meet.recontext.online")

    # Start health check server
    asyncio.create_task(start_health_server())

    bot = JitsiWebRTCBot(
        room_name=room_name,
        output_dir=output_dir,
        jitsi_url=jitsi_url
    )

    try:
        await bot.connect()

        # Записываем пока не прервут
        logger.info("📹 Recording... (Press Ctrl+C to stop)")
        while True:
            await asyncio.sleep(1)

    except KeyboardInterrupt:
        logger.info("⚠️  Interrupted by user")
    finally:
        await bot.disconnect()


if __name__ == "__main__":
    asyncio.run(main())

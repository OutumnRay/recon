"""
GStreamer WebRTC Recorder для Jitsi
Использует webrtcbin для подключения к JVB и записи индивидуальных audio tracks
"""
import asyncio
import logging
import json
import os
from datetime import datetime

import gi
gi.require_version('Gst', '1.0')
gi.require_version('GstWebRTC', '1.0')
gi.require_version('GstSdp', '1.0')
from gi.repository import Gst, GstWebRTC, GstSdp
import websockets

logger = logging.getLogger(__name__)

# Инициализация GStreamer
Gst.init(None)


class GStreamerWebRTCRecorder:
    """
    WebRTC recorder используя GStreamer webrtcbin
    Подключается к Jitsi JVB и записывает audio streams
    """

    def __init__(self, room_name: str, conference_id: str, jvb_ws_url: str, output_dir: str):
        self.room_name = room_name
        self.conference_id = conference_id
        self.jvb_ws_url = jvb_ws_url
        self.output_dir = output_dir

        self.pipeline = None
        self.webrtc = None
        self.ws = None
        self.is_recording = False
        self.active_pads = {}  # pad_id -> recording info

    async def start(self):
        """Запускает WebRTC соединение и начинает запись"""
        try:
            logger.info(f"🎙️  Starting GStreamer WebRTC recorder for room: {self.room_name}")

            # Создаем GStreamer pipeline
            self.pipeline = Gst.Pipeline.new("jitsi-recorder")

            # Создаем webrtcbin element
            self.webrtc = Gst.ElementFactory.make("webrtcbin", "webrtc")
            self.webrtc.set_property("bundle-policy", "max-bundle")
            self.pipeline.add(self.webrtc)

            # Подключаем обработчик для новых audio pads
            self.webrtc.connect("pad-added", self._on_pad_added)

            # Подключаемся к JVB через WebSocket
            await self._connect_jvb()

            # Запускаем pipeline
            self.pipeline.set_state(Gst.State.PLAYING)
            self.is_recording = True

            logger.info(f"✅ GStreamer WebRTC recorder started")

        except Exception as e:
            logger.error(f"❌ Failed to start GStreamer recorder: {e}", exc_info=True)
            raise

    async def _connect_jvb(self):
        """Подключается к JVB через WebSocket Colibri"""
        try:
            logger.info(f"🔌 Connecting to JVB WebSocket: {self.jvb_ws_url}")

            self.ws = await websockets.connect(self.jvb_ws_url)

            # Создаем SDP offer
            promise = Gst.Promise.new_with_change_func(self._on_offer_created)
            self.webrtc.emit("create-offer", None, promise)

            # Обработчик ICE candidates
            self.webrtc.connect("on-ice-candidate", self._on_ice_candidate)

            # Обработчик для incoming SDP
            asyncio.create_task(self._handle_jvb_messages())

        except Exception as e:
            logger.error(f"❌ Failed to connect to JVB: {e}", exc_info=True)
            raise

    def _on_offer_created(self, promise):
        """Callback когда SDP offer создан"""
        try:
            promise.wait()
            reply = promise.get_reply()
            offer = reply['offer']

            # Устанавливаем local description
            self.webrtc.emit("set-local-description", offer, None)

            # Отправляем offer в JVB
            sdp_text = offer.sdp.as_text()
            message = {
                "colibriClass": "EndpointMessage",
                "type": "offer",
                "sdp": sdp_text
            }

            asyncio.create_task(self._send_to_jvb(message))

            logger.debug(f"📤 Sent SDP offer to JVB")

        except Exception as e:
            logger.error(f"❌ Error creating offer: {e}", exc_info=True)

    async def _send_to_jvb(self, message):
        """Отправляет сообщение в JVB через WebSocket"""
        try:
            await self.ws.send(json.dumps(message))
        except Exception as e:
            logger.error(f"❌ Error sending to JVB: {e}")

    async def _handle_jvb_messages(self):
        """Обрабатывает входящие сообщения от JVB"""
        try:
            async for message in self.ws:
                data = json.loads(message)

                if data.get('type') == 'answer':
                    # Получили SDP answer от JVB
                    sdp_text = data['sdp']

                    ret, sdp = GstSdp.SDPMessage.new_from_text(sdp_text)
                    answer = GstWebRTC.WebRTCSessionDescription.new(GstWebRTC.WebRTCSDPType.ANSWER, sdp)

                    self.webrtc.emit("set-remote-description", answer, None)
                    logger.info(f"✅ Received SDP answer from JVB")

                elif data.get('type') == 'candidate':
                    # Получили ICE candidate от JVB
                    candidate = data['candidate']
                    sdp_m_line_index = data['sdpMLineIndex']

                    self.webrtc.emit("add-ice-candidate", sdp_m_line_index, candidate)
                    logger.debug(f"📥 Added ICE candidate from JVB")

        except Exception as e:
            logger.error(f"❌ Error handling JVB messages: {e}", exc_info=True)

    def _on_ice_candidate(self, webrtc, sdp_m_line_index, candidate):
        """Callback для новых ICE candidates"""
        try:
            message = {
                "colibriClass": "EndpointMessage",
                "type": "candidate",
                "candidate": candidate,
                "sdpMLineIndex": sdp_m_line_index
            }

            asyncio.create_task(self._send_to_jvb(message))
            logger.debug(f"📤 Sent ICE candidate to JVB")

        except Exception as e:
            logger.error(f"❌ Error sending ICE candidate: {e}")

    def _on_pad_added(self, webrtc, pad):
        """Callback когда появляется новый audio pad (от участника)"""
        try:
            if pad.direction != Gst.PadDirection.SRC:
                return

            # Получаем caps чтобы определить тип media
            caps = pad.get_current_caps()
            if not caps:
                return

            struct = caps.get_structure(0)
            media_type = struct.get_name()

            logger.info(f"📥 New pad added: {pad.get_name()}, type: {media_type}")

            # Обрабатываем только audio pads
            if media_type.startswith("audio"):
                participant_id = pad.get_name()  # Используем pad name как participant ID
                asyncio.create_task(self._record_audio_pad(pad, participant_id))

        except Exception as e:
            logger.error(f"❌ Error handling new pad: {e}", exc_info=True)

    async def _record_audio_pad(self, pad, participant_id):
        """Записывает audio pad в файл"""
        try:
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S_%f')[:-3]
            filename = f"{self.room_name}_{participant_id}_{timestamp}.opus"
            filepath = os.path.join(self.output_dir, filename)

            logger.info(f"🎙️  Recording audio pad: {filename}")

            # Создаем pipeline элементы для записи
            queue = Gst.ElementFactory.make("queue", f"queue_{participant_id}")
            audioconvert = Gst.ElementFactory.make("audioconvert", f"audioconvert_{participant_id}")
            audioresample = Gst.ElementFactory.make("audioresample", f"audioresample_{participant_id}")
            opusenc = Gst.ElementFactory.make("opusenc", f"opusenc_{participant_id}")
            filesink = Gst.ElementFactory.make("filesink", f"filesink_{participant_id}")

            filesink.set_property("location", filepath)
            opusenc.set_property("bitrate", 48000)

            # Добавляем элементы в pipeline
            self.pipeline.add(queue)
            self.pipeline.add(audioconvert)
            self.pipeline.add(audioresample)
            self.pipeline.add(opusenc)
            self.pipeline.add(filesink)

            # Связываем элементы
            queue.link(audioconvert)
            audioconvert.link(audioresample)
            audioresample.link(opusenc)
            opusenc.link(filesink)

            # Связываем pad с queue
            queue_sink_pad = queue.get_static_pad("sink")
            pad.link(queue_sink_pad)

            # Синхронизируем состояния
            queue.sync_state_with_parent()
            audioconvert.sync_state_with_parent()
            audioresample.sync_state_with_parent()
            opusenc.sync_state_with_parent()
            filesink.sync_state_with_parent()

            # Сохраняем информацию о записи
            self.active_pads[participant_id] = {
                'filename': filename,
                'filepath': filepath,
                'pad': pad,
                'elements': [queue, audioconvert, audioresample, opusenc, filesink],
                'start_time': datetime.now()
            }

            logger.info(f"✅ Started recording: {filename}")

        except Exception as e:
            logger.error(f"❌ Error recording audio pad: {e}", exc_info=True)

    async def stop(self):
        """Останавливает запись"""
        try:
            logger.info(f"🛑 Stopping GStreamer WebRTC recorder")

            self.is_recording = False

            # Останавливаем pipeline
            if self.pipeline:
                self.pipeline.set_state(Gst.State.NULL)

            # Закрываем WebSocket
            if self.ws:
                await self.ws.close()

            logger.info(f"✅ Recorder stopped")

            # Возвращаем информацию о записанных файлах
            recordings = []
            for participant_id, info in self.active_pads.items():
                duration = (datetime.now() - info['start_time']).total_seconds()
                recordings.append({
                    'participant_id': participant_id,
                    'filename': info['filename'],
                    'filepath': info['filepath'],
                    'duration': duration
                })

            return recordings

        except Exception as e:
            logger.error(f"❌ Error stopping recorder: {e}", exc_info=True)
            return []

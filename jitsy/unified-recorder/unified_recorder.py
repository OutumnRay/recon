#!/usr/bin/env python3
"""
Jitsi Unified Recorder - Прямая запись через aiortc (WebRTC)
"""

import os
import asyncio
import json
import logging
import hashlib
import uuid
import xml.etree.ElementTree as ET
from datetime import datetime, timezone
from pathlib import Path
from typing import Dict, Set

# ИСПРАВЛЕНИЕ: Импортируем 'aiohttp' целиком, чтобы получить доступ к ClientSession
import aiohttp
from aiohttp import web, WSMsgType
import boto3
import redis.asyncio as redis

from aiortc import RTCPeerConnection, RTCSessionDescription
from aiortc.contrib.media import MediaRecorder
import sdp_transform

# Настройка логирования
LOG_LEVEL = os.getenv('LOG_LEVEL', 'DEBUG').upper()
logging.basicConfig(level=LOG_LEVEL, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logging.getLogger("aiortc").setLevel(logging.WARNING)
logging.getLogger("aioice").setLevel(logging.WARNING)
logger = logging.getLogger(__name__)

# ========== Конфигурация ==========
JITSI_DOMAIN = os.getenv('JITSI_DOMAIN', 'meet.recontext.online')
XMPP_MUC_DOMAIN = os.getenv('XMPP_MUC_DOMAIN', 'muc.meet.jitsi')
MINIO_ENDPOINT = os.getenv('MINIO_ENDPOINT')
MINIO_ACCESS_KEY = os.getenv('MINIO_ACCESS_KEY')
MINIO_SECRET_KEY = os.getenv('MINIO_SECRET_KEY')
MINIO_BUCKET = os.getenv('MINIO_BUCKET')
REDIS_HOST = os.getenv('REDIS_HOST', 'redis')
REDIS_PORT = int(os.getenv('REDIS_PORT', '6379'))
RECORD_DIR = Path(os.getenv('RECORD_DIR', '/recordings'))
AUTO_RECORD = os.getenv('AUTO_RECORD', 'true').lower() == 'true'

RECORD_DIR.mkdir(exist_ok=True, parents=True)

logger.info("=" * 60)
logger.info("🎬 Jitsi aiortc Recorder Configuration")
logger.info("=" * 60)
logger.info(f"  JITSI_DOMAIN: {JITSI_DOMAIN}")
logger.info(f"  XMPP_MUC_DOMAIN: {XMPP_MUC_DOMAIN}")
logger.info(f"  MINIO_BUCKET: {MINIO_BUCKET}")
logger.info("=" * 60)

s3_client = boto3.client('s3', endpoint_url=MINIO_ENDPOINT, aws_access_key_id=MINIO_ACCESS_KEY, aws_secret_access_key=MINIO_SECRET_KEY) if MINIO_ENDPOINT else None
active_sessions: Dict[str, 'SessionInfo'] = {}
active_conferences: Dict[str, Set[str]] = {}

class SessionInfo:
    def __init__(self, participant_id: str, room_name: str, display_name: str):
        self.participant_id = participant_id
        self.room_name = room_name
        self.display_name = display_name
        self.started_at = datetime.now(timezone.utc)
        self.pc = RTCPeerConnection()
        self.recorder = None
        self.ws = None
        self.run_task = None
        timestamp = self.started_at.strftime('%Y%m%d_%H%M%S_%f')
        safe_name = "".join(c for c in display_name if c.isalnum())[:50]
        self.filename = f"{self.room_name}_{safe_name}_{timestamp}.opus"
        self.filepath = str(RECORD_DIR / self.filename)

    async def start_recording(self):
        log_prefix = f"[{self.room_name}][{self.display_name}]"
        logger.info(f"{log_prefix} 🎙️ Starting direct WebRTC recording -> {self.filename}")
        try:
            self.recorder = MediaRecorder(self.filepath, format="opus")
            @self.pc.on("track")
            async def on_track(track):
                logger.info(f"{log_prefix} Audio track received from Jitsi: {track.kind}")
                if track.kind == "audio":
                    self.recorder.add_track(track)
            self.run_task = asyncio.create_task(self._run_signaling())
            await self.recorder.start()
            logger.info(f"{log_prefix} ✅ Recorder is running and waiting for audio track.")
        except Exception as e:
            logger.error(f"{log_prefix} ❌ Failed to start recording procedure: {e}", exc_info=True)
            await self.stop_recording()

    async def _run_signaling(self):
        log_prefix = f"[{self.room_name}][{self.display_name}]"
        if XMPP_WEBSOCKET_URL:
            ws_url = f"{XMPP_WEBSOCKET_URL}?room={self.room_name}"
        else:
            ws_url = f"wss://{JITSI_DOMAIN}/xmpp-websocket?room={self.room_name}"

        logger.debug(f"{log_prefix} Connecting to Jitsi WebSocket: {ws_url}")
        try:
            # ИСПРАВЛЕНИЕ: Используем aiohttp.ClientSession()
            async with aiohttp.ClientSession() as session:
                async with session.ws_connect(ws_url, ssl=False) as self.ws:
                    logger.info(f"{log_prefix} WebSocket connected.")
                    await self.ws.send_str(f'<open to="{JITSI_DOMAIN}" version="1.0" xmlns="urn:ietf:params:xml:ns:xmpp-framing"/>')
                    room_jid = f"{self.room_name}@{XMPP_MUC_DOMAIN}/{uuid.uuid4()}"
                    await self.ws.send_str(f'<presence to="{room_jid}" xmlns="jabber:client"><x xmlns="http://jabber.org/protocol/muc"/></presence>')
                    offer = await self.pc.createOffer()
                    await self.pc.setLocalDescription(offer)
                    sdp_text = self.pc.localDescription.sdp
                    sdp_text = sdp_text.replace("a=sendrecv", "a=recvonly")
                    jingle_iq = self._create_jingle_offer(sdp_text, room_jid)
                    logger.debug(f"{log_prefix} Sending Jingle IQ offer...")
                    await self.ws.send_str(ET.tostring(jingle_iq, encoding='unicode'))
                    async for msg in self.ws:
                        if msg.type == WSMsgType.TEXT:
                            await self._handle_ws_message(msg.data)
                        elif msg.type in (WSMsgType.CLOSED, WSMsgType.ERROR):
                            break
        except Exception as e:
            logger.error(f"{log_prefix} ❌ Signaling task failed: {e}", exc_info=True)
        finally:
            logger.warning(f"{log_prefix} Signaling task finished.")
            if self.participant_id in active_sessions:
                 await handle_participant_left(self.participant_id)

    async def _handle_ws_message(self, data):
        log_prefix = f"[{self.room_name}][{self.display_name}]"
        try:
            root = ET.fromstring(data)
            jingle = root.find('{urn:xmpp:jingle:1}jingle')
            if jingle is not None and jingle.attrib.get('action') == 'session-accept':
                logger.debug(f"{log_prefix} Received Jingle session-accept.")
                answer_sdp = self._parse_jingle_answer(root)
                if answer_sdp:
                    logger.info(f"{log_prefix} SDP Answer received, setting remote description.")
                    await self.pc.setRemoteDescription(RTCSessionDescription(sdp=answer_sdp, type="answer"))
        except ET.ParseError:
            pass
        except Exception as e:
            logger.error(f"{log_prefix} Error handling WebSocket message: {e}")

    def _create_jingle_offer(self, sdp_text, to_jid):
        iq = ET.Element("iq", to=to_jid.split('/')[0] + '/focus', type="set", xmlns="jabber:client")
        jingle = ET.SubElement(iq, "jingle", action="session-initiate", sid=str(uuid.uuid4()), xmlns="urn:xmpp:jingle:1")
        parsed_sdp = sdp_transform.parse(sdp_text)
        for media in parsed_sdp['media']:
            content = ET.SubElement(jingle, "content", name=media['type'])
            description = ET.SubElement(content, "description", media="RTP/SAVPF", xmlns="urn:xmpp:jingle:apps:rtp:1")
            transport = ET.SubElement(content, "transport", pwd=parsed_sdp['icePwd'], ufrag=parsed_sdp['iceUfrag'], xmlns="urn:xmpp:jingle:transports:ice-udp:1")
            for pt in media['payloads'].split():
                payload = next((p for p in media['rtp'] if p['payload'] == int(pt)), None)
                if payload:
                    ET.SubElement(description, "payload-type", id=str(payload['payload']), name=payload['codec'], clockrate=str(payload['rate']))
            if media.get('ssrcs'):
                ET.SubElement(description, "ssrc", **{'xmlns': "urn:xmpp:jingle:apps:rtp:ssrc:0"})
            for cand in parsed_sdp.get('candidates', []):
                 ET.SubElement(transport, "candidate", **cand)
            if parsed_sdp.get('fingerprint'):
                ET.SubElement(transport, "fingerprint", hash=parsed_sdp['fingerprint']['type'], xmlns="urn:xmpp:jingle:dtls:0").text = parsed_sdp['fingerprint']['hash']
        return iq

    def _parse_jingle_answer(self, iq_element):
        jingle = iq_element.find('{urn:xmpp:jingle:1}jingle')
        if jingle is None: return None
        sdp_obj = {'version': 0, 'origin': {'username': '-', 'sessionId': '0', 'sessionVersion': 0, 'netType': 'IN', 'ipVer': 4, 'address': '0.0.0.0'}, 'name': '-', 'timing': {'start': 0, 'stop': 0}, 'media': []}
        for content in jingle.findall('{urn:xmpp:jingle:1}content'):
            transport = content.find('{urn:xmpp:jingle:transports:ice-udp:1}transport')
            description = content.find('{urn:xmpp:jingle:apps:rtp:1}description')
            media_type = description.attrib['media']
            media = {'type': media_type, 'port': 9, 'protocol': 'UDP/TLS/RTP/SAVPF', 'payloads': ' '.join([pt.attrib['id'] for pt in description.findall('{urn:xmpp:jingle:apps:rtp:1}payload-type')])}
            sdp_obj['media'].append(media)
            sdp_obj['iceUfrag'] = transport.attrib['ufrag']
            sdp_obj['icePwd'] = transport.attrib['pwd']
            sdp_obj['fingerprint'] = {'type': 'sha-256', 'hash': transport.find('{urn:xmpp:jingle:dtls:0}fingerprint').text}
            sdp_obj['candidates'] = [{'foundation': c.attrib['foundation'], 'component': int(c.attrib['component']), 'transport': 'udp', 'priority': int(c.attrib['priority']), 'ip': c.attrib['ip'], 'port': int(c.attrib['port']), 'type': c.attrib['type']} for c in transport.findall('{urn:xmpp:jingle:transports:ice-udp:1}candidate')]
        return sdp_transform.write(sdp_obj)

    async def stop_recording(self):
        log_prefix = f"[{self.room_name}][{self.display_name}]"
        logger.info(f"{log_prefix} ⏹️ Stopping recording...")
        if self.run_task and not self.run_task.done():
            self.run_task.cancel()
        if self.recorder:
            await self.recorder.stop()
        if self.ws and not self.ws.closed:
            await self.ws.close()
        await self.pc.close()
        filepath = Path(self.filepath)
        if filepath.exists() and filepath.stat().st_size > 1024:
            logger.info(f"{log_prefix} 📁 File saved: {self.filepath} ({filepath.stat().st_size} bytes)")
            await self.upload_to_s3(filepath)
        else:
            logger.warning(f"{log_prefix} ⚠️ Empty or missing file, skipping upload.")

    async def upload_to_s3(self, filepath: Path):
        if not s3_client: return
        try:
            s3_key = f"recordings/{self.room_name}/{filepath.name}"
            logger.info(f"☁️ Uploading to s3://{MINIO_BUCKET}/{s3_key}")
            await asyncio.to_thread(s3_client.upload_file, str(filepath), MINIO_BUCKET, s3_key)
            filepath.unlink()
            logger.info(f"✅ Uploaded and removed local file: {s3_key}")
        except Exception as e:
            logger.error(f"❌ S3 Upload failed: {e}", exc_info=True)

async def handle_participant_joined(room_name: str, participant_id: str, display_name: str):
    logger.info(f"📥 Handling PARTICIPANT JOINED for {display_name} (ID: {participant_id})")
    if 'focus' in participant_id:
        logger.debug("Skipping system component 'focus'")
        return
    if room_name not in active_conferences:
        active_conferences[room_name] = set()
    active_conferences[room_name].add(participant_id)
    if AUTO_RECORD:
        if participant_id in active_sessions:
            logger.warning(f"⚠️ Participant {display_name} already has a session. Stopping old one.")
            await handle_participant_left(participant_id)
        session = SessionInfo(participant_id, room_name, display_name)
        active_sessions[participant_id] = session
        await session.start_recording()

async def handle_participant_left(participant_id: str):
    logger.info(f"📤 Handling PARTICIPANT LEFT for ID: {participant_id}")
    if participant_id in active_sessions:
        session = active_sessions.pop(participant_id)
        await session.stop_recording()
        if session.room_name in active_conferences:
            active_conferences[session.room_name].discard(participant_id)
            if not active_conferences[session.room_name]:
                logger.info(f"Last participant left room {session.room_name}. Cleaning up conference entry.")
                del active_conferences[session.room_name]

async def webhook_handler(request):
    try:
        data = await request.json()
        logger.info(f"🔔 Webhook received:\n{json.dumps(data, indent=2)}")
        event_type = data.get('eventType')
        room_name = data.get('roomName')
        participant_id = data.get('participantId')
        if event_type == 'participantJoined':
            display_name = data.get('displayName', participant_id)
            await handle_participant_joined(room_name, participant_id, display_name)
        elif event_type == 'participantLeft':
            await handle_participant_left(participant_id)
        return web.json_response({'status': 'ok'})
    except Exception as e:
        logger.error(f"❌ Webhook handler error: {e}", exc_info=True)
        return web.json_response({'status': 'error', 'message': str(e)}, status=500)

async def health_handler(request):
    return web.json_response({'status': 'healthy', 'active_sessions': len(active_sessions)})

async def main():
    logger.info("🚀 Starting Unified aiortc Recorder")
    app = web.Application()
    app.router.add_get('/health', health_handler)
    app.router.add_post('/events', webhook_handler)
    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, '0.0.0.0', 8080)
    await site.start()
    logger.info("✅ Recorder ready.")
    await asyncio.Event().wait()

if __name__ == '__main__':
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("👋 Shutting down...")
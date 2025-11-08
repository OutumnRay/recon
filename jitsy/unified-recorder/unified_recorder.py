#!/usr/bin/env python3
"""
Jitsi Unified Recorder - автоматическая запись через Kurento
Объединяет auto-recorder (webhook listener) и kurento-recorder (API)
"""

import os
import asyncio
import json
import logging
import hashlib
from datetime import datetime, timezone
from pathlib import Path
from typing import Optional, Dict, Any
from aiohttp import web
import boto3
import websockets
import redis.asyncio as redis

from colibri_client import ColibriClient

LOG_LEVEL = os.getenv('LOG_LEVEL', 'DEBUG').upper()
logging.basicConfig(
    level=LOG_LEVEL,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

KURENTO_WS_URI = os.getenv('KURENTO_URI', 'ws://kurento:8080/kurento')
MINIO_ENDPOINT = os.getenv('MINIO_ENDPOINT', 'https://api.storage.recontext.online')
MINIO_ACCESS_KEY = os.getenv('MINIO_ACCESS_KEY', 'minioadmin')
MINIO_SECRET_KEY = os.getenv('MINIO_SECRET_KEY', 'minioadmin')
MINIO_BUCKET = os.getenv('MINIO_BUCKET', 'jitsi-recordings')
MINIO_SECURE = os.getenv('MINIO_SECURE', 'true').lower() == 'true'
REDIS_HOST = os.getenv('REDIS_HOST', 'redis')
REDIS_PORT = int(os.getenv('REDIS_PORT', '6379'))
JVB_HOST = os.getenv('JVB_HOST', 'jvb')
JVB_COLIBRI_PORT = int(os.getenv('JVB_COLIBRI_PORT', '8080'))
RECORD_DIR = Path(os.getenv('RECORD_DIR', '/recordings'))
AUTO_RECORD = os.getenv('AUTO_RECORD', 'true').lower() == 'true'
WORKER_ID = os.getenv('HOSTNAME', 'unified-recorder')

RECORD_DIR.mkdir(exist_ok=True, parents=True)

logger.info("=" * 60)
logger.info("🎬 Jitsi Unified Recorder Configuration")
logger.info("=" * 60)
logger.info(f"  LOG_LEVEL: {LOG_LEVEL}")
logger.info(f"  KURENTO_URI: {KURENTO_WS_URI}")
logger.info(f"  MINIO_ENDPOINT: {MINIO_ENDPOINT}")
logger.info(f"  MINIO_BUCKET: {MINIO_BUCKET}")
logger.info(f"  JVB_HOST: {JVB_HOST}:{JVB_COLIBRI_PORT}")
logger.info(f"  AUTO_RECORD: {AUTO_RECORD}")
logger.info(f"  WORKER_ID: {WORKER_ID}")
logger.info("=" * 60)

s3_client = None
if MINIO_ACCESS_KEY and MINIO_SECRET_KEY:
    try:
        s3_client = boto3.client(
            's3', endpoint_url=MINIO_ENDPOINT,
            aws_access_key_id=MINIO_ACCESS_KEY, aws_secret_access_key=MINIO_SECRET_KEY,
            region_name='us-east-1'
        )
        s3_client.head_bucket(Bucket=MINIO_BUCKET)
        logger.info(f"✅ MinIO bucket '{MINIO_BUCKET}' exists and is accessible")
    except Exception:
        logger.warning(f"⚠️ Bucket '{MINIO_BUCKET}' not accessible, trying to create...")
        try:
            s3_client.create_bucket(Bucket=MINIO_BUCKET)
            logger.info(f"✅ Created bucket '{MINIO_BUCKET}'")
        except Exception as e:
            logger.error(f"❌ Failed to create bucket: {e}")
else:
    logger.warning("⚠️ S3 credentials not set, uploads will be skipped")

redis_client = None
active_sessions: Dict[str, 'SessionInfo'] = {}
active_conferences: Dict[str, Dict] = {}


class KurentoClient:
    def __init__(self, ws_uri: str):
        self.ws_uri = ws_uri
        self.request_id = 0
        self.websocket = None

    async def connect(self):
        if self.websocket is None or not self.websocket.open:
            logger.info("Connecting/Reconnecting to Kurento Media Server...")
            self.websocket = await websockets.connect(self.ws_uri, ping_interval=20, ping_timeout=10)
            logger.info("🔗 Connected to Kurento Media Server")

    async def _send_request(self, method: str, params: Dict[str, Any] = None) -> Any:
        await self.connect()
        self.request_id += 1
        request = {"id": self.request_id, "method": method, "params": params or {}, "jsonrpc": "2.0"}
        logger.debug(f"-> KMS: {json.dumps(request)}")
        await self.websocket.send(json.dumps(request))
        response_str = await self.websocket.recv()
        response = json.loads(response_str)
        logger.debug(f"<- KMS: {response}")
        if "error" in response:
            raise Exception(f"Kurento error: {response['error']}")
        return response.get("result", {})

    async def create(self, type: str, constructor_params: Dict) -> str:
        res = await self._send_request("create", {"type": type, "constructorParams": constructor_params})
        return res['value']

    async def invoke(self, object_id: str, operation: str, operation_params: Dict = None) -> Any:
        return await self._send_request("invoke", {"object": object_id, "operation": operation, "operationParams": operation_params or {}})

    async def release(self, object_id: str):
        try:
            await self.invoke(object_id, "release")
            logger.debug(f"Kurento object {object_id} released.")
        except Exception as e:
            logger.warning(f"⚠️ Failed to release Kurento object {object_id}: {e}")

kurento = KurentoClient(KURENTO_WS_URI)
colibri = ColibriClient(JVB_HOST, JVB_COLIBRI_PORT)

class SessionInfo:
    def __init__(self, participant_id: str, room_name: str, display_name: str):
        self.participant_id = participant_id
        self.room_name = room_name
        self.display_name = display_name
        self.started_at = datetime.now(timezone.utc)
        self.pipeline_id = None
        self.rtp_endpoint_id = None
        self.recorder_id = None
        timestamp = self.started_at.strftime('%Y%m%d_%H%M%S_%f')
        safe_participant = "".join(c if c.isalnum() else '_' for c in display_name)[:50]
        safe_room = "".join(c if c.isalnum() else '_' for c in room_name)[:50]
        self.filename = f"{safe_room}_{safe_participant}_{timestamp}.opus"
        self.filepath = RECORD_DIR / self.filename
        self.s3_key = None
        self.duration = 0
        self.uploaded = False

    async def start_recording(self, colibri_conference_id: str):
        log_prefix = f"[Room: {self.room_name}] [User: {self.display_name}]"
        logger.info(f"{log_prefix} 🎙️ Starting recording procedure -> {self.filename}")
        try:
            logger.debug(f"{log_prefix} Step 1: Creating Kurento resources...")
            self.pipeline_id = await kurento.create("MediaPipeline", {})
            self.rtp_endpoint_id = await kurento.create("RtpEndpoint", {"mediaPipeline": self.pipeline_id})
            recorder_params = {"mediaPipeline": self.pipeline_id, "uri": f"file://{self.filepath}", "mediaProfile": "WEBM_AUDIO_ONLY"}
            self.recorder_id = await kurento.create("RecorderEndpoint", recorder_params)
            await kurento.invoke(self.rtp_endpoint_id, "connect", {"sink": self.recorder_id})
            logger.info(f"{log_prefix} Kurento pipeline created: {self.pipeline_id}")

            logger.debug(f"{log_prefix} Step 2: Generating SDP offer from Kurento...")
            local_sdp_offer = await kurento.invoke(self.rtp_endpoint_id, "generateOffer")
            logger.debug(f"{log_prefix} Kurento SDP Offer generated:\n{local_sdp_offer}")

            logger.debug(f"{log_prefix} Step 3: Sending SDP offer to JVB Colibri conference {colibri_conference_id}...")
            jvb_response = await colibri.add_rtp_endpoint(colibri_conference_id, local_sdp_offer)
            remote_sdp_answer = jvb_response.get("sdp")
            if not remote_sdp_answer:
                raise Exception("JVB did not return an SDP answer")
            logger.debug(f"{log_prefix} JVB SDP Answer received:\n{remote_sdp_answer}")

            logger.debug(f"{log_prefix} Step 4: Processing JVB's SDP answer in Kurento...")
            await kurento.invoke(self.rtp_endpoint_id, "processAnswer", {"answer": remote_sdp_answer})
            logger.info(f"{log_prefix} ✅ SDP negotiation complete.")

            logger.debug(f"{log_prefix} Step 5: Starting Kurento recorder...")
            await kurento.invoke(self.recorder_id, "record")
            logger.info(f"{log_prefix} ✅ RECORDING IS ACTIVE")

        except Exception as e:
            logger.error(f"{log_prefix} ❌ Failed to start recording: {e}", exc_info=True)
            await self.cleanup()
            raise

    async def stop_recording(self):
        log_prefix = f"[Room: {self.room_name}] [User: {self.display_name}]"
        self.duration = (datetime.now(timezone.utc) - self.started_at).total_seconds()
        logger.info(f"{log_prefix} ⏹️ Stopping recording (duration: {self.duration:.1f}s)")
        if self.recorder_id:
            try:
                await kurento.invoke(self.recorder_id, "stopAndWait")
                logger.debug(f"{log_prefix} Kurento stopAndWait completed.")
            except Exception as e:
                logger.error(f"{log_prefix} Error on Kurento stopAndWait: {e}")

        await asyncio.sleep(1)

        if self.filepath.exists():
            file_size = self.filepath.stat().st_size
            logger.info(f"{log_prefix} 📁 File found: {self.filepath} ({file_size} bytes)")
            if file_size > 100:
                await self.upload_to_s3()
            else:
                logger.warning(f"{log_prefix} ⚠️ File is empty, skipping upload.")
                self.filepath.unlink()
        else:
            logger.error(f"{log_prefix} ❌ File NOT FOUND: {self.filepath}")

        await self.cleanup()

    async def upload_to_s3(self):
        log_prefix = f"[Room: {self.room_name}] [User: {self.display_name}]"
        if not s3_client:
            return
        try:
            self.s3_key = f"recordings/{self.room_name}/{self.filename}"
            logger.info(f"{log_prefix} ☁️ Uploading to s3://{MINIO_BUCKET}/{self.s3_key}")
            await asyncio.to_thread(
                s3_client.upload_file, str(self.filepath), MINIO_BUCKET, self.s3_key,
                ExtraArgs={
                    'ContentType': 'audio/webm',
                    'Metadata': {
                        'participantId': self.participant_id, 'displayName': self.display_name,
                        'roomName': self.room_name, 'startedAt': self.started_at.isoformat(),
                        'durationSeconds': str(round(self.duration, 2))
                    }
                }
            )
            self.filepath.unlink()
            self.uploaded = True
            logger.info(f"{log_prefix} ✅ Uploaded and removed local file.")
        except Exception as e:
            logger.error(f"{log_prefix} ❌ S3 Upload failed: {e}", exc_info=True)

    async def cleanup(self):
        log_prefix = f"[Room: {self.room_name}] [User: {self.display_name}]"
        logger.debug(f"{log_prefix} Cleaning up Kurento resources...")
        if self.pipeline_id:
            await kurento.release(self.pipeline_id)
        logger.debug(f"{log_prefix} Cleanup complete.")

async def handle_participant_joined(room_name: str, participant_id: str, display_name: str):
    log_prefix = f"[Room: {room_name}]"
    logger.info(f"{log_prefix} 📥 Handling PARTICIPANT JOINED for {display_name} (ID: {participant_id})")

    if 'focus' in participant_id.lower():
        logger.debug(f"{log_prefix} Skipping system component 'focus'")
        return

    if room_name not in active_conferences:
        logger.info(f"{log_prefix} First real participant. Creating new Colibri conference...")
        try:
            colibri_conf = await colibri.create_conference()
            active_conferences[room_name] = {'colibri_id': colibri_conf['id'], 'started_at': datetime.now(timezone.utc), 'participants': set()}
            logger.info(f"{log_prefix} 📹 New Colibri conference created: {colibri_conf['id']}")
        except Exception as e:
            logger.error(f"{log_prefix} ❌ Failed to create Colibri conference: {e}", exc_info=True)
            return

    active_conferences[room_name]['participants'].add(participant_id)
    logger.debug(f"{log_prefix} Current participants: {list(active_conferences[room_name]['participants'])}")

    if AUTO_RECORD:
        if participant_id in active_sessions:
            logger.warning(f"{log_prefix} ⚠️ {display_name} already has an active session. Stopping old one.")
            await handle_participant_left(participant_id)

        try:
            session = SessionInfo(participant_id, room_name, display_name)
            active_sessions[participant_id] = session
            colibri_id = active_conferences[room_name]['colibri_id']
            await session.start_recording(colibri_id)
        except Exception:
            if participant_id in active_sessions:
                del active_sessions[participant_id]

async def handle_participant_left(participant_id: str):
    logger.info(f"📤 Handling PARTICIPANT LEFT for ID: {participant_id}")
    if participant_id in active_sessions:
        session = active_sessions.pop(participant_id)
        room_name = session.room_name
        await session.stop_recording()

        if room_name in active_conferences:
            conf = active_conferences[room_name]
            conf['participants'].discard(participant_id)
            logger.debug(f"[Room: {room_name}] Participant removed. Remaining: {list(conf['participants'])}")

            # ИСПРАВЛЕНО: Завершаем конференцию, только если не осталось реальных участников
            is_real_participant_left = any('focus' not in p for p in conf['participants'])
            if not is_real_participant_left:
                 logger.info(f"[Room: {room_name}] Last real participant left. Ending conference.")
                 await handle_conference_ended(room_name)
    else:
        logger.debug(f"Participant {participant_id} had no active session.")

async def handle_conference_ended(room_name: str):
    logger.info(f"🏁 Handling CONFERENCE ENDED for room: {room_name}")
    if room_name in active_conferences:
        conf = active_conferences.pop(room_name)
        colibri_id = conf['colibri_id']
        logger.info(f"[Room: {room_name}] Expiring Colibri conference {colibri_id}...")
        await colibri.expire_conference(colibri_id)
        duration = (datetime.now(timezone.utc) - conf['started_at']).total_seconds()
        logger.info(f"[Room: {room_name}] ✅ Colibri conference expired. (Duration: {duration:.1f}s)")

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
        elif event_type == 'conferenceEnded':
            await handle_conference_ended(room_name)
        else:
            logger.debug(f"Ignoring event type '{event_type}'")

        return web.json_response({'status': 'ok'})
    except Exception as e:
        logger.error(f"❌ Webhook handler error: {e}", exc_info=True)
        return web.json_response({'status': 'error', 'message': str(e)}, status=500)

async def health_handler(request):
    kurento_healthy = False
    try:
        await kurento._send_request("ping", {"interval": 5000})
        kurento_healthy = True
    except Exception as e:
        logger.warning(f"Kurento health check failed: {e}")

    status_code = 200 if kurento_healthy else 503
    status = {
        'status': 'healthy' if kurento_healthy else 'unhealthy',
        'details': {
            'worker_id': WORKER_ID, 'auto_record': AUTO_RECORD,
            'active_conferences': len(active_conferences), 'active_sessions': len(active_sessions),
            'kurento_status': 'connected' if kurento_healthy else 'disconnected'
        }
    }
    return web.json_response(status, status=status_code)

async def stats_handler(request):
    stats = {
        'active_conferences': [{'room': room, **conf, 'participants': list(conf['participants'])} for room, conf in active_conferences.items()],
        'active_sessions': [{
            'participant_id': s.participant_id, 'display_name': s.display_name, 'room': s.room_name,
            'started_at': s.started_at.isoformat(), 'filename': s.filename,
            'duration_seconds': (datetime.now(timezone.utc) - s.started_at).total_seconds()
        } for s in active_sessions.values()]
    }
    return web.json_response(stats, dumps=lambda x: json.dumps(x, default=str))

async def main():
    global redis_client
    logger.info("🚀 Starting Unified Recorder")
    try:
        redis_client = await redis.from_url(f"redis://{REDIS_HOST}:{REDIS_PORT}", decode_responses=True)
        await redis_client.ping()
        logger.info("✅ Redis connected")
    except Exception as e:
        logger.warning(f"⚠️ Redis connection failed: {e}")

    app = web.Application()
    app.router.add_get('/health', health_handler)
    app.router.add_get('/stats', stats_handler)
    app.router.add_post('/events', webhook_handler)

    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, '0.0.0.0', 8080)
    await site.start()

    logger.info("🌐 HTTP Server started on :8080")
    logger.info("   GET  /health, /stats")
    logger.info("   POST /events")
    logger.info("✅ Unified Recorder ready.")

    await asyncio.Event().wait()

if __name__ == '__main__':
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("👋 Shutting down...")
    finally:
        loop = asyncio.get_event_loop()
        if colibri and colibri.session and not colibri.session.closed:
            loop.run_until_complete(colibri.close())
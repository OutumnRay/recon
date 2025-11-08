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
import subprocess
from datetime import datetime, timezone
from pathlib import Path
from typing import Optional, Dict, Any
from aiohttp import web, ClientSession
import boto3
import websockets
import redis.asyncio as redis

# Настройка логирования
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# ========== Конфигурация ==========
# Kurento
KURENTO_WS_URI = os.getenv('KURENTO_URI', 'ws://kurento:8888/kurento')

# MinIO / S3
MINIO_ENDPOINT = os.getenv('MINIO_ENDPOINT', 'https://api.storage.recontext.online')
MINIO_ACCESS_KEY = os.getenv('MINIO_ACCESS_KEY', 'minioadmin')
MINIO_SECRET_KEY = os.getenv('MINIO_SECRET_KEY', 'minioadmin')
MINIO_BUCKET = os.getenv('MINIO_BUCKET', 'jitsi-recordings')
MINIO_SECURE = os.getenv('MINIO_SECURE', 'true').lower() == 'true'

# Redis
REDIS_HOST = os.getenv('REDIS_HOST', 'redis')
REDIS_PORT = int(os.getenv('REDIS_PORT', '6379'))

# Jitsi
JITSI_DOMAIN = os.getenv('JITSI_DOMAIN', 'meet.recontext.online')
JVB_HOST = os.getenv('JVB_HOST', 'jvb')
JVB_COLIBRI_PORT = os.getenv('JVB_COLIBRI_PORT', '8080')

# Recording
RECORD_DIR = Path(os.getenv('RECORD_DIR', '/recordings'))
AUTO_RECORD = os.getenv('AUTO_RECORD', 'true').lower() == 'true'
AUDIO_BITRATE = os.getenv('AUDIO_BITRATE', '128k')
VIDEO_CODEC = os.getenv('VIDEO_CODEC', 'vp8')  # vp8, vp9, h264

WORKER_ID = os.getenv('HOSTNAME', 'unified-recorder')

RECORD_DIR.mkdir(exist_ok=True, parents=True)

logger.info("=" * 60)
logger.info("🎬 Jitsi Unified Recorder Configuration")
logger.info("=" * 60)
logger.info(f"  KURENTO_URI: {KURENTO_WS_URI}")
logger.info(f"  MINIO_ENDPOINT: {MINIO_ENDPOINT}")
logger.info(f"  MINIO_BUCKET: {MINIO_BUCKET}")
logger.info(f"  JITSI_DOMAIN: {JITSI_DOMAIN}")
logger.info(f"  JVB_HOST: {JVB_HOST}:{JVB_COLIBRI_PORT}")
logger.info(f"  AUTO_RECORD: {AUTO_RECORD}")
logger.info(f"  WORKER_ID: {WORKER_ID}")
logger.info("=" * 60)

# ========== S3 Client ==========
s3_client = None
if MINIO_ACCESS_KEY and MINIO_SECRET_KEY:
    try:
        s3_client = boto3.client(
            's3',
            endpoint_url=MINIO_ENDPOINT,
            aws_access_key_id=MINIO_ACCESS_KEY,
            aws_secret_access_key=MINIO_SECRET_KEY,
            region_name='us-east-1'
        )

        # Проверка bucket
        try:
            s3_client.head_bucket(Bucket=MINIO_BUCKET)
            logger.info(f"✅ MinIO bucket '{MINIO_BUCKET}' exists")
        except:
            logger.warning(f"⚠️ Bucket '{MINIO_BUCKET}' not accessible, trying to create...")
            try:
                s3_client.create_bucket(Bucket=MINIO_BUCKET)
                logger.info(f"✅ Created bucket '{MINIO_BUCKET}'")
            except Exception as e:
                logger.error(f"❌ Failed to create bucket: {e}")
    except Exception as e:
        logger.error(f"❌ S3 initialization failed: {e}")
else:
    logger.warning("⚠️ S3 credentials not set")

# ========== Redis ==========
redis_client = None

# ========== Global State ==========
active_sessions = {}  # {participant_id: SessionInfo}
active_conferences = {}  # {room_name: ConferenceInfo}


# ========== Kurento Client ==========
class KurentoClient:
    """WebSocket client for Kurento Media Server"""

    def __init__(self, ws_uri: str):
        self.ws_uri = ws_uri
        self.session_id = None
        self.request_id = 0

    async def _send_request(self, method: str, params: Dict[str, Any]) -> Any:
        """Send JSON-RPC request"""
        self.request_id += 1

        request = {
            "id": self.request_id,
            "method": method,
            "params": params,
            "jsonrpc": "2.0"
        }

        try:
            # Ensure URI has proper format
            uri = self.ws_uri
            if not uri.startswith('ws://') and not uri.startswith('wss://'):
                uri = f'ws://{uri}'

            logger.debug(f"Connecting to Kurento: {uri}")

            async with websockets.connect(
                uri,
                ping_interval=20,
                ping_timeout=10,
                subprotocols=['kurento']  # Kurento requires this subprotocol
            ) as websocket:
                logger.debug(f"Connected to Kurento, sending: {method}")
                await websocket.send(json.dumps(request))
                response_str = await websocket.recv()
                response = json.loads(response_str)

                if "error" in response:
                    raise Exception(f"Kurento error: {response['error']}")

                if self.session_id is None and "sessionId" in response.get("result", {}):
                    self.session_id = response["result"]["sessionId"]

                logger.debug(f"Kurento response: {method} -> OK")
                return response.get("result")
        except Exception as e:
            logger.error(f"Kurento request failed ({method}): {e}")
            raise

    async def create_media_pipeline(self) -> str:
        """Create MediaPipeline"""
        result = await self._send_request("create", {
            "type": "MediaPipeline",
            "constructorParams": {},
            "properties": {}
        })
        pipeline_id = result.get("value")
        logger.debug(f"Created MediaPipeline: {pipeline_id}")
        return pipeline_id

    async def create_recorder_endpoint(self, pipeline_id: str, output_path: str, profile: str = "WEBM") -> str:
        """Create RecorderEndpoint"""
        uri = f"file://{output_path}"

        result = await self._send_request("create", {
            "type": "RecorderEndpoint",
            "constructorParams": {
                "mediaPipeline": pipeline_id,
                "uri": uri,
                "mediaProfile": profile  # WEBM, WEBM_AUDIO_ONLY, WEBM_VIDEO_ONLY, MP4
            },
            "properties": {}
        })
        recorder_id = result.get("value")
        logger.debug(f"Created RecorderEndpoint: {recorder_id} -> {uri}")
        return recorder_id

    async def create_rtp_endpoint(self, pipeline_id: str) -> str:
        """Create RtpEndpoint to receive media from JVB"""
        result = await self._send_request("create", {
            "type": "RtpEndpoint",
            "constructorParams": {
                "mediaPipeline": pipeline_id
            },
            "properties": {}
        })
        rtp_id = result.get("value")
        logger.debug(f"Created RtpEndpoint: {rtp_id}")
        return rtp_id

    async def connect_endpoints(self, source_id: str, sink_id: str):
        """Connect two endpoints"""
        await self._send_request("invoke", {
            "object": source_id,
            "operation": "connect",
            "operationParams": {
                "sink": sink_id
            }
        })
        logger.debug(f"Connected {source_id} -> {sink_id}")

    async def record(self, recorder_id: str):
        """Start recording"""
        await self._send_request("invoke", {
            "object": recorder_id,
            "operation": "record",
            "operationParams": {}
        })
        logger.debug(f"Started recording: {recorder_id}")

    async def stop_and_wait(self, recorder_id: str):
        """Stop recording"""
        try:
            await self._send_request("invoke", {
                "object": recorder_id,
                "operation": "stopAndWait",
                "operationParams": {}
            })
            logger.debug(f"Stopped recording: {recorder_id}")
        except Exception as e:
            logger.warning(f"Error stopping recorder: {e}")

    async def release(self, object_id: str):
        """Release media object"""
        try:
            await self._send_request("release", {
                "object": object_id
            })
            logger.debug(f"Released object: {object_id}")
        except Exception as e:
            logger.debug(f"Error releasing object: {e}")


kurento = KurentoClient(KURENTO_WS_URI)


# ========== Session Management ==========
class SessionInfo:
    """Recording session information"""

    def __init__(self, participant_id: str, room_name: str, display_name: str):
        self.participant_id = participant_id
        self.room_name = room_name
        self.display_name = display_name
        self.started_at = datetime.now(timezone.utc)

        # Kurento objects
        self.pipeline_id = None
        self.rtp_endpoint_id = None
        self.recorder_id = None

        # File info
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
        safe_participant = self._sanitize(participant_id.split('@')[0])
        safe_room = self._sanitize(room_name)

        self.filename = f"{safe_room}_{safe_participant}_{timestamp}.webm"
        self.filepath = RECORD_DIR / self.filename
        self.s3_key = None

        # Stats
        self.duration = 0
        self.uploaded = False

    @staticmethod
    def _sanitize(name: str) -> str:
        """Sanitize filename"""
        return "".join(c if c.isalnum() or c in ('-', '_') else '_' for c in name)[:50]

    async def start_recording(self):
        """Start Kurento recording pipeline"""
        try:
            logger.info(f"🎙️ Starting recording: {self.display_name} -> {self.filename}")

            # 1. Create pipeline
            self.pipeline_id = await kurento.create_media_pipeline()

            # 2. Create RTP endpoint (to receive from JVB)
            self.rtp_endpoint_id = await kurento.create_rtp_endpoint(self.pipeline_id)

            # 3. Create recorder endpoint
            self.recorder_id = await kurento.create_recorder_endpoint(
                self.pipeline_id,
                str(self.filepath),
                "WEBM"  # или WEBM_AUDIO_ONLY для только аудио
            )

            # 4. Connect RTP -> Recorder
            await kurento.connect_endpoints(self.rtp_endpoint_id, self.recorder_id)

            # 5. Start recording
            await kurento.record(self.recorder_id)

            logger.info(f"✅ Recording started: {self.filename}")
            logger.info(f"   Pipeline: {self.pipeline_id}")
            logger.info(f"   RTP Endpoint: {self.rtp_endpoint_id}")
            logger.info(f"   Recorder: {self.recorder_id}")

            # NOTE: В реальности нужно настроить JVB чтобы он отправлял RTP в этот endpoint
            # Для этого нужен SDP offer/answer exchange через Colibri REST API

        except Exception as e:
            logger.error(f"❌ Failed to start recording: {e}", exc_info=True)
            await self.cleanup()
            raise

    async def stop_recording(self):
        """Stop recording and upload"""
        try:
            self.duration = (datetime.now(timezone.utc) - self.started_at).total_seconds()

            logger.info(f"⏹️ Stopping recording: {self.display_name} (duration: {self.duration:.1f}s)")

            # Stop recorder
            if self.recorder_id:
                await kurento.stop_and_wait(self.recorder_id)

            # Wait a bit for file to be written
            await asyncio.sleep(2)

            # Upload to S3
            if self.filepath.exists():
                file_size = self.filepath.stat().st_size
                logger.info(f"📁 Recording file: {self.filepath} ({file_size} bytes)")

                if file_size > 0:
                    await self.upload_to_s3()
                else:
                    logger.warning(f"⚠️ Empty file, skipping upload: {self.filepath}")
                    self.filepath.unlink()
            else:
                logger.error(f"❌ Recording file not found: {self.filepath}")

        except Exception as e:
            logger.error(f"❌ Error stopping recording: {e}", exc_info=True)
        finally:
            await self.cleanup()

    async def upload_to_s3(self):
        """Upload to MinIO/S3"""
        if not s3_client:
            logger.warning("S3 client not configured")
            return

        try:
            conference_id = hashlib.md5(f"{self.room_name}_{self.started_at.isoformat()}".encode()).hexdigest()[:8]
            self.s3_key = f"recordings/{self._sanitize(self.room_name)}/{conference_id}/{self.filename}"

            logger.info(f"☁️ Uploading to s3://{MINIO_BUCKET}/{self.s3_key}")

            await asyncio.to_thread(
                s3_client.upload_file,
                str(self.filepath),
                MINIO_BUCKET,
                self.s3_key,
                ExtraArgs={
                    'ContentType': 'video/webm',
                    'Metadata': {
                        'participantId': self.participant_id,
                        'displayName': self.display_name,
                        'roomName': self.room_name,
                        'startedAt': self.started_at.isoformat(),
                        'durationSeconds': str(round(self.duration, 2))
                    }
                }
            )

            self.filepath.unlink()
            self.uploaded = True

            logger.info(f"✅ Uploaded and removed local file: {self.s3_key}")

        except Exception as e:
            logger.error(f"❌ Upload failed: {e}", exc_info=True)

    async def cleanup(self):
        """Release Kurento resources"""
        try:
            if self.recorder_id:
                await kurento.release(self.recorder_id)
            if self.rtp_endpoint_id:
                await kurento.release(self.rtp_endpoint_id)
            if self.pipeline_id:
                await kurento.release(self.pipeline_id)
        except Exception as e:
            logger.debug(f"Cleanup error: {e}")


# ========== Event Handlers ==========
async def handle_participant_joined(room_name: str, endpoint_id: str, participant_id: str, display_name: str):
    """Handle participant joined event"""
    logger.info(f"📥 PARTICIPANT JOINED: {display_name} (ID: {participant_id}) in room {room_name}")

    # Skip system components
    if endpoint_id == 'focus' or 'focus' in participant_id.lower():
        logger.debug(f"Skipping system component: {endpoint_id}")
        return

    # Create conference if not exists
    if room_name not in active_conferences:
        active_conferences[room_name] = {
            'started_at': datetime.now(timezone.utc),
            'participants': []
        }
        logger.info(f"📹 Conference started: {room_name}")

    active_conferences[room_name]['participants'].append(participant_id)

    # Start recording if auto-record enabled
    if AUTO_RECORD:
        try:
            session = SessionInfo(participant_id, room_name, display_name)
            await session.start_recording()

            active_sessions[participant_id] = session
            logger.info(f"✅ Auto-recording started for {display_name}")

        except Exception as e:
            logger.error(f"❌ Failed to start auto-recording: {e}", exc_info=True)


async def handle_participant_left(room_name: str, endpoint_id: str, participant_id: str):
    """Handle participant left event"""
    logger.info(f"📤 PARTICIPANT LEFT: {participant_id} from room {room_name}")

    if participant_id in active_sessions:
        session = active_sessions[participant_id]
        await session.stop_recording()
        del active_sessions[participant_id]

        logger.info(f"✅ Recording stopped for {session.display_name}")
    else:
        logger.debug(f"Participant {participant_id} not in active sessions (already stopped or focus)")


async def handle_conference_ended(room_name: str):
    """Handle conference ended event"""
    logger.info(f"🏁 CONFERENCE ENDED: {room_name}")

    # Stop all active sessions for this room
    participants_to_stop = [
        pid for pid, session in active_sessions.items()
        if session.room_name == room_name
    ]

    for participant_id in participants_to_stop:
        await handle_participant_left(room_name, '', participant_id)

    # Remove conference
    if room_name in active_conferences:
        conf = active_conferences[room_name]
        duration = (datetime.now(timezone.utc) - conf['started_at']).total_seconds()
        logger.info(f"✅ Conference completed: {room_name} (duration: {duration:.1f}s)")
        del active_conferences[room_name]


# ========== HTTP API ==========
async def webhook_handler(request):
    """Webhook endpoint for Prosody events"""
    try:
        data = await request.json()
        event_type = data.get('eventType')
        room_name = data.get('roomName')
        endpoint_id = data.get('endpointId', '')

        logger.debug(f"🔔 Webhook event: {event_type} for room {room_name}")

        if event_type == 'participantJoined':
            participant_id = data.get('participantId', endpoint_id)
            display_name = data.get('displayName', 'Unknown')
            await handle_participant_joined(room_name, endpoint_id, participant_id, display_name)

        elif event_type == 'participantLeft':
            # Need to map endpoint_id to participant_id
            participant_id = None
            for pid, session in active_sessions.items():
                if session.room_name == room_name:
                    participant_id = pid
                    break

            if participant_id:
                await handle_participant_left(room_name, endpoint_id, participant_id)

        elif event_type == 'conferenceEnded':
            await handle_conference_ended(room_name)

        return web.json_response({'status': 'ok'})

    except Exception as e:
        logger.error(f"❌ Webhook error: {e}", exc_info=True)
        return web.json_response({'status': 'error', 'message': str(e)}, status=500)


async def health_handler(request):
    """Health check"""
    status = {
        'status': 'healthy',
        'worker_id': WORKER_ID,
        'auto_record': AUTO_RECORD,
        'active_conferences': len(active_conferences),
        'active_sessions': len(active_sessions),
        'kurento_uri': KURENTO_WS_URI,
        's3_configured': s3_client is not None
    }
    return web.json_response(status)


async def stats_handler(request):
    """Statistics"""
    stats = {
        'active_conferences': [
            {
                'room': room,
                'started_at': conf['started_at'].isoformat(),
                'participants': len(conf['participants'])
            }
            for room, conf in active_conferences.items()
        ],
        'active_sessions': [
            {
                'participant_id': session.participant_id,
                'display_name': session.display_name,
                'room': session.room_name,
                'started_at': session.started_at.isoformat(),
                'duration': (datetime.now(timezone.utc) - session.started_at).total_seconds(),
                'filename': session.filename
            }
            for session in active_sessions.values()
        ]
    }
    return web.json_response(stats)


async def http_server():
    """Start HTTP server"""
    app = web.Application()
    app.router.add_get('/health', health_handler)
    app.router.add_get('/stats', stats_handler)
    app.router.add_post('/events', webhook_handler)

    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, '0.0.0.0', 8080)
    await site.start()

    logger.info("🌐 HTTP Server started on :8080")
    logger.info("   GET  /health - Health check")
    logger.info("   GET  /stats  - Statistics")
    logger.info("   POST /events - Webhook endpoint")


async def main():
    global redis_client

    logger.info("🚀 Starting Unified Recorder")

    # Connect to Redis
    try:
        redis_client = await redis.from_url(
            f"redis://{REDIS_HOST}:{REDIS_PORT}",
            decode_responses=True
        )
        await redis_client.ping()
        logger.info(f"✅ Redis connected")
    except Exception as e:
        logger.warning(f"⚠️ Redis connection failed: {e}")

    # Start HTTP server
    await http_server()

    logger.info("✅ Unified Recorder ready")

    # Keep running
    while True:
        await asyncio.sleep(3600)


if __name__ == '__main__':
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("👋 Shutting down...")

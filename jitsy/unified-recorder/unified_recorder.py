#!/usr/bin/env python3
"""
Jitsi Unified Recorder - aiortc WebRTC recorder
Based on a combination of best practices from provided scripts.
This version listens for webhook events from Prosody to start and stop recordings.
"""
import os
import asyncio
import json
import logging
from datetime import datetime, timezone
from pathlib import Path
from typing import Dict, Optional

from aiohttp import web
import boto3
import websockets
from aiortc import RTCPeerConnection, RTCSessionDescription, RTCIceCandidate
from aiortc.contrib.media import MediaRecorder

# --- Configuration ---
# Logging configuration
LOG_LEVEL = os.getenv('LOG_LEVEL', 'INFO').upper()
logging.basicConfig(level=LOG_LEVEL, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# Jitsi and WebRTC configuration
JITSI_DOMAIN = os.getenv('JITSI_DOMAIN', 'meet.recontext.online')
JVB_WS_URL = os.getenv('JVB_WS_URL', f'wss://{JITSI_DOMAIN}/colibri-ws/default-id/default')

# MinIO S3 storage configuration
MINIO_ENDPOINT = os.getenv('MINIO_ENDPOINT', 'https://api.storage.recontext.online')
MINIO_ACCESS_KEY = os.getenv('MINIO_ACCESS_KEY', 'minioadmin')
MINIO_SECRET_KEY = os.getenv('MINIO_SECRET_KEY', 'minioadmin')
MINIO_BUCKET = os.getenv('MINIO_BUCKET', 'jitsi-recordings')

# Recording settings
RECORD_DIR = Path(os.getenv('RECORD_DIR', '/recordings'))
AUTO_RECORD = os.getenv('AUTO_RECORD', 'true').lower() == 'true'

# Ensure the recording directory exists
RECORD_DIR.mkdir(exist_ok=True, parents=True)

# --- Logging Initial Configuration ---
logger.info("=" * 60)
logger.info("🎬 Jitsi Unified aiortc WebRTC Recorder")
logger.info("=" * 60)
logger.info(f"  JITSI_DOMAIN: {JITSI_DOMAIN}")
logger.info(f"  JVB_WS_URL: {JVB_WS_URL}")
logger.info(f"  MINIO_BUCKET: {MINIO_BUCKET}")
logger.info(f"  AUTO_RECORD: {AUTO_RECORD}")
logger.info("=" * 60)

# --- S3 Client Initialization ---
s3_client = None
if MINIO_ACCESS_KEY and MINIO_SECRET_KEY:
    try:
        s3_client = boto3.client(
            's3',
            endpoint_url=MINIO_ENDPOINT,
            aws_access_key_id=MINIO_ACCESS_KEY,
            aws_secret_access_key=MINIO_SECRET_KEY
        )
        # Verify connection by checking for the bucket
        s3_client.head_bucket(Bucket=MINIO_BUCKET)
        logger.info(f"✅ MinIO connection successful to bucket: {MINIO_BUCKET}")
    except Exception as e:
        logger.error(f"❌ Failed to initialize S3 client: {e}")
        s3_client = None

# --- Global State ---
# Dictionary to store active recording sessions, keyed by participant ID
active_sessions: Dict[str, 'SessionInfo'] = {}

class SessionInfo:
    """
    Manages a recording session for a single participant using aiortc.
    """

    def __init__(self, participant_id: str, room_name: str, display_name: str):
        self.participant_id = participant_id
        self.room_name = room_name
        self.display_name = display_name
        self.log_prefix = f"[{self.room_name}][{self.display_name}]"
        self.started_at = datetime.now(timezone.utc)

        # WebRTC components
        self.pc: Optional[RTCPeerConnection] = None
        self.websocket: Optional[websockets.WebSocketClientProtocol] = None
        self.recorder: Optional[MediaRecorder] = None
        self.signaling_task: Optional[asyncio.Task] = None

        # File naming and path
        timestamp = self.started_at.strftime('%Y%m%d_%H%M%S')
        # Sanitize display name for use in filenames
        safe_name = "".join(c for c in display_name if c.isalnum())[:50] or "unknown"
        self.filename = f"{self.room_name}_{safe_name}_{timestamp}.opus"
        self.filepath = RECORD_DIR / self.filename

        logger.info(f"{self.log_prefix} 📝 Session created. Recording will be saved to: {self.filename}")

    async def start_recording(self):
        """Initializes the WebRTC peer connection and starts the recording process."""
        logger.info(f"{self.log_prefix} 🎙️ Starting recording...")
        try:
            self.pc = RTCPeerConnection()

            @self.pc.on("track")
            async def on_track(track):
                logger.info(f"{self.log_prefix} ✅ Received {track.kind} track from JVB.")
                if track.kind == "audio":
                    self.recorder = MediaRecorder(str(self.filepath))
                    self.recorder.addTrack(track)
                    await self.recorder.start()
                    logger.info(f"{self.log_prefix} 🔴 Started recording audio to {self.filepath}")

            @self.pc.on("connectionstatechange")
            async def on_connectionstatechange():
                logger.info(f"{self.log_prefix} Connection state changed to: {self.pc.connectionState}")
                if self.pc.connectionState in ["failed", "closed", "disconnected"]:
                    logger.warning(f"{self.log_prefix} Connection state is {self.pc.connectionState}. Stopping session.")
                    await self.stop_recording()

            # The signaling task handles all WebSocket communication with JVB
            self.signaling_task = asyncio.create_task(self._run_signaling())
            logger.info(f"{self.log_prefix} ✅ Recording session started successfully.")
        except Exception as e:
            logger.error(f"{self.log_prefix} ❌ Failed to start recording session: {e}", exc_info=True)
            await self.stop_recording()

    async def _run_signaling(self):
        """Handles the WebSocket signaling with Jitsi Videobridge (JVB)."""
        try:
            logger.info(f"{self.log_prefix} 🔌 Connecting to JVB WebSocket at {JVB_WS_URL}")
            async with websockets.connect(JVB_WS_URL) as ws:
                self.websocket = ws
                logger.info(f"{self.log_prefix} ✅ WebSocket connection to JVB established.")

                # Create and send a join message
                join_msg = json.dumps({
                    "colibriClass": "EndpointMessage",
                    "type": "join",
                    "roomName": self.room_name,
                    "endpointId": self.participant_id,
                    "displayName": self.display_name
                })
                await ws.send(join_msg)
                logger.info(f"{self.log_prefix} 📤 Sent 'join' request to JVB.")

                # Create and send SDP offer
                offer = await self.pc.createOffer()
                await self.pc.setLocalDescription(offer)
                offer_msg = json.dumps({
                    "colibriClass": "EndpointMessage",
                    "type": "offer",
                    "sdp": self.pc.localDescription.sdp
                })
                await ws.send(offer_msg)
                logger.info(f"{self.log_prefix} 📤 Sent SDP offer to JVB.")

                # Process incoming messages from JVB
                async for message in ws:
                    try:
                        msg = json.loads(message)
                        msg_type = msg.get("type")
                        logger.debug(f"{self.log_prefix} 📥 Received message of type: {msg_type}")

                        if msg_type == "answer":
                            answer = RTCSessionDescription(sdp=msg.get("sdp"), type="answer")
                            await self.pc.setRemoteDescription(answer)
                            logger.info(f"{self.log_prefix} ✅ Set remote SDP answer from JVB.")
                        elif msg_type == "ice-candidate":
                            candidate_data = msg.get("candidate")
                            if candidate_data:
                                candidate = RTCIceCandidate(
                                    candidate=candidate_data.get("candidate"),
                                    sdpMid=candidate_data.get("sdpMid"),
                                    sdpMLineIndex=candidate_data.get("sdpMLineIndex")
                                )
                                await self.pc.addIceCandidate(candidate)
                        elif msg.get("colibriClass") == "EndpointConnectivityStatusChangeEvent":
                            # This event can indicate connection issues
                            is_connected = msg.get("connected", False)
                            if not is_connected:
                                logger.warning(f"{self.log_prefix} ⚠️ Endpoint connectivity lost.")
                                await self.stop_recording()
                                break
                    except json.JSONDecodeError:
                        logger.error(f"{self.log_prefix} ❌ Could not decode JSON from message: {message}")
                    except Exception as e:
                        logger.error(f"{self.log_prefix} ❌ Error handling incoming message: {e}", exc_info=True)
        except websockets.exceptions.ConnectionClosed as e:
            logger.warning(f"{self.log_prefix} 🔌 WebSocket connection closed unexpectedly: {e}")
        except Exception as e:
            logger.error(f"{self.log_prefix} ❌ An error occurred in the signaling task: {e}", exc_info=True)
        finally:
            logger.info(f"{self.log_prefix} 🔌 Signaling task ended.")
            if self.participant_id in active_sessions:
                await self.stop_recording()

    async def stop_recording(self):
        """Stops the recording, cleans up resources, and initiates the upload."""
        if self.participant_id not in active_sessions:
            logger.debug(f"{self.log_prefix} Stop recording called, but session already removed.")
            return

        logger.info(f"{self.log_prefix} ⏹️ Stopping recording session...")
        active_sessions.pop(self.participant_id, None)

        if self.signaling_task and not self.signaling_task.done():
            self.signaling_task.cancel()
        if self.recorder:
            await self.recorder.stop()
        if self.pc:
            await self.pc.close()

        logger.info(f"{self.log_prefix} 🛑 All WebRTC resources have been released.")

        # Brief pause to ensure the file is written to disk
        await asyncio.sleep(1)

        if self.filepath.exists():
            file_size = self.filepath.stat().st_size
            logger.info(f"{self.log_prefix} 📁 Recording file size: {file_size} bytes.")
            # Avoid uploading empty or tiny files
            if file_size > 1024:
                await self.upload_to_s3()
            else:
                logger.warning(f"{self.log_prefix} ⚠️ File is too small to upload. Deleting.")
                self.filepath.unlink()
        else:
            logger.warning(f"{self.log_prefix} ⚠️ Recording file not found at path: {self.filepath}")

    async def upload_to_s3(self):
        """Uploads the completed recording file to MinIO S3 storage."""
        if not s3_client:
            logger.warning(f"{self.log_prefix} ⚠️ S3 client is not configured. Skipping upload.")
            return

        s3_key = f"recordings/{self.room_name}/{self.filename}"
        logger.info(f"{self.log_prefix} ☁️ Uploading to s3://{MINIO_BUCKET}/{s3_key}")

        try:
            await asyncio.to_thread(
                s3_client.upload_file,
                str(self.filepath),
                MINIO_BUCKET,
                s3_key,
                ExtraArgs={
                    'ContentType': 'audio/opus',
                    'Metadata': {
                        'participant_id': self.participant_id,
                        'display_name': self.display_name,
                        'room_name': self.room_name,
                        'started_at': self.started_at.isoformat()
                    }
                }
            )
            logger.info(f"{self.log_prefix} ✅ Successfully uploaded to S3.")
            self.filepath.unlink() # Delete local file after successful upload
            logger.info(f"{self.log_prefix} 🗑️ Local recording file deleted.")
        except Exception as e:
            logger.error(f"{self.log_prefix} ❌ Failed to upload to S3: {e}", exc_info=True)

# --- Webhook Event Handlers ---
async def handle_participant_joined(room_name: str, participant_id: str, display_name: str):
    logger.info(f"📥 PARTICIPANT JOINED: '{display_name}' ({participant_id}) in room '{room_name}'")

    if 'focus' in display_name.lower():
        logger.info("Ignoring 'focus' user, which is a Jitsi internal component.")
        return

    if not AUTO_RECORD:
        logger.info("AUTO_RECORD is disabled. Not starting recording.")
        return

    if participant_id in active_sessions:
        logger.warning(f"A session for participant {participant_id} already exists. Stopping the old one.")
        await handle_participant_left(participant_id)

    session = SessionInfo(participant_id, room_name, display_name)
    active_sessions[participant_id] = session
    await session.start_recording()

async def handle_participant_left(participant_id: str):
    logger.info(f"📤 PARTICIPANT LEFT: {participant_id}")
    if participant_id in active_sessions:
        session = active_sessions[participant_id]
        await session.stop_recording()
    else:
        logger.debug(f"Received participant left event for an untracked session: {participant_id}")

async def handle_conference_ended(room_name: str):
    logger.info(f"🏁 CONFERENCE ENDED: {room_name}. Stopping all related recordings.")
    # Create a copy of the keys to avoid issues with modifying the dictionary while iterating
    participant_ids_to_stop = [
        pid for pid, session in active_sessions.items() if session.room_name == room_name
    ]
    for pid in participant_ids_to_stop:
        await handle_participant_left(pid)

# --- HTTP Server ---
async def webhook_handler(request: web.Request):
    """Handles incoming webhook events from Prosody."""
    try:
        data = await request.json()
        logger.info(f"🔔 Webhook received: {json.dumps(data, indent=2)}")

        event_type = data.get('eventType')
        room_name = data.get('roomName')
        participant_id = data.get('participantId')

        if event_type == 'participantJoined' and all([room_name, participant_id]):
            display_name = data.get('displayName', 'Unknown')
            await handle_participant_joined(room_name, participant_id, display_name)
        elif event_type == 'participantLeft' and participant_id:
            await handle_participant_left(participant_id)
        elif event_type == 'conferenceEnded' and room_name:
            await handle_conference_ended(room_name)
        else:
            logger.warning(f"Received unhandled or incomplete event: {event_type}")

        return web.json_response({'status': 'ok'})
    except Exception as e:
        logger.error(f"❌ Error processing webhook: {e}", exc_info=True)
        return web.json_response({'status': 'error', 'message': str(e)}, status=500)

async def health_handler(request: web.Request):
    """Provides a health check endpoint."""
    return web.json_response({
        'status': 'healthy',
        'active_sessions': len(active_sessions),
        'sessions': [{
            'participant': s.display_name,
            'room': s.room_name,
            'duration_seconds': (datetime.now(timezone.utc) - s.started_at).total_seconds()
        } for s in active_sessions.values()]
    })

async def main():
    logger.info("🚀 Starting Jitsi Unified Recorder...")
    app = web.Application()
    app.router.add_post('/events', webhook_handler)
    app.router.add_get('/health', health_handler)

    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, '0.0.0.0', 8080)
    await site.start()

    logger.info("=" * 60)
    logger.info("✅ Recorder is running and listening on port 8080")
    logger.info("  - POST /events : Webhook endpoint for Jitsi events")
    logger.info("  - GET  /health  : Health check and session status")
    logger.info("=" * 60)

    # Keep the application running indefinitely
    await asyncio.Event().wait()

if __name__ == '__main__':
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("👋 Shutting down recorder...")
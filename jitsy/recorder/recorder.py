"""
Jitsi Individual Stream Recorder with Kurento Media Server
REST API for recording individual participant streams and uploading to MinIO
"""

import os
import json
import uuid
import asyncio
import logging
from typing import Dict, Optional, Any
from datetime import datetime
from pathlib import Path

import boto3
from botocore.client import Config
from fastapi import FastAPI, HTTPException, BackgroundTasks
from fastapi.responses import JSONResponse
from pydantic import BaseModel
import websockets

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s'
)
logger = logging.getLogger(__name__)

# Configuration from environment
KURENTO_URI = os.getenv("KURENTO_URI", "ws://kurento:8888/kurento")
MINIO_ENDPOINT = os.getenv("MINIO_ENDPOINT", "https://api.storage.recontext.online")
MINIO_ACCESS_KEY = os.getenv("MINIO_ACCESS_KEY", "minioadmin")
MINIO_SECRET_KEY = os.getenv("MINIO_SECRET_KEY", "minioadmin")
MINIO_BUCKET = os.getenv("MINIO_BUCKET", "jitsi-recordings")
MINIO_SECURE = os.getenv("MINIO_SECURE", "true").lower() == "true"
RECORDINGS_DIR = Path("/recordings")

# Ensure recordings directory exists
RECORDINGS_DIR.mkdir(parents=True, exist_ok=True)


# ============================================================================
# Kurento WebSocket Client
# ============================================================================

class KurentoClient:
    """Simple Kurento Media Server client using WebSocket and JSON-RPC"""

    def __init__(self, ws_uri: str):
        self.ws_uri = ws_uri
        self.session_id = None
        self.request_id = 0

    async def _send_request(self, method: str, params: Dict[str, Any]) -> Any:
        """Send JSON-RPC request to Kurento"""
        self.request_id += 1

        request = {
            "id": self.request_id,
            "method": method,
            "params": params,
            "jsonrpc": "2.0"
        }

        async with websockets.connect(self.ws_uri) as websocket:
            await websocket.send(json.dumps(request))
            response_str = await websocket.recv()
            response = json.loads(response_str)

            if "error" in response:
                raise Exception(f"Kurento error: {response['error']}")

            # Save session ID from first response
            if self.session_id is None and "sessionId" in response.get("result", {}):
                self.session_id = response["result"]["sessionId"]

            return response.get("result")

    async def create_media_pipeline(self) -> str:
        """Create a MediaPipeline and return its ID"""
        result = await self._send_request("create", {
            "type": "MediaPipeline",
            "constructorParams": {},
            "properties": {}
        })
        return result.get("value")

    async def create_recorder_endpoint(self, pipeline_id: str, uri: str) -> str:
        """Create a RecorderEndpoint in the pipeline"""
        result = await self._send_request("create", {
            "type": "RecorderEndpoint",
            "constructorParams": {
                "mediaPipeline": pipeline_id,
                "uri": uri,
                "mediaProfile": "WEBM"
            },
            "properties": {}
        })
        return result.get("value")

    async def record(self, recorder_id: str):
        """Start recording"""
        await self._send_request("invoke", {
            "object": recorder_id,
            "operation": "record",
            "operationParams": {}
        })

    async def stop_and_wait(self, recorder_id: str):
        """Stop recording and wait for it to finish"""
        await self._send_request("invoke", {
            "object": recorder_id,
            "operation": "stopAndWait",
            "operationParams": {}
        })

    async def release(self, object_id: str):
        """Release a media object"""
        await self._send_request("invoke", {
            "object": object_id,
            "operation": "release",
            "operationParams": {}
        })


# FastAPI app
app = FastAPI(
    title="Jitsi Kurento Recorder API",
    description="REST API for recording individual Jitsi participant streams",
    version="1.0.0"
)

# MinIO S3 client
s3_client = boto3.client(
    "s3",
    endpoint_url=MINIO_ENDPOINT,
    aws_access_key_id=MINIO_ACCESS_KEY,
    aws_secret_access_key=MINIO_SECRET_KEY,
    config=Config(signature_version='s3v4'),
    use_ssl=MINIO_SECURE
)

# Active recordings storage
# Format: {user_id: {"pipeline": pipeline, "recorder": recorder, "filepath": path, "started_at": timestamp}}
active_recordings: Dict[str, Dict] = {}


class RecordingRequest(BaseModel):
    """Recording request model"""
    user: str
    room: Optional[str] = None
    sdp_offer: Optional[str] = None  # SDP offer from client for WebRTC negotiation


class RecordingResponse(BaseModel):
    """Recording response model"""
    status: str
    user: str
    message: str
    filepath: Optional[str] = None
    minio_key: Optional[str] = None
    started_at: Optional[str] = None
    stopped_at: Optional[str] = None
    duration: Optional[float] = None


def ensure_bucket_exists():
    """Ensure MinIO bucket exists, create if not"""
    try:
        s3_client.head_bucket(Bucket=MINIO_BUCKET)
        logger.info(f"✅ MinIO bucket '{MINIO_BUCKET}' exists")
    except Exception:
        try:
            s3_client.create_bucket(Bucket=MINIO_BUCKET)
            logger.info(f"✅ Created MinIO bucket '{MINIO_BUCKET}'")
        except Exception as e:
            logger.error(f"❌ Failed to create bucket: {e}")


async def upload_to_minio(filepath: Path, user_id: str, room: Optional[str] = None) -> str:
    """
    Upload recording file to MinIO

    Args:
        filepath: Local file path
        user_id: User identifier
        room: Room name (optional)

    Returns:
        MinIO object key
    """
    try:
        # Generate MinIO object key
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        room_prefix = f"{room}/" if room else ""
        object_key = f"{room_prefix}{user_id}_{timestamp}.webm"

        # Upload file
        logger.info(f"⬆️  Uploading {filepath.name} to MinIO as {object_key}")
        s3_client.upload_file(
            str(filepath),
            MINIO_BUCKET,
            object_key,
            ExtraArgs={'ContentType': 'video/webm'}
        )

        logger.info(f"✅ Uploaded to MinIO: {object_key}")

        # Delete local file after successful upload
        if filepath.exists():
            filepath.unlink()
            logger.info(f"🗑️  Deleted local file: {filepath}")

        return object_key

    except Exception as e:
        logger.error(f"❌ Failed to upload to MinIO: {e}")
        raise HTTPException(status_code=500, detail=f"MinIO upload failed: {str(e)}")


async def create_recording(user_id: str, room: Optional[str] = None) -> Dict:
    """
    Create Kurento recording pipeline for a user

    Args:
        user_id: User identifier
        room: Room name (optional)

    Returns:
        Dictionary with pipeline_id, recorder_id, and filepath
    """
    try:
        # Connect to Kurento
        logger.info(f"🔌 Connecting to Kurento at {KURENTO_URI}")
        kms = KurentoClient(KURENTO_URI)

        # Create media pipeline
        pipeline_id = await kms.create_media_pipeline()
        logger.info(f"📹 Created MediaPipeline for user {user_id}: {pipeline_id}")

        # Generate filename
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        filename = f"{user_id}_{timestamp}.webm"
        filepath = RECORDINGS_DIR / filename

        # Create recorder endpoint
        recorder_id = await kms.create_recorder_endpoint(
            pipeline_id,
            f"file://{filepath}"
        )

        logger.info(f"🎙️  Created RecorderEndpoint: {filepath}")

        # Start recording
        await kms.record(recorder_id)
        logger.info(f"▶️  Recording started for user {user_id}")

        return {
            "pipeline_id": pipeline_id,
            "recorder_id": recorder_id,
            "filepath": filepath,
            "started_at": datetime.now().isoformat(),
            "kms": kms
        }

    except Exception as e:
        logger.error(f"❌ Error creating recording: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to create recording: {str(e)}")


async def stop_recording_task(user_id: str, room: Optional[str] = None):
    """
    Background task to stop recording and upload to MinIO

    Args:
        user_id: User identifier
        room: Room name (optional)
    """
    try:
        if user_id not in active_recordings:
            logger.warning(f"⚠️  No active recording found for user {user_id}")
            return

        recording = active_recordings[user_id]
        pipeline_id = recording["pipeline_id"]
        recorder_id = recording["recorder_id"]
        filepath = recording["filepath"]
        started_at = recording["started_at"]
        kms = recording["kms"]

        # Stop recorder
        logger.info(f"⏹️  Stopping recording for user {user_id}")
        await kms.stop_and_wait(recorder_id)

        # Release recorder and pipeline
        await kms.release(recorder_id)
        await kms.release(pipeline_id)
        logger.info(f"🔓 Released MediaPipeline for user {user_id}")

        # Wait a bit for file to be fully written
        await asyncio.sleep(2)

        # Upload to MinIO
        if filepath.exists() and filepath.stat().st_size > 0:
            minio_key = await upload_to_minio(filepath, user_id, room)

            # Calculate duration
            stopped_at = datetime.now()
            started_dt = datetime.fromisoformat(started_at)
            duration = (stopped_at - started_dt).total_seconds()

            logger.info(f"✅ Recording completed for {user_id} (duration: {duration:.1f}s, MinIO: {minio_key})")
        else:
            logger.warning(f"⚠️  Recording file is empty or missing: {filepath}")

        # Remove from active recordings
        del active_recordings[user_id]

    except Exception as e:
        logger.error(f"❌ Error stopping recording: {e}")
        # Clean up even on error
        if user_id in active_recordings:
            del active_recordings[user_id]


@app.on_event("startup")
async def startup_event():
    """Initialize on startup"""
    logger.info("🚀 Jitsi Kurento Recorder starting...")
    logger.info(f"   Kurento URI: {KURENTO_URI}")
    logger.info(f"   MinIO Endpoint: {MINIO_ENDPOINT}")
    logger.info(f"   MinIO Bucket: {MINIO_BUCKET}")
    logger.info(f"   Recordings Dir: {RECORDINGS_DIR}")

    # Ensure MinIO bucket exists
    ensure_bucket_exists()

    logger.info("✅ Recorder ready")


@app.get("/")
async def root():
    """Root endpoint"""
    return {
        "service": "Jitsi Kurento Recorder",
        "version": "1.0.0",
        "status": "running",
        "active_recordings": len(active_recordings),
        "endpoints": {
            "start": "POST /record/start?user=<user_id>&room=<room_name>",
            "stop": "POST /record/stop?user=<user_id>&room=<room_name>",
            "status": "GET /record/status?user=<user_id>",
            "list": "GET /record/list"
        }
    }


@app.get("/health")
async def health():
    """Health check endpoint"""
    return {"status": "healthy", "active_recordings": len(active_recordings)}


@app.post("/record/start", response_model=RecordingResponse)
async def start_recording(user: str, room: Optional[str] = None):
    """
    Start recording for a user

    Args:
        user: User identifier
        room: Room name (optional)

    Returns:
        Recording status
    """
    logger.info(f"📥 Received start request for user: {user}, room: {room}")

    if user in active_recordings:
        raise HTTPException(
            status_code=400,
            detail=f"Recording already active for user {user}"
        )

    try:
        # Create recording
        recording = await create_recording(user, room)
        active_recordings[user] = recording

        return RecordingResponse(
            status="started",
            user=user,
            message=f"Recording started for user {user}",
            filepath=str(recording["filepath"]),
            started_at=recording["started_at"]
        )

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"❌ Unexpected error starting recording: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/record/stop", response_model=RecordingResponse)
async def stop_recording(
    user: str,
    room: Optional[str] = None,
    background_tasks: BackgroundTasks = None
):
    """
    Stop recording for a user

    Args:
        user: User identifier
        room: Room name (optional)
        background_tasks: FastAPI background tasks

    Returns:
        Recording status
    """
    logger.info(f"📥 Received stop request for user: {user}, room: {room}")

    if user not in active_recordings:
        raise HTTPException(
            status_code=404,
            detail=f"No active recording found for user {user}"
        )

    try:
        recording = active_recordings[user]

        # Schedule stop and upload in background
        background_tasks.add_task(stop_recording_task, user, room)

        return RecordingResponse(
            status="stopping",
            user=user,
            message=f"Recording stopping for user {user}",
            filepath=str(recording["filepath"]),
            started_at=recording["started_at"],
            stopped_at=datetime.now().isoformat()
        )

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"❌ Unexpected error stopping recording: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/record/status")
async def get_recording_status(user: str):
    """
    Get recording status for a user

    Args:
        user: User identifier

    Returns:
        Recording status
    """
    if user not in active_recordings:
        return JSONResponse(
            status_code=404,
            content={"status": "not_found", "user": user, "message": "No active recording"}
        )

    recording = active_recordings[user]
    started_at = datetime.fromisoformat(recording["started_at"])
    duration = (datetime.now() - started_at).total_seconds()

    return {
        "status": "recording",
        "user": user,
        "filepath": str(recording["filepath"]),
        "started_at": recording["started_at"],
        "duration": duration
    }


@app.get("/record/list")
async def list_recordings():
    """
    List all active recordings

    Returns:
        List of active recordings
    """
    recordings = []

    for user_id, recording in active_recordings.items():
        started_at = datetime.fromisoformat(recording["started_at"])
        duration = (datetime.now() - started_at).total_seconds()

        recordings.append({
            "user": user_id,
            "filepath": str(recording["filepath"]),
            "started_at": recording["started_at"],
            "duration": duration
        })

    return {
        "active_recordings": len(recordings),
        "recordings": recordings
    }


@app.delete("/record/stop-all")
async def stop_all_recordings(background_tasks: BackgroundTasks):
    """
    Stop all active recordings

    Args:
        background_tasks: FastAPI background tasks

    Returns:
        Status of all stopped recordings
    """
    logger.info("🛑 Stopping all active recordings")

    stopped_users = list(active_recordings.keys())

    for user_id in stopped_users:
        background_tasks.add_task(stop_recording_task, user_id)

    return {
        "status": "stopping_all",
        "count": len(stopped_users),
        "users": stopped_users
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080, log_level="info")

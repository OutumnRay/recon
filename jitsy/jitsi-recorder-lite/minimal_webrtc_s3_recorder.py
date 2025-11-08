"""
Минимальный WebRTC recorder для Jitsi:
- подключается к JVB напрямую через WebSocket (без браузера)
- сохраняет каждый поступивший audio track в Opus
- сразу заливает готовый файл на S3/MinIO

Чтобы не усложнять код, все настройки берутся из переменных окружения, а
протокол Colibri обрабатывается в самом простом виде (offer/answer + ICE).
"""
import asyncio
import json
import logging
import os
from datetime import datetime
from pathlib import Path
from typing import Optional

import boto3
import websockets
from aiortc import (RTCConfiguration, RTCPeerConnection, RTCSessionDescription,
                    RTCIceServer)
from av import open as av_open


LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO").upper()
logging.basicConfig(
    level=getattr(logging, LOG_LEVEL, logging.INFO),
    format="%(asctime)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger("minimal_webrtc_recorder")


class S3Uploader:
    """Тривиальный загрузчик файлов на S3."""

    def __init__(self) -> None:
        self.bucket = os.getenv("S3_BUCKET", "jitsi-recordings")
        endpoint = os.getenv("S3_ENDPOINT")
        aws_key = os.getenv("AWS_ACCESS_KEY_ID")
        aws_secret = os.getenv("AWS_SECRET_ACCESS_KEY")
        region = os.getenv("AWS_REGION", "us-east-1")

        if not (endpoint and aws_key and aws_secret):
            logger.warning("⚠️  S3 env vars not set => uploads disabled")
            self.client = None
            return

        self.client = boto3.client(
            "s3",
            endpoint_url=endpoint,
            aws_access_key_id=aws_key,
            aws_secret_access_key=aws_secret,
            region_name=region,
        )

        logger.info("☁️  S3 client ready: %s/%s", endpoint, self.bucket)

    def upload(self, local_path: Path, key_prefix: str) -> Optional[str]:
        if not self.client:
            return None

        key = f"{key_prefix.rstrip('/')}/{local_path.name}"
        self.client.upload_file(str(local_path), self.bucket, key)
        logger.info("☁️  Uploaded %s to s3://%s/%s", local_path.name, self.bucket, key)
        return key


class TrackWriter:
    """Записывает один audio track в Opus и триггерит заливку на S3."""

    def __init__(
        self,
        room: str,
        participant_id: str,
        track,
        output_dir: Path,
        uploader: S3Uploader,
        s3_prefix: str,
    ) -> None:
        self.room = room
        self.participant_id = participant_id
        self.track = track
        self.output_dir = output_dir
        self.uploader = uploader
        self.s3_prefix = s3_prefix

        timestamp = datetime.utcnow().strftime("%Y%m%d_%H%M%S_%f")[:-3]
        safe_room = "".join(c if c.isalnum() or c in "-_" else "_" for c in room)
        safe_participant = "".join(
            c if c.isalnum() or c in "-_" else "_" for c in participant_id
        )
        self.filepath = output_dir / f"{safe_room}_{safe_participant}_{timestamp}.opus"

        self._container = None
        self._stream = None
        self._task = None
        self._closed = False

    async def start(self) -> None:
        self._container = av_open(str(self.filepath), "w")
        self._stream = self._container.add_stream("opus", rate=48000)
        self._stream.channels = 2
        self._task = asyncio.create_task(self._record_loop())
        logger.info("🎙️  Recording started: %s", self.filepath.name)

    async def _record_loop(self) -> None:
        try:
            while True:
                frame = await self.track.recv()
                for packet in self._stream.encode(frame):
                    self._container.mux(packet)
        except Exception as exc:
            logger.info(
                "🔚 Track %s finished (%s)", self.participant_id, getattr(exc, "args", exc)
            )
        finally:
            await self.finish()

    async def finish(self) -> None:
        if self._closed:
            return
        self._closed = True

        if self._stream:
            for packet in self._stream.encode():
                self._container.mux(packet)
        if self._container:
            self._container.close()
        size = self.filepath.stat().st_size if self.filepath.exists() else 0
        logger.info("💾 Saved %s (%d bytes)", self.filepath.name, size)

        if self.uploader:
            self.uploader.upload(self.filepath, self.s3_prefix)

    async def wait(self) -> None:
        if self._task:
            try:
                await self._task
            except asyncio.CancelledError:
                pass


class MinimalWebRTCRecorder:
    """Прямое подключение к Colibri WebSocket и сбор audio tracks."""

    def __init__(
        self,
        room: str,
        conference_id: str,
        jvb_host: str,
        jvb_port: str,
        output_dir: Path,
        uploader: S3Uploader,
    ) -> None:
        self.room = room
        self.conference_id = conference_id
        self.jvb_host = jvb_host
        self.jvb_port = jvb_port
        self.output_dir = output_dir
        self.uploader = uploader

        self.pc: Optional[RTCPeerConnection] = None
        self.ws = None
        self.recorders: dict[str, TrackWriter] = {}
        self.s3_prefix = f"recordings/{self.room}/{self.conference_id}"

    async def start(self) -> None:
        rtc_configuration = RTCConfiguration(
            iceServers=[RTCIceServer(urls=["stun:stun.l.google.com:19302"])]
        )
        self.pc = RTCPeerConnection(rtc_configuration)

        @self.pc.on("track")
        async def _on_track(track):
            if track.kind != "audio":
                return
            recorder = TrackWriter(
                room=self.room,
                participant_id=track.id,
                track=track,
                output_dir=self.output_dir,
                uploader=self.uploader,
                s3_prefix=self.s3_prefix,
            )
            await recorder.start()
            self.recorders[track.id] = recorder
            asyncio.create_task(self._cleanup_on_end(track.id, recorder))

        ws_url = f"ws://{self.jvb_host}:{self.jvb_port}/colibri-ws/default-id/{self.conference_id}"
        logger.info("🔌 Connecting to %s", ws_url)
        self.ws = await websockets.connect(ws_url)

        asyncio.create_task(self._read_ws())

        offer = await self.pc.createOffer()
        await self.pc.setLocalDescription(offer)

        await self.ws.send(
            json.dumps(
                {
                    "colibriClass": "EndpointMessage",
                    "type": "offer",
                    "sdp": offer.sdp,
                }
            )
        )

        @self.pc.on("icecandidate")
        async def _on_ice(candidate):
            if candidate:
                await self.ws.send(
                    json.dumps(
                        {
                            "colibriClass": "EndpointMessage",
                            "type": "candidate",
                            "candidate": candidate.candidate,
                            "sdpMid": candidate.sdpMid,
                            "sdpMLineIndex": candidate.sdpMLineIndex,
                        }
                    )
                )

        logger.info("✅ Offer sent, waiting for media…")

    async def _read_ws(self) -> None:
        try:
            async for raw in self.ws:
                message = json.loads(raw)
                msg_type = message.get("type")
                if msg_type == "answer" and self.pc:
                    await self.pc.setRemoteDescription(
                        RTCSessionDescription(sdp=message["sdp"], type="answer")
                    )
                    logger.info("✅ SDP answer applied")
                elif msg_type == "candidate" and self.pc:
                    # aiortc обрабатывает trickle ICE автоматически
                    pass
                elif msg_type == "ping":
                    await self.ws.send(json.dumps({"type": "pong"}))
                else:
                    logger.debug("📨 %s", message)
        except websockets.ConnectionClosed:
            logger.info("🔌 WebSocket closed by JVB")

    async def _cleanup_on_end(self, track_id: str, recorder: TrackWriter) -> None:
        await recorder.wait()
        self.recorders.pop(track_id, None)

    async def stop(self) -> None:
        for recorder in list(self.recorders.values()):
            await recorder.finish()
        self.recorders.clear()

        if self.pc:
            await self.pc.close()
        if self.ws:
            await self.ws.close()


async def main():
    room = os.getenv("JITSI_ROOM", "testmeet")
    conference_id = os.getenv("JITSI_CONFERENCE_ID", "test123")
    jvb_host = os.getenv("JVB_HOST", "localhost")
    jvb_port = os.getenv("JVB_PORT", "8080")
    output_dir = Path(os.getenv("RECORD_DIR", "/tmp/recordings"))
    output_dir.mkdir(parents=True, exist_ok=True)

    uploader = S3Uploader()
    recorder = MinimalWebRTCRecorder(
        room=room,
        conference_id=conference_id,
        jvb_host=jvb_host,
        jvb_port=jvb_port,
        output_dir=output_dir,
        uploader=uploader,
    )

    await recorder.start()

    keep_alive = int(os.getenv("RECORD_SECONDS", "0"))
    try:
        if keep_alive > 0:
            await asyncio.sleep(keep_alive)
        else:
            while True:
                await asyncio.sleep(1)
    except KeyboardInterrupt:
        logger.info("⏹️  Stop requested")
    finally:
        await recorder.stop()


if __name__ == "__main__":
    asyncio.run(main())

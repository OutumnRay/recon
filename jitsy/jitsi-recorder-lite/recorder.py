import os
import asyncio
import json
import subprocess
import uuid
import hashlib
from datetime import datetime, timezone
from pathlib import Path
import boto3
import redis.asyncio as redis
from aiohttp import ClientSession, web
import logging

# Import WebRTC recorder
from simple_webrtc_recorder import SimpleWebRTCRecorder

# Настройка логирования с детализацией
logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s - %(name)s - %(levelname)s - [%(funcName)s:%(lineno)d] - %(message)s'
)
logger = logging.getLogger(__name__)

# Конфигурация
JITSI_DOMAIN = os.getenv('JITSI_DOMAIN', 'meet.recontext.online')
JVB_HOST = os.getenv('JVB_HOST', 'jvb')
JVB_PORT = os.getenv('JVB_PORT', '8080')
PROSODY_HOST = os.getenv('PROSODY_HOST', 'prosody')

# S3
S3_ENDPOINT = os.getenv('S3_ENDPOINT')
S3_BUCKET = os.getenv('S3_BUCKET', 'jitsi-recordings')
AWS_ACCESS_KEY = os.getenv('AWS_ACCESS_KEY_ID', 'minioadmin')
AWS_SECRET_KEY = os.getenv('AWS_SECRET_ACCESS_KEY','minioadmin')
AWS_REGION = os.getenv('AWS_REGION', 'us-east-1')

# Webhook
WEBHOOK_URL = os.getenv('WEBHOOK_URL', '')

# Redis
REDIS_HOST = os.getenv('REDIS_HOST', 'redis')
REDIS_PORT = int(os.getenv('REDIS_PORT', '6379'))

# Параметры
RECORD_DIR = '/tmp/recordings'
AUDIO_BITRATE = os.getenv('AUDIO_BITRATE', '48k')
CHECK_INTERVAL = int(os.getenv('CHECK_INTERVAL', '5'))
RECONNECT_TIMEOUT = 30

WORKER_ID = os.getenv('HOSTNAME', str(uuid.uuid4())[:8])

Path(RECORD_DIR).mkdir(exist_ok=True, parents=True)

logger.info(f"🔧 Configuration loaded:")
logger.info(f"  JITSI_DOMAIN: {JITSI_DOMAIN}")
logger.info(f"  JVB_HOST: {JVB_HOST}:{JVB_PORT}")
logger.info(f"  PROSODY_HOST: {PROSODY_HOST}")
logger.info(f"  REDIS_HOST: {REDIS_HOST}:{REDIS_PORT}")
logger.info(f"  S3_BUCKET: {S3_BUCKET}")
logger.info(f"  WEBHOOK_URL: {WEBHOOK_URL if WEBHOOK_URL else 'NOT SET (optional)'}")
logger.info(f"  WORKER_ID: {WORKER_ID}")

# Инициализация S3 только если credentials заданы
s3_client = None
if AWS_ACCESS_KEY and AWS_SECRET_KEY:
    try:
        s3_client = boto3.client(
            's3',
            endpoint_url=S3_ENDPOINT,
            aws_access_key_id=AWS_ACCESS_KEY,
            aws_secret_access_key=AWS_SECRET_KEY,
            region_name=AWS_REGION
        )
        logger.info(f"✅ S3 client initialized - endpoint: {S3_ENDPOINT}, bucket: {S3_BUCKET}")

        # Test S3 connection and create bucket if needed
        try:
            s3_client.head_bucket(Bucket=S3_BUCKET)
            logger.info(f"✅ S3 bucket '{S3_BUCKET}' exists and is accessible")
        except Exception as e:
            logger.warning(f"⚠️  Bucket '{S3_BUCKET}' not accessible: {e}")
            try:
                s3_client.create_bucket(Bucket=S3_BUCKET)
                logger.info(f"✅ Created S3 bucket '{S3_BUCKET}'")
            except Exception as create_error:
                logger.error(f"❌ Failed to create bucket: {create_error}")

    except Exception as e:
        logger.error(f"❌ S3 client initialization failed: {e}", exc_info=True)
else:
    logger.warning("⚠️  S3 credentials not set - uploads will be disabled")
    logger.warning(f"   AWS_ACCESS_KEY_ID: {'SET' if AWS_ACCESS_KEY else 'NOT SET'}")
    logger.warning(f"   AWS_SECRET_ACCESS_KEY: {'SET' if AWS_SECRET_KEY else 'NOT SET'}")
    logger.warning(f"   S3_ENDPOINT: {S3_ENDPOINT if S3_ENDPOINT else 'NOT SET'}")

redis_client = None
active_conferences = {}


class ParticipantSession:
    """Одна сессия участника"""

    def __init__(self, session_id, participant_id, participant_name, display_name, conference_start_time):
        self.sessionId = session_id
        self.participantId = participant_id
        self.participantName = participant_name
        self.displayName = display_name
        self.joinTime = datetime.now(timezone.utc)
        self.joinOffset = (self.joinTime - conference_start_time).total_seconds()
        self.leaveTime = None
        self.durationSeconds = 0
        self.process = None
        self.filepath = None
        self.filename = None
        self.s3Key = None
        self.isReconnection = False
        logger.debug(f"Created session {session_id} for participant {display_name}")

    async def start_recording(self, room_name, conference_id, stream_url):
        """Начинает запись"""
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S_%f')[:19]

        safe_room = self._sanitize_filename(room_name)
        safe_participant_id = self._sanitize_filename(self.participantId)

        # Extract short participant ID (everything before @)
        short_id = self.participantId.split('@')[0] if '@' in self.participantId else self.participantId
        safe_short_id = self._sanitize_filename(short_id)

        # Format: {room}_{participant_id}_{timestamp}.opus
        self.filename = f"{safe_room}_{safe_short_id}_{timestamp}.opus"
        self.filepath = os.path.join(RECORD_DIR, self.filename)

        cmd = [
            'ffmpeg',
            '-re',
            '-i', stream_url,
            '-vn',
            '-c:a', 'libopus',
            '-b:a', AUDIO_BITRATE,
            '-application', 'voip',
            '-frame_duration', '60',
            '-packet_loss', '15',
            '-vbr', 'on',
            '-compression_level', '10',
            '-f', 'opus',
            '-y',
            self.filepath
        ]

        logger.info(f"🎙️  START RECORDING: {self.filename} (offset: {self.joinOffset:.1f}s, reconnect: {self.isReconnection})")
        logger.info(f"   Stream URL: {stream_url}")
        logger.info(f"   Output file: {self.filepath}")
        logger.debug(f"   FFmpeg command: {' '.join(cmd)}")

        try:
            self.process = await asyncio.create_subprocess_exec(
                *cmd,
                stdout=asyncio.subprocess.DEVNULL,
                stderr=asyncio.subprocess.PIPE
            )
            logger.info(f"✅ FFmpeg process started (PID: {self.process.pid})")

            # Start monitoring stderr in background
            asyncio.create_task(self._monitor_ffmpeg_stderr())
        except Exception as e:
            logger.error(f"❌ Failed to start FFmpeg: {e}", exc_info=True)
            raise

    async def _monitor_ffmpeg_stderr(self):
        """Мониторит stderr FFmpeg для отладки"""
        if not self.process or not self.process.stderr:
            return

        try:
            while True:
                line = await self.process.stderr.readline()
                if not line:
                    break

                line_str = line.decode('utf-8', errors='ignore').strip()
                if line_str:
                    # Log only important FFmpeg messages
                    if any(keyword in line_str.lower() for keyword in ['error', 'failed', 'invalid', 'could not', 'unable']):
                        logger.error(f"[FFmpeg ERROR] {line_str}")
                    elif 'warning' in line_str.lower():
                        logger.warning(f"[FFmpeg WARN] {line_str}")
                    else:
                        logger.debug(f"[FFmpeg] {line_str}")
        except Exception as e:
            logger.debug(f"FFmpeg stderr monitoring stopped: {e}")

    async def stop_recording(self, room_name, conference_id):
        """Останавливает запись и загружает файл на S3"""
        self.leaveTime = datetime.now(timezone.utc)
        self.durationSeconds = (self.leaveTime - self.joinTime).total_seconds()

        logger.info(f"⏹️  STOP RECORDING: {self.filename} (duration: {self.durationSeconds:.1f}s)")

        # Stop FFmpeg process
        if self.process:
            try:
                # Check if process is still running before terminating
                if self.process.returncode is None:
                    self.process.terminate()
                    try:
                        await asyncio.wait_for(self.process.wait(), timeout=5)
                    except asyncio.TimeoutError:
                        logger.warning(f"FFmpeg didn't stop in time, killing process")
                        self.process.kill()
                        await self.process.wait()
                else:
                    logger.debug(f"Process already terminated with code {self.process.returncode}")
            except ProcessLookupError:
                logger.debug(f"Process already terminated (ProcessLookupError)")
            except Exception as e:
                logger.warning(f"Error stopping process: {e}")

        # Upload file to S3 immediately after stopping
        logger.info(f"📂 Checking for recording file: {self.filepath}")

        if os.path.exists(self.filepath):
            file_size = os.path.getsize(self.filepath)
            logger.info(f"📁 Recording file found: {self.filepath} ({file_size} bytes)")

            if file_size > 0:
                if s3_client:
                    # Upload to S3 right away (don't wait for conference end)
                    logger.info(f"→ Starting S3 upload for {self.filename}")
                    await self.upload_to_s3(room_name, conference_id)
                else:
                    logger.warning(f"⚠️  S3 client not initialized, file will be kept locally: {self.filepath}")
            else:
                logger.warning(f"⚠️  Recording file is EMPTY (0 bytes), skipping upload and removing: {self.filepath}")
                try:
                    os.remove(self.filepath)
                    logger.info(f"🗑️  Removed empty file: {self.filepath}")
                except Exception as e:
                    logger.warning(f"Failed to remove empty file: {e}")
        else:
            logger.error(f"❌ Recording file NOT FOUND: {self.filepath}")
            logger.error(f"   Expected path: {self.filepath}")
            logger.error(f"   Directory exists: {os.path.exists(os.path.dirname(self.filepath))}")
            logger.error(f"   FFmpeg process returncode: {self.process.returncode if self.process else 'No process'}")
            # List files in recording directory
            try:
                record_dir_files = os.listdir(RECORD_DIR)
                logger.error(f"   Files in {RECORD_DIR}: {record_dir_files}")
            except Exception as e:
                logger.error(f"   Failed to list directory: {e}")

    async def upload_to_s3(self, room_name, conference_id):
        """Загружает на S3"""
        if not s3_client:
            logger.error(f"❌ Cannot upload {self.filename}: S3 client not initialized")
            return

        try:
            safe_room = self._sanitize_filename(room_name)
            # Path structure: recordings/{room_name}/{conference_id}/{filename}
            self.s3Key = f"recordings/{safe_room}/{conference_id}/{self.filename}"

            logger.info(f"☁️  Starting upload to s3://{S3_BUCKET}/{self.s3Key}")
            logger.info(f"   Local file: {self.filepath} ({os.path.getsize(self.filepath)} bytes)")

            # Upload to S3
            await asyncio.to_thread(
                s3_client.upload_file,
                self.filepath,
                S3_BUCKET,
                self.s3Key,
                ExtraArgs={
                    'ContentType': 'audio/opus',
                    'Metadata': {
                        'roomName': room_name,
                        'conferenceId': conference_id,
                        'participantId': self.participantId,
                        'displayName': self.displayName,
                        'joinOffsetSeconds': str(round(self.joinOffset, 2)),
                        'durationSeconds': str(round(self.durationSeconds, 2)),
                        'isReconnection': str(self.isReconnection)
                    }
                }
            )

            # Remove local file after successful upload
            os.remove(self.filepath)
            logger.info(f"✅ Upload completed and local file removed: {self.s3Key}")

        except Exception as e:
            logger.error(f"❌ Upload failed for {self.filename}: {e}", exc_info=True)
            logger.error(f"   S3 Key: {self.s3Key if hasattr(self, 's3Key') else 'not set'}")
            logger.error(f"   Bucket: {S3_BUCKET}")
            logger.error(f"   Endpoint: {S3_ENDPOINT}")

    @staticmethod
    def _sanitize_filename(name):
        """Очищает имя файла"""
        return "".join(c if c.isalnum() or c in ('-', '_') else '_' for c in name)[:50]

    def to_dict(self):
        return {
            'sessionId': self.sessionId,
            'participantId': self.participantId,
            'participantName': self.participantName,
            'displayName': self.displayName,
            'joinTime': self.joinTime.isoformat(),
            'joinOffsetSeconds': round(self.joinOffset, 2),
            'leaveTime': self.leaveTime.isoformat() if self.leaveTime else None,
            'durationSeconds': round(self.durationSeconds, 2),
            'filename': self.filename,
            's3Key': self.s3Key,
            'isReconnection': self.isReconnection
        }


class Participant:
    """Участник конференции"""

    def __init__(self, participant_id, participant_name, display_name):
        self.participantId = participant_id
        self.participantName = participant_name
        self.displayName = display_name
        self.sessions = []
        self.lastLeaveTime = None
        self.totalDuration = 0
        logger.debug(f"Created participant: {display_name} ({participant_id})")

    def add_session(self, session: ParticipantSession):
        if self.lastLeaveTime:
            time_since_leave = (session.joinTime - self.lastLeaveTime).total_seconds()
            if time_since_leave < RECONNECT_TIMEOUT:
                session.isReconnection = True
                logger.info(f"🔄 Reconnection detected for {self.displayName} after {time_since_leave:.1f}s")
        self.sessions.append(session)

    def update_after_session_end(self, session: ParticipantSession):
        self.lastLeaveTime = session.leaveTime
        self.totalDuration += session.durationSeconds

    def to_dict(self):
        return {
            'participantId': self.participantId,
            'participantName': self.participantName,
            'displayName': self.displayName,
            'totalDurationSeconds': round(self.totalDuration, 2),
            'sessionsCount': len(self.sessions),
            'sessions': [s.to_dict() for s in self.sessions]
        }


class ConferenceRecording:
    """Запись конференции"""

    def __init__(self, room_name):
        self.roomName = room_name
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
        room_hash = hashlib.md5(room_name.encode()).hexdigest()[:8]
        self.conferenceId = f"{room_hash}_{timestamp}"
        self.startTime = datetime.now(timezone.utc)
        self.endTime = None
        self.participants = {}
        self.activeSessions = {}
        self.webrtc_recorder = None  # WebRTC recorder для этой конференции
        self.webrtc_connected = False
        logger.info(f"📹 Conference created: {room_name} (ID: {self.conferenceId})")

    async def start_webrtc_recording(self):
        """Запускает WebRTC recorder для конференции"""
        if self.webrtc_connected:
            logger.debug(f"WebRTC recorder already connected for {self.roomName}")
            return

        try:
            logger.info(f"🔌 Starting WebRTC recorder for room: {self.roomName}")

            #Создаем recorder
            self.webrtc_recorder = SimpleWebRTCRecorder(
                room_name=self.roomName,
                conference_id=self.conferenceId,
                jvb_host=JVB_HOST,
                jvb_port=JVB_PORT,
                output_dir=os.path.join(RECORD_DIR, self.roomName)
            )

            # Подключаемся к JVB
            await self.webrtc_recorder.connect()
            self.webrtc_connected = True

            logger.info(f"✅ WebRTC recorder started for {self.roomName}")

        except Exception as e:
            logger.error(f"❌ Failed to start WebRTC recorder: {e}", exc_info=True)

    async def stop_webrtc_recording(self):
        """Останавливает WebRTC recorder"""
        if not self.webrtc_connected or not self.webrtc_recorder:
            return

        try:
            logger.info(f"🛑 Stopping WebRTC recorder for room: {self.roomName}")

            await self.webrtc_recorder.disconnect()
            self.webrtc_connected = False

            logger.info(f"✅ WebRTC recorder stopped for {self.roomName}")

        except Exception as e:
            logger.error(f"❌ Error stopping WebRTC recorder: {e}", exc_info=True)

    def get_or_create_participant(self, participant_id, participant_name, display_name):
        if participant_id not in self.participants:
            self.participants[participant_id] = Participant(participant_id, participant_name, display_name)
        return self.participants[participant_id]

    def to_dict(self):
        return {
            'conferenceId': self.conferenceId,
            'roomName': self.roomName,
            'startTime': self.startTime.isoformat(),
            'endTime': self.endTime.isoformat() if self.endTime else None,
            'durationSeconds': round((self.endTime - self.startTime).total_seconds(), 2) if self.endTime else None,
            'participantsCount': len(self.participants),
            'totalSessions': sum(len(p.sessions) for p in self.participants.values()),
            'participants': [p.to_dict() for p in self.participants.values()]
        }

    def save_metadata(self):
        safe_room = ParticipantSession._sanitize_filename(self.roomName)
        metadata_dir = os.path.join(RECORD_DIR, safe_room, self.conferenceId)
        Path(metadata_dir).mkdir(parents=True, exist_ok=True)

        metadata_file = os.path.join(metadata_dir, 'metadata.json')
        with open(metadata_file, 'w', encoding='utf-8') as f:
            json.dump(self.to_dict(), f, indent=2, ensure_ascii=False)

        logger.info(f"💾 Metadata saved: {metadata_file}")
        return metadata_file


async def send_webhook(conference: ConferenceRecording):
    """Отправляет webhook"""
    if not WEBHOOK_URL:
        logger.debug("Webhook URL not configured, skipping")
        return

    try:
        payload = {
            'event': 'conferenceEnded',
            'conferenceId': conference.conferenceId,
            'roomName': conference.roomName,
            'startTime': conference.startTime.isoformat(),
            'endTime': conference.endTime.isoformat(),
            'durationSeconds': round((conference.endTime - conference.startTime).total_seconds(), 2),
            'participantsCount': len(conference.participants),
            'totalSessions': sum(len(p.sessions) for p in conference.participants.values()),
            'participants': [
                {
                    'participantId': p.participantId,
                    'participantName': p.participantName,
                    'displayName': p.displayName,
                    'totalDurationSeconds': round(p.totalDuration, 2),
                    'sessionsCount': len(p.sessions),
                    'recordings': [
                        {
                            'filename': s.filename,
                            's3Key': s.s3Key,
                            's3Url': f"s3://{S3_BUCKET}/{s.s3Key}" if s.s3Key else None,
                            'joinOffsetSeconds': round(s.joinOffset, 2),
                            'durationSeconds': round(s.durationSeconds, 2),
                            'isReconnection': s.isReconnection
                        }
                        for s in p.sessions
                    ]
                }
                for p in conference.participants.values()
            ],
            's3Path': f"recordings/{ParticipantSession._sanitize_filename(conference.roomName)}/{conference.conferenceId}/"
        }

        logger.info(f"🔔 Sending webhook to {WEBHOOK_URL}")
        logger.debug(f"Webhook payload: {json.dumps(payload, indent=2)}")

        async with ClientSession() as session:
            async with session.post(WEBHOOK_URL, json=payload, timeout=30) as resp:
                resp_text = await resp.text()
                if resp.status == 200:
                    logger.info(f"✅ Webhook sent successfully")
                else:
                    logger.warning(f"⚠️  Webhook failed: HTTP {resp.status} - {resp_text}")

    except Exception as e:
        logger.error(f"❌ Webhook error: {e}", exc_info=True)


async def upload_metadata_to_s3(conference: ConferenceRecording):
    """Загружает метаданные"""
    if not s3_client:
        logger.warning("S3 client not initialized, skipping metadata upload")
        return

    try:
        metadata_file = conference.save_metadata()
        safe_room = ParticipantSession._sanitize_filename(conference.roomName)
        s3_key = f"recordings/{safe_room}/{conference.conferenceId}/metadata.json"

        logger.info(f"☁️  Uploading metadata to s3://{S3_BUCKET}/{s3_key}")

        await asyncio.to_thread(
            s3_client.upload_file,
            metadata_file,
            S3_BUCKET,
            s3_key,
            ExtraArgs={'ContentType': 'application/json'}
        )

        os.remove(metadata_file)
        logger.info(f"✅ Metadata uploaded successfully")

    except Exception as e:
        logger.error(f"❌ Metadata upload failed: {e}", exc_info=True)


async def claim_room(room_name):
    """Захватывает комнату"""
    lock_key = f"recording:lock:{room_name}"
    worker_key = f"recording:worker:{room_name}"

    locked = await redis_client.set(lock_key, WORKER_ID, ex=60, nx=True)
    if locked:
        await redis_client.set(worker_key, WORKER_ID, ex=3600)
        logger.debug(f"🔒 Claimed room: {room_name}")
        return True

    owner = await redis_client.get(worker_key)
    if owner and owner == WORKER_ID:
        await redis_client.expire(lock_key, 60)
        return True

    logger.debug(f"🔓 Room {room_name} owned by another worker: {owner}")
    return False


async def release_room(room_name):
    """Освобождает комнату"""
    await redis_client.delete(f"recording:lock:{room_name}", f"recording:worker:{room_name}")
    logger.debug(f"🔓 Released room: {room_name}")


async def handle_participant_joined(room_name, endpoint_id, participant_id, participant_name, display_name, stream_url):
    """Обработка присоединения"""
    logger.info(f"📥 PARTICIPANT JOINED: {display_name} (ID: {participant_id}, endpoint: {endpoint_id}) in room {room_name}")

    # Skip system components like 'focus'
    if endpoint_id == 'focus' or participant_name.endswith('/focus') or 'focus@' in participant_id:
        logger.debug(f"Skipping system component: {endpoint_id}")
        return

    if not await claim_room(room_name):
        logger.warning(f"Cannot claim room {room_name}, skipping")
        return

    if room_name not in active_conferences:
        active_conferences[room_name] = ConferenceRecording(room_name)

    conference = active_conferences[room_name]
    participant = conference.get_or_create_participant(participant_id, participant_name, display_name)

    # Запускаем WebRTC recorder при первом участнике
    if not conference.webrtc_connected:
        await conference.start_webrtc_recording()

    # Create a NEW session for tracking metadata
    session_id = f"{endpoint_id}_{datetime.now().timestamp()}"
    session = ParticipantSession(session_id, participant_id, participant_name, display_name, conference.startTime)

    # Add session to participant's session list (for tracking all streams)
    participant.add_session(session)

    # Set as active session (will be stopped on participantLeft)
    conference.activeSessions[endpoint_id] = session

    # NOTE: Запись теперь происходит через WebRTC recorder автоматически
    # Файлы создаются когда WebRTC получает audio tracks


async def handle_participant_left(room_name, endpoint_id):
    """Обработка выхода"""
    logger.info(f"📤 PARTICIPANT LEFT: endpoint {endpoint_id} from room {room_name}")

    if room_name not in active_conferences:
        logger.warning(f"Room {room_name} not found in active conferences")
        return

    conference = active_conferences[room_name]
    if endpoint_id in conference.activeSessions:
        session = conference.activeSessions[endpoint_id]

        # Update session metadata (recording stops automatically in WebRTC)
        session.leaveTime = datetime.now(timezone.utc)
        session.durationSeconds = (session.leaveTime - session.joinTime).total_seconds()

        # Check if participant exists before accessing it
        if session.participantId in conference.participants:
            participant = conference.participants[session.participantId]
            participant.update_after_session_end(session)
        else:
            logger.warning(f"Participant {session.participantId} not found")

        # Safely remove from active sessions
        try:
            del conference.activeSessions[endpoint_id]
        except KeyError:
            logger.debug(f"Endpoint {endpoint_id} already removed from active sessions")
    else:
        logger.debug(f"Endpoint {endpoint_id} not in active sessions (may be system component like 'focus')")


async def handle_conference_ended(room_name):
    """Обработка завершения"""
    logger.info(f"🏁 CONFERENCE ENDED: {room_name}")

    if room_name not in active_conferences:
        logger.warning(f"Room {room_name} not found in active conferences")
        return

    conference = active_conferences[room_name]
    conference.endTime = datetime.now(timezone.utc)

    # Stop WebRTC recorder
    await conference.stop_webrtc_recording()

    # Stop all active sessions
    for endpoint_id in list(conference.activeSessions.keys()):
        await handle_participant_left(room_name, endpoint_id)

    # Upload metadata to S3
    await upload_metadata_to_s3(conference)

    # Send webhook notification
    await send_webhook(conference)

    # Release room lock
    await release_room(room_name)

    # Log summary
    duration = (conference.endTime - conference.startTime).total_seconds()
    total_sessions = sum(len(p.sessions) for p in conference.participants.values())
    total_recordings = sum(1 for p in conference.participants.values() for s in p.sessions if s.s3Key)

    logger.info(f"✅ Conference {room_name} completed:")
    logger.info(f"   Conference ID: {conference.conferenceId}")
    logger.info(f"   Duration: {duration:.1f}s")
    logger.info(f"   Participants: {len(conference.participants)}")
    logger.info(f"   Total sessions: {total_sessions}")
    logger.info(f"   Recordings uploaded to S3: {total_recordings}/{total_sessions}")

    if s3_client and total_recordings > 0:
        safe_room = ParticipantSession._sanitize_filename(room_name)
        s3_path = f"s3://{S3_BUCKET}/recordings/{safe_room}/{conference.conferenceId}/"
        logger.info(f"   S3 Path: {s3_path}")
        logger.info(f"   📁 Recordings:")
        for participant in conference.participants.values():
            for session in participant.sessions:
                if session.s3Key:
                    logger.info(f"      ✅ {session.filename} ({session.durationSeconds:.1f}s)")
                else:
                    logger.info(f"      ❌ {session.filename} (upload failed or file empty)")

    del active_conferences[room_name]


async def events_webhook_handler(request):
    """Обработчик webhook от Prosody"""
    try:
        # Логируем входящий запрос
        logger.info(f"🌐 Incoming webhook request from {request.remote}")

        data = await request.json()
        logger.info(f"📨 Webhook data: {json.dumps(data, indent=2)}")

        event_type = data.get('eventType')
        room_name = data.get('roomName')
        endpoint_id = data.get('endpointId')

        logger.info(f"🔔 Event: {event_type}, Room: {room_name}, Endpoint: {endpoint_id}")

        if event_type == 'participantJoined':
            # TODO: Получить реальный stream URL из JVB
            stream_url = f"rtp://{JVB_HOST}:10000"

            logger.info(f"→ Calling handle_participant_joined for endpoint {endpoint_id}")

            await handle_participant_joined(
                room_name,
                endpoint_id,
                data.get('participantId'),
                data.get('participantName'),
                data.get('displayName'),
                stream_url
            )

        elif event_type == 'participantLeft':
            logger.info(f"→ Calling handle_participant_left for endpoint {endpoint_id}")

            await handle_participant_left(
                room_name,
                endpoint_id
            )

        elif event_type == 'conferenceEnded':
            logger.info(f"→ Calling handle_conference_ended for room {room_name}")

            await handle_conference_ended(room_name)
        else:
            logger.warning(f"⚠️  Unknown event type: {event_type}")

        return web.json_response({'status': 'ok'})

    except json.JSONDecodeError as e:
        logger.error(f"❌ Invalid JSON in webhook: {e}")
        return web.json_response({'status': 'error', 'message': 'Invalid JSON'}, status=400)
    except Exception as e:
        logger.error(f"❌ Webhook processing error: {e}", exc_info=True)
        return web.json_response({'status': 'error', 'message': str(e)}, status=500)


async def health_check_server():
    """HTTP сервер"""
    async def health(request):
        status = {
            'status': 'healthy',
            'workerId': WORKER_ID,
            'activeConferences': len(active_conferences),
            'activeSessions': sum(len(c.activeSessions) for c in active_conferences.values()),
            'totalParticipants': sum(len(c.participants) for c in active_conferences.values()),
            's3Configured': s3_client is not None,
            'webhookConfigured': bool(WEBHOOK_URL)
        }
        logger.debug(f"Health check: {status}")
        return web.json_response(status)

    app = web.Application()
    app.router.add_get('/health', health)
    app.router.add_post('/events', events_webhook_handler)

    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, '0.0.0.0', 8080)
    await site.start()
    logger.info("🌐 HTTP server started on :8080 (/health, /events)")
    logger.info("📝 Webhook endpoint: http://recorder:8080/events")


async def monitor_conferences():
    """Мониторинг"""
    logger.info(f"👀 [{WORKER_ID}] Monitor started")

    while True:
        try:
            for room_name in list(active_conferences.keys()):
                if not await claim_room(room_name):
                    logger.warning(f"Lost ownership of {room_name}")
                    del active_conferences[room_name]

            if active_conferences:
                logger.debug(f"Active conferences: {list(active_conferences.keys())}")

            await asyncio.sleep(CHECK_INTERVAL)

        except Exception as e:
            logger.error(f"Monitor error: {e}", exc_info=True)
            await asyncio.sleep(CHECK_INTERVAL)


async def main():
    global redis_client

    logger.info("=" * 60)
    logger.info("🚀 JITSI RECORDER STARTING")
    logger.info("=" * 60)

    try:
        redis_client = await redis.from_url(
            f"redis://{REDIS_HOST}:{REDIS_PORT}",
            decode_responses=True
        )
        logger.info(f"✅ Redis connected: {REDIS_HOST}:{REDIS_PORT}")
    except Exception as e:
        logger.error(f"❌ Redis connection failed: {e}")
        logger.error("Cannot continue without Redis")
        return

    logger.info(f"🔧 Worker {WORKER_ID} initialized")
    logger.info("=" * 60)

    asyncio.create_task(health_check_server())
    await monitor_conferences()


if __name__ == '__main__':
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("👋 Shutting down gracefully...")
    except Exception as e:
        logger.error(f"💥 Fatal error: {e}", exc_info=True)
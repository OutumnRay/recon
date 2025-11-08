#!/usr/bin/env python3
"""
Jitsi Auto Recorder - автоматическая запись через Kurento API
Слушает события от Prosody и управляет записью через Kurento Recorder API
"""

import os
import asyncio
import json
import logging
from datetime import datetime
from aiohttp import web, ClientSession
import redis.asyncio as redis

# Настройка логирования
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Конфигурация
KURENTO_API_URL = os.getenv('KURENTO_API_URL', 'http://kurento-recorder:8080')
REDIS_HOST = os.getenv('REDIS_HOST', 'redis')
REDIS_PORT = int(os.getenv('REDIS_PORT', '6379'))
WORKER_ID = os.getenv('HOSTNAME', 'auto-recorder')
AUTO_RECORD = os.getenv('AUTO_RECORD', 'true').lower() == 'true'

logger.info(f"🔧 Auto Recorder Configuration:")
logger.info(f"  KURENTO_API_URL: {KURENTO_API_URL}")
logger.info(f"  REDIS_HOST: {REDIS_HOST}:{REDIS_PORT}")
logger.info(f"  WORKER_ID: {WORKER_ID}")
logger.info(f"  AUTO_RECORD: {AUTO_RECORD}")

redis_client = None
active_sessions = {}  # {endpoint_id: {user, room, started_at}}
active_conferences = {}  # {room_name: {started_at, participants: []}}


class KurentoRecorderClient:
    """Клиент для Kurento Recorder API"""

    def __init__(self, api_url: str):
        self.api_url = api_url
        self.session = None

    async def get_session(self):
        if self.session is None or self.session.closed:
            self.session = ClientSession()
        return self.session

    async def start_recording(self, user_id: str, room_name: str):
        """Начать запись участника"""
        try:
            session = await self.get_session()
            url = f"{self.api_url}/record/start"
            params = {"user": user_id, "room": room_name}

            logger.info(f"🎙️  Starting recording: {user_id} in room {room_name}")
            logger.debug(f"   API URL: {url}?user={user_id}&room={room_name}")

            async with session.post(url, params=params, timeout=30) as resp:
                if resp.status == 200:
                    data = await resp.json()
                    logger.info(f"✅ Recording started: {user_id} -> {data.get('filepath', 'unknown')}")
                    return data
                elif resp.status == 400:
                    error = await resp.json()
                    logger.warning(f"⚠️  Recording already active for {user_id}: {error.get('detail')}")
                    return None
                else:
                    text = await resp.text()
                    logger.error(f"❌ Failed to start recording for {user_id}: HTTP {resp.status} - {text}")
                    return None

        except asyncio.TimeoutError:
            logger.error(f"❌ Timeout starting recording for {user_id}")
            return None
        except Exception as e:
            logger.error(f"❌ Error starting recording for {user_id}: {e}", exc_info=True)
            return None

    async def stop_recording(self, user_id: str, room_name: str):
        """Остановить запись участника"""
        try:
            session = await self.get_session()
            url = f"{self.api_url}/record/stop"
            params = {"user": user_id, "room": room_name}

            logger.info(f"⏹️  Stopping recording: {user_id} in room {room_name}")

            async with session.post(url, params=params, timeout=30) as resp:
                if resp.status == 200:
                    data = await resp.json()
                    logger.info(f"✅ Recording stopped: {user_id} -> {data.get('filepath', 'unknown')}")
                    return data
                elif resp.status == 404:
                    logger.warning(f"⚠️  No active recording for {user_id}")
                    return None
                else:
                    text = await resp.text()
                    logger.error(f"❌ Failed to stop recording for {user_id}: HTTP {resp.status} - {text}")
                    return None

        except asyncio.TimeoutError:
            logger.error(f"❌ Timeout stopping recording for {user_id}")
            return None
        except Exception as e:
            logger.error(f"❌ Error stopping recording for {user_id}: {e}", exc_info=True)
            return None

    async def stop_all_recordings(self):
        """Остановить все активные записи"""
        try:
            session = await self.get_session()
            url = f"{self.api_url}/record/stop-all"

            logger.info(f"🛑 Stopping all recordings")

            async with session.delete(url, timeout=30) as resp:
                if resp.status == 200:
                    data = await resp.json()
                    logger.info(f"✅ All recordings stopped: {data.get('count', 0)} sessions")
                    return data
                else:
                    text = await resp.text()
                    logger.error(f"❌ Failed to stop all recordings: HTTP {resp.status} - {text}")
                    return None

        except Exception as e:
            logger.error(f"❌ Error stopping all recordings: {e}", exc_info=True)
            return None

    async def get_status(self, user_id: str):
        """Получить статус записи"""
        try:
            session = await self.get_session()
            url = f"{self.api_url}/record/status"
            params = {"user": user_id}

            async with session.get(url, params=params, timeout=10) as resp:
                if resp.status == 200:
                    return await resp.json()
                else:
                    return None

        except Exception as e:
            logger.debug(f"Error getting status for {user_id}: {e}")
            return None

    async def close(self):
        if self.session and not self.session.closed:
            await self.session.close()


# Global client
recorder_client = KurentoRecorderClient(KURENTO_API_URL)


async def handle_participant_joined(room_name: str, endpoint_id: str, participant_id: str,
                                      participant_name: str, display_name: str):
    """Обработка присоединения участника"""
    logger.info(f"📥 PARTICIPANT JOINED: {display_name} (ID: {participant_id}, endpoint: {endpoint_id}) in room {room_name}")

    # Пропускаем системные компоненты
    if endpoint_id == 'focus' or 'focus' in participant_name.lower() or 'focus@' in participant_id:
        logger.debug(f"Skipping system component: {endpoint_id}")
        return

    # Создаем или обновляем конференцию
    if room_name not in active_conferences:
        active_conferences[room_name] = {
            'started_at': datetime.now().isoformat(),
            'participants': []
        }
        logger.info(f"📹 Conference started: {room_name}")

    # Добавляем участника в список
    if participant_id not in active_conferences[room_name]['participants']:
        active_conferences[room_name]['participants'].append(participant_id)

    # Если автозапись включена - начинаем запись
    if AUTO_RECORD:
        # Используем participant_id как user_id для Kurento
        result = await recorder_client.start_recording(participant_id, room_name)

        if result:
            # Сохраняем информацию о сессии
            active_sessions[endpoint_id] = {
                'user': participant_id,
                'room': room_name,
                'display_name': display_name,
                'started_at': datetime.now().isoformat()
            }
            logger.info(f"✅ Auto-recording started for {display_name}")
        else:
            logger.warning(f"⚠️  Failed to start auto-recording for {display_name}")
    else:
        logger.debug(f"Auto-recording disabled, skipping {display_name}")


async def handle_participant_left(room_name: str, endpoint_id: str):
    """Обработка выхода участника"""
    logger.info(f"📤 PARTICIPANT LEFT: endpoint {endpoint_id} from room {room_name}")

    # Проверяем есть ли активная сессия
    if endpoint_id not in active_sessions:
        logger.debug(f"No active session for endpoint {endpoint_id}")
        return

    session_info = active_sessions[endpoint_id]
    user_id = session_info['user']
    display_name = session_info['display_name']

    # Останавливаем запись
    if AUTO_RECORD:
        result = await recorder_client.stop_recording(user_id, room_name)
        if result:
            duration = (datetime.now() - datetime.fromisoformat(session_info['started_at'])).total_seconds()
            logger.info(f"✅ Recording stopped for {display_name} (duration: {duration:.1f}s)")
        else:
            logger.warning(f"⚠️  Failed to stop recording for {display_name}")

    # Удаляем сессию
    del active_sessions[endpoint_id]

    # Убираем участника из конференции
    if room_name in active_conferences:
        participants = active_conferences[room_name]['participants']
        if user_id in participants:
            participants.remove(user_id)


async def handle_conference_ended(room_name: str):
    """Обработка завершения конференции"""
    logger.info(f"🏁 CONFERENCE ENDED: {room_name}")

    # Останавливаем все активные сессии для этой комнаты
    endpoints_to_stop = [
        endpoint_id for endpoint_id, info in active_sessions.items()
        if info['room'] == room_name
    ]

    for endpoint_id in endpoints_to_stop:
        await handle_participant_left(room_name, endpoint_id)

    # Логируем итоги конференции
    if room_name in active_conferences:
        conf = active_conferences[room_name]
        duration = (datetime.now() - datetime.fromisoformat(conf['started_at'])).total_seconds()
        logger.info(f"✅ Conference {room_name} completed:")
        logger.info(f"   Duration: {duration:.1f}s")
        logger.info(f"   Participants: {len(conf['participants'])}")

        # Удаляем конференцию
        del active_conferences[room_name]

    logger.info(f"📁 Recordings for room '{room_name}' uploaded to MinIO bucket 'jitsi-recordings'")


async def events_webhook_handler(request):
    """Обработчик webhook событий от Prosody"""
    try:
        logger.debug(f"🌐 Incoming webhook from {request.remote}")

        data = await request.json()
        logger.debug(f"📨 Webhook data: {json.dumps(data, indent=2)}")

        event_type = data.get('eventType')
        room_name = data.get('roomName')
        endpoint_id = data.get('endpointId')

        if not event_type or not room_name:
            logger.warning(f"⚠️  Invalid webhook data: missing eventType or roomName")
            return web.json_response({'status': 'error', 'message': 'Missing required fields'}, status=400)

        logger.info(f"🔔 Event: {event_type}, Room: {room_name}, Endpoint: {endpoint_id}")

        if event_type == 'participantJoined':
            await handle_participant_joined(
                room_name,
                endpoint_id,
                data.get('participantId', endpoint_id),
                data.get('participantName', ''),
                data.get('displayName', 'Unknown')
            )

        elif event_type == 'participantLeft':
            await handle_participant_left(room_name, endpoint_id)

        elif event_type == 'conferenceEnded':
            await handle_conference_ended(room_name)

        else:
            logger.warning(f"⚠️  Unknown event type: {event_type}")

        return web.json_response({'status': 'ok'})

    except json.JSONDecodeError as e:
        logger.error(f"❌ Invalid JSON: {e}")
        return web.json_response({'status': 'error', 'message': 'Invalid JSON'}, status=400)
    except Exception as e:
        logger.error(f"❌ Webhook error: {e}", exc_info=True)
        return web.json_response({'status': 'error', 'message': str(e)}, status=500)


async def health_handler(request):
    """Health check endpoint"""
    status = {
        'status': 'healthy',
        'worker_id': WORKER_ID,
        'auto_record': AUTO_RECORD,
        'active_conferences': len(active_conferences),
        'active_sessions': len(active_sessions),
        'kurento_api': KURENTO_API_URL
    }

    # Проверяем доступность Kurento API
    try:
        session = await recorder_client.get_session()
        async with session.get(f"{KURENTO_API_URL}/health", timeout=5) as resp:
            if resp.status == 200:
                kurento_status = await resp.json()
                status['kurento_status'] = kurento_status
            else:
                status['kurento_status'] = 'unreachable'
    except Exception as e:
        status['kurento_status'] = f'error: {str(e)}'

    logger.debug(f"Health check: {status}")
    return web.json_response(status)


async def stats_handler(request):
    """Статистика активных записей"""
    stats = {
        'active_conferences': [],
        'active_sessions': []
    }

    for room_name, conf in active_conferences.items():
        stats['active_conferences'].append({
            'room': room_name,
            'started_at': conf['started_at'],
            'participants': len(conf['participants'])
        })

    for endpoint_id, session in active_sessions.items():
        stats['active_sessions'].append({
            'endpoint_id': endpoint_id,
            'user': session['user'],
            'room': session['room'],
            'display_name': session['display_name'],
            'started_at': session['started_at']
        })

    return web.json_response(stats)


async def http_server():
    """HTTP сервер для webhook и health checks"""
    app = web.Application()
    app.router.add_get('/health', health_handler)
    app.router.add_get('/stats', stats_handler)
    app.router.add_post('/events', events_webhook_handler)

    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, '0.0.0.0', 8080)
    await site.start()

    logger.info("=" * 60)
    logger.info("🌐 Auto Recorder HTTP Server Started")
    logger.info("=" * 60)
    logger.info(f"  Health: http://0.0.0.0:8080/health")
    logger.info(f"  Stats:  http://0.0.0.0:8080/stats")
    logger.info(f"  Events: http://0.0.0.0:8080/events (webhook endpoint)")
    logger.info("=" * 60)


async def cleanup_stale_sessions():
    """Периодическая очистка зависших сессий"""
    while True:
        try:
            await asyncio.sleep(60)  # Проверка каждую минуту

            # Проверяем статус всех активных сессий через Kurento API
            for endpoint_id, session in list(active_sessions.items()):
                user_id = session['user']
                status = await recorder_client.get_status(user_id)

                # Если запись не найдена в Kurento, удаляем локальную сессию
                if status is None or status.get('status') == 'not_found':
                    logger.warning(f"⚠️  Stale session detected: {user_id} - removing")
                    del active_sessions[endpoint_id]

        except Exception as e:
            logger.error(f"Error in cleanup task: {e}", exc_info=True)


async def main():
    global redis_client

    logger.info("=" * 60)
    logger.info("🚀 JITSI AUTO RECORDER STARTING")
    logger.info("=" * 60)

    # Подключаемся к Redis (опционально)
    try:
        redis_client = await redis.from_url(
            f"redis://{REDIS_HOST}:{REDIS_PORT}",
            decode_responses=True
        )
        await redis_client.ping()
        logger.info(f"✅ Redis connected: {REDIS_HOST}:{REDIS_PORT}")
    except Exception as e:
        logger.warning(f"⚠️  Redis connection failed: {e}")
        logger.warning("   Continuing without Redis (single-worker mode)")

    # Запускаем HTTP сервер
    asyncio.create_task(http_server())

    # Запускаем cleanup task
    asyncio.create_task(cleanup_stale_sessions())

    logger.info("✅ Auto Recorder ready")
    logger.info("=" * 60)

    # Держим процесс живым
    while True:
        await asyncio.sleep(3600)


if __name__ == '__main__':
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("👋 Shutting down gracefully...")
        asyncio.run(recorder_client.close())
    except Exception as e:
        logger.error(f"💥 Fatal error: {e}", exc_info=True)

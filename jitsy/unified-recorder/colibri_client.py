# --- unified-recorder/colibri_client.py ---

import logging
import json
from typing import Optional, Dict, Any
from aiohttp import ClientSession

logger = logging.getLogger(__name__)

class ColibriClient:
    """Client for Jitsi Videobridge Colibri REST API"""
    def __init__(self, jvb_host: str, jvb_port: int = 8080):
        self.base_url = f"http://{jvb_host}:{jvb_port}/colibri/conferences"
        self.session: Optional[ClientSession] = None
        # --- ДОБАВЛЕНА СТРОКА ЛОГИРОВАНИЯ ---
        logger.info(f"ColibriClient initialized for JVB at {self.base_url}")

    async def get_session(self) -> ClientSession:
        if self.session is None or self.session.closed:
            self.session = ClientSession()
        return self.session

    async def close(self):
        if self.session and not self.session.closed:
            await self.session.close()

    async def create_conference(self) -> Dict[str, Any]:
        session = await self.get_session()
        url = self.base_url
        payload = {}
        logger.debug(f"Creating Colibri conference: POST {url}")
        async with session.post(url, json=payload, timeout=10) as resp:
            resp_text = await resp.text()
            if resp.status != 201:
                raise Exception(f"Failed to create conference: HTTP {resp.status} - {resp_text}")
            data = json.loads(resp_text)
            conference_id = data.get("id")
            logger.info(f"✅ Created Colibri conference: {conference_id}")
            return data

    # ... (остальная часть файла без изменений)
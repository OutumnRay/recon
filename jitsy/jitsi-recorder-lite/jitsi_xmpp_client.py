"""
Минимальный XMPP клиент для Jitsi Meet.
Подключается к xmpp-websocket, проходит SASL ANONYMOUS,
присоединяется к MUC и запрашивает конференцию у focus.meet.jitsi.

Это первый шаг к полноценному подключению рекордера к реальному
https://meet.recontext.online/<room>: получаем реальные Jingle SDP от Jicofo
без запуска браузера.
"""
import asyncio
import contextlib
import logging
import os
import ssl
import uuid
from urllib.parse import urlparse, quote
from xml.etree import ElementTree as ET

import websockets


logger = logging.getLogger(__name__)

NAMESPACES = {
    "stream": "http://etherx.jabber.org/streams",
    "framing": "urn:ietf:params:xml:ns:xmpp-framing",
    "sasl": "urn:ietf:params:xml:ns:xmpp-sasl",
    "bind": "urn:ietf:params:xml:ns:xmpp-bind",
    "session": "urn:ietf:params:xml:ns:xmpp-session",
    "client": "jabber:client",
    "ping": "urn:xmpp:ping",
    "muc": "http://jabber.org/protocol/muc",
    "focus": "http://jitsi.org/protocol/focus",
    "jingle": "urn:xmpp:jingle:1",
}


def strip_ns(tag: str) -> str:
    """Удаляет xmlns префиксы вида {namespace}tag."""
    if tag.startswith("{"):
        return tag.split("}", 1)[1]
    return tag


class JitsiXmppClient:
    """Минимальный XMPP клиент, работающий поверх xmpp-websocket."""

    def __init__(
        self,
        room: str,
        jitsi_url: str = "https://meet.recontext.online",
        xmpp_domain: str = "meet.jitsi",
        muc_domain: str = "muc.meet.jitsi",
        focus_jid: str = "focus.meet.jitsi",
        nickname: str = "recorder-bot",
    ) -> None:
        parsed = urlparse(jitsi_url)
        base_path = parsed.path.rstrip("/")
        ws_path = f"{base_path}/xmpp-websocket" if base_path else "/xmpp-websocket"
        query = f"?room={quote(room)}"
        self.ws_url = f"wss://{parsed.netloc}{ws_path}{query}"
        self.origin = f"{parsed.scheme}://{parsed.netloc}"

        self.room = room
        self.jitsi_url = jitsi_url
        self.domain = xmpp_domain
        self.muc_domain = muc_domain
        self.focus_jid = focus_jid
        self.nickname = nickname
        self.room_jid = f"{room}@{muc_domain}"
        self.full_room_jid = f"{self.room_jid}/{nickname}"

        self.ssl_context = self._build_ssl_context()
        self.ws = None
        self.loop = asyncio.get_event_loop()
        self.pending_iq: dict[str, asyncio.Future] = {}
        self.jid = None
        self.recv_task = None
        self.on_jingle = None  # callback

    async def connect(self):
        logger.info("🔌 Connecting to xmpp-websocket: %s", self.ws_url)
        self.ws = await websockets.connect(
            self.ws_url,
            subprotocols=["xmpp"],
            origin=self.origin,
            ssl=self.ssl_context,
        )

        await self._open_stream(initial=True)
        await self._authenticate()
        await self._open_stream(initial=False)
        await self._bind_resource()
        await self._establish_session()
        await self._send_initial_presence()
        await self._join_muc()

        self.recv_task = asyncio.create_task(self._recv_loop())
        logger.info("✅ XMPP connection ready as %s", self.jid)

    async def request_conference(self):
        """Отправляет IQ conference к focus для старта сессии."""
        if not self.ws:
            raise RuntimeError("Not connected")

        props = (
            "<property name='disableRtx' value='false'/>"
            "<property name='channelLastN' value='-1'/>"
            "<property name='startAudioOnly' value='false'/>"
            "<property name='openSctp' value='true'/>"
        )
        payload = (
            f"<conference xmlns='{NAMESPACES['focus']}' "
            f"room='{self.room_jid}' machine-uid='{uuid.uuid4().hex}'>"
            f"{props}"
            "</conference>"
        )
        response = await self._send_iq(payload, iq_type="set", to=self.focus_jid)
        logger.info(
            "🎯 Focus response: type=%s, attrs=%s",
            response.get("type"),
            response.attrib,
        )

    async def disconnect(self):
        if self.recv_task:
            self.recv_task.cancel()
            with contextlib.suppress(asyncio.CancelledError):
                await self.recv_task

        if self.ws:
            await self.ws.close()

    async def _recv_loop(self):
        try:
            while True:
                elem = await self._next_stanza()
                if elem is None:
                    return
                await self._handle_stanza(elem)
        except websockets.ConnectionClosed:
            logger.info("🔌 xmpp-websocket closed")

    async def _handle_stanza(self, elem: ET.Element):
        tag = strip_ns(elem.tag)
        if tag == "iq":
            iq_id = elem.get("id")
            iq_type = elem.get("type")
            if iq_id and iq_id in self.pending_iq:
                fut = self.pending_iq.pop(iq_id)
                if not fut.done():
                    fut.set_result(elem)
                return

            if iq_type == "get":
                ping = elem.find(f".//{{{NAMESPACES['ping']}}}ping")
                if ping is not None:
                    await self._send_raw(
                        f"<iq to='{elem.get('from')}' id='{elem.get('id')}' "
                        f"type='result' xmlns='{NAMESPACES['client']}'/>"
                    )
                    return

            jingle = elem.find(f".//{{{NAMESPACES['jingle']}}}jingle")
            if jingle is not None:
                logger.info(
                    "📨 Received Jingle action=%s sid=%s",
                    jingle.get("action"),
                    jingle.get("sid"),
                )
                if self.on_jingle:
                    await self.on_jingle(elem, jingle)
                return

            logger.debug("Unhandled IQ: %s", ET.tostring(elem))
        elif tag == "presence":
            from_jid = elem.get("from")
            nick = from_jid.split("/", 1)[1] if from_jid and "/" in from_jid else from_jid
            logger.debug("👥 Presence from %s (type=%s)", nick, elem.get("type"))
        elif tag == "message":
            logger.debug("💬 Message: %s", ET.tostring(elem, encoding="unicode"))
        else:
            logger.debug("↩️  %s", ET.tostring(elem, encoding="unicode"))

    async def _next_stanza(self) -> ET.Element | None:
        while True:
            frame = await self.ws.recv()
            if frame is None:
                return None
            if isinstance(frame, bytes):
                frame = frame.decode()
            frame = frame.strip()
            if not frame:
                continue
            for elem in self._split_frame(frame):
                tag = strip_ns(elem.tag)
                if tag == "open" or tag == "close":
                    continue
                return elem
            logger.debug("No full stanza extracted from frame: %s", frame)

    @staticmethod
    def _split_frame(frame: str) -> list[ET.Element]:
        """Разбивает WebSocket payload на отдельные XMPP stanzas."""
        data = frame.strip()
        if not data:
            return []
        wrapped = f"<root>{data}</root>"
        try:
            root = ET.fromstring(wrapped)
            return list(root)
        except ET.ParseError as exc:
            logger.error("❌ Failed to parse XMPP chunk: %s", data)
            raise exc

    async def _open_stream(self, initial: bool):
        open_tag = (
            f"<open xmlns='{NAMESPACES['framing']}' to='{self.domain}' version='1.0'/>"
        )
        await self._send_raw(open_tag)

        await self._expect_tag("{urn:ietf:params:xml:ns:xmpp-framing}open")
        features = await self._expect_tag(
            "{http://etherx.jabber.org/streams}features"
        )
        logger.debug(
            "%s stream features: %s",
            "Initial" if initial else "Post-auth",
            ET.tostring(features, encoding="unicode"),
        )

    async def _authenticate(self):
        await self._send_raw(
            f"<auth xmlns='{NAMESPACES['sasl']}' mechanism='ANONYMOUS'/>"
        )
        await self._expect_tag("{urn:ietf:params:xml:ns:xmpp-sasl}success")
        logger.info("🔐 SASL ANONYMOUS success")

    async def _bind_resource(self):
        iq_id = self._new_id("bind")
        stanza = (
            f"<iq type='set' id='{iq_id}' xmlns='{NAMESPACES['client']}'>"
            f"<bind xmlns='{NAMESPACES['bind']}'/>"
            "</iq>"
        )
        await self._send_raw(stanza)
        result = await self._expect_iq(iq_id)
        jid_elem = result.find(f".//{{{NAMESPACES['bind']}}}jid")
        if not jid_elem or not jid_elem.text:
            raise RuntimeError("Failed to bind resource")
        self.jid = jid_elem.text
        logger.info("📛 Bound JID: %s", self.jid)

    async def _establish_session(self):
        iq_id = self._new_id("sess")
        stanza = (
            f"<iq type='set' id='{iq_id}' xmlns='{NAMESPACES['client']}'>"
            f"<session xmlns='{NAMESPACES['session']}'/>"
            "</iq>"
        )
        await self._send_raw(stanza)
        await self._expect_iq(iq_id)

    async def _send_initial_presence(self):
        presence = (
            f"<presence xmlns='{NAMESPACES['client']}' xml:lang='en'>"
            "<c xmlns='http://jabber.org/protocol/caps' node='jitsi' ver='1.0'/>"
            "</presence>"
        )
        await self._send_raw(presence)

    async def _join_muc(self):
        presence = (
            f"<presence to='{self.full_room_jid}' xmlns='{NAMESPACES['client']}'>"
            f"<x xmlns='{NAMESPACES['muc']}'/>"
            "</presence>"
        )
        await self._send_raw(presence)
        logger.info("🚪 Joined MUC %s as %s", self.room_jid, self.nickname)

    async def _expect_tag(self, expected_tag: str) -> ET.Element:
        while True:
            elem = await self._next_stanza()
            if elem is None:
                raise ConnectionError("xmpp stream closed")
            if elem.tag == expected_tag:
                return elem
            logger.debug("Skipping unexpected stanza during handshake: %s", elem.tag)

    async def _expect_iq(self, iq_id: str) -> ET.Element:
        while True:
            elem = await self._next_stanza()
            if elem is None:
                raise ConnectionError("xmpp stream closed")
            if strip_ns(elem.tag) == "iq" and elem.get("id") == iq_id:
                return elem
            logger.debug("Waiting for IQ %s, saw %s", iq_id, elem.tag)

    async def _send_iq(self, inner_xml: str, iq_type: str = "get", to: str | None = None):
        iq_id = self._new_id("iq")
        attrs = [f"id='{iq_id}'", f"type='{iq_type}'"]
        if to:
            attrs.append(f"to='{to}'")
        stanza = (
            f"<iq {' '.join(attrs)} xmlns='{NAMESPACES['client']}'>"
            f"{inner_xml}"
            "</iq>"
        )
        fut = self.loop.create_future()
        self.pending_iq[iq_id] = fut
        await self._send_raw(stanza)
        return await fut

    async def _send_raw(self, data: str):
        if not self.ws:
            raise RuntimeError("WebSocket not connected")
        await self.ws.send(data)

    def _new_id(self, prefix: str) -> str:
        return f"{prefix}-{uuid.uuid4().hex[:8]}"

    @staticmethod
    def _build_ssl_context() -> ssl.SSLContext:
        # Отключаем проверку сертификата по требованию
        context = ssl.create_default_context()
        context.check_hostname = False
        context.verify_mode = ssl.CERT_NONE
        logging.getLogger(__name__).warning("⚠️ XMPP TLS verification disabled")
        return context


async def main():
    logging.basicConfig(level=logging.INFO)
    room = os.getenv("JITSI_ROOM", "testmeet")
    client = JitsiXmppClient(
        room=room,
        jitsi_url=os.getenv("JITSI_URL", "https://meet.recontext.online"),
        xmpp_domain=os.getenv("XMPP_DOMAIN", "meet.jitsi"),
        muc_domain=os.getenv("MUC_DOMAIN", "muc.meet.jitsi"),
        focus_jid=os.getenv("FOCUS_JID", "focus.meet.jitsi"),
        nickname=os.getenv("BOT_NICKNAME", "recorder"),
    )

    await client.connect()
    await client.request_conference()

    try:
        while True:
            await asyncio.sleep(1)
    except KeyboardInterrupt:
        logger.info("⏹️  Interrupt received")
    finally:
        await client.disconnect()


if __name__ == "__main__":
    asyncio.run(main())

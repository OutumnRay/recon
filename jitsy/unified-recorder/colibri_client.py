"""
Colibri REST API client for JVB integration
Handles SDP negotiation, ICE candidates, and media forwarding
"""

import logging
import json
from typing import Optional, Dict, List, Any
from aiohttp import ClientSession

logger = logging.getLogger(__name__)


class ColibriClient:
    """Client for Jitsi Videobridge Colibri REST API"""

    def __init__(self, jvb_host: str, jvb_port: int = 8080):
        self.base_url = f"http://{jvb_host}:{jvb_port}/colibri"
        self.session: Optional[ClientSession] = None

    async def get_session(self) -> ClientSession:
        """Get or create HTTP session"""
        if self.session is None or self.session.closed:
            self.session = ClientSession()
        return self.session

    async def close(self):
        """Close HTTP session"""
        if self.session and not self.session.closed:
            await self.session.close()

    async def create_conference(self) -> Dict[str, Any]:
        """
        Create a new conference in JVB

        Returns:
            Conference data with ID
        """
        session = await self.get_session()
        url = f"{self.base_url}/conferences"

        payload = {
            "contents": [
                {
                    "name": "audio",
                    "channels": []
                },
                {
                    "name": "video",
                    "channels": []
                }
            ]
        }

        logger.debug(f"Creating Colibri conference: POST {url}")

        async with session.post(url, json=payload, timeout=10) as resp:
            if resp.status != 200:
                text = await resp.text()
                raise Exception(f"Failed to create conference: HTTP {resp.status} - {text}")

            data = await resp.json()
            conference_id = data.get("id")

            logger.info(f"✅ Created Colibri conference: {conference_id}")
            return data

    async def create_channel(
        self,
        conference_id: str,
        endpoint_id: str,
        sdp_offer: Optional[str] = None,
        include_audio: bool = True,
        include_video: bool = True
    ) -> Dict[str, Any]:
        """
        Create a channel (participant endpoint) in the conference

        Args:
            conference_id: Conference ID from create_conference
            endpoint_id: Unique participant identifier
            sdp_offer: Optional SDP offer from Kurento
            include_audio: Include audio channel
            include_video: Include video channel

        Returns:
            Channel data with SDP answer and ICE candidates
        """
        session = await self.get_session()
        url = f"{self.base_url}/conferences/{conference_id}"

        # Build channel bundle
        channel_bundle = {
            "id": f"bundle-{endpoint_id}",
            "transport": {
                "xmlns": "urn:xmpp:jingle:transports:ice-udp:1",
                "rtcp-mux": True,
                "ice-controlling": True
            }
        }

        # Add fingerprint if we have SDP offer
        if sdp_offer:
            fingerprint = self._extract_fingerprint_from_sdp(sdp_offer)
            if fingerprint:
                channel_bundle["transport"]["fingerprints"] = [fingerprint]

        contents = []

        # Audio channel
        if include_audio:
            audio_channel = {
                "id": f"{endpoint_id}-audio",
                "endpoint": endpoint_id,
                "initiator": False,
                "direction": "sendrecv",
                "channel-bundle-id": channel_bundle["id"],
                "rtp-level-relay-type": "translator",
                "sources": [],
                "payload-types": [
                    {
                        "id": 111,
                        "name": "opus",
                        "clockrate": 48000,
                        "channels": 2,
                        "parameters": {
                            "minptime": "10",
                            "useinbandfec": "1"
                        }
                    }
                ]
            }
            contents.append({
                "name": "audio",
                "channels": [audio_channel]
            })

        # Video channel
        if include_video:
            video_channel = {
                "id": f"{endpoint_id}-video",
                "endpoint": endpoint_id,
                "initiator": False,
                "direction": "sendrecv",
                "channel-bundle-id": channel_bundle["id"],
                "rtp-level-relay-type": "translator",
                "sources": [],
                "payload-types": [
                    {
                        "id": 96,
                        "name": "VP8",
                        "clockrate": 90000,
                        "parameters": {}
                    },
                    {
                        "id": 100,
                        "name": "VP9",
                        "clockrate": 90000,
                        "parameters": {}
                    }
                ]
            }
            contents.append({
                "name": "video",
                "channels": [video_channel]
            })

        payload = {
            "channel-bundles": [channel_bundle],
            "contents": contents
        }

        logger.debug(f"Creating channel for endpoint {endpoint_id}: PATCH {url}")

        async with session.patch(url, json=payload, timeout=10) as resp:
            if resp.status != 200:
                text = await resp.text()
                raise Exception(f"Failed to create channel: HTTP {resp.status} - {text}")

            data = await resp.json()
            logger.info(f"✅ Created Colibri channel for {endpoint_id}")
            return data

    async def update_channel_transport(
        self,
        conference_id: str,
        endpoint_id: str,
        ice_candidates: List[Dict[str, Any]],
        dtls_fingerprint: Optional[Dict[str, str]] = None
    ) -> Dict[str, Any]:
        """
        Update channel transport info (ICE candidates, DTLS fingerprint)

        Args:
            conference_id: Conference ID
            endpoint_id: Endpoint ID
            ice_candidates: List of ICE candidates
            dtls_fingerprint: DTLS fingerprint dict

        Returns:
            Updated channel data
        """
        session = await self.get_session()
        url = f"{self.base_url}/conferences/{conference_id}"

        transport = {
            "xmlns": "urn:xmpp:jingle:transports:ice-udp:1",
            "rtcp-mux": True,
            "ice-controlling": False,
            "candidates": ice_candidates
        }

        if dtls_fingerprint:
            transport["fingerprints"] = [dtls_fingerprint]

        payload = {
            "channel-bundles": [
                {
                    "id": f"bundle-{endpoint_id}",
                    "transport": transport
                }
            ]
        }

        logger.debug(f"Updating transport for {endpoint_id}: PATCH {url}")

        async with session.patch(url, json=payload, timeout=10) as resp:
            if resp.status != 200:
                text = await resp.text()
                raise Exception(f"Failed to update transport: HTTP {resp.status} - {text}")

            data = await resp.json()
            logger.info(f"✅ Updated transport for {endpoint_id}")
            return data

    async def expire_conference(self, conference_id: str):
        """Expire (delete) a conference"""
        session = await self.get_session()
        url = f"{self.base_url}/conferences/{conference_id}"

        logger.debug(f"Expiring conference: DELETE {url}")

        try:
            async with session.delete(url, timeout=5) as resp:
                if resp.status == 200:
                    logger.info(f"✅ Expired conference {conference_id}")
                else:
                    logger.warning(f"Failed to expire conference: HTTP {resp.status}")
        except Exception as e:
            logger.debug(f"Error expiring conference: {e}")

    @staticmethod
    def _extract_fingerprint_from_sdp(sdp: str) -> Optional[Dict[str, str]]:
        """Extract DTLS fingerprint from SDP"""
        for line in sdp.split('\n'):
            line = line.strip()
            if line.startswith('a=fingerprint:'):
                parts = line.replace('a=fingerprint:', '').split(' ', 1)
                if len(parts) == 2:
                    return {
                        "hash": parts[0],  # e.g., "sha-256"
                        "fingerprint": parts[1],  # e.g., "AA:BB:CC:..."
                        "setup": "actpass"
                    }
        return None

    @staticmethod
    def parse_ice_candidates_from_colibri(channel_bundle: Dict[str, Any]) -> List[str]:
        """Parse ICE candidates from Colibri response to SDP format"""
        candidates = []
        transport = channel_bundle.get("transport", {})

        for candidate in transport.get("candidates", []):
            # Convert Colibri candidate format to SDP format
            # Colibri format: {"component": 1, "foundation": "...", "ip": "...", ...}
            # SDP format: "candidate:foundation component protocol priority ip port typ type ..."

            foundation = candidate.get("foundation", "1")
            component = candidate.get("component", 1)
            protocol = candidate.get("protocol", "udp")
            priority = candidate.get("priority", 2130706431)
            ip = candidate.get("ip", "")
            port = candidate.get("port", 0)
            typ = candidate.get("type", "host")

            sdp_candidate = f"candidate:{foundation} {component} {protocol} {priority} {ip} {port} typ {typ}"

            # Add related address for relay/srflx candidates
            rel_addr = candidate.get("rel-addr")
            rel_port = candidate.get("rel-port")
            if rel_addr and rel_port:
                sdp_candidate += f" raddr {rel_addr} rport {rel_port}"

            candidates.append(sdp_candidate)

        return candidates

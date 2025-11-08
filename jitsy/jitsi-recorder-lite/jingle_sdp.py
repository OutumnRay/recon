"""
Конвертер Jingle XML <-> SDP для Jitsi Meet
"""
import logging
from xml.etree import ElementTree as ET

logger = logging.getLogger(__name__)


class JingleToSDP:
    """Конвертирует Jingle XML в SDP формат"""

    @staticmethod
    def convert(jingle_elem):
        """
        Главная функция конвертации
        Принимает <jingle> element, возвращает SDP string
        """
        sdp_lines = [
            "v=0",
            "o=- 0 0 IN IP4 0.0.0.0",
            "s=-",
            "t=0 0"
        ]

        # Извлекаем contents
        contents = []
        for content in jingle_elem:
            if content.tag.endswith('content'):
                contents.append(content)

        # Group BUNDLE
        bundle_names = [c.get('name') for c in contents if c.get('name')]
        if bundle_names:
            sdp_lines.append(f"a=group:BUNDLE {' '.join(bundle_names)}")

        # Обрабатываем каждый content
        for content in contents:
            media_lines = JingleToSDP._parse_content(content)
            sdp_lines.extend(media_lines)

        return "\r\n".join(sdp_lines) + "\r\n"

    @staticmethod
    def _parse_content(content):
        """Парсит один content (audio/video/data)"""
        lines = []

        name = content.get('name')
        creator = content.get('creator', 'initiator')
        senders = content.get('senders', 'both')

        # Description
        desc = None
        for child in content:
            if child.tag.endswith('description'):
                desc = child
                break

        # Transport
        transport = None
        for child in content:
            if child.tag.endswith('transport'):
                transport = child
                break

        if not desc and not transport:
            logger.warning(f"Content {name} has no description or transport")
            return lines

        # Media type
        media_type = desc.get('media', 'application') if desc is not None else 'application'

        # Codecs
        payload_types = []
        if desc is not None:
            for pt in desc:
                if pt.tag.endswith('payload-type'):
                    payload_types.append(pt.get('id'))

        # Media line
        pt_list = ' '.join(payload_types) if payload_types else '0'
        lines.append(f"m={media_type} 9 UDP/TLS/RTP/SAVPF {pt_list}")
        lines.append("c=IN IP4 0.0.0.0")

        # Transport parameters
        if transport is not None:
            ufrag = transport.get('ufrag')
            pwd = transport.get('pwd')

            if ufrag:
                lines.append(f"a=ice-ufrag:{ufrag}")
            if pwd:
                lines.append(f"a=ice-pwd:{pwd}")

            # DTLS fingerprint
            for child in transport:
                if child.tag.endswith('fingerprint'):
                    hash_func = child.get('hash', 'sha-256')
                    setup = child.get('setup', 'actpass')
                    fp_value = child.text.strip() if child.text else ""
                    lines.append(f"a=fingerprint:{hash_func} {fp_value}")
                    lines.append(f"a=setup:{setup}")
                    break

            # ICE candidates
            for cand in transport:
                if cand.tag.endswith('candidate'):
                    cand_line = JingleToSDP._parse_candidate(cand)
                    if cand_line:
                        lines.append(cand_line)

        # Codecs details
        if desc is not None:
            for pt in desc:
                if pt.tag.endswith('payload-type'):
                    pt_lines = JingleToSDP._parse_payload_type(pt)
                    lines.extend(pt_lines)

            # RTCP-mux
            for child in desc:
                if child.tag.endswith('rtcp-mux'):
                    lines.append("a=rtcp-mux")
                    break

            # RTP header extensions
            for ext in desc:
                if ext.tag.endswith('rtp-hdrext'):
                    ext_id = ext.get('id')
                    uri = ext.get('uri')
                    if ext_id and uri:
                        lines.append(f"a=extmap:{ext_id} {uri}")

        # MID
        if name:
            lines.append(f"a=mid:{name}")

        # Direction
        direction_map = {
            'both': 'sendrecv',
            'initiator': 'recvonly',
            'responder': 'sendonly',
            'none': 'inactive'
        }
        lines.append(f"a={direction_map.get(senders, 'sendrecv')}")

        return lines

    @staticmethod
    def _parse_candidate(cand_elem):
        """Парсит ICE candidate в SDP формат"""
        foundation = cand_elem.get('foundation', '0')
        component = cand_elem.get('component', '1')
        protocol = cand_elem.get('protocol', 'udp').upper()
        priority = cand_elem.get('priority', '0')
        ip = cand_elem.get('ip')
        port = cand_elem.get('port')
        typ = cand_elem.get('type', 'host')

        if not ip or not port:
            return None

        line = f"a=candidate:{foundation} {component} {protocol} {priority} {ip} {port} typ {typ}"

        # raddr/rport для relay/srflx
        if typ in ['relay', 'srflx']:
            rel_addr = cand_elem.get('rel-addr')
            rel_port = cand_elem.get('rel-port')
            if rel_addr and rel_port:
                line += f" raddr {rel_addr} rport {rel_port}"

        return line

    @staticmethod
    def _parse_payload_type(pt_elem):
        """Парсит payload-type в SDP rtpmap/fmtp"""
        lines = []

        pt_id = pt_elem.get('id')
        name = pt_elem.get('name')
        clockrate = pt_elem.get('clockrate')
        channels = pt_elem.get('channels')

        if not pt_id or not name:
            return lines

        # rtpmap
        rtpmap = f"a=rtpmap:{pt_id} {name}/{clockrate}"
        if channels and channels != '1':
            rtpmap += f"/{channels}"
        lines.append(rtpmap)

        # fmtp parameters
        params = []
        for child in pt_elem:
            if child.tag.endswith('parameter'):
                param_name = child.get('name')
                param_value = child.get('value')
                if param_name and param_value:
                    params.append(f"{param_name}={param_value}")

        if params:
            lines.append(f"a=fmtp:{pt_id} {';'.join(params)}")

        # RTCP feedback
        for child in pt_elem:
            if child.tag.endswith('rtcp-fb'):
                fb_type = child.get('type')
                subtype = child.get('subtype')
                if fb_type:
                    fb_line = f"a=rtcp-fb:{pt_id} {fb_type}"
                    if subtype:
                        fb_line += f" {subtype}"
                    lines.append(fb_line)

        return lines


class SDPToJingle:
    """Конвертирует SDP в Jingle XML формат"""

    @staticmethod
    def convert(sdp_text, session_id, initiator):
        """
        Конвертирует SDP answer в Jingle session-accept
        """
        lines = sdp_text.strip().split('\n')

        # Парсим SDP
        media_sections = SDPToJingle._parse_sdp(lines)

        # Создаем Jingle XML
        jingle = ET.Element('jingle', {
            'xmlns': 'urn:xmpp:jingle:1',
            'action': 'session-accept',
            'sid': session_id,
            'initiator': initiator
        })

        # BUNDLE group
        bundle_group = ET.SubElement(jingle, 'group', {
            'xmlns': 'urn:xmpp:jingle:apps:grouping:0',
            'semantics': 'BUNDLE'
        })

        for media in media_sections:
            ET.SubElement(bundle_group, 'content', {'name': media['mid']})

        # Contents
        for media in media_sections:
            content = SDPToJingle._create_content(media)
            jingle.append(content)

        return jingle

    @staticmethod
    def _parse_sdp(lines):
        """Парсит SDP в структуру данных"""
        media_sections = []
        current_media = None

        for line in lines:
            line = line.strip()

            if line.startswith('m='):
                # Новая media секция
                if current_media:
                    media_sections.append(current_media)

                parts = line[2:].split()
                media_type = parts[0]
                port = parts[1]
                proto = parts[2]
                pts = parts[3:]

                current_media = {
                    'type': media_type,
                    'port': port,
                    'proto': proto,
                    'payload_types': pts,
                    'mid': None,
                    'ufrag': None,
                    'pwd': None,
                    'fingerprint': None,
                    'setup': None,
                    'candidates': [],
                    'rtpmap': {},
                    'fmtp': {}
                }

            elif line.startswith('a=mid:') and current_media:
                current_media['mid'] = line[6:]

            elif line.startswith('a=ice-ufrag:') and current_media:
                current_media['ufrag'] = line[12:]

            elif line.startswith('a=ice-pwd:') and current_media:
                current_media['pwd'] = line[10:]

            elif line.startswith('a=fingerprint:') and current_media:
                parts = line[14:].split(None, 1)
                if len(parts) == 2:
                    current_media['fingerprint'] = {
                        'hash': parts[0],
                        'value': parts[1]
                    }

            elif line.startswith('a=setup:') and current_media:
                current_media['setup'] = line[8:]

            elif line.startswith('a=candidate:') and current_media:
                current_media['candidates'].append(line[12:])

            elif line.startswith('a=rtpmap:') and current_media:
                parts = line[9:].split(None, 1)
                if len(parts) == 2:
                    pt_id = parts[0]
                    current_media['rtpmap'][pt_id] = parts[1]

            elif line.startswith('a=fmtp:') and current_media:
                parts = line[7:].split(None, 1)
                if len(parts) == 2:
                    pt_id = parts[0]
                    current_media['fmtp'][pt_id] = parts[1]

        if current_media:
            media_sections.append(current_media)

        return media_sections

    @staticmethod
    def _create_content(media):
        """Создает Jingle content из media секции"""
        content = ET.Element('content', {
            'creator': 'responder',
            'name': media['mid'] or 'media',
            'senders': 'responder'
        })

        # Description
        desc = ET.SubElement(content, 'description', {
            'xmlns': 'urn:xmpp:jingle:apps:rtp:1',
            'media': media['type']
        })

        # Payload types
        for pt_id in media['payload_types']:
            if pt_id in media['rtpmap']:
                rtpmap = media['rtpmap'][pt_id]
                parts = rtpmap.split('/')

                pt_attrs = {
                    'id': pt_id,
                    'name': parts[0]
                }

                if len(parts) > 1:
                    pt_attrs['clockrate'] = parts[1]
                if len(parts) > 2:
                    pt_attrs['channels'] = parts[2]

                ET.SubElement(desc, 'payload-type', pt_attrs)

        # RTCP-mux
        ET.SubElement(desc, 'rtcp-mux')

        # Transport
        transport_attrs = {
            'xmlns': 'urn:xmpp:jingle:transports:ice-udp:1'
        }

        if media['ufrag']:
            transport_attrs['ufrag'] = media['ufrag']
        if media['pwd']:
            transport_attrs['pwd'] = media['pwd']

        transport = ET.SubElement(content, 'transport', transport_attrs)

        # Fingerprint
        if media['fingerprint']:
            fp = ET.SubElement(transport, 'fingerprint', {
                'xmlns': 'urn:xmpp:jingle:apps:dtls:0',
                'hash': media['fingerprint']['hash'],
                'setup': media['setup'] or 'active'
            })
            fp.text = media['fingerprint']['value']

        return content


def test_conversion():
    """Тест конвертации"""
    # Пример Jingle XML
    jingle_xml = """
    <jingle xmlns='urn:xmpp:jingle:1' action='session-initiate' sid='test123'>
        <content name='audio' creator='initiator'>
            <description xmlns='urn:xmpp:jingle:apps:rtp:1' media='audio'>
                <payload-type id='111' name='opus' clockrate='48000' channels='2'>
                    <parameter name='minptime' value='10'/>
                    <parameter name='useinbandfec' value='1'/>
                </payload-type>
                <rtcp-mux/>
            </description>
            <transport xmlns='urn:xmpp:jingle:transports:ice-udp:1' ufrag='test' pwd='testpwd'>
                <fingerprint xmlns='urn:xmpp:jingle:apps:dtls:0' hash='sha-256' setup='actpass'>
                    AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99
                </fingerprint>
                <candidate foundation='1' component='1' protocol='udp' priority='2130706431'
                          ip='192.168.1.1' port='10000' type='host'/>
            </transport>
        </content>
    </jingle>
    """

    jingle_elem = ET.fromstring(jingle_xml)
    sdp = JingleToSDP.convert(jingle_elem)

    print("Generated SDP:")
    print(sdp)


if __name__ == "__main__":
    logging.basicConfig(level=logging.DEBUG)
    test_conversion()

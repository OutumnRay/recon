-- MUC Recorder Events Module
-- Отправляет события участников в recorder через HTTP

local http = require "net.http";
local json = require "util.json";
local jid = require "util.jid";
local timer = require "util.timer";

local RECORDER_WEBHOOK_URL = os.getenv("RECORDER_WEBHOOK_URL") or "http://recorder:8080/events";

module:log("warn", "=================================================");
module:log("warn", "MUC RECORDER EVENTS MODULE LOADED");
module:log("warn", "Webhook URL: %s", RECORDER_WEBHOOK_URL);
module:log("warn", "=================================================");

local event_count = 0;

local function send_event(event_type, room_jid, occupant)
    event_count = event_count + 1;

    local room_name = jid.node(room_jid);
    local participant_id = occupant.bare_jid;
    local display_name = occupant.nick or "Unknown";

    -- Извлекаем endpoint_id из occupant
    local endpoint_id = jid.resource(occupant.nick) or tostring(occupant):match("occupant:(.+)") or "unknown";

    local payload = {
        eventType = event_type,
        roomName = room_name,
        roomJid = room_jid,
        participantId = participant_id,
        participantName = participant_id,
        displayName = display_name,
        endpointId = endpoint_id,
        timestamp = os.time()
    };

    local payload_json = json.encode(payload);

    module:log("warn", "📤 [Event #%d] Sending %s for %s in %s", event_count, event_type, display_name, room_name);
    module:log("debug", "Payload: %s", payload_json);

    http.request(RECORDER_WEBHOOK_URL, {
        method = "POST",
        headers = {
            ["Content-Type"] = "application/json",
            ["Content-Length"] = tostring(#payload_json)
        },
        body = payload_json
    }, function(response_body, response_code, response_headers)
        if response_code == 200 then
            module:log("info", "✅ Event sent successfully: %s for %s", event_type, display_name);
        else
            module:log("error", "❌ Event failed: HTTP %s - %s", tostring(response_code), tostring(response_body));
        end
    end);
end

-- Участник присоединился
module:hook("muc-occupant-joined", function(event)
    module:log("warn", "🔔 MUC-OCCUPANT-JOINED hook triggered!");
    send_event("participantJoined", event.room.jid, event.occupant);
    return nil;
end, 50);

-- Участник вышел
module:hook("muc-occupant-left", function(event)
    module:log("warn", "🔔 MUC-OCCUPANT-LEFT hook triggered!");
    send_event("participantLeft", event.room.jid, event.occupant);
    return nil;
end, 50);

-- Комната уничтожена
module:hook("muc-room-destroyed", function(event)
    module:log("warn", "🔔 MUC-ROOM-DESTROYED hook triggered!");
    local room = event.room;
    local room_name = jid.node(room.jid);

    local payload = {
        eventType = "conferenceEnded",
        roomName = room_name,
        roomJid = room.jid,
        timestamp = os.time()
    };

    local payload_json = json.encode(payload);

    module:log("warn", "📤 Sending conferenceEnded for %s", room_name);

    http.request(RECORDER_WEBHOOK_URL, {
        method = "POST",
        headers = {
            ["Content-Type"] = "application/json",
            ["Content-Length"] = tostring(#payload_json)
        },
        body = payload_json
    }, function(response_body, response_code, response_headers)
        if response_code == 200 then
            module:log("info", "✅ Conference ended event sent");
        else
            module:log("error", "❌ Conference ended event failed: HTTP %s", tostring(response_code));
        end
    end);

    return nil;
end, 50);

-- Тестовый таймер для проверки что модуль работает
timer.add_task(10, function()
    module:log("warn", "❤️  MUC Recorder Events module heartbeat (events sent: %d)", event_count);
    return 10; -- повторять каждые 10 секунд
end);

module:log("warn", "✅ MUC Recorder Events module fully initialized");
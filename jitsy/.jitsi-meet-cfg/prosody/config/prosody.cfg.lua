-- ============================================================ --
-- Prosody Configuration for Jitsi Meet                         --
-- Этот файл был очищен и дополнен для интеграции с рекордером. --
-- ============================================================ --

---------- Глобальные настройки сервера ----------
admins = { }
component_admins_as_room_owners = true

-- Используем libevent/epoll для лучшей производительности
use_libevent = true
network_backend = "epoll"
network_settings = {
  tcp_backlog = 511;
}

-- Модули, необходимые для работы Jitsi
modules_enabled = {
	-- Основные
	"roster";
	"saslauth";
	"tls";
	"disco";
	"posix";

	-- Важные для Jitsi
	"private";
	"limits";
	"version";
	"ping";
	"http_health";
	"external_services";

	-- === ДОБАВЛЕНО: Модуль для отправки событий === --
	"muc_webhook";
}

-- Отключаем неиспользуемые модули
modules_disabled = {
	"offline";
	"register";
	"s2s";
}

-- Запрещаем регистрацию новых пользователей через XMPP
allow_registration = false

-- Требуем шифрование для клиентских подключений
c2s_require_encryption = true
s2s_secure_auth = false -- Для совместимости

-- Аутентификация
authentication = "internal_hashed"

-- Настройки логгирования: выводим информацию уровня 'info' и выше в консоль
log = {
	{ levels = { min = "info" }, timestamps = "%Y-%m-%d %X", to = "console" };
}

-- Настройки сборщика мусора для оптимизации
gc = {
	mode = "incremental";
	threshold = 400;
	speed = 250;
	step_size = 13;
}

-- Пути и порты
pidfile = "/config/data/prosody.pid"
data_path = "/config/data"
c2s_ports = { 5222 }
http_ports = { 5280 }
trusted_proxies = { "127.0.0.1", "::1" }


-- =================================================================== --
-- === НАСТРОЙКИ ДЛЯ WEBHOOK, ОТПРАВЛЯЮЩЕГО ДАННЫЕ В РЕКОРДЕР === --
-- =================================================================== --

-- URL, на который будут отправляться POST-запросы с событиями.
-- 'recorder' - это имя сервиса из вашего docker-compose.yml.
muc_webhook_url = "http://recorder:8080/webhook"

-- Список событий, которые нужно отправлять.
-- Это позволяет не нагружать рекордер лишними данными.
muc_webhook_events = {
    "conference-created",
    "conference-expired",
    "endpoint-created",
    "endpoint-expired"
}


-- =================================================== --
-- === НАСТРОЙКИ TURN-СЕРВЕРА (из вашего файла) === --
-- =================================================== --
external_services = {
	{
		type = "turn",
		host = "5.129.227.21",
		port = 3478,
		transport = "tcp",
		ttl = 86400,
		secret = true,
		algorithm = "turn",
	},
	{
		type = "turns",
		host = "5.129.227.21",
		port = 5349,
		transport = "tcp",
		ttl = 86400,
		secret = true,
		algorithm = "turn",
	}
}

consider_websocket_as_trusted = true;

-- Подключаем остальные файлы конфигурации для Jitsi (виртуальные хосты и компоненты)
-- ВАЖНО: Эта строка должна быть в конце файла.
Include "conf.d/*.cfg.lua"
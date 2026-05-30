// Конфигурация нагрузочного тестирования Recontext.online
// Все параметры можно переопределить через переменные окружения k6:
//   k6 run -e SERVER=192.168.1.10 -e MANAGING_PORT=20080 full-suite.js

export const BASE = {
  managingPortal: `http://${__ENV.SERVER || 'localhost'}:${__ENV.MANAGING_PORT || '20080'}`,
  userPortal:     `http://${__ENV.SERVER || 'localhost'}:${__ENV.USER_PORT || '20081'}`,
};

export const CREDS = {
  admin: { username: __ENV.ADMIN_USER || 'admin', password: __ENV.ADMIN_PASS || 'admin123' },
  user:  { username: __ENV.USER_USER  || 'user',  password: __ENV.USER_PASS  || 'user123'  },
};

// Пороги качества (thresholds) — применяются во всех сценариях
export const THRESHOLDS = {
  http_req_failed:   ['rate<0.05'],        // < 5% ошибок
  http_req_duration: ['p(95)<2000'],       // 95-й перцентиль < 2 с
  'http_req_duration{type:health}':   ['p(99)<300'],   // health < 300 мс
  'http_req_duration{type:auth}':     ['p(95)<1000'],  // auth < 1 с
  'http_req_duration{type:api_read}': ['p(95)<1500'],  // чтение API < 1.5 с
  'http_req_duration{type:api_write}':['p(95)<2000'],  // запись API < 2 с
};

export const JSON_HEADERS = {
  'Content-Type': 'application/json',
  'Accept': 'application/json',
};

export function authHeaders(token) {
  return {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
    'Authorization': `Bearer ${token}`,
  };
}

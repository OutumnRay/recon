/**
 * Spike Test — внезапный пик нагрузки
 * Проверяем, как система восстанавливается после резкого скачка трафика.
 *
 * Запуск:
 *   k6 run -e SERVER=<IP> load-tests/spike-test.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { BASE, CREDS, JSON_HEADERS, authHeaders } from './config.js';

export const options = {
  stages: [
    { duration: '30s', target: 5   },  // нормальный трафик
    { duration: '10s', target: 150 },  // резкий спайк
    { duration: '1m',  target: 150 },  // держим пик
    { duration: '10s', target: 5   },  // резкое снижение
    { duration: '2m',  target: 5   },  // фаза восстановления
    { duration: '30s', target: 0   },
  ],
  thresholds: {
    http_req_failed:   ['rate<0.20'],   // допускаем до 20% ошибок на пике
    http_req_duration: ['p(90)<5000'],
  },
};

const state = { managingToken: null, userToken: null };

export default function () {
  // Авторизация
  if (!state.managingToken) {
    const r = http.post(`${BASE.managingPortal}/api/v1/auth/login`,
      JSON.stringify(CREDS.admin), { headers: JSON_HEADERS, tags: { type: 'auth' } });
    if (r.status === 200) state.managingToken = r.json('token');
  }
  if (!state.userToken) {
    const r = http.post(`${BASE.userPortal}/api/v1/auth/login`,
      JSON.stringify(CREDS.user), { headers: JSON_HEADERS, tags: { type: 'auth' } });
    if (r.status === 200) state.userToken = r.json('token');
  }

  // Во время пика — только самые критичные эндпоинты
  const r1 = http.get(`${BASE.managingPortal}/health`, { tags: { type: 'health' } });
  check(r1, { 'health up': (r) => r.status === 200 });

  if (state.userToken) {
    const r2 = http.get(`${BASE.userPortal}/api/v1/files`,
      { headers: authHeaders(state.userToken), tags: { type: 'api_read' } });
    check(r2, { 'files reachable': (r) => r.status < 500 });
  }

  sleep(0.2);
}

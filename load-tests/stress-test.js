/**
 * Stress Test — поиск точки отказа
 * Нагрузка нарастает до 200 VU, чтобы найти предел системы.
 * После теста смотрим: при каком VU начинают расти ошибки / latency.
 *
 * Запуск:
 *   k6 run -e SERVER=<IP> load-tests/stress-test.js
 *
 * ВНИМАНИЕ: тест намеренно выходит за безопасные пределы.
 * Запускать при наличии возможности быстро остановить (Ctrl+C).
 */

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { BASE, CREDS, JSON_HEADERS, authHeaders } from './config.js';

export const options = {
  stages: [
    { duration: '2m',  target: 25  },
    { duration: '2m',  target: 50  },
    { duration: '2m',  target: 100 },
    { duration: '2m',  target: 150 },
    { duration: '2m',  target: 200 },
    { duration: '3m',  target: 200 },  // удержание на пике
    { duration: '2m',  target: 0   },  // завершение
  ],
  // Пороги намеренно мягкие — нас интересует динамика, а не pass/fail
  thresholds: {
    http_req_failed:   ['rate<0.30'],
    http_req_duration: ['p(95)<10000'],
  },
};

const state = { managingToken: null, userToken: null };

function login(url, creds) {
  const r = http.post(`${url}/api/v1/auth/login`,
    JSON.stringify(creds), { headers: JSON_HEADERS, tags: { type: 'auth' } });
  return r.status === 200 ? r.json('token') : null;
}

export default function () {
  if (!state.managingToken) state.managingToken = login(BASE.managingPortal, CREDS.admin);
  if (!state.userToken)     state.userToken     = login(BASE.userPortal, CREDS.user);

  const roll = Math.random();

  if (roll < 0.3) {
    // Тяжёлые чтения — managing portal
    group('stress-managing', () => {
      http.get(`${BASE.managingPortal}/health`, { tags: { type: 'health' } });
      if (state.managingToken) {
        http.get(`${BASE.managingPortal}/api/v1/users`,
          { headers: authHeaders(state.managingToken), tags: { type: 'api_read' } });
        http.get(`${BASE.managingPortal}/api/v1/livekit/rooms`,
          { headers: authHeaders(state.managingToken), tags: { type: 'api_read' } });
        http.get(`${BASE.managingPortal}/api/v1/metrics/system`,
          { headers: authHeaders(state.managingToken), tags: { type: 'api_read' } });
      }
    });
  } else if (roll < 0.5) {
    // Webhook flood
    group('stress-webhook', () => {
      const sid = `RM_STRESS_${__VU}_${__ITER}`;
      http.post(`${BASE.managingPortal}/webhook/meet`,
        JSON.stringify({
          event: 'participant_joined',
          room: { sid, name: `stress-${__VU}` },
          participant: {
            sid: `PA_${__VU}_${__ITER}`,
            identity: `user_${__VU}`,
            state: 'ACTIVE',
            joinedAt: `${Math.floor(Date.now() / 1000)}`,
          },
          id: `EV_${__VU}_${__ITER}`,
          createdAt: `${Math.floor(Date.now() / 1000)}`,
        }),
        { headers: JSON_HEADERS, tags: { type: 'api_write' } },
      );
    });
  } else {
    // User portal reads
    group('stress-user', () => {
      if (state.userToken) {
        http.get(`${BASE.userPortal}/api/v1/files?page_size=50`,
          { headers: authHeaders(state.userToken), tags: { type: 'api_read' } });
        http.get(`${BASE.userPortal}/api/v1/meetings`,
          { headers: authHeaders(state.userToken), tags: { type: 'api_read' } });
      }
    });
  }

  sleep(0.1 + Math.random() * 0.5);
}

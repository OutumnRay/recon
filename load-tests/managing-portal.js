/**
 * Managing Portal — детальный тест всех API-групп
 * 20 VU, 5 минут
 *
 * Запуск:
 *   k6 run -e SERVER=<IP> load-tests/managing-portal.js
 */

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { BASE, CREDS, JSON_HEADERS, authHeaders } from './config.js';

export const options = {
  vus: 20,
  duration: '5m',
  thresholds: {
    http_req_failed:   ['rate<0.05'],
    http_req_duration: ['p(95)<2000'],
    'http_req_duration{type:health}': ['p(99)<300'],
    'http_req_duration{type:auth}':   ['p(95)<1000'],
  },
};

let token = null;

export default function () {
  if (!token) {
    const r = http.post(
      `${BASE.managingPortal}/api/v1/auth/login`,
      JSON.stringify(CREDS.admin),
      { headers: JSON_HEADERS, tags: { type: 'auth' } },
    );
    check(r, { 'login 200': (x) => x.status === 200 });
    token = r.json('token');
    if (!token) return;
  }

  const hdrs = authHeaders(token);

  group('health-and-status', () => {
    let r = http.get(`${BASE.managingPortal}/health`, { tags: { type: 'health' } });
    check(r, { '/health 200': (x) => x.status === 200 });

    r = http.get(`${BASE.managingPortal}/api/v1/status`, { headers: hdrs, tags: { type: 'api_read' } });
    check(r, { '/status 200': (x) => x.status === 200 });
  });

  group('users', () => {
    let r = http.get(`${BASE.managingPortal}/api/v1/users?page_size=10`,
      { headers: hdrs, tags: { type: 'api_read' } });
    check(r, { 'GET /users 200': (x) => x.status === 200 });

    // Получаем первого пользователя и читаем детали
    try {
      const users = r.json('users') || r.json('items') || [];
      if (users.length > 0) {
        const uid = users[0].id;
        r = http.get(`${BASE.managingPortal}/api/v1/users/${uid}`,
          { headers: hdrs, tags: { type: 'api_read' } });
        check(r, { 'GET /users/:id 200': (x) => x.status === 200 });
      }
    } catch (_) {}
  });

  group('groups', () => {
    let r = http.get(`${BASE.managingPortal}/api/v1/groups`,
      { headers: hdrs, tags: { type: 'api_read' } });
    check(r, { 'GET /groups 200': (x) => x.status === 200 });
  });

  group('departments', () => {
    let r = http.get(`${BASE.managingPortal}/api/v1/departments`,
      { headers: hdrs, tags: { type: 'api_read' } });
    check(r, { 'GET /departments 200': (x) => x.status === 200 });

    r = http.get(`${BASE.managingPortal}/api/v1/departments?format=tree`,
      { headers: hdrs, tags: { type: 'api_read' } });
    check(r, { 'GET /departments?tree 200': (x) => x.status === 200 });
  });

  group('metrics', () => {
    let r = http.get(`${BASE.managingPortal}/api/v1/metrics/system`,
      { headers: hdrs, tags: { type: 'api_read' } });
    check(r, { 'GET /metrics/system 200': (x) => x.status === 200 });

    r = http.get(`${BASE.managingPortal}/api/v1/metrics`,
      { headers: hdrs, tags: { type: 'api_read' } });
    check(r, { 'GET /metrics 200': (x) => x.status === 200 });

    // Отправка метрики (write)
    r = http.post(
      `${BASE.managingPortal}/api/v1/metrics`,
      JSON.stringify({
        service_id: 'k6-load-test',
        metrics: [
          { name: 'requests_total', type: 'counter', value: 1 },
          { name: 'latency_ms',     type: 'gauge',   value: 42.0 },
        ],
      }),
      { headers: hdrs, tags: { type: 'api_write' } },
    );
    check(r, { 'POST /metrics 200/204': (x) => [200, 201, 204].includes(x.status) });
  });

  group('livekit', () => {
    let r = http.get(`${BASE.managingPortal}/api/v1/livekit/rooms`,
      { headers: hdrs, tags: { type: 'api_read' } });
    check(r, { 'GET /livekit/rooms 200': (x) => x.status === 200 });

    r = http.get(`${BASE.managingPortal}/api/v1/livekit/participants?room_sid=NONE`,
      { headers: hdrs, tags: { type: 'api_read' } });
    check(r, { 'GET /livekit/participants 200/404': (x) => [200, 404].includes(x.status) });

    r = http.get(`${BASE.managingPortal}/api/v1/livekit/webhook-events`,
      { headers: hdrs, tags: { type: 'api_read' } });
    check(r, { 'GET /webhook-events 200': (x) => x.status === 200 });
  });

  group('webhook-simulation', () => {
    const ts = Math.floor(Date.now() / 1000);
    const sid = `RM_K6_${__VU}_${__ITER}`;

    // Симулируем полный жизненный цикл комнаты
    const events = [
      { event: 'room_started', room: { sid, name: `k6-room-${__VU}`, emptyTimeout: 300, departureTimeout: 20, creationTime: `${ts}`, creationTimeMs: `${ts * 1000}` } },
      { event: 'participant_joined', room: { sid, name: `k6-room-${__VU}` }, participant: { sid: `PA_${__VU}_${__ITER}`, identity: `k6_user_${__VU}`, state: 'ACTIVE', joinedAt: `${ts}` } },
      { event: 'participant_left', room: { sid, name: `k6-room-${__VU}` }, participant: { sid: `PA_${__VU}_${__ITER}`, identity: `k6_user_${__VU}`, state: 'DISCONNECTED' } },
      { event: 'room_finished', room: { sid, name: `k6-room-${__VU}`, numParticipants: 0 } },
    ];

    for (const payload of events) {
      const r = http.post(
        `${BASE.managingPortal}/webhook/meet`,
        JSON.stringify({ ...payload, id: `EV_${__VU}_${__ITER}_${payload.event}`, createdAt: `${ts}` }),
        { headers: JSON_HEADERS, tags: { type: 'api_write' } },
      );
      check(r, { [`webhook ${payload.event} ok`]: (x) => [200, 204].includes(x.status) });
      sleep(0.05);
    }
  });

  sleep(0.5 + Math.random());
}

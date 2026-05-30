/**
 * Full Suite — комплексный тест с именованными сценариями
 * Запускает все группы параллельно с разными профилями нагрузки.
 *
 * Запуск:
 *   k6 run -e SERVER=<IP> load-tests/full-suite.js
 *
 *   Только один сценарий:
 *   k6 run -e SERVER=<IP> --scenario managing_load load-tests/full-suite.js
 *
 *   С HTML-отчётом:
 *   k6 run -e SERVER=<IP> --out json=results.json load-tests/full-suite.js
 */

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Counter, Trend } from 'k6/metrics';
import { BASE, CREDS, THRESHOLDS, JSON_HEADERS, authHeaders } from './config.js';

// ─── Кастомные метрики ────────────────────────────────────────────────────────
const loginSuccess   = new Counter('login_success_total');
const loginFail      = new Counter('login_fail_total');
const healthFail     = new Counter('health_fail_total');
const webhookLatency = new Trend('webhook_latency', true);

export const options = {
  scenarios: {
    // Сценарий 1: постоянная нагрузка health-check
    health_check: {
      executor: 'constant-vus',
      vus: 3,
      duration: '10m',
      exec: 'healthScenario',
      tags: { scenario: 'health' },
    },

    // Сценарий 2: управляющий портал — рабочая нагрузка
    managing_load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 10 },
        { duration: '4m', target: 20 },
        { duration: '3m', target: 20 },
        { duration: '2m', target: 0  },
      ],
      exec: 'managingScenario',
      tags: { scenario: 'managing' },
    },

    // Сценарий 3: пользовательский портал — рабочая нагрузка
    user_load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m',  target: 15 },
        { duration: '4m',  target: 30 },
        { duration: '3m',  target: 30 },
        { duration: '2m',  target: 0  },
      ],
      exec: 'userScenario',
      tags: { scenario: 'user' },
    },

    // Сценарий 4: webhook flood (LiveKit events)
    webhook_flood: {
      executor: 'constant-arrival-rate',
      rate: 20,           // 20 событий/сек
      timeUnit: '1s',
      duration: '8m',
      preAllocatedVUs: 10,
      maxVUs: 30,
      exec: 'webhookScenario',
      tags: { scenario: 'webhook' },
    },

    // Сценарий 5: периодические записи (metrics send)
    metrics_push: {
      executor: 'constant-arrival-rate',
      rate: 5,            // 5 метрик/сек
      timeUnit: '1s',
      duration: '8m',
      preAllocatedVUs: 5,
      maxVUs: 15,
      exec: 'metricsPushScenario',
      tags: { scenario: 'metrics' },
    },
  },
  thresholds: {
    ...THRESHOLDS,
    login_fail_total:  ['count<50'],
    health_fail_total: ['count<5'],
    webhook_latency:   ['p(95)<500'],
  },
};

// ─── Кэш токенов per-VU ───────────────────────────────────────────────────────
const _tokens = {};

function getToken(portalBase, creds, key) {
  if (_tokens[key]) return _tokens[key];
  const r = http.post(`${portalBase}/api/v1/auth/login`,
    JSON.stringify(creds), { headers: JSON_HEADERS, tags: { type: 'auth' } });
  if (r.status === 200) {
    loginSuccess.add(1);
    _tokens[key] = r.json('token');
  } else {
    loginFail.add(1);
  }
  return _tokens[key];
}

// ─── Сценарий: health ─────────────────────────────────────────────────────────
export function healthScenario() {
  const r = http.get(`${BASE.managingPortal}/health`, { tags: { type: 'health' } });
  if (!check(r, { 'health 200': (x) => x.status === 200 })) healthFail.add(1);
  sleep(2);
}

// ─── Сценарий: managing portal ────────────────────────────────────────────────
export function managingScenario() {
  const token = getToken(BASE.managingPortal, CREDS.admin, `mp_${__VU}`);
  if (!token) { sleep(2); return; }

  const hdrs = authHeaders(token);
  const roll = Math.random();

  if (roll < 0.25) {
    group('users-and-groups', () => {
      http.get(`${BASE.managingPortal}/api/v1/users?page_size=20`,
        { headers: hdrs, tags: { type: 'api_read' } });
      http.get(`${BASE.managingPortal}/api/v1/groups`,
        { headers: hdrs, tags: { type: 'api_read' } });
    });
  } else if (roll < 0.50) {
    group('departments', () => {
      http.get(`${BASE.managingPortal}/api/v1/departments`,
        { headers: hdrs, tags: { type: 'api_read' } });
    });
  } else if (roll < 0.75) {
    group('livekit-data', () => {
      http.get(`${BASE.managingPortal}/api/v1/livekit/rooms`,
        { headers: hdrs, tags: { type: 'api_read' } });
      http.get(`${BASE.managingPortal}/api/v1/livekit/webhook-events`,
        { headers: hdrs, tags: { type: 'api_read' } });
    });
  } else {
    group('metrics-read', () => {
      http.get(`${BASE.managingPortal}/api/v1/metrics/system`,
        { headers: hdrs, tags: { type: 'api_read' } });
      http.get(`${BASE.managingPortal}/api/v1/status`,
        { headers: hdrs, tags: { type: 'api_read' } });
    });
  }

  sleep(0.5 + Math.random());
}

// ─── Сценарий: user portal ────────────────────────────────────────────────────
export function userScenario() {
  const token = getToken(BASE.userPortal, CREDS.user, `up_${__VU}`);
  if (!token) { sleep(2); return; }

  const hdrs = authHeaders(token);
  const roll = Math.random();

  if (roll < 0.40) {
    group('files-browse', () => {
      http.get(`${BASE.userPortal}/api/v1/files?page=1&page_size=20`,
        { headers: hdrs, tags: { type: 'api_read' } });
    });
  } else if (roll < 0.70) {
    group('meetings-browse', () => {
      http.get(`${BASE.userPortal}/api/v1/meetings`,
        { headers: hdrs, tags: { type: 'api_read' } });
    });
  } else if (roll < 0.85) {
    group('assistant-search', () => {
      const r = http.post(`${BASE.userPortal}/api/v1/assistant/chat`,
        JSON.stringify({ mode: 'find_videos', query: 'встреча проект' }),
        { headers: hdrs, tags: { type: 'api_write' } });
      check(r, { 'assistant 200/503': (x) => [200, 503].includes(x.status) });
    });
  } else {
    group('file-init-cancel', () => {
      const r = http.post(`${BASE.userPortal}/api/v1/files/init`,
        JSON.stringify({
          filename: `suite-${__VU}-${__ITER}.mp4`,
          content_type: 'video/mp4',
          file_size: 5 * 1024 * 1024,
          title: `Suite VU${__VU}`,
          language: 'ru',
        }),
        { headers: hdrs, tags: { type: 'api_write' } });
      if (r.status === 200) {
        try {
          const fid = r.json('file_id');
          if (fid) {
            http.post(`${BASE.userPortal}/api/v1/files/${fid}/confirm`,
              JSON.stringify({ success: false }),
              { headers: hdrs, tags: { type: 'api_write' } });
          }
        } catch (_) {}
      }
    });
  }

  sleep(0.3 + Math.random() * 0.7);
}

// ─── Сценарий: webhook flood ──────────────────────────────────────────────────
export function webhookScenario() {
  const ts   = Math.floor(Date.now() / 1000);
  const sid  = `RM_SUITE_${__VU}_${__ITER}`;
  const pSid = `PA_SUITE_${__VU}_${__ITER}`;

  const events = [
    {
      event: 'room_started',
      room: { sid, name: `suite-room-${__VU}`, emptyTimeout: 300, departureTimeout: 20, creationTime: `${ts}`, creationTimeMs: `${ts}000` },
    },
    {
      event: 'participant_joined',
      room: { sid, name: `suite-room-${__VU}` },
      participant: { sid: pSid, identity: `k6_${__VU}`, state: 'ACTIVE', joinedAt: `${ts}` },
    },
  ];

  const idx   = __ITER % events.length;
  const event = events[idx];
  const start = Date.now();

  const r = http.post(
    `${BASE.managingPortal}/webhook/meet`,
    JSON.stringify({ ...event, id: `EV_SUITE_${__VU}_${__ITER}`, createdAt: `${ts}` }),
    { headers: JSON_HEADERS, tags: { type: 'api_write' } },
  );

  webhookLatency.add(Date.now() - start);
  check(r, { 'webhook 200/204': (x) => [200, 204].includes(x.status) });
}

// ─── Сценарий: metrics push ───────────────────────────────────────────────────
export function metricsPushScenario() {
  const token = getToken(BASE.managingPortal, CREDS.admin, `mp_metrics_${__VU}`);
  if (!token) return;

  const r = http.post(
    `${BASE.managingPortal}/api/v1/metrics`,
    JSON.stringify({
      service_id: `k6-suite-vu-${__VU}`,
      metrics: [
        { name: 'req_count', type: 'counter', value: 1 },
        { name: 'latency',   type: 'gauge',   value: Math.random() * 500 },
      ],
    }),
    { headers: authHeaders(token), tags: { type: 'api_write' } },
  );
  check(r, { 'metrics push ok': (x) => [200, 201, 204].includes(x.status) });
}

// default export не нужен при использовании именованных сценариев,
// но k6 требует его наличия — оставляем пустым
export default function () {}

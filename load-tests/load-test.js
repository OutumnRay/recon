/**
 * Load Test — нормальная рабочая нагрузка
 * Сценарий: плавный разгон до 50 VU → держим 5 минут → плавный спуск
 *
 * Охватывает:
 *  - Managing Portal: health, auth, users, groups, departments, metrics, livekit
 *  - User Portal: auth, files, meetings, assistant
 *  - LiveKit webhook (публичный эндпоинт)
 *
 * Запуск:
 *   k6 run -e SERVER=<IP> load-tests/load-test.js
 */

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import { BASE, CREDS, THRESHOLDS, JSON_HEADERS, authHeaders } from './config.js';

// ─── Кастомные метрики ────────────────────────────────────────────────────────
const loginErrors   = new Counter('login_errors');
const apiErrors     = new Counter('api_errors');
const webhookErrors = new Counter('webhook_errors');

export const options = {
  stages: [
    { duration: '1m',  target: 10  },  // разогрев
    { duration: '2m',  target: 30  },  // рост
    { duration: '5m',  target: 50  },  // плато
    { duration: '1m',  target: 10  },  // спуск
    { duration: '30s', target: 0   },  // остановка
  ],
  thresholds: {
    ...THRESHOLDS,
    login_errors:   ['count<20'],
    api_errors:     ['count<100'],
    webhook_errors: ['count<10'],
  },
};

// ─── Хранилище токенов per-VU ─────────────────────────────────────────────────
const state = {
  managingToken: null,
  userToken: null,
  managingLoginAt: 0,
  userLoginAt: 0,
};

const TOKEN_TTL_MS = 50 * 60 * 1000; // 50 минут (меньше JWT expiry)

function ensureManagingToken() {
  if (!state.managingToken || Date.now() - state.managingLoginAt > TOKEN_TTL_MS) {
    const r = http.post(
      `${BASE.managingPortal}/api/v1/auth/login`,
      JSON.stringify(CREDS.admin),
      { headers: JSON_HEADERS, tags: { type: 'auth' } },
    );
    if (r.status === 200) {
      state.managingToken   = r.json('token');
      state.managingLoginAt = Date.now();
    } else {
      loginErrors.add(1);
    }
  }
  return state.managingToken;
}

function ensureUserToken() {
  if (!state.userToken || Date.now() - state.userLoginAt > TOKEN_TTL_MS) {
    const r = http.post(
      `${BASE.userPortal}/api/v1/auth/login`,
      JSON.stringify(CREDS.user),
      { headers: JSON_HEADERS, tags: { type: 'auth' } },
    );
    if (r.status === 200) {
      state.userToken     = r.json('token');
      state.userLoginAt   = Date.now();
    } else {
      loginErrors.add(1);
    }
  }
  return state.userToken;
}

// ─── Сценарии ─────────────────────────────────────────────────────────────────

function scenarioManagingPortal(token) {
  group('managing-portal', () => {
    let r;

    // Health
    r = http.get(`${BASE.managingPortal}/health`, { tags: { type: 'health' } });
    if (!check(r, { 'MP health 200': (x) => x.status === 200 })) apiErrors.add(1);

    // System status
    r = http.get(`${BASE.managingPortal}/api/v1/status`,
      { headers: authHeaders(token), tags: { type: 'api_read' } });
    check(r, { 'MP status 200': (x) => x.status === 200 });

    // Users list
    r = http.get(`${BASE.managingPortal}/api/v1/users?page_size=20`,
      { headers: authHeaders(token), tags: { type: 'api_read' } });
    if (!check(r, { 'MP users 200': (x) => x.status === 200 })) apiErrors.add(1);

    // Groups
    r = http.get(`${BASE.managingPortal}/api/v1/groups`,
      { headers: authHeaders(token), tags: { type: 'api_read' } });
    check(r, { 'MP groups 200': (x) => x.status === 200 });

    // Departments
    r = http.get(`${BASE.managingPortal}/api/v1/departments`,
      { headers: authHeaders(token), tags: { type: 'api_read' } });
    check(r, { 'MP departments 200': (x) => x.status === 200 });

    // Meeting subjects
    r = http.get(`${BASE.managingPortal}/api/v1/meeting-subjects`,
      { headers: authHeaders(token), tags: { type: 'api_read' } });
    check(r, { 'MP meeting-subjects 200': (x) => [200, 404].includes(x.status) });

    // System metrics
    r = http.get(`${BASE.managingPortal}/api/v1/metrics/system`,
      { headers: authHeaders(token), tags: { type: 'api_read' } });
    check(r, { 'MP metrics/system 200': (x) => x.status === 200 });

    // LiveKit rooms
    r = http.get(`${BASE.managingPortal}/api/v1/livekit/rooms`,
      { headers: authHeaders(token), tags: { type: 'api_read' } });
    check(r, { 'MP livekit/rooms 200': (x) => x.status === 200 });

    sleep(0.2);
  });
}

function scenarioWebhook() {
  group('livekit-webhook', () => {
    const payload = JSON.stringify({
      event: 'room_started',
      room: {
        sid: `RM_LOAD_${__VU}_${Date.now()}`,
        name: `load-test-room-${__VU}`,
        emptyTimeout: 300,
        departureTimeout: 20,
        creationTime: `${Math.floor(Date.now() / 1000)}`,
        creationTimeMs: `${Date.now()}`,
      },
      id: `EV_LOAD_${__VU}_${Date.now()}`,
      createdAt: `${Math.floor(Date.now() / 1000)}`,
    });

    const r = http.post(
      `${BASE.managingPortal}/webhook/meet`,
      payload,
      { headers: JSON_HEADERS, tags: { type: 'api_write' } },
    );
    if (!check(r, { 'webhook 200/204': (x) => [200, 204].includes(x.status) })) {
      webhookErrors.add(1);
    }
  });
}

function scenarioUserPortal(token) {
  group('user-portal', () => {
    let r;

    // Files list
    r = http.get(`${BASE.userPortal}/api/v1/files?page=1&page_size=20`,
      { headers: authHeaders(token), tags: { type: 'api_read' } });
    if (!check(r, { 'UP files 200': (x) => x.status === 200 })) apiErrors.add(1);

    // Files with filters
    r = http.get(`${BASE.userPortal}/api/v1/files?status=completed&sort_by=uploaded_at&sort_order=desc`,
      { headers: authHeaders(token), tags: { type: 'api_read' } });
    check(r, { 'UP files filtered 200': (x) => x.status === 200 });

    // Meetings list
    r = http.get(`${BASE.userPortal}/api/v1/meetings?page_size=20`,
      { headers: authHeaders(token), tags: { type: 'api_read' } });
    check(r, { 'UP meetings 200': (x) => [200, 404].includes(x.status) });

    // AI assistant — поиск по видео (find_videos) — лёгкий режим (без LLM)
    r = http.post(
      `${BASE.userPortal}/api/v1/assistant/chat`,
      JSON.stringify({ mode: 'find_videos', query: 'test' }),
      { headers: authHeaders(token), tags: { type: 'api_write' } },
    );
    // 200 если LLM настроен, 503 если нет — оба варианта допустимы
    check(r, { 'UP assistant 200/503': (x) => [200, 503].includes(x.status) });

    sleep(0.3);
  });
}

// ─── Главная функция ──────────────────────────────────────────────────────────

export default function () {
  const mToken = ensureManagingToken();
  const uToken = ensureUserToken();

  // Распределение нагрузки: 40% managing, 15% webhook, 45% user portal
  const roll = Math.random();

  if (roll < 0.40) {
    scenarioManagingPortal(mToken);
  } else if (roll < 0.55) {
    scenarioWebhook();
  } else {
    scenarioUserPortal(uToken);
  }

  sleep(0.5 + Math.random());
}

/**
 * Smoke Test — 2 VU, 2 минуты
 * Цель: убедиться, что все сервисы отвечают до начала серьёзной нагрузки.
 *
 * Запуск:
 *   k6 run -e SERVER=<IP> load-tests/smoke-test.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { BASE, CREDS, JSON_HEADERS, authHeaders } from './config.js';

export const options = {
  vus: 2,
  duration: '2m',
  thresholds: {
    http_req_failed:   ['rate<0.01'],
    http_req_duration: ['p(95)<1000'],
  },
};

// Кэш токенов — одна авторизация на VU
let managingToken = null;
let userToken = null;

function loginManaging() {
  const res = http.post(
    `${BASE.managingPortal}/api/v1/auth/login`,
    JSON.stringify(CREDS.admin),
    { headers: JSON_HEADERS, tags: { type: 'auth' } },
  );
  check(res, { 'managing login 200': (r) => r.status === 200 });
  const body = res.json();
  return body && body.token ? body.token : null;
}

function loginUser() {
  const res = http.post(
    `${BASE.userPortal}/api/v1/auth/login`,
    JSON.stringify(CREDS.user),
    { headers: JSON_HEADERS, tags: { type: 'auth' } },
  );
  check(res, { 'user login 200': (r) => r.status === 200 });
  const body = res.json();
  return body && body.token ? body.token : null;
}

export default function () {
  // Авторизуемся один раз per-VU
  if (!managingToken) managingToken = loginManaging();
  if (!userToken)     userToken     = loginUser();

  // ─── Managing Portal ───────────────────────────────────────────────
  let r;

  r = http.get(`${BASE.managingPortal}/health`, { tags: { type: 'health' } });
  check(r, { 'MP /health 200': (x) => x.status === 200 });

  r = http.get(`${BASE.managingPortal}/api/v1/status`,
    { headers: authHeaders(managingToken), tags: { type: 'api_read' } });
  check(r, { 'MP /status 200': (x) => x.status === 200 });

  r = http.get(`${BASE.managingPortal}/api/v1/users`,
    { headers: authHeaders(managingToken), tags: { type: 'api_read' } });
  check(r, { 'MP /users 200': (x) => x.status === 200 });

  // ─── User Portal ───────────────────────────────────────────────────
  r = http.get(`${BASE.userPortal}/api/v1/files`,
    { headers: authHeaders(userToken), tags: { type: 'api_read' } });
  check(r, { 'UP /files 200': (x) => x.status === 200 });

  r = http.get(`${BASE.userPortal}/api/v1/meetings`,
    { headers: authHeaders(userToken), tags: { type: 'api_read' } });
  check(r, { 'UP /meetings 200': (x) => [200, 404].includes(x.status) });

  sleep(1);
}

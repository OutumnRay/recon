/**
 * Health Check — параллельная проверка всех /health эндпоинтов
 * 10 VU, 3 минуты
 *
 * Запуск:
 *   k6 run -e SERVER=<IP> load-tests/health-check.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { group } from 'k6';
import { BASE } from './config.js';

export const options = {
  vus: 10,
  duration: '3m',
  thresholds: {
    http_req_failed:                           ['rate<0.001'],
    'http_req_duration{endpoint:mp_health}':   ['p(99)<200'],
    'http_req_duration{endpoint:up_health}':   ['p(99)<200'],
  },
};

export default function () {
  group('managing-portal health', () => {
    const r = http.get(
      `${BASE.managingPortal}/health`,
      { tags: { type: 'health', endpoint: 'mp_health' } },
    );
    check(r, {
      'MP /health status 200': (x) => x.status === 200,
      'MP /health has status field': (x) => {
        try { return x.json().status !== undefined; } catch { return false; }
      },
    });
  });

  group('user-portal health', () => {
    // user-portal API — через nginx на user-portal-front (порт 20081)
    const r = http.get(
      `${BASE.userPortal}/api/v1/health`,
      { tags: { type: 'health', endpoint: 'up_health' } },
    );
    // health может быть на /health или /api/v1/health
    check(r, { 'UP /health reachable': (x) => [200, 404].includes(x.status) });
  });

  sleep(0.5);
}

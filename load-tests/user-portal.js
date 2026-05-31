/**
 * User Portal — детальный тест пользовательских API
 * 30 VU, 5 минут
 *
 * Тестирует: login, files (init/list/status/transcript/summary),
 *            meetings (list/detail), AI-assistant
 *
 * Запуск:
 *   k6 run -e SERVER=<IP> load-tests/user-portal.js
 */

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { BASE, CREDS, JSON_HEADERS, authHeaders } from './config.js';

export const options = {
  vus: 30,
  duration: '5m',
  thresholds: {
    http_req_failed:                    ['rate<0.05'],
    http_req_duration:                  ['p(95)<2000'],
    'http_req_duration{type:auth}':     ['p(95)<1000'],
    'http_req_duration{type:api_read}': ['p(95)<1500'],
  },
};

let token = null;
let knownFileId = null;
let knownMeetingId = null;

function doLogin() {
  const r = http.post(
    `${BASE.userPortal}/api/v1/auth/login`,
    JSON.stringify(CREDS.user),
    { headers: JSON_HEADERS, tags: { type: 'auth' } },
  );
  check(r, { 'login 200': (x) => x.status === 200 });
  return r.json('token');
}

export default function () {
  if (!token) {
    token = doLogin();
    if (!token) { sleep(1); return; }
  }

  const hdrs = authHeaders(token);

  // ─── Files ─────────────────────────────────────────────────────────────────
  group('files', () => {
    // Список файлов (разные фильтры)
    let r = http.get(`${BASE.userPortal}/api/v1/files?page=1&page_size=20`,
      { headers: hdrs, tags: { type: 'api_read' } });
    if (check(r, { 'GET /files 200': (x) => x.status === 200 })) {
      try {
        const items = r.json('items') || r.json('files') || [];
        if (items.length > 0 && !knownFileId) knownFileId = items[0].id;
      } catch (_) {}
    }

    r = http.get(`${BASE.userPortal}/api/v1/files?status=completed&sort_by=uploaded_at&sort_order=desc`,
      { headers: hdrs, tags: { type: 'api_read' } });
    check(r, { 'GET /files?completed 200': (x) => x.status === 200 });

    r = http.get(`${BASE.userPortal}/api/v1/files?search=test`,
      { headers: hdrs, tags: { type: 'api_read' } });
    check(r, { 'GET /files?search 200': (x) => x.status === 200 });

    // Детали файла (если есть)
    if (knownFileId) {
      r = http.get(`${BASE.userPortal}/api/v1/files/${knownFileId}`,
        { headers: hdrs, tags: { type: 'api_read' } });
      check(r, { 'GET /files/:id 200': (x) => [200, 404].includes(x.status) });

      r = http.get(`${BASE.userPortal}/api/v1/files/${knownFileId}/status`,
        { headers: hdrs, tags: { type: 'api_read' } });
      check(r, { 'GET /files/:id/status 200': (x) => [200, 404].includes(x.status) });

      r = http.get(`${BASE.userPortal}/api/v1/files/${knownFileId}/transcript?page=1&page_size=50`,
        { headers: hdrs, tags: { type: 'api_read' } });
      check(r, { 'GET /files/:id/transcript 200': (x) => [200, 404].includes(x.status) });

      r = http.get(`${BASE.userPortal}/api/v1/files/${knownFileId}/summary`,
        { headers: hdrs, tags: { type: 'api_read' } });
      check(r, { 'GET /files/:id/summary 200': (x) => [200, 404].includes(x.status) });
    }

    // Init upload (без реальной загрузки — только проверяем ответ API)
    r = http.post(
      `${BASE.userPortal}/api/v1/files/init`,
      JSON.stringify({
        filename:     `k6-test-${__VU}-${__ITER}.mp4`,
        content_type: 'video/mp4',
        file_size:    10 * 1024 * 1024,  // 10 МБ
        title:        `K6 Load Test ${__VU}`,
        language:     'ru',
        chunk_size:   10 * 1024 * 1024,
      }),
      { headers: hdrs, tags: { type: 'api_write' } },
    );
    if (check(r, { 'POST /files/init 200': (x) => x.status === 200 })) {
      // Если upload создался — отменяем его через confirm(success=false)
      try {
        const fid = r.json('file_id');
        if (fid) {
          http.post(
            `${BASE.userPortal}/api/v1/files/${fid}/confirm`,
            JSON.stringify({ success: false }),
            { headers: hdrs, tags: { type: 'api_write' } },
          );
        }
      } catch (_) {}
    }
  });

  // ─── Meetings ──────────────────────────────────────────────────────────────
  group('meetings', () => {
    let r = http.get(`${BASE.userPortal}/api/v1/meetings?page_size=20`,
      { headers: hdrs, tags: { type: 'api_read' } });
    if (check(r, { 'GET /meetings 200': (x) => [200, 404].includes(x.status) })) {
      try {
        const items = r.json('items') || [];
        if (items.length > 0 && !knownMeetingId) knownMeetingId = items[0].id;
      } catch (_) {}
    }

    // Фильтрация по статусу и типу
    r = http.get(`${BASE.userPortal}/api/v1/meetings?status=scheduled&type=conference`,
      { headers: hdrs, tags: { type: 'api_read' } });
    check(r, { 'GET /meetings?filtered 200': (x) => [200, 404].includes(x.status) });

    if (knownMeetingId) {
      r = http.get(`${BASE.userPortal}/api/v1/meetings/${knownMeetingId}`,
        { headers: hdrs, tags: { type: 'api_read' } });
      check(r, { 'GET /meetings/:id 200': (x) => [200, 403, 404].includes(x.status) });
    }
  });

  // ─── AI Assistant ──────────────────────────────────────────────────────────
  group('assistant', () => {
    // find_videos — не требует file_id
    const r = http.post(
      `${BASE.userPortal}/api/v1/assistant/chat`,
      JSON.stringify({ mode: 'find_videos', message: 'нагрузочное тестирование' }),
      { headers: hdrs, tags: { type: 'api_write' } },
    );
    // 200 — LLM есть, 503 — LLM не настроен; оба ок для НТ
    check(r, { 'POST /assistant/chat 200/503': (x) => [200, 503].includes(x.status) });
  });

  sleep(0.5 + Math.random() * 0.5);
}

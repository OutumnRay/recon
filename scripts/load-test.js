/**
 * Полноценное нагрузочное тестирование Recontext
 *
 * Запуск:
 *   k6 run scripts/load-test.js
 *   k6 run --env ADMIN_PASSWORD=secret scripts/load-test.js
 *   k6 run --env SCENARIO=smoke scripts/load-test.js
 *
 * Сценарии (SCENARIO=):
 *   smoke   — 1 пользователь, 1 минута (базовая проверка)
 *   load    — до 50 пользователей (дефолт)
 *   stress  — до 200 пользователей (поиск предела)
 *   soak    — 30 пользователей, 30 минут (проверка утечек)
 */

import http from 'k6/http';
import ws from 'k6/ws';
import { check, group, sleep, fail } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import { SharedArray } from 'k6/data';

// ─── Конфигурация ────────────────────────────────────────────────────────────

const BASE_URL  = __ENV.BASE_URL  || 'https://24recontext.ru';
const ADMIN_USER = __ENV.ADMIN_USER || 'admin@recontext.online';
const ADMIN_PASS = __ENV.ADMIN_PASS || 'admin123';

const SCENARIO  = __ENV.SCENARIO  || 'load';

// ─── Кастомные метрики ───────────────────────────────────────────────────────

const authErrors      = new Counter('auth_errors');
const meetingCreated  = new Counter('meetings_created');
const searchRequests  = new Counter('search_requests');
const fileListRequests = new Counter('file_list_requests');
const wsConnected     = new Counter('ws_connected');
const wsErrors        = new Counter('ws_errors');

const meetingDuration = new Trend('meeting_create_duration', true);
const searchDuration  = new Trend('search_duration', true);

const errorRate = new Rate('errors');

// ─── Сценарии нагрузки ───────────────────────────────────────────────────────

const SCENARIOS = {
  smoke: {
    executor: 'constant-vus',
    vus: 1,
    duration: '1m',
  },
  load: {
    executor: 'ramping-vus',
    stages: [
      { duration: '1m',  target: 10  },  // прогрев
      { duration: '3m',  target: 50  },  // нормальная нагрузка
      { duration: '2m',  target: 100 },  // повышенная нагрузка
      { duration: '1m',  target: 50  },  // снижение
      { duration: '30s', target: 0   },  // сброс
    ],
  },
  stress: {
    executor: 'ramping-vus',
    stages: [
      { duration: '2m',  target: 50  },
      { duration: '5m',  target: 100 },
      { duration: '2m',  target: 200 },
      { duration: '5m',  target: 200 },  // держим предел
      { duration: '2m',  target: 0   },
    ],
  },
  soak: {
    executor: 'constant-vus',
    vus: 30,
    duration: '30m',
  },
};

export const options = {
  scenarios: { main: SCENARIOS[SCENARIO] },
  thresholds: {
    // Глобальные
    http_req_duration:            ['p(95)<1000', 'p(99)<3000'],
    http_req_failed:              ['rate<0.05'],
    errors:                       ['rate<0.05'],

    // Авторизация критична — жёсткие пороги
    'http_req_duration{group:::auth}':     ['p(95)<500'],

    // Список встреч
    'http_req_duration{group:::meetings}': ['p(95)<1000'],

    // Поиск — может быть медленнее
    search_duration:              ['p(95)<3000'],

    // Создание встречи
    meeting_create_duration:      ['p(95)<2000'],
  },
};

// ─── Setup: один раз получаем токен и данные ─────────────────────────────────

export function setup() {
  const loginRes = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ username: ADMIN_USER, password: ADMIN_PASS }),
    { headers: { 'Content-Type': 'application/json' } },
  );

  const loginOk = check(loginRes, {
    'setup: login 200':    (r) => r.status === 200,
    'setup: has token':    (r) => r.json('token') !== undefined,
  });

  if (!loginOk) {
    fail(`Не удалось получить токен: ${loginRes.status} ${loginRes.body}`);
  }

  const token = loginRes.json('token');

  // Получаем список существующих встреч для read-only сценариев
  const meetingsRes = http.get(
    `${BASE_URL}/api/v1/meetings`,
    { headers: authHeaders(token) },
  );
  const meetingsBody = meetingsRes.status === 200 ? meetingsRes.json() : null;
  const meetings = (meetingsBody && meetingsBody.items) || [];
  const meetingIds = meetings.slice(0, 10).map((m) => m.id);

  // Получаем список файлов для read-only сценариев
  const filesRes = http.get(
    `${BASE_URL}/api/v1/files`,
    { headers: authHeaders(token) },
  );
  const filesBody = filesRes.status === 200 ? filesRes.json() : null;
  const files = (filesBody && (filesBody.items || filesBody.files)) || [];
  const fileIds = files.slice(0, 10).map((f) => f.id);

  return { token, meetingIds, fileIds };
}

// ─── Вспомогательные функции ─────────────────────────────────────────────────

function authHeaders(token) {
  return {
    'Content-Type':  'application/json',
    'Authorization': `Bearer ${token}`,
  };
}

function randomItem(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

function randomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

const SEARCH_QUERIES = [
  'бюджет',
  'задача',
  'проект',
  'встреча',
  'отчёт',
  'клиент',
  'договор',
  'сроки',
  'результат',
  'итог',
];

const MEETING_TITLES = [
  'k6 test встреча',
  'k6 нагрузочный тест',
  'k6 load test meeting',
];

// ─── Сценарии (функции для group) ────────────────────────────────────────────

function testHealth() {
  group('health', () => {
    const res = http.get(`${BASE_URL}/health`);
    const ok = check(res, {
      'health: status 200': (r) => r.status === 200,
    });
    if (!ok) errorRate.add(1);
    else errorRate.add(0);
  });
}

function testAuth(token) {
  group('auth', () => {
    // Проверка токена
    const checkRes = http.get(
      `${BASE_URL}/api/v1/auth/check`,
      { headers: authHeaders(token) },
    );
    const ok = check(checkRes, {
      'auth/check: 200': (r) => r.status === 200,
    });
    if (!ok) {
      authErrors.add(1);
      errorRate.add(1);
    } else {
      errorRate.add(0);
    }
  });
}

function testProfile(token) {
  group('profile', () => {
    const res = http.get(
      `${BASE_URL}/api/v1/profile`,
      { headers: authHeaders(token) },
    );
    const ok = check(res, {
      'profile: 200':      (r) => r.status === 200,
      'profile: has data': (r) => r.json('id') !== undefined || r.json('email') !== undefined,
    });
    errorRate.add(ok ? 0 : 1);
  });
}

function testMeetingsList(token) {
  group('meetings', () => {
    // Список без фильтров
    const allRes = http.get(
      `${BASE_URL}/api/v1/meetings`,
      { headers: authHeaders(token) },
    );
    check(allRes, {
      'meetings list: 200':      (r) => r.status === 200,
      'meetings list: has items': (r) => {
        const b = r.json();
        return b !== null && Array.isArray(b.items);
      },
    });
    errorRate.add(allRes.status === 200 ? 0 : 1);

    // С пагинацией
    const pageRes = http.get(
      `${BASE_URL}/api/v1/meetings?limit=10&offset=0`,
      { headers: authHeaders(token) },
    );
    check(pageRes, {
      'meetings paginated: 200': (r) => r.status === 200,
    });
  });
}

function testMeetingGet(token, meetingIds) {
  if (!meetingIds || meetingIds.length === 0) return;

  group('meeting_get', () => {
    const id  = randomItem(meetingIds);
    const res = http.get(
      `${BASE_URL}/api/v1/meetings/${id}`,
      { headers: authHeaders(token) },
    );
    check(res, {
      'meeting get: 200 or 404': (r) => r.status === 200 || r.status === 404,
    });
    errorRate.add(res.status >= 500 ? 1 : 0);
  });
}

function testMeetingCreate(token) {
  group('meeting_create', () => {
    const startTime = Date.now();
    const res = http.post(
      `${BASE_URL}/api/v1/meetings`,
      JSON.stringify({
        title:        `${randomItem(MEETING_TITLES)} ${__VU}-${__ITER}`,
        scheduled_at: new Date(Date.now() + 3600000).toISOString(),
        duration:     60,
        type:         'conference',
        recurrence:   'none',
      }),
      { headers: authHeaders(token) },
    );
    meetingDuration.add(Date.now() - startTime);

    const ok = check(res, {
      'meeting create: 200 or 201': (r) => r.status === 200 || r.status === 201,
      'meeting create: has id':     (r) => r.json('id') !== undefined,
    });

    if (ok) {
      meetingCreated.add(1);
      errorRate.add(0);

      // Удаляем созданную тестовую встречу (чтобы не засорять БД)
      const meetingId = res.json('id');
      if (meetingId) {
        http.del(
          `${BASE_URL}/api/v1/meetings/${meetingId}`,
          null,
          { headers: authHeaders(token) },
        );
      }
    } else {
      errorRate.add(1);
    }
  });
}

function testFilesList(token) {
  group('files', () => {
    fileListRequests.add(1);

    const res = http.get(
      `${BASE_URL}/api/v1/files`,
      { headers: authHeaders(token) },
    );
    check(res, {
      'files list: 200': (r) => r.status === 200,
    });
    errorRate.add(res.status === 200 ? 0 : 1);

    // Пагинация
    const pageRes = http.get(
      `${BASE_URL}/api/v1/files?limit=20&offset=0`,
      { headers: authHeaders(token) },
    );
    check(pageRes, {
      'files paginated: 200': (r) => r.status === 200,
    });
  });
}

function testFileGet(token, fileIds) {
  if (!fileIds || fileIds.length === 0) return;

  group('file_get', () => {
    const id  = randomItem(fileIds);
    const res = http.get(
      `${BASE_URL}/api/v1/files/${id}`,
      { headers: authHeaders(token) },
    );
    check(res, {
      'file get: not 500': (r) => r.status !== 500,
    });
    errorRate.add(res.status >= 500 ? 1 : 0);
  });
}

function testFileTranscript(token, fileIds) {
  if (!fileIds || fileIds.length === 0) return;

  group('file_transcript', () => {
    const id  = randomItem(fileIds);
    const res = http.get(
      `${BASE_URL}/api/v1/files/${id}/transcript`,
      { headers: authHeaders(token) },
    );
    check(res, {
      'transcript: not 500': (r) => r.status !== 500,
    });
    errorRate.add(res.status >= 500 ? 1 : 0);
  });
}

function testSearch(token) {
  group('search', () => {
    searchRequests.add(1);

    const q        = randomItem(SEARCH_QUERIES);
    const startTime = Date.now();

    const res = http.get(
      `${BASE_URL}/api/v1/search?query=${encodeURIComponent(q)}&page_size=10`,
      { headers: authHeaders(token) },
    );
    searchDuration.add(Date.now() - startTime);

    check(res, {
      'search: 200':      (r) => r.status === 200,
      'search: not slow': (r) => r.timings.duration < 5000,
    });
    errorRate.add(res.status === 200 ? 0 : 1);
  });
}

function testRAG(token) {
  group('rag', () => {
    // Проверка разрешений
    const permRes = http.get(
      `${BASE_URL}/api/v1/rag/permission`,
      { headers: authHeaders(token) },
    );
    check(permRes, {
      'rag permission: not 500': (r) => r.status !== 500,
    });

    // Статус RAG
    const statusRes = http.get(
      `${BASE_URL}/api/v1/rag/status`,
      { headers: authHeaders(token) },
    );
    check(statusRes, {
      'rag status: not 500': (r) => r.status !== 500,
    });

    // Семантический поиск (только если разрешён)
    if (permRes.status === 200) {
      const searchRes = http.post(
        `${BASE_URL}/api/v1/rag/search`,
        JSON.stringify({ query: randomItem(SEARCH_QUERIES), limit: 5 }),
        { headers: authHeaders(token) },
      );
      check(searchRes, {
        'rag search: not 500': (r) => r.status !== 500,
      });
      errorRate.add(searchRes.status >= 500 ? 1 : 0);
    }
  });
}

function testMyTasks(token) {
  group('tasks', () => {
    const res = http.get(
      `${BASE_URL}/api/v1/my-tasks`,
      { headers: authHeaders(token) },
    );
    check(res, {
      'my-tasks: 200': (r) => r.status === 200,
    });
    errorRate.add(res.status === 200 ? 0 : 1);
  });
}

function testDepartments(token) {
  group('departments', () => {
    const res = http.get(
      `${BASE_URL}/api/v1/departments`,
      { headers: authHeaders(token) },
    );
    check(res, {
      'departments: 200': (r) => r.status === 200,
    });
    errorRate.add(res.status === 200 ? 0 : 1);
  });
}

function testMeetingSubjects(token) {
  group('meeting_subjects', () => {
    const res = http.get(
      `${BASE_URL}/api/v1/meeting-subjects`,
      { headers: authHeaders(token) },
    );
    check(res, {
      'meeting-subjects: 200': (r) => r.status === 200,
    });
    errorRate.add(res.status === 200 ? 0 : 1);
  });
}

function testWebSocket(token, meetingIds) {
  if (!meetingIds || meetingIds.length === 0) return;

  // WS тестируем только в ~10% итераций чтобы не перегружать
  if (Math.random() > 0.1) return;

  group('websocket', () => {
    const id  = randomItem(meetingIds);
    const url = `${BASE_URL.replace('https://', 'wss://').replace('http://', 'ws://')}/api/v1/meetings/${id}/ws?token=${token}`;

    const response = ws.connect(url, {}, (socket) => {
      wsConnected.add(1);

      socket.on('open', () => {
        socket.send(JSON.stringify({ type: 'ping' }));
      });

      socket.on('message', () => {
        socket.close();
      });

      socket.on('error', (e) => {
        wsErrors.add(1);
      });

      // Закрываем через 3 секунды если нет ответа
      socket.setTimeout(() => socket.close(), 3000);
    });

    check(response, {
      'ws: connected': (r) => r && r.status === 101,
    });
  });
}

// ─── Главная функция (выполняется для каждого VU) ────────────────────────────

export default function (data) {
  const { token, meetingIds, fileIds } = data;

  // Распределяем нагрузку по типам поведения пользователей
  const userType = __VU % 5;

  // ── Общие проверки ──────────────────────────────────────────────────────────
  testHealth();
  testAuth(token);
  testProfile(token);
  sleep(0.5);

  // ── Паттерн поведения зависит от VU ────────────────────────────────────────

  if (userType === 0) {
    // Тип 1: Активный участник встреч (20% пользователей)
    testMeetingsList(token);
    sleep(0.3);
    testMeetingGet(token, meetingIds);
    sleep(0.3);
    testMeetingCreate(token);
    sleep(0.5);
    testSearch(token);
    sleep(0.5);
    testMyTasks(token);

  } else if (userType === 1) {
    // Тип 2: Работает с файлами (20% пользователей)
    testFilesList(token);
    sleep(0.3);
    testFileGet(token, fileIds);
    sleep(0.3);
    testFileTranscript(token, fileIds);
    sleep(0.5);
    testSearch(token);

  } else if (userType === 2) {
    // Тип 3: Аналитик (ищет и использует RAG) (20% пользователей)
    testSearch(token);
    sleep(0.5);
    testRAG(token);
    sleep(0.5);
    testMeetingsList(token);
    sleep(0.3);
    testMeetingSubjects(token);

  } else if (userType === 3) {
    // Тип 4: Просматривает структуру (20% пользователей)
    testMeetingsList(token);
    sleep(0.3);
    testDepartments(token);
    sleep(0.3);
    testMeetingSubjects(token);
    sleep(0.3);
    testFilesList(token);

  } else {
    // Тип 5: Типичный пользователь — всё понемногу (20% пользователей)
    testMeetingsList(token);
    sleep(0.3);
    testSearch(token);
    sleep(0.3);
    testProfile(token);
    sleep(0.3);
    testMyTasks(token);
    testWebSocket(token, meetingIds);
  }

  // Имитируем паузу между действиями пользователя
  sleep(randomInt(1, 3));
}

// ─── Teardown ─────────────────────────────────────────────────────────────────

export function teardown(data) {
  console.log(`Тестирование завершено. Сценарий: ${SCENARIO}`);
}

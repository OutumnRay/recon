#!/bin/bash
# Запуск всех нагрузочных тестов через Docker внутри recontext-network.
# k6 обращается к сервисам напрямую по именам Docker-контейнеров,
# без необходимости открывать дополнительные порты.
#
# Использование (запускать из корня проекта /opt/recontext):
#   chmod +x load-tests/run-all-docker.sh
#   ./load-tests/run-all-docker.sh
#
# Переопределить credentials:
#   ADMIN_PASS=mypass USER_PASS=mypass ./load-tests/run-all-docker.sh

set -e

NETWORK="${DOCKER_NETWORK:-recontext_recontext-network}"
RESULTS_DIR="$(pwd)/load-tests/results/$(date +%Y%m%d_%H%M%S)"
TESTS_DIR="$(pwd)/load-tests"

# Имена контейнеров внутри Docker-сети
MANAGING_HOST="${MANAGING_HOST:-managing-portal}"
MANAGING_PORT="${MANAGING_PORT:-8080}"
USER_HOST="${USER_HOST:-user-portal}"
USER_PORT="${USER_PORT:-8081}"

ADMIN_USER="${ADMIN_USER:-admin@recontext.online}"
ADMIN_PASS="${ADMIN_PASS:-admin123}"
USER_USER="${USER_USER:-user@recontext.online}"
USER_PASS="${USER_PASS:-user123}"

mkdir -p "$RESULTS_DIR"

K6_ENV="-e MANAGING_HOST=$MANAGING_HOST \
        -e MANAGING_PORT=$MANAGING_PORT \
        -e USER_HOST=$USER_HOST \
        -e USER_PORT=$USER_PORT \
        -e ADMIN_USER=$ADMIN_USER \
        -e ADMIN_PASS=$ADMIN_PASS \
        -e USER_USER=$USER_USER \
        -e USER_PASS=$USER_PASS"

echo "========================================================"
echo " Recontext Load Tests (Docker Network Mode)"
echo " Network:         $NETWORK"
echo " Managing Portal: http://$MANAGING_HOST:$MANAGING_PORT"
echo " User Portal:     http://$USER_HOST:$USER_PORT"
echo " Results:         $RESULTS_DIR"
echo "========================================================"

run_test() {
  local name="$1"
  local script="/load-tests/${2}"
  local out="/load-tests/results/$(basename $RESULTS_DIR)/${name}.json"

  echo ""
  echo ">>> [$name] Start: $(date)"

  if docker run --rm \
      --network "$NETWORK" \
      --user root \
      -v "$TESTS_DIR:/load-tests" \
      $K6_ENV \
      grafana/k6 run --out "json=$out" "$script"; then
    echo ">>> [$name] PASSED"
  else
    echo ">>> [$name] FAILED (см. $RESULTS_DIR/${name}.json)"
  fi

  echo ">>> [$name] End: $(date)"
}

# 1. Smoke — базовая проверка
run_test "smoke"    "smoke-test.js"

# 2. Health — только /health
run_test "health"   "health-check.js"

# 3. Managing Portal — детальный тест
run_test "managing" "managing-portal.js"

# 4. User Portal — детальный тест
run_test "user"     "user-portal.js"

# 5. Комбинированная нагрузка
run_test "load"     "load-test.js"

# 6. Полный набор сценариев
run_test "suite"    "full-suite.js"

echo ""
echo "========================================================"
echo " Все тесты завершены."
echo " Результаты: $RESULTS_DIR"
echo "========================================================"

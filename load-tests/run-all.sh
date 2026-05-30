#!/bin/bash
# Запуск всех нагрузочных тестов последовательно
# Использование: ./load-tests/run-all.sh <SERVER_IP> [MANAGING_PORT] [USER_PORT]
#
# Пример:
#   ./load-tests/run-all.sh 192.168.1.10
#   ./load-tests/run-all.sh 192.168.1.10 20080 20081

set -e

SERVER="${1:-localhost}"
MANAGING_PORT="${2:-20080}"
USER_PORT="${3:-20081}"
RESULTS_DIR="load-tests/results/$(date +%Y%m%d_%H%M%S)"

mkdir -p "$RESULTS_DIR"

K6_OPTS="-e SERVER=$SERVER -e MANAGING_PORT=$MANAGING_PORT -e USER_PORT=$USER_PORT"

echo "========================================================"
echo " Recontext Load Tests"
echo " Server:          $SERVER"
echo " Managing Portal: http://$SERVER:$MANAGING_PORT"
echo " User Portal:     http://$SERVER:$USER_PORT"
echo " Results:         $RESULTS_DIR"
echo "========================================================"

run_test() {
  local name="$1"
  local script="$2"
  local out="$RESULTS_DIR/${name}.json"

  echo ""
  echo ">>> [$name] Start: $(date)"
  if k6 run $K6_OPTS --out "json=$out" "$script"; then
    echo ">>> [$name] PASSED"
  else
    echo ">>> [$name] FAILED (see $out)"
  fi
  echo ">>> [$name] End: $(date)"
}

# 1. Smoke — быстрая проверка живучести
run_test "smoke"   "load-tests/smoke-test.js"

# 2. Health check — только /health эндпоинты
run_test "health"  "load-tests/health-check.js"

# 3. Managing Portal — детальный тест
run_test "managing" "load-tests/managing-portal.js"

# 4. User Portal — детальный тест
run_test "user"    "load-tests/user-portal.js"

# 5. Load Test — комбинированная нагрузка
run_test "load"    "load-tests/load-test.js"

# 6. Full Suite — параллельные сценарии
run_test "suite"   "load-tests/full-suite.js"

echo ""
echo "========================================================"
echo " All tests done. Results: $RESULTS_DIR"
echo "========================================================"

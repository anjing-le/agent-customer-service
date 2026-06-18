#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

PORT="${ANJING_DEMO_PORT:-10002}"
BASE_URL="${ANJING_DEMO_BASE_URL:-http://localhost:${PORT}}"
LOG_FILE="$(mktemp -t agent-customer-service-demo.XXXXXX.log)"
SERVER_PID=""

cleanup() {
  if [[ -n "$SERVER_PID" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
  rm -f "$LOG_FILE"
}

fail() {
  echo "demo-classroom-local: $*" >&2
  if [[ -f "$LOG_FILE" ]]; then
    echo "demo-classroom-local: server log tail" >&2
    tail -n 80 "$LOG_FILE" >&2 || true
  fi
  exit 1
}

trap cleanup EXIT INT TERM

if command -v lsof >/dev/null 2>&1 && lsof -iTCP:"$PORT" -sTCP:LISTEN -n -P >/dev/null 2>&1; then
  fail "port ${PORT} is already in use; stop the existing service or set ANJING_DEMO_PORT"
fi

echo "demo-classroom-local: building console"
pnpm build:console

echo "demo-classroom-local: starting platform-all on ${BASE_URL}"
ANJING_ADDR=":${PORT}" ANJING_DATABASE_URL="${ANJING_DEMO_DATABASE_URL:-}" go run ./cmd/platform-all >"$LOG_FILE" 2>&1 &
SERVER_PID="$!"

for _ in $(seq 1 30); do
  if curl -fsS "${BASE_URL}/healthz" >/dev/null 2>&1; then
    echo "demo-classroom-local: service is ready"
    ./scripts/demo-classroom-smoke.sh "$BASE_URL"
    echo "demo-classroom-local: ok"
    exit 0
  fi
  if ! kill -0 "$SERVER_PID" 2>/dev/null; then
    fail "platform-all exited before health check was ready"
  fi
  sleep 1
done

fail "health check timed out for ${BASE_URL}"

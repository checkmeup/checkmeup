#!/usr/bin/env bash
# Driver for the run-checkmeup skill: launches apps/api + apps/web,
# waits for both to be reachable, then exercises a real signed-cookie
# flow (sign-up -> /me -> create a cron monitor -> list it back) purely
# over curl. No browser driver — see SKILL.md for why.
#
# Usage:
#   .claude/skills/run-checkmeup/smoke.sh start   # build + launch both apps, wait for health
#   .claude/skills/run-checkmeup/smoke.sh check   # run the curl flow against already-running apps
#   .claude/skills/run-checkmeup/smoke.sh stop    # kill both apps

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
PID_DIR="/tmp/checkmeup-smoke"
API_LOG="$PID_DIR/api.log"
WEB_LOG="$PID_DIR/web.log"
API_BIN="$PID_DIR/api"

mkdir -p "$PID_DIR"

wait_for() {
  local url="$1" name="$2" tries=30
  until curl -s -o /dev/null "$url"; do
    tries=$((tries - 1))
    if [ "$tries" -le 0 ]; then
      echo "FAIL: $name never became reachable at $url" >&2
      return 1
    fi
    sleep 1
  done
}

start() {
  echo "== migrating db =="
  (cd "$ROOT" && make migrate)

  echo "== building api =="
  (cd "$ROOT/apps/api" && go build -o "$API_BIN" ./cmd/api)

  echo "== launching api =="
  (cd "$ROOT/apps/api" && set -a && source .env && set +a && exec "$API_BIN") >"$API_LOG" 2>&1 &
  echo $! >"$PID_DIR/api.pid"

  echo "== launching web (vite) =="
  (cd "$ROOT/apps/web" && exec bunx vite) >"$WEB_LOG" 2>&1 &
  echo $! >"$PID_DIR/web.pid"

  wait_for "http://localhost:8080/api/v1/health" "api" || { tail -n 40 "$API_LOG"; exit 1; }
  wait_for "http://localhost:5173/" "web" || { tail -n 40 "$WEB_LOG"; exit 1; }
  echo "api: http://localhost:8080  (log: $API_LOG)"
  echo "web: http://localhost:5173  (log: $WEB_LOG)"
}

check() {
  local cookies="$PID_DIR/cookies.txt"
  local email="smoke-$(date +%s)@example.com"
  rm -f "$cookies"

  echo "== health =="
  curl -sf http://localhost:8080/api/v1/health; echo

  echo "== sign-up ($email) =="
  curl -sf -c "$cookies" -X POST http://localhost:8080/api/v1/auth/sign-up \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$email\",\"password\":\"testpass1234\",\"acceptedTerms\":true}"
  echo

  echo "== /me =="
  curl -sf -b "$cookies" http://localhost:8080/api/v1/me; echo

  echo "== create cron monitor =="
  local monitor
  monitor=$(curl -sf -b "$cookies" -X POST http://localhost:8080/api/v1/monitors/cron \
    -H "Content-Type: application/json" \
    -d '{"name":"smoke-test-job","schedule":"every 1h","gracePeriodMins":5,"maxAlertsPerIncident":3,"alertAfterNFailures":1}')
  echo "$monitor"

  echo "== list cron monitors (expect the one just created) =="
  curl -sf -b "$cookies" http://localhost:8080/api/v1/monitors/cron; echo

  echo "== web dev server proxies /api correctly =="
  curl -sf http://localhost:5173/api/v1/health; echo

  echo "== web dev server serves the SPA shell =="
  curl -sf http://localhost:5173/ | grep -q '<div id="app">' && echo "OK: #app root present"

  echo
  echo "PASS"
}

stop() {
  for f in "$PID_DIR/api.pid" "$PID_DIR/web.pid"; do
    [ -f "$f" ] || continue
    pid=$(cat "$f")
    kill "$pid" 2>/dev/null || true
    rm -f "$f"
  done
  echo "stopped"
}

case "${1:-}" in
  start) start ;;
  check) check ;;
  stop) stop ;;
  *) echo "usage: $0 {start|check|stop}" >&2; exit 1 ;;
esac

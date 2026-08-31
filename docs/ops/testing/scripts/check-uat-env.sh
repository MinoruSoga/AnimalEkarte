#!/usr/bin/env bash
# UAT environment readiness check — no secrets printed.
set -euo pipefail

readonly EXIT_NOT_READY=1
ROOT="$(cd "$(dirname "$0")/../../../.." && pwd)"
cd "$ROOT"

ok=0
fail=0
pass() { echo "PASS  $1"; ok=$((ok+1)); }
bad()  { echo "FAIL  $1"; fail=$((fail+1)); }
info() { echo "INFO  $1"; }

echo "=== UAT env check ==="
echo "root: $ROOT"
echo "time: $(date -Iseconds)"

# Keep every request bounded. On transport failure, report curl's exit code instead
# of appending a second synthetic HTTP code to curl's output.
check_http() {
  local label=$1
  local url=$2
  local code
  local curl_status

  if code=$(curl --silent --show-error --output /dev/null --write-out '%{http_code}' \
      --connect-timeout 3 --max-time 10 "$url"); then
    if [ "$code" = "200" ]; then
      pass "$label HTTP $code"
    else
      bad "$label HTTP $code"
    fi
  else
    curl_status=$?
    bad "$label request failed (curl exit $curl_status)"
  fi
}

check_http "frontend :3003" "http://localhost:3003/"
check_http "backend  :8080/health" "http://localhost:8080/health"

# env files
if [ -f .env.local ]; then pass ".env.local exists"; else bad ".env.local missing"; fi

# Read only the two required dotenv keys. Never source the file, because sourcing
# executes arbitrary shell content. Existing exported values take precedence.
load_uat_env() {
  local name
  local value

  [ -f .env.local ] || return 0
  while IFS= read -r -d '' name && IFS= read -r -d '' value; do
    case "$name" in
      E2E_LOGIN_EMAIL)
        if [ -z "${E2E_LOGIN_EMAIL+x}" ]; then printf -v E2E_LOGIN_EMAIL '%s' "$value"; fi
        ;;
      E2E_LOGIN_PASSWORD)
        if [ -z "${E2E_LOGIN_PASSWORD+x}" ]; then printf -v E2E_LOGIN_PASSWORD '%s' "$value"; fi
        ;;
    esac
  done < <(python3 - .env.local <<'PY'
from pathlib import Path
import json
import sys

wanted = {"E2E_LOGIN_EMAIL", "E2E_LOGIN_PASSWORD"}
for raw_line in Path(sys.argv[1]).read_text(encoding="utf-8").splitlines():
    line = raw_line.strip()
    if not line or line.startswith("#") or "=" not in line:
        continue
    name, raw_value = line.split("=", 1)
    name = name.strip()
    if name.startswith("export "):
        name = name[7:].strip()
    if name not in wanted:
        continue
    value = raw_value.strip()
    if len(value) >= 2 and value[0] == value[-1] == "'":
        value = value[1:-1]
    elif len(value) >= 2 and value[0] == value[-1] == '"':
        try:
            value = json.loads(value)
        except json.JSONDecodeError:
            # Keep malformed quoted input intact; the login check will fail safely.
            pass
    sys.stdout.buffer.write(name.encode("utf-8") + b"\0")
    sys.stdout.buffer.write(value.encode("utf-8") + b"\0")
PY
  )
}
load_uat_env

if [ -n "${E2E_LOGIN_EMAIL:-}" ]; then pass "E2E_LOGIN_EMAIL set"; else bad "E2E_LOGIN_EMAIL unset"; fi
if [ -n "${E2E_LOGIN_PASSWORD:-}" ]; then pass "E2E_LOGIN_PASSWORD set"; else bad "E2E_LOGIN_PASSWORD unset"; fi

# LIFF mock hints (local)
if [ "${LIFF_MOCK:-}" = "true" ] || grep -q 'LIFF_MOCK: "true"' docker-compose.yml 2>/dev/null; then
  pass "LIFF_MOCK expected true for local (compose or env)"
else
  bad "LIFF_MOCK not clearly true — required for local LIFF scenarios"
fi
if grep -q -- '- VITE_LIFF_MOCK=true' docker-compose.yml 2>/dev/null; then
  pass "compose sets VITE_LIFF_MOCK=true"
elif [ -f frontend/.env.local ] && grep -q '^[[:space:]]*VITE_LIFF_MOCK=true[[:space:]]*$' frontend/.env.local; then
  pass "frontend/.env.local sets VITE_LIFF_MOCK=true"
else
  bad "VITE_LIFF_MOCK=true not found — required for local LIFF scenarios"
fi

# Compose health is a documented prerequisite. Inspect status only; do not print
# container environment or logs.
check_container_health() {
  local service=$1
  local container_id
  local health

  container_id=$(docker compose ps -q "$service" 2>/dev/null || true)
  container_id=${container_id%%$'\n'*}
  if [ -z "$container_id" ]; then
    bad "compose $service container not running"
    return 1
  fi
  health=$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' \
    "$container_id" 2>/dev/null || true)
  if [ "$health" = "healthy" ]; then
    pass "compose $service container healthy"
    return 0
  fi
  bad "compose $service container health: ${health:-unknown}"
  return 1
}

frontend_healthy=false
if command -v docker >/dev/null 2>&1; then
  if check_container_health db; then :; fi
  if check_container_health backend; then :; fi
  if check_container_health frontend; then frontend_healthy=true; fi
else
  bad "docker CLI unavailable; cannot verify db/backend/frontend health"
fi

# API login. JSON encoding is delegated to Python and credentials travel through
# stdin, never through process arguments. Temporary files are private and removed
# on every exit path.
login_payload=''
login_response=''
cleanup() {
  [ -z "$login_payload" ] || rm -f -- "$login_payload"
  [ -z "$login_response" ] || rm -f -- "$login_response"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

if [ -n "${E2E_LOGIN_EMAIL:-}" ] && [ -n "${E2E_LOGIN_PASSWORD:-}" ]; then
  old_umask=$(umask)
  umask 077
  login_payload=$(mktemp "${TMPDIR:-/tmp}/uat-env-login-payload.XXXXXX")
  login_response=$(mktemp "${TMPDIR:-/tmp}/uat-env-login-response.XXXXXX")
  umask "$old_umask"
  chmod 0600 "$login_payload" "$login_response"

  printf '%s\0%s\0' "$E2E_LOGIN_EMAIL" "$E2E_LOGIN_PASSWORD" | \
    python3 -c 'import json, sys; email, password, _ = sys.stdin.buffer.read().decode("utf-8").split("\0", 2); json.dump({"email": email, "password": password}, sys.stdout)' \
    >"$login_payload"

  login_code=''
  if login_code=$(curl --silent --show-error --output "$login_response" --write-out '%{http_code}' \
      --connect-timeout 5 --max-time 15 \
      --request POST http://localhost:8080/api/v1/login \
      --header 'Content-Type: application/json' \
      --header 'X-Requested-With: XMLHttpRequest' \
      --header 'Origin: http://localhost:3003' \
      --data-binary @- <"$login_payload"); then
    if [ "$login_code" = "200" ]; then
      pass "POST /api/v1/login HTTP $login_code"
    else
      bad "POST /api/v1/login HTTP $login_code"
    fi
  else
    curl_status=$?
    bad "POST /api/v1/login request failed (curl exit $curl_status)"
  fi
fi

# Playwright is installed in the frontend container's named node_modules volume.
# Check the declared @playwright/test dependency in that container, not the host.
playwright_ready=false
if [ "$frontend_healthy" = true ] && \
    docker compose exec -T frontend node -e "require.resolve('@playwright/test/package.json')" >/dev/null 2>&1; then
  playwright_ready=true
  pass "frontend container has @playwright/test"
else
  info "frontend container @playwright/test unavailable"
fi

# Chrome CDP
chrome_ready=false
if command -v lsof >/dev/null 2>&1 && lsof -iTCP:9222 -sTCP:LISTEN >/dev/null 2>&1; then
  chrome_ready=true
  pass "Chrome remote debugging :9222 listening"
else
  info "Chrome :9222 not listening"
fi
if [ "$chrome_ready" = true ] || [ "$playwright_ready" = true ]; then
  pass "browser route ready (Chrome CDP or container Playwright)"
else
  bad "no browser route ready; start Chrome :9222 or prepare container Playwright"
fi

# Docs
for f in \
  docs/ops/testing/TEST_ARCHITECTURE.md \
  docs/ops/testing/UAT-ENV-SETUP.md \
  docs/ops/testing/scenarios/FIELD-LEVEL-PROTOCOL.md \
  docs/ops/testing/scenarios/FORM-FIELD-INVENTORY.md \
  docs/ops/testing/scenarios/README.md
do
  if [ -f "$f" ]; then pass "doc $f"; else bad "doc missing $f"; fi
done

echo "=== summary: PASS=$ok FAIL=$fail ==="
if [ "$fail" -gt 0 ]; then
  echo "UAT env NOT ready (exit $EXIT_NOT_READY)"
  exit "$EXIT_NOT_READY"
fi
echo "UAT env ready"
exit 0

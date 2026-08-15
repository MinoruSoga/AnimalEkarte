#!/usr/bin/env bash
# UAT environment readiness check — no secrets printed.
set -euo pipefail
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

# FE / BE HTTP
fe_code=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 3 http://localhost:3003/ || echo "000")
if [ "$fe_code" = "200" ]; then pass "frontend :3003 HTTP $fe_code"; else bad "frontend :3003 HTTP $fe_code"; fi

be_code=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 3 http://localhost:8080/health || echo "000")
if [ "$be_code" = "200" ]; then pass "backend  :8080/health HTTP $be_code"; else bad "backend  :8080/health HTTP $be_code"; fi

# env files
if [ -f .env.local ]; then pass ".env.local exists"; else bad ".env.local missing"; fi

# load env without printing values
if [ -f .env.local ]; then
  set -a
  # shellcheck disable=SC1091
  source .env.local
  set +a
fi

if [ -n "${E2E_LOGIN_EMAIL:-}" ]; then pass "E2E_LOGIN_EMAIL set (len=${#E2E_LOGIN_EMAIL})"; else bad "E2E_LOGIN_EMAIL unset"; fi
if [ -n "${E2E_LOGIN_PASSWORD:-}" ]; then pass "E2E_LOGIN_PASSWORD set (len=${#E2E_LOGIN_PASSWORD})"; else bad "E2E_LOGIN_PASSWORD unset"; fi

# LIFF mock hints (local)
if [ "${LIFF_MOCK:-}" = "true" ] || grep -q 'LIFF_MOCK: "true"' docker-compose.yml 2>/dev/null; then
  pass "LIFF_MOCK expected true for local (compose or env)"
else
  info "LIFF_MOCK not clearly true — required for local LIFF scenarios"
fi
if [ -f frontend/.env.local ] && grep -q 'VITE_LIFF_MOCK' frontend/.env.local; then
  pass "frontend/.env.local has VITE_LIFF_MOCK"
else
  info "frontend/.env.local VITE_LIFF_MOCK not found (compose may still inject)"
fi

# API login
if [ -n "${E2E_LOGIN_EMAIL:-}" ] && [ -n "${E2E_LOGIN_PASSWORD:-}" ]; then
  login_code=$(curl -s -o /tmp/uat-env-login.json -w "%{http_code}" --connect-timeout 5 \
    -X POST http://localhost:8080/api/v1/login \
    -H 'Content-Type: application/json' \
    -H 'X-Requested-With: XMLHttpRequest' \
    -H 'Origin: http://localhost:3003' \
    -d "{\"email\":\"${E2E_LOGIN_EMAIL}\",\"password\":\"${E2E_LOGIN_PASSWORD}\"}" || echo "000")
  if [ "$login_code" = "200" ]; then pass "POST /api/v1/login HTTP $login_code"; else bad "POST /api/v1/login HTTP $login_code"; fi
  rm -f /tmp/uat-env-login.json
fi

# Playwright package
if [ -f frontend/node_modules/playwright/package.json ]; then
  pass "frontend playwright package present"
else
  info "frontend playwright package missing (needed for scripted UAT)"
fi

# Chrome CDP
if lsof -iTCP:9222 -sTCP:LISTEN >/dev/null 2>&1; then
  pass "Chrome remote debugging :9222 listening"
else
  info "Chrome :9222 not listening — start for chrome-devtools MCP (see UAT-ENV-SETUP.md §4A)"
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
  echo "UAT env NOT ready"
  exit 1
fi
echo "UAT env ready (browser path: start Chrome :9222 if using DevTools MCP)"
exit 0

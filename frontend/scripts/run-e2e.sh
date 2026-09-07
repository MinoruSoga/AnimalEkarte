#!/bin/sh
# E2E test runner using the official Playwright Docker image.
#
# Why Docker: the project frontend container is Alpine-based and cannot run
# Playwright's Chromium. The official playwright jammy image has Chromium 1223.
#
# Why isolated /test workdir: the host node_modules has playwright-core pinned
# to chromium-1217 (pnpm-lock.yaml). Mounting that dir causes a browser-version
# mismatch with the v1.60.0-jammy image (which ships chromium-1223).
# Copying only the spec/config files and running fresh npm install avoids this.
#
# Usage:
#   ./scripts/run-e2e.sh                              # all specs
#   ./scripts/run-e2e.sh e2e/owners-search.spec.ts   # specific file
#   ./scripts/run-e2e.sh --auth-smoke                # CI と同じ auth smoke
#   ./scripts/run-e2e.sh --clinical                  # clinical allowlist + disposable clinic
#   ./scripts/run-e2e.sh --headed                     # headed mode (requires display)
#   ./scripts/run-e2e.sh e2e/owners-search.spec.ts --headed --timeout=30000
#
# Prerequisites:
#   app must be reachable from the host at http://localhost:3003
#   --clinical requires APP_ENV=test on the backend container and E2E_LOGIN_PASSWORD
#
# Note: On Apple Silicon (arm64), use this Docker script as the primary path.
# The official Playwright image can launch Chromium on linux/arm64.

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
FRONTEND_DIR="$(dirname "$SCRIPT_DIR")"
REPO_ROOT="$(dirname "$FRONTEND_DIR")"
BASE_URL="${PLAYWRIGHT_TEST_BASE_URL:-http://host.docker.internal:3003}"

CLINICAL_SPECS="e2e/clinical-flows.spec.ts e2e/clinical-smoke.spec.ts e2e/medical-records-create.spec.ts e2e/medical-records-patient-search.spec.ts e2e/medical-records-pagination-sort.spec.ts e2e/examinations-flow.spec.ts e2e/vaccinations-flow.spec.ts e2e/checkups-flow.spec.ts e2e/hospitalization-flow.spec.ts e2e/estimates-flow.spec.ts"

compose_backend() {
  docker compose -f "${REPO_ROOT}/docker-compose.yml" --project-directory "${REPO_ROOT}" exec -T "$@"
}

MODE=""
case "${1:-}" in
  --clinical)
    MODE="clinical"
    shift
    ;;
  --auth-smoke)
    MODE="auth-smoke"
    shift
    ;;
esac

CLINIC_ID=""
teardown_clinical_fixture() {
  if [ -n "$CLINIC_ID" ]; then
    compose_backend backend go run ./cmd/clinical-e2e-fixture teardown --clinic-id="$CLINIC_ID"
    CLINIC_ID=""
  fi
}

if [ "$MODE" = "clinical" ]; then
  if [ -z "${E2E_LOGIN_PASSWORD:-}" ]; then
    echo "run-e2e.sh: E2E_LOGIN_PASSWORD is required for --clinical" >&2
    exit 1
  fi
  BACKEND_APP_ENV="$(compose_backend backend printenv APP_ENV)"
  if [ "$BACKEND_APP_ENV" != "test" ]; then
    echo "run-e2e.sh: backend APP_ENV must be test for --clinical" >&2
    exit 1
  fi
  case "$BASE_URL" in
    http://localhost:3003|http://localhost:3003/*|http://127.0.0.1:3003|http://127.0.0.1:3003/*|http://host.docker.internal:3003|http://host.docker.internal:3003/*)
      ;;
    *)
      echo "run-e2e.sh: PLAYWRIGHT_TEST_BASE_URL must be the local compose frontend" >&2
      exit 1
      ;;
  esac
  FIXTURE_JSON="$(compose_backend -e E2E_LOGIN_PASSWORD backend go run ./cmd/clinical-e2e-fixture setup)"
  CLINIC_ID="$(printf '%s' "$FIXTURE_JSON" | sed -n 's/.*"clinicId":\([0-9][0-9]*\).*/\1/p')"
  if [ -z "$CLINIC_ID" ] || [ "$CLINIC_ID" = "1" ] || [ "$CLINIC_ID" = "2" ]; then
    echo "run-e2e.sh: clinical fixture clinic id is missing or reserved" >&2
    exit 1
  fi
  export APP_ENV=test
  export E2E_CLINICAL_FIXTURE="$FIXTURE_JSON"
  export E2E_CLINICAL_TEARDOWN=registered
  export E2E_LOGIN_EMAIL="e2e-clinical-${CLINIC_ID}@example.test"
  set -- $CLINICAL_SPECS "$@"
fi

if [ "$MODE" = "auth-smoke" ]; then
  set -- e2e/auth-flows.spec.ts "$@"
fi

# Forward host auth env into the container only when set (name-only -e; no =value on argv).
DOCKER_ENV="-e PLAYWRIGHT_TEST_BASE_URL=${BASE_URL}"
if [ -n "${E2E_LOGIN_EMAIL:-}" ]; then DOCKER_ENV="$DOCKER_ENV -e E2E_LOGIN_EMAIL"; fi
if [ -n "${E2E_LOGIN_PASSWORD:-}" ]; then DOCKER_ENV="$DOCKER_ENV -e E2E_LOGIN_PASSWORD"; fi
if [ -n "${E2E_AUTH_STATE_PATH:-}" ]; then DOCKER_ENV="$DOCKER_ENV -e E2E_AUTH_STATE_PATH"; fi
if [ -n "${APP_ENV:-}" ]; then DOCKER_ENV="$DOCKER_ENV -e APP_ENV"; fi
if [ -n "${E2E_CLINICAL_FIXTURE:-}" ]; then DOCKER_ENV="$DOCKER_ENV -e E2E_CLINICAL_FIXTURE"; fi
if [ -n "${E2E_CLINICAL_TEARDOWN:-}" ]; then DOCKER_ENV="$DOCKER_ENV -e E2E_CLINICAL_TEARDOWN"; fi

# All args passed through safely as positional params to sh -c via -- "$@".
# Single-quoted sh -c command prevents host-side shell expansion (injection-safe).
# shellcheck disable=SC2086 # intentional for DOCKER_ENV flag list only
PLAYWRIGHT_STATUS=0
docker run --rm \
  --add-host=host.docker.internal:host-gateway \
  $DOCKER_ENV \
  -v "${FRONTEND_DIR}/e2e:/test/e2e:ro" \
  -v "${FRONTEND_DIR}/playwright.config.ts:/test/playwright.config.ts:ro" \
  --workdir /test \
  mcr.microsoft.com/playwright:v1.60.0-jammy \
  sh -c 'npm install @playwright/test@1.60.0 --ignore-scripts --silent && node_modules/.bin/playwright test --reporter=list "$@"' \
  -- "$@" || PLAYWRIGHT_STATUS=$?

TEARDOWN_STATUS=0
if [ "$MODE" = "clinical" ]; then
  teardown_clinical_fixture || TEARDOWN_STATUS=$?
fi
if [ "$TEARDOWN_STATUS" -ne 0 ]; then
  exit "$TEARDOWN_STATUS"
fi
exit "$PLAYWRIGHT_STATUS"

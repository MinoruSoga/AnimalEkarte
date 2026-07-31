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
#   ./scripts/run-e2e.sh --headed                     # headed mode (requires display)
#   ./scripts/run-e2e.sh e2e/owners-search.spec.ts --headed --timeout=30000
#
# Prerequisites:
#   app must be reachable from the host at http://localhost:3003
#
# Note: On Apple Silicon (arm64), use this Docker script as the primary path.
# The official Playwright image can launch Chromium on linux/arm64.

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
FRONTEND_DIR="$(dirname "$SCRIPT_DIR")"
BASE_URL="${PLAYWRIGHT_TEST_BASE_URL:-http://host.docker.internal:3003}"

# Forward host auth env into the container only when set (name-only -e; no =value on argv).
DOCKER_ENV="-e PLAYWRIGHT_TEST_BASE_URL=${BASE_URL}"
if [ -n "${E2E_LOGIN_EMAIL:-}" ]; then DOCKER_ENV="$DOCKER_ENV -e E2E_LOGIN_EMAIL"; fi
if [ -n "${E2E_LOGIN_PASSWORD:-}" ]; then DOCKER_ENV="$DOCKER_ENV -e E2E_LOGIN_PASSWORD"; fi
if [ -n "${E2E_AUTH_STATE_PATH:-}" ]; then DOCKER_ENV="$DOCKER_ENV -e E2E_AUTH_STATE_PATH"; fi

# All args passed through safely as positional params to sh -c via -- "$@".
# Single-quoted sh -c command prevents host-side shell expansion (injection-safe).
# shellcheck disable=SC2086 # intentional for DOCKER_ENV flag list only
docker run --rm \
  --add-host=host.docker.internal:host-gateway \
  $DOCKER_ENV \
  -v "${FRONTEND_DIR}/e2e:/test/e2e:ro" \
  -v "${FRONTEND_DIR}/playwright.config.ts:/test/playwright.config.ts:ro" \
  --workdir /test \
  mcr.microsoft.com/playwright:v1.60.0-jammy \
  sh -c 'npm install @playwright/test@1.60.0 --ignore-scripts --silent && node_modules/.bin/playwright test --reporter=list "$@"' \
  -- "$@"

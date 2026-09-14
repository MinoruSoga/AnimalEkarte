#!/usr/bin/env bash
# P4-5(試行10): ECS `animalekarte-stg-migrate` one-shot task 相当の置換。
# Cloudflare Worker の POST /_internal/migrate を叩き、Container内 `/app/migrate` の
# exec結果(exit code)を検証する。ECS版(.github/workflows/backend-deploy.yml)と同じ
# 「exit code非0なら abort(非ゼロ終了)」契約をローカル/CIから再現する。
#
# 前提:
#   - MIGRATE_RUN_SECRET 環境変数(wrangler secret putで投入した値と同一)
#   - WORKER_URL 環境変数(省略時は試行9のworkers.dev URLを既定値とする)
#
# 使い方:
#   MIGRATE_RUN_SECRET=xxxx ./infra/scripts/cf-run-migrate.sh

set -euo pipefail

WORKER_URL="${WORKER_URL:-https://animalekarte-stg-api.baritech-soga.workers.dev}"
ENDPOINT="${WORKER_URL}/_internal/migrate"

if [[ -z "${MIGRATE_RUN_SECRET:-}" ]]; then
  echo "::error::MIGRATE_RUN_SECRET が未設定です" >&2
  exit 1
fi

# curl timeouts: self-test is immediate; migrate must exceed Worker
# AnimalEkarteApiContainer.MIGRATE_TIMEOUT_MS (240s) with margin.
SELFTEST_TIMEOUT=15
MIGRATE_TIMEOUT=270

# Harness Improvement Feedback P0: 認証なしアクセスの self-test。
# 401以外が返る場合は認証が機能していない可能性が高いため即中断する。
echo "==> self-test: unauthenticated request should be rejected"
UNAUTH_CODE=$(curl -s --max-time "${SELFTEST_TIMEOUT}" -o /dev/null -w '%{http_code}' -X POST "${ENDPOINT}")
if [[ "${UNAUTH_CODE}" != "401" ]]; then
  echo "::error::認証なしリクエストが ${UNAUTH_CODE} を返しました(401を期待)。認証が機能していない可能性があるため中断します" >&2
  exit 1
fi
echo "    OK (401)"

# After Container image replace, first migrate can fail while the DO/container
# finishes booting (bare migrate_exec_failed, no exitCode). Retry with backoff.
MAX_ATTEMPTS="${MIGRATE_MAX_ATTEMPTS:-3}"
SLEEP_SECS="${MIGRATE_RETRY_SLEEP_SECS:-20}"

attempt=1
while [[ "${attempt}" -le "${MAX_ATTEMPTS}" ]]; do
  echo "==> POST ${ENDPOINT} (attempt ${attempt}/${MAX_ATTEMPTS})"
  RESPONSE=$(curl -s --max-time "${MIGRATE_TIMEOUT}" -w '\n%{http_code}' -X POST "${ENDPOINT}" \
    -H "Authorization: Bearer ${MIGRATE_RUN_SECRET}")
  HTTP_CODE=$(echo "${RESPONSE}" | tail -n1)
  BODY=$(echo "${RESPONSE}" | sed '$d')

  echo "--- Migration response (HTTP ${HTTP_CODE}) ---"
  echo "${BODY}" | jq . 2>/dev/null || echo "${BODY}"

  EXIT_CODE=$(echo "${BODY}" | jq -r '.exitCode // "unknown"' 2>/dev/null || echo "unknown")
  if [[ "${HTTP_CODE}" == "200" && "${EXIT_CODE}" == "0" ]]; then
    echo "==> Migration succeeded (exitCode 0)"
    exit 0
  fi

  # Non-zero migrate exit is deterministic; do not retry application failures.
  if [[ "${EXIT_CODE}" != "unknown" && "${EXIT_CODE}" != "null" && "${EXIT_CODE}" != "0" ]]; then
    echo "::error::Migration failed (HTTP ${HTTP_CODE}, exitCode ${EXIT_CODE})" >&2
    exit 1
  fi

  if [[ "${attempt}" -eq "${MAX_ATTEMPTS}" ]]; then
    echo "::error::Migration failed after ${MAX_ATTEMPTS} attempts (HTTP ${HTTP_CODE}, exitCode ${EXIT_CODE})" >&2
    exit 1
  fi
  echo "==> Transient migrate failure; sleeping ${SLEEP_SECS}s before retry"
  sleep "${SLEEP_SECS}"
  attempt=$((attempt + 1))
done

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

# code-reviewer指摘 MEDIUM: curlにタイムアウト無指定だとネットワーク異常時にスクリプトが
# 無期限にハングする。self-testは即応するはずなので短め、本実行はWorker側exec timeout
# (index.ts MIGRATE_TIMEOUT_MS=120s)より長めのマージンを持たせる。
SELFTEST_TIMEOUT=15
MIGRATE_TIMEOUT=150

# Harness Improvement Feedback P0: 認証なしアクセスの self-test。
# 401以外が返る場合は認証が機能していない可能性が高いため即中断する。
echo "==> self-test: unauthenticated request should be rejected"
UNAUTH_CODE=$(curl -s --max-time "${SELFTEST_TIMEOUT}" -o /dev/null -w '%{http_code}' -X POST "${ENDPOINT}")
if [[ "${UNAUTH_CODE}" != "401" ]]; then
  echo "::error::認証なしリクエストが ${UNAUTH_CODE} を返しました(401を期待)。認証が機能していない可能性があるため中断します" >&2
  exit 1
fi
echo "    OK (401)"

echo "==> POST ${ENDPOINT}"
RESPONSE=$(curl -s --max-time "${MIGRATE_TIMEOUT}" -w '\n%{http_code}' -X POST "${ENDPOINT}" \
  -H "Authorization: Bearer ${MIGRATE_RUN_SECRET}")
HTTP_CODE=$(echo "${RESPONSE}" | tail -n1)
BODY=$(echo "${RESPONSE}" | sed '$d')

echo "--- Migration response (HTTP ${HTTP_CODE}) ---"
# code-reviewer指摘 LOW: python3依存はCI環境で不確実なため jq に統一(このリポジトリの
# 他スクリプトはpsql/aws/pscale等CLI中心でpython3常設を前提にしていない)。
echo "${BODY}" | jq . 2>/dev/null || echo "${BODY}"

EXIT_CODE=$(echo "${BODY}" | jq -r '.exitCode // "unknown"' 2>/dev/null || echo "unknown")

if [[ "${HTTP_CODE}" != "200" || "${EXIT_CODE}" != "0" ]]; then
  echo "::error::Migration failed (HTTP ${HTTP_CODE}, exitCode ${EXIT_CODE})" >&2
  exit 1
fi

echo "==> Migration succeeded (exitCode 0)"

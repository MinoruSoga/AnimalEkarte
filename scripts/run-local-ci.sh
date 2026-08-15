#!/usr/bin/env bash
# scripts/run-local-ci.sh
#
# ローカル一括 CI（make ci の実体）。
# リモート CI から外した静的ゲート + 手元で再現可能な build/test/lint を順に実行する。
#
# 前提:
#   - docker が使えること
#   - `make up` 済み（backend / frontend コンテナが healthy）であること
#     （Docker 不要のメタ検査は先頭で先に fail-fast する）
#
# 分担の正本: docs/ops/ci-policy.md
# Usage: bash scripts/run-local-ci.sh
# Exit: 最初に失敗したステップで非0
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

GOLANGCI_LINT_VERSION="${GOLANGCI_LINT_VERSION:-v2.11.4}"

step=0
total=18

begin_step() {
  step=$((step + 1))
  echo ""
  echo "=== [${step}/${total}] $1 ==="
}

compose() {
  docker compose --env-file .env.local "$@"
}

require_compose_service() {
  local svc="$1"
  if ! compose exec -T "$svc" true >/dev/null 2>&1; then
    echo "ERROR: ${svc} コンテナが未起動です。先に \`make up\` を実行してください。" >&2
    exit 1
  fi
}

# ── 1–6: Docker 不要のメタゲート（最速 fail-fast）────────────────
begin_step "Reset wait-set contract"
bash scripts/check-reset-wait-services.sh
bash scripts/check-reset-wait-services.test.sh

begin_step "CI step order guardrail"
bash scripts/check-ci-step-order.sh
bash scripts/check-ci-step-order.test.sh

begin_step "Docs symbol drift guardrail"
bash scripts/check-docs-symbol-drift.sh
bash scripts/check-docs-symbol-drift.test.sh

begin_step "ShellCheck scripts"
bash scripts/shellcheck-scripts.sh
bash scripts/shellcheck-scripts.test.sh

begin_step "eslint-disable rationale ratchet"
node frontend/scripts/check-eslint-disable-rationale.mjs
node --test frontend/scripts/check-eslint-disable-rationale.test.mjs

begin_step "Design primary CTA guard"
node scripts/check-design-primary-cta.mjs
bash scripts/check-design-primary-cta.test.sh

begin_step "A4 rehearsal isolation contract"
node --test scripts/check-a4-rehearsal-compose.test.mjs \
  scripts/check-a4-env-file.test.mjs \
  scripts/check-a4-resource-boundary.test.mjs \
  scripts/write-a4-runtime-report.test.mjs

begin_step "Design system audit (C1/C3/C5/C6/C7/C8/C9)"
node frontend/scripts/design-system-audit.mjs --cwd frontend
node --test frontend/scripts/design-system-audit.test.mjs

# ── 7–12: Go inventory / build / test（backend コンテナ）────────
require_compose_service backend

begin_step "Inventory gates (preload / master-FK / audit-tx / CASCADE / OpenAPI date / dbOrTx)"
compose exec -T backend go test ./internal/lintscan/ \
  -run 'TestPreloadClinicScope|TestClinicalResultAuditTxInventory|TestMigrationCascadeInventory|TestDBOrTxInventory|TestMasterFKWriteInventory' \
  -count=1
compose exec -T backend go test ./internal/apicontract/ \
  -run TestOpenAPIDateFormatDrift -count=1

begin_step "Backend: build"
compose exec -T backend go build ./...

begin_step "Backend: test (-race, -p 1, timeout 900s)"
compose exec -T backend go test ./... -count=1 -race -timeout 900s -p 1

begin_step "Backend: golangci-lint (local-only)"
docker run --rm \
  -v "$ROOT/backend:/app" \
  -v ekarte-go-mod-cache:/go/pkg/mod \
  -v ekarte-golangci-cache:/root/.cache \
  -w /app \
  "golangci/golangci-lint:${GOLANGCI_LINT_VERSION}" \
  golangci-lint run

begin_step "Backend: schema drift"
compose exec -T backend go test ./internal/model/ -run TestSchemaDrift -v

begin_step "Codegen sync check"
mkdir -p frontend/src/types/generated
compose run --rm codegen
if ! git diff --exit-code frontend/src/types/generated/; then
  echo "ERROR: models.ts is out of sync. Commit the updated file (make codegen)." >&2
  exit 1
fi

# ── 13–16: Frontend 静的 + build/test ───────────────────────────
require_compose_service frontend

begin_step "Frontend: ESLint"
compose exec -T frontend pnpm run lint

begin_step "Frontend: type-check"
compose exec -T frontend pnpm run type-check

begin_step "Frontend: knip unused"
compose exec -T frontend pnpm run unused

begin_step "Frontend: build + test"
compose exec -T frontend pnpm run build
compose exec -T frontend pnpm run test:run

echo ""
echo "✓ make ci passed"
echo "  local-only: inventory / guardrails / shellcheck / golangci / ESLint / type-check / knip / design CTA + design-audit"
echo "  also covered: backend build+test+schema, frontend build+test, codegen"
echo "  remote CI still runs: gitleaks / path-filtered build+test+coverage / AgentShield"
echo "  E2E is local-only: make e2e (not in automatic PR CI)"
echo "  see docs/ops/ci-policy.md"

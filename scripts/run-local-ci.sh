#!/usr/bin/env bash
# scripts/run-local-ci.sh
#
# ローカル一括 CI（make ci の実体）。
# 旧リモート CI の build/test/coverage/migration/worker/codegen 検証を含め、
# 手元で再現可能な全ゲートを順に実行する。
# リモート CI（.github/workflows/ci.yml）は workflow contracts + gitleaks のみの
# 最小構成であり、品質ゲートの実質はこのスクリプトが担う。
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
NODE_WORKER_IMAGE="${NODE_WORKER_IMAGE:-node:24-bookworm}"
WORKER_PNPM_VERSION="${WORKER_PNPM_VERSION:-10.15.0}"
GITLEAKS_IMAGE="${GITLEAKS_IMAGE:-zricethezav/gitleaks:v8.30.1}"

step=0
total=30
CURRENT_STEP=""

begin_step() {
  step=$((step + 1))
  CURRENT_STEP="$1"
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

# make ci の結果を HEAD の commit status（context `make ci`）として投稿する。
# staging branch protection がこの context を required check にしており、
# main→staging release PR の head SHA に success が無いと merge できない契約
# （docs/ops/ci-policy.md）。自己申告制である点は同文書を参照。
# gh が無い・未認証の環境では WARN のみで検証自体は続行する
# （release 側の強制は status の不在=Pending で効くため）。
# MAKE_CI_STATUS=0 で投稿を無効化できる（実験的実行用）。
MAKE_CI_CONTEXT="make ci"

post_ci_status() {
  local state="$1" description="$2"
  [[ "${MAKE_CI_STATUS:-1}" == "1" ]] || return 0
  command -v gh >/dev/null 2>&1 || {
    echo "WARN: gh CLI が無いため commit status (${MAKE_CI_CONTEXT}) を投稿できません" >&2
    return 0
  }
  local sha repo
  sha="$(git rev-parse HEAD 2>/dev/null)" || return 0
  [[ -n "$sha" ]] || return 0
  repo="$(gh repo view --json nameWithOwner --jq .nameWithOwner 2>/dev/null)" || {
    echo "WARN: gh repo view に失敗したため commit status を投稿できません" >&2
    return 0
  }
  if ! gh api "repos/${repo}/statuses/${sha}" \
      -f state="$state" \
      -f context="$MAKE_CI_CONTEXT" \
      -f description="$description" \
      -f target_url="https://github.com/${repo}/blob/${sha}/scripts/run-local-ci.sh" \
      >/dev/null 2>&1; then
    echo "WARN: commit status 投稿に失敗しました（repo:status 権限を確認）" >&2
  fi
}

on_make_ci_failure() {
  post_ci_status failure "make ci failed at: ${CURRENT_STEP:-unknown}" || true
}
trap on_make_ci_failure ERR

begin_step "Agent instruction and scoped verification contracts"
python3 -B .claude/scripts/sync-codex-mirror.py "$ROOT"
python3 -B .claude/scripts/sync-agents-skills.py "$ROOT"
python3 -B .claude/scripts/test_instruction_safety_contracts.py
python3 -B scripts/test_verify_agent_task.py
python3 -B scripts/test_agent_scope_contracts.py

# ── 2–16: Docker 不要のメタゲート（最速 fail-fast）────────────────
begin_step "Reset wait-set contract"
bash scripts/check-reset-wait-services.sh
bash scripts/check-reset-wait-services.test.sh

begin_step "CI step order guardrail"
bash scripts/check-ci-step-order.sh
bash scripts/check-ci-step-order.test.sh

# 旧 remote job「Workflow Contracts」相当: workflow 定義と repo 契約の整合を
# node:test で検査する。ci.yml を最小構成へ縮小しても本スクリプト側で検査は継続する。
begin_step "Workflow contracts"
node --test scripts/check-workflow-contracts.test.mjs

# actionlint.yml 側の構文検査は remote path-filtered のまま残るが、同一 action の
# バージョン混在検出はホスト bash で十分なためこちらにも配線する（#195 回帰防止）。
begin_step "Actions version drift"
bash scripts/check-actions-version-drift.sh
bash scripts/check-actions-version-drift.test.sh

begin_step "Worker test Makefile contract"
bash scripts/check-test-worker-makefile.test.sh

begin_step "Go coverage shard merge contract"
python3 scripts/merge_go_coverprofiles_test.py

# 旧 remote job の「Verify seed invariants」相当（CSV seed 不変条件・DB 不要）。
begin_step "Seed invariants (verify_seed)"
python3 scripts/verify_seed.py

begin_step "Docs symbol drift guardrail"
bash scripts/check-docs-symbol-drift.sh
bash scripts/check-docs-symbol-drift.test.sh

begin_step "ShellCheck scripts"
bash scripts/shellcheck-scripts.sh
bash scripts/shellcheck-scripts.test.sh

begin_step "STG UAT old_db handoff wrapper"
bash scripts/stg-uat-old-db-handoff.test.sh

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

# 旧 remote job「Gitleaks Secret Scan」相当: remote の PR 差分スキャンと同じ
# 意味論で「base からの新規コミット」だけを検査する（履歴全体スキャンは
# 既知の歴史的検出で fail するため対象外。作業ツリー全体の dir スキャンも
# .git/node_modules の誤検出で使えない）。
# 走査対象: origin/staging..HEAD（main→staging release PR 差分と同一）。
# host に gitleaks があればそれを使い、なければ pin した公式イメージで実行する。
begin_step "Gitleaks secret scan (new commits vs staging)"
if git rev-parse --verify --quiet origin/staging >/dev/null; then
  GITLEAKS_LOG_OPTS="origin/staging..HEAD"
elif git rev-parse --verify --quiet origin/main >/dev/null; then
  GITLEAKS_LOG_OPTS="origin/main..HEAD"
else
  echo "WARN: origin/staging・origin/main が無いため直近50コミットのみ走査します" >&2
  GITLEAKS_LOG_OPTS="-50"
fi
if command -v gitleaks >/dev/null 2>&1; then
  gitleaks git --log-opts="$GITLEAKS_LOG_OPTS" --config .gitleaks.toml --redact "$ROOT"
else
  docker run --rm -v "$ROOT:/repo" -w /repo \
    -e GIT_CONFIG_COUNT=1 \
    -e GIT_CONFIG_KEY_0=safe.directory \
    -e GIT_CONFIG_VALUE_0=/repo \
    "$GITLEAKS_IMAGE" \
    git --log-opts="$GITLEAKS_LOG_OPTS" --config .gitleaks.toml --redact /repo
fi

# ── 17–24: Go inventory / build / test+coverage（backend コンテナ）───────
require_compose_service backend

begin_step "Inventory gates (preload / master-FK / audit-tx / CASCADE / OpenAPI date / dbOrTx)"
compose exec -T backend go test ./internal/lintscan/ \
  -run 'TestPreloadClinicScope|TestClinicalResultAuditTxInventory|TestMigrationCascadeInventory|TestDBOrTxInventory|TestMasterFKWriteInventory' \
  -count=1
compose exec -T backend go test ./internal/apicontract/ \
  -run TestOpenAPIDateFormatDrift -count=1

begin_step "Backend: build"
compose exec -T backend go build ./...

# 旧 remote backend_test shards の統合版: 単一プロセスで全パッケージを
# -race -coverpkg=./internal/... 付きで実行し coverage.out を生成する
# （backend/coverage.out は .gitignore 済み）。
begin_step "Backend: test (-race, -p 1, timeout 900s, coverage)"
compose exec -T backend go test ./... -count=1 -race -timeout 900s -p 1 \
  -covermode=atomic -coverpkg=./internal/... -coverprofile=coverage.out

# 旧 remote job「Backend」集約ステップの coverage ratchet 相当。
begin_step "Backend: coverage ratchet"
compose exec -T backend sh -c \
  'go tool cover -func=coverage.out > coverage-summary.txt'
compose exec -T backend go run ./cmd/coverage-ratchet \
  -baseline .coverage-baseline -func-file coverage-summary.txt

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

# 旧 remote job「Migration Verify」相当: 使い捨て postgres に DDL を全適用する。
begin_step "Migration verify (fresh postgres apply)"
bash scripts/ci-migration-verify.sh

# ── 25–29: Frontend 静的 + audit + build/test+coverage ──────────────
require_compose_service frontend

begin_step "Frontend: ESLint"
compose exec -T frontend pnpm run lint

begin_step "Frontend: type-check"
compose exec -T frontend pnpm run type-check

begin_step "Frontend: knip unused"
compose exec -T frontend pnpm run unused

# 旧 remote job「Frontend Build」の pnpm audit 相当。
# registry の audit endpoint が拒否/timeout した場合は remote CI と同じく WARN で継続する。
begin_step "Frontend: pnpm audit"
set +e
audit_output="$(compose exec -T frontend pnpm audit --audit-level moderate 2>&1)"
audit_exit=$?
set -e
echo "$audit_output"
if [[ "$audit_exit" -ne 0 ]]; then
  if grep -qE "ERR_PNPM_AUDIT_BAD_RESPONSE|ERR_SOCKET_TIMEOUT" <<<"$audit_output"; then
    echo "WARN: pnpm audit unavailable (registry audit endpoint rejected or timed out). Continuing."
  else
    exit "$audit_exit"
  fi
fi

# 旧 remote frontend_test shards の統合版: vitest 全実行 + v8 coverage。
# coverage/coverage-summary.json を生成し baseline ratchet まで同一ステップで行う。
begin_step "Frontend: build + test (with coverage) + ratchet"
compose exec -T frontend pnpm run build
compose exec -T -e NODE_OPTIONS="--max-old-space-size=5120" \
  frontend pnpm exec vitest run --coverage
compose exec -T frontend node scripts/coverage-ratchet.mjs \
  --summary coverage/coverage-summary.json --baseline .coverage-baseline

# 旧 remote job「Worker Tests」相当: typecheck:worker + vitest-pool-workers (workerd)。
# frontend コンテナは ./frontend しかマウントしないため Makefile test-worker と同じ
# 使い捨て node コンテナ構成で root package.json を実行する。
begin_step "Worker: typecheck + tests (workerd)"
docker run --rm \
  -v "$ROOT:/app" \
  -v ekarte-root-node-modules:/app/node_modules \
  -v ekarte-root-pnpm-store:/pnpm-store \
  -e PNPM_STORE_DIR=/pnpm-store \
  -w /app \
  "$NODE_WORKER_IMAGE" \
  bash -lc 'set -euo pipefail; corepack enable; corepack prepare pnpm@'"$WORKER_PNPM_VERSION"' --activate; pnpm config set store-dir /pnpm-store; pnpm install --frozen-lockfile; pnpm run typecheck:worker; pnpm exec vitest run --config backend/worker/vitest.config.mts backend/worker'

post_ci_status success "all ${total} steps passed"

echo ""
echo "✓ make ci passed"
echo "  covered locally: inventory / guardrails / shellcheck / golangci / ESLint / type-check / knip / design audits"
echo "  covered locally: backend build+test+coverage ratchet / schema / codegen / migration verify / seed invariants"
echo "  covered locally: frontend audit+build+test+coverage ratchet / worker typecheck+tests / gitleaks"
echo "  remote CI (minimal): Workflow Contracts / Gitleaks Secret Scan / AgentShield (separate workflow)"
echo "  E2E is local-only: make e2e (not in automatic PR CI)"
echo "  see docs/ops/ci-policy.md"

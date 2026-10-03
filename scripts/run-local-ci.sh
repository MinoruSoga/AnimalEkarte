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
# 途中再開: `bash scripts/run-local-ci.sh 9` / `--from 9` / `CI_FROM=9` /
#   `make ci 9` で step 9 から実行する（失敗 step の再検証・時間節約用。
#   success の commit status 投稿はフル実行（FROM=1）のみ — 部分実行では
#   required check を迂回できるため投稿しない）。
#
# 分担の正本: docs/ops/ci-policy.md
# Usage: bash scripts/run-local-ci.sh [N|--from N]
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
    return 1
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

# ── step 定義（s01–s30。追加時は sXX 関数と total を同時更新）──────────

s01() {
  begin_step "Agent instruction and scoped verification contracts"
  python3 -B .claude/scripts/sync-codex-mirror.py "$ROOT"
  python3 -B .claude/scripts/sync-agents-skills.py "$ROOT"
  python3 -B .claude/scripts/test_instruction_safety_contracts.py
  python3 -B scripts/test_verify_agent_task.py
  python3 -B scripts/test_agent_scope_contracts.py
}

s02() {
  begin_step "Reset wait-set contract"
  bash scripts/check-reset-wait-services.sh
  bash scripts/check-reset-wait-services.test.sh
}

s03() {
  begin_step "CI step order guardrail"
  bash scripts/check-ci-step-order.sh
  bash scripts/check-ci-step-order.test.sh
}

# 旧 remote job「Workflow Contracts」相当: workflow 定義と repo 契約の整合を
# node:test で検査する。ci.yml を最小構成へ縮小しても本スクリプト側で検査は継続する。
s04() {
  begin_step "Workflow contracts"
  node --test scripts/check-workflow-contracts.test.mjs
}

# actionlint.yml 側の構文検査は remote path-filtered のまま残るが、同一 action の
# バージョン混在検出はホスト bash で十分なためこちらにも配線する（#195 回帰防止）。
s05() {
  begin_step "Actions version drift"
  bash scripts/check-actions-version-drift.sh
  bash scripts/check-actions-version-drift.test.sh
}

s06() {
  begin_step "Worker test Makefile contract"
  bash scripts/check-test-worker-makefile.test.sh
}

s07() {
  begin_step "Go coverage shard merge contract"
  python3 scripts/merge_go_coverprofiles_test.py
}

# 旧 remote job の「Verify seed invariants」相当（CSV seed 不変条件・DB 不要）。
s08() {
  begin_step "Seed invariants (verify_seed)"
  python3 scripts/verify_seed.py
}

s09() {
  begin_step "Docs symbol drift guardrail"
  bash scripts/check-docs-symbol-drift.sh
  bash scripts/check-docs-symbol-drift.test.sh
}

s10() {
  begin_step "ShellCheck scripts"
  bash scripts/shellcheck-scripts.sh
  bash scripts/shellcheck-scripts.test.sh
}

s11() {
  begin_step "STG UAT old_db handoff wrapper"
  bash scripts/stg-uat-old-db-handoff.test.sh
}

s12() {
  begin_step "eslint-disable rationale ratchet"
  node frontend/scripts/check-eslint-disable-rationale.mjs
  node --test frontend/scripts/check-eslint-disable-rationale.test.mjs
}

s13() {
  begin_step "Design primary CTA guard"
  node scripts/check-design-primary-cta.mjs
  bash scripts/check-design-primary-cta.test.sh
}

s14() {
  begin_step "A4 rehearsal isolation contract"
  node --test scripts/check-a4-rehearsal-compose.test.mjs \
    scripts/check-a4-env-file.test.mjs \
    scripts/check-a4-resource-boundary.test.mjs \
    scripts/write-a4-runtime-report.test.mjs
}

s15() {
  begin_step "Design system audit (C1/C3/C5/C6/C7/C8/C9)"
  node frontend/scripts/design-system-audit.mjs --cwd frontend
  node --test frontend/scripts/design-system-audit.test.mjs
}

# 旧 remote job「Gitleaks Secret Scan」相当: remote の PR 差分スキャンと同じ
# 意味論で「base からの新規コミット」だけを検査する（履歴全体スキャンは
# 既知の歴史的検出で fail するため対象外。作業ツリー全体の dir スキャンも
# .git/node_modules の誤検出で使えない）。
# 走査対象: origin/staging..HEAD（main→staging release PR 差分と同一）。
# host に gitleaks があればそれを使い、なければ pin した公式イメージで実行する。
s16() {
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
}

s17() {
  begin_step "Inventory gates (preload / master-FK / audit-tx / CASCADE / OpenAPI date / dbOrTx)"
  compose exec -T backend go test ./internal/lintscan/ \
    -run 'TestPreloadClinicScope|TestClinicalResultAuditTxInventory|TestMigrationCascadeInventory|TestDBOrTxInventory|TestMasterFKWriteInventory' \
    -count=1
  compose exec -T backend go test ./internal/apicontract/ \
    -run TestOpenAPIDateFormatDrift -count=1
}

s18() {
  begin_step "Backend: build"
  compose exec -T backend go build ./...
}

# 旧 remote backend_test shards の統合版: 単一プロセスで全パッケージを
# -race -coverpkg=./internal/... 付きで実行し coverage.out を生成する
# （backend/coverage.out は .gitignore 済み）。
s19() {
  begin_step "Backend: test (-race, -p 1, timeout 900s, coverage)"
  compose exec -T backend go test ./... -count=1 -race -timeout 900s -p 1 \
    -covermode=atomic -coverpkg=./internal/... -coverprofile=coverage.out
}

# 旧 remote job「Backend」集約ステップの coverage ratchet 相当。
s20() {
  begin_step "Backend: coverage ratchet"
  compose exec -T backend sh -c \
    'go tool cover -func=coverage.out > coverage-summary.txt'
  compose exec -T backend go run ./cmd/coverage-ratchet \
    -baseline .coverage-baseline -func-file coverage-summary.txt
}

s21() {
  begin_step "Backend: golangci-lint (local-only)"
  docker run --rm \
    -v "$ROOT/backend:/app" \
    -v ekarte-go-mod-cache:/go/pkg/mod \
    -v ekarte-golangci-cache:/root/.cache \
    -w /app \
    "golangci/golangci-lint:${GOLANGCI_LINT_VERSION}" \
    golangci-lint run
}

s22() {
  begin_step "Backend: schema drift"
  compose exec -T backend go test ./internal/model/ -run TestSchemaDrift -v
}

s23() {
  begin_step "Codegen sync check"
  mkdir -p frontend/src/types/generated
  compose run --rm codegen
  if ! git diff --exit-code frontend/src/types/generated/; then
    echo "ERROR: models.ts is out of sync. Commit the updated file (make codegen)." >&2
    return 1
  fi
}

# 旧 remote job「Migration Verify」相当: 使い捨て postgres に DDL を全適用する。
s24() {
  begin_step "Migration verify (fresh postgres apply)"
  bash scripts/ci-migration-verify.sh
}

s25() {
  begin_step "Frontend: ESLint"
  compose exec -T frontend pnpm run lint
}

s26() {
  begin_step "Frontend: type-check"
  compose exec -T frontend pnpm run type-check
}

s27() {
  begin_step "Frontend: knip unused"
  compose exec -T frontend pnpm run unused
}

# 旧 remote job「Frontend Build」の pnpm audit 相当。
# registry の audit endpoint が拒否/timeout した場合は remote CI と同じく WARN で継続する。
s28() {
  begin_step "Frontend: pnpm audit"
  local audit_output audit_exit
  set +e
  audit_output="$(compose exec -T frontend pnpm audit --audit-level moderate 2>&1)"
  audit_exit=$?
  set -e
  echo "$audit_output"
  if [[ "$audit_exit" -ne 0 ]]; then
    if grep -qE "ERR_PNPM_AUDIT_BAD_RESPONSE|ERR_SOCKET_TIMEOUT" <<<"$audit_output"; then
      echo "WARN: pnpm audit unavailable (registry audit endpoint rejected or timed out). Continuing."
    else
      return "$audit_exit"
    fi
  fi
}

# 旧 remote frontend_test shards の統合版: vitest 全実行 + v8 coverage。
# coverage/coverage-summary.json を生成し baseline ratchet まで同一ステップで行う。
s29() {
  begin_step "Frontend: build + test (with coverage) + ratchet"
  compose exec -T frontend pnpm run build
  compose exec -T -e NODE_OPTIONS="--max-old-space-size=5120" \
    frontend pnpm exec vitest run --coverage
  compose exec -T frontend node scripts/coverage-ratchet.mjs \
    --summary coverage/coverage-summary.json --baseline .coverage-baseline
}

# 旧 remote job「Worker Tests」相当: typecheck:worker + vitest-pool-workers (workerd)。
# frontend コンテナは ./frontend しかマウントしないため Makefile test-worker と同じ
# 使い捨て node コンテナ構成で root package.json を実行する。
s30() {
  begin_step "Worker: typecheck + tests (workerd)"
  docker run --rm \
    -v "$ROOT:/app" \
    -v ekarte-root-node-modules:/app/node_modules \
    -v ekarte-root-pnpm-store:/pnpm-store \
    -e PNPM_STORE_DIR=/pnpm-store \
    -w /app \
    "$NODE_WORKER_IMAGE" \
    bash -lc 'set -euo pipefail; corepack enable; corepack prepare pnpm@'"$WORKER_PNPM_VERSION"' --activate; pnpm config set store-dir /pnpm-store; pnpm install --frozen-lockfile; pnpm run typecheck:worker; pnpm exec vitest run --config backend/worker/vitest.config.mts backend/worker'
}

# ── --from / CI_FROM / 位置引数で開始 step を指定できる ────────────────
FROM="${CI_FROM:-1}"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --from=*) FROM="${1#--from=}" ;;
    --from) shift; FROM="${1:?--from requires a step number}" ;;
    *[!0-9]*|'') echo "usage: $0 [N|--from N]  (N: 1..${total})" >&2; exit 2 ;;
    *) FROM="$1" ;;
  esac
  shift
done
if ! [[ "$FROM" =~ ^[0-9]+$ ]] || (( FROM < 1 || FROM > total )); then
  echo "ERROR: --from は 1..${total} の数値で指定してください（got: ${FROM}）" >&2
  exit 2
fi
if (( FROM > 1 )); then
  echo "make ci: --from ${FROM}（step 1..$((FROM - 1)) はスキップ — スキップ分の検証は未実施です）"
fi

# 実行される step 範囲に応じて compose サービスを事前確認する
# （17–23 は backend、25–29 は frontend が必要。24・30 は使い捨て docker のみ）。
if (( FROM <= 23 )); then
  CURRENT_STEP="backend container check"
  require_compose_service backend
fi
if (( FROM <= 29 )); then
  CURRENT_STEP="frontend container check"
  require_compose_service frontend
fi

for (( i = FROM; i <= total; i++ )); do
  step=$i
  "s$(printf '%02d' "$i")"
done

# required check（context `make ci`）の success はフル実行のみ投稿する。
# 部分実行で success を出すと required check を回避できるため、FROM>1 では
# 投稿しない（失敗時の failure 投稿は実施済みのためここでは扱わない）。
if (( FROM == 1 )); then
  post_ci_status success "all ${total} steps passed"
else
  echo "note: --from ${FROM} の部分実行のため commit status は投稿しません（success 投稿はフル実行のみ）"
fi

echo ""
echo "✓ make ci passed"
echo "  covered locally: inventory / guardrails / shellcheck / golangci / ESLint / type-check / knip / design audits"
echo "  covered locally: backend build+test+coverage ratchet / schema / codegen / migration verify / seed invariants"
echo "  covered locally: frontend audit+build+test+coverage ratchet / worker typecheck+tests / gitleaks"
echo "  remote CI (minimal): Workflow Contracts / Gitleaks Secret Scan / AgentShield (separate workflow)"
echo "  resume: make ci <step#> (e.g. make ci 9)"
echo "  E2E is local-only: make e2e (not in automatic PR CI)"
echo "  see docs/ops/ci-policy.md"

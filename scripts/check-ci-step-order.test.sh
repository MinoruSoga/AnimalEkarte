#!/usr/bin/env bash
# scripts/check-ci-step-order.test.sh
#
# check-ci-step-order.sh の回帰テスト。
# 一時ディレクトリに .github/workflows/ci.yml fixture を作り、契約チェックを
#   - ゲート系(Lint/Build)がリスク系(Test/ratchet)より前 → exit 0
#   - リスク系がゲート系より前（過去の実混入パターン）      → exit 1
#   - ゲート系のみ／リスク系のみの job（比較対象なし）      → exit 0
#   - 複数 job のうち1つだけ違反                            → exit 1（かつ違反 job 名を特定）
# で判定できることを確認する。Docker 不要・純テキスト検査。
#
# Usage: bash scripts/check-ci-step-order.test.sh
# Exit codes:
#   0  全ケース PASS
#   1  いずれかのケースが期待と異なる
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHECK="$SCRIPT_DIR/check-ci-step-order.sh"

if [[ ! -f "$CHECK" ]]; then
  echo "FAIL  check script not found at $CHECK"
  exit 1
fi

TMP_ROOT="$(mktemp -d)"
trap 'rm -rf "$TMP_ROOT"' EXIT

failures=0

# fixture を <case>/.github/workflows/ci.yml と <case>/scripts/check-ci-step-order.sh に配置し、
# 本物の check スクリプトを fixture の scripts/ にコピーして実行する。
# check スクリプトは `$SCRIPT_DIR/../.github/workflows/ci.yml` を参照するため、
# scripts/ の親に .github/workflows/ci.yml を置けば fixture を検査できる。
run_case() {
  local name="$1" expected_exit="$2" ci_yml_body="$3"
  local casedir="$TMP_ROOT/$name"
  mkdir -p "$casedir/scripts" "$casedir/.github/workflows"
  printf '%s\n' "$ci_yml_body" > "$casedir/.github/workflows/ci.yml"
  cp "$CHECK" "$casedir/scripts/check-ci-step-order.sh"

  local out actual_exit
  out="$(bash "$casedir/scripts/check-ci-step-order.sh" 2>&1)" && actual_exit=0 || actual_exit=$?

  if [[ "$actual_exit" -eq "$expected_exit" ]]; then
    echo "PASS  [$name] exit=$actual_exit (expected $expected_exit)"
  else
    echo "FAIL  [$name] exit=$actual_exit (expected $expected_exit)"
    echo "----- output -----"
    printf '%s\n' "$out"
    echo "------------------"
    failures=$((failures + 1))
  fi
}

# 1. 正しい順序: Lint/Build が Test より前 → 通る
run_case "good-gate-before-risk" 0 'jobs:
  frontend:
    name: Frontend
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v7.0.0
      - name: Lint
        run: pnpm run lint
      - name: Build
        run: pnpm run build
      - name: Test (with coverage)
        run: pnpm run test:coverage
      - name: Coverage ratchet
        run: node scripts/coverage-ratchet.mjs'

# 2. 過去の実混入パターン: Test が Lint より前 → reject
run_case "bad-risk-before-gate" 1 'jobs:
  frontend:
    name: Frontend
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v7.0.0
      - name: Test (with coverage)
        run: pnpm run test:coverage
      - name: Lint
        run: pnpm run lint
      - name: Build
        run: pnpm run build'

# 3. ゲート系のみ（リスク系なし）: 比較対象が無いので通る
run_case "good-gate-only" 0 'jobs:
  shellcheck:
    name: ShellCheck Scripts
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v7.0.0
      - name: Lint
        run: bash scripts/shellcheck-scripts.sh'

# 4. リスク系のみ（ゲート系なし）: 比較対象が無いので通る
run_case "good-risk-only" 0 'jobs:
  backend:
    name: Backend
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v7.0.0
      - name: Test
        run: go test ./...'

# 5. 複数 job のうち1つだけ違反 → reject し、違反 job を特定できる
run_case "bad-one-of-two-jobs" 1 'jobs:
  backend:
    name: Backend
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v7.0.0
      - name: Build
        run: go build ./...
      - name: Lint
        run: golangci-lint run
      - name: Test
        run: go test ./...
  frontend:
    name: Frontend
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v7.0.0
      - name: Coverage ratchet
        run: node scripts/coverage-ratchet.mjs
      - name: Lint
        run: pnpm run lint'

echo "----"
if [[ "$failures" -gt 0 ]]; then
  echo "RESULT  $failures case(s) failed"
  exit 1
fi
echo "RESULT  all cases passed"
exit 0

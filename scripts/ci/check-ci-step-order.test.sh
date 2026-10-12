#!/usr/bin/env bash
# scripts/ci/check-ci-step-order.test.sh
#
# check-ci-step-order.sh の回帰テスト。
# 一時ディレクトリに .github/workflows/ci.yml fixture を作り、契約チェックを
#   - ゲート系(Lint/Build)がリスク系(Test/ratchet)より前 → exit 0
#   - リスク系がゲート系より前（過去の実混入パターン）      → exit 1
#   - ゲート系のみ／リスク系のみの job（比較対象なし）      → exit 0
#   - 複数 job のうち1つだけ違反                            → exit 1（かつ違反 job 名を特定）
# で判定できることを確認する。Docker 不要・純テキスト検査。
#
# Usage: bash scripts/ci/check-ci-step-order.test.sh
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

# fixture を <case>/.github/workflows/ci.yml と <case>/scripts/ci/check-ci-step-order.sh に配置し、
# 本物の check スクリプトを fixture の scripts/ にコピーして実行する。
# check スクリプトは `$SCRIPT_DIR/../../.github/workflows/ci.yml` を参照するため、
# scripts/ の親に .github/workflows/ci.yml を置けば fixture を検査できる。
run_case() {
  local name="$1" expected_exit="$2" ci_yml_body="$3"
  local casedir="$TMP_ROOT/$name"
  mkdir -p "$casedir/scripts/ci" "$casedir/.github/workflows"
  printf '%s\n' "$ci_yml_body" > "$casedir/.github/workflows/ci.yml"
  cp "$CHECK" "$casedir/scripts/ci/check-ci-step-order.sh"

  local out actual_exit
  out="$(bash "$casedir/scripts/ci/check-ci-step-order.sh" 2>&1)" && actual_exit=0 || actual_exit=$?

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
        run: bash scripts/checks/shellcheck-scripts.sh'

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

RUN_CI="$SCRIPT_DIR/run-local-ci.sh"
if [[ ! -f "$RUN_CI" ]]; then
  echo "FAIL  run-local-ci.sh not found at $RUN_CI"
  exit 1
fi

PROLOGUE="$TMP_ROOT/lanes-prologue.sh"
cat > "$PROLOGUE" <<'EOF'
set -euo pipefail
export MAKE_CI_STATUS=0
release_gate() { printf 'go\n' <>"$1" >&0; }
gate_read() { local __t __rc=0; exec 9<>"$1"; printf 'blocked\n' >"$2"; read -r -t 10 __t <&9 || __rc=$?; exec 9>&-; return "$__rc"; }
__mk_stubs() {
  local __i __name
  for __i in $(seq 1 30); do
    __name="$(printf 's%02d' "$__i")"
    eval "$__name() { printf '%s\n' '$__i' >> started; printf '%s\n' '$__i' >> finished; }"
  done
}
EOF

lane_case() {
  local name=$1
  local d="$TMP_ROOT/lanes-$name"
  mkdir -p "$d"
  {
    printf 'source "%s"\n' "$PROLOGUE"
    printf 'source "%s"\n' "$RUN_CI"
    printf 'require_compose_service() { return 0; }\n'
    printf 'cd "%s"\n' "$d"
    printf 'LANE_DIR="%s/lane"\n' "$d"
    printf 'mkdir -p "$LANE_DIR"\n'
    printf 'mkfifo gate23 gate30\n'
    printf '__mk_stubs\n'
    cat
  } > "$d/scenario.sh"
}

wait_marker() {
  local f=$1 pat=$2 lim=${3:-200} i=0
  while (( i < lim )); do
    [[ -f $f ]] && grep -qE "$pat" "$f" && return 0
    sleep 0.05
    i=$((i + 1))
  done
  return 1
}

start_scenario() {
  D="$TMP_ROOT/lanes-$1"
  : >"$D/out.log"
  bash "$D/scenario.sh" >>"$D/out.log" 2>&1 &
  SPID=$!
}

finish_scenario() {
  local wd
  ( sleep 30; kill -9 "$SPID" 2>/dev/null ) &
  wd=$!
  SRC=0
  wait "$SPID" || SRC=$?
  kill "$wd" 2>/dev/null || true
  wait "$wd" 2>/dev/null || true
}

release_gate() { printf 'go\n' <>"$1" >&0; }

CASE=""
assert() {
  local desc=$1
  shift
  if "$@"; then
    echo "PASS  [$CASE] $desc"
  else
    echo "FAIL  [$CASE] $desc"
    failures=$((failures + 1))
  fi
}

present() { [[ -f $2 ]] && grep -qE "$1" "$2"; }
absent() { [[ -f $2 ]] && ! grep -qE "$1" "$2"; }
rc_eq() { [[ $SRC -eq $1 ]]; }
rc_ne() { [[ $SRC -ne $1 ]]; }

CASE="lanes-gate-holds-frontend"
lane_case "$CASE" <<'CASE_BODY'
s23() { printf '%s\n' 23 >> started; gate_read gate23 s23.blocked; printf '%s\n' 23 >> finished; }
s30() { printf '%s\n' 30 >> started; gate_read gate30 s30.blocked; printf '%s\n' 30 >> finished; }
FROM=1
run_lanes "parallel lane failure" "${LANES[@]}"
CASE_BODY
start_scenario "$CASE"
assert "s23 parked inside gate read" wait_marker "$D/s23.blocked" '.'
assert "s24 started while s23 blocked" wait_marker "$D/started" '^24$'
assert "s30 parked inside gate read" wait_marker "$D/s30.blocked" '.'
assert "s23 still blocked at check time" absent '^23$' "$D/finished"
assert "no s25/s28 started before s23 release" absent '^(25|28)$' "$D/started"
release_gate "$D/gate23"
assert "s25 starts after s23 release" wait_marker "$D/started" '^25$'
assert "frontend started while unrelated long lane still running (no global wait)" \
  absent '^30$' "$D/finished"
release_gate "$D/gate30"
finish_scenario
assert "run_lanes exit 0; every launched lane joined" rc_eq 0
assert "frontend lane completed to s25" present '^25$' "$D/finished"

CASE="lanes-backend-fail-before-23"
lane_case "$CASE" <<'CASE_BODY'
s20() { printf '%s\n' 20 >> started; return 1; }
FROM=1
run_lanes "parallel lane failure" "${LANES[@]}"
CASE_BODY
start_scenario "$CASE"
finish_scenario
assert "run_lanes nonzero on backend failure before s23" rc_ne 0
assert "frontend suppressed on backend failure" absent '^(25|29)$' "$D/started"
assert "migration lane joined" present '^24$' "$D/finished"
assert "worker lane joined" present '^30$' "$D/finished"

CASE="lanes-backend-fail-in-23"
lane_case "$CASE" <<'CASE_BODY'
s23() { printf '%s\n' 23 >> started; return 1; }
FROM=1
run_lanes "parallel lane failure" "${LANES[@]}"
CASE_BODY
start_scenario "$CASE"
finish_scenario
assert "run_lanes nonzero on s23 failure" rc_ne 0
assert "frontend suppressed on s23 failure" absent '^(25|29)$' "$D/started"
assert "migration lane joined" present '^24$' "$D/finished"
assert "worker lane joined" present '^30$' "$D/finished"

CASE="lanes-missing-producer"
lane_case "$CASE" <<'CASE_BODY'
FROM=1
run_lanes "parallel lane failure" "backend:17 18" "frontend:25 26"
CASE_BODY
start_scenario "$CASE"
finish_scenario
assert "run_lanes nonzero when producer is missing" rc_ne 0
assert "frontend suppressed when producer missing" absent '^25$' "$D/started"
assert "missing producer recorded in failures" \
  present 'scheduled codegen producer is missing' "$D/lane/failures"
assert "launched backend lane joined" present '^18$' "$D/finished"

CASE="lanes-resume-from23-gates"
lane_case "$CASE" <<'CASE_BODY'
s23() { printf '%s\n' 23 >> started; gate_read gate23 s23.blocked; printf '%s\n' 23 >> finished; }
FROM=23
run_lanes "parallel lane failure" "${LANES[@]}"
CASE_BODY
start_scenario "$CASE"
assert "s23 parked inside gate read under FROM=23" wait_marker "$D/s23.blocked" '.'
assert "s24 started under FROM=23" wait_marker "$D/started" '^24$'
assert "s23 still blocked at check time" absent '^23$' "$D/finished"
assert "no frontend step before s23 release under FROM=23" absent '^(25|28)$' "$D/started"
release_gate "$D/gate23"
finish_scenario
assert "run_lanes exit 0 under FROM=23" rc_eq 0
assert "frontend completed after release under FROM=23" present '^25$' "$D/finished"

CASE="lanes-resume-from24-waitfree"
lane_case "$CASE" <<'CASE_BODY'
s30() { printf '%s\n' 30 >> started; gate_read gate30 s30.blocked; printf '%s\n' 30 >> finished; }
FROM=24
run_lanes "parallel lane failure" "${LANES[@]}"
CASE_BODY
start_scenario "$CASE"
assert "s30 parked inside gate read under FROM=24" wait_marker "$D/s30.blocked" '.'
assert "s25 starts immediately under FROM=24 (wait-free)" wait_marker "$D/started" '^25$'
assert "s25 started while unrelated lane blocked" absent '^30$' "$D/finished"
release_gate "$D/gate30"
finish_scenario
assert "run_lanes exit 0 under FROM=24" rc_eq 0

CASE="lanes-resume-from25-waitfree"
lane_case "$CASE" <<'CASE_BODY'
s30() { printf '%s\n' 30 >> started; gate_read gate30 s30.blocked; printf '%s\n' 30 >> finished; }
FROM=25
run_lanes "parallel lane failure" "${LANES[@]}"
CASE_BODY
start_scenario "$CASE"
assert "s30 parked inside gate read under FROM=25" wait_marker "$D/s30.blocked" '.'
assert "s25 starts immediately under FROM=25 (wait-free)" wait_marker "$D/started" '^25$'
release_gate "$D/gate30"
finish_scenario
assert "run_lanes exit 0 under FROM=25" rc_eq 0

CASE="lanes-meta-unaffected"
lane_case "$CASE" <<'CASE_BODY'
FROM=1
run_lanes "meta gate failure" "${META_LANES[@]}"
CASE_BODY
start_scenario "$CASE"
finish_scenario
assert "meta lanes exit 0" rc_eq 0
assert "meta-a completed" present '^1$' "$D/finished"
assert "meta-c completed" present '^16$' "$D/finished"

CASE="lanes-no-reader-fails-not-hangs"
lane_case "$CASE" <<'CASE_BODY'
s23() { printf '%s\n' 23 >> started; gate_read gate23 s23.blocked; printf '%s\n' 23 >> finished; }
FROM=1
run_lanes "parallel lane failure" "${LANES[@]}"
CASE_BODY
start_scenario "$CASE"
assert "s23 parked inside gate read" wait_marker "$D/s23.blocked" '.'
finish_scenario
assert "bounded read timeout makes lane fail (no hang)" rc_ne 0
assert "frontend suppressed on s23 gate timeout" absent '^(25|29)$' "$D/started"
assert "migration lane joined" present '^24$' "$D/finished"
assert "worker lane joined" present '^30$' "$D/finished"
assert "release with no reader does not block the harness" release_gate "$D/gate23"

echo "----"
if [[ "$failures" -gt 0 ]]; then
  echo "RESULT  $failures case(s) failed"
  exit 1
fi
echo "RESULT  all cases passed"
exit 0

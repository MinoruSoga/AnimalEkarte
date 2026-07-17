#!/usr/bin/env bash
# scripts/check-actions-version-drift.test.sh
#
# check-actions-version-drift.sh の回帰テスト。
# 一時ディレクトリに workflows fixture を作り、ドリフト検査を
#   - 全 action が単一バージョン                      → exit 0
#   - 同一 action が @v4 と @v6 で混在（#195 の実回帰）→ exit 1
#   - ローカル参照(./)・docker:// は対象外            → exit 0
#   - 同一バージョンの複数回参照はドリフトではない    → exit 0
#   - コメント付き SHA ピンと素の SHA ピンが同一 SHA  → exit 0
# で判定できることを確認する。Docker 不要・純テキスト検査。
#
# Usage: bash scripts/check-actions-version-drift.test.sh
# Exit codes:
#   0  全ケース PASS
#   1  いずれかのケースが期待と異なる
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHECK="$SCRIPT_DIR/check-actions-version-drift.sh"

if [[ ! -f "$CHECK" ]]; then
  echo "FAIL  check script not found at $CHECK"
  exit 1
fi

TMP_ROOT="$(mktemp -d)"
trap 'rm -rf "$TMP_ROOT"' EXIT

failures=0

# fixture を <case>/workflows/*.yml に配置し、check スクリプトへ引数で渡して実行する。
run_case() {
  local name="$1" expected_exit="$2"
  shift 2
  local casedir="$TMP_ROOT/$name/workflows"
  mkdir -p "$casedir"
  local i=0
  for body in "$@"; do
    printf '%s\n' "$body" > "$casedir/wf$i.yml"
    i=$((i + 1))
  done

  local out actual_exit
  out="$(bash "$CHECK" "$casedir" 2>&1)" && actual_exit=0 || actual_exit=$?

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

# ケース1: 全 action 単一バージョン → PASS
run_case "unified" 0 \
'jobs:
  a:
    steps:
      - uses: actions/checkout@v7.0.0
      - uses: actions/setup-node@v6' \
'jobs:
  b:
    steps:
      - uses: actions/checkout@v7.0.0
      - uses: actions/setup-node@v6'

# ケース2: setup-node が v4/v6 混在（#195 実回帰の再現）→ FAIL
run_case "drift-v4-v6" 1 \
'jobs:
  a:
    steps:
      - uses: actions/setup-node@v6' \
'jobs:
  b:
    steps:
      - uses: actions/setup-node@v4'

# ケース3: ローカル参照と docker:// は対象外 → PASS
run_case "local-and-docker-ignored" 0 \
'jobs:
  a:
    steps:
      - uses: ./.github/actions/my-action
      - uses: docker://alpine:3.20
      - uses: actions/checkout@v7.0.0'

# ケース4: 同一ファイル内の同一バージョン複数参照 → PASS
run_case "same-version-repeated" 0 \
'jobs:
  a:
    steps:
      - uses: actions/checkout@v7.0.0
      - uses: actions/checkout@v7.0.0
      - uses: actions/checkout@v7.0.0'

# ケース5: SHA ピン（コメント有無の差はあってよい・SHA 一致）→ PASS
run_case "sha-pin-comment" 0 \
'jobs:
  a:
    steps:
      - uses: some-org/some-action@0b5e1bc95a712285d6bf6289f67d6694de4cf8ef # v1.2.3' \
'jobs:
  b:
    steps:
      - uses: some-org/some-action@0b5e1bc95a712285d6bf6289f67d6694de4cf8ef'

# ケース6: 同一 action の SHA と タグ の混在はドリフト → FAIL
run_case "sha-vs-tag-drift" 1 \
'jobs:
  a:
    steps:
      - uses: some-org/some-action@0b5e1bc95a712285d6bf6289f67d6694de4cf8ef' \
'jobs:
  b:
    steps:
      - uses: some-org/some-action@v1'

if [[ "$failures" -gt 0 ]]; then
  echo "RESULT  $failures case(s) failed"
  exit 1
fi
echo "RESULT  all cases passed"

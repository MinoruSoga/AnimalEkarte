#!/usr/bin/env bash
# scripts/check-reset-wait-services.test.sh
#
# check-reset-wait-services.sh の回帰テスト。
# 一時ディレクトリに Makefile (+ 必要なら contract) fixture を作り、契約チェックを
#   - 正しい wait-set（Makefile 直書き / contract 委譲） → exit 0
#   - one-shot codegen 混入 → exit 1
#   - 必須サービス欠落      → exit 1
#   - 裸の up --wait        → exit 1
#   - 全 volume 一括削除フラグ → exit 1
# で判定できることを確認する。Docker 不要・純テキスト検査。
#
# Usage: bash scripts/check-reset-wait-services.test.sh
# Exit codes:
#   0  全ケース PASS
#   1  いずれかのケースが期待と異なる
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHECK="$SCRIPT_DIR/check-reset-wait-services.sh"

if [[ ! -f "$CHECK" ]]; then
  echo "FAIL  check script not found at $CHECK"
  exit 1
fi

TMP_ROOT="$(mktemp -d)"
trap 'rm -rf "$TMP_ROOT"' EXIT

failures=0

GOOD_CONTRACT_WAIT='#!/usr/bin/env bash
# fixture contract
compose() { :; }
compose up -d --build --wait --wait-timeout 1200 db backend frontend
'

BAD_CONTRACT_CODEGEN='#!/usr/bin/env bash
compose up -d --build --wait --wait-timeout 1200 db backend frontend codegen
'

BAD_CONTRACT_MISSING='#!/usr/bin/env bash
compose up -d --build --wait --wait-timeout 1200 db backend
'

BAD_CONTRACT_BARE='#!/usr/bin/env bash
compose up -d --build --wait --wait-timeout 1200
'

# fixture を <case>/Makefile と <case>/scripts/ に配置し、
# 本物の check スクリプトを fixture の scripts/ にコピーして実行する。
# check スクリプトは `$SCRIPT_DIR/../Makefile` と隣の contract を参照する。
run_case() {
  local name="$1" expected_exit="$2" makefile_body="$3"
  local contract_body="${4:-}"
  local casedir="$TMP_ROOT/$name"
  mkdir -p "$casedir/scripts"
  # Use printf %b so \t becomes a real tab for Make recipe lines.
  printf '%b\n' "$makefile_body" > "$casedir/Makefile"
  cp "$CHECK" "$casedir/scripts/check-reset-wait-services.sh"
  if [[ -n "$contract_body" ]]; then
    printf '%s\n' "$contract_body" > "$casedir/scripts/local-db-reset-contract.sh"
  fi

  local out actual_exit
  out="$(bash "$casedir/scripts/check-reset-wait-services.sh" 2>&1)" && actual_exit=0 || actual_exit=$?

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

# 1. contract 委譲 + 正しい wait-set → 通る
run_case "good-delegate-contract" 0 \
'reset:
\t@bash scripts/local-db-reset-contract.sh

migrate:
\t@echo ok' \
"$GOOD_CONTRACT_WAIT"

# 2. contract に codegen 混入 → reject
run_case "bad-delegate-codegen" 1 \
'reset:
\t@bash scripts/local-db-reset-contract.sh

migrate:
\t@echo ok' \
"$BAD_CONTRACT_CODEGEN"

# 3. contract で必須サービス欠落 → reject
run_case "bad-delegate-missing-required" 1 \
'reset:
\t@bash scripts/local-db-reset-contract.sh

migrate:
\t@echo ok' \
"$BAD_CONTRACT_MISSING"

# 4. contract が裸の up --wait → reject
run_case "bad-delegate-bare-up-wait" 1 \
'reset:
\t@bash scripts/local-db-reset-contract.sh

migrate:
\t@echo ok' \
"$BAD_CONTRACT_BARE"

# 5. Makefile が全 volume 一括削除フラグを使う → reject（委譲していても）
run_case "bad-makefile-volume-wipe" 1 \
'reset:
\tdocker compose down -v
\t@bash scripts/local-db-reset-contract.sh

migrate:
\t@echo ok' \
"$GOOD_CONTRACT_WAIT"

# 6. 旧形: Makefile 直書きの正しい wait-set（contract 無し）→ 通る
run_case "good-inline-makefile" 0 \
'reset:
\tdocker compose down --remove-orphans
\tdocker compose up -d --build --wait --wait-timeout 1200 db backend frontend

migrate:
\t@echo ok'

# 7. 旧形: one-shot codegen を wait 対象に混入 → reject
run_case "bad-inline-codegen" 1 \
'reset:
\tdocker compose up -d --build --wait --wait-timeout 1200 db backend frontend codegen

migrate:
\t@echo ok'

# 8. 旧形: 必須サービス (frontend) 欠落 → reject
run_case "bad-inline-missing-required" 1 \
'reset:
\tdocker compose up -d --build --wait --wait-timeout 1200 db backend

migrate:
\t@echo ok'

# 9. 旧形: 裸の up --wait → reject
run_case "bad-inline-bare-up-wait" 1 \
'reset:
\tdocker compose up -d --build --wait --wait-timeout 1200

migrate:
\t@echo ok'

# 10. 旧形: up --wait 自体が無い → reject
run_case "bad-inline-no-wait" 1 \
'reset:
\tdocker compose up -d db backend frontend

migrate:
\t@echo ok'

# 11. 行継続で codegen を後続物理行に逃がす silent bypass → reject
run_case "bad-inline-continuation-codegen" 1 \
'reset:
\tdocker compose up -d --build --wait --wait-timeout 1200 \\
\t  db backend frontend codegen

migrate:
\t@echo ok'

# 12. 行継続の正しい複数行 wait-set → 通る
run_case "good-inline-continuation" 0 \
'reset:
\tdocker compose up -d --build --wait --wait-timeout 1200 \\
\t  db backend frontend

migrate:
\t@echo ok'

echo "----"
if [[ "$failures" -gt 0 ]]; then
  echo "RESULT  $failures case(s) failed"
  exit 1
fi
echo "RESULT  all cases passed"
exit 0

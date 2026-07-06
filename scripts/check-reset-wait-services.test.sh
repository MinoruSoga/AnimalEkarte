#!/usr/bin/env bash
# scripts/check-reset-wait-services.test.sh
#
# check-reset-wait-services.sh の回帰テスト。
# 一時ディレクトリに Makefile fixture を作り、契約チェックを
#   - 正しい wait-set      → exit 0
#   - one-shot codegen 混入 → exit 1
#   - 必須サービス欠落      → exit 1
#   - 裸の up --wait        → exit 1
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

# fixture を <case>/Makefile と <case>/scripts/check... に配置し、
# 本物の check スクリプトを fixture の scripts/ にリンクして実行する。
# check スクリプトは `$SCRIPT_DIR/../Makefile` を参照するため、
# scripts/ の親に Makefile を置けば fixture を検査できる。
run_case() {
  local name="$1" expected_exit="$2" makefile_body="$3"
  local casedir="$TMP_ROOT/$name"
  mkdir -p "$casedir/scripts"
  printf '%s\n' "$makefile_body" > "$casedir/Makefile"
  cp "$CHECK" "$casedir/scripts/check-reset-wait-services.sh"

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

# 1. 正しい wait-set: 明示サービス列 + codegen 除外 → 通る
run_case "good-explicit-services" 0 'reset:
	$(DC) down -v
	$(DC) up -d --build --wait --wait-timeout 1200 db backend frontend

migrate:
	$(DC) run --rm --entrypoint go backend run ./cmd/migrate'

# 2. one-shot codegen を wait 対象に混入 → reject
run_case "bad-includes-codegen" 1 'reset:
	$(DC) down -v
	$(DC) up -d --build --wait --wait-timeout 1200 db backend frontend codegen

migrate:
	$(DC) run --rm --entrypoint go backend run ./cmd/migrate'

# 3. 必須サービス (frontend) 欠落 → reject
run_case "bad-missing-required" 1 'reset:
	$(DC) down -v
	$(DC) up -d --build --wait --wait-timeout 1200 db backend

migrate:
	$(DC) run --rm --entrypoint go backend run ./cmd/migrate'

# 4. 裸の up --wait（明示サービス列なし = 全サービス対象）→ reject
run_case "bad-bare-up-wait" 1 'reset:
	$(DC) down -v
	$(DC) up -d --build --wait --wait-timeout 1200

migrate:
	$(DC) run --rm --entrypoint go backend run ./cmd/migrate'

# 5. reset レシピに up --wait 自体が無い → reject
run_case "bad-no-wait-invocation" 1 'reset:
	$(DC) down -v
	$(DC) up -d db backend frontend

migrate:
	$(DC) run --rm --entrypoint go backend run ./cmd/migrate'

# 6. 行継続 `\` で codegen を後続物理行に逃がす silent bypass → reject
#    継続行を畳まないと wait 行抽出時に codegen を見逃すため、ここで防ぐ。
run_case "bad-continuation-line-codegen" 1 'reset:
	$(DC) down -v
	$(DC) up -d --build --wait --wait-timeout 1200 \
	  db backend frontend codegen

migrate:
	$(DC) run --rm --entrypoint go backend run ./cmd/migrate'

# 7. 行継続 `\` を使った正しい複数行 wait-set → 通る（継続畳み込みの false-positive 防止）
run_case "good-continuation-line" 0 'reset:
	$(DC) down -v
	$(DC) up -d --build --wait --wait-timeout 1200 \
	  db backend frontend

migrate:
	$(DC) run --rm --entrypoint go backend run ./cmd/migrate'

echo "----"
if [[ "$failures" -gt 0 ]]; then
  echo "RESULT  $failures case(s) failed"
  exit 1
fi
echo "RESULT  all cases passed"
exit 0

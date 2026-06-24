#!/usr/bin/env bash
# scripts/check-reset-wait-services.sh
#
# `make reset` の `docker compose up ... --wait` が待機対象として
# 長寿命サービス (db migrate backend frontend) を明示し、one-shot の
# codegen を含めないことを保証する静的契約チェック。
#
# 背景:
#   codegen は `restart: "no"`・healthcheck なし・依存先なしの one-shot で、
#   tygo 実行後に正常終了する。これを `up --wait` の対象に含めると、
#   正常終了が "running|healthy に到達できない" と判定され `make reset` が
#   cosmetic exit 1 を返す。`up` ターゲットと同じく明示サービス列にすることで
#   この false failure を防ぐ。
#
# このチェックは Docker を必要とせず、Makefile のテキストだけを検査する。
#
# Usage: bash scripts/check-reset-wait-services.sh
# Exit codes:
#   0  契約を満たす
#   1  契約違反（reset の wait 対象が誤っている）
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MAKEFILE="$SCRIPT_DIR/../Makefile"

REQUIRED_SERVICES=(db migrate backend frontend)
FORBIDDEN_SERVICES=(codegen)

if [[ ! -f "$MAKEFILE" ]]; then
  echo "FAIL  Makefile not found at $MAKEFILE"
  exit 1
fi

# `reset:` レシピ本文（次のターゲット見出し行まで）を抽出する。
# レシピ行はタブ始まりなので、行頭から始まる `name:` 見出しで区切る。
# 前提: ターゲット名は英数 _ . - で構成される（この Makefile の慣習）。
# `reset: prereq` のように prereq が付いても `^reset:` でマッチするため問題ない。
# 末尾 `\` の行継続は 1 論理行に畳む。これをしないと、`up --wait` 本体と
# サービス列が複数物理行に分かれた場合に後続行の codegen 混入を見逃す
# （= silent bypass）。継続行を畳んでから wait 行を抽出することで防ぐ。
recipe="$(awk '
  /^reset:/        { inrecipe=1; next }
  inrecipe && /^[A-Za-z0-9_.-]+:/ { if (buf != "") print buf; exit }
  inrecipe {
    sub(/\r$/, "")                 # CRLF 耐性
    if ($0 ~ /\\[[:space:]]*$/) {  # 行末 `\` → 次行を連結
      sub(/\\[[:space:]]*$/, "")
      buf = buf $0 " "
      next
    }
    print buf $0
    buf = ""
  }
  END { if (buf != "") print buf }
' "$MAKEFILE")"

if [[ -z "$recipe" ]]; then
  echo "FAIL  could not locate the 'reset:' recipe in $MAKEFILE"
  exit 1
fi

# レシピ内の `up ... --wait` 起動行を取り出す（コメント行 # は除外）。
wait_line="$(printf '%s\n' "$recipe" | grep -v '^[[:space:]]*#' | grep -E 'up .*--wait' || true)"

if [[ -z "$wait_line" ]]; then
  echo "FAIL  'reset' recipe has no 'up ... --wait' invocation"
  exit 1
fi

errors=0

for svc in "${REQUIRED_SERVICES[@]}"; do
  if ! grep -qE "(^|[[:space:]])$svc([[:space:]]|\$)" <<<"$wait_line"; then
    echo "FAIL  reset wait set is missing required service: $svc"
    errors=$((errors + 1))
  fi
done

for svc in "${FORBIDDEN_SERVICES[@]}"; do
  if grep -qE "(^|[[:space:]])$svc([[:space:]]|\$)" <<<"$wait_line"; then
    echo "FAIL  reset wait set must NOT include one-shot service: $svc"
    echo "      ($svc exits by design; waiting on it causes a cosmetic exit-1)"
    errors=$((errors + 1))
  fi
done

# 明示サービス列が無い（= 全サービス対象）形に戻っていないかも検出する。
# REQUIRED が一つでも欠ける場合は上のループで検出済みだが、裸の `up --wait`
# だと診断メッセージが「個別サービス欠落」になり原因が伝わりにくいため、
# 「明示サービス列なし」という固有の診断をここで出す。
# サービス名のリテラルは REQUIRED_SERVICES から組み立て、二重管理を避ける。
required_alt="$(IFS='|'; echo "${REQUIRED_SERVICES[*]}")"
if ! grep -qE "(^|[[:space:]])($required_alt)([[:space:]]|\$)" <<<"$wait_line"; then
  echo "FAIL  reset uses a bare 'up --wait' (no explicit service list)"
  echo "      this waits on ALL services including one-shot codegen → cosmetic exit-1"
  errors=$((errors + 1))
fi

if [[ "$errors" -gt 0 ]]; then
  echo "----"
  echo "reset wait line was: ${wait_line#"${wait_line%%[![:space:]]*}"}"
  exit 1
fi

echo "PASS  reset waits on [${REQUIRED_SERVICES[*]}] and excludes [${FORBIDDEN_SERVICES[*]}]"
exit 0

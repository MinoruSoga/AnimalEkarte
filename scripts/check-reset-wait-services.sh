#!/usr/bin/env bash
# scripts/check-reset-wait-services.sh
#
# `make reset` の待機対象が長寿命サービス (db backend frontend) を明示し、
# one-shot の codegen を含めないことを保証する静的契約チェック。
#
# TASK-607 以降: reset 本体は scripts/local-db-reset-contract.sh に委譲する。
# Makefile の reset レシピが contract を呼ぶ場合は、wait-set を contract 側から
# 検査する。Makefile 直書きの旧形 fixture も後方互換で検査する。
#
# 加えて Makefile reset が compose の全 volume 一括削除フラグを使っていないことも
# 検証する（DB volume のみ削除は contract スクリプト側の責務）。
#
# migration は専用の migrate サービスではなく backend の entrypoint 内で
# 実行されるため、待機対象サービスに migrate は存在しない。
#
# 背景:
#   codegen は `restart: "no"`・healthcheck なし・依存先なしの one-shot で、
#   tygo 実行後に正常終了する。これを `up --wait` の対象に含めると、
#   正常終了が "running|healthy に到達できない" と判定され `make reset` が
#   cosmetic exit 1 を返す。`up` ターゲットと同じく明示サービス列にすることで
#   この false failure を防ぐ。
#
# このチェックは Docker を必要とせず、Makefile / contract のテキストだけを検査する。
#
# Usage: bash scripts/check-reset-wait-services.sh
# Exit codes:
#   0  契約を満たす
#   1  契約違反（reset の wait 対象が誤っている / 禁止フラグ）
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MAKEFILE="$SCRIPT_DIR/../Makefile"
CONTRACT="$SCRIPT_DIR/local-db-reset-contract.sh"

REQUIRED_SERVICES=(db backend frontend)
FORBIDDEN_SERVICES=(codegen)

if [[ ! -f "$MAKEFILE" ]]; then
  echo "FAIL  Makefile not found at $MAKEFILE"
  exit 1
fi

# `reset:` レシピ本文（次のターゲット見出し行まで）を抽出する。
# レシピ行はタブ始まりなので、行頭から始まる `name:` 見出しで区切る。
# 末尾 `\` の行継続は 1 論理行に畳む。
recipe="$(awk '
  /^reset:/        { inrecipe=1; next }
  inrecipe && /^[A-Za-z0-9_.-]+:/ { if (buf != "") print buf; exit }
  inrecipe {
    sub(/\r$/, "")
    if ($0 ~ /\\[[:space:]]*$/) {
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

errors=0

# Makefile reset が全 volume 一括削除系フラグを使っていないこと。
if printf '%s\n' "$recipe" | grep -v '^[[:space:]]*#' | grep -E 'down[[:space:]]+-v|[[:space:]]--volumes([[:space:]]|$)|volume[[:space:]]+prune' >/dev/null 2>&1; then
  echo "FAIL  reset recipe must not wipe all compose volumes or prune the volume store"
  errors=$((errors + 1))
fi

# wait 検査対象テキスト: contract 委譲なら contract 本体、否则 Makefile レシピ。
wait_source="$recipe"
wait_source_label="Makefile reset recipe"
if printf '%s\n' "$recipe" | grep -F 'local-db-reset-contract.sh' >/dev/null 2>&1; then
  if [[ ! -f "$CONTRACT" ]]; then
    echo "FAIL  reset delegates to local-db-reset-contract.sh but file is missing: $CONTRACT"
    exit 1
  fi
  wait_source="$(cat "$CONTRACT")"
  wait_source_label="scripts/local-db-reset-contract.sh"
fi

# `up ... --wait` 起動行を取り出す（コメント行 # は除外）。
# contract 内の複数行も行継続を畳んだうえで検査するため、単純 grep でよい
# （contract は 1 論理行に up --wait を書く）。
wait_line="$(printf '%s\n' "$wait_source" | grep -v '^[[:space:]]*#' | grep -E 'up .*--wait' || true)"

if [[ -z "$wait_line" ]]; then
  echo "FAIL  $wait_source_label has no 'up ... --wait' invocation"
  exit 1
fi

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

required_alt="$(IFS='|'; echo "${REQUIRED_SERVICES[*]}")"
if ! grep -qE "(^|[[:space:]])($required_alt)([[:space:]]|\$)" <<<"$wait_line"; then
  echo "FAIL  reset uses a bare 'up --wait' (no explicit service list)"
  echo "      this waits on ALL services including one-shot codegen → cosmetic exit-1"
  errors=$((errors + 1))
fi

if [[ "$errors" -gt 0 ]]; then
  echo "----"
  echo "wait source: $wait_source_label"
  echo "wait line was: ${wait_line#"${wait_line%%[![:space:]]*}"}"
  exit 1
fi

echo "PASS  reset waits on [${REQUIRED_SERVICES[*]}] and excludes [${FORBIDDEN_SERVICES[*]}] (via $wait_source_label)"
exit 0

#!/usr/bin/env bash
# scripts/shellcheck-scripts.test.sh
#
# シェル lint ゲート (shellcheck-scripts.sh) の回帰テスト。
# 一時ディレクトリに scripts/ fixture を作り、本物のゲートスクリプトをそこへ
# 配置して実行することで、ゲートが次を満たすことを検証する:
#   - clean なスクリプトのみ        → exit 0
#   - 本物のシェルバグ混入          → exit 1（text-only では拾えない AST 検出）
#   - 行継続 `\` で隠したバグ        → exit 1（整形バイパス耐性）
#   - # shellcheck disable=all       → exit 1（一括抑止での骨抜きを reject）
#   - ファイル全体スコープ disable   → exit 1（shebang 直後の全無効化を reject）
#
# ゲート本体は shellcheck を PATH→Docker の順で解決するため、本テストも
# 同じ実体（実 shellcheck もしくはピン留め Docker イメージ）を通る。
#
# Usage: bash scripts/shellcheck-scripts.test.sh
# Exit codes:
#   0  全ケース期待どおり
#   1  いずれかのケースが期待と異なる、または前提（shellcheck 解決）が満たせない
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GATE="$SCRIPT_DIR/shellcheck-scripts.sh"

if [[ ! -f "$GATE" ]]; then
  echo "FAIL  gate script not found at $GATE"
  exit 1
fi

# 前提チェック: shellcheck か docker のどちらかが無いとゲート自体が実行できず、
# 本テストは契約を検証できない。黙って PASS せず BLOCKED として非ゼロ終了する。
if ! command -v shellcheck >/dev/null 2>&1 && ! command -v docker >/dev/null 2>&1; then
  echo "BLOCKED  neither 'shellcheck' nor 'docker' available — cannot exercise the gate"
  exit 1
fi

TMP_ROOT="$(mktemp -d)"
trap 'rm -rf "$TMP_ROOT"' EXIT

failures=0

# ShellCheck の disable ディレクティブを「テストソース内の素のディレクティブ行」として
# 書くと、このテストファイル自身を lint した際に本物の指令と解釈されてしまう。
# そこでディレクティブ・トークンは実行時に組み立て、ソース中に
# `^# shellcheck disable=...` の素行を一切残さない（HASH = リテラル '#'）。
HASH='#'
DISABLE_ALL="$HASH shellcheck disable=all"
DISABLE_LIST_ALL="$HASH shellcheck disable=SC2034,all"
DISABLE_NAMED_ALL="$HASH shellcheck disable=add-default-case,all"
DISABLE_SC2086="$HASH shellcheck disable=SC2086"
DISABLE_SC2034="$HASH shellcheck disable=SC2034"

# fixture を <case>/scripts/ に配置し、本物のゲートをそこへコピーして実行する。
# ゲートは $SCRIPT_DIR/*.sh（＝自身の隣の .sh）を検査するため、fixture スクリプトを
# ゲートと同じ scripts/ に置けば、それらが検査対象になる。
run_case() {
  local name="$1" expected_exit="$2" fixture_name="$3" fixture_body="$4"
  local casedir="$TMP_ROOT/$name"
  mkdir -p "$casedir/scripts"
  cp "$GATE" "$casedir/scripts/shellcheck-scripts.sh"
  printf '%s\n' "$fixture_body" > "$casedir/scripts/$fixture_name"

  local out actual_exit
  out="$(bash "$casedir/scripts/shellcheck-scripts.sh" 2>&1)" && actual_exit=0 || actual_exit=$?

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

# 1. clean なスクリプト → ゲート通過（ゲート自身 + clean fixture）
run_case "good-clean-script" 0 "sample.sh" '#!/usr/bin/env bash
set -euo pipefail
greeting="hello"
echo "$greeting"'

# 2. 本物のシェルバグ: 代入したが未使用の変数 (SC2034 warning) → reject
#    SC2034 は severity=warning。実スクリプトで実際に修正したのと同じ退行クラスで、
#    text-only の grep では「使われていない」を確実には判定できないが AST なら拾う。
#    （未クォート変数 SC2086 は info 扱いで warning ゲートを通すため採用しない。）
run_case "bad-unused-var" 1 "sample.sh" '#!/usr/bin/env bash
set -euo pipefail
leftover="never consumed"
echo "done"'

# 3. 整形バイパス: 行継続 `\` をまたぐ論理行のバグ → reject
#    代入値を `\` で複数物理行に分割しても、ShellCheck は 1 論理行として解析し
#    未使用 (SC2034) を検出する。1 行単位の text 照合なら見落としうるケース。
run_case "bad-continuation-line-hidden" 1 "sample.sh" '#!/usr/bin/env bash
set -euo pipefail
result="$(echo one) \
$(echo two)"
echo "fixed output"'

# 4. disable=all による一括抑止 → reject（ゲートの骨抜き防止ガード）
#    ディレクティブ行は実行時に組み立て、このテストソースに素行を残さない。
run_case "bad-disable-all" 1 "sample.sh" "#!/usr/bin/env bash
$DISABLE_ALL
set -euo pipefail
path=\$1
rm -rf \$path"

# 4b. カンマ区切りリストの末尾に紛れ込ませた `all`（disable=SC2034,all）→ reject
#     `all` は先頭以外にも書けるため、これを弾けないと末尾 `,all` で全無効化を
#     忍び込ませる整形バイパスがすり抜ける。実コード行の後ろに置いても reject される。
run_case "bad-disable-list-with-all" 1 "sample.sh" "#!/usr/bin/env bash
set -euo pipefail
echo \"code\"
$DISABLE_LIST_ALL
leftover=\"unused\""

# 4c. ハイフンを含む named optional check と `all` の混在（disable=add-default-case,all）→ reject
#     ShellCheck の named check はハイフンを含むため、先頭要素クラスに `-` を入れないと
#     この形で末尾 `,all` を忍び込ませる抜け道が残る。
run_case "bad-disable-named-check-with-all" 1 "sample.sh" "#!/usr/bin/env bash
set -euo pipefail
echo \"code\"
$DISABLE_NAMED_ALL
leftover=\"unused\""

# 5. ファイル全体スコープ disable（shebang 直後の disable）→ reject
run_case "bad-file-scope-disable" 1 "sample.sh" "#!/usr/bin/env bash
$DISABLE_SC2086
set -euo pipefail
path=\$1
rm -rf \$path"

# 6. 行内スコープ disable（1 行に限定・正当な抑止）は許容 → 通過
#    実 warning (SC2034) を出す変数を、直前行の disable で 1 行だけ正当に抑止する。
#    この disable は実コード行の直前＝ファイル全体スコープではないためガードは反応せず、
#    ShellCheck もその 1 行だけ抑止して clean になる。case 5（ファイル全体 disable）が
#    reject されるのと対になり「正当な行内抑止は通す／骨抜きの全体抑止は弾く」を示す。
run_case "good-inline-scoped-disable" 0 "sample.sh" "#!/usr/bin/env bash
set -euo pipefail
$DISABLE_SC2034
exported_for_parent=\"value\""

echo "----"
if [[ "$failures" -gt 0 ]]; then
  echo "RESULT  $failures case(s) failed"
  exit 1
fi
echo "RESULT  all cases passed"
exit 0

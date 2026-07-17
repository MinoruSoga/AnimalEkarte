#!/usr/bin/env bash
# scripts/shellcheck-scripts.sh
#
# scripts/*.sh をまとめて shellcheck にかける検証ゲート。
# `make reset` の wait-set 契約チェック (check-reset-wait-services.sh) と同じく、
# シェルスクリプトの退行を手動レビューではなく自動で弾くために make ci で実行する。
#
# 設計上のポイント:
#   1. shellcheck は AST 解析なので、空白・行継続・整形トリックでは欺けない。
#      naive な grep ベースの検査と違い、`some_cmd \` で次行に逃がしたコードも解析する。
#   2. 検査対象は scripts/*.sh を毎回 glob で動的に列挙する。固定リストにすると
#      新規スクリプトが無検査ですり抜けるため、追加した .sh は自動でゲート対象になる。
#   3. ゲートを骨抜きにする一括抑止ディレクティブ（# shellcheck disable=... を
#      shebang 直後＝ファイル全体スコープに置く、または disable=all）を別途 reject する。
#      これがないと「shellcheck disable を足して握りつぶす」で契約をすり抜けられる。
#
# ShellCheck 本体は次の優先順で解決する（決定的・再現可能なパスを保証）:
#   1. PATH 上の `shellcheck`（ローカル開発機 / CI ランナーに入っている場合）
#   2. Docker のピン留めイメージ koalaman/shellcheck:<SHELLCHECK_IMAGE_TAG>
#   どちらも無ければ環境不備として非ゼロ終了する（チェックを黙ってスキップしない）。
#
# Usage: bash scripts/shellcheck-scripts.sh
# Exit codes:
#   0  対象スクリプトが shellcheck (severity=warning) を通り、抑止ディレクティブ違反もない
#   1  shellcheck 違反、一括抑止ディレクティブ、または shellcheck 本体が見つからない
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# severity: warning 以上を fail にする。info レベルの整形 nit はゲートしない。
SHELLCHECK_SEVERITY="${SHELLCHECK_SEVERITY:-warning}"
# Docker fallback 用のピン留めタグ（再現性のため固定）。
SHELLCHECK_IMAGE_TAG="${SHELLCHECK_IMAGE_TAG:-v0.10.0}"

# 検査対象: scripts/*.sh を動的に列挙する（固定リストにしない）。
# nullglob で「.sh が 0 件」のときに glob リテラルを掴まないようにする。
shopt -s nullglob
targets=("$SCRIPT_DIR"/*.sh)
shopt -u nullglob

if [[ ${#targets[@]} -eq 0 ]]; then
  echo "FAIL  no scripts/*.sh found to lint (expected at least this gate script)"
  exit 1
fi

# ── ガード: ゲートを握りつぶす一括抑止ディレクティブを reject する ──────────
# ShellCheck の disable ディレクティブは「直後の 1 コマンド」スコープと
# 「ファイル全体」スコープがある。後者（shebang の直後に置かれた disable）は
# ファイル丸ごと検査を無効化できるため、ここで静的に検出して fail させる。
# また disable=all はスコープに関わらず全チェック無効化なので常に reject する。
# これにより「shellcheck disable を足してすり抜ける」整形バイパスを塞ぐ。
suppression_violation=0
for f in "${targets[@]}"; do
  rel="${f#"$SCRIPT_DIR/"}"

  # `all` を含む disable は常に禁止（どのスコープでも全チェック無効化になる）。
  # ディレクティブ行＝コメントが `shellcheck disable=` で「始まる」形のみを対象にする。
  # 行頭アンカー (^[[:space:]]*#[[:space:]]*shellcheck) により、説明文中の
  # 引用（echo "... '# shellcheck disable=all' ..." 等）は誤検出しない。
  # `all` はカンマ区切りリストのどの位置にも書ける（"disable=all" / "disable=SC2034,all" /
  # "disable=all,SC2086" 等）ため、`all` が完全な 1 要素として現れる全形を弾く。これを
  # しないと末尾 ",all" で全無効化を忍び込ませる整形バイパスがすり抜ける。
  # パターンは単語境界エスケープ (\< \>) を使わず、要素の前後（disable= か `,` と、
  # `,`/空白/行末）を明示して BSD/GNU 双方の grep で移植可能にする。これにより
  # `disable=allow` や `disable=ballast` のような "all" を含むコードは誤検出しない。
  # 先頭要素のクラスにはハイフンを含める（[A-Za-z0-9-]）。shellcheck の named optional
  # check（add-default-case / require-double-brackets 等）はハイフンを含むため、これを
  # 許さないと `disable=add-default-case,all` で末尾 `,all` を忍び込ませる抜け道が残る。
  # ハイフンはクラス末尾に置きリテラル扱いにする（範囲指定にしない）。
  if grep -nE '^[[:space:]]*#[[:space:]]*shellcheck[[:space:]]+disable=([A-Za-z0-9-]+,)*all([,[:space:]]|$)' "$f" >/dev/null 2>&1; then
    echo "FAIL  $rel uses a blanket 'all' in a disable directive (forbidden)"
    suppression_violation=1
    continue
  fi

  # ファイル全体スコープの disable（shebang 直後・最初の実コード前に置かれた
  # disable ディレクティブ）を検出する。先頭の shebang / 空行 / コメント行を
  # 読み飛ばし、最初に現れた非コメント実コード行より前に disable があれば
  # それはファイル全体スコープ＝ゲート骨抜きとみなす。
  file_scope=$(awk '
    NR == 1 && /^#!/        { next }              # shebang は無視
    /^[[:space:]]*$/        { next }              # 空行は無視
    /^[[:space:]]*#[[:space:]]*shellcheck[[:space:]]+disable=/ { print "HIT"; exit }
    /^[[:space:]]*#/        { next }              # 通常コメントは読み飛ばす
    { exit }                                      # 最初の実コード行に到達 → 探索終了
  ' "$f")
  if [[ "$file_scope" == "HIT" ]]; then
    echo "FAIL  $rel places a file-scope '# shellcheck disable=' before any code (forbidden)"
    suppression_violation=1
  fi
done

if [[ "$suppression_violation" -ne 0 ]]; then
  echo "----"
  echo "Blanket/file-scope shellcheck suppressions defeat the gate. Fix the underlying"
  echo "finding, or scope a 'disable' to the single offending line with justification."
  exit 1
fi

# ── shellcheck 本体を解決して実行する ──────────────────────────────────────
run_shellcheck() {
  if command -v shellcheck >/dev/null 2>&1; then
    shellcheck --severity="$SHELLCHECK_SEVERITY" "${targets[@]}"
    return $?
  fi

  if command -v docker >/dev/null 2>&1; then
    echo "shellcheck not on PATH — using koalaman/shellcheck:$SHELLCHECK_IMAGE_TAG via Docker" >&2
    # コンテナ内では scripts/ をマウントし、相対パスで対象を渡す。
    local rels=()
    local f
    for f in "${targets[@]}"; do
      rels+=("scripts/${f#"$SCRIPT_DIR/"}")
    done
    docker run --rm -v "$SCRIPT_DIR/..:/mnt" -w /mnt \
      "koalaman/shellcheck:$SHELLCHECK_IMAGE_TAG" \
      --severity="$SHELLCHECK_SEVERITY" "${rels[@]}"
    return $?
  fi

  echo "FAIL  neither 'shellcheck' nor 'docker' is available — cannot run the shell lint gate" >&2
  echo "      install shellcheck (brew install shellcheck) or start Docker" >&2
  return 1
}

if run_shellcheck; then
  echo "PASS  shellcheck (severity=$SHELLCHECK_SEVERITY) clean across ${#targets[@]} script(s) in scripts/"
  exit 0
else
  echo "----"
  echo "shellcheck found issues at severity>=$SHELLCHECK_SEVERITY. Fix them above."
  exit 1
fi

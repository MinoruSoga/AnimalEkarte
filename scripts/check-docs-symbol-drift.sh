#!/usr/bin/env bash
# scripts/check-docs-symbol-drift.sh
#
# docs/ が言及するコードシンボル・宣言数値が実装に実在するかを検査する
# （docs 全数同期 99943aaa の再発防止。openapi_route_drift_test と同じ
# 「ドキュメント⇔実装の機械突合」パターンの docs 側ゲート）。
#
# 背景:
#   2026-07-10 の docs 全数監査で、実在しないコンポーネント名
#   （`NotionFilter`→実体は PropertyFilter 等 20 件超）と数値ドリフト
#   （テーブル数・リソース数）が広範に検出された。さらに同期完了から
#   3 時間後の FE-R3 デッドコード削除で新規ドリフトが再発した
#   （docs_full_sweep_sync_20260710）。手動監査では追随できないため、
#   「言及シンボルの実在」と「宣言数値の一致」だけを機械強制する。
#
# 検査内容:
#   1. シンボル実在: docs/screens/**/*.md + docs/DESIGN_SYSTEM.md の
#      バッククォート内から CamelCase トークン（2 ハンプ以上）・
#      use* フック名・ソースファイル名（*.ts/tsx/go/mjs）を抽出し、
#      frontend/{src,liff,line-reserve} + backend/internal のソースに
#      文字列として存在することを確認する（liveness ではなく existence）。
#   2. 宣言数値: docs が宣言する テーブル数 / 権限リソース数 /
#      ハンドラー数 / 配信トリガー数 / 画面数 を実装から算出した値と突合する。
#
# 誤検知の除外は ALLOWLIST に「なぜコードに存在しなくてよいか」の理由
# コメント付きで追加すること。理由のない追加は禁止。
#
# このチェックは Docker を必要とせず、テキストだけを検査する。
#
# Usage: bash scripts/check-docs-symbol-drift.sh
#   DOCS_DRIFT_ROOT=<dir> でリポジトリルートを上書き可能（テスト用）。
# Exit codes:
#   0  全シンボル実在・全数値一致
#   1  ドリフト検出（実在しないシンボル／数値不一致）／解析失敗
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="${DOCS_DRIFT_ROOT:-$(cd "$SCRIPT_DIR/.." && pwd)}"

# ── 誤検知 allowlist（コード非依存のカタカナ英語・外部名など。理由必須） ──
ALLOWLIST=(
  # 外部サービス/製品名（コードシンボルではない）
  "GitHub"
  "TypeScript"
  "JavaScript"
  "TailwindCss"
  # dnd-kit はパッケージ名。docs では `dnd-kit` 表記だが分割抽出で DndKit 等になる場合に備える
  "DndKit"
)

failures=0
checked_tokens=0

fail() {
  echo "FAIL  $1"
  failures=$((failures + 1))
}

# ── 前提ファイル確認 ─────────────────────────────────────────
DOCS_SCREENS="$ROOT/docs/screens"
if [[ ! -d "$DOCS_SCREENS" ]]; then
  echo "FAIL  docs/screens not found under $ROOT"
  exit 1
fi

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

# ── 1. ソースコーパス構築（本文 + ファイル名一覧） ─────────────
CORPUS="$TMP/corpus.txt"
FILENAMES="$TMP/filenames.txt"
: > "$CORPUS"
: > "$FILENAMES"
for dir in frontend/src frontend/liff frontend/line-reserve backend/internal; do
  [[ -d "$ROOT/$dir" ]] || continue
  find "$ROOT/$dir" -type f \( -name '*.ts' -o -name '*.tsx' -o -name '*.go' -o -name '*.mjs' -o -name '*.css' \) -print >> "$FILENAMES.paths"
done
if [[ ! -s "$FILENAMES.paths" ]]; then
  echo "FAIL  no source files found under $ROOT (frontend/backend missing?)"
  exit 1
fi
xargs cat < "$FILENAMES.paths" >> "$CORPUS" 2>/dev/null || true
while IFS= read -r p; do basename "$p"; done < "$FILENAMES.paths" | sort -u > "$FILENAMES"

in_allowlist() {
  local tok="$1"
  local a
  for a in "${ALLOWLIST[@]}"; do
    [[ "$tok" == "$a" ]] && return 0
  done
  return 1
}

# ── 2. シンボル実在チェック ─────────────────────────────────
DOC_FILES="$TMP/doc_files.txt"
{
  find "$DOCS_SCREENS" -type f -name '*.md'
  if [[ -f "$ROOT/docs/DESIGN_SYSTEM.md" ]]; then
    echo "$ROOT/docs/DESIGN_SYSTEM.md"
  fi
} | sort > "$DOC_FILES"

while IFS= read -r doc; do
  rel="${doc#"$ROOT"/}"
  # バッククォート内トークンを単語分解して候補抽出
  grep -oE '`[^`]{1,120}`' "$doc" 2>/dev/null \
    | tr -d '`' \
    | tr -c 'A-Za-z0-9_.-' '\n' \
    | sort -u > "$TMP/words.txt" || true

  while IFS= read -r w; do
    [[ -n "$w" ]] || continue
    # 小文字を1文字も含まないトークン（16進カラーコード等）は対象外
    [[ "$w" =~ [a-z] ]] || continue
    if [[ "$w" =~ ^[A-Z][a-z0-9]+([A-Z][a-z0-9]+)+$ || "$w" =~ ^use[A-Z][A-Za-z0-9]+$ ]]; then
      # CamelCase コンポーネント/型名 or use* フック名
      in_allowlist "$w" && continue
      checked_tokens=$((checked_tokens + 1))
      if ! grep -qF "$w" "$CORPUS"; then
        fail "$rel: シンボル \`$w\` がソースコードに存在しない"
      fi
    elif [[ "$w" =~ ^[a-zA-Z0-9_-]+\.(ts|tsx|go|mjs)$ ]]; then
      # ソースファイル名の実在（basename 一致）
      in_allowlist "$w" && continue
      checked_tokens=$((checked_tokens + 1))
      if ! grep -qFx "$w" "$FILENAMES"; then
        fail "$rel: ファイル \`$w\` がソースツリーに存在しない"
      fi
    fi
  done < "$TMP/words.txt"
done < "$DOC_FILES"

# ── 3. 宣言数値チェック ─────────────────────────────────────
# extract_claims <context-regex> <file...> : マッチ断片から数字列を出力
extract_claims() {
  local regex="$1"
  shift
  local f
  for f in "$@"; do
    [[ -f "$f" ]] || continue
    grep -oE "$regex" "$f" 2>/dev/null | tr -cd '0-9\n' | grep -E '^[0-9]+$' || true
  done
}

# check_number <label> <expected> <context-regex> <file...>
check_number() {
  local label="$1" expected="$2" regex="$3"
  shift 3
  local n
  while IFS= read -r n; do
    [[ -n "$n" ]] || continue
    if [[ "$n" != "$expected" ]]; then
      fail "$label: docs の宣言値 $n が実装の実測値 $expected と不一致"
    fi
  done < <(extract_claims "$regex" "$@")
}

# 3a. テーブル数（正 = 001_init.sql の行頭 CREATE TABLE。コメント行を含めないこと）
# 「全nテーブル」形式の総数宣言のみを対象とし、部分集合の言及（「この5テーブル」等）は対象外。
if [[ -f "$ROOT/backend/migrations/001_init.sql" ]]; then
  tables="$(grep -c '^CREATE TABLE' "$ROOT/backend/migrations/001_init.sql" || true)"
  check_number "テーブル数" "$tables" '全[^0-9]{0,6}[0-9]+[^0-9]{0,6}テーブル|[0-9]+ Tables' \
    "$ROOT/docs/ERD.md" "$ROOT/docs/README.md" "$ROOT/docs/SPECIFICATION.md"
  check_number "テーブル数" "$tables" '[0-9]+ テーブル' \
    "$ROOT/docs/architecture.md"
fi

# 3b. 権限リソース数（正 = permission.go の Resource 定数）
if [[ -f "$ROOT/backend/internal/model/permission.go" ]]; then
  resources="$(grep -cE '^	Resource[A-Za-z0-9]+ +Resource += ' "$ROOT/backend/internal/model/permission.go" || true)"
  check_number "権限リソース数" "$resources" '[0-9]+ ?種類のリソース|[0-9]+ Resources|全[^0-9]{0,4}[0-9]+[^0-9]{0,4}リソース' \
    "$ROOT/docs/AUTH.md" "$ROOT/docs/README.md"
fi

# 3c. ハンドラー数（正 = *_handler.go のファイル数）
if [[ -d "$ROOT/backend/internal/handler" ]]; then
  handlers="$(find "$ROOT/backend/internal/handler" -maxdepth 1 -name '*_handler.go' | wc -l | tr -d ' ')"
  check_number "ハンドラー数" "$handlers" '[0-9]+[^0-9]{0,4}(ハンドラー|handler\.go.{0,3}[0-9]+ ファイル)' \
    "$ROOT/docs/architecture.md" "$ROOT/docs/SPECIFICATION.md"
fi

# 3d. 配信トリガー数（正 = TriggerType 定数）
if [[ -f "$ROOT/backend/internal/model/lstep_delivery_trigger_log.go" ]]; then
  triggers="$(grep -cE '^	Trigger[A-Za-z0-9]+ +TriggerType += ' "$ROOT/backend/internal/model/lstep_delivery_trigger_log.go" || true)"
  check_number "配信トリガー数" "$triggers" '[0-9]+ ?種の自動配信トリガー|[0-9]+ ?種の配信トリガー|[0-9]+種配信トリガー|[0-9]+ 配信トリガー' \
    "$ROOT/docs/architecture.md" "$ROOT/docs/line/README.md" "$ROOT/docs/line/lstep-integration.md"
  check_number "配信トリガー数" "$triggers" '配信トリガー[^0-9]{0,2}[0-9]+' \
    "$ROOT/docs/line/CLAUDE.md"
fi

# 3e. 画面数（正 = docs/screens/ 直下の番号付き .md）
screens="$(find "$DOCS_SCREENS" -maxdepth 1 -name '[0-9]*.md' | wc -l | tr -d ' ')"
check_number "画面数" "$screens" '[0-9]+[^0-9]{0,2}画面' \
  "$ROOT/docs/README.md" "$DOCS_SCREENS/README.md"

# ── 結果 ─────────────────────────────────────────────────────
if [[ "$failures" -gt 0 ]]; then
  echo "NG    docs-symbol-drift: ${failures} 件のドリフト（検査トークン ${checked_tokens} 件）"
  exit 1
fi
echo "OK    docs-symbol-drift: ドリフトなし（検査トークン ${checked_tokens} 件）"

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
#   1. シンボル実在: docs/spec/screens/**/*.md + docs/spec/design-system.md の
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
DOCS_SCREENS="$ROOT/docs/spec/screens"
if [[ ! -d "$DOCS_SCREENS" ]]; then
  echo "FAIL  docs/spec/screens not found under $ROOT"
  exit 1
fi

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

# ── 0. docs/ 直下の構造 allowlist ─────────────────────────────
# 2026-07-16 再編（architecture|spec|ops|delivery ＋ README/product-philosophy）を機械強制する。
# 再編当日に旧 docs/infra/ へ新ファイルが置かれる違反が実際に発生したため、
# シンボル・数値と同様にフォルダ規律もこのゲートで守る。
# 新カテゴリを正式追加する場合は docs/README.md の構成表と本 allowlist を同コミットで更新すること。
TOP_DIR_ALLOWLIST=(architecture spec ops delivery work)
TOP_FILE_ALLOWLIST=("README.md" "product-philosophy.md")
for p in "$ROOT"/docs/*; do
  base="$(basename "$p")"
  if [[ -d "$p" ]]; then
    ok=false
    for a in "${TOP_DIR_ALLOWLIST[@]}"; do [[ "$base" == "$a" ]] && ok=true && break; done
    [[ "$ok" == true ]] || fail "docs/ 直下に allowlist 外のフォルダ \`docs/$base/\`（正: architecture|spec|ops|delivery。docs/README.md の規律参照）"
  elif [[ -f "$p" ]]; then
    ok=false
    for a in "${TOP_FILE_ALLOWLIST[@]}"; do [[ "$base" == "$a" ]] && ok=true && break; done
    [[ "$ok" == true ]] || fail "docs/ 直下に allowlist 外のファイル \`docs/$base\`（直下は README.md と product-philosophy.md のみ。文書は 4 カテゴリ配下へ）"
  fi
done

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
  if [[ -f "$ROOT/docs/spec/design-system.md" ]]; then
    echo "$ROOT/docs/spec/design-system.md"
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
      # -w: 単語境界一致。部分文字列一致だと「実在しない PetSelectionPage が
      # 実在する usePetSelectionPage の内部にマッチしてすり抜ける」ため必須。
      if ! grep -qwF "$w" "$CORPUS"; then
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

# 3a. テーブル数（正 = 全 backend/migrations/*.sql 直下の行頭 CREATE TABLE 合算。seeds/ は対象外）
# 「全nテーブル」形式の総数宣言のみを対象とし、部分集合の言及（「この5テーブル」等）は対象外。
# find+grep: BSD/GNU 両対応（macOS の grep は --include 非対応のため）。-maxdepth 1 で seeds/ を除外。
if [[ -d "$ROOT/backend/migrations" ]]; then
  tables="$(
    {
      find "$ROOT/backend/migrations" -maxdepth 1 -name '*.sql' -type f -print0 2>/dev/null \
        | xargs -0 grep -h '^CREATE TABLE' 2>/dev/null || true
    } | wc -l | tr -d ' '
  )"
  tables="${tables:-0}"
  # 総数の正本は ERD + specification（ゲートが強制する宣言面）。
  # docs/README.md / screens/README.md / overview の索引フッタは別 unit で追随する。
  check_number "テーブル数" "$tables" '全[^0-9]{0,6}[0-9]+[^0-9]{0,6}テーブル|[0-9]+ Tables' \
    "$ROOT/docs/architecture/erd.md" "$ROOT/docs/spec/specification.md"
  check_number "テーブル数" "$tables" '[0-9]+ テーブル' \
    "$ROOT/docs/spec/specification.md"

  # ERD §1.1 のドメイン在庫は old_db の移行レポート生成が読む正本でもある。
  # 上段の「全nテーブル」宣言だけが追随し、個別テーブル名が欠落するドリフトを
  # 生成前に検出するため、migration の物理テーブル集合と完全一致させる。
  migration_tables="$TMP/migration_tables.txt"
  erd_tables="$TMP/erd_tables.txt"
  {
    find "$ROOT/backend/migrations" -maxdepth 1 -name '*.sql' -type f -print0 2>/dev/null \
      | xargs -0 grep -h '^CREATE TABLE' 2>/dev/null || true
  } | sed -E 's/^CREATE TABLE( IF NOT EXISTS)? ([a-zA-Z_][a-zA-Z0-9_]*).*/\2/' \
    | sort -u > "$migration_tables"
  awk '
    /^### 1\.1 / { in_domain_inventory = 1; next }
    /^## 2\./ { in_domain_inventory = 0 }
    in_domain_inventory && /^\|[[:space:]]*\*\*/ { print }
  ' "$ROOT/docs/architecture/erd.md" \
    | grep -oE '`[a-z][a-z0-9_]*`' \
    | tr -d '`' \
    | sort -u > "$erd_tables" || true

  missing_from_erd="$(comm -23 "$migration_tables" "$erd_tables" | paste -sd, -)"
  missing_from_migration="$(comm -13 "$migration_tables" "$erd_tables" | paste -sd, -)"
  if [[ -n "$missing_from_erd" || -n "$missing_from_migration" ]]; then
    fail "ERD §1.1 テーブル在庫が migration と不一致（ERD欠落=${missing_from_erd:-なし}; migration欠落=${missing_from_migration:-なし}）"
  fi
fi

# 3b. 権限リソース数（正 = permission.go の Resource 定数）
if [[ -f "$ROOT/backend/internal/model/permission.go" ]]; then
  resources="$(grep -cE '^	Resource[A-Za-z0-9]+ +Resource += ' "$ROOT/backend/internal/model/permission.go" || true)"
  check_number "権限リソース数" "$resources" '[0-9]+ ?種類のリソース|[0-9]+ Resources|全[^0-9]{0,4}[0-9]+[^0-9]{0,4}リソース' \
    "$ROOT/docs/architecture/auth.md" "$ROOT/docs/README.md" "$DOCS_SCREENS/README.md"
fi

# 3c. ハンドラー数（正 = *_handler.go のファイル数）
if [[ -d "$ROOT/backend/internal/handler" ]]; then
  handlers="$(find "$ROOT/backend/internal/handler" -maxdepth 1 -name '*_handler.go' | wc -l | tr -d ' ')"
  check_number "ハンドラー数" "$handlers" '[0-9]+[^0-9]{0,4}(ハンドラー|handler\.go.{0,3}[0-9]+ ファイル)' \
    "$ROOT/docs/architecture/overview.md" "$ROOT/docs/spec/specification.md"
fi

# 3d. 配信トリガー数（正 = TriggerType 定数）
if [[ -f "$ROOT/backend/internal/model/lstep_delivery_trigger_log.go" ]]; then
  triggers="$(grep -cE '^	Trigger[A-Za-z0-9]+ +TriggerType += ' "$ROOT/backend/internal/model/lstep_delivery_trigger_log.go" || true)"
  check_number "配信トリガー数" "$triggers" '[0-9]+ ?種の自動配信トリガー|[0-9]+ ?種の配信トリガー|[0-9]+種配信トリガー|[0-9]+ 配信トリガー' \
    "$ROOT/docs/architecture/overview.md" "$ROOT/docs/spec/line/README.md" "$ROOT/docs/spec/line/lstep-integration.md"
  check_number "配信トリガー数" "$triggers" '配信トリガー[^0-9]{0,2}[0-9]+' \
    "$ROOT/docs/spec/line/CLAUDE.md"
fi

# 3e. 画面数（正 = docs/spec/screens/ 直下の番号付き .md）
screens="$(find "$DOCS_SCREENS" -maxdepth 1 -name '[0-9]*.md' | wc -l | tr -d ' ')"
check_number "画面数" "$screens" '[0-9]+[^0-9]{0,2}画面' \
  "$ROOT/docs/README.md" "$DOCS_SCREENS/README.md"

# ── 結果 ─────────────────────────────────────────────────────
if [[ "$failures" -gt 0 ]]; then
  echo "NG    docs-symbol-drift: ${failures} 件のドリフト（検査トークン ${checked_tokens} 件）"
  exit 1
fi
echo "OK    docs-symbol-drift: ドリフトなし（検査トークン ${checked_tokens} 件）"

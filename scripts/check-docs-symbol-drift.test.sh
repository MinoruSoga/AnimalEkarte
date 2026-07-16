#!/usr/bin/env bash
# scripts/check-docs-symbol-drift.test.sh
#
# check-docs-symbol-drift.sh の回帰テスト。
# 一時ディレクトリに docs + ソースツリーの fixture を作り、
#   - 実在シンボル・正しい宣言数値のみ                → exit 0
#   - 実在しないコンポーネント名（NotionFilter クラス）→ exit 1
#   - 実在しない use* フック名                        → exit 1
#   - 実在しないソースファイル名                      → exit 1
#   - テーブル総数宣言の不一致（全nテーブル）          → exit 1
#   - 部分集合の言及（「この5テーブル」）は総数扱いしない → exit 0
#   - 16進カラーコード（`A39E98` 等）は検査対象外       → exit 0
# で判定できることを確認する。Docker 不要・純テキスト検査。
#
# Usage: bash scripts/check-docs-symbol-drift.test.sh
# Exit codes:
#   0  全ケース PASS
#   1  いずれかのケースが期待と異なる
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHECK="$SCRIPT_DIR/check-docs-symbol-drift.sh"

if [[ ! -f "$CHECK" ]]; then
  echo "FAIL  check script not found at $CHECK"
  exit 1
fi

TMP_ROOT="$(mktemp -d)"
trap 'rm -rf "$TMP_ROOT"' EXIT

failures=0

# build_fixture <dir> : 検査が通る最小 fixture を構築する
build_fixture() {
  local d="$1"
  mkdir -p "$d/docs/spec/screens" "$d/docs/architecture" "$d/frontend/src/components" \
    "$d/backend/migrations" "$d/backend/internal/model" "$d/backend/internal/handler"

  cat > "$d/frontend/src/components/GoodWidget.tsx" <<'EOF'
export function GoodWidget() { return null; }
export function useGoodHook() { return 1; }
EOF
  printf 'CREATE TABLE a (id int);\nCREATE TABLE b (id int);\n' > "$d/backend/migrations/001_init.sql"
  printf 'package model\n\nconst (\n\tResourceAlpha Resource = "alpha"\n\tResourceBeta Resource = "beta"\n)\n' \
    > "$d/backend/internal/model/permission.go"
  printf 'package model\n\nconst (\n\tTriggerTypeFoo TriggerType = "foo"\n)\n' \
    > "$d/backend/internal/model/lstep_delivery_trigger_log.go"
  printf 'package handler\n' > "$d/backend/internal/handler/one_handler.go"

  cat > "$d/docs/spec/screens/01-sample.md" <<'EOF'
# サンプル画面
- **`GoodWidget`**: 実在するコンポーネント。
- **`useGoodHook`**: 実在するフック。
- 実装は `GoodWidget.tsx` を参照。
- 色は `A39E98` を使用（16進コードは検査対象外）。
EOF
  cat > "$d/docs/spec/screens/README.md" <<'EOF'
全1画面のインデックス。
EOF
  cat > "$d/docs/architecture/erd.md" <<'EOF'
全2テーブルの設計。なお、この5テーブルという表現は部分集合の言及であり総数ではない。
EOF
  cat > "$d/docs/README.md" <<'EOF'
索引 (2 Tables / 2 Resources)。1画面。
EOF
  cat > "$d/docs/architecture/auth.md" <<'EOF'
権限は 2 種類のリソースで構成される。
EOF
  cat > "$d/docs/architecture/overview.md" <<'EOF'
実装規模: 2 テーブル、1 ハンドラー、1 配信トリガー。
EOF
  cat > "$d/docs/spec/specification.md" <<'EOF'
全 **2 テーブル** を audit_logs で追跡。
EOF
}

# run_case <name> <expected-exit> <mutation-fn>
run_case() {
  local name="$1" expected="$2" mutate="$3"
  local d="$TMP_ROOT/$name"
  build_fixture "$d"
  "$mutate" "$d"
  local actual=0
  DOCS_DRIFT_ROOT="$d" bash "$CHECK" > "$TMP_ROOT/$name.out" 2>&1 || actual=$?
  if [[ "$actual" -ne "$expected" ]]; then
    echo "FAIL  case=$name expected=$expected actual=$actual"
    sed 's/^/      /' "$TMP_ROOT/$name.out"
    failures=$((failures + 1))
  else
    echo "PASS  case=$name (exit=$actual)"
  fi
}

mutate_none() { :; }

mutate_phantom_component() {
  echo '- **`PhantomWidget`**: 実在しないコンポーネント。' >> "$1/docs/spec/screens/01-sample.md"
}

mutate_phantom_hook() {
  echo '- **`usePhantomHook`**: 実在しないフック。' >> "$1/docs/spec/screens/01-sample.md"
}

mutate_phantom_file() {
  echo '- 実装は `ghost-file.ts` を参照。' >> "$1/docs/spec/screens/01-sample.md"
}

mutate_wrong_table_total() {
  # 総数宣言を実測(2)と異なる 3 に書き換える
  printf '全3テーブルの設計。\n' > "$1/docs/architecture/erd.md"
}

run_case "clean"             0 mutate_none
run_case "phantom-component" 1 mutate_phantom_component
run_case "phantom-hook"      1 mutate_phantom_hook
run_case "phantom-file"      1 mutate_phantom_file
run_case "wrong-table-total" 1 mutate_wrong_table_total

if [[ "$failures" -gt 0 ]]; then
  echo "NG    docs-symbol-drift self-test: ${failures} case(s) failed"
  exit 1
fi
echo "OK    docs-symbol-drift self-test: all cases passed"

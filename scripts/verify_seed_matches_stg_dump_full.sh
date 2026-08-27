#!/usr/bin/env bash
# scripts/verify_seed_matches_stg_dump_full.sh
#
# 機械的証明: seed(全 DDL migration + 002/003/004 CSV バンドル) が STG dump と全テーブルで
# 一致することを確認する。
#
# DB-A: migrations/*.sql (DDL) を昇順適用 → seeds/{002_master,003_demo,004_staging}/*.csv を
#       manifest.json のテーブル順・固定バンドル順 (002→003→004) で \copy ロード
#       （2026-07 の stub SQL 削除後は 002-004 に .sql ファイルが存在しないため、
#       cmd/migrate と同じ CSV ロード経路をこのスクリプト内で再現する）
# DB-B: migrations/*.sql のスキーマ + dump の INSERT 文のみ適用
#       (session_replication_role='replica' で FK 順序を無視, schema_migrations はスキップ)
#
# 比較対象外カラム (一致不可能 — 除外理由を明記):
#   created_at  : seed は INSERT 時の NOW() で埋まる。dump は実際の STG 生成時刻。値が異なる。
#   updated_at  : 同上。ON CONFLICT DO UPDATE でも NOW() になるため一致しない。
#   deleted_at  : 明示値 (または NULL) のため比較対象に含む。
#
# 比較対象外テーブル:
#   schema_migrations : Go migrate runner が実行時に動的生成。001_init.sql に定義なし。
#
# シークレットカラム (SHA256 ハッシュ化して比較 — 平文を出力しない):
#   accounts.password_hash
#   clinic_integrations.key_value
#
# 前提: Docker が起動していること, prodData/stg_5-21/ekarte が存在すること, jq がインストール済みであること
#       (seeds/<bundle>/manifest.json のテーブル順パースに使用)
#
# 使い方:
#   bash scripts/verify_seed_matches_stg_dump_full.sh [DUMP_FILE]
#   STG_DUMP=<path> bash scripts/verify_seed_matches_stg_dump_full.sh   # 環境変数でも指定可
#   PARSE_ONLY=1   bash scripts/verify_seed_matches_stg_dump_full.sh [DUMP_FILE]  # Docker無しでパース判定のみ
#   引数も $STG_DUMP も無い場合は prodData/ekarte-stg-*.sql の最新を自動選択する。
#   dump は TablePlus 形式("public"."table") / pg_dump 形式(public.table) の双方に対応。

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# Dump path: 第1引数 > $STG_DUMP > prodData/ekarte-stg-*.sql の最新
DUMP_FILE="${1:-${STG_DUMP:-}}"
if [[ -z "$DUMP_FILE" ]]; then
  DUMP_FILE="$(ls -t "$REPO_ROOT"/prodData/ekarte-stg-*.sql 2>/dev/null | head -1 || true)"
fi
MIGRATION_DIR="$REPO_ROOT/backend/migrations"
CONTAINER="ekarte_verify_$$"
# Ephemeral credential per run — never hardcode. Host port is intentionally not
# published; all access is via `docker exec` (no 0.0.0.0 bind, no loopback publish).
POSTGRES_PASSWORD="$(openssl rand -hex 16)"
DB_A="ekarte_a"
DB_B="ekarte_b"

SECRET_COLS=(
  "accounts.password_hash"
  "clinic_integrations.key_value"
)

TMPWORK=""

cleanup_all() {
  [[ -n "$TMPWORK" ]] && rm -rf "$TMPWORK" 2>/dev/null || true
  docker rm -f "$CONTAINER" 2>/dev/null || true
}
trap cleanup_all EXIT

TMPWORK="$(mktemp -d)"

if [[ -z "$DUMP_FILE" || ! -f "$DUMP_FILE" ]]; then
  echo "ERROR: dump file not found: ${DUMP_FILE:-<none>}" >&2
  echo "       第1引数か \$STG_DUMP で指定するか prodData/ekarte-stg-*.sql を配置してください。" >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# Dump INSERT extractor — supports BOTH marker formats:
#   TablePlus : INSERT INTO "public"."table"
#   pg_dump   : INSERT INTO public.table
# schema_migrations は 001_init.sql に定義が無いため除外する（単一行・複数行両対応）。
# ---------------------------------------------------------------------------
extract_dump_inserts() {
  awk '
    /^INSERT INTO ("public"\."schema_migrations"|public\.schema_migrations)[ (]/ { skip=1 }
    skip { if (/;[[:space:]]*$/) skip=0; next }
    /^INSERT INTO ("public"\.|public\.)/ { in_block=1 }
    in_block { print }
    in_block && /;[[:space:]]*$/ { in_block=0 }
  ' "$1"
}

# Fail loudly if the dump has no parseable INSERT table data (silent empty-DB の誤検出防止)。
MARKER_COUNT=$(grep -cE '^INSERT INTO ("public"\."|public\.)' "$DUMP_FILE" || true)
EXTRACT_COUNT=$(extract_dump_inserts "$DUMP_FILE" | grep -cE '^INSERT INTO ' || true)
echo "Dump file : $DUMP_FILE"
echo "Dump parse: ${MARKER_COUNT} INSERT markers / ${EXTRACT_COUNT} extracted statements (schema_migrations 除く)"
if [[ "${MARKER_COUNT:-0}" -eq 0 || "${EXTRACT_COUNT:-0}" -eq 0 ]]; then
  echo "ERROR: dump に解析可能な INSERT テーブルデータがありません。" >&2
  echo "       対応マーカー: 'INSERT INTO \"public\".\"table\"' もしくは 'INSERT INTO public.table'" >&2
  exit 1
fi

# Parse-only mode: フィクスチャ/ユニット検証用（Docker 不要・パース判定のみ）。
if [[ "${PARSE_ONLY:-0}" == "1" ]]; then
  echo "PARSE_ONLY OK (${EXTRACT_COUNT} statements)"
  exit 0
fi

if ! command -v jq > /dev/null 2>&1; then
  echo "ERROR: jq が見つかりません。seeds/<bundle>/manifest.json のパースに必要です（brew install jq 等）。" >&2
  exit 1
fi

SEEDS_DIR="$MIGRATION_DIR/seeds"
BUNDLE_ORDER=(002_master)
DDL_FILES=("$MIGRATION_DIR"/*.sql)

# ---------------------------------------------------------------------------
# 1. Start container
# ---------------------------------------------------------------------------
echo "[1/5] Starting postgres:18-alpine container (name=$CONTAINER, no host port publish)..."
docker run -d --name "$CONTAINER" \
  -e "POSTGRES_PASSWORD=${POSTGRES_PASSWORD}" \
  -e POSTGRES_USER=postgres \
  postgres:18-alpine > /dev/null

for _ in $(seq 1 30); do
  docker exec "$CONTAINER" pg_isready -U postgres -q 2>/dev/null && break
  sleep 1
done
echo "      Ready."

# Base psql invocations
PSQL_Q="docker exec -i $CONTAINER psql -U postgres -v ON_ERROR_STOP=1 -q"
PSQL_T="docker exec -i $CONTAINER psql -U postgres -v ON_ERROR_STOP=1 -t -A"

# ---------------------------------------------------------------------------
# 2. Create databases
# ---------------------------------------------------------------------------
echo "[2/5] Creating databases $DB_A, $DB_B..."
$PSQL_Q postgres -c "CREATE DATABASE ${DB_A};"
$PSQL_Q postgres -c "CREATE DATABASE ${DB_B};"

# ---------------------------------------------------------------------------
# 3. DB-A: 全DDL + seed バンドル (002_master→003_demo→004_staging) CSVロード
#    2026-07 の stub SQL 削除後は 002-004 に .sql が無いため、cmd/migrate の
#    applyCSVBundle と同じ「manifest.json のテーブル順で \copy」をここで再現する。
#    \copy はテーブル名を manifest.json の値からそのまま使う（外部入力ではなく
#    リポジトリにコミット済みの自分自身の manifest のみを読む）。
# ---------------------------------------------------------------------------
echo "[3/5] DB-A: applying DDL migrations + seed bundles..."
for ddl in "${DDL_FILES[@]}"; do
  ddl_name="$(basename "$ddl")"
  printf "      %-48s" "$ddl_name"
  $PSQL_Q "$DB_A" < "$ddl"
  echo "OK"
done

for bundle in "${BUNDLE_ORDER[@]}"; do
  manifest="$SEEDS_DIR/$bundle/manifest.json"
  if [[ ! -f "$manifest" ]]; then
    echo "ERROR: manifest not found: $manifest" >&2
    exit 1
  fi

  while IFS=$'\t' read -r table csvfile; do
    [[ -z "$table" ]] && continue
    csvpath="$SEEDS_DIR/$bundle/$csvfile"
    printf "      %-12s %-40s" "$bundle" "$table"
    $PSQL_Q "$DB_A" -c "\copy public.\"${table}\" FROM STDIN WITH (FORMAT csv, HEADER true)" < "$csvpath"
    echo "OK"
  done < <(jq -r '.tables[] | [.table, .csvFile] | @tsv' "$manifest")
done

# ---------------------------------------------------------------------------
# 4. DB-B: current DDL schema + dump INSERTs
# ---------------------------------------------------------------------------
echo "[4/5] DB-B: applying current DDL schema + dump INSERTs..."

for ddl in "${DDL_FILES[@]}"; do
  $PSQL_Q "$DB_B" < "$ddl"
  echo "      $(basename "$ddl") schema applied."
done

{
  echo "SET session_replication_role = 'replica';"
  extract_dump_inserts "$DUMP_FILE"
} | $PSQL_Q "$DB_B"
echo "      Dump INSERT statements applied."

# Guard: DB-B が無言で空になっていないか確認（dump 形式非対応 / import 失敗の早期検出）。
DBB_CORE=$($PSQL_T "$DB_B" -c "SELECT (SELECT count(*) FROM public.clinics)+(SELECT count(*) FROM public.staffs)+(SELECT count(*) FROM public.accounts);")
if [[ "${DBB_CORE:-0}" -eq 0 ]]; then
  echo "ERROR: dump import 後も DB-B が空です (clinics+staffs+accounts=0)。dump 形式が非対応か import が失敗しています。" >&2
  exit 1
fi
echo "      DB-B core rows (clinics+staffs+accounts): $DBB_CORE"

# ---------------------------------------------------------------------------
# 5. Compare all public tables
# ---------------------------------------------------------------------------
echo "[5/5] Comparing all public tables (excluding schema_migrations)..."
echo ""

TABLES=$($PSQL_T "$DB_A" -c \
  "SELECT tablename FROM pg_tables WHERE schemaname='public' AND tablename<>'schema_migrations' ORDER BY tablename;")

FAIL_COUNT=0
FAIL_TABLES=()
OK_COUNT=0

for tbl in $TABLES; do
  # PK columns (comma-separated, ordered by attnum)
  PK_COLS=$($PSQL_T "$DB_A" -c \
    "SELECT a.attname FROM pg_index i JOIN pg_attribute a ON a.attrelid=i.indrelid AND a.attnum=ANY(i.indkey) WHERE i.indrelid='public.${tbl}'::regclass AND i.indisprimary ORDER BY a.attnum;" \
    | tr '\n' ',' | sed 's/,$//')

  # Comparable columns (exclude created_at, updated_at), ordered by ordinal_position
  COMP_COLS=$($PSQL_T "$DB_A" -c \
    "SELECT column_name FROM information_schema.columns WHERE table_schema='public' AND table_name='${tbl}' AND column_name NOT IN ('created_at','updated_at') ORDER BY ordinal_position;" \
    | tr '\n' ',' | sed 's/,$//')

  if [[ -z "$COMP_COLS" ]]; then
    printf "  SKIP %-45s (no comparable columns)\n" "$tbl"
    continue
  fi

  # Build SELECT expression — hash secret columns
  SEL_EXPRS=""
  IFS=',' read -ra COL_ARR <<< "$COMP_COLS"
  for col in "${COL_ARR[@]}"; do
    [[ -n "$SEL_EXPRS" ]] && SEL_EXPRS+=","
    is_secret=false
    for sc in "${SECRET_COLS[@]}"; do
      [[ "$sc" == "${tbl}.${col}" ]] && { is_secret=true; break; }
    done
    if $is_secret; then
      SEL_EXPRS+="CASE WHEN ${col} IS NULL THEN NULL ELSE encode(sha256(${col}::bytea),'hex') END AS ${col}"
    else
      SEL_EXPRS+="${col}"
    fi
  done

  ORDER_BY="${PK_COLS:-$COMP_COLS}"
  QUERY="SELECT ${SEL_EXPRS} FROM public.${tbl} ORDER BY ${ORDER_BY};"

  TMP_A="$TMPWORK/${tbl}_a.tsv"
  TMP_B="$TMPWORK/${tbl}_b.tsv"

  $PSQL_T "$DB_A" -F$'\t' -c "$QUERY" > "$TMP_A"
  $PSQL_T "$DB_B" -F$'\t' -c "$QUERY" > "$TMP_B"

  ROW_A=$(wc -l < "$TMP_A" | tr -d ' ')
  ROW_B=$(wc -l < "$TMP_B" | tr -d ' ')

  if [[ "$ROW_B" -eq 0 ]]; then
    printf "  SKIP %-45s (not in STG dump; seed-only demo data)\n" "$tbl"
    continue
  fi

  if diff -q "$TMP_A" "$TMP_B" > /dev/null 2>&1; then
    printf "  OK   %-45s %d rows\n" "$tbl" "$ROW_A"
    OK_COUNT=$((OK_COUNT + 1))
  else
    printf "  FAIL %-45s A=%d rows  B=%d rows\n" "$tbl" "$ROW_A" "$ROW_B"
    diff "$TMP_A" "$TMP_B" | head -10 | sed 's/^/       /' || true
    FAIL_TABLES+=("$tbl")
    FAIL_COUNT=$((FAIL_COUNT + 1))
  fi
done

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "======================================================================"
echo "除外カラム (一致不可能):"
echo "  created_at  — seed は INSERT 時の NOW(); dump は実 STG タイムスタンプ"
echo "  updated_at  — 同上"
echo "除外テーブル:"
echo "  schema_migrations — Go migrate runner が実行時生成; 001_init.sql 外"
echo "ハッシュ比較 (SHA256):"
for sc in "${SECRET_COLS[@]}"; do echo "  ${sc}"; done
echo ""
printf "Tables OK:   %d\n" "$OK_COUNT"
printf "Tables FAIL: %d\n" "$FAIL_COUNT"

if [[ $FAIL_COUNT -eq 0 ]]; then
  echo ""
  echo "ALL TABLES MATCH."
  echo "seed(001-004) は STG dump の完全な再現です。"
  exit 0
else
  echo ""
  echo "MISMATCH detected in:"
  for t in "${FAIL_TABLES[@]}"; do echo "  - ${t}"; done
  exit 1
fi

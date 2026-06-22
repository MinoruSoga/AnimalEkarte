#!/usr/bin/env bash
# scripts/dump-stg.sh
#
# STG RDS を SSM ポートフォワード経由で pg_dump し、prodData/ekarte-stg-<実行日>.sql に出力する。
# pg_dump はデフォルトで全テーブルを出力するため、TablePlus の選択エクスポートのような
# テーブル欠落が構造的に発生しない。出力は `--data-only --column-inserts` 形式で、
# scripts/verify_seed_matches_stg_dump_full.sh がそのまま消費できる。
#
# 認証:
#   - AWS: プロファイル(AnimalEkarte)が SSO 等で認証済みであること。
#   - DB : 既定で .env.staging の DB_USER / DB_NAME / DB_PASSWORD を使用する。
#          PGPASSWORD / STG_DB_USER / STG_DB_NAME を環境変数で渡せば上書きできる。
#          スクリプトはパスワードを画面に出力しない。
#
# トンネル:
#   127.0.0.1:<LOCAL_PORT> が既に開いていれば（TablePlus / stg-db-tunnel.sh 等）それを再利用する。
#   開いていなければ本スクリプトが background で起動し、終了時に自動で閉じる。
#
# 使い方: bash scripts/dump-stg.sh   /   make dump-stg

set -euo pipefail

# --- 設定 (env で上書き可) -------------------------------------------------
AWS_PROFILE="${AWS_PROFILE:-AnimalEkarte}"
AWS_REGION="${AWS_REGION:-us-east-1}"
ECS_CLUSTER="${STG_ECS_CLUSTER:-animalekarte-stg-cluster}"
ECS_SERVICE="${STG_ECS_SERVICE:-animalekarte-stg-service}"
RDS_HOST="${STG_RDS_HOST:-animalekarte-stg-db.cqbe28s44fta.us-east-1.rds.amazonaws.com}"
RDS_PORT="${STG_RDS_PORT:-5432}"
LOCAL_PORT="${STG_LOCAL_PORT:-15432}"
TUNNEL_TIMEOUT="${STG_TUNNEL_TIMEOUT:-30}"   # トンネル確立待ちの最大秒数

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_STAGING="${STG_ENV_FILE:-$REPO_ROOT/.env.staging}"
OUT_DIR="${DUMP_DIR:-$REPO_ROOT/prodData}"
OUT_FILE="$OUT_DIR/ekarte-stg-$(date +%Y%m%d).sql"

# .env.staging から KEY の値を取得 (export 前置・前後クォート除去。値は画面に出力しない)
env_val() {
  local key="$1" line val
  [[ -f "$ENV_STAGING" ]] || return 0
  line="$(grep -E "^[[:space:]]*(export[[:space:]]+)?${key}=" "$ENV_STAGING" 2>/dev/null | tail -1)" || true
  [[ -n "$line" ]] || return 0
  val="${line#*=}"
  val="${val#"${val%%[![:space:]]*}"}"; val="${val%"${val##*[![:space:]]}"}"
  case "$val" in
    \"*\") val="${val#\"}"; val="${val%\"}" ;;
    \'*\') val="${val#\'}"; val="${val%\'}" ;;
  esac
  printf '%s' "$val"
}

# DB 接続情報: 明示 env > .env.staging > 既定
DB_USER="${STG_DB_USER:-$(env_val DB_USER)}";  DB_USER="${DB_USER:-ekarte_admin}"
DB_NAME="${STG_DB_NAME:-$(env_val DB_NAME)}";  DB_NAME="${DB_NAME:-ekarte_db}"
if [[ -z "${PGPASSWORD:-}" ]]; then
  PGPASSWORD="$(env_val DB_PASSWORD)"
  export PGPASSWORD
fi

TUNNEL_PID=""
TUNNEL_LOG=""

# --- 後始末: 自分で起動したトンネルのみ閉じる -----------------------------
cleanup() {
  if [[ -n "$TUNNEL_PID" ]]; then
    echo "→ SSM トンネルを終了 (pid=$TUNNEL_PID)"
    kill "$TUNNEL_PID" 2>/dev/null || true
    wait "$TUNNEL_PID" 2>/dev/null || true
  fi
  [[ -n "$TUNNEL_LOG" && -f "$TUNNEL_LOG" ]] && rm -f "$TUNNEL_LOG"
}
trap cleanup EXIT

port_open() { nc -z 127.0.0.1 "$LOCAL_PORT" >/dev/null 2>&1; }

# --- 事前チェック ----------------------------------------------------------
if [[ -z "${PGPASSWORD:-}" ]]; then
  echo "ERROR: DB パスワードを取得できません。"
  echo "       .env.staging に DB_PASSWORD が定義されているか、または PGPASSWORD を環境変数で渡してください。"
  echo "       (参照した .env.staging: $ENV_STAGING)"
  exit 1
fi

mkdir -p "$OUT_DIR"

# --- 1. トンネル確保 -------------------------------------------------------
if port_open; then
  echo "✓ 既存トンネル(127.0.0.1:${LOCAL_PORT})を再利用"
else
  command -v aws >/dev/null || { echo "ERROR: aws CLI が見つかりません"; exit 1; }
  command -v session-manager-plugin >/dev/null || { echo "ERROR: session-manager-plugin が見つかりません"; exit 1; }
  echo "→ STG ECS タスクを取得中..."
  TASK_ARN="$(aws ecs list-tasks --cluster "$ECS_CLUSTER" --service-name "$ECS_SERVICE" \
    --region "$AWS_REGION" --profile "$AWS_PROFILE" --query 'taskArns[0]' --output text)"
  if [[ -z "$TASK_ARN" || "$TASK_ARN" == "None" ]]; then
    echo "ERROR: 実行中の STG タスクが見つかりません (cluster=$ECS_CLUSTER service=$ECS_SERVICE)"
    exit 1
  fi
  TASK_ID="${TASK_ARN##*/}"
  RUNTIME_ID="$(aws ecs describe-tasks --cluster "$ECS_CLUSTER" --tasks "$TASK_ARN" \
    --region "$AWS_REGION" --profile "$AWS_PROFILE" \
    --query 'tasks[0].containers[0].runtimeId' --output text)"
  TARGET="ecs:${ECS_CLUSTER}_${TASK_ID}_${RUNTIME_ID}"

  echo "→ SSM ポートフォワード起動: 127.0.0.1:${LOCAL_PORT} → ${RDS_HOST}:${RDS_PORT}"
  TUNNEL_LOG="$(mktemp -t dump-stg-tunnel.XXXXXX)"
  aws ssm start-session \
    --target "$TARGET" \
    --document-name AWS-StartPortForwardingSessionToRemoteHost \
    --parameters "{\"host\":[\"${RDS_HOST}\"],\"portNumber\":[\"${RDS_PORT}\"],\"localPortNumber\":[\"${LOCAL_PORT}\"]}" \
    --region "$AWS_REGION" --profile "$AWS_PROFILE" >"$TUNNEL_LOG" 2>&1 &
  TUNNEL_PID=$!

  echo "→ トンネル確立を待機中 (最大 ${TUNNEL_TIMEOUT}s)..."
  for _ in $(seq 1 "$TUNNEL_TIMEOUT"); do
    port_open && break
    if ! kill -0 "$TUNNEL_PID" 2>/dev/null; then
      echo "ERROR: SSM セッションが起動に失敗しました:"; sed 's/^/    /' "$TUNNEL_LOG"; exit 1
    fi
    sleep 1
  done
  port_open || { echo "ERROR: ${TUNNEL_TIMEOUT}s 以内にトンネルが開きませんでした:"; sed 's/^/    /' "$TUNNEL_LOG"; exit 1; }
  echo "✓ トンネル確立"
fi

# --- 2. pg_dump (PG18 必須。host pg_dump<18 は postgres:18 コンテナで代替) ---
echo "→ pg_dump → ${OUT_FILE}"
PG_DUMP_ARGS=(-h 127.0.0.1 -p "$LOCAL_PORT" -U "$DB_USER" -d "$DB_NAME"
  --data-only --column-inserts --no-owner --no-privileges)

HOST_PG_MAJOR="$(pg_dump --version 2>/dev/null | grep -oE '[0-9]+' | head -1 || echo 0)"
if command -v pg_dump >/dev/null && [[ "${HOST_PG_MAJOR:-0}" -ge 18 ]]; then
  pg_dump "${PG_DUMP_ARGS[@]}" -f "$OUT_FILE"
else
  echo "  host pg_dump=${HOST_PG_MAJOR:-none} (要 18+)。docker postgres:18 経由で実行"
  # db コンテナ(postgres:18-alpine)の pg_dump を使用。Docker Desktop(Mac) では
  # host.docker.internal でホスト loopback 上のトンネルへ到達できる。
  docker compose --env-file "$REPO_ROOT/.env.local" run --rm --no-deps -T \
    -e PGPASSWORD="$PGPASSWORD" \
    --entrypoint pg_dump db \
    -h host.docker.internal -p "$LOCAL_PORT" -U "$DB_USER" -d "$DB_NAME" \
    --data-only --column-inserts --no-owner --no-privileges >"$OUT_FILE"
fi

# --- 3. サマリ + 完全性チェック -------------------------------------------
TBL_COUNT="$(grep -oE 'INSERT INTO "?[a-zA-Z_]+"?\."?[a-zA-Z_]+"?|INSERT INTO "?[a-zA-Z_]+"?' "$OUT_FILE" \
  | sed -E 's/INSERT INTO //; s/"//g; s/^public\.//' | sort -u | wc -l | tr -d ' ')"
ROW_COUNT="$(grep -c '^INSERT INTO' "$OUT_FILE" 2>/dev/null || true)"
echo ""
echo "✓ 完了: ${OUT_FILE}"
echo "  データ有テーブル: ${TBL_COUNT}  / INSERT 行: ${ROW_COUNT}"
if [[ "${TBL_COUNT:-0}" -lt 90 ]]; then
  echo "  ⚠ テーブル数が想定より少ない。中核テーブル(clinics/staffs/accounts等)が含まれるか確認すること。"
fi

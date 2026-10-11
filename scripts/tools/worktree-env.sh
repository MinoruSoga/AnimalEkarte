#!/usr/bin/env bash
# worktree-env.sh — worktree 用 .env.local を生成/更新する。
#
# docker-compose.yml のホストポートと postgres_data / frontend_node_modules の
# volume 名は ${VAR:-default} 化されており、main repo は既定値を使う。
# worktree でローカルスタックを並行起動するには、このスクリプトで固有値を持つ
# .env.local を作ること（.claude/rules/worktree-runtime.md）。
#
# Usage:
#   scripts/tools/worktree-env.sh <offset> [slug]
#
#   <offset>  1-9256 の整数（14/181 は旧割当衝突で予約・拒否。先頭ゼロは10進数解釈）。
#             worktree ごとの6ポートの独立block。チケット番号を使うと一意になりやすい（EMR-300 → 300 → db 11794 等）。
#   [slug]    compose project / volume 名の接尾辞（default: wt<offset>）。
#             [a-z0-9_-] のみ使用可。
#
# 振る舞い:
#   - .env.local が無ければ .env.example から生成、あれば内容を保持したまま
#     managed block（マーカー間）だけを冪等に更新する
#   - CORS_ALLOWED_ORIGIN は FRONTEND_HOST_PORT に合わせて managed block に一本化
#     （未管理行の重複は除去。env_file の重複キー解決は実装依存のため残さない）
#   - 生成したポートが既に LISTEN 中なら警告する（非致命）
#
# Env overrides (tests / advanced):
#   WORKTREE_ENV_ROOT       対象 worktree ルート（default: script から推定）
#   WORKTREE_ENV_EXAMPLE    テンプレート（default: $ROOT/.env.example）
#   WORKTREE_ENV_FILE       出力先（default: $ROOT/.env.local）
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="${WORKTREE_ENV_ROOT:-$(cd "$SCRIPT_DIR/../.." && pwd)}"
EXAMPLE="${WORKTREE_ENV_EXAMPLE:-$ROOT/.env.example}"
ENV_FILE="${WORKTREE_ENV_FILE:-$ROOT/.env.local}"

BLOCK_BEGIN="# >>> worktree-runtime managed (scripts/tools/worktree-env.sh) >>>"
BLOCK_END="# <<< worktree-runtime managed <<<"

# main の既定ポート（最大8082）を避け、10000から6個ずつ独立blockを使う。
# offset=9256 の末尾は65535。別offset間でサービスの種類が違っても重ならない。
WORKTREE_PORT_BASE=10000
WORKTREE_PORT_COUNT=6
MAX_WORKTREE_OFFSET=9256

die() {
  echo "FAIL  $*" >&2
  exit 1
}

info() {
  echo "INFO  $*" >&2
}

usage() {
  sed -n '3,24p' "${BASH_SOURCE[0]}" >&2
}

# sed -i は BSD(macOS) と GNU で書式が違うため .bak 方式で両対応。
sed_inplace() {
  sed -i.bak "$1" "$2" && rm -f "$2.bak"
}

main() {
  local raw_offset="${1:-}"
  local offset slug="${2:-}"

  if [[ -z "$raw_offset" || "$raw_offset" == "-h" || "$raw_offset" == "--help" ]]; then
    usage
    [[ -z "$raw_offset" ]] && die "offset is required"
    exit 0
  fi
  [[ "$raw_offset" =~ ^[0-9]+$ ]] || die "offset must be an integer: $raw_offset"

  local normalized="${raw_offset#"${raw_offset%%[!0]*}"}"
  [[ -n "$normalized" ]] || die "offset out of range (1-$MAX_WORKTREE_OFFSET): $raw_offset"
  (( ${#normalized} <= ${#MAX_WORKTREE_OFFSET} )) \
    || die "offset out of range (1-$MAX_WORKTREE_OFFSET): $raw_offset"
  offset=$((10#$normalized))
  (( offset >= 1 && offset <= MAX_WORKTREE_OFFSET )) \
    || die "offset out of range (1-$MAX_WORKTREE_OFFSET): $raw_offset"
  case "$offset" in
    14|181)
      die "offset $offset is reserved: 既知の旧 worktree 割当と port block が衝突します。別の offset を選んでください"
      ;;
  esac

  if [[ -z "$slug" ]]; then
    slug="wt${raw_offset}"
  fi
  slug="$(printf '%s' "$slug" | tr '[:upper:]' '[:lower:]')"
  [[ "$slug" =~ ^[a-z0-9][a-z0-9_-]*$ ]] \
    || die "slug must match ^[a-z0-9][a-z0-9_-]*$: $slug"

  local block_start=$((WORKTREE_PORT_BASE + (offset - 1) * WORKTREE_PORT_COUNT))
  local db_port=$block_start
  local backend_port=$((block_start + 1))
  local frontend_port=$((block_start + 2))
  local storybook_port=$((block_start + 3))
  local swagger_port=$((block_start + 4))
  local redoc_port=$((block_start + 5))
  local project="animalekarte-${slug}"

  [[ -f "$EXAMPLE" ]] || die "template not found: $EXAMPLE"
  if [[ ! -f "$ENV_FILE" ]]; then
    cp "$EXAMPLE" "$ENV_FILE"
    info "created $ENV_FILE from ${EXAMPLE##*/}"
  fi

  local tmp_env
  tmp_env="$(mktemp)"
  awk '
    NR == FNR {
      if ($0 == "# >>> worktree-runtime managed (scripts/tools/worktree-env.sh) >>>") b = FNR
      if ($0 == "# <<< worktree-runtime managed <<<") e = FNR
      next
    }
    {
      if (b && FNR == b - 1 && $0 == "") next
      if (b && FNR >= b && (!e || FNR <= e)) next
      print
    }
  ' "$ENV_FILE" "$ENV_FILE" > "$tmp_env"
  cat "$tmp_env" > "$ENV_FILE"
  rm -f "$tmp_env"
  sed_inplace '/^CORS_ALLOWED_ORIGIN=/d' "$ENV_FILE"

  cat >>"$ENV_FILE" <<EOF

$BLOCK_BEGIN
COMPOSE_PROJECT_NAME=$project
DB_HOST_PORT=$db_port
BACKEND_HOST_PORT=$backend_port
FRONTEND_HOST_PORT=$frontend_port
STORYBOOK_HOST_PORT=$storybook_port
SWAGGER_HOST_PORT=$swagger_port
REDOC_HOST_PORT=$redoc_port
POSTGRES_DATA_VOLUME=ekarte-postgres-data-$slug
FRONTEND_NODE_MODULES_VOLUME=ekarte-frontend-node-modules-$slug
CORS_ALLOWED_ORIGIN=http://localhost:$frontend_port
$BLOCK_END
EOF

  info "wrote worktree block to $ENV_FILE"
  echo
  echo "  project   : $project"
  echo "  db        : localhost:$db_port     -> container 5432"
  echo "  backend   : http://localhost:$backend_port"
  echo "  frontend  : http://localhost:$frontend_port"
  echo "  storybook : http://localhost:$storybook_port   (profile docs)"
  echo "  swagger   : http://localhost:$swagger_port   (profile docs)"
  echo "  redoc     : http://localhost:$redoc_port   (profile docs)"
  echo "  volumes   : ekarte-postgres-data-$slug / ekarte-frontend-node-modules-$slug"
  echo
  echo "  next: make up   (ユーザー実行)"

  local busy=""
  local p
  for p in "$db_port" "$backend_port" "$frontend_port" "$storybook_port" "$swagger_port" "$redoc_port"; do
    if command -v lsof >/dev/null 2>&1 && lsof -nP -iTCP:"$p" -sTCP:LISTEN >/dev/null 2>&1; then
      busy="$busy $p"
    fi
  done
  if [[ -n "$busy" ]]; then
    echo "WARN  already listening on host ports:$busy" >&2
    echo "      別 project か別プロセスが使用中。offset を変えるか相手を止めてください" >&2
  fi
}

main "$@"

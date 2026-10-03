#!/usr/bin/env bash
# scripts/ci-migration-verify.sh
#
# make ci 用の migration 検証（旧 GitHub CI migration-verify job のローカル再現）。
# 使い捨て postgres:18-alpine を起動し、backend/migrations/*.sql を昇順に全適用する。
# `psql -v ON_ERROR_STOP=1 -1`（単一トランザクション・エラー即停止）は remote CI と
# 同じ契約。コンテナは EXIT trap で必ず削除する。
#
# 対応範囲の正本（docs/ops/ci-policy.md）:
#   - 本スクリプト: fresh postgres への DDL 全適用（直下 .sql は DDL 専用。
#     CSV seed bundle の COPY FROM STDIN は cmd/migrate 側の経路であり psql 適用対象外）
#   - verify_seed.py: run-local-ci.sh の host メタステップ（CSV 不変条件・DB 不要）
#   - cmd/migrate / lintscan の Go 契約: backend test(-race) フル実行が包含する
#
# Usage: bash scripts/ci-migration-verify.sh
# Env:   PG_IMAGE 上書き可（既定 postgres:18-alpine・CI service と同一）
# Exit:  0 全 migration 適用成功 / 非0 失敗（最初の失敗で停止）
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

PG_IMAGE="${PG_IMAGE:-postgres:18-alpine}"
NAME="ae-migration-verify-$$"

cleanup() {
  docker rm -f "$NAME" >/dev/null 2>&1 || true
}
trap cleanup EXIT

docker run -d --name "$NAME" \
  -e TZ=Asia/Tokyo \
  -e PGTZ=Asia/Tokyo \
  -e POSTGRES_USER=ekarte_user \
  -e POSTGRES_PASSWORD=ekarte_password \
  -e POSTGRES_DB=ekarte_db \
  -v "$ROOT/backend/migrations:/migrations:ro" \
  "$PG_IMAGE" >/dev/null

# pg_isready は「サーバ応答あり」しか確認せず db 存在を見ないため、
# entrypoint の initdb 一時サーバフェーズ（POSTGRES_DB 未作成）で PASS し得る。
# 実際に ekarte_db へ接続できるまで待つ必要がある（並列 lane 実行で露呈）。
ready=0
for _ in $(seq 1 60); do
  if docker exec "$NAME" psql -U ekarte_user -d ekarte_db -tAc 'select 1' >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 1
done
if [[ "$ready" -ne 1 ]]; then
  echo "ERROR: postgres did not become ready within 60s" >&2
  exit 1
fi

shopt -s nullglob
files=("$ROOT"/backend/migrations/*.sql)
shopt -u nullglob
if [[ "${#files[@]}" -eq 0 ]]; then
  echo "ERROR: no migration files under backend/migrations/" >&2
  exit 1
fi

for f in "${files[@]}"; do
  base="$(basename "$f")"
  echo "Applying $base ..."
  docker exec "$NAME" psql -v ON_ERROR_STOP=1 -1 -U ekarte_user -d ekarte_db -f "/migrations/$base" >/dev/null
done

echo "migration verify: ${#files[@]} migration(s) applied cleanly on fresh $PG_IMAGE"

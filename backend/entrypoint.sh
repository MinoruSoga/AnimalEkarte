#!/bin/sh
set -e

# DB migration を起動時に自動 apply (冪等)
# 失敗時は backend を起動しない (set -e により早期終了)
echo "[entrypoint] Applying database migrations..."
go run ./cmd/migrate
echo "[entrypoint] Migrations applied"

# Air でホットリロード起動
echo "[entrypoint] Starting air..."
exec air -c .air.toml

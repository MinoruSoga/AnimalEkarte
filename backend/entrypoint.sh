#!/bin/sh
set -e

# migration を先に適用してから API サーバーを起動する。
# migrate が非ゼロ終了した場合は set -e により本スクリプトも失敗し、
# air は起動しない（backend コンテナは healthy にならず起動フローがブロックされる）。
echo "[entrypoint] Running migrations..."
go run ./cmd/migrate

echo "[entrypoint] Starting air..."
exec air -c .air.toml

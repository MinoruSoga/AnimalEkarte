#!/bin/sh
set -e

# API サーバー本体のみ起動する。
# DB migration は docker compose 側の one-shot service に分離している。
echo "[entrypoint] Starting air..."
exec air -c .air.toml

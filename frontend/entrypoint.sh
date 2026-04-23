#!/bin/sh
set -e

# node_modules が存在しない or 空なら pnpm install
if [ ! -d "node_modules" ] || [ -z "$(ls -A node_modules 2>/dev/null)" ]; then
  echo "Installing dependencies..."
  pnpm install --frozen-lockfile
fi

exec "$@"

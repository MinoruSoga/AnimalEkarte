#!/bin/sh
set -e

# node_modules が存在しない / 空 / vite が欠けている場合は pnpm install
# volume が中途半端な状態でも起動できるようにする。
if [ ! -d "node_modules" ] || [ -z "$(ls -A node_modules 2>/dev/null)" ] || [ ! -x "node_modules/.bin/vite" ]; then
  echo "Installing dependencies..."
  pnpm install --frozen-lockfile
fi

exec "$@"

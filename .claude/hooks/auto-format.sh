#!/bin/bash

# 変更されたファイルを自動フォーマット
# $CLAUDE_FILE_PATH で対象ファイルを取得
#
# ⚠️ 重要: このプロジェクトは Docker 環境で動作する。
#   npm/npx コマンドはローカル実行禁止。Docker 経由で実行する。
#   TS/TSX/JS の整形は Docker 経由の ESLint に委ねる（prettier config なし）。

PROJECT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"

if [[ -n "$CLAUDE_FILE_PATH" ]]; then
    case "$CLAUDE_FILE_PATH" in
        *.go)
            # Go: Docker 経由で gofmt
            RELATIVE="${CLAUDE_FILE_PATH#$PROJECT_DIR/backend/}"
            if [[ "$RELATIVE" != "$CLAUDE_FILE_PATH" ]]; then
                docker compose -f "$PROJECT_DIR/docker-compose.yml" exec -T backend gofmt -w "/app/$RELATIVE" 2>/dev/null || true
            fi
            ;;
        # TS/TSX/JS: prettier config が存在しないためローカル npx は実行しない。
        # lint は `docker compose exec frontend npm run lint` で別途実行すること。
        *.ts|*.tsx|*.js|*.jsx|*.json|*.py)
            # no-op: Docker-first policy に従い、ローカルフォーマッタは使用しない
            ;;
    esac
fi

exit 0

#!/bin/bash

# 変更されたファイルを自動フォーマット
# $CLAUDE_FILE_PATH で対象ファイルを取得

if [[ -n "$CLAUDE_FILE_PATH" ]]; then
    case "$CLAUDE_FILE_PATH" in
        *.ts|*.tsx|*.js|*.jsx)
            npx prettier --write "$CLAUDE_FILE_PATH" 2>/dev/null
            ;;
        *.py)
            black "$CLAUDE_FILE_PATH" 2>/dev/null || true
            ;;
        *.go)
            gofmt -w "$CLAUDE_FILE_PATH" 2>/dev/null || true
            ;;
        *.json)
            npx prettier --write "$CLAUDE_FILE_PATH" 2>/dev/null || true
            ;;
    esac
fi

exit 0

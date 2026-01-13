#!/bin/bash

# コマンド実行をログに記録
LOG_FILE="$CLAUDE_PROJECT_DIR/.claude/logs/commands.log"
mkdir -p "$(dirname "$LOG_FILE")"

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Tool executed" >> "$LOG_FILE"
echo "  Working Dir: $(pwd)" >> "$LOG_FILE"
echo "---" >> "$LOG_FILE"

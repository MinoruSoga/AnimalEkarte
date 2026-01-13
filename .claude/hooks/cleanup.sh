#!/bin/bash
# SessionEnd Hook: セッション終了時のクリーンアップ

set -e

# 一時ファイルの削除
if [ -d "$CLAUDE_PROJECT_DIR/tmp" ]; then
  find "$CLAUDE_PROJECT_DIR/tmp" -type f -mtime +1 -delete 2>/dev/null || true
fi

# ログローテーション（30日以上のログを削除）
if [ -d "$CLAUDE_PROJECT_DIR/.claude/logs" ]; then
  find "$CLAUDE_PROJECT_DIR/.claude/logs" -type f -name "*.log" -mtime +30 -delete 2>/dev/null || true
fi

# セッション終了をログに記録
LOG_FILE="$CLAUDE_PROJECT_DIR/.claude/logs/sessions.log"
mkdir -p "$(dirname "$LOG_FILE")"
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Session ended" >> "$LOG_FILE"

exit 0

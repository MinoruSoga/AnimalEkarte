#!/bin/bash

echo "=== Claude Code Session Started ==="
echo "Project: $CLAUDE_PROJECT_DIR"
echo "Time: $(date)"

# 前回セッションの進捗を読み込み
PROGRESS_FILE="$CLAUDE_PROJECT_DIR/.claude/logs/session-progress.md"
if [[ -f "$PROGRESS_FILE" ]]; then
    echo ""
    echo "=== Previous Session Context ==="
    cat "$PROGRESS_FILE"
    echo "=== End Previous Context ==="
    echo ""
fi

# 進捗ファイルの存在確認（旧形式互換）
if [[ -f "$CLAUDE_PROJECT_DIR/claude-progress.txt" ]]; then
    echo "Progress file found - run /status to check progress"
fi

exit 0

#!/bin/bash

echo "=== Claude Code Session Started ==="
echo "Project: $CLAUDE_PROJECT_DIR"
echo "Time: $(date)"

# 前回セッションの進捗（Branch と最終コミットのみ。git status/diff は context 節約のため省略）
PROGRESS_FILE="$CLAUDE_PROJECT_DIR/.claude/logs/session-progress.md"
if [[ -f "$PROGRESS_FILE" ]]; then
    echo ""
    echo "=== Previous Session Context ==="
    grep -E "^\*\*(Branch|Last commit)\*\*" "$PROGRESS_FILE" || true
    echo "=== End Previous Context ==="
    echo ""
fi

exit 0

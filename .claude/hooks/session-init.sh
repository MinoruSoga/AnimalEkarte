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

# ハーネスが実行中であれば状態を表示（pure bash — python3 不要）
HARNESS_FILE="$CLAUDE_PROJECT_DIR/.claude/logs/harness-active.json"
if [[ -f "$HARNESS_FILE" ]]; then
    echo ""
    echo "=== Harness In Progress ==="
    TASK=$(grep -o '"task"[[:space:]]*:[[:space:]]*"[^"]*"' "$HARNESS_FILE" | head -1 | sed 's/.*: *"\(.*\)"/\1/' || echo "unknown")
    ITER=$(grep -o '"iteration"[[:space:]]*:[[:space:]]*[0-9]*' "$HARNESS_FILE" | head -1 | grep -o '[0-9]*$' || echo "?")
    MAX=$(grep -o '"maxIterations"[[:space:]]*:[[:space:]]*[0-9]*' "$HARNESS_FILE" | head -1 | grep -o '[0-9]*$' || echo "3")
    echo "Task: ${TASK:-unknown} | Iteration: ${ITER:-?} / ${MAX:-3}"
    echo "Resume with: /harness ${TASK:-<task>}"
    echo "=== End Harness Context ==="
    echo ""
fi

exit 0

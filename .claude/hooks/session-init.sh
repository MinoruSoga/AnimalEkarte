#!/bin/bash

echo "=== Claude Code Session Started ==="
echo "Project: $CLAUDE_PROJECT_DIR"
echo "Time: $(date)"

# プロジェクトタイプを検出
if [[ -f "$CLAUDE_PROJECT_DIR/package.json" ]]; then
    echo "Node.js project detected"
fi

if [[ -f "$CLAUDE_PROJECT_DIR/requirements.txt" ]]; then
    echo "Python project detected"
fi

if [[ -f "$CLAUDE_PROJECT_DIR/go.mod" ]]; then
    echo "Go project detected"
fi

if [[ -f "$CLAUDE_PROJECT_DIR/Cargo.toml" ]]; then
    echo "Rust project detected"
fi

# 進捗ファイルの存在確認
if [[ -f "$CLAUDE_PROJECT_DIR/claude-progress.txt" ]]; then
    echo "Progress file found - run /status to check progress"
fi

exit 0

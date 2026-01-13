#!/bin/bash

# 環境変数からツール入力を取得
TOOL_INPUT="$CLAUDE_TOOL_INPUT"

# 危険なパターンをチェック
DANGEROUS_PATTERNS=(
    "rm -rf /"
    "rm -rf ~"
    "> /dev/sda"
    "mkfs"
    "dd if=/dev/zero"
    ":(){:|:&};:"
)

for pattern in "${DANGEROUS_PATTERNS[@]}"; do
    if [[ "$TOOL_INPUT" == *"$pattern"* ]]; then
        echo "BLOCKED: Dangerous command pattern detected: $pattern"
        exit 2  # exit 2 = ブロック
    fi
done

exit 0

#!/bin/bash

# 重要な通知をSlackに送信
WEBHOOK_URL="${SLACK_WEBHOOK_URL}"
MESSAGE="$CLAUDE_NOTIFICATION"

if [[ -n "$WEBHOOK_URL" && -n "$MESSAGE" ]]; then
    curl -X POST -H 'Content-type: application/json' \
        --data "{\"text\":\"Claude Code: $MESSAGE\"}" \
        "$WEBHOOK_URL" 2>/dev/null
fi

exit 0

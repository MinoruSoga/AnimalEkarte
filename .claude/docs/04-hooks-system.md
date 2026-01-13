# Hooks（フック）システム

フックは特定のイベント発生時に自動実行されるシェルコマンドです。設定は `settings.json` の `hooks` セクションで定義します。

## 10種類のフックイベント

| フック名 | 発火タイミング | 用途例 |
|---------|---------------|--------|
| **PreToolUse** | ツール実行前 | 入力検証、権限チェック |
| **PostToolUse** | ツール実行後 | フォーマット、追加検証 |
| **PermissionRequest** | 権限ダイアログ表示時 | 自動承認/拒否 |
| **UserPromptSubmit** | ユーザープロンプト送信時 | コンテキスト追加、プロンプト検証 |
| **Notification** | 通知送信時 | カスタム通知処理 |
| **Stop** | メインエージェント停止時 | 追加作業の要求 |
| **SubagentStop** | サブエージェント停止時 | サブタスク完了処理 |
| **SessionStart** | セッション開始時 | 環境変数設定、依存関係チェック |
| **SessionEnd** | セッション終了時 | クリーンアップ処理 |
| **PreCompact** | コンテキスト圧縮前 | 圧縮前の処理 |

---

## フック設定例

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "echo 'Running command...' >> ~/.claude/audit.log"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "npm run format -- \"$CLAUDE_PROJECT_DIR\""
          }
        ]
      }
    ],
    "SessionStart": [
      {
        "matcher": "startup",
        "hooks": [
          {
            "type": "command",
            "command": "echo 'source ~/.nvm/nvm.sh && nvm use' >> \"$CLAUDE_ENV_FILE\""
          }
        ]
      }
    ]
  }
}
```

---

## フックの決定制御（JSON出力）

フックからJSON出力を返すことで、Claude Codeの動作を制御できます。

### PreToolUse での権限制御

```json
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "allow",
    "permissionDecisionReason": "Auto-approved by CI policy"
  }
}
```

### 操作のブロック

```json
{
  "decision": "block",
  "reason": "This operation is not allowed in this context"
}
```

---

## 環境変数の永続化

`SessionStart` フックで `$CLAUDE_ENV_FILE` に書き込むことで環境変数を永続化：

```bash
#!/bin/bash
if [ -n "$CLAUDE_ENV_FILE" ]; then
  echo 'export NODE_ENV=production' >> "$CLAUDE_ENV_FILE"
  echo 'source ~/.nvm/nvm.sh && nvm use 20' >> "$CLAUDE_ENV_FILE"
fi
exit 0
```

---

## フック用環境変数

| 変数 | 説明 |
|------|------|
| `$CLAUDE_PROJECT_DIR` | プロジェクトルートディレクトリ |
| `$CLAUDE_ENV_FILE` | 環境変数永続化ファイル |
| `$CLAUDE_TOOL_NAME` | 実行中のツール名 |
| `$CLAUDE_TOOL_INPUT` | ツール入力（JSON） |

---

## 実用的なフック例

### 危険なコマンドをブロック

```bash
#!/bin/bash
# .claude/hooks/block-dangerous.sh
INPUT="$CLAUDE_TOOL_INPUT"
COMMAND=$(echo "$INPUT" | jq -r '.command // empty')

if [[ "$COMMAND" =~ rm\ -rf|sudo|chmod\ 777 ]]; then
  echo '{"decision": "block", "reason": "Dangerous command blocked"}'
  exit 0
fi
exit 0
```

### ファイル編集後に自動フォーマット

```bash
#!/bin/bash
# .claude/hooks/auto-format.sh
FILE_PATH="$1"
if [[ "$FILE_PATH" =~ \.(ts|tsx|js|jsx)$ ]]; then
  npx prettier --write "$FILE_PATH" 2>/dev/null
fi
exit 0
```

### コマンドログ記録

```bash
#!/bin/bash
# .claude/hooks/log-commands.sh
echo "[$(date)] $CLAUDE_TOOL_NAME: $CLAUDE_TOOL_INPUT" >> ~/.claude/command.log
exit 0
```

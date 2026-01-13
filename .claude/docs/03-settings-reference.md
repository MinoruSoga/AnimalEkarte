# settings.json 設定リファレンス

## 設定ファイルの階層（優先度順）

1. **Enterprise managed policies** - `managed-settings.json`（最優先）
2. **Command line arguments** - CLIフラグ
3. **Local project settings** - `.claude/settings.local.json`
4. **Shared project settings** - `.claude/settings.json`
5. **User settings** - `~/.claude/settings.json`

---

## 設定ファイルの場所

| 種類 | パス |
|------|------|
| ユーザー設定 | `~/.claude/settings.json` |
| プロジェクト共有設定 | `.claude/settings.json` |
| プロジェクトローカル設定 | `.claude/settings.local.json` |
| エンタープライズポリシー（macOS） | `/Library/Application Support/ClaudeCode/managed-settings.json` |
| エンタープライズポリシー（Linux） | `/etc/claude-code/managed-settings.json` |
| エンタープライズポリシー（Windows） | `C:\Program Files\ClaudeCode\managed-settings.json` |

---

## 包括的な設定例

```json
{
  "permissions": {
    "allow": [
      "Edit",
      "Bash(git diff:*)",
      "Bash(git log:*)",
      "Bash(git branch:*)",
      "Bash(git show:*)",
      "Bash(git stash:*)",
      "Bash(npm run:*)",
      "Bash(make:*)",
      "Read(~/.zshrc)"
    ],
    "ask": [
      "Bash(git push:*)",
      "Bash(git commit:*)"
    ],
    "deny": [
      "WebFetch",
      "Bash(curl:*)",
      "Bash(wget:*)",
      "Bash(rm -rf:*)",
      "Bash(sudo:*)",
      "Read(./.env)",
      "Read(./.env.*)",
      "Read(./secrets/**)"
    ],
    "additionalDirectories": ["../docs/", "../shared/"],
    "defaultMode": "default"
  },
  "env": {
    "CLAUDE_CODE_ENABLE_TELEMETRY": "1",
    "NODE_ENV": "development"
  },
  "model": "claude-sonnet-4-5-20250929",
  "attribution": {
    "commit": "Generated with AI\n\nCo-Authored-By: Claude <noreply@anthropic.com>",
    "pr": ""
  },
  "sandbox": {
    "enabled": true,
    "autoAllowBashIfSandboxed": true,
    "excludedCommands": ["git", "docker"]
  },
  "outputStyle": "Explanatory"
}
```

---

## 権限が必要なツール / 不要なツール

### 権限が必要
- `Bash` - シェルコマンド実行
- `Edit` - ファイル編集
- `Write` - ファイル作成
- `WebFetch` - Web取得
- `WebSearch` - Web検索
- `NotebookEdit` - Jupyter編集

### 権限不要（常に許可）
- `AskUserQuestion` - ユーザーへの質問
- `Glob` - ファイルパターン検索
- `Grep` - テキスト検索
- `Read` - ファイル読み取り

---

## Permission Mode（権限モード）

| モード | 説明 | 使用場面 |
|--------|------|----------|
| `default` | 各操作で承認を求める | 通常の開発 |
| `acceptEdits` | 編集は自動承認、コマンドは確認 | 信頼できるタスク |
| `plan` | 読み取り専用、変更不可 | コードベース分析 |
| `bypassPermissions` | すべて自動承認（危険） | CI/CD等の自動化のみ |

**Shift+Tab** でインタラクティブに切り替え可能。

---

## サンドボックス設定

```json
{
  "sandbox": {
    "enabled": true,
    "autoAllowBashIfSandboxed": true,
    "excludedCommands": ["git", "docker", "make"]
  }
}
```

| 設定 | 説明 |
|------|------|
| `enabled` | サンドボックスを有効化 |
| `autoAllowBashIfSandboxed` | サンドボックス内のBashコマンドを自動承認 |
| `excludedCommands` | サンドボックスから除外するコマンド |

---

## その他の重要な設定

### モデル指定
```json
{
  "model": "claude-sonnet-4-5-20250929"
}
```

### 出力スタイル
```json
{
  "outputStyle": "Explanatory"
}
```

### 環境変数
```json
{
  "env": {
    "NODE_ENV": "development",
    "CLAUDE_CODE_ENABLE_TELEMETRY": "0"
  }
}
```

### パス設定
```json
{
  "plansPath": ".claude/plans",
  "sessionsPath": ".claude/sessions",
  "skillsPath": ".claude/skills",
  "agentsPath": ".claude/agents"
}
```

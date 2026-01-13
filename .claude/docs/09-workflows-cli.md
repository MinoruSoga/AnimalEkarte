# ワークフローとCLIリファレンス

## 主要なCLIフラグ

| フラグ | 説明 |
|--------|------|
| `--print`, `-p` | 非インタラクティブモード（終了後exit） |
| `--continue`, `-c` | 最新の会話を継続 |
| `--resume`, `-r` | セッションIDを指定して再開 |
| `--model` | モデル指定（`sonnet`, `opus`, `haiku`） |
| `--permission-mode` | 権限モード指定 |
| `--output-format` | 出力形式（`text`, `json`, `stream-json`） |
| `--max-turns` | 最大ターン数制限 |
| `--add-dir` | 追加の作業ディレクトリ |
| `--mcp-config` | MCPサーバー設定ファイル |
| `--debug` | デバッグモード |
| `--verbose` | 詳細ログ出力 |

---

## 一般的なワークフロー

### コードベース理解
```bash
> give me an overview of this codebase
> what technologies does this project use?
> trace the login process from front-end to database
```

### バグ修正
```bash
> I'm seeing an error when I run npm test
> suggest a few ways to fix the @ts-ignore in user.ts
> update user.ts to add the null check you suggested
```

### リファクタリング
```bash
> find deprecated API usage in our codebase
> refactor utils.js to use ES2024 features while maintaining the same behavior
> run tests for the refactored code
```

### テスト作成
```bash
> find functions in NotificationsService.swift that are not covered by tests
> add tests for the notification service
> add test cases for edge conditions
```

### PR作成
```bash
> summarize the changes I've made to the authentication module
> create a pr
> enhance the PR description with more context
```

---

## カスタムスラッシュコマンド

`.claude/commands/` ディレクトリにMarkdownファイルを配置：

### 例: `.claude/commands/fix-issue.md`
```markdown
GitHub Issue #$ARGUMENTS を修正します。

## 手順
1. `gh issue view $ARGUMENTS` で詳細を取得
2. 問題を分析し、関連ファイルを検索
3. 修正を実装
4. テストを実行
5. コミットしてPRを作成
```

### 使用方法
```bash
> /fix-issue 123
```

### 変数

| 変数 | 説明 |
|------|------|
| `$ARGUMENTS` | コマンドに渡された引数 |
| `$SELECTION` | IDE選択テキスト |

---

## UNIXスタイルパイプ処理

```bash
# エラーログ分析
cat build-error.txt | claude -p 'explain the root cause' > analysis.txt

# JSON出力
cat code.py | claude -p 'analyze for bugs' --output-format json

# ストリーミングJSON
cat log.txt | claude -p 'parse errors' --output-format stream-json
```

---

## Git Worktreeでの並列セッション

```bash
# 新しいworktree作成
git worktree add ../project-feature-a -b feature-a
cd ../project-feature-a
claude

# worktree管理
git worktree list
git worktree remove ../project-feature-a
```

---

## キーボードショートカット

| ショートカット | 機能 |
|---------------|------|
| `Shift+Tab` | 権限モード切り替え |
| `Tab` | Thinking（思考）モード切り替え |
| `Ctrl+K` | 検索 |
| `↑` | コマンド履歴 |
| `/` | スラッシュコマンド一覧 |
| `?` | 全ショートカット表示 |

---

## スラッシュコマンド一覧

| コマンド | 説明 |
|---------|------|
| `/help` | ヘルプ表示 |
| `/clear` | 会話履歴クリア |
| `/config` | 設定画面 |
| `/memory` | メモリ編集 |
| `/permissions` | 権限確認 |
| `/mcp` | MCP状態 |
| `/agents` | サブエージェント管理 |
| `/init` | プロジェクト初期化 |
| `/bug` | バグ報告 |
| `/tasks` | 実行中タスク一覧 |
| `/compact` | コンテキスト圧縮 |

---

## CI/CD統合

### GitHub Actions
```yaml
name: Claude Code Review
on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install Claude Code
        run: npm install -g @anthropic-ai/claude-code
      - name: Run Review
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
        run: |
          claude -p "Review the changes in this PR" \
            --permission-mode bypassPermissions \
            --output-format json > review.json
```

### GitLab CI
```yaml
claude-review:
  image: node:20
  script:
    - npm install -g @anthropic-ai/claude-code
    - claude -p "Review this merge request" --permission-mode bypassPermissions
```

---

## 出力形式

### テキスト（デフォルト）
```bash
claude -p "explain this code"
```

### JSON
```bash
claude -p "explain this code" --output-format json
```

```json
{
  "type": "result",
  "result": "..."
}
```

### ストリーミングJSON
```bash
claude -p "explain this code" --output-format stream-json
```

各行が独立したJSONオブジェクト：
```json
{"type": "text", "content": "..."}
{"type": "tool_use", "tool": "Read", "input": {...}}
{"type": "result", "result": "..."}
```

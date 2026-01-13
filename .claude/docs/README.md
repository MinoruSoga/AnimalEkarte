# Claude Code ドキュメント

> このディレクトリには、Claude Codeの包括的なガイドが含まれています。

## 目次

### 基礎

| ファイル | 内容 |
|---------|------|
| [01-getting-started.md](01-getting-started.md) | 概要、インストール、初期設定 |
| [02-claude-md-memory.md](02-claude-md-memory.md) | CLAUDE.mdメモリシステム |
| [03-settings-reference.md](03-settings-reference.md) | settings.json設定リファレンス |

### 機能

| ファイル | 内容 |
|---------|------|
| [04-hooks-system.md](04-hooks-system.md) | Hooksフックシステム |
| [05-subagents.md](05-subagents.md) | サブエージェント |
| [06-agent-skills.md](06-agent-skills.md) | Agent Skills（スキル） |
| [07-mcp-integration.md](07-mcp-integration.md) | MCP（Model Context Protocol） |

### プロンプトとワークフロー

| ファイル | 内容 |
|---------|------|
| [08-prompting-techniques.md](08-prompting-techniques.md) | 効果的なプロンプト技法 |
| [09-workflows-cli.md](09-workflows-cli.md) | ワークフローとCLIリファレンス |

### 運用

| ファイル | 内容 |
|---------|------|
| [10-security.md](10-security.md) | セキュリティ対策 |
| [11-troubleshooting.md](11-troubleshooting.md) | トラブルシューティング |

### 高度なトピック

| ファイル | 内容 |
|---------|------|
| [12-advanced-tool-use.md](12-advanced-tool-use.md) | Advanced Tool Use（APIベータ） |
| [13-agent-design-patterns.md](13-agent-design-patterns.md) | エージェント設計パターン |
| [14-context-engineering.md](14-context-engineering.md) | コンテキストエンジニアリング |
| [15-multi-agent-systems.md](15-multi-agent-systems.md) | マルチエージェントシステム |
| [16-quality-monitoring.md](16-quality-monitoring.md) | 品質監視ベストプラクティス |
| [17-long-running-agents.md](17-long-running-agents.md) | 長時間エージェント |

### その他

| ファイル | 内容 |
|---------|------|
| [ollama-guide.md](ollama-guide.md) | Ollamaローカルモデルガイド |

---

## クイックリファレンス

### 重要なファイルパス

| ファイル | 用途 |
|---------|------|
| `CLAUDE.md` | プロジェクト指示 |
| `.claude/settings.json` | Claude Code設定（共有） |
| `.claude/settings.local.json` | Claude Code設定（個人） |
| `.mcp.json` | MCPサーバー設定 |
| `.claude/commands/*.md` | カスタムスラッシュコマンド |
| `.claude/agents/*.md` | サブエージェント定義 |
| `.claude/skills/*/SKILL.md` | スキル定義 |
| `.claude/rules/*.md` | モジュール型ルール |

### よく使うコマンド

```bash
# 基本操作
claude                    # インタラクティブモード
claude -c                 # 会話継続
claude -p "query"         # ワンショット
/help                     # ヘルプ
/clear                    # 履歴クリア

# 設定
/config                   # 設定画面
/memory                   # メモリ編集
/permissions              # 権限確認
/mcp                      # MCP状態

# エージェント
/agents                   # サブエージェント管理
/init                     # プロジェクト初期化
```

---

**公式ドキュメント:** https://code.claude.com/docs/en/
**最終更新:** 2025-12-12

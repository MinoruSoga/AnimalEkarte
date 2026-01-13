# Claude Code 概要とインストール

> 公式ドキュメント: https://code.claude.com/docs/en/

## Claude Code とは

**Claude Code** は、Anthropic公式のCLIベースAIコーディングアシスタントです。ターミナルで動作し、コードベースの理解、バグ修正、機能実装、リファクタリングなどを支援します。

## 主な特徴

| 特徴 | 説明 |
|------|------|
| **自然言語からコード生成** | 説明からコードを計画・実装・検証 |
| **デバッグ支援** | エラーメッセージを解析し修正を提案 |
| **コードベース理解** | プロジェクト構造を把握し質問に回答 |
| **タスク自動化** | lint修正、マージコンフリクト解決、リリースノート作成 |
| **MCP統合** | 外部ツール（Figma、Slack、DB等）との連携 |

## 利用可能なプラットフォーム

- ターミナル（ネイティブCLI）
- Web（claude.ai/code）
- デスクトップアプリ
- Visual Studio Code
- JetBrains IDE
- GitHub Actions / GitLab CI/CD
- Slack

---

## インストール方法

### Homebrew（macOS/Linux）
```bash
brew install --cask claude-code
```

### curl（macOS/Linux/WSL）
```bash
curl -fsSL https://claude.ai/install.sh | bash
```

### Windows PowerShell
```powershell
irm https://claude.ai/install.ps1 | iex
```

### NPM（Node.js 18+）
```bash
npm install -g @anthropic-ai/claude-code
```

---

## 初回起動

```bash
cd /path/to/your/project
claude
# 初回起動時にログインを求められます
```

---

## 基本コマンド

| コマンド | 説明 |
|---------|------|
| `claude` | インタラクティブモード開始 |
| `claude "query"` | 初期プロンプト付きで開始 |
| `claude -p "query"` | ワンショットクエリ（終了後exit） |
| `claude -c` | 最新の会話を継続 |
| `claude -r` / `--resume` | 会話選択ピッカーを表示 |
| `claude update` | 最新版にアップデート |
| `/help` | ヘルプ表示 |
| `/clear` | 会話履歴クリア |

---

## 次のステップ

- [CLAUDE.mdメモリシステム](02-claude-md-memory.md)
- [settings.json設定](03-settings-reference.md)

# MCP（Model Context Protocol）

**Model Context Protocol (MCP)** は、Claude Codeを外部ツールやデータソースに接続するためのオープンスタンダードです。

## MCPでできること

- JIRAからイシューを取得し、GitHubでPRを作成
- SentryやStatisigでモニタリングデータを分析
- PostgreSQLから顧客情報を検索
- Figmaのデザインをコードに反映
- Gmail/Slackでの通知送信

---

## MCPサーバーのインストール

### リモートHTTPサーバー（推奨）
```bash
claude mcp add --transport http notion https://mcp.notion.com/mcp
```

### リモートSSEサーバー
```bash
claude mcp add --transport sse asana https://mcp.asana.com/sse
```

### ローカルstdioサーバー
```bash
claude mcp add --transport stdio airtable --env AIRTABLE_API_KEY=YOUR_KEY \
  -- npx -y airtable-mcp-server
```

---

## スコープ

| スコープ | 保存場所 | 用途 |
|---------|---------|------|
| local（デフォルト） | `~/.claude.json` | 個人用、機密情報含むサーバー |
| project | `.mcp.json` | チーム共有、バージョン管理 |
| user | `~/.claude.json` | 全プロジェクトで利用 |

---

## .mcp.json の形式

```json
{
  "mcpServers": {
    "github": {
      "type": "http",
      "url": "https://api.githubcopilot.com/mcp/",
      "headers": {
        "Authorization": "Bearer ${GITHUB_TOKEN}"
      }
    },
    "database": {
      "command": "npx",
      "args": ["-y", "@bytebase/dbhub", "--dsn", "${DATABASE_URL}"],
      "env": {
        "DATABASE_URL": "postgresql://user:pass@localhost:5432/mydb"
      }
    },
    "figma": {
      "type": "sse",
      "url": "http://localhost:3845/sse"
    }
  }
}
```

---

## MCP管理コマンド

```bash
# サーバー一覧
claude mcp list

# サーバー詳細
claude mcp get github

# サーバー削除
claude mcp remove github

# Claude Code内で状態確認
/mcp
```

---

## MCPリソースの参照

```bash
# @メンションで参照
> Can you analyze @github:issue://123 and suggest a fix?
> Compare @postgres:schema://users with @docs:file://database/user-model
```

---

## 主要なMCPサーバー

### Serena MCP（セマンティックコード解析）

```json
{
  "serena": {
    "command": "uvx",
    "args": [
      "--from", "git+https://github.com/oraios/serena",
      "serena", "start-mcp-server",
      "--context", "ide-assistant",
      "--project", "${PROJECT_ROOT}"
    ],
    "env": {
      "SERENA_MAX_MEM": "1024",
      "SERENA_MAX_WORKERS": "2"
    }
  }
}
```

### Context7 MCP（ライブラリドキュメント検索）

```json
{
  "context7-mcp": {
    "command": "npx",
    "args": ["-y", "@upstash/context7-mcp"],
    "env": {
      "NODE_OPTIONS": "--max-old-space-size=512"
    }
  }
}
```

### Playwright MCP（E2Eテスト）

```json
{
  "playwright": {
    "command": "npx",
    "args": ["-y", "@playwright/mcp@latest"]
  }
}
```

### PostgreSQL MCP

```json
{
  "postgres": {
    "command": "npx",
    "args": ["-y", "@modelcontextprotocol/server-postgres"],
    "env": {
      "DATABASE_URL": "${DATABASE_URL}"
    }
  }
}
```

---

## 設定の優先順位

1. `.mcp.local.json` (個人用設定、最優先)
2. `.mcp.json` (プロジェクト共有設定)
3. `~/.claude.json` (グローバル設定、フォールバック)

---

## トラブルシューティング

### MCPサーバーが接続できない

```bash
# デバッグモードで起動
claude --mcp-debug

# サーバー状態確認
/mcp
```

### 構文チェック

```bash
jq empty .mcp.json
```

### よくある問題

| 問題 | 解決策 |
|------|--------|
| サーバーが表示されない | JSON構文チェック、再起動 |
| 接続タイムアウト | ネットワーク確認、URL確認 |
| 認証エラー | トークン確認、環境変数確認 |
| ツールが利用できない | `--mcp-debug`でログ確認 |

---

## セキュリティベストプラクティス

### DO（推奨）
- `.mcp.local.json`に機密情報を保存
- 定期的なトークンローテーション（3ヶ月ごと）
- `.gitignore`で機密ファイルを除外

### DON'T（禁止）
- `.mcp.json`にAPIキーやトークンを含める
- Git履歴に機密情報を含める
- 平文でトークンを共有

---

## コード実行によるトークン効率化

### 核心的な問題

MCPの採用拡大で2つの非効率性が発生：

1. **ツール定義のオーバーヘッド**: 数百のツール説明をコンテキストにロード
2. **中間結果の肥大化**: ツール出力が複数回モデルを通過

### ソリューション：コードベースのツールアクセス

直接ツール呼び出しではなく、コードAPIを通じてMCPサーバーとやり取り：

**98.7%トークン削減**（150,000 → 2,000トークン）

### 主要なメリット

| メリット | 説明 |
|---------|------|
| プログレッシブツール発見 | 必要な定義のみオンデマンドでロード |
| 効率的なデータフィルタリング | 実行環境内で処理、サマリーのみ返す |
| 制御フロー | ネイティブループと条件分岐 |
| プライバシー保護 | 機密データは実行環境内に留まる |
| 状態の永続化 | 中間作業をファイルとして保存 |

---

## Desktop Extensions（.mcpb）

### 概要

Desktop Extensionsは、Claude Desktop用MCPサーバーのインストールを簡素化するバンドルパッケージ。

### 従来 vs Desktop Extensions

| 従来 | Desktop Extensions |
|------|-------------------|
| Node.js/Python必要 | 不要 |
| JSON手動編集 | 不要 |
| 依存関係管理 | 自動 |
| GitHub検索 | 不要 |

### 構築手順

```bash
# 初期化
npx @anthropic-ai/mcpb init

# パッケージ化
npx @anthropic-ai/mcpb pack

# テスト: .mcpbファイルをClaude Desktopにドラッグ
```

### Manifest構造

```json
{
  "mcpb_version": "1.0",
  "name": "my-extension",
  "version": "1.0.0",
  "server": {
    "command": "node",
    "args": ["server/index.js"]
  },
  "user_config": {
    "api_key": {
      "type": "string",
      "sensitive": true,
      "required": true
    }
  }
}
```

### 高度な機能

- **クロスプラットフォーム**: Windows、macOS、Linux対応
- **動的設定**: `${__dirname}`, `${user_config.key}`, `${HOME}` テンプレート
- **機能宣言**: tools、prompts のマニフェスト宣言

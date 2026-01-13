# Claude Code 設定ガイド

このディレクトリには、プロジェクト専用の Claude Code 設定が含まれています。

## 📁 ファイル構成

- `settings.json` - プロジェクト設定
- `README.md` - 設定ガイド（このファイル）
- `output-styles/` - カスタム出力スタイル
  - `senior-engineer.md` - シニアエンジニア向け出力スタイル
- `skills/` - カスタムSkills
  - `generating-commits/` - コミットメッセージ生成
  - `reviewing-code/` - コードレビュー支援
  - `analyzing-prisma/` - Prismaスキーマ分析
- `agents/` - サブエージェント（モデル別特化）
  - `architect.md` - 設計・セキュリティ監査 (Opus)
  - `implementer.md` - 機能実装 (Sonnet)
  - `reviewer.md` - コードレビュー (Sonnet)
  - `debugger.md` - デバッグ (Sonnet)
  - `researcher.md` - コード検索 (Haiku)
  - `formatter.md` - コミット生成・整形 (Haiku)
  - `prisma-expert.md` - DB設計 (Sonnet)
- `plans/` - 実装計画・タスク管理
- `sessions/` - 会話履歴（自動生成）
- `file-history/` - ファイル変更履歴（自動生成）
- `design_system/` - デザインシステム定義
- `docs/` - プロジェクトドキュメント

## ⚙️ settings.json の設定内容

### 基本設定

#### `outputStyle: "senior-engineer"`
カスタム出力スタイルを使用。`.claude/output-styles/senior-engineer.md` で定義されたスタイルに従い、シニアエンジニアの視点でコードを説明・レビューします。

**カスタムスタイルの特徴:**
- 日本語で明確に説明
- コードの意図を明示
- 潜在的な問題点を指摘
- 改善案を具体的に提示

**組み込みの出力スタイル:**
- `Concise` - 簡潔な出力
- `Explanatory` - 詳細な説明付き
- `Comprehensive` - 最も詳細な説明

---

### パス設定

プロジェクト固有のディレクトリをClaude Codeに認識させる設定です。

| 設定 | パス | 説明 |
|------|------|------|
| `plansPath` | `.claude/plans` | 実装計画ファイルの保存先 |
| `sessionsPath` | `.claude/sessions` | 会話セッション履歴 |
| `fileHistoryPath` | `.claude/file-history` | ファイル変更履歴 |
| `outputStylesPath` | `.claude/output-styles` | カスタム出力スタイル定義 |
| `skillsPath` | `.claude/skills` | スキル定義ファイル |
| `agentsPath` | `.claude/agents` | サブエージェント定義 |
| `designSystemPath` | `.claude/design_system` | デザインシステム定義 |
| `docsPath` | `.claude/docs` | プロジェクトドキュメント |
| `readmePath` | `.claude/README.md` | 設定ガイド（このファイル） |
| `claudePath` | `.claude` | Claude関連ファイルのルート |

**効果:**
- グローバル設定（`~/.claude/`）ではなく、プロジェクトローカルのパスを使用
- 複数プロジェクト間で設定が分離される
- チームでの共有が容易

---

### 環境変数 (env)

#### `USE_DOCKER: "true"`
このプロジェクトがDockerを使用していることを示すフラグ。

#### `DOCKER_PROJECT: "your-project-name"`
Dockerプロジェクト名。コンテナ識別に使用（プロジェクトごとに変更）。

#### `MAX_THINKING_TOKENS: "10000"`
**拡張思考モード**を有効化。Claudeが複雑な問題を解決する際に、より深く考えることができます。
- デフォルト: 無効
- 推奨値: 10000（複雑な実装や設計時に有効）

#### `BASH_DEFAULT_TIMEOUT_MS: "60000"`
Bashコマンドのタイムアウトを60秒に延長。
- デフォルト: 2000ms (2秒)
- Docker操作やビルド処理で有用

#### `CLAUDE_CODE_ENABLE_TELEMETRY: "0"`
テレメトリ（使用状況データ送信）を無効化。プライバシー保護のため。

---

### フック (hooks)

フックは特定のイベント発生時に自動実行される処理です。

#### `UserPromptSubmit` - プロンプト送信時
**タイミング:** ユーザーがメッセージを送信するたびに実行

**目的:** プロジェクトのコンテキストとコーディング原則を自動注入

**自動注入される情報:**

1. **プロジェクト情報**
   - プロジェクト名: JewelryStock（宝石店在庫管理システム）

2. **技術スタック**
   - Frontend: Next.js 14 App Router, React Server Components, TypeScript
   - Styling: Tailwind CSS, shadcn/ui
   - Backend: Next.js API Routes, Server Actions
   - ORM: Prisma
   - Database: PostgreSQL
   - Build: pnpm (Docker経由: make コマンド使用)

3. **重要ディレクトリ**
   - `src/app/` - ページ・ルート
   - `src/components/` - UIコンポーネント
   - `src/lib/` - ユーティリティ
   - `src/actions/` - Server Actions
   - `prisma/schema.prisma` - DBスキーマ

4. **コーディング原則**
   - 型安全性最優先 - anyを避け、適切な型定義
   - SOLID原則 - 単一責任、依存性逆転
   - DRY - 重複を避け、再利用可能なコード
   - エラーハンドリング徹底
   - セキュリティ優先 - XSS、SQLi、CSRF対策

5. **プロンプト最適化ガイド**
   - 「〜作って」→ 既存パターンを確認し、一貫性を保つ
   - 「〜直して」→ 根本原因を特定してから修正
   - 「〜改善して」→ 具体的な改善観点を明確化

6. **作業前チェック**
   - 関連する既存コードを確認
   - 影響範囲を把握
   - セキュリティリスクを検討

**効果:**
- プロジェクトの技術スタックを常に認識
- シニアエンジニアレベルのコーディング原則を適用
- セキュリティとパフォーマンスを常に考慮した実装
- 曖昧な指示でも適切な解釈が可能

#### `PostToolUse` - ツール実行後（Edit/Write）
**タイミング:** ファイルを編集または作成した直後

**目的:** 自動コードレビュー

**チェック項目:**
1. **型安全性** - `any` の使用、型定義の適切性
2. **エラーハンドリング** - `try-catch`、エラーメッセージの品質
3. **セキュリティ** - XSS、SQLインジェクション、CSRF対策
4. **パフォーマンス** - 不要な処理、最適化の余地
5. **テスト** - テストが必要な変更かどうか

**使用モデル:** `haiku` （高速・コスト効率的）

#### `PreToolUse` - ツール実行前（Bash）
**タイミング:** Bashコマンドを実行する直前

**目的:** 危険なコマンドの事前検出

**チェック項目:**
- `rm`, `mv`, `cp --force` など破壊的操作
- システムファイルへの変更
- 予期しない副作用

**使用モデル:** `haiku` （高速・コスト効率的）

#### `Notification` - 通知
**タイミング:** 長時間タスク完了時

**目的:** macOS通知センターでタスク完了を通知

**動作:**
- macOSの場合: 通知センターに「タスクが完了しました」を表示
- その他の場合: コンソールに「Task completed」を出力

#### `SessionStart` - セッション開始時
**タイミング:** Claude Code セッションを開始した時

**実行コマンド:** `git status --short`

**目的:**
- 変更されたファイルを把握してから作業開始
- 前回のセッションで未コミットの変更を確認
- コンテキストを素早く把握

---

### サンドボックス (sandbox)

#### `enabled: true`
サンドボックスモードを有効化。Bash実行を隔離された環境で実行し、安全性を向上。

#### `autoAllowBashIfSandboxed: true`
サンドボックス内であれば、Bashコマンドを自動的に許可。

#### `excludedCommands: ["docker", "make"]`
サンドボックスから除外するコマンド。これらは`permissions`で個別管理。

**理由:** DockerとMakefileコマンドは、プロジェクト固有の複雑な操作を行うため、permissionsで細かく制御する必要がある。

#### `network.allowLocalBinding: true`
ローカルネットワークへのバインディングを許可。

**用途:** 開発サーバー起動（Next.js dev server など）

---

### MCPサーバー

#### `enableAllProjectMcpServers: true`
プロジェクトルートの `.mcp.json` に定義された MCP サーバーを自動的に承認。

**効果:**
- MCP サーバーの手動承認が不要
- セッション開始時から即座に利用可能

**利用可能なMCPサーバー:**
- `serena` - セマンティックコード解析（TypeScriptインデックス済み）
- `context7-mcp` - React/Next.js/TypeScriptドキュメント検索
- `playwright` - E2Eテスト・ブラウザ自動化
- `chrome-devtools` - ブラウザデバッグ・スクレイピング
- `sqlite` - SQLiteデータベース操作
- `figma-mcp` - Figmaデザイン連携
- `cipher` - メモリエージェント（完全ローカル）

---

## Cipher MCP 設定

このテンプレートには Cipher MCP（メモリエージェント）の完全ローカル構成が含まれています。

### 特徴

- ✅ **OpenAI不要** - 完全ローカルLLM（Gemma 2 2B-JPN）を使用
- ✅ **プライバシー保護** - すべてのデータがローカルで処理
- ✅ **ランニングコストゼロ** - API料金不要
- ✅ **マルチエージェント最適化** - 3エージェント同時実行で約8-9GB

### セットアップ手順

#### 1. プロジェクトルートに設定ファイルをコピー

```bash
# プロジェクトルートディレクトリで実行
cp /path/to/Template/cipher.yml ./cipher.yml
cp /path/to/Template/.mcp.json ./.mcp.json
cp /path/to/Template/.env.example ./.env
```

#### 2. `.mcp.json` のパスを更新

`.mcp.json` 内の `CIPHER_AGENT_CONFIG` を実際のプロジェクトパスに変更：

```json
{
  "mcpServers": {
    "cipher": {
      "env": {
        "CIPHER_AGENT_CONFIG": "/absolute/path/to/your-project/cipher.yml"
      }
    }
  }
}
```

**例:**
```json
"CIPHER_AGENT_CONFIG": "/Users/username/projects/my-project/cipher.yml"
```

#### 3. Ollamaモデルをダウンロード

```bash
# LLMモデル（会話生成用）
ollama pull schroneko/gemma-2-2b-jpn-it

# 埋め込みモデル（セマンティック検索用）
ollama pull nomic-embed-text
```

**ダウンロード時間:** 約3-5分（合計3GB）

#### 4. Claude Codeを起動

```bash
claude code
```

### 使用方法

セットアップ完了後、Claude Code内で以下のように使用できます：

```
過去の会話で「認証機能」について話した内容を教えてください
```

Cipher MCPが完全ローカルで動作し、プライバシーを保ちながら検索結果を返します。

### 詳細ドキュメント

- **[../docs/MCP_SETUP.md](../docs/MCP_SETUP.md)** - MCP設定ガイド（全体）
- **[../docs/ollama-guide.md](../docs/ollama-guide.md)** - Ollama使い方・トラブルシューティング

### 構成オプション

| 構成 | LLM | 埋め込み | コスト | プライバシー | 推奨環境 |
|------|-----|---------|--------|------------|----------|
| **完全ローカル** 🏆 | Ollama | Ollama | 無料 | 完全ローカル | マルチエージェント |
| **ハイブリッド** | OpenAI | Ollama | 有料 | 部分クラウド | 高品質会話重視 |

このテンプレートはデフォルトで完全ローカル構成です。

### トラブルシューティング

問題が発生した場合は、`../docs/MCP_SETUP.md` の「トラブルシューティング」セクションを参照してください。

---

### Skills（スキル）

Skillsは、特定のタスクに特化した指示をClaudeに提供する機能です。

#### `generating-commits` - コミットメッセージ生成
**トリガー:** 「コミット」「git commit」「変更をコミット」

**機能:**
- ステージングされた変更を分析
- Conventional Commits形式でメッセージを生成
- feat/fix/refactor/docs/style/test/chore を自動判定

#### `reviewing-code` - コードレビュー支援
**トリガー:** 「レビュー」「PRレビュー」「コード品質」

**機能:**（読み取り専用）
- 型安全性チェック
- エラーハンドリング確認
- セキュリティ問題の検出
- パフォーマンス・保守性の評価

#### `analyzing-prisma` - Prismaスキーマ分析
**トリガー:** 「スキーマ」「マイグレーション」「DBモデル」

**機能:**（読み取り専用）
- スキーマ構造の分析
- リレーションの可視化
- マイグレーション計画の作成
- 破壊的変更の検出

**使用例:**
```
ユーザー: この変更をコミットして
→ generating-commits スキルが有効化

ユーザー: PRをレビューして
→ reviewing-code スキルが有効化

ユーザー: Prismaスキーマを確認して
→ analyzing-prisma スキルが有効化
```

---

### サブエージェント (Sub-agents)

サブエージェントは、独立したコンテキストで動作する特化型AIアシスタントです。
作業タイプに応じて最適なモデルを使用します。

#### モデル選択方針

| モデル | コスト | 速度 | 用途 |
|--------|--------|------|------|
| **Opus** | 高 | 遅 | 複雑な設計、セキュリティ監査 |
| **Sonnet** | 中 | 中 | 実装、レビュー、デバッグ |
| **Haiku** | 低 | 速 | 検索、フォーマット |

#### 利用可能なサブエージェント

| エージェント | モデル | 用途 |
|--------------|--------|------|
| `architect` | Opus | アーキテクチャ設計、重要判断、セキュリティ監査 |
| `implementer` | Sonnet | 機能実装、テスト作成、ドキュメント |
| `reviewer` | Sonnet | コードレビュー、PR確認、品質チェック |
| `debugger` | Sonnet | バグ調査、エラー分析、修正提案 |
| `researcher` | Haiku | コード検索、ファイル探索、パターン調査 |
| `formatter` | Haiku | コミットメッセージ生成、コード整形 |
| `prisma-expert` | Sonnet | スキーマ設計、マイグレーション計画 |

#### 使用方法

```bash
# エージェント一覧を確認
/agents

# 明示的に呼び出し
> Use the architect subagent to design the authentication system
> Have the reviewer subagent check my recent changes
> Ask the debugger subagent to investigate this error
```

#### Skills vs Sub-agents

| 機能 | Skills | Sub-agents |
|------|--------|------------|
| コンテキスト | メイン会話と共有 | **独立** |
| モデル選択 | 不可 | **可能** |
| 複雑なタスク | 簡単な指示向け | **複雑なワークフロー向け** |

---

### 権限管理 (permissions)

#### `deny` - 拒否リスト
**直接実行を拒否するコマンド:**
- パッケージマネージャー（npm, pnpm, yarn, npx）
- ビルドツール（node, next, tsc, eslint, prisma）

**理由:** これらは`make`コマンド経由で実行し、一貫性と安全性を確保。

**読み取り拒否:**
- 環境変数ファイル（`.env*`）
- シークレット、証明書
- ビルド成果物（`.next/`, `node_modules/`など）
- データベースファイル
- ログファイル

#### `allow` - 許可リスト
**自動承認される安全なコマンド:**
- `make` コマンド全般（開発、ビルド、テスト）
- `docker` の情報参照コマンド（`ps`, `logs`, `images`, `inspect`）
- `gh` コマンド（GitHub CLI）

#### `ask` - 確認リスト
**実行前にユーザー確認が必要な危険なコマンド:**
- ファイル削除（`rm`）
- コンテナ削除（`docker rm`, `docker-compose down`）
- Git 強制操作（`push --force`, `reset --hard`）
- データベースリセット（`make db-reset`）
- システム操作（`sudo`）

#### `defaultMode: "default"`
デフォルトの権限モード。

**利用可能な値:**
- `default` - 標準モード（推奨）
- `acceptEdits` - 編集を自動承認（開発速度優先）
- `plan` - 常にプランモード（慎重な開発）

---

### その他の設定

#### `cleanupPeriodDays: 30`
会話履歴の保持期間（日数）。30日経過後、古いセッションは自動削除されます。

#### `includeCoAuthoredBy: true`
Gitコミット時に Claude との共著情報を追加。

**コミットメッセージ例:**
```
feat: ユーザー認証機能を追加

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

---

## 🎯 この設定で実現できること

### 1. シニアエンジニアレベルのコード品質
- 型安全性、SOLID原則、DRYを意識したコード生成
- セキュリティとパフォーマンスを常に考慮
- 自己文書化された読みやすいコード

### 2. 自動コードレビュー
- 編集後に即座に品質チェック
- 潜在的な問題を早期発見
- 改善提案を自動生成

### 3. 安全な開発環境
- サンドボックスによる隔離実行
- 危険な操作は事前確認
- 機密情報へのアクセス制限

### 4. 効率的なワークフロー
- セッション開始時にGit状態を自動確認
- MCPサーバーが即座に利用可能
- 一貫したMakefileベースの操作

---

## 📚 参考リソース

- [Claude Code 公式ドキュメント](https://code.claude.com/docs/ja/)
- [Settings Schema](https://json.schemastore.org/claude-code-settings.json)
- [Hooks ガイド](https://code.claude.com/docs/ja/settings#hooks)
- [MCP サーバー](https://code.claude.com/docs/ja/mcp)

---

## 🔄 設定の更新

設定を変更する場合は、以下のコマンドで設定を再読み込みできます：

```bash
# Claude Code を再起動
claude restart

# または新しいセッションを開始
claude
```

---

**最終更新:** 2025-12-10 (MCP設定更新: playwright, sqlite, context7-mcp追加)

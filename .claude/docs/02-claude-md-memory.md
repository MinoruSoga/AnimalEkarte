# CLAUDE.md メモリシステム

Claude Codeは複数レベルのメモリを階層的に読み込みます。

## メモリの階層構造

| メモリタイプ | 場所 | 用途 | 共有範囲 |
|-------------|------|------|----------|
| **エンタープライズ** | `/Library/Application Support/ClaudeCode/CLAUDE.md`（macOS）<br>`/etc/claude-code/CLAUDE.md`（Linux）<br>`C:\Program Files\ClaudeCode\CLAUDE.md`（Windows） | 組織全体の指示 | 組織全ユーザー |
| **ユーザー** | `~/.claude/CLAUDE.md` | 個人設定（全プロジェクト共通） | 本人のみ |
| **プロジェクト** | `./CLAUDE.md` または `./.claude/CLAUDE.md` | チーム共有指示 | Gitを通じてチーム |
| **プロジェクトルール** | `./.claude/rules/*.md` | モジュール型トピック別指示 | Gitを通じてチーム |
| **プロジェクトローカル** | `./CLAUDE.local.md` | プロジェクト固有の個人設定 | 本人のみ（自動gitignore） |

---

## CLAUDE.md の記述例

```markdown
# プロジェクト名

## 概要
Next.jsベースのEコマースプラットフォーム

## 技術スタック
- Frontend: Next.js 14, TypeScript, Tailwind CSS
- Backend: Node.js, Prisma
- Database: PostgreSQL
- Testing: Jest, Playwright

## 共通コマンド
```bash
# 開発サーバー起動
npm run dev

# ビルド
npm run build

# テスト
npm test

# Lint & Format
npm run lint && npm run format
```

## コードスタイル
- TypeScript strict mode 必須
- 関数名は camelCase
- コンポーネント名は PascalCase
- 定数は UPPER_SNAKE_CASE
- インポート順: React → 外部ライブラリ → 内部モジュール → 型

## 禁止事項
- any型の使用禁止
- console.logの放置禁止
- ハードコード禁止（環境変数を使用）
- 未使用インポートの放置禁止

## ワークフロー
- コミットメッセージは Conventional Commits 形式
- mainブランチへの直接プッシュ禁止
- PRは必ずレビューを通す
```

---

## ファイルインポート機能（@構文）

CLAUDE.md内で他のファイルを参照可能：

```markdown
# プロジェクト情報
プロジェクト概要は @README.md を参照
利用可能なnpmコマンドは @package.json を確認

# 追加指示
- gitワークフロー: @docs/git-instructions.md
- 個人設定: @~/.claude/my-project-instructions.md
```

**インポート機能の特徴：**
- 相対パス・絶対パスに対応
- ホームディレクトリ（`~`）の参照が可能
- 最大5ホップまでの再帰的インポートに対応
- コード内のインポート構文は評価されない

---

## .claude/rules/ によるモジュール管理

大規模プロジェクト向けに、複数のルールファイルで指示を整理：

```
your-project/
├── .claude/
│   ├── CLAUDE.md           # メインプロジェクト指示
│   └── rules/
│       ├── code-style.md   # コード規約
│       ├── testing.md      # テスト規約
│       ├── api.md          # API開発規約
│       └── security.md     # セキュリティ要件
```

### パス特定ルール（YAML frontmatter）

```markdown
---
paths: src/api/**/*.ts
---

# API開発規約

- すべてのAPIエンドポイントに入力検証を含める
- 標準エラーレスポンス形式を使用
- OpenAPIドキュメントコメントを含める
```

### グロブパターン対応

| パターン | マッチ対象 |
|---------|---------|
| `**/*.ts` | 全ディレクトリのTypeScriptファイル |
| `src/**/*` | srcディレクトリ以下の全ファイル |
| `*.md` | プロジェクトルートのMarkdownファイル |
| `src/**/*.{ts,tsx}` | TypeScriptおよびReactファイル |

---

## メモリ追加の方法

### 方法1: # ショートカット（最速）
```
# Always use descriptive variable names
```
入力後、保存先メモリファイルの選択を促されます。

### 方法2: /memory スラッシュコマンド
セッション中に `/memory` を実行するとエディタが開きます。

### 方法3: /init でプロジェクト初期化
```
> /init
```
プロジェクト用のCLAUDE.mdを自動生成します。

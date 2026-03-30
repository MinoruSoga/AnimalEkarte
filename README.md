# Animal Ekarte (アニマル・カルテ)

動物病院向け電子カルテ管理システム。最新の React 19 と Go による高機能かつ保守性の高いアーキテクチャを採用しています。

## 🎯 プロジェクトの現状と規約 (MUST READ)

本プロジェクトは、React 19 への完全移行およびバックエンドエラー処理の標準化を完了しています。開発にあたっては、以下のドキュメントを最優先で参照してください。

- **[.claude/CLAUDE.md](.claude/CLAUDE.md)**: **【Single Source of Truth】** 開発規約・アーキテクチャ・最新ルールの集約地点。
- **[GEMINI.md](GEMINI.md)**: Gemini CLI 向けの最適化されたコンテキスト。
- **[.gemini/styleguide.md](.gemini/styleguide.md)**: 詳細なスタイルガイド・実装パターン。

---

## 🛠 技術スタック

| レイヤー | 技術 |
|---------|------|
| Frontend | React 19 / TypeScript 5.7 / Vite 6 / Tailwind CSS 4 / shadcn/ui |
| Backend | Go 1.25 / Gin / GORM / Air (Hot Reload) |
| Database | PostgreSQL 17 (Docker: postgres:17-alpine) |
| Testing | MSW (Mock Service Worker), Vitest, testify |
| Infrastructure | Docker Compose, Terraform (AWS) |

---

## 🔧 セットアップ & 運用

### 前提条件
- Docker / Docker Compose
- Make

### クイックスタート
```bash
# 1. 環境変数の準備
cp .env.example .env

# 2. コンテナ起動（初期化含む）
make up

# 3. 型定義の同期
make codegen
```

| サービス | URL | 備考 |
|---------|-----|-----|
| Frontend | http://localhost:3000 | |
| Backend API | http://localhost:8080/api/v1 | |
| PostgreSQL | localhost:5434 | |

---

## 開発コマンド (Makefile)

### 基本操作
- `make up` / `make down`: コンテナの起動・停止
- `make build`: コンテナのビルドと起動
- `make logs`: 全ログの表示
- `make db`: DB接続 (psql)
- `make reset`: DBの完全初期化（データの全削除を伴う）

### 品質管理 & 生成
- `make lint`: Go リンター実行
- `make lint-front`: **(New)** フロントエンド リンター実行
- `make test`: Go テスト実行
- `make test-front`: **(New)** フロントエンド テスト (Vitest) 実行
- `make codegen`: Goモデルから TypeScript型を自動生成

> **⚠️ 注意: コマンド実行ルール**
> すべてのコマンドは Docker 経由で実行してください。ローカル環境への npm/go インストールは不要です。

---

## 🏗 アーキテクチャの特徴

### Frontend (React 19 Idiomatic)
- **Action Pattern**: `useActionState` と `SubmitButton` による宣言的フォーム管理。
- **Feature Indexing**: 各機能の `index.ts` (Public API) を通じた厳格なカプセル化。
- **Dependency Inversion**: `app/pages/` で各機能を合成し、機能間直接インポートを排除。
- **Design Tokens**: `design-tokens.ts` による Notion 風テーマの完全制御。

### Backend (Clean Architecture)
- **Unified Error Handling**: `apperrors` パッケージによる統一的なエラーラッピング。
- **Single Source of Truth**: Go のモデル定義からフロントエンドの型を自動生成。

---

## 📚 関連ドキュメント

| カテゴリ | ドキュメント |
|:---|:---|
| **設計** | [architecture.md](docs/architecture.md), [ERD.md](docs/ERD.md), [AUTH.md](docs/AUTH.md) |
| **仕様** | [screens/](docs/screens/) (画面定義), [SPECIFICATION.md](docs/SPECIFICATION.md) |
| **履歴** | [archive/](docs/archive/) (刷新前の設計図、過去のタスク履歴) |

---

## ライセンス

Private / Proprietary

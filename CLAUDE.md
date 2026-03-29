# Animal Ekarte - 動物病院向け電子カルテシステム

> **⚠️ Note for Gemini CLI**: 
> Gemini CLI を使用する場合は、プロジェクトルートの **`GEMINI.md`** および **`.gemini/styleguide.md`** を最優先の真実（Source of Truth）として参照してください。

## ⚠️ Claude Code への指示 (最重要)
**このプロジェクトの詳細な開発ルール・コマンド・ベストプラクティスは `.claude/CLAUDE.md` に集約されています。**
タスクを開始する前に、必ず `.claude/CLAUDE.md` を読み込み、プロジェクトのコンテキストとルールを理解してください。

---

## 🏗 全体アーキテクチャ概要

本システムは、React(TypeScript)のフロントエンドと、Go(Gin)のバックエンドAPI、そしてPostgreSQLデータベースで構成され、すべてDocker Compose上で動作します。

### インフラストラクチャ (Infra)
- **環境**: Docker Compose による完全コンテナ化
- **データベース**: PostgreSQL 18
- **制約事項**: ホストOS（ローカル）での直接のビルド・実行は禁止されています。すべてコンテナ経由（`docker compose exec` や `make`）で行う必要があります（詳細は `.claude/CLAUDE.md` 参照）。

### フロントエンド (Frontend)
- **技術スタック**: React 19, TypeScript 5.7, Vite 6, Tailwind CSS 4, shadcn/ui
- **アーキテクチャ**: `bulletproof-react` をベースにした機能ベース（Feature-based）アーキテクチャ
- **ディレクトリ構造**:
  ```text
  frontend/src/
  ├── app/           # ルーター、プロバイダ、機能間をまたぐ合成ページ (cross-feature)
  ├── features/      # 機能別モジュール (api, components, hooks, routes)
  ├── components/    # 共有UI (shadcn/ui, layouts, errors)
  ├── hooks/         # グローバルhooks
  ├── types/         # 共有型定義（バックエンドから自動生成された models.ts を含む）
  └── lib/           # ライブラリ設定
  ```

### バックエンド (Backend)
- **技術スタック**: Go 1.25, Gin (Webフレームワーク), GORM (ORM)
- **アーキテクチャ**: クリーンアーキテクチャの思想を取り入れた軽量レイヤード設計
- **ディレクトリ構造**:
  ```text
  backend/internal/
  ├── handler/       # HTTPルーティング、リクエスト/レスポンスのバインド
  ├── service/       # ビジネスロジック、ドメイン操作
  ├── repository/    # データアクセス (GORM)
  ├── model/         # ドメインモデル (DBスキーマ対応)
  └── errors/        # センチネルエラー定義
  ```

---

## 📚 ドキュメント構成

| ドキュメント | 説明 |
|-------------|------|
| **[.claude/CLAUDE.md](.claude/CLAUDE.md)** | **【必読】** AI向けの統合開発ルール・コマンド・運用姿勢 |
| [frontend/CODING_RULES.md](frontend/CODING_RULES.md) | フロントエンドの詳細なルールと実装パターン（React 19） |
| [backend/CLAUDE.md](backend/CLAUDE.md) | バックエンドの詳細な実装パターン |
| [docs/CLAUDE_COMMANDS.md](docs/CLAUDE_COMMANDS.md) | Claude Code カスタムコマンド・マニュアル |
| [docs/ERD.md](docs/ERD.md) | データベース設計（ER図） |
| [docs/architecture.md](docs/architecture.md) | アーキテクチャの詳細ドキュメント |

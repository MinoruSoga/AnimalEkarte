# Animal Ekarte - 動物病院向け電子カルテシステム

> **⚠️ Source of Truth (真実の源泉)**: 
> このプロジェクトの開発規約・アーキテクチャ・ベストプラクティスは、
> **`.claude/CLAUDE.md`** に集約されています。

## ⚠️ 開発者への指示 (最重要)
タスクを開始する前に、必ず **`.claude/CLAUDE.md`** を読み込み、最新のプロジェクトコンテキスト（React 19 Action パターン、Feature Indexing 等）を理解してください。

---

## 🏗 全体アーキテクチャ概要
本プロジェクトは、フロントエンドに React 19、バックエンドに Go/Gin を採用した動物病院向けシステムです。backend package は Go/Gin公式ガイドを基準に凝集性・利用者・依存方向で設計し、特定の layer pattern を公式要件として固定しません。

詳細は以下のドキュメントを参照してください：

| ドキュメント | 内容 |
|-------------|------|
| [.claude/CLAUDE.md](.claude/CLAUDE.md) | **【最優先】** 開発ルール・コマンド・最新技術規約 |
| [.claude/rules/go-gin-backend-guidelines.md](.claude/rules/go-gin-backend-guidelines.md) | Go/Gin backend の公式一次資料ベースライン |
| [docs/product-philosophy.md](docs/product-philosophy.md) | **【機能開発前 必読】** 業務効率ソフトウェアとしての意思決定原則（5 ステップ） |
| [docs/architecture/erd.md](docs/architecture/erd.md) | データベース設計（ER図） |
| [docs/architecture/overview.md](docs/architecture/overview.md) | アーキテクチャの詳細ドキュメント |

# Animal Ekarte - 動物病院向け電子カルテシステム

> **⚠️ Source of Truth (真実の源泉)**: 
> このプロジェクトの開発規約・アーキテクチャ・ベストプラクティスは、
> **`.claude/CLAUDE.md`** に集約されています（SSOT）。
>
> AI エージェント運用は **Antigravity CLI (`agy`)** を標準とします（Gemini CLI は 2026-06-18 に提供停止となるため移行済み）。手順は `docs/ANTIGRAVITY_CLI.md` を参照してください。

## ⚠️ 開発者への指示 (最重要)
タスクを開始する前に、必ず **`.claude/CLAUDE.md`** を読み込み、最新のプロジェクトコンテキスト（React 19 パターン、Feature Indexing 等）を理解してください。

---

## 🏗 全体アーキテクチャ概要
本プロジェクトは、フロントエンドに React 19、バックエンドに Go (Clean Architecture) を採用した、スケーラブルな動物病院向けシステムです。

詳細は以下のドキュメントを参照してください：

| ドキュメント | 内容 |
|-------------|------|
| [.claude/CLAUDE.md](.claude/CLAUDE.md) | **【最優先 / SSOT】** 開発ルール・コマンド・最新技術規約 |
| [docs/ANTIGRAVITY_CLI.md](docs/ANTIGRAVITY_CLI.md) | **【現在の標準】** Antigravity CLI (`agy`) 運用メモ（2026-06-18 に提供停止となる Gemini CLI からの移行手順） |
| [GEMINI.md](GEMINI.md) | Gemini CLI 向け最適化コンテキスト（移行後もルール参照・互換用として保持） |
| [.gemini/styleguide.md](.gemini/styleguide.md) | スタイルガイド・実装パターン詳細 |
| [docs/ERD.md](docs/ERD.md) | データベース設計（ER図） |
| [docs/architecture.md](docs/architecture.md) | アーキテクチャの詳細ドキュメント |

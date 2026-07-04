# Animal Ekarte - 動物病院向け電子カルテシステム

> **⚠️ Source of Truth (真実の源泉)**: 
> このプロジェクトの開発規約・アーキテクチャ・ベストプラクティスは、
> **`.claude/CLAUDE.md`** に集約されています。
>
> Gemini CLI を使用する場合、プロジェクトルートの **`GEMINI.md`** も併せて参照してください（`.claude/CLAUDE.md` と同期済みです）。

## ⚠️ 開発者への指示 (最重要)
タスクを開始する前に、必ず **`.claude/CLAUDE.md`** を読み込み、最新のプロジェクトコンテキスト（React 19 Action パターン、Feature Indexing 等）を理解してください。

---

## 🏗 全体アーキテクチャ概要
本プロジェクトは、フロントエンドに React 19、バックエンドに Go (Clean Architecture) を採用した、スケーラブルな動物病院向けシステムです。

詳細は以下のドキュメントを参照してください：

| ドキュメント | 内容 |
|-------------|------|
| [.claude/CLAUDE.md](.claude/CLAUDE.md) | **【最優先】** 開発ルール・コマンド・最新技術規約 |
| [docs/PRODUCT_PHILOSOPHY.md](docs/PRODUCT_PHILOSOPHY.md) | **【機能開発前 必読】** 業務効率ソフトウェアとしての意思決定原則（5 ステップ） |
| [GEMINI.md](GEMINI.md) | Gemini CLI 向け最適化コンテキスト |
| [.gemini/styleguide.md](.gemini/styleguide.md) | スタイルガイド・実装パターン詳細 |
| [docs/ERD.md](docs/ERD.md) | データベース設計（ER図） |
| [docs/architecture.md](docs/architecture.md) | アーキテクチャの詳細ドキュメント |

## モデルルーティング

デフォルト：Claude Sonnet 5
コーディング、ツール利用、リファクタリング、
日常的な作業にはSonnet 5を使用する。
Opus 4.8へ切り替えるのは、次の場合のみ：
- Sonnet 5が同じタスクに2回失敗した場合
- より深い推論が必要な場合
  - 複雑なシステム設計
  - 微妙な正しさを証明する必要があるタスク
まずSonnet 5から始める。
最初からOpusを使うのではなく、結果や証拠に基づいて切り替える。
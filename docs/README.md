# Animal Ekarte プロジェクト・ドキュメント (Documentation Index)

> **目的**: docs/ 配下の公式ドキュメントへの入口（カテゴリ索引）を提供する。
> **読者**: 全開発者・AI エージェント。
> **タイミング**: 関連ドキュメントを探す時。

開発規約・アーキテクチャ・開発手順の統合ルール（SSOT）は **[.claude/CLAUDE.md](../.claude/CLAUDE.md)** を最優先で参照すること。

## 構成（2層索引）

ファイル単位の説明は各フォルダの README.md が正本。本書はカテゴリの入口のみを持つ（二重管理をしない）。

```
docs/
├── product-philosophy.md   … 意思決定原則（何を作るか/作らないか。全文書の上位）
├── architecture/           … 説明系: システムがどう作られているか
├── spec/                   … 仕様系: システムが何をするか（業務・画面・UI）
├── ops/                    … 運用系: デプロイ・CI・テスト・インフラ
├── delivery/               … 納品系: クライアント納品物（Go-live・操作マニュアル）
└── work/                   … 進行中の作業台帳・採択済み決裁・任意ブラウザ結果
```

| カテゴリ | 索引 | 概要 |
|:---|:---|:---|
| **意思決定原則** | [product-philosophy.md](product-philosophy.md) | 5 ステップの意思決定原則。新機能・仕様変更の着手前に必読 |
| **説明系** | [architecture/README.md](architecture/README.md) | レイヤード構造・ERD・RBAC/マルチテナント・データフロー・削除設計・ADR |
| **仕様系** | [spec/README.md](spec/README.md) | 機能要件・全画面仕様・会計/顧客分析/予約フロー・デザイン規約・LINE 連携 |
| **運用系** | [ops/README.md](ops/README.md) | デプロイ・ランブック・テスト・CI/カバレッジポリシー・インフラ構成 |
| **納品系** | [delivery/README.md](delivery/README.md) | 納品パッケージ・Go-live 手順・現場向け操作マニュアル |
| **作業台帳** | [work/README.md](work/README.md) | 進行中の補助メモ・採択済み決裁（正本は root `STATUS.md` / `PO-todo.md`） |
> **フォルダ規律**: docs/ 直下に新カテゴリを追加する場合は本表と CI ゲート（`scripts/check-docs-symbol-drift.sh` の TOP_ALLOWLIST）を同コミットで更新すること。allowlist 外のフォルダは CI が拒否する。

## 横断事項

- **API contract**: 正本は [`backend/docs/api.yaml`](../backend/docs/api.yaml)（Swagger UI 表示は `docker compose -f docker-compose.swagger.yml up`）。
- **docs ドリフトゲート**: `scripts/check-docs-symbol-drift.sh`（CI の docs-symbol-drift ジョブ）が、spec/screens/ 系ドキュメントの言及シンボル実在と宣言数値（テーブル数・リソース数等）の実装一致を機械検査する。
- **タスク台帳（正本）**: リポジトリ直下 [`STATUS.md`](../STATUS.md)（§1 残作業 · §2 Issue · §3 BUG）。USER 実行リストは [`PO-todo.md`](../PO-todo.md)。旧 `todo.md` / `bug.md` / `3-session-agent.html` は STATUS へのスタブのみ。
- **作業補助**: [`work/README.md`](work/README.md)（採択済み Fable 方針のみ。レポート置き場ではない）。
- 旧 `reports/` とブラウザ結果専用 md は削除済み — 必要なら git 履歴。

---

**最新更新**: 2026-08-07 | **ステータス**: work 台帳を `docs/work/` に集約。実装 SoT は root `STATUS.md`

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
└── work/                   … 補助メモ・採択済み決裁・今期外索引（実行 SoT ではない）
```

| カテゴリ | 索引 | 概要 |
|:---|:---|:---|
| **意思決定原則** | [product-philosophy.md](product-philosophy.md) | 5 ステップの意思決定原則。新機能・仕様変更の着手前に必読 |
| **説明系** | [architecture/README.md](architecture/README.md) | domain/capability-first modular monolith・package 境界・ERD・RBAC/マルチテナント・データフロー・削除設計・ADR |
| **仕様系** | [spec/README.md](spec/README.md) | 機能要件・全画面仕様・会計/顧客分析/予約フロー・デザイン規約・LINE 連携 |
| **運用系** | [ops/README.md](ops/README.md) | デプロイ・ランブック・テスト・CI/カバレッジポリシー・インフラ構成 |
| **納品系** | [delivery/README.md](delivery/README.md) | 納品パッケージ・Go-live 手順・現場向け操作マニュアル |
| **作業台帳** | [work/README.md](work/README.md) | 補助メモ・採択済み決裁・今期外索引（実行 SoT は Linear。root 台帳の例外は下記参照） |
> **フォルダ規律**: docs/ 直下に新カテゴリを追加する場合は本表とローカル必須ゲート（`scripts/check-docs-symbol-drift.sh` の TOP_DIR_ALLOWLIST）を同コミットで更新すること。allowlist 外のフォルダは `make ci` が拒否する。

## 横断事項

- **API contract**: 正本は [`backend/docs/api.yaml`](../backend/docs/api.yaml)（Swagger UI 表示は `make docs-ui`）。
- **docs ドリフトゲート**: `scripts/check-docs-symbol-drift.sh`（GitHub CI ではなく `make ci` のローカル必須ゲート。分担は [ops/ci-policy.md](ops/ci-policy.md)）が、spec/screens/ 系ドキュメントの言及シンボル実在と宣言数値（テーブル数・リソース数等）の実装一致を機械検査する。
- **タスク台帳**: 実行 SoT は **Linear**（hub BRT-4）。`todo-po.md` は入口ポインタ。[`todo.md`](../todo.md) は STG 実データ運用テストの例外台帳を含み、[`bug.md`](../bug.md) は UAT 製品 FAIL 台帳を含む。競合時の状態の正本は Linear。root 例外台帳は同一変更で同期し、終了条件と削除予定を台帳内に明記する。時点レポートは CorpVault `evidence/2026-08-20-*` と git 履歴。`reports/` は gitignore（新規 UAT をコミットしない）。
- **作業補助**: [`work/README.md`](work/README.md)（採択済み方針の短いポインタ。レポート置き場ではない）。

---

**最新更新**: 2026-08-31 | **ステータス**: 実行 SoT は Linear（BRT-4）。docs/ は 5 カテゴリ体制で、ファイル単位の説明は各フォルダ README が正本

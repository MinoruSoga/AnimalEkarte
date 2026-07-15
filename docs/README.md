# Animal Ekarte プロジェクト・ドキュメント (Documentation Index)

> **目的**: docs/ 配下の全公式ドキュメントへの索引を提供する。
> **読者**: 全開発者。
> **タイミング**: 関連ドキュメントを探す時。

本ディレクトリには、Animal Ekarte の設計、仕様、および運用に関する全ての公式ドキュメントが集約されています。

---

## ⚠️ 開発ガイドライン (Single Source of Truth)

プロジェクトの最新規約、アーキテクチャ、および開発手順については、以下のファイルを**必ず**最優先で参照してください。

- **[.claude/CLAUDE.md](../.claude/CLAUDE.md)**: 全エージェントおよび開発者向けの統合ルール（SSOT）。
- **[GEMINI.md](../GEMINI.md)**: Gemini CLI 向けの要点ポインタ（詳細規約は `.claude/CLAUDE.md` が正本）。
- **[AI_DEVELOPMENT_WORKFLOW.md](AI_DEVELOPMENT_WORKFLOW.md)**: 仕様駆動・AI エージェント協働開発の標準手順。
- **[CODEGRAPH_PROMPTS.md](CODEGRAPH_PROMPTS.md)**: CodeGraph MCP 調査プロンプト集。

---

## 🏗 システム設計と基盤

| カテゴリ | ドキュメント | 概要 |
|:---|:---|:---|
| **設計思想** | [PRODUCT_PHILOSOPHY.md](PRODUCT_PHILOSOPHY.md) | 業務効率ソフトウェアとしての意思決定原則（5 ステッププロセス）。 |
| **技術設計** | [architecture.md](architecture.md) | 軽量レイヤードアーキテクチャの定義。 |
| **データ** | [ERD.md](ERD.md) | データベース設計（全 **108 テーブル**・リレーション）。 |
| **セキュリティ**| [AUTH.md](AUTH.md) | RBAC 権限モデル（全 **34 リソース**）、マルチテナント隔離。 |
| **UI/UX** | [DESIGN_SYSTEM.md](DESIGN_SYSTEM.md) | Notion ライクなデザイン規約とデザイントークン。 |
| **API** | [openapi.yaml](openapi.yaml) | Swagger UI 表示用。contract 正本は [`backend/docs/api.yaml`](../backend/docs/api.yaml)。 |
| **データフロー** | [data-flow.md](data-flow.md) | リクエストの追跡（Request ID）と非同期同期の仕組み。 |
| **削除設計** | [DELETE_SOFT_DELETE_PATTERNS.md](DELETE_SOFT_DELETE_PATTERNS.md) | Hard Delete と Soft Delete の使い分け、FK 制約との関係。 |
| **意思決定記録** | [adr/](adr/) | アーキテクチャ判断の経緯（ADR）。マルチテナント隔離・支払方法・健診系統など4件。 |

---

## 📱 業務・画面仕様

- **[SPECIFICATION.md](SPECIFICATION.md)**: システム全体の機能要件と主要ビジネスフロー。
- **[screens/README.md](screens/README.md)**: **【全 37 画面インデックス】** 各機能の詳細操作ガイド。
- **[CASH_REGISTER_SPEC.md](CASH_REGISTER_SPEC.md)**: レジ締め・日次/月次売上集計の業務仕様。
- **[CUSTOMER_AGGREGATION_SPEC.md](CUSTOMER_AGGREGATION_SPEC.md)**: 累計売上・来院頻度に基づく顧客分析ダッシュボード。
- **[reservation-to-record-flow.md](reservation-to-record-flow.md)**: 予約からカルテ作成までの統合フロー詳細。

---

## 💬 LINE / Lステップ連携 (CRM)

- **[line/lstep-integration.md](line/lstep-integration.md)**: **Lステップ戦略書**。CPM 判定、全 15 種の配信トリガー。
- **[line/setup.md](line/setup.md)**: LINE Developers Console および管理画面での初期セットアップ。
- **[LINE_LSTEP_COST_ANALYSIS.md](LINE_LSTEP_COST_ANALYSIS.md)**: 外部配信コストと ROI の経済性分析。
- **[line/reservation-spec.md](line/reservation-spec.md)**: 飼い主向け LINE 予約システムの機能と計算エンジン。

---

## 📦 運用とテスト

- **[infra/deploy/README.md](infra/deploy/README.md)**: ステージング環境の運用・デプロイガイド（Cloudflare 正系統 / AWS ECS はロールバック専用）。
- **[testing/SECTION_14_MANUAL_TEST_GUIDE.md](testing/SECTION_14_MANUAL_TEST_GUIDE.md)**: ブラウザによる詳細な手動検証シナリオ。
- **[FUNCTIONAL_TEST_REPORT.md](FUNCTIONAL_TEST_REPORT.md)**: **【全機能検証記録】** 2,000 項目以上の詳細チェックリスト。
- **[DEPLOYMENT_CHECKLIST.md](DEPLOYMENT_CHECKLIST.md)**: 本番リリース前の統合チェックリスト。
- **[ci-policy.md](ci-policy.md)**: CI ワークフローの決定事項記録（Actions バージョン統一方針など）。
- **[coverage-policy.md](coverage-policy.md)**: テストカバレッジ ratchet 方式の運用ポリシー。
- **docs ドリフトゲート**: `scripts/check-docs-symbol-drift.sh`（CI の docs-symbol-drift ジョブ）が、screens/ 系ドキュメントの言及シンボル実在と宣言数値（テーブル数・リソース数等）の実装一致を機械検査する。

---

**最新更新**: 2026-07-10 | **ステータス**: All Sync with Implementation (108 Tables / 34 Resources)

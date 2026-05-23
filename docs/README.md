# Animal Ekarte プロジェクト・ドキュメント (Documentation Index)

本ディレクトリには、Animal Ekarte の設計、仕様、および運用に関する全ての公式ドキュメントが集約されています。

---

## ⚠️ 開発ガイドライン (Single Source of Truth)

プロジェクトの最新規約、アーキテクチャ、および開発手順については、以下のファイルを**必ず**最優先で参照してください。

- **[.claude/CLAUDE.md](../.claude/CLAUDE.md)**: 全エージェントおよび開発者向けの統合ルール。
- **[GEMINI.md](../GEMINI.md)**: Gemini CLI 向けの最適化されたコンテキスト。
- **[AI_DEVELOPMENT_WORKFLOW.md](AI_DEVELOPMENT_WORKFLOW.md)**: 仕様駆動・AI エージェント協働開発の標準手順。

---

## 🏗 システム設計と基盤

| カテゴリ | ドキュメント | 概要 |
|:---|:---|:---|
| **技術設計** | [architecture.md](architecture.md) | 軽量レイヤードアーキテクチャの定義。 |
| **データ** | [ERD.md](ERD.md) | データベース設計（全 **95 テーブル**・リレーション）。 |
| **セキュリティ**| [AUTH.md](AUTH.md) | RBAC 権限モデル（全 **31 リソース**）、マルチテナント隔離。 |
| **UI/UX** | [DESIGN_SYSTEM.md](DESIGN_SYSTEM.md) | Notion ライクなデザイン規約とデザイントークン. |
| **API** | [API_SPEC.md](API_SPEC.md) | バックエンド Go API の詳細リファレンス (v2.3)。 |
| **データフロー** | [data-flow.md](data-flow.md) | リクエストの追跡（Request ID）と非同期同期の仕組み。 |

---

## 📱 業務・画面仕様

- **[SPECIFICATION.md](SPECIFICATION.md)**: システム全体の機能要件と主要ビジネスフロー。
- **[screens/README.md](screens/README.md)**: **【全 38 画面インデックス】** 各機能の詳細操作ガイド。
- **[CASH_REGISTER_SPEC.md](CASH_REGISTER_SPEC.md)**: レジ締め・日次/月次売上集計の業務仕様。
- **[CUSTOMER_AGGREGATION_SPEC.md](CUSTOMER_AGGREGATION_SPEC.md)**: 累計売上・来院頻度に基づく顧客分析ダッシュボード。

---

## 💬 LINE / Lステップ連携 (CRM)

- **[line/lstep-integration.md](line/lstep-integration.md)**: **Lステップ戦略書**。CPM 判定、全 15 種の配信トリガー。
- **[line/setup.md](line/setup.md)**: LINE Developers Console および管理画面での初期セットアップ。
- **[LINE_LSTEP_COST_ANALYSIS.md](LINE_LSTEP_COST_ANALYSIS.md)**: 外部配信コストと ROI の経済性分析。
- **[line/reservation-spec.md](line/reservation-spec.md)**: 飼い主向け LINE 予約システムの機能と計算エンジン。

---

## 📦 運用とテスト

- **[infra/deploy/README.md](infra/deploy/README.md)**: AWS ステージング環境の運用・デプロイガイド。
- **[testing/SECTION_14_MANUAL_TEST_GUIDE.md](testing/SECTION_14_MANUAL_TEST_GUIDE.md)**: ブラウザによる詳細な手動検証シナリオ。
- **[FUNCTIONAL_TEST_REPORT.md](FUNCTIONAL_TEST_REPORT.md)**: **【全機能検証記録】** 2,000 項目以上の詳細チェックリスト。
- **[DEPLOYMENT_CHECKLIST.md](DEPLOYMENT_CHECKLIST.md)**: 本番リリース前の統合チェックリスト。

---

**最新更新**: 2026-05-21 | **ステータス**: All Sync with Implementation (95 Tables / 31 Resources)

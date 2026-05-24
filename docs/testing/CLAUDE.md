# テスト・品質保証ディレクトリ (Testing & Quality Assurance)

> **目的**: システムの信頼性、パフォーマンス、およびコード品質を担保するための戦略と記録の管理。

---

## 📂 ディレクトリ構成

- **`INTEGRATION_TEST_PLAN.md`**: テストの三層構造（Unit/Integration/E2E）および負荷試験の戦略。
- **`E2E_TESTING_GUIDE.md`**: Playwright を使用した、実際の業務フローに基づく自動検証の手順。
- **`SECTION_14_MANUAL_TEST_GUIDE.md`**: ブラウザを使用した詳細な手動検証のチェックリスト。
- **`PERFORMANCE_PROFILING.md`**: レスポンスタイム目標と、ボトルネック特定のプロファイリング手法。
- **`HANDLER_TEST_DOCUMENTATION_STATUS.md`**: 全 31 バックエンドハンドラーのテストカバレッジ状況。

---

## ✅ 品質基準の原則

1.  **自動化の優先**: 回帰バグを即座に検知するため、臨床フローの 90% 以上を E2E テストでカバーする。
2.  **実データに近い検証**: マイグレーションシード（`004_seed_staging.sql`）を使用し、大規模データ下でのパフォーマンスを定常的に確認する。
3.  **証跡の公開**: 全ドメインの検証結果は **[`docs/FUNCTIONAL_TEST_REPORT.md`](../FUNCTIONAL_TEST_REPORT.md)** にてオープンに管理する。

---

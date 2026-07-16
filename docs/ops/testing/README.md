# testing/ — テスト・品質保証

> **目的**: テスト戦略・検証手順・品質記録のドキュメント索引を提供する。
> **読者**: 全開発者・QA・AI エージェント。
> **タイミング**: テスト実施・テスト戦略変更・品質検証の前。

編集時のルール・品質基準の原則は [CLAUDE.md](CLAUDE.md) を参照。

## 索引

| ドキュメント | 内容 | いつ読むか |
|:---|:---|:---|
| [INTEGRATION_TEST_PLAN.md](INTEGRATION_TEST_PLAN.md) | テストの三層構造（Unit/Integration/E2E）と負荷試験の戦略 | テスト方針の確認・変更時 |
| [E2E_TESTING_GUIDE.md](E2E_TESTING_GUIDE.md) | Playwright による業務フロー自動検証の実行・追加手順 | E2E テストの実行・追加時 |
| [SECTION_14_MANUAL_TEST_GUIDE.md](SECTION_14_MANUAL_TEST_GUIDE.md) | ブラウザによる詳細な手動検証シナリオ（browser-test スキルが使用） | 手動検証・ブラウザ QA 時 |
| [PERFORMANCE_PROFILING.md](PERFORMANCE_PROFILING.md) | レスポンスタイム目標とプロファイリング手法（pprof / Lighthouse） | 性能調査・ボトルネック特定時 |
| [HANDLER_TEST_DOCUMENTATION_STATUS.md](HANDLER_TEST_DOCUMENTATION_STATUS.md) | バックエンドハンドラーのテストカバレッジ状況 | ハンドラーテスト整備の計画時 |

## AI エージェント向け注記

- カバレッジ基準の正本は [../coverage-policy.md](../coverage-policy.md)（ratchet 方式）。
- 検証結果はテスト実施時のレポートとして記録する（旧 FUNCTIONAL_TEST_REPORT.md は凍結スナップショットだったため削除済み — git 履歴参照）。

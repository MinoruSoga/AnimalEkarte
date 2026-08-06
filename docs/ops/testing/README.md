# testing/ — テスト・品質保証

> **目的**: テスト戦略・検証手順・品質記録のドキュメント索引を提供する。
> **読者**: 全開発者・QA・AI エージェント。
> **タイミング**: テスト実施・テスト戦略変更・品質検証の前。

編集時のルール・品質基準の原則は [CLAUDE.md](CLAUDE.md) を参照。

## 索引

| ドキュメント | 内容 | いつ読むか |
|:---|:---|:---|
| [scenarios/](scenarios/README.md) | 納品前受け入れテストシナリオ（業務フロー S01〜S12 + フォーム検証 V01〜V05。臨床安全・LIFF 通し・全フォームの入力/更新/DB 整合を担当） | 納品前検証・大きなリリース前 |
| [INTEGRATION_TEST_PLAN.md](INTEGRATION_TEST_PLAN.md) | テストの三層構造（Unit/Integration/E2E）と負荷試験の戦略 | テスト方針の確認・変更時 |
| [E2E_TESTING_GUIDE.md](E2E_TESTING_GUIDE.md) | Playwright による業務フロー自動検証の実行・追加手順 | E2E テストの実行・追加時 |
| [SECTION_14_MANUAL_TEST_GUIDE.md](SECTION_14_MANUAL_TEST_GUIDE.md) | ブラウザによる詳細な手動検証シナリオ（browser-test スキルが使用） | 手動検証・ブラウザ QA 時 |
| [PERFORMANCE_PROFILING.md](PERFORMANCE_PROFILING.md) | レスポンスタイム目標とプロファイリング手法（pprof / Lighthouse） | 性能調査・ボトルネック特定時 |

## AI エージェント向け注記

- カバレッジ基準の正本は [../coverage-policy.md](../coverage-policy.md)（ratchet 方式）。
- BE9完了後のHTTPテストは各`internal/<domain>`と`cmd/api`に配置する。削除済み`internal/handler`を前提にした旧集計はgit履歴で参照する。
- 検証の手順正本は [scenarios/](scenarios/README.md)。FAIL / 要対応は root [`STATUS.md`](../../../STATUS.md) へ短く記録する（専用ブラウザ結果レポートや旧 FUNCTIONAL_TEST_REPORT は置かない — git 履歴参照）。

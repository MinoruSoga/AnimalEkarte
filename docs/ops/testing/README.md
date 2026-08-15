# testing/ — テスト・品質保証

> **目的**: テスト戦略・検証手順・品質記録のドキュメント索引を提供する。  
> **読者**: 全開発者・QA・AI エージェント。  
> **タイミング**: テスト実施・テスト戦略変更・品質検証の前。

編集時のルール・品質基準の原則は [CLAUDE.md](CLAUDE.md) を参照。

## 索引

| ドキュメント | 内容 | いつ読むか |
|:---|:---|:---|
| **[TEST_ARCHITECTURE.md](TEST_ARCHITECTURE.md)** | **テストアーキテクチャ（層 L0–L5・受入 L4・項目単位・記録）** | **受入・戦略確認の最初** |
| [UAT-ENV-SETUP.md](UAT-ENV-SETUP.md) | 受入実行環境の準備・チェックスクリプト | scenarios 実行前 |
| [scenarios/](scenarios/README.md) | 納品前受け入れ（S01–S13 + V01–V05 + 項目単位 F） | 納品前・大きなリリース前 |
| [scenarios/FIELD-LEVEL-PROTOCOL.md](scenarios/FIELD-LEVEL-PROTOCOL.md) | フォーム**項目単位**チェック F0–F6 | V シリーズ実施時 |
| [scenarios/FORM-FIELD-INVENTORY.md](scenarios/FORM-FIELD-INVENTORY.md) | 全フォーム×項目の棚卸し | 項目単位カバー確認 |
| [INTEGRATION_TEST_PLAN.md](INTEGRATION_TEST_PLAN.md) | Unit/Integration/E2E・負荷試験 | 自動テスト方針 |
| [E2E_TESTING_GUIDE.md](E2E_TESTING_GUIDE.md) | Playwright E2E 実行・追加 | L3 回帰 |
| [SECTION_14_MANUAL_TEST_GUIDE.md](SECTION_14_MANUAL_TEST_GUIDE.md) | ドメイン重点の手動/browser-test | L5 補完 |
| [PERFORMANCE_PROFILING.md](PERFORMANCE_PROFILING.md) | pprof / Lighthouse | 性能調査 |

## AI エージェント向け注記

- **受入の正本は [scenarios/](scenarios/README.md)**。アーキテクチャは [TEST_ARCHITECTURE.md](TEST_ARCHITECTURE.md)。
- フォーム受入は **項目単位まで**（[FIELD-LEVEL-PROTOCOL.md](scenarios/FIELD-LEVEL-PROTOCOL.md)）。C1 代表 1 項目だけでは完了にしない。
- 環境: `./docs/ops/testing/scripts/check-uat-env.sh` → [UAT-ENV-SETUP.md](UAT-ENV-SETUP.md)。
- カバレッジ基準の正本は [../coverage-policy.md](../coverage-policy.md)（ratchet 方式）。
- BE9 完了後の HTTP テストは各 `internal/<domain>` と `cmd/api`。
- FAIL / 要対応は root [`todo.md`](../../../todo.md) 受入バグ節。シナリオ md に結果を書かない。証跡は `reports/uat-YYYY-MM-DD/`。

# testing/ — テスト・品質保証

> **目的**: テスト戦略・検証手順・品質記録の索引を提供する。

編集規則は [CLAUDE.md](CLAUDE.md) を参照する。

## 索引

| ドキュメント | 内容 | いつ読むか |
|:---|:---|:---|
| **[TEST_ARCHITECTURE.md](TEST_ARCHITECTURE.md)** | **L0–L5、証跡、環境境界** | **最初** |
| [UAT-ENV-SETUP.md](UAT-ENV-SETUP.md) | stack と明示的な fixture/account provisioning | scenarios 実行前 |
| [scenarios/](scenarios/README.md) | S01–S13、V01–V05、項目単位 F | 納品前・大きなリリース前 |
| [scenarios/FIELD-LEVEL-PROTOCOL.md](scenarios/FIELD-LEVEL-PROTOCOL.md) | フォーム項目単位 F0–F6 | V シリーズ実施時 |
| [scenarios/FORM-FIELD-INVENTORY.md](scenarios/FORM-FIELD-INVENTORY.md) | 宣言済みフォーム群・項目。wildcard/実測待ちを含む | カバー範囲確認 |
| [INTEGRATION_TEST_PLAN.md](INTEGRATION_TEST_PLAN.md) | Unit/API/E2E と負荷試験の現状 | 自動テスト方針 |
| [E2E_TESTING_GUIDE.md](E2E_TESTING_GUIDE.md) | 実装済み Playwright coverage と runner | L3 回帰 |
| [liff-verification.md](liff-verification.md) | LIFF の mock/実 LINE と秘密管理境界 | LIFF/LINE 検証 |
| [SECTION_14_MANUAL_TEST_GUIDE.md](SECTION_14_MANUAL_TEST_GUIDE.md) | L5 の focused exploratory supplement | 補完確認 |
| [PERFORMANCE_PROFILING.md](PERFORMANCE_PROFILING.md) | Lighthouse、k6、SQL 分析、現行 profiler の制約 | 性能調査 |

## 重要な現状

- migration は `002_master` だけをロードし、UAT account/clinical fixture は作らない。準備は [UAT-ENV-SETUP.md](UAT-ENV-SETUP.md) に従う。
- E2E の GitHub workflow は manual・non-gating で、account/fixture provisioning がないため authenticated suite は現在 BLOCKED。
- performance workflow の k6 job も fresh master-only DB に account provisioning がなく、現在 BLOCKED。
- 85 は宣言済みフォーム群の集計であり、inventory の `dynamic_*`、`UI 全項目`、`etc.`、`要実測` が解消されるまで「全 field を網羅」とは言わない。
- カバレッジ基準は [../coverage-policy.md](../coverage-policy.md)。確認済み UAT FAIL は `bug.md` で重複確認・記録後に Linear で追跡する。その他の新規 defect は通常の Linear intake に従う。

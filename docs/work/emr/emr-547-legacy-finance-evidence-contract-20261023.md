# EMR-547 移行原本証拠の文書整備記録

本書は実施記録の索引であり、業務/技術仕様の正本ではない。

2026-10-23、producer静的報告を受領し、技術本文を[会計領域の設計候補](../../design/domains/accounting/legacy-finance-evidence.md#legacy-evidence-purpose-and-ownership)へ移設した。旧本文は複製しない。業務条件は[会計正本](../../spec/domains/accounting/lifecycle.md#requirement-fr-accounting-legacy-evidence-preservation)と[見積正本](../../spec/domains/estimates/business-rules.md#requirement-fr-estimates-legacy-evidence-preservation)。進捗/担当の正本は[EMR-547](https://app.plane.so/baritechllc/browse/EMR-547/)。

設計未完・実装未着手。docs整備/報告受領は実コード/実DB/bundle独立検証や正式import許可ではない。tests/build/codegen/DB/import/Docker/browserはNOT_RUN。

2026-10-23返却設計を[同じ技術正本](../../design/domains/accounting/legacy-finance-evidence.md#legacy-evidence-open-nested-types)へ統合。共通構造・保存制約・readonly API・合成fixture条件は選定済み。6入れ子型、実型/原値表現/profile、manifest/snapshot/artifact bindingと集約元全入力は未閉鎖。正式移行HOLD、19 CSV維持、通常finalize/allocate生成なし。製品保存/read/UI・DDL/API/生成/試験/利用開始は未完、今回NOT_RUN。

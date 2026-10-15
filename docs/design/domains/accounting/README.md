# 会計の技術設計

対象は[業務正本の初回5点](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-first-five)と[第二群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-second)・[第三群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-third)・[第四群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-fourth)・[第五群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-fifth)（計8FR／78AC）の部分設計。D10は進行中。実装・API／DB変更・移行・実機検証は今回行わない。

読む順：業務正本と[語彙](../../../spec/ontology/accounting.md) → [技術仕様](technical-spec.md) → 現行API／コード。確定した業務と提案する実現方式、未決業務を分ける。[ADR003](../../../architecture/adr/003-payment-method-identity-and-consistency.md)・[write-owner](../../../architecture/model-write-owner-catalog.md)を再定義しない。

予約・受付・カルテの確定済み境界、保存§2、通常Undo停止を維持する。締め内部はD12、見積はD11、在庫はD13、税・端数の計算方針は[第五群の計算契約](../../../spec/domains/accounting/lifecycle.md#accounting-calculation-order)へ従い、品目の税務分類・保険具体式・税務適合全体の確認は医院へ残す。正本索引への掲載は実装や運用開始の承認ではない。

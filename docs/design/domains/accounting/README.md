# 会計の技術設計

対象は[業務正本の初回5点](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-first-five)と[第二群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-second)・[第三群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-third)・[第四群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-fourth)・[第五群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-fifth)・[第六群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-sixth)・[第七群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-seventh)（計9FR／99AC）の部分設計。D10は進行中。実装・API／DB変更・移行・実機検証は今回行わない。

読む順：業務正本と[語彙](../../../spec/ontology/accounting.md) → [技術仕様](technical-spec.md) → 現行API／コード。確定した業務と提案する実現方式、未決業務を分ける。[ADR003](../../../architecture/adr/003-payment-method-identity-and-consistency.md)・[write-owner](../../../architecture/model-write-owner-catalog.md)を再定義しない。

予約・受付・カルテの確定済み境界、保存§2、通常Undo停止を維持する。締め内部はD12、見積はD11、在庫はD13、税・端数の計算方針は[第五群の計算契約](../../../spec/domains/accounting/lifecycle.md#accounting-calculation-order)へ従い、品目の税務分類・税務適合全体の確認は医院へ残す。保険の資格／残限度・負担額はスタッフ確認・記録とし、会社／商品別の自動給付算式は将来の別スコープ。正本索引への掲載は実装や運用開始の承認ではない。

第六群の除外指定snapshot・帳票projection・マスタ変更権限の読取mappingは[技術仕様](technical-spec.md#accounting-sixth-design)へ。現行の固定支払選択肢や保険自動設定の有無を新しい将来方針へ昇格させず、grant・設定は変更しない。

第七群の確認結果・契約・負担額snapshotと外部保険業務の境界は[技術仕様](technical-spec.md#accounting-seventh-design)へ。保険会社への請求・入金照合・査定差額管理や同一診療費の複数契約併用を追加しない。

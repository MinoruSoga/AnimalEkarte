# 会計の技術設計

対象は[業務正本の初回5点](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-first-five)と[第二群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-second)・[第三群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-third)・[第四群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-fourth)・[第五群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-fifth)・[第六群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-sixth)・[第七群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-seventh)・[第八群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-eighth)・[第九群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-ninth)・[第十群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-tenth)・[追加の死亡後精算](../../../spec/domains/accounting/lifecycle.md#accounting-posthumous-settlement)（計9FR／144AC）の文書設計。完成条件は[末尾](technical-spec.md#accounting-document-completion)へ。実装・API／DB変更・移行・実機検証は今回行わない。

読む順：業務正本と[語彙](../../../spec/ontology/accounting.md) → [技術仕様](technical-spec.md) → 現行API／コード。確定した業務と提案する実現方式、未決業務を分ける。[ADR003](../../../architecture/adr/003-payment-method-identity-and-consistency.md)・[write-owner](../../../architecture/model-write-owner-catalog.md)を再定義しない。

予約・受付・カルテの確定済み境界、保存§2、通常Undo停止を維持する。締め内部はD12、見積はD11、在庫はD13、税・端数の計算方針は[第五群の計算契約](../../../spec/domains/accounting/lifecycle.md#accounting-calculation-order)へ従い、品目の税務分類・税務適合全体の確認は医院へ残す。保険の資格／残限度・負担額はスタッフ確認・記録とし、会社／商品別の自動給付算式は将来の別スコープ。正本索引への掲載は実装や運用開始の承認ではない。

第六群の除外指定snapshot・帳票projection・マスタ変更権限の読取mappingは[技術仕様](technical-spec.md#accounting-sixth-design)へ。現行の固定支払選択肢や保険自動設定の有無を新しい将来方針へ昇格させず、grant・設定は変更しない。

第七群の確認結果・契約・負担額snapshotと外部保険業務の境界は[技術仕様](technical-spec.md#accounting-seventh-design)へ。保険会社への請求・入金照合・査定差額管理や同一診療費の複数契約併用を追加しない。

第八群の減額時残高復元と併用内訳確認、自動失効なし・同医院同飼主境界、新番号再請求と二重請求防止、同時回収の元請求別配分、共通料金のowner-only参照は[技術仕様](technical-spec.md#accounting-eighth-design)へ。再請求の価格／税／割引継承・旧入金／保険の確認は第九群で確定。現在認可／復旧の目標は全体契約へ、実grant／物理schema／実装は後続。実返金・金銭履歴を自動複製せず、督促・貸倒・医院間／飼主間移管を追加しない。

第九群の元snapshot継承とスタッフ修正、旧入金の帰属変更と元実受領日保持、保険の再確認、返金元内訳・理由の登録と現在残検査は[技術仕様](technical-spec.md#accounting-ninth-design)へ。内訳登録だけで実返金を生成せず、復元済み預り金の二重使用を防ぐ。業務と後続実装条件は[正本の分類](../../../spec/domains/accounting/lifecycle.md#accounting-open-decisions)に分け、実装・全体受入の完了とは扱わない。


第十群の空明細・行別丸め・除外初期値／欠落時false・明示再適用は[第十群設計](technical-spec.md#accounting-tenth-design)。状態／金銭・操作別保存／認可・旧値・3画面／4master・受け手は[全体契約](technical-spec.md#accounting-reconciled-contract)へ。

追加の死亡後精算2ACはユーザー個別承認に基づき反映・独立文書レビュー済み。初回〜第十群の142AC本文を保持し、死亡状態を戻さず生前の元診療・理由・履歴と元飼主資金を保持する。現在の全体状態は[追加9点の進捗](../../../spec/domains/SPECIFICATION_PROGRESS.md#additional-business-decisions-nine)を参照する。

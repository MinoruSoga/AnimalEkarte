# 会計の技術設計

対象は[業務正本の初回5点](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-first-five)と[第二群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-second)・[第三群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-third)・[第四群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-fourth)・[第五群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-fifth)・[第六群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-sixth)・[第七群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-seventh)・[第八群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-eighth)・[第九群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-ninth)・[第十群](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-tenth)・[追加の死亡後精算](../../../spec/domains/accounting/lifecycle.md#accounting-posthumous-settlement)・[支払方法停止](../../../spec/domains/accounting/lifecycle.md#accounting-inactive-payment-method)（前段階9FR／151AC）と[会議連携4AC](../../../spec/domains/accounting/lifecycle.md#accounting-trimming-adjustment-handoff)を含む基準9FR／155ACに[品目別保険設定](../../../spec/domains/accounting/lifecycle.md#requirement-fr-accounting-item-insurance-setting)1FR／12ACを追加した前段階10FR／167ACに、[2026-10-17の医師確認・プラン同期・請求対象](../../../spec/domains/accounting/lifecycle.md#doctor-review-sync-billing-20261017)3FR／7ACを加えた、基準13FR／174ACにNo.8の個別割引権限1AC・No.30の釣銭整合3AC・過受領処理3ACを加え、現在のライフサイクル本文は13FR／184AC（EMR-367追補3ACを含む）の文書対象。完成条件は[末尾](technical-spec.md#accounting-document-completion)へ。実装・API／DB変更・移行・実機検証は今回行わない。

読む順：業務正本と[語彙](../../../spec/ontology/accounting.md) → [技術仕様](technical-spec.md) → 現行API／コード。確定した業務と提案する実現方式、未決業務を分ける。[ADR003](../../../architecture/adr/003-payment-method-identity-and-consistency.md)・[write-owner](../../../architecture/model-write-owner-catalog.md)を再定義しない。

[医師確認集合の非活性内部契約](doctor-review-internal-projection.md#review-internal-scope)は2026-10-23選定。承認済みidentity/数量/record集合だけの失効signature/digestと純粋世代参加、確認時限定signature不変保存の論理方向、T03/T06共用医師reader再利用方向だけを扱う。metadata/技術link-onlyでは失効させず、新current列の自動採用は撤回。既存481 snapshotは現在採用identityを証明できる場合のみ再利用する。実DB known provider・確認時signatureの物理保存・HTTP・確認/Return/Complete・全writer閉包・試験/activationは未供給で、製品Readyではない。

[預り金の初回受領・同口座再読込](owner-deposit-initial-receipt.md)は2026-10-23選定済みの限定技術契約。予約指定なし・新実受領1件全額のbackend単位だけを扱い、共有供給/実装/試験/利用開始は未完。506一般公開wire／全fund CASと509締め設計の採択ではない。

予約・受付・カルテの確定済み境界、保存§2、通常Undo停止を維持する。締め内部はD12、見積はD11、在庫はD13、税・端数の計算方針は[第五群の計算契約](../../../spec/domains/accounting/lifecycle.md#accounting-calculation-order)へ従い、品目の税務分類・税務適合全体の確認は医院へ残す。保険の資格／残限度・負担額はスタッフ確認・記録とし、会社／商品別の自動給付算式は将来の別スコープ。正本索引への掲載は実装や運用開始の承認ではない。

第六群の除外指定snapshot・帳票projection・マスタ変更権限の読取mappingは[技術仕様](technical-spec.md#accounting-sixth-design)へ。現行の固定支払選択肢や保険自動設定の有無を新しい将来方針へ昇格させず、grant・設定は変更しない。

第七群の確認結果・契約・負担額snapshotと外部保険業務の境界は[技術仕様](technical-spec.md#accounting-seventh-design)へ。保険会社への請求・入金照合・査定差額管理や同一診療費の複数契約併用を追加しない。

第八群の減額時残高復元と併用内訳確認、自動失効なし・同医院同飼主境界、新番号再請求と二重請求防止、同時回収の元請求別配分、共通料金のowner-only参照は[技術仕様](technical-spec.md#accounting-eighth-design)へ。再請求の価格／税／割引継承・旧入金／保険の確認は第九群で確定。現在認可／復旧の目標は全体契約へ、実grant／物理schema／実装は後続。実返金・金銭履歴を自動複製せず、督促・貸倒・医院間／飼主間移管を追加しない。

第九群の元snapshot継承とスタッフ修正、旧入金の帰属変更と元実受領日保持、保険の再確認、返金元内訳・理由の登録と現在残検査は[技術仕様](technical-spec.md#accounting-ninth-design)へ。内訳登録だけで実返金を生成せず、復元済み預り金の二重使用を防ぐ。業務と後続実装条件は[正本の分類](../../../spec/domains/accounting/lifecycle.md#accounting-open-decisions)に分け、実装・全体受入の完了とは扱わない。


第十群の空明細・行別丸め・除外初期値／欠落時false・明示再適用は[第十群設計](technical-spec.md#accounting-tenth-design)。状態／金銭・操作別保存／認可・旧値・3画面／4master・受け手は[全体契約](technical-spec.md#accounting-reconciled-contract)へ。

追加の死亡後精算2ACはユーザー個別承認に基づき反映・独立文書レビュー済み。初回〜第十群の142AC本文を保持し、死亡状態を戻さず生前の元診療・理由・履歴と元飼主資金を保持する。現在の全体状態は[追加9点の進捗](../../../spec/domains/SPECIFICATION_PROGRESS.md#additional-business-decisions-nine)を参照する。

支払方法停止の追加7ACは[追加設計](technical-spec.md#accounting-inactive-payment-design)へ対応する。既存144AC本文・IDとレビュー履歴は保持し、追加差分も独立内容レビュー・限定構造検査済み。新受領の選択／保存時検査と過去参照・集計を分離し、既存資金の配分・充当／解除・返金へ新受領のgateを一律適用しない。方法追加・ENUM廃止・初期値／移行は本合意から確定しない。

品目別保険設定は[保存・読取・会計引継ぎ設計](technical-spec.md#accounting-item-insurance-design)へ。診察・処置・予防・健診を薬剤・検査の既存設定と整合させ、実適用／資格／補償保証や全体割引除外とは独立に扱う。新規4種別の新規登録は保険対象外を初期値とし、既存明示値を保持する。旧欠落値の解釈・変換は後続の移行課題として管理し、新規機能の採否未決へ戻さず、snapshot・明示再適用／訂正保護を維持する。

<a id="legacy-finance-evidence-design-entry"></a>
## 移行原本証拠の保存・閲覧設計（EMR-547）

[legacy-finance-evidence.md](legacy-finance-evidence.md#legacy-evidence-purpose-and-ownership)に原本occurrence/派生版の多対多lineage、identity、関連検証、選定read契約、producer報告の受領と設計残を置く。[会計FR/6AC](../../../spec/domains/accounting/lifecycle.md#requirement-fr-accounting-legacy-evidence-preservation)と[見積FR/1AC](../../../spec/domains/estimates/business-rules.md#requirement-fr-estimates-legacy-evidence-preservation)を実現する共通構造・保存tableの役割・関係制約（chunk/field/payload上限は未選定）・readonly APIは2026-10-23選定。6入れ子型・verification_ref検証記録/供給binding・配送local→stable UUID耐久対応と実データprofile/manifest/集約元列挙は未閉鎖で、全wire完成ではない。実DDL/API/UIの実装・試験NOT_RUN。旧work pathは実施記録の索引のみ。

[A1/A2区分締めの限定契約](../finance-operations/cash-closing-contract.md#cash-closing-source-completeness)は464初回受領の2sourceを消費する別の後発選定。預り受領契約自体が509締め全体や非R1 producer/B grantを採択したという意味ではない。実装/供給/利用開始は未完。

<a id="doctor-review-confirm-return-entry"></a>
## 既存会計の医師確認・差戻し

[Confirm/Return契約](doctor-review-confirm-return-contract.md#review-actions-scope)は2026-10-24に限定技術選定。先行の[内部signature](doctor-review-internal-projection.md#review-internal-scope)を変えず、既存billingの現在集合read・明示確認・理由付き差戻し・保存signature/履歴・C06/元target read・旧writer迂回排除を閉じる。billing参照の版更新はstorage guard一回で手動加算なし。共通gate/医師predicate/current provider/writer閉包/API・DDL・model供給/実装・受入は未完。準備billing・Complete・金銭/在庫・生前分C・退院Dは別packetで、全体Ready/実装完了とはしない。

[EMR-573 他ペット未納ヘッダー](other-pet-unpaid-contract.md)は金額を公開しない三値read、全期間候補、元請求の既存精算入口までの限定契約。新金銭writerや独自pet配賦を追加しない。

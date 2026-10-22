# EMR-445 ②全マスタの着手プランと採用証跡

要件判断の窓口はMinoru Soga。既決仕様準拠・重複工程削減・計画coverage指標と未測定の業務成果の区別は[共通の要件検証](emr-445-stage1-contract-start-plan-20261021.md#requirement-validation-20261021)へ従う。実装担当claim・開始許可・実績測定は別。試験NOT_RUN。

2026-10-21。対象はsettings索引21本文と、その索引外の医院/法人・LINE予約設定・検査機器対応・package。医院と法人は別resourceとして行を分けたため下表は26行だが、26件の新規実装票・PRではない。

基準main `a88fa4add58d43cff643590f9f507c8ab1a3c085`（10:45 UTC再取得）。業務正本は[settings索引](../../spec/screens/settings/README.md)と各領域の確定FR/AC。[①②③段階台帳](emr-445-master-stages-ledger-20261021.md)の直接22票＋498は管理対象であって全属性実装完了の証明ではない。現コードは差分確認の証拠であり、正本を縮小しない。実装・試験・DB操作は今回行っておらずNOT_RUN。

新規管理4票EMR-527/528/529/530の起票・Readyは管理責務の具体化であり、実装・生成実行・試験の開始承認ではない。その他の条件付き新規票案は作成済みへ変更しない。

## 1. 責務・開始条件・完了の意味

- 447/448/449は集約・引渡し票。実装claimは既存実装票の限定subset、または下記の具体差分単位に固定する。450は③共通UIであり、全CRUDや認可モデルの代替ではない。
- ①API/schema/model/generatedが必要な行は、その必要subsetの収録/採用commitを単一統合ownerから受領する。api.yamlを複数claimで同時に編集しない。生成型同期は[EMR-527](https://app.plane.so/baritechllc/browse/EMR-527/)の独立後続で担当未割当、①生成型後続の単一claim ownerへ受領する。旧session3表記は履歴で、現claimは着手時に確認する。
- ②は設定保存/再GET・停止削除と保存済み参照保護まで。④の全予約/臨床/会計consumer完成を逆依存させない。ただし既存参照を壊す新writerを有効化しない。
- 各属性の型・null/省略・保存・再GET、旧writer、現在認可、履歴保持、関連fixture期待値を一票の完了条件にする。対応済み関数は再作成せず採用SHAを記録する。無差分は採用証跡だけで閉じ、空PRや試験実施だけの新規開発票を作らない。
- 表中「採用」は記載した操作・主要guardの静的存在確認に限定する。全属性の実行適合や未読の全callerをPASSにしない。未確認を欠陥とせず、追加調査が必要なら属性・経路を特定する。
- 新規restore/reorder UI、一律C06/CAS、物理削除、独自履歴表、臨床値/算式/権限は正本なしに追加しない。試験は後日の本人指示までNOT_RUN。UAT/Doneの既存成果を再オープンしない。

## 2. 全master coverage・専有path

以下pathはrepository rootからの相対path。`B=backend/internal/`、`F=frontend/src/features/`。各行のservice/pathは固定mainで取得確認済み。handlerとserviceをまとめた `{handler,service}` は両ファイルを意味する。固定ソースURL基点は https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/ 。

### 1. 動物種
- 属性/操作：name/active/order。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：445採用証跡 / 無差分採用候補 / pet animal_species＋専有FE。
- BE：B/pet/animal_species_{handler,service}.go。model：B/model/animal_species.go。FE：F/master/routes/AnimalSpeciesSettings.tsx。
- 判定・境界：既採用：CRUD/reorder、handler system_admin、FE isSystemAdmin、使用中削除409、監査DIを確認。新機能不足は認定しない。453検索成果とは別にこの採用を責務票へ記録。
### 2. ケージ
- 属性/操作：name/type/size/price/description/active/order・連結。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：448 / 306既採用のmaster受領＋461 read subset / cage系のみ。
- BE：B/medicalrecord/cage_{handler,service}.go。model：B/model/cage.go。FE：F/master/routes/CageSettings.tsx。
- 判定・境界：既採用：CRUD/merge/unmerge/occupancy、Deleteの同TX lockと連結guardを確認。461必要readの受領と現場occupancy実行は未確認だが、新しいmaster CRUDを再作成しない。
- 追加静的残差：description空clearは§7の448行に接続する。既存CRUD/連結成果は再実装しない。
### 3. 入院・宿泊プラン
- 属性/操作：name/body_size/billing_unit/price/tax/description/active。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：448集約 / EMR-528（MM-HOSPITAL-OPTIONAL） / §3Aの専有差分。
- BE：B/medicalrecord/hospitalization_plan_{handler,service}.go。model：B/model/hospitalization_plan.go。FE：F/master/routes/HospitalizationSettings.tsx。
- 判定・境界：既存CRUD/税nullable model/ケア参照削除guard採用。備考空・体格/料金単位null・明示0送信の差分は§3Aの固定契約で修正する。
- 追加静的残差：同serviceのSortOrder/TaxRate型adapterもEMR-528の受入へ含める（§3A/§7）。
### 4. トリミングコース/option
- 属性/操作：名称/size価格/所要時間/任意略称・説明・画像/曜日/順序/休止。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：448 / 457成果＋458-A / trimming_course/option/T2。
- BE：B/trimming/trimming_course_{handler,service}.go。model：B/model/trimming_master.go（course/option/T2。trimming.goは施術記録で別物）。FE：F/master/routes/TrimmingSettings.tsx。
- 判定・境界：既採用：曜日/画像属性正規化、画像lifecycle監査、参照削除保護。458 T2追加APIは①未収録として明確な残。457 UAT再実装なし。
- 追加静的残差：course/optionのdescription空clearは§7の448行。既存の短縮名/曜日/画像の明示clear成果を維持する。
### 5. トリミングコース分類
- 属性/操作：名称/active/order。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：448 / 445採用証跡（MM-FACILITY） / trimming_course_type系のみ（457と非重複）。
- BE：B/trimming/trimming_course_type_{handler,service}.go。model：B/model/trimming_course_type.go。FE：F/master/routes/TrimmingCourseTypeSettings.tsx。
- 判定・境界：既採用：CRUD/reorder、WithTx内参照countと削除。分類をreservation business_kindへ変える追加実装なし。
### 6. 検査プラン本体CRUD・検査測定field（primary）
- 属性/操作：本体/親/価格/保険・field名称/単位/原値/基準値/順序。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：447 / 445採用証跡（MM-CLINICAL-AUX） / exam_type/field。検査プラン本体CRUDのprimaryは本row6。18は機器mapping、169は健診package、225は健診fieldとして別境界へ参照。
- BE：B/medicalrecord/exam_type_{handler,service}.go。model：B/model/examination_type.go。FE：F/master/routes/TreatmentPlanMaster.tsx。
- 判定・境界：既採用：exam_type CRUD/parent医院検査/参照guardとExaminationType.Items。field操作はexam_type_field.goへ。未承認の基準値投入は技術closureとは別。
- 責務補足：row16の「検査」は本row6参照とし本体CRUDを二重所有しない。検査fieldはexam_type_fields、EMR-225はcheckup_type_fieldsの健診fieldで代用しない。18は機器mapping採用、169は健診package採用へ参照し、checkups/master-medical/lab-importの各grantを相互代用しない。再実装・重複起票なし。
### 7. 支払方法
- 属性/操作：名称/active/標準区分/順序。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：449 / 465-M / payment_method_master＋guard部品。
- BE：B/billing/payment_method_master_{handler,service}.go。model：B/model/payment_method_master.go。FE：F/master/routes/PaymentMethodSettings.tsx。
- 判定・境界：既採用：標準method削除拒否とusage count。465-Mで新fund/event参照との接続、停止writer/共通guard部品/引渡しを閉じる。全受領caller接続465-Cは④であり②の完了gateではない。
### 8. 保険
- 属性/操作：name/coverage_rate/contact/description/active/order。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：449 / 445採用証跡（MM-FINANCE-AUX） / insurance系。
- BE：B/billing/insurance_{handler,service}.go。model：B/model/insurance.go。FE：F/master/routes/InsuranceSettings.tsx。
- 判定・境界：既採用：CoverageRate int、連絡先/備考、CRUD、pet参照削除拒否。新算式なし。全旧consumerの給付計算は④。
- 追加静的残差：description/contact_phoneの空clearは§7の449行。給付算式や全会計consumerは変更しない。
### 9. キャンペーン
- 属性/操作：name/期間/rate又はamount/対象カテゴリ・商品/active。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：449集約 / EMR-530（MM-FINANCE-EXACT-MASTERS、§3C） / campaign数値adapter。
- BE：B/billing/campaign_{handler,service}.go。model：B/model/campaign.go。FE：F/master/routes/CampaignSettings.tsx。
- 判定・境界：期間/率額/商品医院検査/typed対象CRUD採用。master writerがCampaign/Targetだけを更新し会計snapshotを書かない保護は採用するが、DiscountValueのExactNumericとrequest/service/response float64の不一致は②数値adapter残として§3Cで閉じる。
### 10. 物販
- 属性/操作：name/category/price/tax/active/order。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：449 / EMR-530（MM-FINANCE-EXACT-MASTERS、§3C2） / merchandise_item税率adapter。
- BE：B/inventory/merchandise_item_{handler,service}.go。model：B/model/merchandise_item.go。FE：F/master/routes/MerchandiseItemSettings.tsx。
- 判定・境界：既採用：CRUDと同TX削除、請求/見積/campaign参照guard。TaxRateのExactNumericとrequest/service/response float64の不一致は②の保存/再取得差分。§3C2のexact DTO境界へ合わせ、整数単価や全master型を無断変更しない。
### 11. 締め設定
- 属性/操作：標準境界/曜日/特別期間/個別休診日。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：449 / 445採用証跡（MM-CLOSING） / clinic closing/holiday/special period＋closing FE。
- BE：B/clinic/closing_settings_{handler,service}.go。model：B/model/clinic_settings.go。FE：F/closing-settings/routes/ClosingSettingsPage.tsx。
- 判定・境界：既採用：closing setting取得更新service、clinic settings model、専用FE route。特別期間/休診日は別handlerへ。LINE設定との統合はしない。
### 12. 職員
- 属性/操作：人物/account/所属/職種/資格/有効・能力/院別pause/休業/公開profile。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：448 / 493＋366（293追跡） / staff＋assignment専有hunk。
- BE：B/staff/staff_{handler,service}.go。model：B/model/staff.go。FE：F/master/routes/StaffSettings.tsx。
- 判定・境界：既採用：staff model/handler/service/FE。明確な残は493選定leave/eligibility API未収録と366 trusted actor必要subset。profile全体再実装なし。
- 採用境界：既存staff保存の[#702](https://github.com/MinoruSoga/AnimalEkarte/pull/702)（merge `dd1b5a76cdae42c03e55417f38c04083b6fa086f`）も採用。StaffSettings/StaffSidePanel/use-master-saveの基本保存→変更した権限group→所属医院→対応予約区分のawait、現在権限再確認、保存中追加編集/旧パネル遅延保護、観測CREATE成功ID保持→再保存UPDATE、全成功時のみdirty解除/通知/closeを再実装しない。③状態/dirty接続は450/M18、集約448。493のleave/eligibility/勤務/copy/CSV、既知公開4field clearとは別。複数API全体の原子性と未観測通信断unknownの冪等性は#702の保証でなく、実API/実staff writeも今回未確認。[coverageの現在追跡](emr-stage123-coverage-audit-20261022.md#a10-current-trace-20261022)を参照し、新実装scopeを増やさない。
- 追加静的残差：license_numberと予約公開profileの限定4fieldの空clearは§7の448行。password/emailへ一般化しない。
### 13. 職種
- 属性/操作：name/description/active/order。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：448 / 445採用証跡（MM-WORKFORCE-AUX） / occupation系、staff business変更除外。
- BE：B/staff/occupation_{handler,service}.go。model：B/model/occupation.go。FE：F/master/routes/OccupationSettings.tsx。
- 判定・境界：既採用：CRUD/順序、同TX active lock→staff usage→delete。役職名による予約grant復活なし。
- 追加静的残差：description空clearは§7の448行。職種/所属/能力を再設計しない。
### 14. 勤務template
- 属性/操作：name/type/開始終了/複数休憩/notes/active/order。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：448 / 445採用証跡（MM-WORKFORCE-AUX） / shift_template系、493 CSV/copy除外。
- BE：B/staff/shift_template_{handler,service}.go。model：B/model/staff.go。FE：F/shifts/routes/ShiftTemplateSettings.tsx。
- 判定・境界：既採用：staff.go内ShiftTemplate/Break、CRUD・breaks更新・正規化、同TX削除。コピー/CSVは493別操作、template CRUD再作成なし。
### 15. 予約区分/group
- 属性/操作：type/group/所要時間/個別意図/公開/分類/name版/順序。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：448 / 492 / reservation_type/group専有。
- BE：B/reservation/reservation_type_{handler,service}.go。model：B/model/reservation_type.go。FE：F/master/routes/ReservationTypeSettings.tsx。
- 判定・境界：既採用：基本handler/service/model/FE。492 rename/restore/C06追加wireと旧writer切替は明確な残。group解除/任意属性presenceの具体受領条件は§8。今回Dへ追加せず、既存CRUDを削除して全面再構築しない。
### 16. 診療項目5tab（検査本体CRUDはrow6参照）
- 属性/操作：診察/検査/処置/ワクチン/健診の名称/親/description/price/active/対象保険税。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：447 / 454455456既採用＋365 / 5種専有／shared454単一owner。
- BE：B/medicalrecord/consultation_{handler,service}.go。model：B/model/consultation.go。FE：F/master/routes/TreatmentPlanMaster.tsx。
- 判定・境界：既採用：5種個別serviceと共通FE、診察/処置454・ワクチン455・健診456成果。365 vaccine price文字列/numeric23,4は明確な①残。検査fieldとの差分は臨床補助unit。
- 追加静的残差：5tab descriptionの空clearは§7の447行。保険/CAS/既存UAT成果の全面再実装はしない。
### 17. 診断カテゴリ/病名
- 属性/操作：カテゴリ・病名のname/description/active/order/必須category。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：447集約 / EMR-529（MM-DIAGNOSIS-DESCRIPTION） / §3Bの専有差分。
- BE：B/medicalrecord/diagnosis_{handler,service}.go。model：B/model/diagnosis.go。FE：F/master/routes/DiagnosisSettings.tsx。
- 判定・境界：カテゴリ必須NOT NULL/同医院存在検査/参照削除guardは採用済み。備考clearだけ§3Bの限定差分。第1/第2複数診断は495④。
### 18. 薬剤・用量pattern
- 属性/操作：name/親/剤形/単位/price/tax/保険/active・ID別pattern/revision。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：447 / 498 master subset＋基本CRUD採用記録 / medicine/dose専有。
- BE：B/medicalrecord/medicine_{handler,service}.go。model：B/model/medicine.go。FE：F/master/routes/MedicineSettings.tsx。
- 判定・境界：既採用：基本CRUD、親/在庫医院検査、使用中削除guard。498 ID別pattern/revision API未収録が明確な残。既存revision modelは再作成しない。
- 段階移行残：薬剤base model→HTTP/callerのExactNumeric adapterは§7のEMR-447受入へ紐付ける。新規regressionや担当忘れと断定しない。
### 19. 問診template
- 属性/操作：title/category自由text/content plain text/active/order。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：447 / 445採用証跡（MM-CLINICAL-AUX） / inquiry_template系。
- BE：B/medicalrecord/inquiry_template_{handler,service}.go。model：B/model/inquiry_template.go。FE：F/master/routes/InterviewTemplateSettings.tsx。
- 判定・境界：既採用：InquiryTemplate text category/content、CRUD/reorder・使用中削除拒否。plain text/自由categoryを保持し、④挿入/コピーは別。
### 20. 主訴カテゴリ
- 属性/操作：name/description/active/order。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：447 / 445採用証跡（MM-CLINICAL-AUX） / chief_complaint系。
- BE：B/medicalrecord/chief_complaint_{handler,service}.go。model：B/model/chief_complaint_type.go。FE：F/master/routes/ChiefComplaintSettings.tsx。
- 判定・境界：既採用：CRUD/description/sort/active、問診参照削除拒否。飼主原文を更新する機能を足さない。
- 追加静的残差：description空clearは§7の447行。診断専有EMR-529へ主訴を混ぜない。
### 21. 権限group
- 属性/操作：name/description/color/active/order/rules/職員割当境界。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：445採用証跡 / 無差分採用候補 / auth permission＋専有FE。
- BE：B/auth/http_permission.go・auth/permission_group_service{,_mutate,_rules,_self_lockout}.go。model：B/model/permission_group.go。FE：F/master/routes/PermissionGroupSettings.tsx。
- 判定・境界：既採用：http_permissionのCRUD/rules routes、mutation同TX監査、自己view/edit保持判定。新しい権限モデル不足とは認定しない。旧staff割当writerはstaff_service_permissions.goの既存経路と連携。
### 22. 医院設定
- 属性/操作：基本連絡先/登録番号/税率/帳票表示設定。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：445採用証跡 / 無差分採用候補 / clinic基本設定＋専有FE。
- BE：B/clinic/clinic_{handler,service}.go。model：B/model/clinic.go。FE：F/clinic-settings/routes/ClinicMasterSettings.tsx。
- 判定・境界：既採用：医院設定保存field、update+reload同TX、delete時owner/staff/他参照検査、package catalog初期化。実医院設定の変更は対象外。
- 段階移行残：医院税率response adapterは§7のEMR-445既存採用行の限定差分へ紐付ける。加えてMM-CLINICの任意8文字列clearは§7の②医院設定専有subset、単一claim待ち。A共有API/DDL担当へFE実装を暗黙追加せず、実医院設定変更や530への追加をしない。
### 23. 法人設定
- 属性/操作：singleton基本情報/インボイス登録番号。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：445採用証跡 / 医院と同じ責務 / company singleton。
- BE：B/clinic/company_{handler,service}.go。model：B/model/company.go。FE：F/clinic-settings/components/CompanyInvoiceSection.tsx。
- 判定・境界：既採用：singleton Get/Update、医院設定grant付きPATCH、CompanyInvoiceSection。新法人CRUD/会社複数化/権限を追加しない。
### 24. package
- 属性/操作：名称/説明/active/順序/typed構成要素/版。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：445 / 521523既採用 / package master専有、型生成は①独立後続の単一claim owner。
- BE：B/medicalrecord/treatment_package_{handler,service}.go。model：B/model/treatment_package.go。FE：F/master/routes/TreatmentPackageSettings.tsx。
- 判定・境界：6 master操作/typed source/CAS/catalog/監査の#775成果採用。①生成型後続は独立・担当未割当、単一claim ownerへ受領。旧session3は履歴。522/524は④。
- ③追加受入：MasterPageShellのSPA戻るdirty保護はEMR-523/450 P01の操作表へ接続する。522/524の④applyを再開しない。
### 25. LINE予約設定
- 属性/操作：医院別予約設定/公開ページ入力/文言。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：448 / 445採用証跡（MM-LINE-CONFIG） / line_reservation_setting＋設定FE、492/493の属性writer除外。
- BE：B/reservation/line_reservation_setting_{handler,service}.go。model：B/model/line_reservation_setting.go。FE：F/line-reservation/routes/LineReservationSettings.tsx。
- 判定・境界：既採用：設定handler/service/modelとLineReservationSettings/PageEditor。実送信/外部連携credential/marketing実装は非対象。旧公開writerの範囲を設定専有unitへ。
### 26. 検査機器項目対応
- 属性/操作：source/code/unit/value_shape/field/active/order・mapping版。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：447 / 297 master設定subset / lab_device_item_master/mapping管理、通信consumer除外。
- BE：B/medicalrecord/lab_device_item_master_{handler,service}.go。model：B/model/lab_device_item_master.go。FE：F/master/routes/LabDeviceItemMasterSettings.tsx。
- 判定・境界：既存SaveConfiguration同TX/同source/field医院検査を採用。297-A設定writer版参加は公開採択§6/7a/7bに固定、§3D参照。297-B受信は④。

### 分割実装を落とさない補助path
- 診療5tabの専有serviceはB/medicalrecord/{consultation,exam_type,procedure,vaccine,checkup_type}_{handler,service}.go。FEはTreatmentPlanMaster/TreatmentPlanMasterViewへ合流、共有保険adapterは454単一owner。検査field writerはB/medicalrecord/exam_type_field.go。
- 権限割当旧writerはB/staff/staff_service_permissions.go、auth groupのpolicy lock/repositoryと同じ既存保護に従う。
- 締め標準と休診日は別handler（B/clinic/clinic_holiday_{handler,service}.go）。LINEページ編集はF/line-reservation/routes/LineReservationPageEditor.tsx。予約区分公開属性492/能力493との専有境界を保持する。



## 3. 具体差分と受領interface

### A. 入院プラン：EMR-528（MM-HOSPITAL-OPTIONAL）の最小1PR範囲
集約448、専有実装単位[EMR-528](https://app.plane.so/baritechllc/browse/EMR-528/)（MM-HOSPITAL-OPTIONAL、親448、Ready、個人担当未割当）。トリミング分類など無差分対象を同票へ増やさない。
根拠：master-hospitalization-plan.md「詳細編集サイドパネル」「API連携」、D05 master変更は過去/追加済会計へ自動反映しない。BE hospitalization_plan_request.goはUpdate Description/BodySize/BillingUnitを*string、service buildHospitalizationPlanUpdateはnilなら更新なし。FE routes/hospitalization-settings-model.tsは備考空をundefined、体格/料金単位空をnullとして送る。従って備考の空への変更と明示nullのクリアが静的に保存されない。
選定する最小契約:
- 既存PATCH /api/v1/masters/hospitalization-plans/{id}を維持。body_size/billing_unitは省略=保持、null=SQL NULL、許容enum値=設定。空文字をNULLの代用品にしない。
- BE request専有のpresence-aware field（Set/Null/Valueの3値、既存同形helperがあれば再利用）でJSONをdecodeし、service inputに同じpresenceを渡す。build update mapは省略をkeyなし、nullをmap値nil、値をenumにする。DB nullable列/modelは既存のまま。
- 備考はFE Updateでdescriptionを空文字もそのまま送る。BE既存*stringは空文字を保存できるのでこの欄のschema変更不要。
- 新規作成のFE price:number=0をprice || undefinedで落とさず、明示入力0を0として送る。任意値全般のnull→0変換拡張や旧データbackfillは本票でしない。既存画面の数値初期値0/税初期値は変えない。
- response/readのbody_size/billing_unitは既存nullableの意味を維持。DB NULLが現omitemptyで省略される既存responseはFEの??nullで読む。writeのnullと省略の差をread省略の問題へ混同しない。
追加受入：固定main c86c8dd2のhospitalization_plan_service.go L139/L141はinput.SortOrder（int）→model.SortOrder（*int）、taxRate（float64）→model.TaxRate（ExactNumeric）の直接代入が残る。同じserviceのcreate/update/request→model adapterを既存のclear/null/0差分と同じEMR-528で閉じ、presence/既定/明示0を維持して保存後再GETへ対応づける。既存の採択数値境界に従い、model/DDLを旧型へ戻さず、binary float由来の精度損失を新たな正規値としない。これはソースの静的型対応確認であり、ビルド失敗を実行観測した記録ではない。
変更hunk: frontend/src/features/master/routes/hospitalization-settings-model.ts、backend/internal/medicalrecord/hospitalization_plan_request.go/service.go、専有DTO/fixture。①api.yaml既存PATCH propertyのnullable/省略説明は単一ownerが同期。model/DDL新設なし。
履歴境界: 現repositoryはhospitalization_plansだけ更新、Deleteはcare_plan_items参照NOT EXISTS＋soft delete。これを採用し、snapshot新表/過去care更新をこの票へ足さない。405/481の④会計bindingをgateにしない。
受入fixture: 備考非空→空、体格/料金単位値→null、各省略→既存値保持、値→別値、不正enum zero-write、0価格作成、同院/他院、参照中削除拒否、master更新でcare/会計row無変更。NOT_RUN。

### B. 診断：カテゴリ契約は既決、備考clearだけ具体差分
集約447、専有実装単位[EMR-529](https://app.plane.so/baritechllc/browse/EMR-529/)（MM-DIAGNOSIS-DESCRIPTION、親447、Ready、個人担当未割当）。診断/主訴/template一式の再開発票に広げない。
確定: master-diagnosis.md「所属カテゴリ必須選択」。model.DiagnosisName.DiagnosisTypeIDはNOT NULL uint64、create request required、service Create/Updateが同医院カテゴリFindByIDで検査。分類nullableは未選定ではなく不要。PATCH省略は保持、明示0/不存在/他院は既存validation拒否。カテゴリそのものの設計変更なし。
具体差分: frontend/src/features/master/routes/diagnosis-settings-model.tsのType/Name UpdateがCreate builderを再利用し、description空をundefinedへ落とす。BE UpdateDescription *stringとbuild mapは空文字を保存可能。
修正: TypeとNameのUpdate builderだけdescription:data.descriptionを明示する。新規Create既定は変更不要。API/schema/DDL/model/BE変更不要。既存api/diagnosis.tsを経由する両更新callerを維持。
履歴: カテゴリ/病名master CRUDは既存diagnosis_service/repositoryを採用。カテゴリ子参照拒否、病名のカルテ参照拒否を保持。第1/第2複数診断snapshotは495④へ引渡し、本票でカルテ診断を再保存しない。
受入fixture: Type/Nameとも既存備考→空、空→非空、非空維持、有効切替/所属変更が無関係に消えない、カテゴリ必須と他院拒否。実試験NOT_RUN。

### C. campaign：既存保護を採用し、②のexact数値adapterを閉じる

集約449、専有実装単位[EMR-530](https://app.plane.so/baritechllc/browse/EMR-530/)（MM-FINANCE-EXACT-MASTERS、親449、Ready、個人担当未割当）。campaignと次の物販税率だけの保存/再GET/FE DTO整合を一つの限定PR範囲へ束ねる。既存449/465-M等のclaimと担当を確認し、同じhunkを二重実装しない。EMR-530の起票は実装・生成実行・試験の開始承認ではない。その他の条件付き未起票案は変更しない。

固定mainのmodel.Campaign.DiscountValueはExactNumeric/numeric(12,2)。campaign_request.goのcreate/update DiscountValue、campaign_service.goのCreate/UpdateCampaignInput・buildCampaignUpdate・resolveCampaignDiscount・validateCampaignDiscountとCreate代入、campaign_response.goのfloat64 responseが未追従であり、snapshot非更新だけで②無差分とはしない。期間/率額/対象集合/商品医院検査、repositoryのCampaign/Targetだけを更新する境界・同TX/参照保護は再利用する。

必要hunkは backend/internal/billing/campaign_{request,service,response}.go、同repositoryの列bind、frontend/src/features/master/api/campaign.ts（BackendCampaign/request/transform）、frontend/src/features/master/components/CampaignSidePanel.tsxのNumber(v)変換、lib/campaign-side-panel-model.tsのform値、routes/CampaignSettings.tsxのsave builderの当該discount_value接点。採択[483の十進輸送/DiscountTerms](emr-483-price-acknowledgement-selected-20261021.md)とmodel-alignmentのNUMERIC非float方針へ合わせ、当該masterのdiscount_valueをfinite exact decimal文字列の専用HTTP DTO→ExactNumeric service/比較/保存→明示HTTP responseへ接続する。保存modelそのものを公開DTOにしない。旧API number/doubleとFE numberをそのままExactNumericと互換扱いせず、①でこの限定fieldのexact wire/旧client切替をAPI正本へ同期し同PRでBE/専有FEを対応する。汎用MoneyInput/他masterの全変更や旧numberへの黙ったfloat変換fallbackはしない。

rate/amountの現入力意味、個別rateの0..100、非負、期間/対象scope、POST既定値とPATCH未指定/明示0/nullの既定規則は維持する。numeric(12,2)保存尺度は変えず、既存PostgreSQL代入のscale処理をexact値で行い、新しい臨床値/割引上限/合算上限/丸め規則を作らない。桁落ちを避けるため表現不能や不正非数は入口で拒否し、未知/NULLを0補完しない。FEも入力字句と明示decimal DTOを保持し、再GETが同じ保存値を返すことをfixture化する。

②完了条件：request→service→model→response→専有FEでbinary floatを通さず、POST/更新/未指定保持/0・率と額・非数/境界/既存scale/再GETの期待値が一意、旧clientが黙って誤値を送らず、同TX監査/対象集合/参照保護を維持する。生成modelを手編集せず①EMR-527へ必要型差分を受領する。fixture作成と本人指示後の実行を分けNOT_RUN。

④は483 §8.2/8.3、481のtyped割引条件採用・計算・固定snapshot reader。②で会計再計算・過去snapshot更新・新しいsource revision列を追加しない。既存個別入力上限と合算率150%等の共通保存値を混同しない。

### C2. 物販：TaxRateのexact DTO・保存・再GETを②へ含める

同じ449集約候補。model.MerchandiseItem.TaxRateはExactNumeric/numericだが inventory/merchandise_item_request.goの*float64、service.goのCreate/UpdateMerchandiseItemInput・buildMerchandiseItemUpdate・default/モデル代入、response.goのfloat64が不一致。既存CRUD/請求・見積/campaign使用中削除保護を再利用し、型適合済みとはしない。

必要hunkは backend/internal/inventory/merchandise_item_{request,service,response}.go、同repositoryの税率bind、frontend/src/features/master/api/merchandise-items.tsのgenerated modelをOmitしたrequest/response/transform、frontend/src/features/master/components/MerchandiseSidePanel.tsx、lib/merchandise-side-panel-model.tsのtaxRate/default/表示、routes/MerchandiseItemSettings.tsxのtax_rate入力/送信。採択[405税入力](emr-405-tax-input-contract-20261020.md)・483のexact率表現に沿い、tax_rateだけをfinite exact decimal文字列の専用HTTP DTOとExactNumeric serviceへ接続する。10%の比率は0.10（数値として0.1）を保持し、百分率10へ勝手に変換しない。整数unit_priceやtax_type enumは維持する。旧number/double wireと専有FEは①API正本の限定更新と同じ切替で整合し、全master数値型移行をこの票へ混ぜない。

現入口の0..1、0、POST未指定/null時の既存0.10既定、PATCH未指定/null保持、tax_type/税率の現validation意味を維持する。非NULL保存へInvalid/NULLを通さず、明示0をdefaultへ潰さない。旧unknown/不正値を0.10で補修しない。②完了条件はexact request→service→保存→明示response→専有FE再取得の値・presenceが一致し、0/0.10/nullable/未指定/不正非数・精度保持・既存scope/監査/削除保護のfixtureがあること。numericの新scaleや新税則を作らず、試験実行NOT_RUN。会計明細への採用・税額計算・過去帳票/税snapshotは④の405/481/505へ残す。

### D. labmapping：既存設定採用と297-Aの選定済み参加を分離
集約447、実装票297-A。18DoneのList/DeviceCRUD/SaveConfiguration/ItemUpdate/field同医院検査は採用し、再実装しない。
現②既存設定の具体証拠: lab_device_item_master_service.SaveConfigurationは同TXでDevice＋Itemsを保存し、重複item ID、同医院/同source所属を検査。既存master modelはcode/unit/value_shape/optional field/sort/active。表示・編集はLabDeviceItemMasterSettingsとapi/lab-device-item-masters.ts、lab-devices.ts。
新mapping版への移行は「自由に設計」ではなく次の公開採択に固定:
- emr-297-device-mapping-bridge-contract-20261021.md §6：heads PK(clinic,source,species_scope)、immutable revisions、既存master PATCH/configuration/spec writer全参加、同head増分＋snapshot監査。generic旧有効対応のbaselineは未知由来を明示し、過去承認を捏造しない。
- §7a：lab-mapping-revision/v1のexact rows（code、field、exam_type、field版nullable/digest、unit/value_shape、元table/id、draft nullable、由来）、code順/nullable保持。既存master設定はprovenance_kind=manual_master、確定draftはdraft_confirm、初期採用はlegacy_baseline。field版を架空採番しない。
- §7b：field_definition_digestのexact既存列投影、draft contextのmapping scope revision/field definitions reader。これを受領interfaceとして固定し、旧mutable rowのupdated_atを版の代用品にしない。
変更hunk: backend/internal/medicalrecord/lab_device_item_master_service.go（Update/SaveConfigurationの共通participant）、lab_device_mapping_store.go（UpsertDeviceMaster/UpsertSpeciesSpec）、lab_device_mapping_draft.go/handler.go（297-Aのdraft confirm）、該当repository/modelと①選定DDL（stage1 §5.2の297-A設定行で445/446/466が受領）。新しいlegacy設定用C06 intentや第2request identityはここで発明しない。現route/旧DTOはこの参加のためだけには変更しない。confirm新DTO/C06は採択§4C/7の別明示操作を使う。
参加interfaceの意味: 既存認可済の設定変更TX内で、影響するsource/species scopeのheadを採択順でlock→既存master/spec保存→同TXでそのscopeの全active対応を§7a shapeにsnapshot→head増分/immutable revision/必須監査。一部itemだけの不完全snapshotをhead最新版にしない。該当scopeを取りこぼす旧writerを残したまま新readerをactivateしない。全参加が未実装の間は既存設定成果の採用と新版読取未活性を別表示する。
範囲外: 日常frames/wait、job mapping application/resolve、臨床persist/516 linkは297-B/268等④。実機・LLM実送信/臨床mapping承認は⑤/運用で別。②全masterを日常受信完了待ちにしない。
受入fixture: master PATCH/configuration/spec/confirmの各入口が同じscope revisionを進める、設定同TXrollback、generic/spec区別、field定義digest変化、旧NULL版/由来unknown、別院/失権、全writer参加前reader非活性。試験NOT_RUN。API/schema反映とruntime完了を混同しない。


## 4. 採用で閉じる3区分と重複票の扱い

動物種はhandlerのsystem_admin検査、FE isSystemAdmin、serviceのpet参照削除拒否＋監査、composition_owner_pet.goのNewAnimalSpeciesServiceWithAudit配線を採用。権限groupはhttp_permission.goの公開CRUD/rules、permission_group_service_mutate.goの同TX監査、self_lockoutの現在OR合算view/edit維持を採用。医院/法人はclinic update+reload同TX、参照削除guard、package catalog初期化、company GET/PATCHのhospital-settings grantを採用する。この3区分の具体的な新規コード欠陥は今回認定しておらず、追加実装票は作らず445採用証跡候補とする。

10:48 UTCのPlane検索では動物種は453（検索）/223（健診filter）、医院masterは100（登録確認）のみ、権限group/医院設定/法人に一致するEMR票なし。全526票の既存本文snapshotでは111 account/grant、71締めpermission、409/411 UI監査、407合成fixture、369税率回答があるが、全CRUDの同一責務ではない。名称検索だけで全履歴の重複なしを保証せず、新たな具体差分が出た場合だけ対象path/正本条項を示し既存票へのsubset追加か子票化を判断する。既存Done/UATは再実装しない。

表のMM-CLINICAL-AUX/FACILITY/WORKFORCE-AUX/FINANCE-AUX/CLOSING/LINE-CONFIGも専有pathの識別子であり、新規票作成を確定したものではない。当初認定した入院planと診断備考は、それぞれ§3A/§3Bの独立単位に限定する。今回の追加静的残差は§7で447/448/449へ紐付け、診断専有529と金銭exact専有530の範囲を広げない。その他は既存操作採用を優先し、確認だけを大量の実装票や先行試験へ変換しない。

## 5. 支払方法と後段への引渡し

465-M②は停止writer、停止状態判定の共通guard、参照中削除保護、新fund/event型との必要接続、④へ型/fixture/採用commitを引き渡すところまで。停止後新受領不可という業務仕様は維持するが、全受領caller/請求確定/返金/配賦への接続465-Cは④であり②完了gateにしない。未接続callerを対応済みと見せたり、guard未接続で新保証をactivateしたとは扱わない。

薬剤498はmaster pattern/revision保存までで臨床計算④、trimming458-Aは設定までで480/482④、package521/523はmasterまでで522/524④、診断は分類masterと495第1/第2複数診断④を分離する。医院設定変更から過去原本を更新しない。機器mappingの具体臨床値・実機・LLM実送信は別受入で、設定保存の技術設計を止める理由にはしない。

## 6. 正本と引渡しの索引

- [settings全master索引](../../spec/screens/settings/README.md)、[医院設定](../../spec/screens/19-clinic-settings.md)、[マスタ設定総覧](../../spec/screens/20-master-settings.md)
- [D03 master履歴保持](../../spec/domains/medical-records/pages/medical-record-form.md#requirement-fr-medical-records-master-history-snapshots)、[D05入院](../../spec/domains/hospitalization/lifecycle.md)、[D07検査](../../spec/domains/examinations/lifecycle.md)、[D10会計](../../spec/domains/accounting/lifecycle.md)、[D14職員勤務](../../spec/domains/staff-workforce/README.md)、[C04認可](../../spec/shared/access-control.md)
- [498 master契約](emr-498-dose-pattern-master-contract-20261021.md)、[492](emr-490-492-selected-wire-contract-20261021.md)、[458](emr-458-option-assignment-selected-contract-20261021.md)、[365](emr-365-vaccine-price-decimal-contract-20261021.md)、[金銭master](emr-449-459-465-payment-master-contract-20261021.md)
- [493休業](emr-493-staff-leave-selected-contract-20261021.md)、[eligibility](emr-493-reservation-eligibility-contract-20261021.md)、[CSV](emr-448-493-shift-csv-selected-contract-20261021.md)、[copy](emr-448-493-shift-copy-contract-20261021.md)
- [297機器mapping採択](emr-297-device-mapping-bridge-contract-20261021.md)、[483保存条件](emr-483-price-acknowledgement-selected-20261021.md)、[481 source binding](emr-481-source-binding-contract-20261021.md)

計画の網羅・採用コードの存在・実装接続・実試験の適合は別の判定。本文で未実装/未受領としている既選定subsetを完成扱いにせず、新しい業務仕様の判断待ちへも戻さない。

## 7. 既存実装の漏れ確認：限定監査追補

固定基準はmain `c86c8dd2530bbea3d1810e07badb72f3ed1a07e1`。今回の静的読解で認定した経路だけを下表へ追加する。既存の全属性・全callerの適合、実行成功、全マスタ完成の証明ではない。既存Done/UAT成果を全面再実装せず、専有実装ownerは今後の単一claimで決める。新規票は作成していない。527–530はReadyの管理票で、PR803の連携コメントによるReview誤遷移は実装開始根拠ではない。他票のstate/assigneeを変更しない。

以下のFE pathは `frontend/src/features/master/routes/` 相対。更新時の任意文字列は、省略=保持と明示空文字=空を保存を区別する。空をundefinedへ変換するCreate builderをUpdateへそのまま流用すると、BEの既存*string nil=変更なしへ落ちる。修正対象は下記fieldのUpdate builderと既存DTOの接続であり、Create既定/必須name/null全般/認可/履歴/会計計算を一括変更しない。

| 集約・責務 | 確認したFE経路 | 受入へ追加する限定差分 |
|---|---|---|
| 447 主訴・診療5tab | chief-complaint-settings-model.ts L17–23、treatment-plan-master-model.ts L105–211 | descriptionの既存非空→明示空をPATCHに含めて再GETでも空。診察/検査/処置/予防接種/健診の5builderをそれぞれ対応。既存BE各serviceの*string nil保持を維持 |
| 448 ケージ・職種 | cage-settings-model.ts L4–16、occupation-settings-model.ts L4–13 | descriptionの空clearをUpdateで明示送信。既存cage_service.go/occupation_service.goのnil保持・非nil空保存を再利用 |
| 448 トリミングcourse/option | trimming-settings-model.ts L30–88 | 両descriptionの空clearだけ追加。courseのshort_name/allowed_weekdays/image_object_keyは変更フラグ付き既存clearを維持し、全面置換しない |
| 448 staff公開profile | staff-settings-model.ts L109–121 | license_number / reservation_display_name / reservation_comment / reservation_image_urlの4fieldだけ空clear。backend/internal/staff/staff_request.go L63–89→staff_service_builders.go L42–67の*string非nil更新へ接続。password/emailの空clearや認証仕様へ一般化しない |
| 449 保険 | insurance-settings-model.ts L23–34 | description/contact_phoneの空clear。insurance_service.goの各*string nil保持を再利用し、coverage_rateの範囲/既定を変えない |

各行の固定fixtureは非空→空、空→非空、省略保持、明示0/bool等の無関係値保持、保存後再GET、認可失敗時入力保持とする（NOT_RUN）。診断Type/Nameは既知EMR-529専有のまま。上表を529や530へ便宜的に集約しない。

薬剤と医院のadapterは、本人がBE後続へ延期した段階移行の残として扱う。新規regression・担当忘れ・無修正完了とは断定しない。責務は次の既存行へ紐付ける。これは個人担当の割当や実装開始指示ではなく、将来単一claimする範囲の明確化である。

| 既存集約・受入先 | 固定mainで確認した境界 | 限定受入条件・非対象 |
|---|---|---|
| EMR-447 薬剤base master | backend/internal/medicalrecord/medicine_service_create.go L25/L28/L31のDefaultQuantity/TaxRate/Strength→ExactNumeric境界。既存447本文の薬剤create/edit/read往復へ受領 | request/service/base model/HTTP response/callerの型対応を既存選定wireに合わせ、presence/null/0・保存後再GETを対応づける。保存modelを旧型へ戻さない。498 pattern/revision・488製剤量・530 campaign/merchandiseへ混ぜず、新しい臨床値を選定しない |
| EMR-445 医院の既存採用行 | backend/internal/clinic/clinic_response.go L26–27/L66–67のStandardTaxRate/ReducedTaxRate response境界 | modelのExactNumericとHTTP response/callerの採択型を対応づけ、既存医院設定保存/再GETの受入へ追加。下記MM-CLINIC文字列clearとは別hunk。新税則・実医院設定変更・530への追加をしない |

この監査追補の管理先は445/447/448/449/450/523/528の7票。既知529の診断と530のcampaign/merchandiseは専有範囲を維持し、527生成同期も別後続。4票のReady是正は実装開始根拠にしない。実装・ビルド・試験・生成・DB適用は未実施。

### MM-CLINIC：②医院設定専有subset（445、単一claim待ち）

固定main `06213d5778807f9ecbc9a8501dd579e8190a0d3b`。frontend/src/features/clinic-settings/lib/clinic-master-settings-model.tsのbuildUpdateClinicRequestはpostal_code/address/phone_number/fax_number/registration_number/director_name/email/websiteの空文字をundefinedにし、backend/internal/clinic/clinic_service_update.goの*string nil=省略保持へ落ちる。既存非空値を明示空へ変更できない具体差分として、Updateの8fieldだけpresenceを保持して送信し、既存非nil空保存→再GETへ接続する受入を追加する。Create既定・必須name・税率adapter・実医院設定変更は別。受入条件は各非空→空、空→非空、省略保持、別field保持、現在認可拒否で入力保持（NOT_RUN）。既存445で②医院設定専有subset/単一claim待ちとして追跡し、Aの共有API/DDL/model担当へFE実装を暗黙追加しない。

## 8. 492既存設定writer切替の具体受領条件（後続）

frontend/src/features/master/hooks/reservation-type-settings-model.tsのbuildReservationTypeUpdateRequestはCreate builderを流用し、未分類への変更でgroup_idを省略しclear_group_idも送らない。またdescription/reservation_display_name/short_name/reservation_comment/reservation_image_urlの空がundefinedへ落ちる。この静的差分は[既選定492契約§5.2](emr-490-492-selected-wire-contract-20261021.md)のgroup明示nullと属性presenceへ切り替える受入例であり、新規業務判断ではない。

新C06 changesではgroup_idの省略=保持、明示null=解除、値=設定とし、旧clear_group_idと新nullを二重の業務writerにしない。互換adapterが必要なら旧clear flagをpresence付きnullへ一回だけ変換する。任意5文字列の明示空/省略/既存nullable規約を区別し、表示名のrenameやscope/CAS/expected_group_versionsを迂回しない。fixtureはgroup値→未分類→再GET、group省略保持、各文字列非空→空/省略保持、旧flag/新wire混在拒否、他院group拒否（NOT_RUN）。492の後続単一ownerが設定writer切替とともに受領し、今回Dの限定clear作業へ追加しない。

## 9. 今回5枠との境界

[①のA〜E引渡し表](emr-445-stage1-contract-start-plan-20261021.md#stage-current-five-20261021)に従う。C/D/Eの限定hunk以外の既選定②subset、MM-CLINIC、③450/523を同時着手の指示にしない。297-Aは①保存/context受領→447/297の②全設定writer参加の対応を閉じ、daily job application/wait/resolve/persistは④へ残す。計画補正と実装完了を分け、今回はdocs-only/NOT_RUN。

## 10. coverage監査で具体化した②受領残（2026-10-22）

[監査記録A01–A10](emr-stage123-coverage-audit-20261022.md)は旧監査基準と作成基準を分け、計画記載・所有範囲・実装接続・試験を別判定する。今回の所見を進行中C/D/Eや530へ無断追加せず、既存票の後続subsetとして受領する。

| 対象 | ②の具体受領条件 | 責務・設計残/非対象 |
|---|---|---|
| 診療5種＋薬剤＋物販の全体割引除外 | 各設定画面の編集→保存→再GET、明示値と未設定/unknownの保持を①保存契約から受領。旧会計/履歴再計算・false補完なし | 臨床/薬剤447、物販449。①保存型/wireは2026-10-22選定契約Aで技術選定済み、API/DDL/model/製品実装・受領・試験は未完。481/483/505は④採用consumer。現C/E/530へ黙って追加しない |
| 担当者別LINE開始時刻 | 分類/担当者ごとの開始時刻設定と既存予約保持を受領、保存後の候補表示/判定consumerは468へ渡す | 448/492後続単一writer。[2026-10-22時間契約§4](emr-clinic-line-time-selected-contract-20261022.md#4-a02-担当者別trimming開始policy)で保存scope/版/元assignment/日別置換/削除再設定を選定済み。API/保存型/製品実装・受領・試験は未完。現医院＋区分の旧slotで完成とせず、マーケ最終工程とも別 |
| 医院共通営業時間/休診日 | 共通医院基準とLINEの範囲内別設定、院内の警告了承/取消、レジ別日界の既決を保持 | 445医院設定と448/492 LINEの後続。[時間契約§2/3](emr-clinic-line-time-selected-contract-20261022.md)でcommon aggregateとLINE確認版/内側制限、旧レジ休日分離を技術選定済み。156の実値照合・API/保存型/実装受領は未完。実医院設定/現5枠を変更しない |
| 健診種別parent整合 | Create/親変更/Deleteが同じ親row lockと同TX内の存在・子参照再検査に参加し、子保存先行/親削除先行の両順序で片方だけ成立。CAS/品目保険監査は保持 | 447/456 UAT・#732の残存コード差分。Read Committedの通常SELECT→別TX soft-delete→子INSERTの静的raceで、実DB再現NOT_RUN。新DDL必須にせず、Cへ無断追加しない |
| trimming course旧属性の指定なし | 2026-10-22選定契約BでPATCH-only clear_course_type_id/clear_target_sizeを技術選定済み。旧null保持維持、flag true＋非null値/flag null等は400/全体無書込。値→指定なし→GETと無関係項目保持をfixture化。API/製品実装・受領・試験は未完、新DDL/model不要 | 448後続。457新属性/画像/曜日clearとは別、FKへ0/空文字を送らず汎用nullable全面拡張なし。#806でもcourse_type_id schema未収録は残る |
| trimming option Create型adapter | SortOrder int→model *intの具体caller対応を既知型driftの受領表へ追加 | 448の既知型境界確認。新規欠陥件数に二重算入せず、modelを旧型へ戻さない |

EMR-306はDone直前のcomment fd0b67d6-3a21-4d15-821b-d2470f0b243d（2026-10-20 07:51:30.990830Z）の限定STG UAT再実行PASSを採用する。空き2件選択→代表ケージ確認→連結→マスタ/ボード両表示→解除→個別空き復帰は既存成果。ボードDnD合体・連結中割当拒否実演・退院後の連結維持だけUNKNOWN/461・499境界として分離し、証拠欠落・記録不整合・誤Doneにはしない。状態は変更しない。ホテル上限はspec-storage-traceと既存452/461/499に追跡でき、新規漏れではない。今回はdocsのみ、実DB/試験/ビルド/生成NOT_RUN。

### A01/A06の現在判定と後続受領

[2026-10-22の選定契約](emr-master-update-selected-contract-20261022.md)がA01/A06の監査時設計残を具体化した。A01は7resourceの新boolを編集→保存→再GETし、rawNULL/false/trueと旧欠落unknownを区別する。無関係更新/旧writer/importで既存trueやrawNULLを消さず、必要CASと既存同TX auditへ新boolのbefore/afterを保険とは別に追加する。既存is_non_insurance helperだけで監査対応済みにしない。旧会計/治療/見積/帳票/snapshot更新なし。

A06はFE Update builderから対象2propertyを除いて変更差分だけを送る。未編集は省略、指定なし変更は対応flag trueだけ、値指定は値だけ。既存short_name/image_object_keyの空文字・allowed_weekdaysの空配列clearと省略/null保持は変えない。専有FE再GETで確認し、BE対応前に新FEを公開しない。現Dのdescription/既存型driftとは別後続として448が受領する。

選定済みは実装済みではなく、API正本反映・A01の必要schema/model/reader整合・runtime→FE配備・受領・試験は未完。A06の新DDL/modelは不要で、その非対象判定を適用済みとは言わない。A02/A03も下記の技術選定済みだが実反映/受領は未完で、全①〜③完成とはしない。NOT_RUN。

### A02/A03の現在判定と専有後続

[医院・LINE時間契約](emr-clinic-line-time-selected-contract-20261022.md)は監査時設計残を2026-10-22に具体化した。445はcommon initialize/updateと権限限定休診adapter、448/492はLINE acceptance確認/更新と非P分離、staff開始policyの設定UI/単一writerを後続claimで受領する。共通縮小はLINE保存値をmaskするだけで書戻さず、拡大も自動拡張しない。未設定/未照合/正常空/取得失敗を分け、LINE受付可へfallbackしない。

既存予約警告のpreview/了承/取消/CAS再検査を使い、予約全解消や院内全面拒否を新設しない。旧shifts休日はcommon切替時に移行するが、旧closing休日はレジ専用legacy read/writeを維持する。新commonとdual-writeなし。ホテル閉店ちょうどの入退室予定は本人承認済みで、休憩/休日/LINE/定員の独立条件は維持する。④468の利用と実予約切替は別成果、現D等のscopeへ追加しない。実医院値確認・API/DDL/model/生成・runtime/FE・全必要writer閉包・受領・試験は未完。NOT_RUN。

# EMR-445 ②全マスタの着手プランと採用証跡

2026-10-21。対象はsettings索引21本文と、その索引外の医院/法人・LINE予約設定・検査機器対応・package。医院と法人は別resourceとして行を分けたため下表は26行だが、26件の新規実装票・PRではない。

基準main `a88fa4add58d43cff643590f9f507c8ab1a3c085`（10:45 UTC再取得）。業務正本は[settings索引](../../spec/screens/settings/README.md)と各領域の確定FR/AC。[①②③段階台帳](emr-445-master-stages-ledger-20261021.md)の直接22票＋498は管理対象であって全属性実装完了の証明ではない。現コードは差分確認の証拠であり、正本を縮小しない。実装・試験・DB操作は今回行っておらずNOT_RUN。

## 1. 責務・開始条件・完了の意味

- 447/448/449は集約・引渡し票。実装claimは既存実装票の限定subset、または下記の具体差分単位に固定する。450は③共通UIであり、全CRUDや認可モデルの代替ではない。
- ①API/schema/model/generatedが必要な行は、その必要subsetの収録/採用commitを単一統合ownerから受領する。api.yamlを複数claimで同時に編集しない。生成型同期は独立後続で担当未割当、①生成型後続の単一claim ownerへ受領する。旧session3表記は履歴で、現claimは着手時に確認する。
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
### 3. 入院・宿泊プラン
- 属性/操作：name/body_size/billing_unit/price/tax/description/active。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：448集約 / MM-HOSPITAL-OPTIONAL / §3Aの専有差分。
- BE：B/medicalrecord/hospitalization_plan_{handler,service}.go。model：B/model/hospitalization_plan.go。FE：F/master/routes/HospitalizationSettings.tsx。
- 判定・境界：既存CRUD/税nullable model/ケア参照削除guard採用。備考空・体格/料金単位null・明示0送信の差分は§3Aの固定契約で修正する。
### 4. トリミングコース/option
- 属性/操作：名称/size価格/所要時間/任意略称・説明・画像/曜日/順序/休止。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：448 / 457成果＋458-A / trimming_course/option/T2。
- BE：B/trimming/trimming_course_{handler,service}.go。model：B/model/trimming.go。FE：F/master/routes/TrimmingSettings.tsx。
- 判定・境界：既採用：曜日/画像属性正規化、画像lifecycle監査、参照削除保護。458 T2追加APIは①未収録として明確な残。457 UAT再実装なし。
### 5. トリミングコース分類
- 属性/操作：名称/active/order。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：448 / 445採用証跡（MM-FACILITY） / trimming_course_type系のみ（457と非重複）。
- BE：B/trimming/trimming_course_type_{handler,service}.go。model：B/model/trimming_course_type.go。FE：F/master/routes/TrimmingCourseTypeSettings.tsx。
- 判定・境界：既採用：CRUD/reorder、WithTx内参照countと削除。分類をreservation business_kindへ変える追加実装なし。
### 6. 検査種別・測定field
- 属性/操作：本体/親/価格/保険・field名称/単位/原値/基準値/順序。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：447 / 445採用証跡（MM-CLINICAL-AUX） / exam_type/field。既存18/169/225再利用。
- BE：B/medicalrecord/exam_type_{handler,service}.go。model：B/model/examination_type.go。FE：F/master/routes/TreatmentPlanMaster.tsx。
- 判定・境界：既採用：exam_type CRUD/parent医院検査/参照guardとExaminationType.Items。field操作はexam_type_field.goへ。未承認の基準値投入は技術closureとは別。
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
### 9. キャンペーン
- 属性/操作：name/期間/rate又はamount/対象カテゴリ・商品/active。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：449集約 / 445採用証跡 / 新規実装票なし。
- BE：B/billing/campaign_{handler,service}.go。model：B/model/campaign.go。FE：F/master/routes/CampaignSettings.tsx。
- 判定・境界：期間/率額/商品医院検査/typed対象CRUD採用。master writerはCampaign/Targetだけを更新し会計snapshotを書かないため②変更なし。§3Cの固定根拠を採用。
### 10. 物販
- 属性/操作：name/category/price/tax/active/order。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：449 / 445採用証跡（MM-FINANCE-AUX） / merchandise_item系。
- BE：B/inventory/merchandise_item_{handler,service}.go。model：B/model/merchandise_item.go。FE：F/master/routes/MerchandiseItemSettings.tsx。
- 判定・境界：既採用：CRUDと同TX削除、請求/見積/campaign参照guard。数値入力roundtripは既存型のまま仕様適合を確認し、全masterを無断小数化しない。
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
### 13. 職種
- 属性/操作：name/description/active/order。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：448 / 445採用証跡（MM-WORKFORCE-AUX） / occupation系、staff business変更除外。
- BE：B/staff/occupation_{handler,service}.go。model：B/model/occupation.go。FE：F/master/routes/OccupationSettings.tsx。
- 判定・境界：既採用：CRUD/順序、同TX active lock→staff usage→delete。役職名による予約grant復活なし。
### 14. 勤務template
- 属性/操作：name/type/開始終了/複数休憩/notes/active/order。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：448 / 445採用証跡（MM-WORKFORCE-AUX） / shift_template系、493 CSV/copy除外。
- BE：B/staff/shift_template_{handler,service}.go。model：B/model/staff.go。FE：F/shifts/routes/ShiftTemplateSettings.tsx。
- 判定・境界：既採用：staff.go内ShiftTemplate/Break、CRUD・breaks更新・正規化、同TX削除。コピー/CSVは493別操作、template CRUD再作成なし。
### 15. 予約区分/group
- 属性/操作：type/group/所要時間/個別意図/公開/分類/name版/順序。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：448 / 492 / reservation_type/group専有。
- BE：B/reservation/reservation_type_{handler,service}.go。model：B/model/reservation_type.go。FE：F/master/routes/ReservationTypeSettings.tsx。
- 判定・境界：既採用：基本handler/service/model/FE。492 rename/restore/C06追加wireと旧writer切替は明確な残。既存CRUDを削除して全面再構築しない。
### 16. 診療項目5tab
- 属性/操作：診察/検査/処置/ワクチン/健診の名称/親/description/price/active/対象保険税。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：447 / 454455456既採用＋365 / 5種専有／shared454単一owner。
- BE：B/medicalrecord/consultation_{handler,service}.go。model：B/model/consultation.go。FE：F/master/routes/TreatmentPlanMaster.tsx。
- 判定・境界：既採用：5種個別serviceと共通FE、診察/処置454・ワクチン455・健診456成果。365 vaccine price文字列/numeric23,4は明確な①残。検査fieldとの差分は臨床補助unit。
### 17. 診断カテゴリ/病名
- 属性/操作：カテゴリ・病名のname/description/active/order/必須category。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：447集約 / MM-DIAGNOSIS-DESCRIPTION / §3Bの専有差分。
- BE：B/medicalrecord/diagnosis_{handler,service}.go。model：B/model/diagnosis.go。FE：F/master/routes/DiagnosisSettings.tsx。
- 判定・境界：カテゴリ必須NOT NULL/同医院存在検査/参照削除guardは採用済み。備考clearだけ§3Bの限定差分。第1/第2複数診断は495④。
### 18. 薬剤・用量pattern
- 属性/操作：name/親/剤形/単位/price/tax/保険/active・ID別pattern/revision。個別正本にあるCRUDと専用操作のみ。
- 集約 / 実装単位 / 専有hunk：447 / 498 master subset＋基本CRUD採用記録 / medicine/dose専有。
- BE：B/medicalrecord/medicine_{handler,service}.go。model：B/model/medicine.go。FE：F/master/routes/MedicineSettings.tsx。
- 判定・境界：既採用：基本CRUD、親/在庫医院検査、使用中削除guard。498 ID別pattern/revision API未収録が明確な残。既存revision modelは再作成しない。
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

### A. 入院プラン：具体差分あり、MM-HOSPITAL-OPTIONAL の最小1PR案
集約448、専有実装単位MM-HOSPITAL-OPTIONAL（新規子票案、未起票）。トリミング分類など無差分対象を同票へ増やさない。
根拠：master-hospitalization-plan.md「詳細編集サイドパネル」「API連携」、D05 master変更は過去/追加済会計へ自動反映しない。BE hospitalization_plan_request.goはUpdate Description/BodySize/BillingUnitを*string、service buildHospitalizationPlanUpdateはnilなら更新なし。FE routes/hospitalization-settings-model.tsは備考空をundefined、体格/料金単位空をnullとして送る。従って備考の空への変更と明示nullのクリアが静的に保存されない。
選定する最小契約:
- 既存PATCH /api/v1/masters/hospitalization-plans/{id}を維持。body_size/billing_unitは省略=保持、null=SQL NULL、許容enum値=設定。空文字をNULLの代用品にしない。
- BE request専有のpresence-aware field（Set/Null/Valueの3値、既存同形helperがあれば再利用）でJSONをdecodeし、service inputに同じpresenceを渡す。build update mapは省略をkeyなし、nullをmap値nil、値をenumにする。DB nullable列/modelは既存のまま。
- 備考はFE Updateでdescriptionを空文字もそのまま送る。BE既存*stringは空文字を保存できるのでこの欄のschema変更不要。
- 新規作成のFE price:number=0をprice || undefinedで落とさず、明示入力0を0として送る。任意値全般のnull→0変換拡張や旧データbackfillは本票でしない。既存画面の数値初期値0/税初期値は変えない。
- response/readのbody_size/billing_unitは既存nullableの意味を維持。DB NULLが現omitemptyで省略される既存responseはFEの??nullで読む。writeのnullと省略の差をread省略の問題へ混同しない。
変更hunk: frontend/src/features/master/routes/hospitalization-settings-model.ts、backend/internal/medicalrecord/hospitalization_plan_request.go/service.go、専有DTO/fixture。①api.yaml既存PATCH propertyのnullable/省略説明は単一ownerが同期。model/DDL新設なし。
履歴境界: 現repositoryはhospitalization_plansだけ更新、Deleteはcare_plan_items参照NOT EXISTS＋soft delete。これを採用し、snapshot新表/過去care更新をこの票へ足さない。405/481の④会計bindingをgateにしない。
受入fixture: 備考非空→空、体格/料金単位値→null、各省略→既存値保持、値→別値、不正enum zero-write、0価格作成、同院/他院、参照中削除拒否、master更新でcare/会計row無変更。NOT_RUN。

### B. 診断：カテゴリ契約は既決、備考clearだけ具体差分
集約447、専有実装単位MM-DIAGNOSIS-DESCRIPTION（新規子票案、未起票）。診断/主訴/template一式の再開発票に広げない。
確定: master-diagnosis.md「所属カテゴリ必須選択」。model.DiagnosisName.DiagnosisTypeIDはNOT NULL uint64、create request required、service Create/Updateが同医院カテゴリFindByIDで検査。分類nullableは未選定ではなく不要。PATCH省略は保持、明示0/不存在/他院は既存validation拒否。カテゴリそのものの設計変更なし。
具体差分: frontend/src/features/master/routes/diagnosis-settings-model.tsのType/Name UpdateがCreate builderを再利用し、description空をundefinedへ落とす。BE UpdateDescription *stringとbuild mapは空文字を保存可能。
修正: TypeとNameのUpdate builderだけdescription:data.descriptionを明示する。新規Create既定は変更不要。API/schema/DDL/model/BE変更不要。既存api/diagnosis.tsを経由する両更新callerを維持。
履歴: カテゴリ/病名master CRUDは既存diagnosis_service/repositoryを採用。カテゴリ子参照拒否、病名のカルテ参照拒否を保持。第1/第2複数診断snapshotは495④へ引渡し、本票でカルテ診断を再保存しない。
受入fixture: Type/Nameとも既存備考→空、空→非空、非空維持、有効切替/所属変更が無関係に消えない、カテゴリ必須と他院拒否。実試験NOT_RUN。

### C. campaign：②変更なしの採用行として閉じる
集約449、MM-FINANCE-AUXのcampaign行は現時点新規実装票不要。
master-campaigns.mdとD10の既決に対し、campaign_service.goには期間/率額/対象集合/商品医院検査とCRUD、repositoryにはCampaign本体/TargetCategory/TargetItemの更新だけがある。Update L280以降→ReplaceTargets、Delete L350以降→repository Delete L131はsoft deleteであり、billing/calculation snapshotを更新する呼出はない。よって「master更新で既存snapshotを更新しない」という②writer境界はこの採用根拠で閉じる。現在会計readerが過去snapshotを正しく表示することまで証明したとはしない。
④受領先は既採択 emr-483-price-acknowledgement-selected-20261021.md §8.2 DiscountTerms/DiscountResult、§8.3 snapshot.conditions/line.source_snapshot、および481 source binding。masterの率/額・対象定義を変更する新interfaceやsource revision列を②へ足さず、会計ownerが採用時点のtyped割引条件を保存する。個別campaign入力上限と合算率150%等の計算保存型を混同しない。
変更hunkなし。採用証跡へ service/repository/model/FE SHAを記録。会計再計算・reader修正は④既存owner、新規master票のgateでない。

### D. labmapping：既存設定採用と297-Aの選定済み参加を分離
集約447、実装票297-A。18DoneのList/DeviceCRUD/SaveConfiguration/ItemUpdate/field同医院検査は採用し、再実装しない。
現②既存設定の具体証拠: lab_device_item_master_service.SaveConfigurationは同TXでDevice＋Itemsを保存し、重複item ID、同医院/同source所属を検査。既存master modelはcode/unit/value_shape/optional field/sort/active。表示・編集はLabDeviceItemMasterSettingsとapi/lab-device-item-masters.ts、lab-devices.ts。
新mapping版への移行は「自由に設計」ではなく次の公開採択に固定:
- emr-297-device-mapping-bridge-contract-20261021.md §6：heads PK(clinic,source,species_scope)、immutable revisions、既存master PATCH/configuration/spec writer全参加、同head増分＋snapshot監査。generic旧有効対応のbaselineは未知由来を明示し、過去承認を捏造しない。
- §7a：lab-mapping-revision/v1のexact rows（code、field、exam_type、field版nullable/digest、unit/value_shape、元table/id、draft nullable、由来）、code順/nullable保持。既存master設定はprovenance_kind=manual_master、確定draftはdraft_confirm、初期採用はlegacy_baseline。field版を架空採番しない。
- §7b：field_definition_digestのexact既存列投影、draft contextのmapping scope revision/field definitions reader。これを受領interfaceとして固定し、旧mutable rowのupdated_atを版の代用品にしない。
変更hunk: backend/internal/medicalrecord/lab_device_item_master_service.go（Update/SaveConfigurationの共通participant）、lab_device_mapping_store.go（UpsertDeviceMaster/UpsertSpeciesSpec）、lab_device_mapping_draft.go/handler.go（297-Aのdraft confirm）、該当repository/modelと①選定DDL。新しいlegacy設定用C06 intentや第2request identityはここで発明しない。現route/旧DTOはこの参加のためだけには変更しない。confirm新DTO/C06は採択§4C/7の別明示操作を使う。
参加interfaceの意味: 既存認可済の設定変更TX内で、影響するsource/species scopeのheadを採択順でlock→既存master/spec保存→同TXでそのscopeの全active対応を§7a shapeにsnapshot→head増分/immutable revision/必須監査。一部itemだけの不完全snapshotをhead最新版にしない。該当scopeを取りこぼす旧writerを残したまま新readerをactivateしない。全参加が未実装の間は既存設定成果の採用と新版読取未活性を別表示する。
範囲外: 日常frames/wait、job mapping application/resolve、臨床persist/516 linkは297-B/268等④。実機・LLM実送信/臨床mapping承認は⑤/運用で別。②全masterを日常受信完了待ちにしない。
受入fixture: master PATCH/configuration/spec/confirmの各入口が同じscope revisionを進める、設定同TXrollback、generic/spec区別、field定義digest変化、旧NULL版/由来unknown、別院/失権、全writer参加前reader非活性。試験NOT_RUN。API/schema反映とruntime完了を混同しない。


## 4. 採用で閉じる3区分と重複票の扱い

動物種はhandlerのsystem_admin検査、FE isSystemAdmin、serviceのpet参照削除拒否＋監査、composition_owner_pet.goのNewAnimalSpeciesServiceWithAudit配線を採用。権限groupはhttp_permission.goの公開CRUD/rules、permission_group_service_mutate.goの同TX監査、self_lockoutの現在OR合算view/edit維持を採用。医院/法人はclinic update+reload同TX、参照削除guard、package catalog初期化、company GET/PATCHのhospital-settings grantを採用する。この3区分の具体的な新規コード欠陥は今回認定しておらず、追加実装票は作らず445採用証跡候補とする。

10:48 UTCのPlane検索では動物種は453（検索）/223（健診filter）、医院masterは100（登録確認）のみ、権限group/医院設定/法人に一致するEMR票なし。全526票の既存本文snapshotでは111 account/grant、71締めpermission、409/411 UI監査、407合成fixture、369税率回答があるが、全CRUDの同一責務ではない。名称検索だけで全履歴の重複なしを保証せず、新たな具体差分が出た場合だけ対象path/正本条項を示し既存票へのsubset追加か子票化を判断する。既存Done/UATは再実装しない。

表のMM-CLINICAL-AUX/FACILITY/WORKFORCE-AUX/FINANCE-AUX/CLOSING/LINE-CONFIGも専有pathの識別子であり、新規票作成を確定したものではない。具体差分を認定した入院planと診断備考を、それぞれ§3A/§3Bの独立単位に限定する。その他は既存操作採用を優先し、確認だけを大量の実装票や先行試験へ変換しない。

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

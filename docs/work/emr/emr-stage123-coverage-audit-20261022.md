# ①〜③の仕様・チケットcoverage監査記録（2026-10-22）

計画の記載は揃っていても、設定writerの技術設計・実装接続・過去成果の受領がすべて閉じた状態ではない。本記録は、その差を既存ownerへ引き渡すための限定静的監査結果である。全チケット/全ACの実行保証、MECE性の完全証明、①〜③全完了の宣言ではない。製品コード・API YAML・DDL/model・生成物・設定・実行環境は本作業では未変更。make codegen/build/test/Docker/隔離検証/DB操作はすべてNOT_RUN。

要件窓口はMinoru Soga。既存確定仕様に準拠して抜け・重複・受領漏れを整理する本人承認に基づき、新業務要求・KPI・担当再割当を追加しない。削減対象は同じ設定の二重設計/実装と繰返し調査、部分Doneの全体完成への誤換算。計画指標は各所見に正本/根拠/受領先/設計残/現5枠との境界があること。業務効果・実行品質は未測定。

## 1. 基準と読取範囲

| 区分 | 固定値・意味 |
|---|---|
| 監査baseline | main `4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34`（#805採用）。以下の原所見の固定根拠 |
| docs PR作成baseline | main `9eef73feac19c25571724372b0b19da160ba5cfb`。#806/#807が2026-10-22 01:02Zに採用された後、01:05:51Z snapshotの最新main。原所見のbaselineと同一とは扱わない |
| 実装側open PR | #808＝527 caller整理Draft、#809＝528 clear/null/0/type。開始時再読でopen、今回docs集約に取り込まず、未採用を解消済みとしない |
| この記録の方法 | rootの読取専用母集団監査・過去採用資産照合を受領し、編集担当が正本と所見の具体path、#806/#807変更ファイル、代表コード境界を静的再照合。994/1,019コメント全件や全661PRを本編集担当が再コード監査した記録ではない |

初回active530、Review70＋UAT72＝142、Done237。最終固定snapshot **2026-10-22 01:05:51.939 UTC** はReview75＋UAT72＝147票、取得コメント1,019件。増加5票は449/527/528/529/530の実装中票で、完成件数ではない。初回Review/UATコメント994件の母集団調査からの更新として区別する。

Done237票は一次分類で、コメントまで読んだのは14票、残223票は同じ深度ではない。PR全661件の関連参照を棚卸ししたが、全661PRのコードレビューではない。①②旧Done10票は採用資産と現計画受領先を全件特定。③旧残43票は共有成果12・実装成果18・調査/親等13で、40票は所在/責任trace、追加の限定照合は245、318受付write、400/#702 staff write。別枠の追加Done10票も分類し、422 MoneyInput subsetは財務側の受領確認残へ分離した。これらを計画の行数や実装完了件数へ合算しない。

詳細コード照合は①②主要6票＋9境界、③代表6票と分類追加、最近業務18票24PR等の静的範囲。調査票/親票を独立実装数に数えない。①24管理行、②26resource、③72用途（検索対象57＋非keyword/除外15）、非検索D01–D23/P01、7primitive＋18storiesは異なる母集団。72には未来S13と複数入口H03があり、72実装済みではない。18storiesは既存C01契約に分類可能で、18件の実装漏れではない。

## 2. #806/#807採用後の再照合

| 採用PR | 変更ファイル単位の確認 | 今回所見への影響 |
|---|---|---|
| [#806](https://github.com/MinoruSoga/AnimalEkarte/pull/806) | backend/docs/api.yaml、backend/internal/apicontract/openapi_route_drift_test.go | API正本の限定収録。健診親lock、trimming parser/service/FE、治療Enter、SubmitButton、C01患者文脈は変更なし。新保存writerの実装完成ではない。TrimmingCourse.target_size nullableはあるがcourse_type_id schema未収録は残る |
| [#807](https://github.com/MinoruSoga/AnimalEkarte/pull/807) | lab_device_mapping_draft.go・lab_device_mapping_revision.go・schema_drift_test.go、006_operation_anchor_and_lab_mapping_versions.sql、関連mapping/休業anchor ERD・manifest | 採択済みanchor/mapping保存subsetの受領。今回7modelの全体割引除外、担当者別slot、医院共通時間、健診parent race、trimming解除、UI caller各pathは変更なし。006を当該raceの修正/全DDL完成に数えない |

変更一覧は [base→作成基準のAPI](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/backend/docs/api.yaml) と [#807 Files](https://github.com/MinoruSoga/AnimalEkarte/pull/807/files) を参照。model/DDLが採用済みでもDB適用・B最終生成・Q/ContextGate/各writer接続は別。#809で528修正案が存在することとmainへ採用済みであることも分ける。

## 3. 限定所見と既存ownerへの追補

| ID/分類 | 根拠URLと確認した差 | 必要な追補/未確定技術設計 | 既存owner・現5sessionとの境界 |
|---|---|---|---|
| A01 設定writerの設計残 | [会計第六群](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/docs/spec/domains/accounting/lifecycle.md#L88)、[診療品目設定](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/docs/spec/screens/settings/master-treatment.md#L11)、[薬剤](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/docs/spec/screens/settings/master-medicine.md#L13)、[物販](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/docs/spec/screens/settings/master-merchandise.md#L13)、[481未設定/unknown](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/docs/work/emr/emr-481-source-binding-contract-20261021.md#L41)。診療5種＋薬剤＋物販の7model/3画面系統に全体割引除外の設定writerを①②で割り当てられていない | 共有保存型/wireと②編集→保存→GETを受領する**具体保存契約の設計残**。新列/table/API/保存方式を本稿で決めない。正しい元masterの成功読取で未設定確認時だけfalse、未接続/失敗/元不明はunknown。旧履歴へfalse補完/再計算なし | 臨床/薬剤447、物販449。481/483/505は④採用consumerでwriter代用ではない。現C/E・530へ黙って追加しない |
| A02 保存scopeの設計残 | [担当者別開始時刻FR](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/docs/spec/line/reservation-spec.md#L58)、[slot request](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/backend/internal/reservation/reservation_type_request.go#L133)、[保存model](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/backend/internal/model/reservation_type.go#L95)、[scope query](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/backend/internal/reservation/reservation_type_available_slot_repository.go#L31)。現保存は医院＋区分で担当者scopeなし | 単一owner・保存scope/writer/wire/CASの技術設計残。[492既選定](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/docs/work/emr/emr-490-492-selected-wire-contract-20261021.md#L160)は開始枠を拡張しないため、本件収録済みとはしない。既存予約を動かさない | 448/492①②設定writer、468④判定consumer。Lstep/DM最後のマーケ整備とは別、現Dへ追加しない |
| A03 既知保存責務残 | [D17共通時間の採択](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/docs/spec/domains/clinic-settings/README.md#L88)、[EMR-156](https://app.plane.so/baritechllc/browse/EMR-156/) comment fbddf36e-4b2b-49db-9bb7-69c9f2679cd2（2026-10-20 07:37Z、FABLE-QA-20261020-156-345） | clinic_holidaysとLINE曜日/個別日/時間の対応、共通writerの特定・保存/競合/移行が既知残。医院共通基準とLINEの範囲内別設定は確定で業務再質問しない。既存実医院設定を変更しない | 445医院設定と448/LINE境界へ受領。468consumer/492だけでは閉じない。今回A/Dの実装範囲を拡大しない |
| A04 所有重複/参照誤り | [②row6/16](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/docs/work/emr/emr-445-stage2-master-start-plan-20261021.md#L50)、[健診field model](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/backend/internal/model/checkup_type.go) | 検査プラン本体CRUDのprimaryをrow6へ固定しrow16は参照。225はcheckup_type_fieldsでexam_type_fieldsではない。18機器mapping、169健診packageの採用参照へ訂正 | 447が受領、checkups/master-medical/lab-importのgrantを相互代用せず、再実装・重複起票なし。5枠再割当なし |
| A05 静的race差分 | [子Create/親検証](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/backend/internal/medicalrecord/checkup_type_service.go#L169)、[Delete](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/backend/internal/medicalrecord/checkup_type_service.go#L255)、[通常Find/条件付soft-delete](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/backend/internal/medicalrecord/checkup_type_repository.go#L46) | Read Committed通常SELECT後に別TX親soft-deleteがcommitし、その後子INSERTがFKを通り得る。serializable指定/親soft-delete防止triggerなしという静的条件で、実DB再現なし。Create/親変更/Delete共通parent lockと両順序の一方成立、CAS/保険監査維持を受入へ。新DDL必須にしない | 447/456 UAT/#732の残存差分。現Cへ無断追加しない。fixture/実行はNOT_RUN |
| A06 field固有解除/既知型drift | [FE旧属性](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/frontend/src/features/master/routes/trimming-settings-model.ts#L30)、[pointer DTO](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/backend/internal/trimming/trimming_course_request.go#L41)、[nil保持](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/backend/internal/trimming/trimming_course_service.go#L86)、[指定なし可](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/docs/spec/screens/settings/master-trimming.md#L18) | course_type_id空→省略、target_size空→null、BE pointerでnull/省略合流。field固有の省略保持/明示解除/値設定と①API受領、②値→指定なし→GETを追加。0/空文字をFK代用にせず、汎用nullable全面拡張なし。trimming_option_service.go Create SortOrder int→model *intは既知型driftの具体caller確認として重複計数しない | 448後続。457新属性/画像/曜日clearとは別、今回Dのdescription clearへ暗黙追加しない |
| A07 具体Enter caller漏れ | [治療内容Input](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTabParts.tsx#L204)→[onSubmit caller](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTab.tsx#L123)→[create](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/frontend/src/features/medical-records/hooks/use-treatments-tab.ts#L229)、[C01 AC](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/docs/spec/shared/ui-contracts.md#L135) | Enterで無条件onSubmitの残存経路を450受領へ。IME候補確定/isComposing/keyCode229/repeat/pendingを限定fixture化。既存コード残存で新PR回帰とは断定しない。120数量欄を全Input対応証明にしない | 450③後続。通常Enter全面禁止/全欄2回化を新設せず、改行/候補確定/明示保存/既決autosave維持。現5枠へ追加なし |
| A08 wrapper接続差 | [SubmitButton](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/frontend/src/components/shared/Form/SubmitButton.tsx#L31)、[loading正本](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/docs/spec/design-states.md#L80)、[Button既存loading](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/frontend/src/components/ui/button.tsx#L46) | pending→disabledだけ、labelをloadingTextへ置換。既存Button.loadingへwrapperを接続しaria-busy/spinner/元ラベル保持を受領する。primitive自体が未実装とはしない | 450/226227③受領。新しいprimitiveを再実装せず、pending/再押下/表示fixtureはNOT_RUN。5枠scope不変 |
| A09 採択済み構成との文書矛盾 | [C01旧現状](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/docs/spec/shared/ui-contracts.md#L42)、[カード正本](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/docs/spec/screens/common-dialogs.md#L11)、[内包実装](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/frontend/src/features/medical-records/components/MedicalRecordStickyHeader.tsx#L282)、[#679](https://github.com/MinoruSoga/AnimalEkarte/pull/679) | 412「決定2026-10-15 Q1→X13」を根拠に、属性だけPatientInfoCardへ、hideStaff/診療種別非表示、来院種別/診療日/担当医操作はStickyHeader維持とC01現状を訂正。コードに仕様を縮小する変更ではない | 450/C01の文書整合。onOwnerClick等の既存認可を追加許可しない。製品変更や5枠の再割当なし |
| A10 完了証拠/部分成果の確認残 | [306](https://app.plane.so/baritechllc/browse/EMR-306/) 後発Done activity f19d7560-b125-4524-97cf-0dd5c702c2be（2026-10-20 07:51:31Z）と同日03:06本文/コメント、[245](https://app.plane.so/baritechllc/browse/EMR-245/)、[318](https://app.plane.so/baritechllc/browse/EMR-318/)、[400/#702](https://github.com/MinoruSoga/AnimalEkarte/pull/702) | 306はDone直前の限定STG UAT再実行PASS（comment fd0b67d6-3a21-4d15-821b-d2470f0b243d、2026-10-20 07:51:30.990830Z）を採用。空き2件選択→代表確認→連結→マスタ/ボード両表示→解除→個別空き復帰はPASS、DnD合体・連結中割当拒否実演・退院後連結維持はUNKNOWN/461499境界。証拠欠落/記録不整合/誤Doneとしない。245はカルテ3列GET採用までで他一覧filter継承を確認。318受付writeと400staff保存は③表示/④writeを分離。422 MoneyInput subsetは財務受領確認残。状態変更/全画面filter義務化なし | 306は448受領、245/318/400は450と既存write owner、422は財務側。担当漏れ/新欠陥と確定せず、実装5枠へ追加なし |

A01–A03の業務方針は既決だが、具体保存契約の設計残がある。これを単に「試験待ち」「担当未割当」だけへ置換して完了にしない。A04/A09は所有・文書整合、A10は限定PASS/UNKNOWN・部分成果の受領確認で、新規runtime欠陥件数として合算しない。306の原文をrootが投稿前live再読した結果を採用し、旧03:06の未確認記録だけで後発Doneを否定しない。個人名・fixture名は転記しない。

## 4. 新たな欠落に数えない既存追跡

| 対象 | 採用/残件の境界 |
|---|---|
| ホテル上限 | [spec-storage-trace](../db-contracts/spec-storage-trace.md)と[model/storage_reservations.go L40–49](https://github.com/MinoruSoga/AnimalEkarte/blob/4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34/backend/internal/model/storage_reservations.go#L40)、452/461/499へ追跡済み。新しい保存漏れとしない |
| Fable/回復 | 未接続・false gate・Slice Bは既存474/487等の追跡。今回Slice B/機能公開/証明書ルート接続は開始しない |
| RV4証明書・外注健診PDF | 486/487/500等で既存追跡。資料/asset/provider/font/store/writer残を新規設定writer欠落と混算しない |
| 463 / LIFF | 463はBE接続・owners scope flag true、LIFF activation falseを維持。scope不在や新regressionと誤分類しない |
| 519 / 489 / 514 | 519は互換第一段階、489 UATはClaimMetadata/library subset、514 UATはFE子成果。各部分成果を全consumer/全勤務完成へ拡張しない |
| 181 | 後続main873685b0bとSTG記録で過去の証跡不足は解消済み。今回新残件へ再登録しない |
| #808/#809 | openの先行caller/528修正は独立実装PRのまま。根拠として参照するが、今回docsへ他sessionの差分を取り込まない |

## 5. 計画への反映先と未完了の明示

- [段階台帳](emr-445-master-stages-ledger-20261021.md)：母集団と完了判定の分離、設計残の入口。
- [①計画§5.7](emr-445-stage1-contract-start-plan-20261021.md)：A01/A02/A03の共有保存/wire受領設計残とA06 field固有presence。#806/#807の限定採用と旧未受領snapshotを区別。
- [②計画§10](emr-445-stage2-master-start-plan-20261021.md)：設定writer・owner一意化・健診親race・trimming解除・受領証拠。
- [③検索計画](emr-450-stage3-common-ui-start-plan-20261021.md)：72用途の意味と確認残。[③非検索操作表](emr-450-stage3-common-ui-start-plan-20261021-operations.md)：A07/A08/A09/A10の限定caller/証拠受領。
- [C01正本](../../spec/shared/ui-contracts.md#requirement-fr-common-ui-context)：412採択の患者カード内包構成と現在認可を維持した現状記述訂正のみ。

今回の5実装sessionのbranch/claim/worktree、担当範囲、Plane state/assigneeへ操作しない。Plane追補はrootが担当し、このdocs編集で実装票をcloseしない。業務判断を追加する必要が生じれば該当項目だけを返し、既決内容を一括再承認へ戻さない。次に閉じるのはA01–A03の具体保存契約と各owner受領、限定caller/保護差分、A10の限定PASS/UNKNOWN・部分成果受領であり、文書追補だけで完了とはしない。

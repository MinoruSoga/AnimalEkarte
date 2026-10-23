# UI入力・親coverage静的監査（2026-10-23）

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.audit-ui-input-coverage-20261023 |
| 種別 | 固定mainの限定静的監査受領記録 |
| 決定状態 | 既決仕様を継承。新しい業務・技術契約を採択しない |
| 実装状態 | 下記の限定接点の分類。全FR/AC・全source監査完了ではない |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23 |

本稿は提供された監査報告を整形したもので、文書編集時に全sourceを独立再監査した記録ではない。[今回の索引](emr-audit-ready-handoff-index-20261023.md#audit-ready-additional-scope)を参照。

固定main: `77c28fbfbafa257dcca2421af99ff72da6889717`。最新Plane本文・全コメントを5票直接取得。件数: EMR-116=11, EMR-318=9, EMR-319=9, EMR-333=4, EMR-385=4（全票next_page_results=false）。

ブラウザ、実機、テスト、build、DB、製品変更、外部投稿はすべて未実施。見た目/再現/保存PASSを出さない。

<a id="ui-input-coverage-conclusions"></a>
## 結論

新規製品欠陥確定0件。333の既知未接続1原因・2surfaceを現mainで再確認しただけで再計上しない。116元端末、318 Windows/DB/ボード、319訂正再保存、333幾何、385表示受入は未確認を保持。

<a id="ui-input-coverage-adoption"></a>
## 採用証拠

- PR#609: https://github.com/MinoruSoga/AnimalEkarte/pull/609 merged=true / head `4492d2e2ccdeefc8b3f1aa06d7bc1734037e11cf` / merge `ffe981c40eb7e81275045a0e6e8f0cd3cbdf4ec4`。過去検証記録は今回PASSではない。

- PR#614: https://github.com/MinoruSoga/AnimalEkarte/pull/614 merged=true / head `4221765a1c9b0c26a58cc87e828ebfcd71d04469` / merge `e6f8b9b528a5a15504b8ad73a5a82f96045ad466`。過去検証記録は今回PASSではない。

- PR#619: https://github.com/MinoruSoga/AnimalEkarte/pull/619 merged=true / head `1b03eccd0f87a793c365c682612db4b404930c01` / merge `6732c63e1f0ad11e4300bed0ddab32cee738f9b0`。過去検証記録は今回PASSではない。

- PR#662: https://github.com/MinoruSoga/AnimalEkarte/pull/662 merged=true / head `beb855f51013f7b652e253c5100ce5e978d02a2d` / merge `949a55a08996f8d185ef15438030d6c7c51578d0`。過去検証記録は今回PASSではない。

<a id="ui-input-coverage-boundaries"></a>
## MECE・正本境界

- 116=担当候補の適格性/選択、318=患者結果スクロールと当日受付導線、319=既存検証エラーの描画、333=元レイアウトと子への追跡、385=403に集約された2表示所見。業務ルール・検索契約・配信実行を混ぜない。

- 333本文9課題とAC8行の差（カンバン空状態）を独立scope行で保持。全48画面の新規再監査や手検証は行わない。

- 333の採用索引は固定f0687b6の既存記録と最新Plane10/22コメントに基づく受領報告。元の私有manifestは本資料へ添付しておらず、今回用の新採用manifestをrepo上で発見したとは主張しない。本稿ではPR619→固定source path→既存子400/401/403/432の責務を結ぶ。

- 319の旧監査docは docs/work/emr/emr-319-required-field-audit.md に移動。是正前の⚠表を現存欠陥にしない。450 operations430/447行の受領条件どおり既存callerの限定証拠と全画面受入を分離。新規package等を旧11区分の全適用証明へ加算しない。

- 385はPhase1対象・日常業務の後。旧Phase1対象外文は履歴。403への採用は配信許可ではない。

<a id="ui-input-coverage-ac"></a>
## AC別照合

<a id="ui-input-coverage-staff-candidates"></a>
### EMR-116 / 候補状態と適格性

- 判定: STATIC_IMPLEMENTATION; runtime=NOT_RUN

- 所見: 候補なし/取得中/失敗/orphanを分離し、active・出勤・capability交差からoptionsを生成する現在callerを確認。全staff開放はしない。

- 有限後続owner: EMR-116

- 次: 医院・日付・予約区分・権限・staff条件を同一にした候補あり/0/取得中/失敗の受入fixtureを維持。

- [frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:103–332](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx#L103-L332)

- [frontend/src/components/shared/ReservationFormModal/ReservationTypeAndStaffFields.tsx:163–195](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/shared/ReservationFormModal/ReservationTypeAndStaffFields.tsx#L163-L195)

<a id="ui-input-coverage-staff-device-flow"></a>
### EMR-116 / 実Dialog→portaled Popover→cmdk選択→保存→再表示

- 判定: UNVERIFIED_DEVICE_DEPENDENCY; runtime=NOT_RUN

- 所見: modal Popover、pointerdown focus抑止、coarse autofocus抑止は接続済み。旧staff-candidates suiteはstubだが、別staff-select-device suiteは実component＋API mock。旧合成成功/desktop記録は元iPad/一部PCの解消証拠ではない。

- 有限後続owner: EMR-116 / 医院・QA（役割、担当追加なし）

- 次: 元端末機種/OS/ブラウザ版を取得した後、正常PC対照と選択・payload・再読込・focus差を確認。未入手は実機分保留。

- [frontend/src/components/ui/searchable-select.tsx:119–199](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/ui/searchable-select.tsx#L119-L199)

- [frontend/src/components/shared/ReservationFormModal/ReservationFormModal.staff-candidates.test.tsx:9–19](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/shared/ReservationFormModal/ReservationFormModal.staff-candidates.test.tsx#L9-L19)

- [frontend/src/components/shared/ReservationFormModal/ReservationFormModal.staff-select-device.test.tsx:205–299](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/shared/ReservationFormModal/ReservationFormModal.staff-select-device.test.tsx#L205-L299)

- [frontend/src/components/shared/ReservationFormModal/ReservationFormModal.staff-select-device.test.tsx:385–427](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/shared/ReservationFormModal/ReservationFormModal.staff-select-device.test.tsx#L385-L427)

<a id="ui-input-coverage-patient-search-scroll"></a>
### EMR-318 / Windows当日受付→新規追加の患者検索結果

- 判定: STATIC_IMPLEMENTATION_RUNTIME_UNVERIFIED; runtime=NOT_RUN

- 所見: min-h-0の親子連鎖と単一overflow-auto結果region、tabIndex、全20行fixtureが存在。Windows scrollbar・1366x768/125% zoomで最終行を選べるというACは静的証明できない。

- 有限後続owner: EMR-318 表示単位

- 次: 許可後、Windows Chromeの元寸法/zoom/返却件数を記録し末尾行選択を確認。受付登録E2Eと別receipt。

- [frontend/src/components/shared/ReservationFormModal/ReservationFormModalPanels.tsx:140–183](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/shared/ReservationFormModal/ReservationFormModalPanels.tsx#L140-L183)

- [frontend/src/components/shared/ReservationFormModal/PatientSelectionTable.tsx:35–71](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/shared/ReservationFormModal/PatientSelectionTable.tsx#L35-L71)

- [frontend/src/components/shared/ReservationFormModal/PatientSelectionResults.tsx:146–183](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/shared/ReservationFormModal/PatientSelectionResults.tsx#L146-L183)

- [frontend/src/components/shared/ReservationFormModal/PatientSelectionResults.test.tsx:95–123](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/shared/ReservationFormModal/PatientSelectionResults.test.tsx#L95-L123)

<a id="ui-input-coverage-patient-reception-entry"></a>
### EMR-318 / 飼主ペットから当日受付・権限/生存/保存済み条件

- 判定: STATIC_IMPLEMENTATION; runtime=NOT_RUN

- 所見: 生存/死亡ORゲート/永続ID/pending/予約create+viewとclick時境界を保持し、reception=1を当日checked_inへ接続。旧reception grant限定の表現より最終PRのreservations create/edit契約を採用。

- 有限後続owner: EMR-318 受付導線単位

- 次: 当日受付ボードへの到達・権限なし・死亡/不明・未保存の有限fixtureを許可後に受入。

- [frontend/src/features/owners/components/OwnerPetsSection.tsx:148–173](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/owners/components/OwnerPetsSection.tsx#L148-L173)

- [frontend/src/features/owners/components/OwnerPetsSection.tsx:397–400](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/owners/components/OwnerPetsSection.tsx#L397-L400)

- [frontend/src/features/reservations/hooks/use-reservation-modal-state.ts:66–94](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/reservations/hooks/use-reservation-modal-state.ts#L66-L94)

<a id="ui-input-coverage-reception-duplicate-guard"></a>
### EMR-318 / 既存予約再利用・新規同時受付重複防止

- 判定: STATIC_IMPLEMENTATION_DB_DEPENDENCY; runtime=NOT_RUN

- 所見: 0候補POSTだけguard opt-in、既存confirmedはexpected_updated_atでCAS、受付済以降再刻印しない。BE patient intent lock後に医院/生存/当日未完了候補再検査。旧クロスタブ未達を現在欠陥として再計上しない。

- 有限後続owner: EMR-318 / DB受入担当（役割）

- 次: 専用DB二接続、409入力保持、batch rollback、完了後同日別来院、医院/日跨ぎと当日ボードを既存手順で確認。今回DB/試験なし。

- [frontend/src/features/reservations/hooks/use-reservation-save-actions.ts:186–220](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/reservations/hooks/use-reservation-save-actions.ts#L186-L220)

- [frontend/src/features/reservations/hooks/use-reservation-save-actions.ts:233–292](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/reservations/hooks/use-reservation-save-actions.ts#L233-L292)

- [backend/internal/reservation/reservation_reception_create_guard.go:35–82](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/reservation/reservation_reception_create_guard.go#L35-L82)

<a id="ui-input-coverage-validation-primitives"></a>
### EMR-319 / 共通primitive赤枠/エラー関連付け

- 判定: STATIC_IMPLEMENTATION; runtime=NOT_RUN

- 所見: Input/Select/DatePicker/SearchableSelectのinvalid描画とFormFieldErrorを確認。共通描画能力はcallerのerror到達証明ではない。

- 有限後続owner: EMR-319 → EMR-450 required/invalid受領

- 次: 既存対象callerのみをerror→訂正→再保存へ対応。任意項目を必須化せず全画面合格へ外挿しない。

- [frontend/src/components/ui/input.tsx:10–27](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/ui/input.tsx#L10-L27)

- [frontend/src/components/shared/DatePicker/DatePickerSingle.tsx:1–150](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/shared/DatePicker/DatePickerSingle.tsx#L1-L150)

- [frontend/src/components/shared/FormFieldError/FormFieldError.tsx:1–16](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/shared/FormFieldError/FormFieldError.tsx#L1-L16)

- [docs/work/emr/emr-450-stage3-common-ui-start-plan-20261021-operations.md:426–447](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/work/emr/emr-450-stage3-common-ui-start-plan-20261021-operations.md#L426-L447)

<a id="ui-input-coverage-validation-reservations"></a>
### EMR-319 / 11区分和集合: 予約

- 判定: STATIC_IMPLEMENTATION_BOUNDED; runtime=NOT_RUN

- 所見: patient枠とエラー、日時/区分/担当へvalidationErrors伝播。 歴史棚卸しの区分と現source接点の照合であり、その区分の全動的分岐を受入したとはしない。

- 有限後続owner: EMR-319 → EMR-450

- 次: 既存対象フィールドで空保存→該当欄表示→訂正→再保存を限定fixture化。未読caller/後発業務の検証ルール変更は別ownerへ。

- [frontend/src/components/shared/ReservationFormModal/ReservationFormModalPanels.tsx:340–377](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/shared/ReservationFormModal/ReservationFormModalPanels.tsx#L340-L377)

<a id="ui-input-coverage-validation-records"></a>
### EMR-319 / 11区分和集合: カルテ

- 判定: STATIC_IMPLEMENTATION_BOUNDED; runtime=NOT_RUN

- 所見: 日時/体温等の既存エラー表示接続。診断・予防接種・健診callerもsource索引に保持。 歴史棚卸しの区分と現source接点の照合であり、その区分の全動的分岐を受入したとはしない。

- 有限後続owner: EMR-319 → EMR-450

- 次: 既存対象フィールドで空保存→該当欄表示→訂正→再保存を限定fixture化。未読caller/後発業務の検証ルール変更は別ownerへ。

- [frontend/src/features/medical-records/components/VitalsTab/VitalsTabRows.tsx:1–180](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/medical-records/components/VitalsTab/VitalsTabRows.tsx#L1-L180)

<a id="ui-input-coverage-validation-hospitalization"></a>
### EMR-319 / 11区分和集合: 入院

- 判定: STATIC_IMPLEMENTATION_BOUNDED; runtime=NOT_RUN

- 所見: 既存名称/参照/理由/単価の表示配線。ケージ任意を維持し新規必須化しない。 歴史棚卸しの区分と現source接点の照合であり、その区分の全動的分岐を受入したとはしない。

- 有限後続owner: EMR-319 → EMR-450

- 次: 既存対象フィールドで空保存→該当欄表示→訂正→再保存を限定fixture化。未読caller/後発業務の検証ルール変更は別ownerへ。

- [frontend/src/features/hospitalization/components/CarePlanTab/AddForm.tsx:1–250](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/hospitalization/components/CarePlanTab/AddForm.tsx#L1-L250)

<a id="ui-input-coverage-validation-accounting"></a>
### EMR-319 / 11区分和集合: 会計

- 判定: STATIC_IMPLEMENTATION_BOUNDED; runtime=NOT_RUN

- 所見: 返金等の既存入力エラー表示配線。実返金・会計更新は未実施。 歴史棚卸しの区分と現source接点の照合であり、その区分の全動的分岐を受入したとはしない。

- 有限後続owner: EMR-319 → EMR-450

- 次: 既存対象フィールドで空保存→該当欄表示→訂正→再保存を限定fixture化。未読caller/後発業務の検証ルール変更は別ownerへ。

- [frontend/src/features/accounting/components/RefundSection.tsx:1–220](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/accounting/components/RefundSection.tsx#L1-L220)

<a id="ui-input-coverage-validation-owner-pet"></a>
### EMR-319 / 11区分和集合: 飼主/ペット

- 判定: STATIC_IMPLEMENTATION_BOUNDED; runtime=NOT_RUN

- 所見: 既存field errorと赤枠/文字の接続。 歴史棚卸しの区分と現source接点の照合であり、その区分の全動的分岐を受入したとはしない。

- 有限後続owner: EMR-319 → EMR-450

- 次: 既存対象フィールドで空保存→該当欄表示→訂正→再保存を限定fixture化。未読caller/後発業務の検証ルール変更は別ownerへ。

- [frontend/src/features/owners/components/OwnerBasicFields.tsx:1–180](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/owners/components/OwnerBasicFields.tsx#L1-L180)

<a id="ui-input-coverage-validation-trimming"></a>
### EMR-319 / 11区分和集合: トリミング

- 判定: STATIC_IMPLEMENTATION_BOUNDED; runtime=NOT_RUN

- 所見: コースinvalid/errorと担当表示伝播。設定欠落を新必須fieldにしない。 歴史棚卸しの区分と現source接点の照合であり、その区分の全動的分岐を受入したとはしない。

- 有限後続owner: EMR-319 → EMR-450

- 次: 既存対象フィールドで空保存→該当欄表示→訂正→再保存を限定fixture化。未読caller/後発業務の検証ルール変更は別ownerへ。

- [frontend/src/features/trimming/components/TrimmingLeftColumn.tsx:1–181](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/trimming/components/TrimmingLeftColumn.tsx#L1-L181)

<a id="ui-input-coverage-validation-vaccinations"></a>
### EMR-319 / 11区分和集合: 予防接種

- 判定: STATIC_IMPLEMENTATION_BOUNDED; runtime=NOT_RUN

- 所見: ワクチン/接種日の既存fieldErrors接続。 歴史棚卸しの区分と現source接点の照合であり、その区分の全動的分岐を受入したとはしない。

- 有限後続owner: EMR-319 → EMR-450

- 次: 既存対象フィールドで空保存→該当欄表示→訂正→再保存を限定fixture化。未読caller/後発業務の検証ルール変更は別ownerへ。

- [frontend/src/features/vaccinations/components/VaccinationFormPanels.tsx:1–180](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/vaccinations/components/VaccinationFormPanels.tsx#L1-L180)

<a id="ui-input-coverage-validation-checkups"></a>
### EMR-319 / 11区分和集合: 健診

- 判定: STATIC_IMPLEMENTATION_BOUNDED; runtime=NOT_RUN

- 所見: 種別/実施日fieldErrorsからcontrolへ接続。 歴史棚卸しの区分と現source接点の照合であり、その区分の全動的分岐を受入したとはしない。

- 有限後続owner: EMR-319 → EMR-450

- 次: 既存対象フィールドで空保存→該当欄表示→訂正→再保存を限定fixture化。未読caller/後発業務の検証ルール変更は別ownerへ。

- [frontend/src/features/checkups/routes/CheckupFormPanels.tsx:1–140](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/checkups/routes/CheckupFormPanels.tsx#L1-L140)

<a id="ui-input-coverage-validation-examinations"></a>
### EMR-319 / 11区分和集合: 検査

- 判定: STATIC_IMPLEMENTATION_BOUNDED; runtime=NOT_RUN

- 所見: 検査種別/担当医、手動項目名はExamItemsTable198以降に条件付きinvalid。 歴史棚卸しの区分と現source接点の照合であり、その区分の全動的分岐を受入したとはしない。

- 有限後続owner: EMR-319 → EMR-450

- 次: 既存対象フィールドで空保存→該当欄表示→訂正→再保存を限定fixture化。未読caller/後発業務の検証ルール変更は別ownerへ。

- [frontend/src/features/examinations/components/ExaminationFormFields.tsx:54–132](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/examinations/components/ExaminationFormFields.tsx#L54-L132)

<a id="ui-input-coverage-validation-inventory"></a>
### EMR-319 / 11区分和集合: 在庫

- 判定: STATIC_IMPLEMENTATION_BOUNDED; runtime=NOT_RUN

- 所見: 名称/単位、同188-213に数量/下限のinvalidと説明。 歴史棚卸しの区分と現source接点の照合であり、その区分の全動的分岐を受入したとはしない。

- 有限後続owner: EMR-319 → EMR-450

- 次: 既存対象フィールドで空保存→該当欄表示→訂正→再保存を限定fixture化。未読caller/後発業務の検証ルール変更は別ownerへ。

- [frontend/src/features/inventory/components/InventoryFormSections.tsx:49–116](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/inventory/components/InventoryFormSections.tsx#L49-L116)

<a id="ui-input-coverage-validation-settings"></a>
### EMR-319 / 11区分和集合: 設定群

- 判定: STATIC_IMPLEMENTATION_BOUNDED; runtime=NOT_RUN

- 所見: noValidate親下でaction内必須検出。Holiday/StandardClosingはonInvalid、LstepTagはper-field、医院名・shift時刻も接続。 歴史棚卸しの区分と現source接点の照合であり、その区分の全動的分岐を受入したとはしない。

- 有限後続owner: EMR-319 → EMR-450

- 次: 既存対象フィールドで空保存→該当欄表示→訂正→再保存を限定fixture化。未読caller/後発業務の検証ルール変更は別ownerへ。

- [frontend/src/features/closing-settings/components/SpecialPeriodSection.tsx:98–110](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/closing-settings/components/SpecialPeriodSection.tsx#L98-L110)

<a id="ui-input-coverage-validation-corrected-errors"></a>
### EMR-319 / 訂正後の赤枠/期間エラー残留

- 判定: STATIC_IMPLEMENTATION; runtime=NOT_RUN

- 所見: CampaignSidePanelのperiodMissing per-field、全欠落解消時clear、正常再保存時periodError clearが存在。PR614の既知是正であり新規欠陥なし。

- 有限後続owner: EMR-319 → EMR-450

- 次: 空片側/両側、逆転→訂正→再保存の既存回帰を実行許可後に確認。

- [frontend/src/features/master/components/CampaignSidePanel.tsx:67–123](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/master/components/CampaignSidePanel.tsx#L67-L123)

- [frontend/src/features/master/components/CampaignSidePanel.tsx:175–234](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/master/components/CampaignSidePanel.tsx#L175-L234)

<a id="ui-input-coverage-reception-header"></a>
### EMR-333 / 1:390px受付ヘッダー

- 判定: STATIC_IMPLEMENTATION_RUNTIME_UNVERIFIED; runtime=NOT_RUN

- 所見: flex-wrap/min-w-0/full-width小画面action。 旧PR619成果と現caller対応。実見た目PASSなし。

- 有限後続owner: EMR-333 → EMR-400/401 該当子契約

- 次: 元ACの状態/幅だけを既存受入索引へ維持し、後続子票と二重修正しない。

- [frontend/src/components/shared/Form/FormHeader.tsx:20–46](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/shared/Form/FormHeader.tsx#L20-L46)

<a id="ui-input-coverage-reception-telemetry"></a>
### EMR-333 / 2:テレメトリ改行

- 判定: STATIC_IMPLEMENTATION_RUNTIME_UNVERIFIED; runtime=NOT_RUN

- 所見: 件数/時間はnowrap+shrink0、長患者名truncateは別契約。 旧PR619成果と現caller対応。実見た目PASSなし。

- 有限後続owner: EMR-333 → EMR-400/401 該当子契約

- 次: 元ACの状態/幅だけを既存受入索引へ維持し、後続子票と二重修正しない。

- [frontend/src/features/reception/components/ReceptionTelemetryStrip.tsx:31–68](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/reception/components/ReceptionTelemetryStrip.tsx#L31-L68)

<a id="ui-input-coverage-search-spacing"></a>
### EMR-333 / 3:1440px検索余白

- 判定: STATIC_IMPLEMENTATION_RUNTIME_UNVERIFIED; runtime=NOT_RUN

- 所見: 伸縮検索枠と現検索操作を保持。実glyph/余白バランス未測定。 旧PR619成果と現caller対応。実見た目PASSなし。

- 有限後続owner: EMR-333 → EMR-400/401 該当子契約

- 次: 元ACの状態/幅だけを既存受入索引へ維持し、後続子票と二重修正しない。

- [frontend/src/components/shared/PropertyFilter/PropertyFilter.tsx:150–200](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/shared/PropertyFilter/PropertyFilter.tsx#L150-L200)

<a id="ui-input-coverage-owner-address-grid"></a>
### EMR-333 / 4:住所グリッド

- 判定: STATIC_IMPLEMENTATION_RUNTIME_UNVERIFIED; runtime=NOT_RUN

- 所見: 自宅・会社を別groupに整列。 旧PR619成果と現caller対応。実見た目PASSなし。

- 有限後続owner: EMR-333 → EMR-400/401 該当子契約

- 次: 元ACの状態/幅だけを既存受入索引へ維持し、後続子票と二重修正しない。

- [frontend/src/features/owners/components/OwnerAddressFields.tsx:20–149](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/owners/components/OwnerAddressFields.tsx#L20-L149)

<a id="ui-input-coverage-hospitalization-actions"></a>
### EMR-333 / 5:入院詳細ボタン

- 判定: STATIC_IMPLEMENTATION_RUNTIME_UNVERIFIED; runtime=NOT_RUN

- 所見: auto-height/min-height/pb3。詳細action295は通常flow。 旧PR619成果と現caller対応。実見た目PASSなし。

- 有限後続owner: EMR-333 → EMR-400/401 該当子契約

- 次: 元ACの状態/幅だけを既存受入索引へ維持し、後続子票と二重修正しない。

- [frontend/src/features/hospitalization/components/HospitalizationBoard.tsx:171–215](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/hospitalization/components/HospitalizationBoard.tsx#L171-L215)

<a id="ui-input-coverage-cash-register-guidance"></a>
### EMR-333 / 6:レジ初期案内

- 判定: STATIC_IMPLEMENTATION_RUNTIME_UNVERIFIED; runtime=NOT_RUN

- 所見: preview前案内分岐。 旧PR619成果と現caller対応。実見た目PASSなし。

- 有限後続owner: EMR-333 → EMR-400/401 該当子契約

- 次: 元ACの状態/幅だけを既存受入索引へ維持し、後続子票と二重修正しない。

- [frontend/src/features/cash-register/routes/CashRegisterClosePage.tsx:150–168](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/cash-register/routes/CashRegisterClosePage.tsx#L150-L168)

<a id="ui-input-coverage-master-narrow-table"></a>
### EMR-333 / 7:少列master

- 判定: STATIC_IMPLEMENTATION_RUNTIME_UNVERIFIED; runtime=NOT_RUN

- 所見: columns<=5のtableNarrow、custom callerは別途。 旧PR619成果と現caller対応。実見た目PASSなし。

- 有限後続owner: EMR-333 → EMR-400/401 該当子契約

- 次: 元ACの状態/幅だけを既存受入索引へ維持し、後続子票と二重修正しない。

- [frontend/src/features/master/components/MasterCRUDPage.tsx:112–122](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/master/components/MasterCRUDPage.tsx#L112-L122)

<a id="ui-input-coverage-reception-empty-kanban"></a>
### EMR-333 / AC番号外だが元スコープ:受付カンバン空状態

- 判定: STATIC_IMPLEMENTATION; runtime=NOT_RUN

- 所見: 本文の9課題に対し番号付きACは8つ。カンバンempty guidanceを落とさず保持。KanbanColumnの受付なし表示と末尾clearanceを確認。

- 有限後続owner: EMR-333

- 次: 番号付きACだけで網羅率を作らず、この元scope行を親索引へ残す。

- [frontend/src/features/reception/components/KanbanColumn.tsx:73–89](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/reception/components/KanbanColumn.tsx#L73-L89)

<a id="ui-input-coverage-support-clearance"></a>
### EMR-333 / 8:FAB保護/後発package幅

- 判定: KNOWN_UNCONNECTED_CALLERS; runtime=NOT_RUN

- 所見: 共通80px clearance/overlay抑止は存在。ただしpackage独立pane475行と週grid55行には保護接続なし。10/22既知同一原因2surfaceを再確認。package4列表のmaxWidth未指定も既知限定残差。現mainは行番号移動あり。実衝突は未測定。

- 有限後続owner: EMR-333 AC8既存所見 → 共通EMR-432、残存EMR-401 C3-1/403（子契約へ対応）

- 次: package長editor末尾と週grid末尾の既存clearance接続/抑止候補を既存ownerで追跡。全viewport重複・全画面完了を断定しない。

- [frontend/src/features/master/routes/TreatmentPackageSettings.tsx:470–590](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/master/routes/TreatmentPackageSettings.tsx#L470-L590)

- [frontend/src/features/master/components/TreatmentPackageEditor.tsx:124–151](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/master/components/TreatmentPackageEditor.tsx#L124-L151)

- [frontend/src/features/master/components/MasterPageShell.tsx:51–81](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/master/components/MasterPageShell.tsx#L51-L81)

- [frontend/src/features/reservations/components/WeekViewParts.tsx:48–101](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/reservations/components/WeekViewParts.tsx#L48-L101)

- [frontend/src/constants/support-widget-layout.ts:1–21](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/constants/support-widget-layout.ts#L1-L21)

- [frontend/src/styles/globals.css:344–355](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/styles/globals.css#L344-L355)

<a id="ui-input-coverage-delivery-monitor-labels"></a>
### EMR-385 / delivery-monitor全ステータス値の切断

- 判定: STATIC_IMPLEMENTATION_DEPENDENCY; runtime=NOT_RUN

- 所見: SelectTrigger w-full/sm:w-auto/sm:min-w180と完全ラベルが現存。PR662採用済みを再実装しない。

- 有限後続owner: EMR-403（EMR-385所見を受領）

- 次: Phase1内で日常業務の後に、対象幅とfilter状態の可読性を受入。配信再開/資格登録を表示修正条件にしない。

- [frontend/src/features/lstep/routes/LstepDeliveryMonitorPageParts.tsx:78–96](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/lstep/routes/LstepDeliveryMonitorPageParts.tsx#L78-L96)

<a id="ui-input-coverage-checkup-sync-labels"></a>
### EMR-385 / checkup-sync任意ラベルの分断

- 判定: STATIC_IMPLEMENTATION_DEPENDENCY; runtime=NOT_RUN

- 所見: 以降・任意等を意味単位nowrap spanに分割。旧画像/合成証拠と現描画は別。CheckupSyncPage testはAPI mockでshell/permission主体、今回取得した範囲では実glyph幅保証なし。

- 有限後続owner: EMR-403（EMR-385所見を受領）

- 次: 390/768等の既存対象幅でラベル可読性を確認する有限残件。送信や新マーケティング業務を追加しない。

- [frontend/src/features/lstep/components/CheckupSyncFilterForm.tsx:202–225](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/lstep/components/CheckupSyncFilterForm.tsx#L202-L225)

- [frontend/src/features/lstep/components/CheckupSyncFilterForm.tsx:340–363](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/lstep/components/CheckupSyncFilterForm.tsx#L340-L363)

- [frontend/src/features/lstep/routes/CheckupSyncPage.test.tsx:94–116](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/lstep/routes/CheckupSyncPage.test.tsx#L94-L116)

<a id="ui-input-coverage-evidence-limits"></a>
## 証拠限界

取得に失敗した旧pathは移動後の正規pathへ再解決した。推測したLstepDeliveryMonitorPageParts.test.tsxは取得できず、そのfixtureが存在すると主張しない。testファイルの静的読取は実行結果ではない。全5票assignee空のためownerは既存票/役割で表し担当者を新設していない。根拠は本文の固定sourceリンクと採用履歴に限定する。

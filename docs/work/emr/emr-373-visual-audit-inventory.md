# EMR-373 実画面ビジュアル監査 — 静的インベントリ

- タスク: EMR373-INVENTORY revision2 + Lead追試索引
- 対象コミット（候補ベースライン）: `1f5d81ad4ef0e5203270d55495fbfc01d079f237`（フロント挙動は `98a375b5f` と同等）
- 親監査: [EMR-373](https://app.plane.so/baritechllc/browse/EMR-373/)（子 EMR-377..388）

最新の実装修正・撮影範囲は[追試記録](emr-373-visual-audit-results.md)を参照。以下のUNKNOWN理由キーと「未着手」は初回の静的棚卸し時点を示す。撮影結果を未確認の別モード・埋込フォームへ継承しない。

## 1. この文書の位置づけ

**静的インベントリのみ。** ブラウザ受入・コード修正・最終合否は本タスクの範囲外であり、別タスク/別diffで行う。ここに列挙する「ソース確認済」はルート/コンポーネントの存在を静的に確認したことのみを意味し、視覚的合格を意味しない。既存の履歴スクリーンショットは HMR/対象SHA の出所が未検証のため、候補ビルドの視覚根拠には使えない。新規の合成レンダリングは許可済みのタスク所有 Stagehand で Lead が取得・レビューする。個別画面の撮影/レビュー状態は付随レポートおよび Plane コメントに記録され、本書のソース確認から継承しない。Lead が証跡を追記するまで全行 visual=UNKNOWN のままとする。

## 2. 共有メタデータ（全行共通）

- 候補ベースライン: `1f5d81ad4ef0e5203270d55495fbfc01d079f237`
- 予定ビューポート: `1568x873` / `1280x800` / `768x1024` / `390x844`
- ブラウザ/撮影日時/視覚証跡: 本書では記録しない（Lead が付随レポート・Plane コメントで管理。ここでは UNKNOWN）
- ソースパスの接頭辞 `frontend/src/` は各行で省略（別アプリは明示）

## 3. ステータス規約

- **ソース確認** = ルート定義・feature ファイルの静的確認で存在を証明した。
- **visual = UNKNOWN** = 全行一律。ソース存在・テスト・履歴画像は視覚合格を証明しない。

### UNKNOWN 理由キー（各行「理由」列で参照）

| キー | 理由 |
|---|---|
| K1 | 静的ソースのみ確認。撮影/レビュー状態はソースから継承不可（付随レポート・Plane で管理）。履歴画像は HMR/SHA 出所未検証で根拠不可 |
| K2 | データ依存状態（空・長文・死亡・多段・エラー・未払・他医院スコープ）未観測 |
| K3 | 作成/編集/詳細モード別の視覚状態未観測（データゼロの編集モードを含む） |
| K4 | モーダル/ドロワー/サイドパネル/行メニューの開閉・重なり・フォーカス未観測 |
| K5 | 別権限ロール（viewのみ・操作不可等）の表示未観測 |
| K6 | 別アプリ（LIFF / line-reserve）のブラウザ観測なし。トークン・非空・分岐状態含む |
| K7 | 多段フロー/状態機械の各分岐が未観測 |

## 4. インベントリ

凡例: モード＝ list / select / create / edit / detail / dialog / panel / drawer / sidepanel / tabs / nested / menu / operation / print / step / state / overlay / shell / report

### 4.1 受付 `/`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-A01 | 当日受付カンバン（列ボード+DnD） | list | `/` index | features/reception/routes/Reception.tsx, routes/ReceptionPagePanels.tsx | K1,K2 | VQ-001/EMR-377 |
| INV-A02 | 受付フィルタパネル | panel | `/` | features/reception/components/ReceptionFilterPanel.tsx | K1,K2 | VQ-001 |
| INV-A03 | テレメトリ帯 | nested | `/` | features/reception/components/ReceptionTelemetryStrip.tsx | K1 | — |
| INV-A04 | 受付詳細モーダル（操作ボタン群） | dialog detail | `/` | features/reception/components/ReceptionDetailModal.tsx, ReceptionDialogActionButtons.tsx | K1,K4 | — |
| INV-A05 | 予約編集モーダル（受付入口、共有） | dialog edit | `/` | routes/ReceptionPagePanels.tsx→shared/ReservationFormModal | K1,K3,K4 | — |
| INV-A06 | ステータス更新（DnD確定操作） | operation | `/` | Reception.tsx, api/update-appointment-status.ts | K1,K2 | — |
| INV-A07 | 予約キャンセル確認 | dialog | `/` | ReceptionPagePanels.tsx（ConfirmDialog） | K1,K4 | — |
| INV-A08 | 次回予定リセットダイアログ | dialog | `/` | Reception.tsx→shared/ReservationFormModal/NextVisitResetDialog.tsx | K1,K4 | — |

### 4.2 飼主・ペット `/owners`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-O01 | 飼主一覧 | list | `/owners` | app/pages/OwnersListPage.tsx, features/owners/routes/OwnersList.tsx, components/OwnersListTable.tsx | K1,K2 | VQ-002,VQ-012/EMR-378,388 |
| INV-O02 | 一覧行アクション（編集/レポート/削除） | menu | `/owners` | components/OwnerTableRow.tsx, OwnerTableRowButtons.tsx | K1,K4 | 行メニュー未観測 |
| INV-O03 | 飼主新規 | create | `/owners/new` | app/pages/OwnerFormPage.tsx, routes/OwnerForm.tsx | K1,K3 | — |
| INV-O04 | 飼主詳細・編集 | detail+edit | `/owners/:id` | OwnerFormPage.tsx, OwnerForm.tsx, OwnerFormTabs.tsx | K1,K3 | — |
| INV-O05 | ペット行ドロップダウン（操作メニュー） | menu | `/owners/:id` | components/OwnerPetsSection.tsx（DropdownMenu） | K1,K4 | 行メニュー未観測 |
| INV-O06 | PetEditModal 新規 | dialog create | `/owners/:id` | components/PetEditModal.tsx | K1,K3,K4 | — |
| INV-O07 | PetEditModal 編集（`?pet=` ディープリンク含む） | dialog edit | `/owners/:id?pet=:petId` | PetEditModal.tsx, PetEditModalFields/Header/Body.tsx | K1,K3,K4 | — |
| INV-O08 | 副飼主インライン編集 | nested inline | `/owners/:id` ペットモーダル内 | PetSubOwnersSection.tsx | K1,K2,K3 | — |
| INV-O09 | 慢性疾患インライン編集 | nested inline | 同上 | PetChronicConditionsSection.tsx | K1,K2,K3 | — |
| INV-O10 | LINE連携カード | panel | `/owners/:id` | LineIntegrationCard.tsx, LineIntegrationCardParts/Sections.tsx | K1,K5 | — |
| INV-O11 | LINE送信パネル | dialog | `/owners/:id` | LineSendPanel.tsx, LineSendTypeSelector.tsx | K1,K4 | — |
| INV-O12 | Lステップタグ追加ダイアログ | dialog | `/owners/:id` | LstepTagAddDialog.tsx | K1,K4 | — |
| INV-O13 | Lステップタグ一覧・除去インライン | nested | `/owners/:id` | LstepTagList.tsx, LstepTagRemoveInline.tsx | K1,K2 | — |
| INV-O14 | 飼主カルテレポート（Layout外スタンドアロン） | report | `/owners/:id/report` | features/owner-report（OwnerReport）, app/routes/app-routes.tsx:51 | K1 | 一覧を含むため対象 |

### 4.3 予約管理 `/reservations`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-R01 | 予約管理カレンダー（週5/7日・月・フィルタ） | list | `/reservations` | app/pages/ReservationsPage.tsx, features/reservations/routes/ReservationManagement.tsx, components/ReservationManagementCalendar.tsx | K1,K2 | — |
| INV-R02 | 期限超過tentativeパネル | panel | `/reservations` | components/ExpiredTentativePanel.tsx | K1,K2 | — |
| INV-R03 | 新規予約モーダル（共有・reservations入口） | dialog create | `/reservations` | components/shared/ReservationFormModal/ReservationFormModal.tsx | K1,K3,K4 | VQ-011 |
| INV-R04 | 予約編集モーダル（共有） | dialog edit | `/reservations` | 同上 | K1,K3,K4 | 詳細/編集未観測 |
| INV-R05 | 予約詳細モーダル | dialog detail | `/reservations` | components/ReservationDetailModal.tsx | K1,K4 | 同上 |
| INV-R06 | tentative確定モーダル（同一モーダルのconfirmモード） | dialog | `/reservations` | ReservationManagement.tsx（tentativeConfirmId） | K1,K3,K4 | — |
| INV-R07 | 次回予定リセットダイアログ | dialog | `/reservations` | shared/ReservationFormModal/NextVisitResetDialog.tsx | K1,K4 | — |
| INV-R08 | tentative終了確認 | dialog | `/reservations` | ReservationManagement.tsx（ConfirmDialog） | K1,K4 | — |
| INV-R09 | 予約削除確認 | dialog | `/reservations` | 同上 | K1,K4 | — |
| INV-R10 | ステータス変更確認 | dialog | `/reservations` | 同上 | K1,K4 | — |
| INV-R11 | ペット未紐付確認 | dialog | `/reservations` | 同上 | K1,K4 | — |
| INV-R12 | 患者選択テーブル（共有フォーム内） | nested | 共有 | shared/ReservationFormModal/PatientSelectionTable.tsx | K1,K4 | — |
| INV-R13 | 予約種別ピッカーダイアログ | nested dialog | 共有 | shared/ReservationFormModal/ReservationTypePickerDialog.tsx | K1,K4 | — |
| INV-R14 | 新規飼主インラインフォーム | nested inline | 共有 | shared/ReservationFormModal/NewOwnerInlineForm.tsx, ReservationFormModalPanels.tsx | K1,K4 | — |
| INV-R15 | 次回来院確定入口（カルテ画面発） | entry variant | `/medical-records/:id` | features/medical-records/components/NextVisitButton.tsx | K1,K3,K4 | — |

入口バリアント（同一 ReservationFormModal、別初期値・別権限）: **受付** INV-A05 / **予約管理** INV-R03..R06 / **次回来院** INV-R15 の3入口を別面として監査する。

### 4.4 カルテ `/medical-records`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-M01 | カルテ一覧 | list | `/medical-records` | routes/MedicalRecords.tsx（+MedicalRecordsList系） | K1,K2 | VQ-003/EMR-379 |
| INV-M02 | ペット選択 | select | `/medical-records/select-pet` | routes/MedicalRecordPetSelection.tsx | K1,K2 | — |
| INV-M03 | カルテ新規 | create | `/medical-records/new` | routes/MedicalRecordForm.tsx | K1,K3 | VQ-004/EMR-380 |
| INV-M04 | カルテ詳細・編集 | detail+edit | `/medical-records/:id` | 同上 | K1,K2,K3 | VQ-004,007 |
| INV-M05 | スティッキーヘッダ（担当医表示） | nested | カルテ画面 | MedicalRecordStickyHeader.tsx, components/shared/PatientContextHeader/PatientContextHeader.tsx | K1 | VQ-004 |
| INV-M06 | 診療タブ群 | tabs | 同上 | MedicalRecordClinicalTabs.tsx | K1,K4 | — |
| INV-M07 | サービスタブ群 | tabs | 同上 | MedicalRecordServiceTabs.tsx | K1,K4 | — |
| INV-M08 | バイタルタブ | nested | 同上 | VitalsTab/VitalsTab.tsx | K1,K2 | — |
| INV-M09 | バイタル入力モーダル | dialog | 同上 | VitalsModal.tsx | K1,K4 | — |
| INV-M10 | 処置テーブル・タブ | nested | 同上 | TreatmentTable.tsx, TreatmentsTab/*（TreatmentRow/RowEditors/QuantityCell/DoseMessages/TabParts） | K1,K2 | — |
| INV-M11 | 追記（Addendum）モーダル | dialog | 同上 | MedicalRecordAddenda/AddendumModal.tsx, MedicalRecordAddenda.tsx | K1,K4 | — |
| INV-M12 | 担当者選択モーダル | dialog | 同上 | StaffSelectionModal.tsx | K1,K4 | — |
| INV-M13 | 検査取込ダイアログ | dialog | 同上 | ExaminationImportDialog.tsx | K1,K4 | — |
| INV-M14 | 次回来院ボタン＋確定モーダル | nested+dialog | 同上 | NextVisitButton.tsx（=INV-R15） | K1,K4 | — |
| INV-M15 | 画像ギャラリー | nested | 同上 | MedicalRecordImage.tsx, ImageGalleryFilter.tsx, ImageGalleryGroup.tsx | K1,K2 | — |
| INV-M16 | 問診・治療履歴 | nested | 同上 | MedicalRecordInterview.tsx, InterviewChiefComplaint/History/TreatmentPolicy.tsx（InterviewHistory.tsx が「治療履歴」見出しを描画） | K1,K2 | VQ-004（右パネル「治療履歴」折返し） |
| INV-M17 | 診断ヘッダ・臨床計画 | nested | 同上 | DiagnosisHeader*.tsx, ClinicalPlanSection/* | K1,K2 | — |
| INV-M18 | 埋込ワクチンフォーム | nested form | 同上 | MedicalRecordVaccination.tsx, VaccinationForm.tsx | K1,K3 | — |
| INV-M19 | 埋込健診（編集可） | nested edit | 同上 | CheckupsTab/*（CheckupsTab/Rows/Table/Badges） | K1,K3 | INV-C04と同一面 |
| INV-M20 | 埋込検査フォーム | nested form | 同上 | MedicalRecordExamination.tsx | K1,K3 | — |
| INV-M21 | 埋込見積フォーム | nested form | 同上 | MedicalRecordEstimate.tsx, EstimateForm.tsx | K1,K3 | — |
| INV-M22 | 埋込会計確認 | nested form | 同上 | MedicalRecordBillCheck.tsx | K1,K3 | — |
| INV-M23 | 印刷ビュー | print | 同上 | MedicalRecordPrintView.tsx | K1 | — |
| INV-M24 | フォームモーダルホスト群 | dialogs | 同上 | MedicalRecordFormModals.tsx, MedicalRecordLazyModals.tsx | K1,K4 | — |

### 4.5 入院 `/hospitalization`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-H01 | 入院一覧（ボード+展開行+テーブル） | list+expanded | `/hospitalization` | routes/HospitalizationList.tsx, components/HospitalizationBoard.tsx, HospitalizationExpandedView.tsx, HospitalizationTable.tsx | K1,K2 | — |
| INV-H02 | ペット選択 | select | `/hospitalization/select-pet` | routes/HospitalizationPetSelection.tsx | K1,K2 | — |
| INV-H03 | 新規 | create | `/hospitalization/new` | routes/HospitalizationForm.tsx | K1,K3 | — |
| INV-H04 | 編集 | edit | `/hospitalization/:id/edit` | 同上 | K1,K3 | — |
| INV-H05 | 詳細 | detail | `/hospitalization/:id` | routes/HospitalizationDetail.tsx（+DetailPanels） | K1,K2 | — |
| INV-H06 | 日別バイタル インライン編集 | nested inline | 詳細内 | DailyRecordsTab/DailyVitalsSection.tsx | K1,K2,K3 | — |
| INV-H07 | 日別ケアログ（ダイアログ含む） | nested+dialog | 同上 | DailyRecordsTab/DailyCareLogsSection.tsx | K1,K4 | — |
| INV-H08 | スタッフノート（ダイアログ含む） | nested+dialog | 同上 | DailyRecordsTab/DailyStaffNotesSection.tsx | K1,K4 | — |
| INV-H09 | ケアプランタブ | nested tab | 同上 | CarePlanTab/CarePlanTab.tsx | K1,K2 | — |
| INV-H10 | ケアプラン追加フォーム | nested form | 同上 | CarePlanTab/AddForm.tsx | K1,K3 | — |
| INV-H11 | ケアプラン行編集 | nested edit | 同上 | CarePlanTab/EditRow.tsx | K1,K3 | — |

### 4.6 トリミング `/trimming`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-T01 | トリミング一覧 | list | `/trimming` | routes/TrimmingList.tsx | K1,K2 | VQ-002/EMR-378 |
| INV-T02 | ペット選択 | select | `/trimming/select-pet` | routes/TrimmingPetSelection.tsx | K1,K2 | — |
| INV-T03 | 新規 | create | `/trimming/new` | routes/TrimmingForm.tsx（+TrimmingImageUploadField） | K1,K3 | — |
| INV-T04 | 詳細・編集 | detail+edit | `/trimming/:id` | 同上 | K1,K3 | — |

### 4.7 ワクチン `/vaccinations`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-V01 | ワクチン一覧 | list | `/vaccinations` | routes/VaccinationList.tsx | K1,K2 | VQ-002/EMR-378 |
| INV-V02 | ペット選択 | select | `/vaccinations/select-pet` | routes/VaccinationPetSelection.tsx | K1,K2 | — |
| INV-V03 | 新規（独立フォーム、BUG-501） | create | `/vaccinations/new` | routes/VaccinationForm.tsx | K1,K3 | — |
| INV-V04 | 詳細・編集 | detail+edit | `/vaccinations/:id` | 同上 | K1,K3 | — |

### 4.8 検査 `/examinations`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-E01 | 検査一覧 | list | `/examinations` | routes/ExaminationsList.tsx | K1,K2 | VQ-006/EMR-382 |
| INV-E02 | ペット選択 | select | `/examinations/select-pet` | routes/ExaminationPetSelection.tsx | K1,K2 | — |
| INV-E03 | 新規 | create | `/examinations/new` | routes/ExaminationForm.tsx | K1,K3 | — |
| INV-E04 | 詳細・編集 | detail+edit | `/examinations/:id` | 同上 | K1,K2,K3 | VQ-005,007/EMR-381,383 |
| INV-E05 | 検査項目テーブル | nested | 詳細内 | ExamItemsTable.tsx, ExamPivotTable.tsx | K1,K2 | — |
| INV-E06 | 患者変更ダイアログ | dialog | 同上 | ExaminationPatientChangeDialog.tsx | K1,K4 | — |
| INV-E07 | 確定解除ダイアログ | dialog | 同上 | ExaminationUnconfirmDialog.tsx | K1,K4 | — |
| INV-E08 | 検査履歴パネル | nested | 同上 | ExaminationHistoryPanel.tsx | K1,K2 | — |
| INV-E09 | 印刷エリア | print | 同上 | ExaminationPrintArea.tsx | K1 | — |

### 4.9 検査機器 `/lab-device`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-L01 | 機器ボード（受信/保存済/未紐付 各リスト） | list | `/lab-device` | routes/LabDeviceBoard.tsx, LabDeviceBoardPanels.tsx, components/LabDeviceBoardCards.tsx | K1,K2 | VQ-005/EMR-381 |
| INV-L02 | 紐付操作（attach） | operation | `/lab-device` | routes/LabDeviceBoardPanels.tsx（UnlinkedPanel）, useAttachLabDeviceJob | K1,K2 | — |
| INV-L03 | 紐付解除操作（detach） | operation | `/lab-device` | LabDeviceBoardCards.tsx（JobCardView）, useDetachLabDeviceJob | K1,K2 | — |

### 4.10 健診 `/checkups`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-C01 | 健診一覧 | list | `/checkups` | routes/CheckupsList.tsx | K1,K2 | — |
| INV-C02 | ペット選択（medical-records create+edit 二重ガード） | select | `/checkups/select-pet` | routes/CheckupPetSelection.tsx | K1,K2,K5 | — |
| INV-C03 | 新規 | create | `/checkups/new` | routes/CheckupForm.tsx | K1,K3 | — |
| INV-C04 | 埋込健診編集（カルテ内 CheckupsTab） | nested edit | `/medical-records/:id` | features/medical-records/components/CheckupsTab/* | K1,K3 | INV-M19と同一面 |

**注記**: `/checkups/:id` ルートは存在しない（clinical-care-routes.tsx）。健診の編集面はカルテ内埋込のみ。除外表 EX-01 参照。

### 4.11 在庫 `/inventory`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-I01 | 在庫一覧 | list | `/inventory` | routes/InventoryList.tsx | K1,K2 | VQ-002/EMR-378 |
| INV-I02 | 新規 | create | `/inventory/new` | routes/InventoryForm.tsx | K1,K3 | — |
| INV-I03 | 詳細・編集 | detail+edit | `/inventory/:id` | 同上 | K1,K3 | — |

### 4.12 見積 `/estimates`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-S01 | 見積一覧 | list | `/estimates` | routes/EstimateList.tsx | K1,K2 | — |
| INV-S02 | 新規 | create | `/estimates/new` | routes/EstimateForm.tsx, EstimateFormSections.tsx, EstimateLineItems/* | K1,K3 | — |
| INV-S03 | 編集 | edit | `/estimates/:id/edit` | routes/EstimateForm.tsx | K1,K3 | — |
| INV-S04 | 詳細 | detail | `/estimates/:id` | routes/EstimateDetail.tsx, EstimateDetailPanels.tsx | K1,K2 | VQ-007/EMR-383 |
| INV-S05 | 明細行 | nested | フォーム内 | components/EstimateLineItems/EstimateLineItems.tsx | K1,K2 | — |
| INV-S06 | 後続見積（successor）作成ダイアログ＋成功遷移 | dialog+state | `/estimates/:id`→`/new` | EstimateDetail.tsx（useCreateEstimateSuccessor）, EstimateForm.tsx（formState.success） | K1,K3,K4 | — |

### 4.13 会計 `/accounting`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-B01 | 会計一覧（日別+未払タブ） | list+tabs | `/accounting` | routes/AccountingList.tsx, AccountingListTable.tsx, DailyAccountingTab*.tsx, UnpaidTab*.tsx | K1,K2 | — |
| INV-B02 | ペット選択 | select | `/accounting/select-pet` | routes/AccountingPetSelection.tsx | K1,K2 | — |
| INV-B03 | 新規 | create | `/accounting/new` | app/pages/AccountingDetailPage.tsx | K1,K3 | VQ-008/EMR-384 |
| INV-B04 | 詳細・編集 | detail+edit | `/accounting/:id` | 同上（+AccountingDetailPanels） | K1,K2,K3 | VQ-008/EMR-384 |
| INV-B05 | 明細テーブル | nested | 詳細内 | ItemListCard.tsx, AccountingItemRow.tsx | K1,K2 | VQ-008 |
| INV-B06 | 支払カード | nested form | 同上 | PaymentCard.tsx | K1,K3 | — |
| INV-B07 | 返金セクション | nested form | 同上 | RefundSection.tsx | K1,K3 | — |
| INV-B08 | 貸方修正ダイアログ | dialog | 同上 | CreditCorrectionDialog.tsx | K1,K4 | — |
| INV-B09 | 飼主会計履歴・未払残高 | nested | 同上 | OwnerAccountingHistory*.tsx, OwnerUnpaidBalanceCard.tsx, InsuranceCard.tsx | K1,K2 | — |
| INV-B10 | 会計文書・日別印刷 | print | 同上 | AccountingDocument.tsx, DailyAccountingPrintArea.tsx | K1 | — |

### 4.14 レジ締め `/accounting/close*`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-G01 | レジ締め | form | `/accounting/close` | routes/CashRegisterClosePage.tsx, CashRegisterClosePanels.tsx, CashReconciliationCard.tsx, UnifiedClosingSummaryTable.tsx, BillingDetailTable.tsx, ClosePrintArea.tsx | K1,K3 | — |
| INV-G02 | 締め履歴 | list | `/accounting/close/history` | routes/CashRegisterHistoryPage.tsx, CashRegisterHistoryTable.tsx, CashRegisterHistoryFilters.tsx | K1,K2 | — |
| INV-G03 | 履歴詳細ダイアログ | dialog | 同上 | components/CashRegisterHistoryDetailDialog.tsx | K1,K4 | — |

### 4.15 レポート・集計 `/accounting/reports`, `/aggregation`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-Q01 | 会計レポート（埋込表群） | embedded lists | `/accounting/reports` | routes/AccountingReportsPage.tsx, AccountingReportsPagePanels.tsx, MonthlySummaryCards.tsx, CategoryPaymentMatrixTable.tsx, DailyBreakdownTable.tsx, MonthlyReportPrintArea.tsx | K1,K2 | 埋込リストのみ対象（ダッシュボード全体の免除なし） |
| INV-Q02 | 集計ダッシュボード（チップ+埋込表） | embedded lists | `/aggregation` | routes/AggregationDashboardPage.tsx, AggregationDashboardPanels.tsx, AggregationFilterPanel*.tsx, CPMStageSummary.tsx, AggregationOwnerTable*.tsx | K1,K2 | VQ-012/EMR-388 |

### 4.16 シフト `/shifts`, テンプレート `/settings/shift-templates`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-F01 | シフトカレンダー | list | `/shifts` | routes/ShiftCalendarPage.tsx, ShiftCalendar/*, ShiftCell/* | K1,K2 | — |
| INV-F02 | シフト登録・編集ダイアログ | dialog create+edit | `/shifts` | ShiftFormDialog/ShiftFormDialog.tsx | K1,K3,K4 | — |
| INV-F03 | 休診日モーダル | dialog | `/shifts` | ClinicHolidayModal/ClinicHolidayModal.tsx | K1,K4 | — |
| INV-F04 | シフトテンプレート一覧 | list | `/settings/shift-templates` | routes/ShiftTemplateSettings.tsx, ShiftTemplateSettingsWorkspace/List/Parts.tsx | K1,K2 | — |
| INV-F05 | テンプレート登録・編集サイドパネル | sidepanel create+edit | 同上 | ShiftTemplateSidePanelFields.tsx | K1,K3,K4 | — |

### 4.17 マスタ・設定 `/settings/*`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-P01 | 設定トップ（カード一覧） | list | `/settings` | master/routes/MasterSettingsIndex.tsx | K1 | — |
| INV-P02 | スタッフ一覧 | list | `/settings/staff` | master/routes/StaffSettings.tsx | K1,K2 | VQ-002,010/EMR-378,386 |
| INV-P03 | スタッフサイドパネル（新規/編集） | sidepanel create+edit | `/settings/staff` | StaffSidePanel.tsx, StaffSidePanelSections.tsx, StaffBasicInfoSection.tsx, StaffCheckboxSection.tsx, StaffLineReservationSection.tsx | K1,K3,K4 | VQ-010/EMR-386 |
| INV-P04 | 診療項目・プラン（タブ: vaccine/examination/consultation/procedure） | tabs+list+sidepanel | `/settings/treatment-items` | TreatmentPlanMaster.tsx, TreatmentPlanMasterView.tsx, TreatmentPlanTabContent.tsx, TreatmentPlanRows.tsx, TreatmentItemSidePanel.tsx, TreatmentPlanSidePanelHost.tsx, TreatmentPlanDeleteDialog.tsx | K1,K3,K4 | — |
| INV-P05 | 検査種別フィールド表 | nested | `/settings/treatment-items` 内 | ExamTypeFieldsEditor.tsx, ExamTypeFieldsTable.tsx, ExamTypeFieldEditors.tsx, ExamTypeFieldDraftPanel.tsx | K1,K2,K3 | — |
| INV-P06 | 健診種別フィールド表 | nested | 同上 | CheckupTypeFieldsEditor.tsx, CheckupTypeFieldsTable.tsx, CheckupTypeFieldEditors.tsx, CheckupTypeFieldDraftPanel.tsx | K1,K2,K3 | — |
| INV-P07 | 傷病名設定（diagnosis_type/diagnosis_name タブ） | tabs+list+sidepanel | `/settings/diagnosis` | DiagnosisSettings.tsx, DiagnosisTabs.tsx, DiagnosisSortableTable.tsx, DiagnosisTabRows.tsx, DiagnosisTypeSidePanel.tsx, DiagnosisNameSidePanel.tsx, DiagnosisSettingsSidePanels.tsx, DiagnosisDeleteDialogs.tsx | K1,K3,K4 | — |
| INV-P08 | トリミング設定（course/option タブ） | tabs+list+sidepanel | `/settings/trimming` | TrimmingSettings.tsx, TrimmingCourseSidePanel.tsx, TrimmingOptionSidePanel.tsx | K1,K3,K4 | — |
| INV-P09 | トリミングコース種別 | list+sidepanel | `/settings/trimming-course-type` | TrimmingCourseTypeSettings.tsx | K1,K3,K4 | — |
| INV-P10 | 動物種 | list+sidepanel | `/settings/animal-species` | AnimalSpeciesSettings.tsx, AnimalSpeciesSidePanel.tsx, AnimalSpeciesSortableTable.tsx | K1,K3,K4 | — |
| INV-P11 | 薬剤 | list+sidepanel | `/settings/medicine` | MedicineSettings.tsx, MedicineTable*.tsx, MedicineSidePanel*.tsx, MedicineDoseParamsEditor.tsx, MedicineDeleteDialog.tsx | K1,K3,K4 | — |
| INV-P12 | 予約種別設定（グループ表+ツリー+サイドパネル群） | list+sidepanel | `/settings/reservation-type` | ReservationTypeSettings.tsx, ReservationTypeSettingsContent.tsx, ReservationTypeGroupedTable*.tsx, ReservationTypeTree.tsx, ReservationTypeSidePanel.tsx, ReservationTypeGroupSidePanel.tsx, ReservationTypeSettingsSidePanels.tsx, ReservationTypeDeleteDialogs.tsx | K1,K3,K4 | — |
| INV-P13 | 予約可能枠（職種別カレンダー） | nested | `/settings/reservation-type` 内 | ReservationTypeAvailableSlotsSection.tsx, ReservationTypeAvailableSlotsCalendar.tsx | K1,K2 | — |
| INV-P14 | 予約不可時間 | nested | 同上 | ReservationTypeUnavailableTimesSection.tsx | K1,K2 | — |
| INV-P15 | 予約種別×職種 ネスト編集 | nested | 同上 | ReservationTypeOccupationsSection.tsx | K1,K2 | — |
| INV-P16 | 入院設定 | list+sidepanel | `/settings/hospitalization` | HospitalizationSettings.tsx, HospitalizationSidePanel.tsx | K1,K3,K4 | — |
| INV-P17 | ケージ | list+sidepanel | `/settings/cage` | CageSettings.tsx, CageSidePanel.tsx, CageSortableTable.tsx, CageRowOverlay.tsx | K1,K3,K4 | — |
| INV-P18 | 商品 | list+sidepanel | `/settings/merchandise-items` | MerchandiseItemSettings.tsx, MerchandiseSidePanel.tsx, MerchandiseSortableTable.tsx, MerchandiseRowOverlay.tsx | K1,K3,K4 | — |
| INV-P19 | 保険 | list+sidepanel | `/settings/insurance` | InsuranceSettings.tsx, InsuranceSidePanel.tsx | K1,K3,K4 | — |
| INV-P20 | 職種 | list+sidepanel | `/settings/occupations` | OccupationSettings.tsx, OccupationSidePanel.tsx | K1,K3,K4 | — |
| INV-P21 | 権限グループ（ルール表含む） | list+sidepanel+nested | `/settings/permission-groups` | PermissionGroupSettings.tsx, PermissionGroupSidePanel.tsx, PermissionGroupSortableTable.tsx, PermissionRuleTable*.tsx | K1,K3,K4,K5 | — |
| INV-P22 | 問診テンプレート | list+sidepanel | `/settings/inquiry-templates`（＋`/settings/interview/templates` 同一面） | InterviewTemplateSettings.tsx, InterviewTemplateSidePanel.tsx | K1,K3,K4 | EX-04 重複統合 |
| INV-P23 | 主訴種別 | list+sidepanel | `/settings/interview/chief-complaint` | ChiefComplaintSettings.tsx, ChiefComplaintSidePanel.tsx | K1,K3,K4 | — |
| INV-P24 | 支払方法 | list+sidepanel | `/settings/payment-methods` | PaymentMethodSettings.tsx, PaymentMethodSidePanel.tsx | K1,K3,K4 | — |
| INV-P25 | キャンペーン | list+sidepanel | `/settings/campaigns` | CampaignSettings.tsx, CampaignSidePanel.tsx | K1,K3,K4 | — |
| INV-P26 | Lab項目マスタ | list+sidepanel | `/settings/lab-device-item-masters` | LabDeviceItemMasterSettings.tsx, LabDeviceItemMasterSidePanel.tsx | K1,K3,K4 | — |

※ MasterCRUDPage/MasterListPage/MasterPageShell/MasterTabPage はページ共通シェルの内部部品であり、独立面としては除外（EX-07）。各P行に包含。

### 4.18 医院・締め設定 `/settings/clinic`, `/settings/closing-time`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-D01 | 医院設定（一覧+サイドパネル+削除確認） | list+sidepanel | `/settings/clinic` | clinic-settings/routes/ClinicMasterSettings.tsx, ClinicMasterList.tsx, ClinicMasterSidePanel*.tsx, ClinicMasterSettingsPanels.tsx, ClinicDeleteDialog.tsx | K1,K3,K4 | — |
| INV-D02 | 法人請求パネル | nested | `/settings/clinic` 内 | CompanyInvoiceSection.tsx | K1,K2 | — |
| INV-D03 | 締め時間設定ページ | page | `/settings/closing-time` | closing-settings/routes/ClosingSettingsPage.tsx | K1,K3 | — |
| INV-D04 | 標準締め時間 | nested inline | 同上 | StandardClosingTimeSection.tsx | K1,K2 | — |
| INV-D05 | 休日セクション | nested inline | 同上 | HolidaySection.tsx | K1,K2 | — |
| INV-D06 | 特別期間インライン編集 | nested inline | 同上 | SpecialPeriodSection.tsx | K1,K2 | — |

### 4.19 LINE予約管理 `/line-reservation*`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-N01 | LINE予約管理設定 | form | `/line-reservation` および `/line-reservation/settings`（同一コンポーネント、EX-03 統合） | features/line-reservation/routes/LineReservationSettings.tsx | K1,K3,K6 | 設定画面未観測 |
| INV-N02 | ページエディタ | editor | `/line-reservation/page-editor` | routes/LineReservationPageEditor.tsx | K1,K3 | — |
| INV-N03 | LINE予約枠設定 | list | `/line-reservation/slots` | features/master/routes/LineReservationSlotsSettings.tsx（ResourceMasterReservationType ガード） | K1,K2 | — |
| INV-N04 | 連携済みLINE顧客（飼主詳細内） | nested | `/owners/:id` | features/line-reservation/components/LinkedLineCustomers.tsx | K1,K5 | — |
| INV-N05 | 連携検索ダイアログ LinkSearchDialog | dialog | `/owners/:id` | LinkedLineCustomers.tsx 内 | K1,K4 | — |

### 4.20 Lステップ `/lstep/*`, `/settings/lstep/*`, `/settings/integrations/lstep`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-X01 | 健診同期ページ | page | `/lstep/checkup-sync` | lstep/routes/CheckupSyncPage.tsx | K1,K2 | VQ-009/EMR-385 |
| INV-X02 | 同期フィルタフォーム | nested form | 同上 | CheckupSyncFilterForm.tsx | K1,K2 | VQ-009 |
| INV-X03 | 同期プレビュー表 | nested | 同上 | CheckupSyncPreviewTable.tsx | K1,K2 | — |
| INV-X04 | 同期確認ダイアログ | dialog | 同上 | CheckupSyncConfirmDialog.tsx | K1,K4 | — |
| INV-X05 | 配信モニタ | list | `/lstep/delivery-monitor` | routes/LstepDeliveryMonitorPage.tsx | K1,K2 | VQ-009/EMR-385 |
| INV-X06 | 分析＋CSVインポート | page+nested | `/lstep/analytics` | routes/LstepAnalyticsPage.tsx, LstepAnalyticsSections.tsx, LstepCsvImportSection.tsx | K1,K2 | — |
| INV-X07 | タグ管理 | list | `/settings/lstep/tags` | routes/LstepTagManagementPage.tsx, TagSummaryTable.tsx | K1,K2 | — |
| INV-X08 | タグ飼主ドロワー | drawer | 同上 | TagOwnerListDrawer.tsx | K1,K4 | — |
| INV-X09 | タグ一括除去ダイアログ | dialog | 同上 | BulkTagRemoveDialog.tsx | K1,K4 | — |
| INV-X10 | 連携設定 | form | `/settings/integrations/lstep` | features/settings/routes/LstepSettingsPage.tsx, LstepSettingsForm.tsx, LstepSettingsStatus.tsx, LstepSettingsActions.tsx | K1,K3 | — |
| INV-X11 | タグ設定・コードマッピング | nested | 同上 | LstepTagConfigSection.tsx, LstepTagCodeMappingsSection.tsx | K1,K2 | — |
| INV-X12 | トリガー優先度 | nested | 同上 | TriggerPrioritySection.tsx | K1,K2 | — |

### 4.21 同一飼主・ペット連携 `/identity-links`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-Y01 | 連携ワークベンチ | page | `/identity-links` | identity-links/routes/IdentityLinksPage.tsx | K1,K5 | — |
| INV-Y02 | 飼主リンクセクション | nested | 同上 | components/OwnerLinkSection.tsx | K1,K2 | — |
| INV-Y03 | ペットリンクセクション＋履歴 | nested | 同上 | components/PetLinkSection.tsx | K1,K2 | — |
| INV-Y04 | 閲覧のみロール表示 | state | 同上 | IdentityLinksPage.tsx（canEdit=false 分岐、`canView` 非保持時 `/` へ Navigate） | K1,K5 | 別ロール未観測 |

### 4.22 サポート `/settings/bug-reports`, `/settings/chat-history`, 全画面オーバーレイ

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-U01 | サポートFAB | overlay | 全認証画面 | features/support/components/SupportWidget.tsx, app/routes/app-routes.tsx（AuthenticatedSupportWidget） | K1,K4 | VQ-011/EMR-387 |
| INV-U02 | サポートパネル | panel | 同上 | components/SupportPanel.tsx | K1,K4 | VQ-011 |
| INV-U03 | バグ報告タブ（送信フォーム） | nested form | 同上 | components/BugReportTab.tsx | K1,K3 | — |
| INV-U04 | ヘルプチャット | nested | 同上 | components/HelpChat.tsx, HelpChatHistoryProvider.tsx | K1,K2 | — |
| INV-U05 | マニュアルヘルプタブ | nested | 同上 | components/ManualHelpTab.tsx | K1,K2 | — |
| INV-U06 | バグ報告一覧 | list | `/settings/bug-reports` | routes/BugReportsPage.tsx | K1,K2 | — |
| INV-U07 | バグ報告詳細ダイアログ（ステータス編集・canMutate分岐） | dialog | 同上 | BugReportsPage.tsx 内 BugReportDetailDialog | K1,K3,K4,K5 | — |
| INV-U08 | バグ報告削除確認 | dialog | 同上 | BugReportsPage.tsx（ConfirmDialog） | K1,K4 | — |
| INV-U09 | チャット履歴一覧 | list | `/settings/chat-history` | routes/ChatHistoryPage.tsx | K1,K2 | — |

### 4.23 マニュアル `/manual*`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-W01 | マニュアル閲覧（サイドバー/ドロワー/検索） | page | `/manual`, `/manual/:category/:slug` | manual/routes/ManualPage.tsx, ManualPageChrome.tsx, ManualContent.tsx | K1 | 記事本文はドキュメント（EX-05） |
| INV-W02 | マニュアルエディタ | edit | 同上（ResourceManualEdit 権限で到達可） | components/ManualEditor.tsx | K1,K3,K5 | 到達確認済のため対象 |

### 4.24 認証・アプリ共通 `/login` 等

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-Z01 | ログイン | form | `/login` | features/auth（Login）, app/routes/app-routes.tsx | K1 | — |
| INV-Z02 | パスワード再設定要求 | form | `/forgot-password` | ForgotPasswordPage | K1 | — |
| INV-Z03 | パスワードリセット | form | `/reset-password` | ResetPasswordPage | K1 | — |
| INV-Z04 | 404 | page | `*` | app-routes.tsx notFoundRoute | K1 | — |
| INV-Z05 | 共通トースト（アプリ全域オーバーレイ） | overlay | 全画面 | components/ui/sonner.tsx（top-left 指定）, app/provider.tsx | K1,K4 | VQ-005/EMR-381 |

### 4.25 LIFF 別アプリ `frontend/liff`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-J01 | LIFFアプリシェル・ローディング | shell | LIFF URL | frontend/liff/src/App.tsx, pages/LoadingPage.tsx | K1,K6 | — |
| INV-J02 | LINE連携ページ | page states | LIFF URL | frontend/liff/src/pages/LiffLinkPage.tsx（状態: loading/linking/success/conflict/expired/error） | K1,K6 | 自動連携結果のみでユーザーフォームなし。親スコープの登録トークン状態は本画面への包含が **要確認（disputed）** として保留 |
| INV-J03 | ペット健康ページ | page states | LIFF URL | frontend/liff/src/pages/PetHealthPage.tsx（空ペット/成功一覧/エラー/clinic_id欠落） | K1,K2,K6 | 非空状態未観測 |

### 4.26 LINE予約別アプリ `frontend/line-reserve`

| ID | 面 | モード | ルート/入口 | ソース | 理由 | 関連 |
|---|---|---|---|---|---|---|
| INV-K01 | アプリシェル＋内部状態機械 | shell | line-reserve URL | frontend/line-reserve/src/App.tsx, app-page-switch.tsx | K1,K6,K7 | — |
| INV-K02 | トップ | step | 同上 | pages/TopPage.tsx | K1,K6 | — |
| INV-K03 | 顧客情報 | step form | 同上 | pages/CustomerInfoPage.tsx | K1,K3,K6 | — |
| INV-K04 | スタッフ選択 | step | 同上 | pages/StaffSelectPage.tsx | K1,K6 | — |
| INV-K05 | 日付選択 | step | 同上 | pages/DateSelectPage.tsx | K1,K6 | — |
| INV-K06 | 時間選択 | step | 同上 | pages/TimeSelectPage.tsx | K1,K6 | — |
| INV-K07 | コース選択（一般分岐） | step | 同上 | pages/CourseSelectPage.tsx | K1,K6,K7 | — |
| INV-K08 | トリミングコース選択 | step branch | 同上 | pages/TrimmingCourseSelectPage.tsx | K1,K6,K7 | — |
| INV-K09 | トリミングオプション選択 | step branch | 同上 | pages/TrimmingOptionSelectPage.tsx | K1,K6,K7 | — |
| INV-K10 | 確認 | step | 同上 | pages/ConfirmPage.tsx | K1,K3,K6 | — |
| INV-K11 | 完了 | state | 同上 | pages/CompletePage.tsx | K1,K6 | — |
| INV-K12 | メンテナンス | state | 同上 | pages/MaintenancePage.tsx | K1,K6 | — |
| INV-K13 | 予約履歴・既存予約 | list | 同上 | pages/MyReservationsPage.tsx | K1,K2,K6 | — |
| INV-K14 | リクエストフロー | step | 同上 | pages/RequestPage.tsx | K1,K6,K7 | — |

## 5. 除外表（明示）

| ID | 対象 | 種別 | 理由 |
|---|---|---|---|
| EX-01 | `/checkups/:id` | 不存在ルート | clinical-care-routes.tsx に index/select-pet/new のみ。健診編集は INV-C04（カルテ内 CheckupsTab）のみ |
| EX-02 | `/settings/job-title`→`occupations`、`service-type`→`reservation-type`、`inquiry-template`→`inquiry-templates`、`shift-template`→`shift-templates` | リダイレクト | `Navigate` のみ。独立面なし（settings-routes.tsx） |
| EX-03 | `/settings/diagnosis-type`→`diagnosis?tab=diagnosis_type`、`diagnosis-name`→`?tab=diagnosis_name`、`trimming-course`→`trimming?tab=course`、`trimming-option`→`?tab=option`、`examination`/`vaccine`/`consultation`/`procedure`→`treatment-items?tab=*` | リダイレクト（タブエイリアス） | config/paths.ts `withTab`。タブ面は INV-P04/P07/P08 に統合、重複登録なし |
| EX-04 | `/settings/interview/templates` | エイリアス | `/settings/inquiry-templates` と同一 InterviewTemplateSettings → INV-P22 へ統合 |
| EX-05 | `/manual/:category/:slug` 記事本文 | ドキュメントのみ | フォームなし。閲覧面は INV-W01、編集は INV-W02 で包括 |
| EX-06 | SessionPending 等の遷移中表示 | 遷移状態 | 画面ではなく認証待ち表示 |
| EX-07 | MasterCRUDPage/MasterListPage/MasterPageShell/MasterTabPage | 内部共通部品 | ページシェル。各マスタ行に包含、独立面ではない |
| EX-08 | 印刷専用エリアの重複登録 | 印刷バリアント | MedicalRecordPrintView/ExaminationPrintArea/AccountingDocument/DailyAccountingPrintArea/ClosePrintArea/MonthlyReportPrintArea は本体行内で言及済み（INV-M23,E09,B10,G01,Q01） |
| EX-09 | `/line-reservation` index | エイリアス | `/line-reservation/settings` と同一 LineReservationSettings → INV-N01 へ統合 |
| EX-10 | 説明のみ・フォームを持たない詳細系のうち上記に該当しないもの | — | 現行ルート上でフォーム/一覧を持たない独立詳細面は確認されず。一覧を含むレポート（`/owners/:id/report`）は INV-O14 として対象に含めた |

## 6. VQ → 子チケット → ソース対象 対応表

| VQ | 子 | 所見（要約） | 確認済ソース対象 | 実装状態 |
|---|---|---|---|---|
| VQ-001 | EMR-377 | 受付カンバン右端列が1568pxで見切れ | INV-A01: reception/routes/ReceptionPagePanels.tsx（列グリッド） | 未着手（別タスク） |
| VQ-002 | EMR-378 | 一覧テーブル列幅不足・縦積み（owners/trimming/vaccinations/inventory/staff） | INV-O01,T01,V01,I01,P02: OwnersList/TrimmingList/VaccinationList/InventoryList/StaffSettings 各テーブル | 未着手 |
| VQ-003 | EMR-379 | カルテ一覧1字縦積み・placeholder切断・768px右端クリップ | INV-M01: features/medical-records/routes/MedicalRecords.tsx | 未着手 |
| VQ-004 | EMR-380 | カルテ登録/編集 右パネル折返し・ヘッダ見切れ | INV-M03..M05,M16: MedicalRecordForm, InterviewHistory.tsx（右パネル「治療履歴」見出し）, MedicalRecordStickyHeader.tsx + components/shared/PatientContextHeader/PatientContextHeader.tsx（担当医右端） | 未着手 |
| VQ-005 | EMR-381 | トースト左上固定でナビ/ヘッダに重なる | INV-Z05: components/ui/sonner.tsx, app/provider.tsx（再現面: INV-L01,E04,M04） | 未着手 |
| VQ-006 | EMR-382 | 検査一覧 日付「2026-08-/25」折返し | INV-E01: examinations/routes/ExaminationsList.tsx | 未着手 |
| VQ-007 | EMR-383 | 詳細の権限外/取得失敗が空白+赤文字断片 | INV-E04,S04: ExaminationForm, EstimateDetail（同型: M04,T04,V04,I03,B04 要確認） | 未着手 |
| VQ-008 | EMR-384 | 会計明細ヘッダ全列1-2字縦積み | INV-B03..B05: AccountingDetailPage, ItemListCard, AccountingItemRow | 未着手 |
| VQ-009 | EMR-385 | Lステップ セレクト切断・ラベル折返し | INV-X01,X02,X05: CheckupSyncPage, CheckupSyncFilterForm, LstepDeliveryMonitorPage | 未着手 |
| VQ-010 | EMR-386 | スタッフ新規パネル名「無題」 | INV-P03: StaffSidePanel.tsx | 未着手 |
| VQ-011 | EMR-387 | 右下FABが操作要素と重なる | INV-U01,U02: SupportWidget, SupportPanel（被影響面: INV-R03,P03 等モーダル/最下行） | 未着手 |
| VQ-012 | EMR-388 | チップ省略・医院名2ピル視認 | INV-Q02,O01: AggregationDashboardPage（CPMチップ）, OwnersList（医院フィルタチップ） | 未着手 |

※「確認済ソース対象」は静的確認の写像であり、現段階での視覚的真因確定を意味しない。実装は別タスク/別diff。

## 7. ブラウザ再検証チェックリスト（後工程計画。本書はCSS/実行証跡を保証しない）

- [ ] 4ビューポート全走査: `1568x873` / `1280x800` / `768x1024` / `390x844`
- [ ] 長文・長名（飼主名・ペット名・医院名・スタッフ名・予約種別名・タグ名）での折返し/省略
- [ ] 死亡ペット状態（select-pet・カルテ・予約・各編集面）
- [ ] 権限状態（viewのみ/作成可/編集可/削除可の各ロール。identity-links・manual・各マスタ・PetEditModal 内セクション）
- [ ] データゼロ（空一覧・空タブ・空ギャラリー・初回フォーム）と編集モードの空初期値
- [ ] テーブル内スクロール・横スクロール境界・固定列/ヘッダの重なり
- [ ] ソート/フィルタ/ページネーション/リンク遷移（一覧各行・タブ切替・`?pet=` `?tab=` `?checkupId=` `?date=` `?days=` 系）
- [ ] エラー状態（取得失敗・他医院スコープID・権限外・送信失敗）と安全な戻る導線（VQ-007 型）
- [ ] モーダル/ドロワー/サイドパネルのレイヤリング・FAB/トーストとの重なり（VQ-005,VQ-011 型）
- [ ] フォーカス遷移（ダイアログ開閉・フォーム送信後・Esc/Tab）
- [ ] tentative/次回来院系（期限超過パネル・確定・リセット・終了）の状態表示
- [ ] LIFF: トークン有効/期限切れ/競合、PetHealth 空/非空、clinic_id 欠落
- [ ] line-reserve: 一般/トリミング両分岐の全ステップ、履歴、メンテナンス、リクエスト
- [ ] LINE予約管理（設定/ページエディタ/スロット/連携顧客）の状態・権限表示

## 8. 免責

- 本書はソース上の到達可能性と面の分類のみを記録する。CSS・ランタイム挙動・スクリーンショットの合否は証明しない。
- 個別画面の撮影・レビュー状態は付随レポートおよび Plane コメントに記録される（Lead 管理）。本書からレンダリング PASS へ継承しない。
- 新規コード変更・ブラウザ受入・最終合否は別タスク/別diff。

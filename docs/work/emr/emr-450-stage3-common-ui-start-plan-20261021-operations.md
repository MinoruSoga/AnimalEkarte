# EMR-450 第③段階 非検索操作family・受領条件

要件判断の窓口はMinoru Soga。既決仕様準拠・重複工程削減・計画coverage指標と未測定の業務成果の区別は[共通の要件検証](emr-445-stage1-contract-start-plan-20261021.md#requirement-validation-20261021)へ従う。実装担当claim・開始許可・実績測定は別。試験NOT_RUN。

[検索・共通責務・PR単位](emr-450-stage3-common-ui-start-plan-20261021.md)と合わせて適用する。各ownerに同じ5操作が存在するとは仮定せず、実在する保存・確認・保持・戻る・結果照会の接点と将来受領を区別する。全repository writerの網羅監査や実機合格を宣言する文書ではない。

固定code ref: a88fa4add58d43cff643590f9f507c8ab1a3c085。O1実行、O2 clear/resetと永続復元、O3保持/戻る、O4閉じると取消/削除、O5失敗/再試行。E=既存接点、F=④owner受領、N=当該操作familyに非適用、U=受入未確認、G=具体的な③最小接続差。

共通SubmitButton/useFormStatus、ConfirmDialog、NavigationBlocker、FormDialogを再利用する。O2を全フォームに新設せず、検索clear・未保存破棄・履歴付き永続復元を区別する。

## D01 予約登録・更新／受付登録／閉じる

担当境界: 468/460

実接点: [frontend/src/features/reservations/hooks/use-reservation-save-actions.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/reservations/hooks/use-reservation-save-actions.ts) blob b132df719632ca4922c1c85db7841b8daa59d9f0

handler: handleSave L94,122,179,303,368, handleCloseCreateForm L89,163,173,217,227, resetCreateProgress L84,90,92,369

状態/共通接点: ReservationFormModal.onSave → handleSave。error文字列/null、createProgress、savingReceptionRef。閉じるはローカルcreate進捗初期化で取消writerではない

適用: O1:E 保存起動; O2:N 永続復元なし; O3:E modal入力; O4:E 閉じると予約取消別; O5:U 下位intent照会は個別契約

具体差分/無修正根拠: 共有フォームの保存/閉じる接点を再利用。予約operation照会・warning_changedは468/460の既存intent adapter受領で、③に新writerを作らない

④受領subset: appointment ID/version、同scope operation key/fingerprint、warning_changed、receipt/current-state。lookupがunknownなら元key保持し新createしない

## D02 受付status移動・予約編集・取消確認

担当境界: 452/476

実接点: [frontend/src/features/reception/hooks/use-reception-modal-handlers.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/reception/hooks/use-reception-modal-handlers.ts) blob 981e0f2a77f5ff1ba42fd487614841490498af5b

handler: handleAdvanceStatus L151,328, handleEditSave L191,332, handleCancelAppointment L288,329

状態/共通接点: Reception → useReceptionModalHandlers → ReceptionPageBody ConfirmDialog。useReceptionKanbanはisUpdatingStatus/columns snapshotを所有

適用: O1:E 明示status; O2:N board resetは永続Undoでない; O3:E modal; O4:E 取消確認と閉じる分離; O5:U 状態失敗

具体差分/無修正根拠: ReceptionPagePanelsのonCloseCancelConfirmとonConfirmCancelを維持。prepare/read/startや元context照会は452/476受領、shared dialogを受付業務機械へ拡張しない

④受領subset: 専用owner/kind/id/revision、snapshot total/全件性、元context operation結果。unknownを空board/受付成功にしない

## D03 通常保存・診療終了・削除・タブ復元

担当境界: 452/494/495/477

実接点: [frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts) blob a34befcf39b5c3b835119b6df503a64f79a3d9f5

実操作別handler（保存hookへ一括帰属させない）:

| 操作 | file / handler | 限定責務 |
|---|---|---|
| 通常保存 | 上記use-medical-record-save-action.tsのuseMedicalRecordSaveAction L121、formAction L207/335 | 通常saveのpending/catchと成功subsetを確認する |
| 診療終了/確定 | [hooks/use-medical-record-quick-patch-actions.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/medical-records/hooks/use-medical-record-quick-patch-actions.ts#L156) のhandleFinalize→mutateAsync（L156–175） | finalizeの確認/pending・失敗/結果を別に追跡。旧status patchの実在を最新D03終了契約の全実装済み証拠にせず494/495の④受領へつなぐ |
| カルテ削除 | [routes/MedicalRecordForm.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/medical-records/routes/MedicalRecordForm.tsx#L40) のuseDeleteMedicalRecord/handleDeleteConfirm→deleteRecord（L40–53） | canDelete/current pet guard、isDeleting、onSuccessの遷移、onError保持を保存成功と分ける |

状態/共通接点: MedicalRecordFormActions SubmitButton/isSaving と finalize ConfirmDialog/isFinalizing。保存catchはsuccess false

適用: O1:E 通常保存/終了; O2:F 履歴付きtab復元; O3:E dirty保護; O4:E delete確認; O5:F 復元receipt照会

具体差分/無修正根拠: 通常保存と終了Dialogは再利用。tab clear/copy undoをローカルsetStateに流用せず495の復元consumerとして④受領。原文reader477はwrite N

④受領subset: record/pet/clinic、元operation、snapshot、成功subset、expected/current version、restore receipt。固定結果/資金/送信を逆実行しない

## D04 飼主＋pet登録／基本保存／pet削除・主飼主変更確認

担当境界: 451/490/491

実接点: [frontend/src/features/owners/routes/OwnerForm.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/owners/routes/OwnerForm.tsx) blob 9d3c6c10a50af7715ed88f2d28e1f018366237b4

handler: formAction L109,301,304, handleConfirmDeletePet L268,381, handleConfirmOwnerChange L248,392, handleBack L242,307,311

状態/共通接点: OwnerForm → useOwnerForm formState/isLoading。FormHeaderActions + NavigationBlocker + ConfirmDialog、usePostalCodeLookupは別用途

適用: O1:E 保存; O2:N 永続一括clearなし; O3:E NavigationBlocker; O4:E pet確認; O5:F atomic registration receipt

具体差分/無修正根拠: 既存Header/dirty/confirmは再利用。490選定wireへ登録結果照会を接続する後続仕事を具体受領し、メール電話一致を成功receipt代わりにしない

④受領subset: 490/491 selected wire: owner.basic_revision.save、atomic owner+pet、VerifiedCommand/ContextGate、common/basic revision、registration receipt。scopeとcontact/candidate/use分離

## D05 入院フォーム保存・戻る／日次追記／退院

担当境界: 499/461

実接点: [frontend/src/features/hospitalization/hooks/use-hospitalization-form.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/hospitalization/hooks/use-hospitalization-form.ts) blob 661c730aab0fc7d06915cadc6d57066963871652

handler: formAction L101,220, handleFormDataChange L54,222

状態/共通接点: useHospitalizationForm isSaving/formState → FormBody。useHospitalizationFormChrome.handleBack/handleFormChange がdirty、日次FormDialog別

適用: O1:E フォーム/日次; O2:N 汎用復元なし; O3:E dirty; O4:E 閉じる; O5:F 退院・交換結果照会

具体差分/無修正根拠: フォーム保存と日次独立FormDialogを再利用。退院/ケージ交換を共有saveへまとめない。日次FormDialogのpendingとonCloseはconsumer責務で保持

④受領subset: stay/hospitalization/day/care item refs、cage expected allocations、退院state+billing receipt。取得不明を空きへしない

## D06 施術途中保存・削除・戻る／終了

担当境界: 480/482

実接点: [frontend/src/features/trimming/hooks/use-trimming-form.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/trimming/hooks/use-trimming-form.ts) blob 43e6ede4e1c2026d1040b89ac1f1c6dc3a151a0a

handler: formAction L174,291, submitTrimmingMutation L162,188,202,244, handleDelete L266,293

状態/共通接点: TrimmingFormPanels formAction/NavigationBlocker/ConfirmDialog、isSaving/isDeleting。フォーム保存と画像cleanupの責務別

適用: O1:E 保存; O2:N 永続復元なし; O3:E dirty; O4:E 削除確認; O5:F 専用終了receipt

具体差分/無修正根拠: shared footer/dirty/confirm再利用。途中保存の担当/コース未定と終了実績条件は480/482で別判定、③で終了writerを書かない

④受領subset: appointment ID/version、typed trimming record ref、upload保存済key、終了receipt、実担当/記録済施術、会計pull差分

## D07 検査保存・unconfirm・削除・選択印刷／受信attach

担当境界: 516/501

実接点: [frontend/src/features/examinations/routes/ExaminationForm.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/examinations/routes/ExaminationForm.tsx) blob a9a8a25314de5b24a6bc2fe8db34933d0f8fda76

handler: formAction L71,332, handleDeleteConfirm L197,394, retryRead L79,251,252, handleHistoryClear L177,386

状態/共通接点: useExaminationForm → runExaminationSave、NavigationBlocker、ConfirmDialog、ExaminationUnconfirmDialog、print snapshot読取

適用: O1:E 保存/印刷; O2:N 履歴filter clearは結果消去でない; O3:E dirty; O4:E delete; O5:E read retry/F attach照会

具体差分/無修正根拠: 保存失敗success falseと現在ID読取retryを再利用。受信attach/Undoはjob指紋/revisionを所有する516/501へ渡し、form save retryと共通化しない

④受領subset: job/source/specimen ref、received/persisted、exam/revision/item fixed版、usage receipt、attach/Undo operation結果

## D08 接種単体保存・削除／カルテ内一件保存

担当境界: 517/486/487

実接点: [frontend/src/features/vaccinations/hooks/use-vaccination-form.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/vaccinations/hooks/use-vaccination-form.ts) blob 3bc5939a761ef36db9b8a3726e7fabc20c92a05a

handler: formAction L109,179, runVaccinationSave L25,111, handleDelete L148,183

状態/共通接点: useVaccinationForm isSaving/isDeleting + entity-read-result、helpers createVaccinationDeleteHandler; カルテ内は別use-medical-record-vaccination-form

適用: O1:E 一接種保存; O2:N history clearは接種解除でない; O3:E 入力保持; O4:E ID指定delete; O5:F 作成不明照会

具体差分/無修正根拠: 既存一件保存と外側カルテ保存の分離を再利用。unknown時新createを許さない操作receiptは517/486受領、単体finalizedは新設しない

④受領subset: vaccination ID/version、pet/record归属、success timestamp、同日複数と同操作replay区別、pointer null/current version

## D09 独立健診の親→子→field保存・戻る

担当境界: 500

実接点: [frontend/src/features/checkups/hooks/use-checkup-form.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/checkups/hooks/use-checkup-form.ts) blob aff9d89db48855d7f47f2691aeb1af795d757d69

handler: formAction L84,185, createMedicalRecordForCheckup L14,109, createCheckupOnMedicalRecord L13,122, replaceCheckupFieldResults L17,136

状態/共通接点: CheckupForm FormHeaderActions submitDisabled=isPending、hook 108–139は順次await、catch保持前段IDを露出しない。routeはNavigationBlocker未import

適用: O1:E 現保存; O2:N type変更setFieldValues({})は永続消去でない; O3:G 独立フォームdirty接続なし; O4:N deleteなし; O5:F 段階再開

具体差分/無修正根拠: ③最小hunk: CheckupForm入力変更→useUnsavedChanges/NavigationBlocker、成功のみdirty解除、戻る破棄確認。500受領: hookの親/子成功ref保持と再開、これは新業務writer契約なので④。次回日既定は共通UI修正に混ぜない

④受領subset: medical_record_id→checkup_id→fields refs、成功段階、result false/0/未入力、[]明示消去、元operation lookup。後段失敗で前段削除しない

## D10 会計確定・カード訂正・戻る

担当境界: 481/483/505/506/507

実接点: [frontend/src/features/accounting/hooks/use-accounting-completion-action.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/accounting/hooks/use-accounting-completion-action.ts) blob b1b139c308e707457e799996b34a807dd6122384

handler: formAction L250,502, focusAccountingCompletionError L81,485

状態/共通接点: isPending/formState、unbilledConflict。CreditCorrectionDialog formAction/SubmitButton別。版競合のみlocalItems破棄し明示banner

適用: O1:E 確定/訂正; O2:N 汎用clearで資金逆実行なし; O3:E 入力; O4:E dialog close; O5:U original financial receipt

具体差分/無修正根拠: 既存conflictバナー・focus・SubmitButtonを再利用。入力保持原則を理由に意図的なstale明細破棄を戻さない。元資金再照会は505–507専用受領

④受領subset: billing/current version、original payment/source allocation、receipt、operation intent。受領unknown/部分実受領を単純失敗再POSTにしない

## D11 見積保存・削除・後継作成

担当境界: 508

実接点: [frontend/src/features/estimates/routes/EstimateDetail.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/estimates/routes/EstimateDetail.tsx) blob 29438457ec22b4b0de80f62697f794473205bf3d

handler: handleDelete L64,207, successorFormAction L86,226, handleCloseSuccessorDialog L79,225

状態/共通接点: EstimateDetail ConfirmDialog/isDeletePending、successorBusy、entityRead.retry、useEstimateForm.handleCancel

適用: O1:E 保存/後継; O2:N 原本復元なし; O3:E form dirty; O4:E delete/close別; O5:F successor operation lookup

具体差分/無修正根拠: 現在後継dialogは再利用。508の原本不変/後継新ID/lookupを受領し、読取0件で失敗認定しない。後継API接続を③で再実装しない

④受領subset: original estimate ID/version、successor ID、operation ID/result、current parent state、line totals。原本不変/confirmed親維持

## D12 締め確認・実査入力・確認取消

担当境界: 509

実接点: [frontend/src/features/cash-register/routes/CashRegisterClosePage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/cash-register/routes/CashRegisterClosePage.tsx) blob 92b5fcabf0335859190ae2ae1ec7202718d19b2b

handler: formAction L73,179, handleConfirmClose L84,198, handleCancelConfirm L115,194

状態/共通接点: pendingFormData/showConfirm/actualCash。successのみsetActualCash空、catchはconfirmを閉じ値は残す

適用: O1:E 締め; O2:N 原締めclearなし; O3:E 実査保持; O4:E 確認閉じるは締め取消でない; O5:F closing operation lookup

具体差分/無修正根拠: 既存確認段階/成功のみclearを再利用。509元operation結果照会・追記訂正を受領し、retryで原snapshot再作成しない

④受領subset: clinic/day/AM-PM-EMG、original closing snapshot、operation ID、実査/理論/差額メモ、append correction receipt

## D13 在庫品目保存／実物調整

担当境界: 510

実接点: [frontend/src/features/inventory/hooks/use-inventory-form.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/inventory/hooks/use-inventory-form.ts) blob f14c039c64ffe56ede8a6c11bf053afc5694ab4f

handler: formAction L101,165

状態/共通接点: InventoryForm NavigationBlocker + useInventoryForm formState/isPending。現formはcreate/update品目

適用: O1:E 品目/F 調整; O2:N 在庫巻戻しなし; O3:E dirty; O4:E 戻る; O5:F 調整operation lookup

具体差分/無修正根拠: 品目入力sharedフォーム再利用。quantity調整の元操作参照/期待版/正整数は510専用consumerへ。フォームupdateを棚卸の旧絶対値writeに流用しない

④受領subset: inventory item/base/movement event refs、expected revision、delta/real occurred time/reason、original operation、result lookup

## D14 勤務保存・template適用・削除・結果再照会

担当境界: 448/493/514

実接点: [frontend/src/features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx) blob 1a19249c85dbd843e777750dcafd4898695a3d00

handler: formAction L174,248,463,466, handleApplyTemplate L515,663, handleDeleteConfirm L564,1016

状態/共通接点: dispatchFormAction/isPending/pendingOperation、operation_key、照会、ConfirmDialog/SubmitButton、close disabled while pending

適用: O1:E 保存; O2:E template入力コピー/N 既存勤務自動変更; O3:E pending保持; O4:E deleteとclose別; O5:E pendingOperation照会

具体差分/無修正根拠: 既存operation_key/lookupとpending close guardを再利用。G2 template一覧状態だけ別最小差分。CSV/copyは448統合・493 staff採択契約へ受領。514既存日時CRUDは再実装しない

④受領subset: shift/work-break IDs、expected version、operation_key/pendingOperation、scope/date/staff、CSV/copyのbulk preview confirmed setは448統合・493 staff採択契約へ受領。514既存日時CRUDは再実装しない

## D15 CSV upload・履歴読取／手動LINE送信

担当境界: 504

実接点: [frontend/src/features/lstep/components/LstepCsvImportSection.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/lstep/components/LstepCsvImportSection.tsx) blob bdeb99f5f275b6b0d45edb6079aa28fdfd06ea43

handler: CsvUploadSection.formAction L45,62, onReset L62, CsvImportHistoryTable L32,116

状態/共通接点: SubmitButton/isPending、UploadState、onReset setSelectedFile(null)、history hook data/isLoadingのみ

適用: O1:E upload/F送信結果調停; O2:U React reset失敗保持; O3:U file保持; O4:N upload取消なし; O5:G history errorを空へ合流

具体差分/無修正根拠: ③最小hunk: CsvImportHistoryTableでisError/error/refetchを受取り空より先にErrorFallback、retry同条件20。upload失敗時file resetはuseActionStateとfile inputの仕様に沿う受入を固定し、未実機を欠陥断定しない。外部send調停504へ

④受領subset: import ID/snapshot/status、accepted vs inserted件数、send claim/log/external result、最大100 preview。unknown外部成功で再sendしない

優先順: 明示保留: Lstep/DMマーケ仕様整備はPhase1内最後。既存共通DataStates/検索契約の適用先としてのみ記録し③先行必須gateから外す。

## D16 集計条件変更・選択CSV出力

担当境界: 511

実接点: [frontend/src/features/aggregation/routes/AggregationDashboardPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/aggregation/routes/AggregationDashboardPage.tsx) blob e70d706abf05e3cca1c3a189d57db1cd6e027c76

handler: handleParamsChange L58,71,73,121, handleExportCsv L106,133, handleSelectAll L76,128, handleSelectOwner L87,129

状態/共通接点: useGetOwnerAggregations、selectedOwnerIds、buildCsvContent。read-only集計のローカルdownload

適用: O1:E CSV; O2:N 永続復元なし; O3:E 選択状態; O4:N mutation取消なし; O5:E 同条件read

具体差分/無修正根拠: CSVの対象集合選択とread-only境界を再利用。集計検索の差分はL11だけ。V2配信/履歴writerを追加しない

④受領subset: AggregationParams/page、selectedOwnerIds、会計distinct/row counts、取得失敗。CSVは現在選択page、送信claimなし

## D17 医院保存・無効化確認・法人請求情報

担当境界: 448/450

実接点: [frontend/src/features/clinic-settings/hooks/use-clinic-master-settings.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/clinic-settings/hooks/use-clinic-master-settings.ts) blob 3356f8701c8bc23b93f048f5dc39aacc88fe5cc1

handler: formAction L72,171, handleCloseEdit L124,175, handleDeleteConfirm L132,176

状態/共通接点: formDataRef/selectedItemRef、isPending/isError、side panel list status。CompanyInvoiceSection別formAction

適用: O1:E 保存; O2:N scope切替は保存でない; O3:E panel dirty; O4:E delete確認; O5:U scope切替中結果

具体差分/無修正根拠: 既存MasterSidePanel/use-side-peek-dirtyを再利用。医院scope変更で旧入力を新clinicへ書かない最小受入はselectedItemRef/current clinicで固定。法人値変更を発行済みsnapshotへ反映しない

④受領subset: clinic ID、global/company vs clinic fields、current grant、selected target revision。portal viewはwrite grant代用なし

## D18 LINE予約設定保存／本人予約確定・取消

担当境界: 474/519/520

実接点: [frontend/src/features/line-reservation/components/LineReservationSettingsForm.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/line-reservation/components/LineReservationSettingsForm.tsx) blob 47a4a3d9fd5785bfde45994b39fea6bc5598c683

handler: formAction L163,223, handleLoadLatestClosedWeekdays L115,273

状態/共通接点: SubmitButton、hasInvalidClosedWeekdays/hasClosedWeekdaysConflict、明示最新値読取。設定保存と外部稼働別

適用: O1:E 設定/F 本人確定契約; O2:N 設定clearで予約取消しない; O3:E 曜日競合保持; O4:F 本人取消; O5:F 本人操作結果

具体差分/無修正根拠: 設定Form/曜日競合/明示reloadを再利用。本人flowは別frontend/line-reserve、院内overrideを持込まない。519/520受領、③でチャネル有効化しない

④受領subset: authenticated LINE owner/pet/clinic、slot recheck、booking operation/receipt、cancel boundary、notification outcome separate

## D19 飼主・pet group link/unlink

担当境界: 451 identitylink既存owner

実接点: [frontend/src/features/identity-links/components/OwnerLinkSection.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/identity-links/components/OwnerLinkSection.tsx) blob 1d5329bc636f78de1560c27e761c2554e574a6d2

handler: onLinkOwners L17,31,83, toggleOwner L16,30,66, onUnlinkOwner L18,32,35,106, ConfirmDialog.onConfirm L104

状態/共通接点: IdentityLinksPage → useIdentityLinksWorkbench → OwnerLinkSection/PetLinkSection pending、unlinkTarget、確認後別callback

適用: O1:E link; O2:N 台帳merge解除でない; O3:E selected candidates; O4:E unlink確認; O5:U mutation hook結果照会

具体差分/無修正根拠: ConfirmDialog/候補選択を再利用。unlinkをowner deleteへ共通化しない。same-set replay・監査atomicityは既存identity ownerへ受領

④受領subset: group ID/member refs、authorized scope、same-set create、remaining one/zero group、audit receipt。NotFoundだけで初回失敗判断なし

## D20 本人LINE linkと健康read

担当境界: 504/474

実接点: [frontend/liff/src/hooks/use-liff-link.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/liff/src/hooks/use-liff-link.ts) blob fbdc06321a671cf7acd6b98550099e923c9ba207

handler: useLiffLink L13, linkLineAccount L4,59, linkPromiseRef L17,59

状態/共通接点: LiffLinkPage status loading/linking/success/conflict/expired/error。再試行はwindow.location.reload、同マウントpromiseは重複防止

適用: O1:E 認証後link; O2:N token reset不可; O3:E 状態表示; O4:N 永続取消なし; O5:F link状態lookup

具体差分/無修正根拠: staff共通SubmitButtonへ置換しない。504/474受領: reload再POST前に本人link状態照会し単回token消費を失敗証明にしない。共有Spinner/ErrorPageは再利用

④受領subset: LINE idToken、clinic/link token、本人link状態、有限公開health DTO。token済と初回成功を区別。staff grant流用なし

## D21 1頭切替・限定履歴read／専用2帳票

担当境界: 452/497

実接点: [frontend/src/features/owner-report/routes/OwnerReport.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/owner-report/routes/OwnerReport.tsx) blob bb76331c7c6b1bd21d0aa06318aa1c3a3fa20a1f

handler: handleSelectPet L46,59,158

状態/共通接点: useGetOwner/useGetOwnerReportPets/useGetPetFirstVisit、PetSwitcher、selected pet URL、RequirePermission

適用: O1:E pet切替/F専用印刷; O2:N 永続clearなし; O3:E URL context; O4:N writer cancelなし; O5:E section read

具体差分/無修正根拠: 既存read-only入口を再利用。一般印刷/共有ボタン新設なし。497専用予防薬2帳票のみ④受領、全history印刷へ広げない

④受領subset: owner/pet/clinic scope、page+section grant、source fixed result/version、partial read status、専用prescription/yearly projection DTO

## D22 記事編集保存・破棄・同operation再照会

担当境界: 502/513

実接点: [frontend/src/features/manual/components/ManualEditor.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/manual/components/ManualEditor.tsx) blob 720c0156c39fef9e5ecf9adef549e1321072b81f

handler: handleSave L327,400,404,560, handleRequeryOperation L437,656, handleCloseRequest L497,601, handleDiscardConfirm L505,695

状態/共通接点: source.phase/saveStopped/lastOperationKey/resolvingOperation/進行中ref、ConfirmDialog、getManualEditOperationResult

適用: O1:E 保存; O2:N 破棄は旧公開版復元でない; O3:E dirty維持; O4:E discard確認; O5:E 同keylookup

具体差分/無修正根拠: 再実装不要。handleSaveの二重key抑止、unknown入力保持、receipt照合後current readを共有UIの適合例として参照。過去版restoreは513別操作

④受領subset: article category/slug、operation_key、expected revision、receipt/content hash、current version。unknownで新key不可

## D23 不具合報告・外部票作成・報告削除

担当境界: 503

実接点: [frontend/src/features/support/routes/BugReportsPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/support/routes/BugReportsPage.tsx) blob f36e954a0b41f6a20aee363d99c2cac5aa055666

実操作別のhandler・責務（固定ref a88fa4add58d43cff643590f9f507c8ab1a3c085）:

| 操作 | 実在するfile / handler | 受領・再利用の境界 |
|---|---|---|
| 新報告の院内保存 | [components/BugReportTab.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/support/components/BugReportTab.tsx#L102) のhandleSubmit→createReport.mutate（useCreateBugReport、L107） | global [SupportWidget.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/support/components/SupportWidget.tsx#L199)→[SupportPanel.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/support/components/SupportPanel.tsx#L56)→BugReportTabの入口。canSubmit/createReport.isPendingを再利用。BugReportsPage上の作成操作ではない |
| 既存報告からの外部Plane票作成・再試行 | BugReportsPage.tsxのPlaneTicketCellで受けるuseCreatePlaneTicket mutation、L112（再送）/L128（起票） | 新報告の保存と外部票結果を分け、503の現在認可・明示確認・unknown停止契約を受領する。通常の画面再試行を外部再送許可としない |
| 院内報告status変更 | 同file BugReportDetailDialog.handleToggle→useUpdateBugReportStatusのupdateStatus.mutate、L153 | status変更のpending/失敗/現在認可を保持。外部票の作成/削除や新報告保存へ混同しない |
| 院内報告削除 | 同file ConfirmDialog.onConfirm L328→useDeleteBugReportのdeleteReport.mutate L331 | pendingDelete、取消/失敗時保持、削除dialogを再利用。作成済み外部Plane票削除の操作にしない |

状態/共通接点: BugReportTabのform state/canSubmit/createReport.isPending、BugReportsPageのselected/pendingDelete、status変更/外部票作成それぞれのmutation状態、ErrorFallback/LoadingFallback。

適用: O1:E 報告/外部票; O2:N 外部送信取消なし; O3:E form state; O4:E deletion confirm; O5:F 院内保存/外部結果分離

具体差分/無修正根拠: 既存報告pendingとdelete dialogを再利用。外部送信前確認/院内保持/attachment logical-deleteは503受領、共通confirmの有無だけで対応済みにしない

④受領subset: report ID/report clinic、shared ROOM read vs mutation、external issue result、attachment storage+DB result、本人history refs

## P01 package master保存・削除・並替（N-PKGの非検索操作）

担当境界は521/523、型生成受領527。D17の医院/法人設定とは独立したmaster操作family。実接点は[frontend/src/features/master/routes/TreatmentPackageSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/master/routes/TreatmentPackageSettings.tsx)。#775の既存成果を再実装せず、keywordを新設しない。

| 操作 | handler / 共通接点 | 適用・受領境界 |
|---|---|---|
| 作成/更新・保存 | handleSave L66→validateDraft→editor.save→createMut/updateMut、L89のconflict/create-unknown/error入力保持 | O1:E、O3:E、O5:E。saved/noop/conflict/unknownを識別し、成功のみtarget更新・dirty.markClean。unknownを成功や新key再作成へ落とさない |
| 削除・取消 | handleDeleteConfirm L92→deleteMut、pendingDelete/canDelete/version、ConfirmDialog | O4:E、O5:E。取消は未保存破棄/永続削除と区別。成功時だけpendingDelete/target/editorをclear、失敗/競合では確認対象を保持 |
| 並替 | handleMove L113→moveId/buildReorderBody(orderVersion,next)→reorderMut | O1:E、O5:E。順序版・canEditを保持し同順序no-op、競合は再読込を案内。別資源や旧orderへ黙って上書きしない |
| 選択/保持・閉じる | dirty.runWithDiscardCheck→editor.resetFor、MasterPageShell/editorの既存接続 | O3:E、O4:E。保存済み構成の履歴復元を新設しない。O2:N（このfamilyに永続clear/reset操作はない） |

③では上記の実在接点をO1/O3/O4/O5のcoverageへ含め、状態/認可/失敗保持を限定確認する。Uは受入未確認であって追加実装欠陥の断定ではない。522/524のカルテpreview/applyと結果receiptは④で、masterのsave/delete/reorderの代用にしない。実装・実行試験NOT_RUN。

## ③先行作業と後続受領の分離

G3は既存CheckupFormのdirty接続のみ。親→子→fieldの成功ID保持と再開・lookupは500の④業務契約であり同じPRへ混ぜない。G4は既存DataStates適用先の記録であり、Lstep/DMマーケ仕様整備をPhase1内最後にする指定を保持して後順保留する。

D14のCSV/copyは448統合・493 staff採択契約へ受領する。514既存日時CRUDは再実装しない。既存ShiftFormDialogのoperation_key/lookup・ManualEditorの同key結果照会は適合接点として再利用し、共通化を理由に作り直さない。

受入未確認Uは欠陥件数へ加算しない。Fの新業務実装は③の前提とせず、DTO/参照/版・lookup・状態・拒否条件を指定ownerが受領した後に④で実装する。実装・試験は今回未実施。

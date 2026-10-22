# EMR-450 第③段階 共通UI 着手プラン

要件判断の窓口はMinoru Soga。既決仕様準拠・重複工程削減・計画coverage指標と未測定の業務成果の区別は[共通の要件検証](emr-445-stage1-contract-start-plan-20261021.md#requirement-validation-20261021)へ従う。実装担当claim・開始許可・実績測定は別。試験NOT_RUN。

## 目的・実行境界

①DB/API契約、②全マスタ、③共通UIを優先し、その後に④新規業務UIへ進む。本書は③の対象・既存接点・最小差分・受領条件を固定する。実装・試験・runtime有効化は本書の作成に含まれない。既存Done/UAT/Review/merged成果を保持し、進行中claimと重複する実装を開始しない。

Lstep/DMマーケ仕様整備はPhase1内の最後。L12・E05・N05配送監視部分・D15/G4は後順の明示保留とし、③先行必須gateから外す。N05の検査受信部分は別扱い。

## 正本・証拠

業務正本は docs/spec/domains/ と採用済み共通契約。コードは接点の所在・現状差の証拠であり、旧挙動を仕様として採用しない。

- [C01 UI契約](https://github.com/MinoruSoga/AnimalEkarte/blob/main/docs/spec/shared/ui-contracts.md)
- [検索対象・異用途台帳](https://github.com/MinoruSoga/AnimalEkarte/blob/main/docs/spec/shared/interaction-scope.md)
- [owner別共通契約](https://github.com/MinoruSoga/AnimalEkarte/blob/main/docs/spec/shared/owner-contract-coverage.md)
- [非検索操作family・受領条件](emr-450-stage3-common-ui-start-plan-20261021-operations.md)

検索対象は正本66行をMのタブ別consumerへ展開し、keywordのないpackage管理を加えた72行。件数は対象在庫であり、全writer網羅・全実装適合・実機合格の証拠ではない。各行の固定ref/blobを保持する。検索補完のsource pathsは a88fa4add58d43cff643590f9f507c8ab1a3c085 でも同一blobを確認した。

## 共通契約・担当・PR単位

- EMR-450: 本台帳と共通UI受入の責任。EMR-453: 既存共通部品。EMR-451: 患者・飼主候補共通接点。領域票はconsumer・必要subsetの受領先。
- keyword入力/IME/blur/Enterはdraftのみ、検索buttonでappliedへ。clearはdraftのみ。filter/sort/pageは適用済条件を使う。初回取得・URL復元・現在認可・API上限は維持する。
- 読取失敗を空集合へ変換せず、同じapplied条件でretry。候補はdraft不一致・未取得/失敗・scope不明時に確定させない。検索から保存/link/copy/sendを発火しない。
- 必要subsetは現read DTO、scope/limit、現在認可、query状態。新DB/API/writer・臨床実値・業務runtimeを③の共通UI PRへ混ぜない。
- 1PRは同目的shared部品と必要callerの最小adapter差分。claim/ownerを確認してFR/AC、変更path、状態・失敗・scopeの受入記述を含める。H03は二入口を独立に記録する。適合済み部品は再実装せず無修正根拠を残す。

E=既存接点、F=将来業務consumerの受領、N=当該keyword用途の対象外。これらは動作合格Cを意味しない。P/Uは受入未確認であり欠陥扱いしない。Fの新画面実装を③の完了前提にしない。

## 対象ごとの接点と作業

### L01 E-existing

担当: EMR-450。consumer境界: 463/451; 56/195既存履歴

実接点: OwnersListTable.PropertyFilter.onSearchChange → OwnersList.setSearchTerm → URL search effect; handleFilterChange retains species/include_deceased/checkup_history.

状態/無修正境界: PropertyFilter has searchDraft, handleSearchCommit, draft-only handleSearchClear. OwnersList URL state is the applied value. Actual consumer exists; no runtime/context acceptance claimed.
- [frontend/src/features/owners/routes/OwnersList.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/owners/routes/OwnersList.tsx) blob 4057ab361689492e60332dfb1ade41d96982f158
- [frontend/src/features/owners/components/OwnersListTable.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/owners/components/OwnersListTable.tsx) blob 8d79f8d30b3be4df8c18960af22f0fab9c923eb4
- [frontend/src/components/shared/PropertyFilter/PropertyFilter.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/PropertyFilter/PropertyFilter.tsx) blob c5935049495a2180703bf83aa6caec2bbd826b42

具体差分/再利用: 検索発火の修正なし: 実接点の PropertyFilter.onSearchChange はbutton commit後のapplied値。既存URL/page/filterの接続を保持する。未実機context/raceは受入未確認であって再実装理由にしない。

### L02 E-existing

担当: EMR-450。consumer境界: 58/244; 452

実接点: MedicalRecordsListPanels.onSearchChange → handleSearchChange → searchTerm → useMedicalRecordsList; useMedicalRecordsUrlState/resetKey.

状態/無修正境界: Route explicitly documents committed search values. Server predicate remains the medical-record list predicate; clinical history is H01.
- [frontend/src/features/medical-records/routes/MedicalRecords.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/medical-records/routes/MedicalRecords.tsx) blob 187e508f7ef1bca602435e1a511553576256bb6d
- [frontend/src/features/medical-records/routes/MedicalRecordsListPanels.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/medical-records/routes/MedicalRecordsListPanels.tsx) blob f78dc6f32ebd223ab071fbcb6a716f4039e8ebe5
- [frontend/src/features/medical-records/hooks/use-medical-records.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/medical-records/hooks/use-medical-records.ts) blob 71aa7d9777068f8d779b992dee9fc7fcc2b73b4a

具体差分/再利用: 検索発火の修正なし: 実接点の PropertyFilter.onSearchChange はbutton commit後のapplied値。既存URL/page/filterの接続を保持する。未実機context/raceは受入未確認であって再実装理由にしない。

### L03 E-existing

担当: EMR-450。consumer境界: 499/461; 452

実接点: HospitalizationListPageView.onSearchChange → handleSearchChange → setSearchTerm/resetListPage → applyHospitalizationClientFilters for list and board.

状態/無修正境界: List query useGetHospitalizations and board query useGetAllHospitalizations are separate; search does not determine cage occupancy.
- [frontend/src/features/hospitalization/routes/HospitalizationList.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/hospitalization/routes/HospitalizationList.tsx) blob b0a3a69cfbaf14705bbedace9c15c560c3089fa1
- [frontend/src/features/hospitalization/routes/HospitalizationListPanels.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/hospitalization/routes/HospitalizationListPanels.tsx) blob 1a53fa88088fc2582d54253e4290aa2133f0c9f5
- [frontend/src/features/hospitalization/hooks/use-hospitalization-list.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/hospitalization/hooks/use-hospitalization-list.ts) blob da34faa4e587079a1492d149830e4705ce51bd5d

具体差分/再利用: 検索発火の修正なし: 実接点の PropertyFilter.onSearchChange はbutton commit後のapplied値。既存URL/page/filterの接続を保持する。未実機context/raceは受入未確認であって再実装理由にしない。

### L04 E-existing

担当: EMR-450。consumer境界: 516; 415既存UI履歴

実接点: ExaminationsListContent/PropertyFilter.onSearchChange → handleSearchChange → deferredSearch → useFilterExaminationRecords; hasExaminationPageScopedFilter.

状態/無修正境界: Existing date-query/page-local text split and page-scoped banner are explicit route contacts.
- [frontend/src/features/examinations/routes/ExaminationsList.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/examinations/routes/ExaminationsList.tsx) blob 2d85542239d50a74badd362cbc6190416d3f1f86
- [frontend/src/features/examinations/routes/ExaminationsListPanels.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/examinations/routes/ExaminationsListPanels.tsx) blob b2e7d0a3f40e28cf83ebc4cb8df479fff5bc6458

具体差分/再利用: 検索発火の修正なし: 実接点の PropertyFilter.onSearchChange はbutton commit後のapplied値。既存URL/page/filterの接続を保持する。未実機context/raceは受入未確認であって再実装理由にしない。

### L05 E-existing

担当: EMR-450。consumer境界: 517/486; 415既存UI履歴

実接点: VaccinationListContent.onSearchChange → handleSearchChange → deferredSearchTerm → filters.search/useFilterVaccinations.

状態/無修正境界: PetId, date range and page navigation remain distinct; no evidence here that every Q0-Q7 case passes.
- [frontend/src/features/vaccinations/routes/VaccinationList.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/vaccinations/routes/VaccinationList.tsx) blob 6744df6ac161a5ab5fd0572c46ca3ec829afa25b
- [frontend/src/features/vaccinations/routes/VaccinationListPanels.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/vaccinations/routes/VaccinationListPanels.tsx) blob 2b6c917937822b67870ad2472c7a8f81fa589e6b

具体差分/再利用: 検索発火の修正なし: 実接点の PropertyFilter.onSearchChange はbutton commit後のapplied値。既存URL/page/filterの接続を保持する。未実機context/raceは受入未確認であって再実装理由にしない。

### L06 E-existing

担当: EMR-450。consumer境界: 480/482; 412既存UI履歴

実接点: TrimmingListContent.onSearchChange → handleSearchChange → searchKeyword/deferredKeyword → useFilterTrimmingRecords; pagination resetKey.

状態/無修正境界: Existing caller and keyword state located; date filters and sort are separate.
- [frontend/src/features/trimming/routes/TrimmingList.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/trimming/routes/TrimmingList.tsx) blob a56ca1200183a96ba2d63a4fbcd1e7248b695641
- [frontend/src/features/trimming/routes/TrimmingListPanels.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/trimming/routes/TrimmingListPanels.tsx) blob 6efa05665631087627879294e9551ce3ca78517c

具体差分/再利用: 検索発火の修正なし: 実接点の PropertyFilter.onSearchChange はbutton commit後のapplied値。既存URL/page/filterの接続を保持する。未実機context/raceは受入未確認であって再実装理由にしない。

### L07 E-existing

担当: EMR-450。consumer境界: 500; 197/223既存履歴

実接点: CheckupsListContent.onSearchChange → setSearchTerm → deferredSearch → filterCheckupsBySearch; useGetCheckups(requestFilters).

状態/無修正境界: Text is applied to fetched checkups; date/page query is separate. Species lookup is not a new server keyword predicate.
- [frontend/src/features/checkups/routes/CheckupsList.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/checkups/routes/CheckupsList.tsx) blob 81197cf8e4d9504dc1d186068d94d3799d080b11
- [frontend/src/features/checkups/routes/CheckupsListPanels.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/checkups/routes/CheckupsListPanels.tsx) blob 7f7868f8323f76f45c69986973a3d97c72746e76

具体差分/再利用: 検索発火の修正なし: 実接点の PropertyFilter.onSearchChange はbutton commit後のapplied値。既存URL/page/filterの接続を保持する。未実機context/raceは受入未確認であって再実装理由にしない。

### L08 E-existing

担当: EMR-450。consumer境界: 510

実接点: InventoryListContent.onSearchChange → handleSearchChange → deferredSearch → useInventoryList({searchTerm,page,limit}); hasPageScopedFilter.

状態/無修正境界: Fetched-page filtering and page-scoped indication exist; quantity mutation is a separate operation.
- [frontend/src/features/inventory/routes/InventoryList.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/inventory/routes/InventoryList.tsx) blob 90181888bb7d2a3506b2125b391dda4248c5ff38
- [frontend/src/features/inventory/routes/InventoryListPanels.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/inventory/routes/InventoryListPanels.tsx) blob 081b6842a658475c9de86261382482ae2adffe3b
- [frontend/src/features/inventory/hooks/use-inventory.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/inventory/hooks/use-inventory.ts) blob a46963be143994ee99cbf1acf52fe2882690994e

具体差分/再利用: 検索発火の修正なし: 実接点の PropertyFilter.onSearchChange はbutton commit後のapplied値。既存URL/page/filterの接続を保持する。未実機context/raceは受入未確認であって再実装理由にしない。

### L09 E-existing

担当: EMR-450。consumer境界: 508

実接点: EstimateListContent.onSearchChange → setSearchTerm → deferredSearch → filterAndSortEstimates(result.data,activeFilters,deferredSearch,activeSorts).

状態/無修正境界: Client collection filtering exists; usePagination(resetKey:deferredSearch) does not establish server-wide search.
- [frontend/src/features/estimates/routes/EstimateList.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/estimates/routes/EstimateList.tsx) blob 0321cc63149f2ea89c28a6d3affc26c7843af653
- [frontend/src/features/estimates/routes/EstimateListPanels.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/estimates/routes/EstimateListPanels.tsx) blob 7d9b53e2b5aa4dd9485ce9b8f0c00be396fa5425

具体差分/再利用: 検索発火の修正なし: 実接点の PropertyFilter.onSearchChange はbutton commit後のapplied値。既存URL/page/filterの接続を保持する。未実機context/raceは受入未確認であって再実装理由にしない。

### L10 E-existing

担当: EMR-450。consumer境界: 481/505/506

実接点: AccountingListTable.PropertyFilter.onSearchChange → handleSearchChange → buildAccountingListPageFilters → useGetAccountingsPage.

状態/無修正境界: List keyword is distinct from UnpaidTab and financial confirmation. Current PropertyFilter commit path applies; full restore/race verification remains open.
- [frontend/src/features/accounting/routes/AccountingList.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/accounting/routes/AccountingList.tsx) blob 964d6db23bffb72090bc22688b2894de2ab67db6
- [frontend/src/features/accounting/routes/AccountingListPanels.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/accounting/routes/AccountingListPanels.tsx) blob c275c0dd735b6cf57c5ba15382353f5d64f5b5f0
- [frontend/src/features/accounting/components/AccountingListTable.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/accounting/components/AccountingListTable.tsx) blob b7f46a91370af3bac84d91337a226cf1513c2660

具体差分/再利用: 検索発火の修正なし: 実接点の PropertyFilter.onSearchChange はbutton commit後のapplied値。既存URL/page/filterの接続を保持する。未実機context/raceは受入未確認であって再実装理由にしない。

### L11 E-existing

担当: EMR-450。consumer境界: 511; 162既存履歴

実接点: AggregationSearchFilter.Input.onChange → onParamsChange({search,page:1}) → dashboard params → useGetOwnerAggregations.

状態/無修正境界: Existing incremental input, not PropertyFilter. Common explicit-search retrofit has a concrete AggregationParams.search/page contact; existing Phase availability is not reopened.
- [frontend/src/features/aggregation/routes/AggregationDashboardPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/aggregation/routes/AggregationDashboardPage.tsx) blob e70d706abf05e3cca1c3a189d57db1cd6e027c76
- [frontend/src/features/aggregation/components/AggregationFilterPanel.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/aggregation/components/AggregationFilterPanel.tsx) blob 11ae0b795a2d9b04739d595763f92ee8bd41a7da
- [frontend/src/features/aggregation/components/AggregationFilterPanelParts.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/aggregation/components/AggregationFilterPanelParts.tsx) blob b07cb0fe3fbc75de532afb391a1696d687f25911

具体差分/再利用: 最小hunk: AggregationSearchFilter に draftSearch と検索buttonを置き、Input.onChange はdraftのみ。buttonで onParamsChange({search:draftSearch,page:1})。clearはdraftのみ、tab/filter/pageは既適用paramsを継続。GET/CSVの射程は変えない。

### L12 E-existing

担当: EMR-450。consumer境界: 504

実接点: TagSummaryTable.handleSearchChange → searchTerm/useDeferredValue → filteredTags/normalizedIncludes(tag_name).

状態/無修正境界: Existing incremental tag-name local search; categoryFilter and tag actions remain independent.
- [frontend/src/features/lstep/routes/LstepTagManagementPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/lstep/routes/LstepTagManagementPage.tsx) blob 45c5dbf28a4027404bf327cd9a794ca1c9b3d21c
- [frontend/src/features/lstep/components/TagSummaryTable.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/lstep/components/TagSummaryTable.tsx) blob 76da25fb5df833b454a3c8d88669426c76366368

具体差分/再利用: 最小hunk: TagSummaryTable の searchTerm を draft/applied に分離し、filteredTags/normalizedIncludes はappliedのみ。検索button/clear(draftのみ)を接続。categoryFilterとタグ同期mutationは独立のまま。

優先順: 明示保留: Lstep/DMマーケ仕様整備はPhase1内最後。既存共通DataStates/検索契約の適用先としてのみ記録し③先行必須gateから外す。

### M01 E-existing: frontend/src/features/master/routes/MedicineSettings.tsx

担当: EMR-450。consumer境界: 447

対象: 薬剤

実route: [frontend/src/features/master/routes/MedicineSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/MedicineSettings.tsx) blob fbb2036dfbff42bad54c0537c80a2c3e4f69cdef

実component: [frontend/src/features/master/components/MedicineTableSection.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/MedicineTableSection.tsx) blob d214358cabaf5e7c7be523ededa24361212197db

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/MedicineTableSection.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M02a E-existing: frontend/src/features/master/routes/TreatmentPlanMaster.tsx

担当: EMR-450。consumer境界: 454

対象: 診察

実route: [frontend/src/features/master/routes/TreatmentPlanMaster.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/TreatmentPlanMaster.tsx) blob 104d3632e9a4badf704ddbf1b2a15f32bc742f46

実component: [frontend/src/features/master/components/TreatmentPlanTabContent.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/TreatmentPlanTabContent.tsx) blob c355b74353fcf7e08f725a8cdd94c36a29ca122c

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/TreatmentPlanTabContent.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M02b E-existing: frontend/src/features/master/routes/TreatmentPlanMaster.tsx

担当: EMR-450。consumer境界: 454

対象: 処置

実route: [frontend/src/features/master/routes/TreatmentPlanMaster.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/TreatmentPlanMaster.tsx) blob 104d3632e9a4badf704ddbf1b2a15f32bc742f46

実component: [frontend/src/features/master/components/TreatmentPlanTabContent.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/TreatmentPlanTabContent.tsx) blob c355b74353fcf7e08f725a8cdd94c36a29ca122c

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/TreatmentPlanTabContent.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M02c E-existing: frontend/src/features/master/routes/TreatmentPlanMaster.tsx

担当: EMR-450。consumer境界: 455

対象: 予防接種

実route: [frontend/src/features/master/routes/TreatmentPlanMaster.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/TreatmentPlanMaster.tsx) blob 104d3632e9a4badf704ddbf1b2a15f32bc742f46

実component: [frontend/src/features/master/components/TreatmentPlanTabContent.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/TreatmentPlanTabContent.tsx) blob c355b74353fcf7e08f725a8cdd94c36a29ca122c

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/TreatmentPlanTabContent.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M02d E-existing: frontend/src/features/master/routes/TreatmentPlanMaster.tsx

担当: EMR-450。consumer境界: 456

対象: 定期健診

実route: [frontend/src/features/master/routes/TreatmentPlanMaster.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/TreatmentPlanMaster.tsx) blob 104d3632e9a4badf704ddbf1b2a15f32bc742f46

実component: [frontend/src/features/master/components/TreatmentPlanTabContent.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/TreatmentPlanTabContent.tsx) blob c355b74353fcf7e08f725a8cdd94c36a29ca122c

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/TreatmentPlanTabContent.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M03 E-existing: frontend/src/features/master/routes/TreatmentPlanMaster.tsx

担当: EMR-450。consumer境界: 447

対象: 検査

実route: [frontend/src/features/master/routes/TreatmentPlanMaster.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/TreatmentPlanMaster.tsx) blob 104d3632e9a4badf704ddbf1b2a15f32bc742f46

実component: [frontend/src/features/master/components/TreatmentPlanTabContent.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/TreatmentPlanTabContent.tsx) blob c355b74353fcf7e08f725a8cdd94c36a29ca122c

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/TreatmentPlanTabContent.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M04 E-existing: frontend/src/features/master/routes/HospitalizationSettings.tsx

担当: EMR-450。consumer境界: 448

対象: 入院プラン

実route: [frontend/src/features/master/routes/HospitalizationSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/HospitalizationSettings.tsx) blob 7e392d9c5b45c59f6764916a3bcbb71eb8cb7706

実component: [frontend/src/features/master/components/MasterCRUDPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/MasterCRUDPage.tsx) blob 2f32d18c2b05031c445cf80aedef4b500fe4dac2

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/MasterCRUDPage.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M05 E-existing: frontend/src/features/master/routes/MerchandiseItemSettings.tsx

担当: EMR-450。consumer境界: 449

対象: 物販

実route: [frontend/src/features/master/routes/MerchandiseItemSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/MerchandiseItemSettings.tsx) blob 79d8a52ca8c9f7f2370c64d3f77e633c51f8df94

実component: [frontend/src/features/master/components/MasterCRUDPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/MasterCRUDPage.tsx) blob 2f32d18c2b05031c445cf80aedef4b500fe4dac2

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/MasterCRUDPage.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M06a E-existing: frontend/src/features/master/routes/DiagnosisSettings.tsx

担当: EMR-450。consumer境界: 447

対象: 診断カテゴリ

実route: [frontend/src/features/master/routes/DiagnosisSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/DiagnosisSettings.tsx) blob ccfe82d2f5b36fe6c61a8a9665bbcf1ab56d3a99

実component: [frontend/src/features/master/components/DiagnosisTabs.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/DiagnosisTabs.tsx) blob 660fe156b934c71acbd5ff86cba521f08920b674

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/DiagnosisTabs.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M06b E-existing: frontend/src/features/master/routes/DiagnosisSettings.tsx

担当: EMR-450。consumer境界: 447

対象: 病名

実route: [frontend/src/features/master/routes/DiagnosisSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/DiagnosisSettings.tsx) blob ccfe82d2f5b36fe6c61a8a9665bbcf1ab56d3a99

実component: [frontend/src/features/master/components/DiagnosisTabs.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/DiagnosisTabs.tsx) blob 660fe156b934c71acbd5ff86cba521f08920b674

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/DiagnosisTabs.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M07a E-existing: frontend/src/features/master/routes/TrimmingSettings.tsx

担当: EMR-450。consumer境界: 457

対象: トリミングコース

実route: [frontend/src/features/master/routes/TrimmingSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/TrimmingSettings.tsx) blob 240dd6fdcb011f95a798026df8fc3a281d2bb8f5

実component: [frontend/src/features/master/components/TrimmingTabs.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/TrimmingTabs.tsx) blob 88ee745c555ac68714d1a370d84e70eaa15196e1

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/TrimmingTabs.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M07b E-existing: frontend/src/features/master/routes/TrimmingSettings.tsx

担当: EMR-450。consumer境界: 458

対象: トリミングoption

実route: [frontend/src/features/master/routes/TrimmingSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/TrimmingSettings.tsx) blob 240dd6fdcb011f95a798026df8fc3a281d2bb8f5

実component: [frontend/src/features/master/components/TrimmingTabs.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/TrimmingTabs.tsx) blob 88ee745c555ac68714d1a370d84e70eaa15196e1

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/TrimmingTabs.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M08 E-existing: frontend/src/features/master/routes/TrimmingCourseTypeSettings.tsx

担当: EMR-450。consumer境界: 448

対象: コース種別

実route: [frontend/src/features/master/routes/TrimmingCourseTypeSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/TrimmingCourseTypeSettings.tsx) blob 9443e9efb7adabf2b22b5eee6260a49b22af3817

実component: [frontend/src/features/master/components/MasterCRUDPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/MasterCRUDPage.tsx) blob 2f32d18c2b05031c445cf80aedef4b500fe4dac2

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/MasterCRUDPage.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M09 E-existing: frontend/src/features/master/routes/ReservationTypeSettings.tsx

担当: EMR-450。consumer境界: 492

対象: 予約区分/グループ

実route: [frontend/src/features/master/routes/ReservationTypeSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/ReservationTypeSettings.tsx) blob bf8027e895ee012e5b669dcf45f2c44cba43eecb

実component: [frontend/src/features/master/components/ReservationTypeSettingsContent.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/ReservationTypeSettingsContent.tsx) blob a5c956d85682a9e234c6024d95acb7261f642d0a

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

G1: group/type query状態未接続

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=G, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: G1: use-reservation-type-settings query states → ReservationTypeSettings → ReservationTypeSettingsContent

### M10 E-existing: frontend/src/features/master/routes/AnimalSpeciesSettings.tsx

担当: EMR-450。consumer境界: 453

対象: 動物種

実route: [frontend/src/features/master/routes/AnimalSpeciesSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/AnimalSpeciesSettings.tsx) blob 948fd0647908423587e3fcee11aa95929195fa6a

実component: [frontend/src/features/master/components/MasterCRUDPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/MasterCRUDPage.tsx) blob 2f32d18c2b05031c445cf80aedef4b500fe4dac2

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/MasterCRUDPage.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M11 E-existing: frontend/src/features/master/routes/CageSettings.tsx

担当: EMR-450。consumer境界: 448

対象: ケージ

実route: [frontend/src/features/master/routes/CageSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/CageSettings.tsx) blob d50b7fefb2bd7aa40e12a446b0d20ac41224d5d9

実component: [frontend/src/features/master/components/MasterCRUDPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/MasterCRUDPage.tsx) blob 2f32d18c2b05031c445cf80aedef4b500fe4dac2

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/MasterCRUDPage.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M12 E-existing: frontend/src/features/master/routes/CampaignSettings.tsx

担当: EMR-450。consumer境界: 449

対象: キャンペーン

実route: [frontend/src/features/master/routes/CampaignSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/CampaignSettings.tsx) blob 88e531a726b722b6f11b91ddb112dd9b19d6e7fa

実component: [frontend/src/features/master/components/MasterCRUDPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/MasterCRUDPage.tsx) blob 2f32d18c2b05031c445cf80aedef4b500fe4dac2

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/MasterCRUDPage.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M13 E-existing: frontend/src/features/master/routes/ChiefComplaintSettings.tsx

担当: EMR-450。consumer境界: 447

対象: 主訴

実route: [frontend/src/features/master/routes/ChiefComplaintSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/ChiefComplaintSettings.tsx) blob 970fbf7f3ecf45b6f960419e5dc249a43b848cc4

実component: [frontend/src/features/master/components/MasterCRUDPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/MasterCRUDPage.tsx) blob 2f32d18c2b05031c445cf80aedef4b500fe4dac2

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/MasterCRUDPage.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M14 E-existing: frontend/src/features/master/routes/InsuranceSettings.tsx

担当: EMR-450。consumer境界: 449

対象: 保険

実route: [frontend/src/features/master/routes/InsuranceSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/InsuranceSettings.tsx) blob 6a5be686c35b90827ba58648ee3ce8536c7f2334

実component: [frontend/src/features/master/components/MasterCRUDPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/MasterCRUDPage.tsx) blob 2f32d18c2b05031c445cf80aedef4b500fe4dac2

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/MasterCRUDPage.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M15 E-existing: frontend/src/features/master/routes/InterviewTemplateSettings.tsx

担当: EMR-450。consumer境界: 447

対象: 問診テンプレート

実route: [frontend/src/features/master/routes/InterviewTemplateSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/InterviewTemplateSettings.tsx) blob 3eed3bcd7b84436014043f3a2ee1ef5bf6db2a71

実component: [frontend/src/features/master/components/MasterCRUDPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/MasterCRUDPage.tsx) blob 2f32d18c2b05031c445cf80aedef4b500fe4dac2

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/MasterCRUDPage.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M16 E-existing: frontend/src/features/master/routes/OccupationSettings.tsx

担当: EMR-450。consumer境界: 448

対象: 職種

実route: [frontend/src/features/master/routes/OccupationSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/OccupationSettings.tsx) blob 03830c8383a8bf80bbb79f264bbbd023eb9c3aff

実component: [frontend/src/features/master/components/MasterCRUDPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/MasterCRUDPage.tsx) blob 2f32d18c2b05031c445cf80aedef4b500fe4dac2

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/MasterCRUDPage.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M17 E-existing: frontend/src/features/master/routes/PermissionGroupSettings.tsx

担当: EMR-450。consumer境界: 450

対象: 権限グループ

実route: [frontend/src/features/master/routes/PermissionGroupSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/PermissionGroupSettings.tsx) blob f6043159b823303a54fd8ff4db5fc43bb1c1d6e8

実component: [frontend/src/features/master/components/MasterCRUDPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/MasterCRUDPage.tsx) blob 2f32d18c2b05031c445cf80aedef4b500fe4dac2

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/MasterCRUDPage.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M18 E-existing: frontend/src/features/master/routes/StaffSettings.tsx

担当: EMR-450。consumer境界: 448

対象: スタッフ

実route: [frontend/src/features/master/routes/StaffSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/StaffSettings.tsx) blob 34adaab0d46a65e75d0bbde41c3a15a936749789

実component: [frontend/src/features/master/components/MasterCRUDPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/MasterCRUDPage.tsx) blob 2f32d18c2b05031c445cf80aedef4b500fe4dac2

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/MasterCRUDPage.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M19 E-existing: frontend/src/features/master/routes/PaymentMethodSettings.tsx

担当: EMR-450。consumer境界: 449

対象: 支払方法

実route: [frontend/src/features/master/routes/PaymentMethodSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/PaymentMethodSettings.tsx) blob e50396ca7ee39d5a80fc90c46a05d79478a08c67

実component: [frontend/src/features/master/components/MasterCRUDPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/components/MasterCRUDPage.tsx) blob 2f32d18c2b05031c445cf80aedef4b500fe4dac2

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/master/components/MasterCRUDPage.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### M20 E-existing: frontend/src/features/shifts/routes/ShiftTemplateSettings.tsx

担当: EMR-450。consumer境界: 493

対象: 勤務テンプレート

実route: [frontend/src/features/shifts/routes/ShiftTemplateSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/shifts/routes/ShiftTemplateSettings.tsx) blob b842687c933adc0e51e3179c290665507bd1394d

実component: [frontend/src/features/shifts/components/ShiftTemplateSettingsList.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/shifts/components/ShiftTemplateSettingsList.tsx) blob f7fbc8f12ad4ba00cd5726594d2723df949adc2d

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

G2: template query状態未接続

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=G, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: G2: ShiftTemplateSettings query states → Workspace → List

### M21 E-existing: frontend/src/features/clinic-settings/routes/ClinicMasterSettings.tsx

担当: EMR-450。consumer境界: 450

対象: 医院管理

実route: [frontend/src/features/clinic-settings/routes/ClinicMasterSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/clinic-settings/routes/ClinicMasterSettings.tsx) blob 06e37c54996f67f5b8049fcf225bf2ede3dde954

実component: [frontend/src/features/clinic-settings/components/ClinicMasterList.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/clinic-settings/components/ClinicMasterList.tsx) blob f9492f454fc98ad61606cbb7423d77848457ea74

共通接点: PropertyFilter.handleSearchCommit → onSearchChange(applied)。handleSearchClearはdraftのみ。MasterCRUDPage/MasterListPage経由又は直接caller。

既存error/loading分岐あり。scope/race/retryの全経路受入は未確認。clinic scope=all例外維持

FR: FR-common-ui-explicit-business-search, FR-common-ui-interaction-consistency

受入状態（静的/未確認、試験未実施）: trigger=C, clear-reset=P, context=U, result-state=P, race=U, scope=P, no-write=C, consistent-operation=P

具体差分/再利用: 再利用: frontend/src/features/clinic-settings/components/ClinicMasterList.tsx → PropertyFilter.handleSearchCommit / draft clear。trigger/no-write静的C、context/race等未実機を欠陥へ変換しない

### S01 E-existing

担当: EMR-450。consumer境界: 222/451/494

実接点: MedicalRecordPetSelection → PetSelectionSearchForm.setSearchParams → usePetSelectionPage.updateSearchParams → debouncedSearchParams → useGetPets; handleClear/handleSelect/pageBinding.

状態/無修正境界: Existing selector independently fetched. Shared hook explicitly says automatic search, uses 300ms debounce, PAGE_SIZE=20, includeDeceased and auth-derived clinicIds. New selection navigation remains separate from search.
- [frontend/src/features/medical-records/routes/MedicalRecordPetSelection.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/medical-records/routes/MedicalRecordPetSelection.tsx) blob c8e8758fec0a95fcc97dce4d0a31df6415432369
- [frontend/src/components/shared/PetSelection/PetSelectionSearchForm.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/PetSelection/PetSelectionSearchForm.tsx) blob b53ed590a118d44f3df33e4cb8c9b9a15763b591
- [frontend/src/hooks/use-pet-selection-page.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/hooks/use-pet-selection-page.ts) blob 64d0c395a281a2bc0915ec284d8b346b672f4735

具体差分/再利用: 最小共通hunk: PetSelectionSearchForm の入力→draft、検索button→usePetSelectionPage のapplied params、useGetPetsはappliedのみ。300ms debounceは入力からqueryへの発火源にしない。clearはdraftのみ、page/filterはapplied保持。pageBindingにapplied条件・医院・目的を拘束し、draft不一致/取得中/error時選択停止。各route固有permissionを維持。

### S02 E-existing

担当: EMR-450。consumer境界: 222/451/499

実接点: HospitalizationPetSelection → PetSelectionSearchForm.setSearchParams → usePetSelectionPage.updateSearchParams → debouncedSearchParams → useGetPets; handleClear/handleSelect/pageBinding.

状態/無修正境界: Existing selector independently fetched. Shared hook explicitly says automatic search, uses 300ms debounce, PAGE_SIZE=20, includeDeceased and auth-derived clinicIds. New selection navigation remains separate from search.
- [frontend/src/features/hospitalization/routes/HospitalizationPetSelection.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/hospitalization/routes/HospitalizationPetSelection.tsx) blob 75c31ef869cd5d20e4659715fc6477c355ec118a
- [frontend/src/components/shared/PetSelection/PetSelectionSearchForm.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/PetSelection/PetSelectionSearchForm.tsx) blob b53ed590a118d44f3df33e4cb8c9b9a15763b591
- [frontend/src/hooks/use-pet-selection-page.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/hooks/use-pet-selection-page.ts) blob 64d0c395a281a2bc0915ec284d8b346b672f4735

具体差分/再利用: 最小共通hunk: PetSelectionSearchForm の入力→draft、検索button→usePetSelectionPage のapplied params、useGetPetsはappliedのみ。300ms debounceは入力からqueryへの発火源にしない。clearはdraftのみ、page/filterはapplied保持。pageBindingにapplied条件・医院・目的を拘束し、draft不一致/取得中/error時選択停止。各route固有permissionを維持。

### S03 E-existing

担当: EMR-450。consumer境界: 222/451/480

実接点: TrimmingPetSelection → PetSelectionSearchForm.setSearchParams → usePetSelectionPage.updateSearchParams → debouncedSearchParams → useGetPets; handleClear/handleSelect/pageBinding.

状態/無修正境界: Existing selector independently fetched. Shared hook explicitly says automatic search, uses 300ms debounce, PAGE_SIZE=20, includeDeceased and auth-derived clinicIds. New selection navigation remains separate from search.
- [frontend/src/features/trimming/routes/TrimmingPetSelection.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/trimming/routes/TrimmingPetSelection.tsx) blob d583235f78160692be575d07cee068a606e8e31b
- [frontend/src/components/shared/PetSelection/PetSelectionSearchForm.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/PetSelection/PetSelectionSearchForm.tsx) blob b53ed590a118d44f3df33e4cb8c9b9a15763b591
- [frontend/src/hooks/use-pet-selection-page.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/hooks/use-pet-selection-page.ts) blob 64d0c395a281a2bc0915ec284d8b346b672f4735

具体差分/再利用: 最小共通hunk: PetSelectionSearchForm の入力→draft、検索button→usePetSelectionPage のapplied params、useGetPetsはappliedのみ。300ms debounceは入力からqueryへの発火源にしない。clearはdraftのみ、page/filterはapplied保持。pageBindingにapplied条件・医院・目的を拘束し、draft不一致/取得中/error時選択停止。各route固有permissionを維持。

### S04 E-existing

担当: EMR-450。consumer境界: 222/451/516

実接点: ExaminationPetSelection → PetSelectionSearchForm.setSearchParams → usePetSelectionPage.updateSearchParams → debouncedSearchParams → useGetPets; handleClear/handleSelect/pageBinding.

状態/無修正境界: Existing selector independently fetched. Shared hook explicitly says automatic search, uses 300ms debounce, PAGE_SIZE=20, includeDeceased and auth-derived clinicIds. New selection navigation remains separate from search.
- [frontend/src/features/examinations/routes/ExaminationPetSelection.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/examinations/routes/ExaminationPetSelection.tsx) blob 1fd3ce5ad84e585cab4f23b4ae878ae1bb2bd057
- [frontend/src/components/shared/PetSelection/PetSelectionSearchForm.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/PetSelection/PetSelectionSearchForm.tsx) blob b53ed590a118d44f3df33e4cb8c9b9a15763b591
- [frontend/src/hooks/use-pet-selection-page.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/hooks/use-pet-selection-page.ts) blob 64d0c395a281a2bc0915ec284d8b346b672f4735

具体差分/再利用: 最小共通hunk: PetSelectionSearchForm の入力→draft、検索button→usePetSelectionPage のapplied params、useGetPetsはappliedのみ。300ms debounceは入力からqueryへの発火源にしない。clearはdraftのみ、page/filterはapplied保持。pageBindingにapplied条件・医院・目的を拘束し、draft不一致/取得中/error時選択停止。各route固有permissionを維持。

### S05 E-existing

担当: EMR-450。consumer境界: 222/451/500

実接点: CheckupPetSelection → PetSelectionSearchForm.setSearchParams → usePetSelectionPage.updateSearchParams → debouncedSearchParams → useGetPets; handleClear/handleSelect/pageBinding.

状態/無修正境界: Existing selector independently fetched. Shared hook explicitly says automatic search, uses 300ms debounce, PAGE_SIZE=20, includeDeceased and auth-derived clinicIds. New selection navigation remains separate from search. Checkup route is its own permission contact, not interchangeable with other selector gates.
- [frontend/src/features/checkups/routes/CheckupPetSelection.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/checkups/routes/CheckupPetSelection.tsx) blob 92858fe6c03c0d9a60518516b384b6f0c6723bda
- [frontend/src/components/shared/PetSelection/PetSelectionSearchForm.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/PetSelection/PetSelectionSearchForm.tsx) blob b53ed590a118d44f3df33e4cb8c9b9a15763b591
- [frontend/src/hooks/use-pet-selection-page.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/hooks/use-pet-selection-page.ts) blob 64d0c395a281a2bc0915ec284d8b346b672f4735

具体差分/再利用: 最小共通hunk: PetSelectionSearchForm の入力→draft、検索button→usePetSelectionPage のapplied params、useGetPetsはappliedのみ。300ms debounceは入力からqueryへの発火源にしない。clearはdraftのみ、page/filterはapplied保持。pageBindingにapplied条件・医院・目的を拘束し、draft不一致/取得中/error時選択停止。各route固有permissionを維持。

### S06 E-existing

担当: EMR-450。consumer境界: 222/451/505

実接点: AccountingPetSelection → PetSelectionSearchForm.setSearchParams → usePetSelectionPage.updateSearchParams → debouncedSearchParams → useGetPets; handleClear/handleSelect/pageBinding.

状態/無修正境界: Existing selector independently fetched. Shared hook explicitly says automatic search, uses 300ms debounce, PAGE_SIZE=20, includeDeceased and auth-derived clinicIds. New selection navigation remains separate from search.
- [frontend/src/features/accounting/routes/AccountingPetSelection.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/accounting/routes/AccountingPetSelection.tsx) blob 48a6724c9de887e5cbbff44b80d8b07fab95bc1b
- [frontend/src/components/shared/PetSelection/PetSelectionSearchForm.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/PetSelection/PetSelectionSearchForm.tsx) blob b53ed590a118d44f3df33e4cb8c9b9a15763b591
- [frontend/src/hooks/use-pet-selection-page.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/hooks/use-pet-selection-page.ts) blob 64d0c395a281a2bc0915ec284d8b346b672f4735

具体差分/再利用: 最小共通hunk: PetSelectionSearchForm の入力→draft、検索button→usePetSelectionPage のapplied params、useGetPetsはappliedのみ。300ms debounceは入力からqueryへの発火源にしない。clearはdraftのみ、page/filterはapplied保持。pageBindingにapplied条件・医院・目的を拘束し、draft不一致/取得中/error時選択停止。各route固有permissionを維持。

### S07 E-existing

担当: EMR-450。consumer境界: 222/451/517

実接点: VaccinationPetSelection → PetSelectionSearchForm.setSearchParams → usePetSelectionPage.updateSearchParams → debouncedSearchParams → useGetPets; handleClear/handleSelect/pageBinding.

状態/無修正境界: Existing selector independently fetched. Shared hook explicitly says automatic search, uses 300ms debounce, PAGE_SIZE=20, includeDeceased and auth-derived clinicIds. New selection navigation remains separate from search.
- [frontend/src/features/vaccinations/routes/VaccinationPetSelection.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/vaccinations/routes/VaccinationPetSelection.tsx) blob 99b20b97cb12b2cdfcc0ddb2d042d51f2a877b19
- [frontend/src/components/shared/PetSelection/PetSelectionSearchForm.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/PetSelection/PetSelectionSearchForm.tsx) blob b53ed590a118d44f3df33e4cb8c9b9a15763b591
- [frontend/src/hooks/use-pet-selection-page.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/hooks/use-pet-selection-page.ts) blob 64d0c395a281a2bc0915ec284d8b346b672f4735

具体差分/再利用: 最小共通hunk: PetSelectionSearchForm の入力→draft、検索button→usePetSelectionPage のapplied params、useGetPetsはappliedのみ。300ms debounceは入力からqueryへの発火源にしない。clearはdraftのみ、page/filterはapplied保持。pageBindingにapplied条件・医院・目的を拘束し、draft不一致/取得中/error時選択停止。各route固有permissionを維持。

### S08 E-existing

担当: EMR-450。consumer境界: 451/468; 264/318既存履歴

実接点: PatientSearchFilters.onTextFieldChange → usePatientSelectionTable.handleTextFieldChange → debouncedSearchParams → useGetPets(enabled:hasSearchConditions).

状態/無修正境界: Existing debounced patient search; selection/page binding and clear remain separate from reservation confirmation.
- [frontend/src/components/shared/ReservationFormModal/PatientSelectionTable.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/ReservationFormModal/PatientSelectionTable.tsx) blob 80390a086280d5216dc9c8672b7616cbd89a4516
- [frontend/src/components/shared/ReservationFormModal/PatientSearchFilters.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/ReservationFormModal/PatientSearchFilters.tsx) blob aa128b388db2727e4389dc1b06c122ad477b8b57
- [frontend/src/components/shared/ReservationFormModal/hooks/use-patient-selection-table.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/ReservationFormModal/hooks/use-patient-selection-table.ts) blob 24bdbe59607b610d1769ee71410cb79591868a52

具体差分/再利用: 最小hunk: PatientSearchFilters と usePatientSelectionTable のdraft/applied分離。debouncedSearchParamsを入力直結の取得条件にせずbuttonでapplied更新。hasSearchConditions/pageBinding/selected patient guardをappliedに揃え、clear/検索は予約onSaveを呼ばない。

### S09 E-existing

担当: EMR-450。consumer境界: 451/490/491

実接点: PetEditModal.OwnerSearchModal → handleSearch → useOwnerSearch.mutateAsync(trimmed searchTerm); handleOwnerChange is separate from search.

状態/無修正境界: Search button and Enter currently invoke handleSearch. OwnerSearchModal confirmation and caller owner-change boundary must not be folded into search.
- [frontend/src/features/owners/components/PetEditModal.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/owners/components/PetEditModal.tsx) blob 74a327fa952ef9c8a1c4dd30aad9246215091a54
- [frontend/src/components/shared/OwnerSearchModal/OwnerSearchModal.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/OwnerSearchModal/OwnerSearchModal.tsx) blob a57b59fed10d24febb56e0af54410ab64bf1cdc6
- [frontend/src/hooks/use-owner-search.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/hooks/use-owner-search.ts) blob 84ad0afc425e1470d007718087231c206e816404

具体差分/再利用: 最小hunk: OwnerSearchModal のInput Enter→handleSearchを除去しbuttonのみ発火へ。draft clearとapplied候補/query scopeを分離し入力不一致時選択停止。handleOwnerChange/影響確認は別callbackのまま。

### S10 E-existing

担当: EMR-450。consumer境界: 451/491

実接点: Input.onChange → setCandidateSearch → debouncedCandidateSearch.trim → useGetSubOwnerCandidates; candidateOptions exclude primary and existing rows; SearchableSelect.onValueChange → handleAddSubOwner.

状態/無修正境界: Candidate search is an existing business S consumer even though its second local candidate control is SearchableSelect. Replace/save uses metadata version and is a distinct write.
- [frontend/src/features/owners/components/PetSubOwnersSection.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/owners/components/PetSubOwnersSection.tsx) blob 5ef74aab554ef404491fa408d36b09192560cb1f
- [frontend/src/features/owners/hooks/use-pet-sub-owners-section.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/owners/hooks/use-pet-sub-owners-section.ts) blob 89be4eca73627b27b07a13e60e9c65fc207828e1

具体差分/再利用: 最小hunk: PetSubOwnersSection candidateSearch入力とuse-pet-sub-owners-section のdebouncedCandidateSearchをdraft/applied分離。検索buttonでcandidate GET条件だけ更新。SearchableSelect内の局所候補補完はE用途として再利用。handleAddSubOwner/全置換save/metadata.versionは検索から呼ばない。

### S11 E-existing

担当: EMR-450。consumer境界: 451; identitylink既存owner

実接点: OwnerLinkSection.setOwnerQuery → workbench.ownerQuery/debouncedOwnerQuery → useOwnerSearchQuery → searchOwnersForLink.

状態/無修正境界: Existing debounced owner candidate search, generated OwnerSearchItem. toggleOwner/onLinkOwners are separate from read.
- [frontend/src/features/identity-links/routes/IdentityLinksPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/identity-links/routes/IdentityLinksPage.tsx) blob 32cfba652eeb5f60e5f4a61f1f4cbb6214c16751
- [frontend/src/features/identity-links/components/OwnerLinkSection.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/identity-links/components/OwnerLinkSection.tsx) blob 1d5329bc636f78de1560c27e761c2554e574a6d2
- [frontend/src/features/identity-links/hooks/use-identity-links-workbench.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/identity-links/hooks/use-identity-links-workbench.ts) blob 81903270ae10bdf6984c6b4c57365cc14d4e4e1a
- [frontend/src/features/identity-links/hooks/use-identity-link-search.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/identity-links/hooks/use-identity-link-search.ts) blob e118ee4f415331163030635547d85c058386b89d

具体差分/再利用: 最小hunk: OwnerLinkSection.ownerQuery と useIdentityLinksWorkbench の対応queryをdraft/applied分離し、検索buttonでuseOwnerSearchQuery/usePetSearchQueryへ適用。trim/min1/既存上限とownerGroupIdを維持。toggle/link/unlinkは別操作で入力不一致候補を確定させない。

### S12 E-existing

担当: EMR-450。consumer境界: 451; identitylink既存owner

実接点: PetLinkSection.setPetQuery → workbench.petQuery/debouncedPetQuery → usePetSearchQuery → searchPetsForLink.

状態/無修正境界: Existing debounced pet candidate search, generated PetSearchItem. Parent ownerGroupId and link/write guard stay unchanged.
- [frontend/src/features/identity-links/routes/IdentityLinksPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/identity-links/routes/IdentityLinksPage.tsx) blob 32cfba652eeb5f60e5f4a61f1f4cbb6214c16751
- [frontend/src/features/identity-links/components/PetLinkSection.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/identity-links/components/PetLinkSection.tsx) blob 84d917f2b27ca5b84ecae128129ce65432bbaa8f
- [frontend/src/features/identity-links/hooks/use-identity-links-workbench.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/identity-links/hooks/use-identity-links-workbench.ts) blob 81903270ae10bdf6984c6b4c57365cc14d4e4e1a
- [frontend/src/features/identity-links/hooks/use-identity-link-search.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/identity-links/hooks/use-identity-link-search.ts) blob e118ee4f415331163030635547d85c058386b89d

具体差分/再利用: 最小hunk: PetLinkSection.petQuery と useIdentityLinksWorkbench の対応queryをdraft/applied分離し、検索buttonでuseOwnerSearchQuery/usePetSearchQueryへ適用。trim/min1/既存上限とownerGroupIdを維持。toggle/link/unlinkは別操作で入力不一致候補を確定させない。

### S13 F-future

担当: EMR-450。consumer境界: 451/517; 273既存履歴

実接点: Absent inline search contact: VaccinationFormBody renders PatientInfoCard(selectedPet), not owner/pet search; useVaccinationForm owns petSelection.selectedPets/setSelectedPets and useGetPet(route petId).

状態/無修正境界: S07 is the actual existing selection entry. S13 inline owner/pet candidate UI is not present in inspected form/hook; do not double-count S07 or assert an existing inline handler.
- [frontend/src/features/vaccinations/routes/VaccinationForm.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/vaccinations/routes/VaccinationForm.tsx) blob 6290536483ab5699f30c3bbad03a9dbf920c3c51
- [frontend/src/features/vaccinations/routes/VaccinationFormPagePanels.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/vaccinations/routes/VaccinationFormPagePanels.tsx) blob f8945d202ac8bb35c978b48e8c6d5c9331380a74
- [frontend/src/features/vaccinations/hooks/use-vaccination-form.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/vaccinations/hooks/use-vaccination-form.ts) blob 3bc5939a761ef36db9b8a3726e7fabc20c92a05a
- [frontend/src/hooks/use-pet-selection.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/hooks/use-pet-selection.ts) blob ecbbcab2755d0a631da8cff0f75c7460919c0aaf
- [frontend/src/features/vaccinations/routes/VaccinationPetSelection.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/vaccinations/routes/VaccinationPetSelection.tsx) blob 99b20b97cb12b2cdfcc0ddb2d042d51f2a877b19

具体差分/再利用: F-future: 現inline handlerなし。S07へ二重計上しない。451/517受領subsetを固定し、現PatientInfoCardを新検索UIへ変更する実装は④。S13採用範囲とS07置換関係のowner承認が得られるまでは実装着手不可、③部品契約は待たない。

将来受領:
- owner_tickets: [451, 517]
- owner_boundary: D08 vaccination form; D04 patient search/ownership
- existing_state: useVaccinationForm.petSelection: selectedPets/setSelectedPets; route petId; useGetPet; VaccinationPatient.id/ownerId/name/ownerName/status/deceasedAt
- proposed_contact: Separate draft/applied candidate query at VaccinationFormBody boundary, using PetSelectionSearchParams and existing Pet result; selected candidate sets the form patient only after ownership/read/deceased checks. Keep selection distinct from formAction/useCreateVaccination/useUpdateVaccination.
- required_receipt: #451/#517 must reconcile whether S13 remains a distinct adopted inline flow or is superseded by S07. No new UI or acceptance closure authorized by this inventory.
- stage: Stage 4 business implementation/owner reconciliation; not a reason to block Stage 3 shared-contact planning.

### H01 E-existing

担当: EMR-450。consumer境界: 452/494/495/477

実接点: useGetPetMedicalHistory(selectedPet.id,recordId,recordClinicId) → medicalHistoryItems/merged timeline → InterviewHistory.searchTerm/deferredSearch → filteredItems.

状態/無修正境界: Actual history search is incremental title/content/type/author/date matching. onCopyItem/navigation is separate; same-pet/clinic and incomplete/error flags remain owner-owned.
- [frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx) blob 07bdec28490c42f3d3eec328d92a9564ee10442e
- [frontend/src/features/medical-records/components/InterviewHistory.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/medical-records/components/InterviewHistory.tsx) blob a5a7d4bd8af3fc715fe607b78d50e13d5200c887

具体差分/再利用: 最小hunk: InterviewHistory のInput→draft、button→applied、filteredItemsのtitle/content/type/author/date述語はappliedのみ。clearはdraft、copy/navigation/same-pet-clinic/incomplete状態は変更しない。

### H02 E-existing

担当: EMR-450。consumer境界: 516/501

実接点: HistoryFilterPanel.onSearchTermChange → setHistorySearchTerm → deferredHistorySearch → searchedPetHistory/filteredHistory; useGetExaminations receives pet/date filters.

状態/無修正境界: Existing history clear resets keyword/sort/date; explicit-search migration must split draft clear from applied reset without widening fetched history.
- [frontend/src/features/examinations/components/ExaminationHistoryPanel.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/examinations/components/ExaminationHistoryPanel.tsx) blob cd3fa0349dd33fdfd7555462323e11bf5599c31b
- [frontend/src/features/examinations/hooks/use-examination-history-filters.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/examinations/hooks/use-examination-history-filters.ts) blob e9c1237e08f622130987a06c7d1d80e0136ccc26
- [frontend/src/components/shared/HistoryFilterPanel/HistoryFilterPanel.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/HistoryFilterPanel/HistoryFilterPanel.tsx) blob da44ed196f5850a5678d4d2643b32dbacf34b071

具体差分/再利用: 最小hunk: HistoryFilterPanelとuseExaminationHistoryFiltersにdraft/applied keywordを分離し、searchedPetHistoryにはappliedを渡す。history clearはkeyword draftのみ、既存date/sortの独立resetと区別。取得先頭20件の射程を保持。

### H03 E-existing

担当: EMR-450。consumer境界: 517/486

実接点: Standalone: VaccinationHistoryPanel.onHistorySearchTermChange → historyFilter.setHistorySearchTerm → filterVaccinationHistory. Embedded: VaccinationHistory.Input.onChange → searchTerm/deferredSearch → filteredItems.

状態/無修正境界: Two concrete history callers exist and require separate acceptance rows. Standalone shared HistoryFilterPanel does not cover embedded VaccinationHistory; copy and replaced-vaccination display remain separate.
- [frontend/src/features/vaccinations/routes/VaccinationForm.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/vaccinations/routes/VaccinationForm.tsx) blob 6290536483ab5699f30c3bbad03a9dbf920c3c51
- [frontend/src/features/vaccinations/components/VaccinationFormPanels.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/vaccinations/components/VaccinationFormPanels.tsx) blob b3e4ccfbf2aa7facb311ca5307a1dda191b59957
- [frontend/src/features/medical-records/components/VaccinationHistory.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/medical-records/components/VaccinationHistory.tsx) blob 22f856506b66ce1084d4cb8add77f255e4b0527e
- [frontend/src/components/shared/HistoryFilterPanel/HistoryFilterPanel.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/HistoryFilterPanel/HistoryFilterPanel.tsx) blob da44ed196f5850a5678d4d2643b32dbacf34b071

具体差分/再利用: 最小hunkは二入口別: standalone HistoryFilterPanel/useVaccinationHistoryFilter と embedded VaccinationHistory Input/filter。双方にdraft/applied/button、draft clearを接続。履歴copy/置換済表示/上限/期間は保持し、片方の受入を他方へ代入しない。

### H04 E-existing

担当: EMR-450。consumer境界: 480/482

実接点: TrimmingForm chrome.history.sortedHistory → TrimmingRightColumn.historyItems → PastRecordHistoryPanel.searchTerm → filteredItems.

状態/無修正境界: Existing incremental history. onItemClick resolves original history and calls onHistoryClick; search must not copy form values.
- [frontend/src/features/trimming/routes/TrimmingForm.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/trimming/routes/TrimmingForm.tsx) blob 01c5fd146a42fe5ca0ac6cf9fcfe45ce9f0a4701
- [frontend/src/features/trimming/components/TrimmingRightColumn.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/trimming/components/TrimmingRightColumn.tsx) blob f8e82aa4902e32b312db9f341315fe920262eee4
- [frontend/src/components/shared/PastRecordHistoryPanel/PastRecordHistoryPanel.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/PastRecordHistoryPanel/PastRecordHistoryPanel.tsx) blob fc76d45629847a880890cb3d45a3aad25318a356

具体差分/再利用: 最小共通hunk: PastRecordHistoryPanel のsearchTermをdraft/appliedに分離しbutton適用、filteredItemsはappliedのみ、clearはdraft。callerのitems集合・onItemClick/onHistoryClickは維持。施術履歴copyの業務動作を検索handlerへ追加しない。

### H05 E-existing

担当: EMR-450。consumer境界: 500

実接点: CheckupForm.useGetCheckups({petId,...}) → toCheckupHistoryItems → PastRecordHistoryPanel.items; onSearchTermChange/setSearchTerm/filter.

状態/無修正境界: Existing common history caller, with checkup-specific fetched collection and kind/findings text.
- [frontend/src/features/checkups/routes/CheckupForm.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/checkups/routes/CheckupForm.tsx) blob 8f9ab3b1e1b05569985e7b3d0ae8defc67c9914b
- [frontend/src/components/shared/PastRecordHistoryPanel/PastRecordHistoryPanel.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/PastRecordHistoryPanel/PastRecordHistoryPanel.tsx) blob fc76d45629847a880890cb3d45a3aad25318a356

具体差分/再利用: 最小共通hunk: PastRecordHistoryPanel のsearchTermをdraft/appliedに分離しbutton適用、filteredItemsはappliedのみ、clearはdraft。callerのitems集合・onItemClick/onHistoryClickは維持。健診result/false/0の業務動作を検索handlerへ追加しない。

追加静的差分（固定main c86c8dd2）：CheckupForm.tsx L50–59はuseGetCheckupsのerror/refetchを受けずdata??[]へ変換し、L146–151はitems/isLoadingだけをPastRecordHistoryPanelへ渡すため、初回GET失敗と成功0件が「履歴がありません」へ合流する。H05の450 consumer接続としてerror/retryを既存panel又は同じcallerのDataStatesへ通し、取得中/失敗/成功0件を分離する。再試行は同患者・医院・条件に限り、失権/遅延応答で旧履歴を操作根拠にしない。G3のdirty接続とは別差分。500の④receipt/保存再開/lookupをこの受入へ混ぜない。fixtureは初回失敗・成功0件・成功あり・再試行失敗/成功・患者切替後遅延（NOT_RUN）。

### H06 E-existing

担当: EMR-450。consumer境界: 499/461

実接点: HospitalizationForm chrome history → HospitalizationFormFields.PastRecordHistoryPanel.items; searchTerm/deferredSearch → filteredItems.

状態/無修正境界: Existing history caller, type/status local search. Not current room/cage availability.
- [frontend/src/features/hospitalization/routes/HospitalizationForm.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/hospitalization/routes/HospitalizationForm.tsx) blob 49713b95f9bfa96ec075cb94a967d0cedb7b662d
- [frontend/src/features/hospitalization/routes/HospitalizationFormFields.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/hospitalization/routes/HospitalizationFormFields.tsx) blob 0576881067335f98442725c328e8ed9b6d1a6d75
- [frontend/src/components/shared/PastRecordHistoryPanel/PastRecordHistoryPanel.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/PastRecordHistoryPanel/PastRecordHistoryPanel.tsx) blob fc76d45629847a880890cb3d45a3aad25318a356

具体差分/再利用: 最小共通hunk: PastRecordHistoryPanel のsearchTermをdraft/appliedに分離しbutton適用、filteredItemsはappliedのみ、clearはdraft。callerのitems集合・onItemClick/onHistoryClickは維持。入院stay/ケージ占有の業務動作を検索handlerへ追加しない。

### N01 N-purpose

担当: EMR-450。consumer境界: 468/452

実接点: Reception filters.toggleVisitType/setSelectedDoctor/setIsTrimmingOnly; ReservationManagement.setDoctorFilter/setSourceFilter/filterCalendarAppointments.

状態/無修正境界: Independent date/classification/doctor controls, no business keyword input in these page bodies. Reservation patient search separately accounted S08.
- [frontend/src/features/reception/routes/Reception.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/reception/routes/Reception.tsx) blob e3657db61017a213516076da2ffc60230c8b7699
- [frontend/src/features/reservations/routes/ReservationManagement.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/reservations/routes/ReservationManagement.tsx) blob 00bc347d0584494834bffa26f882167488fa792f

具体差分/再利用: keyword修正なし: Independent date/classification/doctor controls, no business keyword input in these page bodies. Reservation patient search separately accounted S08. 共通C01の非検索操作は別operation-family台帳へ。部品名だけで一括置換しない。

### N02 N-purpose

担当: EMR-450。consumer境界: 505/506

実接点: UnpaidTab.handleStartDateChange/handleEndDateChange/handleGroupByChange → URL conditions; AccountingListContent renders tab-specific UI. DailyAccountingTab.handleDateChange → URL selectedDate → useGetAccountings/date_start/date_end + useGetDailySummary(selectedDate,selectedClinicIds); window.print is separate.

状態/無修正境界: 未納period/groupと当日selectedDateは独立filter。DailyAccountingTabにkeywordなし、日付変更と印刷は検索button化しない。
- [frontend/src/features/accounting/components/UnpaidTab.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/accounting/components/UnpaidTab.tsx) blob 6029c8d77989412ca4bf19d05168531544a155af
- [frontend/src/features/accounting/routes/AccountingListPanels.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/accounting/routes/AccountingListPanels.tsx) blob c275c0dd735b6cf57c5ba15382353f5d64f5b5f0
- [frontend/src/features/accounting/components/DailyAccountingTab.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/accounting/components/DailyAccountingTab.tsx) blob 0d2b625cef7bef716b6678082d344bc00108e859

具体差分/再利用: keyword修正なし: Unpaid period/group controls are not keyword; AccountingList keyword remains L10, not excluded by N02. A separate same-day page handler was not established in this read. 共通C01の非検索操作は別operation-family台帳へ。部品名だけで一括置換しない。

### N03 N-purpose

担当: EMR-450。consumer境界: 509/511

実接点: CashRegisterClosePage date/segment state and close action; AccountingReportsPage useGetMonthlyReport(MonthlyReportParams)/exportMonthlyCSV.

状態/無修正境界: Date/month/report selection and aggregation/printing are not business keyword search; close/export writes and data completeness stay distinct.
- [frontend/src/features/cash-register/routes/CashRegisterClosePage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/cash-register/routes/CashRegisterClosePage.tsx) blob 92b5fcabf0335859190ae2ae1ec7202718d19b2b
- [frontend/src/features/accounting-reports/routes/AccountingReportsPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/accounting-reports/routes/AccountingReportsPage.tsx) blob 506ebb92365a705a2f2d994eec93d1a2fcd974dd

具体差分/再利用: keyword修正なし: Date/month/report selection and aggregation/printing are not business keyword search; close/export writes and data completeness stay distinct. 共通C01の非検索操作は別operation-family台帳へ。部品名だけで一括置換しない。

### N04 N-purpose

担当: EMR-450。consumer境界: 493/514/519

実接点: ShiftCalendarPage.handleOccupationChange plus month/staff selection; LineReservationSettings → LineReservationSettingsForm.

状態/無修正境界: Calendar/setting selectors have different purpose; no generic keyword consumer added. Template master keyword remains M20.
- [frontend/src/features/shifts/routes/ShiftCalendarPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/shifts/routes/ShiftCalendarPage.tsx) blob fa2f0bcd59b32bc044b8655dc6ff561e3fd48680
- [frontend/src/features/line-reservation/routes/LineReservationSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/line-reservation/routes/LineReservationSettings.tsx) blob 122e48878d70c720bc65a02a9ea2626e8d9c9711

具体差分/再利用: keyword修正なし: Calendar/setting selectors have different purpose; no generic keyword consumer added. Template master keyword remains M20. 共通C01の非検索操作は別operation-family台帳へ。部品名だけで一括置換しない。

### N05 N-purpose

担当: EMR-450。consumer境界: 504/516

実接点: Delivery monitor handleFromChange/handleToChange/handleTriggerTypeChange/handleStatusChange/handleRefresh; LabDeviceBoard receive and handleSelectVisit/handleAttach/handleDetach.

状態/無修正境界: Monitor read conditions/reload and lab receive/attach operations are not a new keyword search. Attach must not be fired by changing display conditions.
- [frontend/src/features/lstep/routes/LstepDeliveryMonitorPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/lstep/routes/LstepDeliveryMonitorPage.tsx) blob a6b25d24773be7fb9b850e98c34cbd8e1a514b1d
- [frontend/src/features/lab-device/routes/LabDeviceBoard.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/lab-device/routes/LabDeviceBoard.tsx) blob d75c094a7cd3c44584e59082b0a50cc996fb1216

具体差分/再利用: keyword修正なし: Monitor read conditions/reload and lab receive/attach operations are not a new keyword search. Attach must not be fired by changing display conditions. 共通C01の非検索操作は別operation-family台帳へ。部品名だけで一括置換しない。

優先順: 配送監視マーケ部分はPhase1内最後の明示保留。検査受信接点と分離する。

### N06 N-purpose

担当: EMR-450。consumer境界: 497/504/474

実接点: OwnerReport patient/report state; PetHealthPage useFetchState(fetchHealthCard) and data.pets.map.

状態/無修正境界: Registered-patient report display, not business data search. LIFF has fetch retry but no search input in fetched page.
- [frontend/src/features/owner-report/routes/OwnerReport.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/owner-report/routes/OwnerReport.tsx) blob bb76331c7c6b1bd21d0aa06318aa1c3a3fa20a1f
- [frontend/liff/src/pages/PetHealthPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/liff/src/pages/PetHealthPage.tsx) blob 304d1bf5f727f6296bbb1f7f268c154fadfce35b

具体差分/再利用: keyword修正なし: Registered-patient report display, not business data search. LIFF has fetch retry but no search input in fetched page. 共通C01の非検索操作は別operation-family台帳へ。部品名だけで一括置換しない。

### N07 N-purpose

担当: EMR-450。consumer境界: 503/516

実接点: BugReportsPage useGetBugReports/status/detail; ChatHistoryPage useGetSupportChatRooms; LabDeviceItemMasterSettings source/mapping editor.

状態/無修正境界: Fetched page implementations lack a keyword list-search contact. Item mapping value selection remains E02/E03; no global full-text search invented.
- [frontend/src/features/support/routes/BugReportsPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/support/routes/BugReportsPage.tsx) blob f36e954a0b41f6a20aee363d99c2cac5aa055666
- [frontend/src/features/support/routes/ChatHistoryPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/support/routes/ChatHistoryPage.tsx) blob a4931013570e8965a7a03909da4cad0674183951
- [frontend/src/features/master/routes/LabDeviceItemMasterSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/LabDeviceItemMasterSettings.tsx) blob 223487f434799216152de263fca69678d43ce8b3

具体差分/再利用: keyword修正なし: Fetched page implementations lack a keyword list-search contact. Item mapping value selection remains E02/E03; no global full-text search invented. 共通C01の非検索操作は別operation-family台帳へ。部品名だけで一括置換しない。

### E01 N-purpose

担当: EMR-450。consumer境界: 490/491

実接点: OwnerForm.usePostalCodeLookup → address completion callback passed to OwnerInfoSection.

状態/無修正境界: Postal address lookup, not saved owner list search. Preserve explicit lookup and manual entry fallback.
- [frontend/src/features/owners/routes/OwnerForm.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/owners/routes/OwnerForm.tsx) blob 9d3c6c10a50af7715ed88f2d28e1f018366237b4
- [frontend/src/features/owners/components/OwnerInfoSection.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/owners/components/OwnerInfoSection.tsx) blob 8a358aaaebf3ee9a602806ccbde0c3e90ac4b046
- [frontend/src/features/owners/hooks/use-postal-code-lookup.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/owners/hooks/use-postal-code-lookup.ts) blob 4e9dde52fbecf26466807cdc8834db2b4cbe942e

具体差分/再利用: keyword修正なし: Postal address lookup, not saved owner list search. Preserve explicit lookup and manual entry fallback. 共通C01の非検索操作は別operation-family台帳へ。部品名だけで一括置換しない。

### E02 N-purpose

担当: EMR-450。consumer境界: 453/447/516/517

実接点: TreatmentSearchDialog searchTerm/deferredSearchTerm/filterItems and selection; MasterSelectModal ClearableSearchInput.onChange → setSearchTerm → filtered items.

状態/無修正境界: Input-assistance selection intentionally remains incremental. Choice does not execute treatment or financial finalization.
- [frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx) blob dbddcf279041b8140a5160937a5fa837db0868b6
- [frontend/src/components/shared/MasterSelectModal/MasterSelectModal.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/MasterSelectModal/MasterSelectModal.tsx) blob 34a24c5ef481a56f604ce5dac7aaae2b6dc708f6

具体差分/再利用: keyword修正なし: Input-assistance selection intentionally remains incremental. Choice does not execute treatment or financial finalization. 共通C01の非検索操作は別operation-family台帳へ。部品名だけで一括置換しない。

### E03 N-purpose

担当: EMR-450。consumer境界: 448/492/493

実接点: CarePlanRefSelect.SearchableSelect chooses medicine/procedure/plan ref; HospitalizationFormBody.MasterSelectModal.onSelectDoctor; StaffSelectionModal staff selection.

状態/無修正境界: Candidate values, not saved-list business queries. Do not globally rewrite SearchableSelect or MasterSelectModal by component name.
- [frontend/src/features/hospitalization/components/CarePlanTab/CarePlanRefSelect.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/hospitalization/components/CarePlanTab/CarePlanRefSelect.tsx) blob 15c2cad2500733650b52e714311334c8b22abf58
- [frontend/src/features/medical-records/components/StaffSelectionModal.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/medical-records/components/StaffSelectionModal.tsx) blob 1a1c4c5ee4161a4f80b172a18584acbc1045d185
- [frontend/src/features/hospitalization/routes/HospitalizationFormBody.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/hospitalization/routes/HospitalizationFormBody.tsx) blob b8acc9e3425a56255026ef4c9e908a79975352f1

具体差分/再利用: keyword修正なし: Candidate values, not saved-list business queries. Do not globally rewrite SearchableSelect or MasterSelectModal by component name. 共通C01の非検索操作は別operation-family台帳へ。部品名だけで一括置換しない。

### E04 N-purpose

担当: EMR-450。consumer境界: 502/513; 438既存履歴

実接点: ManualPage.query/setQuery → deferredQuery → useManualSearch(allMergedArticles).

状態/無修正境界: Article discovery is guidance browsing; preserve fuzzy article search and keep override editing separate.
- [frontend/src/features/manual/routes/ManualPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/manual/routes/ManualPage.tsx) blob 4f75337e51cc2320ed3ef7798fddee65f7e003d9
- [frontend/src/hooks/use-manual-search.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/hooks/use-manual-search.ts) blob e8016b1738ac16fb3bb6c0cbaa5a3c8ab5697cb6

具体差分/再利用: keyword修正なし: Article discovery is guidance browsing; preserve fuzzy article search and keep override editing separate. 共通C01の非検索操作は別operation-family台帳へ。部品名だけで一括置換しない。

### E05 N-purpose

担当: EMR-450。consumer境界: 469/504

実接点: CheckupSyncFilterForm.onSearch → CheckupSyncPage.handleSearch → CheckupSyncParams/useGetCheckupSyncPreview; LINE card currently uses line ID confirmation/token operations.

状態/無修正境界: Sync preview is explicitly executed external-action target review, not owner/pet list search. Legacy LINE customer candidate picker was not found in current OwnerForm card source; do not claim an existing candidate handler there.
- [frontend/src/features/lstep/routes/CheckupSyncPage.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/lstep/routes/CheckupSyncPage.tsx) blob 5c3458c25f56915b84f72a0aa16eb3f3d670d084
- [frontend/src/features/lstep/components/CheckupSyncFilterForm.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/lstep/components/CheckupSyncFilterForm.tsx) blob 0f1586ca85ccbd1ade060d1b4df25b85713e645c
- [frontend/src/features/owners/components/LineIntegrationCard.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/owners/components/LineIntegrationCard.tsx) blob dc2617148b8d9565ef43c4c4fe8498cc7c641ac6
- [frontend/src/features/owners/hooks/use-line-integration-card-state.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/owners/hooks/use-line-integration-card-state.ts) blob 4f3f648e68f24d41479787e9afd34e7f96d5a103

具体差分/再利用: keyword修正なし: Sync preview is explicitly executed external-action target review, not owner/pet list search. Legacy LINE customer candidate picker was not found in current OwnerForm card source; do not claim an existing candidate handler there. 共通C01の非検索操作は別operation-family台帳へ。部品名だけで一括置換しない。

優先順: 明示保留: Lstep/DMマーケ仕様整備はPhase1内最後。既存共通DataStates/検索契約の適用先としてのみ記録し③先行必須gateから外す。

### E06 N-purpose

担当: EMR-450。consumer境界: 494/516/453

実接点: ExaminationFilter and ImageGalleryFilter local inputs; PropertyFilter FilterAddPopover/SortPopover candidate-menu search.

状態/無修正境界: Current-record item/media and menu-value discovery, not historical business search H01-H06. Keep local behavior separate from underlying media saves.
- [frontend/src/features/medical-records/components/ExaminationFilter.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/medical-records/components/ExaminationFilter.tsx) blob ef4aeae1061a70e95bdea93fc89471eb4bad583a
- [frontend/src/features/medical-records/components/ImageGalleryFilter.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/medical-records/components/ImageGalleryFilter.tsx) blob 998bb5b28d1367d57577dbee5197d052f1029d7d
- [frontend/src/components/shared/PropertyFilter/FilterAddPopover.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/PropertyFilter/FilterAddPopover.tsx) blob 9b0214b67d319d61aa21ae187d20cbbf8511efab
- [frontend/src/components/shared/PropertyFilter/SortPopover.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/components/shared/PropertyFilter/SortPopover.tsx) blob c4299aa5c7bb8a4af470af3b4aaa4ddb04e9e555

具体差分/再利用: keyword修正なし: Current-record item/media and menu-value discovery, not historical business search H01-H06. Keep local behavior separate from underlying media saves. 共通C01の非検索操作は別operation-family台帳へ。部品名だけで一括置換しない。

### E07 N-purpose

担当: EMR-450。consumer境界: 490/491/494/488

実接点: Owner form create/update validation path; useMedicalRecordForm.useMedicalRecordSaveAction/useMedicalRecordQuickPatchActions/useMedicalRecordAutoCreate. useTreatmentDoseGate.useMemo → calculateDose(doseCalcInput); handleRetryDoseParamsLookup is parameter read, not keyword. Owner useOwnerForm.formAction → validateOwnerForm → create/update; selected contact-warning contract remains #490/#491 handoff.

状態/無修正境界: ブラウザ内検索はapp handlerなし。臨床計算は自動派生で検索button待ちへ変更しない。登録validateOwnerFormは既存形式検証で連絡先候補warning実装の証拠ではない。
- [frontend/src/features/owners/hooks/use-owner-form.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/owners/hooks/use-owner-form.ts) blob 5d883b7510dc3cbaa28f04c56023237e457c276e
- [frontend/src/features/medical-records/hooks/use-medical-record-form.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/medical-records/hooks/use-medical-record-form.ts) blob 4e74a0c58ce321ec8f9d8267dfc9f775bd030e18
- [frontend/src/features/medical-records/hooks/use-treatment-dose-gate.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/medical-records/hooks/use-treatment-dose-gate.ts) blob 46b896e48b1a992390dd1dbb19730c9fa1a3d64d
- [frontend/src/lib/medicine-dose.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/lib/medicine-dose.ts) blob 677349cd0398b135e8a60e12358f66cde4fb803a
- [frontend/src/features/owners/hooks/use-owner-form-model.ts](https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/frontend/src/features/owners/hooks/use-owner-form-model.ts) blob 11c496dc21093d4d8c79c7b2333099a7dc013322

具体差分/再利用: keyword修正なし: Purpose exclusion for browser find, registration safety checks and derived/automatic actions is specified. Browser find has no application handler; duplicate-check and clinical calculation specific functions were not established by these fetched excerpts and remain unverified rather than invented. 共通C01の非検索操作は別operation-family台帳へ。部品名だけで一括置換しない。

将来受領:
- owner: [490, 491]
- contract: docs/work/emr/emr-490-492-selected-wire-contract-20261021.md
- subset: contact-candidate authorized result、duplicate別人/同一確認、VerifiedCommand/ContextGate、registration receipt。連絡先一致を同一人物/成功receiptとしない
- stage: ④専用業務adapter; E07 keyword変更対象外

### N-PKG N-purpose

担当: EMR-450。consumer境界: 523

実接点: TreatmentPackageSettings useTreatmentPackages; handleSave/handleDeleteConfirm/handleMove and MasterPageShell editor.

状態/無修正境界: Existing package editor/list has no keyword search field. Exclude only keyword change; general save/delete/reorder state belongs the independent P01 package operation family.
- [frontend/src/features/master/routes/TreatmentPackageSettings.tsx](https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/frontend/src/features/master/routes/TreatmentPackageSettings.tsx) blob 5ad8ec17f57d5f34e33eafc242e471c4da382352

具体差分/再利用: keyword修正なし: Existing package editor/list has no keyword search field. Exclude only keyword change; general save/delete/reorder state belongs the independent P01 package operation family. 共通C01の非検索操作は別operation-family台帳へ。部品名だけで一括置換しない。

## 現在接点の最小差分

- G1 M09: use-reservation-type-settings のtype/group query loading/error/refetchを ReservationTypeSettings → ReservationTypeSettingsContentへ通す。450/492。
- G2 M20: ShiftTemplateSettings のquery状態を Workspace → Listへ通す。450/448/493。514既存日時CRUDを再実装しない。
- G3 D09: CheckupFormのdirty/戻る確認を既存useUnsavedChanges/NavigationBlockerへ接続。450/500。段階保存の成功ID/lookupは④の500受領。
- H05追加: 上記CheckupFormの履歴GET error/retry接続。450、G3 dirtyとは別。
- P01追加: package masterのSPA戻るdirty保護。523/450、非検索操作表の追加受入へ。
- G4 D15: CsvImportHistoryTableのerror/refetchを0件分岐前へ通しDataStatesを再利用。450/504。マーケ優先順により後順明示保留。

## 計画・実装の終了条件

各行の用途、source接点、最小差分又は無修正理由、owner、必要subsetが確認できることを計画の終了条件とする。実装後の受入は別途、draft/applied・clear・context復元・状態・race・scope・無副作用を各consumerで記録する。共通部品1件の成功を全callerへ代入しない。臨床値承認待ち・型生成進行・runtime未完・試験待ちは別状態で管理する。

S13は現inline handlerなし。既存S07と重複計上せず、451/517がdistinct inline flowか既存入口との置換関係を確定するまで新UI実装を開始しない。E05旧LINE候補pickerは現カードsourceに所在を確立していないため既存実装済みとしない。E07登録重複warningは490/491選定wire受領であり、形式validationから適合済みとしない。

## 限定静的監査の範囲（監査追補）

基準main `c86c8dd2530bbea3d1810e07badb72f3ed1a07e1`。今回の認定はH05初回履歴エラーと、[非検索P01操作表](emr-450-stage3-common-ui-start-plan-20261021-operations.md#package-operation-audit-20261021)のSPA戻るdirty接続差だけで、全72用途・全実操作familyの受入完了ではない。P01の部品名訂正は名称整合であり単独の欠陥ではない。focus等はNOT_RUN。create結果不明後の一覧refetch→allowCreateRetryは確認候補で、確定二重作成の欠陥件数へ加算せず、③へ新しい専用lookupを強制しない。試験・runtime修正・公開flag変更なし。

## 2026-10-22 coverage監査の追加受領

[利用者向け監査記録](emr-stage123-coverage-audit-20261022.md)に母集団・静的照合範囲・非欠陥・未閉鎖設計を分けた。検索72用途は検索対象57＋非keyword/除外15で、未来S13・複数入口H03を含み72実装済みではない。非検索D01–D23/P01、7primitive＋18storiesとは別集計。18storiesは既存C01契約へ分類可能で、18件の実装漏れではない。

450への限定受領は、非検索操作表の治療内容Enter caller、SubmitButton pending wrapper差分、EMR-245のカルテ3列GET採用と他一覧filter後続の継承確認。245 Doneを全一覧filter完成の証拠にせず、追加filterを全画面へ義務化しない。318受付write・400/#702 staff保存の部分成果は③表示/④writeを分けて受領確認する。計画記載済みをMECE性・実装完成・試験保証へ換算せず、試験NOT_RUN。

# FE-refactor — FE12-02 active-only ledger

> 更新: 2026-07-26（要件責任者: 曽我）
> 業務目的: 未完了の臨床sentinel・RBAC安全境界を、決裁と実測の証跡が揃うまで追跡する。
> 本ファイルは使い捨てのactive ledger。恒久規約は `DESIGN.md`、`docs/spec/design-system.md`、`docs/spec/ui-design-compliance.md`、`frontend/CLAUDE.md` を正本とする。

## Active scope and authority

- 追跡対象は `FE12-02` の未完了route、`F16`・`F9`、`U10`、`M-01`〜`M-05`、line-reserve font実機確認、C6a安全境界、未裁定riskだけとする。
- 色と臨床semanticは `docs/spec/design-system.md`、恒久route適合は `docs/spec/ui-design-compliance.md`、明示的なPO/USER裁定は `q&a.html` を正本とする。
- authorityから項目が消えたことや判断待ち件数が0であることだけでは完了とみなさない。明示的な決裁または実測証跡が無い項目は保持する。
- 本ledgerの更新は実装・runtime検証・製品決裁を代替しない。

## Active routes

<!-- FE12-ROUTE-TABLE-START -->
| エリア | ページ | パス | コンポーネント | 未完了事項 |
|---|---|---|---|---|
| カルテ | カルテ一覧 | /medical-records | MedicalRecords | C6a danger/異常/RBACレビュー |
| カルテ | カルテ作成 - ペット選択 | /medical-records/select-pet | MedicalRecordPetSelection | C6a死亡/危険ペット選択 |
| カルテ | カルテ作成 | /medical-records/new | MedicalRecordForm | C6a danger/異常/RBACレビュー |
| カルテ | カルテ編集 | /medical-records/:id | MedicalRecordForm | C6a danger/異常/RBACレビュー |
| 入院/ホテル | 入院・ホテル一覧 | /hospitalization | HospitalizationList | C6a死亡表示・操作抑止 |
| 入院/ホテル | 入院・ホテル登録 - ペット選択 | /hospitalization/select-pet | HospitalizationPetSelection | C6a死亡/危険ペット選択 |
| 入院/ホテル | 入院・ホテル登録 | /hospitalization/new | HospitalizationForm | C6a死亡/危険/RBACレビュー |
| 入院/ホテル | 入院・ホテル詳細 | /hospitalization/:id | HospitalizationDetail | C6a死亡表示・child mutationレビュー |
| 入院/ホテル | 入院・ホテル編集 | /hospitalization/:id/edit | HospitalizationForm | C6a死亡/危険/RBACレビュー |
| 検査 | 検査一覧 | /examinations | ExaminationsList | C6a異常値/RBAC |
| 検査 | 検査登録 - ペット選択 | /examinations/select-pet | ExaminationPetSelection | C6a死亡/危険ペット選択 |
| 検査 | 検査登録 | /examinations/new | ExaminationForm | C6a異常/RBACレビュー |
| 検査 | 検査編集 | /examinations/:id | ExaminationForm | C6a異常/RBACレビュー |
| ワクチン | ワクチン一覧 | /vaccinations | VaccinationList | C6a期限超過/死亡/RBAC |
| ワクチン | ワクチン接種 - ペット選択 | /vaccinations/select-pet | VaccinationPetSelection | C6a死亡/危険ペット選択 |
| ワクチン | ワクチン登録 | /vaccinations/new | VaccinationForm | C6a期限超過/RBACレビュー |
| ワクチン | ワクチン編集 | /vaccinations/:id | VaccinationForm | C6a期限超過/RBACレビュー |
| 定期健診 | 定期健診一覧 | /checkups | CheckupsList | C6a要フォロー表示 |
| 定期健診 | 定期健診登録 - ペット選択 | /checkups/select-pet | CheckupPetSelection | C6a死亡/危険ペット選択 |
| 定期健診 | 定期健診登録 | /checkups/new | CheckupForm | C6a臨床status |
| 受付/飼主/予約 | 受付 | / | Reception | C6a危険/死亡ペットの受付表示 |
| 受付/飼主/予約 | 飼主一覧 | /owners | OwnersList | C6a danger/deceased filterレビュー |
| 受付/飼主/予約 | 飼主登録 | /owners/new | OwnerForm | C6a死亡/危険/RBACレビュー |
| 受付/飼主/予約 | 飼主編集 | /owners/:id | OwnerForm | C6a死亡/危険/RBACレビュー |
| 設定/マスタ | 権限グループマスタ | /settings/permission-groups | PermissionGroupSettings | C6a RBACレビュー |
<!-- FE12-ROUTE-TABLE-END -->

## Active task

<!-- FE12-TASK-TABLE-START -->
| ID | Priority | Active frontier | Dependency | Completion evidence |
|---|---|---|---|---|
| FE12-02 | P0 | F16・F9の決裁、M-01〜M-05、line-reserve font実機確認、C6aの臨床/RBAC実測 | U10はF16の決裁待ち | 明示的な決裁と対象実測の証跡が揃うこと |
<!-- FE12-TASK-TABLE-END -->

## Authority drift

- `q&a.html` はPO判断待ち0件と記載する一方、F16/F9の明示裁定を含まない。件数要約から完了を推測せず、F16・F9・U10を保持する。
- `docs/spec/ui-design-compliance.md` にはC18件数ratchetの旧記述が残るが、現行auditはratchetを持たない。本runでは正本側を編集せず、別scopeの文書driftとして扱う。

## C6a 臨床安全レビュー

- 危険/死亡: OwnersList、Reception、各PetSelection、HospitalizationListでsemantic色・明示文言・操作抑止が同時に保たれること。
- 異常/期限: MedicalRecords、Examinations、Vaccinations、Checkupsで正常値をdanger表示せず、異常/期限超過だけをsemantic色と非色手掛かりで示すこと。
- RBAC: PermissionGroupSettingsと各clinical actionで非活性表示だけに頼らず、mutation不能・accessible name維持を確認する。
- 静的レビューで閉じず、残件が実装認可された場合は既存sentinel fixtureのscoped component testを先にRED化してから修正する。

## Active execution rules

決裁・実測後に対応する場合も、次の安全境界を維持する。

1. **臨床sentinelは生成型から表示・操作境界まで欠落させない。** 死亡は明示的なpositive matchで遷移・mutation callbackを拒否し、危険「高」は非色cueを伴う警告として扱う。死亡statusと死亡日時が不整合なら再登録導線を出さない。
2. **権限はaction別の最新値をmutation直前に再検査する。** UIの非表示・disabled・route guardだけを最終防壁にしない。view/edit共用の唯一のdetail routeはread accessを維持し、mutation境界をfail-closedにする。commit直後にも発火し得るcallbackのpermission refは`useLayoutEffect`で同期する。
3. **臨床date-onlyはJSTの厳密過去で判定する。** `YYYY-MM-DD`契約をguardし、`todayJSTISO()`との文字列比較`<`を使う。現在時刻との`Date`比較で当日を期限超過にしない。

## 25 routeの静的検証結果

判定語は次の3値に限定する。`ENFORCED`はcurrent sourceでbinding ruleを再現できる、`GAP`はcurrent sourceがbinding ruleへ明確に違反する、`RUNTIME-ONLY`はseed/persona/browser実測または製品裁定なしにsourceだけでは閉じられない、を表す。

| # | パス | Verdict | current source evidence |
|---:|---|---|---|
| 1 | `/medical-records` | ENFORCED | `frontend/src/features/medical-records/api/transforms.ts:23-35`、`frontend/src/features/medical-records/routes/MedicalRecords.tsx:162-180,316-340,363-379` — backend pet deathを保持し、死亡時のedit/deleteを抑止、delete確定時もlatest permission/record状態を再検査する。 |
| 2 | `/medical-records/select-pet` | ENFORCED | `frontend/src/hooks/use-pet-selection-page.ts:65-71` — `status === "死亡"`をpositive matchし遷移を拒否する。 |
| 3 | `/medical-records/new` | ENFORCED | `frontend/src/features/medical-records/hooks/use-medical-record-auto-create.ts:85-142` — 死亡を明示拒否し、予約作成前とカルテ作成前にlatest `canCreate === true`を再検査する。 |
| 4 | `/medical-records/:id` | ENFORCED | `frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts:83-90,102-171`、`frontend/src/features/medical-records/hooks/use-medical-record-quick-patch-actions.ts:37-170`、`frontend/src/features/medical-records/hooks/use-medical-record-owner-change.ts:47-101`、`frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:144-154,196-203` — save/quick patch/owner change/deleteの全top-level mutationがlatest action permissionと明示死亡を再検査する。 |
| 5 | `/hospitalization` | ENFORCED | `frontend/src/features/hospitalization/hooks/use-hospitalization-list.ts:18-53` — drag/drop mutation直前にlatest `canEdit === true`とsource/target双方の`petIsDeceased`を再検査する。 |
| 6 | `/hospitalization/select-pet` | ENFORCED | `frontend/src/hooks/use-pet-selection-page.ts:65-71` — 死亡ペットの遷移をpositive matchで拒否する。 |
| 7 | `/hospitalization/new` | ENFORCED | `frontend/src/features/hospitalization/hooks/use-hospitalization-form.ts:52-89` — layout同期済みlatest selected petとcreate permissionを読み、`status === "死亡"`をmutation境界で拒否する。 |
| 8 | `/hospitalization/:id` | ENFORCED | `frontend/src/features/hospitalization/components/HospitalizationPatientHeader.tsx:13-28`、`frontend/src/features/hospitalization/components/HospitalizationDetailActions.tsx:29-46`、`frontend/src/features/hospitalization/hooks/use-hospitalization-detail.ts:17-45`、`frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx:48-59`、`frontend/src/features/hospitalization/components/CarePlanTab/CarePlanTab.tsx:31-42` — 死亡表示、check-in、退院、全child mutationをlatest permission/死亡guardで拒否する。 |
| 9 | `/hospitalization/:id/edit` | ENFORCED | `frontend/src/features/hospitalization/hooks/use-hospitalization-form.ts:52-89`、`frontend/src/features/hospitalization/hooks/use-hospitalization-form-model.ts:138-152`、`frontend/src/features/hospitalization/routes/HospitalizationForm.tsx:101-121` — backend死亡statusを保持し、save/deleteをlatest action permissionと死亡で拒否する。 |
| 10 | `/examinations` | RUNTIME-ONLY | `frontend/src/features/examinations/routes/ExaminationsList.tsx:225-234` — 一覧は`resultSummary`と検査statusを表示するが、HIGH/LOW cueを一覧にも要するかはM-02の4 viewport実測・製品確認が必要。 |
| 11 | `/examinations/select-pet` | ENFORCED | `frontend/src/hooks/use-pet-selection-page.ts:65-71` — 死亡ペットの遷移をpositive matchで拒否する。 |
| 12 | `/examinations/new` | ENFORCED | `frontend/src/features/examinations/hooks/use-examination-form.ts:105-125,273-325` — latest create permissionとdirect-query/選択petの明示死亡を親mutation前およびawait後のitems mutation直前に再検査する。 |
| 13 | `/examinations/:id` | ENFORCED | `frontend/src/features/examinations/hooks/use-examination-form.ts:105-125,273-325,350-360` — `existingExam.petId`から対象petを解決し、update/items/deleteの各境界でlatest permissionと明示死亡を拒否する。 |
| 14 | `/vaccinations` | ENFORCED | `frontend/src/features/vaccinations/routes/VaccinationList.tsx:76-103,114-137,193-212,266-285` — editは明示生存時だけ、deleteはlatest permissionと明示生存lookup時だけ許可し、期限超過は`isPastJSTDate`のJST厳密過去だけを表示する。 |
| 15 | `/vaccinations/select-pet` | ENFORCED | `frontend/src/hooks/use-pet-selection-page.ts:65-71` — 死亡ペットの遷移をpositive matchで拒否する。 |
| 16 | `/vaccinations/new` | ENFORCED | `frontend/src/features/vaccinations/hooks/use-vaccination-form.ts:126-147,252-274` — layout同期済みlatest selected/query petとcreate permissionを読み、明示死亡をmutation境界で拒否する。 |
| 17 | `/vaccinations/:id` | ENFORCED | `frontend/src/features/vaccinations/hooks/use-vaccination-form.ts:130-147,232-250,361-374` — authoritative edit petをlayout同期し、update/deleteの双方がlatest action permissionと明示死亡を拒否する。 |
| 18 | `/checkups` | RUNTIME-ONLY | `frontend/src/features/checkups/routes/CheckupsList.tsx:297-305` — 次回期限とraw resultのみで、要フォローfieldはF9 full-readで不在。表示要否は曽我の裁定が必要。 |
| 19 | `/checkups/select-pet` | ENFORCED | `frontend/src/hooks/use-pet-selection-page.ts:65-71` — 死亡ペットの遷移をpositive matchで拒否する。 |
| 20 | `/checkups/new` | ENFORCED | `frontend/src/features/checkups/routes/CheckupForm.tsx:30-50`、`frontend/src/features/checkups/hooks/use-checkup-form.ts:70-82,105-151` — route配線済みlatest create/edit permissionと明示死亡を、カルテ・健診・型付き結果の各mutation前に拒否する。 |
| 21 | `/` | ENFORCED | `frontend/src/features/reception/hooks/use-reception-kanban.ts:132-140,230-266,349-369`、`frontend/src/features/reception/hooks/use-reception-modal-handlers.ts:56-89,113-185` — original/newly-selected双方の死亡とlatest reservation permissionをdrag/status/edit/cancel境界で再検査する。 |
| 22 | `/owners` | RUNTIME-ONLY | `frontend/src/features/owners/components/OwnersListTable.tsx:128-128,191-198`、`frontend/src/lib/status-helpers.ts:176-178` — view-only detail linkとRBACはsourceで再現できるが、死亡のgeneric grayと許可操作範囲はF16/M-01裁定・実測待ち。 |
| 23 | `/owners/new` | ENFORCED | `frontend/src/features/owners/hooks/use-owner-form.ts:125-131,220-241`、`frontend/src/features/owners/hooks/use-pet-form-list-state.ts:38-63,148-177` — owner/pet createはlayout同期済みlatest action permission、pet deleteは明示死亡もmutation境界で拒否する。 |
| 24 | `/owners/:id` | ENFORCED | `frontend/src/features/owners/hooks/use-pet-form-list-state.ts:38-63,84-124`、`frontend/src/features/owners/routes/OwnerForm.tsx:58-60,144-192`、`frontend/src/components/shared/PetDeceasedRecordButton/PetDeceasedDialog.tsx:55-95`、`frontend/src/components/shared/PetDeceasedRecordButton/PetDeceasedBanner.tsx:38-48` — ordinary pet update/owner変更は明示死亡とlatest permissionを拒否し、専用死亡登録/解除もlatest editを再検査する。 |
| 25 | `/settings/permission-groups` | ENFORCED | `frontend/src/features/master/components/MasterCRUDPage.tsx:99-128`、`frontend/src/features/master/routes/PermissionGroupSettings.tsx:46-56,91-124` — create-only panelをwrite可能にし、reorderとpost-save rulesはlatest action permissionを再検査する。 |

## FE12-02 残件

### 残件を位置づける監査表

| ID | Surface | 残る事項 | 対応する手順 |
|---|---|---|---|
| A-01 | OwnersList | F16 死亡表示tokenの決裁 | M-01 |
| B-04 | Examinations list/card | 一覧でHIGH/LOW cueが必要か実測 | M-02 |
| B-08 | Checkups要フォロー | F9 sentinel要否の決裁 | M-05 |
| C-02 | PermissionGroup CRUD/rules | 非活性理由・accessible nameの実測 | M-03 |
| C-04 | MedicalRecords top-level C/U/D | view-only等personaの実測 | M-03 |
| C-06 | Hospitalization top-level C/U/D | view-only等personaの実測 | M-03 |
| C-07 | Hospitalization child mutations | 複合controlの実効性確認 | M-04 |
| C-08 | Vaccinations C/U/D | view-only等personaの実測 | M-03 |
| C-09 | Examinations C/U/D/items | view-only等personaの実測 | M-03 |
| A/B横断 | clinical sentinel responsive | 4 viewportの非色cueとlayout確認 | M-05 |

### 決裁待ち

#### FE12-02-F16: OwnersListの死亡表示がgeneric gray token

- Authority A: `docs/spec/design-system.md:84` は「**危険 / 死亡 / 異常高**: `C.danger`」と明記する。
- Authority B: 同じ§2.4の `docs/spec/design-system.md:89` は「危険バッジ・**死亡グレーアウト**・RBAC 非活性表示」を退行させないと明記する。AとBは死亡表現について相互に競合し、どちらも本unitでは優先しない。
- Current code: `frontend/src/lib/status-helpers.ts:176-178` の`getPetStatusColor`は`status === "生存"`だけをgreen、それ以外をすべて`BADGE.grayHover`へまとめる。死亡だけでなく未知の非生存statusも同じgeneric grayとなる。
- Decision request: 要件責任者=曽我が、(a) generic grayを死亡表示の正本とする、または(b) `C.danger`を死亡へ適用してauthorityの「死亡グレーアウト」文を改訂する、のいずれかを明示裁定する。本ledgerは選択しない。
- Dependency: U10のHospitalizationList/Board death token置換はこの決裁まで着手しない。

#### FE12-02-F9: Checkupsに要フォローsentinel経路がない

- Enumeration command（verbatim）:

```bash
rg -l --glob '*.go' --glob '!**/*_test.go' '^type (Checkup$|Checkup struct|CheckupFieldType|CheckupTypeField|CheckupFieldResult|checkup(TypeField|FieldResult|Response|GlobalResponse|TypeResponse|Field.*Response|.*Response)|petCheckupResultResponse|createCheckupRequest|updateCheckupRequest|listGlobalCheckupsQuery|upsertCheckupFieldResultRequest|replaceCheckupFieldResultsRequest|CreateCheckupInput|UpdateCheckupInput|ListCheckupsByClinicInput|UpsertCheckupFieldResultInput|CheckupFilters)' backend/internal/model backend/internal/medicalrecord | LC_ALL=C sort
```

- Complete backend file list:

```text
backend/internal/medicalrecord/checkup_field_repository.go
backend/internal/medicalrecord/checkup_field_request.go
backend/internal/medicalrecord/checkup_field_response.go
backend/internal/medicalrecord/checkup_field_result_service.go
backend/internal/medicalrecord/checkup_repository.go
backend/internal/medicalrecord/checkup_request.go
backend/internal/medicalrecord/checkup_response.go
backend/internal/medicalrecord/checkup_service.go
backend/internal/medicalrecord/checkup_type_response.go
backend/internal/model/checkup_field.go
backend/internal/model/checkup_record.go
```

- Full-read field evidence: model=`backend/internal/model/checkup_record.go:9-28`（`Result`は自由記述）、`backend/internal/model/checkup_field.go:11-21,36-51,58-84`（`ValueText`/`ValueList`/`IsAbnormal`/`Status`のみ）。request/response DTO=`backend/internal/medicalrecord/checkup_request.go:10-17,45-53,85-91`、`backend/internal/medicalrecord/checkup_response.go:11-26,59-75`、`backend/internal/medicalrecord/checkup_field_request.go:5-19`、`backend/internal/medicalrecord/checkup_field_response.go:11-22,44-58,83-88`、`backend/internal/medicalrecord/checkup_type_response.go:10-23`。service/repository carrier=`backend/internal/medicalrecord/checkup_service.go:16-46`、`backend/internal/medicalrecord/checkup_repository.go:20-43`、`backend/internal/medicalrecord/checkup_field_result_service.go:16-22,39-50,270-272`、`backend/internal/medicalrecord/checkup_field_repository.go:25-39`。frontend DTO=`frontend/src/features/checkups/api/types.ts:1-25`。
- Full-read conclusion: dedicatedな「要フォロー」field、boolean、enum、または同義のDTO contractは存在しない。自由記述`result`と動的fieldの`is_abnormal`/`status`は存在するが、どれを「要フォロー」へ写すかという契約はないため代用しない。
- Decision request: 要件責任者=曽我が要フォローsentinelの要否と意味を裁定する。必要と裁定された場合だけ、別unitでbackend field/API/type/UIを設計する。本unitはfieldを追加しない。

### 要実測項目

#### M-01 OwnersList操作範囲

- Route: `/owners?include_deceased=true`。
- Fixture seed source: disposable local `003_demo`へ`backend/migrations/seeds/003_demo/owners.csv`と`pets.csv`を適用し、`docs/ops/testing/scenarios/S01-deceased-pet-guard.md`/`V03-owner-pet-staff-forms.md`の手順で同一ownerにalive、danger=`高`、deceasedの3頭を準備する。実測後はS01どおり生存へ戻す。
- Persona: owners `view=true, edit=true, delete=true`の通常担当者。曽我または同席する仕様責任者が許可操作を記録する。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: include-deceasedを有効にし、3 rowそれぞれでowner名link、report、編集、削除、pet死亡登録/解除をpointer、Tab/Enter/Space、直接URLで試す。各操作前後のrequestをnetworkで記録する。
- Expected result: aliveは通常操作可、danger=`高`は「⚠ 危険」等の非色cueとaccessible nameを失わない。deceasedはF16裁定前の現行grayを記録し、曽我が許可/禁止する操作を操作単位で明示する。禁止とされた操作はmutation 0件。
- Required evidence artifacts: 4 viewport screenshot、accessibility tree、各操作のaccessible name、GET以外のnetwork HAR、操作可否の曽我裁定メモ。

#### M-02 Examinations一覧意味とlayout

- Route: `/examinations`と対象petの`/medical-records/:id`検査履歴。
- Fixture seed source: `backend/migrations/seeds/003_demo/exams.csv`、`exam_results.csv`、`exam_types.csv`、`exam_type_fields.csv`を基に、`docs/ops/testing/scenarios/S02-exam-abnormal-highlight-lock.md`のhigh/low/normal同居fixtureを作る。
- Persona: examinations `view=true, create=false, edit=false, delete=false`のview-only担当者。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: 同一fixtureを一覧とカルテ履歴で開き、値・summary・statusを比較する。zoom 100%、pointer hover、keyboard focus、横スクロール有無を確認し、一覧でHIGH/LOW非色cueが必要か曽我へ提示する。
- Expected result: normalを異常表示しない。曽我が一覧cueを必要と裁定する場合はHIGH/LOWが色なしでも識別できること、不要と裁定する場合は詳細へのaccessible導線があること。全viewportでwrap/clip/overlapなし。
- Required evidence artifacts: 両surface×4 viewport screenshot、accessible text dump、computed color/token、overflow計測、曽我の一覧cue要否裁定。

#### M-03 RBAC非活性の理由/name

- Route: `/settings/permission-groups`、`/medical-records/:id`、`/hospitalization/:id/edit`、`/vaccinations/:id`、`/examinations/:id`。
- Fixture seed source: `backend/migrations/seeds/003_demo/permission_groups.csv`、`permission_group_rules.csv`、`staffs.csv`、`staff_permission_groups.csv`と各featureの既存record CSV。`docs/ops/testing/scenarios/V03-owner-pet-staff-forms.md`で試験用group/staffを作り、`V01-clinical-forms.md`の既存recordを使う。
- Persona: (1) view-only=`view:true/create:false/edit:false/delete:false`、(2) create-only=`view:true/create:true/edit:false/delete:false`、(3) edit-without-delete=`view:true/create:false/edit:true/delete:false`。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: 各personaでrouteを再ログイン/再読込して開き、pointer、Tab/Enter/Space、formのprogrammatic submit、保存中の権限剥奪後callbackを試す。permission-groupは新規panel、既存panel、reorder、保存後rulesも個別確認する。
- Expected result: view accessは維持する。許可されたactionだけ実行でき、禁止controlはaccessible nameと理由を保持する。禁止personaからのPOST/PATCH/PUT/DELETEは0件で、same-commit剥奪後もmutationしない。
- Required evidence artifacts: persona×routeのaccessibility tree、4 viewport screenshot、network HAR、console log、action別permission matrixと0 mutation集計。

#### M-04 Hospitalization child control実効性

- Route: `/hospitalization`のboard/listと`/hospitalization/:id`。
- Fixture seed source: `backend/migrations/seeds/003_demo/hospitalizations.csv`、`daily_records.csv`、`vital_records.csv`、`care_logs.csv`、`care_plan_items.csv`、`cages.csv`を基に、`docs/ops/testing/scenarios/S05-hospitalization-cycle.md`のadmitted fixtureを準備する。
- Persona: hospitalization view-only=`view:true/create:false/edit:false/delete:false`。対照として通常担当者=`view/create/edit/delete:true`を1回使う。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: board drag/drop、check-in/status、退院（会計あり/なし）、daily、vital、care log、note、care planのcreate/edit/deleteをpointer/keyboard/programmatic callbackで試す。操作dialogを開いた後に権限を剥奪するcaseも含める。
- Expected result: view-onlyおよびsame-commit剥奪後は全child/top-level mutation 0件。死亡fixtureは死亡文言を表示し、drag/check-in等を実行しない。非活性controlのnameと理由は残る。
- Required evidence artifacts: 操作別network HARと0件集計、accessibility tree、4 viewport screenshot、console log、会計あり/なし退院の個別記録。

#### M-05 Clinical sentinel responsive

- Route: 本表25 routeのうちclinical sentinelを表示する`/medical-records`系、`/hospitalization`系、`/examinations`系、`/vaccinations`系、`/checkups`系、`/`、`/owners`。
- Fixture seed source: `003_demo`の`pets.csv`、`medical_records.csv`、`hospitalizations.csv`、`exams.csv`/`exam_results.csv`、`vaccinations.csv`、`checkups.csv`を基に、S01/S02/S03/S05のfixtureを合成してdeath、danger=`高`、HIGH、LOW、past、today、future、emptyを各1件以上用意する。
- Persona: 対象resourceの`view=true`を持つ通常担当者。mutation確認が必要なrowだけ該当action権限ありpersonaを併用する。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: 各fixtureを一覧、選択、登録、編集、詳細で開き、文言、badge、日付、disabled/hidden control、keyboard focus順を確認する。期限は同じ実測日にpast/today/futureを並べる。
- Expected result: death/danger/HIGH/LOW/期限超過が非色cueを持ち、normalとtoday/futureを誤ってdangerにしない。死亡操作はpositive matchで拒否され、全viewportでcue/controlのwrap、clip、overlapなし。
- Required evidence artifacts: route×4 viewport screenshot、accessible name/text dump、computed token、console/network HAR、fixture-to-cue対応表。

#### line-reserve font実機確認

- Route: 顧客向け`/line-reserve/{clinicId}`の顧客情報→要望→確認→完了とマイ予約（clinicId抽出契約=`frontend/line-reserve/src/lib/liff-config.ts:6-14`）。
- Fixture seed source: `backend/migrations/seeds/003_demo/line_reservation_settings.csv`、`reservation_types.csv`、`line_customers.csv`、`pets.csv`を基に`docs/ops/testing/scenarios/V05-auth-line-forms.md` V05-6/V05-7、`S04-liff-reservation-journey.md`の試験用顧客を使う。
- Persona: LINE連携済み顧客persona。業務に影響しない試験用アカウントを使い、送信/予約確定はrunbookの試験環境だけで行う。
- Viewports: iPhone Safari 390×844、Android Chrome 412×915、iPad Safari 768×1024（加えてdesktop 500×900を比較用）。
- Interaction steps: physical deviceでcold loadし、DevTools/remote inspectionのNetworkでGoogle Fonts CSSとfont fileが200であることを確認する。顧客情報から完了/マイ予約まで遷移し、各画面のcomputed `font-family`と実レンダーfontを確認する。offline/reload時のfallbackも確認する。
- Expected result: `frontend/line-reserve/index.html:7-12`からNoto Sans JPがloadされ、`frontend/line-reserve/src/index.css:17-23`の先頭fontとして全画面へ適用される。clip/FOITによる操作不能がなく、font失敗時もfallbackで操作可能。
- Required evidence artifacts: 3実機の画面別screenshot、remote Network HAR、computed font-familyとRendered Fonts capture、端末/OS/browser/version、cold/warm/offline各結果。

## FE12-02 unit execution record

### Harness / boundary

- Saved prompt validation: `node /Users/minoru/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fast-fe12-02-active-ledger.md` → exit 0、`Prompt Craft Harness Validation: PASS`。
- Risk tier: Local write。外部送信、commit/push、authority文書、container lifecycle、DB reset/migrationは実行していない。
- Harness: `tdd-workflow` + `react-testing` + project `scoped-verification-gates`を使用。native Workflow toolはこのsessionに無いため、planner/read-only probe/writer/reviewerのparallel multi-agent fan-outで代替した。
- Repair partition: F16 token、F9 field、U10、M-01/M-02の製品判断は修理gate(b)または製品裁定を満たさないため未変更。sourceで再現でき、F16に依存しない死亡positive-match、latest action permission、JST strict-pastの欠落だけをTDD修理した。

### RED→GREEN evidence

下記は各writerが同一pathのrepair前後に実行したscoped commandから、laneごとのpass/fail counter lineを抽出した表である。完全なcommand出力そのものではない。最終integrationでは同じfeatureをpath指定で再実行する。

| Lane | RED output | GREEN output |
|---|---|---|
| medical-records list/auto-create | `Test Files 1 failed (1)` / `Tests 1 failed \| 11 passed (12)`、および`Tests 3 failed \| 16 passed (19)` | list=`12 passed`、auto-create=`19 passed` |
| medical-record edit latest permission | `Test Files 1 failed (1)` / `Tests 1 failed (1)`、`expected "vi.fn()" to not be called ... called 1 times` | `Test Files 1 passed (1)` / `Tests 1 passed (1)` |
| examinations permission/death | permission RED=`1 failed \| 34 passed (35)`、direct deceased RED=`2 failed \| 35 passed (37)` | `Test Files 1 passed (1)` / `Tests 37 passed (37)` |
| hospitalization list/form/model | list=`2 failed \| 1 passed`、form=`2 failed \| 29 passed`、model=`2 failed \| 11 passed` | list=`3 passed`、form=`31 passed`、model=`13 passed` |
| hospitalization detail/children | `Test Files 4 failed (4)` / `Tests 4 failed \| 8 passed (12)` | `Test Files 9 passed (9)` / `Tests 24 passed (24)` |
| vaccinations | list=`1 failed \| 6 passed (7)`、form=`3 failed \| 36 passed (39)` | list=`7 passed`、form=`39 passed`、permissions=`1 passed` |
| checkups | `Test Files 2 failed (2)` / `Tests 10 failed \| 2 passed (12)` | `Test Files 2 passed (2)` / `Tests 12 passed (12)` |
| reception | `Test Files 3 failed (3)` / `Tests 12 failed \| 36 passed (48)` | focused=`48 passed`、feature regression=`109 passed` |
| OwnersList | table=`1 failed \| 6 passed (7)`、delete=`1 failed (1)` | `Test Files 4 passed (4)` / `Tests 15 passed (15)` |
| OwnerForm/pet/death | initial=`8 failed \| 22 passed`、追加=`2 failed \| 10 passed` | `Test Files 10 passed (10)` / `Tests 57 passed (57)` |
| PermissionGroupSettings | `Test Files 2 failed (2)` / `Tests 4 failed \| 2 passed (6)` | `Test Files 3 passed (3)` / `Tests 10 passed (10)` |

### Orchestration reconciliation

| ID / label | Role | Responsibility / writer-owned paths | Status / evidence | Integration decision |
|---|---|---|---|---|
| `fe12_plan` | planner | 8 acceptance項目、repair gate、phase/ownership設計（read-only） | completed | 8項目へ展開して採用 |
| `probe_medical_exams` | read-only probe | medical-records/examinations source全route | completed、初期verdict/citation | repair対象をwriterへ分配 |
| `probe_hospitalization` | read-only probe | hospitalization 5 route/child mutation | completed | repair対象をwriterへ分配 |
| `probe_vax_checkups` | read-only probe | vaccinations/checkups/JST | completed | F9とrepair対象へ統合 |
| `probe_reception_rbac` | security read-only probe | reception/owners/permission + shared auth | completed | repair対象へ統合 |
| `probe_f9_backend` | read-only probe | checkup model/DTO enumeration + full read | completed | F9 decision packへ統合 |
| `write_medical_records` | TDD writer | `features/medical-records`のlist/auto-create/save-action source+tests | completed、RED→GREEN | 採用 |
| `write_examinations` | TDD writer | `features/examinations/hooks/use-examination-form*` | completed、RED→GREEN | 採用 |
| `write_hosp_list_form` | TDD writer | hospitalization list/form/model source+tests | completed、RED→GREEN | 採用 |
| `write_hosp_detail` | TDD writer | hospitalization detail/actions/header/Daily/CarePlan source+tests | completed、RED→GREEN | 採用 |
| `write_vaccinations` | TDD writer | vaccination list/form hook source+tests | completed、RED→GREEN | 採用 |
| `write_checkups` | TDD writer | checkup form hook/route source+tests | completed、RED→GREEN | 採用 |
| `write_reception` | TDD writer | reception card/kanban/modal/route source+tests | completed、RED→GREEN | 採用 |
| `write_owners_list` | TDD writer | OwnersList/Table source+tests | completed、RED→GREEN | 採用 |
| `write_owner_form` | TDD writer | OwnerForm/owner-pet hooks/shared deceased control source+tests | completed、RED→GREEN | 採用 |
| `write_permission_group` | TDD writer | MasterCRUDPage/PermissionGroupSettings source+tests | completed、RED→GREEN | 採用 |
| `write_owners_list/owners_list_review` | read-only reviewer | OwnersList narrow review | explicitly interrupted after no timely result | output不採用、parent reviewへ委譲 |
| `independent_route_sample` | independent reviewer | reproducible hash sample 8/25 routeをsourceから再導出 | completed、5 HIGHを検出 | verdictを一時reopenし追加TDDへrouting |
| `react_review` | React reviewer | hook/a11y/stale callback review | completed、OwnersList HIGH修理後2/2再検証 | CRITICAL/HIGH=0、Approve |
| `typescript_review` | TypeScript reviewer | type/contract/mock review | completed、CRITICAL/HIGH/MEDIUMなし | Approve |
| `security_review` | security reviewer | RBAC/death/TOCTOU/partial mutation review | completed、全HIGH修理後にcurrent sourceを再監査 | CRITICAL=0/HIGH=0、Approve |
| `repair_hospitalization_review` | TDD writer | hospitalization detail/list/formとchild source+tests | completed、RED 5 / GREEN 21、関連36 pass | 採用 |
| `repair_owner_vax_review` | TDD writer | owners/vaccination review HIGHのsource+tests | completed、4 focused RED、GREEN 58/58 | 採用 |
| `repair_med_check_reception` | TDD writer | medical/checkup/reception review HIGHのsource+tests | completed、medical 24/checkup 16/reception 9 pass | 採用 |
| `repair_examination_second_review` | TDD writer | examination post-await/delete/edit-pet identity source+test | completed、追加RED 5+2、GREEN 44/44 | 採用 |
| `repair_hosp_vax_second_review` | TDD writer | hospitalization submit/vaccination list+form latest death | completed、32/32、42/42、11/11 | 採用 |
| `repair_medical_second_review` | TDD writer | medical quick patch/owner change/form+list delete/edit source+tests | completed、RED 6+4、GREEN 68、wiring 58 | 採用 |
| `post_repair_route_review` | independent reviewer | final frozen ledgerのdeterministic 8/25 route sample + 全citation existence | completed、sample=`25,13,15,3,2,21,7,11`、43 citation in-bounds | CRITICAL/HIGH=0、全8 verdict再現 |

### Failure Signature / assumption deviations

- FS-01: first CarePlan GREENでmemoized childがtest mockだけの変更ではrerenderせず1件FAIL。新仮説として`hospitalizationId`変更を同commitへ含め、broader regression 24/24へ復旧。
- FS-02: hospitalization detail regressionで既存route testに`AuthProvider`が無くFAIL。標準`useAuth` mockを追加し3/3へ復旧。
- FS-03: VaccinationListの最初のREDで`fireEvent`がRadix menuを開かずdefect経路へ到達しなかった。product interactionへ合わせ`userEvent`へ変更し、有効なRED（delete 1回）を取得。
- FS-04: independent 8-route sampleで当初`ENFORCED`とした5 routeの一部mutationに死亡guard不足を再現。verdictを未確定へ戻し、追加repair laneを起動した。
- FS-05: vaccination full REDを最初に一括実行した際、test fixture object identityの再生成でcontainer JS heap limitへ到達。対象を最小specへ縮小しfixture identityを安定化してから、behavioral RED（form 2件、list 2件）とGREEN 42/42・11/11を取得した。
- MEDIUM follow-up（本unitのCRITICAL/HIGH修理後も製品設計を要するため未着手）: reception optimistic UI rollback、permission-group parent/rules partial success、owner parent/pet partial success、medical auto-create拒否後retry UX。backend原子化/UX契約は別unitで裁定する。
- Assumption deviation: 初期prompt contextはF9 backendが`lstep`下にもある可能性を示したが、実enumerationは`backend/internal/model`と`backend/internal/medicalrecord`の11 fileを返した。推測ではなくenumeration/full read結果を採用した。

### Type diagnostics closure（2026-07-26 follow-up）

- Unit status: INCOMPLETE。captured diagnostics 5件をexact pathで分類し、owned 5 / not-owned 0（`5 + 0 = 5`）を確認してsource修理・scoped gateまで完了したが、必須のpost-repair captureが未提供のためclosureは未完了。
- Input: `/tmp/fe12-typecheck.txt`、1425 bytes、SHA-256=`bfe5ebb6822c86b5e148e0cb3dd1e441e6367974e0a3da15537b0ecf39f7d1ef`。promptの禁止に従いproject-level type-checkは再実行していない。
- Saved prompt validation: `node /Users/minoru/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fast-fe12-02-typecheck-closure.md` → exit 0、`Prompt Craft Harness Validation: PASS`（execution前とfinal reconciliationの双方）。
- Harness: `tdd-workflow` + `react-testing` + sequential loop。captured TypeScript diagnosticsをRED、既存のpost-`await`死亡/権限testをbehavior pinとして用い、重複testは追加していない。
- Coverage: `frontend/vite.config.ts:163-169`はthresholdを持たず、`.github/workflows/ci.yml:284-299`のproject-wide coverage + ratchetだけがbinding gate。本promptはfull-project test/coverageを禁止しscoped coverage thresholdも定義しないためcoverage commandは実行せず、対象guardを直接pinする67 testをscoped実行した。
- Follow-up changed files: `frontend/src/features/checkups/hooks/use-checkup-form.ts`、`frontend/src/features/examinations/hooks/use-examination-form.ts`、`frontend/src/features/vaccinations/routes/VaccinationList.tsx`、本ledger。隣接test 3件はfollow-up baselineとbyte一致で未変更。

分類済みdiagnostic全件（container相対`src/...`をrepository相対`frontend/src/...`へ正規化）:

```text
OWNED | frontend/src/features/checkups/hooks/use-checkup-form.ts
src/features/checkups/hooks/use-checkup-form.ts(129,14): error TS2367: This comparison appears to be unintentional because the types '"生存" | undefined' and '"死亡"' have no overlap.

OWNED | frontend/src/features/checkups/hooks/use-checkup-form.ts
src/features/checkups/hooks/use-checkup-form.ts(147,16): error TS2367: This comparison appears to be unintentional because the types '"生存" | undefined' and '"死亡"' have no overlap.

OWNED | frontend/src/features/examinations/hooks/use-examination-form.ts
src/features/examinations/hooks/use-examination-form.ts(295,15): error TS2367: This comparison appears to be unintentional because the types 'false' and 'true' have no overlap.

OWNED | frontend/src/features/examinations/hooks/use-examination-form.ts
src/features/examinations/hooks/use-examination-form.ts(321,15): error TS2367: This comparison appears to be unintentional because the types 'false' and 'true' have no overlap.

OWNED | frontend/src/features/vaccinations/routes/VaccinationList.tsx
src/features/vaccinations/routes/VaccinationList.tsx(83,35): error TS2345: Argument of type 'string | undefined' is not assignable to parameter of type 'string'.
  Type 'undefined' is not assignable to type 'string'.
```

Not-owned diagnostic: 0件（観測・未変更対象なし）。

| Owned diagnostic | Repair walk | Guard preservation |
|---|---|---|
| checkups `TS2367` line 129 | `isMutationPetDeceased()`を追加し、`useLayoutEffect`同期refの厳密な`=== "死亡"`を関数境界で再読取 | カルテ作成await後の死亡/permission拒否順を維持 |
| checkups `TS2367` line 147 | 同readerを結果PUT直前にも再実行 | 健診作成await後の死亡/permission拒否順を維持 |
| examinations `TS2367` line 295 | `isPetExplicitlyDeceased()`を追加し、厳密な`=== true`を関数境界で再読取 | parent update await後のitems拒否を維持 |
| examinations `TS2367` line 321 | 同readerをparent create await後にも再実行 | items作成前の死亡/create permission拒否を維持 |
| vaccinations `TS2345` line 83 | `useGetPet(record.petId ?? "")`へ正規化 | ID欠落時はquery disabledかつ`pet?.status === "生存"`不成立でeditをfail-closed |

Scoped component tests（full command outputからDocker Compose環境warning以外をそのまま記録）:

```text
$ docker compose exec -e NO_COLOR=1 frontend npx vitest run src/features/checkups/hooks/use-checkup-form.test.ts src/features/checkups/routes/CheckupForm.test.tsx --silent
RUN  v4.1.8 /app
✓ src/features/checkups/hooks/use-checkup-form.test.ts (12 tests) 559ms
✓ src/features/checkups/routes/CheckupForm.test.tsx (4 tests) 299ms
Test Files  2 passed (2)
Tests       16 passed (16)
Duration    3.14s
exit 0

$ docker compose exec -e NO_COLOR=1 frontend npx vitest run src/features/examinations/hooks/use-examination-form.test.ts --silent
RUN  v4.1.8 /app
✓ src/features/examinations/hooks/use-examination-form.test.ts (44 tests) 76ms
Test Files  1 passed (1)
Tests       44 passed (44)
Duration    961ms
exit 0

$ docker compose exec -e NO_COLOR=1 frontend npx vitest run src/features/vaccinations/routes/VaccinationList.test.tsx --silent
RUN  v4.1.8 /app
✓ src/features/vaccinations/routes/VaccinationList.test.tsx (11 tests) 755ms
Test Files  1 passed (1)
Tests       11 passed (11)
Duration    3.68s
exit 0
```

Scoped ESLint:

```text
$ docker compose exec frontend npx eslint src/features/checkups/hooks/use-checkup-form.ts src/features/examinations/hooks/use-examination-form.ts src/features/vaccinations/routes/VaccinationList.tsx --max-warnings 0
time="2026-07-26T17:33:39+09:00" level=warning msg="The \"DB_USER\" variable is not set. Defaulting to a blank string."
time="2026-07-26T17:33:39+09:00" level=warning msg="The \"DB_PASSWORD\" variable is not set. Defaulting to a blank string."
time="2026-07-26T17:33:39+09:00" level=warning msg="The \"DB_NAME\" variable is not set. Defaulting to a blank string."
ESLINT_EXIT=0
```

Follow-up baseline相対のadded-line inspection:

```text
ADDED_ANY=0
ADDED_TS_SUPPRESSION=0
ADDED_NON_NULL_ASSERTION_CANDIDATES=0
DIFF_CHECK_EXIT=0
UNCHANGED frontend/src/features/checkups/hooks/use-checkup-form.test.ts
UNCHANGED frontend/src/features/examinations/hooks/use-examination-form.test.ts
UNCHANGED frontend/src/features/vaccinations/routes/VaccinationList.test.tsx
```

Containment / tracked probe:

```text
SESSION_START_FULL_SCOPED_STATUS=79 paths
SESSION_START_OWNED=FE-refactor.md + frontend/src 75 paths
SESSION_START_OBSERVED_NOT_OWNED:
 M frontend/src/constants/item-category.ts
 M frontend/src/features/accounting/components/ItemListCard.test.tsx
 M frontend/src/features/accounting/components/ItemListCard.tsx

SESSION_END_FULL_SCOPED_STATUS=76 paths
SESSION_END_OWNED=FE-refactor.md + frontend/src 75 paths
SESSION_END_OBSERVED_NOT_OWNED=0

FOLLOWUP_TARGET_STATUS:
 M FE-refactor.md
 M frontend/src/features/checkups/hooks/use-checkup-form.ts
 M frontend/src/features/examinations/hooks/use-examination-form.ts
 M frontend/src/features/vaccinations/routes/VaccinationList.tsx

TRACKED:
FE-refactor.md
frontend/src/features/checkups/hooks/use-checkup-form.ts
frontend/src/features/examinations/hooks/use-examination-form.ts
frontend/src/features/vaccinations/routes/VaccinationList.tsx
TRACKED_PROBE_EXIT=0
```

session-startに観測したnot-owned 3 pathは他作業のquiescenceでsession-endにはstatusから消えた。本follow-upのagent ownership・baseline相対content deltaはいずれも上記4 pathだけで、not-owned pathを編集・revert・stageしていない。session-endのfull scoped status 76 pathは、直後の`### Final verification / containment`に貼付済みのledger + 75-path manifestと一致する。

Orchestration（native Workflow toolは未提供のためmulti-agent fan-out）:

| ID / label | Role / responsibility | Writer-owned paths | Status / evidence | Integration |
|---|---|---|---|---|
| `typecheck_closure_plan` | planner、phase/checklist/ownership設計 | read-only | completed、5/5/0分類と3 writer partition | plan採用、prompt矛盾はHarness feedbackへ |
| `diagnostic_classifier` | captured outputと75-path manifestのexact classification | read-only | completed、total 5 / owned 5 / not-owned 0 | 全件採用 |
| `probe_checkups_types` | checkups TS2367/guard test probe | read-only | completed、await跨ぎnarrowingを特定 | repair案採用 |
| `probe_examination_types` | examinations TS2367/guard test probe | read-only | completed、await跨ぎnarrowingを特定 | repair案採用 |
| `probe_vaccination_types` | vaccinations TS2345/fail-closed probe | read-only | completed、optional pet ID contractを特定 | repair案採用 |
| `repair_checkups_type` | TDD writer | `use-checkup-form.ts` | completed、16/16 + ESLint 0 | 採用 |
| `repair_examination_type` | TDD writer | `use-examination-form.ts` | completed、44/44 + ESLint 0 | 採用 |
| `repair_vaccination_type` | TDD writer | `VaccinationList.tsx` | completed、11/11 + ESLint 0 | 採用 |
| `repair_vaccination_type/review_vaccination_type_fix` | TypeScript/React reviewer | read-only | completed、CRITICAL/HIGH/MEDIUM=0 | Approve |
| `independent_type_guard_review` | 全repair hunkのclinical guard reviewer | read-only | completed、CRITICAL/HIGH=0、全hunkで拒否状態不変 | Approve |

- Failure Signature: none。writer/rootのscoped gateは初回で全てexit 0。
- De-Sloppify: test追加なし、console/commented code/broad catch/drive-by refactorなし。source差分はstable reader 2個、呼出置換、optional ID正規化1行だけ。
- Harness Improvement Feedback P1: fresh project-level type-check再実行を禁止したまま「old diagnostic no longer occurs」の実行証跡を要求するため、同一compiler gateの再現証明は本prompt内では生成できない。次版は修理後のuser-captured再実行を必須handoffにするか、許可するscoped type gateを明記する。
- Harness Improvement Feedback P1: containment例の`git status -- FE-refactor.md frontend/src`は既知のunrelated 3 pathも含む。session-start/end deltaまたはexact owned manifest statusをPASS gateとして明記すべき。

#### Post-repair capture re-entry（2026-07-26）

- Run status: `INCOMPLETE`。Acceptance Checklistは8 PASS / 1 BLOCKED / 0 FAIL。
- Post-repair capture gate: `BLOCKED`。`test -e /tmp/fe12-typecheck-after.txt`および`test -s /tmp/fe12-typecheck-after.txt`はいずれもexit 1（file absent）。修理済み5件の消失と新規owned diagnostic 0件は未検証であり、source inspection・Vitest・ESLintを本gateの代替証拠にしない。
- Required input: 非空のuser captureをexact path `/tmp/fe12-typecheck-after.txt`へ配置し、本promptを再実行する。
- Pre-repair capture再確認: `/tmp/fe12-typecheck.txt`は1425 bytes / 14 lines、SHA-256=`bfe5ebb6822c86b5e148e0cb3dd1e441e6367974e0a3da15537b0ecf39f7d1ef`。分類はtotal 5 / owned 5 / not-owned 0のまま。
- Current-source walk: checkupsの`isMutationPetDeceased()`、examinationsの`isPetExplicitlyDeceased()`、vaccinationsの`useGetPet(record.petId ?? "")`は現行sourceに存在する。re-entryではfrontend source/testを変更せず、本ledgerだけを更新した。
- Re-entry containment: session-start/currentの`git status --porcelain -- FE-refactor.md frontend/src`は各76行、SHA-256=`bec1e003c1d1308ee3c1c7f7c306a554871f33db0da639814bc28b03066a4162`でbyte一致（`diff` exit 0）。ledgerのsession-start scratch相対差分は56 additions / 1 deletionで非zero、`git ls-files --error-unmatch FE-refactor.md`はexit 0。
- Protected-section check: type diagnostics節より前のprefixと`### Final verification / containment`以降のsuffixはsession-start scratchとbyte一致（両`diff` exit 0）。route marker/table、F16/F9、6 runbook、残余riskを変更していない。
- Saved prompt validation: `node /Users/minoru/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fast-fe12-02-typecheck-closure.md` → exit 0、`Prompt Craft Harness Validation: PASS`。

Fresh scoped component tests:

```text
$ docker compose exec -e NO_COLOR=1 frontend npx vitest run src/features/checkups/hooks/use-checkup-form.test.ts src/features/checkups/routes/CheckupForm.test.tsx --silent
Test Files  2 passed (2)
Tests       16 passed (16)
Duration    2.57s
exit 0

$ docker compose exec -e NO_COLOR=1 frontend npx vitest run src/features/examinations/hooks/use-examination-form.test.ts --silent
Test Files  1 passed (1)
Tests       44 passed (44)
Duration    1.03s
exit 0

$ docker compose exec -e NO_COLOR=1 frontend npx vitest run src/features/vaccinations/routes/VaccinationList.test.tsx --silent
Test Files  1 passed (1)
Tests       11 passed (11)
Duration    3.18s
exit 0
```

Fresh scoped ESLint:

```text
$ docker compose exec frontend npx eslint src/features/checkups/hooks/use-checkup-form.ts src/features/examinations/hooks/use-examination-form.ts src/features/vaccinations/routes/VaccinationList.tsx --max-warnings 0
ESLINT_EXIT=0
```

Re-entry orchestration（native Workflow toolは未提供のためparallel multi-agent fan-out）:

| ID / label | Role / responsibility | Writer-owned paths | Status / evidence | Integration |
|---|---|---|---|---|
| `reentry_closure_plan` | planner、current contractと9 checklistのreconciliation | read-only | completed、8 PASS / 1 BLOCKEDを導出 | 採用 |
| `reentry_pre_classifier` | pre-captureと75-path manifestのexact classification、current repair walk | read-only | completed、total 5 / owned 5 / not-owned 0 | 採用 |
| `reentry_after_capture_gate` | after-capture existence/size/hash probe | read-only | completed、file absent、gate BLOCKED | 採用 |
| `reentry_guard_review` | TypeScript/React clinical guard review | read-only | completed、CRITICAL/HIGH=0、全hunkで拒否範囲不変 | Approve |
| `reentry_ledger_probe` | ledger contradiction・protected section・containment probe + post-edit review | read-only | completed、旧COMPLETE表記との矛盾を検出し、post-edit protected hash一致 | status訂正へ採用、final Approve |

- Independent Review Gate: checkups TS2367 ×2、examinations TS2367 ×2、vaccinations TS2345の全repairを再読し、death/permission/JST/fail-closed guardの拒否範囲は変わらないと判定。CRITICAL/HIGH=0、Approve。
- De-Sloppify: re-entryでcode/test変更なしのため該当なし。ledgerにはblocking evidenceだけを追記し、既存の分類・repair walk・command evidenceを保持した。
- Failure Signature: none。実行失敗のretryではなく、必須external input欠落を1 checklist itemのBLOCKEDとして記録した。
- Harness Improvement Feedback: none needed。更新済みpromptのpost-capture gateが、green test/lintをtype closureへ誤代用せず正しく`INCOMPLETE`へ止めた。

### Final verification / containment

- Final scoped Vitest（feature別path指定）: 37 files / 399 tests PASS。内訳: medical-records 7/92、examinations 2/47、hospitalization detail/list 10/35、hospitalization form+vaccinations 5/90、checkups 2/16、reception 4/51、owners 4/58、master 3/10。
- Final scoped ESLint: session-owned `frontend/src` 75 filesをcommand lineへ明示し`--max-warnings 0`、exit 0。path内訳はshared deceased=5、checkups=4、examinations=2、hospitalization=22、master=4、medical-records=15、owners=12、reception=7、vaccinations=4。
- Session-start WIP preservation: `frontend/src/features/medical-records/components/MedicalRecordVaccination.tsx`と同testはscratch copyとSHA-1一致。session-owned manifestから除外した。
- Concurrent unrelated delta: `frontend/src/features/accounting/components/ItemListCard.test.tsx`およびbackend/docsの別作業差分はsession中にも変化したが、本unitのwriter ownership外であり編集・revert・stageしていない。unscoped cleanlinessはgateにしていない。
- Owned status: `FE-refactor.md` + session-owned `frontend/src` 75 filesのみ。内訳=`shared 5 / checkups 4 / examinations 2 / hospitalization 22 / master 4 / medical-records 15 / owners 12 / reception 7 / vaccinations 4`。

```text
 M FE-refactor.md
 M frontend/src/components/shared/PetDeceasedRecordButton/PetDeceasedBanner.test.tsx
 M frontend/src/components/shared/PetDeceasedRecordButton/PetDeceasedBanner.tsx
 M frontend/src/components/shared/PetDeceasedRecordButton/PetDeceasedDialog.test.tsx
 M frontend/src/components/shared/PetDeceasedRecordButton/PetDeceasedDialog.tsx
 M frontend/src/components/shared/PetDeceasedRecordButton/PetDeceasedRecordButton.tsx
 M frontend/src/features/checkups/hooks/use-checkup-form.test.ts
 M frontend/src/features/checkups/hooks/use-checkup-form.ts
 M frontend/src/features/checkups/routes/CheckupForm.test.tsx
 M frontend/src/features/checkups/routes/CheckupForm.tsx
 M frontend/src/features/examinations/hooks/use-examination-form.test.ts
 M frontend/src/features/examinations/hooks/use-examination-form.ts
 M frontend/src/features/hospitalization/components/CarePlanTab/CarePlanTab.tsx
 M frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.permissions.test.tsx
 M frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.test.tsx
 M frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx
 M frontend/src/features/hospitalization/components/HospitalizationDetailActions.checkin.test.tsx
 M frontend/src/features/hospitalization/components/HospitalizationDetailActions.tsx
 M frontend/src/features/hospitalization/components/HospitalizationExpandedView.tsx
 M frontend/src/features/hospitalization/components/HospitalizationPatientHeader.tsx
 M frontend/src/features/hospitalization/components/HospitalizationTabbedView.tsx
 M frontend/src/features/hospitalization/hooks/use-hospitalization-detail.ts
 M frontend/src/features/hospitalization/hooks/use-hospitalization-form-model.test.ts
 M frontend/src/features/hospitalization/hooks/use-hospitalization-form-model.ts
 M frontend/src/features/hospitalization/hooks/use-hospitalization-form.test.ts
 M frontend/src/features/hospitalization/hooks/use-hospitalization-form.ts
 M frontend/src/features/hospitalization/hooks/use-hospitalization-list.ts
 M frontend/src/features/hospitalization/routes/HospitalizationDetail.test.tsx
 M frontend/src/features/hospitalization/routes/HospitalizationForm.permissions.test.tsx
 M frontend/src/features/hospitalization/routes/HospitalizationForm.tsx
 M frontend/src/features/hospitalization/routes/HospitalizationList.tsx
 M frontend/src/features/master/components/MasterCRUDPage.tsx
 M frontend/src/features/master/routes/PermissionGroupSettings.tsx
 M frontend/src/features/medical-records/api/transforms.test.ts
 M frontend/src/features/medical-records/api/transforms.ts
 M frontend/src/features/medical-records/hooks/use-medical-record-auto-create.ts
 M frontend/src/features/medical-records/hooks/use-medical-record-form.auto-create.test.ts
 M frontend/src/features/medical-records/hooks/use-medical-record-form.ts
 M frontend/src/features/medical-records/hooks/use-medical-record-owner-change.test.ts
 M frontend/src/features/medical-records/hooks/use-medical-record-owner-change.ts
 M frontend/src/features/medical-records/hooks/use-medical-record-quick-patch-actions.test.ts
 M frontend/src/features/medical-records/hooks/use-medical-record-quick-patch-actions.ts
 M frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts
 M frontend/src/features/medical-records/routes/MedicalRecordForm.permissions.test.tsx
 M frontend/src/features/medical-records/routes/MedicalRecordForm.tsx
 M frontend/src/features/medical-records/routes/MedicalRecords.test.tsx
 M frontend/src/features/medical-records/routes/MedicalRecords.tsx
 M frontend/src/features/owners/components/OwnerPetsSection.test.tsx
 M frontend/src/features/owners/components/OwnerPetsSection.tsx
 M frontend/src/features/owners/components/OwnersListTable.report.test.tsx
 M frontend/src/features/owners/components/OwnersListTable.tsx
 M frontend/src/features/owners/hooks/use-owner-form.test.ts
 M frontend/src/features/owners/hooks/use-owner-form.ts
 M frontend/src/features/owners/hooks/use-pet-form-list-state.test.ts
 M frontend/src/features/owners/hooks/use-pet-form-list-state.ts
 M frontend/src/features/owners/routes/OwnerForm.bug373.test.tsx
 M frontend/src/features/owners/routes/OwnerForm.tsx
 M frontend/src/features/owners/routes/OwnersList.tsx
 M frontend/src/features/reception/components/AppointmentCard.test.tsx
 M frontend/src/features/reception/components/AppointmentCard.tsx
 M frontend/src/features/reception/hooks/use-reception-kanban.test.ts
 M frontend/src/features/reception/hooks/use-reception-kanban.ts
 M frontend/src/features/reception/hooks/use-reception-modal-handlers.test.ts
 M frontend/src/features/reception/hooks/use-reception-modal-handlers.ts
 M frontend/src/features/reception/routes/Reception.tsx
 M frontend/src/features/vaccinations/hooks/use-vaccination-form.test.ts
 M frontend/src/features/vaccinations/hooks/use-vaccination-form.ts
 M frontend/src/features/vaccinations/routes/VaccinationList.test.tsx
 M frontend/src/features/vaccinations/routes/VaccinationList.tsx
?? frontend/src/features/hospitalization/components/HospitalizationPatientHeader.test.tsx
?? frontend/src/features/hospitalization/hooks/use-hospitalization-detail.test.tsx
?? frontend/src/features/hospitalization/hooks/use-hospitalization-list.test.ts
?? frontend/src/features/master/components/MasterCRUDPage.test.tsx
?? frontend/src/features/master/routes/PermissionGroupSettings.test.tsx
?? frontend/src/features/medical-records/hooks/use-medical-record-save-action.test.ts
?? frontend/src/features/owners/routes/OwnersList.permissions.test.tsx
```

- New artifact probe: 新規test 7 pathはすべて`NOT_IGNORED`（HospitalizationPatientHeader、hospitalization detail/list hook、MasterCRUDPage、PermissionGroupSettings、medical save action、OwnersList permission）。
- Ledger baseline diff: session-start scratchとcurrentは`Files ... differ`、`git diff --no-index --numstat`でnonzero。
- Marker/row/runbook gate: marker 4種各1、active route=25、verification row=25、invalid citation/verdict row=0、runbook explicit field=42（6×7）。
- `### 未裁定の残余risk`はsession-start scratchとsection SHA-1一致。
- Full-project TypeScript type-checkはprohibited commandのため未実行。USER manual command=`docker compose exec frontend pnpm type-check`。未実行をPASS扱いしない。

### 未裁定の残余risk

下記はcurrent sourceで未解消を確認したfollow-up候補であり、本整理unitは着手・起票・優先度裁定をしない。

- alive petへ予約編集した直後、Receptionのdanger sentinelがquery invalidation/refetchまで一時的に旧値を保つ可能性がある。
- backend死亡登録APIは既死亡recordの再登録をConflict拒否しない。
- medical-record auto-createの当日予約lookupは実時刻に依存し、対象testに時刻固定がないためJST日付境界でflakyになる可能性がある。
- examination parentとitemsは2 HTTP operationのため、parent成功・items失敗の部分成功riskが残る。
- PermissionGroup以外のmasterでpermission wiringがないcall siteはCRUD 19、save 24。
- tygoに出力へ効かない既存pointer mapping 15行が残る。
- line-reserve `CompletePage`は4文字超のmalformed時刻で旧表示との差があり、専用regression testがない。
- `ConfirmPage`の既存testはpending中の二重送信、LIFF送信順、409時のalert非表示を直接assertしない。
- `manual` chunkの500 kB警告該当は現run未再計測。`manual-index`のMarkdown eager raw bundle構造は継続しているため、次回許可buildで再計測する。
- `AuthProvider`のfeature barrel経由eager importによりauth routeのlazy splitは実効化されていない。feature deep importで迂回せず、公開境界のarchitecture decisionを先に行う。

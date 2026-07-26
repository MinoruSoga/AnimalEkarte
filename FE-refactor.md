# FE-refactor — FE12-02 active-only ledger

> 更新: 2026-07-26（要件責任者: 曽我）
> 業務目的: 未完了の臨床sentinel・RBAC安全境界を、決裁と実測の証跡が揃うまで追跡する。
> 本ファイルは使い捨てのactive ledger。恒久規約は `DESIGN.md`、`docs/spec/design-system.md`、`docs/spec/ui-design-compliance.md`、`frontend/CLAUDE.md` を正本とする。
> 削除した `## FE12-02 unit execution record` と過去のFE/R-F履歴はCommit `657c1a49cd2c37dc63f5af8e530258a36a12d81e` に保存する。

## Active scope and authority

- 追跡対象は `FE12-02` の未完了route、`F16`・`F9`、`U10`、`M-01`〜`M-05`、line-reserve font実機確認、4件のMEDIUM follow-up、未裁定riskだけとする。
- 色と臨床semanticは `docs/spec/design-system.md`、恒久route適合は `docs/spec/ui-design-compliance.md`、明示的なPO/USER裁定は `q&a.html` を正本とする。
- authorityから項目が消えたことや判断待ち件数が0であることだけでは完了とみなさない。明示的な決裁または実測証跡が無い項目は保持する。
- 本ledgerの更新は実装・runtime検証・製品決裁を代替しない。

## Active routes

<!-- FE12-ROUTE-TABLE-START -->
| エリア | ページ | パス | コンポーネント | 未完了事項 |
|---|---|---|---|---|
| 検査 | 検査一覧 | /examinations | ExaminationsList | M-02の一覧HIGH/LOW cue実測・製品確認 |
| 定期健診 | 定期健診一覧 | /checkups | CheckupsList | F9の要フォローsentinel裁定 |
| 受付/飼主/予約 | 飼主一覧 | /owners | OwnersList | F16死亡token裁定・M-01操作範囲実測 |
<!-- FE12-ROUTE-TABLE-END -->

## Active task

<!-- FE12-TASK-TABLE-START -->
| ID | Priority | Active frontier | Dependency | Completion evidence |
|---|---|---|---|---|
| FE12-02 | P0 | F16・F9の決裁、M-01〜M-05、line-reserve font実機確認、4件のMEDIUM follow-up | U10はF16の決裁待ち | 明示的な決裁、対象実測、別unitのatomicity/UX契約が揃うこと |
<!-- FE12-TASK-TABLE-END -->

## Authority drift

- `q&a.html` はPO判断待ち0件と記載する一方、F16/F9の明示裁定を含まない。件数要約から完了を推測せず、F16・F9・U10を保持する。
- `docs/spec/ui-design-compliance.md` にはC18件数ratchetの旧記述が残るが、現行auditはratchetを持たない。本runでは正本側を編集せず、別scopeの文書driftとして扱う。

## C6a 臨床安全レビュー

- `/owners`: F16の死亡token裁定とM-01の操作範囲実測を完了する。
- `/examinations`: M-02で一覧のHIGH/LOW非色cue要否と4 viewport layoutを実測する。
- `/checkups`: F9の要フォローsentinel裁定とM-05のresponsive実測を完了する。
- M-03〜M-05では、静的に閉じたrouteも含めてRBAC・child control・clinical sentinelのruntime証跡を取得する。
- 静的レビューで閉じず、残件が実装認可された場合は既存sentinel fixtureのscoped component testを先にRED化してから修正する。

## Active execution rules

決裁・実測後に対応する場合も、次の安全境界を維持する。

1. **臨床sentinelは生成型から表示・操作境界まで欠落させない。** 死亡は明示的なpositive matchで遷移・mutation callbackを拒否し、危険「高」は非色cueを伴う警告として扱う。死亡statusと死亡日時が不整合なら再登録導線を出さない。
2. **権限はaction別の最新値をmutation直前に再検査する。** UIの非表示・disabled・route guardだけを最終防壁にしない。view/edit共用の唯一のdetail routeはread accessを維持し、mutation境界をfail-closedにする。commit直後にも発火し得るcallbackのpermission refは`useLayoutEffect`で同期する。
3. **臨床date-onlyはJSTの厳密過去で判定する。** `YYYY-MM-DD`契約をguardし、`todayJSTISO()`との文字列比較`<`を使う。現在時刻との`Date`比較で当日を期限超過にしない。

## 3 routeの静的検証結果

判定語は次の3値に限定する。`ENFORCED`はcurrent sourceでbinding ruleを再現できる、`GAP`はcurrent sourceがbinding ruleへ明確に違反する、`RUNTIME-ONLY`はseed/persona/browser実測または製品裁定なしにsourceだけでは閉じられない、を表す。

| # | パス | Verdict | current source evidence |
|---:|---|---|---|
| 1 | `/examinations` | RUNTIME-ONLY | `frontend/src/features/examinations/routes/ExaminationsList.tsx:225-234` — 一覧は`resultSummary`と検査statusを表示するが、HIGH/LOW cueを一覧にも要するかはM-02の4 viewport実測・製品確認が必要。 |
| 2 | `/checkups` | RUNTIME-ONLY | `frontend/src/features/checkups/routes/CheckupsList.tsx:297-305` — 次回期限とraw resultのみで、要フォローfieldはF9 full-readで不在。表示要否は曽我の裁定が必要。 |
| 3 | `/owners` | RUNTIME-ONLY | `frontend/src/features/owners/components/OwnersListTable.tsx:128-128,191-198`、`frontend/src/lib/status-helpers.ts:176-178` — view-only detail linkとRBACはsourceで再現できるが、死亡のgeneric grayと許可操作範囲はF16/M-01裁定・実測待ち。 |

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
| D-01 | Reception | reception optimistic UI rollback | 別unitでatomicity/rollback契約を裁定 |
| D-02 | PermissionGroup | permission-group parent/rules partial success | 別unitで補償・retry契約を裁定 |
| D-03 | Owner/Pet | owner parent/pet partial success | 別unitでatomicity/補償契約を裁定 |
| D-04 | MedicalRecords auto-create | medical auto-create拒否後retry UX | 別unitでretry UX契約を裁定 |

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

### 別unit裁定待ちのMEDIUM follow-up

- `reception optimistic UI rollback`: optimistic更新後の失敗時rollbackとquery再同期の契約を別unitで裁定する。
- `permission-group parent/rules partial success`: parent保存成功・rules保存失敗時の補償とretry契約を別unitで裁定する。
- `owner parent/pet partial success`: owner parent成功・pet処理失敗時のatomicityまたは補償契約を別unitで裁定する。
- `medical auto-create拒否後retry UX`: 自動作成拒否後の再試行導線と重複防止契約を別unitで裁定する。

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

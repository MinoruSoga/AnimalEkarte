# FE-refactor — FE12-02 active-only ledger

> 更新: 2026-07-27（要件責任者: 曽我）
> 業務目的: 未完了の臨床sentinel・RBAC安全境界を、決裁と実測の証跡が揃うまで追跡する。
> 本ファイルは使い捨てのactive ledger。恒久規約は `DESIGN.md`、`docs/spec/design-system.md`、`docs/spec/ui-design-compliance.md`、`frontend/CLAUDE.md` を正本とする。
> 解消済みの証跡: F16/U10/O-A実装=`3b7524748`、R-A/M-A=`e7c978ec9`、P-A=`99bac632e`、exam parent+items原子化=`24929e83d`、死亡API 409 fail-closed=`45b681866`、test-hardening 3件=`3c993420a`。代理決裁6件の記録と調査パックは`a500d424c`および本更新直前までの履歴に保存する。
> 削除した `## FE12-02 unit execution record` と過去のFE/R-F履歴はCommit `657c1a49cd2c37dc63f5af8e530258a36a12d81e` に保存する。

## Active scope and authority

- 追跡対象は `M-01`〜`M-05`、line-reserve font実機確認、未裁定riskだけとする。F16・F9・U10・MEDIUM 4件は代理決裁・実装済みで追跡を終了した。
- 色と臨床semanticは `docs/spec/design-system.md`、恒久route適合は `docs/spec/ui-design-compliance.md`、明示的なPO/USER裁定は `q&a.html` を正本とする。
- authorityから項目が消えたことや判断待ち件数が0であることだけでは完了とみなさない。明示的な決裁または実測証跡が無い項目は保持する。
- 本ledgerの更新は実装・runtime検証・製品決裁を代替しない。

## Active routes

<!-- FE12-ROUTE-TABLE-START -->
| エリア | ページ | パス | コンポーネント | 未完了事項 |
|---|---|---|---|---|
| 検査 | 検査一覧 | /examinations | ExaminationsList | M-02の一覧HIGH/LOW cue実測・製品確認 |
| 受付/飼主/予約 | 飼主一覧 | /owners | OwnersList | M-01操作範囲実測 |
<!-- FE12-ROUTE-TABLE-END -->

M-03〜M-05はroute横断の実測であり、対象routeは各runbookに列挙する。

## Active task

<!-- FE12-TASK-TABLE-START -->
| ID | Priority | Active frontier | Dependency | Completion evidence |
|---|---|---|---|---|
| FE12-02 | P0 | M-01〜M-05、line-reserve font実機確認 | 各runbookの着手ブロッカー（fixture準備）が先行 | 対象実測の証跡が揃うこと |
<!-- FE12-TASK-TABLE-END -->

## Authority drift

- `docs/spec/ui-design-compliance.md` にはC18件数ratchetの旧記述が残るが、現行auditはratchetを持たない。本ledgerは正本側を編集せず、別scopeの文書driftとして扱う。是正は`docs/spec/ui-design-compliance.md`を所有する別unitが行う。

## C6a 臨床安全レビュー

- `/owners`: M-01の操作範囲実測を完了する。
- `/examinations`: M-02で一覧のHIGH/LOW非色cue要否と4 viewport layoutを実測する。
- M-03〜M-05では、静的に閉じたrouteも含めてRBAC・child control・clinical sentinelのruntime証跡を取得する。
- 静的レビューで閉じず、残件が実装認可された場合は既存sentinel fixtureのscoped component testを先にRED化してから修正する。

## Active execution rules

決裁・実測後に対応する場合も、次の安全境界を維持する。

1. **臨床sentinelは生成型から表示・操作境界まで欠落させない。** 死亡は明示的なpositive matchで遷移・mutation callbackを拒否し、危険「高」は非色cueを伴う警告として扱う。死亡statusと死亡日時が不整合なら再登録導線を出さない。
2. **権限はaction別の最新値をmutation直前に再検査する。** UIの非表示・disabled・route guardだけを最終防壁にしない。view/edit共用の唯一のdetail routeはread accessを維持し、mutation境界をfail-closedにする。commit直後にも発火し得るcallbackのpermission refは`useLayoutEffect`で同期する。
3. **臨床date-onlyはJSTの厳密過去で判定する。** `YYYY-MM-DD`契約をguardし、`todayJSTISO()`との文字列比較`<`を使う。現在時刻との`Date`比較で当日を期限超過にしない。

## 要実測項目

### M-01 OwnersList操作範囲

- Route: `/owners?include_deceased=true`。
- Fixture seed source: disposable local `003_demo`へ`backend/migrations/seeds/003_demo/owners.csv`と`pets.csv`を適用し、`docs/ops/testing/scenarios/S01-deceased-pet-guard.md`/`V03-owner-pet-staff-forms.md`の手順で同一ownerにalive、danger=`高`、deceasedの3頭を準備する。実測後はS01どおり生存へ戻す。
- Fixture実査: 全named fileは実在。owner `300588`（`owners.csv:593`）のpet `1001002`（`pets.csv:1005`）、`1001004`（`:1007`）、`1001005`（`:1008`）を同一ownerの3頭候補として固定できる。ただし`pets.csv`全15,654行で`danger_level=high`は0件、`status=deceased + deceased_at + deceased_reason`完備は0件（`deceased_reason`非空自体が0件）。
- 追加すべき具体値: disposable local上で`1001002={status:alive,danger_level:low,death fields:null}`を対照、`1001004={status:alive,danger_level:high,death fields:null}`、`1001005={status:deceased,danger_level:low,deceased_at:D 00:00:00+09,deceased_reason:"M01/M05 fixture"}`とする。CSVは手編集せずS01/V03のUI/API経由で作る。
- 着手ブロッカー: `1001004`のhigh化と`1001005`の死亡日時＋理由登録の2 mutationが未実施。
- 次の一手: fixture担当が2 mutationを準備し、実測担当と曽我または同席する仕様責任者がrunbookを実行し、完了後fixture担当が`1001005`をalive、`1001004`をlowへ復旧する。
- Persona: owners `view=true, edit=true, delete=true`の通常担当者。曽我または同席する仕様責任者が許可操作を記録する。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: include-deceasedを有効にし、3 rowそれぞれでowner名link、report、編集、削除、pet死亡登録/解除をpointer、Tab/Enter/Space、直接URLで試す。各操作前後のrequestをnetworkで記録する。
- Expected result: aliveは通常操作可、danger=`高`は「⚠ 危険」等の非色cueとaccessible nameを失わない。deceasedはF16裁定（一覧はグレーアウト維持・`3b7524748`）どおりの表示を記録し、曽我が許可/禁止する操作を操作単位で明示する。禁止とされた操作はmutation 0件。
- Required evidence artifacts: 4 viewport screenshot、accessibility tree、各操作のaccessible name、GET以外のnetwork HAR、操作可否の曽我裁定メモ。

### M-02 Examinations一覧意味とlayout

- Route: `/examinations`と対象petの`/medical-records/:id`検査履歴。
- Fixture seed source: `backend/migrations/seeds/003_demo/exams.csv`、`exam_results.csv`、`exam_types.csv`、`exam_type_fields.csv`を基に、`docs/ops/testing/scenarios/S02-exam-abnormal-highlight-lock.md`のhigh/low/normal同居fixtureを作る。
- Fixture実査: 全named fileとS02は実在。`exam_types.csv:4`のtype `3`は血液検査、`exam_type_fields.csv:2-4`はWBC `6.0-17.0`、RBC `5.5-8.5`、HCT `37-55`。一方、`exam_results.csv`全1,322,503行は`status=normal,is_abnormal=false`でHIGH/LOWは0件。既存medical recordは全件finalizedで、editable fixtureには使えない。
- 追加すべき具体値: alive dog `pet_id=1000018`（`pets.csv:19`、owner `300003`）へdraft medical record `M`を作り、`M`へ`exam_type_id=3,doctor_id=1,status=result_entered,result_summary="FE12-M02 HIGH/LOW/normal"`のexam `E`を作成する。`E`へWBC=`10.0`（normal）、RBC=`9.0`（high）、HCT=`30.0`（low）を同時登録し、backend導出後のstatus/is_abnormalを確認する。
- 着手ブロッカー: editable medical record `M`とnormal/HIGH/LOW同居exam `E`が未作成。
- 次の一手: fixture担当がM→E→3 resultの順でdisposable localへ作成し、実測担当がview-only personaで一覧とカルテ履歴を比較し、曽我が一覧cue要否を裁定する。
- Persona: examinations `view=true, create=false, edit=false, delete=false`のview-only担当者。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: 同一fixtureを一覧とカルテ履歴で開き、値・summary・statusを比較する。zoom 100%、pointer hover、keyboard focus、横スクロール有無を確認し、一覧でHIGH/LOW非色cueが必要か曽我へ提示する。
- Expected result: normalを異常表示しない。曽我が一覧cueを必要と裁定する場合はHIGH/LOWが色なしでも識別できること、不要と裁定する場合は詳細へのaccessible導線があること。全viewportでwrap/clip/overlapなし。
- Required evidence artifacts: 両surface×4 viewport screenshot、accessible text dump、computed color/token、overflow計測、曽我の一覧cue要否裁定。

### M-03 RBAC非活性の理由/name

- Route: `/settings/permission-groups`、`/medical-records/:id`、`/hospitalization/:id/edit`、`/vaccinations/:id`、`/examinations/:id`。
- Fixture seed source: `backend/migrations/seeds/003_demo/permission_groups.csv`、`permission_group_rules.csv`、`staffs.csv`、`staff_permission_groups.csv`と各featureの既存record CSV。`docs/ops/testing/scenarios/V03-owner-pet-staff-forms.md`で試験用group/staffを作り、`V01-clinical-forms.md`の既存recordを使う。
- Fixture実査: 全named CSV/V01/V03は実在。group `9`（`permission_groups.csv:10`）とstaff `37`（`staff_permission_groups.csv:38`）はclinical 4 resourceでview-onlyだが、`master-permission`は`false/false/false/false`（`permission_group_rules.csv:177`）のため全対象routeを覆わない。create-onlyとedit-without-delete完全一致groupは0件。hospitalization/vaccinationは0行、medical recordは全件finalizedでmutable recordも不足。
- 追加すべき具体値: `master-permission`、`medical-records`、`hospitalization`、`vaccinations`、`examinations`へ共通で、`FE12-M03-VIEW={true,false,false,false}`、`FE12-M03-CREATE={true,true,false,false}`、`FE12-M03-EDIT={true,false,true,false}`の3 groupを作り、各専用staffへ割り当てる。操作対象は別group `FE12-M03-TARGET`とする。secretはseed/ledgerへ記録せず実行時に供給する。
- 着手ブロッカー: 3 persona、`FE12-M03-TARGET`、M-02/M-04/M-05由来のmutable recordが未作成。
- 次の一手: fixture担当がfeature fixture→target group→3 group→3 staff作成・再編集割当を行い、実測担当がpersonaごとに再ログイン/再読込してrunbookを実行する。
- Persona: (1) view-only=`view:true/create:false/edit:false/delete:false`、(2) create-only=`view:true/create:true/edit:false/delete:false`、(3) edit-without-delete=`view:true/create:false/edit:true/delete:false`。
- 注記: 検査登録（POST）は`24929e83d`以降create+editの合成認可であり、create-only personaは拒否されるのが正しい期待値である。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: 各personaでrouteを再ログイン/再読込して開き、pointer、Tab/Enter/Space、formのprogrammatic submit、保存中の権限剥奪後callbackを試す。permission-groupは新規panel、既存panel、reorder、保存後rulesも個別確認する。
- Expected result: view accessは維持する。許可されたactionだけ実行でき、禁止controlはaccessible nameと理由を保持する。禁止personaからのPOST/PATCH/PUT/DELETEは0件で、same-commit剥奪後もmutationしない。
- Required evidence artifacts: persona×routeのaccessibility tree、4 viewport screenshot、network HAR、console log、action別permission matrixと0 mutation集計。

### M-04 Hospitalization child control実効性

- Route: `/hospitalization`のboard/listと`/hospitalization/:id`。
- Fixture seed source: `backend/migrations/seeds/003_demo/hospitalizations.csv`、`daily_records.csv`、`vital_records.csv`、`care_logs.csv`、`care_plan_items.csv`、`cages.csv`を基に、`docs/ops/testing/scenarios/S05-hospitalization-cycle.md`のadmitted fixtureを準備する。
- Fixture実査: 全named CSV/S05は実在。hospitalizations/daily/vital/care log/care planは全てheader-only 0行。`cages.csv`は49行あり、clinic 1のactive dog cageはid `3`（small、`:4`）と`4`（medium、`:5`）。runbookのnote操作に必要な`staff_notes.csv`も0行。
- 追加すべき具体値: full-permission staff `1`でpet `1000018`、owner `300003`、`type=hospitalization,start=D,end=D+2,status=admitted,cage_id=3,doctor_id=1,memo="FE12-M04 admitted"`のHを作成する。Hの日付Dへdaily record DR、`09:00 JST/38.5℃/heart 120/respiration 30/8.0Kg/staff 1`のvital、`09:30/type=food/status=completed/value=完食/staff 1`のcare log、朝食と投薬のactive care plan、`10:00/content="FE12-M04 申し送り"/staff 1`のnoteを作る。
- 着手ブロッカー: admitted Hと全child/noteが未作成。cageだけは準備済み。
- 次の一手: fixture担当がH→DR→vital/care log/care plan/noteを作り、実測担当が通常担当者の対照を1回記録後、view-onlyを検証する。最後にadminが対象petを一時死亡登録し、表示と操作不能を確認後に生存へ戻す。
- 注記: 死亡登録/解除は`45b681866`以降fail-closedであり、既死亡への再登録・生存への解除は409になる。一時死亡→復旧の手順は「生存→死亡登録→死亡解除」の順で1回ずつ行う。
- Persona: hospitalization view-only=`view:true/create:false/edit:false/delete:false`。対照として通常担当者=`view/create/edit/delete:true`を1回使う。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: board drag/drop、check-in/status、退院（会計あり/なし）、daily、vital、care log、note、care planのcreate/edit/deleteをpointer/keyboard/programmatic callbackで試す。操作dialogを開いた後に権限を剥奪するcaseも含める。
- Expected result: view-onlyおよびsame-commit剥奪後は全child/top-level mutation 0件。死亡fixtureは死亡文言を表示し、drag/check-in等を実行しない。非活性controlのnameと理由は残る。
- Required evidence artifacts: 操作別network HARと0件集計、accessibility tree、4 viewport screenshot、console log、会計あり/なし退院の個別記録。

### M-05 Clinical sentinel responsive

- Route: Commit `657c1a49cd2c37dc63f5af8e530258a36a12d81e` に記録した25 routeのうち、clinical sentinelを表示する`/medical-records`系、`/hospitalization`系、`/examinations`系、`/vaccinations`系、`/checkups`系、`/`、`/owners`。
- Fixture seed source: `003_demo`の`pets.csv`、`medical_records.csv`、`hospitalizations.csv`、`exams.csv`/`exam_results.csv`、`vaccinations.csv`、`checkups.csv`を基に、S01/S02/S03/S05のfixtureを合成してdeath、danger=`高`、HIGH、LOW、past、today、future、emptyを各1件以上用意する。
- Fixture実査: 全named CSVと`S01-deceased-pet-guard.md`、`S02-exam-abnormal-highlight-lock.md`、`S03-vaccination-next-due-autocalc.md`、`S05-hospitalization-cycle.md`は実在。`pets.csv`はdanger=high 0、death reason非空0。`exam_results.csv`はHIGH/LOW 0。hospitalizations/vaccinations/checkupsはheader-only 0行。S03の「既存vaccination seedは修正済み」という記述とcurrent CSVはdriftしている。
- 追加すべき具体値: M-01の3頭、M-02のM/E、M-04のHを共有する。実測日Dを宣言し、同じalive petへvaccinationとcheckupを各4行作り、`next_date=[D-1,D,D+1,NULL]`、result/labelを`M05-past/today/future/empty`へ固定する。vaccinationは`vaccine_id=1,date=D-7,next_schedule_type=[other,other,other,NULL]`、checkupは`checkup_type_id=1,date=D-7`。CSVへの直接編集は行わない。
- Fixture対応: death=`1001005`、danger high=`1001004`、normal/HIGH/LOW=`M-02 exam E`、past/today/future/empty=`vaccination/checkup各4行`、hospitalization=`M-04 H`。
- 着手ブロッカー: 上記fixtureは現行seedに存在せず、scenario/UI経由のdisposable local準備が未実施。
- 次の一手: fixture担当がM-01→M-02→M-04→日付4区分を同一fixture setへ合成し、M-03 persona準備後、実測担当がfixture-to-cue表を固定してrunbookを実行する。完了後、一時的な死亡/high状態を復旧する。
- Persona: 対象resourceの`view=true`を持つ通常担当者。mutation確認が必要なrowだけ該当action権限ありpersonaを併用する。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: 各fixtureを一覧、選択、登録、編集、詳細で開き、文言、badge、日付、disabled/hidden control、keyboard focus順を確認する。期限は同じ実測日にpast/today/futureを並べる。
- Expected result: death/danger/HIGH/LOW/期限超過が非色cueを持ち、normalとtoday/futureを誤ってdangerにしない。死亡操作はpositive matchで拒否され、全viewportでcue/controlのwrap、clip、overlapなし。Hospitalization Boardの死亡は`3b7524748`以降「死亡」テキストを持つ。
- Required evidence artifacts: route×4 viewport screenshot、accessible name/text dump、computed token、console/network HAR、fixture-to-cue対応表。

### line-reserve font実機確認

- Route: 顧客向け`/line-reserve/{clinicId}`の顧客情報→要望→確認→完了とマイ予約（clinicId抽出契約=`frontend/line-reserve/src/lib/liff-config.ts:6-14`）。
- Fixture seed source: `backend/migrations/seeds/003_demo/line_reservation_settings.csv`、`reservation_types.csv`、`line_customers.csv`、`pets.csv`を基に`docs/ops/testing/scenarios/V05-auth-line-forms.md` V05-6/V05-7、`S04-liff-reservation-journey.md`の試験用顧客を使う。
- 前提実査: 全named CSV、V05、S04は実在。旧ブロッカーだったwebfont宣言は`frontend/line-reserve/index.html:7-12`のGoogle Fonts stylesheet＋2 preconnectで解消済み。`frontend/line-reserve/src/index.css:17-23`はNoto Sans JPを先頭fontへ指定する。
- 解消判定: source前提は解消済みで、現時点から実機確認へ着手できる。残る入力はLINE連携済み試験用account、試験環境、3実機、remote inspectionであり、source修正待ちは無い。
- 入力受渡し順: QA環境管理者がQAチケットに非秘密の試験環境URL、clinic ID、予約確定を許可する試験範囲、credentialの既存安全チャネル上の取得手順を記録する。次に端末管理者がiPhone/Android/iPadの端末ID、OS/browser version、remote inspection可否を同チケットへ割り当てる。credentialそのものはledgerへ記録しない。
- 次の一手: 上記2担当の受渡し完了を開始条件として、実機QA担当が3実機のcold/warm/offlineを実行し、CSS/font file 200、computed font-family、Rendered Fonts、clip/FOIT、fallback操作可否、端末/OS/browser/versionを記録する。
- Persona: LINE連携済み顧客persona。業務に影響しない試験用アカウントを使い、送信/予約確定はrunbookの試験環境だけで行う。
- Viewports: iPhone Safari 390×844、Android Chrome 412×915、iPad Safari 768×1024（加えてdesktop 500×900を比較用）。
- Interaction steps: physical deviceでcold loadし、DevTools/remote inspectionのNetworkでGoogle Fonts CSSとfont fileが200であることを確認する。顧客情報から完了/マイ予約まで遷移し、各画面のcomputed `font-family`と実レンダーfontを確認する。offline/reload時のfallbackも確認する。
- Expected result: `frontend/line-reserve/index.html:7-12`からNoto Sans JPがloadされ、`frontend/line-reserve/src/index.css:17-23`の先頭fontとして全画面へ適用される。clip/FOITによる操作不能がなく、font失敗時もfallbackで操作可能。
- Required evidence artifacts: 3実機の画面別screenshot、remote Network HAR、computed font-familyとRendered Fonts capture、端末/OS/browser/version、cold/warm/offline各結果。

### 未裁定の残余risk

下記は2026-07-27時点のcurrent sourceで未解消を確認したfollow-up候補である。本ledgerは着手・起票・優先度裁定をしない。

- **frontend型検査の未清算** — `a500d424c`以降の全frontend変更unit（`3b7524748`、`e7c978ec9`、`99bac632e`、`24929e83d`、`3c993420a`）で`pnpm type-check`が未実行。禁止コマンドのためUSERが実行する。手順: `docker compose exec -T frontend pnpm type-check`を実行し、診断が出た場合のみ修正unitを切る。判断者: USER。
- **alive petへ予約編集した直後のReception danger一時旧値** — 最初に開く: `frontend/src/features/reception/hooks/use-reception-modal-handlers.ts:135-168`、`use-reception-kanban.ts:362-369`、`frontend/src/hooks/use-update-reservation.ts:43-58`。確認: 成功時local mergeの`updatedAppointment`が新petのdanger/statusを持つか、API responseとrefetchのどれを正本にするか、失敗時rollback範囲。判断者: frontend reception contract owner。手順: stale値を再現するhook testを作り、API response採用／payload補完のcontract確定後にlocal mergeとquery同期を変更する。
- **pet死亡登録の完全同時二重request** — `45b681866`で逐次requestの再登録/取消は409化済み。残るのは完全同時の二重requestが双方ともtransaction前のstatus readを通過し得る狭い窓である。最初に開く: `backend/internal/lstep/lstep_lifecycle_service.go:95,185`と`backend/internal/pet/`の死亡field更新経路。判断者: clinical API owner。手順: pet repository側の条件付きUPDATE（CAS）または行lockで封鎖し、DB並行testで実証する。
- **backend当日予約lookupのclock seam** — frontend側auto-create testは`3c993420a`でfake timers決定化済み。backendの当日予約lookupと同testのclock注入は未導入。最初に開く: backendのmedical-record auto-create lookup実装と同test。判断者: medical-record contract owner。手順: clock seamを定義し、JST日付境界caseをbackend testへ追加する。
- **PermissionGroup以外のmaster permission未配線call site** — 最初に開く: `frontend/src/features/master/hooks/use-master-crud.ts:162`、`use-master-save.ts`と全caller。確認: 現状列挙したCRUD 19、save 24 call siteごとにroute resource、create/edit/delete action、mutation直前のpositive permission checkが対応するか。判断者: frontend RBAC owner。手順: `rg`で両hookのproduction callerを再列挙し、call site×resource×action表を作り、未配線だけを別実装unitへ渡す。
- **tygo pointer mapping 15行** — 最初に開く: `backend/tygo.yaml:17-35,46-64,75-93`と3 generated output。確認: `*uint64`、`*string`、`*bool`、`*time.Time`、`*float64`の5 mapping×3 packageが生成物diffへ寄与するか。判断者: backend/frontend type contract owner。手順: 許可された`make codegen`で各mappingの出力寄与を個別記録し、寄与0の行だけを設定整理unitへ渡す。
- **manual chunkの未再計測** — 最初に開く: `frontend/src/features/manual/lib/manual-index.ts:50-61,86-87`と`frontend/src/app/routes/operations-routes.tsx:132-150`。確認: 2つのMarkdown globが`eager:true/?raw`でmanual chunkへ入る現在byte、500 kB警告、route lazy境界。判断者: frontend verification gate owner（実行許可）→frontend performance owner（計測）。手順: gate ownerが他sessionのfrontend WIP静止とfull build許可をtask logへ明記し、performance ownerへ引き渡す。引渡し後だけperformance ownerが`docker compose exec frontend pnpm build`を実行し、command/exit、`dist/assets`のmanual chunk名/byte、500 kB警告、build前後statusをartifact化する。警告該当時は比較案を作り、build未実行のまま解消扱いにしない。
- **AuthProviderのfeature barrel経由eager import** — 最初に開く: `frontend/src/app/router.tsx:1-23`、`frontend/src/features/auth/index.ts:1-10`、`hooks/use-auth.tsx:1-19`。確認: routerの同期`@/features/auth` importがLogin/Forgot/Reset等のbarrel exportを同じgraphへ含めるか、public boundaryを維持するentry設計、auth route chunk。判断者: frontend architecture owner。手順: current import graphとchunkを記録し、barrel分割／provider専用public entry／現状維持の各contractを比較してから変更surfaceを確定する。

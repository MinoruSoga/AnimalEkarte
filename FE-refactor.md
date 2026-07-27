# FE-refactor — FE12-02 active-only ledger

> 更新: 2026-07-27（要件責任者: 曽我）
> 業務目的: 未完了の臨床sentinel・RBAC安全境界を、決裁と実測の証跡が揃うまで追跡する。
> 本ファイルは使い捨てのactive ledger。恒久規約は `DESIGN.md`、`docs/spec/design-system.md`、`docs/spec/ui-design-compliance.md`、`frontend/CLAUDE.md` を正本とする。
> 解消済みの証跡: F16/U10/O-A実装=`3b7524748`、R-A/M-A=`e7c978ec9`、P-A=`99bac632e`、exam parent+items原子化=`24929e83d`、死亡API 409 fail-closed=`45b681866`、test-hardening 3件=`3c993420a`。代理決裁6件の記録と調査パックは`a500d424c`および本更新直前までの履歴に保存する。
> S1/S2並行レーン（2026-07-27完走・全項目解消）の証跡: pet死亡CAS封鎖=`18d307076`、カルテlookup JST正規化=`305b50c7f`、ワクチン期限JST契約=`56d18eb18`、auto-create原子化=`5e5868549`、DBOrTx inventory債務=`0eecddb11`、reception danger sentinel=`082be9961`、master permission 45/45配線+fail-closed化=`da550a84d`、auth独立chunk化+checkup resource整合=`8a00a4794`、StaffSettings test追随=`e8a4db982`。各項目の分析・裁定・証跡の全文は本節削除直前の版（commit `0fd1f7d18`）に保存する。裁定の要点2件は後段「維持する裁定」に残す。
> 削除した `## FE12-02 unit execution record` と過去のFE/R-F履歴はCommit `657c1a49cd2c37dc63f5af8e530258a36a12d81e` に保存する。

## Active scope and authority

- 追跡対象は `M-01`〜`M-05`、line-reserve font実機確認、裁定待ちの残余2件とする。F16・F9・U10・MEDIUM 4件およびS1/S2並行レーン全9項目は代理決裁・実装済みで追跡を終了した。
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
| R-1 | P2 | staff入力検証の契約drift 5件 | 仕様裁定（frontend RBAC owner + staff API contract owner） | 各driftの是正または「意図された挙動」の明文化 |
| R-2 | P2 | tygo pointer mapping 15行の寄与測定 | `make codegen`がUSER専権 | 寄与0の行の特定と設定整理 |
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
- Fixture実査: D=`2026-07-27`。API read-backで`1001002={status:alive,danger_level:low,deceased_at:null}`を対照確認し、`1001005={status:deceased,danger_level:low,deceased_at:2026-07-27T00:00:00+09:00}`への死亡登録はHTTP 204で準備済み。R6では`1001004`へ逐語body `{"danger_level":"high"}`を送ってHTTP 400 `{"error":"危険度がhighの場合は危険理由を入力してください"}`、1 fieldだけ足した`{"danger_level":"high","danger_reason":"FE12-M01 fixture"}`もHTTP 400 `{"error":"入力値が正しくありません"}`となった。後者はruntime DBの`pets.danger_reason`列欠落（SQLSTATE 42703）でrollbackされ、read-backは`status=alive,danger_level=low,danger_reason:null`のまま。
- 追加すべき具体値: disposable local上の目標値は引き続き`1001002={status:alive,danger_level:low,death fields:null}`、`1001004={status:alive,danger_level:high,death fields:null}`、`1001005={status:deceased,danger_level:low,deceased_at:2026-07-27T00:00:00+09:00}`。`1001002`と`1001005`は準備済みで、`1001004`だけ未達。schema整合後に再実行する契約上の最小payloadは`{"danger_level":"high","danger_reason":"FE12-M01 fixture"}`である。
- 着手ブロッカー: runtime DBに現行sourceが更新する`pets.danger_reason`列が無く、契約上正しいhigh+非空reason PATCHもSQLSTATE 42703でrollbackされる。pet public responseが`deceased_reason`を公開しないことは現行契約どおりであり、残ブロッカーには数えない。
- 次の一手: 別の認可工程でdisposable local DB schemaを現行migration契約へ合わせた後、fixture担当が最小high+reason payloadを再実行して`1001004={status:alive,danger_level:high,death fields:null}`をread-backする。その後に実測担当と曽我または同席する仕様責任者がrunbookを実行し、完了後fixture担当が`1001005`をalive、`1001004`をlowへ復旧する。
- Persona: owners `view=true, edit=true, delete=true`の通常担当者。曽我または同席する仕様責任者が許可操作を記録する。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: include-deceasedを有効にし、3 rowそれぞれでowner名link、report、編集、削除、pet死亡登録/解除をpointer、Tab/Enter/Space、直接URLで試す。各操作前後のrequestをnetworkで記録する。
- Expected result: aliveは通常操作可、danger=`高`は「⚠ 危険」等の非色cueとaccessible nameを失わない。deceasedはF16裁定（一覧はグレーアウト維持・`3b7524748`）どおりの表示を記録し、曽我が許可/禁止する操作を操作単位で明示する。禁止とされた操作はmutation 0件。
- Required evidence artifacts: 4 viewport screenshot、accessibility tree、各操作のaccessible name、GET以外のnetwork HAR、操作可否の曽我裁定メモ。

### M-02 Examinations一覧意味とlayout

- Route: `/examinations`と対象petの`/medical-records/:id`検査履歴。
- Fixture seed source: `backend/migrations/seeds/003_demo/exams.csv`、`exam_results.csv`、`exam_types.csv`、`exam_type_fields.csv`を基に、`docs/ops/testing/scenarios/S02-exam-abnormal-highlight-lock.md`のhigh/low/normal同居fixtureを作る。
- Fixture実査: D=`2026-07-27`。pet `1000018`のdraft medical record M=`1425546`（record_no=`FE12-M02-M-20260727`、owner=`300003`、doctor=`1`）をread-backした。R6指定の`date="2026-07-27T09:00:00+09:00"`を含む逐語bodyでE作成を実行したがHTTP 400 `{"error":"入力値が正しくありません"}`となり、runtime DBの`exam_type_fields.clinic_id`列欠落（SQLSTATE 42703）によりexam type ownership確認で停止した。作成前後ともEは0件で、items PUTは未実行。
- 追加すべき具体値: M=`1425546`へ`exam_type_id=3,doctor_id=1,date="2026-07-27T09:00:00+09:00",status=result_entered,result_summary="FE12-M02 HIGH/LOW/normal"`のexam Eを作り、EへWBC=`10.0`、RBC=`9.0`、HCT=`30.0`を登録する。`status/is_abnormal`はitem requestへ書かず、backendがspecies別`exam_reference_ranges`から導出したWBC=normal・RBC=high・HCT=lowをread-backする。
- 着手ブロッカー: runtime DBに現行sourceがclinic scopeへ使う`exam_type_fields.clinic_id`列が無く、E作成がSQLSTATE 42703で停止する。導出機構は`exam_reference_ranges`を参照して実装済みだがdirect read APIは無く、live rangeの存在は未確認でBLOCKED。demo seedにrange recordが無いことはsource上確認済みであり、range不在なら3値は未評価normalとなる。
- 次の一手: 別の認可工程でdisposable local DB schemaを現行migration契約へ合わせ、clinic/pet species/type `3`用のWBC/RBC/HCT reference rangeを正規のmaster-data工程で用意する。その後fixture担当がM=`1425546`へE→3 itemsを作成し、`/examinations/:id/items`でbackend導出値をread-backしてから実測担当が一覧とカルテ履歴を比較する。
- Persona: examinations `view=true, create=false, edit=false, delete=false`のview-only担当者。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: 同一fixtureを一覧とカルテ履歴で開き、値・summary・statusを比較する。zoom 100%、pointer hover、keyboard focus、横スクロール有無を確認し、一覧でHIGH/LOW非色cueが必要か曽我へ提示する。
- Expected result: normalを異常表示しない。曽我が一覧cueを必要と裁定する場合はHIGH/LOWが色なしでも識別できること、不要と裁定する場合は詳細へのaccessible導線があること。全viewportでwrap/clip/overlapなし。
- Required evidence artifacts: 両surface×4 viewport screenshot、accessible text dump、computed color/token、overflow計測、曽我の一覧cue要否裁定。

### M-03 RBAC非活性の理由/name

- Route: `/settings/permission-groups`、`/medical-records/:id`、`/hospitalization/:id/edit`、`/vaccinations/:id`、`/examinations/:id`。
- Fixture seed source: `backend/migrations/seeds/003_demo/permission_groups.csv`、`permission_group_rules.csv`、`staffs.csv`、`staff_permission_groups.csv`と各featureの既存record CSV。`docs/ops/testing/scenarios/V03-owner-pet-staff-forms.md`で試験用group/staffを作り、`V01-clinical-forms.md`の既存recordを使う。
- Fixture実査: M-04 H=`1`とM-05 M2=`1425547`/vaccination/checkupは準備済みだが、依存するM-02 exam Eはruntime schema driftによりHTTP 400で未作成。依存順序を守ってR6でもM-03 mutationには着手しておらず、API read-backは`FE12-M03-*` group=`[]`、`fe12-m03-*` staff=`[]`。
- 追加すべき具体値: `master-permission`、`medical-records`、`hospitalization`、`vaccinations`、`examinations`へ共通で、`FE12-M03-VIEW={true,false,false,false}`、`FE12-M03-CREATE={true,true,false,false}`、`FE12-M03-EDIT={true,false,true,false}`の3 groupを作り、`fe12-m03-view@noavet.jp`、`fe12-m03-create@noavet.jp`、`fe12-m03-edit@noavet.jp`の各専用staffへ割り当てる。passwordはdemo seedと同じ`password`。操作対象は別group `FE12-M03-TARGET`とする。
- 着手ブロッカー: M-02 Eが未完了のため、依存順序上M-03の4 group・3 staff・割当・persona login検証を開始できない。加えてsource調査では、指定password=`password`が現行staff作成serviceの英字+数字要件を満たさず、未作成accountはwrite前に拒否される契約不一致がある。
- 次の一手: M-02 E/3 itemsのread-back完了と、demo fixture credential契約・staff password validatorの整合を別の認可工程で解決した後、fixture担当がtarget group→3 group→3 staff作成・割当を行い、3 emailのlogin HTTP 200を確認してから実測担当がpersonaごとに再ログイン/再読込してrunbookを実行する。
- Persona: (1) view-only=`view:true/create:false/edit:false/delete:false`、(2) create-only=`view:true/create:true/edit:false/delete:false`、(3) edit-without-delete=`view:true/create:false/edit:true/delete:false`。
- 注記: 検査登録（POST）は`24929e83d`以降create+editの合成認可であり、create-only personaは拒否されるのが正しい期待値である。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: 各personaでrouteを再ログイン/再読込して開き、pointer、Tab/Enter/Space、formのprogrammatic submit、保存中の権限剥奪後callbackを試す。permission-groupは新規panel、既存panel、reorder、保存後rulesも個別確認する。
- Expected result: view accessは維持する。許可されたactionだけ実行でき、禁止controlはaccessible nameと理由を保持する。禁止personaからのPOST/PATCH/PUT/DELETEは0件で、same-commit剥奪後もmutationしない。
- Required evidence artifacts: persona×routeのaccessibility tree、4 viewport screenshot、network HAR、console log、action別permission matrixと0 mutation集計。

### M-04 Hospitalization child control実効性

- Route: `/hospitalization`のboard/listと`/hospitalization/:id`。
- Fixture seed source: `backend/migrations/seeds/003_demo/hospitalizations.csv`、`daily_records.csv`、`vital_records.csv`、`care_logs.csv`、`care_plan_items.csv`、`cages.csv`を基に、`docs/ops/testing/scenarios/S05-hospitalization-cycle.md`のadmitted fixtureを準備する。
- Fixture実査: D=`2026-07-27`。APIでH=`1`、DR=`1`、vital=`1`、care log=`1`、朝食care plan=`1`、投薬care plan=`3`（medicine_id=`14001`）、note=`1`を作成した。Hはpet=`1000018`/owner=`300003`/cage=`3`/doctor=`1`/status=`admitted`/memo=`FE12-M04 admitted`、全childは指定日時・値でread-back済み。
- 追加すべき具体値: full-permission staff `1`でpet `1000018`、owner `300003`、`type=hospitalization,start=D,end=D+2,status=admitted,cage_id=3,doctor_id=1,memo="FE12-M04 admitted"`のHを作成する。Hの日付Dへdaily record DR、`09:00 JST/38.5℃/heart 120/respiration 30/8.0Kg/staff 1`のvital、`09:30/type=food/status=completed/value=完食/staff 1`のcare log、朝食と投薬のactive care plan、`10:00/content="FE12-M04 申し送り"/staff 1`のnoteを作る。
- 着手ブロッカー: fixture未作成ブロッカーは解消済み。投薬care planはruntime必須contractに従いmedicine_id=`14001`を指定した。
- 次の一手: 実測担当がH=`1`で通常担当者の対照を1回記録後、view-onlyを検証する。最後にadminが対象petを一時死亡登録し、表示と操作不能を確認後に生存へ戻す。
- 注記: 死亡登録/解除は`45b681866`以降fail-closedであり、既死亡への再登録・生存への解除は409になる。一時死亡→復旧の手順は「生存→死亡登録→死亡解除」の順で1回ずつ行う。
- Persona: hospitalization view-only=`view:true/create:false/edit:false/delete:false`。対照として通常担当者=`view/create/edit/delete:true`を1回使う。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: board drag/drop、check-in/status、退院（会計あり/なし）、daily、vital、care log、note、care planのcreate/edit/deleteをpointer/keyboard/programmatic callbackで試す。操作dialogを開いた後に権限を剥奪するcaseも含める。
- Expected result: view-onlyおよびsame-commit剥奪後は全child/top-level mutation 0件。死亡fixtureは死亡文言を表示し、drag/check-in等を実行しない。非活性controlのnameと理由は残る。
- Required evidence artifacts: 操作別network HARと0件集計、accessibility tree、4 viewport screenshot、console log、会計あり/なし退院の個別記録。

### M-05 Clinical sentinel responsive

- Route: Commit `657c1a49cd2c37dc63f5af8e530258a36a12d81e` に記録した25 routeのうち、clinical sentinelを表示する`/medical-records`系、`/hospitalization`系、`/examinations`系、`/vaccinations`系、`/checkups`系、`/`、`/owners`。
- Fixture seed source: `003_demo`の`pets.csv`、`medical_records.csv`、`hospitalizations.csv`、`exams.csv`/`exam_results.csv`、`vaccinations.csv`、`checkups.csv`を基に、S01/S02/S03/S05のfixtureを合成してdeath、danger=`高`、HIGH、LOW、past、today、future、emptyを各1件以上用意する。
- Fixture実査: D=`2026-07-27`、共有alive pet=`1001002`。draft M2=`1425547`（record_no=`FE12-M05-M2-20260727`）配下のcheckup `1/2/3/4`と、top-level vaccination `1/2/3/4`をAPIで作成し、各seriesの`M05-past/today/future/empty`と`next_date=[2026-07-26,2026-07-27,2026-07-28,NULL]`をread-backした。
- 追加すべき具体値: vaccination `1/2/3/4`は`pet_id=1001002,vaccine_id=1,date=2026-07-20,next_schedule_type=[other,other,other,NULL]`でlabelを`remarks`へ保存した。checkup `1/2/3/4`はM2=`1425547`配下の`pet_id=1001002,checkup_type_id=1,date=2026-07-20`でlabelを`result`へ保存した。
- Fixture対応: death=`1001005`（登録済み）、danger high=`1001004`（PATCH 400で未達）、normal/HIGH/LOW=`M-02 exam E`（POST 400で未達）、past/today/future/empty=`vaccination 1/2/3/4 + checkup 1/2/3/4`（準備済み）、hospitalization=`M-04 H 1`（準備済み）。
- 着手ブロッカー: M-05固有のM2/vaccination/checkup fixture未作成ブロッカーは解消済み。ただし合成fixture set全体ではM-01 danger high、M-02 E、M-03 personaが未完了。
- 次の一手: M-01/M-02のAPI blockerを解消してM-03 personaを準備後、実測担当が上記IDでfixture-to-cue表を固定してrunbookを実行する。完了後、一時的な死亡/high状態を復旧する。
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

### 残余risk — 裁定待ち

2026-07-27にS1（backend）/S2（frontend）の2レーン並行で全9項目を解消・清算した。実装・裁定が完了した項目は本節から削除し、証跡commitはファイル冒頭に一覧する。以下は実装ではなく**裁定**を待っている2件である。

- **【裁定待ち】staff入力検証の契約drift 5件** — S2第4走が隣接riskとして列挙（実装未着手）。①新規emailがtruthy判定のみで空白・形式不正をfrontendで阻止しない ②既存staffの非空・短いpasswordをfrontendで阻止しない ③frontend/OpenAPIの「新規email/password必須」とbackendのアカウントなしstaff作成許容が不一致 ④空氏名test名は新規/編集双方を示すが実caseは編集経路のみ ⑤read-only画面でEnterによるform actionが残り、基本staff保存が拒否されても関連付けmutationが独立発行され得る（backend側拒否は未評価）。判断者: frontend RBAC owner ＋ staff API contract owner。次の一手: 各driftについて「是正する」か「意図された挙動として明文化する」かを決める。③は正本がfrontend/OpenAPI側かbackend側かの決定が先行する。

- **【レーン外・USER専権】tygo pointer mapping 15行** — 最初に開く: `backend/tygo.yaml:17-35,46-64,75-93`と3 generated output。確認: `*uint64`、`*string`、`*bool`、`*time.Time`、`*float64`の5 mapping×3 packageが生成物diffへ寄与するか。判断者: backend/frontend type contract owner。手順: 許可された`make codegen`で各mappingの出力寄与を個別記録し、寄与0の行だけを設定整理unitへ渡す。`make codegen`がUSER専権のためエージェントレーンへ割り当てない。

### 維持する裁定（再提案を防ぐため保持）

- **カルテ同日重複にDB unique制約を採らない（2026-07-27）** — 同一pet同日に手で複数カルテを作ることは正当な業務（別々の来院）であり、制約は正当な操作まで禁止する。塞ぐべきは自動生成経路が同じ1回の来院に対して二重に作ることだけであり、これは`5e5868549`のtry-advisory-lockで自動生成経路に限定して解決済み。
- **auto-createにclock seamを導入しない（2026-07-27）** — 重複チェック日は`reservation.StartTime`由来であり現在時刻を参照しない。clock seamを入れると過去/未来予約の検索日が実行日へ変わり挙動を壊す。予約日基準の現行contractが正である。
- **manual chunkの追加分割投資を行わない（2026-07-27）** — 実測522.71 kB（gzip 145.80 kB）で500 kB警告に該当するが、`operations-routes.tsx`のlazy境界により独立chunkとして正しく分割済みで、`/manual`を開いた利用者だけが取得する。build警告は汎用閾値であって業務上の問題の証拠ではなく、表示遅延の申告も無い。存在する問題の証拠なしに最適化するのはproduct-philosophy①違反である。警告閾値の引き上げによる黙らせも行わない（サイレント化は⑤の禁止事項）。再開条件=manual画面の表示時間に関する具体的な業務上の申告が出た場合のみ。

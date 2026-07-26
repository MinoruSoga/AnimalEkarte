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
- `docs/spec/ui-design-compliance.md` にはC18件数ratchetの旧記述が残るが、現行auditはratchetを持たない。本ledgerは正本側を編集せず、別scopeの文書driftとして扱う。是正は`docs/spec/ui-design-compliance.md`を所有する別unitが行う。

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

### 着手可能化全数表（23件）

下表の「次の一手」は本unitが実行する作業ではない。本unitは裁定、runtime実測、seed適用、実装を行わず、開始条件と担当を一意化する。

| 項目 | 解消判定 | 着手ブロッカー（欠けている情報・fixture） | 担当 | 次の一手 |
|---|---|---|---|---|
| F16 | 未解消 | 死亡をgeneric grayと`C.danger`のどちらへ統一するかに加え、Hospitalization Detailを統一対象へ含めるか、現行gray-out＋danger badgeを維持するか、shared contractかhospitalization固有かが未裁定 | 要件責任者=曽我 | 本節の選択肢(a)(b)とU10の3裁定項目を同時に明示する。裁定後に仕様owner、frontend test owner、frontend writerの順で引き渡す |
| F9 | 未解消 | 要フォローsentinelが解決する業務、読む人、set/clear主体、状態型、一覧queueの有無、既存Lステップ無条件配信との関係、監査要否が未定義 | 要件責任者=曽我 | 本節の「追加しない／目的と意味を定義して追加」のいずれかを明示する。追加する場合は7契約を埋めてbackend/frontend実装unitへ渡す |
| U10 | 未解消 | token決定箇所は全数特定済みだが、F16とDetail/shared範囲が未裁定 | 要件責任者=曽我 | F16裁定と同時にDetail/shared範囲を明示し、U10一覧の該当testをRED化するfrontend unitへ渡す |
| M-01 | 未解消 | 同一ownerの3頭候補は特定済みだが、`danger_level=high`が0件、死亡日時＋理由が揃う死亡petが0件 | fixture担当、実測担当、同席する仕様責任者 | owner `300588` のpet `1001004`をhigh、`1001005`を死亡日時＋理由ありへdisposable localで準備してからrunbookを実行し、最後に復旧する |
| M-02 | 未解消 | 全1,322,503 exam resultがnormalで、同一examのnormal/HIGH/LOW fixtureが0件 | fixture担当、実測担当、要件責任者=曽我 | alive pet `1000018`へdraft medical recordとexam type 3のexamを作り、WBC=10.0、RBC=9.0、HCT=30.0を入力して導出status確認後にrunbookを実行する |
| M-03 | 未解消 | view-only既存groupは`master-permission` view=false、create-onlyとedit-without-deleteは0件。mutable clinical recordと操作対象groupも不足 | admin fixture担当、実測担当 | feature fixtureと操作対象groupを作成後、VIEW/CREATE/EDITの3 groupと専用staffを作成・割当し、各personaで再ログインしてrunbookを実行する |
| M-04 | 未解消 | hospitalization、daily、vital、care log、care plan、staff noteがすべて0件。active cageだけ存在 | fixture担当、実測担当 | pet `1000018`、cage `3`でadmitted hospitalizationと全child/noteを作成し、通常担当者の対照実測後にview-onlyと一時死亡状態を検証して復旧する |
| M-05 | 未解消 | death reason、danger=high、HIGH/LOW、past/today/future/empty、hospitalizationの各fixtureが不足 | fixture担当、S01/S02/S03/S05担当、実測担当 | 本節のfixture対応表を同じowner/petへ合成し、fixture-to-cue表を固定してから対象routeを実測し、最後に一時変更を復旧する |
| line-reserve font | 解消済み（source前提）／runtime未実行 | 旧ブロッカーだったwebfont load宣言は解消済み。3実機、LINE連携済み試験用account、試験環境、remote inspectionの受渡しmanifestが未作成 | QA環境管理者→端末管理者→実機QA担当 | QA環境管理者が非秘密の対象URL/clinic IDとcredential受渡し手順をQAチケットへ記録し、端末管理者が3実機とremote inspection可否を割り当てた後、実機QA担当がrunbookを実行する |
| reception optimistic rollback | 未解消 | 即時反映の要否、失敗時のカード位置、成功toast時点、再試行導線、再同期表示が未裁定 | 要件責任者=曽我 | MEDIUM決裁パックのR-A/R-Bと失敗時表示を明示し、failure matrixをfrontend unitへ渡す |
| permission-group parent/rules partial success | 未解消 | parentのみ残る状態の許容、dirty表示、再送単位、create/update共通契約が未裁定 | 要件責任者=曽我 | MEDIUM決裁パックのP-A/P-B/P-CとUI状態を明示し、整合性・監査・権限剥奪matrixを実装unitへ渡す |
| owner parent/pet partial success | 一部解消（backend atomic write実在）／業務裁定は未解消 | backendのowner+pets同一transactionは実在するが、UIがpartialを残すか全体失敗にするか未裁定 | 要件責任者=曽我 | MEDIUM決裁パックのO-A/O-B/O-C、失敗pet入力保持、遷移先、通知を明示してfrontend unitへ渡す |
| medical auto-create retry UX | 未解消 | 同画面滞在／別route、手動／自動／再入場、予約成功後の再利用／取消、通知と上限が未裁定 | 要件責任者=曽我 | MEDIUM決裁パックのM-A/M-B/M-Cと失敗phase契約を明示し、duplicate preventionと権限再確認testを含むunitへ渡す |
| Reception danger一時旧値 | 未解消 | 予約編集成功時のlocal mergeへ新petのdanger値を反映する契約と失敗時rollback境界が未定義 | frontend contract owner | `use-reception-modal-handlers.ts`と`use-reception-kanban.ts`を起点に、API responseを採るかlocal payloadを補完するかを決め、stale/refetch/failure testを作る |
| 死亡record再登録Conflict | 未解消 | 既死亡判定とHTTP Conflict契約がservice/repositoryに無い | clinical API owner、要件責任者=曽我 | death handler→lifecycle service→pet repositoryを開き、既死亡時のidempotent success／Conflictと監査の扱いを定義してtestから着手する |
| auto-create JST境界flaky | 未解消 | frontendの`new Date()`とbackend当日予約lookupのclock注入・JST境界fixtureが未定義 | medical-record contract owner | frontend auto-create hookとbackend auto-create lookup testを開き、clock seamとJST日付境界caseを定義して固定時刻testを作る |
| master permission未配線call site | 未解消 | PermissionGroup以外の`useMasterCRUD` 19、`useMasterSave` 24 call siteのresource/action対応表が無い | frontend RBAC owner | 2 hookの全callerを列挙し、route resourceとcreate/edit/deleteの対応表を作成してから未配線箇所を判定する |
| tygo pointer mapping 15行 | 未解消 | 3 packageに重複する5 pointer mappingが生成物へ効くかをcodegen diffで確認していない | backend/frontend type contract owner | `backend/tygo.yaml`の3 mapping blockと生成物を対応付け、許可されたcodegenで差分を測り、出力0なら設定整理unitへ渡す |
| CompletePage malformed時刻 | 未解消 | 4文字超入力の期待表示とshared formatter契約、専用regression testが無い | line-reserve contract owner | `CompletePage.tsx`と`shared-liff/jst-date.ts`を開き、malformed入力の表示contractを決めて隣接testを作る |
| ConfirmPage未assert 3契約 | 未解消 | pending二重送信、LIFF送信順、409時inline error非表示の直接assertが無い | line-reserve test owner | `ConfirmPage.tsx`と既存testを開き、3観測点を独立caseとして追加する |
| manual chunk未再計測 | 未解消 | 最新許可buildのchunk値と、full buildを実行可能にする明示許可・frontend WIP静止記録が無い | frontend verification gate owner→frontend performance owner | gate ownerがfrontend WIP静止とfull build許可をtask logへ記録し、その引渡し後にperformance ownerがDocker buildのcommand/exit、manual chunk名/byte、500 kB警告、前後statusをartifact化する |
| AuthProvider eager import | 未解消 | `app/router.tsx`がauth feature barrelを同期importし、public boundaryを維持したlazy分離contractが未定義 | frontend architecture owner | routerとauth indexのimport graphを計測し、barrel分割／provider専用public entry／現状維持のcontractを決めてchunk実測へ渡す |

### 決裁済み（2026-07-26・曽我からの委任によりAIが代理決裁）

以下3件を代理決裁した。以降の各節は裁定の根拠となった調査記録である。曽我はこの裁定を覆せる。

**F16裁定: surfaceで分ける。一覧はグレーアウト、単一患者画面は`C.danger`を維持する。**

- 根拠: `C.danger`は「今すぐ対応せよ」を意味する。危険（咬傷リスク）と異常高（値の解釈が要る）はいずれも対応を要求するが、**死亡は対応を要求せず遮断を要求する**。遮断は既にpositive-match guardで実装済みであり、色は遮断機構ではなく状態表示である。
- 一覧で死亡を`C.danger`にすると、`/owners?include_deceased=true`等で死亡行が赤く並び、**その中の危険ペット1頭を見失う**。アラート疲労により赤の安全機能そのものが劣化する。これが一覧でグレーを採る決定的理由である。
- 単一患者画面（`PatientInfoCard.tsx:112-116`の現行`C.bgDanger`【死亡】）は維持する。走査対象が無いためアラート疲労が起きず、かつ操作直前の地点であるため強いmarkerが誤入力を防ぐ。
- したがって`docs/spec/design-system.md:84`と`:89`はどちらも正しく、**文脈修飾が欠けているだけ**である。いずれも削除せず、§2.4へ「死亡は一覧ではグレーアウト、単一患者画面では`C.danger`」を追記する。
- 下位裁定4件: (1) 上記のとおり文脈分離。(2) 一覧surface全て（OwnersList／HospitalizationList／Board）へ共通適用する。走査文脈の一貫性が目的だからである。(3) Detailはgray-outと`C.danger`バッジを両方維持する。前者が操作不能を、後者が状態を示し役割が異なる。(4) shared `PatientInfoCard` contractは変更しない。

**U10裁定: token置換は発生しない。Boardへ死亡テキストラベルを追加する。**

- `HospitalizationListView.tsx:49,52,77-79`と`HospitalizationBoard.tsx:40,69-78`はいずれも既に`opacity-40`のグレーアウトであり、F16裁定「一覧はグレー」の下では置換対象が消滅する。
- 残る実欠陥は**Boardに死亡文言が無いこと**である。グレーのみは色だけの手掛かりであり、`## Active execution rules` 1（死亡・危険は非色cueを伴う）に違反する。ListViewは「死亡」テキストを持つため適合済み。
- 併せて`frontend/src/lib/status-helpers.ts:176-178`の`getPetStatusColor`をpositive match化する。`生存 ? green : gray`の二値では死亡と未知statusが区別できない。これはF16と独立の欠陥である。

**F9裁定: 要フォローsentinelを追加しない。**

- `docs/product-philosophy.md` ①適用。「要フォローsentinelが無い」は観測であって要件ではない。責任者名も業務上の目的（誰の何の作業がどれだけ遅いか）も述べられておらず、名前の無い要件は疑えないため受け入れない。
- ②適用。`TriggerCheckupFollowUp`が健診作成時に無条件でLステップ配信を発火している。per-recordの要フォローflagを足せばフォローアップ概念が2つになり、二重管理の設計禁止に反する。
- 再提起の条件: 誰の作業が、現在どれだけ時間を要し、**既存の無条件配信では何が満たせないか**。この3点が揃うまで着手しない。

#### FE12-02-F16: OwnersListの死亡表示がgeneric gray token

- Authority A: `docs/spec/design-system.md:84` は「**危険 / 死亡 / 異常高**: `C.danger`」と明記する。
- Authority B: 同じ§2.4の `docs/spec/design-system.md:89` は「危険バッジ・**死亡グレーアウト**・RBAC 非活性表示」を退行させないと明記する。AとBは死亡表現について相互に競合し、どちらも本unitでは優先しない。
- Current code: `frontend/src/lib/status-helpers.ts:176-178` の`getPetStatusColor`は`status === "生存"`だけをgreen、それ以外をすべて`BADGE.grayHover`へまとめる。死亡だけでなく未知の非生存statusも同じgeneric grayとなる。
- Decision request: 要件責任者=曽我が、(a) generic grayを死亡表示の正本とする、または(b) `C.danger`を死亡へ適用してauthorityの「死亡グレーアウト」文を改訂する、のいずれかを明示裁定する。本ledgerは選択しない。
- Dependency: U10のHospitalizationList/Board death token置換はこの決裁まで着手しない。

影響範囲の実測結果:

- 検索: `rg -n --glob '*.{ts,tsx}' '\bgetPetStatusColor\b' frontend` は7 hit / 3 file。production定義1、production callerのimport/呼出し各1、test 4で、production call expressionは1件だけ。
- runtime chain: `/owners` → `frontend/src/app/routes/clinical-general-routes.tsx:35-50` → `frontend/src/app/pages/OwnersListPage.tsx:7-11` → `frontend/src/features/owners/routes/OwnersList.tsx:50,299-318` → `frontend/src/features/owners/components/OwnersListTable.tsx:17,216-221`。
- `/owners/:id`と`/owners/:id/report`はcallerが無いため直接影響外。

| 選択肢 | production変更対象 | 仕様変更対象 | 影響route | 既存testと件数 |
|---|---|---|---|---|
| (a) generic gray | なし。`status-helpers.ts:176-178`の現行分岐と`OwnersListTable`のcallerは一致 | `docs/spec/design-system.md:84`の死亡=`C.danger`記述を`:89`の死亡グレーアウトと矛盾しない契約へ変更 | `/owners` | `frontend/src/lib/status-helpers.test.ts` 78件中直接2件（生存green、死亡gray）、`OwnersListTable.report.test.tsx` 7件中直接1件（死亡表示）。unknown fallbackとpresentation tokenのassertは0件 |
| (b) `C.danger` | `frontend/src/lib/status-helpers.ts:176-178`を生存green／死亡danger／未知値grayへ分ける。caller側production editは不要 | `docs/spec/design-system.md:89`の死亡グレーアウト記述を`:84`と矛盾しない契約へ変更 | `/owners` | 同2 file・合計85件。死亡gray期待1件が変更対象、presentation-level danger assertionは現状0件 |

裁定時に曽我が同時に埋める項目:

1. (a)/(b)のどちらか。
2. 選んだ表現をOwnersListだけに適用するか、Hospitalization List/Board/Detailへ共通適用するか。
3. Hospitalization Detail現行のgray-out＋danger badgeを維持するか、単一表現へ置換するか。
4. Detailはshared `PatientInfoCard` contractを変更するか、hospitalization固有表現に限定するか。

裁定後の引渡し: 仕様ownerが該当authorityを更新し、frontend test ownerが対象期待値をRED化し、frontend writerが裁定で確定したproduction surfaceだけを変更し、`docker compose exec frontend npx vitest run src/lib/status-helpers.test.ts src/features/owners/components/OwnersListTable.report.test.tsx`とU10の該当component testを実行する。

#### FE12-02-U10: Hospitalization死亡表示tokenの対象全数

検索コマンド:

```bash
rg -n --glob '*.{ts,tsx}' --glob '!*.test.*' "(HospitalizationPatientHeader|petIsDeceased|isDeceased|deceasedAt|deceased_at|status\s*===\s*[\"'](死亡|deceased)[\"']|[\"']死亡[\"']|[\"']deceased[\"'])" frontend/src/features/hospitalization frontend/src/components/shared/PatientInfoCard/PatientInfoCard.tsx frontend/src/components/shared/PatientContextHeader/PatientContextHeader.tsx
```

対象判定:

| 区分 | path:line | 現行の役割 | route / test |
|---|---|---|---|
| token決定: List | `frontend/src/features/hospitalization/components/HospitalizationListView.tsx:49,52,77-79` | 行`opacity-40`、detail link抑止、`C.text40`の「死亡」 | `/hospitalization` List。隣接test 3件、死亡1件、class assertion 0 |
| token決定: Board | `frontend/src/features/hospitalization/components/HospitalizationBoard.tsx:40,69-78` | `C.bgPage`、`C.borderPrimary20`、`opacity-40`、死亡文言なし | `/hospitalization` Board。隣接test 5件、死亡0 |
| token決定: Detail shared | `frontend/src/components/shared/PatientInfoCard/PatientInfoCard.tsx:78,83,91,109,112-116` | `C.bgPage60`、avatar grayscale、pet名`C.text60`、`C.bgDanger/C.textWhite`の`【死亡】` | `/hospitalization/:id` desktop/mobile。隣接test 4件、死亡0 |
| data bridge | `frontend/src/features/hospitalization/api/transforms.ts:33` | backend `deceased`を`petIsDeceased`へ変換 | 全hospitalization表示。token非決定 |
| Detail bridge | `frontend/src/features/hospitalization/components/HospitalizationPatientHeader.tsx:15-20` | `petIsDeceased`を`PatientInfoCard status`へ変換 | Header test 2件、死亡marker 2件、class assertion 0 |
| placement | `frontend/src/features/hospitalization/routes/HospitalizationDetail.tsx:68-70` | desktop/mobile variantへ同一dataを渡す | Detail test 3件、死亡0 |
| placement | `frontend/src/features/hospitalization/components/HospitalizationExpandedView.tsx:33-37` | desktop Header配置 | 隣接testなし |
| placement | `frontend/src/features/hospitalization/components/HospitalizationTabbedView.tsx:47-51` | mobile Header配置 | 隣接testなし |

対象外hit:

- `HospitalizationDetailActions.tsx:33-46`、`use-hospitalization-detail.ts:19-29`、`CarePlanTab.tsx:20-42`、`DailyRecordsTab.tsx:31-58`は死亡時のmutation拒否であり表示tokenを決めない。
- `HospitalizationForm.tsx:101-121`と`use-hospitalization-form.ts:69`は死亡petの入院登録guard、`use-hospitalization-list.ts:34-43`はboard drag guardであり表示tokenを決めない。
- `use-hospitalization-form-model.ts:151`はbackend statusを日本語form値へ変換する入力mappingであり、表示class/tokenを決めない。
- `PatientContextHeader.tsx:73,100,107,130,134-140`は類似tokenを持つがhospitalization callerがなく、medical-records surfaceなのでU10外。
- Expanded/Tabbedのchild propはmutation guardへの中継で、token非決定。

F16裁定後の変更集合:

- 常時候補: ListView＋test、Board＋test。
- shared contractを変える裁定: `PatientInfoCard`＋test、Header bridge regression test。
- hospitalization固有表現の裁定: `HospitalizationPatientHeader`＋test。shared API追加を選ぶ場合は`PatientInfoCard`も影響。
- verification-only: Detail、Expanded、Tabbedのdesktop/mobile両variant。

着手ブロッカーはF16の4裁定項目だけで、対象探索は解消済み。要件責任者=曽我の裁定後、frontend test ownerが裁定対象の死亡fixtureをRED化し、単一frontend writerへ渡す。

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
- 隣接する既存概念: `backend/internal/lstep/lstep_delivery_trigger_methods.go:168-179` の `TriggerCheckupFollowUp` が `model.TriggerTypeCheckupFollowUp` で飼主へLステップ配信を発火し、`backend/internal/medicalrecord/checkup_service.go:240-254` により健診作成時に**無条件で**（臨床判断によらず）呼ばれる。これは配信triggerであってrecord fieldでも臨床statusでもないため、上記のfield不在結論は変わらない。A/B比較時は同じ2条件を入力にする。(1) 現行の「follow-up」は配信trigger名であり、sentinel名との衝突有無を確認する。(2) 現行配信が対象にする業務と、sentinel案が対象にする業務の同一・差分を記述する。
- Decision request: 要件責任者=曽我が選択肢A/Bのいずれかを明示する。Aの場合は現行contractの維持範囲、Bの場合は下記7契約と配信triggerとの関係を記録し、対応する別unitへ渡す。本unitはfieldを追加しない。

現状の追加実測:

- `/checkups`は`frontend/src/features/checkups/routes/CheckupsList.tsx:296-305`で次回予定日と自由記述resultを表示するだけ。
- `TriggerCheckupFollowUp`のdependency interfaceは`backend/internal/medicalrecord/service_deps.go:86-90`で`clinicID, ownerID`だけを受け、checkup IDやclinical stateを持たない。
- 現行triggerは非同期・非致命的で、sentinel stateとのatomicity contractは無い。

| 選択肢 | 変更対象 | 影響route | 既存test |
|---|---|---|---|
| A: sentinelを追加しない | production 0 file、test 0 file。期限、自由記述result、無条件配信を維持 | 変更0。`/checkups`、`/medical-records/:id`は現行表示 | 変更0 |
| B: 曽我が目的と意味を定義した後にsentinelを追加 | 下記backend/frontend/OpenAPI/codegen surface。配信連動を定義する場合だけLステップ3 fileも追加影響 | frontend `/checkups`、`/checkups/new`、`/medical-records/:id`。HTTP `GET /api/v1/checkups`、medical-records配下checkup GET/POST/PATCH | 下表12 file。sentinel専用testは現状0 |

選択肢Bで曽我が埋める7契約:

1. 業務目的と読む人。
2. boolean／状態enum。
3. set/clear主体とclear条件。
4. 一覧表示だけか、server-side filter/sortを持つ作業queueか。
5. create入力か、健診結果からの導出か。
6. `TriggerCheckupFollowUp`と独立か、配信条件／配信済み状態を表すか。
7. clinical state変更のaudit要否。

選択肢Bの実測影響面:

- schema: `backend/migrations/001_init.sql:1483-1496`を直接編集せず、裁定時点の最大番号に続くappend-only migration。
- model/codegen: `backend/internal/model/checkup_record.go:9-28`、`backend/tygo.yaml:1-7`、`frontend/src/types/generated/models.ts:612-630`。
- request/response: `backend/internal/medicalrecord/checkup_request.go:10-110`、`checkup_response.go:11-111`。
- service/repository: `checkup_service.go:15-71,168-175,183-217,237-254,261-305`、filterを持つ場合は`checkup_repository.go:20-25,54-90`。
- handler/route: method登録は不変で、DTO carrierの影響確認先は`checkup_handler.go:29-117,143-168`、`routes.go:214-230`。
- OpenAPI: `backend/docs/api.yaml:4340-4434,8000-8035,19009-19061,19093-19186`。
- global frontend: `frontend/src/features/checkups/api/types.ts:1-25`、`transforms.ts:3-20`、filter時は`get-checkups.ts:28-50`、`CheckupsList.tsx:35-312`。
- create frontend: `create-checkup-medical-record.ts:17-34`、`use-checkup-form.ts:22-50,137-143`、`CheckupForm.tsx:92-178`。
- medical-record frontend: `frontend/src/features/medical-records/api/checkups.ts:12-56`、`CheckupsTab.tsx:55-82,125-158`、`checkups-tab-table-model.ts:3-17`、`CheckupsTabRows.tsx:27-273`、`CheckupsTabTable.tsx:14-25,81-114`。
- 配信連動時だけ: `service_deps.go:86-90`、`checkup_service.go:237-254`、`lstep_delivery_trigger_methods.go:167-179`。

| 既存test file | 現行件数 | 直接影響 |
|---|---:|---|
| `backend/internal/model/schema_drift_test.go` | 2 | schema drift 1 |
| `backend/internal/medicalrecord/checkup_request_test.go` | 8 | create/update mapping 2 |
| `backend/internal/medicalrecord/checkup_response_test.go` | 6 | normal/global response 2 |
| `backend/internal/medicalrecord/checkup_handler_test.go` | 5 | list/create/update/global 4 |
| `backend/internal/medicalrecord/checkup_service_test.go` | 28 | list/create/update/build updateの5以上。配信連動時はtrigger caseも影響 |
| `backend/internal/medicalrecord/checkup_repository_test.go` | 9 | read/create/update 6。filter契約時はglobal filterも影響 |
| `backend/internal/lstep/lstep_delivery_trigger_service_test.go` | 18 | 配信連動時だけ1 top-level / 4 subtest |
| `frontend/src/features/checkups/api/transforms.test.ts` | 10 | sentinel mapping 0 |
| `frontend/src/features/checkups/hooks/use-checkup-form.test.ts` | 10 | 成功payload 1 |
| `frontend/src/features/checkups/routes/CheckupForm.test.tsx` | 3 | control追加時3 |
| `frontend/src/features/checkups/routes/CheckupsList.test.tsx` | 17 | sentinel表示0 |
| `frontend/src/features/medical-records/components/CheckupsTab/CheckupsTab.test.tsx` | 7 | create/update payload、sentinel専用0 |

着手ブロッカーは7契約の未定義。要件責任者=曽我が選択肢A/Bを明示し、Bの場合は7契約を埋める。その後、schema/API/type/UI/testの各ownerが上記surfaceを分担する。

### 要実測項目

#### M-01 OwnersList操作範囲

- Route: `/owners?include_deceased=true`。
- Fixture seed source: disposable local `003_demo`へ`backend/migrations/seeds/003_demo/owners.csv`と`pets.csv`を適用し、`docs/ops/testing/scenarios/S01-deceased-pet-guard.md`/`V03-owner-pet-staff-forms.md`の手順で同一ownerにalive、danger=`高`、deceasedの3頭を準備する。実測後はS01どおり生存へ戻す。
- Fixture実査: 全named fileは実在。owner `300588`（`owners.csv:593`）のpet `1001002`（`pets.csv:1005`）、`1001004`（`:1007`）、`1001005`（`:1008`）を同一ownerの3頭候補として固定できる。ただし`pets.csv`全15,654行で`danger_level=high`は0件、`status=deceased + deceased_at + deceased_reason`完備は0件（`deceased_reason`非空自体が0件）。
- 追加すべき具体値: disposable local上で`1001002={status:alive,danger_level:low,death fields:null}`を対照、`1001004={status:alive,danger_level:high,death fields:null}`、`1001005={status:deceased,danger_level:low,deceased_at:D 00:00:00+09,deceased_reason:"M01/M05 fixture"}`とする。CSVは手編集せずS01/V03のUI/API経由で作る。
- 着手ブロッカー: `1001004`のhigh化と`1001005`の死亡日時＋理由登録の2 mutationが未実施。
- 次の一手: fixture担当が2 mutationを準備し、実測担当と曽我または同席する仕様責任者がrunbookを実行し、完了後fixture担当が`1001005`をalive、`1001004`をlowへ復旧する。
- Persona: owners `view=true, edit=true, delete=true`の通常担当者。曽我または同席する仕様責任者が許可操作を記録する。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: include-deceasedを有効にし、3 rowそれぞれでowner名link、report、編集、削除、pet死亡登録/解除をpointer、Tab/Enter/Space、直接URLで試す。各操作前後のrequestをnetworkで記録する。
- Expected result: aliveは通常操作可、danger=`高`は「⚠ 危険」等の非色cueとaccessible nameを失わない。deceasedはF16裁定前の現行grayを記録し、曽我が許可/禁止する操作を操作単位で明示する。禁止とされた操作はmutation 0件。
- Required evidence artifacts: 4 viewport screenshot、accessibility tree、各操作のaccessible name、GET以外のnetwork HAR、操作可否の曽我裁定メモ。

#### M-02 Examinations一覧意味とlayout

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

#### M-03 RBAC非活性の理由/name

- Route: `/settings/permission-groups`、`/medical-records/:id`、`/hospitalization/:id/edit`、`/vaccinations/:id`、`/examinations/:id`。
- Fixture seed source: `backend/migrations/seeds/003_demo/permission_groups.csv`、`permission_group_rules.csv`、`staffs.csv`、`staff_permission_groups.csv`と各featureの既存record CSV。`docs/ops/testing/scenarios/V03-owner-pet-staff-forms.md`で試験用group/staffを作り、`V01-clinical-forms.md`の既存recordを使う。
- Fixture実査: 全named CSV/V01/V03は実在。group `9`（`permission_groups.csv:10`）とstaff `37`（`staff_permission_groups.csv:38`）はclinical 4 resourceでview-onlyだが、`master-permission`は`false/false/false/false`（`permission_group_rules.csv:177`）のため全対象routeを覆わない。create-onlyとedit-without-delete完全一致groupは0件。hospitalization/vaccinationは0行、medical recordは全件finalizedでmutable recordも不足。
- 追加すべき具体値: `master-permission`、`medical-records`、`hospitalization`、`vaccinations`、`examinations`へ共通で、`FE12-M03-VIEW={true,false,false,false}`、`FE12-M03-CREATE={true,true,false,false}`、`FE12-M03-EDIT={true,false,true,false}`の3 groupを作り、各専用staffへ割り当てる。操作対象は別group `FE12-M03-TARGET`とする。secretはseed/ledgerへ記録せず実行時に供給する。
- 着手ブロッカー: 3 persona、`FE12-M03-TARGET`、M-02/M-04/M-05由来のmutable recordが未作成。
- 次の一手: fixture担当がfeature fixture→target group→3 group→3 staff作成・再編集割当を行い、実測担当がpersonaごとに再ログイン/再読込してrunbookを実行する。
- Persona: (1) view-only=`view:true/create:false/edit:false/delete:false`、(2) create-only=`view:true/create:true/edit:false/delete:false`、(3) edit-without-delete=`view:true/create:false/edit:true/delete:false`。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: 各personaでrouteを再ログイン/再読込して開き、pointer、Tab/Enter/Space、formのprogrammatic submit、保存中の権限剥奪後callbackを試す。permission-groupは新規panel、既存panel、reorder、保存後rulesも個別確認する。
- Expected result: view accessは維持する。許可されたactionだけ実行でき、禁止controlはaccessible nameと理由を保持する。禁止personaからのPOST/PATCH/PUT/DELETEは0件で、same-commit剥奪後もmutationしない。
- Required evidence artifacts: persona×routeのaccessibility tree、4 viewport screenshot、network HAR、console log、action別permission matrixと0 mutation集計。

#### M-04 Hospitalization child control実効性

- Route: `/hospitalization`のboard/listと`/hospitalization/:id`。
- Fixture seed source: `backend/migrations/seeds/003_demo/hospitalizations.csv`、`daily_records.csv`、`vital_records.csv`、`care_logs.csv`、`care_plan_items.csv`、`cages.csv`を基に、`docs/ops/testing/scenarios/S05-hospitalization-cycle.md`のadmitted fixtureを準備する。
- Fixture実査: 全named CSV/S05は実在。hospitalizations/daily/vital/care log/care planは全てheader-only 0行。`cages.csv`は49行あり、clinic 1のactive dog cageはid `3`（small、`:4`）と`4`（medium、`:5`）。runbookのnote操作に必要な`staff_notes.csv`も0行。
- 追加すべき具体値: full-permission staff `1`でpet `1000018`、owner `300003`、`type=hospitalization,start=D,end=D+2,status=admitted,cage_id=3,doctor_id=1,memo="FE12-M04 admitted"`のHを作成する。Hの日付Dへdaily record DR、`09:00 JST/38.5℃/heart 120/respiration 30/8.0Kg/staff 1`のvital、`09:30/type=food/status=completed/value=完食/staff 1`のcare log、朝食と投薬のactive care plan、`10:00/content="FE12-M04 申し送り"/staff 1`のnoteを作る。
- 着手ブロッカー: admitted Hと全child/noteが未作成。cageだけは準備済み。
- 次の一手: fixture担当がH→DR→vital/care log/care plan/noteを作り、実測担当が通常担当者の対照を1回記録後、view-onlyを検証する。最後にadminが対象petを一時死亡登録し、表示と操作不能を確認後に生存へ戻す。
- Persona: hospitalization view-only=`view:true/create:false/edit:false/delete:false`。対照として通常担当者=`view/create/edit/delete:true`を1回使う。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: board drag/drop、check-in/status、退院（会計あり/なし）、daily、vital、care log、note、care planのcreate/edit/deleteをpointer/keyboard/programmatic callbackで試す。操作dialogを開いた後に権限を剥奪するcaseも含める。
- Expected result: view-onlyおよびsame-commit剥奪後は全child/top-level mutation 0件。死亡fixtureは死亡文言を表示し、drag/check-in等を実行しない。非活性controlのnameと理由は残る。
- Required evidence artifacts: 操作別network HARと0件集計、accessibility tree、4 viewport screenshot、console log、会計あり/なし退院の個別記録。

#### M-05 Clinical sentinel responsive

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
- Expected result: death/danger/HIGH/LOW/期限超過が非色cueを持ち、normalとtoday/futureを誤ってdangerにしない。死亡操作はpositive matchで拒否され、全viewportでcue/controlのwrap、clip、overlapなし。
- Required evidence artifacts: route×4 viewport screenshot、accessible name/text dump、computed token、console/network HAR、fixture-to-cue対応表。

#### line-reserve font実機確認

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

### 別unit裁定待ちのMEDIUM follow-up

**2026-07-26・曽我からの委任によりAIが代理決裁した。以下4件の選択は確定である。各節は裁定の根拠となった調査記録として残す。**

- **owner parent/pet → O-A（現行nested pets contractを使い一括送信）。** backendには既にowner+pet同一transaction作成が実装されており（`repository.go:212-248`、`http_request.go:158-211`）、frontendの`CreateOwnerRequest`がそれを使わず逐次作成しているだけである。**backend production変更ゼロで部分成功という失敗モード自体が消える**。O-B/O-Cは起きるべきでない状態を管理するUIを足す選択であり②違反。
- **permission-group → P-A（parent＋rulesを1 transaction）。** 同じ理屈。P-Bは部分成功stateをpanelに保持し、P-Cは補償削除を実装するが、いずれも不整合を前提にした機構を足す。backend変更は要るが、補償ロジックと監査を作るより小さい。
- **reception → R-A、ただし前提を1つ修正する。** カンバンのドラッグは高頻度操作であり楽観更新は③の観点で正当である。R-Bは全操作を遅くする。**ただし現行の実欠陥は`use-reception-modal-handlers.ts:183-193`がAPI完了前に成功toastを出していることであり、これはユーザーへの虚偽である。** 楽観的なカード位置は許容し、楽観的な成功メッセージは許容しない。R-Aの操作単位rollbackと失敗後再同期を契約化したうえで、toastは成功後のみに直す。
- **medical auto-create → M-A（明示retry actionで失敗phaseから再開）。** M-Bの自動retryは⑤違反（検証されていないプロセスの自動化）であり予約重複を高速に量産し得る。M-Cは作成済み予約を捨てる。**M-Aが安全に成立するのはbackendが既に同一appointmentの既存recordを返すからである**（`medical_record_crud.go:276-293`）。冪等性の土台がある上での再開なので重複を作らない。

以下は裁定の根拠となった現状・選択肢・影響の調査記録である。

#### reception optimistic UI rollback

- 現状: `/`の`frontend/src/features/reception/hooks/use-reception-kanban.ts:230-251`は`opts.rollback`時に`apiColumnDataRef.current`全体へ戻す。drag/advance/cancelはrollbackあり（`:281-299,329-337,345-357`）、terminal「会計済」除外はrollbackなし（`:319-325`）。`update-appointment-status.ts:24-31`は成功時だけreception queryをinvalidateし、`use-reception-modal-handlers.ts:183-193`はcancel API完了前に成功toastを出す。
- 既存test: `use-reception-kanban.test.ts` 23件（drag失敗rollback 1件、`:218-232`）、`use-reception-modal-handlers.test.ts` 9件。advance/cancel/terminal/並行失敗、失敗後refetch、cancel失敗toast契約は0件。
- R-A: optimistic表示を維持し、操作単位snapshotへのrollbackと失敗後query再同期を契約化する。変更面=`use-reception-kanban.ts`、`update-appointment-status.ts`、route `/`、既存23件testへfailure matrix追加。
- R-B: status mutation成功後に移動・除外・成功toastを反映する。変更面=`use-reception-kanban.ts`、`use-reception-modal-handlers.ts`、route `/`、既存23件/9件testの完了時点を変更。
- 着手ブロッカー: 即時移動の業務要否、失敗時のカード位置、toast時点、再試行導線、query再同期表示が未裁定。
- 次の一手: 要件責任者=曽我がR-A/R-Bと上記UI契約を明示し、frontend implementation unitがfailure matrixをRED化する。

#### permission-group parent/rules partial success

- 現状: `/settings/permission-groups`の`PermissionGroupSettings.tsx:101-133`はparent create/update後にrulesを別`mutateAsync`で保存する。APIも`permission-groups.ts:76-104`で分離され、backendも`http_permission.go:278-302`でPOST/PATCH parentとPUT rulesが別route。rules失敗は`use-master-save.ts:97-106,113-122`が通知しpanelを保持する一方、`PermissionGroupSidePanel.tsx:56-64`は非同期結果前にdirty=falseへする。
- 既存test: `PermissionGroupSettings.test.tsx` 4件、`use-master-save.test.ts` 14件（post-save reject一般契約1件）。backendは`http_permission_handlers_test.go` 8、`permission_group_service_test.go` 14、`permission_group_repository_test.go` 11 top-level test。parent成功/rules失敗後のdirty、再送payload、重複、成功表示の一体testは0件。
- P-A: parent＋rulesをcombined writeとして1 transactionへ入れる。変更面=frontend settings/API/save hook、backend handler/service/repository、画面route不変、API contractと上記test群。
- P-B: parent成功を確定状態として残し、rulesだけを再送するpartial-success stateをpanelへ保持する。変更面=Settings/SidePanel/save hook、現行API route、rules-only再送・dirty・close・権限剥奪test。
- P-C: rules失敗時にcreate parent削除、update parent/rules復元の補償を行う。変更面=frontend/API/backendの補償、監査、競合、補償失敗test。
- 着手ブロッカー: parent-only状態の許容、partial表示、dirty、再送対象、成功通知、create/update共通化が未裁定。
- 次の一手: 要件責任者=曽我がP-A/P-B/P-CとUI契約を明示し、implementation unitが整合性・監査・権限剥奪・再送matrixをRED化する。

#### owner parent/pet partial success

- 現状: `/owners/new`の`use-owner-form.ts:227-278`はownerを先に作り、pending petsを`owner-form-followups.ts:1-9`の`Promise.allSettled`で個別作成し、pet失敗があってもwarning＋owner成功を返す。`OwnerForm.tsx:113-124`は成功stateで`/owners/:id`へ遷移する。一方、backend `POST /v1/owners`はnested petsを受け（`http_request.go:158-211`）、`http_owner.go:55-91`→`service_core.go:65-99`→`repository.go:212-248`でowner+petsを同一transaction作成する。frontend `CreateOwnerRequest`（`frontend/src/types/owner.ts:51-72`）はpetsを持たず現行atomic contractを使っていない。
- 既存test: `use-owner-form.test.ts` 5、`owner-form-followups.test.ts` 1、`use-pet-form-list-state.test.ts` 8件。backendは`http_owner_test.go` 4、`service_test.go` 7、`repository_create_with_pets_write_owner_test.go` 7 top-level test。frontend partial pet failureは0件。
- O-A: frontendが現行nested pets contractを使い一括送信する。変更面=`types/owner.ts`、`use-owner-form.ts`、`create-owner.ts`、route `/owners/new`、frontend 5/1/8件とbackend 4/7/7件test。backend production追加は0。
- O-B: partial成功を残し、失敗pet入力と結果をdetailへ渡し、失敗分だけ再送する。変更面=form hook/followups/route、`/owners/new`→`/owners/:id`、成功pet非再送・失敗pet再送・権限剥奪・二重送信test。
- O-C: frontend逐次処理を残し、pet失敗時に作成済みpetとownerを補償削除する。変更面=delete順、補償失敗表示、監査、pet途中失敗・owner削除失敗・権限変化test。
- 解消済み前提: backendにatomic owner+pet writeが無いという技術ブロッカーは解消済み。現行実装の証跡は上記3 backend path。
- 着手ブロッカー: 1頭失敗時に事実を残すか全体失敗か、失敗入力保持、遷移先、再試行単位、通知が未裁定。
- 次の一手: 要件責任者=曽我がO-A/O-B/O-CとUI契約を明示し、frontend implementation unitがrequest型・form state・route・testを変更する。

#### medical auto-create拒否後retry UX

- 現状: `/medical-records/new?petId=...`の`use-medical-record-auto-create.ts:95-145`は開始時にref=true、予約作成後にカルテ作成し、catch（`:146-149`）はerror通知＋ref resetだけを行う。ref resetはeffect dependency変更ではなく、retry callback/error stateを返さない。`MedicalRecordFormActions.tsx:96-102`は作成中だけ保存buttonをdisabledにし、`use-medical-record-save-action.ts:98-108`はrecordIdなしの保存を失敗で返す。backend `medical_record_crud.go:84-92,120-121,276-293`は同一appointmentの既存recordを返す。
- 既存test: `use-medical-record-form.auto-create.test.ts` 21件（auto-create 12、failure 1件`573-595`）、`MedicalRecordFormActions.test.tsx` 11件、backend `medical_record_service_test.go` 27 top-level test。2回目呼出し、利用者のretry導線は0件。
- M-A: 明示retry actionを表示し、失敗phaseと作成済みappointment IDをstateへ保持してそのphaseから再開する。変更面=auto-create hook/form hook/form/actions、同route、21/11件testへ予約失敗・カルテ失敗・retry・二重click・権限剥奪追加。
- M-B: error種別、回数、間隔を定めた自動retryを行い、上限後に手動回復actionを表示する。変更面=auto-create hook/form hook、同route、transient/permanent・上限・cancel/unmount・appointment再利用test。
- M-C: 失敗時にnew routeを終了してpet選択または受付へ戻し、次の作成を新しい明示操作として扱う。変更面=hook/form route/navigation state、`/medical-records/new`・`/medical-records/select-pet`・`/`、戻り先・appointment再利用・二重record test。
- 着手ブロッカー: 同画面／別route、手動／自動／再入場、予約成功後appointmentの残置／再利用／取消、通知と上限が未裁定。
- 次の一手: 要件責任者=曽我がM-A/M-B/M-Cと上記状態を明示し、implementation unitがfailure phaseを型で表現してduplicate preventionと権限再確認testをRED化する。

### 未裁定の残余risk

下記は2026-07-26時点のcurrent sourceで未解消を確認したfollow-up候補である。本ledgerは着手・起票・優先度裁定をしない。

- **alive petへ予約編集した直後のReception danger一時旧値** — 最初に開く: `frontend/src/features/reception/hooks/use-reception-modal-handlers.ts:135-168`、`use-reception-kanban.ts:362-369`、`frontend/src/hooks/use-update-reservation.ts:43-58`。確認: 成功時local mergeの`updatedAppointment`が新petのdanger/statusを持つか、API responseとrefetchのどれを正本にするか、失敗時rollback範囲。判断者: frontend reception contract owner。手順: stale値を再現するhook testを作り、API response採用／payload補完のcontract確定後にlocal mergeとquery同期を変更する。
- **backend死亡登録APIの既死亡再登録** — 最初に開く: `backend/internal/lstep/lstep_lifecycle_handler.go:12-38`、`lstep_lifecycle_service.go:90-128`、`backend/internal/pet/repository.go:291-296`。確認: serviceは既存petを読むが既死亡をrejectせず、repositoryは同じ死亡fieldを再更新する。判断者: clinical API ownerと要件責任者=曽我。手順: idempotent success／Conflict、audit重複、Lステップ副作用を定義し、service/handler testで既死亡caseを固定してから実装する。
- **medical-record auto-createのJST日付境界** — 最初に開く: `frontend/src/features/medical-records/hooks/use-medical-record-auto-create.ts:95-152`と`use-medical-record-form.auto-create.test.ts`、backendの`medical_record_auto_create.go`と同test。確認: `formatJSTDate(new Date())`（hook`:135`）、当日予約lookup、test clockが同じJST日付を使うか。判断者: medical-record contract owner。手順: clock seamとUTC→JST境界caseを定義し、固定時刻testを追加してからlookup/date生成を同じclockへ接続する。
- **PermissionGroup以外のmaster permission未配線call site** — 最初に開く: `frontend/src/features/master/hooks/use-master-crud.ts:162`、`use-master-save.ts`と全caller。確認: 現状列挙したCRUD 19、save 24 call siteごとにroute resource、create/edit/delete action、mutation直前のpositive permission checkが対応するか。判断者: frontend RBAC owner。手順: `rg`で両hookのproduction callerを再列挙し、call site×resource×action表を作り、未配線だけを別実装unitへ渡す。
- **tygo pointer mapping 15行** — 最初に開く: `backend/tygo.yaml:17-35,46-64,75-93`と3 generated output。確認: `*uint64`、`*string`、`*bool`、`*time.Time`、`*float64`の5 mapping×3 packageが生成物diffへ寄与するか。判断者: backend/frontend type contract owner。手順: 許可された`make codegen`で各mappingの出力寄与を個別記録し、寄与0の行だけを設定整理unitへ渡す。
- **line-reserve CompletePageの4文字超malformed時刻** — 最初に開く: `frontend/line-reserve/src/pages/CompletePage.tsx:4,19-52`と`frontend/src/shared-liff/jst-date.ts`の`formatTimeHHMM`。確認: malformed入力の表示contractと他callerへの共有formatter影響、隣接testの不存在。判断者: line-reserve contract owner。手順: 4文字超、空、colon有無の期待表示を決め、CompletePage隣接regression testを作ってからformatterまたはcallerを変更する。
- **ConfirmPageの未assert 3契約** — 最初に開く: `frontend/line-reserve/src/pages/ConfirmPage.tsx:37-68,90-130,205-215`と`ConfirmPage.test.tsx:46-118`。確認: `isPending`中の二重submit、reservation成功後のLIFF送信順、409時にinline `role=alert`が出ないこと。判断者: line-reserve test owner。手順: 3観測点を独立testにし、現行behaviorとの差が出た場合だけcontract ownerへ裁定を戻す。
- **manual chunkの未再計測** — 最初に開く: `frontend/src/features/manual/lib/manual-index.ts:50-61,86-87`と`frontend/src/app/routes/operations-routes.tsx:132-150`。確認: 2つのMarkdown globが`eager:true/?raw`でmanual chunkへ入る現在byte、500 kB警告、route lazy境界。判断者: frontend verification gate owner（実行許可）→frontend performance owner（計測）。手順: gate ownerが他sessionのfrontend WIP静止とfull build許可をtask logへ明記し、performance ownerへ引き渡す。引渡し後だけperformance ownerが`docker compose exec frontend pnpm build`を実行し、command/exit、`dist/assets`のmanual chunk名/byte、500 kB警告、build前後statusをartifact化する。警告該当時は比較案を作り、build未実行のまま解消扱いにしない。
- **AuthProviderのfeature barrel経由eager import** — 最初に開く: `frontend/src/app/router.tsx:1-23`、`frontend/src/features/auth/index.ts:1-10`、`hooks/use-auth.tsx:1-19`。確認: routerの同期`@/features/auth` importがLogin/Forgot/Reset等のbarrel exportを同じgraphへ含めるか、public boundaryを維持するentry設計、auth route chunk。判断者: frontend architecture owner。手順: current import graphとchunkを記録し、barrel分割／provider専用public entry／現状維持の各contractを比較してから変更surfaceを確定する。

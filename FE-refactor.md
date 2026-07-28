# FE-refactor — FE12-02 active-only ledger

> 更新: 2026-07-28（要件責任者: 曽我）
> 業務目的: 未完了の臨床sentinel・RBAC安全境界を、決裁と実測の証跡が揃うまで追跡する。
> 本ファイルは使い捨てのactive ledger。恒久規約は `DESIGN.md`、`docs/spec/design-system.md`、`docs/spec/ui-design-compliance.md`、`frontend/CLAUDE.md` を正本とする。
> 解消済みの証跡: F16/U10/O-A実装=`3b7524748`、R-A/M-A=`e7c978ec9`、P-A=`99bac632e`、exam parent+items原子化=`24929e83d`、死亡API 409 fail-closed=`45b681866`、test-hardening 3件=`3c993420a`。代理決裁6件の記録と調査パックは`a500d424c`および本更新直前までの履歴に保存する。
> S1/S2並行レーン（2026-07-27完走・全項目解消）の証跡: pet死亡CAS封鎖=`18d307076`、カルテlookup JST正規化=`305b50c7f`、ワクチン期限JST契約=`56d18eb18`、auto-create原子化=`5e5868549`、DBOrTx inventory債務=`0eecddb11`、reception danger sentinel=`082be9961`、master permission 45/45配線+fail-closed化=`da550a84d`、auth独立chunk化+checkup resource整合=`8a00a4794`、StaffSettings test追随=`e8a4db982`。各項目の分析・裁定・証跡の全文は本節削除直前の版（commit `0fd1f7d18`）に保存する。裁定の要点2件は後段「維持する裁定」に残す。
> 削除した `## FE12-02 unit execution record` と過去のFE/R-F履歴はCommit `657c1a49cd2c37dc63f5af8e530258a36a12d81e` に保存する。
> **DB reset経路の復旧（2026-07-27・fixture作成の前提工程）**: M-01〜M-05のfixture作成は5走連続でBLOCKEDし、真因は「7/17のmigration統合以降ずっとDB resetが完遂できない状態だった」ことだった。seed CSVがテーブル定義に追随しておらず`COPY`が22P04で落ちていた（`COPY`は列リストを渡さずテーブル定義順に依存する）。証跡: pets/billing_items CSV是正=`0b51c891c`、clinical_plans/exam_results CSV是正=`a4946053b`、migration統合第3回（002-004→001）+clinical_plans列順是正=`edbb162f7`、ERD v31.30（pet_owners/複合FK/テーブル数110）+波及4文書=`2c13b563c`、fixture完成の台帳反映=`976d717c7`。**この破壊はDB resetを実行した瞬間にしか顕在化しないため、7/17から誰も気づいていなかった。** 再発防止のgate2本は **2026-07-28 に実装完了**（`TASK-446`/`TASK-447` として起票 → `backend/internal/lintscan/seed_csv_schema_drift_test.go` と `erd_table_count_drift_test.go` を新設 → 台帳から両sectionを削除済み）。Gate Aはseed CSVヘッダを「`001_init.sql` + 番号順の後続migration」から導いた最終列順と照合し、扱えない列順変更DDLに遭遇したら黙殺せず失敗する。Gate Bは`CREATE TABLE`の実テーブル数（重複定義とコメントを除いた110）とERD宣言値を照合する。両gateとも壊れた入力で確実に失敗することを実証するテストを同梱し、`go test ./...`経由でCIに載る。**コミット済み=`43419cc12`**（`docker-compose.yml`のdocs read-only mount追加を含む）。

## 現在地（2026-07-28 終業時点・ここだけ読めば残件が分かる）

**FE12 実測レーンは実質終結した。** M-03 / M-01-E / M-04 / M-02 / R-4 が COMPLETE、R-1 / R-3 / R-2測定 も完了。**fixture は R-4 で全て撤去済み**（例外: hospitalization `1` が日次記録制約で残存）。臨床状態は実測前の値へ復旧済み（pet `1001005`=alive、`1001004`=low、`1001002`/`1000018`=無傷）。

**残件は5つだけである。**

| # | 残件 | 誰が | 着手前提 |
|---|---|---|---|
| 1 | **M-01-D 裁定** — 死亡ペットの編集/削除をブロックするか | **曽我** | 材料完備。下記「M-01-D 裁定材料」を読めば答えられる |
| 2 | **M-05 残2件** — danger高cueのruntime証明、期限4区分のdanger誤検出有無 | エージェント | **fixture再作成が前提**（R-4で撤去済み）。手順は M-05 節。着手プラン: `3-session-agent.html#TASK-461` |
| 3 | **R-2 実行** — `backend/tygo.yaml` の15行削除＋`make codegen`差分0確認 | **USER専権**（`make codegen`） | 測定完了。削除対象は確定済み。着手プラン: `3-session-agent.html#TASK-462` |
| 4 | **line-reserve** — font実機確認 | QA/端末管理者 | 実機3台とQA環境の受け渡し |
| 5 | **BUG-455〜458 の修正実装** | エージェント | 起票済み。`3-session-agent.html#ledger` が正本。本ledgerの所掌外 |

**この実測が生んだ実装findings 4件は全て起票済みである**（`BUG-455` CRITICAL / `BUG-456` HIGH / `BUG-457` HIGH / `BUG-458` MEDIUM）。以降それらの追跡は `3-session-agent.html#ledger` を正本とする。

### M-01-D 裁定材料（M-01-E の実測結果・これで答えられる）

`/owners?include_deceased=true` の行アクションメニューを全権限 persona で観測した結果、**死亡行と生存行で提示される操作に差分が無い**。

| 操作 | 死亡行（pet `1001005`） | 生存行（pet `1001002`） | disabled |
|---|---|---|---|
| 編集 | 提示あり | 提示あり | **なし** |
| レポート | 提示あり | 提示あり | **なし** |
| 削除 | 提示あり | 提示あり | **なし** |

`docs/spec/specification.md:21` は「死亡ペットに対する誤操作の物理的ブロック」を規定するが、**一覧の行アクションはこれを満たしていない。** 製品哲学により確認ダイアログは選択肢に入らない（ロック・Undo・物理ブロックのいずれか）。

**曽我が決めるのは2点だけ**: (a) 死亡ペットの**編集**をブロックするか。するなら誤記訂正の正規経路をどうするか（追記のみ許す／管理者権限でのみ許す／死亡解除→訂正→再登録を強制する）。(b) 死亡ペットの**削除**をブロックするか。

なお F16（死亡バッジのグレー）は裁定済みで、実装（badge のみ）が合格である。死亡登録/解除は `45b681866` 以降 fail-closed で確定済み。**一覧に死亡登録/解除の導線は存在しない**（`PetCareSection.tsx` = 飼主詳細にある）。

## Active scope and authority

- 追跡対象は `M-01`〜`M-05` の実測、line-reserve font実機確認、残余2件（R-2/R-3）、実測後のfixture復旧（R-4）とする。R-1・F16・F9・U10・MEDIUM 4件およびS1/S2並行レーン全9項目は裁定・実装済みで追跡を終了した。
- **2026-07-28 時点の到達点**: R-1は全5件完了。**PO裁定を待って着手できないのは M-02（R-3経由）・M-01-D の2件**。実行順と干渉回避は「実測レーン分割」節を正本とする。
- **2026-07-28 訂正 — 「fixture は全て作成済み」は誤り。API作成物の過半が消滅している**（下記「fixture生存実測」節が正本）。7/27に作成した fixture のうち生存するのは S1 の pets 4頭と H=`1` だけで、これも 7/28 01:45 に再作成されたものである。M-02・M-05・M-03 の fixture は**再作成が着手前提**として復活した。したがって「M-03・M-04・M-01-E は曽我の判断を要さず着手できる」は現時点で成立しない。
- **2026-07-28 更新 — R-3 の前提が変わった**: #249 U4（`afd8404a4` API＋`1adf55b6e` UI）で `exam_reference_ranges` への投入経路が開通し、R-3 の選択肢 (A) は「rangeのAPIが無いため実装が要る」という制約から解放された。**ただし `exam_reference_ranges` は依然0行**であり、獣医師が動物種ごとの臨床値を投入するまで HIGH/LOW cue は導出されない。したがって **M-02 を「基準値不在のまま実測する」(B) の妥当性は変わらない**。判断が要るのは「臨床値の投入を待つか、待たずに実測するか」だけになった（機構の有無は論点から消えた）。
- **2026-07-28 更新 — 未評価/正常の誤読対策は3 surface全てで完了**: `ExamItemsTable`（`19bbdbae2`）・`ExamPivotTable`・`ExaminationGroup` の全てで `isAssessed === false` を「未判定」として区別表示する。あわせて `BUG-450`（`GET /v1/examinations` のDTOに明細が無く、カルテ検査結果一覧が常に0件・飼主レポートの異常件数が常に0だった silent contract break）を `include_items` の opt-in 追加で解消した。**M-02 の実測時、カルテ側の検査結果一覧は初めて実データを描画する**（従来は空だったため、過去の実測結果があれば無効）。
- **2026-07-28 更新 — S1レーン（M-01-E + M-04）を試行し BLOCKED。「着手可」は環境前提が揃っている場合の話だった**: 実行記録は M-04 節末尾の「2026-07-28 S1レーン実測結果」を正本とする。臨床状態は一切動いておらず、fixture の生存も未確認のままである。**この試行で判明した4件のうち3件は runbook 側ではなく着手プロンプト側の欠陥**であり、次に同じ轍を踏まないよう下記「S1レーン試行の結果」へ分離して記録する。M-03（S2レーン）は未試行だが、ブラウザ経路の前提は共有するため同じ理由で止まる可能性が高い。
- **2026-07-28 レーン分割の訂正 — 並行化の軸は「データ干渉」ではなく「ブラウザ」である**: 後段の「実測レーン分割」節は pet ID / route の重なりで並行可否を判定しているが、実際のボトルネックは **CDP endpoint が `127.0.0.1:9222` の1つしかないこと**である（`.codex/config.toml` の `mcp_servers.chrome-devtools` も 9222 決め打ち）。1プロファイル＝1ログインセッションのため、persona 切替のたびに直列化する。**「データ的に並行可」でも実際には順番待ちになる。** 効果があるのは browser lane から非ブラウザ作業を剥がすことであり、ブラウザ同士を並べても効かない。実際のレーン構成は次のとおり。
  - **レーン1（ブラウザ・直列）**: M-03 closeout → M-01-E → M-04 → M-02 → M-05 → R-4
  - **レーン2（API のみ・ブラウザ不要）**: fixture 再作成。**2026-07-28 完了**（上記「レーン2」節）
  - **レーン3（環境不要）**: R-2 測定・GORM欠陥トリアージ。**2026-07-28 完了**（下記「残余risk」節）
  - **レーン4（人手）**: line-reserve 実機3台。エージェントレーン外
- **2026-07-28 確定実行順（曽我裁定4件を反映）**: 判断待ちは M-01-D のみになった。レーン1は**直列**とし、次の順で進める。
  1. **Chrome 3点gate** — `9222` 起動 → `list_pages` 成功 → **アプリ画面のスクショを1枚実取得**。3点目まで green にならない限り以降へ進まない。
  2. **M-03（S2）** — **2026-07-28 に実質完遂**。証跡パイプライン（a11y 15 / screenshot 60 / network 15 / matrix 未解決0）が全て揃い、9 gate 中 8 が PASS。残るのは「view access 維持」の観測対象を `/hospitalization/:id/edit`（edit gate 付き）から `/hospitalization/:id`（action guard 無し）へ差し替える closeout のみ。**edit route で VIEW/CREATE が拒否されたのは正しい RBAC 挙動であり defect ではない**（`frontend/src/app/routes/clinical-care-routes.tsx:117-141`）。
  3. **S1（M-01-E → M-04）** — M-01-E の合否基準は「死亡行のバッジがグレー」に確定済み。M-04 の権限剥奪caseは M-03 で使った group を借用する。
  4. **M-02** — 基準値不在のまま実測（R-3=B確定）。fixture 再作成済み。
  5. **M-05** — 直列尾部。fixture 再作成済み。**期限4区分は実測日 7/28 基準で作ってある。7/29 以降へずれる場合は `next_date` を切り直すこと。**
  6. **R-4** — fixture復旧。
  - **fixtureは消滅の前科があるため、実測着手前に必ず read-back で生存を確認する。日をまたぐ設計は採らない。**
  - **プロンプト側の欠陥が3連続で実測を止めた**（①証跡保存先が MCP の workspace roots 外 ②許可action と禁止action を区別せず矛盾指示 ③view 維持の観測対象に edit-gated route を指定）。**route を実測対象に指定するときは、その route の guard の `action` と persona の権限を突き合わせてから書くこと。** M-01-E 以降でも同じ構造で再発する。
  - **証跡が1枚も取れていない段階で prompt-craft / 外部agent へ委譲しない。** 3走・証跡0の原因はプロンプト精度ではなく、環境未検証のまま委譲した構造にある。
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
| M-03 | — | RBAC非活性の理由/name — **2026-07-28 COMPLETE**（4走） | 完了。fixtureはR-4で撤去済み | a11y 18・screenshot 76・network 15・matrix未解決0。`tmp/fe12-m03-evidence/2026-07-28/` |
| M-01-E | — | OwnersList操作範囲の証跡収集 — **2026-07-28 COMPLETE** | 完了。**死亡行に物理ブロックが無いことを確定** | a11y 4・screenshot 4・network 1。`tmp/fe12-m01e-evidence/2026-07-28/` |
| M-04 | — | Hospitalization child control実効性 — **2026-07-28 COMPLETE** | 完了。死亡遷移1回・確実に復帰。H=`1`は日次記録制約でR-4撤去不可（残存） | a11y 6・screenshot 5・network 5・control matrix 33行。`tmp/fe12-m04-evidence/2026-07-28/` |
| M-02 | — | Examinations一覧意味とlayout — **2026-07-28 COMPLETE** | 完了。**実欠陥1件を検出→`BUG-456`起票** | screenshot 8・a11y 2・network。`tmp/fe12-m02-evidence/2026-07-28/` |
| M-05 | P0 | Clinical sentinel responsive — **2026-07-28 部分完了 / 未確定2件** | **残るのは (a) danger高cueのruntime未証明（絞り込み無しでpet `1001004`が可視ページに載らなかった疑い）、(b) 期限4区分（past/today/future/empty）のdanger誤検出有無。** fixtureはR-4で撤去済みのため**再作成が着手前提** | 28 PNG・a11y 7・network 7は取得済み。残2件の逐語証跡が要る |
| M-01-D | P0 | 死亡ペットに対する操作可否の**裁定** | **曽我**。M-01-Eで材料は完備した（下記「M-01-D 裁定材料」） | 死亡ペットに対する許可/禁止を操作単位で明示した裁定メモ |
| R-3 | — | M-02の実測方針 — **2026-07-28 裁定済み=(B) 基準値不在のまま実測** | 完了。恒久是正は`3-session-agent.html#BUG-449` | M-02がこの方針で完走した |
| line-reserve | P1 | font実機確認 | QA環境管理者と端末管理者の受け渡し（実機3台・remote inspection）。**曽我の判断は不要** | 3実機のscreenshot・HAR・computed font-family |
| R-4 | P1 | 実測後のfixture復旧 — **2026-07-28 完了**（`1001005` alive、`1001004` low、staff `38-42` / group `10-13`撤去） | M-01〜M-05のcleanup対象read-back完了（M-05自体の判定はBLOCKED）。P1はroute不在・日次記録制約のため一部撤去不可 | `tmp/fe12-r4-evidence/2026-07-28/{before,after,cleanup-report,completion-report}.md` |
| R-1 | P2 | staff入力検証の契約drift 5件 — **2026-07-28 完了** | ①②と③frontend側は`166e4acd7`、③OpenAPI側と④は本ユニット、⑤は`a44fa0ebe`で完了。判断ブロッカー・残件なし | backend binding・OpenAPI required/description・frontend validation/test名・mutation直前の権限再検査が一致 |
| R-2 | P2 | tygo pointer mapping 15行の寄与測定 — **2026-07-28 測定完了。結論=15行すべて寄与0** | 測定は生成物の実読で完了（`make codegen` 不要だった）。**残るのは15行削除の実行のみで、その検証に `make codegen` が要る＝USER専権** | 下記「R-2 測定結果」節。削除後に `make codegen` 差分0であることの確認が残件 |
<!-- FE12-TASK-TABLE-END -->

#### 2026-07-28 R-4 実施結果

- status: **COMPLETE**。P0-aを最優先で実施し、pet `1001005` は `deceased` → `alive` / `deceased_at=null`、pet `1001004` は `danger_level=high` → `low` / `danger_reason`空へ復旧した。解除APIの初回・再試行はHTTP 409（初回後のread-backで既に復旧済み）だったが、最終状態は受入条件を満たす。
- P0-b: staff `38`-`42` はFE12識別子をGETで確認後、全件 `DELETE 204` → `GET 404`。permission group `10`-`13`も名前を確認後、全件 `DELETE 204` → `GET 404`。seed group `1`=`執行`、`2`=`一般`、`9`=`閲覧専用`は前後ともHTTP 200で無傷。
- P1: exam `1014562` は親MR削除後に撤去済み（初回DELETEはitemsのため409、再確認GETは404）。MR `1425546`、checkup `1`-`4`、vaccination `1`-`4`、MR `1425547`、care plan `1`-`2`は撤去済み。exam item `2325052`-`2325054`、hospitalization配下の日次記録・vital・care log・staff noteはDELETE route不在のため未撤去。hospitalization `1`は日次記録制約（HTTP 409）のため撤去不可。未撤去理由は `tmp/fe12-r4-evidence/2026-07-28/cleanup-report.md` にID単位で記録した。
- 巻き込み確認: pet `1001002` / `1000018` は前後とも `alive`・`danger_level=low`・`deceased_at=null`。DBへ直接SQLは実行していない。production code、migration、seedは未変更。
- evidence: `tmp/fe12-r4-evidence/2026-07-28/before.txt`、`after.txt`、`cleanup-report.md`、`completion-report.md`。R-4前のfixture生存実測は履歴であり、staff/group・MR/exam/checkup/vaccinationは本結果で撤去済み、hospitalizationは日次記録制約で残存する。

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

## PO判断待ち2件の代理検討とR-1完了記録（2026-07-28）

コードと仕様書の実読で判断可能なところまで詰めた。**裁定**＝証拠が答えを一意に決めるもの、**提案**＝判断が要るが材料を揃えたもの（曽我承認待ち）、**委任外**＝7/20に定めた委任境界（臨床値・外部書き込み・課金/物理操作）に触れるもの、の3ラベルで区別する。

### 【裁定】R-1③ — backend が正本。frontend/OpenAPI の「新規staffにemail/password必須」が defect である

証拠が一意に決めるため裁定とする。

- `001_init.sql:271` — `account_id bigint REFERENCES accounts(id) ON DELETE SET NULL`。**nullable であり、account削除時もstaff行は残る設計**。アカウント無しstaffはschemaの想定内。
- seed `003_demo/staffs.csv` — 全37件中 **20件が `account_id` 空**。内訳は `resource` 9 / `nurse` 9 / `doctor` 2。
- `models.ts:2836` — `StaffTypeResource = "resource"` がenum値として存在。
- `StaffLineReservationSection.tsx:56-57` — staff formは `staffType` をStaffType全値域のSelectで公開している。**UIから `resource` staffを作成できる。**
- `resource` は予約枠として選択される非人物リソース（部屋・機材等）であり、**原理的にログインを持ち得ない**。ここにemail/passwordを必須化すると、正当な業務データを作成不能にする。

→ **backendの「アカウント無しstaff作成許容」が正しい。** 連動する①②④⑤の是正方針: ①emailは「入力された場合のみ形式検証」（必須化しない）、②既存staffのpasswordは「非空なら8文字以上かつ英数字混在」をfrontendでも検証（`ValidatePassword`と同契約）、③OpenAPIのrequiredからemail/passwordを外す、④空氏名testの名称を編集経路と明記、⑤関連付けmutationの直前に権限グループ・所属医院の権限を再検査し、権限喪失時は発行しない。判断者への確認事項は**残っていない**。

### 【裁定】R-3の実現可能性 — 既存データからの基準値自動生成は採れない

- `exam_reference_ranges` のキーは `(clinic_id, exam_type_field_id, animal_species_id)`（`001_init.sql` の該当CREATE TABLE、UNIQUE制約 `uq_exam_reference_ranges_clinic_field_species`）。**species次元を持つ。**
- 一方 `exam_type_fields.normal_value` は `6.0-17.0 x10^3/uL`・`37-55%`・`10-125 U/L` のような**単位混在の自由文字列**で、**species次元を持たない**（seed実査）。
- 犬と猫で基準値は異なる。normal_valueから機械変換すると「全speciesで同一の基準値」を暗黙に主張することになり、**臨床的に誤った判定を生む**。

→ **normal_valueからのマイグレーションは禁止。** 基準値を入れるなら獣医師が種別ごとに与える必要がある。

### 【深刻度の訂正】R-3は「M-02が実測できない」問題ではない — 臨床機能が全体で停止している

**本節の初出時に深刻度を過小評価していた。訂正する。** 追加調査で、基準値不在はM-02のfixture固有の事情ではなく、**システム全体で検査異常値の自動ハイライトが機能していない**状態であることが判明した。`docs/spec/specification.md:21` が標準装備と謳う機能が動いていない。

- `examination_service.go:472-490` の `ReplaceItems` は item ごとの `refMin`/`refMax` を **`resolvedRanges`（`exam_reference_ranges` の解決結果）からしか設定しない**。request が運ぶ `in.RefMin`/`in.RefMax` は一度も読まれない。
- ~~一方 request DTO は今も `ref_min`/`ref_max` を受け取り、frontendは `parseNormalRange` で送り続けている。APIは受理して黙って破棄する（silent contract break）。~~ **2026-07-28 に解消済み**: 死んだ旧経路は撤去された（`546e26f80` / `e7721b6bd`、台帳 `65ca66bd2` で反映）。request DTO の `ref_min`/`ref_max` も frontend の `parseNormalRange` も現存しない（実測確認済み）。二重管理は解消し、基準値の正本は `exam_reference_ranges` の一本になった。
- 新経路の**投入手段は #249 U4（`afd8404a4` API＋`1adf55b6e` UI）で開通した**が、`exam_reference_ranges` は依然0行であり、獣医師が動物種ごとの臨床値を投入するまで `unassessedExamResult()` が全件へ適用される状態は変わらない。

**`3-session-agent.html#BUG-449` としてCRITICAL起票済み。** FE12-02の枠を超えるため、以降の追跡は同台帳を正本とする。本ledgerのR-3は「M-02をどう実測するか」に限定する。

### 【提案・曽我承認待ち】R-3のA/B — M-02の実測についてはBを推す（基準値不在のまま進める）

BUG-449の恒久是正（基準値マスタの投入経路を作る）とM-02の実測は分けられる。Aを待つとM-02が無期限に止まるが、M-02の主要な問い（一覧cueの要否）は基準値なしで答えられる（下記）。layout実測も基準値と無関係に実施できる。したがって**M-02はBで進め、BUG-449は別途恒久対応する**のが速い。ただしM-02の裁定メモには「判定機能が停止した状態で観測した」旨を必ず残し、BUG-449解消後にcue要否を再確認するかを曽我が決める。

### 【提案・曽我承認待ち】M-02の一覧cue要否 — 一覧にcueを追加しない

基準値が無くても、この問いは既存実装から答えられる。

- **詳細面には非色cueが既にある**: `ExamItemsTable.tsx:82` の `aria-label="基準値内"`、`:138` の `data-abnormal={String(!!item.isAbnormal)}`、`:163` の項目別 `aria-label`。色に依存しない識別手段は実装済み。
- **一覧はexam単位であり、異常はitem単位である**: `/examinations` が表示するのは `resultSummary` と `status` バッジ（`ExaminationsList.tsx:225,229-230`）。`exam_results.is_abnormal` はitem単位の値で、一覧のデータモデルに存在しない。
- 一覧へ集約フラグを出すにはbackendのexamレベルrollupが要る＝**新規実装**。「あれば便利」で追加すると製品哲学①②に反する。

→ 提案は「一覧にcueを追加せず、行→詳細のaccessible導線で担保する」。runbookが想定する「不要と裁定する場合は詳細へのaccessible導線があること」の分岐に該当する。**ただし4 viewportのlayout実測（wrap/clip/overlap）は別途必要であり、これは基準値と無関係に実施できる。**

### 【提案・曽我承認待ち】M-01-D — 死亡ペットに対する操作別の可否

適用原則は2つ。`docs/spec/specification.md:21`「**死亡ペットに対する誤操作の物理的ブロック**」と、製品哲学「確認ダイアログによる安全対策は禁止。ロック・Undo・物理ブロックで解決する」。したがって「ダイアログで確認して続行可」という選択肢は最初から採らない。

| 操作 | 提案 | 根拠 | 現状 |
|---|---|---|---|
| owner名link | **許可** | 飼主情報の閲覧は個体の生死と無関係。遮断する業務理由がない | 未確認 |
| report（飼主レポート） | **許可** | 過去の診療記録の閲覧であり、死亡後にこそ必要になり得る | 未確認 |
| pet編集 | **物理ブロック** | 死亡個体の属性変更は誤操作の典型。ただし**誤記訂正の正規経路を別に用意する必要がある**（ブロックのみだと訂正手段が消える） | 未確認 |
| 削除 | **物理ブロック** | 臨床記録の保全。削除は記録の消失であり、死亡個体でこそ不可逆 | 未確認 |
| 死亡登録/解除 | **確定済み** | `45b681866` 以降fail-closedで、既死亡への再登録・生存への解除は409 | 実装済み |

**「未確認」は現状のUI挙動を実測していないという意味であり、提案の根拠が弱いという意味ではない。** M-01-Eの証跡収集で現状が判明すれば、曽我は上表の提案列と実測列を突き合わせて承認/修正するだけで済む。**pet編集をブロックする場合の誤記訂正経路**だけは、曽我の判断が実装方針を分ける（追記のみ許す／管理者権限でのみ許す／死亡解除→訂正→再登録の順を強制する）。

### 【委任外】臨床値そのもの

WBC・RBC・HCTの**犬別/猫別の具体的な基準値**は、7/20に定めた委任境界（臨床値はユーザーに残す）に該当する。獣医師の確認を経ない値をマスタへ入れることは、判定結果が診療判断に直結する以上、行わない。Aを採る場合はこの工程が先行する。

## 実測レーン分割（2026-07-27・fixture完成後）

**2026-07-28 訂正**: 本節冒頭は「fixtureは全て揃っている」と書いていたが、実測の結果 S1 以外の fixture は消滅していた（「fixture生存実測」節）。**S2・M-02・M-05 は fixture 再作成が着手工程に加わる。** PO裁定を待つのが **M-02（R-3経由）・M-01-D の2件**である点は変わらない。以下の干渉分析は resource の重なりの分析としては引き続き有効である。全レーンが同一のlocal環境を共有するため、素直な2レーン並行は組めない。

> **2026-07-28 訂正**: 本節は当初「全レーンが同一の**disposable local**を共有する」と書いていたが、**disposable な stack は存在しない**。稼働しているのは名前付き永続ボリュームを持つ通常の Compose stack（`backend`/`db`/`frontend`）であり、config だけでは破棄可能性を証明できない。M-04 は pet `1000018` を実際に死亡させて戻すため、**この環境で実施してよいかは曽我の判断が要る**（下記「S1レーン試行の結果」の決定事項①）。

**干渉の実査**

| 組 | 重なるresource | 判定 |
|---|---|---|
| M-01 × M-05 | pet `1001002`/`1001004`/`1001005` | M-01が死亡登録/解除を試行し状態を動かす。M-05はその状態のsnapshotを観測する → **並行不可** |
| M-04 × M-05 | H=`1`、pet `1000018` | M-04が一時死亡登録→復旧を行う → **並行不可** |
| M-03 × M-04 | `/hospitalization` 系route | reset後のhospitalizationはH=`1`の1件のみで、M-03のroute巡回と対象が重なる → **条件付き**（下記制約で回避） |
| M-01 × M-04 | pet `1001005` vs `1000018` | 対象petが異なり、routeも `/owners` と `/hospitalization` で分かれる → **並行可** |

**S1 — 臨床lifecycleレーン**（所有: `/owners`・`/hospitalization`、pet `1001002`/`1001004`/`1001005`/`1000018`、H=`1`）

- M-01-E（証跡収集）と M-04 を担当する。両者は対象petもrouteも重ならないためレーン内で並行してよい。
- 死亡登録/解除を実際に行うため、**このレーンだけが臨床状態を動かす**。各操作の前後でread-backを取り、レーン終了時点の状態をM-05へ引き渡す。
- M-01の「どの操作を許可するか」の裁定（M-01-D）は曽我の担当でありレーン外。証跡だけ揃えて渡す。

**S2 — RBACレーン**（所有: `/settings/permission-groups`、staff `38`/`39`/`40`、group `10`-`13`）

- M-03 を担当する。判定基準は「禁止personaからのPOST/PATCH/PUT/DELETEが0件」であり、大半は**拒否されることの観測**であって成功するmutationではない。
- **制約**: 成功するmutationは自分が所有するTARGET group（`13`）に対してのみ行う。臨床record（medical-records/hospitalization/vaccinations/examinations）へのmutation試行は「拒否される」ことの確認に留め、create-only personaで臨床recordを実際に作らない。これによりS1との干渉を回避する。
- 注記: 検査登録POSTは`24929e83d`以降create+editの合成認可であり、create-only personaが拒否されるのは正しい期待値である（fixtureの欠陥ではない）。

**直列尾部（S1/S2の両完了後・単独実行）**

- M-05 — 横断snapshot観測。death・danger高・期限4区分・入院を同時に固定した状態で証跡を取るため、他レーンが動いている間は結果が無効になる。
- R-4 — fixture復旧。M-05完了後に実行する。

**着手できないもの（レーン外）**

- M-02 — R-3のPO裁定待ち。基準値が無いとHIGH/LOW cueが出ず、「一覧にcueが要るか」という問い自体が成立しない。
- M-01-D — 曽我の裁定。S1がM-01-Eを終えていれば、成果物を見て答えるだけで済む。
- line-reserve — 実機3台とQA環境の受け渡し。PO待ちではない。

**証跡収集の自動化余地（未検証・要設計）**: M-01-E/M-03/M-04/M-05が要求する成果物（4 viewport screenshot・accessibility tree・accessible name・network HAR・computed token・overflow計測）は判断を含まない機械的な収集である。`.codex/config.toml`にChrome DevTools MCPが定義済みで、`resize_page`/`take_screenshot`/`take_snapshot`/`list_network_requests`が対応する。1ルート（M-01-E）で試行し、成果物が曽我の判断材料として成立するかを確かめてから横展開するのが安い。ただし禁止操作のmutation 0件検証は実際に操作を試みる必要があり、復旧手順込みで設計すること。

**2026-07-28 試行結果 — M-01-E で試し、着手前提の段階で BLOCKED した。**収集の質は評価に至っていない（1枚も取得できていない）。判明したのは「経路が使えるか」以前の前提4件であり、うち3件は着手プロンプト側の欠陥である。**横展開（M-03/M-05）は前提が揃うまで行わない。**

## ブラウザ経路の確立（2026-07-28・3点gate 全green）

過去3走を止めていた `9222` 不通は解消した。**再現手順は下記1コマンドである。**

```
nohup "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" \
  --remote-debugging-port=9222 \
  --user-data-dir=<専用プロファイルパス> \
  --no-first-run --no-default-browser-check \
  http://127.0.0.1:3003 >/dev/null 2>&1 &
```

- **gate 1** `curl http://127.0.0.1:9222/json/version` → 200、`Chrome/150.0.7871.182` / `Protocol-Version 1.3`。
- **gate 2** chrome-devtools MCP `list_pages` → `1: Animal Ekarte - 電子カルテ (http://127.0.0.1:3003/login?from=%2F) [selected]`。
- **gate 3** `resize_page 1440×900` → `take_screenshot` でログイン画面のスクリーンショットを実取得。**証跡が実際に1枚取れることまで確認済み。**

**日常使いのChromeは `--remote-debugging-port` 無しで起動しているため 9222 は開かない。** 専用 `--user-data-dir` の別インスタンスを立てるのが正解で、既存プロファイルと競合しない。M-03 の3 persona login がセッション混線しない副次利点もある。ログイン画面が `password` とデモアカウントを自ら表示するため、credential の外部受け渡しも不要である。

## fixture生存実測（2026-07-28 02:5x・admin API・read-onlyのGETのみ）

台帳が「fixture完成済」と書く一方で、7/28 の2回目S1実行記録は「staff `38` は login `401` / read `404`」「Phase 0 の S1 fixture を**APIで復元した**」と書いており自己矛盾していた。admin（`admin@noavet.jp`）でloginし全fixtureをGETで実査した。**mutationは一切行っていない。**

経路の妥当性は同一routeの既存IDを対照に取って確認した（`/masters/staffs/1`=200、`/masters/permission-groups`=200、`/medical-records/1080036`=200）。したがって下記の404は経路誤りではなく**レコード不在**である。

| fixture | 実測 | 判定 |
|---|---|---|
| pet `1001002` | 200 `alive/low` | **生存** |
| pet `1001004` | 200 `alive/high`・`danger_reason=FE12-M01 fixture` | **生存** |
| pet `1001005` | 200 `deceased/low`・`deceased_at=2026-07-27T00:00:00+09:00` | **生存** |
| pet `1000018` | 200 `alive/low` | **生存** |
| hospitalization `1` | 200 `pet_id=1000018,status=admitted,cage_id=3,doctor_id=1,memo=FE12-M04 admitted`。ただし `created_at=2026-07-28T01:45:22+09:00` | **生存（7/28 01:45に再作成されたもの）** |
| medical record `1425546` / `1425547` | 404 / 404 | **消滅** |
| examination `1014562` と items `2325052`-`2325054` | 親MR不在 | **消滅** |
| staff `38` / `39` / `40` | 404 / 404 / 404 | **消滅** |
| permission group `10`-`13` | 全て404 | **消滅** |
| checkup `1`-`4`・vaccination `1`-`4` | M2=`1425547`配下のため | **消滅** |

### 2026-07-28 レーン2 — 消滅fixtureを全て再作成した（ブラウザ不要のAPI作業として並行実施）

M-02・M-05 の fixture を admin API で再作成し、read-back まで完了した。**IDは台帳の指定値と全て一致した**（連番の払い出しが同じ位置に戻ったため）。

| fixture | 実測 |
|---|---|
| M-02 medical record | `1425546`（`MR-20260728-1-ayMjhV`・draft・pet `1000018`/owner `300003`） |
| M-02 examination | `1014562`（`exam_type_id=3`・`doctor_id=1`・`date=2026-07-28T09:00:00+09:00`・`status=result_entered`） |
| M-02 items | WBC=`2325052`(10.0) / RBC=`2325053`(9.0) / HCT=`2325054`(30.0)。**3件とも `status=normal, is_assessed=false, is_abnormal=false`** |
| M-05 medical record | `1425547`（`MR-20260728-1-kg78h7`・draft・pet `1001002`/owner `300588`） |
| M-05 checkup | `1`-`4`（`checkup_type_id=1`・`date=2026-07-20`・`result`にラベル） |
| M-05 vaccination | `1`-`4`（`vaccine_id=1`・`date=2026-07-20`・`remarks`にラベル・`next_schedule_type=[other,other,other,NULL]`） |

- **期限日を実測日基準へ切り直した**: 旧fixtureは 7/27 基準で past/today/future を並べていたが、実測は 7/28 以降になるため **past=`2026-07-27` / today=`2026-07-28` / future=`2026-07-29` / empty=`NULL`** で作成した。7/27 基準のまま作ると「today」が翌日には past になり、M-05 の4区分が崩れる。**実測日が 7/29 以降へずれる場合は再度切り直しが必要である。**
- **M-02 の `is_assessed=false` は R-3 の再確認である。** `exam_reference_ranges` が0行のままなので RBC=9.0（基準5.5-8.5超過）も HCT=30.0（基準37-55未満）も未評価fallbackの `normal` を返す。**この `normal` を臨床的な正常と読み違えてはならない。** 曽我裁定 R-3=(B) のとおり、この状態のまま M-02 を実測する。
- exam field ID の正本は `backend/migrations/seeds/003_demo/exam_type_fields.csv`（`exam_type_id=3` の WBC=`1` / RBC=`2` / HCT=`3`）。**`exam-types` の read API はbackendに存在しない**（route literal は `/:id/checkups`・`/checkups`・`/vaccinations` のみ）ため、seed CSV が唯一の参照元である。
- **checkup には detail route が無い。** `GET /api/v1/checkups/:id` は404を返すが、これはレコード不在ではなく**routeが存在しない**ためである（`/checkups` は list のみ）。生存確認は `GET /api/v1/medical-records/1425547/checkups` で行うこと。次に fixture 生存を疑う人が誤診しないよう明記する。
- S1（pets 4頭・H=`1`）と S2（staff `38`-`40`・group `10`-`13`）の fixture も同時に read-back し、全て HTTP 200 で健在を確認した。**これはR-4前の履歴であり、staff/groupはR-4（2026-07-28）で撤去済み。**

**帰結（着手計画への影響）**

- **2026-07-28 R-4前、全レーンの fixture は再作成済みで揃っていた**（M-03=S2 fixture 再作成節、M-02/M-05=レーン2節、S1=pets 4頭とH=`1`）。**これは実測前の履歴であり、R-4後はstaff/group・MR/exam/checkup/vaccinationが撤去済み、hospitalization `1`は日次記録制約で残存している。**
- **R-4（復旧工程）の対象**: `1001005`→alive、`1001004`→low、M-04で一時死亡させた場合の `1000018` の復帰、staff `38`-`40` と group `10`-`13` の削除、M-02/M-05 の MR `1425546`/`1425547` 配下（exam・items・checkup・vaccination）の削除。
- API作成物が消えた原因は特定していない（DB resetの実行が最有力だが証跡なし）。**同じ消滅が再発する前提で計画すべきであり、実測に着手する前に必ず read-back で生存を確認する。** 作成から実測までの間に日をまたぐ設計は採らない。

## S2 fixture 再作成（2026-07-28・完了）

消滅していた M-03 の fixture を admin API で再作成し、read-back まで完了した。**IDは台帳の指定値と完全に一致した**（連番の払い出しが同じ位置に戻ったため）。

- permission group: VIEW=`10`、CREATE=`11`、EDIT=`12`、TARGET=`13`。VIEW/CREATE/EDIT は `master-permission`・`medical-records`・`hospitalization`・`vaccinations`・`examinations` の5 resourceへ rule 5件ずつ。TARGET は rule 0件の器。resource名は `backend/internal/model/permission.go:7-58` の定数を全数読んで確定した。
- staff: VIEW=`38`（`fe12-m03-view@noavet.jp`）、CREATE=`39`（`fe12-m03-create@noavet.jp`）、EDIT=`40`（`fe12-m03-edit@noavet.jp`）、password=`Fe12pass1`。
- 割当 read-back: `38->[10]`、`39->[11]`、`40->[12]`。3 personaのloginは全て HTTP `200`。`GET /me` で staff `38` = clinic `1` / `is_system_admin=false` を確認。
- **落とし穴**: 割当APIのbody keyは `permission_group_ids` ではなく **`group_ids`**（`staff_request.go:95`）。誤ったkey名で送ると `binding:"max=50,dive"` に `required` が無いため空sliceとして通り、**HTTP 200 で「全グループ解除」が実行される**。permission を変更するendpointで「未送信」と「空配列」が区別できない契約であり、再作成時はread-back必須。

### 実装drift ①の根本原因を特定 — 台帳の記述より深刻で、systemicである

台帳は「staff作成POSTが `reservation_visible:false` を無視する」と記録していた。**再現し、原因を特定した。staff固有の問題ではない。**

- HTTP DTO（`staff_request.go:19` `ReservationVisible *bool`）も service（`staff_service_account.go:63-66,92` / `staff_service_core.go:92-95,112`）も **`false` を正しく伝搬している**。両方とも nil判定付きで正しい。
- 真因は `backend/internal/model/staff.go:35` の `ReservationVisible bool \`gorm:"not null;default:true"\`` である。**GORM は `default` tag を持つfieldの zero value を INSERT 文から除外し、DB default を適用させる。** `bool` の zero value は `false` なので、`false` を明示指定しても INSERT に載らず DB default の `true` が入る。
- **`.Create()` 経路では `false` を新規作成できない。** PATCH（`Updates` map 経由）は正しく `false` を書けるため、「作成時だけ落ちる」挙動になる。
- **同じ形のfieldは `backend/internal/model/*.go` に 39箇所・26ファイル存在し、`*bool` で回避しているものは 0件である。** 全てが構造的に同じ欠陥を持つ。ただし実害は「作成時に `false` を受け付けるAPIがあるか」に依存するため、39件全部が現に壊れているという意味ではない。
- **`reservation_visible` は実害が確認できる。** LINE予約に出さない意図で作成したstaffが黙って予約可能側に入る。業務上の誤りが表示面に出ないまま顧客導線へ露出する。
- 修正方針は2択。(A) 該当fieldを `*bool` へ変える（GORMは非nilなら zero value も書く）。(B) 作成時に `Select` で当該列を明示する。**(A)を推す。** (B)は新しい作成経路を書くたびに再発する。**起票先は `3-session-agent.html#ledger`。本ledgerの所掌外。**

## S1レーン試行の結果（2026-07-28）— 再開前に潰す4件

M-01-E + M-04 の試行は preflight で停止した。実行記録の全文は M-04 節末尾を正本とする。ここには**再開に必要な決定と、繰り返してはならない設計上の誤り**だけを残す。

**環境側（2026-07-28 に全て決着）**

- **① 実測環境 — 決定: (A) この dev stack で実施する（曽我裁定・2026-07-28）**。disposable stack は存在しないが、M-04 の死亡遷移はこの環境で行う。手順は「生存→死亡登録→死亡解除」を1回ずつ、各段でread-backを取る。**死亡登録/解除は `45b681866` 以降 fail-closed であり、順序を誤ると409で戻せなくなる**点だけが実務上の注意である。
- **② ブラウザ経路 — 決定: Chrome DevTools MCP（`127.0.0.1:9222`）を正とする**。専用プロファイルでリモートデバッグ付きChromeを起動する。既存の日常ブラウザと混ぜないため `--user-data-dir` を分ける（M-03の3 persona loginがセッション混線しない利点もある）。**着手条件は3点すべてgreen**: (1) `curl 127.0.0.1:9222/json/version` が200、(2) chrome-devtools MCP の `list_pages` 成功、(3) **アプリ画面のスクリーンショットが実際に1枚取れる**。過去3走は(1)だけを見て委譲したため死んだ。`claude-in-chrome` extension も接続可能だが、`resize_window` はウィンドウサイズであってviewportではなく、500×900のoverflow実測が chrome UI 分ずれて無意味になるため**採らない**（どうしても9222が立たない場合のスクショ専用fallbackに留める）。

**着手プロンプト側の欠陥（次に書く人が繰り返さないための記録）**

- **③ `browser-test` skill を Codex 向けプロンプトの backing に指名してはならない**: `.agents/skills/browser-test/SKILL.md` は冒頭で「**必須: Haiku Agent で実行せよ**」と規定する。Codex セッションに Haiku 経路は無い。**スキル名だけで指名せず SKILL.md の中身を読むこと。** 収集経路は Chrome DevTools MCP を直接使う形で書き、browser-test skill は参照に留める。**なお実行セッション自身が Chrome DevTools MCP を直接叩く場合、この制約そのものが消滅する。幻のブロッカーとして再輸入しないこと。**
- **④ 権限剥奪caseの persona — 決定: S2→S1 を直列化し、group `10`-`13` の1つを借りる（実行計画上の決定・曽我判断不要）**。台帳の「2レーン並行」は実行者が2人いる前提で書かれていたが、その前提は成立していない。直列化すれば M-03 完了時点で S2 の group は空くため、M-04 の剥奪対象として借用できる。**これは妥協ではなく前提の訂正である。** 借用したgroupはR-4で削除する。

**あわせて判明した仕様の曖昧点（M-01-E の合否に直結）**

- **F16 は badge のみか、行全体か**: 実装は死亡ステータス**バッジ**をグレーにしている（`OwnersListTable.tsx:236-241` が `StatusBadge` に `getPetStatusColor` を渡し、`status-helpers.ts:176-180` が死亡へ `BADGE.grayHover` を返す）。**行全体のグレーアウトではない。** 本ledgerの M-01 節は「一覧はグレーアウト維持」としか書いておらず、どちらを期待値とするかが確定していない。**曽我の裁定が要る**（下記「残余risk」へ再掲）。
- **view-only の hospitalization board カードが edit 権限なしでは開けない**（list / 直接 detail 遷移は可能）。意図的なら問題ないが、view-only 利用者の導線として要確認。
- **M-04 が要求する child 操作の一部が現行 UI/API に存在しない可能性**: daily / vital / care log / note の edit・delete が現行 frontend モジュールに見当たらないとの静的所見。**runbook の要求と実装面の突合が未了**であり、再開時は「存在しない操作」を未実施として区別できる形にする。

## 要実測項目

### M-01 OwnersList操作範囲

- Route: `/owners?include_deceased=true`。
- Fixture seed source: local `003_demo`へ`backend/migrations/seeds/003_demo/owners.csv`と`pets.csv`を適用し、`docs/ops/testing/scenarios/S01-deceased-pet-guard.md`/`V03-owner-pet-staff-forms.md`の手順で同一ownerにalive、danger=`高`、deceasedの3頭を準備する。実測後はS01どおり生存へ戻す。
- Fixture実査: D=`2026-07-27`。schema canaryを兼ねた`PATCH /api/v1/pets/1001004`の逐語body `{"danger_level":"high","danger_reason":"FE12-M01 fixture"}`はHTTP 200。`PATCH /api/v1/pets/1001005/death`はAPI field `reason`へ`"M01/M05 fixture"`を渡してHTTP 204。3頭の最終`GET /api/v1/pets/:id`をHTTP 200でread-back済み。
- 追加すべき具体値: local環境上で`1001002={status:alive,danger_level:low,deceased_at:null}`、`1001004={status:alive,danger_level:high,danger_reason:"FE12-M01 fixture",deceased_at:null}`、`1001005={status:deceased,danger_level:low,deceased_at:2026-07-27T00:00:00+09:00}`を確認済み。pet public responseが`deceased_reason`を公開しないことは現行契約どおり。
- 着手ブロッカー: runtime schemaとM-01 fixtureのブロッカーは解消済み。canaryはSQLSTATE 42703を返さず、目標3状態のread-backまで完了した。
- 次の一手: 実測担当と曽我または同席する仕様責任者が上記固定IDでrunbookを実行する。完了後の別cleanup工程で`1001005`をalive、`1001004`をlowへ復旧する。
- Persona: owners `view=true, edit=true, delete=true`の通常担当者。曽我または同席する仕様責任者が許可操作を記録する。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: include-deceasedを有効にし、3 rowそれぞれでowner名link、report、編集、削除、pet死亡登録/解除をpointer、Tab/Enter/Space、直接URLで試す。各操作前後のrequestをnetworkで記録する。
- Expected result: aliveは通常操作可、danger=`高`は「⚠ 危険」等の非色cueとaccessible nameを失わない。deceasedはF16裁定（一覧はグレーアウト維持・`3b7524748`）どおりの表示を記録し、曽我が許可/禁止する操作を操作単位で明示する。禁止とされた操作はmutation 0件。
- Required evidence artifacts: 4 viewport screenshot、accessibility tree、各操作のaccessible name、GET以外のnetwork HAR、操作可否の曽我裁定メモ。

#### 2026-07-28 M-01-E 実測結果

- status: **COMPLETE**。`/owners?include_deceased=true` を開いたままでは対象 fixture が初期表示に出ず、`search=原田` を併用して live runtime 上の `1001002` / `1001004` / `1001005` を確認した。prompt の fixture 名は現行 seed とずれていたが、ID と状態は一致したため、そのまま実測を完了した。
- live read-back: `1001002={status:alive,danger_level:low,deceased_at:null}`、`1001004={status:alive,danger_level:high,danger_reason:"FE12-M01 fixture",deceased_at:null}`、`1001005={status:deceased,danger_level:low,deceased_at:2026-07-27T00:00:00+09:00}`。`GET /api/v1/me` は `id=41, display_name="FE12-M01 通常担当", is_system_admin=false, main_clinic_id="1"`、`/api/v1/masters/staffs/41/permission-groups` は `group_ids=[1]` だった。
- row action matrix: `1001002` / `1001004` / `1001005` の各行で操作メニューを開き、`編集`、`レポート`、`削除` がすべて表示された。`ariaDisabled` と `disabled` はいずれも false で、操作単位の非活性は未実装だった。
- danger cue: `1001004` の高危険行は `⚠ 危険` を表示し、`role=button name="クロの危険理由を表示"` から popover を開くと `role=dialog name="クロの危険理由"` と `危険理由` が accessible tree に出た。色だけに依存しない cue は維持されていた。
- deceased badge: `1001005` は一覧上で死亡状態を灰色 badge として表示し、行全体のグレーアウトではなかった。これは既存裁定どおりで、badge のみの gray 表示を合格基準として満たした。
- network: interaction window の `GET` 以外は `# non-get count: 0`。login 時の `POST /api/v1/login` は baseline auth であり、menu 操作区間の mutation は 0 件だった。
- before/after: `pets-before-after.txt` の read-back 差分はなし。`1001002` / `1001004` / `1001005` は操作前後で不変だった。
- evidence paths: `tmp/fe12-m01e-evidence/2026-07-28/shot-owners-1440x900.png`、`tmp/fe12-m01e-evidence/2026-07-28/shot-owners-1200x800.png`、`tmp/fe12-m01e-evidence/2026-07-28/shot-owners-800x1024.png`、`tmp/fe12-m01e-evidence/2026-07-28/shot-owners-500x900.png`、`tmp/fe12-m01e-evidence/2026-07-28/a11y-owners-list.txt`、`tmp/fe12-m01e-evidence/2026-07-28/a11y-owners-menu-1001002.txt`、`tmp/fe12-m01e-evidence/2026-07-28/a11y-owners-menu-1001004.txt`、`tmp/fe12-m01e-evidence/2026-07-28/a11y-owners-menu-1001005.txt`、`tmp/fe12-m01e-evidence/2026-07-28/net-owners.txt`、`tmp/fe12-m01e-evidence/2026-07-28/pets-before-after.txt`、`tmp/fe12-m01e-evidence/2026-07-28/completion-report.md`。

### M-02 Examinations一覧意味とlayout

- Route: `/examinations`と対象petの`/medical-records/:id`検査履歴。
- Fixture seed source: `backend/migrations/seeds/003_demo/exams.csv`、`exam_results.csv`、`exam_types.csv`、`exam_type_fields.csv`を基に、`docs/ops/testing/scenarios/S02-exam-abnormal-highlight-lock.md`のhigh/low/normal同居fixtureを作る。
- Fixture実査: D=`2026-07-27`。pet `1000018`へdraft medical record M=`1425546`（record_no=`MR-20260727-1-eS5A7a`）とexam E=`1014562`をAPIで作成し、items PUTでWBC item=`2325052`、RBC item=`2325053`、HCT item=`2325054`を登録した。M/E作成はHTTP 201、items置換と`GET /api/v1/examinations/1014562/items`はHTTP 200。
- 追加すべき具体値: Eは`exam_type_id=3,doctor_id=1,date="2026-07-27T09:00:00+09:00",status=result_entered,result_summary="FE12-M02 HIGH/LOW/normal"`。itemsはWBC=`10.0`、RBC=`9.0`、HCT=`30.0`で、requestに`status/is_assessed/is_abnormal/ref_min/ref_max`を書いていない。read-backは3件とも`status=normal,is_assessed=false,is_abnormal=false`であり、これは基準値不在時の未評価fallbackで、RBC/HCTが臨床的に正常という証跡ではない。
- 着手ブロッカー: schema blockerとM/E/items未作成は解消済み。残るブロッカーは`exam_reference_ranges`の対象clinic/species/field基準値不在だけである。現行seedにrange recordは無く、rangeのread/write APIも存在せず、`exam_type_fields.normal_value`表示文字列は導出に使われない。基準値投入・導出値偽装は実施していない。
- 次の一手: master-data責任者が正規の別工程でWBC/RBC/HCT reference rangeを用意するか、基準値不在のまま実測するかを裁定する。rangeを用意した場合は同じE/itemsを正規APIで再評価し、WBC=normal・RBC=high・HCT=lowのread-back後に実測担当が一覧とカルテ履歴を比較する。
- Persona: examinations `view=true, create=false, edit=false, delete=false`のview-only担当者。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: 同一fixtureを一覧とカルテ履歴で開き、値・summary・statusを比較する。zoom 100%、pointer hover、keyboard focus、横スクロール有無を確認し、一覧でHIGH/LOW非色cueが必要か曽我へ提示する。
- Expected result: normalを異常表示しない。曽我が一覧cueを必要と裁定する場合はHIGH/LOWが色なしでも識別できること、不要と裁定する場合は詳細へのaccessible導線があること。全viewportでwrap/clip/overlapなし。
- Required evidence artifacts: 両surface×4 viewport screenshot、accessible text dump、computed color/token、overflow計測、曽我の一覧cue要否裁定。

#### 2026-07-28 M-02 実測結果

- status: INCOMPLETE（観測完了、表示突き合わせ不一致と responsive clip を記録）。view-only staff 42（group 9, is_system_admin=false）で両 route を開いた。access denied は両 surface 0 件。
- fixture/read-back: medical record 1425546、examination 1014562、items は HTTP 200。E、WBC=10.0、RBC=9.0、HCT=30.0、3 items の status=normal,is_assessed=false,is_abnormal=false を確認。normal は未評価 fallback であり臨床的正常ではない。
- surface comparison: /examinations は E の summary FE12-M02 HIGH/LOW/normal と exam status 結果入力済みを表示するが item values/status は表示しない。カルテ検査履歴は item names と 未判定（基準値未設定のため判定していない）を表示するが item values と summary は表示しない。詳細は tmp/fe12-m02-evidence/2026-07-28/surface-comparison.md。
- 未判定 safety: 3項目とも逐語表示に 未判定 と 基準値未設定の説明があり、正常/基準値内とは表示されなかった。
- 一覧 cue: HIGH/LOW の非色 item cue は観測されなかった。accessible detail link は 検査詳細: チビチビ 2026-07-28 ID 1014562。cue 要否の裁定はしていない。
- layout: 1440×900 と 1200×800 は主要要素の崩れなし。800×1024 と 500×900 では一覧の右側列およびカルテの右側情報/列に viewport 外表示または横方向 clip を観測し、500×900 では HCT 名も wrap した。
- network: 各 surface の hover/Tab/横スクロール試行区間で method 位置一致の POST/PATCH/PUT/DELETE は 0 件（login POST は区間外）。
- evidence: tmp/fe12-m02-evidence/2026-07-28/ に a11y 2件、PNG 8件、network raw/normalized、comparison、CDP probe を保存。全ファイル mode 0600、directory 0700。
- changed files: FE-refactor.md と tmp/fe12-m02-evidence/** のみ。backend/frontend production code は変更していない。

### M-03 RBAC非活性の理由/name

- Route: `/settings/permission-groups`、`/medical-records/:id`、`/hospitalization/:id/edit`、`/vaccinations/:id`、`/examinations/:id`。
- Fixture seed source: `backend/migrations/seeds/003_demo/permission_groups.csv`、`permission_group_rules.csv`、`staffs.csv`、`staff_permission_groups.csv`と各featureの既存record CSV。`docs/ops/testing/scenarios/V03-owner-pet-staff-forms.md`で試験用group/staffを作り、`V01-clinical-forms.md`の既存recordを使う。
- Fixture実査: permission groupはVIEW=`10`、CREATE=`11`、EDIT=`12`、TARGET=`13`、専用staffはVIEW=`38`、CREATE=`39`、EDIT=`40`。per-staff read-backは`38->[10]`、`39->[11]`、`40->[12]`で、3 emailのloginは全てHTTP 200。TARGETは操作対象の器としてruleを持たない。
- 追加すべき具体値: `master-permission`、`medical-records`、`hospitalization`、`vaccinations`、`examinations`へ共通で、VIEW group=`{view:true,create:false,edit:false,delete:false}`、CREATE group=`{view:true,create:true,edit:false,delete:false}`、EDIT group=`{view:true,create:false,edit:true,delete:false}`をread-back済み。loginは`fe12-m03-view@noavet.jp`、`fe12-m03-create@noavet.jp`、`fe12-m03-edit@noavet.jp`とpassword=`Fe12pass1`を使う。
- 着手ブロッカー: 4 group・3 staff・割当・persona loginのfixtureブロッカーは解消済み。staff作成bodyの`reservation_visible:false`が201 responseで`true`となるdriftは、各staffへ正規PATCHを行い最終read-backを`false`へ揃えた。production codeは変更していない。
- 次の一手: 実測担当が3 personaで再ログイン/再読込してrunbookを実行する。**実測後のR-4（2026-07-28）でstaff `38`-`42`とgroup `10`-`13`は撤去済み。**
- Persona: (1) view-only=`view:true/create:false/edit:false/delete:false`、(2) create-only=`view:true/create:true/edit:false/delete:false`、(3) edit-without-delete=`view:true/create:false/edit:true/delete:false`。
- 注記: 検査登録（POST）は`24929e83d`以降create+editの合成認可であり、create-only personaは拒否されるのが正しい期待値である。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: 各personaでrouteを再ログイン/再読込して開き、pointer、Tab/Enter/Space、formのprogrammatic submit、保存中の権限剥奪後callbackを試す。permission-groupは新規panel、既存panel、reorder、保存後rulesも個別確認する。
- Expected result: view accessは維持する。許可されたactionだけ実行でき、禁止controlはaccessible nameと理由を保持する。禁止personaからのPOST/PATCH/PUT/DELETEは0件で、same-commit剥奪後もmutationしない。
- Required evidence artifacts: persona×routeのaccessibility tree、4 viewport screenshot、network HAR、console log、action別permission matrixと0 mutation集計。

#### 2026-07-28 M-03 実測結果

- status: **COMPLETE**。v3で収集した15件の資産を再利用し、v4で `/hospitalization/1` の 3 persona 追加収集を完了。判定は「`/hospitalization/1` は3 personaで閲覧可能、`/hospitalization/1/edit` は VIEW/CREATE で拒否が期待値どおり」と更新。
- 実装確認: `frontend/src/app/routes/clinical-care-routes.tsx` の hospitalization ブロックでは、`path: ":id"`（`118-123`）に `RequirePermission` がなく、`path: ":id/edit"`（`125-130`）で `action="edit"` を強制。したがって、v3で `/hospitalization/1/edit` におけるVIEW/CREATE拒否は実装不備ではない。
- fixture/read-back: groups `10`/`11`/`12`は各5 rules、`13`はrule 0件。staff `38`/`39`/`40`の割当は`38->[10]`・`39->[11]`・`40->[12]`。3 persona login は全て HTTP 200。
- v4 追加収集: `/hospitalization/1` の a11y が3件 (`a11y-view-hospitalization-1-detail.txt` / `a11y-create-hospitalization-1-detail.txt` / `a11y-edit-hospitalization-1-detail.txt`)、screenshotが12件、network record が3件。3つの a11y すべてで `grep -c 'アクセス権限がありません' = 0`。3つのネットワーク差分で `NON_GET_IN_INTERVAL` は全て `0 件`。
- matrix/qualification: `permission-matrix.md` に `/hospitalization/1` の行を追加し、`/hospitalization/1/edit` の VIEW/CREATE 列を「拒否が期待値 → 一致」に再分類。`UNREPORTED=0` を維持。
- hospitalization contrast: `/hospitalization/1` の9業務フィールドはbefore/after同一（`hospitalization-before-after.txt`）。`updated_at` は同一。
- workflow orchestration: 本closeoutは `multi_agent_v1.spawn_agent` による multi-agent fan-out を使用し、read-only probes を 3 本に分割したうえで main thread が join/reconcile した。実行した subagent/role は Volta (`019fa50d-30f6-7822-a4d2-6dc2d4046bb8`, explorer), Tesla (`019fa50d-4906-7492-a054-4b879f5a064e`, explorer), Banach (`019fa50d-66a9-7a21-9588-b471fc5af979`, reviewer) の 3 件で、いずれも completed、write-owned paths なし、結果は route truth / artifact inventory / independent review の証拠として採用した。native Workflow tool はこの session では使っていない。
- changed files: `FE-refactor.md` と `tmp/fe12-m03-evidence/**` のみ。`backend/**`/`frontend/**` へ edit command は 0。

### M-04 Hospitalization child control実効性

- Route: `/hospitalization`のboard/listと`/hospitalization/:id`。
- Fixture seed source: `backend/migrations/seeds/003_demo/hospitalizations.csv`、`daily_records.csv`、`vital_records.csv`、`care_logs.csv`、`care_plan_items.csv`、`cages.csv`を基に、`docs/ops/testing/scenarios/S05-hospitalization-cycle.md`のadmitted fixtureを準備する。
- Fixture実査: D=`2026-07-27`。APIでH=`1`、DR=`1`、vital=`1`、care log=`1`、朝食care plan=`1`、投薬care plan=`2`（medicine_id=`14001`）、note=`1`を作成した。全mutationはHTTP 201、H/daily graph/care planのread-backはHTTP 200。
- 追加すべき具体値: Hはpet=`1000018`、owner=`300003`、`type=hospitalization,start=2026-07-27,end=2026-07-29,status=admitted,cage_id=3,doctor_id=1,memo="FE12-M04 admitted"`。DRには`09:00 JST/38.5℃/heart 120/respiration 30/8.0Kg/staff 1`のvital、`09:30/type=food/status=completed/value=完食/staff 1`のcare log、朝食と投薬のactive care plan、`10:00/content="FE12-M04 申し送り"/staff 1`のnoteが存在する。
- 着手ブロッカー: fixture未作成ブロッカーは解消済み。全childとnoteの指定値をread-backし、投薬care planはruntime必須contractに従いmedicine_id=`14001`を指定した。
- 次の一手: 実測担当がH=`1`で通常担当者の対照を1回記録後、view-onlyを検証する。最後にadminが対象petを一時死亡登録し、表示と操作不能を確認後に生存へ戻す。
- 注記: 死亡登録/解除は`45b681866`以降fail-closedであり、既死亡への再登録・生存への解除は409になる。一時死亡→復旧の手順は「生存→死亡登録→死亡解除」の順で1回ずつ行う。
- Persona: hospitalization view-only=`view:true/create:false/edit:false/delete:false`。対照として通常担当者=`view/create/edit/delete:true`を1回使う。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: board drag/drop、check-in/status、退院（会計あり/なし）、daily、vital、care log、note、care planのcreate/edit/deleteをpointer/keyboard/programmatic callbackで試す。操作dialogを開いた後に権限を剥奪するcaseも含める。
- Expected result: view-onlyおよびsame-commit剥奪後は全child/top-level mutation 0件。死亡fixtureは死亡文言を表示し、drag/check-in等を実行しない。非活性controlのnameと理由は残る。
- Required evidence artifacts: 操作別network HARと0件集計、accessibility tree、4 viewport screenshot、console log、会計あり/なし退院の個別記録。

#### 2026-07-28 S1レーン実測結果 — BLOCKED

- 対象unit: M-01-E + M-04。
- status: **BLOCKED**。Chrome DevTools MCPが`http://127.0.0.1:9222/json/version`からbrowser WebSocket URLを取得できず、着手前提のブラウザ経路を確立できなかった。推測または直接APIへの置換は行わず、fixture read-back、login、画面操作、臨床状態mutation、証跡収集を未実施のまま停止した。
- required input: (1) remote debugging port `9222`で待ち受けるChromeを起動し、Chrome DevTools MCPの`list_pages`が成功する状態にする、(2) 対象stackがfixture破棄可能なdisposable localであることを確定する、(3) binding browser-test skillを満たすHaiku-capable browser executor、または同skillの明示的な更新・逸脱承認を用意する。MCP approvalは今回のprobeを停止させなかった。
- changed files: 本unitのwriter変更は本execution recordだけ。本executorは`backend/**`と`frontend/**`へのedit commandを発行していない。既存のshared-worktree WIPは保持したが、要求された前後status snapshotが一致しないためworktree差分0は証明不能。
- gate evidence:
  - `node /Users/minoru/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fast-fe12-s1-measurement.md` → exit `0`; `Prompt Craft Harness Validation: PASS` / `Profile: standard (declared-risk-tier)` / `Execution contract: dynamic-workflow/v1`
  - `curl -sS --max-time 3 -D - http://127.0.0.1:9222/json/version -o /dev/null` → exit `7`; `curl: (7) Failed to connect to 127.0.0.1 port 9222 after 0 ms: Couldn't connect to server`
  - Chrome DevTools MCP `list_pages`（1回だけ実行）→ `Could not connect to Chrome. Check if Chrome is running.` / `Cause: Failed to fetch browser webSocket URL from http://127.0.0.1:9222/json/version: fetch failed`
  - `curl -sS --max-time 5 -o /dev/null -w 'frontend_http=%{http_code} remote=%{remote_ip}:%{remote_port}\n' http://127.0.0.1:3003` → `frontend_http=200 remote=127.0.0.1:3003`
  - `curl -sS --max-time 5 -o /dev/null -w 'backend_root_http=%{http_code} remote=%{remote_ip}:%{remote_port}\n' http://127.0.0.1:8080` → `backend_root_http=404 remote=127.0.0.1:8080`
  - `docker compose ps --status running --format '{{.Service}} | {{.Status}} | {{.Ports}}'` → `backend | Up 2 hours (healthy) | 0.0.0.0:8080->8080/tcp, [::]:8080->8080/tcp`; `db | Up 41 minutes (healthy) | 0.0.0.0:5434->5432/tcp, [::]:5434->5432/tcp`; `frontend | Up 41 minutes (healthy) | 0.0.0.0:3003->3000/tcp, [::]:3003->3000/tcp`
  - 実測前 `git -c core.quotepath=false status --porcelain -- backend/ frontend/ | shasum -a 256` → `625456ae87feb019722c1b4e696ea5ababa3aa80dcc6a460cfc9e2ab5bbfc246  -`
  - 停止後の同command → `d9084d03d99dc187565fa353066204d5e963f4b8b9b63aee67f8b7f216622670  -`。snapshotは変化した。変化したpathのwriter・時刻・原因はUNREPORTEDであり、要求されたstatus比較による「本実測起因の差分0件」はBLOCKEDとし、PASSへ読み替えていない。
- assumption deviations: `~/fe12-s1-evidence/2026-07-28/`は証跡収集開始前に停止したため未作成。browser-test skillが指定するHaiku Agentは本Codex sessionで起動不能だったため、browser readiness probeに限って単一のdelegated executorへfallbackした。このfallbackはbrowser-test skill充足とは扱わない。local Composeであることは確認したが、named persistent DB volumeのためdisposable性はconfigだけでは完全証明していない。
- clinical handoff: fixture read-back未実施。mutation操作は未試行で、M-04の非GET request数はUNREPORTED。本executorはpet `1000018`の死亡登録/解除を含む臨床mutation commandを発行していない。pet `1001005`には触れていないが、現在もdeceasedであることは未確認。M-05が前提にできる最終read-backは未取得。

#### 2026-07-28 S1レーン段階実行結果 — Phase 0臨床fixture復元済み / BLOCKED

- 対象unit: M-01-E + M-04。saved promptは`agent-fast-fe12-s1-phased.md`、validatorはexit `0`で`Prompt Craft Harness Validation: PASS`、`Execution contract: dynamic-workflow/v1`。全validator出力は`/Users/minoru/fe12-s1-evidence/2026-07-28/completion-report.md`のSaved Prompt Validation Gateへ逐語記録した。
- status: **BLOCKED**。Phase 0のS1-owned臨床fixtureはAPIで復元しexact read-backまで完了した。一方、S2-owned view-only persona staff `38`はlogin `401`、staff read `404`、該当permission groupなしであり、S2 resource `staff 38–40 / group 10–13`を変更しない境界に従って再作成しなかった。Phase 1–3はChrome `127.0.0.1:9222`不在のため未実施。
- Phase 0 final read-back: pet `1001002=alive/low`、`1001004=alive/high`（`danger_reason=FE12-M01 fixture`）、`1001005=deceased/low`（`deceased_at=2026-07-27T00:00:00+09:00`）、`1000018=alive/low`。H=`1`は`pet_id=1000018,status=admitted,cage_id=3,doctor_id=1`。daily record `1`、vital `1`、care log `1`、staff note `1`、care plan `1/2`をHTTP `200`でread-backした。sanitized exact evidence: `/Users/minoru/fe12-s1-evidence/2026-07-28/phase0-fixtures.md`（SHA-256 `fd57fdee675d61b2aef95033ed03ff358dcb9061273fd8ae01bb6cc955a36f8a`）。
- Phase 1 / M-01-E: **BLOCKED / UNREPORTED**。4 viewport screenshot、accessibility tree、accessible name、pointer/keyboard/direct URL、非GET network記録は未取得。死亡行のgrayがbadgeかrowかのruntime observationも未取得で、正誤判断を行っていない。
- Phase 2 / M-04: **BLOCKED / UNREPORTED**。view-only personaが存在せずChromeも不在のため、mutation 0件をPASSとしていない。通常担当者control mutationも未試行。
- Phase 3: **BLOCKED / 未実施**。pet `1000018`のalive→deceased→alive操作は開始していない。最終read-backは`alive`。pet `1001005`は`deceased`、pet `1001002/1001004`は`alive`を維持した。
- browser gate: `curl -s -m 3 http://127.0.0.1:9222/json/version` → exit `7`。Chrome DevTools MCP `list_pages`（1回）→ `Could not connect to Chrome. Check if Chrome is running.` / `Cause: Failed to fetch browser webSocket URL from http://127.0.0.1:9222/json/version: fetch failed`。Chromeの起動や別browserへの置換は行っていない。
- product non-modification: 本unitは`backend/**` / `frontend/**`へedit commandを発行していない。実行前status hash `619656aa2b37e9ea891f5e81d97834c2f05a03dd84e0e60ad4922d78e3283061`、初回実行後hash `9648ed7b78d637564b8716c9ab4a6a20739a206ea570f6bffbb6e79951b6098c`、review終了時hash `62796e2718f9cb92b01f9d99e189afb63bfe0706de046cbdfe2435ddadb4e2d4`。status差分の全pathは`/Users/minoru/fe12-s1-evidence/2026-07-28/completion-report.md`へ記録した。本unitのagent ownership外だが外部writer identityと時刻の独立証跡はないため、非改変checklistは**BLOCKED**。
- orchestration: read-only fan-out `/root/runbook_probe`、`/root/api_probe`、`/root/ui_contract_probe`、`/root/harness_probe`をjoinし、root writerだけが本ledgerとrepo外evidenceを更新した。native Workflow toolはsessionに存在しなかった。
- failure signatures: login初回は`X-Requested-With`欠落でHTTP `403`、header追加後にadmin HTTP `200`となった。staff `38`は同header付きでもHTTP `401`で、S2-owned fixture欠落としてBLOCKED。誤ったcare-plan endpointはsourceでrouteを確認し`/care-plan-items`へ訂正後HTTP `200`。
- assumption deviation / prompt defect: Phase 0は欠損fixture復元を要求する一方、staff `38–40` / group `10–13`の変更を禁止しており、S2 fixture欠損時の契約が衝突する。次回promptは「S1-owned臨床fixtureの復元」と「S2 persona不在時のBLOCKED」を分離する。新規implementation defectはruntime未実測のため起票draftなし。
- independent healthcare review: S1-owned最終状態、S2境界、Phase 1–3のBLOCKED、患者状態保全、観測/判断分離はPASS。clinic/audit provenanceと製品コード非改変はBLOCKED。臨床証跡の初期permissionがdirectory `0755` / file `0644`だったHIGH指摘はdirectory `0700` / file `0600`へ修正した。CRITICALな患者状態破壊・clinic跨ぎaccessは観測されていない。
- changed files: repo内writer変更は本`FE-refactor.md`追記のみ。commit、push、migration、reset、codegen、direct DB writeは未実施。

#### 2026-07-28 M-04 実測結果

- status: **COMPLETE**。証跡は`tmp/fe12-m04-evidence/2026-07-28/`（a11y `6`件、screenshot `5`件、network `5`件、全file `0600` / directory `0700`）。saved prompt validatorはexit `0`、`Prompt Craft Harness Validation: PASS`、`Execution contract: dynamic-workflow/v1`。
- persona/fixture: staff `38->[10]`（hospitalization view-only）・staff `41->[1]`（CRUD全許可）、両者`is_system_admin=false`。H=`1`は`pet_id=1000018,status=admitted,cage_id=3,doctor_id=1`、daily/vital/care log/staff note=`1`、care plan=`1/2`をread-backした。
- view-only: `/hospitalization`のList Viewと`/hospitalization/1`は閲覧可能で、pet・care plan・vital・care log・staff noteを逐語取得。mutation controlはdisabled表示ではなく大半がDOMから非表示で、ページlevel cueは`閲覧のみ`。31行のcontrol matrixは空欄`0`、view-only試行区間のmethod位置一致による非GET requestは`0`件。
- Board View補足: staff `38->[10]`だけでは`GET /api/v1/masters/cages`がHTTP `403`となったため、独立review後に既存group `9`の実rule（hospitalization=view-only、master-hospitalization=view-only）をlive read-backし、staff `38->[10,9]`へ一時変更して再実測した。cage card描画後もH=`1`はrole/draggable/tabindex/click handlerなしの`cursor-default`で、click・cage `3→4` pointer drag・Enter・Spaceの前後network差分は非GET `0`件。終了時はstaff `38->[10]`へ復帰した。
- 全権限対照: staff `41`でalive boardのoccupied cardがaccessible draggable buttonとして提示され、detailでは`退院処理`、`入院情報の編集`、care-plan追加/編集/削除、vital/care log/staff note追加が提示された。保存・削除・退院確定・board drag/openは押していない。
- 権限剥奪case: staff `39`のvital dialogを開いてからgroup `11`を`[]`へ剥奪し、`保存` callbackを1回試行。`POST /hospitalizations/1/daily-records/2026-07-27/vitals`はHTTP `403`で拒否された。終了時はstaff `39->[11]`へ復帰し、`38->[10]`・`40->[12]`・`41->[1]`もread-backした。
- 死亡遷移: pet `1000018`は`alive → PATCH /death 204 → deceased(deceased_at=2026-07-28T00:00:00+09:00) → boardで「死亡」表示・click/drag不可・非GET 0件 → DELETE /death 204 → alive`。最終read-backは`status=alive,deceased_at=null`。
- H=`1`前後突合: `pet_id` `owner_id` `hospitalization_type` `start_date` `end_date` `status` `cage_id` `doctor_id` `memo`の9業務fieldは全て一致し、`updated_at`も不変。
- viewport: 1440×900は異常なし、1200×800はsidebarのicon collapseのみで異常なし。800×1024はheaderの退院予定cardがclip、500×900はtitle・patient metadata・`プラン管理・詳細` tabがclip。死亡board 1440×900は死亡文言・cage gridとも可読。
- 意図的非実行: 退院（会計あり/なし）はM-05 fixtureを不可逆に壊すため提示観測のみ。成功するchild mutation、care-plan削除、入院編集保存も実行していない。
- static security finding: FEの退院control/callbackはhospitalization `delete`を要求する一方、backendの退院routeは`edit`を要求する。仕様裁定と整合は別unitの対象とし、本実測では修正していない。
- orchestration: native Workflow toolは未提供のためmulti-agent fan-outを使用。`/root/m04_planner`、`/root/m04_controls`、`/root/m04_permission`、`/root/m04_death_contract`、`/root/m04_ledger_harness`をread-onlyでjoinし、mainだけがwriterを担当。死亡遷移中は全agentをquiescentにした。収集後はfreshな`/root/m04_healthcare_review`と`/root/m04_evidence_review`が不足していたBoard runtime試行とraw network証跡をHIGH/MEDIUMとして検出し、修復後に再reviewした。
- changed files: `FE-refactor.md`と`tmp/fe12-m04-evidence/**`のみ。本executorが`backend/**`/`frontend/**`へ発行したedit commandは`0`。commit、push、migration、reset、codegen、seed差し替え、direct DB writeは未実施。

### M-05 Clinical sentinel responsive

- Route: Commit `657c1a49cd2c37dc63f5af8e530258a36a12d81e` に記録した25 routeのうち、clinical sentinelを表示する`/medical-records`系、`/hospitalization`系、`/examinations`系、`/vaccinations`系、`/checkups`系、`/`、`/owners`。
- Fixture seed source: `003_demo`の`pets.csv`、`medical_records.csv`、`hospitalizations.csv`、`exams.csv`/`exam_results.csv`、`vaccinations.csv`、`checkups.csv`を基に、S01/S02/S03/S05のfixtureを合成してdeath、danger=`高`、HIGH、LOW、past、today、future、emptyを各1件以上用意する。
- Fixture実査: D=`2026-07-27`、共有alive pet=`1001002`。draft M2=`1425547`（record_no=`MR-20260727-1-ru3bB8`）配下のcheckup `1/2/3/4`とtop-level vaccination `1/2/3/4`をAPIで作成し、各seriesの`M05-past/today/future/empty`と`next_date=[2026-07-26,2026-07-27,2026-07-28,NULL]`をHTTP 200でread-backした。
- 追加すべき具体値: vaccination `1/2/3/4`は`pet_id=1001002,vaccine_id=1,date=2026-07-20,next_schedule_type=[other,other,other,NULL]`でlabelを`remarks`へ保存した。checkup `1/2/3/4`はM2=`1425547`配下の`pet_id=1001002,checkup_type_id=1,date=2026-07-20`でlabelを`result`へ保存した。全作成はHTTP 201。
- Fixture対応: death=`1001005`、danger high=`1001004`、normal候補=`M-02 E 1014562 / item 2325052`、HIGH候補=`item 2325053`、LOW候補=`item 2325054`、past/today/future/empty=`vaccination 1/2/3/4 + checkup 1/2/3/4`、hospitalization=`M-04 H 1`、RBAC persona=`staff 38/39/40`。M-02の3 itemは基準値不在のため全て未評価normalで、HIGH/LOW cueは未成立。
- 着手ブロッカー: M-01、M-03、M-04、M-05固有fixtureとM-02 M/E/items作成は完了。合成fixture set全体で残るのは、M-02の正規reference range不在によりHIGH/LOWが導出されない点だけである。
- 次の一手: master-data責任者のM-02基準値裁定後、実測担当が上記IDでfixture-to-cue表を固定してrunbookを実行する。**一時的な死亡/high状態とM-03 group/staffはR-4（2026-07-28）で復旧・撤去済み。**
- Persona: 対象resourceの`view=true`を持つ通常担当者。mutation確認が必要なrowだけ該当action権限ありpersonaを併用する。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: 各fixtureを一覧、選択、登録、編集、詳細で開き、文言、badge、日付、disabled/hidden control、keyboard focus順を確認する。期限は同じ実測日にpast/today/futureを並べる。
- Expected result: death/danger/HIGH/LOW/期限超過が非色cueを持ち、normalとtoday/futureを誤ってdangerにしない。死亡操作はpositive matchで拒否され、全viewportでcue/controlのwrap、clip、overlapなし。Hospitalization Boardの死亡は`3b7524748`以降「死亡」テキストを持つ。
- Required evidence artifacts: route×4 viewport screenshot、accessible name/text dump、computed token、console/network HAR、fixture-to-cue対応表。

#### 2026-07-28 M-05 実測結果

- status: **BLOCKED（観測完了・臨床cue証拠とresponsive未達のため）**。本unitはsnapshot観測のみで、production codeを修正できない。全28 PNGは取得済みだが、`/` 500×900、`/hospitalization` 800×1024・500×900、`/examinations` 500×900でclip/overflowを確認した。高危険 `pet 1001004` の非色cueは対象detailを含めてruntime証拠を取得できず、患者安全上PASSにしない。検査未評価3項目は許可されたdetail route `/examinations/1014562` で各 `未判定` を確認した（`tmp/fe12-m05-evidence/2026-07-28/detail-examination-1014562.txt:75-93`）。
- 実行日: `Tue Jul 28 11:57:55 JST 2026`。fixture基準日 `2026-07-28` とのずれはない。CDP gate: `curl .../json/version` HTTP `200`、MCP `list_pages`成功、CDP probe `gate-cdp-probe.png` は `PNG image data, 1440 x 900`。
- persona gate: `GET /api/v1/me` は staff `42`、`is_system_admin=false`、`main_clinic_id=1`、`permissions` は対象view true/create-edit-delete false。`GET /api/v1/masters/staffs/42/permission-groups` は `{"group_ids":[9]}`。
- fixture before/after: pets `1001002=alive/low`、`1001004=alive/high/danger_reason=FE12-M01 fixture`、`1001005=deceased/deceased_at=2026-07-27T00:00:00+09:00`、`1000018=alive`、H=`1` admitted/cage `3`、exam `1014562` items `2325052/3/4` は `is_assessed=false,is_abnormal=false,status=normal`、vaccination/checkup `1/2/3/4` は past/today/future/empty。`cmp fixtures-before.txt fixtures-after.txt` は exit `0`。
- evidence: `tmp/fe12-m05-evidence/2026-07-28/` に a11y `7`件、route screenshot `28`件、network `7`件、fixture before/after、detail a11y、`fixture-to-cue.md`、`layout-review.md`、`computed-tokens.txt`、`completion-report.md`（後述）を保存する。directory `0700`、files `0600`へ整える。件数実測は a11y `7`、PNG `28`、PNG判定 `28`、network `7`、fixture-to-cue `UNREPORTED=0`。
- fixture-to-cue: `fixture-to-cue.md` に5 fixture行×7 routeの全35セルを記録し、空欄および`UNREPORTED`は0件。death `1001005` は owner detailで `チロ`・`死亡`（`detail-owner-300588.txt:157-163`）を確認したが、7一覧routeでは表示されない。hospitalizationは `入院`・`チビチビ`・`犬用ケージ（小）`、aliveであり死亡個体なし。
- deadline判定: vaccinationは `2026-07-27` のみ `（期限超過）`（`a11y-vaccinations.txt:61-86`）。`2026-07-28` todayと`2026-07-29` futureは日付のみ。checkupは `M05-past` のみ `期限切れ`、today/futureは `期限間近`、emptyは表示なし（`a11y-checkups.txt:65-97`）。today/futureを期限超過/danger扱いする誤検出は0件。
- view access: 7 routeのa11y treeで `アクセス権限がありません` の件数は全て0。`閲覧のみ`は各routeで確認でき、static guard根拠は `frontend/src/app/routes/clinical-general-routes.tsx:14-42`、`frontend/src/app/routes/clinical-care-routes.tsx:17-21,72-78,203-207,259-263,315-319`。medical-records/examinationsは一覧mainが空の初回取得をwait-repairし、route screenshotは正しいrouteをCDP target pathで再取得した。
- network/mutation: routeごとの観測区間で method位置一致 `POST|PATCH|PUT|DELETE` は全7ファイル `0 件`。fixture作成は本unitの観測区間外であり、M-05中に成功するwrite・死亡登録/解除・保存・削除を発行していない。
- layout 28件: 詳細は `layout-review.md`。PASS 20件、BLOCKED 8件。BLOCKED全件は `/` 800×1024（見出しwrap）、`/` 500×900（見出し・新規予約登録ボタン）、`/medical-records` 800×1024・500×900（右端列）、`/hospitalization` 800×1024（board右側3列目）、`/examinations` 800×1024・500×900（右側列）、`/checkups` 500×900（ペット名列右端）。
- clinical review: deathの文字cueはPASS。unassessedはdetailで`未判定`を確認し、HIGH/LOWは基準値0行のため成立しない。高危険 `1001004` はAPI状態のみで、非色の`危険理由を表示` accessible nameを対応付けられずBLOCKED。これは患者安全上、normal扱いのPASSへ置き換えない。
- failure signatures: (1) 初回 snapshotがlazy content前で4 routeのmain空 → route heading wait後に7 route再取得。 (2) CDP helperが`/json`の先頭pageを選びPNG route誤対応 → target pathname選択へ修正し28枚再取得。各修復後に件数・画像形式を再確認した。3回目の同一失敗はない。
- orchestration: native Workflow toolはsessionに無かったため、multi-agent fan-outを使用。preflight `preflight-routes`（019fa6a2-f89d-7693-a7a5-d2d89e904600）、`preflight-fixtures`（019fa6a2-f8f0-7be3-b36d-7ab629961cdc）、`preflight-harness`（019fa6a2-f94c-7621-991b-5c18cb6293eb）、`preflight-review-plan`（019fa6a2-f9ad-72c3-911b-aebf18dcd64c）はread-only join済み。post `post-healthcare-review`（019fa6a8-1223-7b71-acb3-6779f6d74957）、`post-evidence-reconcile`（019fa6a8-127b-7ba2-8a29-c5f846bf15be）、`post-layout-review`（019fa6a8-12d6-7720-94b4-02d5d6bfc90f）はread-only join済み。mainのみが`FE-refactor.md`と`tmp/fe12-m05-evidence/**`を書いた。全agentは完了後close済み。レビューの採用判断は、高危険cue未証明・未評価detail確認・layout clipを本記録へ反映、UNREPORTEDのPASS化は不採用。
- harness: Chosen `eval`、backing `~/.agents/skills/verification-loop/SKILL.md`、補助 `~/.agents/skills/e2e-testing/SKILL.md` と `.agents/skills/scoped-verification-gates/SKILL.md` を実読。browser-test skillはHaiku必須のため実行経路には使わず、CDP/MCP手順で代替した。saved prompt validatorは exit `0`、`Prompt Craft Harness: PASS`、`Execution contract: dynamic-workflow/v1`。
- write scope: 本executorのrepo内write対象は `FE-refactor.md` と `tmp/fe12-m05-evidence/**` のみ。`backend/**` / `frontend/**`へのedit commandは0回。commit、push、PR、migration、DB reset、seed差替、codegenは未実施。
- remaining risk: 高危険 `1001004` の非色cueがruntimeで対応付けられていないこと、4 viewportのclipが未修正であること。次工程は `frontend/**`の変更承認後に別unitで行う。M-05完了はR-4/M-01-D/R-2/line-reserve着手をauthorizeしない。

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

2026-07-27にS1（backend）/S2（frontend）の2レーン並行で全9項目を解消・清算した。実装・裁定が完了した項目は原則として本節から削除し、証跡commitはファイル冒頭に一覧する。以下は実装ではなく**裁定**を待っている項目である（2026-07-28 に F16 の解釈を追加）。R-1のみ、本ユニットの完了記録として残す。

- **【裁定済・2026-07-28・曽我】F16 の「グレーアウト」は badge のみで合格** — 現行実装（`OwnersListTable.tsx:236-241` → `getPetStatusColor` → `status-helpers.ts:176-180` が死亡へ `BADGE.grayHover`）を是とする。行全体のグレーアウトは求めない。根拠は「死亡ペットの情報自体は正常に読めるべきであり、行全体を落とすと可読性を下げる」。**M-01-E の合否基準は「死亡行のステータスバッジがグレーであること」に確定した。** 行全体のグレーアウト不在は実装未達ではない。

- **【裁定済・2026-07-28・曽我】R-3 は選択肢 (B) — 基準値不在のまま M-02 を実測する** — `exam_reference_ranges` の0行は解消を待たない。M-02 の主要な問い（一覧HIGH/LOW cueの要否）と 4 viewport layout は基準値と独立に答えられるため。**M-02 の裁定メモには「異常判定機能が停止した状態で観測した」旨を必ず記載する。** 基準値マスタの恒久是正は `3-session-agent.html#BUG-449` を正本として別途進める。以下は判断の材料として保持する。

- **【背景・R-3】M-02の`exam_reference_ranges`基準値が存在しない** — 2026-07-27 のfixture作成で判明。異常判定は`examination_service.go`が`referenceRanges.ResolveByFieldIDs(ctx, clinicID, animalSpeciesID, fieldIDs)`で引いた基準値から導出する（`exam_reference_range_repository.go`の述語は`clinic_id + animal_species_id + exam_type_field_id`）。**現行seedにrange recordが無く、rangeのread/write APIも存在しない。** `exam_type_fields.normal_value`（`6.0-17.0 x10^3/uL`等）は表示用文字列であって導出には使われない。結果としてRBC=`9.0`（基準5.5-8.5超過）・HCT=`30.0`（基準37-55未満）が`is_assessed=false, is_abnormal=false, status=normal`のまま返る。**この`normal`は「正常」ではなく「未評価fallback」であり**（`exam_result_assessment.go`の`unassessedExamResult()`）、臨床的な正常判定と読み違えてはならない。判断者: master-data責任者。選択肢: (A) 正規の基準値を投入して同じE/itemsを再評価しHIGH/LOW cueを成立させる（**2026-07-28 更新: #249 U4 で投入APIとmaster画面が実装されたため、実装待ちではなくなった。残るのは獣医師による動物種別の臨床値の投入だけである**）、(B) 基準値不在のままM-02を実測しcue無しの状態で裁定する（M-02の裁定内容が「cueの要否」から「導線の妥当性」へ変質する）。**Aを選ぶ場合、投入は正規の別工程で行う。fixture作成unitでの値の偽装は禁止であり、実際に行っていない。**

- **【完了・P2・R-1】staff入力検証の契約drift 5件** — ①emailは入力時のみ形式検証、②既存staffの非空passwordはbackend同等契約で検証、③backendのアカウントなしstaff許容へfrontendとOpenAPIを統一、④空氏名test名を実際の編集経路へ限定、⑤関連付けmutation直前の権限再検査を実装した。①②と③frontend側=`166e4acd7`、③OpenAPI側と④=本ユニット、⑤=`a44fa0ebe`。**残余risk・判断待ち・未完了項目なし。**

- **【要起票・本ledgerの所掌外】fixture作成中に実測した実装drift** — production code を変更せず報告に留めた。正式な起票先は`3-session-agent.html#ledger`。①**staff作成POSTが`reservation_visible:false`を無視する** — **真因を特定済み。上記「GORM `default:true`」項へ統合した**（staff固有ではなくmodel全体で39箇所の機構欠陥）。本項では重複記載しない。②**DBOrTx inventory gateが赤** — `exam_reference_range_repository.go`の`ResolveByFieldIDs`/`FindAnimalSpeciesID`が`persistence.DBOrTx`参加者として未登録（`docker compose exec backend go test ./internal/lintscan/ -run DBOrTx`で再現）。pet側2件は`0eecddb11`で清算済みだが、この2件は#249 U3を実装したセッションの所掌。ゲートの要求どおりambient-tx参加を実証するtestを添えて登録する必要がある。

- **【測定完了・2026-07-28】R-2 tygo pointer mapping 15行 — 全行が寄与0である。削除してよい**

  `make codegen` を実行せずに決着した。**生成物そのものが証拠になる**からである。mappingが効いているなら出力に `| null` が現れるはずだが、5種すべてが `?:`（optional）で出力されている。

  | mapping | 宣言値 | 実際の生成結果 |
  |---|---|---|
  | `*uint64: "number \| null"` | — | `medical_record_id?: number /* uint64 */` |
  | `*string: "string \| null"` | — | `other_reason?: string` |
  | `*bool: "boolean \| null"` | — | `value_bool?: boolean` |
  | `*time.Time: "string \| null"` | — | `completed_at?: string` |
  | `*float64: "number \| null"` | — | `min_value?: number /* float64 */` |

  対照として **mappingに存在しない `*int64`** を見ると `price?: number /* int64 */` であり、**mapping済みの5種と生成形が完全に同一**である。つまり tygo は `*T` キーを一切参照せず、ポインタを deref して**基底型のmapping**（`uint64`→`number`、`time.Time`→`string`）を適用し、`?` で optional を表現している。`*time.Time` が `string | null` でなく `string` になっている点が決定的で、これは `time.Time: "string"` の基底mappingが効いた証拠である。

  **生成物の鮮度も確認済み**: `models.ts` の最終生成は `44ce538a8`（2026-07-27）で、`*T` mapping の導入（`f57289fb4` 2026-07-11 / `dad69bc6a` 2026-07-24）および `tygo.yaml` の最終変更（`0bafc2770` 2026-07-25）より新しい。**古い生成物を見て「効いていない」と誤断定したのではない。**

  → **`backend/tygo.yaml` の3 package × 5行 = 15行を削除してよい。** 残件は削除の実行と、`make codegen` で生成物差分が0であることの確認（USER専権）。判断者の確認事項は残っていない。

- **【起票済 BUG-456・HIGH・2026-07-28 M-02発・生成セッションが根本原因まで確定】カルテの検査タブが入力済み検査値を1つも表示しない**

  本ledgerの所掌外。**起票先は `3-session-agent.html#ledger`。** M-02 の実測で「カルテ側の結果値欄が空」と観測され、生成セッションがブラウザと API の両方で追って**原因を特定した。**

  **症状**: `/medical-records/1425546` の「検査」タブで、`検査結果一覧` の表が `項目名` と `判定` だけを描画し、**`結果値`・`単位`・`基準値` の3列がすべて空**になる。獣医師が入力した WBC=10.0 / RBC=9.0 / HCT=30.0 が1つも出ない。screenshot: `tmp/m02-triage-tab.png`。

  **API は正しく返している**（`GET /api/v1/examinations/1014562/items` の実測）:

  ```
  inspection_value  '10.0'                  ← 入力値はここ
  normal_value      '6.0-17.0 x10^3/uL'     ← 基準値表示文字列はここ
  result            ''                       ← 常に空
  reference_value   ''                       ← 常に空
  unit              ''
  ```

  **根本原因**: `frontend/src/features/medical-records/components/ExaminationGroup.tsx` が `結果値` に `{item.result}`（`:102`）、`基準値` に `{item.referenceValue}`（`:108`）を描画している。一方 mapper `frontend/src/lib/transforms/examination.ts:15-33` は `result ← item.result` / `referenceValue ← item.reference_value` と写しており、**入力値が入る `inspection_value` → `inspectionValue` と、基準値文字列が入る `normal_value` → `normalValue` は写されているのに描画されていない。**

  **さらに `result` は書き込み経路が存在しない。** `backend/internal/medicalrecord/examination_request.go:151-159` の `upsertExamItemRequest` は `InspectionValue` / `NormalValue` / `Unit` / `ReferenceValue` を持つが **`Result` フィールドを持たない**。つまり通常の検査登録 API からは `result` を永続化できず、**この列は構造的に常に空である。**

  **修正方針**: `ExaminationGroup.tsx` の描画を `item.inspectionValue` / `item.normalValue` へ向ける。ただし `result` / `reference_value` を使う別経路（lab_import 等）が存在するなら、どちらを正とするか、あるいは fallback（`result || inspectionValue`）にするかを決める必要がある。**`3-session-agent.html#ledger` 起票時に write 経路の全数調査を含めること。**

  **検出できなかった理由**: 生成型 `models.ts` は GORM model 由来であり、`result` も `inspection_value` も両方フィールドとして存在する。したがって**型検査でも lint でも捕まらない。** 「データが出ない」系はこのクラスであり、実画面を開くまで見えない。

- **【起票済 BUG-457・HIGH・2026-07-28 M-04で実測】退院の認可が FE と BE で食い違う** — 本ledgerの所掌外。**起票先は `3-session-agent.html#ledger`。** frontend の退院の表示制御と callback は hospitalization の **`delete`** 権限を使うが、backend の退院 route は **`edit`** を要求する。したがって「`delete` は持つが `edit` は持たない」persona には退院ボタンが見えて押せるが backend が拒否し、逆に「`edit` は持つが `delete` は持たない」persona は実行できるのにボタンが出ない。**どちらの action を正とするかは製品判断であり、決めた側へ FE/BE を揃える必要がある。** M-04 の independent healthcare reviewer が pre-existing HIGH として指摘し、production code 変更禁止の制約下で報告に留めた。退院は会計を伴う不可逆操作であり、認可の二重定義は放置できない。

- **【起票済 BUG-458・MEDIUM・2026-07-28 M-04で実測】入院詳細のレスポンシブ崩れ 2件** — `800×1024` で予定日ヘッダカードが縦長に潰れて clip、`500×900` でヘッダタイトル・患者メタ情報・Plan タブが clip する。証跡は `tmp/fe12-m04-evidence/2026-07-28/shot-detail-800x1024.png` / `shot-detail-500x900.png`。`1440×900` と `1200×800` は異常なし。**起票先は `3-session-agent.html#ledger`。** 本ledgerの所掌外だが M-05 の横断 layout 実測と重なるため、M-05 完了後にまとめて是正するのが安い。

- **【裁定不要・2026-07-28 M-04で実測】Board View の cage 403 は権限設計どおりで defect ではない** — 入院 Board は cage master を読むため `master-hospitalization` の view 権限を要求する。M-03 用に作った group `10` は `hospitalization` view しか持たないため `GET /api/v1/masters/cages` が 403 になる。**これは fixture の権限設計が Board の要件を満たしていなかっただけで、実装の欠陥ではない。** M-04 は既存 group `9`（hospitalization view + master-hospitalization view）を一時的に付与して runtime 観測を完了し、staff `38` を group `10` へ復元した。起票不要。

- **【起票済 BUG-455・CRITICAL・2026-07-28実証】GORM `default:true` により bool の `false` を新規作成できない — systemic**

  本ledgerの所掌外だが、M-03 fixture 作成中に実証したため取りこぼし防止として記録する。**起票先は `3-session-agent.html#ledger`。**

  当初は「staff作成POSTが `reservation_visible:false` を無視する」という個別driftとして記録していたが、**staff固有ではなく機構由来である**ことを特定した。

  - **機構**: GORM は `default` タグを持つfieldの zero value を INSERT 文から除外し、DB default を適用させる。`bool` の zero value は `false` なので、**`false` を明示指定しても INSERT に載らず DB default の `true` が入る**。HTTP DTO も service も `false` を正しく伝搬しているのに、永続化の直前で消える。
  - **PATCH は通る**（`Updates` map 経由のため）。したがって「作成時だけ落ちる」挙動になり、発見が遅れる。
  - **実証（2026-07-28）**: `POST /api/v1/masters/permission-groups` へ `{"is_active":false, ...}` を送信 → response `is_active=true`、read-back も `true`。probe用に作成したgroup `14` は `DELETE` (204) → `GET` 404 で撤去済み。**production codeは一切変更していない。**
  - **影響範囲**: `backend/internal/model/*.go` に `bool` + `gorm:"default:true"` の field が **39箇所・26ファイル**。`*bool` で回避しているものは **0件**。内訳は `is_active` 29件、`reservation_visible` 2件、`accounting_document_show_*` 6件、`show_no_staff_option` 1件、`is_combinable` 1件。
  - **到達可能性**: create request DTO が当該fieldを受けている経路が実在する（`auth/http_permission.go:132`、`trimming/trimming_option_request.go:6`、`trimming/trimming_course_request.go:6`、`reservation/reservation_type_group_request.go:7` 等）。**「無効」を指定して作ったマスタが有効な状態で作られる。** 権限グループでは「無効化した状態で作った権限グループが有効で作成される」ことを意味し、安全境界に触れる。
  - **修正方針は2択**: (A) 該当fieldを `*bool` へ変える（GORMは非nilなら zero value も書く）。(B) 作成時に `Select` で当該列を明示する。**(A)を推す。** (B)は新しい作成経路を書くたびに再発し、lintでも型検査でも捕まらない。

### 維持する裁定（再提案を防ぐため保持）

- **カルテ同日重複にDB unique制約を採らない（2026-07-27）** — 同一pet同日に手で複数カルテを作ることは正当な業務（別々の来院）であり、制約は正当な操作まで禁止する。塞ぐべきは自動生成経路が同じ1回の来院に対して二重に作ることだけであり、これは`5e5868549`のtry-advisory-lockで自動生成経路に限定して解決済み。
- **auto-createにclock seamを導入しない（2026-07-27）** — 重複チェック日は`reservation.StartTime`由来であり現在時刻を参照しない。clock seamを入れると過去/未来予約の検索日が実行日へ変わり挙動を壊す。予約日基準の現行contractが正である。
- **manual chunkの追加分割投資を行わない（2026-07-27）** — 実測522.71 kB（gzip 145.80 kB）で500 kB警告に該当するが、`operations-routes.tsx`のlazy境界により独立chunkとして正しく分割済みで、`/manual`を開いた利用者だけが取得する。build警告は汎用閾値であって業務上の問題の証拠ではなく、表示遅延の申告も無い。存在する問題の証拠なしに最適化するのはproduct-philosophy①違反である。警告閾値の引き上げによる黙らせも行わない（サイレント化は⑤の禁止事項）。再開条件=manual画面の表示時間に関する具体的な業務上の申告が出た場合のみ。

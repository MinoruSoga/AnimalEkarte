# FE-refactor — FE12-02 active-only ledger

> 更新: 2026-07-28（要件責任者: 曽我）
> 業務目的: 未完了の臨床sentinel・RBAC安全境界を、決裁と実測の証跡が揃うまで追跡する。
> 本ファイルは使い捨てのactive ledger。恒久規約は `DESIGN.md`、`docs/spec/design-system.md`、`docs/spec/ui-design-compliance.md`、`frontend/CLAUDE.md` を正本とする。
> 解消済みの証跡: F16/U10/O-A実装=`3b7524748`、R-A/M-A=`e7c978ec9`、P-A=`99bac632e`、exam parent+items原子化=`24929e83d`、死亡API 409 fail-closed=`45b681866`、test-hardening 3件=`3c993420a`。代理決裁6件の記録と調査パックは`a500d424c`および本更新直前までの履歴に保存する。
> S1/S2並行レーン（2026-07-27完走・全項目解消）の証跡: pet死亡CAS封鎖=`18d307076`、カルテlookup JST正規化=`305b50c7f`、ワクチン期限JST契約=`56d18eb18`、auto-create原子化=`5e5868549`、DBOrTx inventory債務=`0eecddb11`、reception danger sentinel=`082be9961`、master permission 45/45配線+fail-closed化=`da550a84d`、auth独立chunk化+checkup resource整合=`8a00a4794`、StaffSettings test追随=`e8a4db982`。各項目の分析・裁定・証跡の全文は本節削除直前の版（commit `0fd1f7d18`）に保存する。裁定の要点2件は後段「維持する裁定」に残す。
> 削除した `## FE12-02 unit execution record` と過去のFE/R-F履歴はCommit `657c1a49cd2c37dc63f5af8e530258a36a12d81e` に保存する。
> **DB reset経路の復旧（2026-07-27・fixture作成の前提工程）**: M-01〜M-05のfixture作成は5走連続でBLOCKEDし、真因は「7/17のmigration統合以降ずっとDB resetが完遂できない状態だった」ことだった。seed CSVがテーブル定義に追随しておらず`COPY`が22P04で落ちていた（`COPY`は列リストを渡さずテーブル定義順に依存する）。証跡: pets/billing_items CSV是正=`0b51c891c`、clinical_plans/exam_results CSV是正=`a4946053b`、migration統合第3回（002-004→001）+clinical_plans列順是正=`edbb162f7`、ERD v31.30（pet_owners/複合FK/テーブル数110）+波及4文書=`2c13b563c`、fixture完成の台帳反映=`976d717c7`。**この破壊はDB resetを実行した瞬間にしか顕在化しないため、7/17から誰も気づいていなかった。** 再発防止のgate2本は **2026-07-28 に実装完了**（`TASK-446`/`TASK-447` として起票 → `backend/internal/lintscan/seed_csv_schema_drift_test.go` と `erd_table_count_drift_test.go` を新設 → 台帳から両sectionを削除済み）。Gate Aはseed CSVヘッダを「`001_init.sql` + 番号順の後続migration」から導いた最終列順と照合し、扱えない列順変更DDLに遭遇したら黙殺せず失敗する。Gate Bは`CREATE TABLE`の実テーブル数（重複定義とコメントを除いた110）とERD宣言値を照合する。両gateとも壊れた入力で確実に失敗することを実証するテストを同梱し、`go test ./...`経由でCIに載る。**コミット済み=`43419cc12`**（`docker-compose.yml`のdocs read-only mount追加を含む）。

## Active scope and authority

- 追跡対象は `M-01`〜`M-05` の実測、line-reserve font実機確認、残余2件（R-2/R-3）、実測後のfixture復旧（R-4）とする。R-1・F16・F9・U10・MEDIUM 4件およびS1/S2並行レーン全9項目は裁定・実装済みで追跡を終了した。
- **2026-07-28 時点の到達点**: M-01〜M-05 の fixture は全て作成済み。**PO裁定を待って着手できないのは M-02（R-3経由）・M-01-D の2件のみ**で、R-1は全5件完了、M-03・M-04・M-01-E は曽我の判断を要さず着手できる（判定基準が客観的に確定しているため）。実行順と干渉回避は「実測レーン分割」節を正本とする。
- **2026-07-28 更新 — R-3 の前提が変わった**: #249 U4（`afd8404a4` API＋`1adf55b6e` UI）で `exam_reference_ranges` への投入経路が開通し、R-3 の選択肢 (A) は「rangeのAPIが無いため実装が要る」という制約から解放された。**ただし `exam_reference_ranges` は依然0行**であり、獣医師が動物種ごとの臨床値を投入するまで HIGH/LOW cue は導出されない。したがって **M-02 を「基準値不在のまま実測する」(B) の妥当性は変わらない**。判断が要るのは「臨床値の投入を待つか、待たずに実測するか」だけになった（機構の有無は論点から消えた）。
- **2026-07-28 更新 — 未評価/正常の誤読対策は3 surface全てで完了**: `ExamItemsTable`（`19bbdbae2`）・`ExamPivotTable`・`ExaminationGroup` の全てで `isAssessed === false` を「未判定」として区別表示する。あわせて `BUG-450`（`GET /v1/examinations` のDTOに明細が無く、カルテ検査結果一覧が常に0件・飼主レポートの異常件数が常に0だった silent contract break）を `include_items` の opt-in 追加で解消した。**M-02 の実測時、カルテ側の検査結果一覧は初めて実データを描画する**（従来は空だったため、過去の実測結果があれば無効）。
- **2026-07-28 更新 — S1レーン（M-01-E + M-04）を試行し BLOCKED。「着手可」は環境前提が揃っている場合の話だった**: 実行記録は M-04 節末尾の「2026-07-28 S1レーン実測結果」を正本とする。臨床状態は一切動いておらず、fixture の生存も未確認のままである。**この試行で判明した4件のうち3件は runbook 側ではなく着手プロンプト側の欠陥**であり、次に同じ轍を踏まないよう下記「S1レーン試行の結果」へ分離して記録する。M-03（S2レーン）は未試行だが、ブラウザ経路の前提は共有するため同じ理由で止まる可能性が高い。
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
| M-03 | P0 | RBAC非活性の理由/name（S2レーン） | **未試行だがS1と同じ環境前提を共有する**（Chrome 9222・browser executor）。fixture・persona完成済、判定基準は客観（禁止personaのmutation 0件）で曽我不要 | persona×routeのa11y tree・4 viewport・HAR・0 mutation集計 |
| M-04 | P0 | Hospitalization child control実効性（S1レーン）— **2026-07-28 BLOCKED** | **環境4件**（下記「S1レーン試行の結果」）。fixture完成済だが生存未確認。判定基準は客観だが、**権限剥奪caseは剥奪可能なpersonaが未定義**で成立しない | 操作別HARと0件集計・a11y tree・4 viewport |
| M-01-E | P0 | OwnersList操作範囲の**証跡収集**（S1レーン）— **2026-07-28 BLOCKED** | **環境4件**（同上）。加えて**F16の合否基準が未確定**（badge のみか行全体か・下記「裁定待ち」） | 4 viewport screenshot・a11y tree・accessible name・GET以外のHAR |
| M-01-D | P0 | 同・**操作可否の裁定** | **曽我**。M-01-Eの証跡が先行すると判断が速い | 死亡ペットに対する許可/禁止を操作単位で明示した裁定メモ |
| M-05 | P0 | Clinical sentinel responsive — 着手可だが**直列尾部** | S1/S2の完了。横断snapshot観測のため、他レーンが状態を動かしている間は証跡が無効になる | route×4 viewport・a11y dump・fixture-to-cue対応表 |
| M-02 | P0 | Examinations一覧意味とlayout | **R-3の裁定が先行**（基準値不在ではHIGH/LOW cueが出ず実測が成立しない） | 両surface×4 viewport・曽我の一覧cue要否裁定 |
| R-3 | P0 | **M-02をどう実測するか**（基準値の恒久是正は`3-session-agent.html#BUG-449`へ移管） | **曽我の裁定**。「基準値不在のまま実測」を推奨（提案節参照） | 実測方針の明示決定と、裁定メモへの「判定機能停止下での観測」注記 |
| line-reserve | P1 | font実機確認 | QA環境管理者と端末管理者の受け渡し（実機3台・remote inspection）。**曽我の判断は不要** | 3実機のscreenshot・HAR・computed font-family |
| R-4 | P1 | 実測後のfixture復旧（`1001005`をalive、`1001004`をlowへ、staff 38-40とgroup 10-13を削除） | M-01〜M-05の実測完了が先行。判断は不要 | 復旧後のread-back証跡 |
| R-1 | P2 | staff入力検証の契約drift 5件 — **2026-07-28 完了** | ①②と③frontend側は`166e4acd7`、③OpenAPI側と④は本ユニット、⑤は`a44fa0ebe`で完了。判断ブロッカー・残件なし | backend binding・OpenAPI required/description・frontend validation/test名・mutation直前の権限再検査が一致 |
| R-2 | P2 | tygo pointer mapping 15行の寄与測定 | `make codegen`がUSER専権。**曽我の判断は不要** | 寄与0の行の特定と設定整理 |
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

fixtureは全て揃っており、PO裁定を待つのは **M-02（R-3経由）・M-01-D の2件だけ**である。R-1は完了済みで、残りは曽我の判断を要さず、実行する人手さえあれば着手できる。ただし全レーンが同一のlocal環境を共有するため、素直な2レーン並行は組めない。所有resourceの重なりを実査した結果を下に示す。

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

## S1レーン試行の結果（2026-07-28）— 再開前に潰す4件

M-01-E + M-04 の試行は preflight で停止した。実行記録の全文は M-04 節末尾を正本とする。ここには**再開に必要な決定と、繰り返してはならない設計上の誤り**だけを残す。

**環境側（曽我さんの操作・判断が要る）**

- **① 実測環境をどう扱うか（未決定）**: disposable stack は存在しない。M-04 は pet `1000018` を実際に死亡させて戻す。選択肢は (A) この dev stack で実施してよい（死亡登録→解除を1回ずつ、確実に復帰させる）／(B) 別の使い捨て DB を用意する／(C) 死亡遷移を外し view-only の mutation 0件検証だけに絞る。
- **② ブラウザ経路（未起動）**: Chrome DevTools MCP は `http://127.0.0.1:9222` を見るが、7/28 時点で応答しない。リモートデバッグを有効にした Chrome の起動が要る。`curl http://127.0.0.1:9222/json/version` が 200 を返すことが着手条件。

**着手プロンプト側の欠陥（次に書く人が繰り返さないための記録）**

- **③ `browser-test` skill を Codex 向けプロンプトの backing に指名してはならない**: `.agents/skills/browser-test/SKILL.md` は冒頭で「**必須: Haiku Agent で実行せよ**」と規定する。Codex セッションに Haiku 経路は無い。**スキル名だけで指名せず SKILL.md の中身を読むこと。** 収集経路は Chrome DevTools MCP を直接使う形で書き、browser-test skill は参照に留める。
- **④ 権限剥奪caseに剥奪可能な persona を残していなかった**: M-04 の「dialog を開いた後に権限を剥奪」は権限グループの mutation を要するのに、S2 の所有物（staff `38`-`40` / group `10`-`13`）を login 以外禁止としたため、剥奪対象が1つも無くなった。選択肢は (A) S1 専用の persona と group を1組増やす／(B) M-03 非稼働を確認して group `10`-`13` の1つを借りる／(C) この case を別ユニットへ送る。**checklist と constraint の突合を保存前に1回行うこと。**

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

### M-03 RBAC非活性の理由/name

- Route: `/settings/permission-groups`、`/medical-records/:id`、`/hospitalization/:id/edit`、`/vaccinations/:id`、`/examinations/:id`。
- Fixture seed source: `backend/migrations/seeds/003_demo/permission_groups.csv`、`permission_group_rules.csv`、`staffs.csv`、`staff_permission_groups.csv`と各featureの既存record CSV。`docs/ops/testing/scenarios/V03-owner-pet-staff-forms.md`で試験用group/staffを作り、`V01-clinical-forms.md`の既存recordを使う。
- Fixture実査: permission groupはVIEW=`10`、CREATE=`11`、EDIT=`12`、TARGET=`13`、専用staffはVIEW=`38`、CREATE=`39`、EDIT=`40`。per-staff read-backは`38->[10]`、`39->[11]`、`40->[12]`で、3 emailのloginは全てHTTP 200。TARGETは操作対象の器としてruleを持たない。
- 追加すべき具体値: `master-permission`、`medical-records`、`hospitalization`、`vaccinations`、`examinations`へ共通で、VIEW group=`{view:true,create:false,edit:false,delete:false}`、CREATE group=`{view:true,create:true,edit:false,delete:false}`、EDIT group=`{view:true,create:false,edit:true,delete:false}`をread-back済み。loginは`fe12-m03-view@noavet.jp`、`fe12-m03-create@noavet.jp`、`fe12-m03-edit@noavet.jp`とpassword=`Fe12pass1`を使う。
- 着手ブロッカー: 4 group・3 staff・割当・persona loginのfixtureブロッカーは解消済み。staff作成bodyの`reservation_visible:false`が201 responseで`true`となるdriftは、各staffへ正規PATCHを行い最終read-backを`false`へ揃えた。production codeは変更していない。
- 次の一手: 実測担当が3 personaで再ログイン/再読込してrunbookを実行する。完了後の別cleanup工程で3 staffと4 groupを削除する。
- Persona: (1) view-only=`view:true/create:false/edit:false/delete:false`、(2) create-only=`view:true/create:true/edit:false/delete:false`、(3) edit-without-delete=`view:true/create:false/edit:true/delete:false`。
- 注記: 検査登録（POST）は`24929e83d`以降create+editの合成認可であり、create-only personaは拒否されるのが正しい期待値である。
- Viewports: 1440×900、1200×800、800×1024、500×900。
- Interaction steps: 各personaでrouteを再ログイン/再読込して開き、pointer、Tab/Enter/Space、formのprogrammatic submit、保存中の権限剥奪後callbackを試す。permission-groupは新規panel、既存panel、reorder、保存後rulesも個別確認する。
- Expected result: view accessは維持する。許可されたactionだけ実行でき、禁止controlはaccessible nameと理由を保持する。禁止personaからのPOST/PATCH/PUT/DELETEは0件で、same-commit剥奪後もmutationしない。
- Required evidence artifacts: persona×routeのaccessibility tree、4 viewport screenshot、network HAR、console log、action別permission matrixと0 mutation集計。

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

### M-05 Clinical sentinel responsive

- Route: Commit `657c1a49cd2c37dc63f5af8e530258a36a12d81e` に記録した25 routeのうち、clinical sentinelを表示する`/medical-records`系、`/hospitalization`系、`/examinations`系、`/vaccinations`系、`/checkups`系、`/`、`/owners`。
- Fixture seed source: `003_demo`の`pets.csv`、`medical_records.csv`、`hospitalizations.csv`、`exams.csv`/`exam_results.csv`、`vaccinations.csv`、`checkups.csv`を基に、S01/S02/S03/S05のfixtureを合成してdeath、danger=`高`、HIGH、LOW、past、today、future、emptyを各1件以上用意する。
- Fixture実査: D=`2026-07-27`、共有alive pet=`1001002`。draft M2=`1425547`（record_no=`MR-20260727-1-ru3bB8`）配下のcheckup `1/2/3/4`とtop-level vaccination `1/2/3/4`をAPIで作成し、各seriesの`M05-past/today/future/empty`と`next_date=[2026-07-26,2026-07-27,2026-07-28,NULL]`をHTTP 200でread-backした。
- 追加すべき具体値: vaccination `1/2/3/4`は`pet_id=1001002,vaccine_id=1,date=2026-07-20,next_schedule_type=[other,other,other,NULL]`でlabelを`remarks`へ保存した。checkup `1/2/3/4`はM2=`1425547`配下の`pet_id=1001002,checkup_type_id=1,date=2026-07-20`でlabelを`result`へ保存した。全作成はHTTP 201。
- Fixture対応: death=`1001005`、danger high=`1001004`、normal候補=`M-02 E 1014562 / item 2325052`、HIGH候補=`item 2325053`、LOW候補=`item 2325054`、past/today/future/empty=`vaccination 1/2/3/4 + checkup 1/2/3/4`、hospitalization=`M-04 H 1`、RBAC persona=`staff 38/39/40`。M-02の3 itemは基準値不在のため全て未評価normalで、HIGH/LOW cueは未成立。
- 着手ブロッカー: M-01、M-03、M-04、M-05固有fixtureとM-02 M/E/items作成は完了。合成fixture set全体で残るのは、M-02の正規reference range不在によりHIGH/LOWが導出されない点だけである。
- 次の一手: master-data責任者のM-02基準値裁定後、実測担当が上記IDでfixture-to-cue表を固定してrunbookを実行する。完了後、一時的な死亡/high状態とM-03 group/staffを別cleanup工程で復旧・削除する。
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

2026-07-27にS1（backend）/S2（frontend）の2レーン並行で全9項目を解消・清算した。実装・裁定が完了した項目は原則として本節から削除し、証跡commitはファイル冒頭に一覧する。以下は実装ではなく**裁定**を待っている項目である（2026-07-28 に F16 の解釈を追加）。R-1のみ、本ユニットの完了記録として残す。

- **【裁定待ち・P0・2026-07-28追加】F16 の「グレーアウト」は badge のみか、行全体か** — M-01-E の合否基準が確定しない。実装は死亡ステータス**バッジ**だけをグレーにする（`OwnersListTable.tsx:236-241` → `getPetStatusColor` → `status-helpers.ts:176-180` が死亡へ `BADGE.grayHover`）。本ledgerの M-01「Expected result」は「一覧はグレーアウト維持・`3b7524748`」としか書いておらず、行全体を意図したのか badge を指したのかが読み取れない。**現状の実装を是とするなら M-01-E は「badge がグレー」を合格とし、行全体を意図していたなら実装が未達である。** 判断者: 曽我。この裁定が無いと M-01-E の証跡を集めても合否を付けられない。

- **【裁定待ち・P0・R-3】M-02の`exam_reference_ranges`基準値が存在しない** — 2026-07-27 のfixture作成で判明。異常判定は`examination_service.go`が`referenceRanges.ResolveByFieldIDs(ctx, clinicID, animalSpeciesID, fieldIDs)`で引いた基準値から導出する（`exam_reference_range_repository.go`の述語は`clinic_id + animal_species_id + exam_type_field_id`）。**現行seedにrange recordが無く、rangeのread/write APIも存在しない。** `exam_type_fields.normal_value`（`6.0-17.0 x10^3/uL`等）は表示用文字列であって導出には使われない。結果としてRBC=`9.0`（基準5.5-8.5超過）・HCT=`30.0`（基準37-55未満）が`is_assessed=false, is_abnormal=false, status=normal`のまま返る。**この`normal`は「正常」ではなく「未評価fallback」であり**（`exam_result_assessment.go`の`unassessedExamResult()`）、臨床的な正常判定と読み違えてはならない。判断者: master-data責任者。選択肢: (A) 正規の基準値を投入して同じE/itemsを再評価しHIGH/LOW cueを成立させる（**2026-07-28 更新: #249 U4 で投入APIとmaster画面が実装されたため、実装待ちではなくなった。残るのは獣医師による動物種別の臨床値の投入だけである**）、(B) 基準値不在のままM-02を実測しcue無しの状態で裁定する（M-02の裁定内容が「cueの要否」から「導線の妥当性」へ変質する）。**Aを選ぶ場合、投入は正規の別工程で行う。fixture作成unitでの値の偽装は禁止であり、実際に行っていない。**

- **【完了・P2・R-1】staff入力検証の契約drift 5件** — ①emailは入力時のみ形式検証、②既存staffの非空passwordはbackend同等契約で検証、③backendのアカウントなしstaff許容へfrontendとOpenAPIを統一、④空氏名test名を実際の編集経路へ限定、⑤関連付けmutation直前の権限再検査を実装した。①②と③frontend側=`166e4acd7`、③OpenAPI側と④=本ユニット、⑤=`a44fa0ebe`。**残余risk・判断待ち・未完了項目なし。**

- **【要起票・本ledgerの所掌外】fixture作成中に実測した実装drift 2件** — いずれも production code を変更せず報告に留めた。正式な起票先は`3-session-agent.html#ledger`であり、本節は取りこぼし防止の控えである。①**staff作成POSTが`reservation_visible:false`を無視する** — M-03のpersona staff作成で送った`false`が201 responseで`true`になった。各staffへ正規PATCHを打ち最終read-backは`false`へ揃えたが、作成時に指定が落ちる挙動自体は未修正。②**DBOrTx inventory gateが赤** — `exam_reference_range_repository.go`の`ResolveByFieldIDs`/`FindAnimalSpeciesID`が`persistence.DBOrTx`参加者として未登録（`docker compose exec backend go test ./internal/lintscan/ -run DBOrTx`で再現）。pet側2件は`0eecddb11`で清算済みだが、この2件は#249 U3を実装したセッションの所掌。ゲートの要求どおりambient-tx参加を実証するtestを添えて登録する必要がある。

- **【レーン外・USER専権】tygo pointer mapping 15行** — 最初に開く: `backend/tygo.yaml:17-35,46-64,75-93`と3 generated output。確認: `*uint64`、`*string`、`*bool`、`*time.Time`、`*float64`の5 mapping×3 packageが生成物diffへ寄与するか。判断者: backend/frontend type contract owner。手順: 許可された`make codegen`で各mappingの出力寄与を個別記録し、寄与0の行だけを設定整理unitへ渡す。`make codegen`がUSER専権のためエージェントレーンへ割り当てない。

### 維持する裁定（再提案を防ぐため保持）

- **カルテ同日重複にDB unique制約を採らない（2026-07-27）** — 同一pet同日に手で複数カルテを作ることは正当な業務（別々の来院）であり、制約は正当な操作まで禁止する。塞ぐべきは自動生成経路が同じ1回の来院に対して二重に作ることだけであり、これは`5e5868549`のtry-advisory-lockで自動生成経路に限定して解決済み。
- **auto-createにclock seamを導入しない（2026-07-27）** — 重複チェック日は`reservation.StartTime`由来であり現在時刻を参照しない。clock seamを入れると過去/未来予約の検索日が実行日へ変わり挙動を壊す。予約日基準の現行contractが正である。
- **manual chunkの追加分割投資を行わない（2026-07-27）** — 実測522.71 kB（gzip 145.80 kB）で500 kB警告に該当するが、`operations-routes.tsx`のlazy境界により独立chunkとして正しく分割済みで、`/manual`を開いた利用者だけが取得する。build警告は汎用閾値であって業務上の問題の証拠ではなく、表示遅延の申告も無い。存在する問題の証拠なしに最適化するのはproduct-philosophy①違反である。警告閾値の引き上げによる黙らせも行わない（サイレント化は⑤の禁止事項）。再開条件=manual画面の表示時間に関する具体的な業務上の申告が出た場合のみ。

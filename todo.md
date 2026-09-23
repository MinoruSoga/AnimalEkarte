# タスク台帳 — 入口

最終照合: 2026-09-23（JST）。HEAD `936610d75` のコードベースで再照合した。前回（2026-09-22・`cd2feaa14`）以降の着地分: 尿検査M4表示回帰と主訴C0解除UI（`a95835cb9`）、実行シナリオ S14–S33 とフォーム wire-key 総覧（`5a5828a1d`）、handoff CSV 契約 digest 解消と cross-clinic staff 連携（`d34e09df7`・`0c91f5dc9`）、STG 遅延対策の実装と curl 実測（`d96c4ba27`・cities/DO pinning は能力不足で撤回）。製品 FAIL 10件はコード照合で全件「記載どおり未修復」を再確認。作業ツリーには未コミットの3単位がある: 子レコード stale 更新防御（backend・EXCLUSIVE-LOCK）、入院プラン単価転記（FE・MASTER-PATH）、stg-uat-staff-attach RLS bypass。remote・CI・配備・Linear の現在状態は再照会していない。投稿・状態変更は未実施。**未完了の作業だけを掲載する。** 完了の詳細はGit履歴と元のUAT記録を参照する。ローカル準備・修正・回帰追加を、課題全体・STG読取・医院受入・本番実行の完了にしない。

## 着手プランの確認

**回答を受けた4件は、仕様入力待ちを解除済み。現在はコード対応済みの受入と、未対応範囲の調査・検証に分かれる。** 下の実行キューは現行コードに合わせた次工程を示す。実環境の検証・データ移行・本番作業には別の開始条件があるため、それらまで一括READYとはしない。[開発・調査](todo-issue.md)、[検証・受入](todo-verification.md)、[運用・納品](todo-operations.md) が詳細の正本。`todo-performance.md` は性能の技術記録、`bug.md` / `bug-2.md` / 医院報告は根拠・履歴、`docs/ops/backlog-spreadsheet.md` は外部Q&A管理規則。

### 着手判定の読み方

- **READY（調査/設計）**: 対象と成果物が具体化済みで、repo内の調査、失敗ケースの設計、変更案の作成を開始できる。既存準備票を作り直さず、不足する次工程へ進む。runtime実行可・実装済みを意味しない。
- **準備可**: 既存証拠照合や実行条件を揃える作業が可能。十分な成果物があれば完了扱いとし、入力待ちのまま同じ設計票を増やさない。
- **条件付き**: 対象版・環境・fixture・実行者・必要な承認参照・証拠保存先が揃った単位だけ実行できる。現時点で揃ったと確認していないものを READY に昇格しない。
- **BLOCKED / 判断待ち**: 本文の不足入力が必要。準備可の作業は先行できるが、医院の事実、責任者名、承認、現在の外部状態を推測で埋めない。
- **DEFERRED**: 再開条件まで実装しない。既存の調査メモで十分なら新たな作業を起こさない。

担当欄の「開発・QA・運用・PO・医院・producer」は必要な役割で、個人への割当や承認取得を表さない。実行担当は claim 取得時、要件責任者・受入者は実装/受入前に記録する。通常のローカル調査・文書下書きに追加承認は不要。

<a id="ローカル準備票の配置2026-09-21"></a>

### ローカル準備票の配置（2026-09-22）

リンク先の調査/設計票は作成済み。一部コード対応も進んだが、実機受入・DB適用・配備などの現在の実行証拠は今回照合していない。[Issue の5エリア](todo-issue.md#open) は、ローカル調査・検証・設計継続5件 / PO判断待ち9件 / 証拠・環境待ち17件 / 保留3件 / コード対応済み・受入待ち5件。別掲の解決済み7トピックは再登録しない。

| 配置 | 対象 | 備考 |
|---|---|---|
| [todo-campaign-20260918/](docs/work/todo-campaign-20260918/) | UAT-R2 3件、Q2/Q4 集計、処置移行、死亡日訂正、Linear照合 | 2026-09-19 キャンペーン |
| [todo-campaign-20260919-ready17/](docs/work/todo-campaign-20260919-ready17/) | Slack READY 13件の設計票（作成時の名称） | `cd2feaa14` で主訴null hydrate修正、尿検査/接種/プラン導線回帰、遅延条件を追加。現在の開始区分はIssueの5エリアを参照 |
| [linmig-campaign-20260919/](docs/work/linmig-campaign-20260919/) | P1–P5/P8、検査機器、実LINE、migrate、件数、フォント、締め時間、bundle | 実行条件は [運用](todo-operations.md) / [検証](todo-verification.md) |
| [remaining-campaign-20260920/](docs/work/remaining-campaign-20260920/) | P6/P7 と PO/evidence の Slack 14件（計16票） | 判断材料・ケース票。PO裁定と実機/STGの現在の結果は UNKNOWN |
| [todo-ledger-20260922/](docs/work/todo-ledger-20260922/) | EXCLUSIVE-LOCK 子レコード実装記録、GENDER-MAP 対応表とSTG訂正ドラフト、handoff CSV 契約差分、OWNER-DB 分類・実行案 | 2026-09-22 キャンペーン（5票）。進捗は下の該当行と各専用TODOへ反映済み |

### 今開始する4件（仕様入力の回収は完了）

| 順 / ID | 最初に行う作業 | この単位の完了条件 / 後続 |
|---|---|---|
| 1 / UAT-R2-MASTER-PATH | [全経路票](docs/work/todo-campaign-20260918/UAT-R2-MASTER-PATH.md) の複数マスタrequest/model回帰は追加済み。入院参照の下流（プラン単価転記）はコード対応済み・未コミット（named price-loss の解消、回帰は反転固定済み）。次はexam_types配線と予防分類のunit_price維持を合成ケースへ（実行票 S20・docs/ops/testing/scenarios/） | 全12フォームの新規/編集→再読込→利用/会計は未実証。専用fixtureで未カバー箇所を検証し、再現した金額不一致だけ修正 |
| 2 / UAT-R2-CHART-FIT | 高さ制約・タブ内scrollのコード対応済み。[受入キュー](todo-verification.md#uat-followup) で1366×625、全9タブ、sidebar両状態の残ケースを確認 | 必須情報/保存/フォーカスに到達できる対象buildの証拠。Windows 8/Chrome実機の版・CSS領域・100%/既報125%はQAが採取し、最新Chromiumと分ける |
| 3 / UAT-R2-EXCLUSIVE-LOCK | 二重会計409・明細一意制約と、古い合計/確定後明細/検査重複のmock回帰は追加済み。[実施記録](docs/work/todo-ledger-20260922/UAT-R2-EXCLUSIVE-LOCK.md) のとおり子レコード4種別（治療/バイタル/処方/接種）のstale更新CASをbackend実装済み・scoped GREEN（migration 005は未適用）。次はhandler/DTO・OpenAPI・FE生成typesへのversion配線 | 実DB/2セッションは [DB適用証拠](todo-operations.md#billing-schema-readiness) と専用環境待ち。Deleteと一括並べ替えはversion対象外のまま。全面ロックは未採用 |
| 4 / UAT-Q2-TREATMENTS-IMPORT | [全期間移行票](docs/work/todo-campaign-20260918/UAT-Q2-TREATMENTS-IMPORT.md) の列写像は調査済み。未確定FK/分類/日時/用量/価格等を契約案・合成fixture設計へ | 現行21表にはtreatments/prescriptionsなし。全種類・全期間の履歴契約を両repoでレビュー後に実装。実データ投入は別承認 |

上記は4件の現在の次工程。既存の設計票や対応済みコードを作り直さず、依存しない調査は別worktreeで並行可。実装時に同じカルテ/会計ファイルへ触れる単位は直列化する。対象の全列挙・実機値の計測・元列の特定は担当者の作業であり、同じ内容を依頼者へ聞き直さない。個人名を伴う要件/受入担当の参照は製品仕様変更前に実行票へ記録する。

### 追加対応8件の残作業（2026-09-22）

上のMASTER / EXCLUSIVE / IMPORTに加え、[尿検査M4](todo-issue.md#slack-manual-urine) のカルテ表示回帰は `a95835cb9` で着地済み（残は医院承認の項目/凡例/単位/基準・実行票 S28）。[主訴C0](todo-issue.md#slack-complaint) はnull再読込修正・回帰に加え clearable 解除UIまで実装済み（`a95835cb9`・実行票 S27）。残は実機保存・再読込の受入。実装済み回帰は再実装しない。[プラン手入力](todo-issue.md#slack-plan-manual) は案内/導線採否のPO待ち、[ワクチン複数入力](todo-issue.md#slack-vaccine-multi) は元症状・実機証拠待ち、[遅延](todo-issue.md#slack-latency) はSTG curl実測を `todo-performance.md`（E4/E5・未コミット）へ記録済みで、残は実機ブラウザ区間測定の計測設計。受入の残条件は [検証TODO](todo-verification.md#ready8-followup-20260922)。これらの対応は8課題全体の完了を意味しない。

### その他の残件を開始する順序

| 残件 | 次の具体作業 / 終了または停止条件 |
|---|---|
| 性別修正・handoff | [対応表](docs/work/todo-ledger-20260922/UAT-Q3-GENDER-MAP.md) 確定済み: hachioji 1院は統合済みrevisionと一致、jouto/shikishima/hakobuneco の3院は修正前mapping世代でbundle再生成対象。[STG限定訂正ドラフト](docs/work/todo-ledger-20260922/UAT-Q3-GENDER-MAP-CORRECTION-DRAFT.md) 完成・operator承認待ち。handoff CSV契約はdigest bump（`d34e09df7`）でローカルrehearsal BLOCKは解消、formal F6はREHEARSAL_ONLYのまま [BLOCKED](docs/work/todo-ledger-20260922/BUG-LOCAL-HANDOFF-CSV-CONTRACT.md)。PostgreSQL検証receipt取得には未コミットの [stg-uat-staff-attach RLS bypass](backend/cmd/stg-uat-staff-attach/main.go) が前提 |
| ワクチン種・未納・死亡日訂正 | 既存集計/訂正設計は完成済み。[運用担当](todo-operations.md#uat-data-operations) が対象医院・期間・読取範囲/承認・保存先を埋める。条件が揃うまで実データ操作は停止 |
| OWNER / S09 / V04 / clinical / その他UAT | [OWNER実DB 5ケースは過去receipt皆無・全件未実行](docs/work/todo-ledger-20260922/DEV-V-OWNER-DB.md) と分類済み。次は専用disposable DBでの実行（共有DB不可）。S09 / V04 / clinical のexactケース対応づけは [検証表](todo-verification.md#readiness-preparation) どおり継続。起動中の他タスク用コンテナは借用しない |
| P2 / P5 | 差分・remote方式は [LINMIG-232](docs/work/linmig-campaign-20260919/LINMIG-232.md) / [LINMIG-230](docs/work/linmig-campaign-20260919/LINMIG-230.md)。provider有効化・配備・発行は承認まで停止 |
| STGデータ / release / P1・P3・P6・P7 / 認証 | 不足表は linmig / remaining 票へ作成済。[運用表](todo-operations.md#readiness-preparation) の実行条件が揃うまで送信・資格情報変更・本番作業は停止 |
| 性能 | STGのcurl実測は `todo-performance.md`（E4/E5・未コミット）へ記録済み（`d96c4ba27` の認証キャッシュとsleepAfter延長の効果、配置pinningは能力不足で撤回）。残は [測定票](todo-performance.md#測定準備で決めること) での実機ブラウザ区間測定の対象固定。未測定の因果関係で改善を実装しない |
| Linear | [既存照合票](docs/work/todo-campaign-20260918/TODO-V-LINEAR.md) の履歴を保持。接続復旧後に残る既存Issueをfull local ID/元報告/受入条件で照合。投稿・状態変更は別承認 |
| TASK-444-ADDENDUM-CODEGEN | 必要性と別スコープの採用までDEFERRED。準備票を増やすために再開しない |

今回の文書改訂と準備票はタスク実装、テスト実行、外部反映、受入の完了を意味しない。着手時には現行 HEAD と該当 claim・対象入力を再確認する。

## 残作業の入口

| 区分 | 次に行うこと | 正本 |
|---|---|---|
| 実装済み項目の受入 | カルテ高さ・飼主検索・バイタル・MC・撮影の5件を追加。治療 Enter と一覧高さの既存検証とは分け、実機・臨床受入を確認 | [検証キュー](todo-verification.md#uat-followup) |
| 回答済み項目の開発・調査 | 全金額経路の未カバーケース、上書き/二重会計防止の残ケース、全期間処置移行。1366×625 UIは受入へ | [未完了 Issue](todo-issue.md#open) |
| 移行・データ調査 | 性別修正後のbundle、ワクチン種・未納の調査、根拠のある死亡日だけの限定訂正 | [未完了 Issue](todo-issue.md#open) |
| 実装後の確認 | 検索・保険割合・過去カルテ導線の STG ブラウザ確認、全ページ UAT の未確認操作 | [検証 TODO](todo-verification.md#uat-followup) |
| 既存の検証残件 | OWNER 実DB、S09 / V04 / clinical E2E、性能実測、認証 D1 | [検証 TODO](todo-verification.md) |
| 環境・本番・納品 | handoff 再生成、承認後のデータ訂正、Lane 3–4、本番構築・移行・研修 | [運用 TODO](todo-operations.md) |
| deferred | `TASK-444-ADDENDUM-CODEGEN` | [未完了 Issue](todo-issue.md#task-444-addendum-codegen) |

## 管理先と完了の扱い

- 新規の実装・調査・PO 課題は [todo-issue.md](todo-issue.md)。検証・外部作業は各専用 TODO に置き、同じ残件の詳細を複製しない。
- 既存チケットの実行状態は Linear Team **Baritech** / Project **ノア動物病院電子カルテ** / [BRT-4](https://linear.app/baritechllc/issue/BRT-4)。新規 Issue を作らない既存方針を維持する。コメント・状態変更は明示承認後。
- 実装完了と、STG ブラウザ受入・production 配備・go-live を分ける。実装済みの項目は開発キューから削除し、未完了の受入・運用だけを残す。
- 確認元: [医院フィードバック](docs/work/stg-uat-clinic-feedback-q1-q4.md)、[予約・スタッフの UAT 記録](bug.md)、[全ページ UAT 記録](bug-2.md)。元記録の古い計画・当時の状態は現在のキューに読み替えない。

<a id="development-tasks"></a>

## 開発タスク

医院フィードバックのうち **調査・対応・外部証拠照合が残るもの**。原文と回答は [stg-uat-clinic-feedback-q1-q4.md](docs/work/stg-uat-clinic-feedback-q1-q4.md)。詳細・受入は [todo-issue.md](todo-issue.md#open)。コード対応済み（カルテ高さ・飼主検索・バイタル・MC・撮影、および検索 AND・保険 50/70・問診抜粋リンク・治療 Enter・検索一覧高さ）の受入は [検証 TODO](todo-verification.md#uat-followup) に置く。明日以降の予約編集は手順回答済み。

| ID | 内容 | 状態 |
|---|---|---|
| [UAT-R2-MASTER-PATH](todo-issue.md#uat-r2-master-path) | 金額を持つ全マスタの保存・再読込・下流経路を検証 | 価格request/model回帰拡充済み／入院参照下流はコード対応済み（未コミット）／exam_types配線・予防分類と全経路の検証継続 |
| [UAT-R2-EXCLUSIVE-LOCK](todo-issue.md#uat-r2-exclusive-lock) | カルテ上書きと二重会計の両方を防止 | 子レコードstale防御はbackend実装済み（migration 005未適用・FE/handler配線なし）／DB適用・実並行は残る |
| [UAT-Q3-GENDER-MAP](todo-issue.md#uat-q3-gender-map) | コード3/4修正はold_db main統合済み。対応表とSTG訂正ドラフトは作成済み | 3院bundle再生成とoperator承認待ち（bundle・運用） |
| [UAT-Q2-VACCINE-SPECIES](todo-issue.md#uat-q2-vaccine-species) | 猫に犬用ワクチン（Proheart・6種等）。件数調査のあと種を付ける | 集計設計票作成済／STG 調査待ち |
| [UAT-Q4-UNPAID-TRIAGE](todo-issue.md#uat-q4-unpaid-triage) | 未納はデモではない。実未納と突合漏れを集計で切る。一括完了しない | 集計設計票作成済／STG 調査待ち |
| [UAT-Q2-TREATMENTS-IMPORT](todo-issue.md#uat-q2-treatments-import) | 処置マスタと患者ごとの全種類・全期間の履歴を移行 | 列写像調査済み／履歴契約案・fixture設計READY、実装/投入は未実施 |

既存のsource調査と追加回答を各IDへ反映した。READYの範囲と、実行環境・外部承認待ちを分離する。`TASK-444` / `BE-RC-009` / `BE-RC-017` の既存完了は維持する。

<a id="product-bugs"></a>

## 確認済み製品 FAIL

現在の未解消報告は [todo-issue.md](todo-issue.md) の症状別調査・データ課題に整理した。`bug.md` / `bug-2.md` の FIXED / SPEC-OK を未修正の製品 FAIL として再登録しない。環境・fixture 不足は運用、未確認操作は検証に置く。新たに製品欠陥が確定したら、既存 ID のまま再現・原因・担当範囲を確定する。2026-09-23 HEAD `936610d75` のコード照合で、下記10件はいずれも記載どおり未修復を再確認（`cd2feaa14`..HEAD にこれら領域の fix commit なし）。

- **BUG-ACCT-CLOSE-PERM-DEFAULT（OPEN / Medium / billing・permission）**: 既定権限モデルで `cash-register-close:create` が全権限グループ未付与（seed CSV 9グループ・新規 clinic 既定テーブルともに create なし）のため、権限グループ手動編集なしにレジ締め不能。執行への `edit` 付与は消費 endpoint のない dead grant。UAT S08（2026-09-22）で確定。台帳・切り分け詳細は [bug.md](bug.md#bug-acct-close-perm-default)、修正計画は [同](bug.md#plan-bug-acct-close-perm-default)。
- **BUG-S09-FIXTURE-TEARDOWN（OPEN / Low / testing・fixture）**: S09 合成 clinic で締め実行後、`synthetic-closing-fixture teardown` が `cash_register_close_adjustments` の RESTRICT FK と append-only trigger（closes/adjustments とも UPDATE/DELETE 不可）で削除不能。scoped 削除リストに closes/adjustments/holidays がない。UAT S09（2026-09-22）で確定。台帳・切り分け詳細は [bug.md](bug.md#bug-s09-fixture-teardown)、修正計画は [同](bug.md#plan-bug-s09-fixture-teardown)。
- **BUG-AGG-NO-VISIT-REVENUE（OPEN / Medium / aggregation・revenue）**: 完了会計を持つ来院なし飼主が売上ランキングに一切出ない。`filterLTVRows` の `include_no_visit=false` 既定除外（最終来院タブ用の「来院なしを含む」制御）が売上・来院クエリにも適用され、FE の売上タブに回避手段がない。UAT S10（2026-09-22）で確定（飼主 B: 完了会計 ¥3,300・MR なし → `include_no_visit=true` 時のみ出現）。台帳・切り分け詳細は [bug.md](bug.md#bug-agg-no-visit-revenue)、修正計画は [同](bug.md#plan-bug-agg-no-visit-revenue)。
- **BUG-TRIM-KANBAN-IN-CONSULTATION（OPEN / High / trimming・reception）**: 受付済トリミングカードの「トリミングカルテ作成」（UI 表示は「カルテ作成と同時に診療中へ移動」）が `PATCH status=in_consultation` で 409 になる。`validateInConsultationHasMedicalRecord` が `medical_records` 件数を予約区分不問で要求するため、trimming（`appointment_trimming_details` のみ）は永久に受付済滞留。FE 側も既存 appointment への detail 作成 POST に `status` を送らない。UAT S11（2026-09-23）で確定（appointment 1000000006、409 を2回実測）。台帳・切り分け詳細は [bug.md](bug.md#bug-trim-kanban-in-consultation)、修正計画は [同](bug.md#plan-bug-trim-kanban-in-consultation)。
- **BUG-BILLING-UNBILLED-MR-EXCLUSION（OPEN / Medium / billing・unbilled）**: `medical_record_id` 付き pending 会計から明細を soft-delete しても請求元 treatment が未請求候補に復帰しない。`FindUnbilledByPetID` の billing 単位除外（`b.medical_record_id = mr.id AND status != 'cancelled'`）が明細有無に関係なく効き続けるため。会計 cancel でのみ復帰。S11 A3 仕様（削除行は再候補）と矛盾・請求漏れリスク。UAT S11（2026-09-23）で確定（billing 1000000016 / treatment 3）。台帳・切り分け詳細は [bug.md](bug.md#bug-billing-unbilled-mr-exclusion)、修正計画は [同](bug.md#plan-bug-billing-unbilled-mr-exclusion)。
- **BUG-RES-OVERLAP-500（OPEN / Medium / reservation・API）**: 同一スタッフの時間帯部分重複の予約作成が PostgreSQL 排他制約 `excl_appointments_doctor_timerange`（23P01）の 500 として漏れる。完全一致は正しく 409。エラーマッピングに 23P01→Conflict の変換がない。UAT S11 fixture 作成時（2026-09-23）に確定。台帳・切り分け詳細は [bug.md](bug.md#bug-res-overlap-500)、修正計画は [同](bug.md#plan-bug-res-overlap-500)。
- **BUG-LIFF-HEALTHCARD-OWNER-SYNC（OPEN / High / liff・line_customers）**: LIFF トークン連携（`LinkAccount`）は `owners.line_user_id` のみ書き、`line_customers.owner_id` を更新しない。health-card は `line_customers.owner_id` 経由のみで飼主を解決するため、連携済み飼主のヘルスカードが「ペット情報はありません」のまま。`line_customers.owner_id` の書き手はスタッフ手動 `link-owner` のみで、LIFF 連携→ヘルスカード閲覧の導線が断絶。意図的2段ゲートか要裁定。UAT S12（2026-09-23）で確定（fixture 状態で空カードを実測）。台帳・切り分け詳細は [bug.md](bug.md#bug-liff-healthcard-owner-sync)、修正計画は [同](bug.md#plan-bug-liff-healthcard-owner-sync)。
- **BUG-ACCT-INS-SIGN-MISMATCH（OPEN / High / billing・insurance）**: `POST /accountings/complete` で `insurance_amount` の符号規約が FE（負値「マイナスのみ」）と BE（`billing=total−insurance−discount` の正値前提）で不一致。FE 送信値 −1000 で BE は請求額を 3200 と計算し支払内訳一致検証で 400 — UI から保険付き会計を確定不能。直 API で −1000→400 / +1000→201 を実測。UAT S15（2026-09-23）で確定。台帳・切り分け詳細は [bug.md](bug.md#bug-acct-ins-sign-mismatch)、修正計画は [同](bug.md#plan-bug-acct-ins-sign-mismatch)。
- **BUG-ACCT-INS-EDIT-REWRITE（OPEN / Medium / billing・insurance）**: 確定済み会計の修正保存が未変更の `insurance_amount`/`billing_amount` を recalc 値で上書き（観測: −220→0・1980→2200）。`hasInsurance` が `insurance_amount<0` 推定のため保存後に保険表示が消え、BE 正値規約の保存会計も OFF・金額不整合で開く。「変更した項目だけが変わる」違反。UAT S15（2026-09-23）で確定（billing 1000000019/1000000027）。台帳・切り分け詳細は [bug.md](bug.md#bug-acct-ins-edit-rewrite)、修正計画は [同](bug.md#plan-bug-acct-ins-edit-rewrite)。
- **BUG-DIALOG-FOCUS-RESTORE（OPEN / Low / shared UI・a11y）**: 治療プラン検索ダイアログ（`TreatmentSearchDialog`）を Escape で閉じるとフォーカスが呼出元ボタンに戻らず `document.body` に落下。外部 `open` 制御＋`lazy` マウントで Radix FocusScope が復帰対象を捕捉しない疑い。実クリック・キーボード開閉の両経路で再現。S18 手順6 の focus restoration 要件に不合致。同一ダイアログに矢印キー移動なし（Tab 順送りのみ）のギャップも併記。UAT S18（2026-09-23）で確定。台帳・切り分け詳細は [bug.md](bug.md#bug-dialog-focus-restore)、修正計画は [同](bug.md#plan-bug-dialog-focus-restore)。
- **BUG-BILLING-TAX-TYPE-DROPPED（OPEN / High / billing・master）**: マスタ登録の `tax_type`（内税/非課税）が会計明細へ伝播せず、全行 `excluded`・10% で計算・保存される。カルテ連携の未請求候補（`treatmentToUnbilledBillingItem` が `TaxTypeExcluded` をハードコード）と物販マスタ追加（`use-accounting-item-actions.ts` が `tax_type:"excluded"` 固定送信、`get-merchandise-items.ts` が transform で `tax_type` を欠落）の双方で再現。billing 1000000028 の DB 値で確認（内税¥1,100・非課税¥1,000 ×2 が全て外税）。期待合計 ¥5,850 に対し実請求 ¥6,270（税の過剰計上 ¥420）。UAT S20 手順5（2026-09-23）で確定。台帳・切り分け詳細は [bug.md](bug.md#bug-billing-tax-type-dropped)、修正計画は [同](bug.md#plan-bug-billing-tax-type-dropped)。
- **BUG-ACCT-DUP-COMPLETE-500（OPEN / Medium / billing・idempotency）**: 同一カルテへの二重会計確定（別 Idempotency-Key・同一 `medical_record_id`）が意図した 409「このカルテには既に会計があります」に届かず 500。`createCompleteBillingHeader` が UNIQUE 競合後の replay 判定クエリを abort 済み tx 上で実行し 25P02 になるため、`accounting_complete_tx.go` の replay/409 分岐は実環境で到達不能。逐次・真並行の双方で再現（勝者 201・敗者 500）。二重会計行は作られずデータは守られるが、クライアント契約は破損。UAT S21 手順4（2026-09-23）で確定。`UAT-R2-EXCLUSIVE-LOCK` の実DB確認項目に接続。台帳・切り分け詳細は [bug.md](bug.md#bug-acct-dup-complete-500)、修正計画は [同](bug.md#plan-bug-acct-dup-complete-500)。

### UAT 2026-09-23 V01 追加分（証拠: `reports/uat-2026-09-23/V01-clinical-forms.md`、詳細: `bug.md` 末尾「確認済み製品欠陥」）

| ID | severity | 領域 | 症状 | シナリオ |
|:---|:---|:---|:---|:---|
| BUG-MR-DOCTOR-HEADER-STALE | Medium | medical-record / UI | カルテヘッダー担当医が再読込で保存済み doctor_id でなくログインユーザー名を表示 | V01 |
| BUG-VITAL-NOTE-KEY-MISMATCH | Medium | medical-record / vitals | バイタルメモが FE `note` ↔ BE `notes` の key 不一致で保存・表示とも消失 | V01 |
| BUG-MR-VACCINE-FORM-NESTED | High | medical-record / vaccination | カルテ内接種フォームがネスト `<form>` で送信不能（javascript: action が CSP ブロック） | V01 |
| BUG-TRIM-EXCL-TIMERANGE-500 | High | trimming / reservation | 同一担当の90分以内連続トリミング登録が `excl_appointments_doctor_timerange` で 500（409未マップ・UI無音失敗） | V01 |

<a id="human-lane"></a>

## PO / 人間レーン

残る製品判断は [PO判断待ちエリア](todo-issue.md#area-po)。旧Chromeの互換性受入は [カルテ表示の残条件](todo-issue.md#uat-r2-chart-fit)、実環境の承認と操作は [todo-operations.md](todo-operations.md)、受入・go-live は [todo-verification.md](todo-verification.md)。既存の人間ゲートは BRT-4 配下で追跡する。回答済み4件の範囲を未回答へ戻さない。

<a id="refactor-constraints"></a>

## FE 維持制約

- `design-tokens.ts` / `query-keys.ts` / `paths.ts` の表分割や行数だけを目的とした機械的分割をしない。
- `utils/` 再作成・generated/models 一括移行をしない。必要性は [裁定記録](docs/work/development-task-decisions.md#task-444) に従って判断する。
- `app/pages` の合成と owners `loaders.ts` の例外、権限 ref、死亡 sentinel、`useActionState`、queryKey タプルを維持する。
- FE12 却下（manual chunk、死亡行グレーアウト、owners 行アクションをペット生死で止める）を維持する。

着手時は [AGENTS.md](AGENTS.md) の claim・worktree 規則に従う。秘密・患者情報は台帳へ書かない。

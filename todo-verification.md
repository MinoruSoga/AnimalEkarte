# 未完了の検証・受入 TODO

最終照合: 2026-09-18（JST）。完了した実装・unit 検証の列挙を削除し、残る検証と受入だけを扱う。新規実装は [todo-issue.md](todo-issue.md)、外部操作は [todo-operations.md](todo-operations.md)。

着手プラン確認: 2026-09-18。各 ID の入口・前提・手順・証拠に加え、実行前に作れるケース票を下記に具体化した。これは計画の充足確認であり、テストや受入の完了判定ではない。

## 判定原則

- source・静的チェック・unit・CI・配備成功と、ブラウザ UAT・production・go-live を分ける。
- 過去の未実施記録だけから、現在も未実施と断定しない。新しい証拠を照合していないものは UNKNOWN、既知の前提不足は BLOCKED とする。
- 実行前に対象 revision、環境、操作者、承認、fixture、証拠保存先を固定する。秘密・接続文字列・患者情報は台帳や共有ログに含めない。
- データ書込み・migration・配備・外部送信は承認された単位のみ。今回の文書更新ではテスト、DB 操作、ブラウザ UAT を実行していない。

<a id="readiness-preparation"></a>

## 今着手する検証準備

QA/開発が作るケース票の共通列は `ID / case / revision / 環境・fixture参照 / 操作 / 期待値 / 実際値 / 証拠参照 / 後処理 / 判定`。準備時の実際値は「未実行」とし、同じ ID の既存 run に十分な証拠があれば再実行せず対応づける。準備完了は「未実行の行と不足入力が特定できた」状態。ブラウザが使えることだけで共有環境のデータ操作まで許可されたとは扱わない。

| ID | 準備可: 最初に作るケース票・照合表 | 実行へ進む条件 |
|---|---|---|
| UAT-R2-TREATMENT-COMMIT | Enter 1回/2回、Blur、Escape、長押し、IME確定を別ケースにして保存回数と再読込値を記録 | 対象 build、物理 IME 端末、変更可能 fixture、後処理 |
| UAT-R2-MASTER-LIST-HEIGHT | 短い/長い一覧、検索/閉じる、キーボード選択、復帰先フォーカス × viewport/zoom | 実端末の CSS 表示領域と zoom。CHART-FIT の未回答値を推定しない |
| UAT-Q1-SEARCH-AND | 複数語AND/1語/0件/他院候補非表示の期待件数 | STG 配信版と医院別の合成検索 fixture |
| UAT-Q4-INSURANCE-RATES | 新規50/70、既存90/100 × 選択/保存/再読込の期待金額 | 承認済み検証会計・後処理。実請求は使わない |
| UAT-Q2-HISTORY-NAV | 同一ペットの問診行→詳細→戻る、処置未移行の記録 | 対象 build と合成の履歴参照 |
| NOTE2-SWEEP-COVERAGE | `bug-2.md` の未確認 route×操作へ必要 ID・必須値を対応づける | 有効 cage 等の fixture、対象 schema、変更可能範囲 |
| NOTE-STAFF-STARTTIME-RDT | 通常/拡張なし環境で同一操作の stack・発生有無を比較する票 | ユーザー環境の利用。製品由来か拡張由来か未確定の間は修正しない |
| DEV-V-OWNER-DB | 下記5ケースの既存結果を PASS/FAIL/SKIP/未実行に分類 | 専用 disposable DB、cleanup、候補 mount。共有 DB は不可 |
| TODO-V-S09 / QA-UAT-S09-FIXTURE | 下記5時刻の帰属期待表と setup/cleanup 欄 | 起動済み専用 local、helper 条件、合成 identity、clinic 1/2 除外 |
| TODO-V-V04 / QA-UAT-V04-RETEST | V04 のフォーム×C1/C2/C3 を9月13日の証拠へ対応づける | 未収録ケースと disposable clinic・権限 account |
| TODO-V-CLINICAL-E2E / QA-FULL-CLINICAL-E2E | `--clinical` allowlist と DB保存/stub/未実行を分ける | local/CI、APP_ENV=test、専用 fixture/teardown。full job は別定義 |
| TODO-V-STG-DATA | 医院×manifest×Lane3 verify×H3-11×5営業日の証拠一覧 | 運用側の同一入力・対象に結び付く receipt |
| TODO-V-RELEASE / P4 / P8 | P1–P8/E1/E2 の個別結果を close checklist の項目へ対応づける | P4 sign-off と P8 当日 window/判断者/復旧担当。未達は HOLD/No-Go |
| E1 / QA-UAT-LSTEP-REAL | primary 保存・外部タグ・再取得結果・後処理の期待表 | 実 LSTEP write の対象・範囲・復旧・承認 |
| E2 / QA-UAT-LINE-IDTOKEN | 正規idToken、再連携409、無効/期限切れlinkTokenの400系 | 実 LINE の対象・正規 token 取得経路・後処理・承認 |
| TODO-V-LINEAR / META-LINEAR-APPLY | ローカル ID・根拠・残件・更新下書きを既存照合文書へまとめる | 接続復旧後の本文/コメント読取で対応 URL を確定。投稿は別承認 |
| PERF-V-LINEAR | 原因未確定・区間別測定・導入済み観測・受入残の下書き | 直接対応する既存 Issue の読取 |
| AUTH-V-LINEAR-READ / AUTH-V-LINEAR-WRITE | D1の付与/login/mailを分けた下書き。WRITE は READ の後 | 接続復旧、確定 URL、直前再読取、exact 下書きの承認 |
| AUTH-V-D1-PREFLIGHT | 環境別経路・既存admin・staff/主所属・schema・監査・復旧の不足表 | 対象環境・実行者・承認参照の確定 |
| AUTH-V-D1-APPLY | PREFLIGHT→COMMIT→監査receipt→通常loginの確認票 | PREFLIGHT 完了と付与承認。通信断時は再発行せず照合 |
| AUTH-V-D1-MAIL | 送信/受信/リンク利用/再利用拒否/期限切れ拒否/後処理のケース票 | 通常login成功、宛先・回数・受信担当・送信承認 |

`PERF-V-CLIENT-TRACE` / `PERF-V-CF-EVENTS` / `PERF-V-DECIDE-OBSERVATION` / `PERF-V-MITIGATION` / `PERF-V-BUNDLE` / `PERF-V-STG-ACCEPTANCE` は [性能の6単位](#性能タスクの着手順と成果物) がケース票の正本。因果測定前の MITIGATION/BUNDLE 実装は DEFERRED のまま。依頼・承認が必要な行も、ケース票作成はローカルで先行できる。

<a id="uat-followup"></a>

## 直近 UAT の残作業

| ID / 対象 | 残る確認 | 状態・完了条件 |
|---|---|---|
| [UAT-R2-TREATMENT-COMMIT](#uat-r2-treatment-commit) | 実機 IME、Blur、2回 Enter 後の保存・再読込を確認 | ローカル28 tests済み。対象 browser/build の receipt 待ち |
| [UAT-R2-MASTER-LIST-HEIGHT](#uat-r2-master-list-height) | viewport・ズーム・検索/閉じる操作・キーボード選択・フォーカス復帰 | ローカル10 tests済み。対象端末での可視範囲の受入待ち |
| [UAT-Q1-SEARCH-AND](#uat-q1-search-and) | STG で複数語検索・1語検索・医院分離を確認 | ブラウザ未確認。対象 build と結果を記録 |
| [UAT-Q4-INSURANCE-RATES](#uat-q4-insurance-rates) | 新規の保険割合 50/70、既存 90/100 の保持と金額を確認 | ブラウザ未確認。既存値のサイレント丸めなし |
| [UAT-Q2-HISTORY-NAV](#uat-q2-history-nav) | 問診抜粋の行から同一ペットのカルテ詳細へ進めることを確認 | ブラウザ未確認。未移行の処置が空でも詳細を開ける |
| [NOTE2-SWEEP-COVERAGE](bug-2.md#plan-note2-coverage) | 未確認の詳細画面、入院、検査、カルテ・健診の操作を補完 | [全ページ UAT の残範囲](bug-2.md#plan-note2-coverage)。82ページ到達を全 CRUD 完了にしない |
| [NOTE-STAFF-STARTTIME-RDT](bug.md#plan-note-staff-starttime-rdt) | 通常環境と拡張なし環境を比較し、再現時の発生元を確認 | ユーザー環境の確認待ち。製品起因と断定せず、[元の調査](bug.md#plan-note-staff-starttime-rdt) に結果を対応づけ |

Q1 / Q4保険 / Q2履歴の実装は再開しない。根拠は [医院フィードバック](docs/work/stg-uat-clinic-feedback-q1-q4.md) と、9月15日に読取確認した [PR #411](https://github.com/MinoruSoga/AnimalEkarte/pull/411)（merged）、[Backend Deploy](https://github.com/MinoruSoga/AnimalEkarte/actions/runs/34923018516) / [Frontend Deploy](https://github.com/MinoruSoga/AnimalEkarte/actions/runs/34923018544)（ともに success、`d337f016`）。この配備記録はブラウザ確認や本番反映の代替ではない。

`NOTE2-SWEEP-COVERAGE` は `/accounting/:id`、`/hospitalization/:id`、`/hospitalization/:id/edit`、`/inventory/:id` の前提 ID を確保し、有効な cage 等の必須値で入院・検査を確認する。カルテ actor の修正は再実装せず、対象環境の schema を確認した後に新規カルテ→再読込→健診を再検証する。古い「カルテバグでブロック」を現行判定に流用しない。元の `reports/uat-2026-09-13/` を保持し、新しい run の route/action・前提・結果・証拠・未確認理由を残す。

### UAT-R2-TREATMENT-COMMIT

`72807128` の [TreatmentQuantityCell](frontend/src/features/medical-records/components/TreatmentsTab/TreatmentQuantityCell.tsx) は Enter 2回で確定、Blur 保存、Escape 取消を実装済み。repeat / isComposing / keyCode229 の無視を含む28 testsは [既存の同一コミット検証](.planning/agent-fast-campaign/four-candidate-integration-20260916/evidence/rev7-reverify-72807128-codex/controller/RECONCILIATION.md) による。今回の文書更新では再実行していない。残るのは承認された対象 build・fixture での物理 IME と、1回目で未保存→2回目/Blur→再読込で値が残ることのブラウザ確認。実測前に高速化完了とはしない。

### UAT-R2-MASTER-LIST-HEIGHT

同じ検証の10 testsで [TreatmentSearchDialog](frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx) の一覧上限 `max-h-[calc(80vh-12rem)]` を確認済み。残るのは対象端末の viewport・ズーム別の可視行、画面内の検索/閉じる操作、キーボード選択・フォーカス復帰の受入。長い一覧のスクロールは残す。ノートPC125%の記録だけでは解像度・対象タブを確定できない。

### UAT-Q1-SEARCH-AND

承認された STG の対象 build と自医院の検索 fixture を固定し、「飼主名 ペット名」の複数語 AND → 1語 → 0件 → 他医院の候補非表示を確認する。結果は検索語そのものを共有せず、ケース番号・期待件数・表示件数と機密除去した画面証拠を新しい UAT run に記録する。元の [検索仕様](docs/spec/screens/03-owners-list.md) と異なる結果だけを再現付きで開発へ戻す。

### UAT-Q4-INSURANCE-RATES

承認された検証会計を用意し、[会計仕様](docs/spec/screens/11-accounting-detail.md) の新規 50/70 と既存 90/100 を別ケースで確認する。選択 → 金額 → 保存 → 再読込を通し、既存値が変更なしでも丸められないことを照合する。成果物は割合別の期待/実際金額と保存結果。既存の実請求を検証用に変更せず、対象・後処理がなければ開始しない。

### UAT-Q2-HISTORY-NAV

承認済み fixture で問診抜粋の行 → 同一ペットのカルテ詳細 → 戻るを確認し、処置が未移行の記録も詳細へ進めることを照合する。[InterviewHistory](frontend/src/features/medical-records/components/InterviewHistory.tsx) のリンク先と表示対象の対応を証拠にする。成果物は対象 build と行/詳細の対応、未確認理由。これで処置明細の移行完了とはしない。

<a id="development-verification"></a>

## 既存の検証キュー

| ID | 状態 | 次の作業・完了条件 |
|---|---|---|
| [DEV-V-OWNER-DB](#dev-v-owner-db) | UNKNOWN（追加証拠未照合） | 前回は disposable DB URL 未設定。現在の専用 DB と過去実行証拠を確認し、下記の実DBテストの不足だけを実行 |
| [TODO-V-S09](#todo-v-s09--qa-uat-s09-fixture) | BLOCKED（fixture・対象環境待ち） | `QA-UAT-S09-FIXTURE` の #2–#6 を専用 fixture で確認。既存の [UAT 状態](docs/ops/testing/UAT-DOMAIN-STATUS.md) と run の対応を記録 |
| [TODO-V-V04](#todo-v-v04--qa-uat-v04-retest) | UNKNOWN | `QA-UAT-V04-RETEST` と9月13日の master CRUD 証拠を項目単位で対応づけ、削除・後処理・未収録項目を補完 |
| [TODO-V-CLINICAL-E2E](#todo-v-clinical-e2e--qa-full-clinical-e2e) | BLOCKED（実行条件待ち） | `QA-FULL-CLINICAL-E2E` の承認済み test 環境・identity・fixture と full job 証拠を確保 |
| [TODO-V-STG-DATA](#todo-v-stg-data) | UNKNOWN（受入の追加証拠未照合） | H0–H3 / Lane 3–4 の対象入力・医院・件数・金額・画面証拠を確認。配備 success でデータ受入を代用しない |
| [TODO-V-RELEASE](#todo-v-release) | BLOCKED（受入条件未充足） | P1–P8 / E1–E2 の個別 receipt を確認。未解消の臨床安全・会計・分離 FAIL があれば go-live は No-Go |

OWNER の対象は `TestOwnerRepository_UpdateAndFind_ReloadFailureRollsBackUpdate`、`TestOwnerRepository_Update_ClinicIsolation`、`TestOwnerService_Update_DiscountTOCTOU_*`（LockedDiffWithoutPermission を除く）、`TestOwnerRepository_LockByIDForUpdate_RequiresAmbientTransaction`。既存の unit 完了は再登録しない。共有 DB をテスト用にせず、未実行・SKIP は PASS にしない。

9月18日の source 照合で、ワイルドカード部分は `TestOwnerService_Update_DiscountTOCTOU_StaleZeroRejected` と `TestOwnerService_Update_DiscountTOCTOU_NonDiscountFieldStillOK` の2件。実行票にはこの完全名を使い、上記の他3件と合わせて **5件それぞれの実行結果**を記録する。パッケージの exit 0 だけでは充足しない。

### DEV-V-OWNER-DB

1. [owner テスト](backend/internal/owner/) と [testdb helper](backend/internal/testdb/) の接続・schema 作成/cleanup 条件を確認し、現在の専用 disposable DB と過去 receipt を照合する。`TEST_DATABASE_URL` は秘密管理から供給し、接続先の実体が共有 DB でないことを実行者が確認する。
2. 対象 worktree を mount した Docker で `./internal/owner` の上記5ケースだけを `go test -run` に列挙して実行する。TOCTOU は `StaleZeroRejected` と `NonDiscountFieldStillOK` が対象で、`LockedDiffWithoutPermission` は除外する。
3. 成果物は revision、専用 DB の承認参照、各ケースの実行/PASS/FAIL/SKIP、cleanup 結果。接続未設定・全体 exit 0 でもケースが SKIP なら未完了。共有 DB への fallback や自動 migration で補わない。

### TODO-V-S09 / QA-UAT-S09-FIXTURE

1. [S09 fixture 設計](docs/ops/testing/S09-FIXTURE-DESIGN.md) と [対象 spec](frontend/e2e/s09-closing-time-boundaries.spec.ts) を読み、起動済みの専用 local Docker、許可された APP_ENV、合成パスワードの安全な供給、clinic 1/2 を使わない条件を確認する。
2. helper で新規合成 fixture を作成 → S09 #2–#6 の帰属プレビューをブラウザ確認 → cleanup token による teardown の順で実行する。fixture の完了時刻は設計の5時刻を使い、既存会計やシステム時計を変えない。
3. 成果物は秘密除去済みの fixture 参照、ケース別の対象時刻・期待/実際集計、browser report、cleanup 結果。token/password を report に残さず、cleanup 未完了も明記する。共有 STG/PROD には接続しない。

準備時に確定できる期待値は [S09 シナリオ](docs/ops/testing/scenarios/S09-closing-time-boundaries.md) と既存 spec の合成設定（AM開始09:00、境界13:30、平日終了19:00）を使う。時刻は JST。対象日をDとして次の5件を固定し、実医院の締め設定には適用しない。

| 合成会計の完了時刻 | 期待する帰属 |
|---|---|
| D 10:00 | D の午前のみ |
| D 13:30:00 | D の午後。午前へ重複しない |
| D 14:00 | D の午後 |
| D 20:00 | D の緊急 |
| D+1 02:00 | D の緊急。D+1 の緊急へ重複しない |

各スロットの件数は午前1・午後2・緊急2。金額はfixtureに設定した額から独立に計算してケース票へ記入する。

### TODO-V-V04 / QA-UAT-V04-RETEST

1. [V04](docs/ops/testing/scenarios/V04-settings-master-forms.md) の各フォームを9月13日の証拠に対応づけ、未収録の項目だけを選ぶ。対象は一般設定・マスタ・検査機器項目で、LINE/LSTEP の V05 と分ける。
2. disposable clinic と権限別 account を固定し、作成 → validation → 編集 → 再読込 → 未使用行の削除と使用中行の拒否を、該当フォームの C1/C2/C3 に従って確認する。system master の削除を期待しない。
3. 成果物はフォーム×操作の coverage 表、保存/拒否/権限の証拠と後処理結果。以前の DELETE regression だけで全フォームを PASS にせず、未収録・対象外は理由付きで残す。

coverage の母数は V04 本文の標準マスタ16種、診療項目5タブ、薬剤と用量、予約区分、予約枠、締め時間3フォーム、シフト、lab-device、法人invoice。LINE/LSTEPはV05へ残す。現行 [V04 spec](frontend/e2e/v04-settings-master-forms.spec.ts) の4テストは動物種類、主訴、薬剤価格保存、system支払方法削除拒否だけなので、自動 spec の成功を母数全体へ広げない。9月13日の手動証拠を照合できないセルは未収録のままとする。

### TODO-V-CLINICAL-E2E / QA-FULL-CLINICAL-E2E

1. [clinical E2E 設計](docs/ops/testing/CLINICAL-E2E-DESIGN.md) と [runner](frontend/scripts/run-e2e.sh) の allowlist を照合する。承認された起動済み local/CI、`APP_ENV=test`、許可された local base URL、合成 identity、clinic 1/2 除外、teardown を固定する。
2. Docker 内で専用 fixture の setup → runner の `--clinical` → teardown を行う。対象外の auth smoke・全 suite job は分け、full job は別の実行承認と対象定義を満たしてから確認する。
3. 成果物は revision、allowlist と実行ケース数、機密除去済み report、fixture/cleanup の結果。stub で create する spec は DB 保存の証拠に数えない。失敗・未実行・環境違いは後続の全体 PASS にまとめない。

9月18日の runner の `--clinical` は `e2e/` 下の10 spec: `clinical-flows`、`clinical-smoke`、`medical-records-create`、`medical-records-patient-search`、`medical-records-pagination-sort`、`examinations-flow`、`vaccinations-flow`、`checkups-flow`、`hospitalization-flow`、`estimates-flow`（各 `.spec.ts`）。ケース票はこの集合に固定する。[CI workflow](.github/workflows/e2e.yml) の実行対象は `auth-flows.spec.ts` のみで、clinical/full suite job は未配線。`--clinical` の実行証拠と、全suite CIを要求するかの判断・配線作業は別の欄にし、auth smoke成功でどちらも閉じない。

<a id="stg-データレーン"></a>

## STG データレーン

入力受領と操作は [運用 TODO](todo-operations.md#stg-data-lanes)。受入では source manifest・対象医院・件数・金額・監査・画面を照合する。Lane 4 は両院の Lane 3 verify、H3-11、所定の運用日数の証拠が揃うまで完了にしない。9月13日の一部 CRUD 成功をこの受入全体に拡張しない。

### TODO-V-STG-DATA

最初に [H0–Lane 4 の個別計画](todo-operations.md#stg-データレーンの着手プラン) から、医院別の最新 manifest・投入/skip/verify・staff attach・画面・5営業日の証拠を一覧にする。同一入力・同一対象と対応する証拠だけを採用し、不足部分の確認依頼を作る。運用側の実施後に件数・参照・医院分離・金額・監査と画面を再照合し、成果物を医院別の受入表にする。wrapper の exit 0、過去の apply PASS、配備 success のいずれも受入全体の代替にしない。

### TODO-V-RELEASE

最初に [本番・納品 P1/P2/P3/P5/P6/P7](todo-operations.md#本番納品の着手プラン) の receipt と、下表の P4/P8/E1/E2 を対象 revision・環境に対応づける。個別の判定・不足・担当・参照先を [go-live runbook](docs/delivery/GOLIVE_RUNBOOK.md) に集約する。P4 に必要な E1/E2 を省略せず、P8 の当日判定まで完了扱いにしない。

| ID / 対象 | 着手手順 | 成果物・停止条件 |
|---|---|---|
| P4 / #254 AUTHENTICATED-UAT | [close checklist](docs/ops/testing/scenarios/UAT-254-CLOSE-CHECKLIST.md) の5業務フロー、実 LINE/token、DB/audit、残件処理を現在の証拠へ対応づけ、不足操作を承認済み fixture で補完 | 同じ対象版の run 一覧と実施者以外の sign-off。臨床安全・会計・分離・認証・データ消失の未解消 FAIL があれば No-Go |
| E1 / QA-UAT-LSTEP-REAL | [S01 のタグ同期](docs/ops/testing/scenarios/S01-deceased-pet-guard.md) と [V05-17](docs/ops/testing/scenarios/V05-auth-line-forms.md) を、write 有効な承認済み LSTEP 対象・テスト対象者・復旧範囲に限定して確認。設定変更/送信は別承認 | primary 保存と best-effort 同期、外部タグと再取得件数を別々に照合。mock/停止中の204や toast だけで PASS にしない |
| E2 / QA-UAT-LINE-IDTOKEN | [V05 の実 LINE 連携](docs/ops/testing/scenarios/V05-auth-line-forms.md) で正規 idToken による link → 再連携409 → 無効/期限切れ linkToken の400系を確認 | token値/URLを残さないケース別 receipt、二重紐付けなしと後処理。idToken と linkToken を区別し、mock を実 LINE の証拠にしない |
| P8 / #257 GOLIVE | [当日手順](docs/delivery/GOLIVE_RUNBOOK.md) の新 window・判断者・support/rollback owner を記入し、pre-window 全項目 → 当日 import 突合 → smoke → Go/No-Go → 支援へ進む | 判断者の署名、時刻、当日 receipt、復旧判断。window 未確定・前段不足・当日突合未達なら HOLD/No-Go。過去日程を再利用しない |

<a id="linear-reconciliation"></a>

## Linear 照合の残り

9月15日の読取結果: [BRT-4](https://linear.app/baritechllc/issue/BRT-4) は Backlog、[BRT-45](https://linear.app/baritechllc/issue/BRT-45) / [BRT-68](https://linear.app/baritechllc/issue/BRT-68) は Needs Human。これは当時の読取記録。9月18日の Linear MCP 再照会も未接続（`USER_NOT_LOGGED_IN`）で失敗し、現在の状態・対応先は UNKNOWN。完了済みチケットは残件表から除く。

| ID | 状態 | 残作業 |
|---|---|---|
| [TODO-V-LINEAR / META-LINEAR-APPLY](#todo-v-linear--meta-linear-apply) | UNKNOWN（再認証・照合待ち） | 9月13〜17日の残件と既存 Issue の対応を確認し、反映先・URL・現状・更新案を作る |
| [PERF-V-LINEAR](#perf-v-linear) | 対応先未確定 | PERF-STG-LOGIN と通信区間の調査を、既存 Issue に直接対応づける |
| [AUTH-V-LINEAR-READ](#auth-v-linear-read) | 対応先未確定 | D1 対象環境・メール・反映範囲と、既存 Issue 本文を照合 |
| [AUTH-V-LINEAR-WRITE](#auth-v-linear-write) | 承認待ち | 確定した既存 Issue への更新案を明示承認後に反映。新規作成枠を前提にしない |

新規 Issue を作らない方針は維持するが、過去の free issue limit を read-only 照会や既存 Issue 更新の技術的ブロッカーにしない。類似語だけでチケットを割り当てず、不明なら UNKNOWN とする。今回、外部投稿・状態変更は未実施。

### TODO-V-LINEAR / META-LINEAR-APPLY

Team/Project/BRT-4 配下で ID・元報告・実装・受入条件を照会し、既存本文とコメントを読む。検索ヒットだけで直接対応を確定しない。成果物は「ローカル ID → 既存 Issue URL → 読取日時/状態 → 一致根拠 → 本文/コメントの更新案」の対応表で、[既存の照合下書き](docs/work/linear-f1-f6-mapping.md) を利用する。反映は対象 URL と exact な下書きの明示承認後に行い、再読取で反映を確認する。読取不可・対応不明は UNKNOWN のまま残す。

### PERF-V-LINEAR

同じ Team/Project/BRT-4 内で `PERF-STG-LOGIN`、`/login`、`/me`、OPTIONS、Container を手掛かりに既存本文を照合する。通信全体とサーバー区間を分けた計測結果、原因未確定、観測コードのローカル検証済み範囲、STG 受入を別欄にした更新案を作る。直接対応する Issue と日時が確認できれば下書き完了。投稿/Done は別承認で、類似する別案件へ割り当てない。

### AUTH-V-LINEAR-READ

D1 の対象環境・既存スタッフ・通常 login・メールの受入条件を BRT-4 配下の既存本文/コメントと照合し、各残件の直接対応を確認する。成果物は URL/取得日時/現在状態/差分と更新案。合成 DB 成功と対象環境の付与・メールを分け、対象が不明なら UNKNOWN と記録する。

### AUTH-V-LINEAR-WRITE

READ の確定 URL と下書きに対する承認後、更新直前に対象本文・状態を再読取して他者更新を確認する。承認範囲の既存 Issue へ反映 → 再読取 → 更新差分の照合で完了とする。成果物は投稿/更新参照と時刻。競合・拒否・対象変更なら停止し、新規 Issue 作成や未完了 D1/mail の Done 化で代用しない。

<a id="perf-stg-login"></a>

## PERF-STG-LOGIN

技術記録は [todo-performance.md](todo-performance.md)。Worker 観測は現在 tracked code に存在し、未導入 WIP の扱いを終了した。`PERF-V-IMPLEMENT-OBSERVATION` の実 proxy 4 tests・worker typecheck（`index.test.ts` include済み）は `72807128` の [既存検証](.planning/agent-fast-campaign/four-candidate-integration-20260916/evidence/rev7-reverify-72807128-codex/controller/RECONCILIATION.md) で完了し、開いたキューから外した。残るのは遅延の因果測定、常時観測の必要性・出力範囲の再判定、STG 受入である。

| 順 | ID | 状態 | 次の作業・完了条件 |
|---|---|---|---|
| 1 | [PERF-V-CLIENT-TRACE](#性能タスクの着手順と成果物) | 承認・観測条件待ち | `/login` 遷移前から OPTIONS / GET、FCP、操作可能時刻を記録。通常読込と再読込を分ける |
| 2 | [PERF-V-CF-EVENTS](#性能タスクの着手順と成果物) | provider 証拠待ち | 同じ時刻の Worker 受付・forwarding・Container 起動を関連づける。時刻対応できなければ UNKNOWN |
| 3 | [PERF-V-DECIDE-OBSERVATION](#性能タスクの着手順と成果物) | 調査待ち | 実装済み観測の常時出力が必要かを1・2の結果から再判定。必要なら対象を限定する変更、不要なら撤去を別実装単位にする |
| 4 | [PERF-V-MITIGATION](#性能タスクの着手順と成果物) | DEFERRED（因果待ち） | OPTIONS / GET の遅延箇所を特定してから通信・設定変更を選ぶ |
| 5 | [PERF-V-BUNDLE](#性能タスクの着手順と成果物) | DEFERRED（実測待ち） | 固定 revision の転送・parse/execute への寄与を測り、必要な変更だけを判断 |
| 6 | [PERF-V-STG-ACCEPTANCE](#性能タスクの着手順と成果物) | ブラウザ受入待ち | 対象 build の匿名・既存 session・復旧・医院選択を確認し、待機表示・操作可能・認証成功の時刻を分離 |

- `PERF-V-CLIENT-TRACE`: 承認済み対象・時間枠・停止担当・証拠保存先を固定する。相対時刻、method、status、protocol、initiator、OPTIONS/GET 対応、FCP を保存し、Cookie・Authorization・本文・個人情報は含めない。HAR 等は保存前に機密除去。単発値や未使用時間だけで p95/p99・cold start・改善完了と判定しない。
- `PERF-V-CF-EVENTS`: provider 時刻と browser 時刻を対応づけ、Container 起動証拠がない場合は Worker 所要時間だけで起動待ちと断定しない。観測で設定・配備を変更しない。
- 将来の変更時だけ、既存依存を利用し対象候補を mount した隔離 Docker の scoped Vitest と worker typecheck を別々に確認する。依存インストールや今回の再実行は不要。
- `PERF-V-BUNDLE`: cache 条件を分ける。過去の HTML load 約0.31秒だけで約23秒の待ちを bundle 起因としない。公開 entrypoint を壊す deep import や一括 chunk 再編を先行させない。

[STG パフォーマンス測定チェックシート](docs/ops/testing/STG-PERFORMANCE-CHECKLIST.md) に従い、証拠は既存の非公開 run 保存先へ置く。STG traffic・Cloudflare 読取・設定変更・配備・Linear 更新は各対象の承認範囲で行う。

### 性能タスクの着手順と成果物

各 run の保存先は `reports/uat-YYYY-MM-DD/performance/<RUN-ID>/`。チェックシートを複製し、対象 revision・承認範囲・測定条件を先に記入する。Git ignore と機密除去を確認し、測定していない p95/p99 や未合意の SLO を達成済みにしない。

| ID | 最初の作業 → 次の手順 | 成果物・進行条件 |
|---|---|---|
| PERF-V-CLIENT-TRACE | 通常利用を変えず `/login` 遷移前から記録 → OPTIONS/GET の各区間、FCP、操作可能時刻を収集 | 相対時刻表と条件。初回/再読込を分け、秘密除去後に保存 |
| PERF-V-CF-EVENTS | 同一の時刻窓と時刻基準を固定 → 受付/forwarding/Container起動/Go区間を対応づける | client/provider の対応表。相関できない区間は UNKNOWN |
| PERF-V-DECIDE-OBSERVATION | 上記2表で遅延区間を特定 → 現行常時ログが判断に寄与するか評価 | 維持・対象限定・撤去・保留の理由と必要な変更範囲。再現/因果がなければ判断保留 |
| PERF-V-MITIGATION | 因果が分かった区間に対し1変更の仮説と比較条件を作る → 承認された対象で前後測定 | 時間内訳の比較と CORS/CSRF/認証の回帰結果。因果未確定なら通信/設定を変えない |
| PERF-V-BUNDLE | 固定版を cache なし/ありで記録 → 転送/parse/execute が操作可能時刻に占める割合を測る | bundle 寄与と採否。寄与がなければ分割を採用せず、公開 entrypoint を維持 |
| PERF-V-STG-ACCEPTANCE | 対象 run の成功と実配信 revision を照合 → 匿名/既存session/復旧/医院選択を確認 | 待機表示・操作可能・認証成功の3時刻とケース結果。改善候補があれば同条件の前後比較を付ける |

<a id="認証認可の外部境界"></a>

## 認証・認可の外部境界

認証 D1 の対象環境での付与・メール確認をここで一元管理する。契約は [認証設計](docs/architecture/auth.md)、実行手順は [初回管理者手順](docs/ops/deploy/FIRST_SYSTEM_ADMIN.md)、合成検証の参照は [D1 SQL fixture](backend/internal/auth/testdata/first_system_admin.sql)。ローカル実装・合成 fixture の完了を、対象環境への付与やメール経路の成功に読み替えない。

| ID | 状態 | 次の作業・完了条件 |
|---|---|---|
| [AUTH-V-D1-PREFLIGHT](#auth-v-d1-preflight) | BLOCKED | 対象環境・操作者・承認・既存 staff/主所属・影響範囲・rollback を確定 |
| [AUTH-V-D1-APPLY](#auth-v-d1-apply) | BLOCKED（前段待ち） | 承認済み手順で初回管理者を付与し、通常 login の非機密 receipt を取得 |
| [AUTH-V-D1-MAIL](#auth-v-d1-mail) | BLOCKED（前段待ち） | 承認済みの対象と送信・後処理範囲でメール経路を検証 |

Linear の残りは [照合の残り](#linear-reconciliation)。共有 `ekarte_db` / `old-db-postgres` / STG / PROD をテスト DB に使わず、本番付与・メール・migration を自動実行しない。receipt は対象環境、日時、担当、承認、結果、復旧可否のみを既存 runbook へ保存する。外部受入が残れば完了にしない。

### AUTH-V-D1-PREFLIGHT

最初に [初回管理者手順](docs/ops/deploy/FIRST_SYSTEM_ADMIN.md) で対象環境の経路を選ぶ。本番専用手順は無効/削除済みも含め system_admin が0件、既存 staff と有効な主所属があることが条件。操作者、適用済み schema、安全な接続先照合、保守枠、監査、通信断時の確認/復旧を不足表にする。成果物は実行案と非機密承認参照。既存管理者がいる場合は bootstrap を再実行せず復旧/通常の追加経路へ戻す。環境未確定でも入力・手順の照合まで先行できる。

### AUTH-V-D1-APPLY

PREFLIGHT と対象操作の明示承認後、人間が runbook の secure service・repo 外入力・1 transaction の手順を実行する。exit 0、COMMIT、監査付き receipt 1行を確認し、本人の通常 login → 医院選択まで照合する。成果物は保護された account/staff/clinic/audit の対応と、台帳用の非機密結果参照。通信断や COMMIT 不明なら再発行せず read-only receipt で照合し、成功後の取消に account/staff/audit の DELETE を使わない。

### AUTH-V-D1-MAIL

通常 login の確認後、[認証設計](docs/architecture/auth.md) と [V05 のパスワード再設定](docs/ops/testing/scenarios/V05-auth-line-forms.md) から確認するメール経路を選び、承認された宛先・送信回数・受信確認担当・秘密の後処理を固定する。送信 → 受信 → 対象リンクの利用 → 使用済み/期限切れの拒否を確認する。成果物は経路別の非機密 receipt と後処理結果で、メール本文・宛先・token を保存しない。送信不能・想定外宛先・所属不整合は停止し、付与成功でメール成功を代用しない。

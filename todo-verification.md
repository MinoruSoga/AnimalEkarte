# 検証・受入の履歴（現在のタスク状態はPlane）

> Current task state is in Plane. Local task details were migrated 2026-09-23; see [crosswalk](docs/work/plane-md-migration-20260923-receipt.md).

最終タスク移行: 2026-09-23。未完了の検証・受入タスクと状態はPlaneを正本とする。実行記録・検証原則・証拠はローカルで保持する。ローカルケース票の一部は [linmig-campaign-20260919](docs/work/linmig-campaign-20260919/) と [remaining-campaign-20260920](docs/work/remaining-campaign-20260920/) に作成済み。今回 runtime は再実行していない。過去の結果は当時の revision に限定し、追加実装の受入は現在の receipt 未照合として UNKNOWN を維持する。

着手プラン確認: 2026-09-22。前回の [追加実装のキュー](#code-followup-20260921) を維持し、今回8件の追加対応から [残検証・受入](#ready8-followup-20260922) を更新した。既存ケース票を再利用し、追加済みunit/mockと、ローカル未カバー・実DB・実機・医院/POの未確認を分ける。調査票のGREENは当時の記録で、今回の再実行結果ではない。

受入タスクの現行状態・未実施条件はPlaneの移行済み項目を参照する。過去runの証拠と判定原則はこのファイルに残し、他タスクの起動中コンテナや共有DBを検証先に流用しない。

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
| `UAT-R2-TREATMENT-COMMIT` → `EMR-120` | Planeへ移行済み |  |
| `UAT-R2-MASTER-LIST-HEIGHT` → `EMR-121` | Planeへ移行済み |  |
| `UAT-Q1-SEARCH-AND` → `EMR-122` | Planeへ移行済み |  |
| `UAT-Q4-INSURANCE-RATES` → `EMR-123` | Planeへ移行済み |  |
| `UAT-Q2-HISTORY-NAV` → `EMR-124` | Planeへ移行済み |  |
| `NOTE2-SWEEP-COVERAGE` → `EMR-200` | Planeへ移行済み |  |
| `NOTE-STAFF-STARTTIME-RDT` → `EMR-180` | Planeへ移行済み |  |
| `DEV-V-OWNER-DB` → `EMR-125` | Planeへ移行済み |  |
| `TODO-V-S09 / QA-UAT-S09-FIXTURE` → `EMR-126` | Planeへ移行済み |  |
| `TODO-V-V04 / QA-UAT-V04-RETEST` → `EMR-127` | Planeへ移行済み |  |
| `SLACK-CLINICAL-UAT` → `EMR-110`, `TODO-V-CLINICAL-E2E / QA-FULL-CLINICAL-E2E` → `EMR-128` | Planeへ移行済み |  |
| `TODO-V-STG-DATA` → `EMR-129`, `H3-11` → `EMR-145` | Planeへ移行済み |  |
| `TODO-V-RELEASE` → `EMR-130`, `P4 / #254 AUTHENTICATED-UAT` → `EMR-131`, `P8 / #257 GOLIVE` → `EMR-134` | Planeへ移行済み |  |
| `E1 / QA-UAT-LSTEP-REAL` → `EMR-132` | Planeへ移行済み |  |
| `E2 / QA-UAT-LINE-IDTOKEN` → `EMR-133` | Planeへ移行済み |  |
| `TODO-V-LINEAR / META-LINEAR-APPLY` → `EMR-135` | Planeへ移行済み |  |
| `PERF-V-LINEAR` → `EMR-136` | Planeへ移行済み |  |
| `AUTH-V-LINEAR-READ` → `EMR-137`, `AUTH-V-LINEAR-WRITE` → `EMR-138` | Planeへ移行済み |  |
| `AUTH-V-D1-PREFLIGHT` → `EMR-139` | Planeへ移行済み |  |
| `AUTH-V-D1-APPLY` → `EMR-140` | Planeへ移行済み |  |
| `AUTH-V-D1-MAIL` → `EMR-141` | Planeへ移行済み |  |

`PERF-V-CLIENT-TRACE` / `PERF-V-CF-EVENTS` / `PERF-V-DECIDE-OBSERVATION` / `PERF-V-MITIGATION` / `PERF-V-BUNDLE` / `PERF-V-STG-ACCEPTANCE` は [性能の6単位](#性能タスクの着手順と成果物) がケース票の正本。因果測定前の MITIGATION/BUNDLE 実装は DEFERRED のまま。依頼・承認が必要な行も、ケース票作成はローカルで先行できる。

<a id="uat-followup"></a>

## 直近 UAT の残作業

| ID / 対象 | 残る確認 | 状態・完了条件 |
|---|---|---|
| `UAT-R2-TREATMENT-COMMIT` → `EMR-120` | Planeへ移行済み |  |
| `UAT-R2-MASTER-LIST-HEIGHT` → `EMR-121` | Planeへ移行済み |  |
| `UAT-Q1-SEARCH-AND` → `EMR-122` | Planeへ移行済み |  |
| `UAT-Q4-INSURANCE-RATES` → `EMR-123` | Planeへ移行済み |  |
| `UAT-Q2-HISTORY-NAV` → `EMR-124` | Planeへ移行済み |  |
| `NOTE2-SWEEP-COVERAGE` → `EMR-200` | Planeへ移行済み |  |
| `NOTE-STAFF-STARTTIME-RDT` → `EMR-180` | Planeへ移行済み |  |

<a id="code-followup-20260921"></a>

### 追加実装に伴う受入（2026-09-21照合）

以下は同じ Issue ID の検証範囲であり、別の開発チケットではない。コード対応済み5件と、部分対応の STAFF / EXCLUSIVE を分ける。各行の実行前に対象 build・端末/環境・合成 fixture・操作者・操作範囲/承認・後処理・証拠保存先を固定する。新たに確認したのはコードとテスト定義の存在までで、テスト実行・配備・実機成功は今回確認していない。

| ID / 現行コード | 残る確認・次の作業 | 現在の状態 / 完了条件 |
|---|---|---|
| [UAT-R2-CHART-FIT](todo-issue.md#uat-r2-chart-fit) / [タブ高さ制約](frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx#L198) | 1366×625、全9タブ、sidebar両状態、長文/長一覧/ダイアログで必須情報・保存・フォーカス到達を確認。実機Chrome版・CSS領域・100%/既報125%を別記 | コード対応済み・受入 UNKNOWN。対象端末で見切れず操作できる証拠。最新ChromiumだけではWindows 8/旧Chrome受入にならない |
| [SLACK-OWNER-HEIGHT](todo-issue.md#slack-owner-height) / [検索結果scroll](frontend/src/components/shared/OwnerSearchModal/OwnerSearchModal.tsx#L189) | 飼主検索モーダルの候補多数/0件、検索欄、末尾行、閉じる、キーボード選択とフォーカス復帰を確認 | モーダルのコード対応済み・実機受入 UNKNOWN。元報告が飼主/ペット一覧画面なら、そのsurfaceは別途照合 |
| [SLACK-VITALS](todo-issue.md#slack-vitals) / [最新測定値の抽出](frontend/src/features/medical-records/lib/visit-vital-chips.ts) | 同一カルテの取得結果のうちrecorded_atが最新の1行だけを表示し、体温/心拍/呼吸/体重の欠損を古い行から補完しない現行動作を確認。測定なし・患者切替・全9タブ・狭い画面、時刻の保存と表示省略も確認 | 表示コード対応済み・臨床受入 UNKNOWN。現行の表示範囲が医院の期待を満たすか確認し、時刻を非表示にしても保存値を失わない。ヘッダーでの新規入力を実装済みとしない |
| [SLACK-MICROCHIP](todo-issue.md#slack-microchip) / [ヘッダー表示](frontend/src/components/shared/PatientContextHeader/PatientContextHeader.tsx#L151) | 番号有無、長い番号、API再取得、患者切替、1366×625で対象ペットと表示の一致を確認 | コード対応済み・受入 UNKNOWN。空欄/長い値でも操作を隠さず、前患者の番号が残らない証拠 |
| [SLACK-CAMERA](todo-issue.md#slack-camera) / [撮影入力](frontend/src/features/medical-records/components/ImageGalleryFilter.tsx#L118) | 対象端末の撮影→確認/取消→正しいカルテへ保存→再読込。権限拒否・容量/形式・通信失敗も確認 | 撮影入口コード対応済み・実機受入 UNKNOWN。JPEG/PNG/GIFの撮影入力と、PDFも扱う通常アップロードを分ける。capture属性だけでカメラ起動成功としない |
| `SLACK-STAFF-SELECT` → `EMR-116` | Planeへ移行済み |  |
| `UAT-R2-EXCLUSIVE-LOCK` → `EMR-85` | Planeへ移行済み |  |

再現した不一致は同じ Issue ID に戻す。元報告の画面・端末や臨床上の期待値が一致しない場合、既存修正の成功から補外せず、未確認ケースと必要な判断を残す。

<a id="ready8-followup-20260922"></a>

### 追加対応8件の残検証・受入（2026-09-22照合）

基準は `cd2feaa14`。以下は既存IDの残条件で、新規課題や追加済み回帰の再作成ではない。実行前の対象build・専用環境/fixture・操作者・操作範囲/承認・後処理・証拠保存先は上の共通条件に従う。unit/mockは実DB永続化・実端末操作・医院の臨床期待の代替にしない。

| ID | 確認できた追加対応 | 残る確認・完了条件 |
|---|---|---|
| `UAT-R2-MASTER-PATH` → `EMR-84` | Planeへ移行済み |  |
| `UAT-R2-EXCLUSIVE-LOCK` → `EMR-85` | Planeへ移行済み |  |
| `UAT-Q2-TREATMENTS-IMPORT` → `MIG-15` | Planeへ移行済み |  |
| `SLACK-COMPLAINT` → `EMR-87` | Planeへ移行済み |  |
| `SLACK-MANUAL-URINE` → `EMR-86` | Planeへ移行済み |  |
| `SLACK-VACCINE-MULTI` → `EMR-105` | Planeへ移行済み |  |
| `SLACK-PLAN-MANUAL` → `EMR-103` | Planeへ移行済み |  |
| `SLACK-LATENCY` → `EMR-104` | Planeへ移行済み |  |

Q1 / Q4保険 / Q2履歴の実装は再開しない。根拠は [医院フィードバック](docs/work/stg-uat-clinic-feedback-q1-q4.md) と、9月15日に読取確認した [PR #411](https://github.com/MinoruSoga/AnimalEkarte/pull/411)（merged）、[Backend Deploy](https://github.com/MinoruSoga/AnimalEkarte/actions/runs/34923018516) / [Frontend Deploy](https://github.com/MinoruSoga/AnimalEkarte/actions/runs/34923018544)（ともに success、`d337f016`）。この配備記録はブラウザ確認や本番反映の代替ではない。

`NOTE2-SWEEP-COVERAGE` の受入条件はPlane `EMR-200` へ移行済み。

### UAT-R2-TREATMENT-COMMIT

> 移行済み: Plane `EMR-120`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### UAT-R2-MASTER-LIST-HEIGHT

> 移行済み: Plane `EMR-121`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### UAT-Q1-SEARCH-AND

> 移行済み: Plane `EMR-122`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### UAT-Q4-INSURANCE-RATES

> 移行済み: Plane `EMR-123`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### UAT-Q2-HISTORY-NAV

> 移行済み: Plane `EMR-124`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
## 既存の検証キュー

| ID | 状態 | 次の作業・完了条件 |
|---|---|---|
| `DEV-V-OWNER-DB` → `EMR-125` | Planeへ移行済み |  |
| `TODO-V-S09 / QA-UAT-S09-FIXTURE` → `EMR-126` | Planeへ移行済み |  |
| `TODO-V-V04 / QA-UAT-V04-RETEST` → `EMR-127` | Planeへ移行済み |  |
| `SLACK-CLINICAL-UAT` → `EMR-110`, `TODO-V-CLINICAL-E2E / QA-FULL-CLINICAL-E2E` → `EMR-128` | Planeへ移行済み |  |
| `TODO-V-STG-DATA` → `EMR-129` | Planeへ移行済み |  |
| `TODO-V-RELEASE` → `EMR-130`, `E1 / QA-UAT-LSTEP-REAL` → `EMR-132`, `E2 / QA-UAT-LINE-IDTOKEN` → `EMR-133`, `P8 / #257 GOLIVE` → `EMR-134`, `P1 / SEC-SECRETS-5 / #89 / #97` → `EMR-147` | Planeへ移行済み |  |

OWNER の対象は `TestOwnerRepository_UpdateAndFind_ReloadFailureRollsBackUpdate`、`TestOwnerRepository_Update_ClinicIsolation`、`TestOwnerService_Update_DiscountTOCTOU_*`（LockedDiffWithoutPermission を除く）、`TestOwnerRepository_LockByIDForUpdate_RequiresAmbientTransaction`。既存の unit 完了は再登録しない。共有 DB をテスト用にせず、未実行・SKIP は PASS にしない。

9月18日の source 照合で、ワイルドカード部分は `TestOwnerService_Update_DiscountTOCTOU_StaleZeroRejected` と `TestOwnerService_Update_DiscountTOCTOU_NonDiscountFieldStillOK` の2件。実行票にはこの完全名を使い、上記の他3件と合わせて **5件それぞれの実行結果**を記録する。パッケージの exit 0 だけでは充足しない。

### DEV-V-OWNER-DB

> 移行済み: Plane `EMR-125`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### TODO-V-S09 / QA-UAT-S09-FIXTURE

> 移行済み: Plane `EMR-126`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### TODO-V-V04 / QA-UAT-V04-RETEST

> 移行済み: Plane `EMR-127`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### TODO-V-CLINICAL-E2E / QA-FULL-CLINICAL-E2E

> 移行済み: Plane `EMR-128`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
## STG データレーン

入力受領と操作は [運用 TODO](todo-operations.md#stg-data-lanes)。受入では source manifest・対象医院・件数・金額・監査・画面を照合する。Lane 4 は両院の Lane 3 verify、H3-11、所定の運用日数の証拠が揃うまで完了にしない。9月13日の一部 CRUD 成功をこの受入全体に拡張しない。

### TODO-V-STG-DATA

> 移行済み: Plane `EMR-129`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### TODO-V-RELEASE

> 移行済み: Plane `EMR-130`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
## Linear 照合の過去記録（現在の状態はPlane）

9月15日の読取結果（当時）: [BRT-4](https://linear.app/baritechllc/issue/BRT-4) は Backlog、[BRT-45](https://linear.app/baritechllc/issue/BRT-45) / [BRT-68](https://linear.app/baritechllc/issue/BRT-68) は Needs Human。これは当時の読取記録。9月18日の Linear MCP 再照会も未接続（`USER_NOT_LOGGED_IN`）で失敗し、現在の状態・対応先は UNKNOWN。完了済みチケットは残件表から除く。

| ID | 状態 | 残作業 |
|---|---|---|
| `TODO-V-LINEAR / META-LINEAR-APPLY` → `EMR-135` | Planeへ移行済み |  |
| `PERF-V-LINEAR` → `EMR-136` | Planeへ移行済み |  |
| `AUTH-V-LINEAR-READ` → `EMR-137` | Planeへ移行済み |  |
| `AUTH-V-LINEAR-WRITE` → `EMR-138` | Planeへ移行済み |  |

新規 Issue を作らない方針は維持するが、過去の free issue limit を read-only 照会や既存 Issue 更新の技術的ブロッカーにしない。類似語だけでチケットを割り当てず、不明なら UNKNOWN とする。今回、外部投稿・状態変更は未実施。

### TODO-V-LINEAR / META-LINEAR-APPLY

> 移行済み: Plane `EMR-135`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### PERF-V-LINEAR

> 移行済み: Plane `EMR-136`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### AUTH-V-LINEAR-READ

> 移行済み: Plane `EMR-137`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### AUTH-V-LINEAR-WRITE

> 移行済み: Plane `EMR-138`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
## PERF-STG-LOGIN

技術記録は [todo-performance.md](todo-performance.md)。Worker 観測は現在 tracked code に存在し、未導入 WIP の扱いを終了した。`PERF-V-IMPLEMENT-OBSERVATION` の実 proxy 4 tests・worker typecheck（`index.test.ts` include済み）は `72807128` の [既存検証](.planning/agent-fast-campaign/four-candidate-integration-20260916/evidence/rev7-reverify-72807128-codex/controller/RECONCILIATION.md) で完了し、開いたキューから外した。残るのは遅延の因果測定、常時観測の必要性・出力範囲の再判定、STG 受入である。（2026-09-23 E5 追記: 因果は `containerFetch` 区間まで局在、常時観測は KEEP 判定、STG 受入証拠は4ケース取得済。下表の各状態を参照。）

| 順 | ID | 状態 | 次の作業・完了条件 |
|---|---|---|---|
| 1 | [PERF-V-CLIENT-TRACE](#性能タスクの着手順と成果物) | 承認・観測条件待ち → 証拠取得済（2026-09-23 E5）→ post-deploy 再取得済（2026-09-23 E6 browser-pages/stg-acceptance） | `/login` 遷移前から OPTIONS / GET、FCP、操作可能時刻を記録。通常読込と再読込を分ける |
| 2 | [PERF-V-CF-EVENTS](#性能タスクの着手順と成果物) | provider 証拠待ち → 証拠取得済（2026-09-23 E5）→ post-deploy 再取得済（2026-09-23 E6） | 同じ時刻の Worker 受付・forwarding・Container 起動を関連づける。時刻対応できなければ UNKNOWN |
| 3 | [PERF-V-DECIDE-OBSERVATION](#性能タスクの着手順と成果物) | 調査待ち → 判定済 KEEP（2026-09-23 E5）→ post-deploy でも `container_fetch_timing` 出力を確認（2026-09-23 E6、KEEP 判定と整合） | 実装済み観測の常時出力が必要かを1・2の結果から再判定。必要なら対象を限定する変更、不要なら撤去を別実装単位にする |
| `PERF-V-MITIGATION` → `EMR-142` | Planeへ移行済み |  |  |
| `PERF-V-BUNDLE` → `EMR-143` | Planeへ移行済み |  |  |
| 6 | [PERF-V-STG-ACCEPTANCE](#性能タスクの着手順と成果物) | ブラウザ受入待ち → 証拠取得済（2026-09-23 E5、4ケース n=1）→ post-deploy 再取得済（2026-09-23 E6、4ケース n=1） | 対象 build の匿名・既存 session・復旧・医院選択を確認し、待機表示・操作可能・認証成功の時刻を分離 |

2026-09-23 E5 追記（証拠は `reports/perf-e5-residual-20260923/`）: CLIENT-TRACE は `/login` cold/warm 2遷移を記録し OPTIONS 0・FCP 692/88ms を取得（[client-trace](reports/perf-e5-residual-20260923/client-trace/README.md)）。CF-EVENTS は cf-ray 相関で `container_fetch` 866–3848ms 支配・edge+worker 約60–115ms・稼働 instance `maa01`・`scheduling_policy` deployed=`default` vs config=`regional` の乖離を取得（[cf-events](reports/perf-e5-residual-20260923/cf-events/README.md)）。DECIDE-OBSERVATION は KEEP と判定（[obs-decision](reports/perf-e5-residual-20260923/obs-decision/README.md)）。STG-ACCEPTANCE は匿名/既存session/復旧/ログイン後の4ケースを記録し login POST 3882ms・`/v1/me` 1475ms ゲート・OPTIONS 0（[stg-acceptance](reports/perf-e5-residual-20260923/stg-acceptance/README.md)）。測定値は現行 STG 配信版のもので、E5 のコード変更は未配備。PERF-V-LINEAR は対応先未確定のまま。

2026-09-23 E6 追記（post-deploy、証拠は `reports/perf-e5-postdeploy-verify-20260923/`、配信版 worker `15a85636`・コンテナ v70→v71）: DEPLOY-VERIFY は配信版が perf コミット `453be4ecc` を後置することと edge OPTIONS 204（Go ヘッダ無し）を確認し SERVES-NEW-BUILD（[deploy-verify](reports/perf-e5-postdeploy-verify-20260923/deploy-verify/README.md)）。OPTIONS-EDGE は 6 probe で edge 契約を検証（allowlist echo・非 allowlist は ACAO 無し・`_internal` 404・OPTIONS 中央値 ~60ms、[options-edge](reports/perf-e5-postdeploy-verify-20260923/options-edge/README.md)）。WARM-MEASURE は warm 中央値が E5 と ±0.01s 内で同一（[warm-measure](reports/perf-e5-postdeploy-verify-20260923/warm-measure/README.md)）。LOGIN-MEASURE は中央値 1.471s vs E5 ~3.33s（−56%）で bcrypt-skip を方向的に確認（warmth 混在のため単独寄与は UNKNOWN、[login-measure](reports/perf-e5-postdeploy-verify-20260923/login-measure/README.md)）。PLACEMENT は `maa01` 継続・`scheduling_policy` 乖離は未解消・再抽選を推奨（[placement](reports/perf-e5-postdeploy-verify-20260923/placement/README.md)）。CF-EVENTS は `container_fetch` 1001–1988ms・edge+worker ~57–63ms・観測窓内で `sin14` へ再作成（[cf-events](reports/perf-e5-postdeploy-verify-20260923/cf-events/README.md)）。BROWSER-PAGES は実ブラウザで N+1 解消を確認し `accountings?owner_id=` 1.9–2.7s を新規候補として記録（[browser-pages](reports/perf-e5-postdeploy-verify-20260923/browser-pages/README.md)）。STG-ACCEPTANCE は login POST 1451ms（E5 比 −62.6%）・`/v1/me` ゲート 1100.7ms・全ケース OPTIONS 0（[stg-acceptance](reports/perf-e5-postdeploy-verify-20260923/stg-acceptance/README.md)）。未解決の正直な記録: AXIOS-RETRY は配備済みだが観測窓に 503 が無くフィールド未検証（unverified-in-field、失敗ではない）、INSTANCE-TYPE/MITIGATION/BUNDLE はトリガー付き DEFERRED のまま、SLACK-LATENCY はユーザーレーン（`EMR-104`）、PERF-V-LINEAR は Linear MCP 未接続（`USER_NOT_LOGGED_IN`）で BLOCKED のまま。

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
| `PERF-V-MITIGATION` → `EMR-142` | Planeへ移行済み |  |
| `PERF-V-BUNDLE` → `EMR-143` | Planeへ移行済み |  |
| PERF-V-STG-ACCEPTANCE | 対象 run の成功と実配信 revision を照合 → 匿名/既存session/復旧/医院選択を確認 | 待機表示・操作可能・認証成功の3時刻とケース結果。改善候補があれば同条件の前後比較を付ける |

<a id="認証認可の外部境界"></a>

## 認証・認可の外部境界

認証 D1 の対象環境での付与・メール確認をここで一元管理する。契約は [認証設計](docs/architecture/auth.md)、実行手順は [初回管理者手順](docs/ops/deploy/FIRST_SYSTEM_ADMIN.md)、合成検証の参照は [D1 SQL fixture](backend/internal/auth/testdata/first_system_admin.sql)。ローカル実装・合成 fixture の完了を、対象環境への付与やメール経路の成功に読み替えない。

| ID | 状態 | 次の作業・完了条件 |
|---|---|---|
| `AUTH-V-D1-PREFLIGHT` → `EMR-139` | Planeへ移行済み |  |
| `AUTH-V-D1-APPLY` → `EMR-140` | Planeへ移行済み |  |
| `AUTH-V-D1-MAIL` → `EMR-141` | Planeへ移行済み |  |

過去のLinear照合記録は [こちら](#linear-reconciliation)。現在のタスク状態は各Plane項目を参照。共有 `ekarte_db` / `old-db-postgres` / STG / PROD をテスト DB に使わず、本番付与・メール・migration を自動実行しない。receipt は対象環境、日時、担当、承認、結果、復旧可否のみを既存 runbook へ保存する。外部受入が残れば完了にしない。

### AUTH-V-D1-PREFLIGHT

> 移行済み: Plane `EMR-139`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### AUTH-V-D1-APPLY

> 移行済み: Plane `EMR-140`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。
### AUTH-V-D1-MAIL

> 移行済み: Plane `EMR-141`。詳細と現在の状態は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照。

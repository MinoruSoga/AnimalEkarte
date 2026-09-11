# タスク台帳の検証 TODO

作成日: 2026-09-11 JST  
最終更新: 2026-09-11 / `origin/main` = `dd3da59f9`。ローカル開発 READY はなし（検証・外部ゲートのみ残存）。  
対象: [todo.md](todo.md)・[todo-performance.md](todo-performance.md)・[todo-fix-auth.md](todo-fix-auth.md) が参照する開発検証・測定・受入・外部環境・Linear境界。

## 判定原則

- source、static check、ローカル候補、CIの一部成功は、UAT、STG、PROD、go-live、Linear Done を示さない。
- 実行直前に対象revision、環境、権限、証跡保存先を再確認する。未取得の外部事実は UNKNOWN、前提不足は BLOCKED とする。
- 秘密、cookie、token、患者・飼主・スタッフの個人情報、臨床データをこの文書・Git・共有ログに書かない。
- push、merge、deploy、migration apply、Linear書込み、共有環境への作成・更新・削除は、各実行単位で別途承認を得る。

## 検証キュー

| 順 | ID | 範囲 | 状態 | PASS 条件 |
|---|---|---|---|---|
| 1 | TODO-V-LINEAR | `META-LINEAR-APPLY` の検索結果と反映対象 | BLOCKED | Team/Project/hub内のライブ読取結果、直接対応の根拠、更新前の下書きを分離して記録する。書込み・Doneは別承認 |
| 2 | TODO-V-S09 | `QA-UAT-S09-FIXTURE` | BLOCKED | 専用fixtureと起動済みの対象stackで #2–#6 を再実行し、各結果の帰属を保存する |
| 3 | TODO-V-V04 | `QA-UAT-V04-RETEST` | UNKNOWN | disposable clinicでCRUD/DELETEをブラウザー再実行し、作成・更新・削除・後処理を承認範囲内で照合する |
| 4 | TODO-V-CLINICAL-E2E | `QA-FULL-CLINICAL-E2E` | BLOCKED | 承認済みの `APP_ENV=test`、E2E資格情報、起動済みstackで clinical E2E と full job の証跡を取得する |
| 5 | TODO-V-STG-DATA | H0–H3 / Lane 3–4 | BLOCKED | 対象clinic、完全入力、backup/rollback、operator、maintenance window、承認を固定し、実行後に件数・clinic_id・金額・画面証跡を照合する |
| 6 | TODO-V-RELEASE | P1–P8 / E1–E2 | BLOCKED | 各ゲートの非機密receiptを個別に確認し、前提が揃うまで go-live を HOLD とする |

<a id="development-verification"></a>

## 開発タスク検証（残り）

DEV-V-PET-REQUEST / DEV-V-LIFF-READERS / DEV-V-OWNER unit は完了済み（Git 履歴）。残るのは OWNER の disposable DB のみ。

| ID | 状態 | 内容 |
|---|---|---|
| DEV-V-OWNER-DB | BLOCKED | disposable `TEST_DATABASE_URL` 未設定。`TestOwnerRepository_UpdateAndFind_ReloadFailureRollsBackUpdate`、`TestOwnerRepository_Update_ClinicIsolation`、`TestOwnerService_Update_DiscountTOCTOU_*`（LockedDiffWithoutPermission を除く）、`TestOwnerRepository_LockByIDForUpdate_RequiresAmbientTransaction` を共有DB以外で実行する。未設定・SKIPはPASSにしない |

## ID 1: Linear

- 検索対象は `todo.md` 記載の Team、Project、hub、関連ID、本文・コメントに固定する。
- 類似語だけでは直接対応と見なさない。候補が複数なら、候補・相違点・UNKNOWNを分ける。
- 外部書込み前に、URL、ライブ状態、照合日時、反映下書きが一致することを確認する。
- free issue limit、read connector不在、権限不足は BLOCKED として残し、課金・契約変更で回避しない。

## ID 2–4: 受入とE2E

- 実行対象の commit、fixture、環境、ブラウザー、データ条件、開始・終了時刻を固定する。
- 画面上の成功表示だけでなく、許可されたデータ状態・clinic境界・失敗時の復旧を確認する。
- 共有環境・臨床データへの書込みは実施しない。必要な作成・更新・削除は disposable または承認済み専用環境に限定する。
- 未実行のscenario、利用不能なfixture、対象外の環境は SKIP / BLOCKED / UNKNOWN を理由とともに残す。

## ID 5: STG データレーン

- 実行前に対象clinic、data owner、operator、対象入力、maintenance window、backup、rollback、停止条件を確認する。
- 八王子の完全KNJOまたは承認済みの城東主経路がない限り、H0-3b以降へ進まない。
- migration / load 後は、source manifest・行数・clinic_id・金額・監査可能な結果を照合する。既知破損入力の再実行や共有環境の上書きはしない。
- Lane 4 は、両院の Lane 3 verify と H3-11 の画面証跡が揃うまで PASS にしない。

## ID 6: リリース境界

- P1–P8 を順序どおりに確認する。秘密の値ではなく、実行日時、対象環境、確認者、結果、失敗時の復旧可否だけを receipt に残す。
- 認証・権限、clinic / owner / pet / staff 分離、臨床安全、会計金額、データ消失に未解消FAILがあれば go-live は No-Go とする。
- E1 / E2 は実外部環境の受入であり、mock・ローカルテスト・CIを代替証拠にしない。

結果は該当する既存の runbook または Git ignore 対象の受入証跡へ保存し、この入口台帳では状態と参照先だけを更新する。

---

<a id="perf-stg-login"></a>

## PERF-STG-LOGIN

対象: [todo-performance.md](todo-performance.md) の STG `/login` 初回表示遅延。現状の Worker 観測WIP（`backend/worker/index.ts`、`backend/worker/index.test.ts`）は**導入不要**として保全する。削除・commit・deploy はしない。

理由:

- E1 で大きかったのは final GET の送信前、ブラウザー側の接続開始前待ちである。Worker 内の `container.fetch` 所要時間だけでは、この区間を測れない。
- WIP は Container 起動イベントを記録しないため、OPTIONS、接続待ち、Container 起動の因果を分離できない。
- WIP は `/api/v1/me` 以外も含む通常プロキシ要求ごとに `console.info` を追加する。出力項目は固定されているが、必要性未確定の常時ログは運用量を増やす。
- テストは純粋関数出力だけで、実際の proxy 経路の出力・forwarding 不変条件を検証していない。既存の `make test-worker ARGS='backend/worker/index.test.ts'` はrootをmountする専用Docker runnerを使う。「runner不在」は撤回する。2026-09-11にMakefileの静的契約8項目はPASS、WIPのruntime testは未実施。`backend/worker/tsconfig.test.json` の明示includeに `index.test.ts` がない点も将来の型検査対象に含める。

この判定は性能改善、STG受入、リリース可否を示さない。

旧開発候補3件は [裁定記録](docs/work/development-task-decisions.md) に保存し、現在の開発キューには採用しない。下記の証跡から必要な変更箇所を特定できた場合に限り、対象・最初の変更・完了条件を確定した開発タスクを [todo.md](todo.md#development-tasks) に戻す。

| 順 | ID | 実施内容 | 状態 | 次へ進む条件 |
|---|---|---|---|---|
| 1 | PERF-V-CLIENT-TRACE | 通常の `/login` 読込で、遷移前から Network/Performance を開始し、OPTIONS と GET を時刻・method・status・protocol・timing で関連付ける | BLOCKED | 承認済みの STG 対象、時間枠、停止担当、証跡保存先 |
| 2 | PERF-V-CF-EVENTS | 同じ時刻の Worker/Container 起動・forwardingイベントを read-only で照合する | BLOCKED | Cloudflare の読取権限または機密除去済み event export |
| 3 | PERF-V-DECIDE-OBSERVATION | 1 と 2 の証跡で、Worker 観測が必要かを再判定する | DEFERRED | OPTIONS または GET の遅延を含む再現記録と、対応時刻の provider 証跡 |
| 4 | PERF-V-IMPLEMENT-OBSERVATION | 観測を必要と判断した場合のproxy回帰・出力制約を検証する | DEFERRED | ID 3の採用判定と対象実装。既存Docker runnerで実proxy経路、出力制約、型検査を確認 |
| 5 | PERF-V-MITIGATION | 通信経路変更の因果前提と安全条件を確認する | DEFERRED | ID 1–3 の因果証跡。変更候補なしの現状では検証PASSにしない |
| 6 | PERF-V-BUNDLE | 初回JSのbundle寄与を測定する | DEFERRED | 固定revisionのcache状態別の転送量、parse/execute、操作可能時刻を保存 |
| 7 | PERF-V-STG-ACCEPTANCE | 固定revisionのSTG候補を検証し、配信後にブラウザー受入を実施する | BLOCKED | PR競合・product CI・配信revision・外部更新承認・redacted browser evidence |
| 8 | PERF-V-LINEAR | 関連issueをread-only照会し、ローカル反映案を作る | BLOCKED | Team/Project/hubのライブ読取手段。書込みは別承認 |

### PERF-V-CLIENT-TRACE

- 通常利用環境を変更せず、`/login` 遷移**前**から記録する。
- 保存するのは相対時刻、method、status、protocol、initiator、OPTIONS/GET の対応関係、FCP、フォーム操作可能時刻だけとする。
- Cookie、Authorization、request/response body、患者・飼主・医院情報、画面に表示された個人情報を保存しない。HAR、trace、画面画像を保存する場合は事前に機密情報を除去する。
- 自然な未使用期間後の初回と直後の再読込を分ける。未使用時間だけで cold start と判定しない。
- 1回の正常値・異常値で p95/p99、修正済み、Container 起動、性能改善を主張しない。

PASS 条件: 対象revision・観測条件・時刻が固定され、OPTIONS と GET の各区間および FCP/操作可能時刻が機密情報なしで保存されていること。

### PERF-V-CF-EVENTS

- browser 証跡の時刻窓だけを対象に、Worker受付、Container forwarding、起動・再起動のイベントを read-only で照合する。
- provider のログ時刻と browser の相対時刻の基準を記録する。時刻基準が対応できない場合は UNKNOWN とする。
- Container 起動証跡がない場合、Worker forwarding 所要時間だけから cold start と判定しない。
- Cloudflare設定の変更、keep-alive、`sleepAfter`変更、edge OPTIONS応答、deploy はこの検証に含めない。

### PERF-V-DECIDE-OBSERVATION

| 証跡 | 判定 | 次の対応 |
|---|---|---|
| 遅延が Worker 到達前にあり、OPTIONS/GET の Worker 所要時間は短い | Worker WIP は不要 | proxy、接続交渉、拡張機能、再試行を限定して調査する |
| OPTIONS または GET の Worker→Container 所要時間が遅く、provider 起動イベントと相関する | 最小観測を検討可 | 対象を login 関連の OPTIONS と `/api/v1/me` に限定した設計を作る |
| Worker 所要時間は遅いが起動イベントがない | 原因 UNKNOWN | Go request latency、Container状態、接続を別々に照合する |
| 証跡が再現しない、または時刻対応できない | 判定保留 | WIPを導入せず、次の通常読込で再取得する |

### PERF-V-IMPLEMENT-OBSERVATION の受入条件

- 出力対象は login 関連の OPTIONS と `/api/v1/me` に限定し、固定の method/path 分類、非負の丸めた所要時間、status または固定 failure code、検証済み相関ID以外を出さない。
- raw URL・query・Cookie・Authorization・本文・IP・例外文・任意ヘッダを出さない。
- proxy の response、`CF-Connecting-IP` からの既存 `X-Forwarded-For` 処理、503 fallback、CORS/認証の振る舞いを回帰テストで確認する。
- 純粋関数テストだけでなく、実際の Worker proxy 経路で許可された出力だけが発生することを確認する。
- `make test-worker ARGS='backend/worker/index.test.ts'` が既存のDocker scoped test経路。型検査はrootをmountしたDocker内で `pnpm run typecheck:worker` を実行し、新規テストもtsconfigの対象に含める。`make test-worker` 自体は型検査を実行しない。共有依存volumeの同時利用は避け、hostのnpm/pnpmで代用しない。

STG の実測は、既存の [STG パフォーマンス測定チェックシート](docs/ops/testing/STG-PERFORMANCE-CHECKLIST.md) の対象、時間枠、停止担当、証跡保存先を事前に満たす。結果は Git に入れず、`reports/uat-YYYY-MM-DD/performance/<RUN-ID>/` に保存する。Cloudflare読取、STG traffic、push、merge、deploy、設定変更、Linear更新はそれぞれ別途承認が必要である。

### PERF-V-BUNDLE

- cacheなし・ありを分け、loginに不要な静的resourceの転送量、parse/execute、フォーム操作可能時刻を記録する。
- E1のload約0.31秒を23秒待ちの主因にしない。実測で寄与が確認されるまで charts / LIFFの分離、deep import、manualChunks全面再編を行わない。
- 実装する場合は、既存公開entrypointとfeature境界を保ち、前後比較で独立した改善量を示す。

### PERF-V-STG-ACCEPTANCE

- 実行直前に `main` / `staging` のSHA、PR head/base/mergeable/checksを読み取り、候補は隔離worktreeで作る。migration、seed、env、workflowの差分は別枠で確認する。
- product jobが実際にSUCCESSし、provider revisionとSTG asset/revisionが一致するまで受入を始めない。
- 匿名、既存session、復旧画面、医院選択を確認し、待機表示、フォーム操作可能、認証成功の3時刻とOPTIONS/GET timingを分けて保存する。
- 単発値、spinner、security checkだけで性能改善・p95/p99・release readinessを主張しない。

### PERF-V-LINEAR

- Team `Baritech`、Project `ノア動物病院電子カルテ`、hub `BRT-4` 内で、`PERF-STG-LOGIN`、`/login`、`/v1/me`、`preflight`、`cold start`、`Container`、PR #388 / #393を検索する。
- 直接対応するissueだけを対応付け、候補が複数なら相違点とUNKNOWNを残す。原因未確定、C/D未着手、STG受入の状態を分けたローカル下書きを作る。

---

<a id="認証認可の外部境界"></a>

## 認証・認可の外部境界

対象: [todo-fix-auth.md](todo-fix-auth.md) の未完了項目 D1 と LINEAR。D2、D3、D5、D1合成の disposable 実DB証跡は同書の完了記録を正本とし、再実行対象にしない。

- disposable 実DB、static procedure test、offline check は本番付与、対象環境メール、Linear反映の PASS を示さない。
- 共有 `ekarte_db`、`old-db-postgres`、STG、PRODをテストDBに使わない。agentは migration を共有環境へ apply しない。
- credential、接続文字列、メールアドレス、cookie、token、患者情報を記録しない。必要な証跡は環境識別子、実施日時、operator、承認記録、結果だけにする。
- 本番付与、メール送信、Linear書込み、STG/PROD変更は個別の明示承認が必要である。

| 順 | ID | 実施内容 | 状態 | PASS 条件 |
|---|---|---|---|---|
| 1 | AUTH-V-D1-PREFLIGHT | D1対象環境、operator、承認、既存staff/主所属、rollbackを確定する | BLOCKED | 対象環境と責任者、非機密承認記録、既存状態の確認、失敗時の復旧手順が揃う |
| 2 | AUTH-V-D1-APPLY | 承認済み手順で初回管理者を付与し、通常loginを確認する | BLOCKED | ID 1、明示実行承認、対象環境の安全な操作経路 |
| 3 | AUTH-V-D1-MAIL | 対象環境のメール経路を確認する | BLOCKED | ID 2、承認済みテスト先、外部送信と後処理の範囲 |
| 4 | AUTH-V-LINEAR-READ | 関連issueをライブで検索し、ローカル下書きと対応付ける | BLOCKED | read-only connector または機密除去済み検索結果 |
| 5 | AUTH-V-LINEAR-WRITE | 承認済みの反映案を投稿する | BLOCKED | ID 4、workspaceの作成枠、明示書込み承認 |

### D1 の受入条件

1. 実行前に対象環境、operator、承認記録、既存staffと主所属、影響範囲、rollbackを記録する。対象・承認・復旧が不明なら実行しない。
2. 承認済みの本番手順だけを用い、初回管理者付与後に通常loginを確認する。成功の根拠は対象環境での非機密 receipt とする。
3. メールは承認済みのテスト先だけへ送り、宛先・本文・tokenを証跡に残さない。送信失敗、想定外の宛先、既存staff/所属の不整合は FAIL とし、後続へ進まない。
4. D1合成のPASS、D2/D3/D5のdisposable PASS、ローカルCIは補助証跡であり、本項目の代替にしない。

### Linear の受入条件

- `todo-fix-auth.md` の下書き内容と、ライブで読めたissueのURL、状態、担当、受入条件を照合する。
- 直接対応が見つからない、検索手段がない、workspaceの free issue limit により作成が拒否される場合は UNKNOWN / BLOCKED を維持する。
- 作成・コメント・状態変更・Doneは明示承認後にだけ行う。拒否を課金・契約変更で回避しない。
- 反映後も、本番D1・メール・外部受入が完了していなければ `LOCAL COMPLETE / EXTERNAL INCOMPLETE` を維持する。

結果は非機密 receipt と既存の設計書・runbookに保存し、各TODOでは状態と参照先だけを更新する。

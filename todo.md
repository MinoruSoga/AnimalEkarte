# タスク台帳 — 入口

最終照合: 2026-09-22（JST）。ローカル HEAD `cd2feaa14` の追加対応8件を照合し、関連する残作業を同期した。その他の判定は9月21日の照合を保持。remote・CI・配備・Linear の現在状態は再照会していない。投稿・状態変更は未実施。**未完了の作業だけを掲載する。** 完了の詳細はGit履歴と元のUAT記録を参照する。ローカル準備・修正・回帰追加を、課題全体・STG読取・医院受入・本番実行の完了にしない。

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

### 今開始する4件（仕様入力の回収は完了）

| 順 / ID | 最初に行う作業 | この単位の完了条件 / 後続 |
|---|---|---|
| 1 / UAT-R2-MASTER-PATH | [全経路票](docs/work/todo-campaign-20260918/UAT-R2-MASTER-PATH.md) の複数マスタrequest/model回帰は追加済み。次はexam_types配線・予防分類・入院参照の下流を合成ケースへ | 全12フォームの新規/編集→再読込→利用/会計は未実証。専用fixtureで未カバー箇所を検証し、再現した金額不一致だけ修正 |
| 2 / UAT-R2-CHART-FIT | 高さ制約・タブ内scrollのコード対応済み。[受入キュー](todo-verification.md#uat-followup) で1366×625、全9タブ、sidebar両状態の残ケースを確認 | 必須情報/保存/フォーカスに到達できる対象buildの証拠。Windows 8/Chrome実機の版・CSS領域・100%/既報125%はQAが採取し、最新Chromiumと分ける |
| 3 / UAT-R2-EXCLUSIVE-LOCK | 二重会計409・明細一意制約と、古い合計/確定後明細/検査重複のmock回帰は追加済み。次は [競合票](docs/work/todo-campaign-20260918/UAT-R2-EXCLUSIVE-LOCK.md) の所見以外stale更新API・不足防御を設計 | 実DB/2セッションは [DB適用証拠](todo-operations.md#billing-schema-readiness) と専用環境待ち。古い合計/後追い明細の実並行も未確認。全面ロックは未採用 |
| 4 / UAT-Q2-TREATMENTS-IMPORT | [全期間移行票](docs/work/todo-campaign-20260918/UAT-Q2-TREATMENTS-IMPORT.md) の列写像は調査済み。未確定FK/分類/日時/用量/価格等を契約案・合成fixture設計へ | 現行21表にはtreatments/prescriptionsなし。全種類・全期間の履歴契約を両repoでレビュー後に実装。実データ投入は別承認 |

上記は4件の現在の次工程。既存の設計票や対応済みコードを作り直さず、依存しない調査は別worktreeで並行可。実装時に同じカルテ/会計ファイルへ触れる単位は直列化する。対象の全列挙・実機値の計測・元列の特定は担当者の作業であり、同じ内容を依頼者へ聞き直さない。個人名を伴う要件/受入担当の参照は製品仕様変更前に実行票へ記録する。

### 追加対応8件の残作業（2026-09-22）

上のMASTER / EXCLUSIVE / IMPORTに加え、[尿検査M4](todo-issue.md#slack-manual-urine) のカルテ表示回帰と [主訴C0](todo-issue.md#slack-complaint) の実UI解除導線はローカル作業を継続できる。主訴のnull再読込修正・追加済み回帰は再実装しない。[プラン手入力](todo-issue.md#slack-plan-manual) は案内/導線採否のPO待ちへ、[ワクチン複数入力](todo-issue.md#slack-vaccine-multi) と [遅延](todo-issue.md#slack-latency) は元症状・実機/実測の証拠待ちへ移した。受入の残条件は [検証TODO](todo-verification.md#ready8-followup-20260922)。今回のコード差分は主訴修正、その他は回帰追加/調査票の具体化であり、8課題全体が完了したという意味ではない。

### その他の残件を開始する順序

| 残件 | 次の具体作業 / 終了または停止条件 |
|---|---|
| 性別修正・handoff | old_db修正は `a2cea37` でmain統合済み。再統合せず、同revisionを含むbundle/header/hashとPostgreSQL検証receiptを [運用計画](todo-operations.md#uat-q3-gender-map) へ対応づける。未照合を未実施と断定しない |
| ワクチン種・未納・死亡日訂正 | 既存集計/訂正設計は完成済み。[運用担当](todo-operations.md#uat-data-operations) が対象医院・期間・読取範囲/承認・保存先を埋める。条件が揃うまで実データ操作は停止 |
| OWNER / S09 / V04 / clinical / その他UAT | [検証表](todo-verification.md#readiness-preparation) のexactケースを対象revisionの既存receiptへ対応づける。未収録ケースだけ専用fixtureと実行先を確保して検証。起動中の他タスク用コンテナは借用しない |
| P2 / P5 | 差分・remote方式は [LINMIG-232](docs/work/linmig-campaign-20260919/LINMIG-232.md) / [LINMIG-230](docs/work/linmig-campaign-20260919/LINMIG-230.md)。provider有効化・配備・発行は承認まで停止 |
| STGデータ / release / P1・P3・P6・P7 / 認証 | 不足表は linmig / remaining 票へ作成済。[運用表](todo-operations.md#readiness-preparation) の実行条件が揃うまで送信・資格情報変更・本番作業は停止 |
| 性能 | [測定票](todo-performance.md#測定準備で決めること) に対象build/通常操作/区間/時刻/回数を固定。未測定の因果関係で改善を実装しない |
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
| [UAT-R2-MASTER-PATH](todo-issue.md#uat-r2-master-path) | 金額を持つ全マスタの保存・再読込・下流経路を検証 | 価格request/model回帰拡充済み／下流配線・全経路の検証継続 |
| [UAT-R2-EXCLUSIVE-LOCK](todo-issue.md#uat-r2-exclusive-lock) | カルテ上書きと二重会計の両方を防止 | 会計防御/追加mock回帰あり／所見以外stale設計、DB適用・実並行は残る |
| [UAT-Q3-GENDER-MAP](todo-issue.md#uat-q3-gender-map) | コード3/4修正はold_db main統合済み。bundle・DB/STG証拠を照合 | 未完了（bundle・運用） |
| [UAT-Q2-VACCINE-SPECIES](todo-issue.md#uat-q2-vaccine-species) | 猫に犬用ワクチン（Proheart・6種等）。件数調査のあと種を付ける | 集計設計票作成済／STG 調査待ち |
| [UAT-Q4-UNPAID-TRIAGE](todo-issue.md#uat-q4-unpaid-triage) | 未納はデモではない。実未納と突合漏れを集計で切る。一括完了しない | 集計設計票作成済／STG 調査待ち |
| [UAT-Q2-TREATMENTS-IMPORT](todo-issue.md#uat-q2-treatments-import) | 処置マスタと患者ごとの全種類・全期間の履歴を移行 | 列写像調査済み／履歴契約案・fixture設計READY、実装/投入は未実施 |

既存のsource調査と追加回答を各IDへ反映した。READYの範囲と、実行環境・外部承認待ちを分離する。`TASK-444` / `BE-RC-009` / `BE-RC-017` の既存完了は維持する。

<a id="product-bugs"></a>

## 確認済み製品 FAIL

現在の未解消報告は [todo-issue.md](todo-issue.md) の症状別調査・データ課題に整理した。`bug.md` / `bug-2.md` の FIXED / SPEC-OK を未修正の製品 FAIL として再登録しない。環境・fixture 不足は運用、未確認操作は検証に置く。新たに製品欠陥が確定したら、既存 ID のまま再現・原因・担当範囲を確定する。

### UAT 2026-09-23 確定分（証拠: `reports/uat-2026-09-23/`、詳細: `bug.md` 末尾「確認済み製品欠陥」）

| ID | severity | 領域 | 症状 | シナリオ |
|:---|:---|:---|:---|:---|
| BUG-LIFF-HEALTHCARD-OWNER-SYNC | High | liff / health-card | LIFF 連携済み飼主のヘルスカードが空（`line_customers.owner_id` 未同期） | S12 |
| BUG-ACCT-INS-SIGN-MISMATCH | High | accounting / insurance | 保険付き会計が FE/BE 符号規約不整合で UI から確定不能（400） | S15 |
| BUG-ACCT-INS-EDIT-REWRITE | Medium | accounting / insurance | 確定会計の無変更保存で `insurance_amount`/`billing_amount` が recalc 値に上書き・保険表示消失 | S15 |
| BUG-DIALOG-FOCUS-RESTORE | Medium | shared UI / a11y | 外部 open 制御ダイアログの閉鎖後フォーカスが body へ落下（TreatmentSearchDialog・OwnerSearchModal） | S18 / S32 |
| BUG-BILLING-TAX-TYPE-DROPPED | High | accounting / master | 内税・非課税マスタが会計明細へ `excluded` として伝播・税過剰計上 | S20 |
| BUG-ACCT-DUP-COMPLETE-500 | Medium | accounting / idempotency | 同一カルテ別キー確定が UNIQUE 競合経路で 409 でなく 500 | S21 |
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

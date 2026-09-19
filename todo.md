# タスク台帳 — 入口

最終照合: 2026-09-19（JST）。ローカル `main` `71c583992` のTODO・既存準備票・sourceと、本会話の依頼者回答4点を照合。Linear は2026-09-18 23:22–23:23 JSTに BRT-4 / BRT-41 / BRT-42 / BRT-45 / BRT-68 を読取確認（ローカルIDとの直接対応0件）。9月19日の接続試行は `USER_NOT_LOGGED_IN` で、現在の外部状態はUNKNOWN、投稿・状態変更は未実施。**未完了の作業だけを掲載する。** 完了の詳細はGit履歴と元のUAT記録を参照する。

## 着手プランの確認

**回答を受けた4件は、対象・最初の作業・完了条件を固定して調査/設計に着手できる。** 「医院に聞いてから検討する」という待ちを解除し、下の実行キューから始める。実環境の検証・データ移行・本番作業には別の開始条件があるため、それらまで一括READYとはしない。[開発・調査](todo-issue.md)、[検証・受入](todo-verification.md)、[運用・納品](todo-operations.md) が詳細の正本。`todo-performance.md` は性能の技術記録、`bug.md` / `bug-2.md` / 医院報告は根拠・履歴、`docs/ops/backlog-spreadsheet.md` は外部Q&A管理規則。

### 着手判定の読み方

- **READY（調査/設計）**: 対象と成果物が具体化済みで、repo内の調査、失敗ケースの設計、変更案の作成を開始できる。既存準備票を作り直さず、不足する次工程へ進む。runtime実行可・実装済みを意味しない。
- **準備可**: 既存証拠照合や実行条件を揃える作業が可能。十分な成果物があれば完了扱いとし、入力待ちのまま同じ設計票を増やさない。
- **条件付き**: 対象版・環境・fixture・実行者・必要な承認参照・証拠保存先が揃った単位だけ実行できる。現時点で揃ったと確認していないものを READY に昇格しない。
- **BLOCKED / 判断待ち**: 本文の不足入力が必要。準備可の作業は先行できるが、医院の事実、責任者名、承認、現在の外部状態を推測で埋めない。
- **DEFERRED**: 再開条件まで実装しない。既存の調査メモで十分なら新たな作業を起こさない。

担当欄の「開発・QA・運用・PO・医院・producer」は必要な役割で、個人への割当や承認取得を表さない。実行担当は claim 取得時、要件責任者・受入者は実装/受入前に記録する。通常のローカル調査・文書下書きに追加承認は不要。

### 2026-09-19 のローカル準備状況

次の8 IDの準備票を作成し、固定セットのローカルキャンペーンは `COMPLETE`。準備票は `docs/work/todo-campaign-20260918/` に配置した。各本体の実装、STG読取、医院受入、外部反映は完了していない。

- `UAT-R2-MASTER-PATH`、`UAT-R2-EXCLUSIVE-LOCK`、`UAT-R2-CHART-FIT`
- `UAT-Q2-VACCINE-SPECIES`、`UAT-Q4-UNPAID-TRIAGE`、`UAT-Q2-TREATMENTS-IMPORT`
- `PO-PET-DECEASED-DATA-BACKFILL`、`TODO-V-LINEAR`

### 今開始する4件（仕様入力の回収は完了）

| 順 / ID | 最初に行う作業 | この単位の完了条件 / 後続 |
|---|---|---|
| 1 / UAT-R2-MASTER-PATH | [全経路票](docs/work/todo-campaign-20260918/UAT-R2-MASTER-PATH.md) の11単価＋1割引フォームをroute/API/既存テストへ対応づける | 新規/編集→再読込→利用/会計の未カバーケースと期待値が揃う。専用fixtureで検証し、再現した金額不一致を修正 |
| 2 / UAT-R2-CHART-FIT | [UI票](docs/work/todo-campaign-20260918/UAT-R2-CHART-FIT.md) の1366×625、全9タブ、sidebar両状態でoverflow原因と修正箇所を特定 | 必須情報/保存/フォーカスの到達を保つ修正案とテスト。Windows 8/旧Chrome互換性は別検証、実機値はQAが採取 |
| 3 / UAT-R2-EXCLUSIVE-LOCK | [競合票](docs/work/todo-campaign-20260918/UAT-R2-EXCLUSIVE-LOCK.md) を更新API/既存防御へ対応づける | 古い保存・二重会計・再送/異なるkey・通信断・医院分離の不足テスト/最小修正案。全面ロックは未採用 |
| 4 / UAT-Q2-TREATMENTS-IMPORT | [全期間移行票](docs/work/todo-campaign-20260918/UAT-Q2-TREATMENTS-IMPORT.md) の旧列→producer→AE→表示を両repoで埋める | マスタ/全種類/全期間の履歴、参照・精度・重複・欠損・復旧を契約化。レビュー後に実装、実データ投入は別承認 |

4件は上から着手を推奨するが、依存しない調査は別worktreeで並行可。実装時に同じカルテ/会計ファイルへ触れる単位は直列化する。対象の全列挙・実機値の計測・元列の特定は担当者の作業であり、同じ内容を依頼者へ聞き直さない。個人名を伴う要件/受入担当の参照は製品仕様変更前に実行票へ記録する。

### その他の残件を開始する順序

| 残件 | 次の具体作業 / 終了または停止条件 |
|---|---|
| 性別修正・handoff | old_db修正は `a2cea37` でmain統合済み。再統合せず、同revisionを含むbundle/header/hashとPostgreSQL検証receiptを [運用計画](todo-operations.md#uat-q3-gender-map) へ対応づける。未照合を未実施と断定しない |
| ワクチン種・未納・死亡日訂正 | 既存集計/訂正設計は完成済み。[運用担当](todo-operations.md#uat-data-operations) が対象医院・期間・読取範囲/承認・保存先を埋める。条件が揃うまで実データ操作は停止 |
| OWNER / S09 / V04 / clinical / その他UAT | [検証表](todo-verification.md#readiness-preparation) のexactケースを対象revisionの既存receiptへ対応づける。未収録ケースだけ専用fixtureと実行先を確保して検証。起動中の他タスク用コンテナは借用しない |
| P2 / P5 | [本番経路](todo-operations.md#p2--prod-setup) の保護環境・production config選択と、[staff remote実行](todo-operations.md#p5--staff-provision) のsecret供給/対象固定/rollback/receiptを設計。provider有効化・配備・発行は承認まで停止 |
| STGデータ / release / P1・P3・P6・P7 / 認証 | [運用表](todo-operations.md#readiness-preparation) と [検証表](todo-verification.md#readiness-preparation) の入力・receiptを照合し、不足分だけ供給者/担当へ依頼する下書き。送信・資格情報変更・本番作業は別承認 |
| 性能 | [測定票](todo-performance.md#測定準備で決めること) に対象build/通常操作/区間/時刻/回数を固定。未測定の因果関係で改善を実装しない |
| Linear | [既存照合票](docs/work/todo-campaign-20260918/TODO-V-LINEAR.md) の履歴を保持。接続復旧後に残る既存Issueをfull local ID/元報告/受入条件で照合。投稿・状態変更は別承認 |
| TASK-444-ADDENDUM-CODEGEN | 必要性と別スコープの採用までDEFERRED。準備票を増やすために再開しない |

今回の文書改訂と準備票はタスク実装、テスト実行、外部反映、受入の完了を意味しない。着手時には現行 HEAD と該当 claim・対象入力を再確認する。

## 残作業の入口

| 区分 | 次に行うこと | 正本 |
|---|---|---|
| 実装済み項目の受入 | 治療 Enter 2回と検索一覧高さはローカル検証済み。実機 IME・再読込・viewport を確認 | [検証キュー](todo-verification.md#uat-followup) |
| 回答済み4件の開発・調査 | 全金額経路、1366×625 UI、上書き/二重会計防止、全期間処置移行 | [未完了 Issue](todo-issue.md#open) |
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

医院フィードバックのうち **まだ直すもの**。原文と回答は [stg-uat-clinic-feedback-q1-q4.md](docs/work/stg-uat-clinic-feedback-q1-q4.md)。詳細・受入は [todo-issue.md](todo-issue.md#open)。実装済み（検索 AND・保険 50/70・問診抜粋リンク・治療 Enter・検索一覧高さ）の受入は [検証 TODO](todo-verification.md#uat-followup) に置く。明日以降の予約編集は手順回答済み。

| ID | 内容 | 状態 |
|---|---|---|
| [UAT-R2-MASTER-PATH](todo-issue.md#uat-r2-master-path) | 金額を持つ全マスタの保存・再読込・下流経路を検証 | 調査/検証設計READY |
| [UAT-R2-CHART-FIT](todo-issue.md#uat-r2-chart-fit) | Windows 8 / Chrome、15.6インチ、1366×625。全9タブの見切れを解消 | 寸法/UI設計READY、旧Chrome受入は別ゲート |
| [UAT-R2-EXCLUSIVE-LOCK](todo-issue.md#uat-r2-exclusive-lock) | カルテ上書きと二重会計の両方を防止 | 競合調査/設計READY |
| [UAT-Q3-GENDER-MAP](todo-issue.md#uat-q3-gender-map) | コード3/4修正はold_db main統合済み。bundle・DB/STG証拠を照合 | 未完了（bundle・運用） |
| [UAT-Q2-VACCINE-SPECIES](todo-issue.md#uat-q2-vaccine-species) | 猫に犬用ワクチン（Proheart・6種等）。件数調査のあと種を付ける | 集計設計票作成済／STG 調査待ち |
| [UAT-Q4-UNPAID-TRIAGE](todo-issue.md#uat-q4-unpaid-triage) | 未納はデモではない。実未納と突合漏れを集計で切る。一括完了しない | 集計設計票作成済／STG 調査待ち |
| [UAT-Q2-TREATMENTS-IMPORT](todo-issue.md#uat-q2-treatments-import) | 処置マスタと患者ごとの全種類・全期間の履歴を移行 | 契約設計READY、実投入は別承認 |

既存のsource調査と追加回答を各IDへ反映した。READYの範囲と、実行環境・外部承認待ちを分離する。`TASK-444` / `BE-RC-009` / `BE-RC-017` の既存完了は維持する。

<a id="product-bugs"></a>

## 確認済み製品 FAIL

現在の未解消報告は [todo-issue.md](todo-issue.md) の症状別調査・データ課題に整理した。`bug.md` / `bug-2.md` の FIXED / SPEC-OK を未修正の製品 FAIL として再登録しない。環境・fixture 不足は運用、未確認操作は検証に置く。新たに製品欠陥が確定したら、既存 ID のまま再現・原因・担当範囲を確定する。

<a id="human-lane"></a>

## PO / 人間レーン

残る製品判断（旧Chrome対応等）は [todo-issue.md](todo-issue.md#open)、実環境の承認と操作は [todo-operations.md](todo-operations.md)、受入・go-live は [todo-verification.md](todo-verification.md)。既存の人間ゲートは BRT-4 配下で追跡する。回答済み4件の範囲を未回答へ戻さない。

<a id="refactor-constraints"></a>

## FE 維持制約

- `design-tokens.ts` / `query-keys.ts` / `paths.ts` の表分割や行数だけを目的とした機械的分割をしない。
- `utils/` 再作成・generated/models 一括移行をしない。必要性は [裁定記録](docs/work/development-task-decisions.md#task-444) に従って判断する。
- `app/pages` の合成と owners `loaders.ts` の例外、権限 ref、死亡 sentinel、`useActionState`、queryKey タプルを維持する。
- FE12 却下（manual chunk、死亡行グレーアウト、owners 行アクションをペット生死で止める）を維持する。

着手時は [AGENTS.md](AGENTS.md) の claim・worktree 規則に従う。秘密・患者情報は台帳へ書かない。

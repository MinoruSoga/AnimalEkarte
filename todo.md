# タスク台帳 — Linear が正本

- Performance (STG login): see `todo-performance.md` (PERF-STG-LOGIN; child unit PERF-STG-LOGIN-A in progress on candidate).
統合日: 2026-09-08。最終ローカル照合: 2026-09-09 / main `48e89dbe4`。Linear 読み取り結果・認証修正の main 反映・旧 claim 削除を確認した。Linear の状態は同日の記録であり、今回のローカル更新では再照会していない。UAT・STG/PROD・go-live は再判定していない。

| 項目 | 値 |
|------|-----|
| **実行 SoT** | Linear Team **Baritech** · Project **ノア動物病院電子カルテ** · hub **[BRT-4](https://linear.app/baritechllc/issue/BRT-4)** |
| **セキュリティ修正** | **[BRT-226](https://linear.app/baritechllc/issue/BRT-226)**（Review · `origin/main` 済み · Done は人間） |
| **本ファイルの範囲** | repo と強く結び付く **未完了作業の入口**、確認済み製品 FAIL、PO 入口、分離した完了履歴・維持制約 |

状態・Done は Linear を正本とする。行値・秘密は書かない。完了項目は実行キューから除く。以下の履歴は当時の証跡への入口であり、現在の受入・release 判定ではない。

`bug.md`・`todo-now.md`・`todo-po.md`・`todo-refactor.md` は本ファイルへ統合して削除した。別台帳として再作成しない。`todo-fix-auth.md` 等、今回指定外の文書は統合・削除していない。

入口: [2026-09-08 の対応履歴](#session-2026-09-08) · [実行キュー](#対応順実行キュー) · [製品 FAIL](#product-bugs) · [PO / 人間レーン](#human-lane) · [Astra 完了履歴](#astra-history) · [FE 完了履歴・維持制約](#refactor-history)

エージェントは PlanetScale、共有 STG apply、`DROP SCHEMA`、本番 cutover、`make reset`、八王子 CSV の producer 出力を実行しない。push / dispatch / Linear Done / 秘密変更は明示承認が必要。

claim は ID ごとに初回編集前に確認・取得する。作成者別の削除条件は [AGENTS.md](AGENTS.md#branch-deletion-by-creator-mandatory) を正本とする。ユーザー作成は AI による削除禁止。AI 作成は統合・明示終了・成果を保全した引き継ぎと未使用を確認して削除可能。2026-09-09 の本更新開始前に `git branch --list 'claim/*'` が空であることを確認した。旧 META / QA / 認証 claim の解除待ちは解消済み。claim の削除は UAT や受入の完了を意味しない。新規着手時は現在の claim を再確認する。

---

<a id="session-2026-09-08"></a>

## 0. 2026-09-08 全項目対応結果

以下は 2026-09-08 セッション当時の記録。MCP・compose・claim の現在状態を示すものではない。当時は main のまま、Linear 書き込み・秘密・STG/PROD・`make up` を実行せず、S09 の HTTP/CLI 以外は原因付きスキップとした。

| ID | 分類 | 実行者 | 今回 | 結果 / 原因 |
|----|------|--------|------|-------------|
| **META-LINEAR-APPLY** | Linear 書き込み | USER | SKIP | Linear MCP なし。`LINEAR_API_KEY` unset。公開ページはログイン壁。エージェントは書かない |
| **H0-2 / HAC-CSV-1** | STG 実データ | old_db / USER | SKIP | HAC-INPUT-2。完全 KNJO 未受領。同一 BAK 再実行と producer は禁止 |
| **H0-3b / H1-2** | STG | USER | SKIP | H0-2 待ち |
| **AE-STG-UAT-LANE3-HAC** | STG | USER | SKIP | 現行状態 UNKNOWN。H0-2 / H0-3b 待ち。共有 STG apply 禁止 |
| **H3-9 staff attach apply** | STG | USER | SKIP | apply 実施有無 UNKNOWN。STG 実行ゲートは USER |
| **H3-11 画面確認** | STG | USER | SKIP | H3-9 と自医院ログインが必要。証跡未取得 |
| **Lane 4** | STG UAT | 医院 / USER | SKIP | 両院 Lane 3 未証明。5営業日証跡なし |
| **P1 SEC-SECRETS-5** | 秘密 rotation | USER | SKIP | 秘密の作成・表示・投入・revoke はエージェント禁止 |
| **P2 #253 PROD-SETUP** | Production 構築 | USER | SKIP | Production 未構築。本番構築は自動実行しない |
| **P3 #250 PROD-DATA-MIGRATION** | 本番移行 | USER | SKIP | 事前準備待ち。本番 cutover 禁止 |
| **P4 #254 AUTHENTICATED-UAT** | 全業務 UAT | USER | SKIP | full UAT 未証明。PARTIAL/BLOCKED/UNKNOWN を PASS にしない |
| **P5 #255 STAFF-PROVISION** | 職員投入 | USER | SKIP | roster/email/PII 入力は USER。値は書かない |
| **P6 #258 DELIVERY** | 納品パッケージ | USER | SKIP | P1・P2 と契約責任者の非機密事実が未反映 |
| **P7 #256 TRAINING** | 操作説明会 | USER | SKIP | 日程・形式・結果は人間レーン |
| **P8 #257 GOLIVE** | go-live | USER | SKIP | HOLD。P1〜P7 未達 |
| **E1 QA-UAT-LSTEP-REAL** | 外部 LSTEP | USER | SKIP | write 有効な実 LSTEP 環境なし |
| **E2 QA-UAT-LINE-IDTOKEN** | 実 LINE | USER | SKIP | 実 LINE idToken は mock 外。未証明 |
| **QA-UAT-S09-FIXTURE** | 受入 helper | agent / USER | **PARTIAL** | HTTP/CLI/原子性/staff/支払/明細/cleanup を実装。ブラウザ #2–#6 は compose 停止と `make up` 禁止のため未。S09 は BLOCKED のまま |
| **QA-UAT-V04-RETEST** | マスタ DELETE | USER / agent | SKIP | live HTTP 403。clinic 1/2 の権限昇格なし。compose 停止。disposable clinic 再実行も stack 必要 |
| **QA-FULL-CLINICAL-E2E** | clinical E2E | USER / agent | SKIP | `--clinical` は APP_ENV=test の起動済み stack と `E2E_LOGIN_PASSWORD` が必要。`make up` 禁止。e2e.yml full job は USER |
| **TASK-444** | deferred | agent | SKIP | generated/models 公開契約・codegen・consumer 移行計画が未揃い。横断キャンペーンにしない |
| **BE-RC-005** | deferred | agent | SKIP | 新規・変更 service から着手。今回その面を触っていない |
| **BE-RC-009** | deferred | agent | SKIP | 新規 consumer 時のみ。今回対象なし |
| **BE-RC-014** | deferred | agent | SKIP | typed error 利用面の変更時。今回対象なし |
| **BE-RC-015** | deferred | agent | SKIP | 新規・変更面の stutter 回避。今回対象なし |
| **BE-RC-017** | deferred | agent | SKIP | 対象 repository 変更時。今回対象なし |
| **BE-RC-019** | deferred | agent | SKIP | lab / hospitalization 境界の成立する変更時だけ |
| **BRT-226** | セキュリティ Review | USER | SKIP | Done は人間。エージェントは遷移しない |
| **製品 FAIL** | 確認済み FAIL | — | なし | 旧台帳時点で未対応なし。現在の製品全体無欠陥の判定ではない |
| **PO / 人間レーン** | Linear hub | USER | SKIP | Linear が正本。書き込みなし |
| **Astra F1〜F6** | 完了履歴 | — | 履歴 | `origin/main` 統合済み。未完了の後続は上表 |
| **FE リファクタ維持** | 維持制約 | agent | 維持 | 表分割・utils 再作成・FE12 再提案をしない。新規作業ではない |

S09 局所検証（2026-09-08）: fail-closed / CLI / OpenAPI drift GREEN。fixture+HTTP の testdb は disposable Postgres で GREEN。共有 compose は停止のまま。`make codegen` は未実行（USER）。

---

## 対応順（実行キュー）

依存関係と実行権限を満たす項目から 1 件ずつ着手する。USER / old_db の入力・承認待ちはその項目と依存先を停止し、独立した読み取り照合・受入準備は継続できる。deferred はキューに入れない。

| 順 | ID | 実行者 | なぜこの順 | 状態 |
|----|----|--------|------------|------|
| 1 | **META-LINEAR-APPLY** | agent（追加照合） / USER（反映） | 検索範囲・結果は [linear-f1-f6-mapping.md](docs/work/linear-f1-f6-mapping.md)。F1〜F6 の直接対応を確定してから反映 | **PARTIAL**（2026-09-09: プロジェクト55件・hub・関連語を照合。直接対応 ID は未特定。BRT-226 は Review。追加の本文・コメント照合は可能。書き込み・Done は USER 待ち） |
| 2 | **H0-2 / HAC-CSV-1** | old_db / USER | STG 八王子の先頭。これより前の STG 行は進めない | **BLOCKED**（HAC-INPUT-2。完全 KNJO 未受領。同一 BAK 再実行と producer は禁止） |
| 3 | **H0-3b → Lane3 HAC → H3-9 → H3-11 → Lane 4** | USER | 2 の依存どおり | 待ち |
| 4 | **P1 → P2 → P3 → P4 → P5 → P6 → P7 → P8** | USER | go-live 依存。E1 / E2 は P4 の一部 | 待ち |

STG レーンの次は医院/ベンダーからの完全 KNJO 再取得、または城東主経路（JOU-G2-2 の Azure 承認）。H0-3b には入らない。独立して F1〜F6 の Linear 読み取り照合と、下記受入残の実行前提・証跡の確認を進められる。S09 / V04 / clinical E2E の実行は各設計の環境・承認条件を満たしてから行う。Linear 書き込みと Done は USER。

2026-09-09 引き継ぎ: Cursor の Linear 読取不可はそのセッションの制約。Codex の照会結果は `48e89dbe4` で main に反映済み。#254 全体の受入入口は [BRT-45](https://linear.app/baritechllc/issue/BRT-45)（同日観測 Needs Human）。[BRT-68](https://linear.app/baritechllc/issue/BRT-68) の同日観測の残は実 LINE / LIFF。両者を S09 等の PASS と混同しない。旧 META / QA claim は削除済みで、受入残は下記に維持する。

認証の独立した残検証: `950404408`（権限ロック・復旧修正）と `9be825a66`（認可経路・分離検証の追加）は main に反映済み。詳細と開始条件は [todo-fix-auth.md](todo-fix-auth.md) を参照する。実DB並行・初回管理者SQL実行・全経路の実DB返却データ分離・2サーバー即時失効・対象環境のメール・負荷測定は未検証の記録。限定テストの成功をこれらの完了と扱わない。ローカルの前提確認は独立して進められるが、DB/環境操作は個別の実行条件に従う。

---

## 1. 受入残（PASS にしていない）

helper / 再実行スライスは済。UAT / E2E を PASS にしない。正本は [UAT-DOMAIN-STATUS.md](docs/ops/testing/UAT-DOMAIN-STATUS.md)。

| ID | 残 | 状態 |
|----|----|------|
| **QA-UAT-S09-FIXTURE** | ブラウザ #2–#6 再実行。HTTP/CLI は 2026-09-08 実装済み | S09 は BLOCKED |
| **QA-UAT-V04-RETEST** | live HTTP は 403。clinic 1/2 の権限昇格なし | V04 は UNKNOWN |
| **QA-FULL-CLINICAL-E2E** | `--clinical` 未実行。e2e.yml job は未 | E2E は未証明 |

設計: [S09-FIXTURE-DESIGN.md](docs/ops/testing/S09-FIXTURE-DESIGN.md) · [CLINICAL-E2E-DESIGN.md](docs/ops/testing/CLINICAL-E2E-DESIGN.md)。

---

## 2. USER ゲート（秘密・本番・外部環境）

外部状態は実行直前に再確認する。エージェントは秘密値の作成・表示・投入、共有 STG/PROD apply、production 構築、go-live を自動実行しない。

| 順 | ID | 実行者 | 状態 | 完了条件 |
|----|----|--------|------|----------|
| P1 | **SEC-SECRETS-5 / #89 / #97** | USER | 4系統 rotation receipt 未記入 | 新発行→投入→再 deploy→health→旧値 revoke→旧値拒否。値は記録しない |
| P2 | **#253 / U12 PROD-SETUP** | USER / 開発 | Production 未構築 | Cloudflare 本番、Required reviewers、workflow、rollback、backup rehearsal、URL/CI receipt |
| P3 | **#250 PROD-DATA-MIGRATION** | USER / 開発 | 事前準備待ち | rehearsal、最終 import、入力停止、backup/rollback、件数・clinic_id・金額突合 |
| P4 | **#254 AUTHENTICATED-UAT** | USER / agent | full UAT 未証明 | 全業務 scenario の受入結果を確定。PARTIAL / BLOCKED / UNKNOWN を PASS にしない |
| P5 | **#255 STAFF-PROVISION** | USER | 入力未記入 | roster、email 方針、clinic、role、actor、環境承認。PII-free receipt |
| P6 | **#258 / U1〜U12 DELIVERY** | USER | 最終承認待ち | P1・P2 と契約責任者の非機密事実を `DELIVERY_PACKAGE.md` へ反映 |
| P7 | **#256 / U13 TRAINING** | USER | 操作説明会未完 | 日程・形式・範囲・結果・opaque receipt |
| P8 | **#257 GOLIVE** | USER | HOLD | P1〜P7 の受入と第2段階条件。重大 FAIL や当日 import 未達なら No-Go |
| E1 | **QA-UAT-LSTEP-REAL** | USER | 外部環境待ち | write 有効な LSTEP で S01 同期と V05-17 remove |
| E2 | **QA-UAT-LINE-IDTOKEN** | USER | mock 外・未証明 | 実 LINE idToken で link / 409 / 期限切れ 400 |

P4 の延期例外: 臨床安全、会計金額、clinic / owner / pet / staff 分離、認証・権限、データ消失の未解消 FAIL は go-live 前に解消する。それ以外は Linear に受容条件を残し、USER の明示受容がある場合だけ延期できる。

---

## 3. STG 実データ（USER / old_db）

対象は八王子 `clinic_id=1` と城東 `clinic_id=2`。証跡が無いだけなら **UNKNOWN**。再 apply の前に USER が現行状態を確認する。

| 優先 | ID | 実行者 | 状態 | blocked-by |
|------|----|--------|------|------------|
| 1 | **H0-2 / HAC-CSV-1** | old_db / USER | HAC-CSV-1 は C1 `HAC-INPUT-2` 待ち。AE `hachioji/` 空。export なし | CHECKDB clean な新規 BAK、または全32列完全 KNJO。同一/既知破損 BAK の再復元は禁止。城東 live を上書きする load も禁止 |
| 2 | **H0-3b / H1-2** | USER | 待ち | H0-2 |
| 3 | **AE-STG-UAT-LANE3-HAC** | USER | UNKNOWN・投入判断待ち | 現行状態、H0-2、H0-3b |
| 4 | **H3-9 staff attach apply** | USER | 入力あり・apply 実施有無 UNKNOWN | 現行 attach と STG 実行ゲート |
| 5 | **H3-11 画面確認** | USER | UNKNOWN・証跡未取得 | H3-9 と自医院ログイン |
| 6 | **Lane 4** | 医院スタッフ / USER | 完了未証明 | 両院 Lane 3 verify、H3-11 |

索引から外したもの: AE-OLD-DB-MR-UNIQ、Lane3 城東 21表、H3-7 敷島 / Hako。

STG 実行ゲート: 対象環境、data owner、operator、maintenance window、backup / restore、rollback、承認。正本は [STG 手順の停止ゲート](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md#2-pre-deploy-stop-gates)。

観測（2026-09-06）: 城東・敷島・箱は AE handoff に bundle あり。八王子ディレクトリは空。old_db export も八王子 run なし。producer は未実行。Lane 4 は 5営業日証跡なし。

---

## 4. 触ったときだけ / deferred

横断キャンペーンにしない。

| ID | 再開条件 |
|----|----------|
| **TASK-444** | generated/models の公開契約・codegen・consumer 移行計画が揃ってから |
| **BE-RC-005** | 新規・変更 service から 5xx 二重ログを解消 |
| **BE-RC-009** | 新規 consumer または対象機能変更時に利用側最小 port へ分割 |
| **BE-RC-014** | typed error が使えるようになったら `errors.As` へ |
| **BE-RC-015** | 新規・変更面から package.Type stutter を避ける |
| **BE-RC-017** | 対象 repository 変更時に unexported update + typed command |
| **BE-RC-019** | lab / hospitalization 等の境界が成立する変更時だけ |

---

<a id="product-bugs"></a>

## 5. 確認済み製品 FAIL（旧 bug.md）

記録対象は確認済み製品 FAIL のみ（[TEST_ARCHITECTURE.md](docs/ops/testing/TEST_ARCHITECTURE.md) §6）。環境・seed・権限・fixture 不足による BLOCKED / PARTIAL や受入未実施を混ぜない。証跡に credential・token・cookie・idToken・個人情報（PHI）を含めない。起票後は Linear で追跡し、見出し ID は本節内で重複させない。新規項目は `### BUG-XXX` で本節に追加する。

| ID | status | area | severity | scenario | 層 |
|:---|:---|:---|:---|:---|:---|
| （2026-09-06 の旧台帳では未対応なし） | — | — | — | — | — |

これは旧台帳の記録であり、現在の製品全体に不具合がないという判定ではない。認証レビューの修正計画は [todo-fix-auth.md](todo-fix-auth.md) を参照する（本統合の対象外）。対応済み項目は本節の未対応一覧から除き、履歴は Git と `reports/uat-YYYY-MM-DD/` を参照する。

旧台帳の解消記録（2026-09-06）:

- S06 UAT の auto-draft FAIL は、general reservation type id=5 による V01 再確認 PASS 後に除外。証跡識別子: `reports/uat-2026-09-06/v01/`。
- console-full-sweep の `BUG-20260906-001..004` は実装対応済みとの記録。詳細は Git。当時の Linear 照会は未実施。

<a id="human-lane"></a>

## 6. PO / 人間レーン（旧 todo-po.md）

実行 SoT は Linear hub [BRT-4](https://linear.app/baritechllc/issue/BRT-4) · Project ノア動物病院電子カルテ。人間ゲート（UAT・PO 確認）も Linear を正本とし、別の Open 行台帳を再構築しない。本ファイルの受入残・USER ゲートを入口にする。

会社側索引: CorpVault `50_Projects/ノア動物病院電子カルテ/05_Linearマップ.md`。旧詳細本文は Git 履歴。

<a id="astra-history"></a>

## 7. Astra 品質監査 — 完了履歴（旧 todo-now.md）

以下は旧台帳の観測記録。Astra 監査 F1〜F6 の実装は当時 `origin/main` へ統合済み。未完了の後続は本ファイルの実行キューを入口とし、実行状態の正本は Linear（[作業台帳ルール](docs/work/README.md)）。

- 監査対象: `main` / `c41ba8b1c1aef7f8150457dc310fe7f61fca1a75`（2026-09-05）。
- 最終観測: 2026-09-06 / `b987729fd57b05b5f94f0a7e0d5860d515401421`。
- 旧更新時は `claim/TODO-NOW` が不在だったため、`todo.md` の `LEDGER-TODO-NOW-POINTER` として入口ポインタ化した。現在の claim 状態を示すものではない。
- F1〜F6 の実装と F3 manual E2E auth smoke（run `33972458396`）。
- F4〜F6 の push 後通常 CI。
- F3 k6 aggregate の原因は `metric.values.*` のみ参照。修正の追跡 ID は `CI-K6-SUMMARY-SCHEMA`（当時の `todo.md`、詳細は Git）。
- 「診断カテゴリ」「診断病名」label 接続の追跡 ID は `FE-CLINICAL-PLAN-SELECT-LABELS`。

Linear の F1〜F6 対応案は [linear-f1-f6-mapping.md](docs/work/linear-f1-f6-mapping.md)。旧台帳時点の Linear 本体は UNKNOWN、Done は USER。今回の統合では外部状態を再照会していない。

<a id="refactor-history"></a>

## 8. FE リファクタ — 完了履歴・維持制約（旧 todo-refactor.md）

2026-09-02 の FE 規約ゲート完了記録。対象は `frontend/src`・`frontend/liff/src`・`frontend/line-reserve/src` の production TS/TSX。現在の未完了作業は本ファイル、実行状態の正本は Linear（[台帳ルール](docs/work/README.md)）。

### 当時の証跡

- 当時の完了結果: cross-feature deep import、queryKey 配列リテラル、生産の `window.confirm` を対象範囲で解消。150 行超の生産関数は 0、800 行超ファイルは意図的残置の `design-tokens.ts` のみ。
- Docker scoped Vitest: 第1束 16 files / 400 passed、第2束 11 files / 85 passed。
- 2026-09-02 にユーザーが `make lint-front` / `make test-front` の成功を報告。`pnpm build` は未実行。
- 上記は当時の観測であり、現在の runtime・UAT・release 判定ではない。詳細なカテゴリ別変更、検証、claim の解放記録は固定履歴を参照する。ここでは claim の現在状態を断定しない。

```bash
git show 12f15fe4fc85f3e9900f70575277b5ab5bc645e4:todo-refactor.md
# BE フェーズの完了時本文
git show ad63bdf28:todo-refactor.md
```

### 維持する対象外・制約

- `design-tokens.ts` / `query-keys.ts` / `paths.ts` の表分割、50 行までの機械分割、200–399 行ファイルの薄型化だけを目的とした切断は行わない。
- `utils/` を再作成しない。generated/models の一括移行は TASK-444 の別トラック。当時の allowlist 267 件は分割追従のみ。
- `app/pages` の合成と owners `loaders.ts` の例外を維持する。
- 権限 ref、死亡 sentinel、`useActionState`、queryKey タプルの契約を維持する。
- FE12 却下（manual chunk、死亡行グレーアウト、owners 行アクションをペット生死で止める）は維持する。
- 当時のトリミングフォームの権限・死亡ガード欠落は対象外だった。本履歴から現在の未修正・修正済みを判断しない。

---

## 参照

| 文書 | 役割 |
|------|------|
| [Astra 完了履歴](#astra-history) | Astra F1〜F6 の完了履歴 |
| [今回の対応結果](#session-2026-09-08) | 2026-09-08 の全 ID 対応 / スキップ原因 |
| [docs/work/linear-f1-f6-mapping.md](docs/work/linear-f1-f6-mapping.md) | F1〜F6 対応案。Linear は UNKNOWN |
| [docs/ops/testing/S09-FIXTURE-DESIGN.md](docs/ops/testing/S09-FIXTURE-DESIGN.md) | S09 helper 設計 |
| [docs/ops/testing/CLINICAL-E2E-DESIGN.md](docs/ops/testing/CLINICAL-E2E-DESIGN.md) | clinical E2E 設計 |
| [docs/ops/testing/UAT-DOMAIN-STATUS.md](docs/ops/testing/UAT-DOMAIN-STATUS.md) | UAT 集計の正本 |
| [製品 FAIL](#product-bugs) | 確認済み製品 FAIL |
| [docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md) | ローカル handoff |
| [docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md) | STG 破壊境界 |

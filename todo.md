# タスク台帳 — Linear が正本

統合日: 2026-09-08。最終ローカル/GitHub照合の基準点: 2026-09-10 / 更新着手時の `main` = `origin/main` = `5a19989ad`。`main` → `staging` PR #388 は OPEN / CONFLICTING。Codex の Linear 読み取り結果は `48e89dbe4` 以降も保持。Grok 追加照合では Linear MCP がセッション限定で UNAVAILABLE のためライブ再照会はせず、定義再確認と次照会リストを [linear-f1-f6-mapping.md](docs/work/linear-f1-f6-mapping.md) に記録した。UAT・STG/PROD・go-live は再判定していない。

| 項目 | 値 |
|------|-----|
| **実行 SoT** | Linear Team **Baritech** · Project **ノア動物病院電子カルテ** · hub **[BRT-4](https://linear.app/baritechllc/issue/BRT-4)** |
| **セキュリティ修正** | **[BRT-226](https://linear.app/baritechllc/issue/BRT-226)**（Review · `origin/main` 済み · Done は人間） |
| **本ファイルの範囲** | repo と強く結び付く **未完了作業の入口**、確認済み製品 FAIL、PO 入口、維持制約 |

状態・Done は Linear を正本とする。行値・秘密は書かない。対応済み項目と完了証跡は本ファイルから除き、Git 履歴を参照する。

入口: [実行キュー](#対応順実行キュー) · [受入残](#acceptance-remains) · [製品 FAIL](#product-bugs) · [PO / 人間レーン](#human-lane) · [FE 維持制約](#refactor-constraints)

性能調査・改善: [todo-performance.md](todo-performance.md)（PERF-STG-LOGIN、最終照合 2026-09-10）。A（待機表示）は PR #393 / `9b06b551c`、B（起動時session restoreの8秒上限・障害表示）は `b5be27be6` で `main` 統合済み。GitHub product jobsはSKIPで、STG配信・Browser/E2E・改善後実測・LinearはUNKNOWNまたは未実施。preflight・接続待ち・Container起動の内訳とC/Dは未完了。

エージェントは PlanetScale、共有 STG apply、`DROP SCHEMA`、本番 cutover、`make reset`、八王子 CSV の producer 出力を実行しない。push / dispatch / Linear Done / 秘密変更は明示承認が必要。

claim は ID ごとに初回編集前に確認・取得する。作成者別の削除条件は [AGENTS.md](AGENTS.md#branch-deletion-by-creator-mandatory) を正本とする。ユーザー作成は AI による削除禁止。AI 作成は統合・明示終了・成果を保全した引き継ぎと未使用を確認して削除可能。過去セッションの claim 記録は historical snapshot として読み、現在の保有状態は新規着手時に `git branch --list 'claim/<TASK-ID>'` で再確認する。本 META 追加照合に再着手する場合は `claim/META-LINEAR-APPLY` を確認・取得する。claim の削除は UAT や受入の完了を意味しない。

---

## 対応順（実行キュー）

依存関係と実行権限を満たす項目から 1 件ずつ着手する。USER / old_db の入力・承認待ちはその項目と依存先を停止し、独立した読み取り照合・受入準備は継続できる。deferred はキューに入れない。

| 順 | ID | 実行者 | なぜこの順 | 状態 |
|----|----|--------|------------|------|
| 1 | **META-LINEAR-APPLY** | agent（追加照合） / USER（反映） | 検索範囲・結果は [linear-f1-f6-mapping.md](docs/work/linear-f1-f6-mapping.md)。F1〜F6 の直接対応を確定してから反映 | **PARTIAL**（直接対応 ID は未特定。Linear ライブ再照会・書き込み・Done が未実施） |
| 2 | **H0-2 / HAC-CSV-1** | old_db / USER | STG 八王子の先頭。これより前の STG 行は進めない | **BLOCKED**（HAC-INPUT-2。完全 KNJO 未受領。同一 BAK 再実行と producer は禁止） |
| 3 | **H0-3b → Lane3 HAC → H3-9 → H3-11 → Lane 4** | USER | 2 の依存どおり | 待ち |
| 4 | **P1 → P2 → P3 → P4 → P5 → P6 → P7 → P8** | USER | go-live 依存。E1 / E2 は P4 の一部 | 待ち |

STG レーンの次は医院/ベンダーからの完全 KNJO 再取得、または城東主経路（JOU-G2-2 の Azure 承認）。H0-3b には入らない。独立して F1〜F6 の Linear 読み取り照合と、下記受入残の実行前提・証跡の確認を進められる。S09 / V04 / clinical E2E の実行は各設計の環境・承認条件を満たしてから行う。Linear 書き込みと Done は USER。

認証の残検証と開始条件は [todo-fix-auth.md](todo-fix-auth.md) を正本とする。実DB並行・初回管理者SQL実行・全経路の実DB返却データ分離・2サーバー即時失効・対象環境のメール・負荷測定は未検証。DB/環境操作は個別の実行条件に従う。

---

<a id="acceptance-remains"></a>

## 1. 受入残（PASS にしていない）

UAT / E2E を PASS にしない。正本は [UAT-DOMAIN-STATUS.md](docs/ops/testing/UAT-DOMAIN-STATUS.md)。

| ID | 残 | 状態 |
|----|----|------|
| **QA-UAT-S09-FIXTURE** | ブラウザ #2–#6 を再実行して帰属を証明する | S09 は BLOCKED（compose 停止のためブラウザ未実行） |
| **QA-UAT-V04-RETEST** | disposable clinic で CRUD/DELETE をブラウザ再実行する | V04 は UNKNOWN（browser runtime 未実行） |
| **QA-FULL-CLINICAL-E2E** | `--clinical` と e2e.yml full job を実行する | E2E は未証明（APP_ENV=test + E2E_LOGIN_PASSWORD + 起動済み stack が必要） |

設計: [S09-FIXTURE-DESIGN.md](docs/ops/testing/S09-FIXTURE-DESIGN.md) · [CLINICAL-E2E-DESIGN.md](docs/ops/testing/CLINICAL-E2E-DESIGN.md)。

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
| （現在の確認済み未対応項目なし） | — | — | — | — | — |

これは現在の製品全体に不具合がないという判定ではない。認証レビューの修正計画は [todo-fix-auth.md](todo-fix-auth.md) を参照する。対応済み項目は本節に残さず、履歴は Git と `reports/uat-YYYY-MM-DD/` を参照する。

<a id="human-lane"></a>

## 6. PO / 人間レーン（旧 todo-po.md）

実行 SoT は Linear hub [BRT-4](https://linear.app/baritechllc/issue/BRT-4) · Project ノア動物病院電子カルテ。人間ゲート（UAT・PO 確認）も Linear を正本とし、別の Open 行台帳を再構築しない。本ファイルの受入残・USER ゲートを入口にする。

会社側索引: CorpVault `50_Projects/ノア動物病院電子カルテ/05_Linearマップ.md`。旧詳細本文は Git 履歴。

<a id="refactor-constraints"></a>

## 7. FE 維持制約

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
| [docs/work/linear-f1-f6-mapping.md](docs/work/linear-f1-f6-mapping.md) | F1〜F6 対応案。Linear は UNKNOWN |
| [docs/ops/testing/S09-FIXTURE-DESIGN.md](docs/ops/testing/S09-FIXTURE-DESIGN.md) | S09 helper 設計 |
| [docs/ops/testing/CLINICAL-E2E-DESIGN.md](docs/ops/testing/CLINICAL-E2E-DESIGN.md) | clinical E2E 設計 |
| [docs/ops/testing/UAT-DOMAIN-STATUS.md](docs/ops/testing/UAT-DOMAIN-STATUS.md) | UAT 集計の正本 |
| [製品 FAIL](#product-bugs) | 確認済み製品 FAIL |
| [docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md) | ローカル handoff |
| [docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md) | STG 破壊境界 |

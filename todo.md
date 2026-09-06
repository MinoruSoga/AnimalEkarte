# タスク台帳 — Linear が正本

更新日: 2026-09-06

| 項目 | 値 |
|------|-----|
| **実行 SoT** | Linear Team **Baritech** · Project **ノア動物病院電子カルテ** · hub **[BRT-4](https://linear.app/baritechllc/issue/BRT-4)** |
| **セキュリティ修正** | **[BRT-226](https://linear.app/baritechllc/issue/BRT-226)**（Review · `origin/main` 済み · Done は人間） |
| **本ファイルの範囲** | repo と強く結び付く **未完了作業の入口** |

状態・Done は Linear を正本とする。行値・秘密は書かない。完了項目は削除する。

エージェントは PlanetScale、共有 STG apply、`DROP SCHEMA`、本番 cutover、`make reset`、八王子 CSV の producer 出力を実行しない。push / dispatch / Linear Done / 秘密変更は明示承認が必要。

claim は ID ごとに初回編集前に取得する。エージェントは claim を削除しない。

この更新の claim: `claim/LEDGER-TODO-PRUNE`。

---

## 対応順（実行キュー）

上から 1 件だけ着手する。USER / old_db に当たったら止めて提示する。deferred はキューに入れない。

| 順 | ID | 実行者 | なぜこの順 | 状態 |
|----|----|--------|------------|------|
| 1 | **META-LINEAR-APPLY** | USER | repo の対応案は [linear-f1-f6-mapping.md](docs/work/linear-f1-f6-mapping.md)。書き込みと Done は USER | **BLOCKED**（Linear MCP / `LINEAR_API_KEY` なし。公開ページはログイン壁。エージェントは書かない） |
| 2 | **H0-2 / HAC-CSV-1** | old_db / USER | STG 八王子の先頭。これより前の STG 行は進めない | **BLOCKED**（HAC-INPUT-2。完全 KNJO 未受領。同一 BAK 再実行と producer は禁止） |
| 3 | **H0-3b → Lane3 HAC → H3-9 → H3-11 → Lane 4** | USER | 2 の依存どおり | 待ち |
| 4 | **P1 → P2 → P3 → P4 → P5 → P6 → P7 → P8** | USER | go-live 依存。E1 / E2 は P4 の一部 | 待ち |

次は医院/ベンダーからの完全 KNJO 再取得、または城東主経路（JOU-G2-2 の Azure 承認）。H0-3b には入らない。Linear 書き込みと Done は USER。

---

## 1. 受入残（PASS にしていない）

helper / 再実行スライスは済。UAT / E2E を PASS にしない。正本は [UAT-DOMAIN-STATUS.md](docs/ops/testing/UAT-DOMAIN-STATUS.md)。

| ID | 残 | 状態 |
|----|----|------|
| **QA-UAT-S09-FIXTURE** | HTTP/CLI とブラウザ再実行 | S09 は BLOCKED |
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
| **BE-RC-021** | 新規 export へ GoDoc を追加 |
| **BE-RC-023** | clinic validator の `init()` 副作用が出たら constructor 登録へ |
| **BE-RC-034** | 対象 command 変更時に helper へ寄せる |
| **BE-RC-035** | 当該 auth test 変更時に `testdb.Truncate` へ |

---

## 参照

| 文書 | 役割 |
|------|------|
| [`todo-now.md`](todo-now.md) | Astra F1〜F6 の完了履歴 |
| [docs/work/linear-f1-f6-mapping.md](docs/work/linear-f1-f6-mapping.md) | F1〜F6 対応案。Linear は UNKNOWN |
| [docs/ops/testing/S09-FIXTURE-DESIGN.md](docs/ops/testing/S09-FIXTURE-DESIGN.md) | S09 helper 設計 |
| [docs/ops/testing/CLINICAL-E2E-DESIGN.md](docs/ops/testing/CLINICAL-E2E-DESIGN.md) | clinical E2E 設計 |
| [docs/ops/testing/UAT-DOMAIN-STATUS.md](docs/ops/testing/UAT-DOMAIN-STATUS.md) | UAT 集計の正本 |
| [`bug.md`](bug.md) | 確認済み製品 FAIL |
| [docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md) | ローカル handoff |
| [docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md) | STG 破壊境界 |

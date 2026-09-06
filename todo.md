# タスク台帳 — Linear が正本

更新日: 2026-09-06（調査・設計・スキル正本）

| 項目 | 値 |
|------|-----|
| **実行 SoT** | Linear Team **Baritech** · Project **ノア動物病院電子カルテ** · hub **[BRT-4](https://linear.app/baritechllc/issue/BRT-4)** |
| **直下整理** | **[BRT-105](https://linear.app/baritechllc/issue/BRT-105)**（Done） |
| **セキュリティ修正** | **[BRT-226](https://linear.app/baritechllc/issue/BRT-226)**（Review · `origin/main` 済み · Done は人間） |
| **会社側ログ** | CorpVault `50_Projects/ノア動物病院電子カルテ/` |
| **完了履歴** | Git 履歴と Linear の完了 Issue。完了項目は本ファイルへ残さない |
| **本ファイルの範囲** | repo と強く結び付く **未完了作業の入口** |

## 使い方

- 状態・担当・Done は Linear を正本とする。repo 記録だけで Done にしない。
- 確認済みの製品 FAIL は [`bug.md`](bug.md) に記録し、その後 Linear Issue 化する。
- 行値、患者情報、飼主情報、パスワード、接続資格情報は書かない。
- 完了した項目は削除する。経緯が必要な場合は Git 履歴を参照する。
- 現行ローカル handoff 表の正本は「STG 観測」節。手順は [`OLD_DB_HANDOFF_LOCAL.md`](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md)。
- live 共有 STG への再接続はしていない。判定は owner-only report、gitignored 配置、GitHub Actions、zsh 履歴、sibling `old_db` 台帳に限る。

### 実行者

| 実行者 | 意味 |
|--------|------|
| **old_db** | sibling producer。この repo からは CSV を生成できない |
| **USER** | 共有 STG、PlanetScale、医院調整、本番 cutover、workflow dispatch、Linear Done |
| **agent** | この repo のローカル作業と、許可された Linear / GitHub の読み取り照合。共有 STG には触れない |

エージェントは PlanetScale、共有 STG への apply、schema 再作成、`DROP SCHEMA`、本番 cutover、`make reset`、八王子 CSV の producer 出力を実行しません。

着手依頼の範囲内では、ローカル調査・設計・可逆な修正・許可された局所検証に追加承認は不要。本ファイルへの掲載だけでは全件の実行承認にならない。Linear 更新、外部投稿、push / merge / dispatch、秘密変更、共有環境操作は明示承認が必要。migration 適用、codegen、全体検証などの自動実行禁止は [プロジェクトルール](.claude/CLAUDE.md) に従う。

claim は作業 ID ごとに初回編集前に確認・取得し、既存 claim があれば BLOCKED。エージェントは claim ブランチを削除しない。main 統合または明示中止のあと USER が解放する。同じファイルを扱う ID は直列実行する。

この更新の claim: `claim/META-LINEAR-F1-F6` · `claim/QA-UAT-EVIDENCE-SYNC` · `claim/QA-UAT-S09-FIXTURE` · `claim/QA-FULL-CLINICAL-E2E` · `claim/SKILL-GO-REFS` · `claim/SKILL-API-EXAMPLES` · `claim/SKILL-REVIEW-EVIDENCE` · `claim/SKILL-ENTRY-SLIM` · `claim/SKILL-DEDUP` · `claim/SKILL-CODEX-MIRROR` · `claim/SKILL-REEVAL`。解放は USER。

---

## 1. 設計済み・実装は承認待ち

エージェントが閉じられる調査・設計・スキル正本修正は完了。実装・UAT 再実行・Linear Done・push は残さない（USER ゲートへ）。

| ID | 成果物 | 次の USER 操作 |
|----|--------|----------------|
| **META-LINEAR-F1-F6** | [linear-f1-f6-mapping.md](docs/work/linear-f1-f6-mapping.md)。Linear 本体は UNKNOWN | 明示承認後に必要な ID だけ紐付け。Done は USER |
| **QA-UAT-EVIDENCE-SYNC** | [UAT-DOMAIN-STATUS.md](docs/ops/testing/UAT-DOMAIN-STATUS.md) を最終実行と現行 open FAIL で分離。V04 は UNKNOWN。製品 FAIL は追加していない | V04 再実行と Linear 更新は別承認 |
| **QA-UAT-S09-FIXTURE** | [S09-FIXTURE-DESIGN.md](docs/ops/testing/S09-FIXTURE-DESIGN.md)。S09 は BLOCKED のまま | helper 実装の承認 |
| **QA-FULL-CLINICAL-E2E** | [CLINICAL-E2E-DESIGN.md](docs/ops/testing/CLINICAL-E2E-DESIGN.md)。auth smoke と混同しない | fixture 実装と full suite 実行は別承認 |

閉じた項目（履歴は git）: 上記 4 ID の調査/設計、スキル監査一式（GO-REFS〜REEVAL）、`CI-BE-DBORTX-INVENTORY`、`CI-K6-SUMMARY-SCHEMA`、`CI-K6-RUNTIME-CLOSEOUT`、`LEDGER-CI-EVIDENCE-SYNC`、`FE-CLINICAL-PLAN-SELECT-LABELS`、`DOC-MANUAL-SOURCE-SYNC`、`CLINICAL-IRREVERSIBLE-GUARD`（確認のみ）、`FE-TRIMMING-GUARDS`、`BE-RC-036`、`LEDGER-TODO-NOW-POINTER`。

### 記録済み証拠と現在の確認範囲

- 前回 CI closeout 記録時の HEAD / `origin/main`: `bc5626960399b720292bbea2a0d9b9f902d61dad`。下記 run は前回記録の転記で、今回 GitHub へ再照会していない。各 run の対象 SHA と現行差分への適用可否は利用時に確認する。
- 今回の編集前観測（2026-09-06）: ローカル HEAD `b4bbd71f93e86c7c331fd07df8e4c2375379d119`、ローカル remote-tracking ref `origin/main` は `fb49393b7bad71c67794ac871c0afd0b1b1ae442`。fetch 未実施で、リモート最新状態の証明ではない。
- Performance Tests / run `34025435577`（`workflow_dispatch`）: workflow success。endpoint k6、spike k6、aggregate validation、always-run cleanup、Lighthouse、summary が同一 run で success。
- 同 run の活動量: endpoints `http_reqs=5755`、`iterations=1918`、`checks=11508`、`successful_logins=1`。spike `http_reqs=825`、`iterations=824`、`checks=1648`、`successful_logins=1`。credential・body・cookie・token は記録しない。
- 旧 run `34020760108`（head `9ed814fc0`）は旧 validator による aggregate failure。現行 closeout の正本は `34025435577`。
- PR #369 / run `34025435907`: `Backend Test (remaining)` と集約 `Backend` が success。push 側 CI `34025433324` は paths-filter で Backend shards を skip したため、inventory の証明は PR run を正とする。
- 上記は CI/負荷経路の証拠であり、共有 STG、PROD、full clinical E2E、release readiness の証明ではない。

---

## 2. USER ゲート（秘密・本番・外部環境）

外部状態は実行直前に provider・Linear・権限・対象環境を再確認する。エージェントは秘密値の作成・表示・投入、共有 STG/PROD apply、production 構築、go-live を自動実行しない。

| 順 | ID | 実行者 | 状態 | 完了条件 |
|----|----|--------|------|----------|
| P1 | **SEC-SECRETS-5 / #89 / #97** | USER | 4系統 rotation receipt 未記入 | 新発行→投入→再 deploy→health→旧値 revoke→旧値拒否。値は記録しない |
| P2 | **#253 / U12 PROD-SETUP** | USER / 開発 | Production 未構築 | Cloudflare 本番、Required reviewers、workflow、rollback、backup rehearsal、URL/CI receipt |
| P3 | **#250 PROD-DATA-MIGRATION** | USER / 開発 | 事前準備待ち | rehearsal、最終 import、入力停止、backup/rollback、件数・clinic_id・金額突合 |
| P4 | **#254 AUTHENTICATED-UAT** | USER / agent | full UAT 未証明 | `QA-UAT-EVIDENCE-SYNC` で証跡を照合し、全業務 scenario の受入結果を確定。残 FAIL の延期は下記例外条件を満たす場合のみ。PARTIAL / BLOCKED / UNKNOWN を PASS にしない。A4 と混同しない |
| P5 | **#255 STAFF-PROVISION** | USER | 入力未記入 | roster、email 方針、clinic、role、actor、環境承認。PII-free receipt |
| P6 | **#258 / U1〜U12 DELIVERY** | USER | 最終承認待ち | P1・P2 と契約責任者の非機密事実を `DELIVERY_PACKAGE.md` へ反映 |
| P7 | **#256 / U13 TRAINING** | USER | 操作説明会未完 | 日程・形式・範囲・結果・opaque receipt |
| P8 | **#257 GOLIVE** | USER | HOLD | P1〜P7 の受入と第2段階の条件を USER が確認後に切替日。未解消の重大 FAIL、必要証跡不足、当日最終 import 未達なら No-Go |
| E1 | **QA-UAT-LSTEP-REAL** | USER | 外部環境待ち | write 有効な LSTEP で S01 同期と V05-17 remove。行値・credential は残さない |
| E2 | **QA-UAT-LINE-IDTOKEN** | USER | mock 外・未証明 | 実 LINE idToken で link / 409 / 期限切れ 400。token と個人情報は残さない |

P4 の延期例外: 臨床安全、会計金額の整合、clinic / owner / pet / staff のデータ分離、認証・権限、データ消失に関わる未解消 FAIL は go-live 前の解消対象とする。それ以外も、重大度・業務影響・回避策・受容責任者・対応担当・期限・再検証条件を Linear に残し、USER の明示受容がある場合だけ延期できる。延期は PASS と区別し、E1 / E2 を含む必須範囲の未検証を自動的に受容済みにしない。

---

## 3. STG 実データ（USER / old_db）

ここには未完了作業と、完了判定に必要な証跡確認を残す。ローカル配置の不在と外部での未実施は区別する。report / 履歴が無いだけなら **UNKNOWN（実施有無未確認）** とし、再 apply の前に USER が既存証跡と対象状態を確認する。

対象は八王子 `clinic_id=1` と城東 `clinic_id=2`。`shikishima` / `hakobuneco` は UAT 対象外だが STG には載っている。

| 優先 | ID | 実行者 | 状態 | blocked-by | 根拠 |
|------|----|--------|------|------------|------|
| 1 | **H0-2 / HAC-CSV-1** | old_db / USER | ローカル未配置・producer 実施有無 UNKNOWN | USER が producer の現行証跡を確認 | 前回観測では AE `hachioji/` 空、old_db csv-export に八王子 run なし |
| 2 | **H0-3b / H1-2** | USER | 待ち | H0-2 | 八王子 bundle が無い |
| 3 | **AE-STG-UAT-LANE3-HAC** | USER | UNKNOWN・投入判断待ち | 現行状態確認、H0-2、H0-3b / H1-2、下記 STG 実行ゲート | 八王子の STG apply report なし |
| 4 | **H3-9 staff attach apply** | USER | 入力あり・apply 実施有無 UNKNOWN | 現行 attach 状態確認、下記 STG 実行ゲート | 前回観測では roster / secrets は 0600、apply stdout / zsh 履歴なし。値は読まない |
| 5 | **H3-11 画面確認** | USER | UNKNOWN・証跡未取得 | H3-9 の成否と対象スタッフの自医院ログインを確認 | 飼主検索の確認ログなし。行値は残さない |
| 6 | **Lane 4** | 医院スタッフ / USER | 完了未証明 | 両院の Lane 3 verify、H3-11、医院調整 | 2026-09-05 の local UAT は Lane 4 ではない |

索引から外したもの:

- **AE-OLD-DB-MR-UNIQ（現行 rehearsal 3院）** — 非 NULL `medical_record_id` 重複 0。STG apply が unique index を通過
- **AE-STG-UAT-LANE3-JOU の 21表投入** — owner-only apply report `PASS`
- **H3-7 の敷島 / Hako bu neco** — STG apply `PASS`（UAT 対象外）

任意: H1-3 A4 UI rehearsal、H1-4 F8 G4、H3-10 未来日 shift。H1-5 は規則（失敗 bundle を STG へ送らない）。

### 受け入れ（残レーン）

| ID | 目的 | 受け入れ |
|----|------|----------|
| H0-2 / HAC-CSV-1 | 八王子 21 CSV + manifest | `_old_db_handoff/hachioji/`。各 CSV 512MiB 未満。旧 7 CSV 不使用。非 NULL `medical_record_id` 重複なし |
| H0-3b / H1-2 | 隔離確認とローカル rehearsal | `CLINIC_CODE=hachioji` で `make old-db-handoff-check` が PASS |
| AE-STG-UAT-LANE3-HAC | 共有 STG 投入 | maintenance window。失敗時は成功側を残す |
| H3-9 | staff attach | `make stg-uat-staff-attach-preflight` → `make stg-uat-staff-attach` |
| H3-11 | 飼主検索 | 対象医院へ切り替え、実データ表示。行値は残さない |
| Lane 4 | 並行運用 | 必須4業務を両院で連続5営業日。上限8週。local UAT を完了としない |

STG 実行ゲート（USER が実行直前に確認）: H3-1 他エンジニア確認・医院連絡と、H3-2 検証済み full backup・復元担当は repo に作業票がなく UNKNOWN。対象環境、data owner、operator、maintenance window、backup / restore 証跡、rollback 判断、対象操作の承認を確認できるまで新たな apply を実行しない。証跡不在を理由に再投入しない。正本は [STG 手順の停止ゲート](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md#2-pre-deploy-stop-gates)。

### STG 観測（2026-09-06）

以下は前回のローカル資料による観測記録。今回の台帳修正では handoff 配置・秘密ファイル・live STG を再照会していない。過去の apply PASS は現行残存を証明しない。行値は書かない。

| レーン | 状態 | 根拠 |
|--------|------|------|
| Lane 0 城東 | 配置済み | `_old_db_handoff/jouto/` 21 CSV + manifest。layout check PASS |
| Lane 0 八王子 | **未** | AE `hachioji/` 空 |
| Lane 1 城東ほか | ローカル apply 済み | `*-apply.json` が `targetHost=db` / `ekarte_db` で PASS（2026-09-05） |
| Lane 1 八王子 | H0-2待ち | bundle が無い |
| Lane 2 コード | **完了** | STG UAT importer / Make 経路は main 済み |
| Lane 3 城東 CSV | **投入済み** | `stg-uat-apply` `PASS`、2026-09-05 00:19 JST、PlanetScale STG |
| Lane 3 敷島 / Hako | **投入済み（対象外）** | 同日の STG apply `PASS` |
| Lane 3 八王子 | **UNKNOWN** | apply report なし。live 未照会 |
| Lane 4 | **完了未証明** | 5営業日の STG 現場証跡なし。[`bug.md`](bug.md) の未対応 FAIL なしだけでは完了しない |

| clinic | 配置 | layout | STG apply | ローカル apply |
|--------|------|--------|-----------|----------------|
| jouto | 21 CSV + manifest | PASS | PASS（2026-09-05 00:19 JST） | PASS（`db` / `ekarte_db`） |
| hachioji | なし | FAIL | UNKNOWN（証跡なし） | 証跡なし |
| shikishima | 21 CSV + manifest | PASS | PASS | PASS |
| hakobuneco | 21 CSV + manifest | PASS | PASS | PASS |

H3-3 TTL 資格情報は前回観測で gitignored `scripts/stg-uat-old-db-handoff.local.env` が 0600（TTL そのものは未証明）。H3-4〜H3-6 は城東の過去 PASS を参考証拠とし、次回実行や別医院への充足を推定しない。USER が対象ごとに現行条件を再確認する。2026-07-22 の八王子 F6 disposable apply `PASS` は現行 H0-2 ではない。

契約: `make stg-uat-import` / `make stg-uat-handoff`。接続値は repo へ書かない。破壊境界は [STG_PLANETSCALE_SEED_RUNBOOK.md](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md)。

第2段階へ進む条件: 両院投入と verify、対象スタッフの自医院ログイン、必須4業務の5営業日、ブロッキング製品バグなし、STG 入力を本番へ移さない理解、F6 `PASS` / `TRUSTED_CANDIDATE` 維持、移行日と当日出力が決まっていること。

---

## 4. 触ったときだけ / deferred

横断キャンペーンにしない。該当機能を変更するときだけ閉じる。

| ID | 種別 | 再開条件 |
|----|------|----------|
| **TASK-444** | FE deferred | generated/models の domain import 267件。公開契約・codegen・consumer 移行計画が別単位で揃うまで一括置換しない |
| **BE-RC-005** | BE MEDIUM | 新規・変更 service から 5xx 二重ログを解消。既知 4xx は service でログしない |
| **BE-RC-009** | BE MEDIUM | 新規 consumer または対象機能変更時に利用側最小 port へ分割 |
| **BE-RC-014** | BE MEDIUM | typed error が使えるようになったら `errors.As` へ。`err.Error()` 例外を増やさない |
| **BE-RC-015** | BE LOW | 新規・変更面から package.Type stutter を避ける |
| **BE-RC-017** | BE LOW | 対象 repository 変更時に unexported update + typed command |
| **BE-RC-019** | BE LOW | lab / hospitalization 等の業務能力境界が成立する変更時だけ分割を検討 |
| **BE-RC-021** | BE LOW | 新規 export へ GoDoc を追加 |
| **BE-RC-023** | BE LOW | clinic validator の global `init()` でテスト順副作用が出たら constructor 登録へ |
| **BE-RC-034** | BE LOW | 対象 command 変更時に auth / clinic / csvimport helper へ寄せる |
| **BE-RC-035** | BE LOW | 当該 auth test 変更時に `testdb.Truncate` へ |

`todo-refactor.md` のローカル `pnpm build` 未実行は独立タスクにしない。`readiness-report.md` の 300 行超 file も行数だけで分割しない。

---

## 5. スキル監査

正本修正は完了。再評価: [skill-reeval-2026-09-06.md](docs/work/skill-reeval-2026-09-06.md)。Linear 反映は USER。

閉じた ID: `SKILL-GO-REFS`、`SKILL-API-EXAMPLES`、`SKILL-REVIEW-EVIDENCE`、`SKILL-ENTRY-SLIM`、`SKILL-DEDUP`、`SKILL-CODEX-MIRROR`（正本修正済み。ミラーは `.agents/` gitignore。commit hook / `sync-agents-skills.sh`）、`SKILL-REEVAL`。

---

## 次の一手

1. USER: [linear-f1-f6-mapping.md](docs/work/linear-f1-f6-mapping.md) を見て、必要な Linear 紐付けだけ承認する。Done は USER。
2. USER: S09 helper / clinical E2E fixture の実装を承認するまで、UAT/E2E を PASS にしない。
3. USER / old_db: **HAC-CSV-1** の現行証跡を確認する。agent はこの repo から八王子 CSV を出さない。
4. USER: STG 実行ゲートと **H3-9** の現行成否を確認する。

---

## 参照

| 文書 | 役割 |
|------|------|
| [`todo-now.md`](todo-now.md) | Astra F1〜F6 の完了履歴と本ファイルへの入口 |
| [docs/work/linear-f1-f6-mapping.md](docs/work/linear-f1-f6-mapping.md) | F1〜F6 と ledger ID の対応案。Linear は UNKNOWN |
| [docs/ops/testing/S09-FIXTURE-DESIGN.md](docs/ops/testing/S09-FIXTURE-DESIGN.md) | S09 `completed_at` helper 設計。実装は承認後 |
| [docs/ops/testing/CLINICAL-E2E-DESIGN.md](docs/ops/testing/CLINICAL-E2E-DESIGN.md) | clinical E2E 設計。auth smoke と分離 |
| [`todo-po.md`](todo-po.md) | 入口ポインタ |
| [`bug.md`](bug.md) | 確認済み製品 FAIL |
| [docs/ops/deploy/CLINIC_CSV_IMPORT.md](docs/ops/deploy/CLINIC_CSV_IMPORT.md) | F6 21表 cutover |
| [docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md) | `_old_db_handoff` のローカル隔離 |
| [docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md](docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md) | seed と 21表の境界 |
| [docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md) | STG 再作成、直結、破壊境界 |
| [docs/ops/deploy/STG-DEMO-DATA-LIFECYCLE.md](docs/ops/deploy/STG-DEMO-DATA-LIFECYCLE.md) | STG データのライフサイクル |
| [docs/ops/deploy/STAFF_ACCOUNT_PROVISIONING.md](docs/ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) | staff account 運用 |
| [docs/ops/deploy/A4_UI_REHEARSAL.md](docs/ops/deploy/A4_UI_REHEARSAL.md) | 隔離画面確認 |
| [docs/ops/deploy/LOCAL_DB_RESET.md](docs/ops/deploy/LOCAL_DB_RESET.md) | ローカル reset |
| [docs/ops/infra/staging/runbook.md](docs/ops/infra/staging/runbook.md) | STG 障害初動 |

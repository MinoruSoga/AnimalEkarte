# タスク台帳 — Linear が正本

更新日: 2026-09-06

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
| **agent** | この repo のローカル専用。共有 STG には触れない |

エージェントは PlanetScale、共有 STG への apply、schema 再作成、`DROP SCHEMA`、本番 cutover、`make reset`、八王子 CSV の producer 出力を実行しません。

claim は実装単位ごと。エージェントは claim ブランチを削除しない。main 統合または明示中止のあと USER が解放する。

この更新の claim: `claim/CI-K6-RUNTIME-CLOSEOUT`、`claim/LEDGER-CI-EVIDENCE-SYNC`。

---

## 1. いま進める（agent / 直後の USER）

| 優先 | ID | 実行者 | 状態 | 次の一手 | 完了条件 |
|------|----|--------|------|----------|----------|
| 1 | **META-LINEAR-F1-F6** | USER / agent | 未照会 | Linear で F1〜F6 と本ファイル ID の重複を確認し、必要な ID だけ紐付ける | Linear 上の対応が残作業と一致。repo 記録だけで Done にしない |
| 2 | **QA-UAT-S09-FIXTURE** | agent | BLOCKED | `completed_at` を指定できる承認済み fixture API または scoped UAT helper を設計する | 既存会計・DB・システム時計を直接変えず S09 #2〜#6 を再実行できる |
| 3 | **QA-FULL-CLINICAL-E2E** | agent / USER | 未準備 | auth smoke とは別に clinical/data-dependent E2E の fixture・allowlist・cleanup を設計する | 実行は別承認。full E2E 成功と秘密非出力を証拠化 |

閉じた項目（履歴は git）: `CI-BE-DBORTX-INVENTORY`、`CI-K6-SUMMARY-SCHEMA`、`CI-K6-RUNTIME-CLOSEOUT`、`LEDGER-CI-EVIDENCE-SYNC`、`FE-CLINICAL-PLAN-SELECT-LABELS`、`DOC-MANUAL-SOURCE-SYNC`、`CLINICAL-IRREVERSIBLE-GUARD`（確認のみ）、`FE-TRIMMING-GUARDS`（現 HEAD で充足）、`BE-RC-036`、`LEDGER-TODO-NOW-POINTER`。

### 現在の証拠境界

- HEAD / `origin/main`: `bc5626960399b720292bbea2a0d9b9f902d61dad`
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
| P4 | **#254 AUTHENTICATED-UAT** | USER / agent | full UAT 未証明 | 全業務 scenario 完走、または残 FAIL を納品後対応へ隔離。A4 と混同しない |
| P5 | **#255 STAFF-PROVISION** | USER | 入力未記入 | roster、email 方針、clinic、role、actor、環境承認。PII-free receipt |
| P6 | **#258 / U1〜U12 DELIVERY** | USER | 最終承認待ち | P1・P2 と契約責任者の非機密事実を `DELIVERY_PACKAGE.md` へ反映 |
| P7 | **#256 / U13 TRAINING** | USER | 操作説明会未完 | 日程・形式・範囲・結果・opaque receipt |
| P8 | **#257 GOLIVE** | USER | HOLD | P1〜P7 が green のあと切替日。当日最終 import 未達なら No-Go |
| E1 | **QA-UAT-LSTEP-REAL** | USER | 外部環境待ち | write 有効な LSTEP で S01 同期と V05-17 remove。行値・credential は残さない |
| E2 | **QA-UAT-LINE-IDTOKEN** | USER | mock 外・未証明 | 実 LINE idToken で link / 409 / 期限切れ 400。token と個人情報は残さない |

---

## 3. STG 実データ（USER / old_db）

推測で「未」にしていた項目は観測で見直した。ここに残すのは **今も未完了と証明できるもの** だけ。

対象は八王子 `clinic_id=1` と城東 `clinic_id=2`。`shikishima` / `hakobuneco` は UAT 対象外だが STG には載っている。

| 優先 | ID | 実行者 | 状態 | blocked-by | 根拠 |
|------|----|--------|------|------------|------|
| 1 | **H0-2 / HAC-CSV-1** | old_db / USER | 未 | なし | AE `hachioji/` 空。old_db csv-export に八王子 run なし |
| 2 | **H0-3b / H1-2** | USER | 待ち | H0-2 | 八王子 bundle が無い |
| 3 | **AE-STG-UAT-LANE3-HAC** | USER | 待ち | H0-2、H0-3b / H1-2 | 八王子の STG apply report なし |
| 4 | **H3-9 staff attach apply** | USER | 入力あり・apply 証跡なし | なし | roster / secrets は 0600 で存在。apply stdout も zsh 履歴もなし |
| 5 | **H3-11 画面確認** | USER | 証跡なし | H3-9 が未ならログイン不能の可能性 | 飼主検索の確認ログなし。行値は残さない |
| 6 | **Lane 4** | 医院スタッフ / USER | 待ち | 両院の Lane 3。八王子が未 | 2026-09-05 の local UAT は Lane 4 ではない |

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

作業票が repo に無いもの（未実施とは限らない）: H3-1 他エンジニア確認と医院連絡、H3-2 検証済み full backup と復元担当。

### STG 観測（2026-09-06）

行値は書かない。live STG の残存は未照会。

| レーン | 状態 | 根拠 |
|--------|------|------|
| Lane 0 城東 | 配置済み | `_old_db_handoff/jouto/` 21 CSV + manifest。layout check PASS |
| Lane 0 八王子 | **未** | AE `hachioji/` 空 |
| Lane 1 城東ほか | ローカル apply 済み | `*-apply.json` が `targetHost=db` / `ekarte_db` で PASS（2026-09-05） |
| Lane 1 八王子 | H0-2待ち | bundle が無い |
| Lane 2 コード | **完了** | STG UAT importer / Make 経路は main 済み |
| Lane 3 城東 CSV | **投入済み** | `stg-uat-apply` `PASS`、2026-09-05 00:19 JST、PlanetScale STG |
| Lane 3 敷島 / Hako | **投入済み（対象外）** | 同日の STG apply `PASS` |
| Lane 3 八王子 | **未** | apply report なし |
| Lane 4 | **未** | 5営業日の STG 現場証跡なし。[`bug.md`](bug.md) の未対応 FAIL はなし |

| clinic | 配置 | layout | STG apply | ローカル apply |
|--------|------|--------|-----------|----------------|
| jouto | 21 CSV + manifest | PASS | PASS（2026-09-05 00:19 JST） | PASS（`db` / `ekarte_db`） |
| hachioji | なし | FAIL | なし | なし |
| shikishima | 21 CSV + manifest | PASS | PASS | PASS |
| hakobuneco | 21 CSV + manifest | PASS | PASS | PASS |

H3-3 TTL 資格情報は gitignored `scripts/stg-uat-old-db-handoff.local.env` が 0600 で存在（TTL そのものは未証明）。H3-4〜H3-6 は城東 PASS から充足とみなす。2026-07-22 の八王子 F6 disposable apply `PASS` は現行 H0-2 ではない。

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

## 5. スキル・AGENTS.md 監査の残

監査基準: [Rethinking skills and prompts for GPT-6 Astra](https://x.com/pvncher/status/2095991462416490862)。依存パッケージ内の Redux/Playwright/Recharts は対象外。

このセッションで閉じたもの: Docker 復旧の破壊例、フロント秘密/ログ例、DB 診断の秘密表示、一時改変の HEAD 巻き戻し、fresh-DB の共有 DROP、YAML 引用符、旧台帳書込み、モデル固定の手動差戻し、AGENTS.md の作業別案内、task-create / implement / browser-test の記録先。

残るのは現行実装との例合わせと、長い教材の入口整理。各 ID を個別に claim する。

| ID | 優先 | 対象 | 完了条件 |
|----|------|------|----------|
| **SKILL-CODEX-MIRROR** | 2 | sync 後の Codex ミラー、`pre-bash-commit-quality.js`、`scoped-verification-gates` ミラー | 正本修正後に同期し、生成物差分を読み取りで検出できる |
| **SKILL-GO-REFS** | 2 | golang-testing / golang-refactoring / test-generation / clinic-isolation-auditor / tdd-workflow | 現行 domain package・実在 mock へ案内。Session A/B と存在しない go-linting を必須にしない |
| **SKILL-API-EXAMPLES** | 2 | go-security / react-security / api-documentation / test-generation / performance-profiling / postgres-patterns | Wrap、Gin handler、CookieAuth/CSRF、応答、コンテナ内ポートを正本と一致 |
| **SKILL-REVIEW-EVIDENCE** | 2 | review / harness / implement-issue / go-security | 未測定 coverage や未実行テストを PASS と表示しない |
| **SKILL-ENTRY-SLIM** | 3 | 長い description のスキル群 | 発火条件を狭くし、手法列挙は必要時だけ読む |
| **SKILL-DEDUP** | 3 | react-security×security-checklist、deployment×ci-cd、implement×implement-issue 等 | 正本へ案内し二重管理を減らす |
| **SKILL-REEVAL** | 3 | 代表タスクでの再評価 | docs 誤字・局所 Go・FE テスト・migration・STG 準備で必要資料と停止判断が整合 |

個別確認の残: `refactor.md` の export 機械削除は Feature Indexing の公開契約確認へ。`deployment` の push/dispatch は承認境界をその場で明記。

---

## 次の一手

1. Linear で F1〜F6 と本ファイル ID を照合する。
2. old_db で **HAC-CSV-1** を出す。両院 Lane 4 のブロッカー。
3. 城東で STG ログインするなら **H3-9 apply** の成否を残し、**H3-11** で飼主検索だけ確認する（行値は残さない）。

---

## 参照

| 文書 | 役割 |
|------|------|
| [`todo-now.md`](todo-now.md) | Astra F1〜F6 の完了履歴と本ファイルへの入口 |
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

# タスク台帳 — Linear が正本

更新日: 2026-09-06

| 項目 | 値 |
|------|-----|
| **実行 SoT** | Linear Team **Baritech** · Project **ノア動物病院電子カルテ** · hub **[BRT-4](https://linear.app/baritechllc/issue/BRT-4)** |
| **直下整理** | **[BRT-105](https://linear.app/baritechllc/issue/BRT-105)**（Done） |
| **セキュリティ修正** | **[BRT-226](https://linear.app/baritechllc/issue/BRT-226)**（Review · `origin/main` 済み · Done は人間） |
| **会社側ログ** | CorpVault `50_Projects/ノア動物病院電子カルテ/` |
| **完了履歴** | Git 履歴と Linear の完了 Issue。完了項目は本ファイルへ残さない |
| **本ファイルの範囲** | repo と強く結び付く **未完了の STG 実データ運用作業だけ** |

## 使い方

- 状態・担当・次の一手は Linear を正本とする。
- 確認済みの製品 FAIL は [`bug.md`](bug.md) に記録し、その後 Linear Issue 化する。
- 行値、患者情報、飼主情報、パスワード、接続資格情報は書かない。
- 完了した項目は削除する。経緯が必要な場合は Git 履歴を参照する。
- 現行ローカル handoff 表の正本は本ファイルの「観測」節。手順は [`OLD_DB_HANDOFF_LOCAL.md`](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md)。

### 読み方

1. **残作業索引** — 未完了だけ。実行者と blocked-by を先に見る。
2. **観測** — この worktree で確認した事実。Linear / 共有 STG の中身は未照会。
3. **タスク詳細** — 受け入れ条件。完了したら索引から消す。
4. **レーン手順** — USER が共有 STG で踏む手順。エージェントは実行しない。

### 実行者

| 実行者 | 意味 |
|--------|------|
| **old_db** | sibling producer。この repo からは CSV を生成できない |
| **USER** | 共有 STG、PlanetScale、医院調整、本番 cutover |
| **agent** | この repo のローカル専用（layout check、importer 契約）。共有 STG には触れない |

---

## エージェント境界

エージェントは次を実行しません。PlanetScale、共有 STG への apply、schema 再作成、`DROP SCHEMA`、本番 cutover、`make reset`（ローカル DB 破壊）、八王子 CSV の producer 出力。

この worktree で実施済み（2026-09-06）:

- `make old-db-handoff-check` を 21 CSV + manifest の clinicCode / run ID 照合まで強化
- CSV preflight が `billings.medical_record_id` の非 NULL 重複を fail-closed にする（`AE-OLD-DB-MR-UNIQ` の producer 完了ではない）
- 既存 `jouto` / `shikishima` / `hakobuneco` の layout check は PASS。`hachioji` は manifest なしで FAIL（期待どおり）

claim 枝 `claim/TODO-STG-LEDGER` はエージェントが削除しない。main 統合後に USER が `git branch -D claim/TODO-STG-LEDGER` する。

---

## 残作業索引

| 優先 | ID | 実行者 | 状態 | blocked-by | エージェント |
|------|----|--------|------|------------|--------------|
| 1 | **AE-OLD-DB-MR-UNIQ** | old_db | 未完了 | なし | 不可。importer 側の回帰防止のみ実施済み |
| 2 | **H0-2 / HAC-CSV-1** | old_db / USER | 未完了 | なし | 不可。ローカル `hachioji/` は空 |
| 3 | **H0-3b / H1-2** | USER | 待ち | H0-2 | 不可。bundle 到着後の check / rehearsal |
| 4 | **AE-STG-UAT-LANE3-JOU** | USER | 未着手 | なし（ローカル jouto はある） | 不可。共有 STG 投入 |
| 5 | **AE-STG-UAT-LANE3-HAC** | USER | 待ち | H0-2、H0-3b / H1-2 | 不可 |
| 6 | **Lane 4** | 医院スタッフ / USER | 待ち | Lane 3 両院 | 不可。必須4業務を STG で連続5営業日 |

STG UAT 対象は八王子 `clinic_id=1` と城東 `clinic_id=2` のみ。ローカルの `shikishima` / `hakobuneco` は対象外。

---

## 観測（2026-09-06）

Linear の Issue 状態と共有 STG の中身は未照会。行値は書かない。

| レーン | 状態 | 根拠 |
|--------|------|------|
| Lane 0 城東 | ローカル bundle あり | `_old_db_handoff/jouto/` に 21 CSV + `manifest.json`。layout check PASS |
| Lane 0 八王子 | **未** | `_old_db_handoff/hachioji/` は空。layout check は manifest なしで FAIL |
| Lane 1 八王子 | H0-2待ち | bundle が無い |
| Lane 2 コード | **完了** | STG UAT importer / Make 経路は main 済み |
| Lane 3 共有STG | **未着手** | 城東のローカル配置は Lane 3 完了を意味しない。STG apply の証跡なし |
| Lane 4 並行運用 | **未着手** | Lane 3 待ち |
| ローカル製品 UAT | 別レーン | 2026-09-05 は local Docker（`uat/20260905`）。STG Lane 4 ではない。[`bug.md`](bug.md) の未対応 FAIL はなし |
| AE-OLD-DB-MR-UNIQ | **未完了** | producer 完了証跡なし。既存3院のローカル CSV は非 NULL `medical_record_id` 重複 0 |

配置があるだけでは formal F6（`PASS` / `TRUSTED_CANDIDATE`）ではない。`sourceComplete=false` の rehearsal bundle は正式 cutover に使わない。

### 現行ローカル handoff（gitignored）

存在確認と manifest の status のみ。`OLD_DB_HANDOFF_LOCAL.md` が参照する正本表。

| clinic | 配置 | layout check | manifest `status` | `handoffEligibility` | `sourceRunId` |
|--------|------|--------------|-------------------|----------------------|----------------|
| jouto | 21 CSV + manifest | PASS | `REHEARSAL_ONLY` | `REHEARSAL_ONLY` | `jouto-intake-20260822-01` |
| hachioji | なし | FAIL（manifest なし） | — | — | — |
| shikishima | 21 CSV + manifest | PASS | `REHEARSAL_ONLY` | `REHEARSAL_ONLY` | `jouto-intake-20260822-01` |
| hakobuneco | 21 CSV + manifest | PASS | `REHEARSAL_ONLY` | `REHEARSAL_ONLY` | `jouto-intake-20260822-01` |

---

## タスク詳細

### 1. AE-OLD-DB-MR-UNIQ

| 項目 | 内容 |
|------|------|
| 実行者 | old_db |
| 目的 | 1カルテにつき非 NULL の billing リンクは1件。余剰 billing はリンクを外す |
| 受け入れ | producer 出力の `billings.csv` で非 NULL `medical_record_id` が重複しない。スキーマの `idx_billings_medical_record_id_unique` に載る |
| 本 repo のガード | CSV preflight は重複を拒否する。これをもって本タスク完了とはしない |
| 観測 | 既存 `jouto` / `shikishima` / `hakobuneco` のローカル CSV は重複 0。八王子は未出力 |

### 2. H0-2 / HAC-CSV-1

| 項目 | 内容 |
|------|------|
| 実行者 | old_db / USER |
| 目的 | 八王子の医院 identity 付き 21 CSV と manifest を出す |
| 受け入れ | `_old_db_handoff/hachioji/` に 21 CSV + manifest。各 CSV は 512MiB 未満。旧 7 CSV は使わない |
| blocked-by | なし |

### 3. H0-3b / H1-2

| 項目 | 内容 |
|------|------|
| 実行者 | USER |
| 目的 | 八王子 bundle の隔離確認とローカル rehearsal |
| 受け入れ | `CLINIC_CODE=hachioji` で `make old-db-handoff-check` が PASS。必要な場合だけ preflight / apply / verify をローカル rehearsal する。失敗 bundle を共有 STG へ送らない（H1-5） |
| blocked-by | H0-2 |

### 4. AE-STG-UAT-LANE3-JOU

| 項目 | 内容 |
|------|------|
| 実行者 | USER |
| 目的 | 城東 bundle を共有 STG へ投入し、第1段階を開始する |
| 受け入れ | [STG_PLANETSCALE_SEED_RUNBOOK.md](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md) の Lane 3 手順を城東について完了。飼主検索が実データで表示される（行値はログへ残さない） |
| 注意 | ローカル jouto は `REHEARSAL_ONLY`。配置済み ≠ STG 投入済み |

### 5. AE-STG-UAT-LANE3-HAC

| 項目 | 内容 |
|------|------|
| 実行者 | USER |
| 目的 | maintenance window で八王子 bundle を共有 STG へ投入する |
| 受け入れ | Lane 3 手順を八王子について完了 |
| blocked-by | H0-2、H0-3b / H1-2 |

### 6. Lane 4

| 項目 | 内容 |
|------|------|
| 実行者 | 医院スタッフ / USER |
| 目的 | 必須4業務（検索、受付、カルテ、会計）を両院の STG で連続5営業日再現する |
| 受け入れ | 現場が切替可と判断し、第2段階の条件を満たす。上限は8週。ローカル製品 UAT の PASS を Lane 4 完了としない |
| blocked-by | Lane 3 両院 |

---

## STG実データ運用テスト

### 目的

共有 STG に旧システムの医院別データを投入し、医院スタッフが現行業務と並行して新システムを検証します。

- 現行業務の正本は旧システム。
- STG への入力は検証用であり、本番へ移しません。
- 対象は八王子 `clinic_id=1` と城東 `clinic_id=2`。
- 必須業務は検索、受付、カルテ、会計。
- 両院投入後、必須4業務を連続5営業日再現し、現場が切替可と判断したら第2段階へ進みます。上限は8週です。
- 本番 cutover は別工程です。入力は移行日の旧システム出力とし、`PASS` / `TRUSTED_CANDIDATE` 契約を緩めません。

### データ境界

- `002_master`: 医院骨格と参照マスタのみ。accounts や臨床行を含めない。
- `003_demo` / `004_staging`: 退役済み。STG 実データ運用には使わない。
- 業務データ: old_db の医院別 21 CSV と manifest を正本とする。
- `_old_db_handoff/`: ローカル隔離専用。Git、CI、イメージ、通常 migration へ載せない。
- 同一 STG DB で医院ごとに 10M ID band を分ける。
- STG の並行登録データを本番へコピー、昇格、差分追加しない。

### 禁止事項

- デモ臨床と実データを混在させない。
- `pscale connect` で remote DB を localhost に見せかけない。
- 本番用 F6 へ `--allow-local-rehearsal` を流用しない。
- 医院間で band、staff、account、患者、飼主、カルテ、会計を越境させない。
- 共有 STG への投入中に同じ STG で業務入力を続けない。
- PHI や資格情報を標準ログ、Git、Issue、チャットへ出さない。

---

## Lane 0 — 八王子入力

- [ ] **H0-2 / HAC-CSV-1**: 医院 identity 付き 21 CSV と manifest を old_db から出す。各 CSV が 512MiB 未満であることを確認する。
- [ ] **H0-3b**: `old-db-handoff-check` を通す。manifest なしの旧 `hachioji/` や旧 7 CSV は使わない。

配置先は repo 外または gitignored の次の領域です。

```text
backend/migrations/seeds/_old_db_handoff/hachioji/
```

---

## Lane 1 — 未完了のローカル証明

- [ ] **H1-2**: 八王子 bundle で preflight、apply、verify をローカル rehearsal する。
- [ ] **H1-5**: ローカルで失敗する bundle を共有 STG へ送らない。

必要な場合だけ実施します。

- [ ] **H1-3**: [A4_UI_REHEARSAL.md](docs/ops/deploy/A4_UI_REHEARSAL.md) で隔離画面確認を行う。
- [ ] **H1-4**: [F8_G4_FAILURE_REHEARSAL.md](docs/ops/deploy/F8_G4_FAILURE_REHEARSAL.md) で失敗側を確認する。本番 CSV は渡さない。

---

## Lane 3 — 共有 STG 投入（USER のみ）

[STG_PLANETSCALE_SEED_RUNBOOK.md](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md) の破壊境界に従います。城東を先に投入し、八王子は bundle 到着後の maintenance window で投入します。

- [ ] **H3-1**: 他エンジニアが STG を使用していないことを確認し、医院へ開始範囲を伝える。
- [ ] **H3-2**: 検証済み full backup と復元担当を作業票へ記録する。
- [ ] **H3-3**: TTL 付き接続資格情報を用意する。schema 再作成が必要な場合は USER 承認下で実施する。
- [ ] **H3-4**: `backend-deploy.yml` を `staging` で実行し、migration 成功を確認する。
- [ ] **H3-5**: `make stg-uat-skeleton` で医院骨格を作る。21 表の ID band を占有しないことを確認する。
- [ ] **H3-6**: 各医院の 10M ID band が 21 表すべてで空であることを確認する。
- [ ] **H3-7**: 医院ごとに `make stg-uat-import` を実行する。城東を先に行う。失敗時は手動 fallback の preflight / apply / verify を使う。
- [ ] **H3-8**: 一方が失敗した場合は成功側を残し、失敗側だけを修正する。
- [ ] **H3-9**: staff attach の preflight → apply を実行する。名簿と secrets は repo 外の mode 0600 を使う。
- [ ] **H3-11**: 対象医院へ切り替え、飼主検索が実データで表示されることを確認する。行値はログへ残さない。

予約も検証する場合だけ実施します。

- [ ] **H3-10 / AE-STG-UAT-SHIFT**: 未来日の shift を用意する。古い絶対日付をコピーしない。

### 実行契約

- `make stg-uat-import`（21 表の preflight → apply → verify。run sheet で確認した 6 seed ID を明示）
- 手動 fallback: `make stg-uat-csv-import-preflight` / `make stg-uat-csv-import` / `make stg-uat-csv-import-verify`
- `make stg-uat-staff-attach-preflight`
- `make stg-uat-staff-attach`

remote 実行では、対象 host と一致する確認値、SSL 設定、各コマンドの明示 sentinel が必要です。具体値は repo へ記録しません。

---

## Lane 4 — 第1段階の並行運用

- [ ] 期間中は STG DB を reset しない。
- [ ] backend デプロイ前に医院へ日時を知らせる。適用済み migration 編集や seed 差替えを行わない。
- [ ] STG で新規作成した予約、カルテ、会計を本番へ移さない。
- [ ] STG 障害時も旧システムで現行業務を継続する。
- [ ] ログへ PHI が出た場合は検証を止めて修正する。業務監査の正本は `audit_logs` とする。
- [ ] 製品不具合は `bug.md` へ記録し、Linear Issue 化する。本ファイルへバグ本文を複製しない。
- [ ] 必須4業務を両院で連続5営業日再現し、現場の切替判断を記録する。

---

## 第2段階へ進む条件

- 両院の bundle 投入と verify が完了している。
- 対象スタッフが自医院へログインできる。
- 必須4業務を連続5営業日再現している。
- ブロッキングな製品バグが残っていない。
- STG 入力を本番へ移さないことを医院が理解している。
- 本番 F6 の `PASS` / `TRUSTED_CANDIDATE` 契約が維持されている。
- 移行日、当日の旧データ出力、本番 run sheet が決まっている。

---

## 次の一手

1. old_db で **AE-OLD-DB-MR-UNIQ** を完了する（既存3院のローカル CSV は重複 0。八王子と producer 契約は未）。
2. USER が城東の **Lane 3** を実施する（ローカル jouto bundle はある。共有 STG 投入は未）。
3. old_db が **HAC-CSV-1** を作成し、八王子の Lane 0 / 1 / 3 を進める（ローカル `hachioji/` は空）。
4. 両院で Lane 4 を実施し、第2段階の移行日を決める。ローカル製品 UAT の PASS を Lane 4 完了としない。

---

## 参照

| 文書 | 役割 |
|------|------|
| [docs/ops/deploy/CLINIC_CSV_IMPORT.md](docs/ops/deploy/CLINIC_CSV_IMPORT.md) | F6 21表 cutover |
| [docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md) | `_old_db_handoff` のローカル隔離 |
| [docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md](docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md) | seed と 21表の境界 |
| [docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md) | STG 再作成、直結、破壊境界 |
| [docs/ops/deploy/STG-DEMO-DATA-LIFECYCLE.md](docs/ops/deploy/STG-DEMO-DATA-LIFECYCLE.md) | STG データのライフサイクル |
| [docs/ops/deploy/STAFF_ACCOUNT_PROVISIONING.md](docs/ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) | staff account 運用 |
| [docs/ops/deploy/A4_UI_REHEARSAL.md](docs/ops/deploy/A4_UI_REHEARSAL.md) | 隔離画面確認 |
| [docs/ops/deploy/LOCAL_DB_RESET.md](docs/ops/deploy/LOCAL_DB_RESET.md) | ローカル reset |
| [docs/ops/infra/staging/runbook.md](docs/ops/infra/staging/runbook.md) | STG 障害初動 |

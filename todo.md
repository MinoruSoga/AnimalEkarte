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
- live 共有 STG への再接続はしていない。判定は owner-only report、gitignored 配置、GitHub Actions、zsh 履歴、sibling `old_db` 台帳に限る。

### 実行者

| 実行者 | 意味 |
|--------|------|
| **old_db** | sibling producer。この repo からは CSV を生成できない |
| **USER** | 共有 STG、PlanetScale、医院調整、本番 cutover |
| **agent** | この repo のローカル専用。共有 STG には触れない |

エージェントは PlanetScale、共有 STG への apply、schema 再作成、`DROP SCHEMA`、本番 cutover、`make reset`、八王子 CSV の producer 出力を実行しません。

claim 枝 `claim/TODO-STG-LEDGER` はエージェントが削除しない。main 統合後に USER が `git branch -D claim/TODO-STG-LEDGER` する。

---

## 残作業索引

推測で「未」にしていた項目は、下の観測で見直した。ここに残すのは **今も未完了と証明できるもの** だけ。

| 優先 | ID | 実行者 | 状態 | blocked-by | 根拠 |
|------|----|--------|------|------------|------|
| 1 | **H0-2 / HAC-CSV-1** | old_db / USER | 未 | なし | AE `hachioji/` 空。old_db の csv-export に八王子 run なし。old_db `todo.md` の HAC-CSV-1 は未チェック |
| 2 | **H0-3b / H1-2** | USER | 待ち | H0-2 | 八王子 bundle が無い |
| 3 | **AE-STG-UAT-LANE3-HAC** | USER | 待ち | H0-2、H0-3b / H1-2 | 八王子の STG apply report なし |
| 4 | **H3-9 staff attach apply** | USER | 入力あり・apply 証跡なし | なし | roster / secrets は 0600 で存在。apply の stdout report も zsh 履歴もなし |
| 5 | **H3-11 画面確認** | USER | 証跡なし | H3-9 が未ならログイン不能の可能性 | 飼主検索の確認ログなし。行値は残さない |
| 6 | **Lane 4** | 医院スタッフ / USER | 待ち | 両院の Lane 3。八王子が未 | 2026-09-05 の local UAT は Lane 4 ではない |

STG UAT 対象は八王子 `clinic_id=1` と城東 `clinic_id=2`。`shikishima` / `hakobuneco` は対象外だが、STG には載っている（下表）。

索引から外したもの:

- **AE-OLD-DB-MR-UNIQ（現行 rehearsal 3院）** — ローカル CSV は非 NULL `medical_record_id` 重複 0。STG apply が unique index を通っている。producer に uniquify 実装あり。八王子分は HAC-CSV-1 に含める
- **AE-STG-UAT-LANE3-JOU の 21表投入** — owner-only apply report `PASS`
- **H3-7 の敷島 / Hako bu neco** — 同じく STG apply `PASS`（UAT 対象外）

---

## 観測（2026-09-06）

行値は書かない。live STG の残存は未照会。

### レーン

| レーン | 状態 | 根拠 |
|--------|------|------|
| Lane 0 城東 | 配置済み | `_old_db_handoff/jouto/` 21 CSV + manifest。layout check PASS。old_db `local-ae-v2` と manifest 契約ハッシュが一致 |
| Lane 0 八王子 | **未** | AE `hachioji/` 空。old_db csv-export に八王子ディレクトリなし |
| Lane 1 城東ほか | ローカル apply 済み | `*-apply.json` が `targetHost=db` / `ekarte_db` で PASS（2026-09-05）。八王子の H1-2 ではない |
| Lane 1 八王子 | H0-2待ち | bundle が無い |
| Lane 2 コード | **完了** | STG UAT importer / Make 経路は main 済み |
| Lane 3 城東 CSV | **投入済み** | `stg-uat-apply` `PASS`、lane `stg-uat-rehearsal`、2026-09-05 00:19 JST、PlanetScale STG |
| Lane 3 敷島 / Hako | **投入済み（対象外）** | 同日の STG apply `PASS` |
| Lane 3 八王子 | **未** | apply report なし |
| Lane 4 | **未** | 5営業日の STG 現場証跡なし。local UAT（`uat/20260905`）は別レーン。[`bug.md`](bug.md) の未対応 FAIL はなし |

2026-07-22 の八王子 F6 disposable apply `PASS` は **別物**（disposable DB、`clinicOrdinal=2`、現行 H0-2 ではない）。

### 現行ローカル handoff（gitignored）

| clinic | 配置 | layout check | STG apply | ローカル apply | manifest |
|--------|------|--------------|-----------|----------------|----------|
| jouto | 21 CSV + manifest | PASS | PASS（2026-09-05 00:19 JST） | PASS（`db` / `ekarte_db`） | `REHEARSAL_ONLY` / `jouto-intake-20260822-01` |
| hachioji | なし | FAIL | なし | なし | — |
| shikishima | 21 CSV + manifest | PASS | PASS | PASS | 同上 |
| hakobuneco | 21 CSV + manifest | PASS | PASS | PASS | 同上 |

`make stg-uat-handoff` / `make stg-uat-handoff-preflight` が zsh 履歴にある。wrapper は全医院。八王子は manifest なしで skip。

### Lane 3 手順の証跡

| 手順 | 判定 | 根拠 | 確度 |
|------|------|------|------|
| H3-1 他エンジニア確認・医院連絡 | 作業票なし | repo に記録なし | 不明（未実施とは断定しない） |
| H3-2 backup / 復元担当 | 作業票なし | 同上 | 不明 |
| H3-3 TTL 資格情報 | 入力あり | gitignored `scripts/stg-uat-old-db-handoff.local.env` が 0600 で存在。TTL そのものは未証明 | 中 |
| H3-4 staging deploy / migration | 充足とみなす | CSV apply が PlanetScale STG で PASS。`backend-deploy.yml` の staging 成功は 2026-09-05 09:18 JST（投入後）にもある | 中 |
| H3-5 skeleton | 充足とみなす | apply の preflight を通っている | 中 |
| H3-6 空 band | 充足とみなす | 9/4 失敗のあと 9/5 に PASS。一致しない非空 band は上書きしない契約 | 中 |
| H3-7 城東 / 敷島 / Hako | **PASS** | owner-only `stg-uat-apply` | 高 |
| H3-7 八王子 | **未** | bundle なし | 高 |
| H3-8 失敗側だけ修正 | 城東は失敗→成功 | 9/4 `FAILED_*` をリネームし、9/5 `PASS` | 高 |
| H3-9 staff attach | 入力あり・apply なし | roster `stg-uat-staff-attach-v1` と secrets が 0600。clinic 1–4 を含む。コマンド履歴なし。cmd は stdout のみで report ファイルを書かない | 高 |
| H3-10 未来日 shift | 任意・証跡なし | — | — |
| H3-11 飼主検索 | 証跡なし | — | — |

### AE-OLD-DB-MR-UNIQ

現行 rehearsal 3院は充足。残作業索引には置かない。

- AE 配置 CSV は非 NULL `medical_record_id` 重複 0
- その bundle の STG apply が unique index を通過
- old_db に uniquify（`scripts/lib/local-ae-billing-mr-uniq.mjs`）とテストがある
- AE `jouto` manifest の契約ハッシュは old_db `...-local-ae-v2` と一致（未加工 `rehearsal-current` とは不一致）
- 八王子は未出力のため、今後の HAC-CSV-1 で同じ一意制約を満たす

### Lane 1 任意

| 項目 | 判定 |
|------|------|
| H1-3 A4 UI rehearsal | 任意。`sensitive-local/a4-rehearsal-reports/` なし |
| H1-4 F8 G4 | 任意。`sensitive-local/f8-g4-rehearsal/` なし |
| H1-5 | 規則。失敗 bundle を STG へ送らない |

---

## タスク詳細

### 1. H0-2 / HAC-CSV-1

| 項目 | 内容 |
|------|------|
| 実行者 | old_db / USER |
| 目的 | 八王子の医院 identity 付き 21 CSV と manifest を出す |
| 受け入れ | `_old_db_handoff/hachioji/` に 21 CSV + manifest。各 CSV は 512MiB 未満。旧 7 CSV は使わない。非 NULL `medical_record_id` 重複なし |
| 観測 | old_db の HAC-CSV-1 は未。KNJO 破損により formal TRUSTED は別 blocker。現行 rehearsal 経路でも八王子 run が無い |

### 2. H0-3b / H1-2

| 項目 | 内容 |
|------|------|
| 実行者 | USER |
| 目的 | 八王子 bundle の隔離確認とローカル rehearsal |
| 受け入れ | `CLINIC_CODE=hachioji` で `make old-db-handoff-check` が PASS。必要な場合だけ preflight / apply / verify |
| blocked-by | H0-2 |

### 3. AE-STG-UAT-LANE3-HAC

| 項目 | 内容 |
|------|------|
| 実行者 | USER |
| 目的 | maintenance window で八王子 bundle を共有 STG へ投入する |
| blocked-by | H0-2、H0-3b / H1-2 |

### 4. H3-9 staff attach apply

| 項目 | 内容 |
|------|------|
| 実行者 | USER |
| 目的 | 移行 staffs.id へ account を後付けする |
| 受け入れ | `make stg-uat-staff-attach-preflight` → `make stg-uat-staff-attach` が成功する |
| 観測 | 入力ファイルはある。apply 成功の証跡はない |

### 5. H3-11 画面確認

| 項目 | 内容 |
|------|------|
| 実行者 | USER |
| 目的 | 対象医院へ切り替え、飼主検索が実データで表示されることを確認する |
| 注意 | 行値はログへ残さない |

### 6. Lane 4

| 項目 | 内容 |
|------|------|
| 実行者 | 医院スタッフ / USER |
| 目的 | 必須4業務（検索、受付、カルテ、会計）を両院の STG で連続5営業日再現する |
| 受け入れ | 現場が切替可と判断する。上限8週。local UAT の PASS を Lane 4 完了としない |
| blocked-by | Lane 3 両院（八王子が未） |

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

- [ ] **H0-2 / HAC-CSV-1**: 医院 identity 付き 21 CSV と manifest を old_db から出す。各 CSV が 512MiB 未満。非 NULL `medical_record_id` 重複なし。
- [ ] **H0-3b**: `old-db-handoff-check` を通す。manifest なしの旧 `hachioji/` や旧 7 CSV は使わない。

```text
backend/migrations/seeds/_old_db_handoff/hachioji/
```

---

## Lane 1 — 八王子のローカル証明

- [ ] **H1-2**: 八王子 bundle で preflight、apply、verify をローカル rehearsal する。
- [ ] **H1-5**: ローカルで失敗する bundle を共有 STG へ送らない。

必要な場合だけ:

- [ ] **H1-3**: [A4_UI_REHEARSAL.md](docs/ops/deploy/A4_UI_REHEARSAL.md)
- [ ] **H1-4**: [F8_G4_FAILURE_REHEARSAL.md](docs/ops/deploy/F8_G4_FAILURE_REHEARSAL.md)。本番 CSV は渡さない。

---

## Lane 3 — 共有 STG（USER のみ）

[STG_PLANETSCALE_SEED_RUNBOOK.md](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md) の破壊境界に従う。城東 CSV は投入済み。八王子は bundle 到着後の maintenance window。

残チェック:

- [ ] **H3-9**: staff attach の preflight → apply。名簿と secrets は gitignored・mode 0600。
- [ ] **H3-11**: 対象医院へ切り替え、飼主検索が実データで表示される。行値はログへ残さない。
- [ ] **H3-7 八王子**: `make stg-uat-import` または `make stg-uat-handoff`。失敗時は成功側を残す。

作業票が repo に無いもの（未実施とは限らない）:

- H3-1 他エンジニア確認と医院への開始範囲連絡
- H3-2 検証済み full backup と復元担当

予約も検証する場合だけ:

- [ ] **H3-10 / AE-STG-UAT-SHIFT**: 未来日の shift。古い絶対日付をコピーしない。

### 実行契約

- `make stg-uat-import` / `make stg-uat-handoff`
- 手動 fallback: `make stg-uat-csv-import-preflight` / `make stg-uat-csv-import` / `make stg-uat-csv-import-verify`
- `make stg-uat-staff-attach-preflight` / `make stg-uat-staff-attach`

具体的な接続値は repo へ記録しない。

---

## Lane 4 — 第1段階の並行運用

- [ ] 期間中は STG DB を reset しない。
- [ ] backend デプロイ前に医院へ日時を知らせる。適用済み migration 編集や seed 差替えを行わない。
- [ ] STG で新規作成した予約、カルテ、会計を本番へ移さない。
- [ ] STG 障害時も旧システムで現行業務を継続する。
- [ ] ログへ PHI が出た場合は検証を止めて修正する。業務監査の正本は `audit_logs` とする。
- [ ] 製品不具合は `bug.md` へ記録し、Linear Issue 化する。
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

1. old_db で **HAC-CSV-1** を出す。両院 Lane 4 のブロッカー。
2. 城東で STG ログインするなら **H3-9 apply** の成否を残し、**H3-11** で飼主検索だけ確認する（行値は残さない）。
3. 八王子 bundle 到着後に Lane 0/1/3。
4. 両院で Lane 4。local UAT を Lane 4 完了としない。

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

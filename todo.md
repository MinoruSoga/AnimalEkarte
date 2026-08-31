# タスク台帳 — Linear が正本

更新日: 2026-08-31

| 項目 | 値 |
|------|-----|
| **実行 SoT** | Linear Team **Baritech** · Project **ノア動物病院電子カルテ** · hub **[BRT-4](https://linear.app/baritechllc/issue/BRT-4)** |
| **直下整理** | **[BRT-105](https://linear.app/baritechllc/issue/BRT-105)** |
| **会社側ログ** | CorpVault `50_Projects/ノア動物病院電子カルテ/` |
| **完了履歴** | Git 履歴と Linear の完了 Issue。完了項目は本ファイルへ残さない |

## 使い方

- 状態・担当・次の一手は Linear を正本とする。
- 確認済みの製品 FAIL は [`bug.md`](bug.md) に記録し、その後 Linear Issue 化する。
- 本ファイルには、repo と強く結び付く **未完了のSTG実データ運用作業だけ**を残す。
- 行値、患者情報、飼主情報、パスワード、接続資格情報は書かない。
- 完了した項目は削除する。経緯が必要な場合は Git 履歴を参照する。

---

## 現在の残作業

| 優先 | ID | 担当 | 状態 | 次の一手 |
|------|----|------|------|------------|
| 1 | **AE-OLD-DB-MR-UNIQ** | old_db | 未完了 | 非NULL `medical_record_id` を1カルテ1billingにし、余剰billingのリンクを外す |
| 2 | **H0-2 / HAC-CSV-1** | old_db / USER | 未完了 | 八王子の医院identity付き21 CSVとmanifestを出力する。旧7 CSVは使わない |
| 3 | **H0-3b / H1-2** | USER | H0-2待ち | 八王子bundleのhandoff check、preflight、ローカルrehearsalを行う |
| 4 | **AE-STG-UAT-LANE3-JOU** | USER | 未着手 | 城東bundleを共有STGへ投入し、第1段階を開始する |
| 5 | **AE-STG-UAT-LANE3-HAC** | USER | H0-2待ち | maintenance windowで八王子bundleを共有STGへ投入する |
| 6 | **Lane 4** | 医院スタッフ / USER | Lane 3待ち | 必須4業務をSTGで連続5営業日再現する |

エージェントは PlanetScale、共有STGへのapply、schema再作成、`DROP SCHEMA`、本番cutoverを実行しません。

---

## STG実データ運用テスト

### 目的

共有STGに旧システムの医院別データを投入し、医院スタッフが現行業務と並行して新システムを検証します。

- 現行業務の正本は旧システム。
- STGへの入力は検証用であり、本番へ移しません。
- 対象は八王子 `clinic_id=1` と城東 `clinic_id=2`。
- 必須業務は検索、受付、カルテ、会計。
- 両院投入後、必須4業務を連続5営業日再現し、現場が切替可と判断したら第2段階へ進みます。上限は8週です。
- 本番cutoverは別工程です。入力は移行日の旧システム出力とし、`PASS` / `TRUSTED_CANDIDATE` 契約を緩めません。

### データ境界

- `002_master`: 医院骨格と参照マスタのみ。accountsや臨床行を含めない。
- `003_demo` / `004_staging`: 退役済み。STG実データ運用には使わない。
- 業務データ: old_dbの医院別21 CSVとmanifestを正本とする。
- `_old_db_handoff/`: ローカル隔離専用。Git、CI、イメージ、通常migrationへ載せない。
- 同一STG DBで医院ごとに10M ID bandを分ける。
- STGの並行登録データを本番へコピー、昇格、差分追加しない。

### 禁止事項

- デモ臨床と実データを混在させない。
- `pscale connect`でremote DBをlocalhostに見せかけない。
- 本番用F6へ`--allow-local-rehearsal`を流用しない。
- 医院間でband、staff、account、患者、飼主、カルテ、会計を越境させない。
- 共有STGへの投入中に同じSTGで業務入力を続けない。
- PHIや資格情報を標準ログ、Git、Issue、チャットへ出さない。

---

## Lane 0 — 八王子入力

- [ ] **H0-2 / HAC-CSV-1**: 医院identity付き21 CSVとmanifestをold_dbから出す。各CSVが512MiB未満であることを確認する。
- [ ] **H0-3b**: `old-db-handoff-check`を通す。manifestなしの旧`hachioji/`や旧7 CSVは使わない。

配置先はrepo外またはgitignoredの次の領域です。

```text
backend/migrations/seeds/_old_db_handoff/hachioji/
```

---

## Lane 1 — 未完了のローカル証明

- [ ] **H1-2**: 八王子bundleでpreflight、apply、verifyをローカルrehearsalする。
- [ ] **H1-5**: ローカルで失敗するbundleを共有STGへ送らない。

必要な場合だけ実施します。

- [ ] **H1-3**: [A4_UI_REHEARSAL.md](docs/ops/deploy/A4_UI_REHEARSAL.md)で隔離画面確認を行う。
- [ ] **H1-4**: [F8_G4_FAILURE_REHEARSAL.md](docs/ops/deploy/F8_G4_FAILURE_REHEARSAL.md)で失敗側を確認する。本番CSVは渡さない。

---

## Lane 3 — 共有STG投入（USERのみ）

[STG_PLANETSCALE_SEED_RUNBOOK.md](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md)の破壊境界に従います。城東を先に投入し、八王子はbundle到着後のmaintenance windowで投入します。

- [ ] **H3-1**: 他エンジニアがSTGを使用していないことを確認し、医院へ開始範囲を伝える。
- [ ] **H3-2**: 検証済みfull backupと復元担当を作業票へ記録する。
- [ ] **H3-3**: TTL付き接続資格情報を用意する。schema再作成が必要な場合はUSER承認下で実施する。
- [ ] **H3-4**: `backend-deploy.yml`を`staging`で実行し、migration成功を確認する。
- [ ] **H3-5**: `make stg-uat-skeleton`で医院骨格を作る。21表のID bandを占有しないことを確認する。
- [ ] **H3-6**: 各医院の10M ID bandが21表すべてで空であることを確認する。
- [ ] **H3-7**: 医院ごとにpreflight → 承認 → apply → verifyを実行する。城東を先に行う。
- [ ] **H3-8**: 一方が失敗した場合は成功側を残し、失敗側だけを修正する。
- [ ] **H3-9**: staff attachのpreflight → applyを実行する。名簿とsecretsはrepo外のmode 0600を使う。
- [ ] **H3-11**: 対象医院へ切り替え、飼主検索が実データで表示されることを確認する。行値はログへ残さない。

予約も検証する場合だけ実施します。

- [ ] **H3-10 / AE-STG-UAT-SHIFT**: 未来日のshiftを用意する。古い絶対日付をコピーしない。

### 実行契約

- `make stg-uat-csv-import-preflight`
- `make stg-uat-csv-import`
- `make stg-uat-csv-import-verify`
- `make stg-uat-staff-attach-preflight`
- `make stg-uat-staff-attach`

remote実行では、対象hostと一致する確認値、SSL設定、各コマンドの明示sentinelが必要です。具体値はrepoへ記録しません。

---

## Lane 4 — 第1段階の並行運用

- [ ] 期間中はSTG DBをresetしない。
- [ ] backendデプロイ前に医院へ日時を知らせる。適用済みmigration編集やseed差替えを行わない。
- [ ] STGで新規作成した予約、カルテ、会計を本番へ移さない。
- [ ] STG障害時も旧システムで現行業務を継続する。
- [ ] ログへPHIが出た場合は検証を止めて修正する。業務監査の正本は`audit_logs`とする。
- [ ] 製品不具合は`bug.md`へ記録し、Linear Issue化する。本ファイルへバグ本文を複製しない。
- [ ] 必須4業務を両院で連続5営業日再現し、現場の切替判断を記録する。

---

## 第2段階へ進む条件

- 両院のbundle投入とverifyが完了している。
- 対象スタッフが自医院へログインできる。
- 必須4業務を連続5営業日再現している。
- ブロッキングな製品バグが残っていない。
- STG入力を本番へ移さないことを医院が理解している。
- 本番F6の`PASS` / `TRUSTED_CANDIDATE`契約が維持されている。
- 移行日、当日の旧データ出力、本番run sheetが決まっている。

---

## 次の一手

1. old_dbで **AE-OLD-DB-MR-UNIQ** を完了する。
2. USERが城東の **Lane 3** を実施する。
3. old_dbが **HAC-CSV-1** を作成し、八王子のLane 0 / 1 / 3を進める。
4. 両院でLane 4を実施し、第2段階の移行日を決める。

---

## 参照

| 文書 | 役割 |
|------|------|
| [docs/ops/deploy/CLINIC_CSV_IMPORT.md](docs/ops/deploy/CLINIC_CSV_IMPORT.md) | F6 21表cutover |
| [docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md) | `_old_db_handoff`のローカル隔離 |
| [docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md](docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md) | seedと21表の境界 |
| [docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md) | STG再作成、直結、破壊境界 |
| [docs/ops/deploy/STG-DEMO-DATA-LIFECYCLE.md](docs/ops/deploy/STG-DEMO-DATA-LIFECYCLE.md) | STGデータのライフサイクル |
| [docs/ops/deploy/STAFF_ACCOUNT_PROVISIONING.md](docs/ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) | staff account運用 |
| [docs/ops/deploy/A4_UI_REHEARSAL.md](docs/ops/deploy/A4_UI_REHEARSAL.md) | 隔離画面確認 |
| [docs/ops/deploy/LOCAL_DB_RESET.md](docs/ops/deploy/LOCAL_DB_RESET.md) | ローカルreset |
| [docs/ops/infra/staging/runbook.md](docs/ops/infra/staging/runbook.md) | STG障害初動 |

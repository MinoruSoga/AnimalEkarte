# Seed / Migration 運用

> **目的**: current migration/seed contract、変更手順、既知blockerを定義する。

## 1. Current contract

- `backend/migrations/` 直下の `*.sql` はDDL。顔ぶれと本数はcurrent checkoutから導出する。
- `cmd/migrate` はDDLを昇順適用後、`seedbundle.BundleOrderForEnv(APP_ENV)`を適用する。
- **全 `APP_ENV` で CSV order は `002_master` のみ。** `003_demo` / `004_staging` は退役済み。accounts.csv は置かない。
- フェーズ3 のログイン upsert（`internal/seedlogin`）は CSV ではない。適用時だけ `seeds/003_login` を記録する。デモパスワードはコード定数（全カタログ共通）。production / empty / unknown はスキップ。開発/STG のログインは bcrypt に加え、カタログ email + 共通パスワードを許可する。
- migration keyはDDL filenameと`seeds/002_master`、およびログイン seed を適用した環境では `seeds/003_login`。fresh DBのexpected historyはcurrent DDL keys + current bundle order + 適用したログイン seedから導出し、固定行数を文書へ複製しない。
- bundle checksumは`manifest.json`と全CSVから導出される。CSV手編集や適用済みbundleの変更はchecksum mismatchの対象になる。
- `002_master/manifest.json` がtable inventoryとload orderのSSOT。現在は12 tableだが、runbookはmanifestから導出する。
- COPY後のsequence advanceも`cmd/migrate`の同じpathに任せる。

空のmigration historyに既存`clinics` tableがある場合はfail-closedする。checksumを手でbaselineしない。異なる内容の統合前`001_init.sql`が記録済みの場合も、reviewed recovery/rebuild planが必要になる。

## 2. Known blocker: legacy seed keys

現行legacy translation codeは、削除済み`003_demo` / `004_staging` directoryのchecksumを計算しようとする。したがって旧stub keyだけが残る典型的DBでtranslation成功を約束できない。

- legacy keyをmanual checksumでbaselineしない。
- 「reset不要」「自動互換」とrelease noteへ書かない。
- shared environmentで再構築やhistory editをしない。
- code taskとして、current valid targetだけへtranslateするか、testで固定したdurable archived checksumを使うreviewed fix/recovery planを先に作る。

このsource defectが直り、scoped testsとapproved rehearsalが通るまでlegacy-only DB migrationは **BLOCKED**。

## 3. Bundle変更

CSVを直接編集しない。使い捨てlocal DBへの実適用から`cmd/seed-export`で`COPY ... TO STDOUT`するcurrent project pathを使う。

```bash
# DB serviceが起動済みのlocal environmentで、projectのDocker ruleに従う
# 実行は変更ownerが行う
docker compose exec backend go run ./cmd/seed-export
```

`cmd/seed-export`のcurrent outputは`002_master`が所有する12 tableだけ。retained `SEED_EXPORT_CSV_SOURCE` adapterはowners/pets/medical_records/exams/exam_results/billings/billing_itemsの7 table入力を扱うが、実行可能なclinical/demo seedを生成しない。retired `003_demo` regeneration pathとして扱わない。

PHIを含み得るinput/outputはGit、chat、artifactへ出さない。manifest/table order、fresh apply、checksum mismatch、sequenceをscoped testで確認する。実DB fresh applyやmigration applyはagentが自動実行しない。

## 4. old_dbとの境界

| input | path | seed result |
|---|---|---|
| formal 21-table `PASS` / `TRUSTED_CANDIDATE` | [CLINIC_CSV_IMPORT.md](./CLINIC_CSV_IMPORT.md) | target DB cutover。seed filesは変更しない |
| `REHEARSAL_ONLY` / `PARTIAL` | [OLD_DB_HANDOFF_LOCAL.md](./OLD_DB_HANDOFF_LOCAL.md)、STG は `make stg-uat-handoff` | gitignored isolation。`cmd/migrate`は読まない。共有 STG は sentinel 付き STG UAT importer |
| executable clinical/demo seed | **未実装** | 生成・適用しない |

21-table cutover bundleとseed manifestは別contractである。handoff CSVを`002_master`へcopyしない。

## 5. Validation and recovery

1. `python3 scripts/verify_seed.py`でcurrent CSV contractを確認する。
2. scoped migrate/seed testsをDocker経由で行う。
3. current DDL keys + `BundleOrderForEnv(APP_ENV)`（+ ログイン seed を適用したなら `seeds/003_login`）についてmigration coverage `missing=0`を確認する。
4. checksum mismatch、legacy-key blocker、rebuild要否をrelease前に記録する。

local checksum mismatchは[LOCAL_DB_RESET.md](./LOCAL_DB_RESET.md)へ進む。shared STGは[STG_PLANETSCALE_SEED_RUNBOOK.md](./STG_PLANETSCALE_SEED_RUNBOOK.md)の承認境界に従う。direct SQLでschema/historyを修正しない。

## Historical procedures

旧`make seed-old-db`、旧SQL seed、`003_demo`/`004_staging`、AWS/RDS reset手順は退役済み。必要な調査はgit historyで行い、live commandとして復元しない。

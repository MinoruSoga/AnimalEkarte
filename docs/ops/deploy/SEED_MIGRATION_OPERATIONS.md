# Seed / Migration 運用

> **目的**: current migration/seed contract、変更手順、既知blockerを定義する。

## 1. Current contract

- `backend/migrations/` 直下の `*.sql` はDDL。顔ぶれと本数はcurrent checkoutから導出する。
- `cmd/migrate` はDDLを昇順適用後、`seedbundle.BundleOrderForEnv(APP_ENV)`を適用する。
- **全 `APP_ENV` で CSV order は `002_master` のみ。** `003_demo` / `004_staging` は退役済み。accounts.csv は置かない。
- フェーズ3 のログイン upsert（`internal/seedlogin`）は CSV ではない。適用時だけ `seeds/003_login` を記録する。デモパスワードはコード定数（全カタログ共通）。任意のシステム管理者は `SEEDLOGIN_OPERATOR_*` 環境変数からの upsert（値は git に置かない）。production / empty / unknown はスキップ。開発/STG のログインは bcrypt に加え、カタログ email + 共通パスワードを許可する。オペレータ email は共通パスワード対象外。
- migration keyはDDL filenameと`seeds/002_master`、およびログイン seed を適用した環境では `seeds/003_login`。fresh DBのexpected historyはcurrent DDL keys + current bundle order + 適用したログイン seedから導出し、固定行数を文書へ複製しない。
- bundle checksumは`manifest.json`と全CSVから導出される。CSV手編集や適用済みbundleの変更はchecksum mismatchの対象になる。
- `002_master/manifest.json` がtable inventoryとload orderのSSOT。現在は12 tableだが、runbookはmanifestから導出する。
- COPY後のsequence advanceも`cmd/migrate`の同じpathに任せる。

### アカウント関連CSVの配置

- 権限グループ・権限ルールは `002_master/accounts/` に置く。
- 旧DBのスタッフも `002_master/accounts/_old_db_handoff/<医院コード>/staffs.csv` に集約し、Git管理外・Docker image対象外・所有者限定の権限を維持する。臨床CSVとmanifestは従来の `_old_db_handoff/<医院コード>/` に残す。
- manifestは従来の論理ファイル名（例: `staffs.csv`）を保持する。readerが物理配置を解決するため、移動だけではmanifest・CSVの内容、bundle checksum、handoffのmanifest SHAは変わらない。
- スタッフの別置きディレクトリは実ディレクトリ・所有者限定で、中身は `staffs.csv` のみ。臨床CSV側との二重配置やシンボリックリンクは拒否する。旧来の自己完結した直下配置も入力として読み込める。
- リポジトリ内のhandoffはGo readerとMakeが中央の医院別ディレクトリを解決する。別の配置を使う場合は `CSV_IMPORT_ACCOUNT_SOURCE_DIR`（Make）または `--account-source-dir`（CLI）で明示する。Composeは対象を `/migration-accounts:ro` にmountする。通常のmaster manifestにはスタッフを追加せず、cutoverとしての検証・取り込みを維持する。
- `cmd/seed-export` と `stage-old-db-handoff.sh` は新しい配置を出力する。職種・ログイン情報の専用CSVは現時点では存在せず、ログイン生成は引き続き `internal/seedlogin` が担当する。

空のmigration historyに既存`clinics` tableがある場合はfail-closedする。checksumを手でbaselineしない。異なる内容の統合前`001_init.sql`が記録済みの場合も、reviewed recovery/rebuild planが必要になる。

### デモログインの本人・所属契約

- デモは10人・10スタッフ・10アカウントで、主所属は八王子、所属医院は全4院。医院ごとに同じ人のアカウントを作らない。執行1人・一般9人を維持し、それぞれの医院の権限グループを割り当てる。
- 職種は表示だけでなく主所属医院の `occupations` と `staffs.occupation_id` に保存する。必要なデモ職種を用意し、再適用時も欠落を補完する。別医院の職種を参照しない。
- 旧デモの医院別複製は既知の合成ID・emailの一致を確認して無効化する。旧DB由来の履歴スタッフは氏名だけで統合・削除せず、ログイン対象と区別する。
- LoginFormのデモ行を選ぶとemailと共通デモパスワードが入力される。ログイン後は所属医院を切り替える。productionではデモを有効にしない。
- コード変更だけでは既存DBは更新されない。隔離DBへのfresh apply・再適用・ログイン・医院切替の実検証は、対象を明示した承認後に行う。既存DBのresetは不要な前提を置かず、checksumエラー時は停止する。

### pull後の開発環境更新

migrationを追加・変更したcommitをpullした開発者は、更新後のアプリを利用する前に `make migrate` を手動実行する。エージェントは自動適用しない。失敗した場合は更新後アプリの利用を止め、checksum・schema/history・ログの失敗段階を確認する。checksumの手修正や `make reset` への自動切替は行わず、対象固有のrecovery手順を確定する。

## 2. Legacy seed keys

`backend/cmd/migrate/main.go` の `legacyTranslationTargets()` は現在 `seeds/002_master` だけを返す。削除済み `003_demo` / `004_staging` の checksum を読もうとして失敗する旧不具合は修正済みで、`legacy_seed_keys_test.go` が対象集合を固定する。

この翻訳は旧 stub key を残したまま、現行 master key を transaction 内で記録する **mark-applied** 処理であり、CSV の再ロードや既存データの完全性確認ではない。legacy key の存在だけで一律 BLOCKED とせず、対象 DB の DDL/checksum と必要 master rows を承認済み read-only 検証で照合する。

- legacy key を手書き checksum で baseline しない。
- 「reset不要」「全旧DBと自動互換」とは断定しない。
- schema/history 不整合・checksum mismatch・master 欠落なら停止し、対象固有の reviewed recovery plan を用意する。
- 実 DB migration/rehearsal の成否は本書の静的照合では未検証。

## 3. Bundle変更

CSVを直接編集しない。使い捨てlocal DBへの実適用から`cmd/seed-export`で`COPY ... TO STDOUT`するcurrent project pathを使う。

```bash
# DB serviceが起動済みのlocal environmentで、projectのDocker ruleに従う
# 実行は変更ownerが行う
docker compose exec backend go run ./cmd/seed-export
```

`cmd/seed-export`のcurrent outputは`002_master`が所有する12 tableだけ。`SEED_EXPORT_CSV_SOURCE` adapterはowners/pets/medical_records/exams/exam_results/billings/billing_itemsの7 table入力を扱うが、実行可能なclinical/demo seedを生成しない。retired `003_demo` regeneration pathとして扱わない。

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
4. checksum mismatch、legacy translation後のmaster完全性、rebuild要否をrelease前に記録する。

local checksum mismatchは[LOCAL_DB_RESET.md](./LOCAL_DB_RESET.md)へ進む。shared STGは[STG_PLANETSCALE_SEED_RUNBOOK.md](./STG_PLANETSCALE_SEED_RUNBOOK.md)の承認境界と§6の再構築手順に従う。direct SQLでschema/historyを修正しない。

## Historical procedures

旧`make seed-old-db`、旧SQL seed、`003_demo`/`004_staging`、AWS/RDS reset手順は退役済み。必要な調査はgit historyで行い、live commandとして復元しない。

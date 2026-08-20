# Seed / Migration 運用メモ

> **目的**: seed/migration変更時の安全な運用手順を定義する。
> **読者**: 開発者。
> **タイミング**: seed/migration変更時。

更新日: 2026-07-31

## 前提

- `backend/migrations/` 直下の `.sql` はDDL専用。顔ぶれ・本数は固定ではなく `ls backend/migrations/*.sql` の実測を正とする（増分の追加と `001_init.sql` への統合で変わる）。統合した旧ファイルの原文・元コミット・SHA-256は001末尾のアーカイブ節に保持する。旧seed stub SQL 002/003/004 は **2026-07 に削除済み**。seed 002〜004 は `backend/migrations/seeds/{002_master,003_demo,004_staging}/` の CSV + `manifest.json` として管理する。
- **cmd/migrate は二段フェーズ構成**（`backend/cmd/migrate/main.go`）:
  1. 直下の `*.sql` を昇順適用し `schema_migrations` にファイル名で記録
  2. 完了後、`internal/seedbundle.BundleOrderForEnv(APP_ENV)` の順で CSV バンドルを pgx `COPY FROM STDIN` ロードし、`internal/seedbundle.BundleMigrationKey(bundleDir)`（`"seeds/002_master"` 等）で `schema_migrations` に記録する
  - **APP_ENV ゲート（SEC-CS-F01 / fail-closed）**: `development` / `local` / `dev` / `test` のみフル順（`002_master → 003_demo → 004_staging`）。`staging` / `production` / `prod` / 空 / 未知は **`002_master` のみ**（demo の active system admin を共有環境へ載せない）。`BundleOrder` 定数はディスク上のフル順の正本（seed-export / lint 用）として残す。
  - 正データの唯一の生成経路は **使い捨てDBへの実適用 → `COPY ... TO STDOUT` ダンプ**（`backend/cmd/seed-export`）。SQL の静的パースによる生成は禁止（ON CONFLICT の最終マージ状態や `random()` 依存データは静的パースでは再現できないため）。
  - `schema_migrations` に記録される seed バンドルの checksum（`bundleChecksum`）は `manifest.json` + 全 CSV ファイルを合成したもの — CSV のみの変更でも通常の migration ファイル編集と同じ checksum mismatch ガードが働く。
  - COPY はシーケンス（BIGSERIAL）を進めないため、`cmd/migrate` は各テーブルロード後に自動で `setval` を実行する（`advanceSerialSequence`）。
- fresh DB 適用後の正しい終了状態は、`schema_migrations` の行数が **直下 DDL 本数 + その環境で許可された seed バンドル数** に一致すること。直下 DDL は `ls backend/migrations/*.sql`、seed は `BundleOrderForEnv(APP_ENV)`（明示したlocal/dev/testは3 bundle、staging/production/空/未知は`seeds/002_master`のみ）を正とする。本節に固定値を書かない。
- `schema_migrations`が空で既存の`clinics`テーブルを検出した場合、`guardEmptyMigrationHistory`はschema完全性を検証できないためfail-closedで停止する。現行DDL/seedのchecksumを適用済みとして記録するbaseline処理は存在しない。USER承認済みのreset/再構築後、通常のDDL・seed適用経路を完走させる。
- 2026-07-27統合前の `001_init.sql` が適用済みのDBでは、統合後001とのchecksum mismatchが必ず発生する。ローカルは`DB_RESET=true`相当の手動再構築が必要で、現行Cloudflare workflowにはSTGを自動resetする経路がない。共有STGは破壊的操作の明示承認後に再構築する。
- **旧形式（stub SQL 時代）互換**（P1-3, PR #186 review で fail-fast から変更）: `schema_migrations` に `002_seed_master.sql` 等の旧キーが残る DB（2026-07 削除より前のバイナリで migrate 済み）を現行バイナリで起動すると、`detectLegacySeedKeys` が旧キーを検出し、旧 stub に対応する `seeds/002_master` / `seeds/003_demo` / `seeds/004_staging` の3キー全てを現行キーへ翻訳して「適用済み」として記録する（見つかった旧キーに対応するものだけでなく、旧形式相当3件を常に全件）。これは履歴が存在するDBだけの互換処理で、CSVは再ロードせず、DDL checksum検証も迂回しない。旧キー移行だけを理由とする DB 再作成は不要。

## 今回の事故で確認したこと

- seed master 差し替えは静的 grep だけでは不十分で、**fresh DB apply** まで通して初めて `(clinic_id, name)` の実衝突を検知できた。この教訓が CSV 移行時の「正データ=DBダンプ・静的パース禁止」の根拠になっている。
- 今回の demo/master 差し替えは **DB reset 前提** で判断した。既存 DB にそのまま上書き適用する前提ではない。
- ローカル復旧で必要だったのは `make reset` 相当の DB 再構築であり、`make db` は `psql` 接続用コマンドであって reset ではない。
- STG で適用済み migration/seed を編集して反映する場合、DB 再作成が必要になる可能性が高い。stub SQL 削除自体は 002〜004 の記録キーを変える変更だが、**旧キーが残った既存環境（STG 等）は `detectLegacySeedKeys` が旧形式相当3件を現行キーへ自動翻訳するため、旧キー移行だけを理由とする DB 再作成は不要**（上記「旧形式互換」参照）。空履歴＋既存schemaはこの互換処理の対象外でfail-closedする。現行 Cloudflare の `backend-deploy.yml` に `db_reset` 入力はなく、AWS ECS/RDS 経路も廃止済み。共有 STG の再作成が必要な場合は、明示承認を得て [STG_PLANETSCALE_SEED_RUNBOOK.md](./STG_PLANETSCALE_SEED_RUNBOOK.md) の破壊的操作境界に従う。

## CSV シードバンドルの再生成（seed データ内容を変更する場合）

seed データの内容（行の追加・削除・値変更）を変えたい場合は、CSV を直接手編集せず、以下の手順で **使い捨てDBからの再エクスポート** を行う。

```bash
# 1. db が起動していること（docker compose ps で確認）
# 2. 使い捨てDB seed_export_tmp を作成 → 現行DDL/seedを適用
#    → bundleTablesの全テーブルを COPY ダンプ → seed_export_tmp を削除、まで一括実行
docker compose exec backend go run ./cmd/seed-export
```

- `cmd/seed-export` は `DB_HOST` が `db`/`localhost`/`127.0.0.1` 以外なら拒否し、DB名は常に固定の `seed_export_tmp`（環境変数 `DB_NAME` は無視）— 本番/STG DBを誤操作する経路が構造的に存在しない。
- `SEED_EXPORT_CSV_SOURCE` を指定した旧overlay adapterは `owners` / `pets` / `medical_records` / `exams` / `exam_results` / `billings` / `billing_items` の7表だけを取り込む。正式21表consumerの代替ではなく、残り14表を忠実にseedへ変換できない。
- 003 の高密度予約デモ生成（`random()`使用）は、この使い捨てDB適用の中で一度だけ実行され、その結果行がそのまま CSV としてダンプされる。**同じデータを再現したい場合に「2回実行して同じハッシュになるか」を確認する検証方法は誤り** — 実行のたびに新しい使い捨てDBを作るため、`random()` は毎回新しい値を引く。凍結の担保は「dump 側が読み取り専用の `COPY TO STDOUT` しか実行しない」ことと「移行後は `cmd/migrate` 側に `random()`/DO ブロックが一切残っていない」ことの両方で保証される。
- 生成された CSV / `manifest.json` にPHIが含まれる場合、データ管理・Git配布の明示承認までは `git add` / commit / pushしない。

### old_db 21表CSVとの境界

| 入力・目的 | 正規経路 | 結果 |
|---|---|---|
| `PASS` / `TRUSTED_CANDIDATE` の正式21表bundle | [CLINIC_CSV_IMPORT.md](./CLINIC_CSV_IMPORT.md) の `make csv-import-*` | 対象DBへ投入する。seedファイルは生成しない |
| `REHEARSAL_ONLY` / `PARTIAL` の暫定21表bundle | `backend/migrations/seeds/_old_db_handoff/<clinic>/<run>/`（`make old-db-handoff-stage`） | このworktreeだけのローカル保管。`cmd/migrate` / `make seed` からは読まれない。手順は [OLD_DB_HANDOFF_LOCAL.md](./OLD_DB_HANDOFF_LOCAL.md) |
| 実行可能な `003_demo` seed | **未実装 / BLOCKED** | 現行 `cmd/seed-export` は21表handoffを入力できない。専用adapter実装・検証後に使い捨てDBから全bundleを再生成する |

`_old_db_handoff/` はPHIを含み得るため、コピー前にcheckout固有の
`.git/info/exclude`へ `backend/migrations/seeds/_old_db_handoff/` を登録し、
`git check-ignore -q --no-index backend/migrations/seeds/_old_db_handoff/`
がPASSすることを必須とする。tracked `.gitignore` に規則がないcheckoutでは、
この確認を省略しない。

21表CSVを `003_demo` へ直接コピーしてはいけない。cutover manifestとseed bundle
manifestは別契約であり、placeholder解決、target-only seed ID、全テーブル依存、
checksumを使い捨てDB上で確定する必要がある。現行コードに21表専用の
seed変換経路はないため、暫定bundleは隔離保管までとし、実行可能seedとは呼ばない。

## 変更時の最低確認

1. `python3 scripts/verify_seed.py`（CSV ベース。SQL の静的パースは行わない）
2. fresh DB への migration apply 検証（`docker compose exec backend go test ./cmd/migrate/...` はスコープ限定で自動実行可。実DBへの fresh apply はユーザー手動）
3. seed/migration を編集した場合は、checksum mismatch と `db_reset=true` 要否を事前に整理する

## ローカル復旧

- checksum mismatch が出たローカル環境は [LOCAL_DB_RESET.md](./LOCAL_DB_RESET.md) の手順で DB volume を再構築する。
- 開発 DB への直接 SQL 実行や共有環境での手動 reset は行わない。

---

## 旧DB移行：正式経路は CSV import（F6）

正式な医院カットオーバーは [CLINIC_CSV_IMPORT.md](./CLINIC_CSV_IMPORT.md) に従い、`old_db` の21表 CSV + manifest を read-only mountして `make csv-import-preflight` → 承認済み `make csv-import` → `make csv-import-verify` の順で対象DBへ投入する。このF6経路はseedファイルを生成・更新しない。

Issue #250（stage-import 拡張・rehearsal・cutover）の consumer 側受け入れ対応表・21表 mapping・dry-run/idempotency/非PHI エラー ID は [CLINIC_CSV_IMPORT.md](./CLINIC_CSV_IMPORT.md) の「Issue #250 受け入れ条件との対応」を正本とする。production cutover apply は #253/#254/#255 gate 後の USER 操作であり、本ドキュメントの seed 経路では実行しない。

## 旧DB移行データのローカル投入 (old-db seed) — **retired history / 実行禁止**

更新日: 2026-06-24

> **⚠️ 廃止済みの履歴。** `docker-compose.seed-old-db.yml` と現役の
> `make seed-old-db` / `make verify-old-db-seed` は存在しない。以下のコマンドを
> 実行手順として使用せず、正式21表はF6、seed再生成は上記の使い捨てDB経路を使う。

`old_db/sensitive-local/migration-output/` にある TSV ファイル（旧DB→新スキーマのマッピング済みデータ）をローカル開発DBへ投入する手順。

**スコープ**: ローカル開発DB専用。本番・STG への投入は禁止。

### 前提

| 項目 | 内容 |
|------|------|
| 入力パス | `/Users/minoru/Dev/Case/AnimalHospital/old_db/sensitive-local/migration-output/` |
| マニフェスト | `old_db/docs/generated/new-schema-import-manifest.json` |
| 安全境界 | `DB_HOST` が `db` / `localhost` / `127.0.0.1` 以外では自動的に拒否 |

### 旧手順（履歴のみ・実行禁止）

```bash
# 1. DBをリセットして標準 seed を適用
make reset

# 2. 旧DB移行データを投入
make seed-old-db

# 3. 件数・FK整合性を検証
make verify-old-db-seed
```

カスタムパスを使う場合:

```bash
make seed-old-db OLD_DB_MIGRATION_OUTPUT_DIR=/other/path OLD_DB_DOCS_DIR=/other/docs
```

### 仕組み

1. `docker-compose.seed-old-db.yml` が `old_db/sensitive-local/migration-output/` を `/old-db-data:ro` としてマウント。
2. `backend/cmd/seed-old-db/main.go` がマニフェストを読み、各 TSV を `bufio.Scanner` でストリーム読み込み。
3. `pgx.Batch` による 1,000 行単位のパラメータ化 `INSERT` でバルク投入（型キャスト・重複スキップ付き）。
4. FK 依存順（clinics → owners/exam_types → pets → …）に並び替えて処理。
5. 旧DBの `pet_number` / `record_no` は seed 実行中の Go 側キャッシュで新DBの `pets.id` / `medical_records.id` に解決する。
6. 以下のエントリはスキップ（理由はログに出力）:

**loadされるエントリ (13件)**:
   - `animal_species`, `clinics` (×2 source), `owners`, `pets`
   - `exam_types` (×2 source), `exam_type_fields`
   - `merchandise_items`, `procedures`
   - `medical_records::TBL_KARTE_DATA`, `billings::TBL_TRI_DATA`, `exams::TBL_KNS_HIST`

   `owners` は旧 `SiiDM_Kbn`（DM発送区分）を `owners.dm_preference` (boolean) に変換して保持する:
   `01`(DM発送可)→`true` / `02`(DM発送不可)→`false` / その他・空→`NULL`（未設定, unknown-safe）。
   この変換は `transform.go` の `dmPreferenceExpr` で実装され、go unit test と `verify-old-db-seed.sh` の
   `owners.dm_preference populated` チェックで保証される（以前は whitelist 漏れで silent drop されていた）。

**スキップされるエントリ (14件)** — 3 区分に分類:

(A) **子明細クロスウォーク BLOCKED**（親キーが export にないため load 不可。required input は old_db `final-handoff-status.json` の `子明細クロスウォーク` blocked エントリ参照）:
   - `medical_records::TBL_TRI_DATA` — `record_no` (TReat_Sno) のみ。`record_no` 単独は非一意（同値が多重）。必要: 複合キー `(pet_id=TPK_PET_No, record_no=TReat_Sno)` を export が emit + seeder cache を複合キー化
   - `billing_items` (×2) — `billing_id` が TSV にない。必要: `(pet_id=TPK_PET_No, record_no=TReat_Sno)` 複合キー emit + 2-hop `(pet_id,record_no)→billing.id` cache
   - `treatments` (×2) — `medical_record_id` が TSV にない。必要: `(pet_id=TPK_PET_No, record_no=TReat_Sno)` 複合キー emit + 複合キー medical_records cache
   - `exam_results` — `exam_id` が TSV にない。必要: `(pet_id=HkPK_PETNo, sno=HkSK_SNo)` 複合キー emit + 2-hop `(pet_id,sno)→exam.id` cache
   - `clinical_plans`, `inquiries` — `medical_record_id` が TSV にない。必要: `(pet_id=KPK_PetNo, record_no=Ksk_KarteNo)` 複合キー emit + 複合キー medical_records cache
   - `vital_records` — `medical_record_id`/`daily_record_id` が TSV にない。必要: 上記 `TBL_KARTE_DATA` 複合キー crosswalk
   - `medical_record_addenda` — `medical_record_id` + `author_user_id` が必要。複合キー crosswalk に加え `author_user_id` は legacy source 無し（staff mapping 欠落）

(B) **意図的除外**（FK 値のない巨大重複 source。load しても dev で無価値）:
   - `medical_records::MST_PET_INFO` — 次回来院推奨日の更新用データで、単独 INSERT に必要な `record_no` / `date` がない（update-only source）
   - `exam_types::TBL_KNS_HIST` — `name` 列のみ 1.3M 行、FK 値なし（重複行を量産するため除外）
   - `staffs::TBL_KARTE_DATA` — `name` 列のみ 425K 行、FK 値なし

(C) **source 欠落**（必須列が export 自体に存在しない）:
   - `shared_files` — uploaded_by/file_type/file_key/file_size が TSV にない

### 最新検証結果

2026-06-25 の fresh DB 投入結果（`make reset` → `make seed-old-db` → `make verify-old-db-seed`）。
子明細クロスウォーク解決後の値で、`vital_records` / `clinical_plans` / `inquiries` /
`treatments` / `billing_items` / `exam_results` が複合キーで load されるようになったため
loaded/skipped と件数が以前（loaded=13/skipped=14）から更新されている。

```text
make reset
  exit=0   # one-shot codegen を wait 対象から除外し cosmetic exit-1 を解消
           # この wait-set 契約は make check-reset-contract / CI で自動検証され、
           # 裸の `up --wait` への退行は merge 前に reject される（手動確認不要）

make seed-old-db
  loaded=19
  skipped=8
  errors=0
  totalRowsInserted=7,880,430

make verify-old-db-seed
  PASS=34
  FAIL=0
  WARN=0
```

`make verify-old-db-seed` は以下のテーブル件数と主要 FK 孤児ゼロを確認する。

| テーブル | 検証時件数 |
|---|---:|
| animal_species | 9 |
| clinics | 7 |
| owners | 10,427 |
| pets | 15,601 |
| exam_types | 31 |
| exam_type_fields | 368 |
| merchandise_items | 5,073 |
| procedures | 5,088 |
| medical_records | 425,681 |
| billings | 392,433 |
| exams | 1,343,725 |
| vital_records | 424,938 |
| clinical_plans | 425,544 |
| inquiries | 425,544 |
| treatments | 1,542,116 |
| billing_items | 1,542,116 |
| exam_results | 1,322,321 |

### 注意事項

- `*.import.tsv` および `sensitive-local/` は `.gitignore` 済み。AnimalEkarte リポジトリには絶対にコミットしない。
- `make seed-old-db` は DB の drop/recreate を行わない。スキーマ変更は `make reset` で行う。
- 全エントリが loaded/skipped になれば成功（error が 1 件でもあれば非ゼロ終了）。
- `make reset` の `up --wait` wait-set 契約は `scripts/check-reset-wait-services.sh` が静的に検証する。
  `make check-reset-contract`（および `make ci` 先頭ステップ）で
  自動実行されるため、one-shot codegen 混入や裸 `up --wait` への退行は手動確認なしで検出される。
  契約チェック自体の回帰テストは `make check-reset-contract-test` で実行できる。
- `scripts/*.sh` の lint は `make shellcheck`（`make ci` の ShellCheck ステップ）で
  shellcheck により**自動**実行される。シェルスクリプトのレビューは手動目視ではなく、整形・行継続トリックで
  欺けない AST 検査でゲートされる（`make reset` 等の運用スクリプトの退行は merge 前に reject される）。
  ゲート自体の回帰テストは `make shellcheck-test` で実行できる。shellcheck がローカルに無い場合は
  ピン留め Docker イメージ経由で再現可能に実行される。

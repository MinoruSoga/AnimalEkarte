# Seed / Migration 運用メモ

> **目的**: seed/migration変更時の安全な運用手順を定義する。
> **読者**: 開発者。
> **タイミング**: seed/migration変更時。

更新日: 2026-07-24

## 前提

- `backend/migrations/` 直下の `.sql` はDDL専用で、`001_init.sql`（統合初期スキーマ、変更しない）と追記専用incremental `002_lstep_snapshot_import_clinic_fk.sql` / `003_medical_records_appointment_id_index.sql` / `004_payment_splits_billing_id_index.sql` が存在する。旧seed stub SQL 002/003/004 は **2026-07 に削除済み**。seed 002〜004 は `backend/migrations/seeds/{002_master,003_demo,004_staging}/` の CSV + `manifest.json` として管理する。
- **cmd/migrate は二段フェーズ構成**（`backend/cmd/migrate/main.go`）:
  1. 直下の `*.sql` を昇順適用し `schema_migrations` にファイル名で記録
  2. 完了後、`internal/seedbundle.BundleOrder` の固定順（`002_master → 003_demo → 004_staging`）で CSV バンドルを pgx `COPY FROM STDIN` ロードし、`internal/seedbundle.BundleMigrationKey(bundleDir)`（`"seeds/002_master"` 等）で `schema_migrations` に記録する
  - 正データの唯一の生成経路は **使い捨てDBへの実適用 → `COPY ... TO STDOUT` ダンプ**（`backend/cmd/seed-export`）。SQL の静的パースによる生成は禁止（ON CONFLICT の最終マージ状態や `random()` 依存データは静的パースでは再現できないため）。
  - `schema_migrations` に記録される seed バンドルの checksum（`bundleChecksum`）は `manifest.json` + 全 CSV ファイルを合成したもの — CSV のみの変更でも通常の migration ファイル編集と同じ checksum mismatch ガードが働く。
  - COPY はシーケンス（BIGSERIAL）を進めないため、`cmd/migrate` は各テーブルロード後に自動で `setval` を実行する（`advanceSerialSequence`）。
- fresh DB 適用後の正しい終了状態は `schema_migrations` に **7行**（DDL 4 + seed 3）: `001_init.sql` + `002_lstep_snapshot_import_clinic_fk.sql` + `003_medical_records_appointment_id_index.sql` + `004_payment_splits_billing_id_index.sql` + `seeds/002_master` + `seeds/003_demo` + `seeds/004_staging`。
- `schema_migrations`が空で既存テーブルを検出するlegacy baselineでは、`001_init.sql`とseed 3バンドルを適用済み記録する。`002`以降のappend-only DDLはbaseline対象外で、直後のmigration phaseで実行する。
- 既に適用済みの `001_init.sql` / seed バンドル（CSV・manifest.json）を編集すると、既存 DB の `schema_migrations` に記録された checksum と不一致になる。
- **旧形式（stub SQL 時代）互換**（P1-3, PR #186 review で fail-fast から変更）: `schema_migrations` に `002_seed_master.sql` 等の旧キーが残る DB（2026-07 削除より前のバイナリで migrate 済み）を現行バイナリで起動すると、`detectLegacySeedKeys` が旧キーを検出し、旧 stub に対応する `seeds/002_master` / `seeds/003_demo` / `seeds/004_staging` の3キー全てを「適用済み」として baseline する（見つかった旧キーに対応するものだけでなく、旧形式相当3件を常に全件）。旧キー移行だけを理由とする DB 再作成は不要。

## 今回の事故で確認したこと

- seed master 差し替えは静的 grep だけでは不十分で、**fresh DB apply** まで通して初めて `(clinic_id, name)` の実衝突を検知できた。この教訓が CSV 移行時の「正データ=DBダンプ・静的パース禁止」の根拠になっている。
- 今回の demo/master 差し替えは **DB reset 前提** で判断した。既存 DB にそのまま上書き適用する前提ではない。
- ローカル復旧で必要だったのは `make reset` 相当の DB 再構築であり、`make db` は `psql` 接続用コマンドであって reset ではない。
- STG で適用済み migration/seed を編集して反映する場合、DB 再作成が必要になる可能性が高い。stub SQL 削除自体は 002〜004 の記録キーを変える変更だが、**旧キーが残った既存環境（STG 等）は `detectLegacySeedKeys` が旧形式相当3件を自動 baseline するため、旧キー移行だけを理由とする DB 再作成は不要**（上記「旧形式互換」参照）。現行 Cloudflare の `backend-deploy.yml` に `db_reset` 入力はなく、AWS ECS/RDS 経路も廃止済み。共有 STG の再作成が必要な場合は、明示承認を得て [STG_PLANETSCALE_SEED_RUNBOOK.md](./STG_PLANETSCALE_SEED_RUNBOOK.md) の破壊的操作境界に従う。

## CSV シードバンドルの再生成（seed データ内容を変更する場合）

seed データの内容（行の追加・削除・値変更）を変えたい場合は、CSV を直接手編集せず、以下の手順で **使い捨てDBからの再エクスポート** を行う。

```bash
# 1. db が起動していること（docker compose ps で確認）
# 2. 使い捨てDB seed_export_tmp を作成 → 未改変の 002-004 フル INSERT 版を適用
#    → 90テーブルを COPY ダンプ → seed_export_tmp を削除、まで一括実行
docker compose exec backend go run ./cmd/seed-export
```

- `cmd/seed-export` は `DB_HOST` が `db`/`localhost`/`127.0.0.1` 以外なら拒否し、DB名は常に固定の `seed_export_tmp`（環境変数 `DB_NAME` は無視）— 本番/STG DBを誤操作する経路が構造的に存在しない。
- 003 の高密度予約デモ生成（`random()`使用）は、この使い捨てDB適用の中で一度だけ実行され、その結果行がそのまま CSV としてダンプされる。**同じデータを再現したい場合に「2回実行して同じハッシュになるか」を確認する検証方法は誤り** — 実行のたびに新しい使い捨てDBを作るため、`random()` は毎回新しい値を引く。凍結の担保は「dump 側が読み取り専用の `COPY TO STDOUT` しか実行しない」ことと「移行後は `cmd/migrate` 側に `random()`/DO ブロックが一切残っていない」ことの両方で保証される。
- 生成された CSV / `manifest.json` は `git add` して通常のコミットフローでレビューする。

## 変更時の最低確認

1. `python3 scripts/verify_seed.py`（CSV ベース。SQL の静的パースは行わない）
2. fresh DB への migration apply 検証（`docker compose exec backend go test ./cmd/migrate/...` はスコープ限定で自動実行可。実DBへの fresh apply はユーザー手動）
3. seed/migration を編集した場合は、checksum mismatch と `db_reset=true` 要否を事前に整理する

## ローカル復旧

- checksum mismatch が出たローカル環境は [LOCAL_DB_RESET.md](./LOCAL_DB_RESET.md) の手順で DB volume を再構築する。
- 開発 DB への直接 SQL 実行や共有環境での手動 reset は行わない。

---

## 旧DB移行：正式経路は CSV import（F6）

正式な医院カットオーバーは [CLINIC_CSV_IMPORT.md](./CLINIC_CSV_IMPORT.md) に従い、`old_db` の21表 CSV + manifest を read-only mount して `make csv-import-preflight` → 承認済み `make csv-import` → `make csv-import-verify` の順で実行する。

`stage-import` は old_db Postgres へ直接接続する旧ローカル互換経路であり、F6/F7の本番カットオーバーには使用しない。

## 旧ローカル互換経路：stage-import（animalekarte_stage → 本テーブル）

更新日: 2026-06-25

> **⚠️ 直下の TSV ベース直接 seeder (`make seed-old-db`) は deprecated（comparison-only）。**
> F6の本番移行には使用禁止。過去のローカル比較だけで **`make stage-import`** を使う。下記
> 「旧DB移行データのローカル投入 (old-db seed)」節は比較用に残す。

### なぜ stage-import か

直接 seeder は変換ロジックを2リポジトリ（old_db の Node generator + AnimalEkarte の Go seeder）
に分割していたため、マッピング正しさを単一クエリで検証できなかった。空 clinic / branch code
leakage / owner `000001` 衝突 / `record_no`単独 mis-link / 子明細 orphan / lineage 欠落 といった
既知の失敗はすべて下流で個別に潰すしかなかった。

`old_db` の3層パイプライン（`legacy_raw` → `legacy_canonical` → `animalekarte_stage`、
old_db `make migration-pipeline`）は全変換を純 SQL に集約し、`make migration-stage-check` が
各失敗モードを決定的に検証する。stage-import はその **検証済み stage** を唯一の投入元として
本テーブルへ取り込むため、投入前に正しさが保証される。

### 手順（推奨）

```bash
# 0. old_db 側（別 repo）でパイプラインを構築・検証
#    cd old_db && make local-postgres-up && make migration-pipeline && make migration-stage-check

# 1. stage の TCP パスワードを渡す（old_db の POSTGRES_PASSWORD と同一）
export OLD_DB_POSTGRES_PASSWORD=...

# 2. dry-run（件数表示・本テーブルへの書き込みは0）
make stage-import-dry-run

# 3. apply（破壊的：old_db 由来行を削除し stage から再投入。demo/master/config は保持）
make stage-import

# 4. 投入後検証（空clinic / branch leakage / owner collision / orphan / record_no / demo混入）
make verify-stage-import
```

### 仕組み（要点）

| 項目 | 内容 |
|------|------|
| 投入元 | `animalekarte_stage.*`（old_db の `ani_legacy` Postgres）のみ。`legacy_raw`/`legacy_canonical` から本テーブルへは書かない |
| 対象行 | `mapping_status IN ('confirmed','inferred')` のみ。`needs_review`/`archive_only`/`blocked` は投入せず件数と理由を report |
| 親子整合 | hard-FK 子（billing_items→billings、exam_results→exams）は親が非対象なら子もスキップ（dangling FK 防止） |
| clinic | 本テーブルの 八王子病院 を名前解決した id に固定。legacy branch code は使わない |
| NOT NULL 補完 | `animal_species_id`/`exam_type_id` は本テーブルの fallback id、`date`/`scheduled_date` は 1900-01-01 へ COALESCE、`name_kana` は katakana CHECK 回避のため非投入（DEFAULT '') |
| 安全境界 | 非ローカル TARGET DB_HOST を拒否。stage 接続は read-only。apply は `--apply` かつ `--confirm-local-destroy` 必須 |
| トランザクション | 削除→投入を単一トランザクションで実行。失敗時は全ロールバック（中途半端な本テーブル状態を残さない） |
| 冪等性 | 再実行で old_db 由来行を削除してから再投入するため重複しない |
| 実装 | `backend/cmd/stage-import/`、`docker-compose.stage-import.yml`、`scripts/verify-stage-import.sh` |

### テスト

- unit: `make test`（SQL生成・status filter・親子cascade・guard・delete scope・FK順）
- rollback / read-only 統合: `make stage-import-rollback-test`（注入失敗後に件数不変・stage は read-only）

---

## 旧DB移行データのローカル投入 (old-db seed) — **deprecated / comparison-only**

更新日: 2026-06-24

> **⚠️ deprecated。** 上の `make stage-import` を使うこと。本節は比較用に保持。

`old_db/sensitive-local/migration-output/` にある TSV ファイル（旧DB→新スキーマのマッピング済みデータ）をローカル開発DBへ投入する手順。

**スコープ**: ローカル開発DB専用。本番・STG への投入は禁止。

### 前提

| 項目 | 内容 |
|------|------|
| 入力パス | `/Users/minoru/Dev/Case/AnimalHospital/old_db/sensitive-local/migration-output/` |
| マニフェスト | `old_db/docs/generated/new-schema-import-manifest.json` |
| 安全境界 | `DB_HOST` が `db` / `localhost` / `127.0.0.1` 以外では自動的に拒否 |

### 一発手順（推奨）

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

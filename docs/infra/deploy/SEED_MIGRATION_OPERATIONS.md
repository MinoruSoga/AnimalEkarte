# Seed / Migration 運用メモ

更新日: 2026-06-04

## 前提

- `001_init.sql` / `002_seed_master.sql` / `003_seed_demo.sql` / `004_seed_staging.sql` は、fresh DB への初期構築前提で順次適用される。
- 既に適用済みの migration / seed を編集すると、既存 DB の `schema_migrations` に記録された checksum と不一致になる。

## 今回の事故で確認したこと

- seed master 差し替えは静的 grep だけでは不十分で、**fresh DB apply** まで通して初めて `(clinic_id, name)` の実衝突を検知できた。
- 今回の demo/master 差し替えは **DB reset 前提** で判断した。既存 DB にそのまま上書き適用する前提ではない。
- ローカル復旧で必要だったのは `make reset` 相当の DB 再構築であり、`make db` は `psql` 接続用コマンドであって reset ではない。
- STG で適用済み migration/seed を編集して反映する場合、`backend-deploy.yml` の `db_reset=true` が必要になる可能性が高い。

## 変更時の最低確認

1. `python3 scripts/verify_seed.py`
2. fresh DB への migration apply 検証
3. seed/migration を編集した場合は、checksum mismatch と `db_reset=true` 要否を事前に整理する

## ローカル復旧

- checksum mismatch が出たローカル環境は [LOCAL_DB_RESET.md](./LOCAL_DB_RESET.md) の手順で DB volume を再構築する。
- 開発 DB への直接 SQL 実行や共有環境での手動 reset は行わない。

---

## 旧DB移行データのローカル投入 (old-db seed)

更新日: 2026-06-24

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

2026-06-24 の fresh DB 投入結果。`owners.dm_preference` の保持チェックを追加したため
`verify-old-db-seed` の PASS 件数は 21 → 22 に増える（投入後に下表を更新する）。

```text
make seed-old-db
  loaded=13
  skipped=14
  errors=0
  totalRowsInserted=1,382,225

make verify-old-db-seed
  PASS=22   # owners.dm_preference populated チェックを追加
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
| medical_records | 1,433 |
| billings | 1,054 |
| exams | 1,343,725 |

### 注意事項

- `*.import.tsv` および `sensitive-local/` は `.gitignore` 済み。AnimalEkarte リポジトリには絶対にコミットしない。
- `make seed-old-db` は DB の drop/recreate を行わない。スキーマ変更は `make reset` で行う。
- 全エントリが loaded/skipped になれば成功（error が 1 件でもあれば非ゼロ終了）。

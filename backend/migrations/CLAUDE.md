# Migrations

## 命名規則

```
{連番}_{description}.sql
例: 005_add_trimming_course.sql
```

連番は既存の最大番号 + 1。説明は snake_case。

## 必須チェック

- **clinic_id スコープ**: 新テーブルにクリニック間分離が必要な場合は `clinic_id NOT NULL` を付ける
- **ソフトデリート対象**: 業務データは `deleted_at TIMESTAMPTZ` を追加する
- **CASCADE DELETE（原則禁止・例外あり）**: 下記の考え方に従うこと

  **禁止（絶対）**: `owners` / `pets` / `medical_records` 等の PHI・業務データが親となるCASCADEで、削除により診療履歴・会計・バイタル等が連鎖消去されうる設計は禁止。  
  service 層で依存チェックして 409 を返す（P10 参照）を優先し、DBレベルのCASCADE に頼らない。

  **許容される例外**: "純粋な従属データ" で、親が消える時に必ず消えてよい行のみCASCADEを許容する。
  - 中間テーブル / join table（例: `permission_group_staffs`、`shift_template_reservation_types`）
  - 親レコードの構成要素として不可分な子行（例: `vital_records` → `medical_records`、`exam_items` → `exams`、`billing_items` → `billings`）
  - マスタ参照のカスケードで業務履歴を失わないもの（例: `exam_types` / `diagnosis_types` 等の lookup FK）

  **現行スキーマの注記**: `001_init.sql` には `medical_records(id) ON DELETE CASCADE` 等の例外的CASCADE定義が複数存在する（設計当初の判断）。これらは上記「許容される例外」に該当するものとして現状維持。将来の整理タスクで見直す。スキーマ変更はこのタスクのスコープ外。
- **インデックス**: `clinic_id` を含む複合インデックスを追加する

## 実行禁止コマンド（自動実行禁止）

```bash
make db            # DB リセット（高い副作用）
docker compose exec db psql ...  # 直接 SQL 実行
```

マイグレーション適用はユーザーが手動で実施する。

## migration 統合後のローカル復旧

`001_init.sql` の checksum mismatch が出たローカル環境は、`docs/ops/deploy/LOCAL_DB_RESET.md` の手順で DB volume を再構築する。

## seed データは CSV が正、SQL は DDL のみ（2026-07 stub 削除 + 001 統合 + incremental upgrade）

`backend/migrations/` 直下の `.sql` は DDL 専用。`001_init.sql` は **fresh install 用の統合スキーマ**（旧インデックス増分 + 旧 005–012 + `appointments.checked_in_at` を含む）のまま維持する。
すでに薄い/統合前の `001_init.sql` だけが `schema_migrations` に記録されている STG/PROD 向けに、畳み込み DDL を **冪等な incremental**（`003` 以降）として再出荷する。fresh DB では incremental は `IF NOT EXISTS` / `duplicate_object` ガードで no-op 相当になる。

現行 DDL ファイル（2026-07-17 時点）:

1. `001_init.sql` — 統合スキーマ（旧 002–004 インデックス + §7 の旧 005–012 相当 + `checked_in_at`）
2. `002_checkup_field_clinic_composite_fk.sql` — 健診 `checkup_type_fields`↔`checkup_types` clinic 複合 FK（旧 012 の field_results 複合 FK とは別）
3. `003_add_lab_import_tables.sql` — 由来: 旧 005（`lab_import_*`）
4. `004_add_exam_results_exam_id_index.sql` — 由来: 旧 006
5. `005_add_exams_dup_check_index.sql` — 由来: 旧 007
6. `006_add_exams_job_id.sql` — 由来: 旧 008（`exams.job_id`）
7. `007_add_medicine_dose_params.sql` — 由来: 旧 009（`medicine_dose_params` 等）
8. `008_add_checkup_packages.sql` — 由来: 旧 010（`checkup_type_fields` / `checkup_field_results`）
9. `009_add_closing_am_start.sql` — 由来: 旧 011
10. `010_add_clinical_result_composite_fk.sql` — 由来: 旧 012（`checkup_field_results`↔`checkup_type_fields` 複合 FK。002 と重複しない）
11. `011_add_appointment_checked_in_at.sql` — 由来: 旧 `005_add_appointment_checked_in_at.sql`

旧 002/003/004 の seed stub SQL ファイル（`SELECT 1;` の no-op）は削除済みで、seed の実体は `backend/migrations/seeds/{002_master,003_demo,004_staging}/*.csv` + `manifest.json` というディレクトリだけになった。
旧インデックス増分ファイル（`002_add_checkup_vaccination_indexes.sql` / `003_add_pets_batch_living_count_index.sql` / `004_add_billings_hospitalization_id_unique_index.sql`）も `001_init.sql` へ統合済みで、独立ファイルとしては存在しない。

- **cmd/migrate は二段フェーズ構成**: ①直下の `*.sql`（DDL: `001_init.sql` + incremental 002 以降）を昇順適用 → ②`internal/seedbundle.BundleOrder`（`002_master → 003_demo → 004_staging`固定順）で CSV バンドルを pgx `COPY FROM STDIN` ロード（`backend/cmd/migrate/csvbundle.go`）。DDL 失敗時は seed フェーズへ進まない
- **schema_migrations の記録キー**: DDL は従来通り各ファイル名。seed バンドルは `internal/seedbundle.BundleMigrationKey(bundleDir)` が返す `"seeds/<bundle>"`（例: `seeds/002_master`）— stub SQL ファイル名には二度と紐付かない。fresh DB 適用後の正しい終了状態は、**直下 DDL ファイル数 + seed 3 バンドル**（2026-07-17 時点は DDL 11 + seed 3 = 14 行）
- **upgrade path 注意**: 統合スキーマを畳み込んだあと `001` だけ適用済みの DB では、migrate が `001` をスキップするため §7 相当オブジェクトが欠ける。`003`–`011` がその欠落を埋める。STG/PROD への適用は USER 手動（`make db` / 自動 migrate 禁止）
- seed バンドルの checksum（`bundleChecksum`）は manifest.json + 全 CSV ファイルの内容を合成したもの。CSV だけの編集でも、既に適用済みの DB では通常の migration ファイル編集と同じ checksum mismatch ガードが働く
- COPY はシーケンスを進めないため、テーブルロード後に自動 setval される（`advanceSerialSequence`）
- **旧形式（stub SQL 時代）からの移行は非対応**: `schema_migrations` に `002_seed_master.sql` 等の旧キーが残っている DB を現行バイナリで起動すると `detectLegacySeedKeys` が fail-fast する。対処は `db_reset` またはボリューム再構築のみ（in-place 移行は実装しない）
- CSV の正データ生成経路は使い捨てDBダンプのみ: `docker compose exec backend go run ./cmd/seed-export`。SQL の静的パース/手編集による CSV 生成は禁止（ON CONFLICT の最終マージ状態・`random()` 依存データは静的パースでは再現できない）
- テーブル→バンドル割当は「最初に触れたファイル」基準（earliest-file-wins、`cmd/seed-export/tables.go` の `bundleTables`）で固定済み。新しいシードテーブルの追加は `cmd/seed-export` の再設計が必要

## seed / migration 差し替え時の注意

- 適用済み migration（`001_init.sql` および incremental DDL）/ seed バンドル（CSV・manifest.json）を編集すると既存 DB は checksum mismatch になる
- seed データの内容変更は CSV の手編集ではなく `cmd/seed-export` の再実行で行う（上記参照）
- DB 非依存の最低検証は `python3 scripts/verify_seed.py`（CSV ベース）
- 運用メモ: `docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md`

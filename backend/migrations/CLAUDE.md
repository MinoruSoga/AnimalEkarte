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

## seed データは CSV が正、SQL は DDL のみ（2026-07 stub 削除 + 001 完全統合）

`backend/migrations/` 直下の `.sql` は DDL 専用。2026-07-27 に当時の incremental 002–009 を原文のまま `001_init.sql` 末尾へ統合し、同日夕に追加分の incremental 002–004（`002_pets_owners_clinic_composite_unique` / `003_add_pet_owners` / `004_add_exam_result_qualitative_bounds`）も同じ方式で統合して、直下 DDL を単一ファイルへ戻した。2026-07-29 にさらに incremental 002–007（`pets.version` / exam_results index / inventory quantity CHECK / payments.clinic_id+clinic軸複合FK / payment method system_key 一致トリガー / owners phone 部分 unique）をセクション9へ同様式で統合した。2026-07-31 に append-only だった incremental 002–006（LSTEP day unique / closing EXCLUDE / identity links / LINE bot user id / medical-record image upload quota）をセクション10へ同様式で統合し、直下 DDL を再び `001_init.sql` 単一ファイルへ戻した。今後スキーマ変更を追加する場合も、適用済みファイルの checksum を変える影響を先に評価する。

直下 DDL の顔ぶれ・本数は固定ではない（増分の追加・`001_init.sql` への統合で変わる）。正の在庫は次の実測とする:

```bash
ls backend/migrations/*.sql
```

seed 側の構成は `seeds/{002_master,003_demo,004_staging}/` — CSV シードバンドル（`*.csv` + `manifest.json`。SQL ファイルではない。`internal/seedbundle.BundleOrder` 固定順）。`001_init.sql` に取り込まれた旧増分の本文は、末尾の統合セクション（セクション8・9・10 等）に原文・元コミット・SHA-256 付きで残る。

旧 002/003/004 の seed stub SQL、旧インデックス増分（`002_add_checkup_vaccination_indexes.sql` 等）、旧 005–012、2026-07-17 朝に一時的に存在した upgrade path incremental（`002_checkup_field_clinic_composite_fk.sql` / `003`–`011`）、2026-07-22〜27に追加された旧 incremental 002–009、2026-07-27 夕に統合した `002_pets_owners_clinic_composite_unique.sql` / `003_add_pet_owners.sql` / `004_add_exam_result_qualitative_bounds.sql`、2026-07-29に統合した `002_add_pets_version.sql`〜`007_owners_clinic_phone_unique.sql`、および2026-07-31に統合した `002_lstep_delivery_trigger_log_daily_unique.sql`〜`006_medical_record_image_upload_quota.sql` は全て独立ファイルとしては削除済み（統合当時の事実）。それら統合済み本文の所在は `001_init.sql` の統合セクション。**直下の現行増分の在庫は `ls backend/migrations/*.sql` を正とする**（本節にファイル名・本数を列挙しない）。

**統合前DBのno-resetアップグレード経路は存在しない**: 旧 `001_init.sql` が `schema_migrations` に記録済みの既存 DB（ローカル/STG/PROD）は、001 統合による checksum 変更で migrate が fail する。適用経路は `DB_RESET=true` のスキーマ再構築（USER手動）のみ。ローカルは `LOCAL_DB_RESET.md`、STGは明示承認後の再構築計画に従う。現行Cloudflare workflowに自動reset経路はない。

- **cmd/migrate は二段フェーズ構成**: ①直下の `*.sql`（本数は固定ではない。検算: `ls backend/migrations/*.sql`）を昇順適用 → ②`internal/seedbundle.BundleOrderForEnv(APP_ENV)` が許可した順で CSV バンドルを pgx `COPY FROM STDIN` ロード（`backend/cmd/migrate/csvbundle.go`）。local/dev/testは3 bundle、staging/production/空/未知は`002_master`のみ。DDL 失敗時は seed フェーズへ進まない
- 実行対象seedは `BundleOrderForEnv(APP_ENV)` が列挙する exact directoryだけ。`backend/migrations/seeds/_old_db_handoff/` はローカル保管用であり、`cmd/migrate` の入力bundleとして列挙しない
- **schema_migrations の記録キー**: DDL は従来通りファイル名。seed バンドルは `internal/seedbundle.BundleMigrationKey(bundleDir)` が返す `"seeds/<bundle>"`（例: `seeds/002_master`）— stub SQL ファイル名には二度と紐付かない。fresh DB 適用後の正しい終了状態の行数は **直下 `*.sql` の本数 + `BundleOrderForEnv(APP_ENV)` が許可したseed数**。DDL 本数の検算は `ls backend/migrations/*.sql`
- **両フェーズ後のキー突合（fail-closed）**: `cmd/migrate` は適用完了後に `Migration key coverage` 行を1行出す（`missing` / `extra` / `expected` / `recorded`）。`missing=0` なら期待キーは全て記録済み。欠落があれば非ゼロ終了する。`extra` は統合・削除でディスクから消えた履歴キーであり失敗にしない。再構築の成否は固定在庫数ではなくこのサマリー行で判定する
- **空履歴 + 既存schemaはfail-closed**: `schema_migrations`が空で`clinics`が既に存在する場合、`guardEmptyMigrationHistory`はschema完全性を検証できないため起動を拒否する。現行DDL/seedのchecksumを適用済みとして記録するbaseline処理は存在しない。USER承認済みのreset/再構築を行い、通常のDDL・seed適用経路を完走させる
- seed バンドルの checksum（`bundleChecksum`）は manifest.json + 全 CSV ファイルの内容を合成したもの。CSV だけの編集でも、既に適用済みの DB では通常の migration ファイル編集と同じ checksum mismatch ガードが働く
- COPY はシーケンスを進めないため、テーブルロード後に自動 setval される（`advanceSerialSequence`）
- **旧形式（stub SQL 時代）の seed キー**: `schema_migrations` に `002_seed_master.sql` 等の旧キーが残っている DB では、履歴が存在するため上記の空履歴ガードとは別に、`detectLegacySeedKeys` が旧stubと同等な `seeds/{002_master,003_demo,004_staging}` を現行キーへ翻訳して適用済み記録する（PR #186 P1-3 で fail-fast から translate 方式へ変更済み・CSV の再ロードはしない）。この翻訳はDDL側のchecksum検証を迂回せず、統合前001のchecksumを持つDBには上記のreset/再構築が必要
- CSV の正データ生成経路は使い捨てDBダンプのみ: `docker compose exec backend go run ./cmd/seed-export`。SQL の静的パース/手編集による CSV 生成は禁止（ON CONFLICT の最終マージ状態・`random()` 依存データは静的パースでは再現できない）
- `SEED_EXPORT_CSV_SOURCE` は旧7表adapterであり21表には使わない。現行 `cmd/seed-export` に21表handoff入力経路はない。21表CSVを `003_demo` へ直接コピーせず、専用adapter実装・検証までは `_old_db_handoff/` に隔離する
- テーブル→バンドル割当は「最初に触れたファイル」基準（earliest-file-wins、`cmd/seed-export/tables.go` の `bundleTables`）で固定済み。新しいシードテーブルの追加は `cmd/seed-export` の再設計が必要
- 実スタッフの氏名・email・password hashなどのPII/credential verifierをseedやGit履歴へ追加しない。スタッフ初期登録は、データ管理承認を得たsecret-managedな一回限りのimport経路で行う

## seed / migration 差し替え時の注意

- 適用済み migration（`001_init.sql` および incremental DDL）/ seed バンドル（CSV・manifest.json）を編集すると既存 DB は checksum mismatch になる
- seed データの内容変更は CSV の手編集ではなく `cmd/seed-export` の再実行で行う（上記参照）
- DB 非依存の最低検証は `python3 scripts/verify_seed.py`（CSV ベース）
- 運用メモ: `docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md`

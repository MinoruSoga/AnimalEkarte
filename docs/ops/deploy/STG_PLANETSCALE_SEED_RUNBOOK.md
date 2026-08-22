# PlanetScale STG シードデータ投入 Runbook

> **目的**: 2026-07-15 に `DROP SCHEMA public CASCADE` で初期化した PlanetScale STG（`animalekarte-stg` / `main` ブランチ）へ、次の CF デプロイ後にデータ（seed）を復元・検証する手順を定義する。
> **読者**: DevOps/開発者。
> **タイミング**: PlanetScale STG のスキーマ初期化直後〜次回 CF デプロイ前後。

---

## 0. 背景

2026-07-15、migration 002〜004 の stub SQL を 001 へ統合したコミット（`fb758d50` chore(backend): migration 002-004 を 001 に統合し削除する）により、STG の `schema_migrations` に残っていた旧キー（`002_seed_master.sql` 等）と現行バイナリの整合性リスクを避けるため、PlanetScale STG（`animalekarte-stg`）に対して直接 `DROP SCHEMA public CASCADE` を実行し初期化した。次の CF デプロイで `wrangler deploy` → `POST /_internal/migrate` が走れば `backend/migrations/001_init.sql`（DDL）は復元されるが、**そのデータ（seed）が復元されるかどうかは別途確認が必要**というのが本書の出発点。

調査の結論を先取りすると、**cmd/migrate は DDL 適用後に seed バンドルの投入まで同一プロセス内で行う**（§2.1〜2.2 参照）。したがって「schema 復元 → seed は別途手作業」という前提は、STG が *真に空のスキーマ*（`public` スキーマごと存在しない状態）である限り成立しない。本書は、この自動投入を前提にした**検証中心の手順**と、それでも人手が要る分岐（`public` スキーマの実在確認、フルデモ投入）を扱う。

---

## 1. 方式比較（結論: seed 再投入を主手順とする）

| 案 | 概要 | 所要目安 | リスク | 採否 |
|---|---|---|---|---|
| **A. seed 再投入**（本書の主手順） | `backend/migrations/seeds/{002_master,003_demo,004_staging}/` の CSV を `cmd/migrate` の既存メカニズムで自動投入し、検証する | 数分（migrate 自体）＋検証 10〜15分 | 低（既存の起動時経路をそのまま使うだけ。新規コードなし） | **採用** |
| B. 旧 RDS ダンプ移行（凍結比較） | AWS 廃止前に検討した方式。基盤廃止後は実行不能 | — | 本番同等の PII を STG に持ち込むリスク。旧基盤の復元も禁止 | **退役・実行禁止**（§6） |

STG はデモデータ運用（`docs/ops/deploy/STG-DEMO-DATA-LIFECYCLE.md` §2.1「Seed Data」）であり、実データの忠実な移行を要求されていない。したがって A を主手順とする。

---

## 2. 調査結果（根拠）

### 2.1 cmd/migrate は DB_RESET に関わらず seed を適用する

`backend/cmd/migrate/main.go` の `run()`:

- `DB_RESET` は `resetSchema`（`DROP SCHEMA public CASCADE` → `CREATE SCHEMA public`）の実行有無だけを制御する（main.go:87-94, 525-552）。
- seed バンドルのロード（`runSeedBundles`, main.go:463-523）は `DB_RESET` の値と**無関係**に、`runSQLMigrations`（DDL 適用, main.go:372-461）の直後に必ず呼ばれる（main.go:115-125）。
- `runSeedBundles` は各バンドルが `schema_migrations` に未記録の場合のみ CSV を投入する（`isAlreadyApplied` ガード, main.go:486-494）。**未記録かどうか**が投入有無を決める唯一の条件であり、`DB_RESET` は関係しない。
- 空のmigration履歴に既存アプリケーションschemaがある場合のガードは `guardEmptyMigrationHistory`（main.go:295-328）。`clinics` テーブルが存在すればschema完全性を検証できないためfail-closedで停止し、現行DDL/seedのchecksumは記録しない。`clinics` が存在しないfresh DBだけが、そのまま通常のDDL・seed適用へ進む。

CF 経路（`POST /_internal/migrate` → `Container.exec(["/app/migrate"])`）は起動引数が固定で `DB_RESET` を注入する経路が構造的に存在しない。

**結論**: `DROP SCHEMA public CASCADE` 実行後の STG は「`clinics` テーブルが存在しない = 新規 DB」と cmd/migrate から見える。したがって次の `POST /_internal/migrate` は DB_RESET の値に関係なく、直下 DDL（`ls backend/migrations/*.sql` を正とする）を昇順に適用後、002_master → 003_demo → 004_staging の順で CSV を **自動投入する**（`seedbundle.BundleOrder`, `backend/internal/seedbundle/manifest.go`）。fresh DB の終了状態は、`schema_migrations` の行数が直下 DDL 本数 + seed バンドル数に一致することである。一方、統合前001が記録済みの現行STGへ通常の `POST /_internal/migrate` を実行するとchecksum mismatchでfailする。現行Cloudflare経路は `DB_RESET` を注入できないため、明示承認した再構築を先に完了させる必要がある。

### 2.2 ただし前提条件が一つだけある: `public` スキーマの実在

`resetSchema`（main.go:591-617）は `DROP SCHEMA public CASCADE` と `CREATE SCHEMA public` を**対で**実行する。しかし今回 STG に対して実行されたのはこのペアではなく、手動での `DROP SCHEMA public CASCADE` 単体（背景§0）。`CREATE SCHEMA public` が対で実行されていない場合、`public` スキーマ自体が存在しない状態になる。

この場合、migrate の次のステップ `ensureMigrationsTable`（main.go:97, 140-150）が実行する `CREATE TABLE IF NOT EXISTS schema_migrations (...)` はスキーマ未指定（unqualified）のため、PostgreSQL は `no schema has been selected to create in` で失敗する。これは `POST /_internal/migrate` の非ゼロ終了 → `backend-deploy.yml` の `Run database migration` ステップ失敗 → デプロイ全体 abort という**大声で失敗する**経路になる（サイレント障害にはならない）が、そうなると「次の CF デプロイで seed まで自動復元される」という前提が崩れる。

→ **§4 の事前チェックは省略不可**。`public` スキーマの実在を先に確認し、なければ CF デプロイ前に手動で作成する。

### 2.3 seed バンドルの中身（002_master / 003_demo / 004_staging）

`backend/migrations/seeds/` は CSV + `manifest.json` のディレクトリ3本のみで、stub SQL は 2026-07 に削除済み（`docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md:11`）。`cmd/migrate` は `manifest.json` の `tables` 配列順に pgx `COPY FROM STDIN` で各 CSV をロードする（`backend/cmd/migrate/csvbundle.go:99-116`）。

| バンドル | テーブル数（manifest） | 内容 | git 追跡状態 |
|---|---:|---|---|
| `002_master` | 5 | company, animal_species, lstep 系マスタ | 通常追跡（小容量、合計 <1KB） |
| `003_demo` | 85 | owners/pets/medical_records/billings 等の業務デモデータ一式 | 全 91 ファイルが通常追跡（`H`）。CSV は Git LFS。skip-worktree ではない |
| `004_staging` | 1 | `appointment_trimming_details`（現状 0 行） | 通常追跡 |

**直近コミットで変わった点**（`e273545d5d2604856ecee4639bcac5c201534f2c` feat(backend): CSV 取込アダプタを追加し、フル 003_demo はローカル専用とする）:

- `backend/internal/csvimport/import.go` を新設。旧DB由来の owners/pets/medical_records/exams/exam_results/billings/billing_items の7テーブルを、**使い捨てローカルDB**に対してのみ投入するアダプタ（`import.go:17-45`）。`cmd/seed-export`（`main.go:104-116`）が `SEED_EXPORT_CSV_SOURCE` 環境変数経由でこれを呼び出し、投入結果を再び `COPY ... TO STDOUT` で CSV ダンプして `backend/migrations/seeds/003_demo/` を再生成する、という**ローカル専用**のパイプライン。STG/PlanetScale へ直接書き込む経路ではない。
- 旧 `docs/ops/deploy/ANIMALEKARTE_CSV_IMPORT_COMPLETION.md:43,65-67`（2026-08-20 削除。`git show 1bd219ff9^:docs/ops/deploy/ANIMALEKARTE_CSV_IMPORT_COMPLETION.md`）: GitHub の 100MB ファイルサイズ制限を回避するため、フルデモ（529MB、owners 10,370 / pets 15,654 / medical_records 425,544 / billings 392,105 / billing_items 1,542,422 / exam_results 1,322,503 行）は `old_db/sensitive-local/animalekarte-003-demo-full/` に**ローカルのみ**保持し、リポジトリの `003_demo` は小さいデモのまま維持する方針に確定した。

**本セッションで確認した現在の実体**（2026-07-16 時点）:

- **git にコミットされている内容**は Git LFS 経由のフルデモである（`git show HEAD:backend/migrations/seeds/003_demo/exam_results.csv` は LFS ポインタ、size 174379631）。小さいデモではない。2026-07-16 時点で計測した主要テーブルの行数（下表は当時の小さいデモ。現行 HEAD の行数ではない）:

  | テーブル | 行数（コミット済み） |
  |---|---:|
  | clinics | 4 |
  | staffs | 36 |
  | permission_groups | 8 |
  | exam_types | 23 |
  | owners | 61 |
  | pets | 78 |
  | appointments | 810 |
  | medical_records | 136 |
  | billings | 64 |
  | billing_items | 52 |
  | exam_results | 53 |

- **ローカル作業ツリー**は skip-worktree の隠しオーバーレイではない。`git ls-files -v backend/migrations/seeds/003_demo/` は全 91 ファイル `H`、`git status --porcelain backend/migrations/seeds/` は空で、作業ツリーは HEAD のコミット済みオブジェクトと一致する。HEAD 自体がフルデモ（`exam_results.csv` は Git LFS、size 174379631）であり、GitHub の 100MB 制限は LFS で扱っている。未コミットでも `git add` 対象外でもない。

→ HEAD（LFS smudge 後の作業ツリーも同じ）はフルデモである。committed の小さいデモを取り出す経路はない。

### 2.4 凍結履歴: 廃止した STG seed 投入経路

AWS ECS/RDS の `db_reset` workflow は基盤廃止時に削除済みで、復元・実行しない。
現行 `backend-deploy.yml` に `db_reset` 入力はない。`make csv-import` は old_db（旧カルテ）由来データを
**正式な CSV cutover 経路**として投入するためのコマンドであり、PlanetScale STG への手動 direct-import 代替ではない。
当時の AWS 手順を調査する場合だけ git 履歴の凍結記録を参照する（2026-08-20 削除。`git show e0260d32f^:docs/ops/infra/_archive/aws-legacy/` 配下）。

### 2.5 PlanetScale への直接投入経路: `pscale role` + psql

PlanetScale Postgres は `pscale role create <database> <branch> <name> --inherited-roles <roles> --ttl <duration>` で、期限付き（TTL）の Postgres ロールを都度発行できる（ローカル `pscale role create --help` で確認済み。`--ttl duration` は `"2h"` 等を受け付け、デフォルト無期限）。STG では `noah-animalekarte` 組織の `animalekarte-stg` データベース・`main` ブランチが対象（`infra/scripts/pscale-create-stg.sh:8-11`）。

既存運用は `pscale role reset-default`（アプリ本体が使う既定 `postgres` ロールのパスワードを都度再発行・失効）だが、これはアプリ稼働用の共有クレデンシャルを毎回ローテーションする前提で、本番稼働中の Worker/Hyperdrive 設定にも影響する。本書の検証・任意投入作業は本番トラフィックに影響しない**別ロール**を使うべきなので、`pscale role reset-default` ではなく `pscale role create` で使い捨てロールを発行する（TTL 失効で自動的に片付く）。

直結接続が必要な理由: `cmd/migrate` は `pg_advisory_lock` を使うため Hyperdrive 経由の接続では動作しない（`infra/scripts/pscale-create-stg.sh:34-35` 「Hyperdrive 経由では advisory lock が非対応」）。同じ理由で、手動の `CREATE SCHEMA public;` やロールバック用 `DROP SCHEMA` も **Hyperdrive を経由しない直結接続**（`pscale role` で発行したロールの host/port へ直接 psql）で行う。

`backend/cmd/seed-export` はローカル専用ツール（安全ガードとして `DB_HOST` が `db`/`localhost`/`127.0.0.1` 以外を拒否する。`main.go` package doc・`dbconn.IsLocalHost` 使用箇所）であり、PlanetScale への逆方向（import）に相当する既存ツールは**ない**。フルデモを STG へ入れる場合は psql の `\copy`（クライアント側ストリーム）を手動で使うしかない。`COPY ... FROM '/path'`（サーバ側ファイル読み込み）は PlanetScale 側にファイルが存在しないため使えない点に注意（`cmd/migrate` 自身も `COPY FROM STDIN` を使っており、サーバ側パス読み込みではない。`csvbundle.go:173-178`）。

---

## 3. 前提

- 本書の「推奨手順」は、**次の CF デプロイで `POST /_internal/migrate` が正常終了していること**（`backend-deploy.yml` の `Run database migration` ステップが exit 0）を前提とする。
- ただし §2.2 の通り、`public` スキーマが存在しないまま migrate を叩くとこの前提そのものが崩れる（migrate が失敗する）。**§4 の事前チェックを先に実施してから CF デプロイをトリガーすること。**
- 本書はコード変更を伴わない。`backend/migrations/seeds/` の内容にも触れない（§5 Step E のデータ投入はいずれもデータ投入のみで、seed バンドルファイル自体は編集しない）。

---

## 4. 事前チェック（CF デプロイをトリガーする前に一度だけ）: `public` スキーマの実在確認

```bash
# 1. TTL付きの検証用ロールを発行（2時間で自動失効。値は画面表示のみ、ファイル/ログ/チャットに残さない）
pscale role create animalekarte-stg main stg-seed-precheck \
  --org noah-animalekarte \
  --inherited-roles postgres \
  --ttl 2h

# 2. 表示された host/port/user/password/database をこのシェルセッションのみの変数に代入
export PGHOST="<表示された host>"
export PGPORT="<表示された port>"
export PGUSER="<表示された user>"
export PGPASSWORD="<表示された password>"
export PGDATABASE="<表示された database>"

# 3. public スキーマの実在確認
psql -c "SELECT schema_name FROM information_schema.schemata WHERE schema_name = 'public';"
```

**分岐（人間の判断が必要）**:

- 1行返る（`public` が存在） → 何もしなくてよい。§5 へ進む。
- 0行（`public` が存在しない） → 次を実行してから §5 へ進む:
  ```bash
  psql -c "CREATE SCHEMA public;"
  ```
  実行前に、STG に他エンジニアがアクセス中でないか確認すること（`STG-DEMO-DATA-LIFECYCLE.md` §7.4 の確認事項に準ずる）。

```bash
# 4. 検証が終わったら忘れずに失効（TTL任せでもよいが即時失効推奨）
pscale role list animalekarte-stg main --org noah-animalekarte
pscale role delete animalekarte-stg main <role-id> --org noah-animalekarte
```

所要目安: 5分以内。

---

## 5. 推奨手順: seed 再投入（cmd/migrate の自動投入を使う）

### Step A. CF デプロイをトリガー

以下いずれか（実行は人間が行う。本書は手順の提示のみ）:

```bash
# 方法1: staging ブランチへの push（backend/** に変更がある場合。今回はデータのみの復元なので
#        変更がなければ push では走らない点に注意）
# 方法2: 手動トリガー
gh workflow run backend-deploy.yml --ref staging
```

`backend-deploy.yml` は `wrangler deploy` → `POST /_internal/migrate`（`infra/scripts/cf-run-migrate.sh`）→ `/health` ポーリングの順で実行される（`.github/workflows/backend-deploy.yml:60-121`）。

所要目安: `wrangler deploy` 数十秒 + migrate（seed バンドル合計 <1MB・約1,200行なので数秒〜長くても Worker 側 exec timeout の 120秒以内）+ health check 最大 6分（30秒×12回ポーリング）。

### Step B. デプロイ結果の確認

```bash
gh run list --workflow backend-deploy.yml --branch staging --limit 1
gh run view <run-id> --log | grep -E "Seed bundle summary|✓ Migration completed|✓ Seed bundle loaded|schema_migrations"
```

`Run database migration` ステップが成功していれば、`cmd/migrate` のログに `Seed bundle summary applied=3 skipped=0 total=3`（`main.go:582-585`）相当の出力が出るはず。

### Step C. 検証用ロールを発行し、直接 psql で確認

```bash
pscale role create animalekarte-stg main stg-seed-verify \
  --org noah-animalekarte \
  --inherited-roles pg_read_all_data \
  --ttl 1h
export PGHOST="<host>"; export PGPORT="<port>"; export PGUSER="<user>"
export PGPASSWORD="<password>"; export PGDATABASE="<database>"
```

読み取り専用の検証なら `--inherited-roles pg_read_all_data` で十分（`postgres` ロール継承は書き込み系操作が必要な Step E でのみ使う）。

### Step D. 検証クエリ（テーブル別件数 + 主要マスタの存在確認）

```sql
-- 1. schema_migrations が「直下 DDL 本数 + seed バンドル数」そろっているか（fresh apply の正しい終了状態）
--    SEED_MIGRATION_OPERATIONS.md:18 の期待値
SELECT filename, checksum, executed_at FROM schema_migrations ORDER BY filename;
-- 期待: 001_init.sql /
--       seeds/002_master / seeds/003_demo / seeds/004_staging の4行
-- checksum の期待値（2026-07-16 時点、git HEAD の committed 内容から算出。
--   seeds/*.csv や manifest.json を編集した場合は再計算が必要。算出方法は §5 Step E-a 参照）:
--   seeds/002_master = 5a46c460e51bf617602c0812f100d077df36a3f5855a85d23ba84f63a2bf9945
--   seeds/003_demo   = c3e86a7c78d6d1b654ecd2ce4657e77402e8fb3a4f896b5c0172df3416c095e5
--   seeds/004_staging= 3cb6a3292700248ef2c3835154c070b34218122de99196f054e4721aec4319d1

-- 2. 002_master（5テーブル）
SELECT 'companies' t, count(*) FROM companies
UNION ALL SELECT 'animal_species', count(*) FROM animal_species
UNION ALL SELECT 'lstep_auto_managed_prefixes', count(*) FROM lstep_auto_managed_prefixes
UNION ALL SELECT 'lstep_condition_tag_mappings', count(*) FROM lstep_condition_tag_mappings
UNION ALL SELECT 'lstep_send_purpose_tag_prefixes', count(*) FROM lstep_send_purpose_tag_prefixes;
-- 期待: companies=1, animal_species=6, lstep_auto_managed_prefixes=19,
--       lstep_condition_tag_mappings=7, lstep_send_purpose_tag_prefixes=4

-- 3. 003_demo（主要マスタ・業務データの存在確認。小さいデモの期待値）
SELECT 'clinics' t, count(*) FROM clinics
UNION ALL SELECT 'staffs', count(*) FROM staffs
UNION ALL SELECT 'permission_groups', count(*) FROM permission_groups
UNION ALL SELECT 'exam_types', count(*) FROM exam_types
UNION ALL SELECT 'owners', count(*) FROM owners
UNION ALL SELECT 'pets', count(*) FROM pets
UNION ALL SELECT 'appointments', count(*) FROM appointments
UNION ALL SELECT 'medical_records', count(*) FROM medical_records
UNION ALL SELECT 'billings', count(*) FROM billings
UNION ALL SELECT 'billing_items', count(*) FROM billing_items
UNION ALL SELECT 'exam_results', count(*) FROM exam_results;
-- 期待: clinics=4, staffs=36, permission_groups=8, exam_types=23, owners=61, pets=78,
--       appointments=810, medical_records=136, billings=64, billing_items=52, exam_results=53
-- (§2.3 の表と同一。85テーブル全件を確認したい場合は各テーブル名で
--  `git show HEAD:backend/migrations/seeds/003_demo/<table>.csv | wc -l` からヘッダ1行を引いた値と突合する)

-- 4. system_admin グループ・初期 clinic の存在（STG-DEMO-DATA-LIFECYCLE.md §2.1 の必須 seed data）
SELECT id, name FROM permission_groups WHERE id = 1;
SELECT id, name FROM clinics ORDER BY id LIMIT 1;

-- 5. 004_staging
SELECT count(*) FROM appointment_trimming_details;
-- 期待: 0（現状の manifest では空テーブル）

-- 6. デモアカウントでログイン可能か（DB上ではなくAPI経由で確認。STG-DEMO-DATA-LIFECYCLE.md §7.4）
--    curl -X POST https://animalekarte-stg-api.baritech-soga.workers.dev/api/v1/auth/login ...
```

すべて期待値通りであれば復元完了。ズレがあれば `schema_migrations` の該当行の有無・`checksum mismatch` エラーの有無を migrate ログで再確認する。

所要目安: 10〜15分。

### Step E（オプション・人間の判断が必要）: pscale role + psql による直接投入

投入対象を取り違えないこと。LFS 化（`2d58e64d2`）以降、HEAD の `003_demo` はフルデモである。E-a は staging 自動投入の再現（`002_master` のみ）。フルデモを STG に載せる判断は E-b（本書スコープ外）に残す。

#### E-a. 緊急フォールバック — 自動投入（Step A）が失敗した場合の手動再現

`POST /_internal/migrate` 自体が届かない・恒常的に失敗するなど、CF デプロイ経由の自動投入（§2.1）が使えない場合の最終手段。**スキーマが真に空（91テーブル全て 0 行）の場合にのみ**実施する。

重要な注意点: HEAD および LFS smudge 後の作業ツリーは同じフルデモである。`git archive` は 2026-07-16 の小さいデモを得る手段ではない（小さいデモは HEAD に無い）。E-a は Step A の手動再現であり、staging の `cmd/migrate` は `BundleOrderForEnv("staging")` で `002_master` のみを載せる。作業ツリーから `003_demo` を `\copy` すると committed のフルデモが STG に入る。それは自動投入の再現ではない。

```bash
# 1. staging 自動投入と同じく 002_master のみ（003_demo のフルデモは載せない）
SEEDS="backend/migrations/seeds"

# 2. 書き込み可能な検証用ロールを発行
pscale role create animalekarte-stg main stg-seed-fallback \
  --org noah-animalekarte \
  --inherited-roles postgres \
  --ttl 1h
export PGHOST="<host>"; export PGPORT="<port>"; export PGUSER="<user>"
export PGPASSWORD="<password>"; export PGDATABASE="<database>"

# 3. 002_master のみ。manifest.json のテーブル順。
psql -v seeds="$SEEDS" <<'EOSQL'
\copy companies FROM :'seeds'/002_master/companies.csv WITH (FORMAT csv, HEADER true)
\copy animal_species FROM :'seeds'/002_master/animal_species.csv WITH (FORMAT csv, HEADER true)
\copy lstep_auto_managed_prefixes FROM :'seeds'/002_master/lstep_auto_managed_prefixes.csv WITH (FORMAT csv, HEADER true)
\copy lstep_condition_tag_mappings FROM :'seeds'/002_master/lstep_condition_tag_mappings.csv WITH (FORMAT csv, HEADER true)
\copy lstep_send_purpose_tag_prefixes FROM :'seeds'/002_master/lstep_send_purpose_tag_prefixes.csv WITH (FORMAT csv, HEADER true)
EOSQL

# 4. SERIAL シーケンスを cmd/migrate と同じロジックで一括前進（csvbundle.go:124-151 相当）
psql <<'EOSQL'
DO $$
DECLARE
  r record; seq regclass; max_id bigint;
BEGIN
  FOR r IN SELECT table_name FROM information_schema.columns
           WHERE table_schema = 'public' AND column_name = 'id'
  LOOP
    seq := pg_get_serial_sequence('public.' || quote_ident(r.table_name), 'id');
    IF seq IS NOT NULL THEN
      EXECUTE format('SELECT max(id) FROM %I', r.table_name) INTO max_id;
      PERFORM setval(seq, COALESCE(max_id, 1), max_id IS NOT NULL);
    END IF;
  END LOOP;
END $$;
EOSQL

# 5. 必須の後処理: 投入したバンドルだけ schema_migrations に記録する。
#    staging 自動投入は seeds/002_master のみ。003_demo/004_staging の旧 checksum を
#    書かない（実データを載せていないキーを記録すると次回 migrate がスキップする）。
psql <<'EOSQL'
INSERT INTO schema_migrations (filename, checksum, executed_at) VALUES
  ('seeds/002_master', '5a46c460e51bf617602c0812f100d077df36a3f5855a85d23ba84f63a2bf9945', now())
ON CONFLICT (filename) DO NOTHING;
EOSQL
```

これは通常の `runSeedBundles` がCSV投入成功後に `recordMigration` する順序（main.go:496-510）と同じく、実データ投入の完了後に対応する履歴を記録する手動復旧である。`guardEmptyMigrationHistory`（main.go:295-328）は読取り専用で、空履歴の既存schemaへ現行checksumを書き込まない。`translateLegacySeedKeys`（main.go:220-272）はmigration履歴に旧stubキーが存在する場合だけ、CSVを再投入せず現行seedキーへ翻訳する別の互換処理である。

#### E-b. フルデモ投入（ローカルのみ・本書のスコープ外 — 着手前に別途判断すること）

§2.3 の通り、フルデモは Git LFS で HEAD にコミット済みである（skip-worktree の隠しオーバーレイではない。GitHub 100MB 制限は LFS が担う）。staging の `cmd/migrate` はなお `002_master` のみを自動投入する。フルデモを STG に載せるには migrate 経路を使わず、別判断が要る。

これを STG に入れたい場合、**E-a（002_master のみ）と単純に組み合わせることはできない**。理由:
- E-a のあとに `003_demo` フルデモの一部テーブルだけを上書きすると、参照先が欠けた dangling FK になり、データ整合性が壊れる。
- 実施する場合はスキーマを空の状態（§7 ロールバック後）から、専用手順で扱う。
- schema_migrations に書く checksum は実際に投入した内容と一致させないと、次回 migrate が不一致/再投入を試みる。

本書はここまでの制約整理に留める。505MB・91テーブル規模の初回実行を無検証の手順書だけで進めるのはリスクが高く、実施する場合は本書とは別に、実施前提（実施理由・実施者・許容ダウンタイム）を明確にした専用の手順で扱うべきと判断した。

**判断が必要な点**:
- E-a と E-b のどちらも、通常の git 管理下のシード投入経路をバイパスする手動操作。**再現性なし**（§7 のロールバック後は自動では戻らない。次に必要になった時は同じ手順を再実行する）。
- E-b を実施する場合、505MB を東京リージョンの PlanetScale へ psql `\copy` で送る所要時間は未実測（数十分規模を想定）。ネットワーク切断リスクを許容できるか、業務時間内に実施できるかは実施者の判断。
- 実施要否そのものが判断事項: STG は「デモデータ運用」（§1）なので、`002_master` の自動投入（または E-a）で業務上十分なら E-b は不要というのが本書の既定スタンス。

---

## 6. 凍結比較: 旧 RDS ダンプ移行（退役・実行禁止）

> AWS ECS/RDS 基盤は 2026-07-20 に廃止済みで、ライブな取得元や切り戻し先ではない。
> 以下は当時の方式選定理由を残す比較記録であり、旧基盤や IaC を復元して実行してはならない。

| 観点 | seed 再投入（採用） | 旧 RDS ダンプ案（凍結） |
|---|---|---|
| 実データ忠実性 | デモデータのみ（本番同等ではない） | 当時は旧 DB の実データを反映できたが、現在は基盤廃止済み |
| 実装コスト | ゼロ（既存 cmd/migrate をそのまま使う） | 退役済み経路の再構築が必要なため禁止 |
| スキーマ整合性 | `cmd/migrate` が保証（同一 DDL から生成） | 旧スキーマとの差分検証が別途必要 |
| PII/コンプライアンス | デモデータのみ、リスク低 | 本番同等データを STG に複製することになり、旧 `ANIMALEKARTE_CSV_IMPORT_COMPLETION.md`（削除済み・git 履歴）で懸念された「PHI が STG に残る」問題を re-introduce しかねない |
| 運用方針との整合 | `STG-DEMO-DATA-LIFECYCLE.md` の「STG=デモデータ運用」方針に合致 | 方針からの逸脱（要 PO 判断） |

**結論**: 旧 RDS ダンプ案は不採用の凍結履歴であり、現在は実行禁止。STG は現行 seed / snapshot 手順だけで復旧する。

---

## 7. ロールバック

現在の状態（スキーマなし、または投入直後で問題が見つかった場合）から再度やり直す手順。

```bash
# 1. 直結ロールを発行
pscale role create animalekarte-stg main stg-rollback \
  --org noah-animalekarte \
  --inherited-roles postgres \
  --ttl 1h
export PGHOST="<host>"; export PGPORT="<port>"; export PGUSER="<user>"
export PGPASSWORD="<password>"; export PGDATABASE="<database>"

# 2. スキーマを完全初期化（今回と同じ操作。必ず対で実行し public 不在状態を残さない）
psql -c "DROP SCHEMA public CASCADE;"
psql -c "CREATE SCHEMA public;"

# 3. ロールを失効
pscale role delete animalekarte-stg main <role-id> --org noah-animalekarte
```

その後は §5 Step A（CF デプロイのトリガー）からやり直す。`public` を対で再作成しているため §4 の事前チェックは今回は不要（ただし習慣として一度確認しても良い）。

---

## 8. 人間の判断が必要な分岐（まとめ）

1. **§2.2 / §4**: `public` スキーマが実在するか。実在しなければ CF デプロイ前に手動で `CREATE SCHEMA public;` が必須（自動化されていない・migrate 側では検出できず単に失敗する）。
2. **§5 Step E-b**: HEAD の LFS フルデモを STG に投入するか。staging 自動投入および E-a は `002_master` のみで、業務要件をそれで満たすなら不要。投入する場合は本書がスコープ外とした専用手順を別途起こす必要があり、所要時間未実測・再現性なしの手動操作である点を許容できるか。
3. **§7 ロールバック**: 検証で問題が見つかった場合、再度スキーマごと作り直すか、部分修正で済ませるか（本書はスキーマごと作り直す手順のみ用意）。
4. **§5 Step B**: `POST /_internal/migrate` が非ゼロ終了した場合、`checksum mismatch`（seed ファイル改変）なのか `public` スキーマ不在（§2.2）なのか、ログを見て切り分けるのは実施者の判断。

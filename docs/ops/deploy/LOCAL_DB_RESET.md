# ローカル開発環境：DB再構築手順 (Local DB Reset)

> **目的**: ローカルDB再構築の標準手順を定義する（単一入口・回復可能・fail-closed）。
> **読者**: 開発者（USER）。エージェントは実行しない。
> **タイミング**: ローカル checksum mismatch 発生時、またはローカルDBを意図的に初期化したい時。

> **Animal Ekarte**: マイグレーション統合に伴う不整合の解消手順
> **最新更新**: 2026-07-30（TASK-607: `make reset` を recoverable 単一入口に置換）

---

## 1. 発生する問題

スキーマの整理やマイグレーションファイルの統合（Squash）を行った際、ローカル環境の `schema_migrations` テーブルに記録されたチェックサムが、最新の SQL ファイルと一致しなくなり、バックエンドの起動に失敗することがあります。

現行の migrate runner は `DB_RESET=true` 以外に checksum mismatch の非破壊更新経路を持たないため、**local では volume 再構築が正規経路**です。

---

## 2. 単一入口（USER のみ）

```bash
make reset
```

実体は `scripts/local-db-reset-contract.sh` です。次を **1 アクションで fail-closed** に実行します。

| 段階 | 内容 |
|------|------|
| 1. 環境照合 | 固定 project 名 `animalekarte` と固定 volume 名を compose 実測と照合。`APP_ENV` が production/staging 系なら拒否。 |
| 2. 回復 snapshot | `umask 077` で追跡外 `.local-db-backups/<UTC>/` に owner-only の gzip 済み `pg_dumpall`、SHA-256、対象 volume + DDL/seed key manifest を作成。空 dump / 空 digest / 書き込み失敗なら **ここで停止**（volume は消さない）。 |
| 3. 削除 | サービス停止（named volume は保持）。**`ekarte-postgres-data` だけ**を `docker volume rm`。cache 3 件（`ekarte-frontend-node-modules` / `ekarte-go-mod-cache` / `ekarte-go-build-cache`）は保持。 |
| 4. 再起動 + postflight | `db backend frontend` を `--wait` で起動。migration key coverage `missing=0`、直下 DDL 全件、seed `002_master`、`schema_migrations` 契約、backend healthy、`/health` HTTP 200 を確認。不足があれば非 0。 |

### 2.1 失われるもの / 保持されるもの

| 失われるもの | 保持されるもの |
|--------------|----------------|
| local Postgres cluster 内の **全 DB / global role / schema / data / migration 履歴 / seed 適用結果 / 手入力行** | bind mount の source / docs / `.env*` |
| （上に伴い）local でしか持っていない未コミット相当の DB 行 | cache 3 volume（frontend `node_modules`、Go module cache、Go build cache） |
| | object storage や repo 外のバックアップ |
| | `.local-db-backups/` に残る直前 snapshot（今回の実行で作成） |

### 2.2 禁止事項

- STG / PROD / 共有環境で `make reset` を実行しない。
- エージェントは `make reset` を自動実行しない（USER 明示操作のみ）。
- compose の全 volume 一括削除や volume store の prune を reset 経路に使わない（cache を巻き込むため）。

---

## 3. 正常終了の確認

適用対象の直下 DDL 本数は固定ではない。先に数える:

```bash
ls backend/migrations/*.sql
```

バックエンドのログでは、**直下の各 DDL ファイルごとに** 1 行の `Migration completed file=<ファイル名>` が出たあと、サマリが出る（1 本だけの例を正常終了の条件にしない）。形:

```text
Migration completed file=<各トップレベル *.sql のファイル名>
…（`ls backend/migrations/*.sql` の件数ぶん並ぶ）
Migration summary applied=N skipped=0 total=N
Seed bundle loaded bundle=002_master

Seed bundle summary applied=1 skipped=0 total=1
Migration key coverage missing=0 extra=X expected=E recorded=R
```

ここで `N` は `ls backend/migrations/*.sql` の件数。

**成否の一次判定は `Migration key coverage` の 1 行**とする。`missing=0` なら、ディスク上の直下 DDL と seed バンドルの期待キーがすべて `schema_migrations` に記録されている。`extra` は統合・削除でディスクから消えた履歴キーの件数であり、0 でなくても失敗ではない。固定の行数期待で照合しない。

`schema_migrations` の行数も固定値ではない。**行数 = 直下 `*.sql` の本数 + seed 1**（キーは各 DDL ファイル名と `seeds/002_master`）。検算は上記 `ls` の件数に 1 を足したものと、DB の `SELECT COUNT(*) FROM schema_migrations` を照合する（余剰履歴がある環境では `recorded` がこの導出より大きくなり得る）。

`/health` の HTTP 200 と `make reset` postflight は backend の liveness と migration postflight の成功だけを示します。臨床入力の準備完了には、スタッフアカウントの払い出し、handoff/import、ログイン、必要権限を別に確認します。

---

## 4. 手動の分解手順（参考・通常は不要）

単一入口が使えない緊急時のみ。通常は `make reset` を使うこと。

### 4.1 回復 snapshot（推奨）

```bash
umask 077
mkdir -p ".local-db-backups/$(date -u +%Y%m%dT%H%M%SZ)"
# db コンテナ上で pg_dumpall を gzip し、sha256 と manifest を同じディレクトリへ
```

### 4.2 ボリュームの削除（DB のみ）

```bash
docker compose --env-file .env.local down --remove-orphans
docker volume rm ekarte-postgres-data
```

※ `docker-compose.yml` 内でボリュームの `name` が `ekarte-postgres-data` として明示的に指定されているため、プロジェクト名のプレフィックス（`animalekarte_` 等）は付加されません。cache volume（`ekarte-frontend-node-modules` / `ekarte-go-mod-cache` / `ekarte-go-build-cache`）は削除しないでください。

### 4.3 再起動と自動構築

```bash
make up
```

起動時に `backend/migrations/` 直下の `*.sql` がファイル名昇順で適用された後、`002_master` の CSV シードバンドルがロードされます。直下 DDL の本数は固定ではない。臨床行がある場合は postflight 後の `_old_db_handoff` import。

2026-07-27 統合前の 001 が適用済みの DB は checksum mismatch になるため、この再構築が必須です。

---

## 5. 契約テスト（破壊的操作なし）

```bash
bash scripts/local-db-reset-contract.test.sh
bash scripts/check-reset-wait-services.test.sh
```

実 DB に対する `make reset` は **USER が明示承認したときだけ** 実行する。エージェントの自動実行対象外。

---

## 6. 注意事項

- **データ消失**: local Postgres cluster 内のデータは全て削除されます。直前 snapshot は `.local-db-backups/`（git 非追跡）に残ります。
- **共有環境**: ステージング等の共有環境では、決してこの手順（ボリューム削除）を実行しないでください。現行 workflow は DB を再作成しません。再構築が必要な場合は、破壊的操作の明示承認を得て [STG_PLANETSCALE_SEED_RUNBOOK.md](./STG_PLANETSCALE_SEED_RUNBOOK.md) に従います。

---

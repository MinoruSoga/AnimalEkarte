# ローカル開発環境：DB再構築手順 (Local DB Reset)

> **目的**: ローカルDB再構築の標準手順を定義する。
> **読者**: 開発者。
> **タイミング**: ローカルchecksum mismatch発生時。

> **Animal Ekarte**: マイグレーション統合に伴う不整合の解消手順
> **最新更新**: 2026-06-12

---

## 1. 発生する問題
スキーマの整理やマイグレーションファイルの統合（Squash）を行った際、ローカル環境の `schema_migrations` テーブルに記録されたチェックサムが、最新の SQL ファイルと一致しなくなり、バックエンドの起動に失敗することがあります。

---

## 2. 解決手順（クリーンリセット）
既存のデータベースボリュームを一旦削除し、空の状態でマイグレーションを再実行させます。

### 2.1 コンテナの停止
```bash
# 全てのサービスを停止
docker compose down
```

### 2.2 ボリュームの削除
```bash
# データベースの永続化ボリュームを削除
docker volume rm ekarte-postgres-data
```
※ `docker-compose.yml` 内でボリュームの `name` が `ekarte-postgres-data` として明示的に指定されているため、プロジェクト名のプレフィックス（`animalekarte_` 等）は付加されません。

### 2.3 再起動と自動構築
```bash
# 再びコンテナを起動
make up
```
起動時に `backend/migrations/` 直下の `*.sql` がファイル名昇順で適用された後、`002_master` → `003_demo` → `004_staging` の CSV シードバンドルが順次ロードされます（seed 002-004 は stub SQL ではなく CSV バンドルのみ）。直下 DDL の本数は固定ではない。検算:

```bash
ls backend/migrations/*.sql
```

2026-07-27統合前の001が適用済みのDBはchecksum mismatchになるため、この再構築が必須です。

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
Seed bundle loaded bundle=003_demo
Seed bundle loaded bundle=004_staging
Seed bundle summary applied=3 skipped=0 total=3
```

ここで `N` は `ls backend/migrations/*.sql` の件数。

`schema_migrations` の行数も固定値ではない。**行数 = 直下 `*.sql` の本数 + seed 3**（キーは各 DDL ファイル名と `seeds/002_master` / `seeds/003_demo` / `seeds/004_staging`）。検算は上記 `ls` の件数に 3 を足したものと、DB の `SELECT COUNT(*) FROM schema_migrations` を照合する。

`/health` エンドポイントが HTTP 200 を返せば、臨床データの入力準備が整いました。

---

## 4. 注意事項
- **データ消失**: この操作により、ローカル環境に入力したテスト用データは全て削除されます。
- **共有環境**: ステージング等の共有環境では、決してこの手順（ボリューム削除）を実行しないでください。現行workflowはDBを再作成しません。再構築が必要な場合は、破壊的操作の明示承認を得て [STG_PLANETSCALE_SEED_RUNBOOK.md](./STG_PLANETSCALE_SEED_RUNBOOK.md) に従います。

---

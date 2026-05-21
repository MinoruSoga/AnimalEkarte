# ローカルDB再構築手順

## 対象

migration 統合後にローカル backend が以下のような checksum mismatch で起動できない場合に使う。

```text
checksum mismatch for 001_init.sql:
  applied=<old checksum>, current=<new checksum>
```

STG は `workflow_dispatch` の `db_reset=true` で空 DB から `001`〜`004` を適用するため、この手順は主にローカル開発環境向け。

## 原因

`001_init.sql`〜`004_seed_staging.sql` は STG DB reset 前提で統合されている。既存ローカル volume には過去の `schema_migrations.checksum` が残るため、同じファイル名の内容が変わると migration runner が停止する。

## 手順

`.env.local` の中身は表示しない。Docker Compose には明示的に env file だけ渡す。

```bash
docker compose --env-file .env.local down
docker volume ls | grep ekarte
docker volume rm animalekarte_ekarte-postgres-data
docker compose --env-file .env.local up -d db
docker compose --env-file .env.local up -d backend
docker compose --env-file .env.local logs backend --tail=120
curl -sf http://localhost:8080/health
```

volume 名が環境で異なる場合は `docker volume ls | grep postgres` で確認してから削除する。

## 期待結果

backend log に以下の流れが出る。

```text
Migration completed file=001_init.sql
Migration completed file=002_seed_master.sql
Migration completed file=003_seed_demo.sql
Migration completed file=004_seed_staging.sql
Migration summary applied=4 skipped=0 total=4
```

`/health` が 200 を返せば復旧完了。

## 注意

- `schema_migrations` の checksum を手で UPDATE するのは、DB を捨てられない環境だけに限定する。
- ローカル開発環境では volume 再作成を優先する。
- STG deploy では `.env.staging` を書き換えず、`gh workflow run backend-deploy.yml --ref staging -f db_reset=true` を使う。

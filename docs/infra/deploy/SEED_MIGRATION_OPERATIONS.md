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

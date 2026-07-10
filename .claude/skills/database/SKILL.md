---
name: database
description: PostgreSQL操作、マイグレーション、最適化
---

# データベース管理スキル

## このスキルを使用するタイミング

- データベースマイグレーションの作成と実行
- クエリパフォーマンスの最適化
- バックアップとリストア操作

## 基本コマンド

### マイグレーション（Raw SQL・新規採番）

> ⚠️ 以下の `psql` 直接実行は CLAUDE.md の自動実行禁止コマンド。ユーザーに手動実行を依頼する。

適用は backend の migration ランナー（`backend/cmd/migrate`、schema_migrations + checksum 管理）が行う
（`/migrations` は db コンテナにマウントされていないため psql での直接適用は不可）。

```bash
# 接続確認
docker compose exec db psql -U "$DB_USER" -d "$DB_NAME" -c "\dt"
```

### データベース操作

> ⚠️ 以下はいずれも CLAUDE.md の自動実行禁止コマンド。ユーザーに手動実行を依頼する。

```bash
# psql 接続
docker compose exec db psql -U "$DB_USER" -d "$DB_NAME"

# バックアップ
docker compose exec db pg_dump -U "$DB_USER" "$DB_NAME" > backup.sql

# リストア
docker compose exec db psql -U "$DB_USER" -d "$DB_NAME" < backup.sql
```

## 命名規則

- マイグレーション: NNN_description.sql（例: 001_init.sql）
- テーブル: snake_case、複数形
- カラム: snake_case

## 重要な注意事項

- 適用済み migration の編集は禁止（checksum mismatch を招く）。最終番号+1 の新規ファイルで追加（migration-seed-safety スキル参照）
- GORM モデル変更後は `make codegen` で `models.ts` を再生成
- 破壊的変更は段階的に実施
- migration ランナーは tx 内実行のため migration 内では `CREATE INDEX CONCURRENTLY` 不可。通常の `CREATE INDEX` を使用（実例: 002_add_checkup_vaccination_indexes.sql）

## 詳細リファレンス

- リファレンス: [reference.md](./reference.md)
- トラブルシューティング: [troubleshooting.md](./troubleshooting.md)

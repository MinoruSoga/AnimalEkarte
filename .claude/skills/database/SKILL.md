---
name: Database Management
description: PostgreSQL操作、マイグレーション、最適化
---

# データベース管理スキル

## このスキルを使用するタイミング

- データベースマイグレーションの作成と実行
- クエリパフォーマンスの最適化
- バックアップとリストア操作

## 基本コマンド

### マイグレーション（Raw SQL、リリース前直接編集運用）

```bash
# マイグレーションファイルを直接編集（リリース前のみ）
# backend/migrations/001_init.sql

# コンテナ内で psql を使って手動適用
docker compose exec db psql -U postgres -d animalekarte -f /migrations/001_init.sql

# 接続確認
docker compose exec db psql -U postgres -d animalekarte -c "\dt"
```

### データベース操作

```bash
# psql 接続
docker compose exec db psql -U postgres -d animalekarte

# バックアップ
docker compose exec db pg_dump -U postgres animalekarte > backup.sql

# リストア
docker compose exec db psql -U postgres -d animalekarte < backup.sql
```

## 命名規則

- マイグレーション: NNN_description.sql（例: 001_init.sql）
- テーブル: snake_case、複数形
- カラム: snake_case

## 重要な注意事項

- スキーマ変更は `backend/migrations/001_init.sql` を直接編集（リリース前運用）
- GORM モデル変更後は `make codegen` で `models.ts` を再生成
- 破壊的変更は段階的に実施
- インデックスは `CREATE INDEX CONCURRENTLY` を使用

## 詳細リファレンス

- スキーマガイド: [schema-guide.md](./schema-guide.md)
- パフォーマンス: [performance.md](./performance.md)

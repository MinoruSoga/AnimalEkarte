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

### マイグレーション

```bash
# マイグレーション作成
npx prisma migrate dev --name [migration_name]

# マイグレーション適用
npx prisma migrate deploy

# スキーマ同期
npx prisma db push
```

### データベース操作

```bash
# Prisma Studio起動
npx prisma studio

# シード実行
npx prisma db seed
```

## 命名規則

- マイグレーション: YYYYMMDD_description
- テーブル: snake_case、複数形
- カラム: snake_case

## 重要な注意事項

- 本番環境では `migrate deploy` を使用
- 破壊的変更は段階的に実施
- インデックスは `CREATE INDEX CONCURRENTLY` を使用

## 詳細リファレンス

- スキーマガイド: [schema-guide.md](./schema-guide.md)
- パフォーマンス: [performance.md](./performance.md)

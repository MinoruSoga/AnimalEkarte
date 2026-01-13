# Database Reference

## Prisma Schema ガイド

### モデル定義

```prisma
model User {
  id        String   @id @default(cuid())
  email     String   @unique
  name      String?
  posts     Post[]
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt

  @@index([email])
  @@map("users")
}

model Post {
  id        String   @id @default(cuid())
  title     String
  content   String?
  published Boolean  @default(false)
  author    User     @relation(fields: [authorId], references: [id])
  authorId  String

  @@index([authorId])
  @@map("posts")
}
```

### リレーション種別

| 種別 | 記法 | 例 |
|------|------|-----|
| 1:1 | `@relation` | User - Profile |
| 1:N | `[]` | User - Posts |
| M:N | 中間テーブル | Post - Tags |

### データ型マッピング

| Prisma | PostgreSQL | TypeScript |
|--------|------------|------------|
| String | TEXT | string |
| Int | INTEGER | number |
| BigInt | BIGINT | bigint |
| Float | DOUBLE PRECISION | number |
| Boolean | BOOLEAN | boolean |
| DateTime | TIMESTAMP | Date |
| Json | JSONB | JsonValue |

## インデックス設計

### 推奨パターン

```prisma
// 単一カラム
@@index([email])

// 複合インデックス（検索順序重要）
@@index([status, createdAt])

// ユニーク制約
@@unique([email, tenantId])
```

### パフォーマンス指針

- WHERE句で頻繁に使用するカラムにインデックス
- 外部キーには自動でインデックス作成
- 複合インデックスは左から順に使用される

## マイグレーション

### コマンド一覧

```bash
# 開発環境
npx prisma migrate dev --name <name>

# 本番環境
npx prisma migrate deploy

# リセット（開発のみ）
npx prisma migrate reset

# 状態確認
npx prisma migrate status
```

### 命名規則

```
YYYYMMDD_description
例: 20250115_add_user_role
```

## クエリ最適化

### N+1問題の回避

```typescript
// NG: N+1クエリ
const users = await prisma.user.findMany();
for (const user of users) {
  const posts = await prisma.post.findMany({ where: { authorId: user.id } });
}

// OK: include で解決
const users = await prisma.user.findMany({
  include: { posts: true }
});
```

### 選択的フィールド取得

```typescript
// 必要なフィールドのみ取得
const users = await prisma.user.findMany({
  select: {
    id: true,
    name: true,
    email: true
  }
});
```

## 参照リンク

- [Prisma公式ドキュメント](https://www.prisma.io/docs)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

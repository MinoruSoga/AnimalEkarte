# Database Troubleshooting

## よくある問題と解決策

### 1. 接続エラー

**症状**: `Can't reach database server`

**原因と対処**:
```bash
# PostgreSQL が起動しているか確認
pg_isready -h localhost -p 5432

# Docker の場合
docker ps | grep postgres

# 接続文字列の確認
echo $DATABASE_URL
```

**チェックリスト**:
- [ ] PostgreSQL サービスが起動している
- [ ] ポート番号が正しい（デフォルト: 5432）
- [ ] ユーザー名/パスワードが正しい
- [ ] データベースが存在する

---

### 2. マイグレーション失敗

**症状**: `Migration failed to apply`

**対処手順**:
```bash
# 1. 状態確認
npx prisma migrate status

# 2. 失敗したマイグレーションを確認
ls -la prisma/migrations/

# 3. 開発環境: リセット
npx prisma migrate reset

# 4. 本番環境: 手動修正
psql $DATABASE_URL -c "DELETE FROM _prisma_migrations WHERE migration_name = 'XXXXX';"
```

**予防策**:
- マイグレーション前にバックアップ
- ステージング環境で先にテスト

---

### 3. スキーマ同期エラー

**症状**: `The database schema is not in sync`

**対処**:
```bash
# スキーマとDBの差分確認
npx prisma db pull --force

# 差分を解消
npx prisma migrate dev
```

---

### 4. パフォーマンス問題

**症状**: クエリが遅い

**診断**:
```sql
-- 実行計画の確認
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'test@example.com';

-- インデックス確認
\di+ users
```

**対処**:
```prisma
// インデックス追加
model User {
  email String
  @@index([email])
}
```

---

### 5. ロック待ち

**症状**: `Lock wait timeout exceeded`

**診断**:
```sql
-- ロック状況確認
SELECT * FROM pg_locks WHERE NOT granted;

-- プロセス確認
SELECT * FROM pg_stat_activity WHERE state = 'active';
```

**対処**:
```sql
-- 問題のプロセスを終了
SELECT pg_terminate_backend(pid);
```

---

### 6. ディスク容量不足

**症状**: `No space left on device`

**対処**:
```bash
# 容量確認
df -h

# 不要データ削除
VACUUM FULL;

# ログクリーンアップ
pg_archivecleanup /var/lib/postgresql/pg_wal/ <checkpoint>
```

---

## 緊急時対応

### バックアップからの復元

```bash
# バックアップ作成
pg_dump $DATABASE_URL > backup.sql

# 復元
psql $DATABASE_URL < backup.sql
```

### ロールバック

```bash
# 特定のマイグレーションまで戻す
npx prisma migrate resolve --rolled-back <migration_name>
```

## 連絡先

問題が解決しない場合:
- DBA チームに連絡
- Slack: #database-support

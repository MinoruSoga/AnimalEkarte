# Database Troubleshooting (GORM + PostgreSQL)

> AnimalEkarte は GORM v2 + PostgreSQL 18 + Raw SQL マイグレーション。

## よくある問題と解決策

### 1. 接続エラー

**症状**: `failed to connect to postgres`

**対処**:
```bash
# DB コンテナの起動確認
docker compose ps db

# コンテナ再起動
docker compose restart db

# 接続確認
docker compose exec db psql -U postgres -d animalekarte -c "\conninfo"

# 環境変数確認（backend コンテナ内）
docker compose exec backend env | grep DB
```

**チェックリスト**:
- [ ] `docker compose up -d` で db コンテナが起動している
- [ ] `.env` の DB_HOST / DB_PORT / DB_USER / DB_PASSWORD が正しい
- [ ] データベースが存在する（`\l` で確認）

---

### 2. マイグレーション失敗

**症状**: SQL エラーでスキーマ適用に失敗

**対処手順**:
```bash
# 1. 現在のスキーマ確認
docker compose exec db psql -U postgres -d animalekarte -c "\dt"

# 2. 失敗箇所を特定してSQLを修正
# backend/migrations/001_init.sql を編集

# 3. 開発環境: DB をリセットして再適用
docker compose down -v   # ボリューム削除
docker compose up -d db
docker compose exec db psql -U postgres -d animalekarte -f /migrations/001_init.sql

# 4. GORM モデル変更を codegen に反映
make codegen
```

**予防策**:
- マイグレーション前にバックアップ取得
- ステージング環境で先にテスト

---

### 3. GORM モデルと DB スキーマの不一致

**症状**: `column "xxx" does not exist` / `relation "xxx" does not exist`

**対処**:
```bash
# DB の実際のスキーマを確認
docker compose exec db psql -U postgres -d animalekarte -c "\d table_name"

# backend/migrations/001_init.sql と GORM モデルを突き合わせて修正
# モデル変更後は codegen 実行
make codegen
```

---

### 4. パフォーマンス問題

**症状**: クエリが遅い

**診断**:
```sql
-- 実行計画の確認
EXPLAIN ANALYZE SELECT * FROM patients WHERE owner_id = 1;

-- インデックス確認
\di+ patients

-- スロークエリ確認
SELECT query, mean_exec_time FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;
```

**対処**:
```sql
-- インデックス追加（001_init.sql に追記して再適用）
CREATE INDEX CONCURRENTLY idx_patients_owner_id ON patients(owner_id);
```

---

### 5. ロック待ち

**症状**: クエリがハング / タイムアウト

**診断**:
```sql
-- ロック状況確認
SELECT * FROM pg_locks WHERE NOT granted;

-- アクティブプロセス確認
SELECT pid, state, query, wait_event FROM pg_stat_activity WHERE state = 'active';
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
# Docker ボリューム容量確認
docker system df

# 不要イメージ・ボリュームの削除
docker system prune -f

# PostgreSQL VACUUM
docker compose exec db psql -U postgres -d animalekarte -c "VACUUM FULL;"
```

---

## 緊急時対応

### バックアップと復元

```bash
# バックアップ作成
docker compose exec db pg_dump -U postgres animalekarte > backup_$(date +%Y%m%d).sql

# 復元
docker compose exec -T db psql -U postgres -d animalekarte < backup_20260101.sql
```

### 開発環境の完全リセット

```bash
docker compose down -v    # ボリュームごと削除
docker compose up -d      # 再起動（init スクリプトが自動実行される）
```

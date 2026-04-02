---
name: database-indexing
description: PostgreSQL インデックス戦略・クエリ最適化（EXPLAIN ANALYZE、複合インデックス）
---

# Database Indexing & Query Optimization

PostgreSQL クエリパフォーマンス最適化。

## インデックス戦略

### 1. 単一カラムインデックス

```sql
-- 検索条件で頻繁に使用されるカラム
CREATE INDEX idx_owners_clinic_id ON owners(clinic_id);

-- 確認
EXPLAIN ANALYZE SELECT * FROM owners WHERE clinic_id = 1;
```

### 2. 複合インデックス（マルチテナント対応）

```sql
-- clinic_id + id (クリニック内での高速検索)
CREATE INDEX idx_clinics_owner_id ON owners(clinic_id, id);

-- ✅ WHERE clinic_id = X AND id = Y
-- ❌ WHERE id = Y（clinic_id がなくても使用可能だが、効率が低い）

-- WHERE clinic_id = X AND name LIKE '太%'
CREATE INDEX idx_clinics_name ON owners(clinic_id, name);
```

### 3. 部分インデックス（論理削除対応）

```sql
-- active なレコードのみインデックス
CREATE INDEX idx_owners_active ON owners(clinic_id, id)
WHERE deleted_at IS NULL;

-- EXPLAIN ANALYZE
-- → Index Scan になるか確認
```

### 4. 外部キーインデックス

```sql
-- ❌ 推奨されない（N+1クエリ）
SELECT * FROM owners;
SELECT * FROM pets WHERE owner_id = ?; -- N回実行

-- ✅ Preload or Join
SELECT o.*, p.* FROM owners o
LEFT JOIN pets p ON o.id = p.owner_id
WHERE o.clinic_id = 1;

-- 外部キーカラムには自動的にインデックスが必要
CREATE INDEX idx_pets_owner_id ON pets(owner_id);
```

### 5. 順序付きインデックス (DESC)

```sql
-- 最新ランクから検索する場合
CREATE INDEX idx_created_at_desc ON owners(clinic_id, created_at DESC);

-- ✅ WHERE clinic_id = X ORDER BY created_at DESC LIMIT 10
```

## クエリ最適化

### EXPLAIN ANALYZE

```sql
EXPLAIN ANALYZE SELECT * FROM owners WHERE clinic_id = 1 AND id = 100;

Output:
Index Scan using idx_clinics_owner_id on owners  (cost=0.28..1.23 rows=1)
  Index Cond: (clinic_id = 1 AND id = 100)
  Planning Time: 0.045 ms
  Execution Time: 0.078 ms

✅ 高速（Index Scan）
```

### Seq Scan 検出

```sql
EXPLAIN ANALYZE SELECT * FROM owners WHERE name LIKE '太%';

Output:
Seq Scan on owners  (cost=0.00..123.45 rows=1000)
  Filter: (name LIKE '太%')
  Planning Time: 0.050 ms
  Execution Time: 12.345 ms

❌ 遅い（Seq Scan）
→ インデックス追加: CREATE INDEX idx_owners_name ON owners(name)
```

### N+1 クエリ検出

```sql
-- ❌ N+1: Owner ごとに Pet を取得
SELECT * FROM owners WHERE clinic_id = 1;  -- 1回
-- アプリ層でループ
  SELECT * FROM pets WHERE owner_id = ?;   -- N回 (所有者数分)

-- ✅ Preload (GORM)
db.Preload("Pets").Where("clinic_id = ?", clinicID).Find(&owners)

-- または JOIN
SELECT o.*, p.* FROM owners o
LEFT JOIN pets p ON o.id = p.owner_id
WHERE o.clinic_id = 1;
```

## インデックス設計パターン

### 所有権テーブル（マルチテナント）

```sql
-- owners テーブル
CREATE INDEX idx_owners_clinic_id ON owners(clinic_id, id);
CREATE INDEX idx_owners_clinic_email ON owners(clinic_id, email) WHERE deleted_at IS NULL;
CREATE INDEX idx_owners_created_at ON owners(clinic_id, created_at DESC);

-- pets テーブル
CREATE INDEX idx_pets_owner_id ON pets(owner_id);
CREATE INDEX idx_pets_clinic_owner ON pets(clinic_id, owner_id);

-- vaccinations テーブル
CREATE INDEX idx_vaccinations_pet_id ON vaccinations(pet_id);
CREATE INDEX idx_vaccinations_clinic_date ON vaccinations(clinic_id, recorded_at DESC);
```

### タイムシリーズデータ（予約・記録）

```sql
-- reservations テーブル
CREATE INDEX idx_reservations_clinic_date ON reservations(clinic_id, reservation_date);
CREATE INDEX idx_reservations_status ON reservations(clinic_id, status)
WHERE cancelled_at IS NULL;

-- medical_records テーブル
CREATE INDEX idx_records_pet_date ON medical_records(pet_id, record_date DESC);
```

## 監視・メンテナンス

### インデックス使用率確認

```sql
SELECT schemaname, tablename, indexname, idx_scan
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

-- idx_scan = 0 → 未使用インデックス（削除候補）
```

### インデックスサイズ確認

```sql
SELECT indexname, pg_size_pretty(pg_relation_size(indexrelid))
FROM pg_stat_user_indexes
ORDER BY pg_relation_size(indexrelid) DESC;
```

### VACUUM & ANALYZE

```sql
-- インデックス最適化（定期実行）
VACUUM ANALYZE owners;
VACUUM ANALYZE pets;
```

## チェックリスト

- [ ] WHERE 句で使用されるカラムにインデックス
- [ ] JOIN カラムにインデックス
- [ ] ORDER BY カラムにインデックス
- [ ] マルチテナント: (clinic_id, id) 複合インデックス
- [ ] 論理削除: 部分インデックス (deleted_at IS NULL)
- [ ] EXPLAIN ANALYZE で実行計画確認
- [ ] N+1 クエリ排除 (GORM Preload)
- [ ] 定期 VACUUM ANALYZE 実行

## パフォーマンス目標

```
Query Time:        < 50ms (p95)
Index Scan Rate:   > 90%
Seq Scan Count:    < 5/日
VACUUM Duration:   < 1分
```

## 関連スキル

- `performance-profiling` - クエリパフォーマンス分析
- `go-security` - SQL インジェクション対策

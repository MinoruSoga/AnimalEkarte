---
description: PostgreSQL設計規約（マルチテナント、論理削除、インデックス戦略）
alwaysApply: false
globs: ["backend/migrations/**", "backend/internal/model/**", "backend/internal/repository/**"]
---

# Database Design Rules

PostgreSQL 18 マルチテナント設計規約。

## 核心ルール

### 1. テーブル設計パターン

```sql
-- ✅ 標準テーブル
CREATE TABLE owners (
  id BIGSERIAL PRIMARY KEY,
  clinic_id BIGINT NOT NULL,  -- マルチテナント必須
  name VARCHAR(100) NOT NULL,
  email VARCHAR(100) NOT NULL,
  phone VARCHAR(20),
  address TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,  -- 論理削除
  CONSTRAINT fk_owners_clinic FOREIGN KEY (clinic_id) REFERENCES clinics(id),
  CONSTRAINT uk_owners_clinic_email UNIQUE (clinic_id, email) WHERE deleted_at IS NULL
);

-- ✅ 論理削除インデックス（active レコードのみ）
CREATE INDEX idx_owners_active ON owners(clinic_id, id) WHERE deleted_at IS NULL;
CREATE INDEX idx_owners_email ON owners(clinic_id, email) WHERE deleted_at IS NULL;

-- ✅ タイムシリーズインデックス（新しい順）
CREATE INDEX idx_created_at_desc ON owners(clinic_id, created_at DESC);
```

### 2. マルチテナント設計（clinic_id 必須）

```sql
-- ❌ 危険: clinic_id なしのクエリ（データリーク可能性）
SELECT * FROM owners WHERE id = 1;

-- ✅ 安全: 常に clinic_id を条件に含める
SELECT * FROM owners WHERE clinic_id = $1 AND id = $2;

-- ✅ インデックス設計（clinic_id を必ず先に）
CREATE INDEX idx_owners_clinic_id ON owners(clinic_id, id);
```

### 3. 複合インデックス戦略

```sql
-- WHERE clinic_id = X AND id = Y
CREATE INDEX idx_clinic_owner ON owners(clinic_id, id);

-- WHERE clinic_id = X AND status = Y
CREATE INDEX idx_clinic_status ON vaccinations(clinic_id, status);

-- WHERE clinic_id = X ORDER BY created_at DESC LIMIT 10
CREATE INDEX idx_clinic_created_desc ON owners(clinic_id, created_at DESC);

-- WHERE clinic_id = X AND name LIKE '%X%'
-- ⚠️ LIKE '%X' では B-tree インデックス効果なし → GIN インデックス検討
CREATE INDEX idx_owners_name_gin ON owners USING GIN(to_tsvector('japanese', name));
```

### 4. 論理削除対応

```sql
-- ✅ 部分インデックス（deleted_at IS NULL）
CREATE INDEX idx_owners_active ON owners(clinic_id, id) WHERE deleted_at IS NULL;

-- ✅ UNIQUE 制約も論理削除対応
CREATE UNIQUE INDEX uk_owners_email
ON owners(clinic_id, email) WHERE deleted_at IS NULL;

-- ✅ App層でのフィルタ
-- repository/owner_repository.go
func (r *OwnerRepository) GetByID(ctx context.Context, id uint64) (*Owner, error) {
  var owner Owner
  return &owner, r.db.WithContext(ctx)
    .Where("id = ? AND deleted_at IS NULL", id)  // 必須
    .First(&owner)
    .Error
}

// ✅ Global Scope (GORM)
func (Owner) TableName() string {
  return "owners"
}

// ← Global Scope で deleted_at を自動フィルタ
db.Scopes(SoftDeleteScope).Where(...).Find(&owners)

func SoftDeleteScope(db *gorm.DB) *gorm.DB {
  return db.Where("deleted_at IS NULL")
}
```

### 5. 外部キー・リレーション設計

```sql
-- ✅ 外部キーは必須、DELETE CASCADE は要検討
CREATE TABLE pets (
  id BIGSERIAL PRIMARY KEY,
  clinic_id BIGINT NOT NULL,
  owner_id BIGINT NOT NULL,
  name VARCHAR(100) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  CONSTRAINT fk_pets_owner FOREIGN KEY (owner_id) REFERENCES owners(id)
    ON DELETE CASCADE,  -- owner 削除時に pet も削除
  CONSTRAINT fk_pets_clinic FOREIGN KEY (clinic_id) REFERENCES clinics(id)
);

-- ✅ 外部キーカラムにはインデックス必須
CREATE INDEX idx_pets_owner ON pets(owner_id);
CREATE INDEX idx_pets_clinic_owner ON pets(clinic_id, owner_id);
```

### 6. N+1 クエリ対策

```go
// ❌ N+1: Owner 取得 → 各 Owner の Pet をループで取得
owners, _ := r.GetOwners(ctx, clinicID)
for _, owner := range owners {
  pets, _ := r.GetPetsByOwner(ctx, owner.ID)  // N回実行
  owner.Pets = pets
}

// ✅ GORM Preload（単一クエリで関連データ取得）
var owners []Owner
db.WithContext(ctx)
  .Preload("Pets")
  .Where("clinic_id = ?", clinicID)
  .Find(&owners)

// ✅ JOIN（複雑フィルタが必要な場合）
var owners []Owner
db.WithContext(ctx)
  .Joins("LEFT JOIN pets ON owners.id = pets.owner_id")
  .Where("owners.clinic_id = ?", clinicID)
  .Distinct("owners.*")
  .Find(&owners)
```

### 7. Enum/Status カラム設計

```sql
-- ✅ PostgreSQL ENUM型（推奨）
CREATE TYPE billing_status AS ENUM ('waiting', 'paid', 'failed');

CREATE TABLE invoices (
  id BIGSERIAL PRIMARY KEY,
  clinic_id BIGINT NOT NULL,
  status billing_status DEFAULT 'waiting',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Go model
type BillingStatus string
const (
  BillingStatusWaiting BillingStatus = "waiting"
  BillingStatusPaid    BillingStatus = "paid"
  BillingStatusFailed  BillingStatus = "failed"
)

type Invoice struct {
  ID        uint64        `gorm:"primaryKey"`
  ClinicID  uint64
  Status    BillingStatus `gorm:"type:billing_status"`
  CreatedAt time.Time
}
```

### 8. スキーママイグレーション

```sql
-- backend/migrations/001_init.sql
-- リリース前は直接編集 OK（incremental migration 不要）
-- リリース後は incremental migration 推奨

-- 002_add_field.sql
ALTER TABLE owners ADD COLUMN middle_name VARCHAR(100);

-- 003_create_index.sql
CREATE INDEX idx_owners_clinic ON owners(clinic_id);
```

## チェックリスト

- [ ] 全テーブルに `clinic_id` (マルチテナント)
- [ ] 全テーブルに `created_at`, `updated_at`, `deleted_at`
- [ ] WHERE 句は `clinic_id` から開始
- [ ] インデックス: `clinic_id` を先に（複合インデックス）
- [ ] 論理削除: 部分インデックス `WHERE deleted_at IS NULL`
- [ ] 外部キー: ON DELETE CASCADE 検討
- [ ] 外部キーカラム: インデックス必須
- [ ] UNIQUE 制約: 論理削除対応（部分インデックス）
- [ ] N+1 クエリ: Preload または JOIN で対策
- [ ] EXPLAIN ANALYZE でクエリ確認

## パフォーマンス目標

```
Query Time:      < 50ms (p95)
Index Scan Rate: > 90%
Seq Scan Count:  < 5/日
```

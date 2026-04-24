---
description: PostgreSQL design standards (multi-tenant, soft delete, indexing strategy)
alwaysApply: false
globs: ["backend/migrations/**", "backend/internal/model/**", "backend/internal/repository/**"]
---

# Database Design Rules

PostgreSQL 18 multi-tenant design standards.

## Core Rules

### 1. Table Design Pattern

```sql
-- ✅ Standard table
CREATE TABLE owners (
  id BIGSERIAL PRIMARY KEY,
  clinic_id BIGINT NOT NULL,  -- Multi-tenant required
  name VARCHAR(100) NOT NULL,
  email VARCHAR(100) NOT NULL,
  phone VARCHAR(20),
  address TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,  -- Soft delete
  CONSTRAINT fk_owners_clinic FOREIGN KEY (clinic_id) REFERENCES clinics(id),
  CONSTRAINT uk_owners_clinic_email UNIQUE (clinic_id, email) WHERE deleted_at IS NULL
);

-- ✅ Soft delete index (active records only)
CREATE INDEX idx_owners_active ON owners(clinic_id, id) WHERE deleted_at IS NULL;
CREATE INDEX idx_owners_email ON owners(clinic_id, email) WHERE deleted_at IS NULL;

-- ✅ Time-series index (recent first)
CREATE INDEX idx_created_at_desc ON owners(clinic_id, created_at DESC);
```

### 2. Multi-Tenant Design (clinic_id required)

```sql
-- ❌ Dangerous: clinic_id omitted (data leak possible)
SELECT * FROM owners WHERE id = 1;

-- ✅ Safe: Always include clinic_id in WHERE
SELECT * FROM owners WHERE clinic_id = $1 AND id = $2;

-- ✅ Index design (clinic_id first)
CREATE INDEX idx_owners_clinic_id ON owners(clinic_id, id);
```

### 3. Composite Index Strategy

```sql
-- WHERE clinic_id = X AND id = Y
CREATE INDEX idx_clinic_owner ON owners(clinic_id, id);

-- WHERE clinic_id = X AND status = Y
CREATE INDEX idx_clinic_status ON vaccinations(clinic_id, status);

-- WHERE clinic_id = X ORDER BY created_at DESC LIMIT 10
CREATE INDEX idx_clinic_created_desc ON owners(clinic_id, created_at DESC);

-- WHERE clinic_id = X AND name LIKE '%X%'
-- ⚠️ LIKE '%X' won't use B-tree index → consider GIN
CREATE INDEX idx_owners_name_gin ON owners USING GIN(to_tsvector('japanese', name));
```

### 4. Soft Delete Support

```sql
-- ✅ Partial index (deleted_at IS NULL)
CREATE INDEX idx_owners_active ON owners(clinic_id, id) WHERE deleted_at IS NULL;

-- ✅ UNIQUE constraint with soft delete
CREATE UNIQUE INDEX uk_owners_email
ON owners(clinic_id, email) WHERE deleted_at IS NULL;

-- ✅ Application filter
-- repository/owner_repository.go
func (r *OwnerRepository) GetByID(ctx context.Context, id uint64) (*Owner, error) {
  var owner Owner
  return &owner, r.db.WithContext(ctx)
    .Where("id = ? AND deleted_at IS NULL", id)  -- Required
    .First(&owner)
    .Error
}

// ✅ Global Scope (GORM)
func (Owner) TableName() string {
  return "owners"
}

// Auto-filter deleted_at with Global Scope
db.Scopes(SoftDeleteScope).Where(...).Find(&owners)

func SoftDeleteScope(db *gorm.DB) *gorm.DB {
  return db.Where("deleted_at IS NULL")
}
```

### 5. Foreign Key / Relation Design

```sql
-- ✅ FK required; DELETE CASCADE on deliberation
CREATE TABLE pets (
  id BIGSERIAL PRIMARY KEY,
  clinic_id BIGINT NOT NULL,
  owner_id BIGINT NOT NULL,
  name VARCHAR(100) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  CONSTRAINT fk_pets_owner FOREIGN KEY (owner_id) REFERENCES owners(id)
    ON DELETE CASCADE,  -- Delete pet when owner deleted
  CONSTRAINT fk_pets_clinic FOREIGN KEY (clinic_id) REFERENCES clinics(id)
);

-- ✅ FK column must have index
CREATE INDEX idx_pets_owner ON pets(owner_id);
CREATE INDEX idx_pets_clinic_owner ON pets(clinic_id, owner_id);
```

### 6. N+1 Query Prevention

```go
// ❌ N+1: Fetch Owner → Loop fetching each Owner's Pets
owners, _ := r.GetOwners(ctx, clinicID)
for _, owner := range owners {
  pets, _ := r.GetPetsByOwner(ctx, owner.ID)  // N queries
  owner.Pets = pets
}

// ✅ GORM Preload (single query for related data)
var owners []Owner
db.WithContext(ctx)
  .Preload("Pets")
  .Where("clinic_id = ?", clinicID)
  .Find(&owners)

// ✅ JOIN (complex filters)
var owners []Owner
db.WithContext(ctx)
  .Joins("LEFT JOIN pets ON owners.id = pets.owner_id")
  .Where("owners.clinic_id = ?", clinicID)
  .Distinct("owners.*")
  .Find(&owners)
```

### 7. Enum/Status Column Design

```sql
-- ✅ PostgreSQL ENUM type (recommended)
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

### 8. Schema Migration

```sql
-- backend/migrations/001_init.sql
-- Pre-release: direct edit OK (incremental not needed)
-- Post-release: incremental migration recommended

-- 002_add_field.sql
ALTER TABLE owners ADD COLUMN middle_name VARCHAR(100);

-- 003_create_index.sql
CREATE INDEX idx_owners_clinic ON owners(clinic_id);
```

## Checklist

- [ ] All tables have `clinic_id` (multi-tenant)
- [ ] All tables have `created_at`, `updated_at`, `deleted_at`
- [ ] WHERE always starts with `clinic_id`
- [ ] Indexes: `clinic_id` first (composite)
- [ ] Soft delete: partial indexes `WHERE deleted_at IS NULL`
- [ ] FK: ON DELETE CASCADE deliberated
- [ ] FK column: index required
- [ ] UNIQUE constraint: soft delete aware (partial index)
- [ ] N+1 prevention: Preload or JOIN used
- [ ] EXPLAIN ANALYZE confirms query plan

## Performance Targets

```
Query Time:      < 50ms (p95)
Index Scan Rate: > 90%
Seq Scan Count:  < 5/day
```

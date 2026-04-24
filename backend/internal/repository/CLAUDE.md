# Repository Layer — P2 / P3 / P4 / P9 / P16

## P2: deleted_at IS NULL in CountUsage (MANDATORY)

```go
// ✅
err := r.db.WithContext(ctx).Model(&model.Vaccination{}).
    Where("vaccine_id = ? AND deleted_at IS NULL", vaccineID).Count(&count).Error

// ❌
Where("vaccine_id = ?", vaccineID)  // deleted_at IS NULL なし
```

## P3: Preload with deleted_at IS NULL (MANDATORY)

ソフトデリート対象エンティティのすべての `Preload` に条件を付ける。

```go
// ✅
db.Preload("ReservationType", "deleted_at IS NULL").
   Preload("Doctor", "deleted_at IS NULL").Find(&reservations)

// ❌
db.Preload("ReservationType").Preload("Doctor").Find(&reservations)
```

**注意**: `Preload("Doctor")` は `Staff` モデルへのエイリアス — 対象。

対象 42 エンティティ（抜粋）: `Account`, `Billing`, `Cage`, `Checkup`, `Consultation`,
`Examination`, `Hospitalization`, `MedicalRecord`, `Medicine`, `Owner`, `Pet`,
`Reservation`, `Staff`, `Treatment`, `Vaccination`, `Vaccine` など。
完全なリストは `.claude/refs/gin-architecture-compliance.md` P3 を参照。

## P4: clinicScope on Update/Upsert (MANDATORY — 最重要)

UPDATE/UPSERT に `Scopes(clinicScope(clinicID))` を必ず付ける。

```go
// ✅
err := r.db.WithContext(ctx).Model(&model.Vaccine{}).
    Scopes(clinicScope(clinicID)).
    Where("id = ?", id).Updates(fields).Error

// ❌ クロスクリニックデータ更新リスク
r.db.Model(&model.Vaccine{}).Where("id = ?", id).Updates(fields)
```

**例外（clinicScope 不要）**: `clinic_repository.go`, `company_repository.go`,
`account_repository.go`, `password_reset_token_repository.go`, `audit_repository.go`

## P9: apperrors.FromGORM on GORM errors (MANDATORY)

```go
// ✅
if err := r.db.WithContext(ctx).First(&vaccine, id).Error; err != nil {
    return nil, apperrors.FromGORM(err, "vaccine", fmt.Sprintf("%d", id))
}

// ❌
if err := r.db.First(&vaccine, id).Error; err != nil {
    return nil, err
}
```

## P16: Method naming conventions (MANDATORY)

```
FindAll / FindByClinicID  ← 一覧（GetAll, List, Fetch は違反）
FindByID                  ← 単件（GetByID, Get, Find は違反）
Create / Update / Delete  ← 標準
CountBy{Xxx}              ← カウント
CountUsageBy{Xxx}         ← 使用数カウント
```

```go
// ✅
type VaccineRepository interface {
    FindAll(ctx context.Context, clinicID uint64) ([]*model.Vaccine, error)
    FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error)
    Create(ctx context.Context, vaccine *model.Vaccine) error
    Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccine, error)
    Delete(ctx context.Context, clinicID, id uint64) error
}

// ❌
GetAll(...)  GetByID(...)  List(...)  Fetch(...)
```

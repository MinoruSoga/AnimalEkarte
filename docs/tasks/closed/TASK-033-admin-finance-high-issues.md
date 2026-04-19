# TASK-033: 管理・会計系 HIGH 問題 5件

## 優先度

HIGH

---

## 問題 1: clinic_handler の CreateClinic が *model.Clinic を直接構築

### ファイル
`backend/internal/handler/clinic_handler.go:171-182`

### 問題
`CreateClinic` ハンドラが `&model.Clinic{...}` を直接組み立てて service に渡している。Input DTO 未使用。

```go
// 現状（規約違反）
clinic := &model.Clinic{
    Name:               req.Name,
    PostalCode:         req.PostalCode,
    // ...
}
result, err := h.svc.Clinic.CreateClinic(c.Request.Context(), clinic)
```

### 修正案
```go
// service/clinic_service.go — Input DTO 追加
type CreateClinicInput struct {
    Name               string
    PostalCode         *string
    Address            *string
    PhoneNumber        *string
    // ...
}

// service シグネチャ変更
Create(ctx context.Context, input CreateClinicInput) (*model.Clinic, error)

// handler 側: DTO のみ構築
input := service.CreateClinicInput{
    Name:    req.Name,
    // ...
}
result, err := h.svc.Clinic.CreateClinic(c.Request.Context(), input)
```

---

## 問題 2: inventory_handler の CreateInventory が *model.InventoryItem を直接構築

### ファイル
`backend/internal/handler/inventory_handler.go:89-103`

### 問題
```go
item := &model.InventoryItem{
    ClinicID:      clinicID,
    Name:          input.Name,
    Category:      model.InventoryCategory(input.Category),
    // ...
}
h.svc.Inventory.Create(c.Request.Context(), clinicID, item)
```
`model.InventoryCategory(input.Category)` の型変換もハンドラ内で行われており、service 層の責務が侵食されている。

### 修正案
```go
// service/inventory_service.go — Input DTO 追加
type CreateInventoryInput struct {
    Name          string
    Category      string
    Quantity      float64
    Unit          string
    MinStockLevel float64
    Location      *string
    ExpiryDate    *time.Time
    Supplier      *string
    LastRestocked *time.Time
    Status        *string
}

// service 内で model.InventoryCategory 変換・ClinicID 設定を行う
```

---

## 問題 3: inventory_repository の CountUsageByInventoryID で vaccines / medicines が deleted_at IS NULL 欠落

### ファイル
`backend/internal/repository/inventory_repository.go:132-143`

### 問題
```go
err := r.db.WithContext(ctx).
    Raw(`SELECT (
        SELECT COUNT(*) FROM treatments WHERE inventory_id = ? AND deleted_at IS NULL
    ) + (
        SELECT COUNT(*) FROM vaccines WHERE inventory_id = ?          -- deleted_at IS NULL なし
    ) + (
        SELECT COUNT(*) FROM medicines WHERE inventory_id = ?         -- deleted_at IS NULL なし
    ) AS total`, inventoryID, inventoryID, inventoryID).
    Scan(&count).Error
```
`treatments` には `deleted_at IS NULL` があるが `vaccines` と `medicines` にない。論理削除済みのレコードでも参照カウントに含まれ、削除可能な在庫が削除不可と誤判定される。

### 修正案
```sql
SELECT COUNT(*) FROM vaccines  WHERE inventory_id = ? AND deleted_at IS NULL
SELECT COUNT(*) FROM medicines WHERE inventory_id = ? AND deleted_at IS NULL
```

---

## 問題 4: clinic_repository の FindByStaffID が staff_clinic_assignments.deleted_at IS NULL 欠落

### ファイル
`backend/internal/repository/clinic_repository.go:42-53`

### 問題
```go
err := r.db.WithContext(ctx).
    Joins("INNER JOIN staff_clinic_assignments ON staff_clinic_assignments.clinic_id = clinics.id").
    Where("staff_clinic_assignments.staff_id = ?", staffID).
    // ...
```
`staff_clinic_assignments` の論理削除（`deleted_at IS NULL`）が JOIN 条件に含まれていない。退職済みスタッフの割当が残っている場合、削除済みのアサインメントが有効なものとして扱われる。

### 修正案
```go
Joins("INNER JOIN staff_clinic_assignments ON staff_clinic_assignments.clinic_id = clinics.id"+
    " AND staff_clinic_assignments.deleted_at IS NULL").
Where("staff_clinic_assignments.staff_id = ?", staffID).
```

---

## 問題 5: billing_confirmation_service の Return メソッドで状態チェックが不完全

### ファイル
`backend/internal/service/billing_confirmation_service.go:95-122`

### 問題
`Return` は `GetOrCreate` を呼ぶが、`ConfirmationStatusPending` 以外でも差し戻しを実行できてしまう。`Confirm` と対称的に `ConfirmationStatusConfirmed` のみ差し戻し可能とする状態チェックが必要。

```go
// Confirm には状態チェックあり:
if review.Status == model.ConfirmationStatusConfirmed {
    return nil, apperrors.WrapInvalidInput("billing review is already confirmed")
}

// Return には状態チェックなし（問題）
// pending 以外（例: 既に returned）の場合でも更新されてしまう
```

### 修正案
```go
func (s *billingConfirmationService) Return(...) (*model.BillingConfirmation, error) {
    review, err := s.GetOrCreate(ctx, clinicID, medicalRecordID)
    if err != nil { ... }

    // 確認済みのみ差し戻し可能
    if review.Status != model.ConfirmationStatusConfirmed {
        return nil, apperrors.WrapInvalidInput("差し戻しは確認済みの場合のみ可能です")
    }
    // ...
}
```

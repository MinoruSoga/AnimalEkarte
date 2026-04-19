# BUG-427: CountUsage 系メソッドで子テーブルの soft delete 未対応により削除済みレコードがカウントされる

## 概要

BUG-414（FindAll/FindByID の deleted_at フィルタ問題）は GORM soft delete が全マスタモデルに
設定済みのため **実害なし**と確認された（GORM が自動的に `deleted_at IS NULL` を付加）。

しかし CountUsage 系メソッドが参照する一部の**子テーブルモデル**に `DeletedAt` フィールドが存在せず、
GORM の auto soft delete が適用されないため、削除済みレコードもカウントされてしまう。
結果として「使用中」と誤判定され、実際には未使用のマスタが削除できないバグが発生する。

## 影響を受けるモデルと CountUsage メソッド

### 1. `EstimateItem` — DeletedAt フィールドなし

**モデルファイル**: `backend/internal/model/estimate.go:51-77`
```go
type EstimateItem struct {
    ID                  uint64 `gorm:"primaryKey"`
    // ...
    // DeletedAt フィールドなし ← GORM auto soft delete が効かない
    UpdatedAt           time.Time `gorm:"autoUpdateTime"`
}
```

**影響を受けるメソッド**: `merchandise_item_repository.go:87-108`（CountUsageByMerchandiseItemID）
```go
// EstimateItem のカウント部分（行 100-108）
if err := r.db.WithContext(ctx).
    Model(&model.EstimateItem{}).
    Joins("JOIN estimates ON estimates.id = estimate_items.estimate_id AND estimates.clinic_id = ?", clinicID).
    Where("estimate_items.merchandise_item_id = ?", merchandiseItemID).
    // ↑ deleted_at フィルタなし（EstimateItem に DeletedAt がないため追加不可）
    Count(&estimateCount).Error; err != nil { ... }
```

**実害**: 見積もりから削除された物販品目でも「使用中」と判定され、マスタ削除できない。

---

### 2. `ClinicalPlan` — DeletedAt フィールドなし

**モデルファイル**: `backend/internal/model/clinical_plan.go:8-29`
```go
type ClinicalPlan struct {
    ID              uint64 `gorm:"primaryKey"`
    // ...
    // DeletedAt フィールドなし
    UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}
```

**影響を受けるメソッド**: `diagnosis_repository.go:226-236`（CountClinicalPlansByDiagnosisNameID）
```go
if err := r.db.WithContext(ctx).
    Model(&model.ClinicalPlan{}).
    Joins("JOIN medical_records ON ... AND medical_records.deleted_at IS NULL", clinicID).
    Where("clinical_plans.diagnosis_name_id = ? OR clinical_plans.diagnosis_2_name_id = ?", ...).
    // ↑ medical_records は deleted_at フィルタあり、clinical_plans はなし
    Count(&count).Error; err != nil { ... }
```

**実害**: 削除済みカルテの診断名参照でも「使用中」と判定される可能性。

---

### 3. `StaffClinicAssignment` — DeletedAt フィールドなし

**モデルファイル**: `backend/internal/model/account.go:25-38`
```go
type StaffClinicAssignment struct {
    StaffID   uint64
    ClinicID  uint64
    // DeletedAt フィールドなし
}
```

**影響を受けるメソッド**: `occupation_repository.go:91-101`（CountStaffsByOccupationID）
```go
if err := r.db.WithContext(ctx).
    Model(&model.Staff{}).
    Joins("JOIN staff_clinic_assignments ON staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ?", clinicID).
    Where("staffs.occupation_id = ?", occupationID).
    // ↑ Staff は GORM auto soft delete あり、StaffClinicAssignment はなし
    Count(&count).Error; err != nil { ... }
```

**実害**: クリニックからスタッフ割り当てを解除しても、occupation の「使用中」カウントに含まれる可能性。

## 修正方針

### Option A（推奨）: 各モデルに `DeletedAt gorm.DeletedAt` を追加

```go
// estimate.go
type EstimateItem struct {
    ID        uint64         `gorm:"primaryKey"`
    // ...
    DeletedAt gorm.DeletedAt `gorm:"index"`  // 追加
    UpdatedAt time.Time      `gorm:"autoUpdateTime"`
}

// clinical_plan.go
type ClinicalPlan struct {
    ID        uint64         `gorm:"primaryKey"`
    // ...
    DeletedAt gorm.DeletedAt `gorm:"index"`  // 追加
    UpdatedAt time.Time      `gorm:"autoUpdateTime"`
}
```

**注意**: マイグレーションで `deleted_at` カラムを追加する必要がある。

### Option B: CountUsage メソッドにコメントで制限事項を明示し、WHERE 句を調整

EstimateItem の場合、物理削除が採用されているなら現状のカウントが正しい可能性もある。
設計意図を確認した上で対応方針を決定すること。

## 影響ファイル

- `backend/internal/model/estimate.go` — EstimateItem 構造体
- `backend/internal/model/clinical_plan.go` — ClinicalPlan 構造体
- `backend/internal/model/account.go` — StaffClinicAssignment 構造体
- `backend/internal/repository/merchandise_item_repository.go` — 行 100-108
- `backend/internal/repository/diagnosis_repository.go` — 行 226-236
- `backend/internal/repository/occupation_repository.go` — 行 91-101
- `backend/migrations/001_init.sql` — deleted_at カラム追加（Option A の場合）

## 優先度

**Medium** — 削除済みレコードが「使用中」と誤判定されるユーザー影響あり。ただし設計意図確認が先。

## BUG-414 との関係

BUG-414（FindAll/FindByID の deleted_at フィルタ欠落）は GORM soft delete が全マスタモデルに設定済みのため **クローズ推奨**。
本チケット（BUG-427）が BUG-414 の実質的な問題として継続すべき課題。

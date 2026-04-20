# CODE-QUALITY-211: CountUsage 系メソッドの clinicScope 非使用（手動 WHERE 混在）

## 概要

4つの Repository の `CountUsageByXxxID` メソッドで `clinicScope(clinicID)` を使わず
手動 WHERE で `clinic_id = ?` を混在指定している。
他の全 CountUsage 実装は `Scopes(clinicScope(clinicID))` で統一されており、
これらだけ規約から逸脱している。

## 優先度

HIGH

## 影響ファイル

| ファイル | メソッド | 行番号 |
|---------|---------|--------|
| `backend/internal/repository/exam_type_repository.go` | CountUsageByExamTypeID | ~L97 |
| `backend/internal/repository/vaccine_repository.go` | CountUsageByVaccineID | ~L97 |
| `backend/internal/repository/insurance_repository.go` | CountPetsByInsuranceID | ~L90 |
| `backend/internal/repository/diagnosis_repository.go` | CountNamesByCategoryID | ~L107 |

---

## 正しいパターン（参照: occupation_repository.go）

```go
func (r *occupationRepository) CountStaffsByOccupationID(...) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&model.StaffClinicAssignment{}).
        Scopes(clinicScope(clinicID)).              // ← clinicScope 使用
        Where("occupation_id = ?", occupationID).   // ← FK 条件のみ
        Count(&count).Error
    ...
}
```

---

## 問題詳細

### 1. exam_type_repository.go:97

```go
// 現状（手動 WHERE）
Where("exam_type_id = ? AND clinic_id = ?", examTypeID, clinicID)

// 修正後
Scopes(clinicScope(clinicID)).
Where("exam_type_id = ?", examTypeID)
```

### 2. vaccine_repository.go:97

```go
// 現状（手動 WHERE）
Where("vaccine_id = ? AND clinic_id = ?", vaccineID, clinicID)

// 修正後
Scopes(clinicScope(clinicID)).
Where("vaccine_id = ?", vaccineID)
```

### 3. insurance_repository.go:90

```go
// 現状（手動 WHERE）
Where("insurance_id = ? AND clinic_id = ?", insuranceID, clinicID)

// 修正後
Scopes(clinicScope(clinicID)).
Where("insurance_id = ?", insuranceID)
```

`pets` は `gorm.DeletedAt` を持つため、`Model(&model.Pet{})` を起点にすることで
GORM が自動的に soft delete フィルタも適用する。clinicScope への変更でこちらも解消される。

### 4. diagnosis_repository.go:107

```go
// 現状（手動 WHERE）
Where("diagnosis_type_id = ? AND clinic_id = ?", categoryID, clinicID)

// 修正後
Scopes(clinicScope(clinicID)).
Where("diagnosis_type_id = ?", categoryID)
```

---

## 規約参照

- `.claude/rules/database-design.md`: WHERE 句は `clinic_id` から開始（clinicScope 使用）
- `backend/internal/repository/repository.go` で `clinicScope` が定義されており、全リポジトリで使用するのが規約

## テスト

修正後も `CountUsageByXxxID` が正しいカウントを返すことを確認（clinicScope の機能的同等性検証）。

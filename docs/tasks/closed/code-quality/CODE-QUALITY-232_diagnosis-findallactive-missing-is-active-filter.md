# CODE-QUALITY-232: diagnosis_repository FindAllActive の is_active フィルタ欠落

## 概要

`backend/internal/repository/diagnosis_repository.go` の `FindAllActive` メソッドが、
メソッド名に "Active" と付いているにもかかわらず `is_active = true` のフィルタを適用していない。
`model.DiagnosisName` には `IsActive bool` フィールドが存在する（model/diagnosis.go:30）。

---

## 該当コード

**ファイル:** `backend/internal/repository/diagnosis_repository.go:177-189`

```go
// FindAllActive はページネーションなしで全件取得する（#418: ListNames 用）。
// typeID が非 nil の場合は該当カテゴリのみ、nil の場合はクリニック全件を返す。
func (r *diagnosisNameRepository) FindAllActive(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error) {
    q := r.db.WithContext(ctx).Model(&model.DiagnosisName{}).Scopes(clinicScope(clinicID))
    if typeID != nil {
        q = q.Where("diagnosis_type_id = ?", *typeID)
    }
    names := make([]model.DiagnosisName, 0)
    if err := q.Order("sort_order ASC, name ASC").Find(&names).Error; err != nil {
        return nil, apperrors.FromGORM(err, "diagnosis_name", "")
    }
    return names, nil
}
```

---

## 問題

`model.DiagnosisName` は `IsActive bool` フィールドを持つ（`model/diagnosis.go:30`）が、
`FindAllActive` のクエリに `WHERE is_active = true` が含まれていない。

結果として:
- `is_active = false` の無効化された診断名がフロントエンドの選択肢に表示される
- 診察記録で非アクティブな診断名が選択可能になってしまう

---

## モデル定義（参照）

```go
// model/diagnosis.go:27-35
type DiagnosisName struct {
    ID              uint64         `gorm:"primaryKey"   json:"id"`
    ClinicID        uint64         `gorm:"not null"     json:"-"`
    DiagnosisTypeID uint64         `gorm:"not null"     json:"diagnosis_type_id"`
    Name            string         `gorm:"not null"     json:"name"`
    SortOrder       int            `gorm:"default:0"    json:"sort_order"`
    IsActive        bool           `gorm:"default:true" json:"is_active"`
    DeletedAt       gorm.DeletedAt `                    json:"-"`
}
```

---

## 修正案

```go
func (r *diagnosisNameRepository) FindAllActive(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error) {
    q := r.db.WithContext(ctx).
        Model(&model.DiagnosisName{}).
        Scopes(clinicScope(clinicID)).
        Where("is_active = ?", true)  // ← 追加
    if typeID != nil {
        q = q.Where("diagnosis_type_id = ?", *typeID)
    }
    names := make([]model.DiagnosisName, 0)
    if err := q.Order("sort_order ASC, name ASC").Find(&names).Error; err != nil {
        return nil, apperrors.FromGORM(err, "diagnosis_name", "")
    }
    return names, nil
}
```

---

## 優先度

**HIGH** — 非アクティブな診断名が業務画面で選択可能になる機能バグ。
診断名の is_active 管理を行っているクリニックでは直接影響が出る。

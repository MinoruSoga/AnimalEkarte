# BUG-415: diagnosis_repository の FindAll でクエリビルダーが重複定義されている

## 概要

`diagnosis_repository.go` の `FindAll` メソッドで、同一クエリビルダーが2回作成されている。
`medicine_repository.go` では `buildBase()` 関数で解決済みのパターンが適用されていない。
コードの一貫性とDRY原則に違反する。

## 問題箇所

```go
// diagnosis_repository.go:32-46
func (r *diagnosisRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.DiagnosisType, error) {
    var total int64
    var categories []model.DiagnosisType

    base := r.db.WithContext(ctx).Model(&model.DiagnosisType{}).Scopes(clinicScope(clinicID))
    if err := base.Count(&total).Error; err != nil {
        return nil, apperrors.FromGORM(err, "diagnosis_type", fmt.Sprintf("%d", clinicID))
    }

    // ← ここで base を再利用せず同じビルダーを再度作成
    if err := r.db.WithContext(ctx).Model(&model.DiagnosisType{}).Scopes(clinicScope(clinicID)).
        Preload("Names").
        Order("sort_order ASC").
        Find(&categories).Error; err != nil {
        return nil, apperrors.FromGORM(err, "diagnosis_type", fmt.Sprintf("%d", clinicID))
    }
    // ...
}
```

## 期待する実装

`medicine_repository.go` で実装済みの `buildBase()` パターンを適用する。

```go
// medicine_repository.go:31-48（参考パターン）
buildBase := func() *gorm.DB {
    return r.db.WithContext(ctx).Model(&model.Medicine{}).Scopes(clinicScope(clinicID))
}
if err := buildBase().Count(&total).Error; err != nil { ... }
if err := buildBase().Offset(offset).Limit(limit).Order(...).Find(&medicines).Error; err != nil { ... }
```

```go
// 修正後の diagnosis_repository.go
buildBase := func() *gorm.DB {
    return r.db.WithContext(ctx).Model(&model.DiagnosisType{}).Scopes(clinicScope(clinicID))
}
if err := buildBase().Count(&total).Error; err != nil { ... }
if err := buildBase().Preload("Names").Order("sort_order ASC").Find(&categories).Error; err != nil { ... }
```

## 影響ファイル

- `backend/internal/repository/diagnosis_repository.go` — 行 32-46

## 優先度

**Low** — DRY原則違反。動作への影響なし。リファクタリング対象。

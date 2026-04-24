# TASK-085: diagnosis_type_repository — FindAll に Preload("Names") 欠落

## 優先度

MEDIUM

---

## 概要

`diagnosis_repository.go` の `DiagnosisType` に関する `FindAll` が `Preload("Names")` を行っていない。
`FindByID` は `Preload("Names")` を実装しており、一覧 API と詳細 API で返却データ構造が非対称になっている。

---

## 問題箇所

### diagnosis_repository.go

```go
// ❌ FindAll: Preload なし（L32-41）
func (r *diagnosisTypeRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, error) {
    categories := make([]model.DiagnosisType, 0)
    if err := r.db.WithContext(ctx).
        Model(&model.DiagnosisType{}).
        Scopes(clinicScope(clinicID)).
        Offset((page - 1) * limit).Limit(limit).
        Order("sort_order ASC, name ASC").
        Find(&categories).Error; err != nil {
        return nil, apperrors.FromGORM(err, "diagnosis_type", "")
    }
    return categories, nil
}

// ✅ FindByID: Preload あり（L43-52）
func (r *diagnosisTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisType, error) {
    var category model.DiagnosisType
    err := r.db.WithContext(ctx).
        Preload("Names").           // ← あり
        Scopes(clinicScope(clinicID)).
        Where("id = ?", id).
        First(&category).Error
    // ...
}
```

---

## 影響

- 一覧 API（`GET /v1/masters/diagnosis-types`）のレスポンスに `Names` が含まれない
- FE は一覧表示のたびに各 DiagnosisType の Detail API を N 回呼ぶか、名前情報なしで表示するしかない
- `permission_group_repository` は FindAll/FindByID 両方で `Preload("Rules")` を実装しており、こちらが正しいパターン

---

## 参照実装（permission_group_repository.go）

```go
// ✅ FindAll でも Preload("Rules")
func (r *permissionGroupRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error) {
    groups := make([]model.PermissionGroup, 0)
    err := r.db.WithContext(ctx).
        Scopes(clinicScope(clinicID)).
        Preload("Rules").    // ← 両方で Preload
        Order("sort_order ASC, name ASC").
        Find(&groups).Error
    // ...
}
```

---

## 修正方針

```go
// ✅ 修正後: FindAll に Preload("Names") を追加
func (r *diagnosisTypeRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, error) {
    categories := make([]model.DiagnosisType, 0)
    if err := r.db.WithContext(ctx).
        Model(&model.DiagnosisType{}).
        Preload("Names").           // 追加
        Scopes(clinicScope(clinicID)).
        Offset((page - 1) * limit).Limit(limit).
        Order("sort_order ASC, name ASC").
        Find(&categories).Error; err != nil {
        return nil, apperrors.FromGORM(err, "diagnosis_type", "")
    }
    return categories, nil
}
```

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `diagnosis_repository.go` | `FindAll`（DiagnosisType）に `Preload("Names")` を追加 |

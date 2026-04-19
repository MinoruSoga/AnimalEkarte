# TASK-075: exam_type / checkup_type — CountChildrenByParentID に clinic_id フィルタ欠落

## 優先度

HIGH

---

## 概要

階層構造（親子）を持つマスタ（exam_type / checkup_type）の `CountChildrenByParentID` メソッドが
`clinic_id` でスコープされていない。

Delete 前の「子レコード存在チェック」が他クリニックの子を含めてカウントするため、
TASK-071（CountUsageBy* 同問題）と同じ組織的バグが存在する。

---

## 問題箇所

### exam_type_repository.go

```go
// ❌ clinic_id フィルタなし
func (r *examinationTypeRepository) CountChildrenByParentID(ctx context.Context, parentID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.ExaminationType{}).
        Where("parent_id = ?", parentID).   // 全テナントをカウント
        Count(&count).Error; err != nil {
        return 0, apperrors.FromGORM(err, "examination_type", "")
    }
    return count, nil
}
```

### checkup_type_repository.go

```go
// ❌ clinic_id フィルタなし
func (r *checkupTypeRepository) CountChildrenByParentID(ctx context.Context, parentID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.CheckupType{}).
        Where("parent_id = ?", parentID).   // 全テナントをカウント
        Count(&count).Error; err != nil {
        return 0, apperrors.FromGORM(err, "checkup_type", "")
    }
    return count, nil
}
```

### service 呼び出し側（exam_type_service.go, checkup_type_service.go）

```go
// ❌ clinicID を渡していない
childCount, err := s.repo.CountChildrenByParentID(ctx, id)
```

---

## 修正方針

```go
// ✅ clinicID パラメータを追加し clinicScope でスコープ
func (r *examinationTypeRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.ExaminationType{}).
        Scopes(clinicScope(clinicID)).
        Where("parent_id = ?", parentID).
        Count(&count).Error; err != nil {
        return 0, apperrors.FromGORM(err, "examination_type", "")
    }
    return count, nil
}
```

---

## Interface 変更が必要なメソッド

```go
// ❌ 修正前
CountChildrenByParentID(ctx context.Context, parentID uint64) (int64, error)

// ✅ 修正後
CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
```

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `exam_type_repository.go` | `CountChildrenByParentID` に `clinicID` 追加・`Scopes(clinicScope(clinicID))` 追加 |
| `checkup_type_repository.go` | 同上 |
| `exam_type_service.go` | Delete で `CountChildrenByParentID(ctx, clinicID, id)` に変更 |
| `checkup_type_service.go` | 同上 |
| 各 interface 定義 | シグネチャ更新 |

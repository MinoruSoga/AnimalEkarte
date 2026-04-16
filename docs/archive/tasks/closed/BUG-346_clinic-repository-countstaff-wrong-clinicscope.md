# BUG-346: clinic_repository.CountStaffByClinicID が staffs に誤った clinicScope を適用

## 概要

`clinic_repository.go:118-128` の `CountStaffByClinicID` が、`clinic_id` カラムを持たない `staffs` テーブルに `clinicScope(clinicID)` を適用している。
実行時に PostgreSQL が "column clinic_id does not exist" エラーを返す。

## 影響範囲

- **呼び出し元**: `clinic_service.go:175` — クリニック削除時に `CountStaffByClinicID` を呼ぶ
- **エンドポイント**: `DELETE /v1/masters/clinics/:id`
- **症状**: クリニック削除が常に 500 エラーで失敗する

## 根本原因

`staffs` テーブルは v31 スキーマ再設計（`001_init.sql:1343,1550`）で `clinic_id` カラムを削除し、`staff_clinic_assignments` 中間テーブルに移行した。
しかし `CountStaffByClinicID` の実装がこの変更に追従していない。

```go
// backend/internal/repository/clinic_repository.go:118-128（現在・バグあり）
func (r *clinicRepository) CountStaffByClinicID(ctx context.Context, clinicID uint64) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&model.Staff{}).
        Scopes(clinicScope(clinicID)).Where("deleted_at IS NULL"). // staffs に clinic_id なし → 実行時エラー
        Count(&count).Error
    if err != nil {
        return 0, apperrors.FromGORM(err, "staff", fmt.Sprintf("clinic_id=%d", clinicID))
    }
    return count, nil
}
```

```
-- 001_init.sql (マイグレーションコメント)
-- Deleted: idx_staffs_clinic_id (staffs now uses account_id; clinic membership tracked via staff_clinic_assignments)
-- Deleted: idx_staffs_clinic_name (staffs no longer has clinic_id directly)
```

## 修正方針

`staff_clinic_assignments` を JOIN してカウントする。

```go
// 修正後
func (r *clinicRepository) CountStaffByClinicID(ctx context.Context, clinicID uint64) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&model.Staff{}).
        Joins("INNER JOIN staff_clinic_assignments ON staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ?", clinicID).
        Where("staffs.deleted_at IS NULL").
        Count(&count).Error
    if err != nil {
        return 0, apperrors.FromGORM(err, "staff", fmt.Sprintf("clinic_id=%d", clinicID))
    }
    return count, nil
}
```

## 優先度

**HIGH** — クリニック削除 API が常に 500 エラー。

## 確認方法

1. `DELETE /v1/masters/clinics/:id` を実行 → 現在: 500 エラー
2. 修正後: スタッフが存在するクリニックは 409、存在しないクリニックは 204

## 関連ファイル

- `backend/internal/repository/clinic_repository.go:118-128`
- `backend/internal/service/clinic_service.go:175`
- `backend/migrations/001_init.sql:1343,1550`（削除コメント）

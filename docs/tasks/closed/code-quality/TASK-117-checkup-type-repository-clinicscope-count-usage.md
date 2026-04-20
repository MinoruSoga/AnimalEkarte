# TASK-117: `checkup_type_repository.go` — `CountUsageByCheckupTypeID` が `clinicScope` を使用していない

## 優先度

**Medium** — 同一ファイル内でテナント分離の実装パターンが混在しており、保守性と一貫性が低下している。

---

## 概要

`checkup_type_repository.go` の `CountUsageByCheckupTypeID`（行 93-101）は、
`model.Checkup`（`checkups` テーブル）に対して `clinic_id` を直接 `Where` 句に記述している。

`checkups` テーブルは直接 `clinic_id` を持つため、`clinicScope` の使用対象であり、
同ファイル内の他のメソッド（`FindAll`, `FindByID` 等）や `CountChildrenByParentID`（行 105-114）との
実装パターンが不統一になっている。

---

## 問題箇所

### `repository/checkup_type_repository.go:93-101`

```go
// ❌ clinic_id を直接 WHERE に記述（clinicScope を使っていない）
func (r *checkupTypeRepository) CountUsageByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.Checkup{}).
        Where("checkup_type_id = ? AND clinic_id = ?", checkupTypeID, clinicID).
        Count(&count).Error; err != nil {
        return 0, apperrors.FromGORM(err, "checkup_type", "")
    }
    return count, nil
}
```

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ 同ファイル CountChildrenByParentID（行 105-114） — clinicScope を正しく使用
func (r *checkupTypeRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.CheckupType{}).
        Scopes(clinicScope(clinicID)).
        Where("parent_id = ?", parentID).
        Count(&count).Error; err != nil {
        return 0, apperrors.FromGORM(err, "checkup_type", "")
    }
    return count, nil
}

// ✅ 同ファイル FindAll（行 33-39） — clinicScope を正しく使用
err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Order(...).Find(&checkupTypes).Error
```

---

## 修正方針

### `repository/checkup_type_repository.go:93-101`

`Where("checkup_type_id = ? AND clinic_id = ?")` を `Scopes(clinicScope(clinicID)).Where("checkup_type_id = ?")` に変更する。

```go
// ✅ 修正後
func (r *checkupTypeRepository) CountUsageByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.Checkup{}).
        Scopes(clinicScope(clinicID)).
        Where("checkup_type_id = ?", checkupTypeID).
        Count(&count).Error; err != nil {
        return 0, apperrors.FromGORM(err, "checkup_type", "")
    }
    return count, nil
}
```

---

## 影響範囲

| ファイル | 対象 | 状態 |
|---------|------|------|
| `repository/checkup_type_repository.go:93-101` | CountUsageByCheckupTypeID | ❌ clinic_id を直接 WHERE に記述 |

---

## 準拠すべきプロジェクト規約

### `backend/CLAUDE.md` — マルチテナント: clinicScope（必須）

> プライマリテーブルが直接 `clinic_id` を持つ場合は、`clinicScope` を使用する。
> ❌ 禁止: 手動で clinic_id を WHERE に記述
> ✅ 必須: clinicScope を使用

### プロジェクト内参照実装

- `repository/checkup_type_repository.go:105-114` — `CountChildrenByParentID` が `clinicScope` を正しく使用（同ファイル内の不統一）

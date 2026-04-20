# TASK-110: `billing_item_repository.go` — テナント分離実装の不統一

## 優先度

**Medium** — 同一ファイル内でテナント分離の実装パターンが混在しており、保守性と可読性が低下している。

---

## 概要

`billing_item_repository.go` で 2 つのテナント分離パターンが混在している:

### パターン1（正しい）— `FindByID`
`clinic_id` を **JOIN 条件** に含める（backend/CLAUDE.md 推奨パターン）
```go
Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID)
```

### パターン2（不統一）— `FindByBillingID`
`clinic_id` を **WHERE 句** に記述（JOIN 条件に含まない）
```go
Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.deleted_at IS NULL").
Where("billings.clinic_id = ? AND billing_items.billing_id = ?", clinicID, billingID).
```

### パターン3（TOCTOU 懸念）— `UpdateFields`, `Delete`
`FindByID` で事前確認後、mutation は `id = ?` のみ（clinic_id が mutation に含まれない）
```go
if _, err := r.FindByID(ctx, clinicID, id); err != nil { return err }
result := r.db.WithContext(ctx).Model(&model.BillingItem{}).
    Where("id = ?", id).  // ← clinic_id なし
    Updates(fields)
```

---

## 問題箇所

### `repository/billing_item_repository.go:46`

```go
// ❌ clinic_id が WHERE 句（JOIN 条件に含まれていない）
func (r *billingItemRepository) FindByBillingID(...) {
    if err := r.db.WithContext(ctx).
        Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.deleted_at IS NULL").
        Where("billings.clinic_id = ? AND billing_items.billing_id = ?", clinicID, billingID).
        ...
```

### `repository/billing_item_repository.go:67-70`

```go
// ❌ 事前確認後 id = ? のみで UPDATE（clinic_id なし）
result := r.db.WithContext(ctx).
    Model(&model.BillingItem{}).
    Where("id = ?", id).
    Updates(fields)
```

### `repository/billing_item_repository.go:85`

```go
// ❌ 事前確認後 id = ? のみで DELETE（clinic_id なし）
result := r.db.WithContext(ctx).Delete(&model.BillingItem{}, "id = ?", id)
```

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ FindByID（同ファイル:34）— JOIN 条件に clinic_id 含む
Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
Where("billing_items.id = ?", id).

// ✅ backend/CLAUDE.md — JOIN 経由のテナント判定パターン
// JOIN 先テーブルの clinic_id フィルタが JOIN 条件に含まれているか
// JOIN 先テーブルの deleted_at IS NULL が JOIN 条件に含まれているか
```

---

## 修正方針

### 1. `repository/billing_item_repository.go:43-52` — FindByBillingID

`clinic_id` を JOIN 条件に移動する（WHERE 句から削除）。

```go
// ✅ 修正後
func (r *billingItemRepository) FindByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingItem, error) {
    items := make([]model.BillingItem, 0)
    if err := r.db.WithContext(ctx).
        Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
        Where("billing_items.billing_id = ?", billingID).
        Order("sort_order ASC, id ASC").
        Find(&items).Error; err != nil {
        return nil, apperrors.FromGORM(err, "billing_item", "")
    }
    return items, nil
}
```

### 2. `repository/billing_item_repository.go:62-77` — UpdateFields

事前 FindByID を廃止し、JOIN ベースの mutation に変更する。

```go
// ✅ 修正後
func (r *billingItemRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
    result := r.db.WithContext(ctx).
        Model(&model.BillingItem{}).
        Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
        Where("billing_items.id = ?", id).
        Updates(fields)
    if result.Error != nil {
        return apperrors.FromGORM(result.Error, "billing_item", fmt.Sprintf("%d", id))
    }
    if result.RowsAffected == 0 {
        return apperrors.WrapNotFound("billing_item", fmt.Sprintf("%d", id))
    }
    return nil
}
```

### 3. `repository/billing_item_repository.go:80-93` — Delete

同様に JOIN ベースの DELETE に変更する。

```go
// ✅ 修正後
func (r *billingItemRepository) Delete(ctx context.Context, clinicID, id uint64) error {
    result := r.db.WithContext(ctx).
        Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
        Delete(&model.BillingItem{}, "billing_items.id = ?", id)
    if result.Error != nil {
        return apperrors.FromGORM(result.Error, "billing_item", fmt.Sprintf("%d", id))
    }
    if result.RowsAffected == 0 {
        return apperrors.WrapNotFound("billing_item", fmt.Sprintf("%d", id))
    }
    return nil
}
```

---

## 影響範囲

| ファイル | 対象メソッド | 問題 |
|---------|------------|------|
| `repository/billing_item_repository.go:46` | FindByBillingID | clinic_id が WHERE 句（JOIN 条件に含まれていない） |
| `repository/billing_item_repository.go:67-70` | UpdateFields | mutation に clinic_id なし |
| `repository/billing_item_repository.go:85` | Delete | mutation に clinic_id なし |

---

## 準拠すべきプロジェクト規約

### `backend/CLAUDE.md` — JOIN を含む repository メソッドのレビューチェックリスト

> - [ ] JOIN 先テーブルの `clinic_id` フィルタが JOIN 条件に含まれているか
> - [ ] JOIN 先テーブルの `deleted_at IS NULL` が JOIN 条件に含まれているか

### プロジェクト内参照実装

- `repository/billing_item_repository.go:34` — `FindByID` が正しいパターンを実装済み（同ファイル内の不統一）

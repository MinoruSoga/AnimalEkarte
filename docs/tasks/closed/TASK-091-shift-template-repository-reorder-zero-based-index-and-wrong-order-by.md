# TASK-091: shift_template_repository — Reorder 0始まりインデックス + ORDER BY 非標準

## 優先度

HIGH

---

## 概要

`shift_template_repository.go` の Reorder 実装が共通ヘルパー（`reorderByClinicID`）と異なり、
sort_order に 0 始まりの値を設定している（実質バグ）。
また FindAll の ORDER BY が全マスタ標準と異なる。

---

## 問題箇所

### 1. Reorder: 0始まりインデックス（L118-136）

```go
// ❌ shift_template_repository.go:123 — sort_order が 0 から始まる
func (r *shiftTemplateRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
    if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        for i, id := range ids {
            result := tx.Model(&model.ShiftTemplate{}).
                Scopes(clinicScope(clinicID)).Where("id = ?", id).
                Update("sort_order", i)   // ← i は 0 始まり（バグ）
```

共通ヘルパー `reorderByClinicID`（`helpers.go:19`）は `i+1`（1始まり）を使用：

```go
// ✅ helpers.go:19 — 全ドメイン共通実装
Update("sort_order", i+1)
```

この不一致により、シフトテンプレートのみ sort_order が 0 始まりになり、
FindAll の `ORDER BY sort_order ASC` でソート順が 1 つずれる。

---

### 2. FindAll: 非標準 ORDER BY（L37）

```go
// ❌ L37: name ではなく id でセカンダリソート
Order("sort_order ASC, id ASC")

// ✅ 全マスタ標準
Order("sort_order ASC, name ASC")
```

`model.ShiftTemplate` には `Name` フィールドが存在する（`shiftTemplateResponse.Name` で確認済み）。

---

## 修正方針

```go
// ✅ 修正後: 共通ヘルパーに委譲する
func (r *shiftTemplateRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
    return reorderByClinicID(ctx, r.db, &model.ShiftTemplate{}, "shift_template", clinicID, ids)
}
```

FindAll ORDER BY:
```go
// ✅ 修正後
Order("sort_order ASC, name ASC")
```

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `repository/shift_template_repository.go` | `Reorder` を `reorderByClinicID` 委譲に変更（0→1始まり修正）、`FindAll` ORDER BY を `sort_order ASC, name ASC` に変更 |

# TASK-118: `inquiry_repository.go` — `clinicScope` 未使用 + Update 後の手動フィールド再代入

## 優先度

**Medium** — テナント分離実装の不統一 + 保守性リスクのある冗長コード。

---

## 概要

`inquiry_repository.go` に 2 つの問題が混在している。

1. `CountByChiefComplaintTypeID`（行 98-108）で `clinic_id` を直接 `Where` 句に記述（`clinicScope` 未使用）
2. `UpsertByMedicalRecordID`（行 78-91）で `Updates()` 後に全フィールドを手動再代入しており、
   updates map との二重管理になっている

---

## 問題箇所

### 問題 1: `repository/inquiry_repository.go:98-108` — `clinicScope` 未使用

```go
// ❌ clinic_id を直接 WHERE に記述
func (r *inquiryRepository) CountByChiefComplaintTypeID(ctx context.Context, clinicID, categoryID uint64) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&model.Inquiry{}).
        Where("clinic_id = ? AND chief_complaint_type_id = ?", clinicID, categoryID).
        Count(&count).Error
```

`inquiries` テーブルは直接 `clinic_id` を持つため `clinicScope` を使うのがプロジェクト規約。
一方、同ファイルの `UpsertByMedicalRecordID`（行 39-40）では `Scopes(clinicScope(clinicID))` を正しく使用しており、
同一ファイル内で不統一になっている。

### 問題 2: `repository/inquiry_repository.go:78-91` — Update 後の手動フィールド再代入

```go
// ❌ updates map（13フィールド）と手動代入（13フィールド）の二重管理
updates := map[string]any{
    "chief_complaint": inquiry.ChiefComplaint,
    "notes":           inquiry.Notes,
    // ... 11フィールド
}
if err := r.db.WithContext(ctx).Model(&existing).Updates(updates).Error; err != nil { ... }

// ← 全フィールドを手動で再代入（DRY 違反）
existing.ChiefComplaint = inquiry.ChiefComplaint
existing.Notes          = inquiry.Notes
// ... 11フィールド
```

フィールドが追加・削除された際に updates map と手動代入の両方を変更する必要があり、
片方を更新し忘れると戻り値と DB の状態が乖離する。

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ repository/inquiry_repository.go:39-40 — clinicScope を正しく使用（同ファイル内）
r.db.WithContext(ctx).
    Table("medical_records").
    Scopes(clinicScope(clinicID)).Where("id = ?", inquiry.MedicalRecordID).
    Count(&mrCount)

// ✅ repository/billing_item_repository.go — UpdateFields 後に FindByID で最新状態を取得
result := r.db.WithContext(ctx).
    Model(&model.BillingItem{}).
    Scopes(...).
    Where("id = ?", id).
    Updates(fields)
...
return r.FindByID(ctx, clinicID, id)  // 最新状態を DB から再取得
```

---

## 修正方針

### 問題 1: `CountByChiefComplaintTypeID` の clinic_id 条件

```go
// ✅ 修正後
func (r *inquiryRepository) CountByChiefComplaintTypeID(ctx context.Context, clinicID, categoryID uint64) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&model.Inquiry{}).
        Scopes(clinicScope(clinicID)).
        Where("chief_complaint_type_id = ?", categoryID).
        Count(&count).Error
    if err != nil {
        return 0, apperrors.FromGORM(err, "inquiry", "")
    }
    return count, nil
}
```

### 問題 2: Update 後の手動再代入を DB 再取得に置き換え

```go
// ✅ 修正後（手動再代入を廃止し DB から再取得）
if err := r.db.WithContext(ctx).Model(&existing).Updates(updates).Error; err != nil {
    return nil, apperrors.FromGORM(err, "inquiry", "")
}

// 最新状態を DB から取得（updated_at 等の DB 管理フィールドも正確に反映）
var refreshed model.Inquiry
if err := r.db.WithContext(ctx).
    Scopes(clinicScope(clinicID)).
    Where("id = ?", existing.ID).
    First(&refreshed).Error; err != nil {
    return nil, apperrors.FromGORM(err, "inquiry", "")
}
return &refreshed, nil
```

---

## 影響範囲

| ファイル | 対象 | 状態 |
|---------|------|------|
| `repository/inquiry_repository.go:98-108` | CountByChiefComplaintTypeID | ❌ clinic_id を直接 WHERE に記述 |
| `repository/inquiry_repository.go:78-91` | UpsertByMedicalRecordID の手動再代入 | ❌ updates map との二重管理（DRY 違反） |

---

## 準拠すべきプロジェクト規約

### `backend/CLAUDE.md` — マルチテナント: clinicScope（必須）

> プライマリテーブルが直接 `clinic_id` を持つ場合は、`clinicScope` を使用する。

### プロジェクト内参照実装

- `repository/inquiry_repository.go:39-40` — 同ファイル内で clinicScope を正しく使用済み
- `repository/billing_item_repository.go` — UpdateFields 後に FindByID で最新状態を返すパターン

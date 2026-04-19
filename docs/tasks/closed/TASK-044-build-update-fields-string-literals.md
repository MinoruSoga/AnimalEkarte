# TASK-044: buildUpdateFields でカラム名をベタ文字列リテラル使用 — 6ドメイン

## 優先度

MEDIUM

---

## 概要

参照実装 `medicine_service.go` は `buildMedicineUpdateFields` の中で `colMedicineName` 等の**名前付き定数**を使用している。しかし以下の 6 ドメインはカラム名をベタ文字列で直接 map キーに書いており、タイポによる無音バグ・カラムリネーム時のスコープ漏れのリスクがある。

```go
// ✅ medicine（参照実装）
const (
    colMedicineName        = "name"
    colMedicinePrice       = "price"
    colMedicineIsActive    = "is_active"
    // ...
)
func buildMedicineUpdateFields(input *UpdateMedicineInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil { fields[colMedicineName] = *input.Name }
    // ...
}

// ❌ 各ドメイン（以下）
func buildCheckupTypeUpdateFields(input UpdateCheckupTypeInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil { fields["name"] = *input.Name }  // ← ベタ文字列
    // ...
}
```

---

## 対象ドメインと該当ファイル

| ドメイン | ファイル |
|---------|---------|
| checkup_type | `backend/internal/service/checkup_type_service.go:144-173` |
| chief_complaint_type | `backend/internal/service/chief_complaint_type_service.go:134-149` |
| reservation_type_group | `backend/internal/service/reservation_type_group_service.go:111-126` |
| reservation_type | `backend/internal/service/reservation_type_service.go`（buildReservationTypeUpdateFields 内） |
| company | `backend/internal/service/company_service.go:68-104` |
| hospitalization_plan | `backend/internal/service/hospitalization_plan_service.go:113-143` |

---

## 修正方針（共通）

各 service ファイルの先頭に `const` ブロックを追加し、`buildXxxUpdateFields` 内のベタ文字列をすべて定数参照に置き換える。

```go
// 例: checkup_type_service.go
const (
    colCheckupTypeName        = "name"
    colCheckupTypePrice       = "price"
    colCheckupTypeIsActive    = "is_active"
    colCheckupTypeDescription = "description"
    colCheckupTypeInterval    = "interval"
    colCheckupTypeTargetAge   = "target_age"
    colCheckupTypeParentID    = "parent_id"
    colCheckupTypeSortOrder   = "sort_order"
)

func buildCheckupTypeUpdateFields(input UpdateCheckupTypeInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil        { fields[colCheckupTypeName]        = *input.Name }
    if input.Price != nil       { fields[colCheckupTypePrice]       = *input.Price }
    if input.IsActive != nil    { fields[colCheckupTypeIsActive]    = *input.IsActive }
    // ...
}
```

## 備考

`diagnosis_service.go` はすでに定数化済み（参照実装以外で正しい実装の例）。全ドメインでこれに統一すること。

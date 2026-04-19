# TASK-080: permission_group_service — buildUpdateFields に const 定数未定義

## 優先度

MEDIUM

---

## 概要

`permission_group_service.go` の `buildPermissionGroupUpdateFields` 関数が
DB カラム名をベアストリングリテラル（`"name"`, `"description"` 等）で直接記述しており、
参照実装（medicine_service.go）の名前付き定数パターンと不統一。

---

## 問題箇所

### permission_group_service.go:154-172

```go
// ❌ ベアストリングリテラル（const なし）
func buildPermissionGroupUpdateFields(input *UpdatePermissionGroupInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil {
        fields["name"] = *input.Name
    }
    if input.Description != nil {
        fields["description"] = *input.Description
    }
    if input.Color != nil {
        fields["color"] = *input.Color
    }
    if input.SortOrder != nil {
        fields["sort_order"] = *input.SortOrder
    }
    if input.IsActive != nil {
        fields["is_active"] = *input.IsActive
    }
    return fields
}
```

---

## 参照実装（medicine_service.go）

```go
// ✅ 定数でカラム名を管理
const (
    colMedicineName        = "name"
    colMedicineDosageForm  = "dosage_form"
    colMedicineDescription = "description"
    colMedicineIsActive    = "is_active"
    colMedicineSortOrder   = "sort_order"
)

func buildMedicineUpdateFields(input *UpdateMedicineInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil {
        fields[colMedicineName] = *input.Name
    }
    // ...
}
```

---

## 修正方針

```go
// ✅ 修正後: permission_group_service.go
const (
    colPermissionGroupName        = "name"
    colPermissionGroupDescription = "description"
    colPermissionGroupColor       = "color"
    colPermissionGroupSortOrder   = "sort_order"
    colPermissionGroupIsActive    = "is_active"
)

func buildPermissionGroupUpdateFields(input *UpdatePermissionGroupInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil {
        fields[colPermissionGroupName] = *input.Name
    }
    if input.Description != nil {
        fields[colPermissionGroupDescription] = *input.Description
    }
    if input.Color != nil {
        fields[colPermissionGroupColor] = *input.Color
    }
    if input.SortOrder != nil {
        fields[colPermissionGroupSortOrder] = *input.SortOrder
    }
    if input.IsActive != nil {
        fields[colPermissionGroupIsActive] = *input.IsActive
    }
    return fields
}
```

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `permission_group_service.go` | const ブロック追加、`buildPermissionGroupUpdateFields` のベアストリングをconst参照に変更 |

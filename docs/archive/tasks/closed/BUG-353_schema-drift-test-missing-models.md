# BUG-353: schema_drift_test.go の allModels() に 5 モデルが未登録

## 概要

`schema_drift_test.go` の `allModels()` に以下の 5 モデルが含まれていない。
これらのテーブルでスキーマドリフト（Go モデルと DB スキーマの乖離）が発生しても検出できない。

## 欠落モデル

| モデル | テーブル | ファイル |
|--------|---------|---------|
| `PermissionGroup` | `permission_groups` | `permission_group.go:27` |
| `PermissionGroupRule` | `permission_group_rules` | `permission_group.go:45` |
| `StaffPermissionGroup` | `staff_permission_groups` | `permission_group.go:58` |
| `ReservationTypeGroup` | `reservation_type_groups` | `reservation_type_group.go:17` |
| `ClinicHoliday` | `clinic_holidays` | `staff.go:84` |

## 修正内容

`schema_drift_test.go` の `allModels()` に追加:

```go
func allModels() []any {
    return []any{
        // ... 既存 ...
        // 以下を追加
        &model.PermissionGroup{},
        &model.PermissionGroupRule{},
        &model.StaffPermissionGroup{},
        &model.ReservationTypeGroup{},
        &model.ClinicHoliday{},
    }
}
```

## 優先度

**MEDIUM** — CI のスキーマドリフト検出の網羅率に影響。5 テーブル分の Go ↔ DB 乖離を検出できない。

## 関連ファイル

- `backend/internal/model/schema_drift_test.go:39-103`

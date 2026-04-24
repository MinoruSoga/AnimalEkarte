# TASK-089: slog エンティティ ID フィールド命名不統一（複数サービス）

## 優先度

MEDIUM

---

## 概要

複数の master service で `slog` に渡すエンティティ ID フィールド名が、
プロジェクト規約（`{domain_name}_id` 形式）に準拠していない。

規約: slog のエンティティ ID フィールドは **`{ドメイン名}_id`** とする。
（例: `medicine_id`, `exam_type_id`, `permission_group_id`）

---

## 問題箇所

### diagnosis_service.go

```go
// ❌ L117: "type_id" — 曖昧（どの type か不明）
slog.Uint64("type_id", diagType.ID)
// → 正しくは "diagnosis_type_id"

// ❌ L257: "name_id" — 曖昧（どの name か不明）
slog.Uint64("name_id", name.ID)
// → 正しくは "diagnosis_name_id"
```

### chief_complaint_type_service.go

```go
// ❌ L82: "category_id" — 曖昧（chief_complaint のカテゴリとわかりにくい）
slog.Uint64("category_id", category.ID)
// → 正しくは "chief_complaint_type_id"
```

### inquiry_template_service.go

```go
// ❌ L84: "template_id" — 曖昧（どの template か不明）
slog.Uint64("template_id", template.ID)
// → 正しくは "inquiry_template_id"
```

### merchandise_item_service.go

```go
// ❌ L147 (Create): "item_id" — 曖昧（どの item か不明）
slog.Uint64("item_id", item.ID)
// → 正しくは "merchandise_item_id"

// ✅ L175 (Update): "merchandise_item_id" — 正しい
slog.Uint64("merchandise_item_id", id)
// ↑ Create と Update でフィールド名が不統一
```

### permission_group_service.go

```go
// ❌ L83: "group_id" — 曖昧（どの group か不明）
slog.Uint64("group_id", group.ID)
// → 正しくは "permission_group_id"
```

---

## 参照実装（medicine_service.go）

```go
// ✅ medicine_service.go: "medicine_id"
slog.InfoContext(ctx, "medicine created",
    slog.Uint64("clinic_id", clinicID),   // clinic_id は FIRST
    slog.Uint64("medicine_id", medicine.ID),
)
```

---

## 修正方針

各サービスのエンティティ ID フィールド名を `{ドメイン名}_id` 形式に統一する。

| ファイル | 行 | 修正前 | 修正後 |
|---------|-----|-------|-------|
| `diagnosis_service.go` | L117 | `"type_id"` | `"diagnosis_type_id"` |
| `diagnosis_service.go` | L257 | `"name_id"` | `"diagnosis_name_id"` |
| `chief_complaint_type_service.go` | L82 | `"category_id"` | `"chief_complaint_type_id"` |
| `inquiry_template_service.go` | L84 | `"template_id"` | `"inquiry_template_id"` |
| `merchandise_item_service.go` | L147 | `"item_id"` | `"merchandise_item_id"` |
| `permission_group_service.go` | L83 | `"group_id"` | `"permission_group_id"` |

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `service/diagnosis_service.go` | `"type_id"` → `"diagnosis_type_id"`、`"name_id"` → `"diagnosis_name_id"` |
| `service/chief_complaint_type_service.go` | `"category_id"` → `"chief_complaint_type_id"` |
| `service/inquiry_template_service.go` | `"template_id"` → `"inquiry_template_id"` |
| `service/merchandise_item_service.go` | `"item_id"` → `"merchandise_item_id"`（Create の slog を Update と統一） |
| `service/permission_group_service.go` | `"group_id"` → `"permission_group_id"` |

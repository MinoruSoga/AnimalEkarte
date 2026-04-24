# TASK-066: permission_group / staff handler — model 型をハンドラ内で直接構築

## 優先度

MEDIUM

---

## 概要

`permission_group_handler.go` と `staff_handler.go` が handler 層で `model.*` 型を直接構築しており、
TASK-049（procedure / cage / vaccine の handler→model 依存）と同じパターンの違反が存在する。

handler は service の Input DTO を組み立てる責務のみを持つべきであり、
`model.PermissionGroup{}` / `model.PermissionGroupRule{}` / `model.StaffClinicAssignment{}` の構築は service 層で行うべき。

---

## 問題箇所

### 1. permission_group_handler.go — CreatePermissionGroup（L64-71 概算）

```go
// ❌ handler 内で model 型を直接構築
pg := &model.PermissionGroup{
    ClinicID:    clinicID,
    Name:        req.Name,
    Description: req.Description,
    Color:       req.Color,
    IsActive:    req.IsActive,
    SortOrder:   req.SortOrder,
}
if err := h.svc.PermissionGroup.Create(c.Request.Context(), pg); err != nil {
```

### 2. permission_group_handler.go — SetPermissionGroupRules（L261-270 概算）

```go
// ❌ handler 内で model スライスを構築
rules := make([]model.PermissionGroupRule, 0, len(req.Rules))
for _, r := range req.Rules {
    rules = append(rules, model.PermissionGroupRule{
        Resource:  r.Resource,
        CanView:   r.CanView,
        CanCreate: r.CanCreate,
        CanEdit:   r.CanEdit,
        CanDelete: r.CanDelete,
    })
}
```

### 3. staff_handler.go — CreateStaff（L107-111 概算）

```go
// ❌ handler 内で model 型を直接構築
if asgErr := h.svc.StaffClinicAssignment.Create(ctx, &model.StaffClinicAssignment{
    StaffID:  staff.ID,
    ClinicID: clinicID,
    IsMain:   true,
}); asgErr != nil {
```

---

## 修正方針

### permission_group — Create

```go
// ✅ service.CreatePermissionGroupInput を定義して使用
type CreatePermissionGroupInput struct {
    Name        string
    Description string
    Color       string
    IsActive    bool
    SortOrder   int
}

// handler 側
if err := h.svc.PermissionGroup.Create(c.Request.Context(), clinicID, service.CreatePermissionGroupInput{
    Name:        req.Name,
    Description: req.Description,
    Color:       req.Color,
    IsActive:    req.IsActive,
    SortOrder:   req.SortOrder,
}); err != nil {
```

### permission_group — SetRules

```go
// ✅ service.SetPermissionGroupRulesInput を定義して使用
// handler は req.Rules をそのまま渡し、model 変換は service 層で行う
```

### staff — CreateStaff assignment

```go
// ✅ StaffClinicAssignment の構築を service.Create 内部に移動
// service.CreateStaffInput に IsMain フラグを持たせる、または
// service.CreateStaff が内部でクリニック割当も処理する
```

---

## 影響ファイル

- `backend/internal/handler/permission_group_handler.go`
- `backend/internal/handler/staff_handler.go`
- `backend/internal/service/permission_group_service.go`（Input DTO 追加）
- `backend/internal/service/staff_service.go`（assignment 構築をここに移動）

---

## 追記: permission_group_service の slog 順序違反

`permission_group_service.go` の slog 呼び出しで `clinic_id` が先頭でない箇所がある（TASK-057 パターン）:

```go
// ❌ group_id が先頭
slog.InfoContext(ctx, "permission group created",
    slog.Uint64("group_id", group.ID),
    slog.Uint64("clinic_id", group.ClinicID))

// ✅ 修正後
slog.InfoContext(ctx, "permission group created",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("group_id", group.ID))
```

対象: Create / Update / Delete / SetRules の各 slog 呼び出しで clinic_id を先頭に統一すること。

---

## 参照実装

`occupation_handler.go` の Create — service.CreateOccupationInput を使用する正しい実装。

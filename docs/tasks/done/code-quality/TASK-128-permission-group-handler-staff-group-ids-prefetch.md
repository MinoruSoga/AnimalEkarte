# TASK-128: `permission_group_handler.go` — Handler が staffGroupIDs を事前取得して Service に渡している

## 優先度

**Medium** — Handler が外部データの取得責務を担っており、責務分離が崩れている。

---

## 概要

`permission_group_handler.go` の `SetPermissionGroupRules` ハンドラ（行 202-270）において、
Handler がスタッフのグループ ID 一覧（`staffGroupIDs`）を自ら取得してから Service に渡している。

`GetPermissionGroupIDs` の呼び出しと staffGroupIDs の構築は Service 層の責務であり、
Handler は staffID のみを渡せば十分である。

---

## 問題箇所

### `handler/permission_group_handler.go:224-247`

```go
// ❌ Handler が staffGroupIDs を事前取得して Service に渡している
staffID, ok := extractStaffID(c)
if !ok {
    return
}
// staffID が所属するグループ ID 一覧を取得（self-reference チェックに使用）
var staffGroupIDs []uint64
if myGroupIDs, groupErr := h.svc.Staff.GetPermissionGroupIDs(c.Request.Context(), staffID); groupErr == nil {
    staffGroupIDs = myGroupIDs
}

// Convert request rules to model
rules := make([]model.PermissionGroupRule, 0, len(req.Rules))
for _, r := range req.Rules {
    rules = append(rules, model.PermissionGroupRule{...})
}

if err := h.svc.PermissionGroup.SetRules(c.Request.Context(), id, rules, staffGroupIDs); err != nil {
    ...
}
```

**2 つの問題:**
1. Handler が `h.svc.Staff.GetPermissionGroupIDs(...)` を呼び出して staffGroupIDs を取得 → Handler が複数の Service に依存しデータ収集を担っている
2. `model.PermissionGroupRule` の構築を Handler が行っている → モデル構築は Service の責務

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ handler/vaccine_handler.go — Handler は raw データを渡すのみ
vaccine, err := h.svc.Vaccine.Create(c.Request.Context(), clinicID, &service.CreateVaccineInput{
    Name: req.Name, ...
})

// ✅ service/vaccine_service.go — モデル構築と検索は Service で実行
func (s *vaccineService) Create(...) (*model.Vaccine, error) {
    ...
    v := &model.Vaccine{Name: input.Name, ...}
    ...
}
```

---

## 修正方針

### `service/permission_group_service.go` — SetRules シグネチャ変更

```go
// 修正前
func (s *permissionGroupService) SetRules(
    ctx context.Context,
    id uint64,
    rules []model.PermissionGroupRule,
    staffGroupIDs []uint64,  // ← Handler が取得して渡している
) error

// ✅ 修正後: staffID を渡し、Service 内で staffGroupIDs を取得
func (s *permissionGroupService) SetRules(
    ctx context.Context,
    id uint64,
    rules []model.PermissionGroupRule,
    actorStaffID uint64,  // ← staffID のみ渡す
) error {
    // Service 内で staffGroupIDs を取得
    staffGroupIDs, err := s.staffRepo.GetPermissionGroupIDs(ctx, actorStaffID)
    if err != nil {
        staffGroupIDs = []uint64{}  // エラー時は空（自己参照チェック不能なら許可方向）
    }
    return validateNotSelfReference(id, rules, staffGroupIDs)
}
```

### `handler/permission_group_handler.go` — Handler をシンプル化

```go
// ✅ 修正後
staffID, ok := extractStaffID(c)
if !ok {
    return
}
rules := make([]model.PermissionGroupRule, 0, len(req.Rules))
for _, r := range req.Rules {
    rules = append(rules, model.PermissionGroupRule{...})
}
if err := h.svc.PermissionGroup.SetRules(c.Request.Context(), id, rules, staffID); err != nil {
    ...
}
```

---

## 影響範囲

| ファイル | 行 | 状態 |
|---------|---|------|
| `handler/permission_group_handler.go:229-233` | `GetPermissionGroupIDs` 呼び出し | ❌ Handler が外部データ取得 |
| `handler/permission_group_handler.go:247` | `SetRules(..., staffGroupIDs)` 呼び出し | ❌ Handler から取得データを渡す |
| `service/permission_group_service.go` | `SetRules` シグネチャ | staffID を受け取り内部で取得するよう変更 |

---

## 準拠すべきプロジェクト規約

### `backend/CLAUDE.md` — 依存関係の方向

> Handler はリクエスト解析と Service への委譲のみを担う。

### プロジェクト内参照実装

- `handler/vaccine_handler.go` — Handler が raw データのみ渡す正しいパターン

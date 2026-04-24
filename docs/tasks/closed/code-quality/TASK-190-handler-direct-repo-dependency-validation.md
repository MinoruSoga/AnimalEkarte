# TASK-190: handler/validation.go — Handler が Repository を直接呼び出している（責任分離違反）

## 優先度
High

## 対象ファイル
- `backend/internal/handler/validation.go`
- `backend/internal/handler/handler.go`
- `backend/internal/service/staff_service.go`

## 問題概要
`validation.go` の `verifyStaffClinicMembership` が `h.repos.StaffClinicAssignment.ExistsByStaffAndClinic`
を Handler 層から直接呼び出している。

Handler 構造体が `repos *repository.Repositories` フィールドを持つことで、
handler → repository の直接依存が生まれており、アーキテクチャ規約
「handler はサービス経由のみでデータアクセスする」に違反している。

この直接依存は以下7箇所の Handler メソッドに波及している：
- `GetStaff`
- `GetStaffPermissionGroups`
- `SetStaffPermissionGroups`
- `GetStaffClinicAssignments`
- `SetStaffClinicAssignments`
- `GetStaffExcludedReservationTypes`
- `SetStaffExcludedReservationTypes`

## 現状コード（validation.go:46）

```go
func (h *Handler) verifyStaffClinicMembership(c *gin.Context, staffID, clinicID uint64) bool {
    exists, err := h.repos.StaffClinicAssignment.ExistsByStaffAndClinic(c.Request.Context(), staffID, clinicID)
    // ↑ repository 直接呼び出し（NG）
    ...
}
```

## あるべき姿

```go
// service/staff_service.go に追加
type StaffService interface {
    ...
    VerifyClinicMembership(ctx context.Context, staffID, clinicID uint64) error
}

// handler/validation.go
func (h *Handler) verifyStaffClinicMembership(c *gin.Context, staffID, clinicID uint64) bool {
    if err := h.svc.Staff.VerifyClinicMembership(c.Request.Context(), staffID, clinicID); err != nil {
        RespondError(c, err)
        return false
    }
    return true
}
```

`handler.go` の `Handler` 構造体から `repos` フィールドを除去できる（他に利用箇所がないことを確認の上）。

## 完了条件
- [ ] `StaffService` に `VerifyClinicMembership` メソッドを追加
- [ ] `validation.go` が `h.svc.Staff.VerifyClinicMembership` を呼ぶように修正
- [ ] `handler.go` の `repos` フィールドを除去（他の利用箇所がない場合）
- [ ] `go test ./backend/internal/...` がパス

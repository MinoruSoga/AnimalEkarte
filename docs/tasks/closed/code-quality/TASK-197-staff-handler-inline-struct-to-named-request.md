# TASK-197: staff_handler.go — 匿名インライン struct を名前付きリクエスト型に移動

## 優先度
Medium

## 対象ファイル
- `backend/internal/handler/staff_handler.go`
- `backend/internal/handler/staff_request.go`

## 問題概要
以下3メソッドがリクエスト型を匿名インライン struct で定義している。
他のすべてのハンドラは `*_request.go` に名前付き型を定義しており、このファイルのみ規約が異なる。
匿名 struct はテストコードでの型参照が困難になる。

## 該当箇所

```go
// staff_handler.go:246-249（NG）
var req struct {
    GroupIDs []uint64 `json:"group_ids"`
}

// staff_handler.go:303付近（NG）
var req struct {
    ClinicIDs []uint64 `json:"clinic_ids"`
}

// staff_handler.go:360付近（NG）
var req struct {
    ExcludedTypeIDs []uint64 `json:"excluded_type_ids"`
}
```

## あるべき姿

`staff_request.go` に名前付き型として定義：

```go
// staff_request.go
type setStaffPermissionGroupsRequest struct {
    GroupIDs []uint64 `json:"group_ids"`
}

type setStaffClinicAssignmentsRequest struct {
    ClinicIDs []uint64 `json:"clinic_ids"`
}

type setStaffExcludedReservationTypesRequest struct {
    ExcludedTypeIDs []uint64 `json:"excluded_type_ids"`
}
```

`staff_handler.go` では `var req setStaffPermissionGroupsRequest` に変更する。

## 完了条件
- [ ] 3つの匿名 struct を `staff_request.go` に名前付き型として定義
- [ ] `staff_handler.go` が名前付き型を参照するよう修正
- [ ] `go test ./backend/internal/...` がパス

# TASK-226: reservation_line_routes.go — LINE管理エンドポイント全体に RequirePermission が未設定 [CRITICAL]

## 優先度
Critical

## 対象ファイル
- `backend/internal/handler/reservation_line_routes.go`

## 問題概要
`RegisterLineReservationRoutes` は `protected` グループ（JWT認証済み）に登録されているが、
書き込み・削除系エンドポイントに `RequirePermission` ミドルウェアが**一切設定されていない**。

認証さえ通れば、権限に関係なくLINE予約設定・予約区分・スタッフ・スケジュール・予約・顧客すべてを操作できる。

## 現状コード（抜粋）

```go
func (h *Handler) RegisterLineReservationRoutes(rg *gin.RouterGroup) {
    clinics := rg.Group("/clinics/:id")

    // 全エンドポイントに RequirePermission なし
    clinics.PUT("/line-reservation-settings", h.UpsertLineReservationSetting)  // ❌

    types := clinics.Group("/reservation-types")
    types.POST("", h.CreateReservationTypeLiff)         // ❌
    types.PUT("/:id", h.UpdateReservationTypeLiff)      // ❌
    types.DELETE("/:id", h.DeleteReservationTypeLiff)   // ❌
    types.PATCH("/:id/status", ...)                     // ❌
    types.PATCH("/:id/sort-order", ...)                 // ❌
    types.POST("/:id/image", ...)                       // ❌

    staffs.POST("", h.CreateReservationStaff)           // ❌
    staffs.PUT("/:staffId", ...)                        // ❌
    staffs.DELETE("/:staffId", ...)                     // ❌
    // ...
}
```

## 比較（正しい実装例）

```go
// registerOwnerRoutesWithAuth
owners.POST("", h.RequirePermission(string(model.ResourceOwners), "edit"), h.CreateOwner)
owners.PUT("/:id", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwner)
owners.DELETE("/:id", h.RequirePermission(string(model.ResourceOwners), "delete"), h.DeleteOwner)
```

## あるべき姿

LINE管理機能に対応する `model.Resource*` 定数を確認し（または追加し）、
書き込み・削除系の全エンドポイントに `RequirePermission` を追加する。

```go
// GET は "view"、POST/PUT/PATCH は "edit"、DELETE は "delete"
types.POST("", h.RequirePermission(string(model.ResourceReservationTypes), "edit"), h.CreateReservationTypeLiff)
types.DELETE("/:id", h.RequirePermission(string(model.ResourceReservationTypes), "delete"), h.DeleteReservationTypeLiff)
```

## 完了条件
- [ ] `reservation_line_routes.go` の全書き込み・削除エンドポイントに `RequirePermission` を追加
- [ ] GET は `"view"`、POST/PUT/PATCH は `"edit"`、DELETE は `"delete"` の権限を適用
- [ ] 使用する `model.Resource*` 定数が存在することを確認（なければ追加）
- [ ] `go test ./backend/internal/...` がパス

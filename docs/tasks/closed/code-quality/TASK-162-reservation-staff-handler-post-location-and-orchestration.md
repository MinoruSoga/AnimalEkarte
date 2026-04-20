# TASK-162: reservation_staff_handler.go — POST 201 Location ヘッダー欠落 + Handler 内 orchestration

## 概要
`reservation_staff_handler.go` に 2 種類の規約違反がある。

1. **POST 201 に Location ヘッダーなし** — `CreateReservationStaff` が `http.StatusCreated` を返すが `c.Header("Location", ...)` がない。
2. **Handler 内で複数 Service 呼び出し (orchestration)** — `CreateReservationStaff`、`UpdateReservationStaff`、`PatchReservationStaffStatus` の各ハンドラで、`svc.ReservationStaff.Create/Update/PatchStatus` の後に追加で `svc.ReservationStaff.GetExcludedReservationTypes` を呼び出している。Handler は「リクエスト解析 + Service 委譲」のみが責務であり、複数 Service の orchestration は Service 層に移すべき。

## 優先度
Medium（責務分離）

## 対象ファイル
`backend/internal/handler/reservation_staff_handler.go`

---

## 問題 1: Location ヘッダー欠落

### 現状コード（行 61〜67）
```go
staff, err := h.svc.ReservationStaff.Create(c.Request.Context(), clinicID, &service.CreateReservationStaffInput{...})
if err != nil {
    RespondError(c, err)
    return
}
excluded, err := h.svc.ReservationStaff.GetExcludedReservationTypes(c.Request.Context(), staff.ID)
// ...
c.JSON(http.StatusCreated, toReservationStaffResponse(staff, excluded))
```

### 修正後コード
```go
staff, excluded, err := h.svc.ReservationStaff.Create(c.Request.Context(), clinicID, &service.CreateReservationStaffInput{...})
if err != nil {
    RespondError(c, err)
    return
}
c.Header("Location", fmt.Sprintf("/v1/reservation-staffs/%d", staff.ID))
c.JSON(http.StatusCreated, toReservationStaffResponse(staff, excluded))
```

---

## 問題 2: Handler 内 orchestration（GetExcludedReservationTypes の重複呼び出し）

`CreateReservationStaff`（行 62〜66）、`UpdateReservationStaff`（行 97〜100）、`PatchReservationStaffStatus`（行 142〜146）の3ヶ所で、処理後に `GetExcludedReservationTypes` を呼んでいる。これはレスポンス組み立てのためのデータ取得を Handler が直接行っている状態であり、Service 層の責務。

### 現状コード（CreateReservationStaff の例）
```go
func (h *Handler) CreateReservationStaff(c *gin.Context) {
    // ...
    staff, err := h.svc.ReservationStaff.Create(c.Request.Context(), clinicID, &service.CreateReservationStaffInput{...})
    if err != nil {
        RespondError(c, err)
        return
    }
    // ❌ Handler が追加の Service メソッドを呼び出している
    excluded, err := h.svc.ReservationStaff.GetExcludedReservationTypes(c.Request.Context(), staff.ID)
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusCreated, toReservationStaffResponse(staff, excluded))
}
```

### 修正方針
`Create`/`Update`/`PatchStatus` の戻り値として `(*model.Staff, []model.StaffReservationExclusion, error)` を返すか、
または excluded を含む専用のレスポンス用構造体（`StaffWithExclusions`）を返すように Service インターフェースを変更し、
Handler での追加呼び出しを廃止する。

### 修正後コード（Service インターフェース変更案）
```go
// service/reservation_staff_service.go
type ReservationStaffService interface {
    // ...
    Create(ctx context.Context, clinicID uint64, input *CreateReservationStaffInput) (*model.Staff, []model.StaffReservationExclusion, error)
    Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationStaffInput) (*model.Staff, []model.StaffReservationExclusion, error)
    PatchStatus(ctx context.Context, clinicID, id uint64, isActive bool) (*model.Staff, []model.StaffReservationExclusion, error)
    // ...
}

// サービス実装側で excluded 取得を内包
func (s *reservationStaffService) Create(ctx context.Context, clinicID uint64, input *CreateReservationStaffInput) (*model.Staff, []model.StaffReservationExclusion, error) {
    // ... 既存ロジック ...
    excluded, err := s.repo.FindExcludedReservationTypes(ctx, staff.ID)
    if err != nil {
        return nil, nil, apperrors.Wrap(err, "failed to get excluded reservation types")
    }
    return staff, excluded, nil
}
```

```go
// handler/reservation_staff_handler.go (修正後)
func (h *Handler) CreateReservationStaff(c *gin.Context) {
    // ...
    staff, excluded, err := h.svc.ReservationStaff.Create(c.Request.Context(), clinicID, &service.CreateReservationStaffInput{...})
    if err != nil {
        RespondError(c, err)
        return
    }
    c.Header("Location", fmt.Sprintf("/v1/reservation-staffs/%d", staff.ID))
    c.JSON(http.StatusCreated, toReservationStaffResponse(staff, excluded))
}
```

## 備考
`ListReservationStaffs` でも Handler 内で `ListExcludedByStaffIDs` を呼んでいるが（行 27〜30）、
List の場合は N+1 回避のためのバルク取得という性質上、Handler で staffIDs を集計してから呼ぶ必要があるため、
こちらは Service に `ListWithExclusions(ctx, clinicID)` メソッドを追加して Handler を簡潔にする対応が望ましい（任意対応）。

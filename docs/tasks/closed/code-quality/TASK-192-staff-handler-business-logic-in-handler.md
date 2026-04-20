# TASK-192: staff_handler.go — UpdateStaff のビジネスロジックが Handler に混入

## 優先度
High

## 対象ファイル
- `backend/internal/handler/staff_handler.go`
- `backend/internal/service/staff_service.go`

## 問題概要
「プロフィール更新かパスワード更新か」の判定分岐と `validatePassword` の呼び出しが
Handler 層（`staff_handler.go:116-168` および `validators.go`）に実装されている。

パスワードのバリデーションルール（複雑性チェック等）はビジネスロジックであり、
Service 層に閉じるべきである。

## 現状コード（staff_handler.go:116-168 の概略）

```go
// Handler 層でビジネス判定（NG）
hasProfileUpdate := req.Name != nil || req.LicenseNumber != nil || ...
hasPasswordUpdate := req.Password != nil && *req.Password != ""
if !hasProfileUpdate && !hasPasswordUpdate {
    RespondError(c, apperrors.WrapInvalidInput("at least one field must be provided"))
    return
}
if hasPasswordUpdate {
    if err := validatePassword(*req.Password); err != nil { ... }
    if err := h.svc.Staff.UpdatePassword(...); err != nil { ... }
}
if hasProfileUpdate {
    if err := h.svc.Staff.Update(...); err != nil { ... }
}
```

## あるべき姿

```go
// service/staff_service.go の UpdateStaffInput に Password フィールドを追加
type UpdateStaffInput struct {
    Name          *string
    LicenseNumber *string
    // ... 既存フィールド ...
    Password      *string  // 追加: nil なら更新なし
}

// staff_service.go の Update 内でパスワード処理を一括
func (s *staffService) Update(ctx context.Context, ..., input UpdateStaffInput) (*model.Staff, error) {
    if input.Password != nil && *input.Password != "" {
        if err := validatePassword(*input.Password); err != nil {
            return nil, apperrors.WrapInvalidInput(err.Error())
        }
        // ハッシュ化・保存
    }
    // プロフィール更新
    ...
}

// handler 層はシンプルに
func (h *Handler) UpdateStaff(c *gin.Context) {
    ...
    result, err := h.svc.Staff.Update(ctx, clinicID, id, input)
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, toStaffResponse(result))
}
```

## 完了条件
- [ ] `UpdateStaffInput` に `Password *string` フィールドを追加
- [ ] `StaffService.Update` 内でパスワードバリデーション・ハッシュ化を処理
- [ ] `staff_handler.go` から `validatePassword` 呼び出しと分岐ロジックを除去
- [ ] `go test ./backend/internal/...` がパス

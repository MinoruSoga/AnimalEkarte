# TASK-035: スケジューリング・LINE系 HIGH 問題 3件

## 優先度

HIGH

---

## 問題 1: shift_handler / shift_template_handler で model.ShiftType 変換をハンドラ内で実施

### ファイル
- `backend/internal/handler/shift_handler.go:75`
- `backend/internal/handler/shift_handler.go:112`

### 問題
```go
// CreateShiftEntry (L75)
ShiftType: model.ShiftType(req.ShiftType),

// UpdateShiftEntry (L112)
st := model.ShiftType(*req.ShiftType)
input.ShiftType = &st
```
`model.ShiftType` へのキャスト・バリデーションをハンドラ内で行っている。不正な文字列（例: `"invalid_type"`）が渡された場合、GORM が実行時エラーを返すまで検出できない。バリデーションは service 層の `validators.go` または service 内で実施すべきである。

### 修正案
```go
// service/shift_service.go — Input DTO の ShiftType を string のまま受け取り、service 内でバリデーション
type CreateShiftEntryInput struct {
    StaffID   uint64
    Date      time.Time
    ShiftType string // string のまま受け取る
    // ...
}

func validateShiftType(s string) (model.ShiftType, error) {
    switch model.ShiftType(s) {
    case model.ShiftTypeFullDay, model.ShiftTypeMorning, model.ShiftTypeAfternoon, model.ShiftTypeOff:
        return model.ShiftType(s), nil
    default:
        return "", apperrors.WrapInvalidInput(fmt.Sprintf("invalid shift_type: %s", s))
    }
}
```

---

## 問題 2: line_customer_service の LinkOwner に slog.InfoContext なし

### ファイル
`backend/internal/service/line_customer_service.go:33-42`

### 問題
```go
func (s *lineCustomerService) LinkOwner(ctx context.Context, clinicID, id uint64, ownerID *uint64) (*model.LineCustomer, error) {
    if err := s.repo.UpdateOwnerLink(ctx, clinicID, id, ownerID); err != nil {
        return nil, apperrors.Wrap(err, "failed to link owner to reservation customer")
    }
    result, err := s.repo.FindByID(ctx, clinicID, id)
    // ...
}
```
LINE 顧客とオーナーの紐付けは重要な業務操作（監査が必要）にもかかわらず `slog.InfoContext` がない。`List` はログ不要だが、`LinkOwner` は mutation として必須。

### 修正案
```go
func (s *lineCustomerService) LinkOwner(ctx context.Context, clinicID, id uint64, ownerID *uint64) (*model.LineCustomer, error) {
    if err := s.repo.UpdateOwnerLink(ctx, clinicID, id, ownerID); err != nil {
        return nil, apperrors.Wrap(err, "failed to link owner to reservation customer")
    }
    slog.InfoContext(ctx, "line_customer owner linked",
        slog.Uint64("clinic_id", clinicID),
        slog.Uint64("line_customer_id", id),
        slog.Any("owner_id", ownerID))
    result, err := s.repo.FindByID(ctx, clinicID, id)
    // ...
}
```

---

## 問題 3: liff_auth.go ミドルウェアが c.JSON を直接使用

### ファイル
`backend/internal/middleware/liff_auth.go`（複数箇所）

### 問題
```go
// ミドルウェア内で c.JSON を直接呼んでいる
c.JSON(http.StatusUnauthorized, gin.H{"error": "..."})
c.Abort()
```
ハンドラ規約では `RespondError(c, apperrors.WrapXxx(...))` を使う決まりだが、ミドルウェアは `handler` パッケージに属さないため `RespondError` を呼べない循環参照問題がある。

現状の回避策として直接 `c.JSON` を使っているが、エラーレスポンスフォーマット（`code`, `message`, `timestamp` フィールド）が統一されておらず、クライアントのエラーハンドリングが複雑になる。

### 修正案
`RespondError` と同等のレスポンス生成関数を `internal/handler/response.go` から独立した `internal/httputil/respond.go` に切り出し、ミドルウェア・ハンドラ両方から参照可能にする。

```go
// internal/httputil/respond.go
package httputil

func RespondError(c *gin.Context, err error) {
    // 現 handler.RespondError と同実装
}
```

または、最低限、ミドルウェアのエラーレスポンスを handler 層と同一の JSON 構造（`{"code": "...", "message": "...", "timestamp": "..."}`）に合わせる。

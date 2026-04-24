# TASK-225: reservation_type_handler.go — CreateUnavailableTime の相互依存フィールドバリデーション欠如

## 優先度
Low

## 対象ファイル
- `backend/internal/handler/reservation_type_handler.go`
- `backend/internal/handler/reservation_type_request.go`

## 問題概要
予約種別の利用不可時間（UnavailableTime）には以下のビジネスルールがある。

- `unavailable_type == "weekly"` → `day_of_week` フィールドが必須
- `unavailable_type == "specific"` → `specific_date` フィールドが必須

この相互依存バリデーションが `binding` タグでは表現できず、
handler でも service でも明示的なチェックが実装されていない。
不正な組み合わせ（`weekly` なのに `day_of_week` なし）が DB に保存される可能性がある。

## あるべき姿（handler 層でのチェック）

```go
// CreateUnavailableTime handler 内
func (h *Handler) CreateUnavailableTime(c *gin.Context) {
    var req createUnavailableTimeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
        return
    }

    // 相互依存チェック（追加）
    switch req.UnavailableType {
    case "weekly":
        if req.DayOfWeek == nil {
            RespondError(c, apperrors.WrapInvalidInput("weekly タイプでは day_of_week が必要です"))
            return
        }
    case "specific":
        if req.SpecificDate == nil {
            RespondError(c, apperrors.WrapInvalidInput("specific タイプでは specific_date が必要です"))
            return
        }
    }
    ...
}
```

または service 層の `CreateUnavailableTimeInput` の バリデーションとして実装してもよい。

## 完了条件
- [ ] `CreateUnavailableTime` に `unavailable_type` と依存フィールドの相互検証を追加
- [ ] 不正な組み合わせで 400 が返ることをテストで確認
- [ ] `go test ./backend/internal/...` がパス

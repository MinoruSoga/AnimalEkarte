# BUG-412: examination_handler で Enum バリデーションが二重実施されている

## 概要

`examination_handler.go` の Update/Create エンドポイントで、`status` フィールドの Enum 検証が
Request struct の `binding` タグと `validateEnum()` 関数呼び出しの両方で重複実施されている。
冗長な処理であり、一方に統一すべきである。

## 問題箇所

```go
// examination_request.go
type updateExaminationRequest struct {
    Status string `json:"status" binding:"omitempty,oneof=pending in_progress result_entered completed confirmed"`
    // ↑ binding タグで既に検証済み
}

// examination_handler.go:104-116（Update）
s, err := validateEnum(input.Status,
    model.ExaminationStatusPending,
    model.ExaminationStatusInProgress,
    model.ExaminationStatusResultEntered,
    model.ExaminationStatusCompleted,
    model.ExaminationStatusConfirmed,
)
if err != nil {
    RespondError(c, apperrors.WrapInvalidInput("invalid status: "+err.Error()))
    return
}
// ↑ binding タグと同じ検証を手動で再実施
```

同様に行 154-166（Create）も重複。

## 期待する実装

`binding` タグによる自動検証のみで完結させる。`validateEnum()` の手動呼び出しを削除。

```go
// examination_request.go（変更なし、これで十分）
type updateExaminationRequest struct {
    Status string `json:"status" binding:"omitempty,oneof=pending in_progress result_entered completed confirmed"`
}

// examination_handler.go（validateEnum 呼び出しを削除）
if err := c.ShouldBindJSON(&input); err != nil {
    RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
    return
}
// ← ここで binding タグが status の oneof を検証済み。手動 validateEnum 不要
svcInput := service.UpdateExaminationInput{...}
```

## 影響ファイル

- `backend/internal/handler/examination_handler.go` — 行 104-116, 154-166

## 優先度

**Low** — 冗長処理。動作に影響はないが保守性低下につながる。

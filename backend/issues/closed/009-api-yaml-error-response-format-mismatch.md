---
status: open
---

# [api.yaml] ErrorResponse スキーマが実装と乖離している

## 背景

api.yaml で ErrorResponse を `{code, message, details}` 構造として定義しているが、
実装（`handler/response.go` の `RespondError`）は `{error: string}` の単純形式で返している。
Swagger UI を見た開発者が実際と異なるスキーマを参照してしまう状態。

## 問題

```yaml
# api.yaml L26
ErrorResponse:
  type: object
  required: [code, message]
  properties:
    code:
      type: string
      enum: [VALIDATION_ERROR, NOT_FOUND, CONFLICT, UNAUTHORIZED, FORBIDDEN, INTERNAL_ERROR]
    message:
      type: string
    details:
      type: array
      items:
        type: object
```

```go
// handler/response.go: 実装（実際に返るレスポンス）
c.JSON(http.StatusNotFound,    gin.H{"error": "owner not found: 999"})
c.JSON(http.StatusBadRequest,  gin.H{"error": "..."})
c.JSON(http.StatusConflict,    gin.H{"error": "..."})
c.JSON(http.StatusUnauthorized,gin.H{"error": "unauthorized"})
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
```

**実際のレスポンス**:
```json
{"error": "owner not found: 999"}
```

**yaml が示すレスポンス**:
```json
{"code": "NOT_FOUND", "message": "owner not found: 999", "details": []}
```

## 修正方針

2つの選択肢がある。どちらを選ぶかは設計判断。

### 案A: yaml を実装に合わせる（推奨・工数小）

```yaml
ErrorResponse:
  type: object
  required: [error]
  properties:
    error:
      type: string
      description: エラーメッセージ
      example: "owner not found: 999"
```

### 案B: 実装を yaml に合わせる（工数大・API品質向上）

```go
type errorResponse struct {
    Code    string        `json:"code"`
    Message string        `json:"message"`
    Details []errorDetail `json:"details,omitempty"`
}

func RespondError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, apperrors.ErrNotFound):
        c.JSON(http.StatusNotFound, errorResponse{
            Code:    "NOT_FOUND",
            Message: extractMessage(err),
        })
    // ...
    }
}
```

案Bはフロントエンドのエラーハンドリング（`axios.ts`）も変更が必要になる。

**推奨**: まず案Aで yaml を実装に合わせてドキュメントの正確性を確保する。
案Bへの移行は別途チケットとして起票する。

## 完了条件

- [ ] `ErrorResponse` スキーマが実装の `{error: string}` と一致している（案A）
- [ ] 全エンドポイントのエラーレスポンス参照が更新されている
- [ ] Swagger UI でエラーレスポンス例が実際のレスポンスと一致する

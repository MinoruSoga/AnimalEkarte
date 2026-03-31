# BE-030: リクエスト構造体の binding タグ欠落

## 問題
複数の `*_request.go` ファイルで `binding:"required"` タグが欠落しており、
無効なリクエストが service 層まで到達する可能性がある。

## 影響ファイル
- `handler/chief_complaint_request.go` — Name フィールドに binding なし
- `handler/consultation_request.go` — 必須フィールドに binding なし
- `handler/job_title_request.go` — binding タグなし
- `handler/estimate_request.go` — 一部フィールドにタグなし
- `handler/clinical_plan_request.go` — binding タグなし
- `handler/inquiry_template_request.go` — binding タグなし
- `handler/checkup_request.go` — 必須フィールド未検証
- `handler/company_request.go` — binding タグなし

## 参照実装（正しいパターン）
```go
// handler/staff_request.go
type createStaffRequest struct {
    Name     string `json:"name"     binding:"required"`
    Role     string `json:"role"     binding:"required"`
    Email    string `json:"email"    binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}
```

## 修正方針
1. 各リクエスト構造体の必須フィールドに `binding:"required"` を追加
2. 文字列長・形式バリデーションも `binding:"min=1,max=255"` 等で追加
3. `ShouldBindJSON` でバインドエラーをキャッチし `RespondError` に渡す

## 優先度
MEDIUM（入力バリデーション漏れ）

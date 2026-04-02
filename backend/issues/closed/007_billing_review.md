# 会計医師確認（BillingReview）API 実装

## 概要
外来カルテの会計を医師が確認・差し戻しするワークフロー API を実装する。
`model/billing_review.go` の `BillingReview` struct は実装済み。handler・service・repository が未実装。

**依存**: `005_medical_record_treatments.md`（治療項目が存在することが前提）
**ブロック**: accounting（会計）フローはこの確認ステータスに依存する。

## 優先度
high

## 関連テーブル
- `billing_reviews` (id, medical_record_id NOT NULL UNIQUE, status billing_review_status DEFAULT 'pending', confirmed_by, confirmed_at, returned_by, returned_at, return_reason text DEFAULT '', memo text DEFAULT '', created_at, updated_at)
  - `billing_review_status` enum: `pending` / `confirmed` / `returned`
- `medical_records` (親テーブル)
- `staffs` (confirmed_by / returned_by の参照先)

## 実装内容

### モデル
`model/billing_review.go` は実装済み。変更不要。

```go
type BillingReviewStatus string
const (
    BillingReviewStatusPending   BillingReviewStatus = "pending"
    BillingReviewStatusConfirmed BillingReviewStatus = "confirmed"
    BillingReviewStatusReturned  BillingReviewStatus = "returned"
)
```

### リポジトリ
新規ファイル `repository/billing_review_repository.go`:
```go
type BillingReviewRepository interface {
    FindByMedicalRecordID(ctx context.Context, medicalRecordID uint64) (*model.BillingReview, error)
    Create(ctx context.Context, review *model.BillingReview) error
    Update(ctx context.Context, id uint64, fields map[string]any) error
}
```
- `FindByMedicalRecordID`: `billing_reviews` は medical_record_id に UNIQUE 制約があるため 1 件のみ返す。存在しない場合は `apperrors.WrapNotFound` を返す。
- `Create`: 初回作成（status = pending）
- `Update`: `map[string]any` で confirmed_by, confirmed_at, returned_by, returned_at, return_reason, status を更新

`repository/repositories.go` に `BillingReview BillingReviewRepository` を追加。

### サービス
新規ファイル `service/billing_review_service.go`:
```go
type ConfirmBillingReviewInput struct {
    ConfirmedBy uint64  // staff_id（JWT から取得）
    Memo        string
}

type ReturnBillingReviewInput struct {
    ReturnedBy   uint64  // staff_id
    ReturnReason string
    Memo         string
}

type BillingReviewService interface {
    GetOrCreate(ctx context.Context, medicalRecordID uint64) (*model.BillingReview, error)
    Confirm(ctx context.Context, medicalRecordID uint64, input *ConfirmBillingReviewInput) (*model.BillingReview, error)
    Return(ctx context.Context, medicalRecordID uint64, input *ReturnBillingReviewInput) (*model.BillingReview, error)
}
```

- `GetOrCreate`: レコードが存在しない場合は status=pending で新規作成して返す
- `Confirm`: status を `confirmed` に変更し `confirmed_by`、`confirmed_at = now()` を設定
- `Return`: status を `returned` に変更し `returned_by`、`returned_at = now()`、`return_reason` を設定
- すでに `confirmed` のレコードを再度 `Confirm` しようとした場合は `apperrors.WrapInvalidInput("already confirmed")` を返す

### ハンドラ
新規ファイル `handler/billing_review_handler.go`:
```go
func (h *Handler) GetBillingReview(c *gin.Context)
func (h *Handler) ConfirmBillingReview(c *gin.Context)
func (h *Handler) ReturnBillingReview(c *gin.Context)
func (h *Handler) RegisterBillingReviewRoutes(rg *gin.RouterGroup)
```

新規ファイル `handler/billing_review_request.go`:
```go
type confirmBillingReviewRequest struct {
    ConfirmedBy uint64 `json:"confirmed_by" binding:"required"`
    Memo        string `json:"memo"`
}

type returnBillingReviewRequest struct {
    ReturnedBy   uint64 `json:"returned_by"   binding:"required"`
    ReturnReason string `json:"return_reason" binding:"required"`
    Memo         string `json:"memo"`
}
```

新規ファイル `handler/billing_review_response.go`:
```go
type billingReviewResponse struct {
    ID              uint64      `json:"id"`
    MedicalRecordID uint64      `json:"medical_record_id"`
    Status          string      `json:"status"`
    ConfirmedBy     *uint64     `json:"confirmed_by,omitempty"`
    ConfirmedAt     *time.Time  `json:"confirmed_at,omitempty"`
    ReturnedBy      *uint64     `json:"returned_by,omitempty"`
    ReturnedAt      *time.Time  `json:"returned_at,omitempty"`
    ReturnReason    string      `json:"return_reason"`
    Memo            string      `json:"memo"`
    CreatedAt       time.Time   `json:"created_at"`
    UpdatedAt       time.Time   `json:"updated_at"`
    ConfirmedStaff  *staffSummaryResponse `json:"confirmed_staff,omitempty"`
    ReturnedStaff   *staffSummaryResponse `json:"returned_staff,omitempty"`
}
```

### ルート登録
`cmd/api/main.go` の医療記録グループに追加:
```go
medRecords.GET("/:id/billing-review",         h.GetBillingReview)
medRecords.POST("/:id/billing-review/confirm", h.ConfirmBillingReview)
medRecords.POST("/:id/billing-review/return",  h.ReturnBillingReview)
```

## 完了条件
- `GET /v1/medical-records/:id/billing-review` が確認状態を返す（存在しない場合は pending で自動作成）
- `POST /v1/medical-records/:id/billing-review/confirm` で確認済みに変更できる
- `POST /v1/medical-records/:id/billing-review/return` で差し戻しできる（`return_reason` 必須）
- `confirmed` 状態を再度 `confirm` しようとすると 400 エラーを返す
- 存在しない `medical_record_id` に対しては 404 を返す

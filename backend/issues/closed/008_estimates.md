# 見積書 CRUD 実装

## 概要
見積書（Estimate）およびその明細（EstimateItem）の CRUD API を実装する。
`model/estimate.go` の `Estimate` / `EstimateItem` struct は実装済み。handler・service・repository が未実装。

## 優先度
medium

## 関連テーブル
- `estimates` (id, clinic_id NOT NULL, estimate_no text NOT NULL DEFAULT '', medical_record_id, title text DEFAULT '', owner_id, status estimate_status DEFAULT 'draft', subtotal numeric(10,2) DEFAULT 0, tax_total numeric(10,2) DEFAULT 0, total_amount numeric(10,2) DEFAULT 0, insurance_amount numeric(10,2) DEFAULT 0, discount_amount numeric(10,2) DEFAULT 0, valid_until date, comment text DEFAULT '', notes text DEFAULT '', created_by, deleted_at, created_at, updated_at)
  - `estimate_status` enum: `draft` / `sent` / `approved` / `rejected`
- `estimate_items` (id, estimate_id NOT NULL, name text NOT NULL DEFAULT '', category item_category NOT NULL, unit_price numeric(10,2) DEFAULT 0, quantity int DEFAULT 1, tax_rate numeric(3,2) DEFAULT 0.10, discount_rate numeric(5,2) DEFAULT 0, discount_amount numeric(10,2) DEFAULT 0, is_insurance_applicable bool DEFAULT false, consultation_id, procedure_id, medicine_id, sort_order int DEFAULT 0, created_at)
  - `item_category` enum: `consultation` / `procedure` / `medicine` / `other`（確認要）

## 実装内容

### モデル
`model/estimate.go` は実装済み。変更不要。

`EstimateStatus`: `draft` / `sent` / `approved` / `rejected`

### リポジトリ
新規ファイル `repository/estimate_repository.go`:
```go
type EstimateRepository interface {
    List(ctx context.Context, clinicID uint64, filter EstimateFilter) ([]model.Estimate, error)
    FindByID(ctx context.Context, clinicID, id uint64) (*model.Estimate, error)
    Create(ctx context.Context, estimate *model.Estimate) error
    Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
    Delete(ctx context.Context, clinicID, id uint64) error
    ListItems(ctx context.Context, estimateID uint64) ([]model.EstimateItem, error)
    CreateItem(ctx context.Context, item *model.EstimateItem) error
    UpdateItem(ctx context.Context, id uint64, fields map[string]any) error
    DeleteItem(ctx context.Context, id uint64) error
    FindItemByID(ctx context.Context, id uint64) (*model.EstimateItem, error)
}

type EstimateFilter struct {
    OwnerID         *uint64
    MedicalRecordID *uint64
    Status          *model.EstimateStatus
}
```
- `List`: `Preload(Owner)` で Owner 情報付き返却、`ORDER BY created_at DESC`
- `FindByID`: `Preload(Items)` で明細込み返却、clinic_id スコープ
- `Delete`: soft delete（`deleted_at` 付き）
- `DeleteItem`: 物理削除（`estimate_items` には deleted_at なし）

`repository/repositories.go` に `Estimate EstimateRepository` を追加。

### サービス
新規ファイル `service/estimate_service.go`:
```go
type CreateEstimateInput struct {
    EstimateNo      string
    MedicalRecordID *uint64
    Title           string
    OwnerID         *uint64
    Status          model.EstimateStatus
    Subtotal        float64
    TaxTotal        float64
    TotalAmount     float64
    InsuranceAmount float64
    DiscountAmount  float64
    ValidUntil      *time.Time
    Comment         string
    Notes           string
    CreatedBy       *uint64
}

type UpdateEstimateInput struct {
    EstimateNo      *string
    Title           *string
    Status          *model.EstimateStatus
    Subtotal        *float64
    TaxTotal        *float64
    TotalAmount     *float64
    InsuranceAmount *float64
    DiscountAmount  *float64
    ValidUntil      *time.Time
    Comment         *string
    Notes           *string
}

type CreateEstimateItemInput struct {
    Name                  string
    Category              model.ItemCategory
    UnitPrice             float64
    Quantity              int
    TaxRate               float64
    DiscountRate          float64
    DiscountAmount        float64
    IsInsuranceApplicable bool
    ConsultationID        *uint64
    ProcedureID           *uint64
    MedicineID            *uint64
    SortOrder             int
}

type UpdateEstimateItemInput struct {
    Name                  *string
    Category              *model.ItemCategory
    UnitPrice             *float64
    Quantity              *int
    TaxRate               *float64
    DiscountRate          *float64
    DiscountAmount        *float64
    IsInsuranceApplicable *bool
    ConsultationID        *uint64
    ProcedureID           *uint64
    MedicineID            *uint64
    SortOrder             *int
}
```

`buildEstimateUpdateFields()` および `buildEstimateItemUpdateFields()` を実装する。
`service/validators.go` に `validateEstimateStatus` を追加。

### ハンドラ
新規ファイル `handler/estimate_handler.go`:
```go
func (h *Handler) ListEstimates(c *gin.Context)
func (h *Handler) CreateEstimate(c *gin.Context)
func (h *Handler) GetEstimate(c *gin.Context)
func (h *Handler) UpdateEstimate(c *gin.Context)
func (h *Handler) DeleteEstimate(c *gin.Context)
func (h *Handler) ListEstimateItems(c *gin.Context)
func (h *Handler) CreateEstimateItem(c *gin.Context)
func (h *Handler) UpdateEstimateItem(c *gin.Context)
func (h *Handler) DeleteEstimateItem(c *gin.Context)
func (h *Handler) RegisterEstimateRoutes(rg *gin.RouterGroup)
```

新規ファイル `handler/estimate_request.go` / `handler/estimate_response.go` を作成する。

`estimateResponse` には `items []estimateItemResponse` を含め、`GET /v1/estimates/:id` でまとめて返す。

### ルート登録
`cmd/api/main.go` に追加:
```go
estimates := v1.Group("/estimates", authMiddleware)
estimates.GET("",    h.ListEstimates)
estimates.POST("",   h.CreateEstimate)
estimates.GET("/:id",    h.GetEstimate)
estimates.PATCH("/:id",  h.UpdateEstimate)
estimates.DELETE("/:id", h.DeleteEstimate)

estimates.GET("/:id/items",              h.ListEstimateItems)
estimates.POST("/:id/items",             h.CreateEstimateItem)
estimates.PATCH("/:id/items/:itemId",    h.UpdateEstimateItem)
estimates.DELETE("/:id/items/:itemId",   h.DeleteEstimateItem)
```

## 完了条件
- `GET /v1/estimates` がクリニックの見積書一覧を返す（`owner_id`, `medical_record_id`, `status` でフィルタ可能）
- `POST /v1/estimates` で見積書を作成できる
- `GET /v1/estimates/:id` で見積書詳細（明細込み）を返す
- `PATCH /v1/estimates/:id` で見積書を部分更新できる
- `DELETE /v1/estimates/:id` で soft delete できる
- 明細（`/items`）の CRUD が動作する
- 他クリニックの見積書は参照・変更できない（clinic_id スコープ）

# 治療計画（TreatmentPlan）CRUD 実装

## 概要
外来カルテまたは入院に紐づく治療計画（TreatmentPlan）の CRUD API を実装する。
`model/hospitalization.go` の `TreatmentPlan` struct は実装済み。handler・service・repository が未実装。

## 優先度
medium

## 関連テーブル
- `treatment_plans` (id, medical_record_id nullable, hospitalization_id nullable, treatment_content text NOT NULL DEFAULT '', memo text DEFAULT '', insurance bool DEFAULT false, unit_price numeric(10,2) DEFAULT 0, quantity int DEFAULT 1, discount_rate numeric(5,2) DEFAULT 0, discount_amount numeric(10,2) DEFAULT 0, subtotal numeric(10,2) DEFAULT 0, sort_order int DEFAULT 0, deleted_at, created_at, updated_at)
- `medical_records` (外来紐づけ、nullable)
- `hospitalizations` (入院紐づけ、nullable)

備考: `medical_record_id` と `hospitalization_id` のどちらか一方のみを持つ（両方 null も可）。

## 実装内容

### モデル
`model/hospitalization.go` の `TreatmentPlan` は実装済み。変更不要。

### リポジトリ
新規ファイル `repository/treatment_plan_repository.go`:
```go
type TreatmentPlanRepository interface {
    ListByMedicalRecordID(ctx context.Context, medicalRecordID uint64) ([]model.TreatmentPlan, error)
    ListByHospitalizationID(ctx context.Context, hospitalizationID uint64) ([]model.TreatmentPlan, error)
    FindByID(ctx context.Context, id uint64) (*model.TreatmentPlan, error)
    Create(ctx context.Context, plan *model.TreatmentPlan) error
    Update(ctx context.Context, id uint64, fields map[string]any) error
    Delete(ctx context.Context, id uint64) error
}
```
- `ListBy*`: `ORDER BY sort_order ASC`
- `Delete`: soft delete（`deleted_at` 付き）

`repository/repositories.go` に `TreatmentPlan TreatmentPlanRepository` を追加。

### サービス
新規ファイル `service/treatment_plan_service.go`:
```go
type CreateTreatmentPlanInput struct {
    TreatmentContent string
    Memo             string
    Insurance        bool
    UnitPrice        float64
    Quantity         int
    DiscountRate     float64
    DiscountAmount   float64
    Subtotal         float64
    SortOrder        int
}

type UpdateTreatmentPlanInput struct {
    TreatmentContent *string
    Memo             *string
    Insurance        *bool
    UnitPrice        *float64
    Quantity         *int
    DiscountRate     *float64
    DiscountAmount   *float64
    Subtotal         *float64
    SortOrder        *int
}

type TreatmentPlanService interface {
    ListByMedicalRecord(ctx context.Context, medicalRecordID uint64) ([]model.TreatmentPlan, error)
    ListByHospitalization(ctx context.Context, hospitalizationID uint64) ([]model.TreatmentPlan, error)
    Create(ctx context.Context, medicalRecordID, hospitalizationID *uint64, input *CreateTreatmentPlanInput) (*model.TreatmentPlan, error)
    Update(ctx context.Context, id uint64, input *UpdateTreatmentPlanInput) (*model.TreatmentPlan, error)
    Delete(ctx context.Context, id uint64) error
}
```

`buildTreatmentPlanUpdateFields()` を実装する。
`Create` で `subtotal` が 0 の場合は `unit_price * quantity * (1 - discount_rate/100) - discount_amount` として自動計算する（サービス層で計算）。

### ハンドラ
新規ファイル `handler/treatment_plan_handler.go`:
```go
// 外来カルテ経由
func (h *Handler) ListTreatmentPlansByMedicalRecord(c *gin.Context)
func (h *Handler) CreateTreatmentPlanForMedicalRecord(c *gin.Context)
func (h *Handler) UpdateTreatmentPlan(c *gin.Context)
func (h *Handler) DeleteTreatmentPlan(c *gin.Context)

// 入院経由
func (h *Handler) ListTreatmentPlansByHospitalization(c *gin.Context)
func (h *Handler) CreateTreatmentPlanForHospitalization(c *gin.Context)

func (h *Handler) RegisterTreatmentPlanRoutes(rg *gin.RouterGroup)
```

新規ファイル `handler/treatment_plan_request.go`:
```go
type createTreatmentPlanRequest struct {
    TreatmentContent string  `json:"treatment_content" binding:"required"`
    Memo             string  `json:"memo"`
    Insurance        bool    `json:"insurance"`
    UnitPrice        float64 `json:"unit_price"`
    Quantity         int     `json:"quantity"`
    DiscountRate     float64 `json:"discount_rate"`
    DiscountAmount   float64 `json:"discount_amount"`
    Subtotal         float64 `json:"subtotal"`
    SortOrder        int     `json:"sort_order"`
}

type updateTreatmentPlanRequest struct {
    TreatmentContent *string  `json:"treatment_content"`
    Memo             *string  `json:"memo"`
    Insurance        *bool    `json:"insurance"`
    UnitPrice        *float64 `json:"unit_price"`
    Quantity         *int     `json:"quantity"`
    DiscountRate     *float64 `json:"discount_rate"`
    DiscountAmount   *float64 `json:"discount_amount"`
    Subtotal         *float64 `json:"subtotal"`
    SortOrder        *int     `json:"sort_order"`
}
```

新規ファイル `handler/treatment_plan_response.go` で `treatmentPlanResponse` と変換関数を実装。

### ルート登録
`cmd/api/main.go` に追加:
```go
// 外来カルテ経由
medRecords.GET("/:id/treatment-plans",              h.ListTreatmentPlansByMedicalRecord)
medRecords.POST("/:id/treatment-plans",             h.CreateTreatmentPlanForMedicalRecord)
medRecords.PATCH("/:id/treatment-plans/:planId",    h.UpdateTreatmentPlan)
medRecords.DELETE("/:id/treatment-plans/:planId",   h.DeleteTreatmentPlan)

// 入院経由
hosps.GET("/:id/treatment-plans",                   h.ListTreatmentPlansByHospitalization)
hosps.POST("/:id/treatment-plans",                  h.CreateTreatmentPlanForHospitalization)
```

## 完了条件
- `GET /v1/medical-records/:id/treatment-plans` が治療計画一覧を `sort_order` 昇順で返す
- `POST /v1/medical-records/:id/treatment-plans` で治療計画を作成できる（`treatment_content` 必須）
- `PATCH /v1/medical-records/:id/treatment-plans/:planId` で部分更新できる
- `DELETE /v1/medical-records/:id/treatment-plans/:planId` で soft delete できる
- 入院経由の GET / POST も同様に動作する
- `subtotal` が 0 の場合は自動計算される

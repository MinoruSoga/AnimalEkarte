# 外来治療項目 CRUD 実装

## 概要
外来カルテの治療項目（診療費明細）の CRUD API を実装する。会計フローのコアとなる機能。
`model/treatment.go` の `Treatment` struct は実装済み。handler・service・repository が未実装。

**依存**: `007_billing_review.md`（会計確認フロー）の前提となる。

## 優先度
high（会計フローのコア）

## 関連テーブル
- `treatments` (id, medical_record_id NOT NULL, item_type treatment_item_type NOT NULL DEFAULT 'other', consultation_id, procedure_id, medicine_id, inventory_id, unit_price numeric(10,2) DEFAULT 0, quantity int DEFAULT 1, selected bool DEFAULT false, status treatment_status DEFAULT '未完了', content text NOT NULL DEFAULT '', memo text DEFAULT '', insurance bool DEFAULT false, discount_rate numeric(5,2) DEFAULT 0, discount_amount numeric(10,2) DEFAULT 0, sort_order int DEFAULT 0, deleted_at, created_at, updated_at)
- `medical_records` (親テーブル)
- `consultations`, `procedures`, `medicines`, `inventory_items` (参照先マスタ)

## 実装内容

### モデル
`model/treatment.go` は実装済み。変更不要。

定数:
- `TreatmentItemType`: `consultation` / `procedure` / `medicine` / `other`
- `TreatmentStatus`: `未完了` / `完了` / `-`

### リポジトリ
新規ファイル `repository/treatment_repository.go`:
```go
type TreatmentRepository interface {
    ListByMedicalRecordID(ctx context.Context, medicalRecordID uint64) ([]model.Treatment, error)
    Create(ctx context.Context, treatment *model.Treatment) error
    Update(ctx context.Context, id uint64, fields map[string]any) error
    Delete(ctx context.Context, id uint64) error
    FindByID(ctx context.Context, id uint64) (*model.Treatment, error)
    BulkUpdateSortOrder(ctx context.Context, updates []TreatmentSortUpdate) error
}

type TreatmentSortUpdate struct {
    ID        uint64
    SortOrder int
}
```
- `ListByMedicalRecordID`: `ORDER BY sort_order ASC` で Preload(Consultation, Procedure, Medicine, Inventory)
- `Delete`: soft delete（`deleted_at` 付き）
- `BulkUpdateSortOrder`: トランザクションで一括更新（`PUT` 並び替え用）

`repository/repositories.go` に `Treatment TreatmentRepository` を追加。

### サービス
新規ファイル `service/treatment_service.go`:
```go
type CreateTreatmentInput struct {
    ItemType       model.TreatmentItemType
    ConsultationID *uint64
    ProcedureID    *uint64
    MedicineID     *uint64
    InventoryID    *uint64
    UnitPrice      float64
    Quantity       int
    Selected       bool
    Status         model.TreatmentStatus
    Content        string
    Memo           string
    Insurance      bool
    DiscountRate   float64
    DiscountAmount float64
    SortOrder      int
}

type UpdateTreatmentInput struct {
    ItemType       *model.TreatmentItemType
    ConsultationID *uint64
    ProcedureID    *uint64
    MedicineID     *uint64
    InventoryID    *uint64
    UnitPrice      *float64
    Quantity       *int
    Selected       *bool
    Status         *model.TreatmentStatus
    Content        *string
    Memo           *string
    Insurance      *bool
    DiscountRate   *float64
    DiscountAmount *float64
    SortOrder      *int
}

type BulkUpdateTreatmentsInput struct {
    Treatments []BulkTreatmentItem
}

type BulkTreatmentItem struct {
    ID        uint64
    SortOrder int
}
```

`buildTreatmentUpdateFields(input *UpdateTreatmentInput) map[string]any` を実装する。
`service/validators.go` に `validateTreatmentItemType` を追加。

### ハンドラ
新規ファイル `handler/treatment_handler.go`:
```go
func (h *Handler) ListTreatments(c *gin.Context)
func (h *Handler) CreateTreatment(c *gin.Context)
func (h *Handler) UpdateTreatment(c *gin.Context)
func (h *Handler) DeleteTreatment(c *gin.Context)
func (h *Handler) BulkUpdateTreatments(c *gin.Context)  // PUT 並び替え・一括更新
func (h *Handler) RegisterTreatmentRoutes(rg *gin.RouterGroup)
```

新規ファイル `handler/treatment_request.go`:
```go
type createTreatmentRequest struct {
    ItemType       string  `json:"item_type"       binding:"required"`
    ConsultationID *uint64 `json:"consultation_id"`
    ProcedureID    *uint64 `json:"procedure_id"`
    MedicineID     *uint64 `json:"medicine_id"`
    InventoryID    *uint64 `json:"inventory_id"`
    UnitPrice      float64 `json:"unit_price"`
    Quantity       int     `json:"quantity"`
    Selected       bool    `json:"selected"`
    Status         string  `json:"status"`
    Content        string  `json:"content"`
    Memo           string  `json:"memo"`
    Insurance      bool    `json:"insurance"`
    DiscountRate   float64 `json:"discount_rate"`
    DiscountAmount float64 `json:"discount_amount"`
    SortOrder      int     `json:"sort_order"`
}

type updateTreatmentRequest struct {
    ItemType       *string  `json:"item_type"`
    ConsultationID *uint64  `json:"consultation_id"`
    ProcedureID    *uint64  `json:"procedure_id"`
    MedicineID     *uint64  `json:"medicine_id"`
    InventoryID    *uint64  `json:"inventory_id"`
    UnitPrice      *float64 `json:"unit_price"`
    Quantity       *int     `json:"quantity"`
    Selected       *bool    `json:"selected"`
    Status         *string  `json:"status"`
    Content        *string  `json:"content"`
    Memo           *string  `json:"memo"`
    Insurance      *bool    `json:"insurance"`
    DiscountRate   *float64 `json:"discount_rate"`
    DiscountAmount *float64 `json:"discount_amount"`
    SortOrder      *int     `json:"sort_order"`
}

type bulkUpdateTreatmentsRequest struct {
    Treatments []bulkTreatmentItem `json:"treatments" binding:"required"`
}

type bulkTreatmentItem struct {
    ID        uint64 `json:"id"         binding:"required"`
    SortOrder int    `json:"sort_order"`
}
```

新規ファイル `handler/treatment_response.go`:
```go
type treatmentResponse struct {
    ID              uint64    `json:"id"`
    MedicalRecordID uint64    `json:"medical_record_id"`
    ItemType        string    `json:"item_type"`
    ConsultationID  *uint64   `json:"consultation_id,omitempty"`
    ProcedureID     *uint64   `json:"procedure_id,omitempty"`
    MedicineID      *uint64   `json:"medicine_id,omitempty"`
    InventoryID     *uint64   `json:"inventory_id,omitempty"`
    UnitPrice       float64   `json:"unit_price"`
    Quantity        int       `json:"quantity"`
    Selected        bool      `json:"selected"`
    Status          string    `json:"status"`
    Content         string    `json:"content"`
    Memo            string    `json:"memo"`
    Insurance       bool      `json:"insurance"`
    DiscountRate    float64   `json:"discount_rate"`
    DiscountAmount  float64   `json:"discount_amount"`
    SortOrder       int       `json:"sort_order"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
    // Nested relations (optional, loaded by Preload)
    Consultation *consultationSummaryResponse `json:"consultation,omitempty"`
    Procedure    *procedureSummaryResponse    `json:"procedure,omitempty"`
    Medicine     *medicineSummaryResponse     `json:"medicine,omitempty"`
}
```

### ルート登録
`cmd/api/main.go` の医療記録グループに追加:
```go
medRecords.GET("/:id/treatments",                h.ListTreatments)
medRecords.POST("/:id/treatments",               h.CreateTreatment)
medRecords.PATCH("/:id/treatments/:treatmentId", h.UpdateTreatment)
medRecords.DELETE("/:id/treatments/:treatmentId", h.DeleteTreatment)
medRecords.PUT("/:id/treatments",                h.BulkUpdateTreatments)
```

## 完了条件
- `GET /v1/medical-records/:id/treatments` が治療項目一覧を `sort_order` 昇順で返す
- `POST /v1/medical-records/:id/treatments` で治療項目を作成できる
- `PATCH /v1/medical-records/:id/treatments/:treatmentId` で部分更新できる
- `DELETE /v1/medical-records/:id/treatments/:treatmentId` で soft delete できる
- `PUT /v1/medical-records/:id/treatments` で並び替え（`sort_order` 一括更新）できる
- 存在しない `medical_record_id` に対しては 404 を返す
- `treatmentId` が指定の medical record に属さない場合も 404 を返す
- 不正な `item_type` は 400 を返す

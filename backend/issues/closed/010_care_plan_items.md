# 入院ケアプラン項目 CRUD 実装

## 概要
入院患者のケアプラン項目（CarePlanItem）の CRUD API を実装する。
`model/hospitalization.go` の `CarePlanItem` struct は実装済み。handler・service・repository が未実装。

## 優先度
high

## 関連テーブル
- `care_plan_items` (id, hospitalization_id NOT NULL, type care_plan_type NOT NULL, name text NOT NULL DEFAULT '', description text DEFAULT '', timing plan_timing[] (配列), status care_plan_status DEFAULT 'active', notes text DEFAULT '', medicine_id, procedure_id, hospitalization_plan_id, unit_price numeric(10,2) DEFAULT 0, category text DEFAULT '', sort_order int DEFAULT 0, created_at, updated_at)
  - `care_plan_type` enum: `food` / `medicine` / `treatment` / `instruction` / `item`
  - `care_plan_status` enum: `active` / `completed` / `discontinued`
  - `plan_timing` enum: `morning` / `noon` / `night`
- `hospitalizations` (親テーブル)
- `medicines`, `procedures`, `hospitalization_plans` (参照先)

## 実装内容

### モデル
`model/hospitalization.go` の `CarePlanItem` は実装済み。変更不要。

`CarePlanItem.Timing` は `pq.StringArray` 型（PostgreSQL の `plan_timing[]` 配列）。
リクエスト・レスポンスでは `[]string` として扱う。

### リポジトリ
新規ファイル `repository/care_plan_item_repository.go`:
```go
type CarePlanItemRepository interface {
    ListByHospitalizationID(ctx context.Context, hospitalizationID uint64) ([]model.CarePlanItem, error)
    FindByID(ctx context.Context, id uint64) (*model.CarePlanItem, error)
    Create(ctx context.Context, item *model.CarePlanItem) error
    Update(ctx context.Context, id uint64, fields map[string]any) error
    Delete(ctx context.Context, id uint64) error
}
```
- `ListByHospitalizationID`: `ORDER BY sort_order ASC`、`Preload(Medicine, Procedure, HospitalizationPlan)`
- `Delete`: 物理削除（`care_plan_items` に deleted_at なし）

`repository/repositories.go` に `CarePlanItem CarePlanItemRepository` を追加。

### サービス
新規ファイル `service/care_plan_item_service.go`:
```go
type CreateCarePlanItemInput struct {
    Type                  model.CarePlanType
    Name                  string
    Description           string
    Timing                []string   // plan_timing 値の配列
    Status                model.CarePlanStatus
    Notes                 string
    MedicineID            *uint64
    ProcedureID           *uint64
    HospitalizationPlanID *uint64
    UnitPrice             float64
    Category              string
    SortOrder             int
}

type UpdateCarePlanItemInput struct {
    Type                  *model.CarePlanType
    Name                  *string
    Description           *string
    Timing                []string  // nil = 未変更、空スライス = クリア
    Status                *model.CarePlanStatus
    Notes                 *string
    MedicineID            *uint64
    ProcedureID           *uint64
    HospitalizationPlanID *uint64
    UnitPrice             *float64
    Category              *string
    SortOrder             *int
}

type CarePlanItemService interface {
    List(ctx context.Context, hospitalizationID uint64) ([]model.CarePlanItem, error)
    Create(ctx context.Context, hospitalizationID uint64, input *CreateCarePlanItemInput) (*model.CarePlanItem, error)
    Update(ctx context.Context, hospitalizationID, itemID uint64, input *UpdateCarePlanItemInput) (*model.CarePlanItem, error)
    Delete(ctx context.Context, hospitalizationID, itemID uint64) error
}
```

`Timing` フィールドは `pq.StringArray` にキャストして保存する。
`buildCarePlanItemUpdateFields()` を実装。`Timing` が空スライス（`len == 0`）で nil ではない場合は空配列で更新する。

`service/validators.go` に `validateCarePlanType`, `validateCarePlanStatus`, `validatePlanTiming` を追加。

### ハンドラ
新規ファイル `handler/care_plan_item_handler.go`:
```go
func (h *Handler) ListCarePlanItems(c *gin.Context)
func (h *Handler) CreateCarePlanItem(c *gin.Context)
func (h *Handler) UpdateCarePlanItem(c *gin.Context)
func (h *Handler) DeleteCarePlanItem(c *gin.Context)
func (h *Handler) RegisterCarePlanItemRoutes(rg *gin.RouterGroup)
```

新規ファイル `handler/care_plan_item_request.go`:
```go
type createCarePlanItemRequest struct {
    Type                  string   `json:"type"   binding:"required"`
    Name                  string   `json:"name"   binding:"required"`
    Description           string   `json:"description"`
    Timing                []string `json:"timing"`
    Status                string   `json:"status"`
    Notes                 string   `json:"notes"`
    MedicineID            *uint64  `json:"medicine_id"`
    ProcedureID           *uint64  `json:"procedure_id"`
    HospitalizationPlanID *uint64  `json:"hospitalization_plan_id"`
    UnitPrice             float64  `json:"unit_price"`
    Category              string   `json:"category"`
    SortOrder             int      `json:"sort_order"`
}

type updateCarePlanItemRequest struct {
    Type                  *string  `json:"type"`
    Name                  *string  `json:"name"`
    Description           *string  `json:"description"`
    Timing                []string `json:"timing"`  // omitempty なし（空配列でクリア可能にする）
    Status                *string  `json:"status"`
    Notes                 *string  `json:"notes"`
    MedicineID            *uint64  `json:"medicine_id"`
    ProcedureID           *uint64  `json:"procedure_id"`
    HospitalizationPlanID *uint64  `json:"hospitalization_plan_id"`
    UnitPrice             *float64 `json:"unit_price"`
    Category              *string  `json:"category"`
    SortOrder             *int     `json:"sort_order"`
}
```

新規ファイル `handler/care_plan_item_response.go` で `carePlanItemResponse` と `toCarePlanItemResponse()` を実装。
`Timing` は `[]string` としてレスポンスに返す。

### ルート登録
`cmd/api/main.go` の入院グループに追加:
```go
hosps.GET("/:id/care-plan-items",              h.ListCarePlanItems)
hosps.POST("/:id/care-plan-items",             h.CreateCarePlanItem)
hosps.PATCH("/:id/care-plan-items/:itemId",    h.UpdateCarePlanItem)
hosps.DELETE("/:id/care-plan-items/:itemId",   h.DeleteCarePlanItem)
```

## 完了条件
- `GET /v1/hospitalizations/:id/care-plan-items` がケアプラン項目一覧を `sort_order` 昇順で返す
- `POST /v1/hospitalizations/:id/care-plan-items` で項目を作成できる（`type`, `name` 必須）
- `PATCH /v1/hospitalizations/:id/care-plan-items/:itemId` で部分更新できる
- `DELETE /v1/hospitalizations/:id/care-plan-items/:itemId` で削除できる
- `timing` フィールドが `["morning", "noon"]` 等の配列で正しく保存・返却される
- 不正な `type` / `status` / `timing` 値は 400 エラーを返す
- 他院の入院 ID を指定すると 404 を返す

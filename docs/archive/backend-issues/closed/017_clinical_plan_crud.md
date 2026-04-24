# 診察所見・診断・治療方針（ClinicalPlan）CRUD 実装

## 概要
外来カルテに 1:1 で紐づく診察所見・診断・治療方針（ClinicalPlan）の CRUD API を実装する。
`model/clinical_plan.go` の `ClinicalPlan` struct は実装済み。handler・service・repository が未実装。

`clinical_plans` は `medical_record_id` に UNIQUE 制約があるため、1カルテにつき最大1件。
GET/POST/PATCH/DELETE のうち POST は「存在しなければ作成・存在すれば更新」の Upsert 相当になる。

## 優先度
high（診察フローのコア）

## 関連テーブル
- `clinical_plans` (id, medical_record_id NOT NULL UNIQUE, physical_exam text NOT NULL DEFAULT '', diagnosis_category_id, diagnosis_name_id, diagnosis_details text NOT NULL DEFAULT '', treatment_policy text NOT NULL DEFAULT '', created_at, updated_at)
  - UNIQUE制約: medical_record_id（1カルテ1レコード）
  - soft delete なし
- `medical_records` (親テーブル)
- `diagnosis_categories`, `diagnosis_names` (参照先マスタ)

## 実装内容

### モデル
`model/clinical_plan.go` は実装済み。変更不要。

```go
type ClinicalPlan struct {
    ID                  uint64
    MedicalRecordID     uint64    // uniqueIndex
    PhysicalExam        string
    DiagnosisCategoryID *uint64
    DiagnosisNameID     *uint64
    DiagnosisDetails    string
    TreatmentPolicy     string
    CreatedAt           time.Time
    UpdatedAt           time.Time
    // Relations: MedicalRecord, DiagnosisCategory, DiagnosisName
}
```

### リポジトリ
新規ファイル `repository/clinical_plan_repository.go`:
```go
type ClinicalPlanRepository interface {
    FindByMedicalRecordID(ctx context.Context, medicalRecordID uint64) (*model.ClinicalPlan, error)
    Create(ctx context.Context, plan *model.ClinicalPlan) error
    Update(ctx context.Context, id uint64, fields map[string]any) error
    Delete(ctx context.Context, id uint64) error
}
```
- `FindByMedicalRecordID`: UNIQUE制約のため0-1件。存在しない場合は `WrapNotFound`
- `Delete`: 物理削除（`clinical_plans` に deleted_at なし）
- `Update`: `map[string]any` で GORM ゼロ値問題を回避

`repository/repositories.go` の `Repositories` struct に `ClinicalPlan ClinicalPlanRepository` を追加。

### サービス
新規ファイル `service/clinical_plan_service.go`:
```go
type SaveClinicalPlanInput struct {
    PhysicalExam        string
    DiagnosisCategoryID *uint64
    DiagnosisNameID     *uint64
    DiagnosisDetails    string
    TreatmentPolicy     string
}

type UpdateClinicalPlanInput struct {
    PhysicalExam        *string
    DiagnosisCategoryID *uint64
    DiagnosisNameID     *uint64
    DiagnosisDetails    *string
    TreatmentPolicy     *string
}

type ClinicalPlanService interface {
    GetOrCreate(ctx context.Context, medicalRecordID uint64) (*model.ClinicalPlan, error)
    Update(ctx context.Context, medicalRecordID uint64, input *UpdateClinicalPlanInput) (*model.ClinicalPlan, error)
    Delete(ctx context.Context, medicalRecordID uint64) error
}
```
- `GetOrCreate`: `FindByMedicalRecordID` → ErrNotFound の場合は空レコードを `Create` して返す
- `Update`: `GetOrCreate` で取得 → `buildClinicalPlanUpdateFields` → `repo.Update` → `repo.FindByMedicalRecordID` で返す
- `Delete`: `FindByMedicalRecordID` で存在確認 → `repo.Delete`

`service/service.go` の `Services` struct に `ClinicalPlan ClinicalPlanService` を追加。

### ハンドラ
新規ファイル `handler/clinical_plan_request.go`:
```go
type updateClinicalPlanRequest struct {
    PhysicalExam        *string `json:"physical_exam"`
    DiagnosisCategoryID *uint64 `json:"diagnosis_category_id"`
    DiagnosisNameID     *uint64 `json:"diagnosis_name_id"`
    DiagnosisDetails    *string `json:"diagnosis_details"`
    TreatmentPolicy     *string `json:"treatment_policy"`
}
```

新規ファイル `handler/clinical_plan_response.go`:
```go
type clinicalPlanResponse struct {
    ID                  string    `json:"id"`
    MedicalRecordID     string    `json:"medical_record_id"`
    PhysicalExam        string    `json:"physical_exam"`
    DiagnosisCategoryID *string   `json:"diagnosis_category_id,omitempty"`
    DiagnosisNameID     *string   `json:"diagnosis_name_id,omitempty"`
    DiagnosisDetails    string    `json:"diagnosis_details"`
    TreatmentPolicy     string    `json:"treatment_policy"`
    CreatedAt           time.Time `json:"created_at"`
    UpdatedAt           time.Time `json:"updated_at"`
    // Nested
    DiagnosisCategory *diagnosisCategorySummaryResponse `json:"diagnosis_category,omitempty"`
    DiagnosisName     *diagnosisNameSummaryResponse     `json:"diagnosis_name,omitempty"`
}
```

新規ファイル `handler/clinical_plan_handler.go`:
```go
func (h *Handler) GetClinicalPlan(c *gin.Context)    // GET /:id/clinical-plan
func (h *Handler) UpdateClinicalPlan(c *gin.Context)  // PATCH /:id/clinical-plan
func (h *Handler) DeleteClinicalPlan(c *gin.Context)  // DELETE /:id/clinical-plan
func (h *Handler) RegisterClinicalPlanRoutes(rg *gin.RouterGroup)
```

GET は「存在しない場合は空レコードを自動作成（GetOrCreate）」で常に 200 を返す。

### ルート登録
`handler/medical_record_handler.go` の `RegisterMedicalRecordRoutes` に追加:
```go
h.RegisterClinicalPlanRoutes(records)
```

`RegisterClinicalPlanRoutes` の実装:
```go
func (h *Handler) RegisterClinicalPlanRoutes(rg *gin.RouterGroup) {
    rg.GET("/:id/clinical-plan",    h.GetClinicalPlan)
    rg.PATCH("/:id/clinical-plan",  h.UpdateClinicalPlan)
    rg.DELETE("/:id/clinical-plan", h.DeleteClinicalPlan)
}
```

注: api.yaml では `/clinical-plans`（複数形）で定義されているが、1:1関係のため単数形 `/clinical-plan` を推奨。
api.yaml を単数形に修正するか、ルートを複数形に合わせるかは実装時に判断する。

## 完了条件
- `GET /v1/medical-records/:id/clinical-plan` が所見・診断・治療方針を返す（初回は空レコードを自動作成）
- `PATCH /v1/medical-records/:id/clinical-plan` で部分更新できる（nil フィールドはスキップ）
- `DELETE /v1/medical-records/:id/clinical-plan` で削除できる
- 存在しない `medical_record_id` に対しては 404 を返す
- 二重 POST（Upsert）しても 409 を返さずに更新で処理する

# 外来バイタル CRUD 実装

## 概要
外来カルテに対するバイタル記録（体温・心拍数・呼吸数・体重）の CRUD API を実装する。
`model/vital.go` の `Vital` struct は実装済み。handler・service・repository が未実装。

## 優先度
high

## 関連テーブル
- `vitals` (id, medical_record_id NOT NULL, recorded_at NOT NULL DEFAULT now(), staff_id, temperature numeric(4,1), heart_rate int, respiration_rate int, weight numeric(6,2), notes text DEFAULT '', created_at)
- `medical_records` (親テーブル、存在確認に使用)
- `staffs` (staff_id の参照先)

## 実装内容

### モデル
`model/vital.go` は実装済み。変更不要。

```go
type Vital struct {
    ID              uint64
    MedicalRecordID uint64
    RecordedAt      time.Time
    StaffID         *uint64
    Temperature     *float64
    HeartRate       *int
    RespirationRate *int
    Weight          *float64
    Notes           string
    CreatedAt       time.Time
}
```

### リポジトリ
新規ファイル `repository/vital_repository.go`:
```go
type VitalRepository interface {
    ListByMedicalRecordID(ctx context.Context, medicalRecordID uint64) ([]model.Vital, error)
    Create(ctx context.Context, vital *model.Vital) error
    Update(ctx context.Context, id uint64, fields map[string]any) error
    Delete(ctx context.Context, id uint64) error
    FindByID(ctx context.Context, id uint64) (*model.Vital, error)
}
```
- `ListByMedicalRecordID`: `ORDER BY recorded_at ASC`
- `Update`: `map[string]any` で GORM ゼロ値問題を回避
- `Delete`: 物理削除（`DELETE FROM vitals WHERE id = ?`）

`repository/repositories.go` の `Repositories` struct に `Vital VitalRepository` を追加。

### サービス
新規ファイル `service/vital_service.go`:
```go
type CreateVitalInput struct {
    RecordedAt      time.Time
    StaffID         *uint64
    Temperature     *float64
    HeartRate       *int
    RespirationRate *int
    Weight          *float64
    Notes           string
}

type UpdateVitalInput struct {
    RecordedAt      *time.Time
    StaffID         *uint64
    Temperature     **float64  // nil = 未送信、&nil = null に更新
    HeartRate       **int
    RespirationRate **int
    Weight          **float64
    Notes           *string
}

type VitalService interface {
    List(ctx context.Context, medicalRecordID uint64) ([]model.Vital, error)
    Create(ctx context.Context, medicalRecordID uint64, input *CreateVitalInput) (*model.Vital, error)
    Update(ctx context.Context, medicalRecordID, vitalID uint64, input *UpdateVitalInput) (*model.Vital, error)
    Delete(ctx context.Context, medicalRecordID, vitalID uint64) error
}
```

`buildVitalUpdateFields(input *UpdateVitalInput) map[string]any` を実装する。
service DI は `service/services.go` の `Services` struct に追加する。

### ハンドラ
新規ファイル `handler/vital_handler.go`:
```go
func (h *Handler) ListVitals(c *gin.Context)
func (h *Handler) CreateVital(c *gin.Context)
func (h *Handler) UpdateVital(c *gin.Context)
func (h *Handler) DeleteVital(c *gin.Context)
func (h *Handler) RegisterVitalRoutes(rg *gin.RouterGroup)
```

新規ファイル `handler/vital_request.go`:
```go
type createVitalRequest struct {
    RecordedAt      time.Time `json:"recorded_at"`
    StaffID         *uint64   `json:"staff_id"`
    Temperature     *float64  `json:"temperature"`
    HeartRate       *int      `json:"heart_rate"`
    RespirationRate *int      `json:"respiration_rate"`
    Weight          *float64  `json:"weight"`
    Notes           string    `json:"notes"`
}

type updateVitalRequest struct {
    RecordedAt      *time.Time `json:"recorded_at"`
    StaffID         *uint64    `json:"staff_id"`
    Temperature     *float64   `json:"temperature"`
    HeartRate       *int       `json:"heart_rate"`
    RespirationRate *int       `json:"respiration_rate"`
    Weight          *float64   `json:"weight"`
    Notes           *string    `json:"notes"`
}
```

新規ファイル `handler/vital_response.go`:
```go
type vitalResponse struct {
    ID              uint64     `json:"id"`
    MedicalRecordID uint64     `json:"medical_record_id"`
    RecordedAt      time.Time  `json:"recorded_at"`
    StaffID         *uint64    `json:"staff_id,omitempty"`
    Temperature     *float64   `json:"temperature,omitempty"`
    HeartRate       *int       `json:"heart_rate,omitempty"`
    RespirationRate *int       `json:"respiration_rate,omitempty"`
    Weight          *float64   `json:"weight,omitempty"`
    Notes           string     `json:"notes"`
    CreatedAt       time.Time  `json:"created_at"`
    Staff           *staffSummaryResponse `json:"staff,omitempty"`
}
```

### ルート登録
`cmd/api/main.go` の医療記録グループに追加:
```go
medRecords := v1.Group("/medical-records", authMiddleware)
// 既存ルートに追加
medRecords.GET("/:id/vitals",              h.ListVitals)
medRecords.POST("/:id/vitals",             h.CreateVital)
medRecords.PATCH("/:id/vitals/:vitalId",   h.UpdateVital)
medRecords.DELETE("/:id/vitals/:vitalId",  h.DeleteVital)
```

## 完了条件
- `GET /v1/medical-records/:id/vitals` がバイタル一覧を `recorded_at` 昇順で返す
- `POST /v1/medical-records/:id/vitals` でバイタルを作成できる
- `PATCH /v1/medical-records/:id/vitals/:vitalId` で部分更新できる（nil フィールドは更新しない）
- `DELETE /v1/medical-records/:id/vitals/:vitalId` で削除できる
- 存在しない `medical_record_id` に対するリクエストは 404 を返す
- `vitalId` が指定 medical record に属さない場合も 404 を返す

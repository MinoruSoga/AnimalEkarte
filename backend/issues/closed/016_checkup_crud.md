# 健康診断記録 CRUD 実装

## 概要
外来カルテに紐づく健康診断記録（Checkup）の CRUD API を実装する。
`model/checkup_record.go` の `Checkup` struct は実装済み。handler・service・repository が未実装。

## 優先度
high

## 関連テーブル
- `checkups` (id, medical_record_id NOT NULL, pet_id, checkup_type_id NOT NULL, date date NOT NULL, next_date date, doctor_id, result text NOT NULL DEFAULT '', created_at, updated_at, deleted_at)
  - `checkup_type_id` → `checkup_types.id` RESTRICT
  - `doctor_id` → `staffs.id` SET NULL
  - soft delete あり（deleted_at）
- `medical_records` (親テーブル)
- `checkup_types` (マスタ)
- `staffs` (doctor_id 参照先)

## 実装内容

### モデル
`model/checkup_record.go` は実装済み。変更不要。

```go
type Checkup struct {
    ID              uint64
    MedicalRecordID uint64
    PetID           *uint64
    CheckupTypeID   uint64
    Date            time.Time      // type:date
    NextDate        *time.Time     // type:date, nullable
    DoctorID        *uint64
    Result          string
    DeletedAt       gorm.DeletedAt
    CreatedAt       time.Time
    UpdatedAt       time.Time
    // Relations: MedicalRecord, Pet, CheckupType, Doctor
}
```

### リポジトリ
新規ファイル `repository/checkup_repository.go`:
```go
type CheckupRepository interface {
    ListByMedicalRecordID(ctx context.Context, medicalRecordID uint64) ([]model.Checkup, error)
    FindByID(ctx context.Context, id uint64) (*model.Checkup, error)
    Create(ctx context.Context, checkup *model.Checkup) error
    Update(ctx context.Context, id uint64, fields map[string]any) error
    Delete(ctx context.Context, id uint64) error
}
```
- `ListByMedicalRecordID`: `ORDER BY date ASC`、`Preload("CheckupType", "Doctor")`、soft delete 対応
- `Delete`: GORM soft delete（`gorm.DeletedAt` フィールドがあるため `Delete(&model.Checkup{}, id)` で自動処理）
- `Update`: `map[string]any` で GORM ゼロ値問題を回避、`RowsAffected == 0` で WrapNotFound

`repository/repositories.go` の `Repositories` struct に `Checkup CheckupRepository` を追加。

### サービス
新規ファイル `service/checkup_service.go`:
```go
type CreateCheckupInput struct {
    CheckupTypeID uint64
    PetID         *uint64
    Date          time.Time
    NextDate      *time.Time
    DoctorID      *uint64
    Result        string
}

type UpdateCheckupInput struct {
    CheckupTypeID *uint64
    PetID         *uint64
    Date          *time.Time
    NextDate      *time.Time
    DoctorID      *uint64
    Result        *string
}

type CheckupService interface {
    List(ctx context.Context, medicalRecordID uint64) ([]model.Checkup, error)
    Create(ctx context.Context, medicalRecordID uint64, input *CreateCheckupInput) (*model.Checkup, error)
    Update(ctx context.Context, medicalRecordID, checkupID uint64, input *UpdateCheckupInput) (*model.Checkup, error)
    Delete(ctx context.Context, medicalRecordID, checkupID uint64) error
}
```
- `Create`: `MedicalRecordID` をセット → `repo.Create` → `repo.FindByID` で返す
- `Update`: `buildCheckupUpdateFields` → `repo.Update` → `repo.FindByID`
- `Delete`: `repo.FindByID` で親所属確認 → `repo.Delete`

`service/service.go` の `Services` struct に `Checkup CheckupService` を追加。

### ハンドラ
新規ファイル `handler/checkup_request.go`:
```go
type createCheckupRequest struct {
    CheckupTypeID uint64     `json:"checkup_type_id" binding:"required"`
    PetID         *uint64    `json:"pet_id"`
    Date          time.Time  `json:"date"            binding:"required"`
    NextDate      *time.Time `json:"next_date"`
    DoctorID      *uint64    `json:"doctor_id"`
    Result        string     `json:"result"`
}

type updateCheckupRequest struct {
    CheckupTypeID *uint64    `json:"checkup_type_id"`
    PetID         *uint64    `json:"pet_id"`
    Date          *time.Time `json:"date"`
    NextDate      *time.Time `json:"next_date"`
    DoctorID      *uint64    `json:"doctor_id"`
    Result        *string    `json:"result"`
}
```

新規ファイル `handler/checkup_response.go`:
```go
type checkupResponse struct {
    ID              string     `json:"id"`
    MedicalRecordID string     `json:"medical_record_id"`
    CheckupTypeID   string     `json:"checkup_type_id"`
    PetID           *string    `json:"pet_id,omitempty"`
    Date            time.Time  `json:"date"`
    NextDate        *time.Time `json:"next_date,omitempty"`
    DoctorID        *string    `json:"doctor_id,omitempty"`
    Result          string     `json:"result"`
    DeletedAt       *time.Time `json:"deleted_at,omitempty"`
    CreatedAt       time.Time  `json:"created_at"`
    UpdatedAt       time.Time  `json:"updated_at"`
    // Nested
    CheckupType *checkupTypeSummaryResponse `json:"checkup_type,omitempty"`
    Doctor      *staffSummaryResponse       `json:"doctor,omitempty"`
}
```

新規ファイル `handler/checkup_handler.go`:
```go
func (h *Handler) ListCheckups(c *gin.Context)
func (h *Handler) CreateCheckup(c *gin.Context)
func (h *Handler) UpdateCheckup(c *gin.Context)
func (h *Handler) DeleteCheckup(c *gin.Context)
func (h *Handler) RegisterCheckupRoutes(rg *gin.RouterGroup)
```

### ルート登録
`handler/medical_record_handler.go` の `RegisterMedicalRecordRoutes` に追加:
```go
h.RegisterCheckupRoutes(records)
```

`RegisterCheckupRoutes` の実装:
```go
func (h *Handler) RegisterCheckupRoutes(rg *gin.RouterGroup) {
    rg.GET("/:id/checkups",               h.ListCheckups)
    rg.POST("/:id/checkups",              h.CreateCheckup)
    rg.PATCH("/:id/checkups/:checkupId",  h.UpdateCheckup)
    rg.DELETE("/:id/checkups/:checkupId", h.DeleteCheckup)
}
```

## 完了条件
- `GET /v1/medical-records/:id/checkups` が健診記録一覧を `date` 昇順で返す
- `POST /v1/medical-records/:id/checkups` で健診記録を作成できる
- `PATCH /v1/medical-records/:id/checkups/:checkupId` で部分更新できる
- `DELETE /v1/medical-records/:id/checkups/:checkupId` で soft delete できる
- 存在しない `medical_record_id` に対しては 404 を返す
- 他カルテの checkup_id を指定すると 404 を返す

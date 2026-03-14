# 入院日次記録 CRUD 実装

## 概要
入院患者の日次ケア記録（DailyRecord）、バイタル記録（VitalRecord）、ケアログ（CareLogRecord）、スタッフメモ（StaffNoteRecord）の CRUD API を実装する。
これらのモデルは `model/hospitalization.go` に実装済み。handler・service・repository が未実装。

## 優先度
high

## 関連テーブル
- `daily_records` (id, hospitalization_id NOT NULL, date date NOT NULL, created_at, updated_at)
  - UNIQUE(hospitalization_id, date)
- `vital_records` (id, daily_record_id NOT NULL, time text NOT NULL DEFAULT '', temperature numeric(4,1), heart_rate int, respiration_rate int, weight numeric(6,2), notes text DEFAULT '', staff_id, created_at)
- `care_log_records` (id, daily_record_id NOT NULL, time text NOT NULL DEFAULT '', type care_log_type NOT NULL, status care_log_status NOT NULL DEFAULT 'completed', value text DEFAULT '', staff_id, notes text DEFAULT '', created_at)
  - `care_log_type` enum: `food` / `excretion` / `medicine` / `treatment` / `other`
  - `care_log_status` enum: `completed` / `partial` / `skipped`
- `staff_note_records` (id, daily_record_id NOT NULL, time text NOT NULL DEFAULT '', content text NOT NULL DEFAULT '', staff_id, created_at)

## 実装内容

### モデル
`model/hospitalization.go` に `DailyRecord`, `VitalRecord`, `CareLogRecord`, `StaffNoteRecord` は実装済み。変更不要。

### リポジトリ
新規ファイル `repository/daily_record_repository.go`:
```go
type DailyRecordRepository interface {
    ListByHospitalizationID(ctx context.Context, hospitalizationID uint64) ([]model.DailyRecord, error)
    FindByHospitalizationIDAndDate(ctx context.Context, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error)
    GetOrCreateByDate(ctx context.Context, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error)
    CreateVitalRecord(ctx context.Context, vr *model.VitalRecord) error
    CreateCareLogRecord(ctx context.Context, cr *model.CareLogRecord) error
    CreateStaffNoteRecord(ctx context.Context, sn *model.StaffNoteRecord) error
    DeleteVitalRecord(ctx context.Context, id uint64) error
    DeleteCareLogRecord(ctx context.Context, id uint64) error
    DeleteStaffNoteRecord(ctx context.Context, id uint64) error
}
```
- `ListByHospitalizationID`: `ORDER BY date DESC`、各 DailyRecord を `Preload(VitalRecords, CareLogRecords, StaffNoteRecords)` で返す
- `FindByHospitalizationIDAndDate`: 日付でピンポイント取得。存在しない場合は `WrapNotFound`
- `GetOrCreateByDate`: 存在しない場合のみ INSERT（UPSERT or SELECT → INSERT）
- 子テーブルの DELETE は物理削除

`repository/repositories.go` に `DailyRecord DailyRecordRepository` を追加。

### サービス
新規ファイル `service/daily_record_service.go`:
```go
type CreateVitalRecordInput struct {
    Time            string
    Temperature     *float64
    HeartRate       *int
    RespirationRate *int
    Weight          *float64
    Notes           string
    StaffID         *uint64
}

type CreateCareLogRecordInput struct {
    Time    string
    Type    model.CareLogType
    Status  model.CareLogStatus
    Value   string
    StaffID *uint64
    Notes   string
}

type CreateStaffNoteRecordInput struct {
    Time    string
    Content string
    StaffID *uint64
}

type DailyRecordService interface {
    List(ctx context.Context, hospitalizationID uint64) ([]model.DailyRecord, error)
    GetOrCreateByDate(ctx context.Context, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error)
    AddVitalRecord(ctx context.Context, hospitalizationID uint64, date time.Time, input *CreateVitalRecordInput) (*model.DailyRecord, error)
    AddCareLogRecord(ctx context.Context, hospitalizationID uint64, date time.Time, input *CreateCareLogRecordInput) (*model.DailyRecord, error)
    AddStaffNoteRecord(ctx context.Context, hospitalizationID uint64, date time.Time, input *CreateStaffNoteRecordInput) (*model.DailyRecord, error)
}
```

`service/validators.go` に `validateCareLogType`, `validateCareLogStatus` を追加。

### ハンドラ
新規ファイル `handler/daily_record_handler.go`:
```go
func (h *Handler) ListDailyRecords(c *gin.Context)
func (h *Handler) GetDailyRecord(c *gin.Context)   // GET /:id/daily-records/:date
func (h *Handler) CreateDailyRecord(c *gin.Context) // POST（日次記録を明示的に作成）
func (h *Handler) AddVitalRecord(c *gin.Context)
func (h *Handler) AddCareLogRecord(c *gin.Context)
func (h *Handler) AddStaffNoteRecord(c *gin.Context)
func (h *Handler) RegisterDailyRecordRoutes(rg *gin.RouterGroup)
```

URL パラメータ `:date` は `YYYY-MM-DD` 形式。ハンドラで `time.Parse("2006-01-02", dateStr)` でパース。

新規ファイル `handler/daily_record_request.go`:
```go
type addVitalRecordRequest struct {
    Time            string   `json:"time"             binding:"required"`
    Temperature     *float64 `json:"temperature"`
    HeartRate       *int     `json:"heart_rate"`
    RespirationRate *int     `json:"respiration_rate"`
    Weight          *float64 `json:"weight"`
    Notes           string   `json:"notes"`
    StaffID         *uint64  `json:"staff_id"`
}

type addCareLogRecordRequest struct {
    Time    string  `json:"time"    binding:"required"`
    Type    string  `json:"type"    binding:"required"`
    Status  string  `json:"status"`
    Value   string  `json:"value"`
    StaffID *uint64 `json:"staff_id"`
    Notes   string  `json:"notes"`
}

type addStaffNoteRecordRequest struct {
    Time    string  `json:"time"    binding:"required"`
    Content string  `json:"content" binding:"required"`
    StaffID *uint64 `json:"staff_id"`
}
```

新規ファイル `handler/daily_record_response.go` で各 response struct と `toDailyRecordResponse()` を実装。

### ルート登録
`cmd/api/main.go` の入院グループに追加:
```go
hosps := v1.Group("/hospitalizations", authMiddleware)
// 既存ルートに追加
hosps.GET("/:id/daily-records",                            h.ListDailyRecords)
hosps.POST("/:id/daily-records",                           h.CreateDailyRecord)
hosps.GET("/:id/daily-records/:date",                      h.GetDailyRecord)
hosps.POST("/:id/daily-records/:date/vitals",              h.AddVitalRecord)
hosps.POST("/:id/daily-records/:date/care-logs",           h.AddCareLogRecord)
hosps.POST("/:id/daily-records/:date/staff-notes",         h.AddStaffNoteRecord)
```

## 完了条件
- `GET /v1/hospitalizations/:id/daily-records` が日次記録一覧を日付降順で返す（子レコード込み）
- `GET /v1/hospitalizations/:id/daily-records/:date` が指定日の記録を返す（存在しない場合は 404）
- `POST /v1/hospitalizations/:id/daily-records` で日次記録を明示的に作成できる
- `POST /:id/daily-records/:date/vitals` でバイタルを追加できる（DailyRecord が存在しない場合は自動作成）
- `POST /:id/daily-records/:date/care-logs` でケアログを追加できる
- `POST /:id/daily-records/:date/staff-notes` でスタッフメモを追加できる
- 不正な `type` / `status` 値は 400 エラーを返す
- 他院の入院記録には 404 を返す

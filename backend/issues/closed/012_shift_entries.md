# スタッフシフト管理 CRUD 実装

## 概要
スタッフのシフト（ShiftEntry）の CRUD API を実装する。
`model/staff.go` の `ShiftEntry` struct は実装済み。handler・service・repository が未実装。

## 優先度
low

## 関連テーブル
- `shift_entries` (id, clinic_id NOT NULL, staff_id NOT NULL, date date NOT NULL, shift_type shift_type NOT NULL, start_time text DEFAULT '', end_time text DEFAULT '', note text DEFAULT '', created_at, updated_at)
  - `shift_type` enum: `full` / `morning` / `afternoon` / `off` / `paid_leave`
  - UNIQUE(clinic_id, staff_id, date)（推奨: 同日の重複シフトを防ぐ）
- `staffs` (staff_id の参照先)

## 実装内容

### モデル
`model/staff.go` の `ShiftEntry` / `ShiftType` は実装済み。変更不要。

```go
type ShiftType string
const (
    ShiftTypeFull      ShiftType = "full"
    ShiftTypeMorning   ShiftType = "morning"
    ShiftTypeAfternoon ShiftType = "afternoon"
    ShiftTypeOff       ShiftType = "off"
    ShiftTypePaidLeave ShiftType = "paid_leave"
)
```

### リポジトリ
新規ファイル `repository/shift_entry_repository.go`:
```go
type ShiftEntryFilter struct {
    YearMonth string  // "YYYY-MM" 形式
    StaffID   *uint64
}

type ShiftEntryRepository interface {
    List(ctx context.Context, clinicID uint64, filter ShiftEntryFilter) ([]model.ShiftEntry, error)
    FindByID(ctx context.Context, clinicID, id uint64) (*model.ShiftEntry, error)
    Create(ctx context.Context, entry *model.ShiftEntry) error
    Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
    Delete(ctx context.Context, clinicID, id uint64) error
}
```
- `List`: `YearMonth` フィルタは `WHERE date >= '2024-03-01' AND date < '2024-04-01'` 相当に変換。`ORDER BY date ASC, staff_id ASC`
- `Preload(Staff)` で Staff 情報付き返却
- `Create`: 同日・同スタッフの重複は `apperrors.WrapAlreadyExists` で返す

`repository/repositories.go` に `ShiftEntry ShiftEntryRepository` を追加。

### サービス
新規ファイル `service/shift_entry_service.go`:
```go
type CreateShiftEntryInput struct {
    StaffID   uint64
    Date      time.Time
    ShiftType model.ShiftType
    StartTime string
    EndTime   string
    Note      string
}

type UpdateShiftEntryInput struct {
    ShiftType *model.ShiftType
    StartTime *string
    EndTime   *string
    Note      *string
}

type ShiftEntryService interface {
    List(ctx context.Context, clinicID uint64, yearMonth string, staffID *uint64) ([]model.ShiftEntry, error)
    Create(ctx context.Context, clinicID uint64, input *CreateShiftEntryInput) (*model.ShiftEntry, error)
    Update(ctx context.Context, clinicID, id uint64, input *UpdateShiftEntryInput) (*model.ShiftEntry, error)
    Delete(ctx context.Context, clinicID, id uint64) error
}
```

`service/validators.go` に追加:
- `validateShiftType(shiftType string) error`
- `validateYearMonth(yearMonth string) error` — `YYYY-MM` 形式チェック

### ハンドラ
新規ファイル `handler/shift_entry_handler.go`:
```go
func (h *Handler) ListShifts(c *gin.Context)
func (h *Handler) CreateShift(c *gin.Context)
func (h *Handler) UpdateShift(c *gin.Context)
func (h *Handler) DeleteShift(c *gin.Context)
func (h *Handler) RegisterShiftEntryRoutes(rg *gin.RouterGroup)
```

新規ファイル `handler/shift_entry_request.go`:
```go
type createShiftEntryRequest struct {
    StaffID   uint64    `json:"staff_id"   binding:"required"`
    Date      time.Time `json:"date"       binding:"required"`
    ShiftType string    `json:"shift_type" binding:"required"`
    StartTime string    `json:"start_time"`
    EndTime   string    `json:"end_time"`
    Note      string    `json:"note"`
}

type updateShiftEntryRequest struct {
    ShiftType *string `json:"shift_type"`
    StartTime *string `json:"start_time"`
    EndTime   *string `json:"end_time"`
    Note      *string `json:"note"`
}
```

新規ファイル `handler/shift_entry_response.go` で `shiftEntryResponse` と `toShiftEntryResponse()` を実装。
`Staff` フィールドは `staffSummaryResponse` でネストする。

クエリパラメータ:
- `date`: `YYYY-MM` 形式（月単位フィルタ）
- `staff_id`: 特定スタッフで絞り込み（optional）

### ルート登録
`cmd/api/main.go` に追加:
```go
shifts := v1.Group("/shifts", authMiddleware)
shifts.GET("",     h.ListShifts)
shifts.POST("",    h.CreateShift)
shifts.PATCH("/:id", h.UpdateShift)
shifts.DELETE("/:id", h.DeleteShift)
```

## 完了条件
- `GET /v1/shifts?date=2024-03` が月単位でシフト一覧を返す
- `GET /v1/shifts?date=2024-03&staff_id=5` でスタッフ絞り込みができる
- `POST /v1/shifts` でシフトを作成できる
- `PATCH /v1/shifts/:id` でシフトを部分更新できる
- `DELETE /v1/shifts/:id` でシフトを削除できる
- 同日・同スタッフの重複シフト作成は 409 エラーを返す
- 不正な `shift_type` は 400 エラーを返す
- 他クリニックのシフトは参照・変更できない

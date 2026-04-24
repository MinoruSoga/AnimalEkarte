# BE-120: LIFF トリミング予約フロー拡張

**Status**: Open
**Priority**: High
**Affects**: /api/liff/:clinicId — トリミング区分選択後の追加ステップ
**Date Created**: 2026-04-16
**Related**: TASK-002, BE-118（前提）, FE-254

## Summary

LIFF予約フローに「トリミングコース選択」「トリミングオプション選択」ステップを追加する。
`reservation_types.category = 'trimming'` の区分が選択された場合のみ追加ステップが表示される。
2つの新エンドポイントと `CreateReservation` の入力拡張で対応する。

## 現状のコード

```go
// backend/internal/service/liff_service.go:17-27
type LiffService interface {
    GetSettings(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
    GetProfile(ctx context.Context, clinicID, customerID uint64) (*model.LineCustomer, error)
    GetCourses(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
    GetStaffs(ctx context.Context, clinicID, typeID uint64) ([]model.Staff, error)
    GetAvailableDates(ctx context.Context, clinicID, typeID, staffID uint64) ([]AvailableDateResult, BookingWindow, error)
    GetAvailableTimes(ctx context.Context, clinicID, typeID, staffID uint64, date time.Time) ([]TimeSlot, error)
    CreateReservation(ctx context.Context, clinicID, customerID uint64, input *CreateReservationInput) (*model.Appointment, error)
    GetMyReservations(ctx context.Context, clinicID, customerID uint64) ([]model.Appointment, error)
    CancelReservation(ctx context.Context, clinicID, customerID, reservationID uint64) error
}
// ※ GetTrimmingCourses / GetTrimmingOptions が存在しない

// backend/internal/handler/liff_request.go:6-14
type liffCreateReservationRequest struct {
    TypeID         uint64          `json:"course_id"      binding:"required"`
    StaffID        uint64          `json:"staff_id"`
    Date           string          `json:"date"             binding:"required"` // "YYYY-MM-DD"
    StartTime      string          `json:"start_time"       binding:"required"` // "HHMM"
    EndTime        string          `json:"end_time"         binding:"required"` // "HHMM"
    CustomerFields json.RawMessage `json:"customer_fields"`
    RequestText    string          `json:"request_text"`
}
// ※ トリミング詳細フィールドなし

// backend/internal/handler/reservation_line_routes.go:62-83（LIFF ルート定義）
authed.GET("/courses", h.GetLiffTypes)
// ※ /trimming-courses / /trimming-options がない
```

## 必要な変更

### 1. `backend/internal/service/liff_service.go` — インターフェース拡張

#### 1-a. インターフェースに新メソッド追加

```go
type LiffService interface {
    // ... 既存メソッド（変更なし） ...
    GetTrimmingCourses(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) // ★ 追加
    GetTrimmingOptions(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)  // ★ 追加
}
```

#### 1-b. `liffService` struct に TrimmingCourse/Option リポジトリを追加

```go
type liffService struct {
    settingRepo      repository.LineReservationSettingRepository
    typeLiffRepo     repository.ReservationTypeLiffRepository
    staffRepo        repository.ReservationStaffRepository
    scheduleRepo     repository.ReservationScheduleRepository
    adminRepo        repository.ReservationAdminRepository
    reservationRepo  repository.ReservationRepository
    customerRepo     repository.LineCustomerRepository
    ownerRepo        repository.OwnerRepository
    trimmingCourse   repository.TrimmingCourseRepository   // ★ 追加
    trimmingOption   repository.TrimmingOptionRepository   // ★ 追加
    trimmingDetail   repository.AppointmentTrimmingDetailRepository // ★ 追加
    validators       ReservationValidators
    notifier         ReservationNotifier
}
```

#### 1-c. `NewLiffService` 引数追加

```go
func NewLiffService(
    settingRepo repository.LineReservationSettingRepository,
    typeLiffRepo repository.ReservationTypeLiffRepository,
    staffRepo repository.ReservationStaffRepository,
    scheduleRepo repository.ReservationScheduleRepository,
    adminRepo repository.ReservationAdminRepository,
    customerRepo repository.LineCustomerRepository,
    ownerRepo repository.OwnerRepository,
    trimmingCourse repository.TrimmingCourseRepository,    // ★ 追加
    trimmingOption repository.TrimmingOptionRepository,    // ★ 追加
    trimmingDetail repository.AppointmentTrimmingDetailRepository, // ★ 追加
    validators ReservationValidators,
    notifier ReservationNotifier,
) LiffService {
    return &liffService{
        // ... 既存フィールド ...
        trimmingCourse: trimmingCourse,
        trimmingOption: trimmingOption,
        trimmingDetail: trimmingDetail,
    }
}
```

#### 1-d. 新メソッド実装

```go
// GetTrimmingCourses はLIFF向けトリミングコース一覧を返す（is_active=true のみ）
func (s *liffService) GetTrimmingCourses(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
    all, err := s.trimmingCourse.FindAll(ctx, clinicID)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to get trimming courses")
    }
    result := make([]model.TrimmingCourse, 0, len(all))
    for i := range all {
        if all[i].IsActive {
            result = append(result, all[i])
        }
    }
    return result, nil
}

// GetTrimmingOptions はLIFF向けトリミングオプション一覧を返す（is_active=true のみ）
func (s *liffService) GetTrimmingOptions(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
    all, err := s.trimmingOption.FindAll(ctx, clinicID)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to get trimming options")
    }
    result := make([]model.TrimmingOption, 0, len(all))
    for i := range all {
        if all[i].IsActive {
            result = append(result, all[i])
        }
    }
    return result, nil
}
```

#### 1-e. `CreateReservation` でトリミング詳細を作成

`CreateReservationInput` に以下を追加（`reservation_validators.go`）:

```go
type CreateReservationInput struct {
    ClinicID          uint64
    CustomerID        uint64
    ReservationTypeID uint64
    StaffID           uint64 // 0 = 指名なし
    Date              time.Time
    StartTime         string // "HHMM"
    EndTime           string // "HHMM"
    CustomerFields    []byte
    RequestText       string
    Settings          *model.LineReservationSetting
    // ★ 追加: トリミング詳細（category='trimming' の場合のみ使用）
    TrimmingCourseID *uint64
    TrimmingOptionIDs []uint64
    TrimmingStyleRequest string
}
```

`liffService.CreateReservation` の実装末尾に追記:

```go
// appointment 作成完了後、category='trimming' の場合に trimming_detail を作成
if appt.ReservationType != nil && appt.ReservationType.Category == model.ReservationTypeCategoryTrimming {
    detail := &model.AppointmentTrimmingDetail{
        ClinicID:      clinicID,
        AppointmentID: appt.ID,
        CourseID:      input.TrimmingCourseID,
        StyleRequest:  input.TrimmingStyleRequest,
        BWUnit:        model.BodyWeightUnitKg, // デフォルト
    }
    if err := s.trimmingDetail.Create(ctx, detail); err != nil {
        // best-effort: trimming_detail 作成失敗は予約自体をロールバックしない
        // ただし警告ログを出力する
        slog.WarnContext(ctx, "failed to create trimming detail",
            slog.String("error", err.Error()),
            slog.Uint64("appointment_id", appt.ID))
    } else if len(input.TrimmingOptionIDs) > 0 {
        if err := s.trimmingDetail.SetOptions(ctx, appt.ID, input.TrimmingOptionIDs); err != nil {
            slog.WarnContext(ctx, "failed to set trimming options",
                slog.String("error", err.Error()),
                slog.Uint64("appointment_id", appt.ID))
        }
    }
}
```

**注意**: `appt.ReservationType` がロードされているか確認が必要。
`CreateReservation` 内の appointment 取得後に `Preload("ReservationType")` が実行されていない場合は、
`s.reservationRepo.FindByID` で再取得するか、あるいは `typeLiffRepo.FindByID` で事前に category を取得しておくこと。

### 2. `backend/internal/service/service.go` — NewLiffService 引数追加

```go
// Before（現在の引数順序を確認して合わせること）:
Liff: NewLiffService(repos.LineReservationSetting, ..., validators, notifier)

// After:
Liff: NewLiffService(
    repos.LineReservationSetting,
    // ... 既存引数 ...
    repos.TrimmingCourse,
    repos.TrimmingOption,
    repos.AppointmentTrimmingDetail,
    validators,
    notifier,
)
```

### 3. `backend/internal/handler/liff_request.go` — トリミングフィールド追加

```go
type liffCreateReservationRequest struct {
    TypeID         uint64          `json:"course_id"      binding:"required"`
    StaffID        uint64          `json:"staff_id"`
    Date           string          `json:"date"             binding:"required"` // "YYYY-MM-DD"
    StartTime      string          `json:"start_time"       binding:"required"` // "HHMM"
    EndTime        string          `json:"end_time"         binding:"required"` // "HHMM"
    CustomerFields json.RawMessage `json:"customer_fields"`
    RequestText    string          `json:"request_text"`
    // ★ 追加: トリミング詳細（category='trimming' の場合のみ送信）
    TrimmingCourseID     *uint64  `json:"trimming_course_id"`
    TrimmingOptionIDs    []uint64 `json:"trimming_option_ids"`
    TrimmingStyleRequest string   `json:"trimming_style_request"`
}
```

`liff_handler.go` の `CreateLiffReservation` で input 組み立て時に追加:

```go
input := &service.CreateReservationInput{
    // ... 既存フィールド ...
    TrimmingCourseID:     req.TrimmingCourseID,
    TrimmingOptionIDs:    req.TrimmingOptionIDs,
    TrimmingStyleRequest: req.TrimmingStyleRequest,
}
```

### 4. `backend/internal/handler/liff_handler.go` — 新エンドポイント追加

```go
// GetLiffTrimmingCourses はLIFF向けトリミングコース一覧を返す。
func (h *Handler) GetLiffTrimmingCourses(c *gin.Context) {
    clinicID, ok := extractLiffClinicID(c)
    if !ok {
        return
    }
    courses, err := h.svc.Liff.GetTrimmingCourses(c.Request.Context(), clinicID)
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, courses)
}

// GetLiffTrimmingOptions はLIFF向けトリミングオプション一覧を返す。
func (h *Handler) GetLiffTrimmingOptions(c *gin.Context) {
    clinicID, ok := extractLiffClinicID(c)
    if !ok {
        return
    }
    options, err := h.svc.Liff.GetTrimmingOptions(c.Request.Context(), clinicID)
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, options)
}
```

### 5. `backend/internal/handler/reservation_line_routes.go` — ルート追加

```go
// authed グループに追加（認証必須）
authed.GET("/courses", h.GetLiffTypes)
authed.GET("/trimming-courses", h.GetLiffTrimmingCourses) // ★ 追加
authed.GET("/trimming-options", h.GetLiffTrimmingOptions) // ★ 追加
authed.GET("/staffs", h.GetLiffStaffs)
// ... 他のルートは変更なし ...
```

## 新 API エンドポイント仕様

### `GET /api/liff/:clinicId/trimming-courses`

**認証**: LINE IDトークン必須

**レスポンス例**:
```json
[
  {
    "id": 1,
    "clinic_id": 1,
    "name": "フルコース",
    "price": 5000,
    "is_active": true,
    "description": "シャンプー＋カット＋爪切り込み",
    "target_size": "small",
    "duration": 90,
    "sort_order": 0
  }
]
```

### `GET /api/liff/:clinicId/trimming-options`

**認証**: LINE IDトークン必須

**レスポンス例**:
```json
[
  {
    "id": 1,
    "clinic_id": 1,
    "name": "爪切り",
    "price": 500,
    "is_active": true,
    "description": "",
    "is_combinable": true,
    "duration": 10,
    "sort_order": 0
  }
]
```

### `POST /api/liff/:clinicId/reservations`（拡張後）

**追加フィールド**:
```json
{
  "course_id": 1,
  "staff_id": 0,
  "date": "2026-05-10",
  "start_time": "1000",
  "end_time": "1100",
  "trimming_course_id": 1,
  "trimming_option_ids": [1, 2],
  "trimming_style_request": "足回り短めにお願いします"
}
```

`trimming_course_id`, `trimming_option_ids`, `trimming_style_request` は任意フィールド。
`category = 'general'` の予約区分に対してこれらのフィールドを送信しても無視される。

## フロントエンド影響

- LIFF フロントエンド: `FE-254` で `/trimming-courses`, `/trimming-options` エンドポイントを呼び出す新ステップを実装
- `POST /api/liff/:clinicId/reservations` の追加フィールドは `FE-254` で送信

## 完了条件

- [ ] `LiffService` インターフェースに `GetTrimmingCourses`, `GetTrimmingOptions` が追加されている
- [ ] `liffService` struct に `trimmingCourse`, `trimmingOption`, `trimmingDetail` フィールドが追加されている
- [ ] `NewLiffService` に3つの新リポジトリ引数が追加されている
- [ ] `service.go` の `NewLiffService` 呼び出しが更新されている
- [ ] `GET /api/liff/:clinicId/trimming-courses` が動作する（is_active=true のみ返す）
- [ ] `GET /api/liff/:clinicId/trimming-options` が動作する（is_active=true のみ返す）
- [ ] `liffCreateReservationRequest` に `trimming_course_id`, `trimming_option_ids`, `trimming_style_request` が追加されている
- [ ] `CreateReservationInput` に対応フィールドが追加されている
- [ ] `CreateReservation` が `category='trimming'` の場合に `appointment_trimming_details` を作成する
- [ ] `docker compose exec backend go build ./...` が通る
- [ ] `docker compose exec backend go test ./...` が通る

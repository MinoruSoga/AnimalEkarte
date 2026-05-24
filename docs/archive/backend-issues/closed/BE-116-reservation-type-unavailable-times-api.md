# BE-116: 予約不可時間 & 職種紐付け 管理 API 実装

**Status**: Closed（2026-05-19 実装済み確認）
**Priority**: High
**Affects**: 予約区分マスタ管理 API
**Date Created**: 2026-04-16
**Related**: TASK-001, BE-115（先に完了必要）, FE-252

## Summary

BE-115 で追加した2テーブルに対して、カルテ側管理画面（FE-252）が利用する CRUD API を実装する。
既存の予約区分ハンドラ（`reservation_type_handler.go`）に新エンドポイントを追加する。

## 現状のコード

```go
// backend/internal/handler/reservation_type_handler.go（既存エンドポイント）
// GET    /v1/clinics/:clinicId/reservation-types
// POST   /v1/clinics/:clinicId/reservation-types
// GET    /v1/clinics/:clinicId/reservation-types/:id
// PUT    /v1/clinics/:clinicId/reservation-types/:id
// DELETE /v1/clinics/:clinicId/reservation-types/:id
```

```go
// backend/cmd/api/main.go（既存ルーティング）
// reservationTypes.GET("", h.ReservationType.List)
// reservationTypes.POST("", h.ReservationType.Create)
// ...
```

## 必要な変更

### 1. 新規エンドポイント設計

```
# 予約不可時間
GET    /v1/clinics/:clinicId/reservation-types/:id/unavailable-times
POST   /v1/clinics/:clinicId/reservation-types/:id/unavailable-times
DELETE /v1/clinics/:clinicId/reservation-types/:id/unavailable-times/:unavailableTimeId

# 職種紐付け
GET    /v1/clinics/:clinicId/reservation-types/:id/occupations
POST   /v1/clinics/:clinicId/reservation-types/:id/occupations
DELETE /v1/clinics/:clinicId/reservation-types/:id/occupations/:occupationId
```

### 2. Request 定義

`backend/internal/handler/reservation_type_request.go` に追記:

```go
// CreateUnavailableTimeRequest は予約不可時間の作成リクエスト
type CreateUnavailableTimeRequest struct {
    UnavailableType string  `json:"unavailable_type" binding:"required,oneof=weekly specific"`
    DayOfWeek       *int8   `json:"day_of_week"`      // weekly 時: 0-6
    SpecificDate    *string `json:"specific_date"`    // specific 時: "YYYY-MM-DD"
    StartTime       string  `json:"start_time"        binding:"required"` // "HH:MM"
    EndTime         string  `json:"end_time"          binding:"required"` // "HH:MM"
}

// LinkOccupationRequest は職種紐付けリクエスト
type LinkOccupationRequest struct {
    OccupationID uint64 `json:"occupation_id" binding:"required"`
}
```

### 3. Repository 追加

**`backend/internal/repository/repositories.go` の `Repositories` struct と `NewRepositories()` も更新する**（これを忘れると `go build` が通らない）:

```go
// repositories.go の Repositories struct に追加（LINE予約セクションの末尾）
ReservationTypeUnavailableTime ReservationTypeUnavailableTimeRepository // ★ NEW
ReservationTypeOccupation      ReservationTypeOccupationRepository      // ★ NEW

// NewRepositories() に追加
ReservationTypeUnavailableTime: NewReservationTypeUnavailableTimeRepository(db), // ★ NEW
ReservationTypeOccupation:      NewReservationTypeOccupationRepository(db),      // ★ NEW
```

`backend/internal/repository/reservation_type_unavailable_time_repository.go`（新規ファイル）:

```go
// ReservationTypeUnavailableTimeRepository は予約不可時間の永続化インターフェース
type ReservationTypeUnavailableTimeRepository interface {
    // FindAll は指定予約区分の予約不可時間を全件返す（LIFF・管理API 両方が使用）
    FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error)
    Create(ctx context.Context, t *model.ReservationTypeUnavailableTime) error
    // Delete は id で物理削除する（論理削除なし）
    Delete(ctx context.Context, clinicID, id uint64) error
}
```

`backend/internal/repository/reservation_type_occupation_repository.go`（新規ファイル）:

```go
// ReservationTypeOccupationRepository は職種紐付けの永続化インターフェース
type ReservationTypeOccupationRepository interface {
    // FindAll は Occupation を Preload して返す（ListOccupations レスポンスの occupation フィールドに必要）
    FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeOccupation, error)
    Create(ctx context.Context, o *model.ReservationTypeOccupation) error
    // Delete は物理削除（論理削除なし）
    Delete(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) error
    // CountWorkingStaff は指定日に対応職種のスタッフが何人出勤しているかを返す（LIFF 専用）
    // shift_type が 'full' / 'morning' / 'afternoon' のスタッフのみカウント
    CountWorkingStaff(ctx context.Context, clinicID, reservationTypeID uint64, date time.Time) (int64, error)
}
```

`FindAll` の実装は `Preload("Occupation")` を含めること:
```go
// FindAll 実装例（抜粋）
db.WithContext(ctx).
    Preload("Occupation").
    Where("clinic_id = ? AND reservation_type_id = ?", clinicID, reservationTypeID).
    Find(&results)
```

`CountWorkingStaff` の SQL（GORM Raw クエリ用・`?` プレースホルダ）:

```sql
-- 注意事項:
-- 1. staffs テーブルは clinic_id カラムを持たない（FindOnDutyStaffs/shift_entry_repository.go:157 参照）
--    → clinic フィルタは shift_entries.clinic_id で担保する
-- 2. shift_entries は物理削除テーブル（deleted_at カラムなし）
-- 3. 日付は YYYY-MM-DD 文字列で渡す（date.In(jstLocation()).Format("2006-01-02")）
SELECT COUNT(DISTINCT se.staff_id)
FROM reservation_type_occupations rto
JOIN staffs s ON s.occupation_id = rto.occupation_id
    AND s.deleted_at IS NULL
JOIN shift_entries se ON se.staff_id = s.id
    AND se.clinic_id = ?       -- clinicID
    AND se.date = ?            -- date string "YYYY-MM-DD"
    AND se.shift_type NOT IN ('off', 'paid_leave')
WHERE rto.clinic_id = ?        -- clinicID
  AND rto.reservation_type_id = ?  -- reservationTypeID
```

実装例（`reservation_type_occupation_repository.go`）:

```go
func (r *reservationTypeOccupationRepository) CountWorkingStaff(
    ctx context.Context, clinicID, reservationTypeID uint64, date time.Time,
) (int64, error) {
    dateStr := date.In(jstLocation()).Format("2006-01-02") // JST 日付文字列で比較
    var count int64
    err := r.db.WithContext(ctx).Raw(`
        SELECT COUNT(DISTINCT se.staff_id)
        FROM reservation_type_occupations rto
        JOIN staffs s ON s.occupation_id = rto.occupation_id AND s.deleted_at IS NULL
        JOIN shift_entries se ON se.staff_id = s.id
            AND se.clinic_id = ?
            AND se.date = ?
            AND se.shift_type NOT IN ('off', 'paid_leave')
        WHERE rto.clinic_id = ?
          AND rto.reservation_type_id = ?
    `, clinicID, dateStr, clinicID, reservationTypeID).Scan(&count).Error
    return count, apperrors.FromGORM(err, "count_working_staff", "")
}
```

`jstLocation()` は `liff_service.go` の同名ヘルパーと同一（`time.LoadLocation("Asia/Tokyo")`）。
リポジトリ内で再実装するか、`internal/util` 等に切り出す。

### 4. Service 追加

`backend/internal/service/reservation_type_service.go` に追加:

```go
// インターフェースに新メソッドを追加
type ReservationTypeService interface {
    // ... 既存メソッド（List/GetByID/Create/Update/Delete/Reorder）...
    ListUnavailableTimes(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error)
    CreateUnavailableTime(ctx context.Context, clinicID, reservationTypeID uint64, input CreateUnavailableTimeInput) (*model.ReservationTypeUnavailableTime, error)
    DeleteUnavailableTime(ctx context.Context, clinicID, id uint64) error

    ListOccupations(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeOccupation, error)
    LinkOccupation(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) (*model.ReservationTypeOccupation, error)
    UnlinkOccupation(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) error
}

// reservationTypeService struct に2フィールド追加（現状は repo + reservationRepo の2フィールド）
type reservationTypeService struct {
    repo                repository.ReservationTypeRepository
    reservationRepo     repository.ReservationRepository
    unavailableTimeRepo repository.ReservationTypeUnavailableTimeRepository // ★ NEW
    occupationRepo      repository.ReservationTypeOccupationRepository      // ★ NEW
}

// NewReservationTypeService のシグネチャを変更（service.go の呼び出し箇所も更新が必要）
func NewReservationTypeService(
    repo            repository.ReservationTypeRepository,
    reservationRepo repository.ReservationRepository,
    unavailableTimeRepo repository.ReservationTypeUnavailableTimeRepository, // ★ NEW
    occupationRepo      repository.ReservationTypeOccupationRepository,      // ★ NEW
) ReservationTypeService {
    return &reservationTypeService{
        repo:                repo,
        reservationRepo:     reservationRepo,
        unavailableTimeRepo: unavailableTimeRepo,
        occupationRepo:      occupationRepo,
    }
}

type CreateUnavailableTimeInput struct {
    UnavailableType string
    DayOfWeek       *int8
    SpecificDate    *time.Time
    StartTime       string
    EndTime         string
}
```

`backend/internal/service/service.go` の `NewReservationTypeService` 呼び出し箇所も更新:

```go
// service.go の NewServices 関数内（現状: NewReservationTypeService(repos.ReservationType, repos.Reservation)）
ReservationType: NewReservationTypeService(
    repos.ReservationType,
    repos.Reservation,
    repos.ReservationTypeUnavailableTime, // ★ NEW（BE-115 で追加したリポジトリ）
    repos.ReservationTypeOccupation,      // ★ NEW
),
```

バリデーション（`service/validators.go` に追加）:
- `weekly` の場合: `DayOfWeek` が 0-6 の範囲内
- `specific` の場合: `SpecificDate` が非 nil
- `StartTime < EndTime`（HH:MM 文字列比較。VARCHAR(5) 保存のため "12:00" < "13:00" は lexicographic に正しく機能する）
- 重複チェック: 同種別・同曜日 or 同日付で時間帯が重複する設定は 409 を返す
  - `weekly` 同士: 同じ `day_of_week` で時間帯が交差する場合は禁止
  - `specific` 同士: 同じ `specific_date` で時間帯が交差する場合は禁止
  - `weekly` と `specific` の混在: 許可（LIFF側で特定日が曜日より優先される）

`LinkOccupation` のバリデーション:
- クリニック越境防止: `OccupationRepository.FindByID(ctx, clinicID, occupationID)` で事前確認が理想だが、
  `occupation_id REFERENCES occupations(id)` の FK 制約 + `reservation_type_id` の FK 制約により
  **存在しない ID や論理的に不正な組み合わせは DB レベルで弾かれる**。
  `apperrors.FromGORM` が FK 違反を 409 に変換するため、最低限の安全性は確保される。
  厳密なクリニック所有チェックが必要な場合は、`reservationTypeService` に5番目のフィールド
  `occupationMasterRepo repository.OccupationRepository` を追加し、`NewReservationTypeService` のシグネチャと
  `service.go` の DI 配線も同様に更新すること（`OccupationRepository` は `repos.Occupation` として既に存在）。

### 5. Handler 追加

`backend/internal/handler/reservation_type_response.go`（新規または既存ファイルに追記）にレスポンス DTO を定義する。
`SpecificDate` は `*time.Time` だが JSON では `"YYYY-MM-DD"` 文字列で返す必要があるため、
`clinic_holiday_handler.go:22-31` と同じパターンでレスポンス変換関数を用意する:

```go
// reservation_type_response.go に追加
type unavailableTimeResponse struct {
    ID                uint64  `json:"id"`
    ClinicID          uint64  `json:"clinic_id"`
    ReservationTypeID uint64  `json:"reservation_type_id"`
    UnavailableType   string  `json:"unavailable_type"`
    DayOfWeek         *int8   `json:"day_of_week,omitempty"`
    SpecificDate      *string `json:"specific_date,omitempty"` // "YYYY-MM-DD" 形式
    StartTime         string  `json:"start_time"`
    EndTime           string  `json:"end_time"`
}

func toUnavailableTimeResponse(t *model.ReservationTypeUnavailableTime) unavailableTimeResponse {
    resp := unavailableTimeResponse{
        ID:                t.ID,
        ClinicID:          t.ClinicID,
        ReservationTypeID: t.ReservationTypeID,
        UnavailableType:   string(t.UnavailableType),
        DayOfWeek:         t.DayOfWeek,
        StartTime:         t.StartTime,
        EndTime:           t.EndTime,
    }
    if t.SpecificDate != nil {
        s := t.SpecificDate.UTC().Format("2006-01-02") // DATE型はUTC午前0時で格納される
        resp.SpecificDate = &s
    }
    return resp
}
```

`backend/internal/handler/reservation_type_handler.go` に追加:

```go
func (h *ReservationTypeHandler) ListUnavailableTimes(c *gin.Context) { ... }
func (h *ReservationTypeHandler) CreateUnavailableTime(c *gin.Context) { ... }
func (h *ReservationTypeHandler) DeleteUnavailableTime(c *gin.Context) { ... }
func (h *ReservationTypeHandler) ListOccupations(c *gin.Context) { ... }
func (h *ReservationTypeHandler) LinkOccupation(c *gin.Context) { ... }
func (h *ReservationTypeHandler) UnlinkOccupation(c *gin.Context) { ... }
```

エラーハンドリング: 全ハンドラで `RespondError(c, err)` を使用。`ShouldBindJSON` エラーは `apperrors.WrapInvalidInput(parseBindError(err))`。

`CreateUnavailableTime` ハンドラで `specific_date` 文字列 → `*time.Time` への変換が必要:

```go
// CreateUnavailableTime ハンドラ内の型変換（抜粋）
var input service.CreateUnavailableTimeInput
input.UnavailableType = req.UnavailableType
input.DayOfWeek       = req.DayOfWeek
input.StartTime       = req.StartTime
input.EndTime         = req.EndTime
if req.SpecificDate != nil {
    t, err := time.Parse("2006-01-02", *req.SpecificDate)
    if err != nil {
        RespondError(c, apperrors.WrapInvalidInput("specific_date must be YYYY-MM-DD"))
        return
    }
    input.SpecificDate = &t
}
```

### 6. ルーティング追加

`backend/cmd/api/main.go`:

```go
// 予約区分の詳細リソース
rtDetail := reservationTypes.Group("/:id")
{
    rtDetail.GET("/unavailable-times", h.ReservationType.ListUnavailableTimes)
    rtDetail.POST("/unavailable-times", h.ReservationType.CreateUnavailableTime)
    rtDetail.DELETE("/unavailable-times/:unavailableTimeId", h.ReservationType.DeleteUnavailableTime)

    rtDetail.GET("/occupations", h.ReservationType.ListOccupations)
    rtDetail.POST("/occupations", h.ReservationType.LinkOccupation)
    rtDetail.DELETE("/occupations/:occupationId", h.ReservationType.UnlinkOccupation)
}
```

## API レスポンス形式

```json
// GET /unavailable-times
{
  "data": [
    {
      "id": 1,
      "unavailable_type": "weekly",
      "day_of_week": 1,
      "start_time": "12:00",
      "end_time": "13:00"
    },
    {
      "id": 2,
      "unavailable_type": "specific",
      "specific_date": "2026-05-10",
      "start_time": "09:00",
      "end_time": "12:00"
    }
  ]
}

// GET /occupations
{
  "data": [
    { "id": 1, "occupation_id": 3, "occupation": { "id": 3, "name": "トリマー" } }
  ]
}
```

## 完了条件

- [ ] 6エンドポイントが正常に動作する
- [ ] `weekly` / `specific` バリデーションが機能する
- [ ] 重複時間帯登録時に 409 が返る
- [ ] 存在しない予約区分 ID を指定すると 404 が返る
- [ ] `NewReservationTypeService` のシグネチャが4引数に変更されている（`service.go` の DI 配線も更新済み）
- [ ] `go build ./...` が通る
- [ ] `go test ./internal/service/...` が通る（サービス単体テスト追加）

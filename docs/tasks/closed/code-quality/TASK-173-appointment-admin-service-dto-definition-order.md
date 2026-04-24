# TASK-173: appointment_admin_service.go — interface が DTO より前に定義されている

## 優先度
Medium

## 対象ファイル
`backend/internal/service/appointment_admin_service.go`

## 問題概要
`ReservationAdminService interface` が `CreateReservationAdminInput` DTO より
**前に定義**されている。規約の正しい定義順序に違反している。

規約の正しい順序:
```
CreateXxxInput      ← 先に定義
UpdateXxxInput      ← (あれば)
const / helper      ← (あれば)
type XxxService interface  ← interface は後
type xxxService struct
func NewXxxService(...)
func methods...
```

## 現状コード（appointment_admin_service.go 抜粋）

```go
// line 14: ReservationAdminService interface (NG: DTO より先)
type ReservationAdminService interface {
    ListByMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.Reservation, error)
    ListByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.Reservation, error)
    Create(ctx context.Context, clinicID uint64, input *CreateReservationAdminInput) (*model.Reservation, error)
    Delete(ctx context.Context, clinicID, id uint64) error
}

// line 22: CreateReservationAdminInput (NG: interface の後に来ている)
type CreateReservationAdminInput struct {
    StartTime         time.Time
    EndTime           time.Time
    OwnerID           *uint64
    PetID             *uint64
    VisitType         string
    ReservationTypeID uint64
    DoctorID          *uint64
    IsDesignated      bool
    Notes             string
    CreatedBy         *uint64
    LineCustomerID    *uint64
    IsStaffDelegated  bool
    CustomerFields    []byte
}
```

## 修正後コード（定義順序）

```go
// CreateReservationAdminInput は管理者手動予約の入力データ
type CreateReservationAdminInput struct {
    StartTime         time.Time
    EndTime           time.Time
    OwnerID           *uint64
    PetID             *uint64
    VisitType         string
    ReservationTypeID uint64
    DoctorID          *uint64
    IsDesignated      bool
    Notes             string
    CreatedBy         *uint64
    LineCustomerID    *uint64
    IsStaffDelegated  bool
    CustomerFields    []byte
}

// ReservationAdminService は管理者向け予約管理のビジネスロジックインターフェース
type ReservationAdminService interface {
    ListByMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.Reservation, error)
    ListByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.Reservation, error)
    Create(ctx context.Context, clinicID uint64, input *CreateReservationAdminInput) (*model.Reservation, error)
    Delete(ctx context.Context, clinicID, id uint64) error
}

type reservationAdminService struct { ... }

func NewReservationAdminService(...) ReservationAdminService { ... }

// ... メソッド実装
```

## 影響範囲
コンパイル・動作への影響はなし。コードの可読性・規約統一のみ。

## 対応方針
`CreateReservationAdminInput` を `ReservationAdminService interface` の前に移動する。

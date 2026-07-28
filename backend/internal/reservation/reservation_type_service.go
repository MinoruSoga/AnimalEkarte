package reservation

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Input DTOs ----

// CreateReservationTypeInput はサービス種別作成のための入力データ
type CreateReservationTypeInput struct {
	Name        string
	Color       string
	IsActive    bool
	Description string
	SortOrder   int
	GroupID     *uint64
	ParentID    *uint64
	Category    string

	// LINE予約用フィールド
	ReservationDisplayName string
	DurationMinutes        *int
	MaxConcurrent          *int
	ShortName              string
	ShowShortName          bool
	ReservationVisible     *bool
	ReservationComment     string
	ReservationImageURL    string
	ReservationDayOption   string
	IsInternal             bool
}

// UpdateReservationTypeInput はサービス種別更新のための入力データ（ポインタ型でゼロ値を区別する）
type UpdateReservationTypeInput struct {
	Name          *string
	Color         *string
	IsActive      *bool
	Description   *string
	SortOrder     *int
	GroupID       *uint64
	ClearGroupID  bool // true のとき group_id を NULL にクリアする
	ParentID      *uint64
	ClearParentID bool // true のとき parent_id を NULL にクリアする
	Category      *string

	// LINE予約用フィールド
	ReservationDisplayName *string
	DurationMinutes        *int
	MaxConcurrent          *int
	ClearMaxConcurrent     bool // true のとき max_concurrent を NULL にクリアする
	ShortName              *string
	ShowShortName          *bool
	ReservationVisible     *bool
	ReservationComment     *string
	ReservationImageURL    *string
	ReservationDayOption   *string
	IsInternal             *bool
}

// ---- DB column constants ----

// ---- ReservationTypeService ----

// CreateUnavailableTimeInput は予約不可時間の作成入力DTO
type CreateUnavailableTimeInput struct {
	UnavailableType string
	DayOfWeek       *int8
	SpecificDate    *time.Time
	StartTime       string
	EndTime         string
}

// CreateAvailableSlotInput は予約可能開始時刻の作成入力DTO
type CreateAvailableSlotInput struct {
	AvailableType string
	DayOfWeek     *int8
	SpecificDate  *time.Time
	StartTime     string
	IsActive      *bool
}

// ReservationTypeCoreService は予約種別の CRUD・Reorder を担う。
type ReservationTypeCoreService interface { //nolint:revive // ReservationType is a domain entity name, cannot avoid stutter
	List(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
	Create(ctx context.Context, clinicID uint64, input *CreateReservationTypeInput) (*model.ReservationType, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeInput) (*model.ReservationType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

// ReservationTypeUnavailableTimeService は予約不可時間帯の管理を担う。
type ReservationTypeUnavailableTimeService interface { //nolint:revive // ReservationType is a domain entity name, cannot avoid stutter
	ListUnavailableTimes(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error)
	CreateUnavailableTime(ctx context.Context, clinicID, reservationTypeID uint64, input CreateUnavailableTimeInput) (*model.ReservationTypeUnavailableTime, error)
	DeleteUnavailableTime(ctx context.Context, clinicID, reservationTypeID, id uint64) error
}

// ReservationTypeAvailableSlotService は予約可能開始時刻の管理を担う。
type ReservationTypeAvailableSlotService interface { //nolint:revive // ReservationType is a domain entity name, cannot avoid stutter
	ListAvailableSlots(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error)
	CreateAvailableSlot(ctx context.Context, clinicID, reservationTypeID uint64, input CreateAvailableSlotInput) (*model.ReservationTypeAvailableSlot, error)
	DeleteAvailableSlot(ctx context.Context, clinicID, reservationTypeID, id uint64) error
}

// ReservationTypeOccupationService は診療科目ひもづけを担う。
type ReservationTypeOccupationService interface { //nolint:revive // ReservationType is a domain entity name, cannot avoid stutter
	ListOccupations(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeOccupation, error)
	LinkOccupation(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) (*model.ReservationTypeOccupation, error)
	UnlinkOccupation(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) error
}

// ReservationTypeService は3つのサブインターフェースを合成した完全インターフェース。
// handler 層はこのインターフェースを通じてすべての操作を行う。
type ReservationTypeService interface { //nolint:revive // ReservationType is a domain entity name, cannot avoid stutter
	ReservationTypeCoreService
	ReservationTypeUnavailableTimeService
	ReservationTypeAvailableSlotService
	ReservationTypeOccupationService
}

type reservationTypeService struct {
	repo                ReservationTypeRepository
	unavailableTimeRepo ReservationTypeUnavailableTimeRepository
	availableSlotRepo   ReservationTypeAvailableSlotRepository
	occupationRepo      ReservationTypeOccupationRepository
	baseOccupationRepo  occupationFinder
	groupRepo           ReservationTypeGroupRepository
}

// BE-refactor.md X-14/U6b: groupRepo は GroupID のクロステナント write 防止(FindByID
// 所有権検証)に使う。variadic availableSlotRepo より前の必須引数として追加し(全呼び出し元更新)、
// 未設定(nil)の場合は既存の trimmingCourseRepo 等と同型の nil 許容パターンでチェックをスキップする。
func NewReservationTypeService(
	repo ReservationTypeRepository,
	unavailableTimeRepo ReservationTypeUnavailableTimeRepository,
	occupationRepo ReservationTypeOccupationRepository,
	baseOccupationRepo occupationFinder,
	groupRepo ReservationTypeGroupRepository,
	availableSlotRepo ...ReservationTypeAvailableSlotRepository,
) ReservationTypeService {
	var slotRepo ReservationTypeAvailableSlotRepository
	if len(availableSlotRepo) > 0 {
		slotRepo = availableSlotRepo[0]
	}
	return &reservationTypeService{
		repo:                repo,
		unavailableTimeRepo: unavailableTimeRepo,
		availableSlotRepo:   slotRepo,
		occupationRepo:      occupationRepo,
		baseOccupationRepo:  baseOccupationRepo,
		groupRepo:           groupRepo,
	}
}

package service

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
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
	Category    string

	// LINE予約用フィールド
	ReservationDisplayName string
	DurationMinutes        *int
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
	Name         *string
	Color        *string
	IsActive     *bool
	Description  *string
	SortOrder    *int
	GroupID      *uint64
	ClearGroupID bool // true のとき group_id を NULL にクリアする
	Category     *string

	// LINE予約用フィールド
	ReservationDisplayName *string
	DurationMinutes        *int
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
	ReservationTypeOccupationService
}

type reservationTypeService struct {
	repo                repository.ReservationTypeRepository
	unavailableTimeRepo repository.ReservationTypeUnavailableTimeRepository
	occupationRepo      repository.ReservationTypeOccupationRepository
	baseOccupationRepo  repository.OccupationRepository
}

func NewReservationTypeService(
	repo repository.ReservationTypeRepository,
	unavailableTimeRepo repository.ReservationTypeUnavailableTimeRepository,
	occupationRepo repository.ReservationTypeOccupationRepository,
	baseOccupationRepo repository.OccupationRepository,
) ReservationTypeService {
	return &reservationTypeService{
		repo:                repo,
		unavailableTimeRepo: unavailableTimeRepo,
		occupationRepo:      occupationRepo,
		baseOccupationRepo:  baseOccupationRepo,
	}
}

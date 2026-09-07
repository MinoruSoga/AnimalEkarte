package reservation

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// AllowedReservationRoutes は予約経路の許可値ホワイトリスト（FEAT-381-2 Commit 3）。
var AllowedReservationRoutes = map[string]struct{}{
	"line": {}, "phone": {}, "reception": {}, "exam_room": {}, "record_shortcut": {},
}

const colReservationRoute = "reservation_route"
const AllowedReservationRoutesMessage = "reservation_route must be one of: line, phone, reception, exam_room, record_shortcut"

// UpdateReservationRouteInput は予約経路更新の入力DTO（FEAT-381-2 Commit 3）。
type UpdateReservationRouteInput struct{ Route string }

// CreateManualReservationInput は管理画面からの予約作成入力 DTO。
type CreateManualReservationInput struct {
	ClinicID          uint64
	StartTime         time.Time
	EndTime           time.Time
	OwnerID           *uint64
	PetID             *uint64
	VisitType         model.VisitType
	ReservationTypeID uint64
	DoctorID          *uint64
	IsDesignated      bool
	Status            model.ReservationStatus
	Notes             string
	Source            model.ReservationSource
	CreatedBy         *uint64
	ReservationRoute  *string
}

// UpdateReservationInput は予約更新のサービス入力 DTO
type ReservationBatchPet struct {
	OwnerID uint64
	PetID   uint64
}

type UpdateReservationInput struct {
	StartTime         *time.Time
	EndTime           *time.Time
	OwnerID           *uint64
	PetID             *uint64
	VisitType         *model.VisitType
	ReservationTypeID *uint64
	DoctorID          *uint64
	IsDesignated      *bool
	Status            *model.ReservationStatus
	Notes             *string
}

func buildReservationUpdate(input *UpdateReservationInput) map[string]any {
	fields := make(map[string]any)
	if input.StartTime != nil {
		fields["start_time"] = *input.StartTime
	}
	if input.EndTime != nil {
		fields["end_time"] = *input.EndTime
	}
	if input.OwnerID != nil {
		if *input.OwnerID == 0 {
			fields["owner_id"] = nil // 0 は「飼主クリア」として NULL に設定
		} else {
			fields["owner_id"] = *input.OwnerID
		}
	}
	if input.PetID != nil {
		if *input.PetID == 0 {
			fields["pet_id"] = nil // 0 は「ペットクリア」として NULL に設定
		} else {
			fields["pet_id"] = *input.PetID
		}
	}
	if input.VisitType != nil {
		fields["visit_type"] = *input.VisitType
	}
	if input.ReservationTypeID != nil {
		fields["reservation_type_id"] = *input.ReservationTypeID
	}
	if input.DoctorID != nil {
		if *input.DoctorID == 0 {
			fields["doctor_id"] = nil // 0 は「医師未指定に変更」として NULL に設定
		} else {
			fields["doctor_id"] = *input.DoctorID
		}
	}
	if input.IsDesignated != nil {
		fields["is_designated"] = *input.IsDesignated
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	if input.Notes != nil {
		fields["notes"] = *input.Notes
	}
	return fields
}

type ReservationService interface {
	// List は指定した複数医院 (#86 拠点横断) の予約一覧を返す。clinicIDs はハンドラ層で所属検証済みであること。
	List(ctx context.Context, clinicIDs []uint64, page, limit int, date, startDate, endDate *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	// GetByIDForClinics は複数医院スコープで予約を1件取得する (#86 詳細画面拠点横断)。clinicIDs はハンドラ層で所属検証済みであること。
	GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Reservation, error)
	Create(ctx context.Context, input *CreateManualReservationInput) (*model.Reservation, error)
	CreateBatch(ctx context.Context, input *CreateManualReservationInput, pets []ReservationBatchPet) ([]model.Reservation, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationInput) (*model.Reservation, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	UpdateReservationRoute(ctx context.Context, clinicID, id uint64, input UpdateReservationRouteInput) (*model.Reservation, error)
}

type reservationService struct {
	repo                 ReservationRepository
	typeRepo             reservationTypeFinder
	tx                   Transactor
	reservationStaffRepo ReservationStaffWriteGuard
	unavailableTimeRepo  ReservationTypeUnavailableTimeRepository
	availableSlotRepo    ReservationTypeAvailableSlotRepository
	settingFinder        lineReservationSettingFinder
	holidayFinder        clinicHolidayFinder
}

func NewReservationServiceWithAvailabilityAndType(repo ReservationRepository, typeRepo reservationTypeFinder, tx Transactor, reservationStaffRepo ReservationStaffWriteGuard, unavailableTimeRepo ReservationTypeUnavailableTimeRepository, availableSlotRepo ...ReservationTypeAvailableSlotRepository) ReservationService {
	var slotRepo ReservationTypeAvailableSlotRepository
	if len(availableSlotRepo) > 0 {
		slotRepo = availableSlotRepo[0]
	}
	return newReservationService(repo, typeRepo, tx, reservationStaffRepo, unavailableTimeRepo, slotRepo, nil, nil)
}

// NewReservationServiceWithLineSettings is the backward-compatible constructor
// that also injects LINE reservation settings for closed-day checks on constrained Create.
// holidayFinder is omitted (nil). Constrained Create/Update then fail closed with
// "clinic holiday lookup is required"; they do not skip clinic_holidays.
// Production composition must use NewReservationServiceWithClinicHolidays.
func NewReservationServiceWithLineSettings(
	repo ReservationRepository,
	typeRepo reservationTypeFinder,
	tx Transactor,
	reservationStaffRepo ReservationStaffWriteGuard,
	unavailableTimeRepo ReservationTypeUnavailableTimeRepository,
	availableSlotRepo ReservationTypeAvailableSlotRepository,
	settingFinder lineReservationSettingFinder,
) ReservationService {
	return newReservationService(repo, typeRepo, tx, reservationStaffRepo, unavailableTimeRepo, availableSlotRepo, settingFinder, nil)
}

// NewReservationServiceWithClinicHolidays injects LINE settings and clinic_holidays.
func NewReservationServiceWithClinicHolidays(
	repo ReservationRepository,
	typeRepo reservationTypeFinder,
	tx Transactor,
	reservationStaffRepo ReservationStaffWriteGuard,
	unavailableTimeRepo ReservationTypeUnavailableTimeRepository,
	availableSlotRepo ReservationTypeAvailableSlotRepository,
	settingFinder lineReservationSettingFinder,
	holidayFinder clinicHolidayFinder,
) ReservationService {
	return newReservationService(repo, typeRepo, tx, reservationStaffRepo, unavailableTimeRepo, availableSlotRepo, settingFinder, holidayFinder)
}

func newReservationService(
	repo ReservationRepository,
	typeRepo reservationTypeFinder,
	tx Transactor,
	reservationStaffRepo ReservationStaffWriteGuard,
	unavailableTimeRepo ReservationTypeUnavailableTimeRepository,
	availableSlotRepo ReservationTypeAvailableSlotRepository,
	settingFinder lineReservationSettingFinder,
	holidayFinder clinicHolidayFinder,
) *reservationService {
	return &reservationService{
		repo:                 repo,
		typeRepo:             typeRepo,
		tx:                   tx,
		reservationStaffRepo: reservationStaffRepo,
		unavailableTimeRepo:  unavailableTimeRepo,
		availableSlotRepo:    availableSlotRepo,
		settingFinder:        settingFinder,
		holidayFinder:        holidayFinder,
	}
}

// resolveFinalOwnerPet は現在値と PATCH 値を統合した最終 OwnerID/PetID を返す。
// 入力が 0 の場合はクリア（nil）として扱う（DoctorID と同様）。
func resolveFinalOwnerPet(current *model.Reservation, input *UpdateReservationInput) (ownerID, petID *uint64) {
	ownerID = current.OwnerID
	petID = current.PetID
	if input.OwnerID != nil {
		if *input.OwnerID == 0 {
			ownerID = nil
		} else {
			ownerID = input.OwnerID
		}
	}
	if input.PetID != nil {
		if *input.PetID == 0 {
			petID = nil
		} else {
			petID = input.PetID
		}
	}
	return ownerID, petID
}

// ValidateReservationOwnerPetLinksWithRepo は sharedkernel.ValidateReservationOwnerPetLinks への
// 既存呼び出し面互換 delegate（実装正本は sharedkernel・BE9-2D ⑤ Batch B 昇格。
// medicalrecord の hospitalization 系と恒久共有のため）。
func ValidateReservationOwnerPetLinksWithRepo(ctx context.Context, repo sharedkernel.OwnerPetLinkVerifier, clinicID uint64, ownerID, petID *uint64) error {
	return sharedkernel.ValidateReservationOwnerPetLinks(ctx, repo, clinicID, ownerID, petID)
}

// reservationDeceasedPetMessage は予約系 write が死亡ペットを拒否するときの安定メッセージ。
const reservationDeceasedPetMessage = "死亡したペットは予約できません"

// petByIDInClinicFinder は FindPetByIDInClinic を sharedkernel.PetByIDFinder へ適合する。
type petByIDInClinicFinder interface {
	FindPetByIDInClinic(ctx context.Context, clinicID, petID uint64) (*model.Pet, error)
}

type petByIDAdapter struct {
	repo petByIDInClinicFinder
}

func (a petByIDAdapter) FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	return a.repo.FindPetByIDInClinic(ctx, clinicID, id)
}

// ValidateReservationPetNotDeceased は petID が非 nil のとき死亡ペットへの予約 write を拒否する（#261 P0）。
// nil pet（未紐付け・クリア）は従来どおり許可する。
func ValidateReservationPetNotDeceased(ctx context.Context, repo petByIDInClinicFinder, clinicID uint64, petID *uint64) error {
	if petID == nil {
		return nil
	}
	return sharedkernel.ValidatePetNotDeceased(ctx, petByIDAdapter{repo: repo}, clinicID, *petID, reservationDeceasedPetMessage)
}

func (s *reservationService) List(ctx context.Context, clinicIDs []uint64, page, limit int, date, startDate, endDate *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error) {
	items, total, err := s.repo.FindAll(ctx, clinicIDs, page, limit, date, startDate, endDate, status, source, petID, ownerID)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list reservations")
	}
	return items, total, nil
}

func (s *reservationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation")
	}
	return result, nil
}

func (s *reservationService) GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Reservation, error) {
	result, err := s.repo.FindByIDForClinics(ctx, clinicIDs, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation for clinics")
	}
	return result, nil
}

func (s *reservationService) Create(ctx context.Context, input *CreateManualReservationInput) (*model.Reservation, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	if input.ReservationRoute != nil {
		if _, ok := AllowedReservationRoutes[*input.ReservationRoute]; !ok {
			return nil, apperrors.WrapInvalidInput(AllowedReservationRoutesMessage)
		}
	}
	enforceBookingConstraints := ShouldEnforceReservationBookingConstraints(input.Status, input.ReservationRoute)
	if enforceBookingConstraints {
		if err := ValidateReservationTypeAvailableTime(ctx, s.unavailableTimeRepo, input.ClinicID, input.ReservationTypeID, input.StartTime, input.EndTime); err != nil {
			return nil, err
		}
	}
	if err := validateTimeRange(input.StartTime, input.EndTime); err != nil {
		return nil, err
	}
	// BE-refactor.md X-14/U6b: クロステナント write 防止。reception/exam_room/record_shortcut
	// 等の shortcut 経路や確定済みステータスでは enforceBookingConstraints=false となり、
	// 下の CheckReservationTypeCapacity(容量チェック内の FindByID)がスキップされるため、
	// ReservationTypeID の所有権検証は経路・ステータスに関わらず常にここで行う
	// (reservation_validators.go の typeRepo と同型の nil 許容パターン)。
	if s.typeRepo != nil {
		if _, err := s.typeRepo.FindByID(ctx, input.ClinicID, input.ReservationTypeID); err != nil {
			return nil, apperrors.Wrap(err, "failed to verify reservation type ownership")
		}
	}
	reservation := &model.Reservation{
		ClinicID:          input.ClinicID,
		StartTime:         input.StartTime,
		EndTime:           input.EndTime,
		OwnerID:           input.OwnerID,
		PetID:             input.PetID,
		VisitType:         input.VisitType,
		ReservationTypeID: input.ReservationTypeID,
		DoctorID:          input.DoctorID,
		IsDesignated:      input.IsDesignated,
		Status:            input.Status,
		Notes:             input.Notes,
		Source:            input.Source,
		CreatedBy:         input.CreatedBy,
		ReservationRoute:  input.ReservationRoute,
	}

	// SELECT FOR UPDATE + トランザクションで競合を防止
	// LINE予約・電子カルテ予約・管理者手動予約すべてで同一テーブルを使用するため、
	// アプリケーションレベルでの排他制御が必要
	// BE-refactor.md X-9: 空き枠（既存行 0 件）は SELECT FOR UPDATE が何もロックしないため、
	// AcquireBookingLock（clinic 単位 advisory xact lock）で競合チェック～INSERT を直列化する。
	if s.tx == nil {
		return nil, apperrors.WrapInternalServerError("reservation transaction dependency is required")
	}
	if err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := ValidateReservationStaffCapability(ctx, s.reservationStaffRepo, reservation.ClinicID, reservation.DoctorID, reservation.ReservationTypeID); err != nil {
			return err
		}
		if err := ValidateReservationOwnerPetLinksWithRepo(ctx, s.repo, reservation.ClinicID, reservation.OwnerID, reservation.PetID); err != nil {
			return err
		}
		if err := ValidateReservationPetNotDeceased(ctx, s.repo, reservation.ClinicID, reservation.PetID); err != nil {
			return err
		}
		if enforceBookingConstraints {
			if err := s.validateCreateClosedDays(ctx, reservation.ClinicID, reservation.StartTime); err != nil {
				return err
			}
			if err := s.validateCreateClinicHoliday(ctx, reservation.ClinicID, reservation.StartTime); err != nil {
				return err
			}
			if err := s.repo.AcquireBookingLock(ctx, reservation.ClinicID); err != nil {
				return err
			}
			if err := CheckSlotConflict(ctx, s.repo, reservation.ClinicID, reservation.DoctorID, reservation.StartTime, reservation.EndTime, nil); err != nil {
				return err
			}
			if err := CheckReservationTypeCapacity(ctx, s.repo, s.typeRepo, reservation.ClinicID, reservation.ReservationTypeID, reservation.StartTime, nil); err != nil {
				return err
			}
		}
		return s.repo.Create(ctx, reservation)
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to create reservation")
	}

	slog.InfoContext(ctx, "reservation created",
		slog.Uint64("reservation_id", reservation.ID),
		slog.Uint64("clinic_id", reservation.ClinicID))
	return reservation, nil
}

// CreateBatch atomically creates one intentionally shared doctor/time booking for selected pets.
// The batch is the sole exception to the normal one-reservation-per-doctor-slot rule.
// It acquires the same clinic advisory lock and rejects any pre-existing, unrelated overlap.
func (s *reservationService) CreateBatch(ctx context.Context, input *CreateManualReservationInput, pets []ReservationBatchPet) ([]model.Reservation, error) {
	if input == nil || len(pets) < 2 {
		return nil, apperrors.WrapInvalidInput("at least two pets are required for a reservation batch")
	}
	if input.ReservationRoute != nil {
		if _, ok := AllowedReservationRoutes[*input.ReservationRoute]; !ok {
			return nil, apperrors.WrapInvalidInput(AllowedReservationRoutesMessage)
		}
	}
	if err := validateTimeRange(input.StartTime, input.EndTime); err != nil {
		return nil, err
	}
	if s.tx == nil {
		return nil, apperrors.WrapInternalServerError("reservation transaction dependency is required")
	}
	if s.typeRepo != nil {
		if _, err := s.typeRepo.FindByID(ctx, input.ClinicID, input.ReservationTypeID); err != nil {
			return nil, apperrors.Wrap(err, "failed to verify reservation type ownership")
		}
	}
	enforceBookingConstraints := ShouldEnforceReservationBookingConstraints(input.Status, input.ReservationRoute)
	if enforceBookingConstraints {
		if err := ValidateReservationTypeAvailableTime(ctx, s.unavailableTimeRepo, input.ClinicID, input.ReservationTypeID, input.StartTime, input.EndTime); err != nil {
			return nil, err
		}
	}
	created := make([]model.Reservation, 0, len(pets))
	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := ValidateReservationStaffCapability(ctx, s.reservationStaffRepo, input.ClinicID, input.DoctorID, input.ReservationTypeID); err != nil {
			return err
		}
		if enforceBookingConstraints {
			if err := s.repo.AcquireBookingLock(ctx, input.ClinicID); err != nil {
				return err
			}
			if err := s.validateCreateClosedDays(ctx, input.ClinicID, input.StartTime); err != nil {
				return err
			}
			if err := s.validateCreateClinicHoliday(ctx, input.ClinicID, input.StartTime); err != nil {
				return err
			}
			if err := CheckSlotConflict(ctx, s.repo, input.ClinicID, input.DoctorID, input.StartTime, input.EndTime, nil); err != nil {
				return err
			}
			if err := CheckReservationTypeCapacityForCount(ctx, s.repo, s.typeRepo, input.ClinicID, input.ReservationTypeID, input.StartTime, len(pets)); err != nil {
				return err
			}
		}
		seen := make(map[uint64]struct{}, len(pets))
		for _, pet := range pets {
			if pet.OwnerID == 0 || pet.PetID == 0 {
				return apperrors.WrapInvalidInput("owner_id and pet_id are required for a reservation batch")
			}
			if _, duplicate := seen[pet.PetID]; duplicate {
				return apperrors.WrapInvalidInput("each pet may appear only once in a reservation batch")
			}
			seen[pet.PetID] = struct{}{}
			ownerID, petID := pet.OwnerID, pet.PetID
			if err := ValidateReservationOwnerPetLinksWithRepo(ctx, s.repo, input.ClinicID, &ownerID, &petID); err != nil {
				return err
			}
			if err := ValidateReservationPetNotDeceased(ctx, s.repo, input.ClinicID, &petID); err != nil {
				return err
			}
			reservation := model.Reservation{ClinicID: input.ClinicID, StartTime: input.StartTime, EndTime: input.EndTime, OwnerID: &ownerID, PetID: &petID, VisitType: input.VisitType, ReservationTypeID: input.ReservationTypeID, DoctorID: input.DoctorID, IsDesignated: input.IsDesignated, Status: input.Status, Notes: input.Notes, Source: input.Source, CreatedBy: input.CreatedBy, ReservationRoute: input.ReservationRoute}
			if err := s.repo.Create(ctx, &reservation); err != nil {
				return err
			}
			created = append(created, reservation)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to create reservation batch")
	}
	return created, nil
}

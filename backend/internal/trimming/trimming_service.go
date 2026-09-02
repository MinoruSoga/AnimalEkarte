// Package service provides business logic implementations for Trimming entity.
package trimming

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// CreateTrimmingInput はトリミング予約作成の入力DTO（appointments ベース, BE-119）
type CreateTrimmingInput struct {
	ActorID           *uint64 // 認証済み staff actor。request body からは受け取らない。
	AppointmentID     *uint64
	ReservationTypeID uint64
	StartTime         time.Time
	EndTime           time.Time
	PetID             *uint64
	StaffID           *uint64 // appointments.doctor_id にマップ
	Status            model.ReservationStatus
	ReservationRoute  *string
	// トリミング詳細（appointment_trimming_details）
	CourseID        *uint64
	StyleRequest    string
	BodyWeight      *float64
	BWUnit          model.BodyWeightUnit
	BodyTemperature *float64
	UsedShampoo     string
	UsedRibbon      string
	Remarks         string
	StyleImage      string
	CompletedImage  string
	OptionIDs       []uint64
}

// UpdateTrimmingInput はトリミング予約部分更新の入力DTO。nil = 未送信フィールド。
// OptionIDs: nil = 変更なし、non-nil（空スライス含む）= 全置換
type UpdateTrimmingInput struct {
	ActorID         *uint64 // 認証済み staff actor。request body からは受け取らない。
	StartTime       *time.Time
	EndTime         *time.Time
	PetID           *uint64
	StaffID         *uint64
	Status          *model.ReservationStatus
	CourseID        *uint64
	StyleRequest    *string
	BodyWeight      *float64
	BWUnit          *model.BodyWeightUnit
	BodyTemperature *float64
	UsedShampoo     *string
	UsedRibbon      *string
	Remarks         *string
	StyleImage      *string
	CompletedImage  *string
	OptionIDs       *[]uint64
}

func buildTrimmingAppointmentUpdateFields(
	input *CreateTrimmingInput,
	locked *model.Reservation,
	status model.ReservationStatus,
	resolvedStart, resolvedEnd time.Time,
) reservation.UpdateTrimmingReservationInput {
	update := reservation.UpdateTrimmingReservationInput{}
	if input.PetID != nil && locked.PetID == nil {
		update.PetID = input.PetID
	}
	if input.StaffID != nil {
		update.DoctorID = input.StaffID
	}
	if !resolvedStart.Equal(locked.StartTime) {
		update.StartTime = &resolvedStart
	}
	if !resolvedEnd.Equal(locked.EndTime) {
		update.EndTime = &resolvedEnd
	}
	if input.Status != "" && status != locked.Status {
		update.Status = &status
	}
	return update
}

func hasTrimmingAppointmentUpdate(input reservation.UpdateTrimmingReservationInput) bool {
	return input.StartTime != nil || input.EndTime != nil || input.PetID != nil || input.DoctorID != nil || input.Status != nil
}

func needsTrimmingAppointmentUpdate(input reservation.UpdateTrimmingReservationInput, locked *model.Reservation) bool {
	return hasTrimmingAppointmentUpdate(input) || (locked.OwnerID == nil && locked.PetID != nil)
}

// TrimmingService はトリミング管理のビジネスロジックインターフェース（BE-119）
type TrimmingService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Reservation, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	Create(ctx context.Context, clinicID uint64, input *CreateTrimmingInput) (*model.Reservation, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingInput) (*model.Reservation, error)
	Delete(ctx context.Context, clinicID, id uint64, actorID *uint64) error
}

type trimmingService struct {
	reservation        TrimmingReservationRepository
	reservationType    ReservationTypeRepository
	reservationStaff   ReservationStaffRepository
	unavailableTime    ReservationTypeUnavailableTimeRepository
	trimmingDetail     AppointmentTrimmingDetailRepository
	trimmingCourseRepo TrimmingCourseRepository
	trimmingOptionRepo TrimmingOptionRepository
	transactor         Transactor
	auditTx            AuditTxLogger
}

// TrimmingReservationRepository is the appointment-owner capability view used by trimming.
// It intentionally exposes only trimming-specific intents plus the reads needed by this domain.
type TrimmingReservationRepository interface {
	sharedkernel.OwnerPetLinkVerifier
	// FindPetByIDInClinic is required for SD-10 deceased write guard (#261 P0).
	FindPetByIDInClinic(ctx context.Context, clinicID, petID uint64) (*model.Pet, error)
	FindAllByCategory(ctx context.Context, clinicID uint64, category model.ReservationTypeCategory, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Reservation, int64, error)
	FindTrimmingByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	AcquireBookingLock(ctx context.Context, clinicID uint64) error
	LockTrimmingByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	HasDoctorConflict(ctx context.Context, clinicID, doctorID uint64, start, end time.Time, excludeID *uint64) (bool, error)
	CountOnDutyDoctors(ctx context.Context, clinicID uint64, date time.Time) (int64, error)
	CountConflicts(ctx context.Context, clinicID uint64, start, end time.Time, excludeID *uint64) (int64, error)
	CountByTypeAndStartTime(ctx context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error)
	CreateForTrimming(ctx context.Context, clinicID uint64, input reservation.CreateTrimmingReservationInput) (*model.Reservation, error)
	UpdateForTrimming(ctx context.Context, clinicID, id uint64, input reservation.UpdateTrimmingReservationInput) (*model.Reservation, error)
	DeleteForTrimming(ctx context.Context, clinicID, id uint64) error
}

func NewTrimmingService(
	reservationRepo TrimmingReservationRepository,
	reservationType ReservationTypeRepository,
	reservationStaff ReservationStaffRepository,
	unavailableTime ReservationTypeUnavailableTimeRepository,
	trimmingDetail AppointmentTrimmingDetailRepository,
	trimmingCourseRepo TrimmingCourseRepository,
	trimmingOptionRepo TrimmingOptionRepository,
	transactor Transactor,
) TrimmingService {
	return NewTrimmingServiceWithAudit(
		reservationRepo,
		reservationType,
		reservationStaff,
		unavailableTime,
		trimmingDetail,
		trimmingCourseRepo,
		trimmingOptionRepo,
		transactor,
		nil,
	)
}

// NewTrimmingServiceWithAudit wires the durable, transaction-local clinical audit sink.
// A nil sink is retained only for composition compatibility and makes every mutation fail closed.
func NewTrimmingServiceWithAudit(
	reservationRepo TrimmingReservationRepository,
	reservationType ReservationTypeRepository,
	reservationStaff ReservationStaffRepository,
	unavailableTime ReservationTypeUnavailableTimeRepository,
	trimmingDetail AppointmentTrimmingDetailRepository,
	trimmingCourseRepo TrimmingCourseRepository,
	trimmingOptionRepo TrimmingOptionRepository,
	transactor Transactor,
	auditTx AuditTxLogger,
) TrimmingService {
	return &trimmingService{
		reservation:        reservationRepo,
		reservationType:    reservationType,
		reservationStaff:   reservationStaff,
		unavailableTime:    unavailableTime,
		trimmingDetail:     trimmingDetail,
		trimmingCourseRepo: trimmingCourseRepo,
		trimmingOptionRepo: trimmingOptionRepo,
		transactor:         transactor,
		auditTx:            auditTx,
	}
}

func (s *trimmingService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Reservation, int64, error) {
	items, total, err := s.reservation.FindAllByCategory(ctx, clinicID, model.ReservationTypeCategoryTrimming, petID, ownerID, startDate, endDate, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list trimming appointments", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list trimming appointments")
	}
	return items, total, nil
}

func (s *trimmingService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	appt, err := s.reservation.FindTrimmingByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get trimming appointment", "error", err)
		return nil, apperrors.Wrap(err, "failed to get trimming appointment")
	}
	// FindTrimmingByID は TrimmingDetail をプリロードしない。
	// 別途 AppointmentTrimmingDetailRepository から取得する。
	// NotFound は正常系（trimming_detail 未作成の予約もありうる）として無視し、
	// それ以外の DB エラー（接続断・タイムアウト等）は伝播させる。
	detail, detailErr := s.trimmingDetail.FindByAppointmentID(ctx, clinicID, id)
	if detailErr == nil {
		appt.TrimmingDetail = detail
	} else if !apperrors.IsNotFound(detailErr) {
		return nil, apperrors.Wrap(detailErr, "failed to get trimming detail")
	}
	return appt, nil
}

func (s *trimmingService) Create(ctx context.Context, clinicID uint64, input *CreateTrimmingInput) (*model.Reservation, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("trimming create input is required")
	}
	if err := requireTrimmingStaffAuditActor(input.ActorID); err != nil {
		return nil, err
	}
	if err := s.requireAuditTx(); err != nil {
		return nil, err
	}
	status := model.ReservationStatusPending
	if input.Status != "" {
		status = input.Status
	}
	bwUnit := model.BodyWeightUnitKg
	if input.BWUnit != "" {
		bwUnit = input.BWUnit
	}

	if input.AppointmentID != nil {
		return s.createDetailForExistingAppointment(ctx, clinicID, *input.AppointmentID, input, bwUnit, status)
	}
	if input.StartTime.IsZero() || input.EndTime.IsZero() {
		return nil, apperrors.WrapInvalidInput("start_time and end_time are required")
	}
	if input.ReservationRoute != nil {
		if _, ok := reservation.AllowedReservationRoutes[*input.ReservationRoute]; !ok {
			return nil, apperrors.WrapInvalidInput(reservation.AllowedReservationRoutesMessage)
		}
	}
	enforceBookingConstraints := reservation.ShouldEnforceReservationBookingConstraints(status, input.ReservationRoute)
	if err := s.validateTrimmingReservationType(ctx, clinicID, input.ReservationTypeID, true); err != nil {
		return nil, err
	}
	if err := reservation.ValidateReservationStaffCapability(ctx, s.reservationStaff, clinicID, input.StaffID, input.ReservationTypeID); err != nil {
		return nil, err
	}
	if enforceBookingConstraints {
		if err := s.requireBookingConstraintDependencies(); err != nil {
			return nil, err
		}
		if err := reservation.ValidateReservationTypeAvailableTime(ctx, s.unavailableTime, clinicID, input.ReservationTypeID, input.StartTime, input.EndTime); err != nil {
			return nil, err
		}
	} else if err := sharedkernel.ValidateTimeRange(input.StartTime, input.EndTime); err != nil {
		return nil, err
	}
	// Fail fast before opening the write transaction, then revalidate under SHARE locks inside the
	// transaction to close the master-update TOCTOU window.
	if err := s.validateTrimmingCourseAndOptions(ctx, clinicID, input.CourseID, input.OptionIDs, nil, nil); err != nil {
		return nil, err
	}
	var apptID uint64
	var result *model.Reservation
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		created, err := s.createTrimmingAppointmentInTx(txCtx, clinicID, input, bwUnit, status, enforceBookingConstraints)
		if err != nil {
			return err
		}
		result = created
		apptID = created.ID
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to create trimming appointment", "error", err)
		return nil, apperrors.Wrap(err, "failed to create trimming appointment")
	}

	slog.InfoContext(ctx, "trimming appointment created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("appointment_id", apptID))

	return result, nil
}

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
	// appointment → trimming_detail → options の3書き込みを単一トランザクションで実行する。
	// 中間でエラーが発生した場合はロールバックされ、孤立レコードは残らない。
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// Booking operations share one clinic-scoped advisory-lock order. Acquire it before
		// patient/staff/master row locks so Create cannot deadlock with Update/Delete paths.
		if enforceBookingConstraints {
			if err := s.reservation.AcquireBookingLock(txCtx, clinicID); err != nil {
				return apperrors.Wrap(err, "failed to acquire trimming booking lock")
			}
		}
		if err := reservation.ValidateReservationOwnerPetLinksWithRepo(txCtx, s.reservation, clinicID, nil, input.PetID); err != nil {
			return err
		}
		if err := reservation.ValidateReservationPetNotDeceased(txCtx, s.reservation, clinicID, input.PetID); err != nil {
			return err
		}
		if err := reservation.ValidateReservationStaffCapability(txCtx, s.reservationStaff, clinicID, input.StaffID, input.ReservationTypeID); err != nil {
			return err
		}
		if enforceBookingConstraints {
			if err := reservation.CheckSlotConflict(txCtx, s.reservation, clinicID, input.StaffID, input.StartTime, input.EndTime, nil); err != nil {
				return err
			}
			if err := reservation.CheckReservationTypeCapacity(txCtx, s.reservation, s.reservationType, clinicID, input.ReservationTypeID, input.StartTime, nil); err != nil {
				return err
			}
		}
		appt, err := s.reservation.CreateForTrimming(txCtx, clinicID, reservation.CreateTrimmingReservationInput{
			ReservationTypeID: input.ReservationTypeID,
			StartTime:         input.StartTime,
			EndTime:           input.EndTime,
			PetID:             input.PetID,
			DoctorID:          input.StaffID,
			Status:            status,
			ReservationRoute:  input.ReservationRoute,
		})
		if err != nil {
			return apperrors.Wrap(err, "failed to create trimming appointment")
		}
		// The course/option reads use this ambient transaction and hold SHARE locks until
		// the detail and junction rows have been persisted.
		if err := s.validateTrimmingCourseAndOptions(txCtx, clinicID, input.CourseID, input.OptionIDs, nil, nil); err != nil {
			return err
		}

		detail := &model.AppointmentTrimmingDetail{
			ClinicID:        clinicID,
			AppointmentID:   appt.ID,
			CourseID:        input.CourseID,
			StyleRequest:    input.StyleRequest,
			BodyWeight:      input.BodyWeight,
			BWUnit:          bwUnit,
			BodyTemperature: input.BodyTemperature,
			UsedShampoo:     input.UsedShampoo,
			UsedRibbon:      input.UsedRibbon,
			Remarks:         input.Remarks,
			StyleImage:      input.StyleImage,
			CompletedImage:  input.CompletedImage,
		}
		if err := s.trimmingDetail.Create(txCtx, detail); err != nil {
			return apperrors.Wrap(err, "failed to create trimming detail")
		}
		if len(input.OptionIDs) > 0 {
			if err := s.trimmingDetail.SetOptions(txCtx, clinicID, appt.ID, input.OptionIDs); err != nil {
				return apperrors.Wrap(err, "failed to set trimming options")
			}
		}

		apptID = appt.ID
		result, err = s.GetByID(txCtx, clinicID, appt.ID)
		if err != nil {
			return err
		}
		return s.logTrimmingAuditTx(
			txCtx,
			clinicID,
			input.ActorID,
			model.AuditActionTrimmingCreate,
			appt.ID,
			trimmingAuditMutationCreateAppointment,
			nil,
			trimmingAuditValue(result, result.TrimmingDetail),
		)
	}); err != nil {
		slog.ErrorContext(ctx, "failed to create trimming appointment", "error", err)
		return nil, apperrors.Wrap(err, "failed to create trimming appointment")
	}

	slog.InfoContext(ctx, "trimming appointment created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("appointment_id", apptID))

	return result, nil
}

func (s *trimmingService) validateTrimmingReservationType(ctx context.Context, clinicID, reservationTypeID uint64, requireActive bool) error {
	if s.reservationType == nil {
		return apperrors.WrapInternalServerError("reservation type repository is required")
	}

	reservationType, err := s.reservationType.FindByID(ctx, clinicID, reservationTypeID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get trimming reservation type")
	}
	if reservationType.Category != model.ReservationTypeCategoryTrimming {
		return apperrors.WrapInvalidInput("reservation_type_id must be a trimming reservation type")
	}
	if requireActive && !reservationType.IsActive {
		return apperrors.WrapInvalidInput("reservation_type_id references an inactive trimming reservation type")
	}
	return nil
}

func (s *trimmingService) requireBookingConstraintDependencies() error {
	if s.unavailableTime == nil {
		return apperrors.WrapInternalServerError("reservation type unavailable-time repository is required")
	}
	if s.reservation == nil {
		return apperrors.WrapInternalServerError("reservation repository is required")
	}
	if s.reservationType == nil {
		return apperrors.WrapInternalServerError("reservation type repository is required")
	}
	return nil
}

// validateTrimmingCourseAndOptions は appointment_trimming_detail の CourseID / OptionIDs が
// caller の clinic に属し、かつ is_active であることを永続化前に検証する（X-14c: 2 repo ガード
// + #228: 無効化されたコース/オプションを新規に紐付けさせない）。request にIDがあるのに
// 対応repositoryが未DIなら、clinic ownershipを検証できないためfail-closedにする。
// existingCourseID/existingOptionIDs はカルテに既に紐づいている ID（Create 系呼び出しでは常に
// nil/空 — 新規紐付けはすべて active 必須）。既にリンク済みの ID は is_active チェックを免除し
// （データを消さない）、新規に追加される ID のみ active を要求する。
func (s *trimmingService) validateTrimmingCourseAndOptions(
	ctx context.Context, clinicID uint64, courseID *uint64, optionIDs []uint64,
	existingCourseID *uint64, existingOptionIDs []uint64,
) error {
	if courseID != nil && s.trimmingCourseRepo == nil {
		return apperrors.WrapInternalServerError("trimming course repository is required")
	}
	if len(optionIDs) > 0 && s.trimmingOptionRepo == nil {
		return apperrors.WrapInternalServerError("trimming option repository is required")
	}
	if s.trimmingCourseRepo != nil {
		if err := validateOwnedMasterFK(ctx, "trimming course", clinicID, courseID,
			func(actx context.Context, cid, mid uint64) error {
				course, err := s.trimmingCourseRepo.FindByID(actx, cid, mid)
				if err != nil {
					return err
				}
				if !course.IsActive && (existingCourseID == nil || *existingCourseID != mid) {
					return apperrors.WrapInvalidInput("course_id references an inactive trimming course")
				}
				return nil
			}); err != nil {
			return err
		}
	}
	if s.trimmingOptionRepo != nil {
		existingOptions := make(map[uint64]bool, len(existingOptionIDs))
		for _, existingID := range existingOptionIDs {
			existingOptions[existingID] = true
		}
		if err := validateOwnedMasterFKs(ctx, "trimming option", clinicID, optionIDs,
			func(actx context.Context, cid, mid uint64) error {
				option, err := s.trimmingOptionRepo.FindByID(actx, cid, mid)
				if err != nil {
					return err
				}
				if !option.IsActive && !existingOptions[mid] {
					return apperrors.WrapInvalidInput("option_ids references an inactive trimming option")
				}
				return nil
			}); err != nil {
			return err
		}
	}
	return nil
}

func (s *trimmingService) createDetailForExistingAppointment(
	ctx context.Context,
	clinicID uint64,
	appointmentID uint64,
	input *CreateTrimmingInput,
	bwUnit model.BodyWeightUnit,
	status model.ReservationStatus,
) (*model.Reservation, error) {
	appt, err := s.reservation.FindTrimmingByID(ctx, clinicID, appointmentID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get existing trimming appointment", "error", err, "appointment_id", appointmentID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get existing trimming appointment")
	}
	if appt.ReservationType == nil || appt.ReservationType.Category != model.ReservationTypeCategoryTrimming {
		return nil, apperrors.WrapInvalidInput("appointment is not a trimming reservation")
	}
	var result *model.Reservation
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.reservation.AcquireBookingLock(txCtx, clinicID); err != nil {
			return apperrors.Wrap(err, "failed to acquire trimming booking lock")
		}
		locked, err := s.reservation.LockTrimmingByID(txCtx, clinicID, appointmentID)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock existing trimming appointment")
		}
		if err := reservation.ValidateTrimmingAppointmentMutable(locked.Status); err != nil {
			return err
		}
		oldValue := trimmingAuditValue(locked, nil)
		// TRM-04: existing detail must not silently return 201 with discarded payload.
		if existing, err := s.trimmingDetail.FindByAppointmentID(txCtx, clinicID, appointmentID); err == nil && existing != nil {
			return apperrors.WrapConflict("trimming detail already exists for this appointment; use PATCH to update")
		} else if err != nil && !apperrors.IsNotFound(err) {
			return apperrors.Wrap(err, "failed to check existing trimming detail")
		}
		if err := s.validateTrimmingCourseAndOptions(txCtx, clinicID, input.CourseID, input.OptionIDs, nil, nil); err != nil {
			return err
		}
		if input.StaffID != nil {
			if err := reservation.ValidateReservationStaffCapability(txCtx, s.reservationStaff, clinicID, input.StaffID, locked.ReservationTypeID); err != nil {
				return err
			}
		}

		finalPetID := locked.PetID
		if input.PetID != nil {
			finalPetID = input.PetID
		}
		if err := reservation.ValidateReservationOwnerPetLinksWithRepo(txCtx, s.reservation, clinicID, locked.OwnerID, finalPetID); err != nil {
			return err
		}
		// Always validate the effective pet (appointment pet when request omits pet_id).
		if err := reservation.ValidateReservationPetNotDeceased(txCtx, s.reservation, clinicID, finalPetID); err != nil {
			return err
		}

		if input.PetID != nil && locked.PetID != nil && *locked.PetID != *input.PetID {
			return apperrors.WrapInvalidInput("pet_id does not match appointment")
		}
		resolvedStart := input.StartTime
		if resolvedStart.IsZero() {
			resolvedStart = locked.StartTime
		}
		resolvedEnd := input.EndTime
		if resolvedEnd.IsZero() {
			resolvedEnd = locked.EndTime
		}
		if !input.StartTime.IsZero() || !input.EndTime.IsZero() {
			if err := s.requireBookingConstraintDependencies(); err != nil {
				return err
			}
			if err := reservation.ValidateReservationTypeAvailableTime(txCtx, s.unavailableTime, clinicID, locked.ReservationTypeID, resolvedStart, resolvedEnd); err != nil {
				return err
			}
		}
		resolvedDoctorID := locked.DoctorID
		if input.StaffID != nil {
			resolvedDoctorID = input.StaffID
		}
		resolvedStatus := locked.Status
		if input.Status != "" {
			resolvedStatus = status
		}
		if (!input.StartTime.IsZero() || !input.EndTime.IsZero() || input.StaffID != nil) &&
			reservation.ShouldEnforceReservationBookingConstraints(resolvedStatus, locked.ReservationRoute) {
			if err := s.requireBookingConstraintDependencies(); err != nil {
				return err
			}
			if err := reservation.CheckSlotConflict(txCtx, s.reservation, clinicID, resolvedDoctorID, resolvedStart, resolvedEnd, &appointmentID); err != nil {
				return err
			}
			if err := reservation.CheckReservationTypeCapacity(txCtx, s.reservation, s.reservationType, clinicID, locked.ReservationTypeID, resolvedStart, &appointmentID); err != nil {
				return err
			}
		}
		appointmentUpdate := buildTrimmingAppointmentUpdateFields(input, locked, status, resolvedStart, resolvedEnd)
		if needsTrimmingAppointmentUpdate(appointmentUpdate, locked) {
			if _, err := s.reservation.UpdateForTrimming(txCtx, clinicID, appointmentID, appointmentUpdate); err != nil {
				return apperrors.Wrap(err, "failed to update existing trimming appointment")
			}
		}

		detail := &model.AppointmentTrimmingDetail{
			ClinicID:        clinicID,
			AppointmentID:   appointmentID,
			CourseID:        input.CourseID,
			StyleRequest:    input.StyleRequest,
			BodyWeight:      input.BodyWeight,
			BWUnit:          bwUnit,
			BodyTemperature: input.BodyTemperature,
			UsedShampoo:     input.UsedShampoo,
			UsedRibbon:      input.UsedRibbon,
			Remarks:         input.Remarks,
			StyleImage:      input.StyleImage,
			CompletedImage:  input.CompletedImage,
		}
		// TRM-06: do not log here — outer WithTx failure boundary records once.
		if err := s.trimmingDetail.Create(txCtx, detail); err != nil {
			return apperrors.Wrap(err, "failed to create trimming detail")
		}
		if len(input.OptionIDs) > 0 {
			if err := s.trimmingDetail.SetOptions(txCtx, clinicID, appointmentID, input.OptionIDs); err != nil {
				return apperrors.Wrap(err, "failed to set trimming options")
			}
		}
		result, err = s.GetByID(txCtx, clinicID, appointmentID)
		if err != nil {
			return err
		}
		return s.logTrimmingAuditTx(
			txCtx,
			clinicID,
			input.ActorID,
			model.AuditActionTrimmingCreate,
			appointmentID,
			trimmingAuditMutationCreateDetail,
			oldValue,
			trimmingAuditValue(result, result.TrimmingDetail),
		)
	}); err != nil {
		slog.ErrorContext(ctx, "failed to create trimming detail for existing appointment", "error", err, "appointment_id", appointmentID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to create trimming detail for existing appointment")
	}

	return result, nil
}

func (s *trimmingService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingInput) (*model.Reservation, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("trimming update input is required")
	}
	if err := requireTrimmingStaffAuditActor(input.ActorID); err != nil {
		return nil, err
	}
	if err := s.requireAuditTx(); err != nil {
		return nil, err
	}
	var optionIDs []uint64
	if input.OptionIDs != nil {
		optionIDs = *input.OptionIDs
	}
	if input.CourseID != nil && s.trimmingCourseRepo == nil {
		return nil, apperrors.WrapInternalServerError("trimming course repository is required")
	}
	if len(optionIDs) > 0 && s.trimmingOptionRepo == nil {
		return nil, apperrors.WrapInternalServerError("trimming option repository is required")
	}
	if _, err := s.reservation.FindTrimmingByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to get trimming appointment", "error", err)
		return nil, apperrors.Wrap(err, "failed to get trimming appointment")
	}

	appointmentUpdate := reservation.UpdateTrimmingReservationInput{
		StartTime: input.StartTime,
		EndTime:   input.EndTime,
		PetID:     input.PetID,
		DoctorID:  input.StaffID,
		Status:    input.Status,
	}

	// appointment → trimming_detail → options の更新を単一トランザクションで実行する。
	// appointments が更新済みで trimming_detail が失敗すると不整合が生じるため、
	// Create と対称なアトミック性を保証する。
	// #228: course_id/option_ids の is_active 検証は、既存紐付け ID（is_active チェック免除対象）
	// を判定するため、tx 内で trimming_detail を取得した直後・フィールド上書き前に行う。
	var result *model.Reservation
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// Reservation updates always take the clinic advisory lock before the appointment row lock.
		// This preserves the global lock order and makes category/status/conflict checks race-safe.
		if err := s.reservation.AcquireBookingLock(txCtx, clinicID); err != nil {
			return apperrors.Wrap(err, "failed to acquire trimming booking lock")
		}
		locked, err := s.reservation.LockTrimmingByID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock trimming appointment")
		}
		if err := reservation.ValidateTrimmingAppointmentMutable(locked.Status); err != nil {
			return err
		}
		finalPetID := locked.PetID
		if input.PetID != nil {
			finalPetID = input.PetID
		}
		if err := reservation.ValidateReservationOwnerPetLinksWithRepo(txCtx, s.reservation, clinicID, locked.OwnerID, finalPetID); err != nil {
			return err
		}
		// Always validate the effective pet (appointment pet when request omits pet_id).
		if err := reservation.ValidateReservationPetNotDeceased(txCtx, s.reservation, clinicID, finalPetID); err != nil {
			return err
		}
		resolvedStart := locked.StartTime
		if input.StartTime != nil {
			resolvedStart = *input.StartTime
		}
		resolvedEnd := locked.EndTime
		if input.EndTime != nil {
			resolvedEnd = *input.EndTime
		}
		resolvedDoctorID := locked.DoctorID
		if input.StaffID != nil {
			resolvedDoctorID = input.StaffID
			if err := reservation.ValidateReservationStaffCapability(txCtx, s.reservationStaff, clinicID, resolvedDoctorID, locked.ReservationTypeID); err != nil {
				return err
			}
		}
		if input.StartTime != nil || input.EndTime != nil {
			if err := s.requireBookingConstraintDependencies(); err != nil {
				return err
			}
			if err := reservation.ValidateReservationTypeAvailableTime(txCtx, s.unavailableTime, clinicID, locked.ReservationTypeID, resolvedStart, resolvedEnd); err != nil {
				return err
			}
		}
		if input.StartTime != nil || input.EndTime != nil || input.StaffID != nil {
			if err := s.requireBookingConstraintDependencies(); err != nil {
				return err
			}
			if err := reservation.CheckSlotConflict(txCtx, s.reservation, clinicID, resolvedDoctorID, resolvedStart, resolvedEnd, &id); err != nil {
				return err
			}
			if err := reservation.CheckReservationTypeCapacity(txCtx, s.reservation, s.reservationType, clinicID, locked.ReservationTypeID, resolvedStart, &id); err != nil {
				return err
			}
		}
		if needsTrimmingAppointmentUpdate(appointmentUpdate, locked) {
			if _, err := s.reservation.UpdateForTrimming(txCtx, clinicID, id, appointmentUpdate); err != nil {
				return apperrors.Wrap(err, "failed to update trimming appointment")
			}
		}

		// trimming detail の更新
		detail, err := s.trimmingDetail.FindByAppointmentID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to get trimming detail for update")
		}
		oldValue := trimmingAuditValue(locked, detail)

		existingOptionIDs := make([]uint64, len(detail.Options))
		for i := range detail.Options {
			existingOptionIDs[i] = detail.Options[i].ID
		}
		if err := s.validateTrimmingCourseAndOptions(txCtx, clinicID, input.CourseID, optionIDs, detail.CourseID, existingOptionIDs); err != nil {
			return err
		}

		if input.CourseID != nil {
			detail.CourseID = input.CourseID
		}
		if input.StyleRequest != nil {
			detail.StyleRequest = *input.StyleRequest
		}
		if input.BodyWeight != nil {
			detail.BodyWeight = input.BodyWeight
		}
		if input.BWUnit != nil {
			detail.BWUnit = *input.BWUnit
		}
		if input.BodyTemperature != nil {
			detail.BodyTemperature = input.BodyTemperature
		}
		if input.UsedShampoo != nil {
			detail.UsedShampoo = *input.UsedShampoo
		}
		if input.UsedRibbon != nil {
			detail.UsedRibbon = *input.UsedRibbon
		}
		if input.Remarks != nil {
			detail.Remarks = *input.Remarks
		}
		if input.StyleImage != nil {
			detail.StyleImage = *input.StyleImage
		}
		if input.CompletedImage != nil {
			detail.CompletedImage = *input.CompletedImage
		}
		if err := s.trimmingDetail.Update(txCtx, detail); err != nil {
			return apperrors.Wrap(err, "failed to update trimming detail")
		}
		if input.OptionIDs != nil {
			if err := s.trimmingDetail.SetOptions(txCtx, clinicID, id, *input.OptionIDs); err != nil {
				return apperrors.Wrap(err, "failed to set trimming options")
			}
		}
		result, err = s.GetByID(txCtx, clinicID, id)
		if err != nil {
			return err
		}
		return s.logTrimmingAuditTx(
			txCtx,
			clinicID,
			input.ActorID,
			model.AuditActionTrimmingUpdate,
			id,
			trimmingAuditMutationUpdate,
			oldValue,
			trimmingAuditValue(result, result.TrimmingDetail),
		)
	}); err != nil {
		slog.ErrorContext(ctx, "failed to update trimming appointment", "error", err)
		return nil, apperrors.Wrap(err, "failed to update trimming appointment")
	}

	slog.InfoContext(ctx, "trimming appointment updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("appointment_id", id))

	return result, nil
}

func (s *trimmingService) Delete(ctx context.Context, clinicID, id uint64, actorID *uint64) error {
	if err := requireTrimmingStaffAuditActor(actorID); err != nil {
		return err
	}
	if err := s.requireAuditTx(); err != nil {
		return err
	}
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.reservation.AcquireBookingLock(txCtx, clinicID); err != nil {
			return apperrors.Wrap(err, "failed to acquire trimming booking lock")
		}
		locked, err := s.reservation.LockTrimmingByID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock trimming appointment for delete")
		}
		var detail *model.AppointmentTrimmingDetail
		detail, err = s.trimmingDetail.FindByAppointmentID(txCtx, clinicID, id)
		if err != nil && !apperrors.IsNotFound(err) {
			return apperrors.Wrap(err, "failed to get trimming detail for delete")
		}
		oldValue := trimmingAuditValue(locked, detail)
		if err := s.reservation.DeleteForTrimming(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to delete trimming appointment")
		}
		return s.logTrimmingAuditTx(
			txCtx,
			clinicID,
			actorID,
			model.AuditActionTrimmingDelete,
			id,
			trimmingAuditMutationDelete,
			oldValue,
			nil,
		)
	}); err != nil {
		slog.ErrorContext(ctx, "failed to delete trimming appointment", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete trimming appointment")
	}
	slog.InfoContext(ctx, "trimming appointment deleted",
		slog.Uint64("appointment_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}

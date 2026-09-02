// Package service provides business logic implementations for Trimming entity.
package trimming

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
)

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
		updated, err := s.updateTrimmingInTx(txCtx, clinicID, id, input, optionIDs, appointmentUpdate)
		if err != nil {
			return err
		}
		result = updated
		return nil
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

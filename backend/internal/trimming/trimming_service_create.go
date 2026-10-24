package trimming

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

func (s *trimmingService) createTrimmingAppointmentInTx(
	txCtx context.Context,
	clinicID uint64,
	input *CreateTrimmingInput,
	bwUnit model.BodyWeightUnit,
	status model.ReservationStatus,
	enforceBookingConstraints bool,
) (*model.Reservation, *uint64, error) {
	if enforceBookingConstraints {
		if err := s.reservation.AcquireBookingLock(txCtx, clinicID); err != nil {
			return nil, nil, apperrors.Wrap(err, "failed to acquire trimming booking lock")
		}
	}
	if err := reservation.ValidateReservationOwnerPetLinksWithRepo(txCtx, s.reservation, clinicID, nil, input.PetID); err != nil {
		return nil, nil, err
	}
	if err := reservation.ValidateReservationPetNotDeceased(txCtx, s.reservation, clinicID, input.PetID); err != nil {
		return nil, nil, err
	}
	if err := reservation.ValidateReservationStaffCapability(txCtx, s.reservationStaff, clinicID, input.StaffID, input.ReservationTypeID); err != nil {
		return nil, nil, err
	}
	if enforceBookingConstraints {
		if err := reservation.CheckSlotConflict(txCtx, s.reservation, clinicID, input.StaffID, input.StartTime, input.EndTime, nil); err != nil {
			return nil, nil, err
		}
		if err := reservation.CheckReservationTypeCapacity(txCtx, s.reservation, s.reservationType, clinicID, input.ReservationTypeID, input.StartTime, nil); err != nil {
			return nil, nil, err
		}
	}
	appt, previousOwnerID, err := s.reservation.CreateForTrimming(txCtx, clinicID, reservation.CreateTrimmingReservationInput{
		ReservationTypeID: input.ReservationTypeID,
		StartTime:         input.StartTime,
		EndTime:           input.EndTime,
		PetID:             input.PetID,
		DoctorID:          input.StaffID,
		Status:            status,
		ReservationRoute:  input.ReservationRoute,
	})
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to create trimming appointment")
	}
	detail, err := newTrimmingDetailFromCreateInput(clinicID, appt.ID, input, bwUnit)
	if err != nil {
		return nil, nil, err
	}
	if err := s.lockImageReferencesInTx(txCtx, clinicID, detail.StyleImage, detail.CompletedImage); err != nil {
		return nil, nil, err
	}
	if err := s.trimmingDetail.Create(txCtx, detail); err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to create trimming detail")
	}
	result, err := s.GetByID(txCtx, clinicID, appt.ID)
	if err != nil {
		return nil, nil, err
	}
	if err := s.logTrimmingAuditTx(
		txCtx,
		clinicID,
		input.ActorID,
		model.AuditActionTrimmingCreate,
		appt.ID,
		trimmingAuditMutationCreateAppointment,
		nil,
		trimmingAuditValue(result, result.TrimmingDetail),
	); err != nil {
		return nil, nil, err
	}
	return result, previousOwnerID, nil
}

func (s *trimmingService) createTrimmingDetailForExistingInTx(
	txCtx context.Context,
	clinicID uint64,
	appointmentID uint64,
	input *CreateTrimmingInput,
	bwUnit model.BodyWeightUnit,
	status model.ReservationStatus,
	locked *model.Reservation,
) (*model.Reservation, error) {
	lockedStatus, err := sharedkernel.RequireStored(locked.Status, "appointments.status")
	if err != nil {
		return nil, err
	}
	if err := reservation.ValidateTrimmingAppointmentMutable(lockedStatus); err != nil {
		return nil, err
	}
	oldValue := trimmingAuditValue(locked, nil)
	if existing, err := s.trimmingDetail.FindByAppointmentID(txCtx, clinicID, appointmentID); err == nil && existing != nil {
		return nil, apperrors.WrapConflict("trimming detail already exists for this appointment; use PATCH to update")
	} else if err != nil && !apperrors.IsNotFound(err) {
		return nil, apperrors.Wrap(err, "failed to check existing trimming detail")
	}
	if input.StaffID != nil {
		if err := reservation.ValidateReservationStaffCapability(txCtx, s.reservationStaff, clinicID, input.StaffID, locked.ReservationTypeID); err != nil {
			return nil, err
		}
	}

	finalPetID := locked.PetID
	if input.PetID != nil {
		finalPetID = input.PetID
	}
	if err := reservation.ValidateReservationOwnerPetLinksWithRepo(txCtx, s.reservation, clinicID, locked.OwnerID, finalPetID); err != nil {
		return nil, err
	}
	if err := reservation.ValidateReservationPetNotDeceased(txCtx, s.reservation, clinicID, finalPetID); err != nil {
		return nil, err
	}

	if input.PetID != nil && locked.PetID != nil && *locked.PetID != *input.PetID {
		return nil, apperrors.WrapInvalidInput("pet_id does not match appointment")
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
			return nil, err
		}
		if err := reservation.ValidateReservationTypeAvailableTime(txCtx, s.unavailableTime, clinicID, locked.ReservationTypeID, resolvedStart, resolvedEnd); err != nil {
			return nil, err
		}
	}
	resolvedDoctorID := locked.DoctorID
	if input.StaffID != nil {
		resolvedDoctorID = input.StaffID
	}
	resolvedStatus := lockedStatus
	if input.Status != "" {
		resolvedStatus = status
	}
	if (!input.StartTime.IsZero() || !input.EndTime.IsZero() || input.StaffID != nil) &&
		reservation.ShouldEnforceReservationBookingConstraints(resolvedStatus, locked.ReservationRoute) {
		if err := s.requireBookingConstraintDependencies(); err != nil {
			return nil, err
		}
		if err := reservation.CheckSlotConflict(txCtx, s.reservation, clinicID, resolvedDoctorID, resolvedStart, resolvedEnd, &appointmentID); err != nil {
			return nil, err
		}
		if err := reservation.CheckReservationTypeCapacity(txCtx, s.reservation, s.reservationType, clinicID, locked.ReservationTypeID, resolvedStart, &appointmentID); err != nil {
			return nil, err
		}
	}
	appointmentUpdate := buildTrimmingAppointmentUpdateFields(input, locked, status, resolvedStart, resolvedEnd)
	if needsTrimmingAppointmentUpdate(appointmentUpdate, locked) {
		if _, err := s.reservation.UpdateForTrimming(txCtx, clinicID, appointmentID, appointmentUpdate); err != nil {
			return nil, apperrors.Wrap(err, "failed to update existing trimming appointment")
		}
	}

	detail, err := newTrimmingDetailFromCreateInput(clinicID, appointmentID, input, bwUnit)
	if err != nil {
		return nil, err
	}
	if err := s.lockImageReferencesInTx(txCtx, clinicID, detail.StyleImage, detail.CompletedImage); err != nil {
		return nil, err
	}
	if err := s.trimmingDetail.Create(txCtx, detail); err != nil {
		return nil, apperrors.Wrap(err, "failed to create trimming detail")
	}
	result, err := s.GetByID(txCtx, clinicID, appointmentID)
	if err != nil {
		return nil, err
	}
	if err := s.logTrimmingAuditTx(
		txCtx,
		clinicID,
		input.ActorID,
		model.AuditActionTrimmingCreate,
		appointmentID,
		trimmingAuditMutationCreateDetail,
		oldValue,
		trimmingAuditValue(result, result.TrimmingDetail),
	); err != nil {
		return nil, err
	}
	return result, nil
}

// trimmingExactNumeric は float64 の request 値（bw / bt）を exact 列へ最短10進表現で橋渡しする。
// nil（未測定）は nil のまま返し、NaN・±Inf は 400 とする。
func trimmingExactNumeric(value *float64, field string) (*model.ExactNumeric, error) {
	if value == nil {
		return nil, nil
	}
	exact, err := model.ExactNumericFromFloat64(*value)
	if err != nil {
		return nil, apperrors.WrapInvalidInput(field + " は有限な数値で指定してください")
	}
	return &exact, nil
}

func newTrimmingDetailFromCreateInput(
	clinicID, appointmentID uint64,
	input *CreateTrimmingInput,
	bwUnit model.BodyWeightUnit,
) (*model.AppointmentTrimmingDetail, error) {
	bodyWeight, err := trimmingExactNumeric(input.BodyWeight, "bw")
	if err != nil {
		return nil, err
	}
	bodyTemperature, err := trimmingExactNumeric(input.BodyTemperature, "bt")
	if err != nil {
		return nil, err
	}
	return &model.AppointmentTrimmingDetail{
		ClinicID:        clinicID,
		AppointmentID:   appointmentID,
		StyleRequest:    input.StyleRequest,
		BodyWeight:      bodyWeight,
		BWUnit:          sharedkernel.Ptr(bwUnit),
		BodyTemperature: bodyTemperature,
		UsedShampoo:     input.UsedShampoo,
		UsedRibbon:      input.UsedRibbon,
		Remarks:         input.Remarks,
		StyleImage:      input.StyleImage,
		CompletedImage:  input.CompletedImage,
	}, nil
}

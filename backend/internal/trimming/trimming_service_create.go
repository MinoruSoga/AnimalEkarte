package trimming

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
)

func (s *trimmingService) createTrimmingAppointmentInTx(
	txCtx context.Context,
	clinicID uint64,
	input *CreateTrimmingInput,
	bwUnit model.BodyWeightUnit,
	status model.ReservationStatus,
	enforceBookingConstraints bool,
) (*model.Reservation, error) {
	if enforceBookingConstraints {
		if err := s.reservation.AcquireBookingLock(txCtx, clinicID); err != nil {
			return nil, apperrors.Wrap(err, "failed to acquire trimming booking lock")
		}
	}
	if err := reservation.ValidateReservationOwnerPetLinksWithRepo(txCtx, s.reservation, clinicID, nil, input.PetID); err != nil {
		return nil, err
	}
	if err := reservation.ValidateReservationPetNotDeceased(txCtx, s.reservation, clinicID, input.PetID); err != nil {
		return nil, err
	}
	if err := reservation.ValidateReservationStaffCapability(txCtx, s.reservationStaff, clinicID, input.StaffID, input.ReservationTypeID); err != nil {
		return nil, err
	}
	if enforceBookingConstraints {
		if err := reservation.CheckSlotConflict(txCtx, s.reservation, clinicID, input.StaffID, input.StartTime, input.EndTime, nil); err != nil {
			return nil, err
		}
		if err := reservation.CheckReservationTypeCapacity(txCtx, s.reservation, s.reservationType, clinicID, input.ReservationTypeID, input.StartTime, nil); err != nil {
			return nil, err
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
		return nil, apperrors.Wrap(err, "failed to create trimming appointment")
	}
	if err := s.validateTrimmingCourseAndOptions(txCtx, clinicID, input.CourseID, input.OptionIDs, nil, nil); err != nil {
		return nil, err
	}
	detail := newTrimmingDetailFromCreateInput(clinicID, appt.ID, input, bwUnit)
	if err := s.trimmingDetail.Create(txCtx, detail); err != nil {
		return nil, apperrors.Wrap(err, "failed to create trimming detail")
	}
	if len(input.OptionIDs) > 0 {
		if err := s.trimmingDetail.SetOptions(txCtx, clinicID, appt.ID, input.OptionIDs); err != nil {
			return nil, apperrors.Wrap(err, "failed to set trimming options")
		}
	}
	result, err := s.GetByID(txCtx, clinicID, appt.ID)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return result, nil
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
	if err := reservation.ValidateTrimmingAppointmentMutable(locked.Status); err != nil {
		return nil, err
	}
	oldValue := trimmingAuditValue(locked, nil)
	if existing, err := s.trimmingDetail.FindByAppointmentID(txCtx, clinicID, appointmentID); err == nil && existing != nil {
		return nil, apperrors.WrapConflict("trimming detail already exists for this appointment; use PATCH to update")
	} else if err != nil && !apperrors.IsNotFound(err) {
		return nil, apperrors.Wrap(err, "failed to check existing trimming detail")
	}
	if err := s.validateTrimmingCourseAndOptions(txCtx, clinicID, input.CourseID, input.OptionIDs, nil, nil); err != nil {
		return nil, err
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
	resolvedStatus := locked.Status
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

	detail := newTrimmingDetailFromCreateInput(clinicID, appointmentID, input, bwUnit)
	if err := s.trimmingDetail.Create(txCtx, detail); err != nil {
		return nil, apperrors.Wrap(err, "failed to create trimming detail")
	}
	if len(input.OptionIDs) > 0 {
		if err := s.trimmingDetail.SetOptions(txCtx, clinicID, appointmentID, input.OptionIDs); err != nil {
			return nil, apperrors.Wrap(err, "failed to set trimming options")
		}
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

func newTrimmingDetailFromCreateInput(
	clinicID, appointmentID uint64,
	input *CreateTrimmingInput,
	bwUnit model.BodyWeightUnit,
) *model.AppointmentTrimmingDetail {
	return &model.AppointmentTrimmingDetail{
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
}

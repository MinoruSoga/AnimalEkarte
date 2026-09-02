package trimming

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
)

func applyTrimmingDetailFields(detail *model.AppointmentTrimmingDetail, input *UpdateTrimmingInput) {
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
}

func (s *trimmingService) updateTrimmingInTx(
	txCtx context.Context,
	clinicID, id uint64,
	input *UpdateTrimmingInput,
	optionIDs []uint64,
	appointmentUpdate reservation.UpdateTrimmingReservationInput,
) (*model.Reservation, error) {
	if err := s.reservation.AcquireBookingLock(txCtx, clinicID); err != nil {
		return nil, apperrors.Wrap(err, "failed to acquire trimming booking lock")
	}
	locked, err := s.reservation.LockTrimmingByID(txCtx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to lock trimming appointment")
	}
	if err := reservation.ValidateTrimmingAppointmentMutable(locked.Status); err != nil {
		return nil, err
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
			return nil, err
		}
	}
	if input.StartTime != nil || input.EndTime != nil {
		if err := s.requireBookingConstraintDependencies(); err != nil {
			return nil, err
		}
		if err := reservation.ValidateReservationTypeAvailableTime(txCtx, s.unavailableTime, clinicID, locked.ReservationTypeID, resolvedStart, resolvedEnd); err != nil {
			return nil, err
		}
	}
	if input.StartTime != nil || input.EndTime != nil || input.StaffID != nil {
		if err := s.requireBookingConstraintDependencies(); err != nil {
			return nil, err
		}
		if err := reservation.CheckSlotConflict(txCtx, s.reservation, clinicID, resolvedDoctorID, resolvedStart, resolvedEnd, &id); err != nil {
			return nil, err
		}
		if err := reservation.CheckReservationTypeCapacity(txCtx, s.reservation, s.reservationType, clinicID, locked.ReservationTypeID, resolvedStart, &id); err != nil {
			return nil, err
		}
	}
	if needsTrimmingAppointmentUpdate(appointmentUpdate, locked) {
		if _, err := s.reservation.UpdateForTrimming(txCtx, clinicID, id, appointmentUpdate); err != nil {
			return nil, apperrors.Wrap(err, "failed to update trimming appointment")
		}
	}

	detail, err := s.trimmingDetail.FindByAppointmentID(txCtx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get trimming detail for update")
	}
	oldValue := trimmingAuditValue(locked, detail)
	existingOptionIDs := make([]uint64, len(detail.Options))
	for i := range detail.Options {
		existingOptionIDs[i] = detail.Options[i].ID
	}
	if err := s.validateTrimmingCourseAndOptions(txCtx, clinicID, input.CourseID, optionIDs, detail.CourseID, existingOptionIDs); err != nil {
		return nil, err
	}
	applyTrimmingDetailFields(detail, input)
	if err := s.trimmingDetail.Update(txCtx, detail); err != nil {
		return nil, apperrors.Wrap(err, "failed to update trimming detail")
	}
	if input.OptionIDs != nil {
		if err := s.trimmingDetail.SetOptions(txCtx, clinicID, id, *input.OptionIDs); err != nil {
			return nil, apperrors.Wrap(err, "failed to set trimming options")
		}
	}
	result, err := s.GetByID(txCtx, clinicID, id)
	if err != nil {
		return nil, err
	}
	if err := s.logTrimmingAuditTx(
		txCtx,
		clinicID,
		input.ActorID,
		model.AuditActionTrimmingUpdate,
		id,
		trimmingAuditMutationUpdate,
		oldValue,
		trimmingAuditValue(result, result.TrimmingDetail),
	); err != nil {
		return nil, err
	}
	return result, nil
}

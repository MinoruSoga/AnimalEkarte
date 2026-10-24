package trimming

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

func applyTrimmingDetailFields(detail *model.AppointmentTrimmingDetail, input *UpdateTrimmingInput) {
	if input.StyleRequest != nil {
		detail.StyleRequest = *input.StyleRequest
	}
	if input.BodyWeight != nil {
		detail.BodyWeight = input.BodyWeight
	}
	if input.BWUnit != nil {
		detail.BWUnit = sharedkernel.Ptr(*input.BWUnit)
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
	appointmentUpdate reservation.UpdateTrimmingReservationInput,
) (*model.Reservation, *TrimmingReplacedImages, *trimmingNextVisitTagChange, error) {
	if err := s.reservation.AcquireBookingLock(txCtx, clinicID); err != nil {
		return nil, nil, nil, apperrors.Wrap(err, "failed to acquire trimming booking lock")
	}
	locked, err := s.reservation.LockTrimmingByID(txCtx, clinicID, id)
	if err != nil {
		return nil, nil, nil, apperrors.Wrap(err, "failed to lock trimming appointment")
	}
	before := trimmingNextVisitFactsFor(locked)
	lockedStatus, err := sharedkernel.RequireStored(locked.Status, "appointments.status")
	if err != nil {
		return nil, nil, nil, err
	}
	if err := reservation.ValidateTrimmingAppointmentMutable(lockedStatus); err != nil {
		return nil, nil, nil, err
	}
	finalPetID := locked.PetID
	if input.PetID != nil {
		finalPetID = input.PetID
	}
	if err := reservation.ValidateReservationOwnerPetLinksWithRepo(txCtx, s.reservation, clinicID, locked.OwnerID, finalPetID); err != nil {
		return nil, nil, nil, err
	}
	if err := reservation.ValidateReservationPetNotDeceased(txCtx, s.reservation, clinicID, finalPetID); err != nil {
		return nil, nil, nil, err
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
			return nil, nil, nil, err
		}
	}
	if input.StartTime != nil || input.EndTime != nil {
		if err := s.requireBookingConstraintDependencies(); err != nil {
			return nil, nil, nil, err
		}
		if err := reservation.ValidateReservationTypeAvailableTime(txCtx, s.unavailableTime, clinicID, locked.ReservationTypeID, resolvedStart, resolvedEnd); err != nil {
			return nil, nil, nil, err
		}
	}
	if input.StartTime != nil || input.EndTime != nil || input.StaffID != nil {
		if err := s.requireBookingConstraintDependencies(); err != nil {
			return nil, nil, nil, err
		}
		if err := reservation.CheckSlotConflict(txCtx, s.reservation, clinicID, resolvedDoctorID, resolvedStart, resolvedEnd, &id); err != nil {
			return nil, nil, nil, err
		}
		if err := reservation.CheckReservationTypeCapacity(txCtx, s.reservation, s.reservationType, clinicID, locked.ReservationTypeID, resolvedStart, &id); err != nil {
			return nil, nil, nil, err
		}
	}
	if needsTrimmingAppointmentUpdate(appointmentUpdate, locked) {
		if _, err := s.reservation.UpdateForTrimming(txCtx, clinicID, id, appointmentUpdate); err != nil {
			return nil, nil, nil, apperrors.Wrap(err, "failed to update trimming appointment")
		}
	}

	detail, err := s.trimmingDetail.FindByAppointmentID(txCtx, clinicID, id)
	if err != nil {
		return nil, nil, nil, apperrors.Wrap(err, "failed to get trimming detail for update")
	}
	// 置き換え前の画像値を lock 保持の tx 内で捕捉する — 同時 PATCH は
	// booking lock でシリアライズされるため、各呼び出しは直前の書き込みを
	// 正確に見る（handler が tx 前に読むスナップショットは並行書き込みの
	// 新値を掃除候補から漏らす）。
	replaced := &TrimmingReplacedImages{Style: detail.StyleImage, Completed: detail.CompletedImage}
	oldValue := trimmingAuditValue(locked, detail)
	applyTrimmingDetailFields(detail, input)
	if err := s.lockImageReferencesInTx(txCtx, clinicID, detail.StyleImage, detail.CompletedImage); err != nil {
		return nil, nil, nil, err
	}
	if err := s.trimmingDetail.Update(txCtx, detail); err != nil {
		return nil, nil, nil, apperrors.Wrap(err, "failed to update trimming detail")
	}
	result, err := s.GetByID(txCtx, clinicID, id)
	if err != nil {
		return nil, nil, nil, err
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
		return nil, nil, nil, err
	}
	return result, replaced, trimmingNextVisitChange(before, result), nil
}

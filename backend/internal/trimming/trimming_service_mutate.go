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
		return nil, apperrors.Wrap(err, "failed to get existing trimming appointment")
	}
	if !appt.ReservationType.HasBusinessKind(model.ReservationBusinessKindTrimming) {
		return nil, apperrors.WrapInvalidInput("appointment is not a trimming reservation")
	}
	var result *model.Reservation
	var tagChange *trimmingNextVisitTagChange
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.reservation.AcquireBookingLock(txCtx, clinicID); err != nil {
			return apperrors.Wrap(err, "failed to acquire trimming booking lock")
		}
		locked, err := s.reservation.LockTrimmingByID(txCtx, clinicID, appointmentID)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock existing trimming appointment")
		}
		before := trimmingNextVisitFactsFor(locked)
		created, err := s.createTrimmingDetailForExistingInTx(txCtx, clinicID, appointmentID, input, bwUnit, status, locked)
		if err != nil {
			return err
		}
		result = created
		tagChange = trimmingNextVisitChange(before, created)
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to create trimming detail for existing appointment")
	}

	s.syncNextVisitTagChange(ctx, clinicID, tagChange)
	return result, nil
}

func (s *trimmingService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingInput) (*model.Reservation, *TrimmingReplacedImages, error) {
	if input == nil {
		return nil, nil, apperrors.WrapInvalidInput("trimming update input is required")
	}
	if err := requireTrimmingStaffAuditActor(input.ActorID); err != nil {
		return nil, nil, err
	}
	if err := s.requireAuditTx(); err != nil {
		return nil, nil, err
	}
	// EMR-480 / EMR-549 C1: コース・オプションの保存先（trimming_booking_revisions 系）の writer が未接続のため、
	// course_id / option_ids を含む更新は予約も詳細も書く前に 422 trimming_booking_pending で拒否する（部分保存しない）。
	// 空配列の option_ids も「全置換」の書込で、保存先がないため受理できない。
	if input.CourseID != nil || input.OptionIDs != nil {
		return nil, nil, apperrors.WrapTrimmingBookingPending()
	}
	if err := normalizeUpdateTrimmingImages(clinicID, input); err != nil {
		return nil, nil, err
	}
	if _, err := s.reservation.FindTrimmingByID(ctx, clinicID, id); err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to get trimming appointment")
	}

	appointmentUpdate := reservation.UpdateTrimmingReservationInput{
		StartTime: input.StartTime,
		EndTime:   input.EndTime,
		PetID:     input.PetID,
		DoctorID:  input.StaffID,
		Status:    input.Status,
	}

	// appointment → trimming_detail の更新を単一トランザクションで実行する。
	// appointments が更新済みで trimming_detail が失敗すると不整合が生じるため、
	// Create と対称なアトミック性を保証する。
	var result *model.Reservation
	var replaced *TrimmingReplacedImages
	var tagChange *trimmingNextVisitTagChange
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		updated, rep, change, err := s.updateTrimmingInTx(txCtx, clinicID, id, input, appointmentUpdate)
		if err != nil {
			return err
		}
		result = updated
		replaced = rep
		tagChange = change
		return nil
	}); err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to update trimming appointment")
	}
	s.syncNextVisitTagChange(ctx, clinicID, tagChange)

	slog.InfoContext(ctx, "trimming appointment updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("appointment_id", id))

	return result, replaced, nil
}

func (s *trimmingService) Delete(ctx context.Context, clinicID, id uint64, actorID *uint64) error {
	_, err := s.DeleteWithImages(ctx, clinicID, id, actorID)
	return err
}

func (s *trimmingService) DeleteWithImages(ctx context.Context, clinicID, id uint64, actorID *uint64) (*TrimmingDeletedImages, error) {
	if err := requireTrimmingStaffAuditActor(actorID); err != nil {
		return nil, err
	}
	if err := s.requireAuditTx(); err != nil {
		return nil, err
	}
	var ownerID uint64
	var images *TrimmingDeletedImages
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.reservation.AcquireBookingLock(txCtx, clinicID); err != nil {
			return apperrors.Wrap(err, "failed to acquire trimming booking lock")
		}
		locked, err := s.reservation.LockTrimmingByID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock trimming appointment for delete")
		}
		ownerID = trimmingNextVisitFactsFor(locked).ownerID
		var detail *model.AppointmentTrimmingDetail
		detail, err = s.trimmingDetail.FindByAppointmentID(txCtx, clinicID, id)
		if err != nil && !apperrors.IsNotFound(err) {
			return apperrors.Wrap(err, "failed to get trimming detail for delete")
		}
		if detail != nil {
			images = &TrimmingDeletedImages{Style: detail.StyleImage, Completed: detail.CompletedImage}
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
		return nil, apperrors.Wrap(err, "failed to delete trimming appointment")
	}
	s.syncNextVisitTagsBestEffort(ctx, clinicID, &ownerID)
	slog.InfoContext(ctx, "trimming appointment deleted",
		slog.Uint64("appointment_id", id),
		slog.Uint64("clinic_id", clinicID))
	return images, nil
}

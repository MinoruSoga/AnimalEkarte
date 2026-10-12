package reservation

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// guardedCheckInWriter is only needed for the owner-detail reception shortcut.
// Existing generic updates and re-reception keep their established behavior.
type guardedCheckInWriter interface {
	checkInConfirmed(ctx context.Context, clinicID, id uint64, expectedUpdatedAt, checkedInAt time.Time) (*model.Reservation, error)
}

func validateGuardedCheckInInput(input *UpdateReservationInput) error {
	if input.ExpectedUpdatedAt == nil {
		return nil
	}
	if input.ExpectedUpdatedAt.IsZero() || input.Status == nil || *input.Status != model.ReservationStatusCheckedIn ||
		input.StartTime != nil || input.EndTime != nil || input.OwnerID != nil || input.PetID != nil ||
		input.VisitType != nil || input.ReservationTypeID != nil || input.DoctorID != nil || input.IsDesignated != nil || input.Notes != nil {
		return apperrors.WrapInvalidInput("expected_updated_at は status=checked_in のみの受付更新に指定してください")
	}
	return nil
}

func (s *reservationService) guardedCheckIn(ctx context.Context, clinicID, id uint64, expected time.Time) (*model.Reservation, error) {
	writer, ok := s.repo.(guardedCheckInWriter)
	if !ok || s.tx == nil {
		return nil, apperrors.WrapInternalServerError("guarded check-in transaction dependency is required")
	}
	var updated *model.Reservation
	if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		current, err := s.repo.LockAndFindByID(txCtx, clinicID, id)
		if err != nil {
			return err
		}
		if current == nil || current.Status != model.ReservationStatusConfirmed || !current.UpdatedAt.Equal(expected) {
			return apperrors.WrapConflict("予約が更新されています。再読み込みしてください")
		}
		now := time.Now()
		if current.StartTime.In(config.JST).Format(time.DateOnly) != now.In(config.JST).Format(time.DateOnly) {
			return apperrors.WrapConflict("当日の予約ではありません。再読み込みしてください")
		}
		if err := s.validateGuardedCheckInPatient(txCtx, clinicID, current); err != nil {
			return err
		}
		// SHARE-lock acquisition may wait across JST midnight.
		now = time.Now()
		if current.StartTime.In(config.JST).Format(time.DateOnly) != now.In(config.JST).Format(time.DateOnly) {
			return apperrors.WrapConflict("当日の予約ではありません。再読み込みしてください")
		}
		updated, err = writer.checkInConfirmed(txCtx, clinicID, id, expected, now)
		if err == nil && updated == nil {
			return apperrors.WrapInternalServerError("guarded check-in returned no reservation")
		}
		return err
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to check in confirmed reservation")
	}
	syncNextVisitTagsBestEffort(ctx, s.nextVisitTagSync, clinicID, updated.OwnerID)
	return updated, nil
}

// Both reads retain SHARE locks in the same ambient transaction. Check the
// owner and alive state from that one locked pet, without changing the generic
// historical-reservation deceased guard's established contract.
func (s *reservationService) validateGuardedCheckInPatient(ctx context.Context, clinicID uint64, current *model.Reservation) error {
	return validateGuardedReceptionPatient(ctx, s.repo, clinicID, current)
}

type guardedReceptionPatientReader interface {
	AssertOwnerInClinic(context.Context, uint64, uint64) error
	FindPetByIDInClinic(context.Context, uint64, uint64) (*model.Pet, error)
}

func validateGuardedReceptionPatient(ctx context.Context, repo guardedReceptionPatientReader, clinicID uint64, current *model.Reservation) error {
	if current.OwnerID == nil || *current.OwnerID == 0 || current.PetID == nil || *current.PetID == 0 {
		return apperrors.WrapInvalidInput("受付する前に飼主とペットの紐付けが必要です")
	}
	if err := repo.AssertOwnerInClinic(ctx, clinicID, *current.OwnerID); err != nil {
		return apperrors.Wrap(err, "failed to verify check-in owner")
	}
	pet, err := repo.FindPetByIDInClinic(ctx, clinicID, *current.PetID)
	if err != nil {
		return apperrors.Wrap(err, "failed to verify check-in pet")
	}
	if pet == nil || pet.ClinicID != clinicID || pet.ID != *current.PetID || pet.OwnerID != *current.OwnerID {
		return apperrors.WrapConflict("飼主とペットの情報が更新されています。再読み込みしてください")
	}
	if pet.Status != model.PetStatusAlive || pet.DeceasedAt != nil {
		return apperrors.WrapInvalidInput("生存が確認できないペットは受付できません")
	}
	return nil
}

// The CAS also guards the write itself even after the locked validation read.
// The same ambient transaction retains the reservation lock and pet SHARE lock.
func (r *reservationRepository) checkInConfirmed(ctx context.Context, clinicID, id uint64, expectedUpdatedAt, checkedInAt time.Time) (*model.Reservation, error) {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return nil, apperrors.WrapInternalServerError("guarded check-in requires a transaction")
	}
	result := tx.WithContext(ctx).Model(&model.Reservation{}).
		Where("clinic_id = ? AND id = ? AND deleted_at IS NULL AND status = ? AND updated_at = ?", clinicID, id, model.ReservationStatusConfirmed, expectedUpdatedAt).
		Updates(buildGuardedCheckInUpdateFields(checkedInAt))
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "reservation", "")
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapConflict("予約が更新されています。再読み込みしてください")
	}
	return r.FindByID(ctx, clinicID, id)
}

func buildGuardedCheckInUpdateFields(checkedInAt time.Time) map[string]any {
	return map[string]any{
		"status":        model.ReservationStatusCheckedIn,
		"checked_in_at": checkedInAt,
		"updated_at":    checkedInAt,
	}
}

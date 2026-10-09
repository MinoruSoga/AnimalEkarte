package reservation

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type guardedReceptionCreator interface {
	createGuardedReception(context.Context, *model.Reservation) (*model.Reservation, error)
}

func createOrPromoteTentativeWithGuard(ctx context.Context, repo ReservationRepository, booking *model.Reservation, preventDuplicate bool) (*model.Reservation, error) {
	if !preventDuplicate {
		return createOrPromoteTentative(ctx, repo, booking)
	}
	writer, ok := repo.(guardedReceptionCreator)
	if !ok {
		return nil, apperrors.WrapInternalServerError("guarded reception repository is required")
	}
	created, err := writer.createGuardedReception(ctx, booking)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to create guarded reception")
	}
	if created == nil {
		return nil, apperrors.WrapInternalServerError("guarded reception returned no reservation")
	}
	return created, nil
}

func validateReceptionCreationGuard(input *CreateManualReservationInput) error {
	if !input.PreventDuplicateReception {
		return nil
	}
	if input.ReservationRoute == nil || *input.ReservationRoute != "reception" || input.Status != model.ReservationStatusCheckedIn || input.StartTime.In(config.JST).Format(time.DateOnly) != time.Now().In(config.JST).Format(time.DateOnly) {
		return apperrors.WrapInvalidInput("prevent_duplicate_reception は当日の reception/checked_in 作成のみ指定できます")
	}
	return nil
}

// Serialize the empty-candidate check and insertion using the existing patient
// intent lock. Finished visits remain eligible for a legitimate return visit.
func (r *reservationRepository) createGuardedReception(ctx context.Context, booking *model.Reservation) (*model.Reservation, error) {
	if err := requireNextVisitIntentAmbientTransaction(ctx); err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, apperrors.WrapInvalidInput("reservation must not be nil")
	}
	if err := validateReceptionCreationGuard(&CreateManualReservationInput{PreventDuplicateReception: true, ReservationRoute: booking.ReservationRoute, Status: booking.Status, StartTime: booking.StartTime}); err != nil {
		return nil, err
	}
	if booking.OwnerID == nil || *booking.OwnerID == 0 || booking.PetID == nil || *booking.PetID == 0 {
		return nil, apperrors.WrapInvalidInput("owner_id and pet_id are required for guarded reception")
	}
	if err := r.acquireNextVisitIntentLock(ctx, booking.ClinicID, *booking.PetID); err != nil {
		return nil, err
	}
	if err := validateGuardedReceptionPatient(ctx, r, booking.ClinicID, booking); err != nil {
		return nil, err
	}
	now := time.Now().In(config.JST)
	if booking.StartTime.In(config.JST).Format(time.DateOnly) != now.Format(time.DateOnly) {
		return nil, apperrors.WrapConflict("当日ではありません。再読み込みしてください")
	}
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, config.JST)
	var count int64
	err := persistence.TxFromContext(ctx).WithContext(ctx).Model(&model.Reservation{}).
		Where("clinic_id = ? AND pet_id = ? AND deleted_at IS NULL AND start_time >= ? AND start_time < ? AND status IN ?", booking.ClinicID, *booking.PetID, day, day.AddDate(0, 0, 1), []model.ReservationStatus{model.ReservationStatusPending, model.ReservationStatusConfirmed, model.ReservationStatusCheckedIn, model.ReservationStatusInConsultation, model.ReservationStatusAccounting}).Count(&count).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation", "")
	}
	if count > 0 {
		return nil, apperrors.WrapConflict("当日の予約または受付が追加されています。再読み込みしてください")
	}
	copyBooking := *booking
	return r.promoteOrCreateFromTentative(ctx, &copyBooking, true)
}

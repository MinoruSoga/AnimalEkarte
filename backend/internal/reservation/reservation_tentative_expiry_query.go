package reservation

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// FindOwnerIDsWithExpiredTentativeTagCursor returns current owners with a past
// tentative date and its stale next_visit tag. All past dates are considered so
// a paused daily job catches up. Successful cache removal removes the candidate;
// failed removal leaves it available on the next run. Appointment state is unchanged.
func (r *reservationRepository) FindOwnerIDsWithExpiredTentativeTagCursor(
	ctx context.Context, clinicID uint64, asOf time.Time, afterOwnerID uint64, limit int,
) ([]uint64, error) {
	ownerIDs := make([]uint64, 0)
	err := r.expiredTentativeTagOwnersQuery(ctx, clinicID, asOf, afterOwnerID, limit).Find(&ownerIDs).Error
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find expired tentative tag owners")
	}
	return ownerIDs, nil
}

func (r *reservationRepository) expiredTentativeTagOwnersQuery(
	ctx context.Context, clinicID uint64, asOf time.Time, afterOwnerID uint64, limit int,
) *gorm.DB {
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	_, dateOnly := jstCalendarDate(asOf)
	return persistence.DBOrTx(ctx, r.db).Model(&model.Reservation{}).
		Select("DISTINCT "+nextVisitOwnerExpression+" AS owner_id").
		Where("appointments.clinic_id = ? AND appointments.deleted_at IS NULL AND appointments.status = ?", clinicID, model.ReservationStatusTentative).
		Where("appointments.intended_date < ?::date", dateOnly).
		Where("("+nextVisitOwnerExpression+") > ?", afterOwnerID).
		// Compare a generated date string, never cast untrusted cache text to DATE.
		// Both the cache and pet owner lookup are scoped to the appointment clinic.
		Where(`EXISTS (SELECT 1 FROM lstep_tag_cache expiry_tag
			WHERE expiry_tag.clinic_id = appointments.clinic_id
			  AND expiry_tag.owner_id = (` + nextVisitOwnerExpression + `)
			  AND expiry_tag.tag_name = 'next_visit_' || to_char(appointments.intended_date, 'YYYY-MM-DD'))`).
		Order("owner_id ASC").Limit(limit)
}

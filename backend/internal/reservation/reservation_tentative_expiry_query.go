package reservation

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// FindOwnerIDsWithExpiredTentativeTagCursor returns owners directly from past
// next_visit tag cache entries, including former owners and deleted pets. All past dates are considered so
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

	// Strict calendar validation stays inside CASE, so malformed provider/cache
	// strings never reach integer casts. Lexical ISO comparison avoids DATE casts.
	const validPastTag = `CASE WHEN expiry_tag.tag_name ~ '^next_visit_[0-9]{4}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$'
  AND substring(expiry_tag.tag_name, 12, 4) <> '0000'
 THEN substring(expiry_tag.tag_name, 20, 2)::integer <= CASE substring(expiry_tag.tag_name, 17, 2)
  WHEN '02' THEN CASE WHEN substring(expiry_tag.tag_name, 12, 4)::integer % 400 = 0
   OR (substring(expiry_tag.tag_name, 12, 4)::integer % 4 = 0 AND substring(expiry_tag.tag_name, 12, 4)::integer % 100 <> 0) THEN 29 ELSE 28 END
  WHEN '04' THEN 30 WHEN '06' THEN 30 WHEN '09' THEN 30 WHEN '11' THEN 30 ELSE 31 END
  AND substring(expiry_tag.tag_name, 12, 10) < ?
 ELSE FALSE END`
	return persistence.DBOrTx(ctx, r.db).Table("lstep_tag_cache AS expiry_tag").
		Select("DISTINCT expiry_tag.owner_id AS owner_id").
		Joins("JOIN owners owner ON owner.id = expiry_tag.owner_id AND owner.clinic_id = expiry_tag.clinic_id AND owner.deleted_at IS NULL").
		Where("expiry_tag.clinic_id = ? AND expiry_tag.owner_id > ?", clinicID, afterOwnerID).
		Where(validPastTag, dateOnly).
		Order("owner_id ASC").Limit(limit)
}

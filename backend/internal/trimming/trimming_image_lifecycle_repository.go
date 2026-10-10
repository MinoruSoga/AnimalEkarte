package trimming

import (
	"context"
	"fmt"
	"sort"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func (r *appointmentTrimmingDetailRepository) LockImageKeys(ctx context.Context, keys []ImageKey) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("trimming image lock requires an ambient transaction")
	}
	ordered := append([]ImageKey(nil), keys...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].clinicID != ordered[j].clinicID {
			return ordered[i].clinicID < ordered[j].clinicID
		}
		return ordered[i].value < ordered[j].value
	})
	for _, key := range ordered {
		if _, err := ownedImageKey(key.clinicID, key.value); err != nil {
			return err
		}
		// Namespace is distinct from booking/intent locks. Hash collisions only
		// serialize unrelated images; they cannot weaken the protocol.
		name := fmt.Sprintf("trimming-image:%d:%s", key.clinicID, key.value)
		if err := persistence.DBOrTx(ctx, r.db).Exec(
			"SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", name,
		).Error; err != nil {
			return apperrors.Wrap(err, "failed to acquire trimming image key lock")
		}
	}
	return nil
}

func (r *appointmentTrimmingDetailRepository) ImageKeyDeleted(ctx context.Context, key ImageKey) (bool, error) {
	if persistence.TxFromContext(ctx) == nil {
		return false, apperrors.WrapInternalServerError("trimming image lifecycle check requires an ambient transaction")
	}
	var exists bool
	err := persistence.DBOrTx(ctx, r.db).Raw(
		"SELECT EXISTS (SELECT 1 FROM trimming_image_deletions WHERE clinic_id = ? AND object_key = ?)",
		key.clinicID, key.value,
	).Scan(&exists).Error
	return exists, err
}

func (r *appointmentTrimmingDetailRepository) MarkImageKeyDeleted(ctx context.Context, key ImageKey, actorID uint64) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("trimming image deletion requires an ambient transaction")
	}
	return persistence.DBOrTx(ctx, r.db).Exec(`
		INSERT INTO trimming_image_deletions (clinic_id, object_key, actor_staff_id)
		VALUES (?, ?, ?)
		ON CONFLICT (clinic_id, object_key) DO NOTHING`, key.clinicID, key.value, actorID).Error
}

var _ ImageLifecycleRepository = (*appointmentTrimmingDetailRepository)(nil)

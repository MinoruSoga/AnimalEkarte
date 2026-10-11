package lstep

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// nextVisitSyncRetryQueue persists work before any provider write and acknowledges
// it atomically with the cache. A newer enqueue must survive an older sync.
type nextVisitSyncRetryQueue interface {
	EnqueueNextVisitSync(context.Context, uint64, uint64) (uint64, error)
	AcknowledgeNextVisitSync(context.Context, uint64, uint64, uint64) error
}

func (r *lstepTagCacheRepository) EnqueueNextVisitSync(ctx context.Context, clinicID, ownerID uint64) (uint64, error) {
	if r.db == nil || clinicID == 0 || ownerID == 0 {
		return 0, apperrors.WrapInternalServerError("next visit retry requires database and scoped owner")
	}
	// This single statement commits before WithNextVisitOwnerLock starts. Do not
	// enqueue inside that transaction or borrow another connection from its callback.
	var revision uint64
	result := r.db.WithContext(persistence.WithTxValue(ctx, nil)).Raw(`
INSERT INTO lstep_next_visit_sync_retries (clinic_id, owner_id)
SELECT clinic_id, id FROM owners WHERE clinic_id = ? AND id = ? AND deleted_at IS NULL
ON CONFLICT (clinic_id, owner_id) DO UPDATE
SET revision = EXCLUDED.revision, updated_at = NOW()
RETURNING revision`, clinicID, ownerID).Scan(&revision)
	if result.Error != nil {
		return 0, apperrors.Wrap(result.Error, "failed to persist next visit sync retry")
	}
	if revision == 0 {
		return 0, apperrors.WrapNotFound("owner", fmt.Sprint(ownerID))
	}
	return revision, nil
}

func (r *lstepTagCacheRepository) AcknowledgeNextVisitSync(ctx context.Context, clinicID, ownerID, revision uint64) error {
	if r.db == nil || clinicID == 0 || ownerID == 0 || revision == 0 || persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("next visit retry acknowledgement requires scoped transaction and revision")
	}
	// Zero affected rows is valid: another invocation may already have enqueued
	// a newer revision. Commit failure rolls this deletion back with the cache.
	result := persistence.DBOrTx(ctx, r.db).Exec(`DELETE FROM lstep_next_visit_sync_retries
WHERE clinic_id = ? AND owner_id = ? AND revision = ?`, clinicID, ownerID, revision)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "failed to acknowledge next visit sync retry")
	}
	return nil
}

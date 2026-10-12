package lstep

import (
	"context"
	"fmt"
	"regexp"

	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
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
WITH target_owner AS MATERIALIZED (
  SELECT clinic_id, id FROM owners WHERE clinic_id = ? AND id = ? AND deleted_at IS NULL FOR KEY SHARE
)
INSERT INTO lstep_next_visit_sync_retries (clinic_id, owner_id)
SELECT clinic_id, id FROM target_owner
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

var cleanupLineUserIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

type NextVisitCleanupIntent struct {
	ClinicID   uint64 `gorm:"column:clinic_id"`
	OwnerID    uint64 `gorm:"column:owner_id"`
	Revision   uint64 `gorm:"column:revision"`
	LineUserID string `gorm:"column:cleanup_line_user_id"`
}

type nextVisitCleanupQueue interface {
	EnqueueNextVisitCleanup(context.Context, uint64, uint64, string) error
}

func (r *lstepTagCacheRepository) EnqueueNextVisitCleanup(ctx context.Context, clinicID, ownerID uint64, lineUserID string) error {
	if r.db == nil || persistence.TxFromContext(ctx) == nil || clinicID == 0 || ownerID == 0 {
		return apperrors.WrapInternalServerError("next visit cleanup enqueue requires a scoped ambient transaction")
	}
	if !cleanupLineUserIDPattern.MatchString(lineUserID) {
		return apperrors.WrapInvalidInput("invalid cleanup line user id")
	}
	var locked model.Owner
	if err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "clinic_id", "line_user_id").
		Where("clinic_id = ? AND id = ? AND deleted_at IS NULL", clinicID, ownerID).
		First(&locked).Error; err != nil {
		return apperrors.FromGORM(err, "owner", fmt.Sprint(ownerID))
	}
	if locked.ClinicID != clinicID || locked.LineUserID == nil || *locked.LineUserID != lineUserID {
		return apperrors.WrapConflict("next visit cleanup target does not match the locked owner")
	}
	if err := persistence.AcquireLineRecipientLock(ctx, clinicID, lineUserID); err != nil {
		return err
	}
	result := persistence.DBOrTx(ctx, r.db).Exec(`
INSERT INTO lstep_next_visit_sync_retries (clinic_id, owner_id, cleanup_line_user_id)
VALUES (?, ?, ?)
ON CONFLICT (clinic_id, owner_id) DO UPDATE
SET revision = EXCLUDED.revision, cleanup_line_user_id = EXCLUDED.cleanup_line_user_id, updated_at = NOW()`, clinicID, ownerID, lineUserID)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "failed to persist next visit cleanup intent")
	}
	return nil
}

func (r *lstepTagCacheRepository) FindNextVisitCleanupIntent(ctx context.Context, clinicID, ownerID uint64) (*NextVisitCleanupIntent, error) {
	if r.db == nil || clinicID == 0 || ownerID == 0 {
		return nil, apperrors.WrapInternalServerError("next visit cleanup intent requires database and scoped owner")
	}
	var intent NextVisitCleanupIntent
	result := persistence.DBOrTx(ctx, r.db).Raw(`SELECT clinic_id, owner_id, revision, cleanup_line_user_id
FROM lstep_next_visit_sync_retries
WHERE clinic_id = ? AND owner_id = ? AND cleanup_line_user_id IS NOT NULL`, clinicID, ownerID).Scan(&intent)
	if result.Error != nil {
		return nil, apperrors.Wrap(result.Error, "failed to find next visit cleanup intent")
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("next visit cleanup intent", fmt.Sprint(ownerID))
	}
	return &intent, nil
}

func (r *lstepTagCacheRepository) FindOwnerForNextVisitCleanup(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error) {
	if r.db == nil || clinicID == 0 || ownerID == 0 {
		return nil, apperrors.WrapInternalServerError("next visit cleanup owner read requires database and scoped owner")
	}
	var owner model.Owner
	err := persistence.DBOrTx(ctx, r.db).
		Unscoped().
		Select("id", "clinic_id", "line_user_id", "lstep_opt_out", "deleted_at").
		Where("clinic_id = ? AND id = ?", clinicID, ownerID).
		First(&owner).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "owner", fmt.Sprint(ownerID))
	}
	return &owner, nil
}

func (r *lstepTagCacheRepository) CountActiveOwnersWithLineUser(ctx context.Context, clinicID uint64, lineUserID string) (int64, error) {
	if r.db == nil || clinicID == 0 {
		return 0, apperrors.WrapInternalServerError("line recipient ownership check requires database and scoped clinic")
	}
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Owner{}).
		Where("clinic_id = ? AND line_user_id = ? AND deleted_at IS NULL", clinicID, lineUserID).
		Count(&count).Error; err != nil {
		return 0, apperrors.Wrap(err, "failed to count line recipient owners")
	}
	return count, nil
}

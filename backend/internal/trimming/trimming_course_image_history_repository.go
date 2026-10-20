package trimming

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// Saved originals remain referenced by immutable course image audit history
// after replace/clear/soft-delete. A missing/unreadable audit sink fails closed.
func trimmingCourseImageHistoryInUse(ctx context.Context, db *gorm.DB, clinicID uint64, key string) (bool, error) {
	var references []struct{ OldKey, NewKey *string }
	if err := persistence.DBOrTx(ctx, db).Raw(`
		SELECT old_value->>'image_object_key' AS old_key,
		       new_value->>'image_object_key' AS new_key
		FROM audit_logs
		WHERE clinic_id = ? AND resource = ?
		  AND (old_value->>'image_object_key' LIKE ? OR new_value->>'image_object_key' LIKE ?)`,
		clinicID, courseImageAuditResource, "%"+trimmingImageStorageKeyPrefix+"%", "%"+trimmingImageStorageKeyPrefix+"%",
	).Scan(&references).Error; err != nil {
		return false, apperrors.Wrap(err, "failed to check original course image history")
	}
	for _, reference := range references {
		for _, stored := range []*string{reference.OldKey, reference.NewKey} {
			if stored != nil && trimmingImageStoredReferenceMatchesKey(*stored, key) {
				return true, nil
			}
		}
	}
	return false, nil
}

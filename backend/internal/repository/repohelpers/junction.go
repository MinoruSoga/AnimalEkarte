package repohelpers

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// junctionInsertBatchSize is the batch size for junction/child row inserts.
// Mirrors the parent repository package's private constant of the same value
// (kept independent to avoid an import cycle back into repository).
const junctionInsertBatchSize = 100

// ReplaceChildRowsByParentID deletes then re-inserts child rows scoped by a parent
// FK column, inside a single transaction. No clinic validation — for direct
// parent-FK child replacement (e.g. shift entry/template breaks).
// parentFKColumn must be a compile-time literal only (never runtime SQL assembly).
func ReplaceChildRowsByParentID[T any](
	db *gorm.DB,
	parentID uint64,
	rows []T,
	childModel any,
	parentFKColumn string,
	resource string,
	assignParentID func(*T, uint64),
	wrapMessage string,
) error {
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(fmt.Sprintf("%s = ?", parentFKColumn), parentID).
			Delete(childModel).Error; err != nil {
			return apperrors.FromGORM(err, resource, fmt.Sprintf("%d", parentID))
		}
		if len(rows) == 0 {
			return nil
		}
		for i := range rows {
			assignParentID(&rows[i], parentID)
		}
		if err := tx.CreateInBatches(rows, junctionInsertBatchSize).Error; err != nil {
			return apperrors.FromGORM(err, resource, "")
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, wrapMessage)
	}
	return nil
}

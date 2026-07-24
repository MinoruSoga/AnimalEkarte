package repohelpers

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/persistence"
)

// ReplaceChildRowsByParentID delegates to persistence.
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
	return persistence.ReplaceChildRowsByParentID(
		db,
		parentID,
		rows,
		childModel,
		parentFKColumn,
		resource,
		assignParentID,
		wrapMessage,
	)
}

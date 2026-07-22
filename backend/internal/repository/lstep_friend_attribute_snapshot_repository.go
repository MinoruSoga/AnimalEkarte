package repository

// BE9-2C L⑤ compatibility facade. The implementation and owned contract live
// in internal/lstep; repository aggregation remains until BE9-2F.

import (
	"gorm.io/gorm"

	lstepdomain "github.com/animal-ekarte/backend/internal/lstep"
)

type LstepFriendAttributeSnapshotRepository = lstepdomain.LstepFriendAttributeSnapshotRepository

func NewLstepFriendAttributeSnapshotRepository(db *gorm.DB) LstepFriendAttributeSnapshotRepository {
	return lstepdomain.NewLstepFriendAttributeSnapshotRepository(db)
}

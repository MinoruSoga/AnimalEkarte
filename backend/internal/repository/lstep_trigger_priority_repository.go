package repository

// BE9-2C L④ compatibility facade. Repository aggregation remains until BE9-2F,
// while the implementation and contract are owned by internal/lstep.

import (
	"gorm.io/gorm"

	lstepdomain "github.com/animal-ekarte/backend/internal/lstep"
)

type LstepTriggerPriorityRepository = lstepdomain.LstepTriggerPriorityRepository

func NewLstepTriggerPriorityRepository(db *gorm.DB) LstepTriggerPriorityRepository {
	return lstepdomain.NewLstepTriggerPriorityRepository(db)
}

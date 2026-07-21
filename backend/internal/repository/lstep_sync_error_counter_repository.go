package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/lstep"
)

// LstepSyncErrorCounterRepository は internal/lstep への移行facade（BE9-2C L①・BE9-2F削除予定）。
type LstepSyncErrorCounterRepository = lstep.LstepSyncErrorCounterRepository

// NewLstepSyncErrorCounterRepository は internal/lstep の実装を返す（BE9-2C L① facade）。
func NewLstepSyncErrorCounterRepository(db *gorm.DB) LstepSyncErrorCounterRepository {
	return lstep.NewLstepSyncErrorCounterRepository(db)
}

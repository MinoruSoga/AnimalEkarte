package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/lstep"
)

// LstepTagConfigRepository は internal/lstep への移行facade（BE9-2C L②・BE9-2F削除予定）。
type LstepTagConfigRepository = lstep.LstepTagConfigRepository

// NewLstepTagConfigRepository は internal/lstep の実装を返す（BE9-2C L② facade）。
func NewLstepTagConfigRepository(db *gorm.DB) LstepTagConfigRepository {
	return lstep.NewLstepTagConfigRepository(db)
}

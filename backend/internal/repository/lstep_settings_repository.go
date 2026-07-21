package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/lstep"
)

// LstepSettingsRepository は internal/lstep への移行facade（BE9-2C L①・BE9-2F削除予定）。
type LstepSettingsRepository = lstep.LstepSettingsRepository

// NewLstepSettingsRepository は internal/lstep の実装を返す（BE9-2C L① facade）。
func NewLstepSettingsRepository(db *gorm.DB) LstepSettingsRepository {
	return lstep.NewLstepSettingsRepository(db)
}

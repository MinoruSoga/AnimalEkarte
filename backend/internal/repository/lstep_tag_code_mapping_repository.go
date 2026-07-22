package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/lstep"
)

// LstepTagCodeMappingRepository は internal/lstep への移行facade（BE9-2C L③a・BE9-2F削除予定）。
type LstepTagCodeMappingRepository = lstep.LstepTagCodeMappingRepository

// NewLstepTagCodeMappingRepository は internal/lstep の実装を返す（BE9-2C L③a facade）。
func NewLstepTagCodeMappingRepository(db *gorm.DB) LstepTagCodeMappingRepository {
	return lstep.NewLstepTagCodeMappingRepository(db)
}

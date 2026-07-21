package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/billing"
)

// InsuranceRepository は internal/billing への移行facade（BE9-2C B①・BE9-2F削除予定）。
type InsuranceRepository = billing.InsuranceRepository

// NewInsuranceRepository は internal/billing の実装を返す（BE9-2C B① facade）。
func NewInsuranceRepository(db *gorm.DB) InsuranceRepository {
	return billing.NewInsuranceRepository(db)
}

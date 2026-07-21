package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/billing"
)

// EstimateRepository は internal/billing への移行facade（BE9-2C B②・BE9-2F削除予定）。
type EstimateRepository = billing.EstimateRepository

// NewEstimateRepository は internal/billing の実装を返す（BE9-2C B② facade）。
func NewEstimateRepository(db *gorm.DB) EstimateRepository {
	return billing.NewEstimateRepository(db)
}

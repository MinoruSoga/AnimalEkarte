package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/clinic"
)

// ClosingSpecialPeriodRepository is a stable facade alias for closingspecialperiod.
type ClosingSpecialPeriodRepository = clinic.ClosingSpecialPeriodRepository

// NewClosingSpecialPeriodRepository constructs the closing special period repository.
func NewClosingSpecialPeriodRepository(db *gorm.DB) ClosingSpecialPeriodRepository {
	return clinic.NewClosingSpecialPeriodRepository(db)
}

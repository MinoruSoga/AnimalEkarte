package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/closingspecialperiod"
)

// ClosingSpecialPeriodRepository is a stable facade alias for closingspecialperiod.
type ClosingSpecialPeriodRepository = closingspecialperiod.Repository

// NewClosingSpecialPeriodRepository constructs the closing special period repository.
func NewClosingSpecialPeriodRepository(db *gorm.DB) ClosingSpecialPeriodRepository {
	return closingspecialperiod.New(db)
}

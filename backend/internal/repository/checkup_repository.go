package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/checkup"
)

// CheckupFilters is a stable facade alias for the checkup domain package.
type CheckupFilters = checkup.Filters

// CheckupRepository is a stable facade alias for the checkup domain package.
// Service/handler imports keep using repository.* so the domain split does
// not churn all importers.
type CheckupRepository = checkup.Repository

// NewCheckupRepository constructs the checkup repository.
func NewCheckupRepository(db *gorm.DB) CheckupRepository {
	return checkup.New(db)
}

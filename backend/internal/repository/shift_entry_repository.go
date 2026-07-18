package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/shiftentry"
)

// ShiftEntryFilter is a stable facade alias for the shiftentry domain package
// (BE8-4). Service/handler imports keep using repository.* so the split does
// not churn all importers.
type ShiftEntryFilter = shiftentry.Filter

// ShiftEntryRepository is a stable facade alias for the shiftentry domain
// package (BE8-4). Service/handler imports keep using repository.* so the
// split does not churn all importers.
type ShiftEntryRepository = shiftentry.Repository

// NewShiftEntryRepository constructs the shift entry repository.
func NewShiftEntryRepository(db *gorm.DB) ShiftEntryRepository {
	return shiftentry.New(db)
}

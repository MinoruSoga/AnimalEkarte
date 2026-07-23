package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/clinic"
)

// ClinicHolidayRepository is a stable facade alias for the clinicholiday domain
// package (BE8-4). Service/handler imports keep using repository.* so the split
// does not churn all importers.
type ClinicHolidayRepository = clinic.ClinicHolidayRepository

// NewClinicHolidayRepository constructs the clinic holiday repository.
func NewClinicHolidayRepository(db *gorm.DB) ClinicHolidayRepository {
	return clinic.NewClinicHolidayRepository(db)
}

package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/clinic"
)

// ClinicSettingsRepository is a stable facade alias for the clinicsettings domain
// package (BE8-4). Service/handler imports keep using repository.* so the split
// does not churn all importers.
type ClinicSettingsRepository = clinic.ClinicSettingsRepository

// NewClinicSettingsRepository constructs the clinic settings repository.
func NewClinicSettingsRepository(db *gorm.DB) ClinicSettingsRepository {
	return clinic.NewClinicSettingsRepository(db)
}

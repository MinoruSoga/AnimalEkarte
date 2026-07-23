// Package clinicsettings keeps the legacy import path for clinic settings data access.
package clinicsettings

import (
	"gorm.io/gorm"

	clinicrepo "github.com/animal-ekarte/backend/internal/clinic/clinicsettings"
)

type Repository = clinicrepo.Repository

func New(db *gorm.DB) Repository {
	return clinicrepo.New(db)
}

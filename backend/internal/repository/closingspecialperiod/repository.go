// Package closingspecialperiod keeps the legacy import path for special-period data access.
package closingspecialperiod

import (
	"gorm.io/gorm"

	clinicrepo "github.com/animal-ekarte/backend/internal/clinic/closingspecialperiod"
)

type Repository = clinicrepo.Repository

func New(db *gorm.DB) Repository {
	return clinicrepo.New(db)
}

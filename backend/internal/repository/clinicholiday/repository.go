// Package clinicholiday keeps the legacy import path for clinic holiday data access.
package clinicholiday

import (
	"gorm.io/gorm"

	clinicrepo "github.com/animal-ekarte/backend/internal/clinic/clinicholiday"
)

type Repository = clinicrepo.Repository

func New(db *gorm.DB) Repository {
	return clinicrepo.New(db)
}

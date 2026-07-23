// Package company keeps the legacy import path for company data access.
package company

import (
	"gorm.io/gorm"

	clinicrepo "github.com/animal-ekarte/backend/internal/clinic/company"
)

type Repository = clinicrepo.Repository

func New(db *gorm.DB) Repository {
	return clinicrepo.New(db)
}

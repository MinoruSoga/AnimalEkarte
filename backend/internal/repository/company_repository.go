package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/clinic"
)

// CompanyRepository is a stable facade alias for the company domain
// package (BE8-4). Service/handler imports keep using repository.* so the split
// does not churn all importers.
type CompanyRepository = clinic.CompanyRepository

// NewCompanyRepository constructs the company repository.
func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return clinic.NewCompanyRepository(db)
}

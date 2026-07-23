// Package service keeps the legacy company service surface during BE9 migration.
// The implementation lives in internal/clinic.
package service

import (
	"github.com/animal-ekarte/backend/internal/clinic"
	"github.com/animal-ekarte/backend/internal/repository"
)

type UpdateCompanyInput = clinic.UpdateCompanyInput
type CompanyService = clinic.CompanyService

func NewCompanyService(repo repository.CompanyRepository) CompanyService {
	return clinic.NewCompanyService(repo)
}

func buildCompanyUpdate(input *UpdateCompanyInput) map[string]any {
	return clinic.BuildCompanyUpdate(input)
}

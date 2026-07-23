// Package repository keeps the legacy clinic repository surface during BE9 migration.
// The implementation lives in internal/clinic.
package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/clinic"
)

type ClinicRepository = clinic.ClinicRepository
type ClinicDependencyCount = clinic.ClinicDependencyCount

func NewClinicRepository(db *gorm.DB) ClinicRepository {
	return clinic.NewClinicRepository(db)
}

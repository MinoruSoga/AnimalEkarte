package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/vaccine"
)

// VaccineRepository is a stable facade alias for vaccine.
type VaccineRepository = vaccine.Repository

// NewVaccineRepository constructs the vaccine repository.
func NewVaccineRepository(db *gorm.DB) VaccineRepository {
	return vaccine.New(db)
}

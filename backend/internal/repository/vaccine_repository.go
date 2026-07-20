package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// VaccineRepository is a stable facade alias for medicalrecord.VaccineRepository (BE9-2D
// roll-up of the former vaccine subpackage). Kept only because the repositories.go aggregator still constructs it and
// cmd/api/main.go passes that instance into the medicalrecord constructors; no other production
// consumer remains. Delete when main.go switches to calling medicalrecord.New* directly (BE9-2F).
type VaccineRepository = medicalrecord.VaccineRepository

// NewVaccineRepository constructs the vaccine repository.
func NewVaccineRepository(db *gorm.DB) VaccineRepository {
	return medicalrecord.NewVaccineRepository(db)
}

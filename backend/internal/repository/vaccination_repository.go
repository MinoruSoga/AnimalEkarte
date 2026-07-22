package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// VaccinationRepository is a stable facade alias for medicalrecord.VaccinationRepository
// (BE9-2D roll-up: the implementation moved from internal/repository/vaccination_repository.go
// into internal/medicalrecord). Kept because the lstep/liff domains and the
// repositories.go aggregator still depend on the stable constructor.
type VaccinationRepository = medicalrecord.VaccinationRepository

// NewVaccinationRepository constructs the vaccination repository.
func NewVaccinationRepository(db *gorm.DB) VaccinationRepository {
	return medicalrecord.NewVaccinationRepository(db)
}

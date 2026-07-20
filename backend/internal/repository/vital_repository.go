package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// VitalRepository is a stable facade alias for medicalrecord.VitalRepository (BE9-2D sub-batch④a
// roll-up). Kept because treatment_dose_save.go (treatment domain, staying in internal/service
// until its own BE9-2D batch) reads repos.Vital and the repositories.go aggregator still depends on it.
type VitalRepository = medicalrecord.VitalRepository

// NewVitalRepository constructs the vital repository.
func NewVitalRepository(db *gorm.DB) VitalRepository {
	return medicalrecord.NewVitalRepository(db)
}

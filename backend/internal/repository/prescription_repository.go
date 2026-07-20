package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// PrescriptionRepository is a stable facade alias for medicalrecord.PrescriptionRepository
// (BE9-2D roll-up of the former prescription subpackage). Kept because lstep_tag_sync_service.go (lstep
// domain, staying in internal/service until its own BE9-2C batch) and the repositories.go
// aggregator still depend on it.
type PrescriptionRepository = medicalrecord.PrescriptionRepository

// NewPrescriptionRepository constructs the prescription repository.
func NewPrescriptionRepository(db *gorm.DB) PrescriptionRepository {
	return medicalrecord.NewPrescriptionRepository(db)
}

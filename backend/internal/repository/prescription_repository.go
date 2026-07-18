package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/prescription"
)

// PrescriptionRepository is a stable facade alias for the prescription domain
// package (BE8-4). Service/handler imports keep using repository.* so the split
// does not churn all importers.
type PrescriptionRepository = prescription.Repository

// NewPrescriptionRepository constructs the prescription repository.
func NewPrescriptionRepository(db *gorm.DB) PrescriptionRepository {
	return prescription.New(db)
}

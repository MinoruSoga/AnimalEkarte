package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// ClinicalPlanRepository is a stable facade alias for medicalrecord.ClinicalPlanRepository
// (BE9-2D sub-batch④a roll-up). Kept because medical_record_service.go (staying in internal/service
// until its own BE9-2D batch) depends on repos.ClinicalPlan and the repositories.go aggregator still
// depends on it.
type ClinicalPlanRepository = medicalrecord.ClinicalPlanRepository

// NewClinicalPlanRepository constructs the clinical plan repository.
func NewClinicalPlanRepository(db *gorm.DB) ClinicalPlanRepository {
	return medicalrecord.NewClinicalPlanRepository(db)
}

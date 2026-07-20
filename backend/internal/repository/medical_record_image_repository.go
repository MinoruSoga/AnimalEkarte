package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// MedicalRecordImageRepository is a stable facade alias for medicalrecord.MedicalRecordImageRepository
// (BE9-2D sub-batch④a roll-up). No residual internal/service consumer references the type directly,
// but the repositories.go aggregator still injects repos.MedicalRecordImage, so the facade is kept
// (same aggregator-injection rationale as the sub-batch③ leaves that retained a facade).
type MedicalRecordImageRepository = medicalrecord.MedicalRecordImageRepository

// NewMedicalRecordImageRepository constructs the medical record image repository.
func NewMedicalRecordImageRepository(db *gorm.DB) MedicalRecordImageRepository {
	return medicalrecord.NewMedicalRecordImageRepository(db)
}

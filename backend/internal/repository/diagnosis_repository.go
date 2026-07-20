package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// DiagnosisTypeRepository is a facade alias to medicalrecord.DiagnosisTypeRepository
// (BE9-2C roll-up of the former diagnosistype subpackage — BE8-4 batch27). Kept here (not
// deleted) because clinical_plan_service.go/medical_record_service.go, still in
// internal/service pending BE9-2D, depend on this facade type.
type DiagnosisTypeRepository = medicalrecord.DiagnosisTypeRepository

// NewDiagnosisTypeRepository constructs a DiagnosisTypeRepository.
func NewDiagnosisTypeRepository(db *gorm.DB) DiagnosisTypeRepository {
	return medicalrecord.NewDiagnosisTypeRepository(db)
}

// DiagnosisNameRepository is a facade alias to medicalrecord.DiagnosisNameRepository
// (BE9-2C roll-up of the former diagnosisname subpackage — BE8-4 batch28). Kept here (not
// deleted) because clinical_plan_service.go/medical_record_service.go, still in
// internal/service pending BE9-2D, depend on this facade type.
type DiagnosisNameRepository = medicalrecord.DiagnosisNameRepository

// NewDiagnosisNameRepository constructs a DiagnosisNameRepository.
func NewDiagnosisNameRepository(db *gorm.DB) DiagnosisNameRepository {
	return medicalrecord.NewDiagnosisNameRepository(db)
}

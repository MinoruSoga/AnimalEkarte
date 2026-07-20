package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// ExamTypeRepository is a stable facade alias for medicalrecord.ExamTypeRepository (BE9-2C
// roll-up of the former examtype subpackage). Kept here (not deleted) because
// examination_service.go, still in internal/service, depends on this facade type.
// (lab_import_examination_service.go moved to internal/medicalrecord in BE9-2D sub-batch③ and now
// references the unqualified medicalrecord.ExamTypeRepository directly, no longer via this facade.)
type ExamTypeRepository = medicalrecord.ExamTypeRepository

// NewExamTypeRepository constructs the examination type master repository.
func NewExamTypeRepository(db *gorm.DB) ExamTypeRepository {
	return medicalrecord.NewExamTypeRepository(db)
}

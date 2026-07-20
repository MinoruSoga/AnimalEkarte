package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// ExamTypeRepository is a stable facade alias for medicalrecord.ExamTypeRepository (BE9-2C
// roll-up of the former examtype subpackage). Kept here (not deleted) because
// examination_service.go/lab_import_examination_service.go, still in internal/service
// pending BE9-2D, depend on this facade type.
type ExamTypeRepository = medicalrecord.ExamTypeRepository

// NewExamTypeRepository constructs the examination type master repository.
func NewExamTypeRepository(db *gorm.DB) ExamTypeRepository {
	return medicalrecord.NewExamTypeRepository(db)
}

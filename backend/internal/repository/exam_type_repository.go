package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/examtype"
)

// ExamTypeRepository is a stable facade alias for examtype.
type ExamTypeRepository = examtype.Repository

// NewExamTypeRepository constructs the examination type master repository.
func NewExamTypeRepository(db *gorm.DB) ExamTypeRepository {
	return examtype.New(db)
}

package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/diagnosisname"
	"github.com/animal-ekarte/backend/internal/repository/diagnosistype"
)

// DiagnosisTypeRepository is a facade alias to diagnosistype.Repository (BE8-4 batch27).
type DiagnosisTypeRepository = diagnosistype.Repository

// NewDiagnosisTypeRepository constructs a DiagnosisTypeRepository.
func NewDiagnosisTypeRepository(db *gorm.DB) DiagnosisTypeRepository {
	return diagnosistype.New(db)
}

// DiagnosisNameRepository is a facade alias to diagnosisname.Repository (BE8-4 batch28).
type DiagnosisNameRepository = diagnosisname.Repository

// NewDiagnosisNameRepository constructs a DiagnosisNameRepository.
func NewDiagnosisNameRepository(db *gorm.DB) DiagnosisNameRepository {
	return diagnosisname.New(db)
}

package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// CheckupTypeRepository is a stable facade alias for medicalrecord.CheckupTypeRepository
// (BE9-2D roll-up of the former checkuptype subpackage). Kept only because the repositories.go aggregator still constructs it and
// cmd/api/main.go passes that instance into the medicalrecord constructors; no other production
// consumer remains. Delete when main.go switches to calling medicalrecord.New* directly (BE9-2F).
type CheckupTypeRepository = medicalrecord.CheckupTypeRepository

// NewCheckupTypeRepository constructs the checkup type repository.
func NewCheckupTypeRepository(db *gorm.DB) CheckupTypeRepository {
	return medicalrecord.NewCheckupTypeRepository(db)
}

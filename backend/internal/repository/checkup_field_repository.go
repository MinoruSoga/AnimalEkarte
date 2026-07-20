package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// CheckupTypeFieldRepository is a stable facade alias for medicalrecord.CheckupTypeFieldRepository
// (BE9-2D roll-up: the implementation moved from internal/repository/checkup_field_repository.go
// into internal/medicalrecord). Kept only because the repositories.go aggregator still constructs it and
// cmd/api/main.go passes that instance into the medicalrecord constructors; no other production
// consumer remains. Delete when main.go switches to calling medicalrecord.New* directly (BE9-2F).
type CheckupTypeFieldRepository = medicalrecord.CheckupTypeFieldRepository

// NewCheckupTypeFieldRepository constructs the checkup type field repository.
func NewCheckupTypeFieldRepository(db *gorm.DB) CheckupTypeFieldRepository {
	return medicalrecord.NewCheckupTypeFieldRepository(db)
}

// CheckupFieldResultRepository is a stable facade alias for
// medicalrecord.CheckupFieldResultRepository (BE9-2D roll-up: the implementation moved from
// internal/repository/checkup_field_repository.go into internal/medicalrecord). Kept only because the
// repositories.go aggregator still constructs it and cmd/api/main.go passes that instance into
// the medicalrecord constructors; no other production consumer remains. Delete when main.go
// switches to calling medicalrecord.New* directly (BE9-2F).
type CheckupFieldResultRepository = medicalrecord.CheckupFieldResultRepository

// NewCheckupFieldResultRepository constructs the checkup field result repository.
func NewCheckupFieldResultRepository(db *gorm.DB) CheckupFieldResultRepository {
	return medicalrecord.NewCheckupFieldResultRepository(db)
}

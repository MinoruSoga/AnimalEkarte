package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// CheckupFilters is a stable facade alias for medicalrecord.CheckupFilters (BE9-2D roll-up of
// the former checkup subpackage). Kept because lstep_tag_sync_service.go (lstep domain, staying in
// internal/service until its own BE9-2C batch) and the repositories.go aggregator still depend on it.
type CheckupFilters = medicalrecord.CheckupFilters

// CheckupRepository is a stable facade alias for medicalrecord.CheckupRepository (BE9-2D
// roll-up of the former checkup subpackage). Kept because lstep_tag_sync_service.go (lstep domain,
// staying in internal/service until its own BE9-2C batch) and the repositories.go aggregator
// still depend on it.
type CheckupRepository = medicalrecord.CheckupRepository

// NewCheckupRepository constructs the checkup repository.
func NewCheckupRepository(db *gorm.DB) CheckupRepository {
	return medicalrecord.NewCheckupRepository(db)
}

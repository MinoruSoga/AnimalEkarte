package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// ChiefComplaintTypeRepository is a stable facade alias for
// medicalrecord.ChiefComplaintTypeRepository (BE9-2C roll-up of the former chiefcomplaint
// subpackage). Kept here (not deleted) because medical_record_service.go and NewServices (both in
// internal/service) still depend on this facade type; inquiry_service.go moved to
// internal/medicalrecord in BE9-2D and now uses medicalrecord.ChiefComplaintTypeRepository directly.
type ChiefComplaintTypeRepository = medicalrecord.ChiefComplaintTypeRepository

// NewChiefComplaintTypeRepository constructs the chief complaint type repository.
func NewChiefComplaintTypeRepository(db *gorm.DB) ChiefComplaintTypeRepository {
	return medicalrecord.NewChiefComplaintTypeRepository(db)
}

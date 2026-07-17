package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/chiefcomplaint"
)

// ChiefComplaintTypeRepository is a stable facade alias for chiefcomplaint.
type ChiefComplaintTypeRepository = chiefcomplaint.Repository

// NewChiefComplaintTypeRepository constructs the chief complaint type repository.
func NewChiefComplaintTypeRepository(db *gorm.DB) ChiefComplaintTypeRepository {
	return chiefcomplaint.New(db)
}

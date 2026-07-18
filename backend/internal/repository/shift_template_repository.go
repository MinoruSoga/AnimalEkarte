package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/shifttemplate"
)

// ShiftTemplateRepository is a stable facade alias for the shifttemplate domain
// package (BE8-4). Service/handler imports keep using repository.* so the split
// does not churn all importers.
type ShiftTemplateRepository = shifttemplate.Repository

// NewShiftTemplateRepository constructs the shift template repository.
func NewShiftTemplateRepository(db *gorm.DB) ShiftTemplateRepository {
	return shifttemplate.New(db)
}

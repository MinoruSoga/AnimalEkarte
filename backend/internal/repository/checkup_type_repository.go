package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/checkuptype"
)

// CheckupTypeRepository is a stable facade alias for the checkuptype domain
// package (BE8-4). Service/handler imports keep using repository.* so the split
// does not churn all importers.
type CheckupTypeRepository = checkuptype.Repository

// NewCheckupTypeRepository constructs the checkup type repository.
func NewCheckupTypeRepository(db *gorm.DB) CheckupTypeRepository {
	return checkuptype.New(db)
}

package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/cage"
)

// CageRepository is a stable facade alias for the cage domain package.
type CageRepository = cage.Repository

// NewCageRepository constructs the cage repository.
func NewCageRepository(db *gorm.DB) CageRepository {
	return cage.New(db)
}

package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/occupation"
)

// OccupationRepository is a stable facade alias for occupation.
type OccupationRepository = occupation.Repository

// NewOccupationRepository constructs the occupation repository.
func NewOccupationRepository(db *gorm.DB) OccupationRepository {
	return occupation.New(db)
}

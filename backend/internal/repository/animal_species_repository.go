package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/animalspecies"
)

// AnimalSpeciesRepository is a stable facade alias for animalspecies.
type AnimalSpeciesRepository = animalspecies.Repository

// NewAnimalSpeciesRepository constructs the animal species repository.
func NewAnimalSpeciesRepository(db *gorm.DB) AnimalSpeciesRepository {
	return animalspecies.New(db)
}

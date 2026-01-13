package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

// PetRepository defines the interface for pet data access operations.
type PetRepository interface {
	GetAllPets(ctx context.Context) ([]model.Pet, error)
	GetPetByID(ctx context.Context, id uuid.UUID) (*model.Pet, error)
	CreatePet(ctx context.Context, pet *model.Pet) error
	UpdatePet(ctx context.Context, pet *model.Pet) error
	DeletePet(ctx context.Context, id uuid.UUID) error
}

// Ensure Repository implements PetRepository
var _ PetRepository = (*Repository)(nil)

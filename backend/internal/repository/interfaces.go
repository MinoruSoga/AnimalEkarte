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

// OwnerRepository defines the interface for owner data access operations.
type OwnerRepository interface {
	GetAllOwners(ctx context.Context) ([]model.Owner, error)
	GetOwnerByID(ctx context.Context, id uuid.UUID) (*model.Owner, error)
	CreateOwner(ctx context.Context, owner *model.Owner) error
	UpdateOwner(ctx context.Context, owner *model.Owner) error
	DeleteOwner(ctx context.Context, id uuid.UUID) error
}

// Ensure Repository implements interfaces
var _ PetRepository = (*Repository)(nil)
var _ OwnerRepository = (*Repository)(nil)

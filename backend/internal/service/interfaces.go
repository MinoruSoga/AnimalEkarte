package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

// PetService defines the interface for pet business logic.
type PetService interface {
	GetAllPets(ctx context.Context) ([]model.Pet, error)
	GetPetByID(ctx context.Context, id string) (*model.Pet, error)
	CreatePet(ctx context.Context, req *model.CreatePetRequest) (*model.Pet, error)
	UpdatePet(ctx context.Context, id string, req *model.UpdatePetRequest) (*model.Pet, error)
	DeletePet(ctx context.Context, id string) error
}

// OwnerService defines the interface for owner business logic.
type OwnerService interface {
	GetAllOwners(ctx context.Context) ([]model.Owner, error)
	GetOwnerByID(ctx context.Context, id string) (*model.Owner, error)
	CreateOwner(ctx context.Context, req *model.CreateOwnerRequest) (*model.Owner, error)
	UpdateOwner(ctx context.Context, id string, req *model.UpdateOwnerRequest) (*model.Owner, error)
	DeleteOwner(ctx context.Context, id string) error
}

// Ensure Service implements interfaces
var _ PetService = (*Service)(nil)
var _ OwnerService = (*Service)(nil)

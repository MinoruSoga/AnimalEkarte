package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

// PetService defines the interface for pet business logic operations.
type PetService interface {
	GetAllPets(ctx context.Context) ([]model.Pet, error)
	GetPetByID(ctx context.Context, id string) (*model.Pet, error)
	CreatePet(ctx context.Context, req *model.CreatePetRequest) (*model.Pet, error)
	UpdatePet(ctx context.Context, id string, req *model.UpdatePetRequest) (*model.Pet, error)
	DeletePet(ctx context.Context, id string) error
}

// Ensure Service implements PetService
var _ PetService = (*Service)(nil)

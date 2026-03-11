package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type PetService interface {
	List(ctx context.Context, ownerID *uuid.UUID, page, limit int, search string) ([]model.Pet, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Pet, error)
	Create(ctx context.Context, pet *model.Pet) error
	Update(ctx context.Context, pet *model.Pet) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type petService struct {
	repo repository.PetRepository
}

func NewPetService(repo repository.PetRepository) PetService {
	return &petService{repo: repo}
}

func (s *petService) List(ctx context.Context, ownerID *uuid.UUID, page, limit int, search string) ([]model.Pet, int64, error) {
	return s.repo.FindAll(ctx, ownerID, page, limit, search)
}

func (s *petService) GetByID(ctx context.Context, id uuid.UUID) (*model.Pet, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *petService) Create(ctx context.Context, pet *model.Pet) error {
	return s.repo.Create(ctx, pet)
}

func (s *petService) Update(ctx context.Context, pet *model.Pet) error {
	return s.repo.Update(ctx, pet)
}

func (s *petService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

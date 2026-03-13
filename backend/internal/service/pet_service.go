package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type PetService interface {
	List(ctx context.Context, clinicID uint64, ownerID *uint64, page, limit int, search string) ([]model.Pet, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
	Create(ctx context.Context, pet *model.Pet) error
	Update(ctx context.Context, pet *model.Pet) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type petService struct {
	repo repository.PetRepository
}

func NewPetService(repo repository.PetRepository) PetService {
	return &petService{repo: repo}
}

func (s *petService) List(ctx context.Context, clinicID uint64, ownerID *uint64, page, limit int, search string) ([]model.Pet, int64, error) {
	return s.repo.FindAll(ctx, clinicID, ownerID, page, limit, search)
}

func (s *petService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *petService) Create(ctx context.Context, pet *model.Pet) error {
	return s.repo.Create(ctx, pet)
}

func (s *petService) Update(ctx context.Context, pet *model.Pet) error {
	return s.repo.Update(ctx, pet)
}

func (s *petService) Delete(ctx context.Context, clinicID, id uint64) error {
	return s.repo.Delete(ctx, clinicID, id)
}

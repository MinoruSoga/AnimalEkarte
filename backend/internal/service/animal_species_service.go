// Package service provides business logic implementations for AnimalSpecies entity.
package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// AnimalSpeciesService はペット種類マスタのビジネスロジック層
type AnimalSpeciesService interface {
	List(ctx context.Context) ([]model.AnimalSpecies, error)
}

type animalSpeciesService struct{ repo repository.AnimalSpeciesRepository }

// NewAnimalSpeciesService はAnimalSpeciesServiceを初期化して返す
func NewAnimalSpeciesService(repo repository.AnimalSpeciesRepository) AnimalSpeciesService {
	return &animalSpeciesService{repo: repo}
}

func (s *animalSpeciesService) List(ctx context.Context) ([]model.AnimalSpecies, error) {
	return s.repo.FindAll(ctx)
}

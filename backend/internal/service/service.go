package service

import (
	"github.com/animal-ekarte/backend/internal/repository"
)

// Service contains the business logic layer.
type Service struct {
	repo repository.PetRepository
}

// New creates a new Service with the given repository.
func New(repo repository.PetRepository) *Service {
	return &Service{repo: repo}
}

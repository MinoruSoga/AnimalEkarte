// Package service provides business logic implementations for ServiceType entity.
package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- ServiceTypeService ----

type ServiceTypeService interface { //nolint:revive // ServiceType is a domain entity name, cannot avoid stutter
	List(ctx context.Context) ([]model.ServiceType, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.ServiceType, error)
	Create(ctx context.Context, serviceType *model.ServiceType) error
	Update(ctx context.Context, serviceType *model.ServiceType) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type serviceTypeService struct {
	repo repository.ServiceTypeRepository
}

func NewServiceTypeService(repo repository.ServiceTypeRepository) ServiceTypeService {
	return &serviceTypeService{repo: repo}
}

func (s *serviceTypeService) List(ctx context.Context) ([]model.ServiceType, error) {
	return s.repo.FindAll(ctx)
}
func (s *serviceTypeService) GetByID(ctx context.Context, id uuid.UUID) (*model.ServiceType, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *serviceTypeService) Create(ctx context.Context, serviceType *model.ServiceType) error {
	return s.repo.Create(ctx, serviceType)
}
func (s *serviceTypeService) Update(ctx context.Context, serviceType *model.ServiceType) error {
	return s.repo.Update(ctx, serviceType)
}
func (s *serviceTypeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

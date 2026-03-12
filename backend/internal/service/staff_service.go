// Package service provides business logic implementations for Staff entity.
package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- StaffService ----

type StaffService interface {
	List(ctx context.Context, role *string) ([]model.Staff, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Staff, error)
	Create(ctx context.Context, staff *model.Staff) error
	Update(ctx context.Context, staff *model.Staff) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type staffService struct{ repo repository.StaffRepository }

func NewStaffService(repo repository.StaffRepository) StaffService {
	return &staffService{repo: repo}
}

func (s *staffService) List(ctx context.Context, role *string) ([]model.Staff, error) {
	return s.repo.FindAll(ctx, role)
}
func (s *staffService) GetByID(ctx context.Context, id uuid.UUID) (*model.Staff, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *staffService) Create(ctx context.Context, staff *model.Staff) error {
	return s.repo.Create(ctx, staff)
}
func (s *staffService) Update(ctx context.Context, staff *model.Staff) error {
	return s.repo.Update(ctx, staff)
}
func (s *staffService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

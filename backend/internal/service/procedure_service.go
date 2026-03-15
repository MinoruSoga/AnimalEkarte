// Package service provides business logic implementations for Procedure entity.
package service

import (
	"context"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- ProcedureService ----

type ProcedureService interface {
	List(ctx context.Context) ([]model.Procedure, error)
	GetByID(ctx context.Context, id uint64) (*model.Procedure, error)
	Create(ctx context.Context, procedure *model.Procedure) error
	Update(ctx context.Context, procedure *model.Procedure) error
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type procedureService struct {
	repo repository.ProcedureRepository
}

func NewProcedureService(repo repository.ProcedureRepository) ProcedureService {
	return &procedureService{repo: repo}
}

func (s *procedureService) List(ctx context.Context) ([]model.Procedure, error) {
	return s.repo.FindAll(ctx)
}
func (s *procedureService) GetByID(ctx context.Context, id uint64) (*model.Procedure, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *procedureService) Create(ctx context.Context, procedure *model.Procedure) error {
	return s.repo.Create(ctx, procedure)
}
func (s *procedureService) Update(ctx context.Context, procedure *model.Procedure) error {
	return s.repo.Update(ctx, procedure)
}
func (s *procedureService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *procedureService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	return s.repo.Reorder(ctx, clinicID, ids)
}

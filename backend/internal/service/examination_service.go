package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type ExaminationService interface {
	List(ctx context.Context, clinicID uint64, petID *uint64, ownerID *uint64, status *string, page, limit int) ([]model.Exam, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Exam, error)
	Create(ctx context.Context, exam *model.Exam) error
	Update(ctx context.Context, clinicID uint64, exam *model.Exam) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type examinationService struct {
	repo repository.ExaminationRepository
}

func NewExaminationService(repo repository.ExaminationRepository) ExaminationService {
	return &examinationService{repo: repo}
}

func (s *examinationService) List(ctx context.Context, clinicID uint64, petID *uint64, ownerID *uint64, status *string, page, limit int) ([]model.Exam, int64, error) {
	return s.repo.FindAll(ctx, clinicID, petID, ownerID, status, page, limit)
}

func (s *examinationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Exam, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *examinationService) Create(ctx context.Context, exam *model.Exam) error {
	return s.repo.Create(ctx, exam)
}

func (s *examinationService) Update(ctx context.Context, clinicID uint64, exam *model.Exam) error {
	return s.repo.Update(ctx, clinicID, exam)
}

func (s *examinationService) Delete(ctx context.Context, clinicID, id uint64) error {
	return s.repo.Delete(ctx, clinicID, id)
}

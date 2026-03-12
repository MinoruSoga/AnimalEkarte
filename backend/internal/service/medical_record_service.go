package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type MedicalRecordService interface {
	List(ctx context.Context, clinicID uuid.UUID, petID *uuid.UUID, ownerID *uuid.UUID, page, limit int) ([]model.MedicalRecord, int64, error)
	GetByID(ctx context.Context, clinicID, id uuid.UUID) (*model.MedicalRecord, error)
	GetByRecordNo(ctx context.Context, clinicID uuid.UUID, recordNo string) (*model.MedicalRecord, error)
	Create(ctx context.Context, record *model.MedicalRecord) error
	Update(ctx context.Context, record *model.MedicalRecord) error
	Delete(ctx context.Context, clinicID, id uuid.UUID) error
}

type medicalRecordService struct {
	repo repository.MedicalRecordRepository
}

func NewMedicalRecordService(repo repository.MedicalRecordRepository) MedicalRecordService {
	return &medicalRecordService{repo: repo}
}

func (s *medicalRecordService) List(ctx context.Context, clinicID uuid.UUID, petID *uuid.UUID, ownerID *uuid.UUID, page, limit int) ([]model.MedicalRecord, int64, error) {
	return s.repo.FindAll(ctx, clinicID, petID, ownerID, page, limit)
}

func (s *medicalRecordService) GetByID(ctx context.Context, clinicID, id uuid.UUID) (*model.MedicalRecord, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *medicalRecordService) GetByRecordNo(ctx context.Context, clinicID uuid.UUID, recordNo string) (*model.MedicalRecord, error) {
	return s.repo.FindByRecordNo(ctx, clinicID, recordNo)
}

func (s *medicalRecordService) Create(ctx context.Context, record *model.MedicalRecord) error {
	return s.repo.Create(ctx, record)
}

func (s *medicalRecordService) Update(ctx context.Context, record *model.MedicalRecord) error {
	return s.repo.Update(ctx, record)
}

func (s *medicalRecordService) Delete(ctx context.Context, clinicID, id uuid.UUID) error {
	return s.repo.Delete(ctx, clinicID, id)
}

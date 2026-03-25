package service

import (
	"context"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type MedicalRecordService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
	GetByRecordNo(ctx context.Context, clinicID uint64, recordNo string) (*model.MedicalRecord, error)
	Create(ctx context.Context, record *model.MedicalRecord) error
	Update(ctx context.Context, record *model.MedicalRecord) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type medicalRecordService struct {
	repo      repository.MedicalRecordRepository
	ownerRepo repository.OwnerRepository
	petRepo   repository.PetRepository
}

func NewMedicalRecordService(
	repo repository.MedicalRecordRepository,
	ownerRepo repository.OwnerRepository,
	petRepo repository.PetRepository,
) MedicalRecordService {
	return &medicalRecordService{
		repo:      repo,
		ownerRepo: ownerRepo,
		petRepo:   petRepo,
	}
}

func (s *medicalRecordService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error) {
	return s.repo.FindAll(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
}

func (s *medicalRecordService) GetByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *medicalRecordService) GetByRecordNo(ctx context.Context, clinicID uint64, recordNo string) (*model.MedicalRecord, error) {
	return s.repo.FindByRecordNo(ctx, clinicID, recordNo)
}

func (s *medicalRecordService) Create(ctx context.Context, record *model.MedicalRecord) error {
	return s.repo.Create(ctx, record)
}

func (s *medicalRecordService) Update(ctx context.Context, record *model.MedicalRecord) error {
	// owner_id 変更時: クリニック所属確認
	if record.OwnerID != nil {
		if _, err := s.ownerRepo.FindByID(ctx, record.ClinicID, *record.OwnerID); err != nil {
			return apperrors.WrapInvalidInput("owner not found in this clinic")
		}
	}

	// pet_id 変更時: クリニック所属確認
	if record.PetID != nil {
		if _, err := s.petRepo.FindByID(ctx, record.ClinicID, *record.PetID); err != nil {
			return apperrors.WrapInvalidInput("pet not found in this clinic")
		}
	}

	return s.repo.Update(ctx, record)
}

func (s *medicalRecordService) Delete(ctx context.Context, clinicID, id uint64) error {
	return s.repo.Delete(ctx, clinicID, id)
}

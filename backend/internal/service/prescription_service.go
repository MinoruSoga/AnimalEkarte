package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreatePrescriptionInput は処方薬記録作成の入力DTO
type CreatePrescriptionInput struct {
	PrescribedAt time.Time
	DurationDays int
}

// UpdatePrescriptionInput は処方薬記録更新の入力DTO
type UpdatePrescriptionInput struct {
	PrescribedAt *time.Time
	DurationDays *int
}

func buildPrescriptionUpdate(input *UpdatePrescriptionInput) map[string]any {
	fields := map[string]any{}
	if input.PrescribedAt != nil {
		fields["prescribed_at"] = *input.PrescribedAt
	}
	if input.DurationDays != nil {
		fields["duration_days"] = *input.DurationDays
	}
	return fields
}

// PrescriptionService は処方薬記録のビジネスロジックを定義するインターフェース
type PrescriptionService interface {
	List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Prescription, error)
	GetByID(ctx context.Context, clinicID, prescriptionID uint64) (*model.Prescription, error)
	Create(ctx context.Context, clinicID, medicalRecordID uint64, input *CreatePrescriptionInput) (*model.Prescription, error)
	Update(ctx context.Context, clinicID, medicalRecordID, prescriptionID uint64, input *UpdatePrescriptionInput) (*model.Prescription, error)
	Delete(ctx context.Context, clinicID, medicalRecordID, prescriptionID uint64) error
}

type prescriptionService struct {
	repo          repository.PrescriptionRepository
	medRecordRepo repository.MedicalRecordRepository
}

// NewPrescriptionService は PrescriptionService の実装を返す
func NewPrescriptionService(repo repository.PrescriptionRepository, medRecordRepo repository.MedicalRecordRepository) PrescriptionService {
	return &prescriptionService{repo: repo, medRecordRepo: medRecordRepo}
}

func (s *prescriptionService) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Prescription, error) {
	result, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list prescriptions", "error", err)
		return nil, apperrors.Wrap(err, "failed to list prescriptions")
	}
	return result, nil
}

func (s *prescriptionService) GetByID(ctx context.Context, clinicID, prescriptionID uint64) (*model.Prescription, error) {
	result, err := s.repo.FindByID(ctx, clinicID, prescriptionID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get prescription")
	}
	return result, nil
}

func (s *prescriptionService) Create(ctx context.Context, clinicID, medicalRecordID uint64, input *CreatePrescriptionInput) (*model.Prescription, error) {
	mr, err := s.medRecordRepo.FindByID(ctx, clinicID, medicalRecordID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get medical record")
	}
	if mr.OwnerID == nil {
		return nil, apperrors.WrapInvalidInput("medical record has no owner")
	}

	p := &model.Prescription{
		ClinicID:        clinicID,
		OwnerID:         *mr.OwnerID,
		PetID:           mr.PetID,
		MedicalRecordID: &medicalRecordID,
		PrescribedAt:    input.PrescribedAt,
		DurationDays:    input.DurationDays,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		slog.ErrorContext(ctx, "failed to create prescription", "error", err)
		return nil, apperrors.Wrap(err, "failed to create prescription")
	}
	slog.InfoContext(ctx, "prescription created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("prescription_id", p.ID),
		slog.Uint64("medical_record_id", medicalRecordID))
	return p, nil
}

func (s *prescriptionService) Update(ctx context.Context, clinicID, medicalRecordID, prescriptionID uint64, input *UpdatePrescriptionInput) (*model.Prescription, error) {
	existing, err := s.repo.FindByID(ctx, clinicID, prescriptionID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get prescription", "error", err)
		return nil, apperrors.Wrap(err, "failed to get prescription")
	}
	if existing.MedicalRecordID == nil || *existing.MedicalRecordID != medicalRecordID {
		return nil, apperrors.WrapNotFound("prescription", fmt.Sprintf("%d", prescriptionID))
	}
	fields := buildPrescriptionUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, clinicID, prescriptionID, fields); err != nil {
		slog.ErrorContext(ctx, "failed to update prescription", "error", err)
		return nil, apperrors.Wrap(err, "failed to update prescription")
	}
	slog.InfoContext(ctx, "prescription updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("prescription_id", prescriptionID))
	updated, err := s.repo.FindByID(ctx, clinicID, prescriptionID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get prescription after update", "error", err)
		return nil, apperrors.Wrap(err, "failed to get prescription after update")
	}
	return updated, nil
}

func (s *prescriptionService) Delete(ctx context.Context, clinicID, medicalRecordID, prescriptionID uint64) error {
	existing, err := s.repo.FindByID(ctx, clinicID, prescriptionID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get prescription")
	}
	if existing.MedicalRecordID == nil || *existing.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("prescription", fmt.Sprintf("%d", prescriptionID))
	}
	if err := s.repo.Delete(ctx, clinicID, prescriptionID); err != nil {
		return apperrors.Wrap(err, "failed to delete prescription")
	}
	slog.InfoContext(ctx, "prescription deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("prescription_id", prescriptionID))
	return nil
}

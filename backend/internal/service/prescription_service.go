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
	tagSyncSvc    LstepTagSyncService
	transactor    repository.Transactor
}

// NewPrescriptionService は PrescriptionService の実装を返す。transactor は BE-refactor.md X-11
// （確定と子書込の競合防止）のため、子書込を LockByIDForUpdate の行ロックと同一トランザクションに
// 収める目的で注入する。
func NewPrescriptionService(repo repository.PrescriptionRepository, medRecordRepo repository.MedicalRecordRepository, tagSyncSvc LstepTagSyncService, transactor repository.Transactor) PrescriptionService {
	return &prescriptionService{repo: repo, medRecordRepo: medRecordRepo, tagSyncSvc: tagSyncSvc, transactor: transactor}
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
	p := &model.Prescription{
		ClinicID:        clinicID,
		MedicalRecordID: &medicalRecordID,
		PrescribedAt:    input.PrescribedAt,
		DurationDays:    input.DurationDays,
	}

	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// BE-refactor.md X-11: LockByIDForUpdate の行ロックで finalize と直列化し、確定と同時の
		// 処方追加が確定済みカルテに混入する競合を防ぐ。
		mr, err := s.medRecordRepo.LockByIDForUpdate(txCtx, clinicID, medicalRecordID)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get medical record", "error", err)
			return apperrors.Wrap(err, "failed to get medical record")
		}
		if mr.OwnerID == nil {
			return apperrors.WrapInvalidInput("medical record has no owner")
		}
		// Prevent adding prescriptions to finalized medical records
		if mr.Status == model.MedicalRecordStatusFinalized {
			return apperrors.WrapConflict("確定済みの診療記録には処方を追加できません")
		}
		p.OwnerID = *mr.OwnerID
		p.PetID = mr.PetID

		if err := s.repo.Create(txCtx, p); err != nil {
			slog.ErrorContext(txCtx, "failed to create prescription", "error", err)
			return apperrors.Wrap(err, "failed to create prescription")
		}
		return nil
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "prescription created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("prescription_id", p.ID),
		slog.Uint64("medical_record_id", medicalRecordID))
	s.syncPrescriptionTag(ctx, clinicID, p.OwnerID)
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

	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// Prevent updating prescriptions for finalized medical records.
		// BE-refactor.md X-11: LockByIDForUpdate の行ロックで finalize と直列化し、確定と同時の
		// 処方編集が確定済みカルテに混入する競合を防ぐ。
		mr, err := s.medRecordRepo.LockByIDForUpdate(txCtx, clinicID, medicalRecordID)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get medical record", "error", err)
			return apperrors.Wrap(err, "failed to get medical record")
		}
		if mr != nil && mr.Status == model.MedicalRecordStatusFinalized {
			return apperrors.WrapConflict("確定済みの診療記録の処方は編集できません")
		}
		fields := buildPrescriptionUpdate(input)
		if len(fields) == 0 {
			return apperrors.WrapInvalidInput("at least one field must be provided")
		}
		if err := s.repo.Update(txCtx, clinicID, prescriptionID, fields); err != nil {
			slog.ErrorContext(txCtx, "failed to update prescription", "error", err)
			return apperrors.Wrap(err, "failed to update prescription")
		}
		return nil
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "prescription updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("prescription_id", prescriptionID))
	updated, err := s.repo.FindByID(ctx, clinicID, prescriptionID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get prescription after update", "error", err)
		return nil, apperrors.Wrap(err, "failed to get prescription after update")
	}
	s.syncPrescriptionTag(ctx, clinicID, updated.OwnerID)
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
	// Prevent deleting prescriptions for finalized medical records
	mr, err := s.medRecordRepo.FindByID(ctx, clinicID, medicalRecordID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get medical record")
	}
	if mr != nil && mr.Status == model.MedicalRecordStatusFinalized {
		return apperrors.WrapConflict("確定済みの診療記録の処方は削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, prescriptionID); err != nil {
		slog.ErrorContext(ctx, "failed to delete prescription", "error", err, "clinic_id", clinicID, "prescription_id", prescriptionID)
		return apperrors.Wrap(err, "failed to delete prescription")
	}
	slog.InfoContext(ctx, "prescription deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("prescription_id", prescriptionID))
	s.syncPrescriptionTag(ctx, clinicID, existing.OwnerID)
	return nil
}

func (s *prescriptionService) syncPrescriptionTag(ctx context.Context, clinicID, ownerID uint64) {
	if s.tagSyncSvc == nil {
		return
	}
	if err := s.tagSyncSvc.SyncPrescriptionTag(ctx, clinicID, ownerID); err != nil {
		slog.ErrorContext(ctx, "failed to sync prescription tag", "error", err, "clinic_id", clinicID, "owner_id", ownerID)
	}
}

package pet

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// CreateChronicConditionInput は慢性疾患フラグ作成の入力（BE-012）。
type CreateChronicConditionInput struct {
	ConditionCode string    `json:"condition_code"`
	ConditionName string    `json:"condition_name"`
	DiagnosedAt   time.Time `json:"diagnosed_at"`
	Notes         *string   `json:"notes"`
	IsActive      bool      `json:"is_active"`
}

// UpdateChronicConditionInput は慢性疾患フラグ更新の入力（BE-012）。
type UpdateChronicConditionInput struct {
	ConditionCode *string    `json:"condition_code"`
	ConditionName *string    `json:"condition_name"`
	DiagnosedAt   *time.Time `json:"diagnosed_at"`
	Notes         *string    `json:"notes"`
	IsActive      *bool      `json:"is_active"`
}

func buildChronicConditionUpdateFields(input UpdateChronicConditionInput) map[string]any {
	fields := make(map[string]any)
	if input.ConditionCode != nil {
		fields["condition_code"] = *input.ConditionCode
	}
	if input.ConditionName != nil {
		fields["condition_name"] = *input.ConditionName
	}
	if input.DiagnosedAt != nil {
		fields["diagnosed_at"] = *input.DiagnosedAt
	}
	if input.Notes != nil {
		fields["notes"] = *input.Notes
	}
	if input.IsActive != nil {
		fields["is_active"] = *input.IsActive
	}
	return fields
}

// ChronicConditionService は慢性疾患フラグの業務ロジックインターフェース（BE-012）。
type ChronicConditionService interface {
	List(ctx context.Context, clinicID, petID uint64) ([]model.PetChronicCondition, error)
	Create(ctx context.Context, clinicID, petID uint64, input CreateChronicConditionInput) (*model.PetChronicCondition, error)
	Update(ctx context.Context, clinicID, petID, id uint64, input UpdateChronicConditionInput) (*model.PetChronicCondition, error)
	Delete(ctx context.Context, clinicID, petID, id uint64) error
}

type chronicConditionService struct {
	repo       ChronicConditionRepository
	petRepo    PetFinder
	tagSyncSvc ChronicConditionTagSynchronizer
}

// NewChronicConditionService は ChronicConditionService を初期化して返す。
func NewChronicConditionService(
	repo ChronicConditionRepository,
	petRepo PetFinder,
	tagSyncSvc ChronicConditionTagSynchronizer,
) ChronicConditionService {
	return &chronicConditionService{repo: repo, petRepo: petRepo, tagSyncSvc: tagSyncSvc}
}

func (s *chronicConditionService) List(ctx context.Context, clinicID, petID uint64) ([]model.PetChronicCondition, error) {
	records, err := s.repo.FindByPetID(ctx, clinicID, petID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list chronic conditions")
	}
	return records, nil
}

func (s *chronicConditionService) Create(ctx context.Context, clinicID, petID uint64, input CreateChronicConditionInput) (*model.PetChronicCondition, error) {
	pet, err := s.petRepo.FindByID(ctx, clinicID, petID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find pet")
	}

	record := &model.PetChronicCondition{
		ClinicID:      clinicID,
		PetID:         petID,
		ConditionCode: input.ConditionCode,
		ConditionName: input.ConditionName,
		DiagnosedAt:   input.DiagnosedAt,
		Notes:         input.Notes,
		IsActive:      input.IsActive,
	}
	if err := s.repo.Create(ctx, record); err != nil {
		return nil, apperrors.Wrap(err, "failed to create chronic condition")
	}

	s.syncTags(ctx, clinicID, pet.OwnerID)
	return record, nil
}

func (s *chronicConditionService) Update(ctx context.Context, clinicID, petID, id uint64, input UpdateChronicConditionInput) (*model.PetChronicCondition, error) {
	pet, err := s.petRepo.FindByID(ctx, clinicID, petID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find pet")
	}

	if _, err := s.repo.FindByID(ctx, clinicID, petID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to find chronic condition")
	}

	fields := buildChronicConditionUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgAtLeastOneField)
	}
	if err := s.repo.Update(ctx, clinicID, petID, id, input); err != nil {
		return nil, apperrors.Wrap(err, "failed to update chronic condition")
	}

	updated, err := s.repo.FindByID(ctx, clinicID, petID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to reload chronic condition")
	}

	s.syncTags(ctx, clinicID, pet.OwnerID)
	return updated, nil
}

func (s *chronicConditionService) Delete(ctx context.Context, clinicID, petID, id uint64) error {
	pet, err := s.petRepo.FindByID(ctx, clinicID, petID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find pet")
	}

	if _, err := s.repo.FindByID(ctx, clinicID, petID, id); err != nil {
		return apperrors.Wrap(err, "failed to find chronic condition")
	}

	if err := s.repo.Delete(ctx, clinicID, petID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete chronic condition")
	}

	s.syncTags(ctx, clinicID, pet.OwnerID)
	return nil
}

func (s *chronicConditionService) syncTags(ctx context.Context, clinicID, ownerID uint64) {
	codes, err := s.repo.FindActiveConditionCodesByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.WarnContext(ctx, "failed to get active condition codes (non-fatal)", "error", err)
		return
	}
	if err := s.tagSyncSvc.SyncChronicConditionTags(ctx, clinicID, ownerID, codes); err != nil {
		slog.WarnContext(ctx, "failed to sync chronic condition tags (non-fatal)", "owner_id", ownerID, "error", err)
	}
}

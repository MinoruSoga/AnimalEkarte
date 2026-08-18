package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- CageService ----

// CreateCageInput はケージ作成の入力DTO
type CreateCageInput struct {
	Name        string
	CageType    string
	CageSize    string
	Price       *int64
	IsActive    bool
	Description string
	SortOrder   int
}

// UpdateCageInput はケージ更新のサービス入力 DTO
type UpdateCageInput struct {
	Name        *string
	CageType    *string
	CageSize    *string
	Price       *int64
	IsActive    *bool
	Description *string
	SortOrder   *int
}

const (
	colCageName        = "name"
	colCageCageType    = "cage_type"
	colCageCageSize    = "cage_size"
	colCagePrice       = "price"
	colCageIsActive    = "is_active"
	colCageDescription = "description"
	colCageSortOrder   = "sort_order"
)

func buildCageUpdate(input *UpdateCageInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colCageName] = *input.Name
	}
	if input.CageType != nil {
		fields[colCageCageType] = model.CageType(*input.CageType)
	}
	if input.CageSize != nil {
		fields[colCageCageSize] = model.CageSize(*input.CageSize)
	}
	if input.Price != nil {
		fields[colCagePrice] = *input.Price
	}
	if input.IsActive != nil {
		fields[colCageIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colCageDescription] = *input.Description
	}
	if input.SortOrder != nil {
		fields[colCageSortOrder] = *input.SortOrder
	}
	return fields
}

type CageService interface {
	List(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error)
	Create(ctx context.Context, clinicID uint64, input *CreateCageInput) (*model.Cage, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateCageInput) (*model.Cage, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type cageService struct {
	repo       CageRepository
	transactor Transactor
}

// NewCageService constructs CageService.
// transactor is required for Delete: lock + usage check + soft-delete share one
// ambient transaction (SEC-CS-F13; merchandise/procedure atomic-delete pattern).
func NewCageService(repo CageRepository, transactor Transactor) CageService {
	return &cageService{repo: repo, transactor: transactor}
}

func (s *cageService) List(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error) {
	result, err := s.repo.FindAll(ctx, clinicID, cageType)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list cage", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list cage")
	}
	return result, nil
}
func (s *cageService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get cage", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get cage")
	}
	return result, nil
}
func (s *cageService) Create(ctx context.Context, clinicID uint64, input *CreateCageInput) (*model.Cage, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	// Service は Handler 以外から直接呼ばれる可能性があるため防御的バリデーションを維持する
	if err := validateCageType(input.CageType); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate cage type")
	}
	if err := validateCageSize(input.CageSize); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate cage size")
	}
	cage := &model.Cage{
		ClinicID:    clinicID,
		Name:        input.Name,
		CageType:    model.CageType(input.CageType),
		CageSize:    model.CageSize(input.CageSize),
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		SortOrder:   input.SortOrder,
	}
	if err := s.repo.Create(ctx, cage); err != nil {
		if conflict := apperrors.AsNameUniqueConflict(
			err,
			input.Name,
			apperrors.ConstraintCageName,
			apperrors.CodeCageNameConflict,
		); conflict != nil {
			return nil, conflict
		}
		slog.ErrorContext(ctx, "failed to create cage", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to create cage")
	}
	slog.InfoContext(ctx, "cage created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("cage_id", cage.ID))
	return cage, nil
}
func (s *cageService) Update(ctx context.Context, clinicID, id uint64, input *UpdateCageInput) (*model.Cage, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to get cage", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get cage")
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	// Service は Handler 以外から直接呼ばれる可能性があるため防御的バリデーションを維持する
	if input.CageType != nil {
		if err := validateCageType(*input.CageType); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate cage type")
		}
	}
	if input.CageSize != nil {
		if err := validateCageSize(*input.CageSize); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate cage size")
		}
	}
	fields := buildCageUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}
	cage, err := s.repo.Update(ctx, clinicID, id, fields)
	if err != nil {
		nameForConflict := ""
		if input.Name != nil {
			nameForConflict = *input.Name
		}
		if conflict := apperrors.AsNameUniqueConflict(
			err,
			nameForConflict,
			apperrors.ConstraintCageName,
			apperrors.CodeCageNameConflict,
		); conflict != nil {
			return nil, conflict
		}
		slog.ErrorContext(ctx, "failed to update cage", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to update cage")
	}
	slog.InfoContext(ctx, "cage updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("cage_id", id))
	return cage, nil
}
func (s *cageService) Delete(ctx context.Context, clinicID, id uint64) error {
	// SEC-CS-F13: FOR UPDATE → usage count → soft-delete in one ambient tx so a
	// concurrent hospitalization assignment cannot sneak between check and delete.
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("cage write transaction dependency is required")
	}
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := s.repo.LockByIDForUpdate(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to get cage")
		}
		count, err := s.repo.CountUsageByCageID(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to check cage usage", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to check cage usage")
		}
		if count > 0 {
			return apperrors.WrapConflict("このケージは入院データで使用中のため削除できません")
		}
		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			slog.ErrorContext(txCtx, "failed to delete cage", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to delete cage")
		}
		return nil
	}); err != nil {
		return err
	}
	slog.InfoContext(ctx, "cage deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("cage_id", id))
	return nil
}

func (s *cageService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	// SEC-CS-F04: bound cardinality + uniqueness before fan-out UPDATE per id.
	if err := httpapi.ValidateReorderIDs(ids); err != nil {
		return err
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		slog.ErrorContext(ctx, "failed to reorder cage", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to reorder cage")
	}
	slog.InfoContext(ctx, "cage reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}

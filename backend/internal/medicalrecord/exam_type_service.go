package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	colExamTypeName           = "name"
	colExamTypePrice          = "price"
	colExamTypeIsActive       = "is_active"
	colExamTypeDescription    = "description"
	colExamTypeParentID       = "parent_id"
	colExamTypeSortOrder      = "sort_order"
	colExamTypeIsNonInsurance = "is_non_insurance"
)

func buildExamTypeUpdate(input *UpdateExamTypeInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colExamTypeName] = *input.Name
	}
	if input.Price != nil {
		fields[colExamTypePrice] = *input.Price
	}
	if input.IsActive != nil {
		fields[colExamTypeIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colExamTypeDescription] = *input.Description
	}
	setNullableUint64Field(fields, colExamTypeParentID, input.ClearParentID, input.ParentID)
	if input.SortOrder != nil {
		fields[colExamTypeSortOrder] = *input.SortOrder
	}
	if input.IsNonInsurance != nil {
		fields[colExamTypeIsNonInsurance] = *input.IsNonInsurance
	}
	return fields
}

// ---- ExaminationTypeService ----

// CreateExamTypeInput は検査種別作成の入力DTO
type CreateExamTypeInput struct {
	Name           string
	Price          *int64
	IsActive       bool
	Description    string
	ParentID       *uint64
	SortOrder      int
	IsNonInsurance bool
}

// UpdateExamTypeInput は検査種別更新のサービス入力 DTO
type UpdateExamTypeInput struct {
	Name           *string
	Price          *int64
	IsActive       *bool
	Description    *string
	ParentID       *uint64
	ClearParentID  bool
	SortOrder      *int
	IsNonInsurance *bool
}

type ExaminationTypeService interface {
	List(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error)
	Create(ctx context.Context, clinicID uint64, input *CreateExamTypeInput) (*model.ExaminationType, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateExamTypeInput) (*model.ExaminationType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CreateField(ctx context.Context, clinicID uint64, command *CreateExamTypeFieldCommand) (*model.ExamTypeField, error)
	UpdateField(ctx context.Context, clinicID, examTypeID, fieldID uint64, input *UpdateExamTypeFieldInput) (*ExamTypeFieldResult, error)
	DeleteField(ctx context.Context, clinicID, examTypeID, fieldID uint64) error
	ReorderFields(ctx context.Context, clinicID, examTypeID uint64, ids []uint64) error
	ReplaceReferenceRanges(ctx context.Context, clinicID, examTypeID uint64, command *ReplaceReferenceRangesCommand) (*ExamTypeFieldResult, error)
	ListReferenceRanges(ctx context.Context, clinicID uint64, fieldIDs []uint64) (map[uint64][]model.ExamReferenceRange, error)
}

type examTypeService struct {
	repo       ExamTypeRepository
	transactor Transactor
}

// NewExamTypeService constructs ExaminationTypeService.
// transactor is required so Create/Update/Delete and field writes share ambient transactions
// (MRB-07: compile-time wiring; nil is fail-closed as 500 via withTx).
func NewExamTypeService(repo ExamTypeRepository, transactor Transactor) ExaminationTypeService {
	return &examTypeService{repo: repo, transactor: transactor}
}

func (s *examTypeService) List(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list exam types", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list exam types")
	}
	return items, nil
}
func (s *examTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get exam type", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get exam type")
	}
	return result, nil
}
func (s *examTypeService) Create(ctx context.Context, clinicID uint64, input *CreateExamTypeInput) (*model.ExaminationType, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	// MRB-08 / X-05: parent_id 検証と Create を同一 WithTx に収め、FindByID の FOR SHARE を発火させる。
	exType := &model.ExaminationType{
		ClinicID:       clinicID,
		Name:           input.Name,
		Price:          input.Price,
		IsActive:       input.IsActive,
		Description:    input.Description,
		ParentID:       input.ParentID,
		SortOrder:      input.SortOrder,
		IsNonInsurance: input.IsNonInsurance,
	}
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		if err := s.validateParentOwnership(txCtx, clinicID, input.ParentID); err != nil {
			return err
		}
		if err := s.repo.Create(txCtx, exType); err != nil {
			if conflict := apperrors.AsNameUniqueConflict(
				err,
				input.Name,
				apperrors.ConstraintExamTypeName,
				apperrors.CodeExamTypeNameConflict,
			); conflict != nil {
				return conflict
			}
			slog.ErrorContext(txCtx, "failed to create exam type", "error", err, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to create exam type")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "exam type created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("exam_type_id", exType.ID))
	return exType, nil
}
func (s *examTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateExamTypeInput) (*model.ExaminationType, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	fields := buildExamTypeUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}
	// MRB-08: 存在確認・parent 検証・Update を同一 tx に収める。
	var exType *model.ExaminationType
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		if _, err := s.repo.FindByID(txCtx, clinicID, id); err != nil {
			slog.ErrorContext(txCtx, "failed to get exam type", "error", err)
			return apperrors.Wrap(err, "failed to get exam type")
		}
		if err := s.validateParentOwnership(txCtx, clinicID, input.ParentID); err != nil {
			return err
		}
		updated, err := s.repo.Update(txCtx, clinicID, id, fields)
		if err != nil {
			nameForConflict := ""
			if input.Name != nil {
				nameForConflict = *input.Name
			}
			if conflict := apperrors.AsNameUniqueConflict(
				err,
				nameForConflict,
				apperrors.ConstraintExamTypeName,
				apperrors.CodeExamTypeNameConflict,
			); conflict != nil {
				return conflict
			}
			slog.ErrorContext(txCtx, "failed to update exam type", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to update exam type")
		}
		exType = updated
		return nil
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "exam type updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("exam_type_id", id))
	return exType, nil
}
func (s *examTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	// MRB-08: 依存 count と Delete を同一 tx に収め、並行使用中の silent 削除を防ぐ。
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		if _, err := s.repo.FindByID(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to find exam type")
		}
		childCount, err := s.repo.CountChildrenByParentID(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to check exam type children", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to check exam type children")
		}
		if childCount > 0 {
			return apperrors.WrapConflict("この検査種別にはサブ種別が登録されているため削除できません")
		}
		count, err := s.repo.CountUsageByExamTypeID(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to check exam type dependencies", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to check exam type dependencies")
		}
		if count > 0 {
			return apperrors.WrapConflict("この検査種別は検査記録で使用中のため削除できません")
		}
		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			slog.ErrorContext(txCtx, "failed to delete exam type", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to delete exam type")
		}
		return nil
	}); err != nil {
		return err
	}
	slog.InfoContext(ctx, "exam type deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("exam_type_id", id))
	return nil
}

func (s *examTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(errMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		slog.ErrorContext(ctx, "failed to reorder exam types", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to reorder exam types")
	}
	slog.InfoContext(ctx, "exam type reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}

// validateParentOwnership verifies a request-supplied parent_id belongs to the caller's
// clinic before it is persisted (X-14 self-ref master FK guard).
func (s *examTypeService) validateParentOwnership(ctx context.Context, clinicID uint64, parentID *uint64) error {
	return validateOwnedMasterFK(ctx, "parent exam type", clinicID, parentID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.repo.FindByID(actx, cid, mid)
			return err
		})
}

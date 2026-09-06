package lstep

// File organization (this file only, non-mandatory convention — order is not enforced by any
// gate; see service/CLAUDE.md): const → buildFunc → interface → struct → constructor → methods

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// LstepTagConfigService は自動管理タグ設定のCRUDを担う。
type LstepTagConfigService interface {
	// --- auto_managed_prefixes ---
	ListAutoManagedPrefixes(ctx context.Context) ([]*model.LstepAutoManagedPrefix, error)
	CreateAutoManagedPrefix(ctx context.Context, input CreateAutoManagedPrefixInput) (*model.LstepAutoManagedPrefix, error)
	DeleteAutoManagedPrefix(ctx context.Context, id uint64) error

	// --- condition_tag_mappings ---
	ListConditionTagMappings(ctx context.Context) ([]*model.LstepConditionTagMapping, error)
	CreateConditionTagMapping(ctx context.Context, input CreateConditionTagMappingInput) (*model.LstepConditionTagMapping, error)
	DeleteConditionTagMapping(ctx context.Context, id uint64) error

	// --- send_purpose_tag_prefixes ---
	ListSendPurposeTagPrefixes(ctx context.Context) ([]*model.LstepSendPurposeTagPrefix, error)
	CreateSendPurposeTagPrefix(ctx context.Context, input CreateSendPurposeTagPrefixInput) (*model.LstepSendPurposeTagPrefix, error)
	DeleteSendPurposeTagPrefix(ctx context.Context, id uint64) error
}

// CreateAutoManagedPrefixInput は自動管理プレフィックス作成入力。
type CreateAutoManagedPrefixInput struct {
	Prefix      string
	Category    string
	Description *string
}

// CreateConditionTagMappingInput は慢性疾患コードマッピング作成入力。
type CreateConditionTagMappingInput struct {
	ConditionCode string
	TagName       string
	Description   *string
}

// CreateSendPurposeTagPrefixInput はLINE送信目的プレフィックス作成入力。
type CreateSendPurposeTagPrefixInput struct {
	Purpose     string
	TagPrefix   string
	Description *string
}

type lstepTagConfigService struct {
	repo LstepTagConfigRepository
}

// NewLstepTagConfigService はサービスを生成する。
func NewLstepTagConfigService(repo LstepTagConfigRepository) LstepTagConfigService {
	return &lstepTagConfigService{repo: repo}
}

func (s *lstepTagConfigService) ListAutoManagedPrefixes(ctx context.Context) ([]*model.LstepAutoManagedPrefix, error) {
	items, err := s.repo.FindAllAutoManagedPrefixes(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list auto managed prefixes")
	}
	return items, nil
}

func (s *lstepTagConfigService) CreateAutoManagedPrefix(ctx context.Context, input CreateAutoManagedPrefixInput) (*model.LstepAutoManagedPrefix, error) {
	// LSA-14: service boundary mirrors binding oneof so non-HTTP callers cannot skip enum checks.
	switch input.Category {
	case "B", "C1", "C2", "C3":
	default:
		return nil, apperrors.WrapInvalidInput("category must be one of B, C1, C2, C3")
	}
	m := &model.LstepAutoManagedPrefix{
		Prefix:      input.Prefix,
		Category:    input.Category,
		Description: input.Description,
	}
	if err := s.repo.CreateAutoManagedPrefix(ctx, m); err != nil {
		// BUG-026: map measured prefix unique_violation to stable domain code + value echo.
		if conflict := apperrors.AsNameUniqueConflict(
			err,
			input.Prefix,
			apperrors.ConstraintLstepAutoManagedPrefix,
			apperrors.CodeLstepAutoManagedPrefixConflict,
		); conflict != nil {
			return nil, conflict
		}
		return nil, apperrors.Wrap(err, "failed to create auto managed prefix")
	}
	return m, nil
}

func (s *lstepTagConfigService) DeleteAutoManagedPrefix(ctx context.Context, id uint64) error {
	if err := s.repo.DeleteAutoManagedPrefix(ctx, id); err != nil {
		return apperrors.Wrap(err, "failed to delete auto managed prefix")
	}
	return nil
}

func (s *lstepTagConfigService) ListConditionTagMappings(ctx context.Context) ([]*model.LstepConditionTagMapping, error) {
	items, err := s.repo.FindAllConditionTagMappings(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list condition tag mappings")
	}
	return items, nil
}

func (s *lstepTagConfigService) CreateConditionTagMapping(ctx context.Context, input CreateConditionTagMappingInput) (*model.LstepConditionTagMapping, error) {
	m := &model.LstepConditionTagMapping{
		ConditionCode: input.ConditionCode,
		TagName:       input.TagName,
		Description:   input.Description,
	}
	if err := s.repo.CreateConditionTagMapping(ctx, m); err != nil {
		if apperrors.IsAlreadyExists(err) {
			return nil, apperrors.WrapAlreadyExists("lstep_condition_tag_mapping", input.ConditionCode)
		}
		return nil, apperrors.Wrap(err, "failed to create condition tag mapping")
	}
	return m, nil
}

func (s *lstepTagConfigService) DeleteConditionTagMapping(ctx context.Context, id uint64) error {
	if err := s.repo.DeleteConditionTagMapping(ctx, id); err != nil {
		return apperrors.Wrap(err, "failed to delete condition tag mapping")
	}
	return nil
}

func (s *lstepTagConfigService) ListSendPurposeTagPrefixes(ctx context.Context) ([]*model.LstepSendPurposeTagPrefix, error) {
	items, err := s.repo.FindAllSendPurposeTagPrefixes(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list send purpose tag prefixes")
	}
	return items, nil
}

func (s *lstepTagConfigService) CreateSendPurposeTagPrefix(ctx context.Context, input CreateSendPurposeTagPrefixInput) (*model.LstepSendPurposeTagPrefix, error) {
	m := &model.LstepSendPurposeTagPrefix{
		Purpose:     input.Purpose,
		TagPrefix:   input.TagPrefix,
		Description: input.Description,
	}
	if err := s.repo.CreateSendPurposeTagPrefix(ctx, m); err != nil {
		return nil, apperrors.Wrap(err, "failed to create send purpose tag prefix")
	}
	return m, nil
}

func (s *lstepTagConfigService) DeleteSendPurposeTagPrefix(ctx context.Context, id uint64) error {
	if err := s.repo.DeleteSendPurposeTagPrefix(ctx, id); err != nil {
		return apperrors.Wrap(err, "failed to delete send purpose tag prefix")
	}
	return nil
}
